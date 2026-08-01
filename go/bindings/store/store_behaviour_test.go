package store

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newStoreInterpreter registers the store module against a fresh bootstrap
// interpreter and returns the direct caller.
//
//	caller := newStoreInterpreter(t)
//	value, err := caller.Call("core.store", "available")
func newStoreInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register store module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestStore_Available_Good reports that the native binding is pending.
func TestStore_Available_Good(t *core.T) {
	caller := newStoreInterpreter(t)

	value, callErr := caller.Call("core.store", "available")
	if callErr != nil {
		t.Fatalf("available: %v", callErr)
	}
	if value != false {
		t.Fatalf("expected store support unavailable, got %#v", value)
	}
}

// TestStore_UnknownFunction_Bad reports an unregistered function name.
func TestStore_UnknownFunction_Bad(t *core.T) {
	caller := newStoreInterpreter(t)

	if _, callErr := caller.Call("core.store", "open"); callErr == nil {
		t.Fatal("expected error for unknown function")
	}
}
