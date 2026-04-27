package tier2_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dappco.re/go/py/runtime/tier2"
)

func TestRunner_RunSource_StreamsAndCaptures_Good(t *testing.T) {
	python := requirePython(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	runner := tier2.NewRunner(tier2.Options{
		Python: python,
		Environment: map[string]string{
			"COREPY_TIER2_TEST": "enabled",
		},
		Stdout: &stdout,
		Stderr: &stderr,
	})
	result, err := runner.RunSource(context.Background(), `
import os
import sys
print("out:" + os.environ["COREPY_TIER2_TEST"])
print("err:stream", file=sys.stderr)
`)
	if err != nil {
		t.Fatalf("run tier2 source: %v", err)
	}

	if !result.OK() {
		t.Fatalf("expected successful result, got %#v", result)
	}
	if strings.TrimSpace(result.Stdout) != "out:enabled" {
		t.Fatalf("unexpected captured stdout %q", result.Stdout)
	}
	if strings.TrimSpace(result.Stderr) != "err:stream" {
		t.Fatalf("unexpected captured stderr %q", result.Stderr)
	}
	if stdout.String() != result.Stdout {
		t.Fatalf("stdout stream/capture mismatch %q != %q", stdout.String(), result.Stdout)
	}
	if stderr.String() != result.Stderr {
		t.Fatalf("stderr stream/capture mismatch %q != %q", stderr.String(), result.Stderr)
	}
}

func TestRunner_RunFile_UsesWorkingDirectoryAndPythonPath_Good(t *testing.T) {
	python := requirePython(t)
	directory := t.TempDir()
	packageDirectory := filepath.Join(directory, "pkg")
	if err := os.MkdirAll(packageDirectory, 0755); err != nil {
		t.Fatalf("create package directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDirectory, "samplemod.py"), []byte(`VALUE = "from-pythonpath"`), 0600); err != nil {
		t.Fatalf("write module: %v", err)
	}
	script := filepath.Join(directory, "script.py")
	if err := os.WriteFile(script, []byte(`
from pathlib import Path
import samplemod
print(Path.cwd().name + ":" + samplemod.VALUE)
`), 0600); err != nil {
		t.Fatalf("write script: %v", err)
	}

	runner := tier2.NewRunner(tier2.Options{
		Python:           python,
		WorkingDirectory: directory,
		PythonPath:       []string{packageDirectory},
	})
	result, err := runner.RunFile(context.Background(), script)
	if err != nil {
		t.Fatalf("run tier2 file: %v", err)
	}
	if strings.TrimSpace(result.Stdout) != filepath.Base(directory)+":from-pythonpath" {
		t.Fatalf("unexpected stdout %q", result.Stdout)
	}
}

func TestRunner_RunSource_NonZeroExit_Bad(t *testing.T) {
	python := requirePython(t)
	runner := tier2.NewRunner(tier2.Options{Python: python})

	result, err := runner.RunSource(context.Background(), `
import sys
print("about to fail", file=sys.stderr)
sys.exit(7)
`)
	if err == nil {
		t.Fatal("expected nonzero exit to fail")
	}

	var exitErr tier2.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if result.ExitCode != 7 || exitErr.Result.ExitCode != 7 {
		t.Fatalf("unexpected exit result %#v / %#v", result, exitErr.Result)
	}
	if !strings.Contains(err.Error(), "about to fail") {
		t.Fatalf("expected stderr in error, got %v", err)
	}
}

func TestRunner_RunSource_Timeout_Ugly(t *testing.T) {
	python := requirePython(t)
	runner := tier2.NewRunner(tier2.Options{
		Python:  python,
		Timeout: 100 * time.Millisecond,
	})

	result, err := runner.RunSource(context.Background(), `
import time
time.sleep(5)
`)
	if err == nil {
		t.Fatal("expected timeout to fail")
	}
	var exitErr tier2.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if !result.TimedOut || !exitErr.Result.TimedOut {
		t.Fatalf("expected timed out result, got %#v / %#v", result, exitErr.Result)
	}
}

func TestResolvePython_Missing_Bad(t *testing.T) {
	_, err := tier2.ResolvePython("definitely-not-a-corepy-python")
	if err == nil {
		t.Fatal("expected missing python to fail")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unexpected error %v", err)
	}
}

func requirePython(t *testing.T) string {
	t.Helper()

	python, err := tier2.ResolvePython("")
	if err != nil {
		t.Skipf("Tier 2 CPython is not available in this environment: %v", err)
	}
	return python
}
