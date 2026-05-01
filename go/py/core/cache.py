"""JSON cache helpers with path-shaped keys.

from core import cache

store = cache.new("/tmp/corepy-cache", 300)
cache.set(store, "dns/localhost", {"host": "127.0.0.1"})
"""

from __future__ import annotations

import json
from pathlib import Path, PurePosixPath
import time
from typing import Any


_DEFAULT_TTL_SECONDS = 3600


class Cache:
    """File-backed JSON cache with TTL expiry.

    store = cache.Cache("/tmp/corepy-cache", 300)
    """

    def __init__(self, base_dir: str | Path | None = None, ttl_seconds: int = _DEFAULT_TTL_SECONDS) -> None:
        self._base_dir = Path.cwd() / ".core" / "cache" if base_dir in (None, "") else Path(base_dir)
        self._ttl_seconds = _ttl_value(ttl_seconds)
        self._base_dir.mkdir(parents=True, exist_ok=True)

    @property
    def base_dir(self) -> Path:
        """Return the cache root directory.

        store.base_dir
        """

        return self._base_dir

    def path(self, key: str) -> str:
        """Return the storage path for a cache key.

        store.path("dns/localhost")
        """

        return str(self._path_for_key(key))

    def set(self, key: str, value: Any, ttl_seconds: int | None = None) -> str:
        """Store a JSON-serialisable value under a key.

        store.set("dns/localhost", {"host": "127.0.0.1"})
        """

        ttl_value = self._ttl_seconds if ttl_seconds is None else _ttl_value(ttl_seconds)
        target = self._path_for_key(key)
        target.parent.mkdir(parents=True, exist_ok=True)
        now = time.time()
        target.write_text(
            json.dumps(
                {
                    "data": value,
                    "cached_at": now,
                    "expires_at": now + ttl_value,
                },
                indent=2,
                sort_keys=True,
            ),
            encoding="utf-8",
        )
        return str(target)

    def get(self, key: str, default: Any = None) -> Any:
        """Return the cached value or the provided default.

        store.get("dns/localhost", {})
        """

        entry = self._load_entry(key)
        if entry is None:
            return default
        return entry["data"]

    def has(self, key: str) -> bool:
        """Return True when an unexpired cache entry exists.

        store.has("dns/localhost")
        """

        return self._load_entry(key) is not None

    def delete(self, key: str) -> bool:
        """Delete a cached value when present.

        store.delete("dns/localhost")
        """

        target = self._path_for_key(key)
        if not target.exists():
            return False
        target.unlink()
        return True

    def clear(self, prefix: str = "") -> int:
        """Delete all keys that match a prefix.

        store.clear("dns")
        """

        removed = 0
        for key in self.keys(prefix):
            if self.delete(key):
                removed += 1
        return removed

    def keys(self, prefix: str = "") -> list[str]:
        """List stored keys, optionally under a prefix.

        store.keys("dns")
        """

        normalized_prefix = _prefix(prefix)
        if not self._base_dir.exists():
            return []

        keys: list[str] = []
        for target in sorted(self._base_dir.rglob("*.json")):
            if not target.is_file():
                continue
            key = target.relative_to(self._base_dir).as_posix()[:-5]
            if normalized_prefix and key != normalized_prefix and not key.startswith(f"{normalized_prefix}/"):
                continue
            keys.append(key)
        return keys

    def _load_entry(self, key: str) -> dict[str, Any] | None:
        target = self._path_for_key(key)
        if not target.exists():
            return None
        try:
            entry = json.loads(target.read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError):
            return None
        if not isinstance(entry, dict):
            return None
        expires_at = entry.get("expires_at")
        if not isinstance(expires_at, (int, float)) or time.time() > float(expires_at):
            target.unlink(missing_ok=True)
            return None
        return entry

    def _path_for_key(self, key: str) -> Path:
        return self._base_dir.joinpath(*_key_parts(key)).with_suffix(".json")


def new(base_dir: str | Path | None = None, ttl_seconds: int = _DEFAULT_TTL_SECONDS) -> Cache:
    """Create a cache handle.

    cache.new("/tmp/corepy-cache", 300)
    """

    return Cache(base_dir, ttl_seconds)


def path(cache_value: Cache, key: str) -> str:
    """Return the cache storage path for a key.

    cache.path(store, "dns/localhost")
    """

    return cache_value.path(key)


def set(cache_value: Cache, key: str, value: Any) -> str:
    """Store a cache entry with the default TTL.

    cache.set(store, "dns/localhost", {"host": "127.0.0.1"})
    """

    return cache_value.set(key, value)


def set_with_ttl(cache_value: Cache, key: str, value: Any, ttl_seconds: int) -> str:
    """Store a cache entry with an explicit TTL.

    cache.set_with_ttl(store, "dns/localhost", {"host": "127.0.0.1"}, 60)
    """

    return cache_value.set(key, value, ttl_seconds)


def get(cache_value: Cache, key: str, default: Any = None) -> Any:
    """Return a cache entry or the default value.

    cache.get(store, "dns/localhost", {})
    """

    return cache_value.get(key, default)


def has(cache_value: Cache, key: str) -> bool:
    """Return True when a key exists and has not expired.

    cache.has(store, "dns/localhost")
    """

    return cache_value.has(key)


def delete(cache_value: Cache, key: str) -> bool:
    """Delete a cache key when present.

    cache.delete(store, "dns/localhost")
    """

    return cache_value.delete(key)


def clear(cache_value: Cache, prefix: str = "") -> int:
    """Delete cache keys by prefix.

    cache.clear(store, "dns")
    """

    return cache_value.clear(prefix)


def keys(cache_value: Cache, prefix: str = "") -> list[str]:
    """List cache keys, optionally filtered by prefix.

    cache.keys(store, "dns")
    """

    return cache_value.keys(prefix)


def _key_parts(key: str) -> list[str]:
    parts = _parts(key, allow_empty=False, field_name="cache key")
    if not parts:
        raise ValueError("cache key must not be empty")
    return parts


def _prefix(prefix: str) -> str:
    return "/".join(_parts(prefix, allow_empty=True, field_name="cache prefix"))


def _parts(value: str, *, allow_empty: bool, field_name: str) -> list[str]:
    raw = str(value).strip().replace("\\", "/")
    if raw == "":
        if allow_empty:
            return []
        raise ValueError(f"{field_name} must not be empty")
    path_value = PurePosixPath(raw)
    if path_value.is_absolute():
        raise ValueError(f"{field_name} must be relative")
    parts: list[str] = []
    for part in path_value.parts:
        if part in {"", "."}:
            continue
        if part == "..":
            raise ValueError(f"{field_name} must not contain '..'")
        parts.append(part)
    if not parts and not allow_empty:
        raise ValueError(f"{field_name} must not be empty")
    return parts


def _ttl_value(ttl_seconds: int) -> int:
    ttl_value = int(ttl_seconds)
    if ttl_value < 0:
        raise ValueError("ttl_seconds must be zero or positive")
    if ttl_value == 0:
        return _DEFAULT_TTL_SECONDS
    return ttl_value


__all__ = [
    "Cache",
    "clear",
    "delete",
    "get",
    "has",
    "keys",
    "new",
    "path",
    "set",
    "set_with_ttl",
]
