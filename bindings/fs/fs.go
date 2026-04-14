package fs

import (
	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes filesystem bindings backed by core.Fs.
//
//	fs.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.fs",
		Documentation: "Filesystem primitives backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"read_file":  readFile,
			"write_file": writeFile,
			"ensure_dir": ensureDir,
			"temp_dir":   tempDir,
		},
	})
}

func readFile(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "core.fs.read_file")
	if err != nil {
		return nil, err
	}
	return typemap.ResultValue(filesystem().Read(path), "core.fs.read_file")
}

func writeFile(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "core.fs.write_file")
	if err != nil {
		return nil, err
	}
	content, err := typemap.ExpectString(arguments, 1, "core.fs.write_file")
	if err != nil {
		return nil, err
	}
	if _, err := typemap.ResultValue(filesystem().Write(path, content), "core.fs.write_file"); err != nil {
		return nil, err
	}
	return path, nil
}

func ensureDir(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "core.fs.ensure_dir")
	if err != nil {
		return nil, err
	}
	if _, err := typemap.ResultValue(filesystem().EnsureDir(path), "core.fs.ensure_dir"); err != nil {
		return nil, err
	}
	return path, nil
}

func tempDir(arguments ...any) (any, error) {
	prefix := "corepy-"
	if len(arguments) > 0 {
		var err error
		prefix, err = typemap.ExpectString(arguments, 0, "core.fs.temp_dir")
		if err != nil {
			return nil, err
		}
	}
	return filesystem().TempDir(prefix), nil
}

func filesystem() *core.Fs {
	return (&core.Fs{}).NewUnrestricted()
}
