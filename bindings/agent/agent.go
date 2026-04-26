package agent

import "dappco.re/go/py/runtime"

// Register exposes the planned Agent module surface.
//
//	agent.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.agent",
		Documentation: "Agent dispatch helpers for CorePy; native binding pending",
		Functions: map[string]runtime.Function{
			"available": available,
		},
	})
}

func available(arguments ...any) (any, error) {
	return false, nil
}
