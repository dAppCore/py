# core/py — Python Binding for Core Primitives

The fourth corner of the polyglot primitive stack. Python code imports `core`
the way Go code imports `core/go`: same primitive names, same import paths,
different syntax surface.

## Current Implementation

- `runtime/` contains a bootstrap Tier 1 interpreter that validates the CorePy
  module contract and import shape without waiting on the gpython dependency.
- `bindings/` contains Go-backed bindings for `core.echo`, `core.fs`,
  `core.json`, `core.options`, `core.config`, `core.data`, `core.service`,
  `core.log`, and `core.err`.
- `py/core/` contains the Python package surface for the RFC v1 modules,
  including docstrings and concrete fallbacks for CPython validation.

## Validation

```bash
GOWORK=off go test ./...
PYTHONPATH=py python3 -m unittest discover -s py/tests -v
```

## Layout

| Path | Purpose |
|------|---------|
| `bindings/` | Go-side primitive bindings and type conversion helpers |
| `runtime/` | Tier 1 bootstrap interpreter and integration tests |
| `py/core/` | Python package surface for `core.*` modules |
| `py/tests/` | Python package validation |
| `examples/` | Example CorePy programs |

## Spec

`/home/claude/Code/core/plans/code/core/py/RFC.md`
