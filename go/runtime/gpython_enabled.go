//go:build gpython

package runtime

import "github.com/dappcore/py/go/runtime/gpython"

func newGPython(modules []Module) (Interpreter, error) {
	return gpython.New(modules)
}
