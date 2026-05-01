"""Path helpers with slash-shaped examples.

from core import path

location = path.join("deploy", "to", "homelab")
name = path.base(location)
"""

from __future__ import annotations

from glob import glob as glob_paths
import posixpath


def join(*segments: str) -> str:
    """Join path segments with `/`.

    path.join("deploy", "to", "homelab")
    """

    return "/".join(segments)


def base(value: str) -> str:
    """Return the last path element.

    path.base("/tmp/corepy/config.json")
    """

    if value == "":
        return "."
    trimmed = value.rstrip("/")
    if trimmed == "":
        return "/"
    return trimmed.split("/")[-1]


def dir(value: str) -> str:
    """Return all but the last path element.

    path.dir("/tmp/corepy/config.json")
    """

    if value == "":
        return "."
    index = value.rfind("/")
    if index < 0:
        return "."
    directory = value[:index]
    return "/" if directory == "" else directory


def ext(value: str) -> str:
    """Return the file extension including the dot.

    path.ext("config.json")
    """

    name = base(value)
    index = name.rfind(".")
    if index <= 0:
        return ""
    return name[index:]


def is_abs(value: str) -> bool:
    """Return True when the path is absolute.

    path.is_abs("/tmp/corepy")
    """

    return value.startswith("/") or (len(value) >= 3 and value[1] == ":" and value[2] in ("/", "\\"))


def clean(value: str) -> str:
    """Collapse duplicate separators and `..` segments.

    path.clean("deploy//to/../from")
    """

    if value == "":
        return "."
    return posixpath.normpath(value)


def glob(pattern: str) -> list[str]:
    """Return filesystem paths that match a glob pattern.

    path.glob("/tmp/corepy/*.json")
    """

    return glob_paths(pattern)
