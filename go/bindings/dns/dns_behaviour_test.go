package dns

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newDNSInterpreter registers the dns module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newDNSInterpreter(t)
//	value, err := caller.Call("core.dns", "lookup_host", "localhost")
func newDNSInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register dns module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestDNS_LookupHostLocalhost_Good resolves localhost without the network.
func TestDNS_LookupHostLocalhost_Good(t *core.T) {
	caller := newDNSInterpreter(t)

	value, callErr := caller.Call("core.dns", "lookup_host", "localhost")
	if callErr != nil {
		t.Fatalf("lookup_host: %v", callErr)
	}
	hosts, ok := value.([]string)
	if !ok || len(hosts) == 0 {
		t.Fatalf("lookup_host: unexpected %#v", value)
	}
}

// TestDNS_LookupIPLocalhost_Good resolves localhost to a loopback address.
func TestDNS_LookupIPLocalhost_Good(t *core.T) {
	caller := newDNSInterpreter(t)

	value, callErr := caller.Call("core.dns", "lookup_ip", "localhost")
	if callErr != nil {
		t.Fatalf("lookup_ip: %v", callErr)
	}
	addresses, ok := value.([]string)
	if !ok || len(addresses) == 0 {
		t.Fatalf("lookup_ip: unexpected %#v", value)
	}
}

// TestDNS_LookupPort_Good resolves a well-known service port.
func TestDNS_LookupPort_Good(t *core.T) {
	caller := newDNSInterpreter(t)

	value, callErr := caller.Call("core.dns", "lookup_port", "tcp", "http")
	if callErr != nil {
		t.Fatalf("lookup_port: %v", callErr)
	}
	if value != 80 {
		t.Fatalf("lookup_port: expected 80, got %#v", value)
	}
}

// TestDNS_LookupHost_Bad reports a missing name argument.
func TestDNS_LookupHost_Bad(t *core.T) {
	caller := newDNSInterpreter(t)

	if _, callErr := caller.Call("core.dns", "lookup_host"); callErr == nil {
		t.Fatal("expected error for missing name argument")
	}
}

// TestDNS_LookupPort_Ugly reports an unknown service name.
func TestDNS_LookupPort_Ugly(t *core.T) {
	caller := newDNSInterpreter(t)

	if _, callErr := caller.Call("core.dns", "lookup_port", "tcp", "definitely-not-a-service"); callErr == nil {
		t.Fatal("expected error for unknown service")
	}
}
