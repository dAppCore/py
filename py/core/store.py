"""core.store — SQLite KV and workspace primitives.

    available()
"""


def available() -> bool:
    """Return whether the native Store binding is active.

    available()
    """

    return False


__all__ = ["available"]
