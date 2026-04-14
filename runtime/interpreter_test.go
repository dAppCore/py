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

func TestInterpreter_Run_ImportModuleForms_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	filename := filepath.Join(t.TempDir(), "sample.txt")
	if err := os.WriteFile(filename, []byte("hello"), 0600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	script := fmt.Sprintf(`
import core
import core.fs as filesystem
print(core.echo("hello"))
print(filesystem.read_file(%q))
`, filename)

	output, err := interpreter.Run(script)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if !reflect.DeepEqual(lines, []string{"hello", "hello"}) {
		t.Fatalf("unexpected output lines %#v", lines)
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

func TestInterpreter_Run_ConfigEnvFallback_Good(t *testing.T) {
	t.Setenv("DATABASE_HOST", "db.internal")
	t.Setenv("PORT", "8080")
	t.Setenv("DEBUG", "true")

	interpreter := newTestInterpreter(t)

	output, err := interpreter.Run(`
from core import config
cfg = config.new()
print(config.get(cfg, "database.host"))
print(config.int(cfg, "port"))
print(config.bool(cfg, "debug"))
config.set(cfg, "database.host", "override.internal")
config.set(cfg, "port", 9000)
config.set(cfg, "debug", False)
print(config.string(cfg, "database.host"))
print(config.int(cfg, "port"))
print(config.bool(cfg, "debug"))
`)
	if err != nil {
		t.Fatalf("run config env fallback: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if !reflect.DeepEqual(lines, []string{"db.internal", "8080", "True", "override.internal", "9000", "False"}) {
		t.Fatalf("unexpected output lines %#v", lines)
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

func TestInterpreter_Run_ListAndDictTypeMapping_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	output, err := interpreter.Run(`
from core import options, math
values = [3, 1, 2]
items = {"name": "corepy", "port": 8080}
handle = options.new(items)
print(options.string(handle, "name"))
print(math.mean(values))
print(math.sort(values))
`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if !reflect.DeepEqual(lines, []string{"corepy", "2", "[1, 2, 3]"}) {
		t.Fatalf("unexpected output lines %#v", lines)
	}
}

func TestInterpreter_Run_MathExample_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	script, err := os.ReadFile(filepath.Join("..", "examples", "math.py"))
	if err != nil {
		t.Fatalf("read math example: %v", err)
	}

	output, err := interpreter.Run(string(script))
	if err != nil {
		t.Fatalf("run math example: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 output lines, got %#v", lines)
	}
	if lines[0] != "0.5" {
		t.Fatalf("unexpected mean output %q", lines[0])
	}
	if !strings.Contains(lines[1], `"index": 1`) || !strings.Contains(lines[1], `"index": 0`) {
		t.Fatalf("unexpected nearest-neighbour output %q", lines[1])
	}
}

func TestInterpreter_Run_RFCMathImports_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	output, err := interpreter.Run(`
from core.math import kdtree, knn, mean, stdev
embeddings = [[1.0, 0.0], [0.0, 1.0], [0.8, 0.2]]
tree = kdtree.build(embeddings, metric="cosine")
print(mean([1, 2, 3]))
print(stdev([1, 2, 3]))
print(tree.nearest([1.0, 0.0], k=2))
print(knn.search(embeddings, [1.0, 0.0], k=2, metric="cosine"))
`)
	if err != nil {
		t.Fatalf("run RFC math imports: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 output lines, got %#v", lines)
	}
	if lines[0] != "2" {
		t.Fatalf("unexpected mean output %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "0.81649") {
		t.Fatalf("unexpected stdev output %q", lines[1])
	}
	if !strings.Contains(lines[2], `"index": 0`) || !strings.Contains(lines[2], `"index": 2`) {
		t.Fatalf("unexpected tree nearest output %q", lines[2])
	}
	if !strings.Contains(lines[3], `"index": 0`) || !strings.Contains(lines[3], `"index": 2`) {
		t.Fatalf("unexpected knn output %q", lines[3])
	}
}

func TestInterpreter_Run_DirectNestedMathImports_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	output, err := interpreter.Run(`
import core.math.kdtree as kdtree
from core.math.knn import search
tree = kdtree.build([[0.0, 0.0], [1.0, 1.0], [3.0, 3.0]], metric="euclidean")
print(tree.nearest([0.8, 0.8], k=2))
print(search([[1.0, 0.0], [0.0, 1.0], [0.8, 0.2]], [1.0, 0.0], k=2, metric="cosine"))
`)
	if err != nil {
		t.Fatalf("run nested math imports: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 output lines, got %#v", lines)
	}
	if !strings.Contains(lines[0], `"index": 1`) || !strings.Contains(lines[0], `"index": 0`) {
		t.Fatalf("unexpected kdtree output %q", lines[0])
	}
	if !strings.Contains(lines[1], `"index": 0`) || !strings.Contains(lines[1], `"index": 2`) {
		t.Fatalf("unexpected knn output %q", lines[1])
	}
}

func TestInterpreter_Run_MathSignalImports_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	output, err := interpreter.Run(`
from core import math
from core.math import signal
values = [1, 3, 6, 10]
print(math.moving_average(values, window=2))
print(signal.difference(values))
print(math.signal.difference(values, lag=2))
`)
	if err != nil {
		t.Fatalf("run math signal imports: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if !reflect.DeepEqual(lines, []string{"[1, 2, 4.5, 8]", "[2, 3, 4]", "[5, 7]"}) {
		t.Fatalf("unexpected signal output lines %#v", lines)
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

func TestInterpreter_Call_MathPrimitives_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	sortedValues, err := interpreter.Call("core.math", "sort", []any{3, 1, 2})
	if err != nil {
		t.Fatalf("sort values: %v", err)
	}
	if !reflect.DeepEqual(sortedValues, []any{1, 2, 3}) {
		t.Fatalf("unexpected sorted values %#v", sortedValues)
	}

	index, err := interpreter.Call("core.math", "binary_search", []any{1, 2, 3}, 2)
	if err != nil {
		t.Fatalf("binary search: %v", err)
	}
	if index != 1 {
		t.Fatalf("unexpected binary search index %#v", index)
	}

	tree, err := interpreter.Call("core.math.kdtree", "build", []any{
		[]any{0.0, 0.0},
		[]any{1.0, 1.0},
		[]any{3.0, 3.0},
	}, corepyruntime.KeywordArguments{"metric": "euclidean"})
	if err != nil {
		t.Fatalf("build kdtree: %v", err)
	}

	defaultNearest, err := interpreter.Call("core.math.kdtree", "nearest", tree, []any{0.8, 0.8})
	if err != nil {
		t.Fatalf("kdtree nearest default k: %v", err)
	}
	if len(defaultNearest.([]map[string]any)) != 1 {
		t.Fatalf("expected default nearest search to return one neighbour, got %#v", defaultNearest)
	}

	nearest, err := interpreter.Call("core.math.kdtree", "nearest", tree, []any{0.8, 0.8}, corepyruntime.KeywordArguments{"k": 2})
	if err != nil {
		t.Fatalf("kdtree nearest: %v", err)
	}

	neighbors := nearest.([]map[string]any)
	if len(neighbors) != 2 {
		t.Fatalf("expected 2 neighbours, got %#v", neighbors)
	}
	if neighbors[0]["index"] != 1 || neighbors[1]["index"] != 0 {
		t.Fatalf("unexpected neighbour order %#v", neighbors)
	}

	cosine, err := interpreter.Call("core.math.knn", "search", []any{
		[]any{1.0, 0.0},
		[]any{0.0, 1.0},
		[]any{0.8, 0.2},
	}, []any{1.0, 0.0}, corepyruntime.KeywordArguments{"k": 2, "metric": "cosine"})
	if err != nil {
		t.Fatalf("knn search: %v", err)
	}

	cosineNeighbors := cosine.([]map[string]any)
	if cosineNeighbors[0]["index"] != 0 || cosineNeighbors[1]["index"] != 2 {
		t.Fatalf("unexpected cosine neighbour order %#v", cosineNeighbors)
	}

	smoothed, err := interpreter.Call("core.math", "moving_average", []any{1, 3, 6, 10}, corepyruntime.KeywordArguments{"window": 2})
	if err != nil {
		t.Fatalf("moving average: %v", err)
	}
	if !reflect.DeepEqual(smoothed, []float64{1, 2, 4.5, 8}) {
		t.Fatalf("unexpected smoothed values %#v", smoothed)
	}

	delta, err := interpreter.Call("core.math.signal", "difference", []any{1, 3, 6, 10}, corepyruntime.KeywordArguments{"lag": 2})
	if err != nil {
		t.Fatalf("difference: %v", err)
	}
	if !reflect.DeepEqual(delta, []float64{5, 7}) {
		t.Fatalf("unexpected difference values %#v", delta)
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
