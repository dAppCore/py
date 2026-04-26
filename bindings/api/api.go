package api

import "dappco.re/go/py/runtime"

// Register exposes the planned API module surface.
//
//	api.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.api",
		Documentation: "REST server and client helpers for CorePy; native binding pending",
		Functions: map[string]runtime.Function{
			"available": available,
		},
	})
}

func available(arguments ...any) (any, error) {
	return false, nil
}
