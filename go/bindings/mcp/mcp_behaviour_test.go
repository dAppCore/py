package mcp

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newMCPInterpreter registers the mcp module against a fresh bootstrap
// interpreter and returns the direct caller.
//
//	caller := newMCPInterpreter(t)
//	value, err := caller.Call("core.mcp", "available")
func newMCPInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register mcp module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestMCP_Available_Good reports that the native binding is pending.
func TestMCP_Available_Good(t *core.T) {
	caller := newMCPInterpreter(t)

	value, callErr := caller.Call("core.mcp", "available")
	if callErr != nil {
		t.Fatalf("available: %v", callErr)
	}
	if value != false {
		t.Fatalf("expected mcp support unavailable, got %#v", value)
	}
}

// TestMCP_UnknownFunction_Bad reports an unregistered function name.
func TestMCP_UnknownFunction_Bad(t *core.T) {
	caller := newMCPInterpreter(t)

	if _, callErr := caller.Call("core.mcp", "connect"); callErr == nil {
		t.Fatal("expected error for unknown function")
	}
}
