package service

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newServiceInterpreter registers the service module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newServiceInterpreter(t)
//	handle, err := caller.Call("core.service", "new", "corepy")
func newServiceInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register service module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestService_RegisterNames_Good registers a service and lists the names.
func TestService_RegisterNames_Good(t *core.T) {
	caller := newServiceInterpreter(t)

	handle, callErr := caller.Call("core.service", "new", "corepy")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.service", "register", handle, "brain"); callErr != nil {
		t.Fatalf("register: %v", callErr)
	}

	value, callErr := caller.Call("core.service", "names", handle)
	if callErr != nil {
		t.Fatalf("names: %v", callErr)
	}
	names, ok := value.([]string)
	if !ok {
		t.Fatalf("names: expected []string, got %T", value)
	}
	found := false
	for _, name := range names {
		if name == "brain" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected registered service in names, got %#v", names)
	}
}

// TestService_StartStopAll_Good runs the startup and shutdown lifecycle.
func TestService_StartStopAll_Good(t *core.T) {
	caller := newServiceInterpreter(t)

	handle, callErr := caller.Call("core.service", "new")
	if callErr != nil {
		t.Fatalf("new default: %v", callErr)
	}

	started, callErr := caller.Call("core.service", "start_all", handle)
	if callErr != nil {
		t.Fatalf("start_all: %v", callErr)
	}
	if started != true {
		t.Fatalf("start_all: unexpected %#v", started)
	}

	stopped, callErr := caller.Call("core.service", "stop_all", handle)
	if callErr != nil {
		t.Fatalf("stop_all: %v", callErr)
	}
	if stopped != true {
		t.Fatalf("stop_all: unexpected %#v", stopped)
	}
}

// TestService_RegisterMissingName_Bad reports a missing service name argument.
func TestService_RegisterMissingName_Bad(t *core.T) {
	caller := newServiceInterpreter(t)

	handle, callErr := caller.Call("core.service", "new", "corepy")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.service", "register", handle); callErr == nil {
		t.Fatal("expected error for missing service name")
	}
}

// TestService_Names_Ugly rejects a non-Core handle argument.
func TestService_Names_Ugly(t *core.T) {
	caller := newServiceInterpreter(t)

	if _, callErr := caller.Call("core.service", "names", "not-a-core"); callErr == nil {
		t.Fatal("expected error for non-Core handle")
	}
}
