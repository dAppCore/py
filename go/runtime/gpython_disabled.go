//go:build !gpython

package runtime

func newGPython(modules []Module) (Interpreter, error) {
	return nil, BackendNotBuiltError{Backend: BackendGPython, BuildTag: BackendGPython}
}
