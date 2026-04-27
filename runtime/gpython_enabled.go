//go:build gpython

package runtime

import "dappco.re/go/py/runtime/gpython"

func newGPython(modules []Module) (Interpreter, error) {
	return gpython.New(modules)
}
