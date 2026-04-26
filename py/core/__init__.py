"""core — Python binding for Core primitives.

Use the same import paths across Tier 1 and Tier 2:

    from core import action, agent, api, array, cache, container, crypto, dns, echo, entitlement, fs, i18n, info, json, math, mcp, options, path, registry, scm, store, strings, task, ws
    print(echo("hello"))
    fs.write_file("/tmp/corepy.json", json.dumps({"name": "corepy"}))
"""

from . import action, agent, api, array, cache, config, container, crypto, data, dns, entitlement, err, fs, i18n, info, json, log, math, mcp, medium, options, path, process, registry, scm, service, store, strings, task, ws

__version__ = "0.2.0"


def echo(value: str) -> str:
    """Return the value unchanged.

    echo("hello")
    """

    return value


__all__ = [
    "array",
    "action",
    "agent",
    "api",
    "cache",
    "config",
    "container",
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
    "mcp",
    "medium",
    "options",
    "path",
    "process",
    "registry",
    "scm",
    "service",
    "store",
    "strings",
    "task",
    "ws",
]
