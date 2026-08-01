package medium

import (
	"path/filepath"

	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newMediumInterpreter registers the medium module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newMediumInterpreter(t)
//	handle, err := caller.Call("core.medium", "memory", "seed")
func newMediumInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register medium module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestMedium_Memory_Good round-trips text and bytes through a memory medium.
func TestMedium_Memory_Good(t *core.T) {
	caller := newMediumInterpreter(t)

	handle, callErr := caller.Call("core.medium", "memory", "seed")
	if callErr != nil {
		t.Fatalf("memory: %v", callErr)
	}

	text, callErr := caller.Call("core.medium", "read_text", handle)
	if callErr != nil {
		t.Fatalf("read_text: %v", callErr)
	}
	if text != "seed" {
		t.Fatalf("unexpected seed text %#v", text)
	}

	if _, callErr := caller.Call("core.medium", "write_text", handle, "updated"); callErr != nil {
		t.Fatalf("write_text: %v", callErr)
	}
	text, callErr = caller.Call("core.medium", "read_text", handle)
	if callErr != nil {
		t.Fatalf("read_text after write: %v", callErr)
	}
	if text != "updated" {
		t.Fatalf("unexpected updated text %#v", text)
	}

	bytes, callErr := caller.Call("core.medium", "read_bytes", handle)
	if callErr != nil {
		t.Fatalf("read_bytes: %v", callErr)
	}
	if data, ok := bytes.([]byte); !ok || string(data) != "updated" {
		t.Fatalf("unexpected read_bytes %#v", bytes)
	}
}

// TestMedium_File_Good round-trips through a filesystem-backed medium.
func TestMedium_File_Good(t *core.T) {
	caller := newMediumInterpreter(t)
	path := filepath.Join(t.TempDir(), "nested", "content.txt")

	handle, callErr := caller.Call("core.medium", "from_path", path)
	if callErr != nil {
		t.Fatalf("from_path: %v", callErr)
	}

	if _, callErr := caller.Call("core.medium", "write_text", handle, "on disk"); callErr != nil {
		t.Fatalf("write_text: %v", callErr)
	}
	text, callErr := caller.Call("core.medium", "read_text", handle)
	if callErr != nil {
		t.Fatalf("read_text: %v", callErr)
	}
	if text != "on disk" {
		t.Fatalf("unexpected file text %#v", text)
	}

	if _, callErr := caller.Call("core.medium", "write_bytes", handle, []byte("raw bytes")); callErr != nil {
		t.Fatalf("write_bytes: %v", callErr)
	}
	bytes, callErr := caller.Call("core.medium", "read_bytes", handle)
	if callErr != nil {
		t.Fatalf("read_bytes: %v", callErr)
	}
	if data, ok := bytes.([]byte); !ok || string(data) != "raw bytes" {
		t.Fatalf("unexpected file bytes %#v", bytes)
	}
}

// TestMedium_WriteText_Bad reports a missing value argument.
func TestMedium_WriteText_Bad(t *core.T) {
	caller := newMediumInterpreter(t)

	handle, callErr := caller.Call("core.medium", "memory")
	if callErr != nil {
		t.Fatalf("memory: %v", callErr)
	}

	if _, callErr := caller.Call("core.medium", "write_text", handle); callErr == nil {
		t.Fatal("expected error for missing value argument")
	}
}

// TestMedium_ReadText_Ugly rejects a non-medium handle argument.
func TestMedium_ReadText_Ugly(t *core.T) {
	caller := newMediumInterpreter(t)

	if _, callErr := caller.Call("core.medium", "read_text", "not-a-handle"); callErr == nil {
		t.Fatal("expected error for non-medium handle")
	}
}
