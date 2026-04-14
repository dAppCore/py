# core/py — Python Binding for Core Primitives

The fourth corner of the polyglot primitive stack. Python code imports `core` the
way Go code imports `core/go`. Same primitives, same shape, same tests, different
syntax surface.

Two tiers:

- **Tier 1 (gpython-embedded):** ships inside any CoreGO binary that imports
  `dappco.re/go/py`. Pure-Go Python interpreter, no host CPython required.
- **Tier 2 (CPython-via-uv):** managed CPython subprocess for code that needs
  C extensions or 3.14 features beyond gpython's coverage.

## Layout

| Path | Purpose |
|------|---------|
| `bindings/` | Go-side primitive bindings (fs, json, medium, options, process, service, math, typemap) |
| `runtime/` | gpython host integration |
| `py/core/` | Python-side package (installable via uv) |
| `py/tests/` | Python test suite |
| `examples/` | Polyglot example programs |

## Spec

`plans/code/core/py/RFC.md` in the spec tree — read first.

## Status

Bootstrap. Empty skeleton ready for factory dispatches.
