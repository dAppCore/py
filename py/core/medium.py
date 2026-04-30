"""Simple transport wrapper for memory, filesystem, or Qdrant-backed content.

from core import medium

buffer = medium.memory("hello")
buffer.write_text("updated")
remote = medium.open("qdrant://localhost:6333/core_medium")
"""

from __future__ import annotations

import base64
import binascii
from dataclasses import dataclass
import json as stdlib_json
from pathlib import Path
from typing import Any
from urllib import error as url_error
from urllib import parse as url_parse
from urllib import request as url_request

from . import fs as core_fs


class MediumError(RuntimeError):
    """Raised when a medium backend cannot complete an operation."""


class QdrantError(MediumError):
    """Raised when Qdrant returns an error or an unexpected response."""


@dataclass(slots=True)
class Medium:
    """Text or bytes transport for memory and filesystem targets.

    medium.memory("hello")
    """

    location: str | Path | None = None
    text: str = ""
    data: bytes = b""

    def read_text(self) -> str:
        """Read text from the medium.

        buffer.read_text()
        """

        if self.location is None:
            return self.text
        return Path(self.location).read_text(encoding="utf-8")

    def write_text(self, value: str) -> str:
        """Write text into the medium.

        buffer.write_text("updated")
        """

        if self.location is None:
            self.text = value
            self.data = value.encode("utf-8")
            return value
        path = Path(self.location)
        core_fs.ensure_dir(path.parent)
        path.write_text(value, encoding="utf-8")
        return value

    def read_bytes(self) -> bytes:
        """Read bytes from the medium.

        buffer.read_bytes()
        """

        if self.location is None:
            return self.data if self.data else self.text.encode("utf-8")
        return Path(self.location).read_bytes()

    def write_bytes(self, value: bytes) -> bytes:
        """Write bytes into the medium.

        buffer.write_bytes(b"updated")
        """

        if self.location is None:
            self.data = value
            try:
                self.text = value.decode("utf-8")
            except UnicodeDecodeError:
                self.text = ""
            return value
        path = Path(self.location)
        core_fs.ensure_dir(path.parent)
        path.write_bytes(value)
        return value


@dataclass(slots=True)
class QdrantMedium:
    """Text or bytes transport backed by one Qdrant collection point.

    medium.open("qdrant://localhost:6333/core_medium")
    """

    base_url: str
    collection: str
    point_id: int | str = 1
    payload_field: str = "text"
    bytes_field: str = "data_base64"
    timeout: float = 10.0

    @classmethod
    def from_url(cls, url: str) -> QdrantMedium:
        """Create a Qdrant medium from a qdrant://host:port/collection URL."""

        parsed = url_parse.urlparse(url)
        if parsed.scheme != "qdrant":
            raise ValueError("qdrant medium URLs must use the qdrant scheme")
        if not parsed.hostname:
            raise ValueError("qdrant medium URLs must include a host")
        try:
            parsed_port = parsed.port
        except ValueError as exc:
            raise ValueError("qdrant medium URL has an invalid port") from exc

        collection_parts = [url_parse.unquote(part) for part in parsed.path.split("/") if part]
        if len(collection_parts) != 1:
            raise ValueError("qdrant medium URLs must be qdrant://host:port/collection")

        query = url_parse.parse_qs(parsed.query, keep_blank_values=True)
        point_id = _parse_point_id(_query_value(query, "point", "id", default="1"))
        payload_field = _query_value(query, "field", "payload_field", default="text")
        bytes_field = _query_value(query, "bytes_field", default="data_base64")
        timeout = _parse_timeout(_query_value(query, "timeout", default="10"))

        host = parsed.hostname
        host_part = f"[{host}]" if ":" in host and not host.startswith("[") else host
        netloc = f"{host_part}:{parsed_port}" if parsed_port is not None else host_part

        return cls(
            base_url=f"http://{netloc}",
            collection=collection_parts[0],
            point_id=point_id,
            payload_field=payload_field,
            bytes_field=bytes_field,
            timeout=timeout,
        )

    def read_text(self) -> str:
        """Read text from a Qdrant point payload."""

        payload = self._retrieve_payload()
        value = payload.get(self.payload_field)
        if isinstance(value, str):
            return value

        encoded = payload.get(self.bytes_field)
        if isinstance(encoded, str):
            try:
                return _decode_base64(encoded).decode("utf-8")
            except UnicodeDecodeError as exc:
                raise QdrantError(f"qdrant payload field {self.bytes_field!r} is not UTF-8 text") from exc

        raise QdrantError(f"qdrant payload field {self.payload_field!r} is missing or not text")

    def write_text(self, value: str) -> str:
        """Write text into a Qdrant point payload."""

        self._set_payload({self.payload_field: value})
        return value

    def read_bytes(self) -> bytes:
        """Read bytes from a Qdrant point payload."""

        payload = self._retrieve_payload()
        encoded = payload.get(self.bytes_field)
        if isinstance(encoded, str):
            return _decode_base64(encoded)

        value = payload.get(self.payload_field)
        if isinstance(value, str):
            return value.encode("utf-8")

        raise QdrantError(f"qdrant payload fields {self.bytes_field!r} and {self.payload_field!r} are missing")

    def write_bytes(self, value: bytes) -> bytes:
        """Write bytes into a Qdrant point payload."""

        payload = {self.bytes_field: base64.b64encode(value).decode("ascii")}
        try:
            payload[self.payload_field] = value.decode("utf-8")
        except UnicodeDecodeError:
            payload[self.payload_field] = ""
        self._set_payload(payload)
        return value

    def _retrieve_payload(self) -> dict[str, Any]:
        result = self._request(
            "POST",
            self._points_path(),
            {"ids": [self.point_id], "with_payload": True, "with_vector": False},
        ).get("result")
        if not isinstance(result, list):
            raise QdrantError("qdrant retrieve response is missing a result list")
        if not result:
            raise QdrantError(f"qdrant point {self.point_id!r} was not found in collection {self.collection!r}")

        point = result[0]
        if not isinstance(point, dict):
            raise QdrantError("qdrant retrieve response contains an invalid point")
        payload = point.get("payload")
        if not isinstance(payload, dict):
            raise QdrantError("qdrant retrieve response is missing a payload object")
        return payload

    def _set_payload(self, payload: dict[str, Any]) -> None:
        self._request("POST", self._payload_path(), {"payload": payload, "points": [self.point_id]})

    def _points_path(self) -> str:
        collection = url_parse.quote(self.collection, safe="")
        return f"/collections/{collection}/points"

    def _payload_path(self) -> str:
        collection = url_parse.quote(self.collection, safe="")
        return f"/collections/{collection}/points/payload"

    def _request(self, method: str, path: str, body: dict[str, Any] | None = None) -> dict[str, Any]:
        url = f"{self.base_url}{path}"
        data = None
        headers = {"Accept": "application/json"}
        if body is not None:
            data = stdlib_json.dumps(body).encode("utf-8")
            headers["Content-Type"] = "application/json"

        request = url_request.Request(url, data=data, headers=headers, method=method)
        try:
            response = url_request.urlopen(request, timeout=self.timeout)
            try:
                status = getattr(response, "status", getattr(response, "code", 200))
                raw_body = response.read()
            finally:
                close = getattr(response, "close", None)
                if close is not None:
                    close()
        except url_error.HTTPError as exc:
            detail = _read_error_detail(exc)
            raise QdrantError(f"qdrant {method} {path} failed with HTTP {exc.code}: {detail}") from exc
        except url_error.URLError as exc:
            reason = getattr(exc, "reason", exc)
            raise QdrantError(f"qdrant {method} {path} failed: {reason}") from exc

        if status < 200 or status >= 300:
            raise QdrantError(f"qdrant {method} {path} failed with HTTP {status}: {_decode_body(raw_body)}")
        if not raw_body:
            return {}

        try:
            payload = stdlib_json.loads(raw_body.decode("utf-8"))
        except (UnicodeDecodeError, stdlib_json.JSONDecodeError) as exc:
            raise QdrantError(f"qdrant {method} {path} returned invalid JSON") from exc
        if not isinstance(payload, dict):
            raise QdrantError(f"qdrant {method} {path} returned a non-object response")

        status_value = payload.get("status")
        if status_value not in (None, "ok"):
            raise QdrantError(f"qdrant {method} {path} returned status {status_value!r}")
        return payload


def memory(initial_text: str = "") -> Medium:
    """Create an in-memory medium.

    medium.memory("hello")
    """

    return Medium(text=initial_text, data=initial_text.encode("utf-8"))


def from_path(path: str | Path) -> Medium:
    """Create a filesystem-backed medium.

    medium.from_path("/tmp/corepy.txt")
    """

    return Medium(location=path)


def open(target: str | Path) -> Medium | QdrantMedium:
    """Open a filesystem or Qdrant-backed medium.

    medium.open("qdrant://localhost:6333/core_medium")
    """

    target_text = str(target)
    parsed = url_parse.urlparse(target_text)
    if parsed.scheme == "qdrant":
        return QdrantMedium.from_url(target_text)
    if parsed.scheme == "file":
        return from_path(url_parse.unquote(parsed.path))
    if parsed.scheme:
        raise ValueError(f"unsupported medium URL scheme {parsed.scheme!r}")
    return from_path(target)


def read_text(medium_value: Medium | QdrantMedium) -> str:
    """Read text from a medium handle.

    medium.read_text(buffer)
    """

    return medium_value.read_text()


def write_text(medium_value: Medium | QdrantMedium, value: str) -> str:
    """Write text to a medium handle.

    medium.write_text(buffer, "updated")
    """

    return medium_value.write_text(value)


def read_bytes(medium_value: Medium | QdrantMedium) -> bytes:
    """Read bytes from a medium handle.

    medium.read_bytes(buffer)
    """

    return medium_value.read_bytes()


def write_bytes(medium_value: Medium | QdrantMedium, value: bytes) -> bytes:
    """Write bytes to a medium handle.

    medium.write_bytes(buffer, b"updated")
    """

    return medium_value.write_bytes(value)


def _query_value(query: dict[str, list[str]], *names: str, default: str) -> str:
    for name in names:
        values = query.get(name)
        if values:
            return values[0]
    return default


def _parse_point_id(value: str) -> int | str:
    return int(value) if value.isdecimal() else value


def _parse_timeout(value: str) -> float:
    try:
        timeout = float(value)
    except ValueError as exc:
        raise ValueError("qdrant medium URL timeout must be a number") from exc
    if timeout <= 0:
        raise ValueError("qdrant medium URL timeout must be positive")
    return timeout


def _decode_base64(value: str) -> bytes:
    try:
        return base64.b64decode(value.encode("ascii"), validate=True)
    except (UnicodeEncodeError, binascii.Error) as exc:
        raise QdrantError("qdrant payload contains invalid base64 bytes") from exc


def _read_error_detail(error: url_error.HTTPError) -> str:
    try:
        return _decode_body(error.read())
    except OSError:
        return error.reason


def _decode_body(body: bytes) -> str:
    if not body:
        return ""
    try:
        return body.decode("utf-8")
    except UnicodeDecodeError:
        return repr(body)
