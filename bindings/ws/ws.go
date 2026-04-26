package ws

import "dappco.re/go/py/runtime"

// Register exposes the planned WebSocket module surface.
//
//	ws.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.ws",
		Documentation: "WebSocket helpers for CorePy; native binding pending",
		Functions: map[string]runtime.Function{
			"available": available,
		},
	})
}

func available(arguments ...any) (any, error) {
	return false, nil
}
