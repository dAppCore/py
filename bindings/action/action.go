package action

import (
	"fmt" // AX-6-exception: reflection-backed bootstrap call diagnostics need formatted type output.
	"reflect"

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Handle is a named action with an optional handler.
type Handle struct {
	Name        string
	Handler     any
	Description string
	Schema      map[string]any
	Enabled     bool
}

// Registry stores actions in registration order.
type Registry = core.Registry[*Handle]

// Register exposes Core action helpers.
//
//	action.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.action",
		Documentation: "Named action helpers for CorePy",
		Functions: map[string]runtime.Function{
			"new":          newAction,
			"new_registry": newRegistry,
			"register":     registerAction,
			"get":          getAction,
			"names":        names,
			"run":          run,
			"exists":       exists,
			"disable":      disable,
			"enable":       enable,
		},
	})
}

func newAction(arguments ...any) (any, error) {
	item := &Handle{Enabled: true, Schema: map[string]any{}}
	if len(arguments) > 0 {
		name, err := typemap.ExpectString(arguments, 0, "core.action.new")
		if err != nil {
			return nil, err
		}
		item.Name = name
	}
	if len(arguments) > 1 {
		item.Handler = arguments[1]
	}
	if len(arguments) > 2 {
		description, err := typemap.ExpectString(arguments, 2, "core.action.new")
		if err != nil {
			return nil, err
		}
		item.Description = description
	}
	if len(arguments) > 3 {
		schema, err := typemap.ExpectMap(arguments, 3, "core.action.new")
		if err != nil {
			return nil, err
		}
		item.Schema = schema
	}
	return item, nil
}

func newRegistry(arguments ...any) (any, error) {
	return core.NewRegistry[*Handle](), nil
}

func registerAction(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.action.register")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.action.register")
	if err != nil {
		return nil, err
	}
	if len(arguments) < 3 {
		return nil, fmt.Errorf("core.action.register expected argument 2")
	}

	item := &Handle{
		Name:    name,
		Handler: arguments[2],
		Enabled: true,
		Schema:  map[string]any{},
	}
	if len(arguments) > 3 {
		description, err := typemap.ExpectString(arguments, 3, "core.action.register")
		if err != nil {
			return nil, err
		}
		item.Description = description
	}
	if len(arguments) > 4 {
		schema, err := typemap.ExpectMap(arguments, 4, "core.action.register")
		if err != nil {
			return nil, err
		}
		item.Schema = schema
	}

	if _, err := typemap.ResultValue(registryValue.Set(name, item), "core.action.register"); err != nil {
		return nil, err
	}
	return item, nil
}

func getAction(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.action.get")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.action.get")
	if err != nil {
		return nil, err
	}
	result := registryValue.Get(name)
	if !result.OK {
		return &Handle{Name: name, Enabled: true, Schema: map[string]any{}}, nil
	}
	return result.Value, nil
}

func names(arguments ...any) (any, error) {
	registryValue, err := expectRegistry(arguments, 0, "core.action.names")
	if err != nil {
		return nil, err
	}
	return registryValue.Names(), nil
}

func run(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.action.run")
	if err != nil {
		return nil, err
	}
	options := map[string]any{}
	if len(arguments) > 1 {
		options, err = typemap.ExpectMap(arguments, 1, "core.action.run")
		if err != nil {
			return nil, err
		}
	}
	return RunHandle(item, options)
}

func exists(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.action.exists")
	if err != nil {
		return nil, err
	}
	return item.Handler != nil, nil
}

func disable(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.action.disable")
	if err != nil {
		return nil, err
	}
	item.Enabled = false
	return item, nil
}

func enable(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.action.enable")
	if err != nil {
		return nil, err
	}
	item.Enabled = true
	return item, nil
}

// RunHandle executes an action handle with map-based options.
func RunHandle(item *Handle, options map[string]any) (any, error) {
	if item == nil || item.Handler == nil {
		name := "<nil>"
		if item != nil && item.Name != "" {
			name = item.Name
		}
		return nil, fmt.Errorf("action not registered: %s", name)
	}
	if !item.Enabled {
		return nil, fmt.Errorf("action disabled: %s", item.Name)
	}

	switch typed := item.Handler.(type) {
	case runtime.Function:
		return typed(options)
	}

	handlerValue := reflect.ValueOf(item.Handler)
	if !handlerValue.IsValid() || handlerValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("action handler is not callable: %T", item.Handler)
	}

	handlerType := handlerValue.Type()
	callArguments := []reflect.Value{}
	if handlerType.IsVariadic() || handlerType.NumIn() > 1 {
		return nil, fmt.Errorf("action handler signature is not supported: %T", item.Handler)
	}
	if handlerType.NumIn() == 1 {
		argumentType := handlerType.In(0)
		if argumentType.Kind() == reflect.Interface {
			callArguments = append(callArguments, reflect.ValueOf(options))
		} else if reflect.TypeOf(options).AssignableTo(argumentType) {
			callArguments = append(callArguments, reflect.ValueOf(options))
		} else {
			return nil, fmt.Errorf("action handler parameter is not supported: %s", argumentType)
		}
	}

	returnValues := handlerValue.Call(callArguments)
	switch len(returnValues) {
	case 0:
		return nil, nil
	case 1:
		if errValue, ok := returnValues[0].Interface().(error); ok {
			return nil, errValue
		}
		return returnValues[0].Interface(), nil
	case 2:
		var err error
		if !returnValues[1].IsNil() {
			err = returnValues[1].Interface().(error)
		}
		return returnValues[0].Interface(), err
	default:
		return nil, fmt.Errorf("action handler returned unsupported arity: %d", len(returnValues))
	}
}

func expectRegistry(arguments []any, index int, functionName string) (*Registry, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*Registry)
	if !ok {
		return nil, fmt.Errorf("%s expected action registry, got %T", functionName, arguments[index])
	}
	return value, nil
}

func expectHandle(arguments []any, index int, functionName string) (*Handle, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*Handle)
	if !ok {
		return nil, fmt.Errorf("%s expected action handle, got %T", functionName, arguments[index])
	}
	return value, nil
}
