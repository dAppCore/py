"""core — Python binding for Core primitives.

Use the same import paths across Tier 1 and Tier 2:

    from core import action, array, cache, crypto, dns, echo, entitlement, fs, i18n, info, json, math, options, path, registry, scm, strings, task
    print(echo("hello"))
    fs.write_file("/tmp/corepy.json", json.dumps({"name": "corepy"}))
"""

from . import action, array, cache, config, crypto, data, dns, entitlement, err, fs, i18n, info, json, log, math, medium, options, path, process, registry, scm, service, strings, task

__version__ = "0.2.0"


def echo(value: str) -> str:
    """Return the value unchanged.

    echo("hello")
    """

    return value


__all__ = [
    "array",
    "action",
    "cache",
    "config",
    "crypto",
    "data",
    "dns",
    "echo",
    "entitlement",
    "err",
    "fs",
    "i18n",
    "info",
    "json",
    "log",
    "math",
    "medium",
    "options",
    "path",
    "process",
    "registry",
    "scm",
    "service",
    "strings",
    "task",
]
