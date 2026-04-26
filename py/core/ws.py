"""core.ws — WebSocket primitives.

    available()
"""


def available() -> bool:
    """Return whether the native WebSocket binding is active.

    available()
    """

    return False


__all__ = ["available"]
