"""Mounted content helpers with path-first examples.

from core import data

assets = data.Data()
assets.mount("fixtures", "/tmp/corepy-fixtures")
text = assets.read_string("fixtures/example.txt")
"""

from __future__ import annotations

import builtins
from pathlib import Path
import re
import shutil
from typing import Any, Mapping


_GO_TEMPLATE_PATTERN = re.compile(r"\{\{\s*\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}")


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

        assets.extract("fixtures/templates", "/tmp/corepy-workspace", {"Name": "corepy"})
        """

        source_directory = self._resolve(path)
        target_directory = Path(target_dir).expanduser().resolve()
        target_directory.mkdir(parents=True, exist_ok=True)

        for source_path in sorted(source_directory.rglob("*")):
            relative_path = source_path.relative_to(source_directory)
            rendered_relative = _render_go_template(relative_path.as_posix(), template_data)
            destination_path = _safe_destination(target_directory, _strip_template_filter(source_path, rendered_relative))
            if source_path.is_dir():
                destination_path.mkdir(parents=True, exist_ok=True)
                continue

            destination_path.parent.mkdir(parents=True, exist_ok=True)
            if not _is_template_file(source_path):
                shutil.copy2(source_path, destination_path)
                continue

            try:
                text = source_path.read_text(encoding="utf-8")
            except UnicodeDecodeError:
                shutil.copy2(source_path, destination_path)
                continue
            destination_path.write_text(_render_go_template(text, template_data), encoding="utf-8")

        return str(target_directory)

    def mounts(self) -> list[str]:
        """Return mounted names in insertion order.

        assets.mounts()
        """

        return builtins.list(self._mounts.keys())

    def _resolve(self, logical_path: str) -> Path:
        mount_name, _, relative_path = logical_path.partition("/")
        if mount_name not in self._mounts:
            raise KeyError(f"mount not found: {mount_name}")
        root = self._mounts[mount_name]
        return root if relative_path == "" else root / relative_path


def new() -> Data:
    """Create a Data handle.

    data.new()
    """

    return Data()


def mount(data_value: Data, name: str, source: str | Path, path: str = ".") -> Data:
    """Mount a directory onto a Data handle and return it.

    data.mount(assets, "fixtures", "/tmp/corepy-fixtures")
    """

    data_value.mount(name, source, path)
    return data_value


def mount_path(data_value: Data, name: str, source: str | Path, path: str = ".") -> Data:
    """Mount a directory onto a Data handle using the Go binding name.

    data.mount_path(assets, "fixtures", "/tmp/corepy-fixtures")
    """

    return mount(data_value, name, source, path)


def read_file(data_value: Data, path: str) -> bytes:
    """Read file bytes from a Data handle.

    data.read_file(assets, "fixtures/example.txt")
    """

    return data_value.read_file(path)


def read_string(data_value: Data, path: str) -> str:
    """Read text from a Data handle.

    data.read_string(assets, "fixtures/example.txt")
    """

    return data_value.read_string(path)


def list(data_value: Data, path: str) -> list[str]:
    """List child names from a Data handle.

    data.list(assets, "fixtures")
    """

    return data_value.list(path)


def list_names(data_value: Data, path: str) -> list[str]:
    """List child stems from a Data handle.

    data.list_names(assets, "fixtures")
    """

    return data_value.list_names(path)


def extract(data_value: Data, path: str, target_dir: str | Path, template_data: Mapping[str, Any] | None = None) -> str:
    """Extract mounted content from a Data handle.

    data.extract(assets, "fixtures/templates", "/tmp/corepy-workspace", {"Name": "corepy"})
    """

    return data_value.extract(path, target_dir, template_data)


def mounts(data_value: Data) -> list[str]:
    """Return mounted names from a Data handle.

    data.mounts(assets)
    """

    return data_value.mounts()


def _is_template_file(path: Path) -> bool:
    return ".tmpl" in path.name


def _strip_template_filter(source_path: Path, rendered_relative: str) -> Path:
    relative = Path(rendered_relative)
    if not _is_template_file(source_path):
        return relative
    return relative.with_name(relative.name.replace(".tmpl", ""))


def _render_go_template(value: str, template_data: Mapping[str, Any] | None) -> str:
    if template_data is None:
        return value

    def replace(match: re.Match[str]) -> str:
        key = match.group(1)
        if key not in template_data:
            return match.group(0)
        return str(template_data[key])

    return _GO_TEMPLATE_PATTERN.sub(replace, value)


def _safe_destination(target_root: Path, relative_path: Path) -> Path:
    destination = (target_root / relative_path).resolve()
    if destination != target_root and target_root not in destination.parents:
        raise ValueError(f"extracted path escapes target directory: {relative_path}")
    return destination
