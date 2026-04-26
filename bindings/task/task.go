package task

import (
	"fmt" // AX-6-exception: task bootstrap validation reports dynamic Go types and action names.

	core "dappco.re/go/core"
	actionbinding "dappco.re/go/py/bindings/action"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

type Step struct {
	Action string
	With   map[string]any
	Async  bool
	Input  string
}

type Handle struct {
	Name        string
	Description string
	Steps       []Step
}

type Registry = core.Registry[*Handle]

// Register exposes Core task helpers.
//
//	task.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.task",
		Documentation: "Task composition helpers for CorePy",
		Functions: map[string]runtime.Function{
			"new":          newTask,
			"new_registry": newRegistry,
			"new_step":     newStep,
			"register":     registerTask,
			"get":          getTask,
			"names":        names,
			"run":          run,
			"exists":       exists,
		},
	})
}

func newTask(arguments ...any) (any, error) {
	item := &Handle{}
	if len(arguments) > 0 {
		name, err := typemap.ExpectString(arguments, 0, "core.task.new")
		if err != nil {
			return nil, err
		}
		item.Name = name
	}
	if len(arguments) > 1 {
		steps, err := parseSteps(arguments[1], "core.task.new")
		if err != nil {
			return nil, err
		}
		item.Steps = steps
	}
	if len(arguments) > 2 {
		description, err := typemap.ExpectString(arguments, 2, "core.task.new")
		if err != nil {
			return nil, err
		}
		item.Description = description
	}
	return item, nil
}

func newRegistry(arguments ...any) (any, error) {
	return core.NewRegistry[*Handle](), nil
}

func newStep(arguments ...any) (any, error) {
	positional, keywordArguments := runtime.SplitKeywordArguments(arguments)

	actionName, err := typemap.ExpectString(positional, 0, "core.task.new_step")
	if err != nil {
		return nil, err
	}
	step := Step{Action: actionName, With: map[string]any{}}
	if len(positional) > 1 {
		withValues, err := typemap.ExpectMap(positional, 1, "core.task.new_step")
		if err != nil {
			return nil, err
		}
		step.With = withValues
	}
	if len(positional) > 2 {
		async, ok := positional[2].(bool)
		if !ok {
			return nil, fmt.Errorf("core.task.new_step expected argument 2 to be bool, got %T", positional[2])
		}
		step.Async = async
	}
	if len(positional) > 3 {
		input, err := typemap.ExpectString(positional, 3, "core.task.new_step")
		if err != nil {
			return nil, err
		}
		step.Input = input
	}
	if len(keywordArguments) > 0 {
		if withValues, ok := keywordArguments["with"].(map[string]any); ok {
			step.With = cloneMap(withValues)
		}
		if withValues, ok := keywordArguments["with_values"].(map[string]any); ok {
			step.With = cloneMap(withValues)
		}
		if asyncValue, exists := keywordArguments["async"]; exists {
			asyncBool, ok := asyncValue.(bool)
			if !ok {
				return nil, fmt.Errorf("core.task.new_step expected keyword async to be bool, got %T", asyncValue)
			}
			step.Async = asyncBool
		}
		if asyncValue, exists := keywordArguments["async_step"]; exists {
			asyncBool, ok := asyncValue.(bool)
			if !ok {
				return nil, fmt.Errorf("core.task.new_step expected keyword async_step to be bool, got %T", asyncValue)
			}
			step.Async = asyncBool
		}
		if inputValue, exists := keywordArguments["input"]; exists {
			inputString, ok := inputValue.(string)
			if !ok {
				return nil, fmt.Errorf("core.task.new_step expected keyword input to be string, got %T", inputValue)
			}
			step.Input = inputString
		}
	}
	return stepToMap(step), nil
}

func registerTask(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.task.register")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.task.register")
	if err != nil {
		return nil, err
	}
	steps, err := parseSteps(arguments[2], "core.task.register")
	if err != nil {
		return nil, err
	}
	item := &Handle{Name: name, Steps: steps}
	if len(arguments) > 3 {
		description, err := typemap.ExpectString(arguments, 3, "core.task.register")
		if err != nil {
			return nil, err
		}
		item.Description = description
	}
	if _, err := typemap.ResultValue(registryValue.Set(name, item), "core.task.register"); err != nil {
		return nil, err
	}
	return item, nil
}

func getTask(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.task.get")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.task.get")
	if err != nil {
		return nil, err
	}
	result := registryValue.Get(name)
	if !result.OK {
		return &Handle{Name: name}, nil
	}
	return result.Value, nil
}

func names(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.task.names")
	if err != nil {
		return nil, err
	}
	return registryValue.Names(), nil
}

func run(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.task.run")
	if err != nil {
		return nil, err
	}
	actionRegistry, err := expectActionRegistry(arguments, 1, "core.task.run")
	if err != nil {
		return nil, err
	}
	options := map[string]any{}
	if len(arguments) > 2 {
		options, err = typemap.ExpectMap(arguments, 2, "core.task.run")
		if err != nil {
			return nil, err
		}
	}
	return RunHandle(item, actionRegistry, options)
}

func exists(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.task.exists")
	if err != nil {
		return nil, err
	}
	return len(item.Steps) > 0, nil
}

// RunHandle executes a task with an action registry and runtime values.
func RunHandle(item *Handle, actions *actionbinding.Registry, options map[string]any) (any, error) {
	if item == nil || len(item.Steps) == 0 {
		name := "<nil>"
		if item != nil && item.Name != "" {
			name = item.Name
		}
		return nil, fmt.Errorf("task has no steps: %s", name)
	}

	var (
		lastValue any
		lastOK    bool
	)

	for _, step := range item.Steps {
		stepOptions := cloneMap(step.With)
		if len(stepOptions) == 0 {
			stepOptions = cloneMap(options)
		}
		if step.Input == "previous" && lastOK {
			stepOptions["_input"] = lastValue
		}

		actionResult := actions.Get(step.Action)
		if !actionResult.OK {
			return nil, fmt.Errorf("action not found: %s", step.Action)
		}
		actionValue := actionResult.Value.(*actionbinding.Handle)

		if step.Async {
			go func(currentAction *actionbinding.Handle, currentOptions map[string]any) {
				_, _ = actionbinding.RunHandle(currentAction, currentOptions)
			}(actionValue, cloneMap(stepOptions))
			continue
		}

		lastResult, err := actionbinding.RunHandle(actionValue, stepOptions)
		if err != nil {
			return nil, err
		}
		lastValue = lastResult
		lastOK = true
	}
	return lastValue, nil
}

func parseSteps(value any, functionName string) ([]Step, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s expected steps to be []any, got %T", functionName, value)
	}

	steps := make([]Step, 0, len(items))
	for _, item := range items {
		switch typed := item.(type) {
		case map[string]any:
			step, err := mapToStep(typed, functionName)
			if err != nil {
				return nil, err
			}
			steps = append(steps, step)
		default:
			return nil, fmt.Errorf("%s expected step definitions to be map[string]any, got %T", functionName, item)
		}
	}
	return steps, nil
}

func mapToStep(values map[string]any, functionName string) (Step, error) {
	actionValue, ok := values["action"].(string)
	if !ok || actionValue == "" {
		return Step{}, fmt.Errorf("%s expected step action to be a non-empty string", functionName)
	}

	step := Step{Action: actionValue, With: map[string]any{}}
	if withValues, ok := values["with"].(map[string]any); ok {
		step.With = cloneMap(withValues)
	}
	if withValues, ok := values["with_values"].(map[string]any); ok {
		step.With = cloneMap(withValues)
	}
	if asyncValue, exists := values["async"]; exists {
		asyncBool, ok := asyncValue.(bool)
		if !ok {
			return Step{}, fmt.Errorf("%s expected step async to be bool, got %T", functionName, asyncValue)
		}
		step.Async = asyncBool
	}
	if asyncValue, exists := values["async_step"]; exists {
		asyncBool, ok := asyncValue.(bool)
		if !ok {
			return Step{}, fmt.Errorf("%s expected step async_step to be bool, got %T", functionName, asyncValue)
		}
		step.Async = asyncBool
	}
	if inputValue, exists := values["input"]; exists {
		inputString, ok := inputValue.(string)
		if !ok {
			return Step{}, fmt.Errorf("%s expected step input to be string, got %T", functionName, inputValue)
		}
		step.Input = inputString
	}
	return step, nil
}

func stepToMap(step Step) map[string]any {
	return map[string]any{
		"action": actionValue(step),
		"with":   cloneMap(step.With),
		"async":  step.Async,
		"input":  step.Input,
	}
}

func actionValue(step Step) string {
	return step.Action
}

func cloneMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func expectRegistry(arguments []any, index int, functionName string) (*Registry, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*Registry)
	if !ok {
		return nil, fmt.Errorf("%s expected task registry, got %T", functionName, arguments[index])
	}
	return value, nil
}

func expectHandle(arguments []any, index int, functionName string) (*Handle, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*Handle)
	if !ok {
		return nil, fmt.Errorf("%s expected task handle, got %T", functionName, arguments[index])
	}
	return value, nil
}

func expectActionRegistry(arguments []any, index int, functionName string) (*actionbinding.Registry, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*actionbinding.Registry)
	if !ok {
		return nil, fmt.Errorf("%s expected action registry, got %T", functionName, arguments[index])
	}
	return value, nil
}
