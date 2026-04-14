package echo

import "dappco.re/go/py/runtime"

// Register exposes the bootstrap `core.echo` round-trip.
//
//	echo.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core",
		Documentation: "Root CorePy module",
		Functions: map[string]runtime.Function{
			"echo": func(arguments ...any) (any, error) {
				if len(arguments) != 1 {
					return nil, runtimeError("core.echo", "expected exactly one argument")
				}
				return arguments[0], nil
			},
		},
	})
}

func runtimeError(functionName, message string) error {
	return &echoError{functionName: functionName, message: message}
}

type echoError struct {
	functionName string
	message      string
}

func (err *echoError) Error() string {
	return err.functionName + ": " + err.message
}
