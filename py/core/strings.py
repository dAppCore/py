"""String helpers with Core-shaped naming.

from core import strings

strings.contains("hello world", "world")
strings.trim("  corepy  ")
"""

from __future__ import annotations


def contains(value: str, substring: str) -> bool:
    """Return True when the substring exists.

    strings.contains("hello world", "world")
    """

    return substring in value


def trim(value: str) -> str:
    """Trim surrounding whitespace.

    strings.trim("  corepy  ")
    """

    return value.strip()


def trim_prefix(value: str, prefix: str) -> str:
    """Trim a leading prefix when present.

    strings.trim_prefix("--debug", "--")
    """

    return value[len(prefix):] if value.startswith(prefix) else value


def trim_suffix(value: str, suffix: str) -> str:
    """Trim a trailing suffix when present.

    strings.trim_suffix("config.json", ".json")
    """

    return value[:-len(suffix)] if suffix and value.endswith(suffix) else value


def has_prefix(value: str, prefix: str) -> bool:
    """Return True when the value starts with the prefix.

    strings.has_prefix("--debug", "--")
    """

    return value.startswith(prefix)


def has_suffix(value: str, suffix: str) -> bool:
    """Return True when the value ends with the suffix.

    strings.has_suffix("config.json", ".json")
    """

    return value.endswith(suffix)


def split(value: str, separator: str) -> list[str]:
    """Split a string by a separator.

    strings.split("deploy/to/homelab", "/")
    """

    return value.split(separator)


def split_n(value: str, separator: str, limit: int) -> list[str]:
    """Split a string into at most `limit` parts.

    strings.split_n("key=value=extra", "=", 2)
    """

    if limit == 0:
        return []
    if limit < 0:
        return value.split(separator)
    return value.split(separator, limit - 1)


def join(separator: str, *parts: str) -> str:
    """Join parts with a separator.

    strings.join("/", "deploy", "to", "homelab")
    """

    return separator.join(parts)


def replace(value: str, old: str, new: str) -> str:
    """Replace all occurrences of one substring with another.

    strings.replace("deploy/to/homelab", "/", ".")
    """

    return value.replace(old, new)


def lower(value: str) -> str:
    """Return lowercase text.

    strings.lower("HELLO")
    """

    return value.lower()


def upper(value: str) -> str:
    """Return uppercase text.

    strings.upper("hello")
    """

    return value.upper()


def rune_count(value: str) -> int:
    """Return the Unicode code point count.

    strings.rune_count("🔥")
    """

    return len(value)


def concat(*parts: str) -> str:
    """Concatenate string parts without a separator.

    strings.concat("deploy", "/", "to")
    """

    return "".join(parts)
