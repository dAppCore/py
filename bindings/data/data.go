package data

import (
	"io/fs"
	"os"
	"sort"
	"strings" // AX-6-exception: bootstrap data path normalization keeps stdlib contains until the binding is gpython-native.

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
			"mount":       mountPath,
			"mount_path":  mountPath,
			"read_file":   readFile,
			"read_string": readString,
			"list":        list,
			"list_names":  listNames,
			"extract":     extract,
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
	return typemap.ResultValue(data.ReadString(normalizePath(path)), "core.data.read_string")
}

func readFile(arguments ...any) (any, error) {
	data, err := typemap.ExpectData(arguments, 0, "core.data.read_file")
	if err != nil {
		return nil, err
	}
	path, err := typemap.ExpectString(arguments, 1, "core.data.read_file")
	if err != nil {
		return nil, err
	}
	return typemap.ResultValue(data.ReadFile(normalizePath(path)), "core.data.read_file")
}

func list(arguments ...any) (any, error) {
	data, err := typemap.ExpectData(arguments, 0, "core.data.list")
	if err != nil {
		return nil, err
	}
	path, err := typemap.ExpectString(arguments, 1, "core.data.list")
	if err != nil {
		return nil, err
	}

	value, err := typemap.ResultValue(data.List(normalizePath(path)), "core.data.list")
	if err != nil {
		return nil, err
	}

	entries := value.([]fs.DirEntry)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
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
	return typemap.ResultValue(data.ListNames(normalizePath(path)), "core.data.list_names")
}

func extract(arguments ...any) (any, error) {
	data, err := typemap.ExpectData(arguments, 0, "core.data.extract")
	if err != nil {
		return nil, err
	}
	path, err := typemap.ExpectString(arguments, 1, "core.data.extract")
	if err != nil {
		return nil, err
	}
	targetDirectory, err := typemap.ExpectString(arguments, 2, "core.data.extract")
	if err != nil {
		return nil, err
	}

	var templateData any
	if len(arguments) > 3 {
		templateData = arguments[3]
	}

	if _, err := typemap.ResultValue(data.Extract(normalizePath(path), targetDirectory, templateData), "core.data.extract"); err != nil {
		return nil, err
	}
	return targetDirectory, nil
}

func mounts(arguments ...any) (any, error) {
	data, err := typemap.ExpectData(arguments, 0, "core.data.mounts")
	if err != nil {
		return nil, err
	}
	return data.Mounts(), nil
}

func normalizePath(path string) string {
	if path == "" || strings.Contains(path, "/") {
		return path
	}
	return path + "/."
}
