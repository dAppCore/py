package info

import (
	"slices"

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes Core system information helpers.
//
//	info.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.info",
		Documentation: "System information helpers for CorePy",
		Functions: map[string]runtime.Function{
			"env":      env,
			"keys":     keys,
			"snapshot": snapshot,
		},
	})
}

func env(arguments ...any) (any, error) {
	key, err := typemap.ExpectString(arguments, 0, "core.info.env")
	if err != nil {
		return nil, err
	}
	return core.Env(key), nil
}

func keys(arguments ...any) (any, error) {
	values := core.EnvKeys()
	slices.Sort(values)
	return values, nil
}

func snapshot(arguments ...any) (any, error) {
	keys := core.EnvKeys()
	slices.Sort(keys)
	values := make(map[string]any, len(keys))
	for _, key := range keys {
		values[key] = core.Env(key)
	}
	return values, nil
}
