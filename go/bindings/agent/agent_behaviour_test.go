package agent

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newAgentInterpreter registers the agent module against a fresh bootstrap
// interpreter and returns the direct caller.
//
//	caller := newAgentInterpreter(t)
//	value, err := caller.Call("core.agent", "available")
func newAgentInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register agent module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestAgent_Available_Good reports that the native binding is pending.
func TestAgent_Available_Good(t *core.T) {
	caller := newAgentInterpreter(t)

	value, callErr := caller.Call("core.agent", "available")
	if callErr != nil {
		t.Fatalf("available: %v", callErr)
	}
	if value != false {
		t.Fatalf("expected agent support unavailable, got %#v", value)
	}
}

// TestAgent_UnknownFunction_Bad reports an unregistered function name.
func TestAgent_UnknownFunction_Bad(t *core.T) {
	caller := newAgentInterpreter(t)

	if _, callErr := caller.Call("core.agent", "spawn"); callErr == nil {
		t.Fatal("expected error for unknown function")
	}
}
