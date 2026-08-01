package log

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newLogInterpreter registers the log module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newLogInterpreter(t)
//	value, err := caller.Call("core.log", "info", "ready")
func newLogInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register log module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestLog_Levels_Good emits a message at each level after setting the level.
func TestLog_Levels_Good(t *core.T) {
	caller := newLogInterpreter(t)

	for _, level := range []string{"quiet", "error", "warn", "info", "debug"} {
		value, callErr := caller.Call("core.log", "set_level", level)
		if callErr != nil {
			t.Fatalf("set_level %s: %v", level, callErr)
		}
		if value != true {
			t.Fatalf("set_level %s: unexpected %#v", level, value)
		}
	}

	for _, function := range []string{"debug", "info", "warn", "error"} {
		value, callErr := caller.Call("core.log", function, "message from %s", "test")
		if callErr != nil {
			t.Fatalf("%s: %v", function, callErr)
		}
		if value != true {
			t.Fatalf("%s: unexpected %#v", function, value)
		}
	}
}

// TestLog_SetLevel_Bad reports an unknown level name.
func TestLog_SetLevel_Bad(t *core.T) {
	caller := newLogInterpreter(t)

	if _, callErr := caller.Call("core.log", "set_level", "verbose"); callErr == nil {
		t.Fatal("expected error for unknown level")
	}
}

// TestLog_Info_Ugly rejects a non-string message argument.
func TestLog_Info_Ugly(t *core.T) {
	caller := newLogInterpreter(t)

	if _, callErr := caller.Call("core.log", "info", 42); callErr == nil {
		t.Fatal("expected error for non-string message")
	}
}
