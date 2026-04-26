package log

import (
	"fmt" // AX-6-exception: log level parser reports unsupported level names during bootstrap.

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes logging bindings backed by dappco.re/go/core.
//
//	log.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.log",
		Documentation: "Structured logging backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"set_level": setLevel,
			"debug":     debug,
			"info":      info,
			"warn":      warn,
			"error":     errorMessage,
		},
	})
}

func setLevel(arguments ...any) (any, error) {
	levelName, err := typemap.ExpectString(arguments, 0, "core.log.set_level")
	if err != nil {
		return nil, err
	}
	level, err := parseLevel(levelName)
	if err != nil {
		return nil, err
	}
	core.SetLevel(level)
	return true, nil
}

func debug(arguments ...any) (any, error) {
	return logWith(core.Debug, "core.log.debug", arguments...)
}

func info(arguments ...any) (any, error) {
	return logWith(core.Info, "core.log.info", arguments...)
}

func warn(arguments ...any) (any, error) {
	return logWith(core.Warn, "core.log.warn", arguments...)
}

func errorMessage(arguments ...any) (any, error) {
	return logWith(core.Error, "core.log.error", arguments...)
}

func logWith(fn func(string, ...any), functionName string, arguments ...any) (any, error) {
	message, err := typemap.ExpectString(arguments, 0, functionName)
	if err != nil {
		return nil, err
	}
	fn(message, arguments[1:]...)
	return true, nil
}

func parseLevel(levelName string) (core.Level, error) {
	switch levelName {
	case "quiet":
		return core.LevelQuiet, nil
	case "error":
		return core.LevelError, nil
	case "warn":
		return core.LevelWarn, nil
	case "info":
		return core.LevelInfo, nil
	case "debug":
		return core.LevelDebug, nil
	default:
		return core.LevelInfo, fmt.Errorf("unknown log level %q", levelName)
	}
}
