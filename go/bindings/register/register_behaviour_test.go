package register

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// TestRegister_DefaultModules_Real_Good registers the full default module set
// and the stdlib shadows against a bootstrap interpreter.
func TestRegister_DefaultModules_Real_Good(t *core.T) {
	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := DefaultModules(interpreter); err != nil {
		t.Fatalf("DefaultModules: %v", err)
	}

	lister, ok := interpreter.(runtime.ModuleLister)
	if !ok {
		t.Fatalf("interpreter does not list modules: %T", interpreter)
	}
	registered := map[string]struct{}{}
	for _, name := range lister.Modules() {
		registered[name] = struct{}{}
	}

	for _, name := range DefaultModuleNames() {
		if name == "core.echo" {
			// echo registers under the root "core" module name.
			continue
		}
		if _, ok := registered[name]; !ok {
			t.Fatalf("expected module %q to be registered", name)
		}
	}
}

// TestRegister_DefaultModuleNames_Real_Good reports a spec-aligned name list.
func TestRegister_DefaultModuleNames_Real_Good(t *core.T) {
	specs := DefaultModuleSpecs()
	names := DefaultModuleNames()
	if len(specs) == 0 || len(specs) != len(names) {
		t.Fatalf("specs/names length mismatch: %d vs %d", len(specs), len(names))
	}
	for index, spec := range specs {
		if spec.Name != names[index] {
			t.Fatalf("name mismatch at %d: %q vs %q", index, spec.Name, names[index])
		}
		if spec.Register == nil {
			t.Fatalf("spec %q has a nil Register hook", spec.Name)
		}
	}
}

// TestRegister_DefaultShadowModuleNames_Real_Good reports the stdlib shadows.
func TestRegister_DefaultShadowModuleNames_Real_Good(t *core.T) {
	specs := DefaultShadowModuleSpecs()
	names := DefaultShadowModuleNames()
	if len(specs) == 0 || len(specs) != len(names) {
		t.Fatalf("shadow specs/names mismatch: %d vs %d", len(specs), len(names))
	}
	for index, spec := range specs {
		if spec.Name != names[index] {
			t.Fatalf("shadow name mismatch at %d: %q vs %q", index, spec.Name, names[index])
		}
		if spec.Register == nil {
			t.Fatalf("shadow spec %q has a nil Register hook", spec.Name)
		}
	}
}

// TestRegister_ShadowModules_Real_Good registers the stdlib shadows and lists
// them on the interpreter.
func TestRegister_ShadowModules_Real_Good(t *core.T) {
	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	for _, spec := range DefaultShadowModuleSpecs() {
		if err := spec.Register(interpreter); err != nil {
			t.Fatalf("register shadow %q: %v", spec.Name, err)
		}
	}

	lister, ok := interpreter.(runtime.ModuleLister)
	if !ok {
		t.Fatalf("interpreter does not list modules: %T", interpreter)
	}
	if len(lister.Modules()) == 0 {
		t.Fatal("expected shadow modules to be registered")
	}
}
