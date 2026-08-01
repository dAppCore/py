package api

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newAPIInterpreter registers the api module against a fresh bootstrap
// interpreter and returns the direct caller.
//
//	caller := newAPIInterpreter(t)
//	value, err := caller.Call("core.api", "available")
func newAPIInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register api module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestAPI_Available_Good reports that the native binding is pending.
func TestAPI_Available_Good(t *core.T) {
	caller := newAPIInterpreter(t)

	value, callErr := caller.Call("core.api", "available")
	if callErr != nil {
		t.Fatalf("available: %v", callErr)
	}
	if value != false {
		t.Fatalf("expected api support unavailable, got %#v", value)
	}
}

// TestAPI_UnknownFunction_Bad reports an unregistered function name.
func TestAPI_UnknownFunction_Bad(t *core.T) {
	caller := newAPIInterpreter(t)

	if _, callErr := caller.Call("core.api", "serve"); callErr == nil {
		t.Fatal("expected error for unknown function")
	}
}
