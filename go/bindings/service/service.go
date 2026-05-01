package service

import (
	"context"

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes Service bindings backed by dappco.re/go/core.
//
//	service.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.service",
		Documentation: "Service registry backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"new":       newCore,
			"register":  registerService,
			"get":       getService,
			"names":     names,
			"start_all": startAll,
			"stop_all":  stopAll,
		},
	})
}

func newCore(arguments ...any) (any, error) {
	if len(arguments) == 0 {
		return core.New(), nil
	}
	name, err := typemap.ExpectString(arguments, 0, "core.service.new")
	if err != nil {
		return nil, err
	}
	return core.New(core.WithOption("name", name)), nil
}

func registerService(arguments ...any) (any, error) {
	instance, err := typemap.ExpectCore(arguments, 0, "core.service.register")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.service.register")
	if err != nil {
		return nil, err
	}
	if len(arguments) > 2 {
		if _, err := typemap.ResultValue(instance.RegisterService(name, arguments[2]), "core.service.register"); err != nil {
			return nil, err
		}
		return instance, nil
	}
	if _, err := typemap.ResultValue(instance.Service(name, core.Service{}), "core.service.register"); err != nil {
		return nil, err
	}
	return instance, nil
}

func getService(arguments ...any) (any, error) {
	instance, err := typemap.ExpectCore(arguments, 0, "core.service.get")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.service.get")
	if err != nil {
		return nil, err
	}
	return typemap.ResultValue(instance.Service(name), "core.service.get")
}

func names(arguments ...any) (any, error) {
	instance, err := typemap.ExpectCore(arguments, 0, "core.service.names")
	if err != nil {
		return nil, err
	}
	return instance.Services(), nil
}

func startAll(arguments ...any) (any, error) {
	instance, err := typemap.ExpectCore(arguments, 0, "core.service.start_all")
	if err != nil {
		return nil, err
	}
	if _, err := typemap.ResultValue(instance.ServiceStartup(context.Background(), nil), "core.service.start_all"); err != nil {
		return nil, err
	}
	return true, nil
}

func stopAll(arguments ...any) (any, error) {
	instance, err := typemap.ExpectCore(arguments, 0, "core.service.stop_all")
	if err != nil {
		return nil, err
	}
	if _, err := typemap.ResultValue(instance.ServiceShutdown(context.Background()), "core.service.stop_all"); err != nil {
		return nil, err
	}
	return true, nil
}
