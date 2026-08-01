package info

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newInfoInterpreter registers the info module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newInfoInterpreter(t)
//	value, err := caller.Call("core.info", "keys")
func newInfoInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register info module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestInfo_Env_Good reads a known environment key set for the test.
func TestInfo_Env_Good(t *core.T) {
	t.Setenv("COREPY_INFO_TEST", "present")
	caller := newInfoInterpreter(t)

	value, callErr := caller.Call("core.info", "env", "COREPY_INFO_TEST")
	if callErr != nil {
		t.Fatalf("env: %v", callErr)
	}
	if value != "present" {
		t.Fatalf("unexpected env value %#v", value)
	}
}

// TestInfo_Keys_Good returns a sorted slice that includes Core's own system
// keys (the binding reports Core-managed SysInfo keys, not the OS environment).
func TestInfo_Keys_Good(t *core.T) {
	caller := newInfoInterpreter(t)

	value, callErr := caller.Call("core.info", "keys")
	if callErr != nil {
		t.Fatalf("keys: %v", callErr)
	}
	keys, ok := value.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", value)
	}
	if len(keys) == 0 {
		t.Fatal("expected at least Core's own system keys")
	}
	for index := 1; index < len(keys); index++ {
		if keys[index] < keys[index-1] {
			t.Fatalf("keys not sorted at %d: %q < %q", index, keys[index], keys[index-1])
		}
	}
	found := false
	for _, key := range keys {
		if key == "OS" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected Core system key OS in keys snapshot")
	}
}

// TestInfo_Snapshot_Good returns a map whose entries match core.Env for the
// reported keys.
func TestInfo_Snapshot_Good(t *core.T) {
	caller := newInfoInterpreter(t)

	value, callErr := caller.Call("core.info", "snapshot")
	if callErr != nil {
		t.Fatalf("snapshot: %v", callErr)
	}
	snapshot, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", value)
	}
	if len(snapshot) == 0 {
		t.Fatal("expected a non-empty snapshot")
	}
	if snapshot["OS"] != core.Env("OS") {
		t.Fatalf("snapshot OS %#v does not match core.Env %q", snapshot["OS"], core.Env("OS"))
	}
}

// TestInfo_Env_Bad reports a missing key argument.
func TestInfo_Env_Bad(t *core.T) {
	caller := newInfoInterpreter(t)

	if _, callErr := caller.Call("core.info", "env"); callErr == nil {
		t.Fatal("expected error for missing key argument")
	}
}

// TestInfo_Env_Ugly rejects a non-string key argument.
func TestInfo_Env_Ugly(t *core.T) {
	caller := newInfoInterpreter(t)

	if _, callErr := caller.Call("core.info", "env", 1); callErr == nil {
		t.Fatal("expected error for non-string key argument")
	}
}
