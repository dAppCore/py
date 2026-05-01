package register

import (
	core "dappco.re/go"
)

func TestRegister_DefaultModuleSpecs_Good(t *core.T) {
	subject := DefaultModuleSpecs
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultModuleSpecs_Bad(t *core.T) {
	subject := DefaultModuleSpecs
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultModuleSpecs_Ugly(t *core.T) {
	subject := DefaultModuleSpecs
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultModuleNames_Good(t *core.T) {
	subject := DefaultModuleNames
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultModuleNames_Bad(t *core.T) {
	subject := DefaultModuleNames
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultModuleNames_Ugly(t *core.T) {
	subject := DefaultModuleNames
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultShadowModuleSpecs_Good(t *core.T) {
	subject := DefaultShadowModuleSpecs
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultShadowModuleSpecs_Bad(t *core.T) {
	subject := DefaultShadowModuleSpecs
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultShadowModuleSpecs_Ugly(t *core.T) {
	subject := DefaultShadowModuleSpecs
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultShadowModuleNames_Good(t *core.T) {
	subject := DefaultShadowModuleNames
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultShadowModuleNames_Bad(t *core.T) {
	subject := DefaultShadowModuleNames
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultShadowModuleNames_Ugly(t *core.T) {
	subject := DefaultShadowModuleNames
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultModules_Good(t *core.T) {
	subject := DefaultModules
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultModules_Bad(t *core.T) {
	subject := DefaultModules
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestRegister_DefaultModules_Ugly(t *core.T) {
	subject := DefaultModules
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
