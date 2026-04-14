"""Simple transport wrapper for memory or filesystem-backed content.

from core import medium

buffer = medium.memory("hello")
buffer.write_text("updated")
"""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from . import fs as core_fs


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


def read_text(medium_value: Medium) -> str:
    """Read text from a medium handle.

    medium.read_text(buffer)
    """

    return medium_value.read_text()


def write_text(medium_value: Medium, value: str) -> str:
    """Write text to a medium handle.

    medium.write_text(buffer, "updated")
    """

    return medium_value.write_text(value)


def read_bytes(medium_value: Medium) -> bytes:
    """Read bytes from a medium handle.

    medium.read_bytes(buffer)
    """

    return medium_value.read_bytes()


def write_bytes(medium_value: Medium, value: bytes) -> bytes:
    """Write bytes to a medium handle.

    medium.write_bytes(buffer, b"updated")
    """

    return medium_value.write_bytes(value)
