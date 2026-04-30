"""Structured logging helpers with predictable names.

from core import log

log.set_level("info")
log.info("service started", "service", "corepy")
"""

from __future__ import annotations

import logging
from typing import Any


_LOGGER = logging.getLogger("core")
_QUIET_LEVEL = logging.CRITICAL + 10
_LEVELS = {
    "quiet": _QUIET_LEVEL,
    "error": logging.ERROR,
    "warn": logging.WARNING,
    "info": logging.INFO,
    "debug": logging.DEBUG,
}

if not _LOGGER.handlers:
    handler = logging.StreamHandler()
    handler.setFormatter(logging.Formatter("%(levelname)s %(message)s"))
    _LOGGER.addHandler(handler)
_LOGGER.propagate = False
_LOGGER.setLevel(logging.INFO)


def set_level(level: str | int) -> None:
    """Set the logger level.

    log.set_level("debug")
    """

    if isinstance(level, str):
        level_value = _LEVELS.get(level.lower())
        if level_value is None:
            raise ValueError(f"unknown log level: {level}")
        _LOGGER.setLevel(level_value)
        return
    _LOGGER.setLevel(level)


def debug(message: str, *keyvals: Any) -> None:
    """Log a debug message with optional key-value pairs.

    log.debug("service started", "service", "corepy")
    """

    _emit(_LOGGER.debug, message, *keyvals)


def info(message: str, *keyvals: Any) -> None:
    """Log an info message with optional key-value pairs.

    log.info("service started", "service", "corepy")
    """

    _emit(_LOGGER.info, message, *keyvals)


def warn(message: str, *keyvals: Any) -> None:
    """Log a warning message with optional key-value pairs.

    log.warn("service slow", "service", "corepy")
    """

    _emit(_LOGGER.warning, message, *keyvals)


def error(message: str, *keyvals: Any) -> None:
    """Log an error message with optional key-value pairs.

    log.error("service failed", "service", "corepy")
    """

    _emit(_LOGGER.error, message, *keyvals)


def _emit(writer: Any, message: str, *keyvals: Any) -> None:
    if len(keyvals) % 2 != 0:
        raise ValueError("keyvals must be key-value pairs")
    if not keyvals:
        writer(message)
        return
    pairs = [f"{key}={value}" for key, value in zip(keyvals[::2], keyvals[1::2])]
    writer(f"{message} {' '.join(pairs)}")
