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
