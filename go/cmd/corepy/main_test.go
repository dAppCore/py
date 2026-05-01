package main

import (
	"bytes"
	"strings"
	"testing"

	"dappco.re/go/py/runtime/tier2"
)

func TestRun_RunExpression_Good(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runWithIO([]string{"run", "-e", `from core import echo; print(echo("cli"))`}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run expression: %v stderr=%q", err, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "cli" {
		t.Fatalf("unexpected stdout %q", stdout.String())
	}
}

func TestRun_RunExpressionTier2Fallback_Good(t *testing.T) {
	requireTier2Python(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runWithIO([]string{"run", "-tier2-fallback", "-e", `import sys; from core import echo; print(echo("tier2"), sys.version_info[0])`}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run expression with tier2 fallback: %v stderr=%q", err, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "tier2 3" {
		t.Fatalf("unexpected stdout %q", stdout.String())
	}
}

func TestRun_RunExpressionNoFallback_Bad(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runWithIO([]string{"run", "-e", `import sys; print("tier2")`}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected unsupported import without fallback")
	}
	if !strings.Contains(err.Error(), "unsupported import sys") {
		t.Fatalf("unexpected error %v", err)
	}
	if stdout.String() != "" {
		t.Fatalf("expected no stdout, got %q", stdout.String())
	}
}

func TestRun_ReplPreservesNamespace_Good(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	stdin := strings.NewReader("from core import echo\nmessage = echo(\"repl\")\nprint(message)\n:quit\n")

	err := runWithStreams([]string{"repl"}, stdin, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run repl: %v stderr=%q", err, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "repl" {
		t.Fatalf("unexpected stdout %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("unexpected stderr %q", stderr.String())
	}
}

func TestRun_Modules_Good(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runWithIO([]string{"modules"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("list modules: %v stderr=%q", err, stderr.String())
	}
	output := stdout.String()
	for _, expected := range []string{"core", "core.fs", "core.process", "core.math.kdtree"} {
		if !strings.Contains(output, expected+"\n") {
			t.Fatalf("expected %s in module list, got %q", expected, output)
		}
	}
}

func TestRun_Tier2Which_Good(t *testing.T) {
	requireTier2Python(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runWithIO([]string{"tier2", "which"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("tier2 which: %v stderr=%q", err, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) == "" {
		t.Fatal("expected tier2 python path")
	}
}

func TestRun_Tier2RunExpression_Good(t *testing.T) {
	requireTier2Python(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runWithIO([]string{"tier2", "run", "-e", `from core import echo; print(echo("tier2-cli"))`}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("tier2 run: %v stderr=%q", err, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "tier2-cli" {
		t.Fatalf("unexpected stdout %q", stdout.String())
	}
}

func TestRun_Tier2RunFailure_Bad(t *testing.T) {
	requireTier2Python(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runWithIO([]string{"tier2", "run", "-e", `import sys; print("nope", file=sys.stderr); sys.exit(4)`}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected tier2 failure")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Fatalf("expected stderr in error, got %v", err)
	}
	if strings.TrimSpace(stderr.String()) != "nope" {
		t.Fatalf("expected streamed stderr, got %q", stderr.String())
	}
}

func requireTier2Python(t *testing.T) {
	t.Helper()

	if _, err := tier2.ResolvePython(""); err != nil {
		t.Skipf("Tier 2 CPython is not available in this environment: %v", err)
	}
}
