package echo

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newEchoInterpreter registers the echo module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the binding.
//
//	caller := newEchoInterpreter(t)
//	value, err := caller.Call("core", "echo", "hi")
func newEchoInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register echo module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestEcho_Echo_Good returns its single argument unchanged.
func TestEcho_Echo_Good(t *core.T) {
	caller := newEchoInterpreter(t)

	value, callErr := caller.Call("core", "echo", "hello")
	if callErr != nil {
		t.Fatalf("echo: %v", callErr)
	}
	if value != "hello" {
		t.Fatalf("unexpected echo %#v", value)
	}
}

// TestEcho_Echo_Bad reports a missing argument.
func TestEcho_Echo_Bad(t *core.T) {
	caller := newEchoInterpreter(t)

	if _, callErr := caller.Call("core", "echo"); callErr == nil {
		t.Fatal("expected error for missing argument")
	}
}

// TestEcho_Echo_Ugly reports more than one argument.
func TestEcho_Echo_Ugly(t *core.T) {
	caller := newEchoInterpreter(t)

	if _, callErr := caller.Call("core", "echo", "a", "b"); callErr == nil {
		t.Fatal("expected error for extra argument")
	}
}
