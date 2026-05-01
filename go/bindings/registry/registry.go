package registry

import (
	"fmt" // AX-6-exception: registry bootstrap validation reports dynamic Go types.

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes Core registry helpers.
//
//	registry.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.registry",
		Documentation: "Named collection helpers for CorePy",
		Functions: map[string]runtime.Function{
			"new":      newRegistry,
			"set":      set,
			"get":      get,
			"has":      has,
			"names":    names,
			"list":     list,
			"len":      length,
			"delete":   deleteValue,
			"disable":  disable,
			"enable":   enable,
			"disabled": disabled,
			"lock":     lock,
			"locked":   locked,
			"seal":     seal,
			"sealed":   sealed,
			"open":     open,
		},
	})
}

func newRegistry(arguments ...any) (any, error) {
	return core.NewRegistry[any](), nil
}

func set(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.set")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.registry.set")
	if err != nil {
		return nil, err
	}
	if len(arguments) < 3 {
		return nil, fmt.Errorf("core.registry.set expected argument 2")
	}
	if _, err := typemap.ResultValue(registryValue.Set(name, arguments[2]), "core.registry.set"); err != nil {
		return nil, err
	}
	return registryValue, nil
}

func get(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.get")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.registry.get")
	if err != nil {
		return nil, err
	}
	result := registryValue.Get(name)
	if result.OK {
		return result.Value, nil
	}
	if len(arguments) > 2 {
		return arguments[2], nil
	}
	return nil, nil
}

func has(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.has")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.registry.has")
	if err != nil {
		return nil, err
	}
	return registryValue.Has(name), nil
}

func names(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.names")
	if err != nil {
		return nil, err
	}
	return registryValue.Names(), nil
}

func list(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.list")
	if err != nil {
		return nil, err
	}
	pattern, err := typemap.ExpectString(arguments, 1, "core.registry.list")
	if err != nil {
		return nil, err
	}
	return registryValue.List(pattern), nil
}

func length(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.len")
	if err != nil {
		return nil, err
	}
	return registryValue.Len(), nil
}

func deleteValue(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.delete")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.registry.delete")
	if err != nil {
		return nil, err
	}
	if _, err := typemap.ResultValue(registryValue.Delete(name), "core.registry.delete"); err != nil {
		return nil, err
	}
	return true, nil
}

func disable(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.disable")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.registry.disable")
	if err != nil {
		return nil, err
	}
	if _, err := typemap.ResultValue(registryValue.Disable(name), "core.registry.disable"); err != nil {
		return nil, err
	}
	return registryValue, nil
}

func enable(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.enable")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.registry.enable")
	if err != nil {
		return nil, err
	}
	if _, err := typemap.ResultValue(registryValue.Enable(name), "core.registry.enable"); err != nil {
		return nil, err
	}
	return registryValue, nil
}

func disabled(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.disabled")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.registry.disabled")
	if err != nil {
		return nil, err
	}
	return registryValue.Disabled(name), nil
}

func lock(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.lock")
	if err != nil {
		return nil, err
	}
	registryValue.Lock()
	return registryValue, nil
}

func locked(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.locked")
	if err != nil {
		return nil, err
	}
	return registryValue.Locked(), nil
}

func seal(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.seal")
	if err != nil {
		return nil, err
	}
	registryValue.Seal()
	return registryValue, nil
}

func sealed(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.sealed")
	if err != nil {
		return nil, err
	}
	return registryValue.Sealed(), nil
}

func open(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.registry.open")
	if err != nil {
		return nil, err
	}
	registryValue.Open()
	return registryValue, nil
}

func expectRegistry(arguments []any, index int, functionName string) (*core.Registry[any], error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*core.Registry[any])
	if !ok {
		return nil, fmt.Errorf("%s expected registry handle, got %T", functionName, arguments[index])
	}
	return value, nil
}
