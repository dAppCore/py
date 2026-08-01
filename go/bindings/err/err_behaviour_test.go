package err

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newErrInterpreter registers the err module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newErrInterpreter(t)
//	value, err := caller.Call("core.err", "e", "open", "boom")
func newErrInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register err module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestErr_E_Good builds a structured error from operation + message and reads
// its fields back through the bindings.
func TestErr_E_Good(t *core.T) {
	caller := newErrInterpreter(t)

	value, callErr := caller.Call("core.err", "e", "store.open", "cannot open store")
	if callErr != nil {
		t.Fatalf("e: %v", callErr)
	}
	builtError, ok := value.(error)
	if !ok {
		t.Fatalf("expected error, got %T", value)
	}

	message, callErr := caller.Call("core.err", "message", builtError)
	if callErr != nil {
		t.Fatalf("message: %v", callErr)
	}
	if message != "cannot open store" {
		t.Fatalf("unexpected message %#v", message)
	}

	operation, callErr := caller.Call("core.err", "operation", builtError)
	if callErr != nil {
		t.Fatalf("operation: %v", callErr)
	}
	if operation != "store.open" {
		t.Fatalf("unexpected operation %#v", operation)
	}
}

// TestErr_E_WithCode_Good builds a coded error and reads the code back.
func TestErr_E_WithCode_Good(t *core.T) {
	caller := newErrInterpreter(t)

	value, callErr := caller.Call("core.err", "e", "store.open", "cannot open store", nil, "E_OPEN")
	if callErr != nil {
		t.Fatalf("e with code: %v", callErr)
	}
	builtError, ok := value.(error)
	if !ok {
		t.Fatalf("expected error, got %T", value)
	}

	code, callErr := caller.Call("core.err", "error_code", builtError)
	if callErr != nil {
		t.Fatalf("error_code: %v", callErr)
	}
	if code != "E_OPEN" {
		t.Fatalf("unexpected code %#v", code)
	}
}

// TestErr_Wrap_Good wraps a cause and verifies the root unwraps to the cause.
func TestErr_Wrap_Good(t *core.T) {
	caller := newErrInterpreter(t)

	cause := core.E("disk.read", "device offline", nil)

	value, callErr := caller.Call("core.err", "wrap", cause, "store.open", "cannot open store")
	if callErr != nil {
		t.Fatalf("wrap: %v", callErr)
	}
	wrapped, ok := value.(error)
	if !ok {
		t.Fatalf("expected error, got %T", value)
	}

	root, callErr := caller.Call("core.err", "root", wrapped)
	if callErr != nil {
		t.Fatalf("root: %v", callErr)
	}
	rootError, ok := root.(error)
	if !ok {
		t.Fatalf("expected root error, got %T", root)
	}
	if core.Operation(rootError) != "disk.read" {
		t.Fatalf("unexpected root operation %q", core.Operation(rootError))
	}
}

// TestErr_E_Bad reports a missing operation argument as a binding error rather
// than panicking.
func TestErr_E_Bad(t *core.T) {
	caller := newErrInterpreter(t)

	if _, callErr := caller.Call("core.err", "e"); callErr == nil {
		t.Fatal("expected error for missing operation argument")
	}

	if _, callErr := caller.Call("core.err", "e", "store.open"); callErr == nil {
		t.Fatal("expected error for missing message argument")
	}
}

// TestErr_E_Ugly rejects a non-string operation argument.
func TestErr_E_Ugly(t *core.T) {
	caller := newErrInterpreter(t)

	if _, callErr := caller.Call("core.err", "e", 42, "message"); callErr == nil {
		t.Fatal("expected error for non-string operation argument")
	}
}

// TestErr_Message_Nil returns an empty message for a nil error argument.
func TestErr_Message_Nil(t *core.T) {
	caller := newErrInterpreter(t)

	message, callErr := caller.Call("core.err", "message", nil)
	if callErr != nil {
		t.Fatalf("message nil: %v", callErr)
	}
	if message != "" {
		t.Fatalf("expected empty message, got %#v", message)
	}
}

// TestErr_UnknownFunction_Bad reports an unregistered function name.
func TestErr_UnknownFunction_Bad(t *core.T) {
	caller := newErrInterpreter(t)

	if _, callErr := caller.Call("core.err", "does_not_exist"); callErr == nil {
		t.Fatal("expected error for unknown function")
	}
}
