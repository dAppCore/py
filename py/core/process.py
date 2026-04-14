"""Process helpers for Tier 2 and local validation.

from core import process

output = process.run("python3", "-c", "print('hello')")
"""

from __future__ import annotations

import os
from pathlib import Path
import subprocess
from typing import Mapping


def run(command: str, *arguments: str, directory: str | Path | None = None, env: Mapping[str, str] | None = None, check: bool = True) -> str:
    """Run a command and return standard output.

    process.run("python3", "-c", "print('hello')")
    """

    merged_env = None
    if env is not None:
        merged_env = os.environ.copy()
        merged_env.update(env)

    completed = subprocess.run(
        [command, *arguments],
        capture_output=True,
        check=False,
        cwd=None if directory is None else str(directory),
        env=merged_env,
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


def run_with_env(directory: str | Path, env: Mapping[str, str], command: str, *arguments: str, check: bool = True) -> str:
    """Run a command with extra environment variables.

    process.run_with_env("/tmp", {"MODE": "test"}, "python3", "-c", "print('hello')")
    """

    return run(command, *arguments, directory=directory, env=env, check=check)


def exists() -> bool:
    """Return True when subprocess execution is available.

    process.exists()
    """

    return True
