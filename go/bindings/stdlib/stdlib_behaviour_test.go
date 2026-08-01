package stdlib

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newStdlibInterpreter registers the stdlib shadow modules against a fresh
// bootstrap interpreter and returns the direct caller.
//
//	caller := newStdlibInterpreter(t)
//	value, err := caller.Call("os.path", "basename", "/a/b.txt")
func newStdlibInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register stdlib modules: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestStdlib_Base64_Good round-trips through the base64 shadow module.
func TestStdlib_Base64_Good(t *core.T) {
	caller := newStdlibInterpreter(t)

	encoded, callErr := caller.Call("base64", "b64encode", "corepy")
	if callErr != nil {
		t.Fatalf("b64encode: %v", callErr)
	}
	if encoded != "Y29yZXB5" {
		t.Fatalf("unexpected encoding %#v", encoded)
	}

	decoded, callErr := caller.Call("base64", "b64decode", encoded)
	if callErr != nil {
		t.Fatalf("b64decode: %v", callErr)
	}
	if bytes, ok := decoded.([]byte); !ok || string(bytes) != "corepy" {
		t.Fatalf("unexpected decoding %#v", decoded)
	}
}

// TestStdlib_JSON_Good round-trips through the json shadow module.
func TestStdlib_JSON_Good(t *core.T) {
	caller := newStdlibInterpreter(t)

	dumped, callErr := caller.Call("json", "dumps", map[string]any{"name": "corepy"})
	if callErr != nil {
		t.Fatalf("dumps: %v", callErr)
	}
	text, ok := dumped.(string)
	if !ok {
		t.Fatalf("dumps: expected string, got %T", dumped)
	}

	loaded, callErr := caller.Call("json", "loads", text)
	if callErr != nil {
		t.Fatalf("loads: %v", callErr)
	}
	if values, ok := loaded.(map[string]any); !ok || values["name"] != "corepy" {
		t.Fatalf("loads: unexpected %#v", loaded)
	}
}

// TestStdlib_Hashlib_Good computes a known SHA-256 hex digest.
func TestStdlib_Hashlib_Good(t *core.T) {
	caller := newStdlibInterpreter(t)

	digest, callErr := caller.Call("hashlib", "sha256", "abc")
	if callErr != nil {
		t.Fatalf("sha256: %v", callErr)
	}

	hex, callErr := caller.Call("hashlib", "_hexdigest", digest)
	if callErr != nil {
		t.Fatalf("_hexdigest: %v", callErr)
	}
	if hex != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("unexpected sha256 hex %#v", hex)
	}
}

// TestStdlib_OSPath_Good exercises os.path helpers.
func TestStdlib_OSPath_Good(t *core.T) {
	caller := newStdlibInterpreter(t)

	base, callErr := caller.Call("os.path", "basename", "/var/data/store.db")
	if callErr != nil {
		t.Fatalf("basename: %v", callErr)
	}
	if base != "store.db" {
		t.Fatalf("unexpected basename %#v", base)
	}

	joined, callErr := caller.Call("os.path", "join", "var", "data")
	if callErr != nil {
		t.Fatalf("join: %v", callErr)
	}
	if joined == "" {
		t.Fatal("expected non-empty joined path")
	}

	isAbs, callErr := caller.Call("os.path", "isabs", "/var/data")
	if callErr != nil {
		t.Fatalf("isabs: %v", callErr)
	}
	if isAbs != true {
		t.Fatalf("expected absolute path, got %#v", isAbs)
	}
}

// TestStdlib_OSGetenv_Good reads an environment variable with a default.
func TestStdlib_OSGetenv_Good(t *core.T) {
	t.Setenv("COREPY_STDLIB_TEST", "present")
	caller := newStdlibInterpreter(t)

	value, callErr := caller.Call("os", "getenv", "COREPY_STDLIB_TEST")
	if callErr != nil {
		t.Fatalf("getenv: %v", callErr)
	}
	if value != "present" {
		t.Fatalf("unexpected getenv value %#v", value)
	}

	fallback, callErr := caller.Call("os", "getenv", "COREPY_STDLIB_ABSENT", "default")
	if callErr != nil {
		t.Fatalf("getenv default: %v", callErr)
	}
	if fallback != "default" {
		t.Fatalf("unexpected getenv default %#v", fallback)
	}
}

// TestStdlib_OSGetcwd_Good returns the current working directory.
func TestStdlib_OSGetcwd_Good(t *core.T) {
	caller := newStdlibInterpreter(t)

	value, callErr := caller.Call("os", "getcwd")
	if callErr != nil {
		t.Fatalf("getcwd: %v", callErr)
	}
	if dir, ok := value.(string); !ok || dir == "" {
		t.Fatalf("getcwd: unexpected %#v", value)
	}
}

// TestStdlib_OSDirLifecycle_Good makes, lists, and removes a directory tree.
func TestStdlib_OSDirLifecycle_Good(t *core.T) {
	caller := newStdlibInterpreter(t)
	root := t.TempDir()

	child := root + "/nested/child"
	if _, callErr := caller.Call("os", "makedirs", child); callErr != nil {
		t.Fatalf("makedirs: %v", callErr)
	}

	value, callErr := caller.Call("os", "listdir", root)
	if callErr != nil {
		t.Fatalf("listdir: %v", callErr)
	}
	names, ok := value.([]string)
	if !ok || len(names) != 1 || names[0] != "nested" {
		t.Fatalf("listdir: unexpected %#v", value)
	}

	if _, callErr := caller.Call("os", "remove", child); callErr != nil {
		t.Fatalf("remove: %v", callErr)
	}
}

// TestStdlib_Logging_Good emits at each logging level.
func TestStdlib_Logging_Good(t *core.T) {
	caller := newStdlibInterpreter(t)

	for _, function := range []string{"debug", "info", "warning", "error"} {
		if _, callErr := caller.Call("logging", function, "stdlib %s", "log"); callErr != nil {
			t.Fatalf("logging %s: %v", function, callErr)
		}
	}

	// A non-string message is rejected.
	if _, callErr := caller.Call("logging", "info", 42); callErr == nil {
		t.Fatal("expected error for non-string logging message")
	}
}

// TestStdlib_SubprocessGetOutput_Good captures combined output from sh.
func TestStdlib_SubprocessGetOutput_Good(t *core.T) {
	caller := newStdlibInterpreter(t)

	value, callErr := caller.Call("subprocess", "getoutput", "echo stdlib")
	if callErr != nil {
		t.Fatalf("getoutput: %v", callErr)
	}
	text, ok := value.(string)
	if !ok || text != "stdlib\n" {
		t.Fatalf("getoutput: unexpected %#v", value)
	}
}

// TestStdlib_Base64_Bad reports a missing argument.
func TestStdlib_Base64_Bad(t *core.T) {
	caller := newStdlibInterpreter(t)

	if _, callErr := caller.Call("base64", "b64encode"); callErr == nil {
		t.Fatal("expected error for missing argument")
	}
}

// TestStdlib_Base64Decode_Ugly rejects malformed base64.
func TestStdlib_Base64Decode_Ugly(t *core.T) {
	caller := newStdlibInterpreter(t)

	if _, callErr := caller.Call("base64", "b64decode", "*** not base64 ***"); callErr == nil {
		t.Fatal("expected error for malformed base64")
	}
}
