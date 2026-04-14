package runtime_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dappco.re/go/py/bindings/register"
	corepyruntime "dappco.re/go/py/runtime"
)

func TestInterpreter_Run_EchoRoundTrip_Good(t *testing.T) {
	interpreter := corepyruntime.New()
	if err := register.DefaultModules(interpreter); err != nil {
		t.Fatalf("register modules: %v", err)
	}

	output, err := interpreter.Run(`
from core import echo
print(echo("hello"))
`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}
	if strings.TrimSpace(output) != "hello" {
		t.Fatalf("unexpected output %q", output)
	}
}

func TestInterpreter_Run_SubmoduleImport_Good(t *testing.T) {
	interpreter := corepyruntime.New()
	if err := register.DefaultModules(interpreter); err != nil {
		t.Fatalf("register modules: %v", err)
	}

	directory := t.TempDir()
	filename := filepath.Join(directory, "sample.json")
	if err := os.WriteFile(filename, []byte(`{"name":"corepy"}`), 0600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	script := fmt.Sprintf(`
from core import fs, json
data = fs.read_file(%q)
print(json.dumps(json.loads(data)))
`, filename)

	output, err := interpreter.Run(script)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}
	if strings.TrimSpace(output) != `{"name":"corepy"}` {
		t.Fatalf("unexpected output %q", output)
	}
}

func TestInterpreter_Call_Primitives_Good(t *testing.T) {
	interpreter := corepyruntime.New()
	if err := register.DefaultModules(interpreter); err != nil {
		t.Fatalf("register modules: %v", err)
	}

	optionsHandle, err := interpreter.Call("core.options", "new", map[string]any{
		"name": "corepy",
		"port": 8080,
	})
	if err != nil {
		t.Fatalf("create options: %v", err)
	}

	name, err := interpreter.Call("core.options", "string", optionsHandle, "name")
	if err != nil {
		t.Fatalf("options string: %v", err)
	}
	if name != "corepy" {
		t.Fatalf("unexpected option name %#v", name)
	}

	configHandle, err := interpreter.Call("core.config", "new")
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	if _, err := interpreter.Call("core.config", "set", configHandle, "debug", true); err != nil {
		t.Fatalf("set config: %v", err)
	}
	debugEnabled, err := interpreter.Call("core.config", "bool", configHandle, "debug")
	if err != nil {
		t.Fatalf("config bool: %v", err)
	}
	if debugEnabled != true {
		t.Fatalf("unexpected debug flag %#v", debugEnabled)
	}

	dataHandle, err := interpreter.Call("core.data", "new")
	if err != nil {
		t.Fatalf("create data registry: %v", err)
	}
	fixtureDirectory := filepath.Join(t.TempDir(), "fixtures")
	if err := os.MkdirAll(fixtureDirectory, 0755); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fixtureDirectory, "note.txt"), []byte("hello from data"), 0600); err != nil {
		t.Fatalf("write data fixture: %v", err)
	}
	if _, err := interpreter.Call("core.data", "mount_path", dataHandle, "fixtures", fixtureDirectory); err != nil {
		t.Fatalf("mount data path: %v", err)
	}
	content, err := interpreter.Call("core.data", "read_string", dataHandle, "fixtures/note.txt")
	if err != nil {
		t.Fatalf("read mounted data: %v", err)
	}
	if content != "hello from data" {
		t.Fatalf("unexpected mounted content %#v", content)
	}

	serviceHandle, err := interpreter.Call("core.service", "new", "corepy")
	if err != nil {
		t.Fatalf("create service core: %v", err)
	}
	if _, err := interpreter.Call("core.service", "register", serviceHandle, "brain"); err != nil {
		t.Fatalf("register service: %v", err)
	}
	serviceNames, err := interpreter.Call("core.service", "names", serviceHandle)
	if err != nil {
		t.Fatalf("list services: %v", err)
	}
	names := serviceNames.([]string)
	if len(names) == 0 || names[0] != "cli" {
		t.Fatalf("expected built-in cli service first, got %#v", names)
	}
	if names[len(names)-1] != "brain" {
		t.Fatalf("expected registered service in names, got %#v", names)
	}
}
