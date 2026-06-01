package json

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newJSONInterpreter registers the json module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newJSONInterpreter(t)
//	value, err := caller.Call("core.json", "dumps", map[string]any{"k": 1})
func newJSONInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register json module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestJSON_RoundTrip_Good marshals a map and parses it back.
func TestJSON_RoundTrip_Good(t *core.T) {
	caller := newJSONInterpreter(t)

	dumped, callErr := caller.Call("core.json", "dumps", map[string]any{"name": "corepy"})
	if callErr != nil {
		t.Fatalf("dumps: %v", callErr)
	}
	text, ok := dumped.(string)
	if !ok {
		t.Fatalf("dumps: expected string, got %T", dumped)
	}

	loaded, callErr := caller.Call("core.json", "loads", text)
	if callErr != nil {
		t.Fatalf("loads: %v", callErr)
	}
	values, ok := loaded.(map[string]any)
	if !ok {
		t.Fatalf("loads: expected map, got %T", loaded)
	}
	if values["name"] != "corepy" {
		t.Fatalf("unexpected round-trip value %#v", values["name"])
	}
}

// TestJSON_Dumps_Bad reports a wrong argument count.
func TestJSON_Dumps_Bad(t *core.T) {
	caller := newJSONInterpreter(t)

	if _, callErr := caller.Call("core.json", "dumps"); callErr == nil {
		t.Fatal("expected error for missing argument")
	}
	if _, callErr := caller.Call("core.json", "dumps", 1, 2); callErr == nil {
		t.Fatal("expected error for extra argument")
	}
}

// TestJSON_Loads_Bad reports a missing string argument.
func TestJSON_Loads_Bad(t *core.T) {
	caller := newJSONInterpreter(t)

	if _, callErr := caller.Call("core.json", "loads"); callErr == nil {
		t.Fatal("expected error for missing string argument")
	}
}

// TestJSON_Loads_Ugly rejects malformed JSON text.
func TestJSON_Loads_Ugly(t *core.T) {
	caller := newJSONInterpreter(t)

	if _, callErr := caller.Call("core.json", "loads", "{not valid json"); callErr == nil {
		t.Fatal("expected error for malformed JSON")
	}
}
