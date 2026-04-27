package runtime_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExamples_Echo_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	script, err := os.ReadFile(filepath.Join("..", "examples", "echo.py"))
	if err != nil {
		t.Fatalf("read echo example: %v", err)
	}

	output, err := interpreter.Run(string(script))
	if err != nil {
		t.Fatalf("run echo example: %v", err)
	}
	if strings.TrimSpace(output) != "hello" {
		t.Fatalf("unexpected echo output %q", output)
	}
}

func TestExamples_PrimitivePipeline_Good(t *testing.T) {
	interpreter := newTestInterpreter(t)

	script, err := os.ReadFile(filepath.Join("..", "examples", "primitive_pipeline.py"))
	if err != nil {
		t.Fatalf("read primitive pipeline example: %v", err)
	}

	output, err := interpreter.Run(string(script))
	if err != nil {
		t.Fatalf("run primitive pipeline example: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two output lines, got %#v", lines)
	}
	if len(lines[0]) != 64 {
		t.Fatalf("expected sha256 digest, got %q", lines[0])
	}
	if lines[1] != "True" {
		t.Fatalf("expected cache hit output, got %q", lines[1])
	}
}
