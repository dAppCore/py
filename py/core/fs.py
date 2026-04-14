"""Filesystem primitives with path-shaped examples.

from core import fs

fs.ensure_dir("/tmp/corepy")
fs.write_file("/tmp/corepy/config.json", '{"name":"corepy"}')
text = fs.read_file("/tmp/corepy/config.json")
"""

from __future__ import annotations

from pathlib import Path
import tempfile


def read_file(path: str | Path) -> str:
    """Read a UTF-8 file into a string.

    fs.read_file("/tmp/corepy/config.json")
    """

    return Path(path).read_text(encoding="utf-8")


def read_bytes(path: str | Path) -> bytes:
    """Read a file into bytes.

    fs.read_bytes("/tmp/corepy/config.json")
    """

    return Path(path).read_bytes()


def write_file(path: str | Path, content: str) -> str:
    """Write UTF-8 text to a file.

    fs.write_file("/tmp/corepy/config.json", '{"name":"corepy"}')
    """

    filename = Path(path)
    ensure_dir(filename.parent)
    filename.write_text(content, encoding="utf-8")
    return str(filename)


def write_bytes(path: str | Path, content: bytes) -> str:
    """Write bytes to a file.

    fs.write_bytes("/tmp/corepy/config.bin", b"corepy")
    """

    filename = Path(path)
    ensure_dir(filename.parent)
    filename.write_bytes(content)
    return str(filename)


def ensure_dir(path: str | Path) -> str:
    """Create a directory if it does not already exist.

    fs.ensure_dir("/tmp/corepy")
    """

    directory = Path(path)
    directory.mkdir(parents=True, exist_ok=True)
    return str(directory)


def temp_dir(prefix: str = "corepy-") -> str:
    """Create a temporary directory and return its path.

    workdir = fs.temp_dir("corepy-")
    """

    return tempfile.mkdtemp(prefix=prefix)
