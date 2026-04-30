package scm

import (
	"fmt"     // AX-6-exception: SCM bootstrap preserves wrapped git command errors.
	"os/exec" // AX-6-exception: SCM binding shells to git until go-scm is wired as the backing primitive.
	"strings" // AX-6-exception: SCM parses git porcelain output with stdlib line helpers.

	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes Git-backed source-control helpers.
//
//	scm.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.scm",
		Documentation: "Git helpers for CorePy",
		Functions: map[string]runtime.Function{
			"exists":        exists,
			"root":          root,
			"branch":        branch,
			"head":          head,
			"status":        status,
			"tracked_files": trackedFiles,
		},
	})
}

func exists(arguments ...any) (any, error) {
	directory, err := directoryArgument(arguments, "core.scm.exists")
	if err != nil {
		return nil, err
	}
	if _, err := exec.LookPath("git"); err != nil {
		return false, nil
	}
	output, err := git(directory, false, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(output) == "true", nil
}

func root(arguments ...any) (any, error) {
	directory, err := directoryArgument(arguments, "core.scm.root")
	if err != nil {
		return nil, err
	}
	output, err := git(directory, true, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	return strings.TrimSpace(output), nil
}

func branch(arguments ...any) (any, error) {
	directory, err := directoryArgument(arguments, "core.scm.branch")
	if err != nil {
		return nil, err
	}
	output, err := git(directory, true, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, err
	}
	return strings.TrimSpace(output), nil
}

func head(arguments ...any) (any, error) {
	directory, err := directoryArgument(arguments, "core.scm.head")
	if err != nil {
		return nil, err
	}
	output, err := git(directory, true, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	return strings.TrimSpace(output), nil
}

func status(arguments ...any) (any, error) {
	directory, err := directoryArgument(arguments, "core.scm.status")
	if err != nil {
		return nil, err
	}
	output, err := git(directory, true, "status", "--short", "--branch")
	if err != nil {
		return nil, err
	}

	lines := []string{}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimRight(line, "\r")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		lines = append(lines, trimmed)
	}

	branchName := ""
	if len(lines) > 0 && strings.HasPrefix(lines[0], "## ") {
		branchName = strings.TrimSpace(strings.TrimPrefix(lines[0], "## "))
		lines = lines[1:]
	}

	return map[string]any{
		"branch":  branchName,
		"clean":   len(lines) == 0,
		"changes": lines,
	}, nil
}

func trackedFiles(arguments ...any) (any, error) {
	directory, err := directoryArgument(arguments, "core.scm.tracked_files")
	if err != nil {
		return nil, err
	}
	output, err := git(directory, true, "ls-files")
	if err != nil {
		return nil, err
	}
	files := []string{}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		files = append(files, trimmed)
	}
	return files, nil
}

func directoryArgument(arguments []any, functionName string) (string, error) {
	if len(arguments) == 0 {
		return ".", nil
	}
	return typemap.ExpectString(arguments, 0, functionName)
}

func git(directory string, check bool, arguments ...string) (string, error) {
	gitBinary, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("git is not available")
	}

	command := exec.Command(gitBinary, append([]string{"-C", directory}, arguments...)...)
	output, err := command.CombinedOutput()
	if err != nil && check {
		return "", fmt.Errorf("git %s failed: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}
