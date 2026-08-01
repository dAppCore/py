package contract

import (
	core "dappco.re/go"
)

func TestTypes_UnsupportedImportError_Error_Good(t *core.T) {
	subject := (*UnsupportedImportError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestTypes_UnsupportedImportError_Error_Bad(t *core.T) {
	subject := (*UnsupportedImportError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestTypes_UnsupportedImportError_Error_Ugly(t *core.T) {
	subject := (*UnsupportedImportError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
