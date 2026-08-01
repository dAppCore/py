package config

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newConfigInterpreter registers the config module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newConfigInterpreter(t)
//	handle, err := caller.Call("core.config", "new")
func newConfigInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register config module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestConfig_TypedValues_Good sets typed values and reads them back.
func TestConfig_TypedValues_Good(t *core.T) {
	caller := newConfigInterpreter(t)

	handle, callErr := caller.Call("core.config", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.config", "set", handle, "name", "corepy"); callErr != nil {
		t.Fatalf("set name: %v", callErr)
	}
	if _, callErr := caller.Call("core.config", "set", handle, "port", 9000); callErr != nil {
		t.Fatalf("set port: %v", callErr)
	}
	if _, callErr := caller.Call("core.config", "set", handle, "debug", true); callErr != nil {
		t.Fatalf("set debug: %v", callErr)
	}

	name, callErr := caller.Call("core.config", "string", handle, "name")
	if callErr != nil {
		t.Fatalf("string: %v", callErr)
	}
	if name != "corepy" {
		t.Fatalf("unexpected name %#v", name)
	}

	port, callErr := caller.Call("core.config", "int", handle, "port")
	if callErr != nil {
		t.Fatalf("int: %v", callErr)
	}
	if port != 9000 {
		t.Fatalf("unexpected port %#v", port)
	}

	debug, callErr := caller.Call("core.config", "bool", handle, "debug")
	if callErr != nil {
		t.Fatalf("bool: %v", callErr)
	}
	if debug != true {
		t.Fatalf("unexpected debug %#v", debug)
	}

	value, callErr := caller.Call("core.config", "get", handle, "name")
	if callErr != nil {
		t.Fatalf("get: %v", callErr)
	}
	if value != "corepy" {
		t.Fatalf("unexpected get value %#v", value)
	}
}

// TestConfig_Features_Good enables and disables features.
func TestConfig_Features_Good(t *core.T) {
	caller := newConfigInterpreter(t)

	handle, callErr := caller.Call("core.config", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.config", "enable", handle, "telemetry"); callErr != nil {
		t.Fatalf("enable: %v", callErr)
	}
	enabled, callErr := caller.Call("core.config", "enabled", handle, "telemetry")
	if callErr != nil {
		t.Fatalf("enabled: %v", callErr)
	}
	if enabled != true {
		t.Fatalf("expected feature enabled, got %#v", enabled)
	}

	features, callErr := caller.Call("core.config", "enabled_features", handle)
	if callErr != nil {
		t.Fatalf("enabled_features: %v", callErr)
	}
	if list, ok := features.([]string); !ok || len(list) != 1 || list[0] != "telemetry" {
		t.Fatalf("unexpected enabled_features %#v", features)
	}

	if _, callErr := caller.Call("core.config", "disable", handle, "telemetry"); callErr != nil {
		t.Fatalf("disable: %v", callErr)
	}
	disabled, callErr := caller.Call("core.config", "enabled", handle, "telemetry")
	if callErr != nil {
		t.Fatalf("enabled after disable: %v", callErr)
	}
	if disabled != false {
		t.Fatalf("expected feature disabled, got %#v", disabled)
	}
}

// TestConfig_EnvironmentFallback_Good falls back to a normalised environment
// variable when the key is not set in the config handle.
func TestConfig_EnvironmentFallback_Good(t *core.T) {
	t.Setenv("CORE_DB_PORT", "5432")
	caller := newConfigInterpreter(t)

	handle, callErr := caller.Call("core.config", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	value, callErr := caller.Call("core.config", "int", handle, "core.db.port")
	if callErr != nil {
		t.Fatalf("int env fallback: %v", callErr)
	}
	if value != 5432 {
		t.Fatalf("unexpected env-fallback int %#v", value)
	}

	text, callErr := caller.Call("core.config", "string", handle, "core.db.port")
	if callErr != nil {
		t.Fatalf("string env fallback: %v", callErr)
	}
	if text != "5432" {
		t.Fatalf("unexpected env-fallback string %#v", text)
	}
}

// TestConfig_GetMissing_Good returns nil for an absent, unset key.
func TestConfig_GetMissing_Good(t *core.T) {
	caller := newConfigInterpreter(t)

	handle, callErr := caller.Call("core.config", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	value, callErr := caller.Call("core.config", "get", handle, "definitely.absent.key")
	if callErr != nil {
		t.Fatalf("get absent: %v", callErr)
	}
	if value != nil {
		t.Fatalf("expected nil for absent key, got %#v", value)
	}
}

// TestConfig_Set_Bad reports a missing value argument.
func TestConfig_Set_Bad(t *core.T) {
	caller := newConfigInterpreter(t)

	handle, callErr := caller.Call("core.config", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.config", "set", handle, "name"); callErr == nil {
		t.Fatal("expected error for missing value argument")
	}
}

// TestConfig_WrongHandle_Ugly rejects a non-Config handle argument.
func TestConfig_WrongHandle_Ugly(t *core.T) {
	caller := newConfigInterpreter(t)

	if _, callErr := caller.Call("core.config", "get", 42, "name"); callErr == nil {
		t.Fatal("expected error for non-Config handle")
	}
}
