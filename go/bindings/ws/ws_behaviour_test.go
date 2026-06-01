package ws

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newWSInterpreter registers the ws module against a fresh bootstrap
// interpreter and returns the direct caller.
//
//	caller := newWSInterpreter(t)
//	value, err := caller.Call("core.ws", "available")
func newWSInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register ws module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestWS_Available_Good reports that the native binding is pending.
func TestWS_Available_Good(t *core.T) {
	caller := newWSInterpreter(t)

	value, callErr := caller.Call("core.ws", "available")
	if callErr != nil {
		t.Fatalf("available: %v", callErr)
	}
	if value != false {
		t.Fatalf("expected ws support unavailable, got %#v", value)
	}
}

// TestWS_UnknownFunction_Bad reports an unregistered function name.
func TestWS_UnknownFunction_Bad(t *core.T) {
	caller := newWSInterpreter(t)

	if _, callErr := caller.Call("core.ws", "connect"); callErr == nil {
		t.Fatal("expected error for unknown function")
	}
}
