package options

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newOptionsInterpreter registers the options module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newOptionsInterpreter(t)
//	handle, err := caller.Call("core.options", "new")
func newOptionsInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register options module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestOptions_SetGet_Good sets and reads typed values back through a handle.
func TestOptions_SetGet_Good(t *core.T) {
	caller := newOptionsInterpreter(t)

	handle, callErr := caller.Call("core.options", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.options", "set", handle, "name", "corepy"); callErr != nil {
		t.Fatalf("set name: %v", callErr)
	}
	if _, callErr := caller.Call("core.options", "set", handle, "port", 8080); callErr != nil {
		t.Fatalf("set port: %v", callErr)
	}
	if _, callErr := caller.Call("core.options", "set", handle, "debug", true); callErr != nil {
		t.Fatalf("set debug: %v", callErr)
	}

	name, callErr := caller.Call("core.options", "string", handle, "name")
	if callErr != nil {
		t.Fatalf("string: %v", callErr)
	}
	if name != "corepy" {
		t.Fatalf("unexpected name %#v", name)
	}

	port, callErr := caller.Call("core.options", "int", handle, "port")
	if callErr != nil {
		t.Fatalf("int: %v", callErr)
	}
	if port != 8080 {
		t.Fatalf("unexpected port %#v", port)
	}

	debug, callErr := caller.Call("core.options", "bool", handle, "debug")
	if callErr != nil {
		t.Fatalf("bool: %v", callErr)
	}
	if debug != true {
		t.Fatalf("unexpected debug %#v", debug)
	}
}

// TestOptions_HasItems_Good reports presence and exposes the items map.
func TestOptions_HasItems_Good(t *core.T) {
	caller := newOptionsInterpreter(t)

	handle, callErr := caller.Call("core.options", "new", map[string]any{"name": "corepy"})
	if callErr != nil {
		t.Fatalf("new from map: %v", callErr)
	}

	has, callErr := caller.Call("core.options", "has", handle, "name")
	if callErr != nil {
		t.Fatalf("has: %v", callErr)
	}
	if has != true {
		t.Fatalf("expected key present, got %#v", has)
	}

	missing, callErr := caller.Call("core.options", "has", handle, "absent")
	if callErr != nil {
		t.Fatalf("has absent: %v", callErr)
	}
	if missing != false {
		t.Fatalf("expected key absent, got %#v", missing)
	}

	items, callErr := caller.Call("core.options", "items", handle)
	if callErr != nil {
		t.Fatalf("items: %v", callErr)
	}
	values, ok := items.(map[string]any)
	if !ok {
		t.Fatalf("items: expected map, got %T", items)
	}
	if values["name"] != "corepy" {
		t.Fatalf("unexpected items entry %#v", values["name"])
	}
}

// TestOptions_GetMissing_Good returns nil for an absent key.
func TestOptions_GetMissing_Good(t *core.T) {
	caller := newOptionsInterpreter(t)

	handle, callErr := caller.Call("core.options", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	value, callErr := caller.Call("core.options", "get", handle, "absent")
	if callErr != nil {
		t.Fatalf("get absent: %v", callErr)
	}
	if value != nil {
		t.Fatalf("expected nil for absent key, got %#v", value)
	}
}

// TestOptions_Set_Bad reports a missing value argument.
func TestOptions_Set_Bad(t *core.T) {
	caller := newOptionsInterpreter(t)

	handle, callErr := caller.Call("core.options", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.options", "set", handle, "name"); callErr == nil {
		t.Fatal("expected error for missing value argument")
	}
}

// TestOptions_WrongHandle_Ugly rejects a non-Options handle argument.
func TestOptions_WrongHandle_Ugly(t *core.T) {
	caller := newOptionsInterpreter(t)

	if _, callErr := caller.Call("core.options", "get", "not-a-handle", "name"); callErr == nil {
		t.Fatal("expected error for non-Options handle")
	}
}
