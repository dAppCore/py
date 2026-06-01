package pathbinding

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newPathInterpreter registers the path module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newPathInterpreter(t)
//	value, err := caller.Call("core.path", "base", "/a/b.txt")
func newPathInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register path module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestPath_Base_Good extracts the final path element.
func TestPath_Base_Good(t *core.T) {
	caller := newPathInterpreter(t)

	value, callErr := caller.Call("core.path", "base", "/var/data/store.db")
	if callErr != nil {
		t.Fatalf("base: %v", callErr)
	}
	if value != "store.db" {
		t.Fatalf("unexpected base %#v", value)
	}
}

// TestPath_Dir_Good extracts the parent directory.
func TestPath_Dir_Good(t *core.T) {
	caller := newPathInterpreter(t)

	value, callErr := caller.Call("core.path", "dir", "/var/data/store.db")
	if callErr != nil {
		t.Fatalf("dir: %v", callErr)
	}
	if value != "/var/data" {
		t.Fatalf("unexpected dir %#v", value)
	}
}

// TestPath_Ext_Good extracts the file extension.
func TestPath_Ext_Good(t *core.T) {
	caller := newPathInterpreter(t)

	value, callErr := caller.Call("core.path", "ext", "store.db")
	if callErr != nil {
		t.Fatalf("ext: %v", callErr)
	}
	if value != ".db" {
		t.Fatalf("unexpected ext %#v", value)
	}
}

// TestPath_IsAbs_Good reports absolute versus relative paths.
func TestPath_IsAbs_Good(t *core.T) {
	caller := newPathInterpreter(t)

	absolute, callErr := caller.Call("core.path", "is_abs", "/var/data")
	if callErr != nil {
		t.Fatalf("is_abs absolute: %v", callErr)
	}
	if absolute != true {
		t.Fatalf("expected absolute path, got %#v", absolute)
	}

	relative, callErr := caller.Call("core.path", "is_abs", "data/store.db")
	if callErr != nil {
		t.Fatalf("is_abs relative: %v", callErr)
	}
	if relative != false {
		t.Fatalf("expected relative path, got %#v", relative)
	}
}

// TestPath_Join_Good joins multiple segments.
func TestPath_Join_Good(t *core.T) {
	caller := newPathInterpreter(t)

	value, callErr := caller.Call("core.path", "join", "var", "data", "store.db")
	if callErr != nil {
		t.Fatalf("join: %v", callErr)
	}
	text, ok := value.(string)
	if !ok {
		t.Fatalf("expected string, got %T", value)
	}
	if text == "" {
		t.Fatal("expected non-empty joined path")
	}
}

// TestPath_Base_Bad reports a missing argument.
func TestPath_Base_Bad(t *core.T) {
	caller := newPathInterpreter(t)

	if _, callErr := caller.Call("core.path", "base"); callErr == nil {
		t.Fatal("expected error for missing path argument")
	}
}

// TestPath_Base_Ugly rejects a non-string argument.
func TestPath_Base_Ugly(t *core.T) {
	caller := newPathInterpreter(t)

	if _, callErr := caller.Call("core.path", "base", 123); callErr == nil {
		t.Fatal("expected error for non-string path argument")
	}
}

// TestPath_Join_Ugly rejects a non-string segment among the arguments.
func TestPath_Join_Ugly(t *core.T) {
	caller := newPathInterpreter(t)

	if _, callErr := caller.Call("core.path", "join", "var", 7); callErr == nil {
		t.Fatal("expected error for non-string join segment")
	}
}
