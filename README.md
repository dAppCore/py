# core/py — Python Binding for Core Primitives

The fourth corner of the polyglot primitive stack. Python code imports `core`
the way Go code imports `core/go`: same primitive names, same import paths,
different syntax surface.

## Current Implementation

- `runtime/` contains a bootstrap Tier 1 interpreter that validates the CorePy
  module contract, import shape, and Python-style list/dict type mapping
  without waiting on the gpython dependency.
- `bindings/` contains Go-backed bindings for the RFC v1 module surface:
  `core.echo`, `core.fs`, `core.json`, `core.medium`, `core.options`,
  `core.path`, `core.process`, `core.config`, `core.data`, `core.service`,
  `core.log`, `core.err`, `core.strings`, `core.array`, `core.registry`,
  `core.info`, `core.entitlement`, `core.action`, `core.task`, `core.i18n`,
  the first `core.math` surface, plus initial RFC coverage for `core.cache`,
  `core.crypto`, `core.dns`, and `core.scm`
  (`mean`, `median`, `variance`, `stdev`, sorting, scaling, signal helpers,
  and the `core.math.kdtree` / `core.math.knn` / `core.math.signal`
  import paths).
- `py/core/` contains the Python package surface for the RFC v1 modules,
  including docstrings, concrete fallbacks for CPython validation, and
  module-level helpers that mirror the Tier 1 binding shape.

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
