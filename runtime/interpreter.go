// Package runtime hosts backend-neutral CorePy interpreters.
//
// The default backend remains the bootstrap Tier 1 interpreter until the real
// LetheanNetwork/gpython integration lands behind the `gpython` build tag.
//
//	interp, err := runtime.New(runtime.Options{Backend: "bootstrap"})
//	if err != nil {
//	    return err
//	}
//	defer interp.Close()
//	out, err := interp.Run(`from core import echo; print(echo("hi"))`)
package runtime

import (
	"fmt"
	"strings"

	"dappco.re/go/py/runtime/bootstrap"
	"dappco.re/go/py/runtime/internal/contract"
)

const (
	// BackendBootstrap selects the bootstrap interpreter.
	BackendBootstrap = "bootstrap"
	// BackendGPython selects the planned real gpython interpreter.
	BackendGPython = "gpython"
)

// Function is a Python-callable binding exposed by a module.
type Function = contract.Function

// Module defines a registered CorePy module.
type Module = contract.Module

// Interpreter is the backend-neutral CorePy execution contract.
type Interpreter = contract.Interpreter

// ModuleLister is implemented by backends that can report registered modules.
type ModuleLister = contract.ModuleLister

// DirectCaller is implemented by backends that support direct binding calls.
type DirectCaller = contract.DirectCaller

// KeywordArguments carries Python-style `name=value` arguments for bindings that
// opt into keyword handling.
type KeywordArguments = contract.KeywordArguments

// BoundMethod describes a method resolved from an object handle.
type BoundMethod = contract.BoundMethod

// AttributeResolver exposes Python-style attributes from a Go-backed handle.
type AttributeResolver = contract.AttributeResolver

// ModuleReference is an imported module handle inside the bootstrap runtime.
type ModuleReference = contract.ModuleReference

// Options selects a CorePy runtime backend and optional pre-registered modules.
type Options struct {
	// Backend is "bootstrap" by default. "gpython" is reserved for the real
	// LetheanNetwork/gpython backend and currently reports BackendNotBuiltError.
	Backend string
	Modules []Module
}

// BackendNotBuiltError reports a selected backend that is known but absent from
// the current build.
type BackendNotBuiltError struct {
	Backend  string
	BuildTag string
}

func (err BackendNotBuiltError) Error() string {
	return fmt.Sprintf("runtime.New: backend %q not built; recompile with -tags %s", err.Backend, err.BuildTag)
}

// New creates an interpreter for the requested backend.
func New(opts Options) (Interpreter, error) {
	backend := strings.TrimSpace(opts.Backend)
	if backend == "" {
		backend = BackendBootstrap
	}

	switch backend {
	case BackendBootstrap:
		interpreter := bootstrap.New()
		for _, module := range opts.Modules {
			if err := interpreter.RegisterModule(module); err != nil {
				_ = interpreter.Close()
				return nil, err
			}
		}
		return interpreter, nil
	case BackendGPython:
		return newGPython(opts.Modules)
	default:
		return nil, fmt.Errorf("runtime.New: unknown backend %q", backend)
	}
}

// SplitKeywordArguments separates positional arguments from a trailing
// KeywordArguments payload.
//
//	positional, keywordArguments := runtime.SplitKeywordArguments(arguments)
func SplitKeywordArguments(arguments []any) ([]any, KeywordArguments) {
	if len(arguments) == 0 {
		return nil, nil
	}

	keywordArguments, ok := arguments[len(arguments)-1].(KeywordArguments)
	if !ok {
		return append([]any(nil), arguments...), nil
	}
	return append([]any(nil), arguments[:len(arguments)-1]...), keywordArguments
}
