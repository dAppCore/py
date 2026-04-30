package options

import (
	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes Core Options bindings.
//
//	options.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.options",
		Documentation: "Typed option primitives backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"new":    newOptions,
			"set":    setValue,
			"get":    getValue,
			"has":    hasKey,
			"string": stringValue,
			"int":    intValue,
			"bool":   boolValue,
			"items":  items,
		},
	})
}

func newOptions(arguments ...any) (any, error) {
	if len(arguments) == 0 {
		options := core.NewOptions()
		return &options, nil
	}
	values, err := typemap.ExpectMap(arguments, 0, "core.options.new")
	if err != nil {
		return nil, err
	}
	return typemap.MapToOptions(values), nil
}

func setValue(arguments ...any) (any, error) {
	options, err := typemap.ExpectOptions(arguments, 0, "core.options.set")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.options.set")
	if err != nil {
		return nil, err
	}
	if len(arguments) < 3 {
		return nil, core.E("core.options.set", "expected value argument", nil)
	}
	options.Set(key, arguments[2])
	return options, nil
}

func getValue(arguments ...any) (any, error) {
	options, err := typemap.ExpectOptions(arguments, 0, "core.options.get")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.options.get")
	if err != nil {
		return nil, err
	}
	result := options.Get(key)
	if !result.OK {
		return nil, nil
	}
	return result.Value, nil
}

func hasKey(arguments ...any) (any, error) {
	options, err := typemap.ExpectOptions(arguments, 0, "core.options.has")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.options.has")
	if err != nil {
		return nil, err
	}
	return options.Has(key), nil
}

func stringValue(arguments ...any) (any, error) {
	options, err := typemap.ExpectOptions(arguments, 0, "core.options.string")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.options.string")
	if err != nil {
		return nil, err
	}
	return options.String(key), nil
}

func intValue(arguments ...any) (any, error) {
	options, err := typemap.ExpectOptions(arguments, 0, "core.options.int")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.options.int")
	if err != nil {
		return nil, err
	}
	return options.Int(key), nil
}

func boolValue(arguments ...any) (any, error) {
	options, err := typemap.ExpectOptions(arguments, 0, "core.options.bool")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.options.bool")
	if err != nil {
		return nil, err
	}
	return options.Bool(key), nil
}

func items(arguments ...any) (any, error) {
	options, err := typemap.ExpectOptions(arguments, 0, "core.options.items")
	if err != nil {
		return nil, err
	}
	return typemap.OptionsToMap(options), nil
}
