---
module: core/py
repo: core/py
lang: python
tier: consumer
depends:
  - code/core/go
  - code/rfc/core/RFC-CORE-008-AGENT-EXPERIENCE
tags:
  - python
  - polyglot
  - gpython
  - core-primitive
  - embedded-interpreter
  - uv
  - poindexter
---

# core/py RFC — Python Binding for Core Primitives

> The fourth corner of the polyglot primitive stack. Python code imports `core`
> the way Go code imports `core/go`. Same primitives, same shape, same tests,
> different syntax surface. Under the hood, Python runs in **gpython** (a pure-Go
> Python interpreter) embedded inside CoreGO — one binary, no host Python
> interpreter required.

**Python package:** `core` (distributed as `dappco.re/py/core` source tree)
**Go host:** `dappco.re/go/py` (embeds gpython + exposes primitives as Python modules)
**Upstream gpython:** `github.com/go-python/gpython` → forked to `LetheanNetwork/gpython` (dev branch)
**Repo:** `core/py` (polyglot — Go host + Python userland)

---

## 1. Summary

CorePy is the Python binding layer of the Core primitive stack, matching
CoreGO (Go), CorePHP (PHP), and CoreTS (TypeScript) as the fourth
language-native surface over the same underlying primitives.

The architectural innovation is that CorePy does **not** embed CPython.
Instead, it embeds **gpython** — a pure-Go implementation of the Python
interpreter maintained by the `go-python` organisation. This means:

- **Single binary distribution** — CoreGO with CorePy embedded is one
  executable. No host Python interpreter, no pip, no venv, no dependency
  hell, no `python3` vs `python3.12` confusion.
- **Goroutine-native concurrency** — Python code running in gpython can
  participate in Go's concurrency model directly. No GIL to fight.
- **CoreGO primitives exposed as Python modules** — `from core import fs`
  calls `c.Fs()`, `from core import json` calls `core.JSONMarshal`.
  Python import paths mirror Go package paths (AX principle 3: *path is
  documentation*). Core's battle-tested Go implementations back every
  Python import.
- **Symmetry with CorePHP** — CorePHP embeds PHP inside CoreGO. CorePy
  embeds Python inside CoreGO. Same pattern, same distribution story,
  same architecture.

CPython's C-extension stdlib modules (`os`, `io`, `json`, `re`, `socket`,
`hashlib`, `subprocess`, `threading`, etc.) are not available in gpython
because they are C, not Python. **This is not a limitation for CorePy
because CoreGO already has Go-native equivalents for all of them,
exposed to Python code via the import binding layer.**

Heavy numerical / ML work (numpy, torch, mlx, transformers) is out of scope
for Tier 1 and runs in a separate host CPython process managed by
`core.Process`. See §4 (Two-Tier Python).

---

## 2. Goals

1. **First-class Python developer experience** — Python devs write
   regular-looking Python against `core.*` primitives and the code feels
   idiomatic. No cgo wrappers visible, no ctypes, no "this is not real
   Python".

2. **Single-binary distribution** — every Core service that embeds CorePy
   can ship Python tooling without the host machine having Python,
   pip, venvs, or any interpreter-management machinery. Critical for
   Lethean Network edge workers on mixed community hardware.

3. **Zero primitive drift** — a bug fix to `core.Fs()` in Go automatically
   propagates to CorePy users because there is only one implementation
   (Go); CorePy is a thin binding layer, not a reimplementation.

4. **Python version target: 3.14** — modern Python syntax and features
   (pattern matching, `|` union types, walrus operator, `typing.Protocol`,
   full dataclasses, async/await) must be supported. gpython as upstream
   is Python 3.4-ish; upgrading gpython's syntax and runtime coverage to
   3.14 compat is **in-scope** for core/py and will be maintained at
   `LetheanNetwork/gpython`.

5. **Two-tier Python split** — Tier 1 (gpython inside CoreGO) for
   application-layer code, config, data transformation, service glue,
   and mathematical work backed by Poindexter. Tier 2 (host CPython via
   `core.Process` subprocess) for heavy ML ecosystem (torch, mlx,
   transformers, training). Both tiers use `import core` for primitive
   bindings; Tier 1 gets them from gpython's extension mechanism, Tier 2
   gets them via a CPython extension module generated with `gopy`.

---

## 3. Non-Goals

- **Not a CPython replacement.** CorePy does not try to run arbitrary
  third-party Python packages from PyPI that depend on C extensions.
  numpy, torch, mlx, transformers, pandas, scipy — these run in Tier 2
  CPython, not Tier 1 gpython.
- **Not a port of CPython's C-extension stdlib.** `os`, `io`, `json`,
  `socket`, `hashlib`, `subprocess`, etc. are not reimplemented in pure
  Python. They are replaced by CoreGO primitives exposed as Python
  modules.
- **Not a single monolithic `core` module.** Each primitive binds as a
  distinct submodule: `core.fs`, `core.json`, `core.process`, `core.medium`,
  `core.service`, `core.options`, `core.config`, `core.data`, etc.
  Matches the structure of `core/go` packages exactly.
- **Not an ML framework.** CorePy is infrastructure, not inference.
  Lemma/LEM tooling uses CorePy for its Core-side plumbing and calls
  Tier 2 CPython subprocess for the actual model work.

---

## 4. Two-Tier Python Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                  CoreGO (Lethean core)                       │
│  Options, Config, Data, Service, Medium, Process, Fs,        │
│  JSON, Net, DB, WebSocket, MCP, Agent, ... + Poindexter      │
└──────────┬───────────────────────────────┬───────────────────┘
           │ embed                         │ subprocess
           ↓                               ↓
  ┌────────────────────┐          ┌────────────────────────┐
  │  Tier 1 (CorePy)   │          │  Tier 2 (host CPython) │
  │                    │          │                        │
  │  gpython           │          │  uv-managed CPython    │
  │  + core.fs         │          │  + numpy, torch, mlx   │
  │  + core.json       │          │  + transformers, peft  │
  │  + core.medium     │          │  + LEM training        │
  │  + core.process    │          │  + heavy ML eval       │
  │  + core.service    │          │                        │
  │  + core.math       │          │  Optionally calls back │
  │    (Poindexter)    │          │  into CoreGO via gopy  │
  │                    │          │  for primitive access. │
  │  Single binary.    │          │                        │
  │  No host Python.   │          │  Needs CPython host.   │
  │  Pure Go substrate.│          │  Managed by core.Process│
  └────────────────────┘          └────────────────────────┘
```

**Tier 1** handles 80% of Lethean's Python-use cases: config loading,
data transformation, service definitions, KNN over embeddings
(Poindexter), stats, signal processing, Medium-backed I/O, HTTP
services, scripts that Core services invoke at runtime. Zero host
dependencies beyond a CoreGO binary.

**Tier 2** handles the remaining 20%: anything that imports numpy,
torch, mlx, or transformers. Runs as a subprocess managed by
`core.Process`, in a uv-managed venv, with output streamed back
through `core.Medium`. LEM training, MLX inference, HF transformers,
benchmarking harnesses all live in Tier 2.

The boundary is clean because Tier 2 dependencies are obvious at the
`import` level — if your script `import torch`, it's Tier 2; if it
doesn't, it's likely Tier 1.

---

## 5. Primitive Bindings

Each Core primitive is exposed as a Python submodule under the `core`
package. Python import paths mirror Go package paths (`code/core/go/fs`
→ `core.fs`, `code/core/go/process` → `core.process`, etc.). All
bindings are implemented in Go via gpython's extension mechanism
(§7.1) and backed by the canonical CoreGO implementations.

### 5.1 Primitive Coverage Map

| Python module | Go source | CPython equivalent | Purpose |
|---|---|---|---|
| `core.options` | `core/go` Options | — | Typed-option primitive (With*, Must*, For[T] replacement) |
| `core.config` | `core/go/config` | `configparser`, `os.environ` | .core/ config backing, env integration |
| `core.data` | `core/go` Data | — | Fractal DTO primitive |
| `core.service` | `core/go` Service | — | Service lifecycle / handler protocol |
| `core.medium` | `core/go/io` Medium | `io`, `fsspec` | Universal transport (local/S3/cube/memory) |
| `core.fs` | `core/go/fs` | `os`, `os.path`, `pathlib` | Filesystem primitives |
| `core.process` | `core/go/process` | `subprocess`, `os.exec` | Process management (mockable) |
| `core.json` | `core/go` JSONMarshal/Unmarshal | `json` | JSON encode/decode |
| `core.strings` | `core/go` string ops | `str` methods, `re` | String matching, trimming, contains |
| `core.path` | `core/go` JoinPath/PathBase | `os.path`, `pathlib` | Path manipulation |
| `core.log` | `core/go` Print/Error | `logging` | Structured logging |
| `core.err` | `core/go` E() | Exception hierarchy | Scoped error construction |
| `core.api` | `core/api` Gin host | `http.server`, `flask`, `fastapi` | REST server + client |
| `core.ws` | `core/go/ws` | `websockets` | WebSocket primitives |
| `core.store` | `core/go/store` | `sqlite3`, `shelve` | SQLite KV + DuckDB workspace |
| `core.dns` | `go-dns` | `dns.resolver` | DNS resolution |
| `core.cache` | `go-cache` | `functools.lru_cache`, `cachetools` | Caching primitives |
| `core.container` | `go-container` | — | Container orchestration |
| `core.scm` | `go-scm` | `git` (via subprocess) | Git operations |
| `core.math` | `Poindexter` | `numpy` (subset), `scipy.stats` | Sort, search, KDTree, KNN, stats, signal |
| `core.crypto` | `snider/Borg` | `hashlib`, `hmac`, `ssl` | Hashing, HMAC, signing, encryption |
| `core.agent` | `core/agent` | — | Agent dispatch + fleet primitives |
| `core.mcp` | `core/mcp` | — | MCP tool protocol |

### 5.2 Binding Conventions

Each binding follows the same shape:

```go
// core/py/bindings/fs/fs.go
package fs

import (
    "github.com/go-python/gpython/py"
    corefs "dappco.re/go/fs"
)

func init() {
    py.RegisterModule(&py.ModuleImpl{
        Info: py.ModuleInfo{
            Name: "core.fs",
            Doc:  "Filesystem primitives backed by core/go/fs",
        },
        Methods: []*py.Method{
            {Name: "read_file",  Method: readFile,  Flags: py.METH_VARARGS},
            {Name: "write_file", Method: writeFile, Flags: py.METH_VARARGS},
            // ...
        },
    })
}

func readFile(self py.Object, args py.Tuple) (py.Object, error) {
    var path py.String
    if err := py.ParseTuple(args, "s", &path); err != nil {
        return nil, err
    }
    data, err := corefs.ReadFile(string(path))  // <-- CoreGO call
    if err != nil {
        return nil, py.ExceptionNewf(py.OSError, "%s", err.Error())
    }
    return py.Bytes(data), nil
}
```

Pattern:
1. `init()` registers a Python module with gpython
2. Each Python-callable function parses its args, calls the
   corresponding CoreGO function, converts the result to a Python
   object, and returns it
3. CoreGO errors map to Python exceptions via `py.ExceptionNewf`
4. Python types map to Go types via gpython's type conversion layer

Binding coverage target for v1: Options, Config, Data, Service,
Medium, Fs, Process, JSON, log, err. Everything else follows the same
pattern and gets added incrementally.

---

## 6. Math via Poindexter

Poindexter (github.com/Snider/Poindexter) provides pure-Go
implementations of the mathematical primitives CorePy needs at Tier 1:

- **Sorting** (ints, floats, strings, custom comparators) — replaces
  `sorted()`, `list.sort`
- **Binary search** — replaces `bisect`
- **KDTree** (Euclidean, Manhattan, Chebyshev, Cosine metrics) —
  replaces `scipy.spatial.KDTree`, `sklearn.neighbors.NearestNeighbors`
- **Generic KNN** — replaces typical embedding-search loops
- **Statistics** — means, medians, variance, distributions
- **Signal processing** — basic filters, transforms
- **Epsilon comparisons** — stable float equality checks
- **Scale operations** — normalisation, rescaling

`core.math` exposes these as a Python module. Example:

```python
from core.math import kdtree, knn, mean, stdev

# Build a KDTree over 1M embeddings
tree = kdtree.build(embeddings, metric="cosine")

# Find 10 nearest neighbours
neighbours = tree.nearest(query_embedding, k=10)

# Descriptive stats
m = mean(scores)
s = stdev(scores)
```

All operations run in pure Go inside gpython. No numpy, no scipy, no
CPython required. For a 2B-parameter model doing embedding search over
its own KV cache or a RAG corpus, this is the entire math surface you
need.

**Out of scope for Poindexter + CorePy:** dense linear algebra (BLAS
operations, matrix multiplication, SVD, eigendecomposition), FFT,
differential equation solvers, autograd. Anything that wants `A @ B`
on large tensors runs in Tier 2 (numpy/torch/mlx).

---

## 7. gpython Fork — Upgrading to Python 3.14

Upstream gpython targets Python 3.4-ish. That is not acceptable for
CorePy. Modern Python features that must work:

- **Pattern matching** (`match`/`case`, Python 3.10+)
- **Union syntax** (`int | str`, Python 3.10+)
- **Walrus operator** (`:=`, Python 3.8+)
- **f-string improvements** (3.12+)
- **Type hints** (full `typing` module: `Protocol`, `TypedDict`, `Generic`, etc.)
- **Dataclasses** (`@dataclass`, Python 3.7+)
- **Async/await** (full coroutine support, 3.5+)
- **`asyncio` semantics** mapped to Go goroutines where possible

**Scope:** fork gpython to `LetheanNetwork/gpython`, work on `dev`
branch, upstream non-Lethean-specific fixes back to
`github.com/go-python/gpython`. Follow the Lethean fork-and-maintain
pattern established with torchax, optax, CommonLoopUtils, and the rest
of the Google ML stack.

**Milestones:**

- **gpy-0.1:** gpython dev branch fork exists, builds, tests pass.
  Python 3.4 baseline confirmed working inside CoreGO.
- **gpy-0.2:** Pattern matching (`match`/`case`) and union syntax
  (`int | str`) working. Enough for typed primitive bindings.
- **gpy-0.3:** Full `typing` module (`Protocol`, `Generic`, `TypedDict`,
  `dataclass` decorator).
- **gpy-0.4:** Async/await with goroutine-backed event loop.
- **gpy-1.0:** Python 3.14 syntax parity. All CorePy primitive
  bindings testable from gpython without syntax-compat warnings.

### 7.1 Extension Mechanism

gpython's extension mechanism is C-API-like but Go-native. Each binding
module registers with `py.RegisterModule()` and provides `METH_VARARGS`
/ `METH_KEYWORDS` / `METH_NOARGS` methods. CoreGO types convert to
Python objects via a type-mapping layer defined in
`core/py/bindings/typemap/`.

The type-mapping layer handles:
- Go primitives (`int`, `float64`, `string`, `bool`, `[]byte`) ↔ Python primitives
- Go slices/maps ↔ Python lists/dicts
- Go structs ↔ Python classes (via dataclass-style wrapping)
- Go errors ↔ Python exceptions (scoped via `core.E`)
- Go channels ↔ Python async generators / queues
- Go interfaces ↔ Python protocols

---

## 8. Distribution & Packaging

### 8.1 Source Tree

```
core/py/
├── RFC.md                           (this file)
├── bindings/                        (Go-side primitive bindings)
│   ├── fs/
│   ├── json/
│   ├── medium/
│   ├── options/
│   ├── process/
│   ├── service/
│   ├── math/                        (Poindexter wrappers)
│   └── typemap/                     (Go ↔ Python type conversion)
├── runtime/                         (gpython host integration)
│   └── interpreter.go
├── py/                              (Python-side package; installable via uv)
│   ├── pyproject.toml
│   ├── core/
│   │   ├── __init__.py
│   │   ├── fs.py                    (type stubs + docstrings for IDE support)
│   │   ├── json.py
│   │   └── ...
│   └── tests/
├── examples/
└── README.md
```

### 8.2 Tier 1 Distribution (gpython-embedded)

Tier 1 CorePy ships as part of any CoreGO binary that imports
`dappco.re/go/py`. There is no separate install step — the Python
package is embedded in the binary at build time.

Running Python code:

```go
// Go host
interp := py.New()
interp.Run(`
    from core import fs, json
    data = fs.read_file("/etc/config.yaml")
    print(json.loads(data))
`)
```

### 8.3 Tier 2 Distribution (CPython-via-uv)

For Tier 2 use, `core` is installable as a regular Python package via
uv using the `#subdirectory` pip URL trick:

```bash
uv pip install "core @ git+ssh://git@forge.lthn.ai:2223/core#subdirectory=py/core"
```

This installs a CPython extension module built with `gopy` that wraps
the same Go primitives. The Python-side API is identical to Tier 1,
so code written for Tier 1 runs on Tier 2 without changes (modulo Tier
2 also being able to import numpy/torch/mlx if it needs them).

**This is the key symmetry:** the same `import core` works in both
tiers. Tier 1 is faster for pure-Core work because there's no
subprocess / IPC boundary; Tier 2 is needed when you want numpy
alongside `core`.

---

## 9. First-Test Milestone

Before building any primitive bindings, validate the embedding itself:

1. Clone gpython to `LetheanNetwork/gpython`, dev branch.
2. Create `core/py/runtime/interpreter.go` that embeds gpython in
   CoreGO via `py.New()`.
3. Expose ONE trivial Go function as a Python callable: `core.echo(s)`
   returns `s` unchanged.
4. Write a Go integration test that:
   - Creates a gpython interpreter
   - Runs Python code: `from core import echo; print(echo("hello"))`
   - Asserts the output is `"hello"`
5. Commit and push.

**That's proof of life.** Once the round-trip works, the next primitive
is `Options` (pure data, no I/O, easiest to bind). Each subsequent
primitive follows the same pattern and adds test coverage
(`TestFilename_Function_{Good,Bad,Ugly}` per AX principle 10).

---

## 10. Risks & Mitigations

| Risk | Mitigation |
|---|---|
| gpython's Python version is too old for modern syntax | Fork to `LetheanNetwork/gpython`, upgrade incrementally, upstream fixes. Owned scope (§7). |
| gpython doesn't support some bindings pattern we need | Extension mechanism is sufficient for primitive wrapping (verified by `go-python/gopy` design). Worst case: patch gpython. |
| Small upstream community → bus factor | Maintain `LetheanNetwork/gpython` as the Lethean-authoritative fork. Same pattern as torchax, optax, CommonLoopUtils. |
| Tier 1 users hit a wall where they need numpy | Tier 2 path is clean — rewrite the script for Tier 2 via `core.Process`. Boundary is obvious at the import level. |
| Embedding overhead (Python parse/compile at init) | Bytecode cache + pre-compiled marshal file for frequently-invoked scripts. Same technique CPython uses. |
| Python developers unfamiliar with Core primitives | Type stubs (§8.1 `py/core/*.py`) provide full IDE completion and docstrings. Documentation matches core/go naming by convention. |

---

## 11. Status & Next Steps

| Item | Status |
|---|---|
| RFC draft | 🚧 This document — v0.1 |
| gpython fork | Planned: `LetheanNetwork/gpython` dev branch |
| First-test (echo round-trip) | Planned: milestone gpy-0.1 |
| Primitive binding: Options | Planned: gpy-0.2 |
| Python 3.14 parity | Planned: gpy-1.0 |
| Integration with core/agent | Planned: post gpy-1.0 |
| LEM tooling migration to CorePy | Planned: post gpy-1.0 |

**Next step after this RFC lands:** fork gpython, stand up the
interpreter embedding in a CoreGO binary, do the echo round-trip, commit.
Everything else follows once proof-of-life is green.

---

## 12. References

- **gpython upstream:** https://github.com/go-python/gpython
- **go-python organisation:** https://github.com/go-python (gpython, gopy,
  cpy3, py, setuptools-golang, go2py)
- **Poindexter:** https://github.com/Snider/Poindexter (math primitives)
- **CoreGO:** `~/Code/core/go` (primitive source of truth)
- **CorePHP:** `code/core/php/` (architectural precedent for embedded
  interpreter pattern)
- **CoreTS:** `code/core/ts/` (architectural precedent for polyglot
  primitive exposure)
- **AX principles:** `rfc/core/RFC-CORE-008-AGENT-EXPERIENCE.md`
- **uv:** https://docs.astral.sh/uv/ (Python packaging + venv manager,
  used for Tier 2 distribution)
