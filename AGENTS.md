# AGENTS.md

## Repo Overview

- Module: `dappco.re/go/py`
- Purpose: Python binding for Core primitives, with a Go-backed Tier 1 bootstrap runtime and a CPython package surface under `py/core/`.
- Main surfaces: `bindings/`, `runtime/`, `py/core/`, `py/tests/`, and `examples/`.

## First Reads

- Read `README.md` first for the current layout, supported modules, and validation commands.
- Read `CLAUDE.md` for the architecture split between Tier 1 and Tier 2.
- Read `runtime/interpreter.go` before changing import behaviour, module registration, or the Tier 1 execution contract.
- Read `py/tests/test_core.py` before changing observable Python behaviour, since it documents the current public surface well.
- The README references an RFC outside this repo; if that spec is unavailable, trust the current implementation and tests.

## Working Rules

- Keep changes narrow and directly related to the task.
- Preserve parity between the Go bindings/runtime contract and the Python package surface when changing primitive names, signatures, or import paths.
- Prefer current code patterns over broad refactors or speculative abstractions.
- Update nearby tests and docs when behaviour changes.
- Avoid adding new Go or Python dependencies unless they are clearly required.
- Preserve any user changes already present in the worktree.
- Do not commit or rewrite history unless the user asks.

## Project Layout

- `bindings/`: Go-backed primitive bindings such as `config`, `data`, `echo`, `err`, `fs`, `json`, `log`, `math`, `medium`, `options`, `path`, `process`, `service`, and `strings`.
- `bindings/typemap/`: Go-to-Python value conversion helpers used by the binding layer.
- `runtime/`: Tier 1 bootstrap interpreter used to validate module registration, import shape, and simple execution.
- `py/core/`: Python package surface for `core.*`, including docstrings and CPython fallbacks.
- `py/core/math/`: Math submodules that must remain importable as `core.math.kdtree`, `core.math.knn`, and `core.math.signal`.
- `py/tests/`: CPython validation for the package surface.
- `examples/`: Example CorePy programs.

## Coding Conventions

- Match the style of the files you touch; keep Go and Python code straightforward and local.
- Keep public import paths stable: Python users import `core` and `core.*`.
- Maintain the existing module shape where both object-oriented and module-level helpers are exposed, such as in `config`, `data`, `medium`, `options`, and `service`.
- Prefer example-driven doc comments and docstrings where the surrounding file already uses them.
- Keep the Python package typed and consistent with the `pyproject.toml` target of Python 3.12+.
- Avoid unrelated renames, formatting-only churn, and comment rewrites.
- Leave the local `replace dappco.re/go/core => ../go` setup in `go.mod` alone unless the task explicitly requires changing dependency wiring.

## Testing

- Start with the narrowest relevant test surface, then widen scope.
- Run `GOWORK=off go test ./...` after Go-side changes in `bindings/` or `runtime/`.
- Run `PYTHONPATH=py python3 -m unittest discover -s py/tests -v` after Python package changes.
- Run both validation commands when changing cross-surface behaviour or public module contracts.
- Use `py/tests/test_core.py` as the reference for expected package parity and import behaviour.

## Notes for Agents

- This repo is intentionally bootstrap-oriented: the runtime validates the binding contract before the full embedded Python story lands.
- When docs and code disagree, trust the current implementation and tests, then update the docs if needed.
- If you discover a stable repo convention while working, update this file so future agents inherit it.
