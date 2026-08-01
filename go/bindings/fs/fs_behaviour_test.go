package fs

import (
	"os"
	"path/filepath"

	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newFsInterpreter registers the fs module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newFsInterpreter(t)
//	value, err := caller.Call("core.fs", "read_file", path)
func newFsInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register fs module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestFs_WriteRead_Good writes a file and reads it back as text.
func TestFs_WriteRead_Good(t *core.T) {
	caller := newFsInterpreter(t)
	path := filepath.Join(t.TempDir(), "note.txt")

	if _, callErr := caller.Call("core.fs", "write_file", path, "hello corepy"); callErr != nil {
		t.Fatalf("write_file: %v", callErr)
	}

	value, callErr := caller.Call("core.fs", "read_file", path)
	if callErr != nil {
		t.Fatalf("read_file: %v", callErr)
	}
	if value != "hello corepy" {
		t.Fatalf("unexpected content %#v", value)
	}
}

// TestFs_WriteReadBytes_Good writes raw bytes and reads them back.
func TestFs_WriteReadBytes_Good(t *core.T) {
	caller := newFsInterpreter(t)
	path := filepath.Join(t.TempDir(), "nested", "data.bin")

	if _, callErr := caller.Call("core.fs", "write_bytes", path, []byte{1, 2, 3}); callErr != nil {
		t.Fatalf("write_bytes: %v", callErr)
	}

	value, callErr := caller.Call("core.fs", "read_bytes", path)
	if callErr != nil {
		t.Fatalf("read_bytes: %v", callErr)
	}
	bytes, ok := value.([]byte)
	if !ok || len(bytes) != 3 || bytes[2] != 3 {
		t.Fatalf("unexpected bytes %#v", value)
	}
}

// TestFs_EnsureDir_Good creates a directory tree.
func TestFs_EnsureDir_Good(t *core.T) {
	caller := newFsInterpreter(t)
	path := filepath.Join(t.TempDir(), "a", "b", "c")

	if _, callErr := caller.Call("core.fs", "ensure_dir", path); callErr != nil {
		t.Fatalf("ensure_dir: %v", callErr)
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		t.Fatalf("expected directory at %q: %v", path, err)
	}
}

// TestFs_TempDir_Good returns a usable temporary directory.
func TestFs_TempDir_Good(t *core.T) {
	caller := newFsInterpreter(t)

	value, callErr := caller.Call("core.fs", "temp_dir", "corepy-test-")
	if callErr != nil {
		t.Fatalf("temp_dir: %v", callErr)
	}
	dir, ok := value.(string)
	if !ok || dir == "" {
		t.Fatalf("temp_dir: unexpected %#v", value)
	}
}

// TestFs_ReadFile_Bad reports a missing path argument.
func TestFs_ReadFile_Bad(t *core.T) {
	caller := newFsInterpreter(t)

	if _, callErr := caller.Call("core.fs", "read_file"); callErr == nil {
		t.Fatal("expected error for missing path argument")
	}
}

// TestFs_ReadBytes_Ugly reports a read of an absent file.
func TestFs_ReadBytes_Ugly(t *core.T) {
	caller := newFsInterpreter(t)

	if _, callErr := caller.Call("core.fs", "read_bytes", filepath.Join(t.TempDir(), "absent.bin")); callErr == nil {
		t.Fatal("expected error reading an absent file")
	}
}
