package container

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newContainerInterpreter registers the container module against a fresh
// bootstrap interpreter and returns the direct caller.
//
//	caller := newContainerInterpreter(t)
//	value, err := caller.Call("core.container", "available")
func newContainerInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register container module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestContainer_Available_Good reports that the native binding is pending.
func TestContainer_Available_Good(t *core.T) {
	caller := newContainerInterpreter(t)

	value, callErr := caller.Call("core.container", "available")
	if callErr != nil {
		t.Fatalf("available: %v", callErr)
	}
	if value != false {
		t.Fatalf("expected container support unavailable, got %#v", value)
	}
}

// TestContainer_UnknownFunction_Bad reports an unregistered function name.
func TestContainer_UnknownFunction_Bad(t *core.T) {
	caller := newContainerInterpreter(t)

	if _, callErr := caller.Call("core.container", "run"); callErr == nil {
		t.Fatal("expected error for unknown function")
	}
}
