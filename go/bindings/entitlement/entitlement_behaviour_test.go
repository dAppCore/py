package entitlement

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newEntitlementInterpreter registers the entitlement module against a fresh
// bootstrap interpreter and returns the direct caller.
//
//	caller := newEntitlementInterpreter(t)
//	handle, err := caller.Call("core.entitlement", "new", true, false, 100, 80)
func newEntitlementInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register entitlement module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestEntitlement_UsagePercent_Good computes usage from limit and used.
func TestEntitlement_UsagePercent_Good(t *core.T) {
	caller := newEntitlementInterpreter(t)

	handle, callErr := caller.Call("core.entitlement", "new", true, false, 100, 80)
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	value, callErr := caller.Call("core.entitlement", "usage_percent", handle)
	if callErr != nil {
		t.Fatalf("usage_percent: %v", callErr)
	}
	percent, ok := value.(float64)
	if !ok || percent != 80 {
		t.Fatalf("unexpected usage percent %#v", value)
	}
}

// TestEntitlement_NearLimit_Good reports proximity to the limit.
func TestEntitlement_NearLimit_Good(t *core.T) {
	caller := newEntitlementInterpreter(t)

	handle, callErr := caller.Call("core.entitlement", "new", true, false, 100, 90)
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	near, callErr := caller.Call("core.entitlement", "near_limit", handle, 0.8)
	if callErr != nil {
		t.Fatalf("near_limit: %v", callErr)
	}
	if near != true {
		t.Fatalf("expected near limit, got %#v", near)
	}

	notNear, callErr := caller.Call("core.entitlement", "near_limit", handle, 0.95)
	if callErr != nil {
		t.Fatalf("near_limit high threshold: %v", callErr)
	}
	if notNear != false {
		t.Fatalf("expected not near limit, got %#v", notNear)
	}
}

// TestEntitlement_New_Bad rejects a non-bool allowed argument.
func TestEntitlement_New_Bad(t *core.T) {
	caller := newEntitlementInterpreter(t)

	if _, callErr := caller.Call("core.entitlement", "new", "yes"); callErr == nil {
		t.Fatal("expected error for non-bool allowed argument")
	}
}

// TestEntitlement_NearLimit_Ugly reports a missing threshold and a wrong handle.
func TestEntitlement_NearLimit_Ugly(t *core.T) {
	caller := newEntitlementInterpreter(t)

	handle, callErr := caller.Call("core.entitlement", "new", true)
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.entitlement", "near_limit", handle); callErr == nil {
		t.Fatal("expected error for missing threshold")
	}
	if _, callErr := caller.Call("core.entitlement", "usage_percent", "not-an-entitlement"); callErr == nil {
		t.Fatal("expected error for non-entitlement value")
	}
}
