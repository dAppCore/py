package array

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newArrayInterpreter registers the array module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newArrayInterpreter(t)
//	handle, err := caller.Call("core.array", "new", 1, 2)
func newArrayInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register array module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestArray_AddContainsLen_Good adds items and reports membership and length.
func TestArray_AddContainsLen_Good(t *core.T) {
	caller := newArrayInterpreter(t)

	handle, callErr := caller.Call("core.array", "new", "a")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.array", "add", handle, "b", "c"); callErr != nil {
		t.Fatalf("add: %v", callErr)
	}

	length, callErr := caller.Call("core.array", "len", handle)
	if callErr != nil {
		t.Fatalf("len: %v", callErr)
	}
	if length != 3 {
		t.Fatalf("unexpected length %#v", length)
	}

	contains, callErr := caller.Call("core.array", "contains", handle, "b")
	if callErr != nil {
		t.Fatalf("contains: %v", callErr)
	}
	if contains != true {
		t.Fatalf("expected membership, got %#v", contains)
	}
}

// TestArray_AddUniqueDeduplicate_Good keeps the array distinct.
func TestArray_AddUniqueDeduplicate_Good(t *core.T) {
	caller := newArrayInterpreter(t)

	handle, callErr := caller.Call("core.array", "new", "a", "a", "b")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.array", "add_unique", handle, "a", "c"); callErr != nil {
		t.Fatalf("add_unique: %v", callErr)
	}
	if _, callErr := caller.Call("core.array", "deduplicate", handle); callErr != nil {
		t.Fatalf("deduplicate: %v", callErr)
	}

	value, callErr := caller.Call("core.array", "as_list", handle)
	if callErr != nil {
		t.Fatalf("as_list: %v", callErr)
	}
	list, ok := value.([]any)
	if !ok || len(list) != 3 {
		t.Fatalf("expected three distinct items, got %#v", value)
	}
}

// TestArray_RemoveClear_Good removes a value and clears the array.
func TestArray_RemoveClear_Good(t *core.T) {
	caller := newArrayInterpreter(t)

	handle, callErr := caller.Call("core.array", "new", "a", "b", "c")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.array", "remove", handle, "b"); callErr != nil {
		t.Fatalf("remove: %v", callErr)
	}
	length, callErr := caller.Call("core.array", "len", handle)
	if callErr != nil {
		t.Fatalf("len after remove: %v", callErr)
	}
	if length != 2 {
		t.Fatalf("unexpected length after remove %#v", length)
	}

	if _, callErr := caller.Call("core.array", "clear", handle); callErr != nil {
		t.Fatalf("clear: %v", callErr)
	}
	length, callErr = caller.Call("core.array", "len", handle)
	if callErr != nil {
		t.Fatalf("len after clear: %v", callErr)
	}
	if length != 0 {
		t.Fatalf("expected empty array, got %#v", length)
	}
}

// TestArray_Contains_Bad reports a missing value argument.
func TestArray_Contains_Bad(t *core.T) {
	caller := newArrayInterpreter(t)

	handle, callErr := caller.Call("core.array", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.array", "contains", handle); callErr == nil {
		t.Fatal("expected error for missing value argument")
	}
}

// TestArray_WrongHandle_Ugly rejects a non-array handle argument.
func TestArray_WrongHandle_Ugly(t *core.T) {
	caller := newArrayInterpreter(t)

	if _, callErr := caller.Call("core.array", "len", "not-a-handle"); callErr == nil {
		t.Fatal("expected error for non-array handle")
	}
}
