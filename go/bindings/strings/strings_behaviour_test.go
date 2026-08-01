package stringsbinding

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newStringsInterpreter registers the strings module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newStringsInterpreter(t)
//	value, err := caller.Call("core.strings", "upper", "abc")
func newStringsInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register strings module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// callString invokes a binding and asserts a string result.
func callString(t *core.T, caller runtime.DirectCaller, function string, arguments ...any) string {
	t.Helper()
	value, callErr := caller.Call("core.strings", function, arguments...)
	if callErr != nil {
		t.Fatalf("%s: %v", function, callErr)
	}
	text, ok := value.(string)
	if !ok {
		t.Fatalf("%s: expected string, got %T", function, value)
	}
	return text
}

// TestStrings_Transform_Good exercises the single-value transforms.
func TestStrings_Transform_Good(t *core.T) {
	caller := newStringsInterpreter(t)

	if got := callString(t, caller, "lower", "ABC"); got != "abc" {
		t.Fatalf("lower: %q", got)
	}
	if got := callString(t, caller, "upper", "abc"); got != "ABC" {
		t.Fatalf("upper: %q", got)
	}
	if got := callString(t, caller, "trim", "  padded  "); got != "padded" {
		t.Fatalf("trim: %q", got)
	}
	if got := callString(t, caller, "trim_prefix", "core.err", "core."); got != "err" {
		t.Fatalf("trim_prefix: %q", got)
	}
	if got := callString(t, caller, "trim_suffix", "store.db", ".db"); got != "store" {
		t.Fatalf("trim_suffix: %q", got)
	}
	if got := callString(t, caller, "replace", "a-b-c", "-", "_"); got != "a_b_c" {
		t.Fatalf("replace: %q", got)
	}
}

// TestStrings_Predicate_Good exercises the boolean predicates.
func TestStrings_Predicate_Good(t *core.T) {
	caller := newStringsInterpreter(t)

	contains, callErr := caller.Call("core.strings", "contains", "abcdef", "cde")
	if callErr != nil {
		t.Fatalf("contains: %v", callErr)
	}
	if contains != true {
		t.Fatalf("contains: %#v", contains)
	}

	hasPrefix, callErr := caller.Call("core.strings", "has_prefix", "core.err", "core.")
	if callErr != nil {
		t.Fatalf("has_prefix: %v", callErr)
	}
	if hasPrefix != true {
		t.Fatalf("has_prefix: %#v", hasPrefix)
	}

	hasSuffix, callErr := caller.Call("core.strings", "has_suffix", "store.db", ".db")
	if callErr != nil {
		t.Fatalf("has_suffix: %v", callErr)
	}
	if hasSuffix != true {
		t.Fatalf("has_suffix: %#v", hasSuffix)
	}
}

// TestStrings_Split_Good exercises split, split_n, join, and concat.
func TestStrings_Split_Good(t *core.T) {
	caller := newStringsInterpreter(t)

	if got := callString(t, caller, "join", "-", "a", "b", "c"); got != "a-b-c" {
		t.Fatalf("join: %q", got)
	}
	if got := callString(t, caller, "concat", "a", "b", "c"); got != "abc" {
		t.Fatalf("concat: %q", got)
	}

	split, callErr := caller.Call("core.strings", "split", "a-b-c", "-")
	if callErr != nil {
		t.Fatalf("split: %v", callErr)
	}
	if parts, ok := split.([]string); !ok || len(parts) != 3 {
		t.Fatalf("split: unexpected %#v", split)
	}

	splitN, callErr := caller.Call("core.strings", "split_n", "a-b-c", "-", 2)
	if callErr != nil {
		t.Fatalf("split_n: %v", callErr)
	}
	if parts, ok := splitN.([]string); !ok || len(parts) != 2 {
		t.Fatalf("split_n: unexpected %#v", splitN)
	}
}

// TestStrings_RuneCount_Good counts runes including multi-byte characters.
func TestStrings_RuneCount_Good(t *core.T) {
	caller := newStringsInterpreter(t)

	value, callErr := caller.Call("core.strings", "rune_count", "café")
	if callErr != nil {
		t.Fatalf("rune_count: %v", callErr)
	}
	if value != 4 {
		t.Fatalf("rune_count: %#v", value)
	}
}

// TestStrings_MissingArgument_Bad reports a missing second argument.
func TestStrings_MissingArgument_Bad(t *core.T) {
	caller := newStringsInterpreter(t)

	if _, callErr := caller.Call("core.strings", "contains", "abc"); callErr == nil {
		t.Fatal("expected error for missing substring argument")
	}
	if _, callErr := caller.Call("core.strings", "split_n", "a-b-c", "-"); callErr == nil {
		t.Fatal("expected error for missing limit argument")
	}
}

// TestStrings_WrongType_Ugly rejects non-string and non-int arguments.
func TestStrings_WrongType_Ugly(t *core.T) {
	caller := newStringsInterpreter(t)

	if _, callErr := caller.Call("core.strings", "upper", 99); callErr == nil {
		t.Fatal("expected error for non-string value")
	}
	if _, callErr := caller.Call("core.strings", "split_n", "a-b", "-", "two"); callErr == nil {
		t.Fatal("expected error for non-int limit")
	}
	if _, callErr := caller.Call("core.strings", "join", "-", 7); callErr == nil {
		t.Fatal("expected error for non-string join part")
	}
}
