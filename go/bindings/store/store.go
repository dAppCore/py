package store

import "github.com/dappcore/py/go/runtime"

// Register exposes the planned Store module surface.
//
//	store.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.store",
		Documentation: "SQLite KV and workspace helpers for CorePy; native binding pending",
		Functions: map[string]runtime.Function{
			"available": available,
		},
	})
}

func available(arguments ...any) (any, error) {
	return false, nil
}
