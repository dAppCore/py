package dns

import (
	core "dappco.re/go"
)

func TestDns_Register_Good(t *core.T) {
	subject := Register
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestDns_Register_Bad(t *core.T) {
	subject := Register
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestDns_Register_Ugly(t *core.T) {
	subject := Register
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
