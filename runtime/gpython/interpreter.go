//go:build gpython

// Package gpython is reserved for the real LetheanNetwork/gpython backend.
//
// The pass-2 runtime selector deliberately returns BackendNotBuiltError for
// Options{Backend: "gpython"} until this package wires py.RegisterModule,
// py.METH_VARARGS, py.METH_KEYWORDS, and the bindings/typemap conversion layer
// to the forked interpreter.
package gpython
