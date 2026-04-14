package runtime_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"strings"
	"testing"

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/register"
	corepyruntime "dappco.re/go/py/runtime"
)

type lifecycleService struct {
	started bool
	stopped bool
}

func (service *lifecycleService) OnStartup(ctx context.Context) core.Result {
	service.started = true
	return core.Result{OK: ctx.Err() == nil}
}

func (service *lifecycleService) OnShutdown(ctx context.Context) core.Result {
	service.stopped = true
	return core.Result{OK: ctx.Err() == nil}
}

func newTestInterpreter(t *testing.T) *corepyruntime.Interpreter {
	t.Helper()

	interpreter := corepyruntime.New()
	if err := register.DefaultModules(interpreter); err != nil {
		t.Fatalf("register modules: %v", err)
	}
	return interpreter
}

func TestInterpreter_Run_EchoRoundTrip_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

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
	interpreter := newTestInterpreter(t)

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

func TestInterpreter_Run_MediumImport_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	output, err := interpreter.Run(`
from core import medium
buffer = medium.memory("hello")
medium.write_text(buffer, "updated")
print(medium.read_text(buffer))
`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}
	if strings.TrimSpace(output) != "updated" {
		t.Fatalf("unexpected output %q", output)
	}
}

func TestInterpreter_Run_ProcessImport_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("find go binary: %v", err)
	}

	script := fmt.Sprintf(`
from core import process
print(process.run(%q, "env", "GOOS"))
`, goBinary)

	output, err := interpreter.Run(script)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}
	if strings.TrimSpace(output) != goruntime.GOOS {
		t.Fatalf("unexpected output %q", output)
	}
}

func TestInterpreter_Run_PathAndStringsImport_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	output, err := interpreter.Run(`
from core import path, strings
location = path.join("deploy", "to", "homelab")
print(strings.concat(location, ":", path.base(location)))
`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}
	if strings.TrimSpace(output) != "deploy/to/homelab:homelab" {
		t.Fatalf("unexpected output %q", output)
	}
}

func TestInterpreter_Call_Primitives_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

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

func TestInterpreter_Call_FilesystemAndMediumBytes_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	filename := filepath.Join(t.TempDir(), "payload.bin")
	if _, err := interpreter.Call("core.fs", "write_bytes", filename, []byte("corepy")); err != nil {
		t.Fatalf("write bytes: %v", err)
	}

	content, err := interpreter.Call("core.fs", "read_bytes", filename)
	if err != nil {
		t.Fatalf("read bytes: %v", err)
	}
	if string(content.([]byte)) != "corepy" {
		t.Fatalf("unexpected byte payload %#v", content)
	}

	mediumHandle, err := interpreter.Call("core.medium", "from_path", filename)
	if err != nil {
		t.Fatalf("create file-backed medium: %v", err)
	}
	if _, err := interpreter.Call("core.medium", "write_bytes", mediumHandle, []byte("updated")); err != nil {
		t.Fatalf("write medium bytes: %v", err)
	}
	mediumContent, err := interpreter.Call("core.medium", "read_bytes", mediumHandle)
	if err != nil {
		t.Fatalf("read medium bytes: %v", err)
	}
	if string(mediumContent.([]byte)) != "updated" {
		t.Fatalf("unexpected medium payload %#v", mediumContent)
	}

	memoryHandle, err := interpreter.Call("core.medium", "memory", "hello")
	if err != nil {
		t.Fatalf("create memory medium: %v", err)
	}
	if _, err := interpreter.Call("core.medium", "write_text", memoryHandle, "world"); err != nil {
		t.Fatalf("write memory medium: %v", err)
	}
	text, err := interpreter.Call("core.medium", "read_text", memoryHandle)
	if err != nil {
		t.Fatalf("read memory medium: %v", err)
	}
	if text != "world" {
		t.Fatalf("unexpected memory medium text %#v", text)
	}
}

func TestInterpreter_Call_ProcessHelpers_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("find go binary: %v", err)
	}

	output, err := interpreter.Call("core.process", "run", goBinary, "env", "GOOS")
	if err != nil {
		t.Fatalf("process run: %v", err)
	}
	if strings.TrimSpace(output.(string)) != goruntime.GOOS {
		t.Fatalf("unexpected process output %#v", output)
	}

	inDirectoryOutput, err := interpreter.Call("core.process", "run_in", "/home/claude/Code/core/py", goBinary, "env", "GOMOD")
	if err != nil {
		t.Fatalf("process run_in: %v", err)
	}
	if !strings.HasSuffix(strings.TrimSpace(inDirectoryOutput.(string)), "/home/claude/Code/core/py/go.mod") {
		t.Fatalf("unexpected process run_in output %#v", inDirectoryOutput)
	}

	envOutput, err := interpreter.Call("core.process", "run_with_env", "/home/claude/Code/core/py", map[string]string{"GOWORK": "off"}, goBinary, "env", "GOWORK")
	if err != nil {
		t.Fatalf("process run_with_env: %v", err)
	}
	if strings.TrimSpace(envOutput.(string)) != "off" {
		t.Fatalf("unexpected process run_with_env output %#v", envOutput)
	}

	exists, err := interpreter.Call("core.process", "exists")
	if err != nil {
		t.Fatalf("process exists: %v", err)
	}
	if exists != true {
		t.Fatalf("expected process capability to exist, got %#v", exists)
	}
}

func TestInterpreter_Call_DataExtract_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	fixtureDirectory := filepath.Join(t.TempDir(), "fixtures")
	templateDirectory := filepath.Join(fixtureDirectory, "templates")
	if err := os.MkdirAll(templateDirectory, 0755); err != nil {
		t.Fatalf("create fixture directories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fixtureDirectory, "note.txt"), []byte("hello from data"), 0600); err != nil {
		t.Fatalf("write data note: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templateDirectory, "greeting.txt.tmpl"), []byte("hello {{.Name}}"), 0600); err != nil {
		t.Fatalf("write template file: %v", err)
	}

	dataHandle, err := interpreter.Call("core.data", "new")
	if err != nil {
		t.Fatalf("create data registry: %v", err)
	}
	if _, err := interpreter.Call("core.data", "mount", dataHandle, "fixtures", fixtureDirectory); err != nil {
		t.Fatalf("mount data path: %v", err)
	}

	fileContent, err := interpreter.Call("core.data", "read_file", dataHandle, "fixtures/note.txt")
	if err != nil {
		t.Fatalf("read data file: %v", err)
	}
	if string(fileContent.([]byte)) != "hello from data" {
		t.Fatalf("unexpected mounted bytes %#v", fileContent)
	}

	listed, err := interpreter.Call("core.data", "list", dataHandle, "fixtures")
	if err != nil {
		t.Fatalf("list mounted data: %v", err)
	}
	if !strings.Contains(strings.Join(listed.([]string), ","), "note.txt") {
		t.Fatalf("expected note.txt in mounted list, got %#v", listed)
	}

	targetDirectory := filepath.Join(t.TempDir(), "workspace")
	if _, err := interpreter.Call("core.data", "extract", dataHandle, "fixtures/templates", targetDirectory, map[string]string{"Name": "corepy"}); err != nil {
		t.Fatalf("extract mounted data: %v", err)
	}
	extracted, err := os.ReadFile(filepath.Join(targetDirectory, "greeting.txt"))
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(extracted) != "hello corepy" {
		t.Fatalf("unexpected extracted content %q", extracted)
	}
}

func TestInterpreter_Call_ServiceLifecycle_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	serviceHandle, err := interpreter.Call("core.service", "new", "corepy")
	if err != nil {
		t.Fatalf("create service core: %v", err)
	}

	runner := &lifecycleService{}
	if _, err := interpreter.Call("core.service", "register", serviceHandle, "runner", runner); err != nil {
		t.Fatalf("register lifecycle service: %v", err)
	}

	serviceValue, err := interpreter.Call("core.service", "get", serviceHandle, "runner")
	if err != nil {
		t.Fatalf("get service: %v", err)
	}
	if serviceValue != runner {
		t.Fatalf("unexpected service instance %#v", serviceValue)
	}

	if _, err := interpreter.Call("core.service", "start_all", serviceHandle); err != nil {
		t.Fatalf("start services: %v", err)
	}
	if !runner.started {
		t.Fatal("expected lifecycle service to start")
	}

	if _, err := interpreter.Call("core.service", "stop_all", serviceHandle); err != nil {
		t.Fatalf("stop services: %v", err)
	}
	if !runner.stopped {
		t.Fatal("expected lifecycle service to stop")
	}
}

func TestInterpreter_Call_ErrorHelpers_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	issue, err := interpreter.Call("core.err", "e", "core.save", "write failed", nil, "WRITE_FAIL")
	if err != nil {
		t.Fatalf("create structured error: %v", err)
	}

	code, err := interpreter.Call("core.err", "error_code", issue)
	if err != nil {
		t.Fatalf("read error code: %v", err)
	}
	if code != "WRITE_FAIL" {
		t.Fatalf("unexpected error code %#v", code)
	}

	wrapped, err := interpreter.Call("core.err", "wrap", issue, "core.deploy", "deploy failed", "DEPLOY_FAIL")
	if err != nil {
		t.Fatalf("wrap structured error: %v", err)
	}
	root, err := interpreter.Call("core.err", "root", wrapped)
	if err != nil {
		t.Fatalf("read root error: %v", err)
	}
	if !errors.Is(wrapped.(error), root.(error)) {
		t.Fatalf("expected root error to be part of the wrapped chain, got %#v", root)
	}

	nilWrapped, err := interpreter.Call("core.err", "wrap", nil, "core.deploy", "deploy failed")
	if err != nil {
		t.Fatalf("wrap nil error: %v", err)
	}
	if nilWrapped != nil {
		t.Fatalf("expected nil wrapped error, got %#v", nilWrapped)
	}
}

func TestInterpreter_Call_PathAndStringHelpers_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	joined, err := interpreter.Call("core.path", "join", "deploy", "to", "homelab")
	if err != nil {
		t.Fatalf("path join: %v", err)
	}
	if joined != "deploy/to/homelab" {
		t.Fatalf("unexpected joined path %#v", joined)
	}

	baseName, err := interpreter.Call("core.path", "base", "/tmp/corepy/config.json")
	if err != nil {
		t.Fatalf("path base: %v", err)
	}
	if baseName != "config.json" {
		t.Fatalf("unexpected base name %#v", baseName)
	}

	cleaned, err := interpreter.Call("core.path", "clean", "deploy//to/../from")
	if err != nil {
		t.Fatalf("path clean: %v", err)
	}
	if cleaned != "deploy/from" {
		t.Fatalf("unexpected cleaned path %#v", cleaned)
	}

	contains, err := interpreter.Call("core.strings", "contains", "hello world", "world")
	if err != nil {
		t.Fatalf("strings contains: %v", err)
	}
	if contains != true {
		t.Fatalf("expected contains to be true, got %#v", contains)
	}

	parts, err := interpreter.Call("core.strings", "split_n", "key=value=extra", "=", 2)
	if err != nil {
		t.Fatalf("strings split_n: %v", err)
	}
	if !reflect.DeepEqual(parts, []string{"key", "value=extra"}) {
		t.Fatalf("unexpected split parts %#v", parts)
	}
}
