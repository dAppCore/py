# gpython Readiness Audit

Pass 4 audit for swapping the bootstrap interpreter to the planned
`LetheanNetwork/gpython` backend without rewriting primitive bindings.

## Pass 4 Boundary

The public runtime selector now routes `Options{Backend: "gpython"}` through a
build-tagged factory:

- default builds still return `BackendNotBuiltError`
- `-tags gpython` builds instantiate `runtime/gpython`
- `runtime/gpython` currently delegates to the bootstrap interpreter as a smoke
  shell
- TODO: replace that package-local delegate with the real
  `LetheanNetwork/gpython` fork, `py.RegisterModule`, `py.METH_VARARGS`, and
  `py.METH_KEYWORDS` wiring

This keeps the gpython integration boundary testable without making the default
build depend on a fork that is not vendored here yet.

Pass 4 adds two pieces that should survive the real gpython swap:

- stdlib-shaped shadow modules for `json`, `os`, `os.path`, `subprocess`,
  `logging`, `hashlib`, `base64`, and `socket`; these are registered beside
  `core.*` modules and route common Python imports to Core-backed primitives
- a typed unsupported-import error that the CLI can use for
  `corepy run -tier2-fallback`, retrying the script in Tier 2 CPython only when
  Tier 1 cannot satisfy the import table

The real gpython backend should preserve that import policy: first try native
Tier 1 Core/shadow modules, then surface unsupported imports as fallback
eligible instead of mixing implicit CPython execution into the interpreter.

## Audit Basis

Upstream gpython exposes modules through `py.RegisterModule` / `py.ModuleImpl`
and method functions shaped as `py.PyCFunction`, `py.PyCFunctionNoArgs`, and
`py.PyCFunctionWithKeywords`, where arguments arrive as `py.Tuple` plus optional
`py.StringDict` keyword arguments. `py.ParseTuple`, `py.ParseTupleAndKeywords`,
and `py.UnpackTuple` are the native parse surface.

References:
- https://pkg.go.dev/github.com/go-python/gpython/py
- https://github.com/go-python/gpython

The existing CorePy bindings all use `runtime.Function func(...any) (any,
error)`. That is acceptable if the gpython backend provides one shared adapter:

1. Convert `py.Tuple` / `py.StringDict` to `[]any` plus trailing
   `runtime.KeywordArguments`.
2. Invoke the existing `runtime.Function`.
3. Convert the returned Go value or error back to `py.Object` / Python
   exception.

Bindings marked READY need only that shared adapter. Bindings marked
NEEDS-ADAPTATION need module-specific Python-visible handle, class, callable, or
protocol wrappers in addition to the shared adapter. No current binding is
blocked by an absent gpython feature.

## Summary

| Status | Count | Percent |
|---|---:|---:|
| READY | 17 | 54.8% |
| NEEDS-ADAPTATION | 14 | 45.2% |
| BLOCKED | 0 | 0.0% |

## Per-Binding Audit

| Binding | Status | Notes |
|---|---|---|
| `action` | NEEDS-ADAPTATION | Stores and invokes callables. gpython needs a callable proxy that can hold a `py.Object`, call it with converted context/options, and preserve the existing Go `runtime.Function` path. |
| `agent` | READY | Stub exposes `available() -> bool`; no handle or keyword behavior. |
| `api` | READY | Stub exposes `available() -> bool`; no handle or keyword behavior. |
| `array` | NEEDS-ADAPTATION | Returns an internal Go handle. Needs a Python class/capsule wrapper so subsequent calls can safely recover the Go array handle. |
| `cache` | NEEDS-ADAPTATION | Returns a cache handle and persists JSON-like maps. Needs handle wrapping plus map/list conversion in typemap. |
| `config` | NEEDS-ADAPTATION | Returns `*core.Config`. Needs a Python-visible config handle and method/module-level parity wrappers. |
| `container` | READY | Stub exposes `available() -> bool`; no handle or keyword behavior. |
| `crypto` | READY | Primitive string/bytes/int inputs and string/bytes/bool outputs; suitable for shared tuple/typemap adapter. |
| `data` | NEEDS-ADAPTATION | Returns `*core.Data` and mounts host paths. Needs handle wrapping and careful path/bytes/list conversion. |
| `dns` | READY | String inputs with string/list/int outputs; no persistent handles. |
| `echo` | READY | Single object round-trip; useful as the first gpython adapter proof. |
| `entitlement` | NEEDS-ADAPTATION | Returns a Core entitlement value with Python method parity in shims. Needs struct/class wrapping or stable value conversion. |
| `err` | NEEDS-ADAPTATION | Returns and accepts Go `error` values. Needs scoped Python exception objects that preserve operation, code, root, and wrapping semantics. |
| `fs` | READY | Path strings and bytes/text payloads only; maps cleanly through shared typemap. |
| `i18n` | NEEDS-ADAPTATION | Accepts translator interfaces. Needs Python protocol adapter for `translate`, `set_language`, `language`, and `available_languages`. |
| `info` | READY | No-arg/string inputs and map/list/string outputs; no persistent handles. |
| `json` | READY | JSON text and JSON-like primitives/lists/maps; depends only on shared map/list conversion. |
| `log` | READY | Level/message string calls and variadic key/value payloads; can ride shared tuple conversion. |
| `math` | NEEDS-ADAPTATION | Scalar/list helpers are READY, but `kdtree.build` returns a handle with methods and keyword arguments. Needs class wrapper for KDTree and keyword-aware adapters. |
| `mcp` | READY | Stub exposes `available() -> bool`; no handle or keyword behavior. |
| `medium` | NEEDS-ADAPTATION | Returns file/memory medium handles and reads/writes bytes. Needs handle wrapping and lifetime rules for file-backed state. |
| `options` | NEEDS-ADAPTATION | Returns `*core.Options`. Needs a Python-visible options handle and module-level helper parity. |
| `path` | READY | String/list helpers over path primitives; no persistent handles. |
| `process` | READY | Go-backed process execution lets Tier 1 avoid Python `subprocess`; inputs/outputs are strings, lists, and env maps. |
| `registry` | NEEDS-ADAPTATION | Returns a generic registry handle. Needs Python-visible handle, mutability guard errors, and safe `any` item conversion. |
| `scm` | READY | String path inputs and string/list/map outputs; no persistent handles. |
| `service` | NEEDS-ADAPTATION | Stores services and may call lifecycle interfaces. Needs Python protocol wrappers for startup/shutdown-capable objects. |
| `store` | READY | Stub exposes `available() -> bool`; no handle or keyword behavior. |
| `strings` | READY | String/list/int helpers; no persistent handles. |
| `task` | NEEDS-ADAPTATION | Uses task handles and action registry integration. Needs handle wrappers and callable bridge parity with `action`. |
| `ws` | READY | Stub exposes `available() -> bool`; no handle or keyword behavior. |

## Stdlib Shadow Audit

| Import | Tier 1 backing | Coverage |
|---|---|---|
| `json` | `core.JSONMarshalString` / `core.JSONUnmarshalString` | `dumps`, `loads` |
| `os` | Go `os` plus Core path/process conventions | `getcwd`, `getenv`, `listdir`, `makedirs`, `remove`, `system` |
| `os.path` | Go filepath/Core path semantics | `abspath`, `basename`, `dirname`, `exists`, `isabs`, `join` |
| `subprocess` | Go `os/exec`, matching `core.process` result shape | `check_output`, `getoutput`, `run` |
| `logging` | Core log functions | `basicConfig`, `debug`, `info`, `warning`, `error` |
| `hashlib` | Go crypto hashes | `sha1`, `sha256`, `hexdigest` handles |
| `base64` | Go base64 codec | `b64encode`, `b64decode` |
| `socket` | Go `net` DNS/service lookup | `gethostbyname`, `gethostbyname_ex`, `getservbyname` |

## Estimated Lift

Minimum swap once `LetheanNetwork/gpython` is available:

- Shared gpython backend shell, stdout capture, module registration selector:
  1-2 days.
- Shared `py.Tuple` / `py.StringDict` / `py.Object` typemap adapter for
  primitives, bytes, lists, maps, and errors: 2-4 days.
- Python-visible handle/class wrappers for stateful modules (`array`, `cache`,
  `config`, `data`, `math.kdtree`, `medium`, `options`, `registry`): 3-5 days.
- Callable/protocol adapters for `action`, `task`, `i18n`, and `service`:
  3-5 days.
- Parity pass over examples, CLI, and bootstrap/gpython backend tests: 2-3 days.

Pragmatic estimate: roughly 2 engineering weeks for a reliable gpy-0.2 swap if
the fork already builds, supports the required syntax subset, and does not
require gpython runtime patches beyond module registration and stdout capture.
