"""JSON helpers with Core-shaped naming.

from core import json

payload = json.dumps({"name": "corepy"})
data = json.loads(payload)
"""

from __future__ import annotations

import json as jsonlib
from typing import Any


def dumps(value: Any, *, indent: int | None = None, sort_keys: bool = False) -> str:
    """Serialise a value to JSON text.

    json.dumps({"name": "corepy"})
    """

    return jsonlib.dumps(value, indent=indent, sort_keys=sort_keys)


def loads(value: str | bytes) -> Any:
    """Deserialise JSON text or bytes.

    json.loads('{"name":"corepy"}')
    """

    if isinstance(value, bytes):
        value = value.decode("utf-8")
    return jsonlib.loads(value)
