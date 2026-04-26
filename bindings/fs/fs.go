package fs

import (
	"os"
	"path/filepath" // AX-6-exception: byte-write helper needs parent directory resolution for local files.

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes filesystem bindings backed by core.Fs.
//
//	fs.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.fs",
		Documentation: "Filesystem primitives backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"read_file":   readFile,
			"read_bytes":  readBytes,
			"write_file":  writeFile,
			"write_bytes": writeBytes,
			"ensure_dir":  ensureDir,
			"temp_dir":    tempDir,
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

func readBytes(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "core.fs.read_bytes")
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, core.Wrap(err, "core.fs.read_bytes", "read failed")
	}
	return content, nil
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

func writeBytes(arguments ...any) (any, error) {
	path, err := typemap.ExpectString(arguments, 0, "core.fs.write_bytes")
	if err != nil {
		return nil, err
	}
	content, err := typemap.ExpectBytes(arguments, 1, "core.fs.write_bytes")
	if err != nil {
		return nil, err
	}

	if _, err := typemap.ResultValue(filesystem().EnsureDir(filepath.Dir(path)), "core.fs.write_bytes"); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		return nil, core.Wrap(err, "core.fs.write_bytes", "write failed")
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
