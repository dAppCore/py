package mcp

import "dappco.re/go/py/runtime"

// Register exposes the planned MCP module surface.
//
//	mcp.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.mcp",
		Documentation: "MCP tool protocol helpers for CorePy; native binding pending",
		Functions: map[string]runtime.Function{
			"available": available,
		},
	})
}

func available(arguments ...any) (any, error) {
	return false, nil
}
