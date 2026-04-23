"""Array helpers mirroring Core's typed slice primitive.

from core import array

values = array.new("a", "b")
array.add(values, "c")
"""

from __future__ import annotations

import builtins
from typing import Any


class Array:
    """Mutable ordered collection with Core-shaped helpers.

    values = array.Array("a", "b")
    """

    def __init__(self, *items: Any) -> None:
        self._items = list(items)

    def add(self, *values: Any) -> None:
        """Append values in order.

        values.add("c", "d")
        """

        self._items.extend(values)

    def add_unique(self, *values: Any) -> None:
        """Append only values that are not already present.

        values.add_unique("c", "d")
        """

        for value in values:
            if not self.contains(value):
                self._items.append(value)

    def contains(self, value: Any) -> bool:
        """Return True when the value is present.

        values.contains("c")
        """

        return any(item == value for item in self._items)

    def remove(self, value: Any) -> None:
        """Remove the first matching value when present.

        values.remove("b")
        """

        for index, item in enumerate(self._items):
            if item == value:
                del self._items[index]
                return

    def deduplicate(self) -> None:
        """Remove duplicate values while preserving order.

        values.deduplicate()
        """

        unique_items: list[Any] = []
        for item in self._items:
            if any(existing == item for existing in unique_items):
                continue
            unique_items.append(item)
        self._items = unique_items

    def len(self) -> int:
        """Return the number of stored values.

        values.len()
        """

        return builtins.len(self._items)

    def clear(self) -> None:
        """Remove all values.

        values.clear()
        """

        self._items.clear()

    def as_list(self) -> list[Any]:
        """Return a shallow list copy.

        values.as_list()
        """

        return list(self._items)


def new(*items: Any) -> Array:
    """Create a new Array handle.

    array.new("a", "b")
    """

    return Array(*items)


def add(array_value: Array, *values: Any) -> Array:
    """Append values and return the handle.

    array.add(values, "c")
    """

    array_value.add(*values)
    return array_value


def add_unique(array_value: Array, *values: Any) -> Array:
    """Append only missing values and return the handle.

    array.add_unique(values, "c")
    """

    array_value.add_unique(*values)
    return array_value


def contains(array_value: Array, value: Any) -> bool:
    """Return True when the value exists in the handle.

    array.contains(values, "c")
    """

    return array_value.contains(value)


def remove(array_value: Array, value: Any) -> Array:
    """Remove the first matching value and return the handle.

    array.remove(values, "b")
    """

    array_value.remove(value)
    return array_value


def deduplicate(array_value: Array) -> Array:
    """Remove duplicates and return the handle.

    array.deduplicate(values)
    """

    array_value.deduplicate()
    return array_value


def len(array_value: Array) -> int:
    """Return the number of stored values.

    array.len(values)
    """

    return array_value.len()


def clear(array_value: Array) -> Array:
    """Clear the handle and return it.

    array.clear(values)
    """

    array_value.clear()
    return array_value


def as_list(array_value: Array) -> list[Any]:
    """Return a shallow list copy.

    array.as_list(values)
    """

    return array_value.as_list()


__all__ = [
    "Array",
    "add",
    "add_unique",
    "as_list",
    "clear",
    "contains",
    "deduplicate",
    "len",
    "new",
    "remove",
]
