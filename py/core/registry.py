"""Thread-safe named collection helpers.

from core import registry

items = registry.new()
registry.set(items, "brain", {"enabled": True})
"""

from __future__ import annotations

import builtins
from fnmatch import fnmatchcase
from typing import Any


class Registry:
    """Named collection with insertion order and lock modes.

    items = registry.Registry()
    """

    def __init__(self) -> None:
        self._items: dict[str, Any] = {}
        self._disabled: set[str] = builtins.set()
        self._order: list[str] = []
        self._mode = "open"

    def set(self, name: str, value: Any) -> None:
        """Store or update a named value.

        items.set("brain", value)
        """

        if self._mode == "locked":
            raise RuntimeError(f"registry is locked, cannot set: {name}")
        if self._mode == "sealed" and name not in self._items:
            raise RuntimeError(f"registry is sealed, cannot add new key: {name}")
        if name not in self._items:
            self._order.append(name)
        self._items[name] = value

    def get(self, name: str, default: Any = None) -> Any:
        """Return a named value or the provided default.

        items.get("brain")
        """

        return self._items.get(name, default)

    def has(self, name: str) -> bool:
        """Return True when the name exists.

        items.has("brain")
        """

        return name in self._items

    def names(self) -> list[str]:
        """Return names in insertion order.

        items.names()
        """

        return builtins.list(self._order)

    def list(self, pattern: str) -> list[Any]:
        """Return enabled values whose names match a glob.

        items.list("process.*")
        """

        return [
            self._items[name]
            for name in self._order
            if name not in self._disabled and fnmatchcase(name, pattern)
        ]

    def len(self) -> int:
        """Return the number of stored values.

        items.len()
        """

        return builtins.len(self._items)

    def delete(self, name: str) -> bool:
        """Delete a named value.

        items.delete("brain")
        """

        if self._mode == "locked":
            raise RuntimeError(f"registry is locked, cannot delete: {name}")
        if name not in self._items:
            raise KeyError(name)
        del self._items[name]
        self._disabled.discard(name)
        self._order = [item for item in self._order if item != name]
        return True

    def disable(self, name: str) -> None:
        """Soft-disable a named value.

        items.disable("brain")
        """

        if name not in self._items:
            raise KeyError(name)
        self._disabled.add(name)

    def enable(self, name: str) -> None:
        """Re-enable a soft-disabled value.

        items.enable("brain")
        """

        if name not in self._items:
            raise KeyError(name)
        self._disabled.discard(name)

    def disabled(self, name: str) -> bool:
        """Return True when the named value is disabled.

        items.disabled("brain")
        """

        return name in self._disabled

    def lock(self) -> None:
        """Fully freeze the registry.

        items.lock()
        """

        self._mode = "locked"

    def locked(self) -> bool:
        """Return True when the registry is fully locked.

        items.locked()
        """

        return self._mode == "locked"

    def seal(self) -> None:
        """Disallow new keys while permitting updates.

        items.seal()
        """

        self._mode = "sealed"

    def sealed(self) -> bool:
        """Return True when the registry is sealed.

        items.sealed()
        """

        return self._mode == "sealed"

    def open(self) -> None:
        """Reset the registry to open mode.

        items.open()
        """

        self._mode = "open"


def new() -> Registry:
    """Create a new Registry handle.

    registry.new()
    """

    return Registry()


def set(registry_value: Registry, name: str, value: Any) -> Registry:
    """Store or update a named value and return the handle.

    registry.set(items, "brain", value)
    """

    registry_value.set(name, value)
    return registry_value


def get(registry_value: Registry, name: str, default: Any = None) -> Any:
    """Return a named value or the default.

    registry.get(items, "brain")
    """

    return registry_value.get(name, default)


def has(registry_value: Registry, name: str) -> bool:
    """Return True when the name exists.

    registry.has(items, "brain")
    """

    return registry_value.has(name)


def names(registry_value: Registry) -> list[str]:
    """Return names in insertion order.

    registry.names(items)
    """

    return registry_value.names()


def list(registry_value: Registry, pattern: str) -> list[Any]:
    """Return enabled values that match a glob.

    registry.list(items, "process.*")
    """

    return registry_value.list(pattern)


def len(registry_value: Registry) -> int:
    """Return the number of stored values.

    registry.len(items)
    """

    return registry_value.len()


def delete(registry_value: Registry, name: str) -> bool:
    """Delete a named value.

    registry.delete(items, "brain")
    """

    return registry_value.delete(name)


def disable(registry_value: Registry, name: str) -> Registry:
    """Soft-disable a named value and return the handle.

    registry.disable(items, "brain")
    """

    registry_value.disable(name)
    return registry_value


def enable(registry_value: Registry, name: str) -> Registry:
    """Re-enable a named value and return the handle.

    registry.enable(items, "brain")
    """

    registry_value.enable(name)
    return registry_value


def disabled(registry_value: Registry, name: str) -> bool:
    """Return True when a name is disabled.

    registry.disabled(items, "brain")
    """

    return registry_value.disabled(name)


def lock(registry_value: Registry) -> Registry:
    """Lock the registry and return the handle.

    registry.lock(items)
    """

    registry_value.lock()
    return registry_value


def locked(registry_value: Registry) -> bool:
    """Return True when the registry is locked.

    registry.locked(items)
    """

    return registry_value.locked()


def seal(registry_value: Registry) -> Registry:
    """Seal the registry and return the handle.

    registry.seal(items)
    """

    registry_value.seal()
    return registry_value


def sealed(registry_value: Registry) -> bool:
    """Return True when the registry is sealed.

    registry.sealed(items)
    """

    return registry_value.sealed()


def open(registry_value: Registry) -> Registry:
    """Reset the registry to open mode and return the handle.

    registry.open(items)
    """

    registry_value.open()
    return registry_value


__all__ = [
    "Registry",
    "delete",
    "disable",
    "disabled",
    "enable",
    "get",
    "has",
    "len",
    "list",
    "lock",
    "locked",
    "names",
    "new",
    "open",
    "seal",
    "sealed",
    "set",
]
