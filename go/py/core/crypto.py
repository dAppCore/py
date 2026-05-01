"""Cryptographic helpers for hashing, signing, and encoding.

from core import crypto

digest = crypto.sha256("hello")
"""

from __future__ import annotations

import base64
import hashlib
import hmac
import secrets


def sha1(value: bytes | str) -> str:
    """Return the SHA-1 hex digest of a value.

    crypto.sha1("hello")
    """

    return hashlib.sha1(_bytes(value)).hexdigest()


def sha256(value: bytes | str) -> str:
    """Return the SHA-256 hex digest of a value.

    crypto.sha256("hello")
    """

    return hashlib.sha256(_bytes(value)).hexdigest()


def hmac_sha256(key: bytes | str, value: bytes | str) -> str:
    """Return the HMAC-SHA256 hex digest of a value.

    crypto.hmac_sha256("secret", "hello")
    """

    return hmac.new(_bytes(key), _bytes(value), hashlib.sha256).hexdigest()


def compare_digest(left: bytes | str, right: bytes | str) -> bool:
    """Return True when two values match in constant time.

    crypto.compare_digest("a", "a")
    """

    return hmac.compare_digest(_bytes(left), _bytes(right))


def base64_encode(value: bytes | str) -> str:
    """Return a Base64-encoded ASCII string.

    crypto.base64_encode("hello")
    """

    return base64.b64encode(_bytes(value)).decode("ascii")


def base64_decode(value: str) -> bytes:
    """Decode a Base64 string into bytes.

    crypto.base64_decode("aGVsbG8=")
    """

    return base64.b64decode(value.encode("ascii"))


def random_bytes(size: int) -> bytes:
    """Return cryptographically random bytes.

    crypto.random_bytes(16)
    """

    if size < 0:
        raise ValueError("size must be zero or positive")
    return secrets.token_bytes(size)


def _bytes(value: bytes | str) -> bytes:
    if isinstance(value, bytes):
        return value
    return value.encode("utf-8")


__all__ = [
    "base64_decode",
    "base64_encode",
    "compare_digest",
    "hmac_sha256",
    "random_bytes",
    "sha1",
    "sha256",
]
