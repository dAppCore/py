// Package tier2 runs host CPython as CorePy's subprocess escape hatch.
//
// Tier 1 remains the embedded gpython path. Tier 2 is intentionally explicit:
// callers choose a host Python executable, get stdout/stderr/exit semantics
// back, and can stream output through Core mediums or CLI writers.
package tier2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const EnvPython = "COREPY_TIER2_PYTHON"

// Options controls a Tier 2 CPython subprocess.
type Options struct {
	Python           string
	WorkingDirectory string
	Environment      map[string]string
	PythonPath       []string
	Timeout          time.Duration
	Stdout           io.Writer
	Stderr           io.Writer
}

// Runner executes Python source or files with consistent process semantics.
type Runner struct {
	options Options
}

// Result captures a completed or interrupted Tier 2 subprocess.
type Result struct {
	Command  []string
	Stdout   string
	Stderr   string
	ExitCode int
	TimedOut bool
}

// OK reports whether the subprocess completed successfully.
func (result Result) OK() bool {
	return !result.TimedOut && result.ExitCode == 0
}

// ExitError reports a subprocess failure while preserving captured output.
type ExitError struct {
	Result  Result
	Timeout time.Duration
	Cause   error
}

func (err ExitError) Error() string {
	command := compactCommand(err.Result.Command)
	if err.Result.TimedOut {
		if err.Timeout > 0 {
			return fmt.Sprintf("corepy tier2: command timed out after %s: %s", err.Timeout, command)
		}
		return fmt.Sprintf("corepy tier2: command timed out: %s", command)
	}

	if stderr := firstLine(err.Result.Stderr); stderr != "" {
		return fmt.Sprintf("corepy tier2: command exited with status %d: %s", err.Result.ExitCode, stderr)
	}
	if err.Cause != nil {
		return fmt.Sprintf("corepy tier2: command exited with status %d: %v", err.Result.ExitCode, err.Cause)
	}
	return fmt.Sprintf("corepy tier2: command exited with status %d: %s", err.Result.ExitCode, command)
}

func (err ExitError) Unwrap() error {
	return err.Cause
}

// NewRunner creates a Tier 2 subprocess runner.
func NewRunner(options Options) *Runner {
	return &Runner{options: options}
}

// ResolvePython locates the Python executable for Tier 2.
func ResolvePython(requested string) (string, error) {
	if strings.TrimSpace(requested) != "" {
		path, err := exec.LookPath(strings.TrimSpace(requested))
		if err != nil {
			return "", fmt.Errorf("corepy tier2: python executable %q not found", requested)
		}
		return path, nil
	}
	if envPython := strings.TrimSpace(os.Getenv(EnvPython)); envPython != "" {
		path, err := exec.LookPath(envPython)
		if err != nil {
			return "", fmt.Errorf("corepy tier2: python executable %q from %s not found", envPython, EnvPython)
		}
		return path, nil
	}

	candidates := []string{"python3.14", "python3.13", "python3", "python"}
	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}

		path, err := exec.LookPath(candidate)
		if err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("corepy tier2: no Python executable found; set %s", EnvPython)
}

// RunSource executes a Python source string.
func (runner *Runner) RunSource(ctx context.Context, source string, arguments ...string) (Result, error) {
	command := append([]string{"-c", source}, arguments...)
	return runner.run(ctx, command)
}

// RunFile executes a Python file.
func (runner *Runner) RunFile(ctx context.Context, filename string, arguments ...string) (Result, error) {
	if strings.TrimSpace(filename) == "" {
		return Result{}, fmt.Errorf("corepy tier2: filename cannot be empty")
	}
	command := append([]string{filename}, arguments...)
	return runner.run(ctx, command)
}

func (runner *Runner) run(ctx context.Context, pythonArguments []string) (Result, error) {
	python, err := ResolvePython(runner.options.Python)
	if err != nil {
		return Result{}, err
	}

	if ctx == nil {
		ctx = context.Background()
	}
	timeout := runner.options.Timeout
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	commandArguments := append([]string(nil), pythonArguments...)
	cmd := exec.CommandContext(ctx, python, commandArguments...)
	if runner.options.WorkingDirectory != "" {
		cmd.Dir = runner.options.WorkingDirectory
	}
	cmd.Env = runner.environment()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = outputWriter(&stdout, runner.options.Stdout)
	cmd.Stderr = outputWriter(&stderr, runner.options.Stderr)

	runErr := cmd.Run()
	timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded)
	exitCode := 0
	if runErr != nil {
		exitCode = -1
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}

	result := Result{
		Command:  append([]string{python}, commandArguments...),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		TimedOut: timedOut,
	}
	if runErr != nil || timedOut {
		return result, ExitError{Result: result, Timeout: timeout, Cause: runErr}
	}
	return result, nil
}

func (runner *Runner) environment() []string {
	env := os.Environ()
	if len(runner.options.PythonPath) > 0 {
		existing := os.Getenv("PYTHONPATH")
		parts := append([]string(nil), runner.options.PythonPath...)
		if existing != "" {
			parts = append(parts, existing)
		}
		env = append(env, "PYTHONPATH="+strings.Join(parts, string(os.PathListSeparator)))
	}
	if len(runner.options.Environment) == 0 {
		return env
	}

	keys := make([]string, 0, len(runner.options.Environment))
	for key := range runner.options.Environment {
		keys = append(keys, key)
	}
	sortStrings(keys)
	for _, key := range keys {
		env = append(env, key+"="+runner.options.Environment[key])
	}
	return env
}

func outputWriter(capture *bytes.Buffer, stream io.Writer) io.Writer {
	if stream == nil {
		return capture
	}
	return io.MultiWriter(capture, stream)
}

func compactCommand(command []string) string {
	const maxPart = 80
	parts := make([]string, 0, len(command))
	for _, part := range command {
		cleaned := strings.ReplaceAll(part, "\n", `\n`)
		if len(cleaned) > maxPart {
			cleaned = cleaned[:maxPart] + "..."
		}
		if strings.ContainsAny(cleaned, " \t\n\"'") {
			cleaned = strconvQuote(cleaned)
		}
		parts = append(parts, cleaned)
	}
	return strings.Join(parts, " ")
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if index := strings.IndexByte(value, '\n'); index != -1 {
		return strings.TrimSpace(value[:index])
	}
	return value
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

func strconvQuote(value string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `"`, `\"`) + `"`
}

// LocalPythonPath returns the repository-local CPython package path when this
// command is executed from the CorePy source tree.
func LocalPythonPath(start string) (string, bool) {
	if strings.TrimSpace(start) == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", false
		}
		start = cwd
	}

	current, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	for {
		candidate := filepath.Join(current, "py", "core", "__init__.py")
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Join(current, "py"), true
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}
