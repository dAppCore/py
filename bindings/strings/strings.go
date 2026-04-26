package stringsbinding

import (
	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes string helpers backed by dappco.re/go/core.
//
//	stringsbinding.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.strings",
		Documentation: "String helpers backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"contains":    contains,
			"trim":        trim,
			"trim_prefix": trimPrefix,
			"trim_suffix": trimSuffix,
			"has_prefix":  hasPrefix,
			"has_suffix":  hasSuffix,
			"split":       split,
			"split_n":     splitN,
			"join":        join,
			"replace":     replace,
			"lower":       lower,
			"upper":       upper,
			"rune_count":  runeCount,
			"concat":      concat,
		},
	})
}

func contains(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.contains")
	if err != nil {
		return nil, err
	}
	substring, err := typemap.ExpectString(arguments, 1, "core.strings.contains")
	if err != nil {
		return nil, err
	}
	return core.Contains(value, substring), nil
}

func trim(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.trim")
	if err != nil {
		return nil, err
	}
	return core.Trim(value), nil
}

func trimPrefix(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.trim_prefix")
	if err != nil {
		return nil, err
	}
	prefix, err := typemap.ExpectString(arguments, 1, "core.strings.trim_prefix")
	if err != nil {
		return nil, err
	}
	return core.TrimPrefix(value, prefix), nil
}

func trimSuffix(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.trim_suffix")
	if err != nil {
		return nil, err
	}
	suffix, err := typemap.ExpectString(arguments, 1, "core.strings.trim_suffix")
	if err != nil {
		return nil, err
	}
	return core.TrimSuffix(value, suffix), nil
}

func hasPrefix(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.has_prefix")
	if err != nil {
		return nil, err
	}
	prefix, err := typemap.ExpectString(arguments, 1, "core.strings.has_prefix")
	if err != nil {
		return nil, err
	}
	return core.HasPrefix(value, prefix), nil
}

func hasSuffix(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.has_suffix")
	if err != nil {
		return nil, err
	}
	suffix, err := typemap.ExpectString(arguments, 1, "core.strings.has_suffix")
	if err != nil {
		return nil, err
	}
	return core.HasSuffix(value, suffix), nil
}

func split(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.split")
	if err != nil {
		return nil, err
	}
	separator, err := typemap.ExpectString(arguments, 1, "core.strings.split")
	if err != nil {
		return nil, err
	}
	return core.Split(value, separator), nil
}

func splitN(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.split_n")
	if err != nil {
		return nil, err
	}
	separator, err := typemap.ExpectString(arguments, 1, "core.strings.split_n")
	if err != nil {
		return nil, err
	}
	limit, err := typemap.ExpectInt(arguments, 2, "core.strings.split_n")
	if err != nil {
		return nil, err
	}
	return core.SplitN(value, separator, limit), nil
}

func join(arguments ...any) (any, error) {
	separator, err := typemap.ExpectString(arguments, 0, "core.strings.join")
	if err != nil {
		return nil, err
	}
	parts, err := stringArguments(arguments, 1, "core.strings.join")
	if err != nil {
		return nil, err
	}
	return core.Join(separator, parts...), nil
}

func replace(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.replace")
	if err != nil {
		return nil, err
	}
	oldValue, err := typemap.ExpectString(arguments, 1, "core.strings.replace")
	if err != nil {
		return nil, err
	}
	newValue, err := typemap.ExpectString(arguments, 2, "core.strings.replace")
	if err != nil {
		return nil, err
	}
	return core.Replace(value, oldValue, newValue), nil
}

func lower(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.lower")
	if err != nil {
		return nil, err
	}
	return core.Lower(value), nil
}

func upper(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.upper")
	if err != nil {
		return nil, err
	}
	return core.Upper(value), nil
}

func runeCount(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.strings.rune_count")
	if err != nil {
		return nil, err
	}
	return core.RuneCount(value), nil
}

func concat(arguments ...any) (any, error) {
	parts, err := stringArguments(arguments, 0, "core.strings.concat")
	if err != nil {
		return nil, err
	}
	return core.Concat(parts...), nil
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
