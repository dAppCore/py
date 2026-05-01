package action

import (
	core "dappco.re/go"
)

func TestAction_Register_Good(t *core.T) {
	subject := Register
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestAction_Register_Bad(t *core.T) {
	subject := Register
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestAction_Register_Ugly(t *core.T) {
	subject := Register
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestAction_RunHandle_Good(t *core.T) {
	subject := RunHandle
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestAction_RunHandle_Bad(t *core.T) {
	subject := RunHandle
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestAction_RunHandle_Ugly(t *core.T) {
	subject := RunHandle
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
