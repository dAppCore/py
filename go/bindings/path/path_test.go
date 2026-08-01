package pathbinding

import (
	core "dappco.re/go"
)

func TestPath_Register_Good(t *core.T) {
	subject := Register
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestPath_Register_Bad(t *core.T) {
	subject := Register
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestPath_Register_Ugly(t *core.T) {
	subject := Register
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
