# core/py

Python binding for Core primitives. Polyglot repo: Go host + Python userland.

**Module:** `dappco.re/go/py` (Go host)
**Python package:** `core` (under `py/core/`)
**Spec:** `plans/code/core/py/RFC.md`

## Architecture

- **Tier 1:** gpython-embedded — pure-Go Python interpreter, no host CPython
- **Tier 2:** CPython-via-uv — managed subprocess for C extensions / 3.14 features

## Key conventions

- Python imports `core.*` (fs, json, medium, options, process, service, math)
- Bindings live under `bindings/` per-primitive
- `runtime/interpreter.go` is the Go host entrypoint
- Math primitives wrap Poindexter (no NumPy)
- Type conversion in `bindings/typemap/`

## Status

Bootstrap. See `plans/code/core/py/RFC.md` for the implementation roadmap.
