package typemap

import (
	core "dappco.re/go"
)

// TestTypemap_ExpectString_Real_Good returns the string at the index.
func TestTypemap_ExpectString_Real_Good(t *core.T) {
	value, err := ExpectString([]any{"hello"}, 0, "test")
	if err != nil {
		t.Fatalf("ExpectString: %v", err)
	}
	if value != "hello" {
		t.Fatalf("unexpected value %q", value)
	}
}

// TestTypemap_ExpectString_Real_Bad reports a missing index.
func TestTypemap_ExpectString_Real_Bad(t *core.T) {
	if _, err := ExpectString(nil, 0, "test"); err == nil {
		t.Fatal("expected error for missing argument")
	}
}

// TestTypemap_ExpectString_Real_Ugly rejects a non-string value.
func TestTypemap_ExpectString_Real_Ugly(t *core.T) {
	if _, err := ExpectString([]any{42}, 0, "test"); err == nil {
		t.Fatal("expected error for non-string value")
	}
}

// TestTypemap_ExpectInt_Real_Good returns the int at the index.
func TestTypemap_ExpectInt_Real_Good(t *core.T) {
	value, err := ExpectInt([]any{7}, 0, "test")
	if err != nil {
		t.Fatalf("ExpectInt: %v", err)
	}
	if value != 7 {
		t.Fatalf("unexpected value %d", value)
	}
}

// TestTypemap_ExpectBytes_Real_Good accepts string and []byte inputs.
func TestTypemap_ExpectBytes_Real_Good(t *core.T) {
	fromString, err := ExpectBytes([]any{"abc"}, 0, "test")
	if err != nil || string(fromString) != "abc" {
		t.Fatalf("ExpectBytes from string: %v / %q", err, fromString)
	}
	fromBytes, err := ExpectBytes([]any{[]byte{1, 2}}, 0, "test")
	if err != nil || len(fromBytes) != 2 {
		t.Fatalf("ExpectBytes from []byte: %v / %#v", err, fromBytes)
	}
	if _, err := ExpectBytes([]any{42}, 0, "test"); err == nil {
		t.Fatal("expected error for non-bytes value")
	}
}

// TestTypemap_ExpectMap_Real_Good returns a map argument.
func TestTypemap_ExpectMap_Real_Good(t *core.T) {
	value, err := ExpectMap([]any{map[string]any{"k": "v"}}, 0, "test")
	if err != nil || value["k"] != "v" {
		t.Fatalf("ExpectMap: %v / %#v", err, value)
	}
	if _, err := ExpectMap([]any{"not-a-map"}, 0, "test"); err == nil {
		t.Fatal("expected error for non-map value")
	}
}

// TestTypemap_OptionsRoundTrip_Real_Good converts maps to options and back.
func TestTypemap_OptionsRoundTrip_Real_Good(t *core.T) {
	options := MapToOptions(map[string]any{"name": "corepy", "port": 8080})
	values := OptionsToMap(options)
	if values["name"] != "corepy" || values["port"] != 8080 {
		t.Fatalf("round-trip mismatch %#v", values)
	}
	if OptionsToMap(nil) == nil {
		t.Fatal("OptionsToMap(nil) should return an empty map, not nil")
	}
}

// TestTypemap_ExpectOptions_Real_Good accepts pointer, value, and map forms.
func TestTypemap_ExpectOptions_Real_Good(t *core.T) {
	base := core.NewOptions(core.Option{Key: "name", Value: "corepy"})

	fromPointer, err := ExpectOptions([]any{&base}, 0, "test")
	if err != nil || fromPointer == nil {
		t.Fatalf("ExpectOptions pointer: %v", err)
	}
	fromValue, err := ExpectOptions([]any{base}, 0, "test")
	if err != nil || fromValue == nil {
		t.Fatalf("ExpectOptions value: %v", err)
	}
	fromMap, err := ExpectOptions([]any{map[string]any{"name": "corepy"}}, 0, "test")
	if err != nil || fromMap == nil {
		t.Fatalf("ExpectOptions map: %v", err)
	}
	if _, err := ExpectOptions([]any{42}, 0, "test"); err == nil {
		t.Fatal("expected error for incompatible Options value")
	}
}

// TestTypemap_ResultValue_Real_Good unwraps an OK result and surfaces failures.
func TestTypemap_ResultValue_Real_Good(t *core.T) {
	value, err := ResultValue(core.Result{Value: "ok", OK: true}, "test")
	if err != nil || value != "ok" {
		t.Fatalf("ResultValue OK: %v / %#v", err, value)
	}

	if _, err := ResultValue(core.Result{OK: false}, "test"); err == nil {
		t.Fatal("expected error for nil-value failed result")
	}

	cause := core.E("op", "boom", nil)
	if _, err := ResultValue(core.Result{Value: cause, OK: false}, "test"); err == nil {
		t.Fatal("expected error for error-value failed result")
	}
}

// TestTypemap_ExpectError_Real_Good distinguishes errors from non-errors.
func TestTypemap_ExpectError_Real_Good(t *core.T) {
	cause := core.E("op", "boom", nil)
	value, err := ExpectError([]any{cause}, 0, "test")
	if err != nil || value == nil {
		t.Fatalf("ExpectError: %v / %#v", err, value)
	}
	if _, err := ExpectError([]any{"not-an-error"}, 0, "test"); err == nil {
		t.Fatal("expected error for non-error value")
	}
}

// TestTypemap_OptionalError_Real_Good returns nil for a nil argument.
func TestTypemap_OptionalError_Real_Good(t *core.T) {
	value, err := OptionalError([]any{nil}, 0, "test")
	if err != nil || value != nil {
		t.Fatalf("OptionalError nil: %v / %#v", err, value)
	}
	if _, err := OptionalError(nil, 0, "test"); err == nil {
		t.Fatal("expected error for missing argument")
	}
}

// TestTypemap_ExpectCoreConfigData_Real_Ugly rejects mismatched handle types.
func TestTypemap_ExpectCoreConfigData_Real_Ugly(t *core.T) {
	if _, err := ExpectConfig([]any{"x"}, 0, "test"); err == nil {
		t.Fatal("expected error for non-Config value")
	}
	if _, err := ExpectData([]any{"x"}, 0, "test"); err == nil {
		t.Fatal("expected error for non-Data value")
	}
	if _, err := ExpectCore([]any{"x"}, 0, "test"); err == nil {
		t.Fatal("expected error for non-Core value")
	}
}
