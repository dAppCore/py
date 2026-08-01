package process

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newProcessInterpreter registers the process module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newProcessInterpreter(t)
//	value, err := caller.Call("core.process", "run", "echo", "hi")
func newProcessInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register process module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestProcess_Run_Good runs echo and captures stdout.
func TestProcess_Run_Good(t *core.T) {
	caller := newProcessInterpreter(t)

	value, callErr := caller.Call("core.process", "run", "echo", "hello corepy")
	if callErr != nil {
		t.Fatalf("run: %v", callErr)
	}
	text, ok := value.(string)
	if !ok {
		t.Fatalf("run: expected string stdout, got %T", value)
	}
	if text != "hello corepy\n" {
		t.Fatalf("unexpected stdout %q", text)
	}
}

// TestProcess_RunResult_Good runs echo and inspects the result map.
func TestProcess_RunResult_Good(t *core.T) {
	caller := newProcessInterpreter(t)

	value, callErr := caller.Call("core.process", "run_result", "echo", "ok")
	if callErr != nil {
		t.Fatalf("run_result: %v", callErr)
	}
	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("run_result: expected map, got %T", value)
	}
	if result["ok"] != true {
		t.Fatalf("expected ok result, got %#v", result["ok"])
	}
	if result["exit_code"] != 0 {
		t.Fatalf("expected exit_code 0, got %#v", result["exit_code"])
	}
}

// TestProcess_RunIn_Good runs a command in a specific directory.
func TestProcess_RunIn_Good(t *core.T) {
	caller := newProcessInterpreter(t)
	dir := t.TempDir()

	value, callErr := caller.Call("core.process", "run_in", dir, "echo", "located")
	if callErr != nil {
		t.Fatalf("run_in: %v", callErr)
	}
	if value != "located\n" {
		t.Fatalf("unexpected run_in stdout %#v", value)
	}
}

// TestProcess_Exists_Good reports that a process Core is available.
func TestProcess_Exists_Good(t *core.T) {
	caller := newProcessInterpreter(t)

	value, callErr := caller.Call("core.process", "exists")
	if callErr != nil {
		t.Fatalf("exists: %v", callErr)
	}
	if _, ok := value.(bool); !ok {
		t.Fatalf("exists: expected bool, got %T", value)
	}
}

// TestProcess_RunWithEnv_Good passes an environment and reads it back via sh.
func TestProcess_RunWithEnv_Good(t *core.T) {
	caller := newProcessInterpreter(t)
	dir := t.TempDir()

	value, callErr := caller.Call("core.process", "run_with_env", dir,
		map[string]string{"COREPY_ENV": "set"},
		"sh", "-c", "printf %s \"$COREPY_ENV\"")
	if callErr != nil {
		t.Fatalf("run_with_env: %v", callErr)
	}
	if value != "set" {
		t.Fatalf("unexpected env-aware stdout %#v", value)
	}
}

// TestProcess_RunKeyword_Good applies a keyword timeout to a fast command.
func TestProcess_RunKeyword_Good(t *core.T) {
	caller := newProcessInterpreter(t)

	value, callErr := caller.Call("core.process", "run", "echo", "kw",
		runtime.KeywordArguments{"timeout": 30})
	if callErr != nil {
		t.Fatalf("run with keyword timeout: %v", callErr)
	}
	if value != "kw\n" {
		t.Fatalf("unexpected keyword stdout %#v", value)
	}
}

// TestProcess_RunKeyword_Ugly rejects an unexpected keyword argument.
func TestProcess_RunKeyword_Ugly(t *core.T) {
	caller := newProcessInterpreter(t)

	if _, callErr := caller.Call("core.process", "run", "echo", "x",
		runtime.KeywordArguments{"nope": true}); callErr == nil {
		t.Fatal("expected error for unexpected keyword argument")
	}
}

// TestProcess_Run_Bad reports a missing command argument.
func TestProcess_Run_Bad(t *core.T) {
	caller := newProcessInterpreter(t)

	if _, callErr := caller.Call("core.process", "run"); callErr == nil {
		t.Fatal("expected error for missing command argument")
	}
}

// TestProcess_Run_Ugly reports a failure for a non-existent command.
func TestProcess_Run_Ugly(t *core.T) {
	caller := newProcessInterpreter(t)

	if _, callErr := caller.Call("core.process", "run", "definitely-not-a-real-command-xyz"); callErr == nil {
		t.Fatal("expected error running a non-existent command")
	}
}
