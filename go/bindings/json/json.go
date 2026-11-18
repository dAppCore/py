package json

import (
	core "github.com/dappcore/go"
	"github.com/dappcore/py/go/bindings/typemap"
	"github.com/dappcore/py/go/runtime"
)

// Register exposes JSON bindings backed by github.com/dappcore/go.
//
//	json.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.json",
		Documentation: "JSON helpers backed by github.com/dappcore/go",
		Functions: map[string]runtime.Function{
			"dumps": dumps,
			"loads": loads,
		},
	})
}

func dumps(arguments ...any) (any, error) {
	if len(arguments) != 1 {
		return nil, core.E("core.json.dumps", "expected exactly one argument", nil)
	}
	return core.JSONMarshalString(arguments[0]), nil
}

func loads(arguments ...any) (any, error) {
	text, err := typemap.ExpectString(arguments, 0, "core.json.loads")
	if err != nil {
		return nil, err
	}
	var value any
	if _, err := typemap.ResultValue(core.JSONUnmarshalString(text, &value), "core.json.loads"); err != nil {
		return nil, err
	}
	return value, nil
}
