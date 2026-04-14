"""Configuration and feature flags with concrete examples.

from core import config

cfg = config.Config()
cfg.set("database.host", "localhost")
cfg.enable("debug")
"""

from __future__ import annotations

from typing import Any


class Config:
    """Runtime settings plus feature flags.

    cfg = config.Config()
    """

    def __init__(self) -> None:
        self._settings: dict[str, Any] = {}
        self._features: dict[str, bool] = {}

    def set(self, key: str, value: Any) -> None:
        """Store a setting by key.

        cfg.set("database.host", "localhost")
        """

        self._settings[key] = value

    def get(self, key: str, default: Any = None) -> Any:
        """Read a setting by key.

        cfg.get("database.host")
        """

        return self._settings.get(key, default)

    def string(self, key: str) -> str:
        """Read a string setting or an empty string.

        cfg.string("database.host")
        """

        value = self.get(key, "")
        return value if isinstance(value, str) else ""

    def int(self, key: str) -> int:
        """Read an integer setting or zero.

        cfg.int("port")
        """

        value = self.get(key, 0)
        return value if isinstance(value, int) and not isinstance(value, bool) else 0

    def bool(self, key: str) -> bool:
        """Read a boolean setting or False.

        cfg.bool("debug")
        """

        value = self.get(key, False)
        return value if isinstance(value, bool) else False

    def enable(self, feature: str) -> None:
        """Enable a feature flag.

        cfg.enable("debug")
        """

        self._features[feature] = True

    def disable(self, feature: str) -> None:
        """Disable a feature flag.

        cfg.disable("debug")
        """

        self._features[feature] = False

    def enabled(self, feature: str) -> bool:
        """Return True when a feature flag is enabled.

        cfg.enabled("debug")
        """

        return self._features.get(feature, False)

    def enabled_features(self) -> list[str]:
        """Return all enabled feature names.

        cfg.enabled_features()
        """

        return [feature for feature, enabled in self._features.items() if enabled]
