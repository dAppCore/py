"""Git-backed source-control helpers.

from core import scm

if scm.exists("."):
    print(scm.branch("."))
"""

from __future__ import annotations

from pathlib import Path
import shutil
import subprocess
from typing import Any


def exists(directory: str | Path = ".") -> bool:
    """Return True when the directory is inside a Git worktree.

    scm.exists(".")
    """

    if shutil.which("git") is None:
        return False
    completed = _git(directory, "rev-parse", "--is-inside-work-tree", check=False)
    return completed.strip() == "true"


def root(directory: str | Path = ".") -> str:
    """Return the repository root directory.

    scm.root(".")
    """

    return _git(directory, "rev-parse", "--show-toplevel").strip()


def branch(directory: str | Path = ".") -> str:
    """Return the current branch name.

    scm.branch(".")
    """

    return _git(directory, "rev-parse", "--abbrev-ref", "HEAD").strip()


def head(directory: str | Path = ".") -> str:
    """Return the current HEAD commit hash.

    scm.head(".")
    """

    return _git(directory, "rev-parse", "HEAD").strip()


def tracked_files(directory: str | Path = ".") -> list[str]:
    """Return tracked repository paths.

    scm.tracked_files(".")
    """

    output = _git(directory, "ls-files")
    return [line for line in output.splitlines() if line]


def status(directory: str | Path = ".") -> dict[str, Any]:
    """Return branch, cleanliness, and change lines.

    scm.status(".")
    """

    output = _git(directory, "status", "--short", "--branch")
    lines = [line.rstrip() for line in output.splitlines() if line.rstrip()]
    branch_name = ""
    if lines and lines[0].startswith("## "):
        branch_name = lines[0][3:].strip()
        lines = lines[1:]
    return {
        "branch": branch_name,
        "clean": len(lines) == 0,
        "changes": lines,
    }


def _git(directory: str | Path, *arguments: str, check: bool = True) -> str:
    git_binary = shutil.which("git")
    if git_binary is None:
        raise RuntimeError("git is not available")

    completed = subprocess.run(
        [git_binary, "-C", str(directory), *arguments],
        capture_output=True,
        check=False,
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


__all__ = ["branch", "exists", "head", "root", "status", "tracked_files"]
