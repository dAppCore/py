package config

import (
	"os"
	"strconv"
	"strings" // AX-6-exception: environment key normalization uses Replacer until core exposes equivalent composition.

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes Config bindings backed by dappco.re/go/core.
//
//	config.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.config",
		Documentation: "Runtime settings backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"new":              newConfig,
			"set":              setValue,
			"get":              getValue,
			"string":           stringValue,
			"int":              intValue,
			"bool":             boolValue,
			"enable":           enableFeature,
			"disable":          disableFeature,
			"enabled":          enabledFeature,
			"enabled_features": enabledFeatures,
		},
	})
}

func newConfig(arguments ...any) (any, error) {
	return (&core.Config{}).New(), nil
}

func setValue(arguments ...any) (any, error) {
	config, err := typemap.ExpectConfig(arguments, 0, "core.config.set")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.config.set")
	if err != nil {
		return nil, err
	}
	if len(arguments) < 3 {
		return nil, core.E("core.config.set", "expected value argument", nil)
	}
	config.Set(key, arguments[2])
	return config, nil
}

func getValue(arguments ...any) (any, error) {
	config, err := typemap.ExpectConfig(arguments, 0, "core.config.get")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.config.get")
	if err != nil {
		return nil, err
	}
	result := config.Get(key)
	if result.OK {
		return result.Value, nil
	}
	if value, ok := lookupEnvironment(key); ok {
		return value, nil
	}
	return nil, nil
}

func stringValue(arguments ...any) (any, error) {
	config, err := typemap.ExpectConfig(arguments, 0, "core.config.string")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.config.string")
	if err != nil {
		return nil, err
	}
	if result := config.Get(key); result.OK {
		return config.String(key), nil
	}
	if value, ok := lookupEnvironment(key); ok {
		return value, nil
	}
	return "", nil
}

func intValue(arguments ...any) (any, error) {
	config, err := typemap.ExpectConfig(arguments, 0, "core.config.int")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.config.int")
	if err != nil {
		return nil, err
	}
	if result := config.Get(key); result.OK {
		return config.Int(key), nil
	}
	if value, ok := lookupEnvironment(key); ok {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed, nil
		}
	}
	return 0, nil
}

func boolValue(arguments ...any) (any, error) {
	config, err := typemap.ExpectConfig(arguments, 0, "core.config.bool")
	if err != nil {
		return nil, err
	}
	key, err := typemap.ExpectString(arguments, 1, "core.config.bool")
	if err != nil {
		return nil, err
	}
	if result := config.Get(key); result.OK {
		return config.Bool(key), nil
	}
	if value, ok := lookupEnvironment(key); ok {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed, nil
		}
	}
	return false, nil
}

func enableFeature(arguments ...any) (any, error) {
	config, err := typemap.ExpectConfig(arguments, 0, "core.config.enable")
	if err != nil {
		return nil, err
	}
	feature, err := typemap.ExpectString(arguments, 1, "core.config.enable")
	if err != nil {
		return nil, err
	}
	config.Enable(feature)
	return config, nil
}

func disableFeature(arguments ...any) (any, error) {
	config, err := typemap.ExpectConfig(arguments, 0, "core.config.disable")
	if err != nil {
		return nil, err
	}
	feature, err := typemap.ExpectString(arguments, 1, "core.config.disable")
	if err != nil {
		return nil, err
	}
	config.Disable(feature)
	return config, nil
}

func enabledFeature(arguments ...any) (any, error) {
	config, err := typemap.ExpectConfig(arguments, 0, "core.config.enabled")
	if err != nil {
		return nil, err
	}
	feature, err := typemap.ExpectString(arguments, 1, "core.config.enabled")
	if err != nil {
		return nil, err
	}
	return config.Enabled(feature), nil
}

func enabledFeatures(arguments ...any) (any, error) {
	config, err := typemap.ExpectConfig(arguments, 0, "core.config.enabled_features")
	if err != nil {
		return nil, err
	}
	return config.EnabledFeatures(), nil
}

func lookupEnvironment(key string) (string, bool) {
	return os.LookupEnv(environmentKey(key))
}

func environmentKey(key string) string {
	replacer := strings.NewReplacer(".", "_", "-", "_", "/", "_", " ", "_")
	return strings.ToUpper(replacer.Replace(strings.TrimSpace(key)))
}
