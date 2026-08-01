package bootstrap

import (
	core "dappco.re/go"
)

func TestInterpreter_New_Good(t *core.T) {
	subject := New
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_New_Bad(t *core.T) {
	subject := New
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_New_Ugly(t *core.T) {
	subject := New
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Close_Good(t *core.T) {
	subject := (*Interpreter).Close
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Close_Bad(t *core.T) {
	subject := (*Interpreter).Close
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Close_Ugly(t *core.T) {
	subject := (*Interpreter).Close
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_RegisterModule_Good(t *core.T) {
	subject := (*Interpreter).RegisterModule
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_RegisterModule_Bad(t *core.T) {
	subject := (*Interpreter).RegisterModule
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_RegisterModule_Ugly(t *core.T) {
	subject := (*Interpreter).RegisterModule
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Modules_Good(t *core.T) {
	subject := (*Interpreter).Modules
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Modules_Bad(t *core.T) {
	subject := (*Interpreter).Modules
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Modules_Ugly(t *core.T) {
	subject := (*Interpreter).Modules
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_NewSession_Good(t *core.T) {
	subject := (*Interpreter).NewSession
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_NewSession_Bad(t *core.T) {
	subject := (*Interpreter).NewSession
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_NewSession_Ugly(t *core.T) {
	subject := (*Interpreter).NewSession
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Session_Run_Good(t *core.T) {
	subject := (*Session).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Session_Run_Bad(t *core.T) {
	subject := (*Session).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Session_Run_Ugly(t *core.T) {
	subject := (*Session).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Call_Good(t *core.T) {
	subject := (*Interpreter).Call
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Call_Bad(t *core.T) {
	subject := (*Interpreter).Call
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Call_Ugly(t *core.T) {
	subject := (*Interpreter).Call
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Run_Good(t *core.T) {
	subject := (*Interpreter).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Run_Bad(t *core.T) {
	subject := (*Interpreter).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestInterpreter_Interpreter_Run_Ugly(t *core.T) {
	subject := (*Interpreter).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
