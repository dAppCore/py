package entitlement

import (
	"fmt"

	core "dappco.re/go/core"
	"dappco.re/go/py/runtime"
)

// Register exposes Core entitlement helpers.
//
//	entitlement.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.entitlement",
		Documentation: "Entitlement helpers for CorePy",
		Functions: map[string]runtime.Function{
			"new":           newEntitlement,
			"near_limit":    nearLimit,
			"usage_percent": usagePercent,
		},
	})
}

func newEntitlement(arguments ...any) (any, error) {
	values := core.Entitlement{}
	if len(arguments) > 0 {
		allowed, ok := arguments[0].(bool)
		if !ok {
			return nil, fmt.Errorf("core.entitlement.new expected argument 0 to be bool, got %T", arguments[0])
		}
		values.Allowed = allowed
	}
	if len(arguments) > 1 {
		unlimited, ok := arguments[1].(bool)
		if !ok {
			return nil, fmt.Errorf("core.entitlement.new expected argument 1 to be bool, got %T", arguments[1])
		}
		values.Unlimited = unlimited
	}
	if len(arguments) > 2 {
		limit, ok := arguments[2].(int)
		if !ok {
			return nil, fmt.Errorf("core.entitlement.new expected argument 2 to be int, got %T", arguments[2])
		}
		values.Limit = limit
	}
	if len(arguments) > 3 {
		used, ok := arguments[3].(int)
		if !ok {
			return nil, fmt.Errorf("core.entitlement.new expected argument 3 to be int, got %T", arguments[3])
		}
		values.Used = used
	}
	if len(arguments) > 4 {
		remaining, ok := arguments[4].(int)
		if !ok {
			return nil, fmt.Errorf("core.entitlement.new expected argument 4 to be int, got %T", arguments[4])
		}
		values.Remaining = remaining
	} else {
		values.Remaining = values.Limit - values.Used
	}
	if len(arguments) > 5 {
		reason, ok := arguments[5].(string)
		if !ok {
			return nil, fmt.Errorf("core.entitlement.new expected argument 5 to be string, got %T", arguments[5])
		}
		values.Reason = reason
	}
	return values, nil
}

func nearLimit(arguments ...any) (any, error) {
	value, err := expectEntitlement(arguments, 0, "core.entitlement.near_limit")
	if err != nil {
		return nil, err
	}
	if len(arguments) < 2 {
		return nil, fmt.Errorf("core.entitlement.near_limit expected argument 1")
	}
	threshold, ok := arguments[1].(float64)
	if !ok {
		return nil, fmt.Errorf("core.entitlement.near_limit expected argument 1 to be float, got %T", arguments[1])
	}
	return value.NearLimit(threshold), nil
}

func usagePercent(arguments ...any) (any, error) {
	value, err := expectEntitlement(arguments, 0, "core.entitlement.usage_percent")
	if err != nil {
		return nil, err
	}
	return value.UsagePercent(), nil
}

func expectEntitlement(arguments []any, index int, functionName string) (core.Entitlement, error) {
	if index >= len(arguments) {
		return core.Entitlement{}, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	switch typed := arguments[index].(type) {
	case core.Entitlement:
		return typed, nil
	case *core.Entitlement:
		if typed == nil {
			return core.Entitlement{}, fmt.Errorf("%s expected entitlement value, got nil", functionName)
		}
		return *typed, nil
	default:
		return core.Entitlement{}, fmt.Errorf("%s expected entitlement value, got %T", functionName, arguments[index])
	}
}
