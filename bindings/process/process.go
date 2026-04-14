package process

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

var (
	defaultCoreOnce sync.Once
	defaultCore     *core.Core
)

// Register exposes Process bindings backed by core.Process.
//
//	process.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.process",
		Documentation: "Process helpers backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"run":          run,
			"run_in":       runIn,
			"run_with_env": runWithEnv,
			"exists":       exists,
		},
	})
}

func run(arguments ...any) (any, error) {
	command, processArguments, err := commandArgs(arguments, 0, "core.process.run")
	if err != nil {
		return nil, err
	}

	return typemap.ResultValue(processCore().Process().Run(context.Background(), command, processArguments...), "core.process.run")
}

func runIn(arguments ...any) (any, error) {
	directory, err := typemap.ExpectString(arguments, 0, "core.process.run_in")
	if err != nil {
		return nil, err
	}
	command, processArguments, err := commandArgs(arguments, 1, "core.process.run_in")
	if err != nil {
		return nil, err
	}

	return typemap.ResultValue(processCore().Process().RunIn(context.Background(), directory, command, processArguments...), "core.process.run_in")
}

func runWithEnv(arguments ...any) (any, error) {
	directory, err := typemap.ExpectString(arguments, 0, "core.process.run_with_env")
	if err != nil {
		return nil, err
	}
	env, err := envList(arguments, 1, "core.process.run_with_env")
	if err != nil {
		return nil, err
	}
	command, processArguments, err := commandArgs(arguments, 2, "core.process.run_with_env")
	if err != nil {
		return nil, err
	}

	return typemap.ResultValue(processCore().Process().RunWithEnv(context.Background(), directory, env, command, processArguments...), "core.process.run_with_env")
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

	cmd := exec.CommandContext(ctx, command, optionStrings(options.Get("args"))...)
	if directory := options.String("dir"); directory != "" {
		cmd.Dir = directory
	}
	if env := optionStrings(options.Get("env")); len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		cause := err
		if stderrString := strings.TrimSpace(stderr.String()); stderrString != "" {
			cause = fmt.Errorf("%w: %s", err, stderrString)
		}
		return core.Result{
			Value: core.E("core.process.run", core.Concat("command failed: ", command), cause),
			OK:    false,
		}
	}

	return core.Result{Value: stdout.String(), OK: true}
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
