package scm

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newSCMInterpreter registers the scm module against a fresh bootstrap
// interpreter and returns the direct caller.
//
//	caller := newSCMInterpreter(t)
//	value, err := caller.Call("core.scm", "branch", repoDir)
func newSCMInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register scm module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// initRepo creates a git repository with a single committed file and returns
// its directory. The test is skipped when git is unavailable.
func initRepo(t *core.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		(*testing.T)(t).Skip("git is not available")
	}
	dir := t.TempDir()

	run := func(args ...string) {
		command := exec.Command("git", append([]string{"-C", dir}, args...)...)
		command.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, string(output))
		}
	}

	run("init", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("content"), 0o644); err != nil {
		t.Fatalf("write tracked file: %v", err)
	}
	run("add", "tracked.txt")
	run("commit", "-m", "initial")
	return dir
}

// TestSCM_ExistsRootBranch_Good inspects a real git repository.
func TestSCM_ExistsRootBranch_Good(t *core.T) {
	caller := newSCMInterpreter(t)
	dir := initRepo(t)

	exists, callErr := caller.Call("core.scm", "exists", dir)
	if callErr != nil {
		t.Fatalf("exists: %v", callErr)
	}
	if exists != true {
		t.Fatalf("expected a git work tree, got %#v", exists)
	}

	branch, callErr := caller.Call("core.scm", "branch", dir)
	if callErr != nil {
		t.Fatalf("branch: %v", callErr)
	}
	if branch != "main" {
		t.Fatalf("unexpected branch %#v", branch)
	}

	head, callErr := caller.Call("core.scm", "head", dir)
	if callErr != nil {
		t.Fatalf("head: %v", callErr)
	}
	if text, ok := head.(string); !ok || len(text) != 40 {
		t.Fatalf("head: expected a 40-char sha, got %#v", head)
	}
}

// TestSCM_StatusTracked_Good reports a clean status and tracked files.
func TestSCM_StatusTracked_Good(t *core.T) {
	caller := newSCMInterpreter(t)
	dir := initRepo(t)

	value, callErr := caller.Call("core.scm", "status", dir)
	if callErr != nil {
		t.Fatalf("status: %v", callErr)
	}
	status, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("status: expected map, got %T", value)
	}
	if status["clean"] != true {
		t.Fatalf("expected clean status, got %#v", status["clean"])
	}

	tracked, callErr := caller.Call("core.scm", "tracked_files", dir)
	if callErr != nil {
		t.Fatalf("tracked_files: %v", callErr)
	}
	files, ok := tracked.([]string)
	if !ok || len(files) != 1 || files[0] != "tracked.txt" {
		t.Fatalf("tracked_files: unexpected %#v", tracked)
	}
}

// TestSCM_Exists_Good reports false for a non-repository directory.
func TestSCM_Exists_Good(t *core.T) {
	caller := newSCMInterpreter(t)

	exists, callErr := caller.Call("core.scm", "exists", t.TempDir())
	if callErr != nil {
		t.Fatalf("exists: %v", callErr)
	}
	if exists != false {
		t.Fatalf("expected non-repository to report false, got %#v", exists)
	}
}

// TestSCM_Branch_Ugly rejects a non-string directory argument.
func TestSCM_Branch_Ugly(t *core.T) {
	caller := newSCMInterpreter(t)

	if _, callErr := caller.Call("core.scm", "branch", 123); callErr == nil {
		t.Fatal("expected error for non-string directory argument")
	}
}
