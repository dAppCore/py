"""Documentation stub for the future gopy-built CorePy extension."""


class Tier2BuildUnavailableError(NotImplementedError):
    """Raised when the native Tier 2 CorePy extension has not been built."""


def unavailable() -> None:
    raise Tier2BuildUnavailableError(
        "Tier 2 CorePy is not built. Run py/build/build.sh in an environment "
        "with gopy and CPython 3.13+; see py/build/README.md."
    )
