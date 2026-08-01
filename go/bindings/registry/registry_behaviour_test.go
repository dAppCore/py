package registry

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newRegistryInterpreter registers the registry module against a fresh
// bootstrap interpreter and returns the direct caller.
//
//	caller := newRegistryInterpreter(t)
//	handle, err := caller.Call("core.registry", "new")
func newRegistryInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register registry module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestRegistry_SetGetHas_Good stores values and reports membership.
func TestRegistry_SetGetHas_Good(t *core.T) {
	caller := newRegistryInterpreter(t)

	handle, callErr := caller.Call("core.registry", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.registry", "set", handle, "alpha", 1); callErr != nil {
		t.Fatalf("set: %v", callErr)
	}

	value, callErr := caller.Call("core.registry", "get", handle, "alpha")
	if callErr != nil {
		t.Fatalf("get: %v", callErr)
	}
	if value != 1 {
		t.Fatalf("unexpected value %#v", value)
	}

	has, callErr := caller.Call("core.registry", "has", handle, "alpha")
	if callErr != nil {
		t.Fatalf("has: %v", callErr)
	}
	if has != true {
		t.Fatalf("expected key present, got %#v", has)
	}
}

// TestRegistry_GetDefault_Good returns the default for a missing name.
func TestRegistry_GetDefault_Good(t *core.T) {
	caller := newRegistryInterpreter(t)

	handle, callErr := caller.Call("core.registry", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	value, callErr := caller.Call("core.registry", "get", handle, "absent", "fallback")
	if callErr != nil {
		t.Fatalf("get default: %v", callErr)
	}
	if value != "fallback" {
		t.Fatalf("expected fallback, got %#v", value)
	}
}

// TestRegistry_NamesLen_Good reports the registered names and count.
func TestRegistry_NamesLen_Good(t *core.T) {
	caller := newRegistryInterpreter(t)

	handle, callErr := caller.Call("core.registry", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}
	if _, callErr := caller.Call("core.registry", "set", handle, "alpha", 1); callErr != nil {
		t.Fatalf("set alpha: %v", callErr)
	}
	if _, callErr := caller.Call("core.registry", "set", handle, "beta", 2); callErr != nil {
		t.Fatalf("set beta: %v", callErr)
	}

	names, callErr := caller.Call("core.registry", "names", handle)
	if callErr != nil {
		t.Fatalf("names: %v", callErr)
	}
	if list, ok := names.([]string); !ok || len(list) != 2 {
		t.Fatalf("names: unexpected %#v", names)
	}

	length, callErr := caller.Call("core.registry", "len", handle)
	if callErr != nil {
		t.Fatalf("len: %v", callErr)
	}
	if length != 2 {
		t.Fatalf("unexpected length %#v", length)
	}
}

// TestRegistry_DeleteList_Good deletes an entry and lists by pattern.
func TestRegistry_DeleteList_Good(t *core.T) {
	caller := newRegistryInterpreter(t)

	handle, callErr := caller.Call("core.registry", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}
	if _, callErr := caller.Call("core.registry", "set", handle, "alpha", 1); callErr != nil {
		t.Fatalf("set alpha: %v", callErr)
	}
	if _, callErr := caller.Call("core.registry", "set", handle, "beta", 2); callErr != nil {
		t.Fatalf("set beta: %v", callErr)
	}

	listed, callErr := caller.Call("core.registry", "list", handle, "*")
	if callErr != nil {
		t.Fatalf("list: %v", callErr)
	}
	if listed == nil {
		t.Fatal("list: expected a non-nil result")
	}

	deleted, callErr := caller.Call("core.registry", "delete", handle, "alpha")
	if callErr != nil {
		t.Fatalf("delete: %v", callErr)
	}
	if deleted != true {
		t.Fatalf("delete: unexpected %#v", deleted)
	}

	has, callErr := caller.Call("core.registry", "has", handle, "alpha")
	if callErr != nil {
		t.Fatalf("has after delete: %v", callErr)
	}
	if has != false {
		t.Fatalf("expected key removed, got %#v", has)
	}
}

// TestRegistry_LockSealOpen_Good toggles the lock and seal lifecycle.
func TestRegistry_LockSealOpen_Good(t *core.T) {
	caller := newRegistryInterpreter(t)

	handle, callErr := caller.Call("core.registry", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.registry", "lock", handle); callErr != nil {
		t.Fatalf("lock: %v", callErr)
	}
	locked, callErr := caller.Call("core.registry", "locked", handle)
	if callErr != nil {
		t.Fatalf("locked: %v", callErr)
	}
	if locked != true {
		t.Fatalf("expected locked, got %#v", locked)
	}

	if _, callErr := caller.Call("core.registry", "seal", handle); callErr != nil {
		t.Fatalf("seal: %v", callErr)
	}
	sealed, callErr := caller.Call("core.registry", "sealed", handle)
	if callErr != nil {
		t.Fatalf("sealed: %v", callErr)
	}
	if sealed != true {
		t.Fatalf("expected sealed, got %#v", sealed)
	}

	if _, callErr := caller.Call("core.registry", "open", handle); callErr != nil {
		t.Fatalf("open: %v", callErr)
	}
	locked, callErr = caller.Call("core.registry", "locked", handle)
	if callErr != nil {
		t.Fatalf("locked after open: %v", callErr)
	}
	if locked != false {
		t.Fatalf("expected unlocked after open, got %#v", locked)
	}
}

// TestRegistry_Disable_Good disables a registered entry.
func TestRegistry_Disable_Good(t *core.T) {
	caller := newRegistryInterpreter(t)

	handle, callErr := caller.Call("core.registry", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}
	if _, callErr := caller.Call("core.registry", "set", handle, "alpha", 1); callErr != nil {
		t.Fatalf("set: %v", callErr)
	}

	if _, callErr := caller.Call("core.registry", "disable", handle, "alpha"); callErr != nil {
		t.Fatalf("disable: %v", callErr)
	}
}

// TestRegistry_Set_Bad reports a missing value argument.
func TestRegistry_Set_Bad(t *core.T) {
	caller := newRegistryInterpreter(t)

	handle, callErr := caller.Call("core.registry", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.registry", "set", handle, "alpha"); callErr == nil {
		t.Fatal("expected error for missing value argument")
	}
}

// TestRegistry_Get_Ugly rejects a non-registry handle argument.
func TestRegistry_Get_Ugly(t *core.T) {
	caller := newRegistryInterpreter(t)

	if _, callErr := caller.Call("core.registry", "get", "not-a-registry", "alpha"); callErr == nil {
		t.Fatal("expected error for non-registry handle")
	}
}
