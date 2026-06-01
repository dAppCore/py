package contract

import (
	core "dappco.re/go"
)

// TestTypes_UnsupportedImportError_Named_Good includes the module name.
func TestTypes_UnsupportedImportError_Named_Good(t *core.T) {
	err := UnsupportedImportError{Module: "numpy"}
	if err.Error() != "unsupported import numpy" {
		t.Fatalf("unexpected message %q", err.Error())
	}
}

// TestTypes_UnsupportedImportError_Empty_Good handles a blank module name.
func TestTypes_UnsupportedImportError_Empty_Good(t *core.T) {
	err := UnsupportedImportError{}
	if err.Error() != "unsupported import" {
		t.Fatalf("unexpected empty-module message %q", err.Error())
	}
}

// TestTypes_UnsupportedImportError_AsError_Good confirms the value satisfies
// the error interface.
func TestTypes_UnsupportedImportError_AsError_Good(t *core.T) {
	var err error = UnsupportedImportError{Module: "pandas"}
	if err == nil || err.Error() == "" {
		t.Fatal("expected a non-empty error value")
	}
}
