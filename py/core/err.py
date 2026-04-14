"""Structured errors with operation-first context.

from core import err

issue = err.e("core.save", "write failed")
wrapped = err.wrap(issue, "core.deploy", "deploy failed")
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(slots=True)
class CoreError(Exception):
    """Structured error with operation context.

    err.e("core.save", "write failed")
    """

    operation: str
    message: str
    cause: BaseException | None = None
    code: str = ""

    def __str__(self) -> str:
        prefix = f"{self.operation}: " if self.operation else ""
        if self.cause is None and self.code == "":
            return prefix + self.message
        if self.cause is None:
            return f"{prefix}{self.message} [{self.code}]"
        if self.code == "":
            return f"{prefix}{self.message}: {self.cause}"
        return f"{prefix}{self.message} [{self.code}]: {self.cause}"


def e(operation: str, message: str, cause: BaseException | None = None, *, code: str = "") -> CoreError:
    """Create a structured error.

    err.e("core.save", "write failed")
    """

    return CoreError(operation=operation, message=message, cause=cause, code=code)


def wrap(cause: BaseException | None, operation: str, message: str, *, code: str = "") -> CoreError | None:
    """Wrap an existing error with operation context.

    err.wrap(issue, "core.deploy", "deploy failed")
    """

    if cause is None and code == "":
        return None
    return CoreError(operation=operation, message=message, cause=cause, code=code)


def operation(value: BaseException) -> str:
    """Return the error operation when available.

    err.operation(issue)
    """

    return value.operation if isinstance(value, CoreError) else ""


def error_code(value: BaseException) -> str:
    """Return the error code when available.

    err.error_code(issue)
    """

    return value.code if isinstance(value, CoreError) else ""


def message(value: BaseException) -> str:
    """Return the structured message or plain string.

    err.message(issue)
    """

    return value.message if isinstance(value, CoreError) else str(value)


def root(value: BaseException | None) -> BaseException | None:
    """Return the deepest wrapped error.

    err.root(issue)
    """

    current = value
    while isinstance(current, CoreError) and current.cause is not None:
        if not isinstance(current.cause, BaseException):
            break
        current = current.cause
    return current
