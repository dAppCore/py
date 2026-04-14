"""Mounted content helpers with path-first examples.

from core import data

assets = data.Data()
assets.mount("fixtures", "/tmp/corepy-fixtures")
text = assets.read_string("fixtures/example.txt")
"""

from __future__ import annotations

from pathlib import Path
import shutil
from typing import Any, Mapping


class _TemplateValues(dict[str, Any]):
    def __missing__(self, key: str) -> str:
        return "{" + key + "}"


class Data:
    """Mounted content registry for local directories.

    assets = data.Data()
    """

    def __init__(self) -> None:
        self._mounts: dict[str, Path] = {}

    def mount(self, name: str, source: str | Path, path: str = ".") -> str:
        """Mount a local directory under a logical name.

        assets.mount("fixtures", "/tmp/corepy-fixtures")
        """

        root = Path(source).expanduser().resolve()
        mounted_root = (root / path).resolve()
        self._mounts[name] = mounted_root
        return str(mounted_root)

    def read_file(self, path: str) -> bytes:
        """Read mounted file bytes.

        assets.read_file("fixtures/example.txt")
        """

        return self._resolve(path).read_bytes()

    def read_string(self, path: str) -> str:
        """Read mounted file text.

        assets.read_string("fixtures/example.txt")
        """

        return self._resolve(path).read_text(encoding="utf-8")

    def list(self, path: str) -> list[str]:
        """List child names at a mounted path.

        assets.list("fixtures")
        """

        return sorted(child.name for child in self._resolve(path).iterdir())

    def list_names(self, path: str) -> list[str]:
        """List child names without file extensions.

        assets.list_names("fixtures")
        """

        names: list[str] = []
        for child_name in self.list(path):
            names.append(Path(child_name).stem)
        return names

    def extract(self, path: str, target_dir: str | Path, template_data: Mapping[str, Any] | None = None) -> str:
        """Copy a mounted directory into a target directory.

        assets.extract("fixtures/templates", "/tmp/corepy-workspace", {"name": "corepy"})
        """

        source_directory = self._resolve(path)
        target_directory = Path(target_dir)
        target_directory.mkdir(parents=True, exist_ok=True)

        for source_path in source_directory.rglob("*"):
            relative_path = source_path.relative_to(source_directory)
            destination_path = target_directory / relative_path
            if source_path.is_dir():
                destination_path.mkdir(parents=True, exist_ok=True)
                continue

            destination_path.parent.mkdir(parents=True, exist_ok=True)
            if template_data is None:
                shutil.copy2(source_path, destination_path)
                continue

            try:
                text = source_path.read_text(encoding="utf-8")
            except UnicodeDecodeError:
                shutil.copy2(source_path, destination_path)
                continue
            destination_path.write_text(text.format_map(_TemplateValues(template_data)), encoding="utf-8")

        return str(target_directory)

    def mounts(self) -> list[str]:
        """Return mounted names in insertion order.

        assets.mounts()
        """

        return list(self._mounts.keys())

    def _resolve(self, logical_path: str) -> Path:
        mount_name, _, relative_path = logical_path.partition("/")
        if mount_name not in self._mounts:
            raise KeyError(f"mount not found: {mount_name}")
        root = self._mounts[mount_name]
        return root if relative_path == "" else root / relative_path
