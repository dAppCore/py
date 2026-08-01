package bootstrap

import (
	"strings"

	core "dappco.re/go"
	"dappco.re/go/py/runtime/internal/contract"
)

// echoModule returns a module exposing a single round-trip echo function.
//
//	module := echoModule()
func echoModule() contract.Module {
	return contract.Module{
		Name:          "core",
		Documentation: "echo module",
		Functions: map[string]contract.Function{
			"echo": func(arguments ...any) (any, error) {
				if len(arguments) != 1 {
					return nil, core.E("core.echo", "expected one argument", nil)
				}
				return arguments[0], nil
			},
		},
	}
}

// TestInterpreter_RunImportEcho_Good imports and runs the echo function.
func TestInterpreter_RunImportEcho_Good(t *core.T) {
	interpreter := New()
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := interpreter.RegisterModule(echoModule()); err != nil {
		t.Fatalf("register: %v", err)
	}

	output, err := interpreter.Run("from core import echo\nprint(echo(\"hello\"))")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if strings.TrimSpace(output) != "hello" {
		t.Fatalf("unexpected output %q", output)
	}
}

// TestInterpreter_Call_Good invokes a registered function directly.
func TestInterpreter_Call_Good(t *core.T) {
	interpreter := New()
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := interpreter.RegisterModule(echoModule()); err != nil {
		t.Fatalf("register: %v", err)
	}

	value, err := interpreter.Call("core", "echo", "direct")
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if value != "direct" {
		t.Fatalf("unexpected call value %#v", value)
	}
}

// TestInterpreter_Modules_Good lists the registered module names.
func TestInterpreter_Modules_Good(t *core.T) {
	interpreter := New()
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := interpreter.RegisterModule(echoModule()); err != nil {
		t.Fatalf("register: %v", err)
	}

	names := interpreter.Modules()
	found := false
	for _, name := range names {
		if name == "core" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected core module in %#v", names)
	}
}

// TestInterpreter_Session_Good preserves namespace state across runs.
func TestInterpreter_Session_Good(t *core.T) {
	interpreter := New()
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := interpreter.RegisterModule(echoModule()); err != nil {
		t.Fatalf("register: %v", err)
	}

	session := interpreter.NewSession()
	if _, err := session.Run("greeting = echo"); err != nil {
		// echo is only available after import; ensure assignment of a literal works
		if _, err := session.Run("greeting = \"stored\""); err != nil {
			t.Fatalf("session assignment: %v", err)
		}
	}
	output, err := session.Run("print(greeting)")
	if err != nil {
		t.Fatalf("session print: %v", err)
	}
	if strings.TrimSpace(output) == "" {
		t.Fatal("expected preserved namespace value")
	}
}

// TestInterpreter_Call_Bad reports an unregistered module.
func TestInterpreter_Call_Bad(t *core.T) {
	interpreter := New()
	t.Cleanup(func() { _ = interpreter.Close() })

	if _, err := interpreter.Call("absent", "echo"); err == nil {
		t.Fatal("expected error for unregistered module")
	}
}

// TestInterpreter_Call_MissingFunction_Bad reports an unregistered function.
func TestInterpreter_Call_MissingFunction_Bad(t *core.T) {
	interpreter := New()
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := interpreter.RegisterModule(echoModule()); err != nil {
		t.Fatalf("register: %v", err)
	}

	if _, err := interpreter.Call("core", "absent"); err == nil {
		t.Fatal("expected error for unregistered function")
	}
}

// TestInterpreter_Run_Ugly reports a syntactically broken statement.
func TestInterpreter_Run_Ugly(t *core.T) {
	interpreter := New()
	t.Cleanup(func() { _ = interpreter.Close() })

	if _, err := interpreter.Run("print(echo(\"unclosed\")"); err == nil {
		t.Fatal("expected error for an unbalanced print statement")
	}
}
