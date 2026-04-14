package data

import (
	"os"

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes Data bindings backed by dappco.re/go/core.
//
//	data.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.data",
		Documentation: "Embedded content registry backed by dappco.re/go/core",
		Functions: map[string]runtime.Function{
			"new":         newData,
			"mount_path":  mountPath,
			"read_string": readString,
			"list_names":  listNames,
			"mounts":      mounts,
		},
	})
}

func newData(arguments ...any) (any, error) {
	return &core.Data{Registry: core.NewRegistry[*core.Embed]()}, nil
}

func mountPath(arguments ...any) (any, error) {
	data, err := typemap.ExpectData(arguments, 0, "core.data.mount_path")
	if err != nil {
		return nil, err
	}
	name, err := typemap.ExpectString(arguments, 1, "core.data.mount_path")
	if err != nil {
		return nil, err
	}
	sourceDirectory, err := typemap.ExpectString(arguments, 2, "core.data.mount_path")
	if err != nil {
		return nil, err
	}
	mountPath := "."
	if len(arguments) > 3 {
		mountPath, err = typemap.ExpectString(arguments, 3, "core.data.mount_path")
		if err != nil {
			return nil, err
		}
	}

	options := core.NewOptions(
		core.Option{Key: "name", Value: name},
		core.Option{Key: "source", Value: os.DirFS(sourceDirectory)},
		core.Option{Key: "path", Value: mountPath},
	)
	if _, err := typemap.ResultValue(data.New(options), "core.data.mount_path"); err != nil {
		return nil, err
	}
	return data, nil
}

func readString(arguments ...any) (any, error) {
	data, err := typemap.ExpectData(arguments, 0, "core.data.read_string")
	if err != nil {
		return nil, err
	}
	path, err := typemap.ExpectString(arguments, 1, "core.data.read_string")
	if err != nil {
		return nil, err
	}
	return typemap.ResultValue(data.ReadString(path), "core.data.read_string")
}

func listNames(arguments ...any) (any, error) {
	data, err := typemap.ExpectData(arguments, 0, "core.data.list_names")
	if err != nil {
		return nil, err
	}
	path, err := typemap.ExpectString(arguments, 1, "core.data.list_names")
	if err != nil {
		return nil, err
	}
	return typemap.ResultValue(data.ListNames(path), "core.data.list_names")
}

func mounts(arguments ...any) (any, error) {
	data, err := typemap.ExpectData(arguments, 0, "core.data.mounts")
	if err != nil {
		return nil, err
	}
	return data.Mounts(), nil
}
