"""Typed option primitives with AX-style examples.

from core import options

opts = options.Options({"name": "corepy", "port": 8080})
opts.set("debug", True)
"""

from __future__ import annotations

import builtins
from dataclasses import dataclass
from typing import Any, Iterable, Mapping


@dataclass(slots=True)
class Option:
    """Single key-value pair.

    options.Option("name", "corepy")
    """

    key: str
    value: Any


class Options:
    """Core-shaped collection of key-value options.

    opts = options.Options({"name": "corepy"})
    """

    def __init__(self, values: Mapping[str, Any] | Iterable[Option] | None = None) -> None:
        self._items: dict[str, Any] = {}
        if values is None:
            return
        if isinstance(values, Mapping):
            for key, value in values.items():
                self._items[str(key)] = value
            return
        for item in values:
            self._items[item.key] = item.value

    def set(self, key: str, value: Any) -> None:
        """Add or replace an option.

        opts.set("port", 8080)
        """

        self._items[key] = value

    def get(self, key: str, default: Any = None) -> Any:
        """Return an option value or the provided default.

        opts.get("name")
        """

        return self._items.get(key, default)

    def has(self, key: str) -> bool:
        """Return True when the option exists.

        opts.has("debug")
        """

        return key in self._items

    def string(self, key: str) -> str:
        """Return a string value or an empty string.

        opts.string("name")
        """

        value = self.get(key, "")
        return value if isinstance(value, str) else ""

    def int(self, key: str) -> int:
        """Return an integer value or zero.

        opts.int("port")
        """

        value = self.get(key, 0)
        return value if isinstance(value, builtins.int) and not isinstance(value, builtins.bool) else 0

    def bool(self, key: str) -> bool:
        """Return a boolean value or False.

        opts.bool("debug")
        """

        value = self.get(key, False)
        return value if isinstance(value, builtins.bool) else False

    def items(self) -> list[Option]:
        """Return the option items in insertion order.

        opts.items()
        """

        return [Option(key=key, value=value) for key, value in self._items.items()]

    def to_dict(self) -> dict[str, Any]:
        """Return a plain dictionary copy.

        opts.to_dict()
        """

        return dict(self._items)

    def __len__(self) -> int:
        return len(self._items)

    def __contains__(self, key: str) -> bool:
        return self.has(key)


def _coerce(value: Options | Mapping[str, Any]) -> Options:
    if isinstance(value, Options):
        return value
    return Options(value)


def new(values: Mapping[str, Any] | Iterable[Option] | None = None) -> Options:
    """Create an Options handle.

    options.new({"name": "corepy"})
    """

    return Options(values)


def set(options_value: Options | Mapping[str, Any], key: str, value: Any) -> Options:
    """Set an option on a handle and return it.

    options.set(opts, "port", 8080)
    """

    handle = _coerce(options_value)
    handle.set(key, value)
    return handle


def get(options_value: Options | Mapping[str, Any], key: str) -> Any:
    """Read an option value from a handle.

    options.get(opts, "name")
    """

    return _coerce(options_value).get(key)


def has(options_value: Options | Mapping[str, Any], key: str) -> bool:
    """Return True when an option exists on a handle.

    options.has(opts, "debug")
    """

    return _coerce(options_value).has(key)


def string(options_value: Options | Mapping[str, Any], key: str) -> str:
    """Read a string option from a handle.

    options.string(opts, "name")
    """

    return _coerce(options_value).string(key)


def int(options_value: Options | Mapping[str, Any], key: str) -> builtins.int:
    """Read an integer option from a handle.

    options.int(opts, "port")
    """

    return _coerce(options_value).int(key)


def bool(options_value: Options | Mapping[str, Any], key: str) -> builtins.bool:
    """Read a boolean option from a handle.

    options.bool(opts, "debug")
    """

    return _coerce(options_value).bool(key)


def items(options_value: Options | Mapping[str, Any]) -> dict[str, Any]:
    """Return a plain dictionary copy of the option items.

    options.items(opts)
    """

    return _coerce(options_value).to_dict()
