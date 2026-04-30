//go:build gpython

package runtime_test

import (
	"os/exec"
	"reflect"
	goruntime "runtime"
	"strings"
	"testing"

	"dappco.re/go/py/bindings/register"
	corepyruntime "dappco.re/go/py/runtime"
)

func TestGPythonBackend_StdlibShadowRoundTrip_Good(t *testing.T) {
	interpreter, err := corepyruntime.New(corepyruntime.Options{Backend: corepyruntime.BackendGPython})
	if err != nil {
		t.Fatalf("create gpython backend: %v", err)
	}
	defer interpreter.Close()

	if err := register.DefaultModules(interpreter); err != nil {
		t.Fatalf("register gpython backend modules: %v", err)
	}

	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("find go binary: %v", err)
	}

	output, err := interpreter.Run(`
import base64
import hashlib
import json
import os
import subprocess
payload = json.loads('{"name":"corepy"}')
print(json.dumps(payload))
print(os.path.basename(os.path.join("tmp", "corepy.json")))
print(base64.b64encode("hello"))
digest = hashlib.sha256("hello")
print(digest.hexdigest())
print(subprocess.check_output(["` + goBinary + `", "env", "GOOS"]))
`)
	if err != nil {
		t.Fatalf("run gpython stdlib shadow script: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	expected := []string{
		`{"name":"corepy"}`,
		"corepy.json",
		"aGVsbG8=",
		"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		goruntime.GOOS,
	}
	if !reflect.DeepEqual(lines, expected) {
		t.Fatalf("unexpected gpython stdlib shadow output %#v", lines)
	}
}
