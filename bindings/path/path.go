package pathbinding

import (
	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes path helpers backed by dappco.re/go/core.
//
//	pathbinding.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.path",
		Documentation: "Path helpers backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"join":   join,
			"base":   base,
			"dir":    dir,
			"ext":    ext,
			"is_abs": isAbs,
			"clean":  clean,
			"glob":   glob,
		},
	})
}

func join(arguments ...any) (any, error) {
	segments, err := stringArguments(arguments, 0, "core.path.join")
	if err != nil {
		return nil, err
	}
	return core.JoinPath(segments...), nil
}

func base(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.path.base")
	if err != nil {
		return nil, err
	}
	return core.PathBase(value), nil
}

func dir(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.path.dir")
	if err != nil {
		return nil, err
	}
	return core.PathDir(value), nil
}

func ext(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.path.ext")
	if err != nil {
		return nil, err
	}
	return core.PathExt(value), nil
}

func isAbs(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.path.is_abs")
	if err != nil {
		return nil, err
	}
	return core.PathIsAbs(value), nil
}

func clean(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.path.clean")
	if err != nil {
		return nil, err
	}
	separator := core.Env("DS")
	if separator == "" {
		separator = "/"
	}
	return core.CleanPath(value, separator), nil
}

func glob(arguments ...any) (any, error) {
	pattern, err := typemap.ExpectString(arguments, 0, "core.path.glob")
	if err != nil {
		return nil, err
	}
	return core.PathGlob(pattern), nil
}

func stringArguments(arguments []any, startIndex int, functionName string) ([]string, error) {
	values := make([]string, 0, len(arguments)-startIndex)
	for index := startIndex; index < len(arguments); index++ {
		value, err := typemap.ExpectString(arguments, index, functionName)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}
