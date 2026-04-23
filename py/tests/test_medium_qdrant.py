from __future__ import annotations

import json
import unittest
from urllib import error as url_error
from unittest.mock import patch

from core import medium


class FakeResponse:
    def __init__(self, payload: dict[str, object] | bytes, status: int = 200) -> None:
        self.status = status
        self._body = json.dumps(payload).encode("utf-8") if isinstance(payload, dict) else payload
        self.closed = False

    def read(self) -> bytes:
        return self._body

    def close(self) -> None:
        self.closed = True


class FakeQdrantHTTP:
    def __init__(self, *responses: FakeResponse | Exception) -> None:
        self.responses = list(responses)
        self.requests: list[object] = []
        self.timeouts: list[float] = []

    def __call__(self, request: object, timeout: float) -> FakeResponse:
        self.requests.append(request)
        self.timeouts.append(timeout)
        response = self.responses.pop(0)
        if isinstance(response, Exception):
            raise response
        return response


class MediumQdrantTests(unittest.TestCase):
    def test_open_parses_qdrant_url(self) -> None:
        handle = medium.open("qdrant://qdrant.local:6333/articles?point=42&field=body&timeout=2.5")

        self.assertIsInstance(handle, medium.QdrantMedium)
        self.assertEqual(handle.base_url, "http://qdrant.local:6333")
        self.assertEqual(handle.collection, "articles")
        self.assertEqual(handle.point_id, 42)
        self.assertEqual(handle.payload_field, "body")
        self.assertEqual(handle.bytes_field, "data_base64")
        self.assertEqual(handle.timeout, 2.5)

    def test_qdrant_read_text_retrieves_point_payload(self) -> None:
        fake_http = FakeQdrantHTTP(
            FakeResponse({"status": "ok", "result": [{"id": 1, "payload": {"text": "hello"}}]})
        )

        with patch.object(medium.url_request, "urlopen", fake_http):
            handle = medium.open("qdrant://qdrant.local:6333/articles")
            self.assertEqual(medium.read_text(handle), "hello")

        request = fake_http.requests[0]
        self.assertEqual(request.get_method(), "POST")
        self.assertEqual(request.full_url, "http://qdrant.local:6333/collections/articles/points")
        self.assertEqual(
            json.loads(request.data.decode("utf-8")),
            {
                "ids": [1],
                "with_payload": True,
                "with_vector": False,
            },
        )
        self.assertEqual(_headers(request)["content-type"], "application/json")
        self.assertEqual(fake_http.timeouts, [10.0])

    def test_qdrant_write_text_sets_point_payload(self) -> None:
        fake_http = FakeQdrantHTTP(
            FakeResponse({"status": "ok", "result": {"operation_id": 7, "status": "acknowledged"}})
        )

        with patch.object(medium.url_request, "urlopen", fake_http):
            handle = medium.open("qdrant://qdrant.local:6333/articles")
            self.assertEqual(medium.write_text(handle, "updated"), "updated")

        request = fake_http.requests[0]
        self.assertEqual(request.get_method(), "POST")
        self.assertEqual(request.full_url, "http://qdrant.local:6333/collections/articles/points/payload")
        self.assertEqual(
            json.loads(request.data.decode("utf-8")),
            {
                "payload": {"text": "updated"},
                "points": [1],
            },
        )

    def test_qdrant_write_bytes_sets_base64_payload(self) -> None:
        fake_http = FakeQdrantHTTP(FakeResponse({"status": "ok", "result": {"status": "acknowledged"}}))

        with patch.object(medium.url_request, "urlopen", fake_http):
            handle = medium.open("qdrant://qdrant.local:6333/articles")
            self.assertEqual(medium.write_bytes(handle, b"\xff\xfe"), b"\xff\xfe")

        request = fake_http.requests[0]
        self.assertEqual(request.full_url, "http://qdrant.local:6333/collections/articles/points/payload")
        self.assertEqual(
            json.loads(request.data.decode("utf-8")),
            {
                "payload": {"data_base64": "//4=", "text": ""},
                "points": [1],
            },
        )

    def test_qdrant_open_rejects_missing_collection(self) -> None:
        with self.assertRaisesRegex(ValueError, "qdrant://host:port/collection"):
            medium.open("qdrant://qdrant.local:6333")

    def test_qdrant_http_errors_raise_qdrant_error(self) -> None:
        fake_http = FakeQdrantHTTP(url_error.URLError("connection refused"))

        with patch.object(medium.url_request, "urlopen", fake_http):
            handle = medium.open("qdrant://qdrant.local:6333/articles")
            with self.assertRaisesRegex(medium.QdrantError, "connection refused"):
                handle.read_text()

    def test_qdrant_missing_point_raises_qdrant_error(self) -> None:
        fake_http = FakeQdrantHTTP(FakeResponse({"status": "ok", "result": []}))

        with patch.object(medium.url_request, "urlopen", fake_http):
            handle = medium.open("qdrant://qdrant.local:6333/articles")
            with self.assertRaisesRegex(medium.QdrantError, "was not found"):
                handle.read_text()


def _headers(request: object) -> dict[str, str]:
    return {key.lower(): value for key, value in request.header_items()}
