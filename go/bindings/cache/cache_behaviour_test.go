package cache

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newCacheInterpreter registers the cache module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newCacheInterpreter(t)
//	handle, err := caller.Call("core.cache", "new", baseDir)
func newCacheInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register cache module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// newCacheHandle creates a cache rooted in a temp directory.
func newCacheHandle(t *core.T, caller runtime.DirectCaller) any {
	t.Helper()
	handle, callErr := caller.Call("core.cache", "new", t.TempDir())
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}
	return handle
}

// TestCache_SetGetHas_Good stores and retrieves a value.
func TestCache_SetGetHas_Good(t *core.T) {
	caller := newCacheInterpreter(t)
	handle := newCacheHandle(t, caller)

	if _, callErr := caller.Call("core.cache", "set", handle, "alpha", "value-a"); callErr != nil {
		t.Fatalf("set: %v", callErr)
	}

	value, callErr := caller.Call("core.cache", "get", handle, "alpha")
	if callErr != nil {
		t.Fatalf("get: %v", callErr)
	}
	if value != "value-a" {
		t.Fatalf("unexpected cached value %#v", value)
	}

	has, callErr := caller.Call("core.cache", "has", handle, "alpha")
	if callErr != nil {
		t.Fatalf("has: %v", callErr)
	}
	if has != true {
		t.Fatalf("expected key present, got %#v", has)
	}
}

// TestCache_GetDefault_Good returns the default for a missing key.
func TestCache_GetDefault_Good(t *core.T) {
	caller := newCacheInterpreter(t)
	handle := newCacheHandle(t, caller)

	value, callErr := caller.Call("core.cache", "get", handle, "absent", "fallback")
	if callErr != nil {
		t.Fatalf("get default: %v", callErr)
	}
	if value != "fallback" {
		t.Fatalf("expected fallback, got %#v", value)
	}
}

// TestCache_DeleteClearKeys_Good deletes, lists, and clears entries.
func TestCache_DeleteClearKeys_Good(t *core.T) {
	caller := newCacheInterpreter(t)
	handle := newCacheHandle(t, caller)

	if _, callErr := caller.Call("core.cache", "set", handle, "group/one", 1); callErr != nil {
		t.Fatalf("set one: %v", callErr)
	}
	if _, callErr := caller.Call("core.cache", "set", handle, "group/two", 2); callErr != nil {
		t.Fatalf("set two: %v", callErr)
	}

	value, callErr := caller.Call("core.cache", "keys", handle, "group")
	if callErr != nil {
		t.Fatalf("keys: %v", callErr)
	}
	if keys, ok := value.([]string); !ok || len(keys) != 2 {
		t.Fatalf("keys: unexpected %#v", value)
	}

	deleted, callErr := caller.Call("core.cache", "delete", handle, "group/one")
	if callErr != nil {
		t.Fatalf("delete: %v", callErr)
	}
	if deleted != true {
		t.Fatalf("expected delete to report removal, got %#v", deleted)
	}

	removed, callErr := caller.Call("core.cache", "clear", handle, "group")
	if callErr != nil {
		t.Fatalf("clear: %v", callErr)
	}
	if removed != 1 {
		t.Fatalf("expected one remaining entry cleared, got %#v", removed)
	}
}

// TestCache_SetWithTTL_Good stores with an explicit ttl and reads it back.
func TestCache_SetWithTTL_Good(t *core.T) {
	caller := newCacheInterpreter(t)
	handle := newCacheHandle(t, caller)

	if _, callErr := caller.Call("core.cache", "set_with_ttl", handle, "ttl-key", "value", 3600); callErr != nil {
		t.Fatalf("set_with_ttl: %v", callErr)
	}
	value, callErr := caller.Call("core.cache", "get", handle, "ttl-key")
	if callErr != nil {
		t.Fatalf("get ttl: %v", callErr)
	}
	if value != "value" {
		t.Fatalf("unexpected ttl value %#v", value)
	}
}

// TestCache_Set_Bad reports a missing value argument.
func TestCache_Set_Bad(t *core.T) {
	caller := newCacheInterpreter(t)
	handle := newCacheHandle(t, caller)

	if _, callErr := caller.Call("core.cache", "set", handle, "alpha"); callErr == nil {
		t.Fatal("expected error for missing value argument")
	}
}

// TestCache_Path_Ugly rejects an unsafe key containing a parent traversal.
func TestCache_Path_Ugly(t *core.T) {
	caller := newCacheInterpreter(t)
	handle := newCacheHandle(t, caller)

	if _, callErr := caller.Call("core.cache", "path", handle, "../escape"); callErr == nil {
		t.Fatal("expected error for parent-traversal key")
	}
}
