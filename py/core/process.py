"""Process helpers for Tier 2 and local validation.

from core import process

output = process.run("python3", "-c", "print('hello')")
"""

from __future__ import annotations

import os
from pathlib import Path
import subprocess
from collections.abc import Mapping, Sequence


def run(
    command: str,
    *arguments: str,
    directory: str | Path | None = None,
    env: Mapping[str, str] | Sequence[str] | None = None,
    check: bool = True,
) -> str:
    """Run a command and return standard output.

    process.run("python3", "-c", "print('hello')")
    """

    completed = subprocess.run(
        [command, *arguments],
        capture_output=True,
        check=False,
        cwd=None if directory is None else str(directory),
        env=_merged_env(env),
        text=True,
    )
    if check and completed.returncode != 0:
        raise subprocess.CalledProcessError(
            completed.returncode,
            completed.args,
            output=completed.stdout,
            stderr=completed.stderr,
        )
    return completed.stdout


def run_in(directory: str | Path, command: str, *arguments: str, check: bool = True) -> str:
    """Run a command in a specific directory.

    process.run_in("/tmp", "python3", "-c", "print('hello')")
    """

    return run(command, *arguments, directory=directory, check=check)


def run_with_env(
    directory: str | Path,
    env: Mapping[str, str] | Sequence[str],
    command: str,
    *arguments: str,
    check: bool = True,
) -> str:
    """Run a command with extra environment variables.

    process.run_with_env("/tmp", {"MODE": "test"}, "python3", "-c", "print('hello')")
    """

    return run(command, *arguments, directory=directory, env=env, check=check)


def exists() -> bool:
    """Return True when subprocess execution is available.

    process.exists()
    """

    return True


def _merged_env(env: Mapping[str, str] | Sequence[str] | None) -> dict[str, str] | None:
    if env is None:
        return None

    merged_env = os.environ.copy()
    if isinstance(env, Mapping):
        for key, value in env.items():
            merged_env[str(key)] = str(value)
        return merged_env

    if isinstance(env, Sequence) and not isinstance(env, (str, bytes, bytearray)):
        for entry in env:
            if not isinstance(entry, str):
                raise TypeError("environment entries must be strings")
            key, separator, value = entry.partition("=")
            if separator == "":
                raise ValueError("environment entries must be KEY=value strings")
            merged_env[key] = value
        return merged_env

    raise TypeError("env must be a mapping or a sequence of KEY=value strings")
