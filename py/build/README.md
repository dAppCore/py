# CorePy Tier 2 Build Stub

This directory is the scaffold for the Tier 2 CorePy distribution described in
RFC §8.3: CPython imports `core`, but the module is backed by a native extension
generated with `gopy` from the same Go primitives used by Tier 1.

Tier 2 is for Python programs that need CPython-only libraries such as numpy,
torch, mlx, pandas, or scipy alongside `core`. Tier 1 remains the embedded
gpython path for pure-Core workloads.

## Current State

Pass 2 intentionally does not build the native extension. The Codex sandbox does
not provide `gopy` or the target CPython toolchain, so this package installs as a
documentation stub. Importing `core` raises `NotImplementedError` and points
operators back to this build flow.

## Build Flow

Install prerequisites in the operator build environment:

```bash
go install github.com/go-python/gopy@latest
python3.13 --version
```

Then run:

```bash
./build.sh
```

When `gopy` is present, the wrapper invokes it against the Go binding packages
under `bindings/` and writes generated artifacts under `py/build/dist/gopy/`.
Until the real Tier 2 package lands, generated output is not checked in.
