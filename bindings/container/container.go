package container

import "dappco.re/go/py/runtime"

// Register exposes the planned Container module surface.
//
//	container.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.container",
		Documentation: "Container orchestration helpers for CorePy; native binding pending",
		Functions: map[string]runtime.Function{
			"available": available,
		},
	})
}

func available(arguments ...any) (any, error) {
	return false, nil
}
