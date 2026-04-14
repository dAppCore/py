package medium

import (
	"os"
	"path/filepath"
	"unicode/utf8"

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes Medium bindings for memory and filesystem-backed content.
//
//	medium.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.medium",
		Documentation: "Medium-backed content helpers for memory and filesystem transports",
		Functions: map[string]runtime.Function{
			"memory":      memory,
			"from_path":   fromPath,
			"read_text":   readText,
			"write_text":  writeText,
			"read_bytes":  readBytes,
			"write_bytes": writeBytes,
		},
	})
}

type handle struct {
	location string
	text     string
	data     []byte
}

func memory(arguments ...any) (any, error) {
	initialText := ""
	if len(arguments) > 0 {
		var err error
		initialText, err = typemap.ExpectString(arguments, 0, "core.medium.memory")
		if err != nil {
			return nil, err
		}
	}

	return &handle{
		text: initialText,
		data: []byte(initialText),
	}, nil
}

func fromPath(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "core.medium.from_path")
	if err != nil {
		return nil, err
	}

	return &handle{location: path}, nil
}

func readText(arguments ...any) (any, error) {
	mediumHandle, err := expectHandle(arguments, 0, "core.medium.read_text")
	if err != nil {
		return nil, err
	}
	if mediumHandle.location == "" {
		return mediumHandle.text, nil
	}

	return typemap.ResultValue(filesystem().Read(mediumHandle.location), "core.medium.read_text")
}

func writeText(arguments ...any) (any, error) {
	mediumHandle, err := expectHandle(arguments, 0, "core.medium.write_text")
	if err != nil {
		return nil, err
	}
	value, err := typemap.ExpectString(arguments, 1, "core.medium.write_text")
	if err != nil {
		return nil, err
	}

	if mediumHandle.location == "" {
		mediumHandle.text = value
		mediumHandle.data = []byte(value)
		return value, nil
	}

	if err := ensureParentDir(mediumHandle.location, "core.medium.write_text"); err != nil {
		return nil, err
	}
	if _, err := typemap.ResultValue(filesystem().Write(mediumHandle.location, value), "core.medium.write_text"); err != nil {
		return nil, err
	}
	return value, nil
}

func readBytes(arguments ...any) (any, error) {
	mediumHandle, err := expectHandle(arguments, 0, "core.medium.read_bytes")
	if err != nil {
		return nil, err
	}
	if mediumHandle.location == "" {
		if mediumHandle.data != nil {
			return append([]byte(nil), mediumHandle.data...), nil
		}
		return []byte(mediumHandle.text), nil
	}

	data, err := os.ReadFile(mediumHandle.location)
	if err != nil {
		return nil, core.Wrap(err, "core.medium.read_bytes", "read failed")
	}
	return data, nil
}

func writeBytes(arguments ...any) (any, error) {
	mediumHandle, err := expectHandle(arguments, 0, "core.medium.write_bytes")
	if err != nil {
		return nil, err
	}
	value, err := typemap.ExpectBytes(arguments, 1, "core.medium.write_bytes")
	if err != nil {
		return nil, err
	}

	if mediumHandle.location == "" {
		mediumHandle.data = append([]byte(nil), value...)
		if utf8.Valid(value) {
			mediumHandle.text = string(value)
		} else {
			mediumHandle.text = ""
		}
		return append([]byte(nil), value...), nil
	}

	if err := ensureParentDir(mediumHandle.location, "core.medium.write_bytes"); err != nil {
		return nil, err
	}
	if err := os.WriteFile(mediumHandle.location, value, 0644); err != nil {
		return nil, core.Wrap(err, "core.medium.write_bytes", "write failed")
	}
	return append([]byte(nil), value...), nil
}

func expectHandle(arguments []any, index int, functionName string) (*handle, error) {
	if index >= len(arguments) {
		return nil, core.E(functionName, "expected medium handle", nil)
	}
	mediumHandle, ok := arguments[index].(*handle)
	if !ok {
		return nil, core.E(functionName, "expected medium handle", nil)
	}
	return mediumHandle, nil
}

func ensureParentDir(path, functionName string) error {
	parentDirectory := filepath.Dir(path)
	if parentDirectory == "." || parentDirectory == "" {
		return nil
	}
	_, err := typemap.ResultValue(filesystem().EnsureDir(parentDirectory), functionName)
	return err
}

func filesystem() *core.Fs {
	return (&core.Fs{}).NewUnrestricted()
}
