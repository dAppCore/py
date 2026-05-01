"""Read-only system information helpers.

from core import info

print(info.env("OS"))
"""

from __future__ import annotations

from datetime import datetime, timezone
import getpass
import os
from pathlib import Path
import platform
import shutil
import socket
import subprocess
import tempfile


def env(key: str) -> str:
    """Return a system value by key, falling back to the process environment.

    info.env("OS")
    """

    return _SNAPSHOT.get(key, os.environ.get(key, ""))


def keys() -> list[str]:
    """Return the available built-in keys.

    info.keys()
    """

    return sorted(_SNAPSHOT)


def snapshot() -> dict[str, str]:
    """Return a copy of the built-in system snapshot.

    info.snapshot()
    """

    return dict(_SNAPSHOT)


def _build_snapshot() -> dict[str, str]:
    home = Path(os.environ.get("CORE_HOME", Path.home()))
    snapshot = {
        "OS": _os_name(),
        "ARCH": platform.machine().lower(),
        "GO": _go_version(),
        "DS": os.sep,
        "PS": os.pathsep,
        "PID": str(os.getpid()),
        "NUM_CPU": str(os.cpu_count() or 0),
        "USER": getpass.getuser(),
        "HOSTNAME": socket.gethostname(),
        "DIR_HOME": str(home),
        "DIR_DOWNLOADS": str(home / "Downloads"),
        "DIR_CODE": str(home / "Code"),
        "DIR_TMP": tempfile.gettempdir(),
        "DIR_CWD": str(Path.cwd()),
        "CORE_START": _START,
    }
    snapshot["DIR_CONFIG"] = _config_dir(home)
    snapshot["DIR_CACHE"] = _cache_dir(home)
    snapshot["DIR_DATA"] = _data_dir(home)
    return snapshot


def _os_name() -> str:
    name = platform.system().lower()
    if name == "darwin":
        return "darwin"
    if name.startswith("win"):
        return "windows"
    return name


def _go_version() -> str:
    go_binary = shutil.which("go")
    if go_binary is None:
        return os.environ.get("GOVERSION", "")
    completed = subprocess.run(
        [go_binary, "env", "GOVERSION"],
        capture_output=True,
        check=False,
        text=True,
    )
    if completed.returncode == 0:
        return completed.stdout.strip()
    return os.environ.get("GOVERSION", "")


def _config_dir(home: Path) -> str:
    os_name = _os_name()
    if os_name == "darwin":
        return str(home / "Library" / "Application Support")
    if os_name == "windows":
        return os.environ.get("APPDATA", str(home / "AppData" / "Roaming"))
    return os.environ.get("XDG_CONFIG_HOME", str(home / ".config"))


def _cache_dir(home: Path) -> str:
    os_name = _os_name()
    if os_name == "darwin":
        return str(home / "Library" / "Caches")
    if os_name == "windows":
        return os.environ.get("LOCALAPPDATA", str(home / "AppData" / "Local"))
    return os.environ.get("XDG_CACHE_HOME", str(home / ".cache"))


def _data_dir(home: Path) -> str:
    os_name = _os_name()
    if os_name == "darwin":
        return str(home / "Library")
    if os_name == "windows":
        return os.environ.get("LOCALAPPDATA", str(home / "AppData" / "Local"))
    return os.environ.get("XDG_DATA_HOME", str(home / ".local" / "share"))


_START = datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")
_SNAPSHOT = _build_snapshot()


__all__ = ["env", "keys", "snapshot"]
