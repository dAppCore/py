package array

import (
	"fmt"
	"reflect"

	"dappco.re/go/py/runtime"
)

type handle struct {
	items []any
}

// Register exposes Core array helpers.
//
//	array.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.array",
		Documentation: "Array helpers for CorePy",
		Functions: map[string]runtime.Function{
			"new":         newArray,
			"add":         add,
			"add_unique":  addUnique,
			"contains":    contains,
			"remove":      remove,
			"deduplicate": deduplicate,
			"len":         length,
			"clear":       clear,
			"as_list":     asList,
		},
	})
}

func newArray(arguments ...any) (any, error) {
	items := make([]any, 0, len(arguments))
	items = append(items, arguments...)
	return &handle{items: items}, nil
}

func add(arguments ...any) (any, error) {
	arrayValue, err := expectHandle(arguments, 0, "core.array.add")
	if err != nil {
		return nil, err
	}
	arrayValue.items = append(arrayValue.items, arguments[1:]...)
	return arrayValue, nil
}

func addUnique(arguments ...any) (any, error) {
	arrayValue, err := expectHandle(arguments, 0, "core.array.add_unique")
	if err != nil {
		return nil, err
	}
	for _, value := range arguments[1:] {
		if !containsValue(arrayValue.items, value) {
			arrayValue.items = append(arrayValue.items, value)
		}
	}
	return arrayValue, nil
}

func contains(arguments ...any) (any, error) {
	arrayValue, err := expectHandle(arguments, 0, "core.array.contains")
	if err != nil {
		return nil, err
	}
	if len(arguments) < 2 {
		return nil, fmt.Errorf("core.array.contains expected argument 1")
	}
	return containsValue(arrayValue.items, arguments[1]), nil
}

func remove(arguments ...any) (any, error) {
	arrayValue, err := expectHandle(arguments, 0, "core.array.remove")
	if err != nil {
		return nil, err
	}
	if len(arguments) < 2 {
		return nil, fmt.Errorf("core.array.remove expected argument 1")
	}
	for index, value := range arrayValue.items {
		if reflect.DeepEqual(value, arguments[1]) {
			arrayValue.items = append(arrayValue.items[:index], arrayValue.items[index+1:]...)
			break
		}
	}
	return arrayValue, nil
}

func deduplicate(arguments ...any) (any, error) {
	arrayValue, err := expectHandle(arguments, 0, "core.array.deduplicate")
	if err != nil {
		return nil, err
	}
	result := make([]any, 0, len(arrayValue.items))
	for _, value := range arrayValue.items {
		if containsValue(result, value) {
			continue
		}
		result = append(result, value)
	}
	arrayValue.items = result
	return arrayValue, nil
}

func length(arguments ...any) (any, error) {
	arrayValue, err := expectHandle(arguments, 0, "core.array.len")
	if err != nil {
		return nil, err
	}
	return len(arrayValue.items), nil
}

func clear(arguments ...any) (any, error) {
	arrayValue, err := expectHandle(arguments, 0, "core.array.clear")
	if err != nil {
		return nil, err
	}
	arrayValue.items = nil
	return arrayValue, nil
}

func asList(arguments ...any) (any, error) {
	arrayValue, err := expectHandle(arguments, 0, "core.array.as_list")
	if err != nil {
		return nil, err
	}
	result := make([]any, len(arrayValue.items))
	copy(result, arrayValue.items)
	return result, nil
}

func expectHandle(arguments []any, index int, functionName string) (*handle, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*handle)
	if !ok {
		return nil, fmt.Errorf("%s expected array handle, got %T", functionName, arguments[index])
	}
	return value, nil
}

func containsValue(items []any, target any) bool {
	for _, item := range items {
		if reflect.DeepEqual(item, target) {
			return true
		}
	}
	return false
}
