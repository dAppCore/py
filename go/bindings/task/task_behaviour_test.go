package task

import (
	core "dappco.re/go"
	actionbinding "dappco.re/go/py/bindings/action"
	"dappco.re/go/py/runtime"
)

// newTaskInterpreter registers the task and action modules against a fresh
// bootstrap interpreter and returns the direct caller.
//
//	caller := newTaskInterpreter(t)
//	registry, err := caller.Call("core.task", "new_registry")
func newTaskInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register task module: %v", err)
	}
	if err := actionbinding.Register(interpreter); err != nil {
		t.Fatalf("register action module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// echoHandler returns its options map.
func echoHandler(arguments ...any) (any, error) {
	if len(arguments) == 0 {
		return map[string]any{}, nil
	}
	return arguments[0], nil
}

// TestTask_NewStep_Good builds a step definition map.
func TestTask_NewStep_Good(t *core.T) {
	caller := newTaskInterpreter(t)

	value, callErr := caller.Call("core.task", "new_step", "echo", map[string]any{"k": "v"})
	if callErr != nil {
		t.Fatalf("new_step: %v", callErr)
	}
	step, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("new_step: expected map, got %T", value)
	}
	if step["action"] != "echo" {
		t.Fatalf("unexpected step action %#v", step["action"])
	}
}

// TestTask_RegisterNamesExists_Good registers a task and lists names.
func TestTask_RegisterNamesExists_Good(t *core.T) {
	caller := newTaskInterpreter(t)

	registry, callErr := caller.Call("core.task", "new_registry")
	if callErr != nil {
		t.Fatalf("new_registry: %v", callErr)
	}

	steps := []any{map[string]any{"action": "echo", "with": map[string]any{"k": "v"}}}
	handle, callErr := caller.Call("core.task", "register", registry, "greet", steps)
	if callErr != nil {
		t.Fatalf("register: %v", callErr)
	}

	exists, callErr := caller.Call("core.task", "exists", handle)
	if callErr != nil {
		t.Fatalf("exists: %v", callErr)
	}
	if exists != true {
		t.Fatalf("expected task to have steps, got %#v", exists)
	}

	names, callErr := caller.Call("core.task", "names", registry)
	if callErr != nil {
		t.Fatalf("names: %v", callErr)
	}
	if list, ok := names.([]string); !ok || len(list) != 1 || list[0] != "greet" {
		t.Fatalf("names: unexpected %#v", names)
	}
}

// TestTask_Run_Good runs a task whose step targets a registered action.
func TestTask_Run_Good(t *core.T) {
	caller := newTaskInterpreter(t)

	actionRegistry, callErr := caller.Call("core.action", "new_registry")
	if callErr != nil {
		t.Fatalf("action new_registry: %v", callErr)
	}
	if _, callErr := caller.Call("core.action", "register", actionRegistry, "echo", runtime.Function(echoHandler)); callErr != nil {
		t.Fatalf("action register: %v", callErr)
	}

	task, callErr := caller.Call("core.task", "new", "greet",
		[]any{map[string]any{"action": "echo", "with": map[string]any{"name": "corepy"}}})
	if callErr != nil {
		t.Fatalf("task new: %v", callErr)
	}

	value, callErr := caller.Call("core.task", "run", task, actionRegistry)
	if callErr != nil {
		t.Fatalf("run: %v", callErr)
	}
	result, ok := value.(map[string]any)
	if !ok || result["name"] != "corepy" {
		t.Fatalf("run: unexpected %#v", value)
	}
}

// TestTask_Get_Good returns a placeholder handle for an unknown task.
func TestTask_Get_Good(t *core.T) {
	caller := newTaskInterpreter(t)

	registry, callErr := caller.Call("core.task", "new_registry")
	if callErr != nil {
		t.Fatalf("new_registry: %v", callErr)
	}

	handle, callErr := caller.Call("core.task", "get", registry, "absent")
	if callErr != nil {
		t.Fatalf("get: %v", callErr)
	}
	if handle == nil {
		t.Fatal("get: expected a placeholder handle")
	}
}

// TestTask_NewStep_Bad reports a missing action argument.
func TestTask_NewStep_Bad(t *core.T) {
	caller := newTaskInterpreter(t)

	if _, callErr := caller.Call("core.task", "new_step"); callErr == nil {
		t.Fatal("expected error for missing action name")
	}
}

// TestTask_Run_Ugly reports running a task with no steps.
func TestTask_Run_Ugly(t *core.T) {
	caller := newTaskInterpreter(t)

	actionRegistry, callErr := caller.Call("core.action", "new_registry")
	if callErr != nil {
		t.Fatalf("action new_registry: %v", callErr)
	}

	task, callErr := caller.Call("core.task", "new", "empty")
	if callErr != nil {
		t.Fatalf("task new: %v", callErr)
	}

	if _, callErr := caller.Call("core.task", "run", task, actionRegistry); callErr == nil {
		t.Fatal("expected error running a task with no steps")
	}
}
