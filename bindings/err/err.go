package err

import (
	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes error helpers backed by dappco.re/go/core.
//
//	err.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.err",
		Documentation: "Structured errors backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"e":          e,
			"wrap":       wrap,
			"message":    message,
			"operation":  operation,
			"error_code": errorCode,
			"root":       root,
		},
	})
}

func e(arguments ...any) (any, error) {
	operation, err := typemap.ExpectString(arguments, 0, "core.err.e")
	if err != nil {
		return nil, err
	}
	message, err := typemap.ExpectString(arguments, 1, "core.err.e")
	if err != nil {
		return nil, err
	}
	cause, err := typemap.OptionalError(arguments, 2, "core.err.e")
	if len(arguments) > 2 && err != nil {
		return nil, err
	}

	code := ""
	if len(arguments) > 3 {
		code, err = typemap.ExpectString(arguments, 3, "core.err.e")
		if err != nil {
			return nil, err
		}
	}

	if code != "" {
		return core.WrapCode(cause, code, operation, message), nil
	}
	return core.E(operation, message, cause), nil
}

func wrap(arguments ...any) (any, error) {
	sourceError, err := typemap.OptionalError(arguments, 0, "core.err.wrap")
	if err != nil {
		return nil, err
	}
	operation, err := typemap.ExpectString(arguments, 1, "core.err.wrap")
	if err != nil {
		return nil, err
	}
	message, err := typemap.ExpectString(arguments, 2, "core.err.wrap")
	if err != nil {
		return nil, err
	}

	code := ""
	if len(arguments) > 3 {
		code, err = typemap.ExpectString(arguments, 3, "core.err.wrap")
		if err != nil {
			return nil, err
		}
	}

	if code != "" {
		return core.WrapCode(sourceError, code, operation, message), nil
	}
	return core.Wrap(sourceError, operation, message), nil
}

func message(arguments ...any) (any, error) {
	sourceError, err := typemap.OptionalError(arguments, 0, "core.err.message")
	if err != nil {
		return nil, err
	}
	return core.ErrorMessage(sourceError), nil
}

func operation(arguments ...any) (any, error) {
	sourceError, err := typemap.OptionalError(arguments, 0, "core.err.operation")
	if err != nil {
		return nil, err
	}
	return core.Operation(sourceError), nil
}

func errorCode(arguments ...any) (any, error) {
	sourceError, err := typemap.OptionalError(arguments, 0, "core.err.error_code")
	if err != nil {
		return nil, err
	}
	return core.ErrorCode(sourceError), nil
}

func root(arguments ...any) (any, error) {
	sourceError, err := typemap.OptionalError(arguments, 0, "core.err.root")
	if err != nil {
		return nil, err
	}
	return core.Root(sourceError), nil
}
