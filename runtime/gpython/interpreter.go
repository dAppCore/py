//go:build gpython

// Package gpython owns the build-tagged Tier 1 backend boundary.
package gpython

import (
	"dappco.re/go/py/runtime/bootstrap"
	"dappco.re/go/py/runtime/internal/contract"
)

// Interpreter is the gpython backend shell.
//
// TODO(corepy-gpython): replace the bootstrap delegate with
// LetheanNetwork/gpython and the py.RegisterModule adapter once the fork is
// vendored. Keeping this package behind the gpython build tag makes that swap a
// package-local change instead of a public runtime contract change.
type Interpreter struct {
	delegate *bootstrap.Interpreter
}

// New creates the build-tagged gpython backend shell.
func New(modules []contract.Module) (*Interpreter, error) {
	delegate := bootstrap.New()
	for _, module := range modules {
		if err := delegate.RegisterModule(module); err != nil {
			_ = delegate.Close()
			return nil, err
		}
	}
	return &Interpreter{delegate: delegate}, nil
}

// Run executes source through the current gpython shell.
func (interpreter *Interpreter) Run(source string) (string, error) {
	return interpreter.delegate.Run(source)
}

// RegisterModule registers a CorePy module.
func (interpreter *Interpreter) RegisterModule(module contract.Module) error {
	return interpreter.delegate.RegisterModule(module)
}

// Close releases interpreter resources.
func (interpreter *Interpreter) Close() error {
	return interpreter.delegate.Close()
}

// Modules returns registered module names.
func (interpreter *Interpreter) Modules() []string {
	return interpreter.delegate.Modules()
}

// Call invokes a registered binding directly.
func (interpreter *Interpreter) Call(moduleName, functionName string, arguments ...any) (any, error) {
	return interpreter.delegate.Call(moduleName, functionName, arguments...)
}
