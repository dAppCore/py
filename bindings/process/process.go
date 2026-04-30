package process

import (
	"bytes"
	"context"
	"fmt" // AX-6-exception: process bootstrap preserves wrapped stderr context from exec failures.
	"os"
	"os/exec" // AX-6-exception: this binding provides the process primitive before go-process is registered.
	"sort"
	"strings" // AX-6-exception: process bootstrap trims stderr captured from os/exec.
	"sync"
	"time"

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

var (
	defaultCoreOnce sync.Once
	defaultCore     *core.Core
)

type executionOptions struct {
	Directory string
	Env       []string
	Timeout   time.Duration
	Check     bool
}

type executionResult struct {
	Command  []string
	Stdout   string
	Stderr   string
	ExitCode int
	TimedOut bool
}

// Register exposes Process bindings backed by core.Process.
//
//	process.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.process",
		Documentation: "Process helpers backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"run":          run,
			"run_in":       runIn,
			"run_with_env": runWithEnv,
			"run_result":   runResult,
			"exists":       exists,
		},
	})
}

func run(arguments ...any) (any, error) {
	positional, keywordArguments := runtime.SplitKeywordArguments(arguments)
	options := executionOptions{Check: true}
	if err := applyKeywordArguments(&options, "core.process.run", keywordArguments, "directory", "env", "timeout", "check"); err != nil {
		return nil, err
	}

	command, processArguments, err := commandArgs(positional, 0, "core.process.run")
	if err != nil {
		return nil, err
	}

	result, err := executeProcess(context.Background(), command, processArguments, options)
	if err != nil {
		return nil, err
	}
	return result.Stdout, nil
}

func runIn(arguments ...any) (any, error) {
	positional, keywordArguments := runtime.SplitKeywordArguments(arguments)
	options := executionOptions{Check: true}
	if err := applyKeywordArguments(&options, "core.process.run_in", keywordArguments, "env", "timeout", "check"); err != nil {
		return nil, err
	}

	directory, err := typemap.ExpectString(positional, 0, "core.process.run_in")
	if err != nil {
		return nil, err
	}
	options.Directory = directory
	command, processArguments, err := commandArgs(positional, 1, "core.process.run_in")
	if err != nil {
		return nil, err
	}

	result, err := executeProcess(context.Background(), command, processArguments, options)
	if err != nil {
		return nil, err
	}
	return result.Stdout, nil
}

func runWithEnv(arguments ...any) (any, error) {
	positional, keywordArguments := runtime.SplitKeywordArguments(arguments)
	options := executionOptions{Check: true}
	if err := applyKeywordArguments(&options, "core.process.run_with_env", keywordArguments, "timeout", "check"); err != nil {
		return nil, err
	}

	directory, err := typemap.ExpectString(positional, 0, "core.process.run_with_env")
	if err != nil {
		return nil, err
	}
	env, err := envList(positional, 1, "core.process.run_with_env")
	if err != nil {
		return nil, err
	}
	options.Directory = directory
	options.Env = env
	command, processArguments, err := commandArgs(positional, 2, "core.process.run_with_env")
	if err != nil {
		return nil, err
	}

	result, err := executeProcess(context.Background(), command, processArguments, options)
	if err != nil {
		return nil, err
	}
	return result.Stdout, nil
}

func runResult(arguments ...any) (any, error) {
	positional, keywordArguments := runtime.SplitKeywordArguments(arguments)
	options := executionOptions{}
	if err := applyKeywordArguments(&options, "core.process.run_result", keywordArguments, "directory", "env", "timeout", "check"); err != nil {
		return nil, err
	}

	command, processArguments, err := commandArgs(positional, 0, "core.process.run_result")
	if err != nil {
		return nil, err
	}

	result, err := executeProcess(context.Background(), command, processArguments, options)
	if err != nil {
		return nil, err
	}
	return result.Map(), nil
}

func exists(arguments ...any) (any, error) {
	return processCore().Process().Exists(), nil
}

func processCore() *core.Core {
	defaultCoreOnce.Do(func() {
		defaultCore = core.New()
		defaultCore.Action("process.run", handleRun)
	})
	return defaultCore
}

func handleRun(ctx context.Context, options core.Options) core.Result {
	command := options.String("command")
	if command == "" {
		return core.Result{Value: core.E("core.process.run", "command is required", nil), OK: false}
	}

	result, err := executeProcess(ctx, command, optionStrings(options.Get("args")), executionOptions{
		Directory: options.String("dir"),
		Env:       optionStrings(options.Get("env")),
		Check:     true,
	})
	if err != nil {
		return core.Result{
			Value: core.E("core.process.run", core.Concat("command failed: ", command), err),
			OK:    false,
		}
	}

	return core.Result{Value: result.Stdout, OK: true}
}

func executeProcess(ctx context.Context, command string, arguments []string, options executionOptions) (executionResult, error) {
	if command == "" {
		return executionResult{}, fmt.Errorf("core.process: command is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, command, arguments...)
	if options.Directory != "" {
		cmd.Dir = options.Directory
	}
	if len(options.Env) > 0 {
		cmd.Env = append(os.Environ(), options.Env...)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	timedOut := ctx.Err() == context.DeadlineExceeded
	exitCode := 0
	if runErr != nil {
		exitCode = -1
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	result := executionResult{
		Command:  append([]string{command}, arguments...),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		TimedOut: timedOut,
	}
	if (runErr != nil || timedOut) && options.Check {
		return result, processError(result, options.Timeout, runErr)
	}
	return result, nil
}

func (result executionResult) Map() map[string]any {
	return map[string]any{
		"command":   append([]string(nil), result.Command...),
		"stdout":    result.Stdout,
		"stderr":    result.Stderr,
		"exit_code": result.ExitCode,
		"timed_out": result.TimedOut,
		"ok":        result.ExitCode == 0 && !result.TimedOut,
	}
}

func processError(result executionResult, timeout time.Duration, cause error) error {
	if result.TimedOut {
		if timeout > 0 {
			return fmt.Errorf("core.process: command timed out after %s: %s", timeout, strings.Join(result.Command, " "))
		}
		return fmt.Errorf("core.process: command timed out: %s", strings.Join(result.Command, " "))
	}
	if stderr := strings.TrimSpace(result.Stderr); stderr != "" {
		first := strings.SplitN(stderr, "\n", 2)[0]
		return fmt.Errorf("core.process: command exited with status %d: %s", result.ExitCode, first)
	}
	if cause != nil {
		return fmt.Errorf("core.process: command exited with status %d: %w", result.ExitCode, cause)
	}
	return fmt.Errorf("core.process: command exited with status %d", result.ExitCode)
}

func applyKeywordArguments(options *executionOptions, functionName string, keywordArguments runtime.KeywordArguments, allowed ...string) error {
	if len(keywordArguments) == 0 {
		return nil
	}

	allowedSet := map[string]struct{}{}
	for _, name := range allowed {
		allowedSet[name] = struct{}{}
	}
	for name := range keywordArguments {
		if _, ok := allowedSet[name]; !ok {
			return fmt.Errorf("%s got unexpected keyword argument %q", functionName, name)
		}
	}

	if value, ok := keywordArguments["directory"]; ok {
		directory, valid := value.(string)
		if !valid {
			return fmt.Errorf("%s expected directory to be string, got %T", functionName, value)
		}
		options.Directory = directory
	}
	if value, ok := keywordArguments["env"]; ok {
		env, err := envFromValue(value, functionName)
		if err != nil {
			return err
		}
		options.Env = env
	}
	if value, ok := keywordArguments["timeout"]; ok {
		timeout, err := timeoutFromValue(value, functionName)
		if err != nil {
			return err
		}
		options.Timeout = timeout
	}
	if value, ok := keywordArguments["check"]; ok {
		check, valid := value.(bool)
		if !valid {
			return fmt.Errorf("%s expected check to be bool, got %T", functionName, value)
		}
		options.Check = check
	}
	return nil
}

func commandArgs(arguments []any, commandIndex int, functionName string) (string, []string, error) {
	command, err := typemap.ExpectString(arguments, commandIndex, functionName)
	if err != nil {
		return "", nil, err
	}

	processArguments := make([]string, 0, len(arguments)-commandIndex-1)
	for index := commandIndex + 1; index < len(arguments); index++ {
		argument, err := typemap.ExpectString(arguments, index, functionName)
		if err != nil {
			return "", nil, err
		}
		processArguments = append(processArguments, argument)
	}

	return command, processArguments, nil
}

func timeoutFromValue(value any, functionName string) (time.Duration, error) {
	switch typed := value.(type) {
	case nil:
		return 0, nil
	case time.Duration:
		if typed < 0 {
			return 0, fmt.Errorf("%s expected non-negative timeout, got %s", functionName, typed)
		}
		return typed, nil
	case int:
		if typed < 0 {
			return 0, fmt.Errorf("%s expected non-negative timeout, got %d", functionName, typed)
		}
		return time.Duration(typed) * time.Second, nil
	case float64:
		if typed < 0 {
			return 0, fmt.Errorf("%s expected non-negative timeout, got %v", functionName, typed)
		}
		return time.Duration(typed * float64(time.Second)), nil
	case string:
		timeout, err := time.ParseDuration(typed)
		if err != nil {
			return 0, fmt.Errorf("%s expected timeout duration string: %w", functionName, err)
		}
		if timeout < 0 {
			return 0, fmt.Errorf("%s expected non-negative timeout, got %s", functionName, timeout)
		}
		return timeout, nil
	default:
		return 0, fmt.Errorf("%s expected timeout to be seconds or duration string, got %T", functionName, value)
	}
}

func envList(arguments []any, index int, functionName string) ([]string, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	return envFromValue(arguments[index], functionName)
}

func envFromValue(value any, functionName string) ([]string, error) {
	switch typed := value.(type) {
	case nil:
		return nil, nil
	case []string:
		return append([]string(nil), typed...), nil
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("%s expected environment entries to be strings, got %T", functionName, item)
			}
			result = append(result, text)
		}
		return result, nil
	case map[string]string:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		result := make([]string, 0, len(keys))
		for _, key := range keys {
			result = append(result, key+"="+typed[key])
		}
		return result, nil
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		result := make([]string, 0, len(keys))
		for _, key := range keys {
			text, ok := typed[key].(string)
			if !ok {
				return nil, fmt.Errorf("%s expected environment value for %q to be string, got %T", functionName, key, typed[key])
			}
			result = append(result, key+"="+text)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("%s expected environment mapping or []string, got %T", functionName, value)
	}
}

func optionStrings(result core.Result) []string {
	if !result.OK {
		return nil
	}

	switch typed := result.Value.(type) {
	case nil:
		return nil
	case []string:
		return append([]string(nil), typed...)
	case []any:
		resultStrings := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil
			}
			resultStrings = append(resultStrings, text)
		}
		return resultStrings
	default:
		return nil
	}
}
