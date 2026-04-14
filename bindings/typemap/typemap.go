package typemap

import (
	"fmt"
	"sort"

	core "dappco.re/go/core"
)

// ResultValue unwraps a Core Result into a plain Go value.
//
//	value, err := typemap.ResultValue(result, "core.fs.read_file")
func ResultValue(result core.Result, functionName string) (any, error) {
	if result.OK {
		return result.Value, nil
	}
	if result.Value == nil {
		return nil, fmt.Errorf("%s failed", functionName)
	}
	if err, ok := result.Value.(error); ok {
		return nil, err
	}
	return nil, fmt.Errorf("%s failed: %v", functionName, result.Value)
}

// ExpectString returns the string argument at the given index.
//
//	path, err := typemap.ExpectString(arguments, 0, "core.fs.read_file")
func ExpectString(arguments []any, index int, functionName string) (string, error) {
	if index >= len(arguments) {
		return "", fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(string)
	if !ok {
		return "", fmt.Errorf("%s expected argument %d to be string, got %T", functionName, index, arguments[index])
	}
	return value, nil
}

// ExpectBytes returns a byte slice argument at the given index.
//
//	content, err := typemap.ExpectBytes(arguments, 1, "core.fs.write_bytes")
func ExpectBytes(arguments []any, index int, functionName string) ([]byte, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}

	switch typed := arguments[index].(type) {
	case []byte:
		return append([]byte(nil), typed...), nil
	case string:
		return []byte(typed), nil
	default:
		return nil, fmt.Errorf("%s expected argument %d to be []byte, got %T", functionName, index, arguments[index])
	}
}

// ExpectMap returns the map argument at the given index.
//
//	values, err := typemap.ExpectMap(arguments, 0, "core.options.new")
func ExpectMap(arguments []any, index int, functionName string) (map[string]any, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s expected argument %d to be map[string]any, got %T", functionName, index, arguments[index])
	}
	return value, nil
}

// ExpectOptions returns an Options pointer from either a pointer, value, or map.
//
//	options, err := typemap.ExpectOptions(arguments, 0, "core.options.set")
func ExpectOptions(arguments []any, index int, functionName string) (*core.Options, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}

	switch typed := arguments[index].(type) {
	case *core.Options:
		return typed, nil
	case core.Options:
		options := typed
		return &options, nil
	case map[string]any:
		return MapToOptions(typed), nil
	default:
		return nil, fmt.Errorf("%s expected Options-compatible value, got %T", functionName, arguments[index])
	}
}

// OptionsToMap returns a map copy of the option items.
//
//	values := typemap.OptionsToMap(options)
func OptionsToMap(options *core.Options) map[string]any {
	values := map[string]any{}
	if options == nil {
		return values
	}
	for _, item := range options.Items() {
		values[item.Key] = item.Value
	}
	return values
}

// MapToOptions converts a Python-style dict into Core Options.
//
//	options := typemap.MapToOptions(map[string]any{"name": "corepy"})
func MapToOptions(values map[string]any) *core.Options {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	items := make([]core.Option, 0, len(keys))
	for _, key := range keys {
		items = append(items, core.Option{Key: key, Value: values[key]})
	}
	options := core.NewOptions(items...)
	return &options
}

// ExpectConfig returns a Config pointer.
//
//	config, err := typemap.ExpectConfig(arguments, 0, "core.config.set")
func ExpectConfig(arguments []any, index int, functionName string) (*core.Config, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*core.Config)
	if !ok {
		return nil, fmt.Errorf("%s expected *core.Config, got %T", functionName, arguments[index])
	}
	return value, nil
}

// ExpectData returns a Data pointer.
//
//	data, err := typemap.ExpectData(arguments, 0, "core.data.mount_path")
func ExpectData(arguments []any, index int, functionName string) (*core.Data, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*core.Data)
	if !ok {
		return nil, fmt.Errorf("%s expected *core.Data, got %T", functionName, arguments[index])
	}
	return value, nil
}

// ExpectCore returns a Core pointer.
//
//	instance, err := typemap.ExpectCore(arguments, 0, "core.service.register")
func ExpectCore(arguments []any, index int, functionName string) (*core.Core, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*core.Core)
	if !ok {
		return nil, fmt.Errorf("%s expected *core.Core, got %T", functionName, arguments[index])
	}
	return value, nil
}

// ExpectError returns an error argument.
//
//	err, convErr := typemap.ExpectError(arguments, 0, "core.err.wrap")
func ExpectError(arguments []any, index int, functionName string) (error, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(error)
	if !ok {
		return nil, fmt.Errorf("%s expected error argument, got %T", functionName, arguments[index])
	}
	return value, nil
}

// OptionalError returns an error argument or nil when the value is None/nil.
//
//	err, convErr := typemap.OptionalError(arguments, 0, "core.err.wrap")
func OptionalError(arguments []any, index int, functionName string) (error, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	if arguments[index] == nil {
		return nil, nil
	}
	return ExpectError(arguments, index, functionName)
}
