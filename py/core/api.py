"""core.api — REST server and client primitives.

    available()
"""


def available() -> bool:
    """Return whether the native API binding is active.

    available()
    """

    return False


__all__ = ["available"]
