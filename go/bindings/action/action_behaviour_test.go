package action

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newActionInterpreter registers the action module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newActionInterpreter(t)
//	registry, err := caller.Call("core.action", "new_registry")
func newActionInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register action module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// echoHandler is a runtime.Function that returns its options map.
func echoHandler(arguments ...any) (any, error) {
	if len(arguments) == 0 {
		return map[string]any{}, nil
	}
	return arguments[0], nil
}

// TestAction_RegisterRun_Good registers a handler and runs it.
func TestAction_RegisterRun_Good(t *core.T) {
	caller := newActionInterpreter(t)

	registry, callErr := caller.Call("core.action", "new_registry")
	if callErr != nil {
		t.Fatalf("new_registry: %v", callErr)
	}

	handle, callErr := caller.Call("core.action", "register", registry, "echo", runtime.Function(echoHandler))
	if callErr != nil {
		t.Fatalf("register: %v", callErr)
	}

	exists, callErr := caller.Call("core.action", "exists", handle)
	if callErr != nil {
		t.Fatalf("exists: %v", callErr)
	}
	if exists != true {
		t.Fatalf("expected handler present, got %#v", exists)
	}

	value, callErr := caller.Call("core.action", "run", handle, map[string]any{"k": "v"})
	if callErr != nil {
		t.Fatalf("run: %v", callErr)
	}
	options, ok := value.(map[string]any)
	if !ok || options["k"] != "v" {
		t.Fatalf("run: unexpected %#v", value)
	}
}

// TestAction_Names_Good lists registered action names.
func TestAction_Names_Good(t *core.T) {
	caller := newActionInterpreter(t)

	registry, callErr := caller.Call("core.action", "new_registry")
	if callErr != nil {
		t.Fatalf("new_registry: %v", callErr)
	}
	if _, callErr := caller.Call("core.action", "register", registry, "echo", runtime.Function(echoHandler)); callErr != nil {
		t.Fatalf("register: %v", callErr)
	}

	value, callErr := caller.Call("core.action", "names", registry)
	if callErr != nil {
		t.Fatalf("names: %v", callErr)
	}
	names, ok := value.([]string)
	if !ok || len(names) != 1 || names[0] != "echo" {
		t.Fatalf("names: unexpected %#v", value)
	}
}

// TestAction_DisableRun_Good prevents running a disabled action.
func TestAction_DisableRun_Good(t *core.T) {
	caller := newActionInterpreter(t)

	handle, callErr := caller.Call("core.action", "new", "echo", runtime.Function(echoHandler))
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.action", "disable", handle); callErr != nil {
		t.Fatalf("disable: %v", callErr)
	}
	if _, callErr := caller.Call("core.action", "run", handle, map[string]any{}); callErr == nil {
		t.Fatal("expected error running a disabled action")
	}

	if _, callErr := caller.Call("core.action", "enable", handle); callErr != nil {
		t.Fatalf("enable: %v", callErr)
	}
	if _, callErr := caller.Call("core.action", "run", handle, map[string]any{}); callErr != nil {
		t.Fatalf("run after enable: %v", callErr)
	}
}

// TestAction_Get_Good returns a placeholder handle for an unknown name.
func TestAction_Get_Good(t *core.T) {
	caller := newActionInterpreter(t)

	registry, callErr := caller.Call("core.action", "new_registry")
	if callErr != nil {
		t.Fatalf("new_registry: %v", callErr)
	}

	handle, callErr := caller.Call("core.action", "get", registry, "absent")
	if callErr != nil {
		t.Fatalf("get: %v", callErr)
	}
	if handle == nil {
		t.Fatal("get: expected a placeholder handle, got nil")
	}
}

// TestAction_Register_Missing_Bad reports a missing handler argument.
func TestAction_Register_Missing_Bad(t *core.T) {
	caller := newActionInterpreter(t)

	registry, callErr := caller.Call("core.action", "new_registry")
	if callErr != nil {
		t.Fatalf("new_registry: %v", callErr)
	}

	if _, callErr := caller.Call("core.action", "register", registry, "echo"); callErr == nil {
		t.Fatal("expected error for missing handler argument")
	}
}

// TestAction_Run_Ugly reports running an action with no handler.
func TestAction_Run_Ugly(t *core.T) {
	caller := newActionInterpreter(t)

	handle, callErr := caller.Call("core.action", "new", "empty")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.action", "run", handle, map[string]any{}); callErr == nil {
		t.Fatal("expected error running an action without a handler")
	}
}
