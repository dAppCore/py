"""core.container — container orchestration primitives.

    available()
"""


def available() -> bool:
    """Return whether the native Container binding is active.

    available()
    """

    return False


__all__ = ["available"]
