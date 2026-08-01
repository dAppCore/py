package data

import (
	"os"
	"path/filepath"

	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newDataInterpreter registers the data module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newDataInterpreter(t)
//	handle, err := caller.Call("core.data", "new")
func newDataInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register data module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// fixtureDirectory writes a small content tree and returns its root.
func fixtureDirectory(t *core.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "greeting.txt"), []byte("hello corepy"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return root
}

// mountedData returns a data handle with a fixture directory mounted under
// the name "fixtures".
func mountedData(t *core.T, caller runtime.DirectCaller) any {
	t.Helper()
	handle, callErr := caller.Call("core.data", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}
	if _, callErr := caller.Call("core.data", "mount", handle, "fixtures", fixtureDirectory(t)); callErr != nil {
		t.Fatalf("mount: %v", callErr)
	}
	return handle
}

// TestData_ReadString_Good reads a mounted file as text.
func TestData_ReadString_Good(t *core.T) {
	caller := newDataInterpreter(t)
	handle := mountedData(t, caller)

	value, callErr := caller.Call("core.data", "read_string", handle, "fixtures/greeting.txt")
	if callErr != nil {
		t.Fatalf("read_string: %v", callErr)
	}
	if value != "hello corepy" {
		t.Fatalf("unexpected content %#v", value)
	}
}

// TestData_ReadFile_Good reads a mounted file as bytes.
func TestData_ReadFile_Good(t *core.T) {
	caller := newDataInterpreter(t)
	handle := mountedData(t, caller)

	value, callErr := caller.Call("core.data", "read_file", handle, "fixtures/greeting.txt")
	if callErr != nil {
		t.Fatalf("read_file: %v", callErr)
	}
	bytes, ok := value.([]byte)
	if !ok {
		t.Fatalf("read_file: expected []byte, got %T", value)
	}
	if string(bytes) != "hello corepy" {
		t.Fatalf("unexpected bytes %q", string(bytes))
	}
}

// TestData_Mounts_Good reports the mounted names.
func TestData_Mounts_Good(t *core.T) {
	caller := newDataInterpreter(t)
	handle := mountedData(t, caller)

	value, callErr := caller.Call("core.data", "mounts", handle)
	if callErr != nil {
		t.Fatalf("mounts: %v", callErr)
	}
	names, ok := value.([]string)
	if !ok {
		t.Fatalf("mounts: expected []string, got %T", value)
	}
	found := false
	for _, name := range names {
		if name == "fixtures" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected fixtures mount, got %#v", names)
	}
}

// TestData_Extract_Good extracts mounted content to a target directory.
func TestData_Extract_Good(t *core.T) {
	caller := newDataInterpreter(t)
	handle := mountedData(t, caller)

	target := t.TempDir()
	value, callErr := caller.Call("core.data", "extract", handle, "fixtures/greeting.txt", target)
	if callErr != nil {
		t.Fatalf("extract: %v", callErr)
	}
	if value != target {
		t.Fatalf("unexpected extract target %#v", value)
	}
}

// TestData_List_Good lists entries under a mounted path.
func TestData_List_Good(t *core.T) {
	caller := newDataInterpreter(t)
	handle := mountedData(t, caller)

	value, callErr := caller.Call("core.data", "list", handle, "fixtures")
	if callErr != nil {
		t.Fatalf("list: %v", callErr)
	}
	names, ok := value.([]string)
	if !ok {
		t.Fatalf("list: expected []string, got %T", value)
	}
	found := false
	for _, name := range names {
		if name == "greeting.txt" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected greeting.txt in listing, got %#v", names)
	}
}

// TestData_ListNames_Good lists names under a mounted path.
func TestData_ListNames_Good(t *core.T) {
	caller := newDataInterpreter(t)
	handle := mountedData(t, caller)

	value, callErr := caller.Call("core.data", "list_names", handle, "fixtures")
	if callErr != nil {
		t.Fatalf("list_names: %v", callErr)
	}
	if names, ok := value.([]string); !ok || len(names) == 0 {
		t.Fatalf("list_names: unexpected %#v", value)
	}
}

// TestData_ReadString_Bad reports a missing path argument.
func TestData_ReadString_Bad(t *core.T) {
	caller := newDataInterpreter(t)
	handle := mountedData(t, caller)

	if _, callErr := caller.Call("core.data", "read_string", handle); callErr == nil {
		t.Fatal("expected error for missing path argument")
	}
}

// TestData_ReadString_Ugly reports a read of an absent path.
func TestData_ReadString_Ugly(t *core.T) {
	caller := newDataInterpreter(t)
	handle := mountedData(t, caller)

	if _, callErr := caller.Call("core.data", "read_string", handle, "fixtures/absent.txt"); callErr == nil {
		t.Fatal("expected error reading absent path")
	}
}

// TestData_WrongHandle_Ugly rejects a non-Data handle argument.
func TestData_WrongHandle_Ugly(t *core.T) {
	caller := newDataInterpreter(t)

	if _, callErr := caller.Call("core.data", "mounts", "not-a-handle"); callErr == nil {
		t.Fatal("expected error for non-Data handle")
	}
}
