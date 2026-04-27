// Package contract carries the CorePy runtime types shared by backend
// implementations and the public runtime facade.
package contract

// Function is a Python-callable binding exposed by a module.
//
//	module := runtime.Module{
//	    Name: "core",
//	    Functions: map[string]runtime.Function{
//	        "echo": func(arguments ...any) (any, error) { return arguments[0], nil },
//	    },
//	}
type Function func(arguments ...any) (any, error)

// Module defines a registered CorePy module.
//
//	runtime.Module{
//	    Name: "core.fs",
//	    Documentation: "Filesystem primitives",
//	    Functions: map[string]runtime.Function{"read_file": readFile},
//	}
type Module struct {
	Name          string
	Documentation string
	Functions     map[string]Function
}

// Interpreter is the backend-neutral CorePy execution contract.
type Interpreter interface {
	Run(source string) (string, error)
	RegisterModule(m Module) error
	Close() error
}

// ModuleLister is implemented by backends that can report their registered
// module names.
type ModuleLister interface {
	Modules() []string
}

// DirectCaller is implemented by backends that support direct Go invocation of
// registered CorePy functions.
type DirectCaller interface {
	Call(moduleName, functionName string, arguments ...any) (any, error)
}

// KeywordArguments carries Python-style `name=value` arguments for bindings that
// opt into keyword handling.
//
//	bindings := runtime.KeywordArguments{"metric": "cosine", "k": 2}
type KeywordArguments map[string]any

// BoundMethod describes a method resolved from an object handle.
//
//	method := runtime.BoundMethod{ModuleName: "core.math.kdtree", FunctionName: "nearest", Arguments: []any{tree}}
type BoundMethod struct {
	ModuleName   string
	FunctionName string
	Arguments    []any
}

// AttributeResolver exposes Python-style attributes from a Go-backed handle.
//
//	attribute, ok := tree.ResolveAttribute("nearest")
type AttributeResolver interface {
	ResolveAttribute(name string) (any, bool)
}

// ModuleReference is an imported module handle inside the bootstrap runtime.
//
//	from core import fs
//	print(fs.read_file("/tmp/demo.txt"))
type ModuleReference struct {
	Name string
}

// UnsupportedImportError reports an import that the selected Tier 1 backend
// cannot satisfy from its registered module table.
type UnsupportedImportError struct {
	Module string
}

func (err UnsupportedImportError) Error() string {
	if err.Module == "" {
		return "unsupported import"
	}
	return "unsupported import " + err.Module
}
