"""core — Python binding for Core primitives.

Use the same import paths across Tier 1 and Tier 2:

    from core import echo, fs, json, math, options, path, strings
    print(echo("hello"))
    fs.write_file("/tmp/corepy.json", json.dumps({"name": "corepy"}))
"""

from . import config, data, err, fs, json, log, math, medium, options, path, process, service, strings

__version__ = "0.2.0"


def echo(value: str) -> str:
    """Return the value unchanged.

    echo("hello")
    """

    return value


__all__ = [
    "config",
    "data",
    "echo",
    "err",
    "fs",
    "json",
    "log",
    "math",
    "medium",
    "options",
    "path",
    "process",
    "service",
    "strings",
]
