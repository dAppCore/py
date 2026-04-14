"""Configuration and feature flags with concrete examples.

from core import config

cfg = config.Config()
cfg.set("database.host", "localhost")
cfg.enable("debug")
"""

from __future__ import annotations

import builtins
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
        return value if isinstance(value, builtins.int) and not isinstance(value, builtins.bool) else 0

    def bool(self, key: str) -> bool:
        """Read a boolean setting or False.

        cfg.bool("debug")
        """

        value = self.get(key, False)
        return value if isinstance(value, builtins.bool) else False

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


def new() -> Config:
    """Create a Config handle.

    config.new()
    """

    return Config()


def set(config_value: Config, key: str, value: Any) -> Config:
    """Set a configuration value and return the handle.

    config.set(cfg, "debug", True)
    """

    config_value.set(key, value)
    return config_value


def get(config_value: Config, key: str) -> Any:
    """Read a configuration value from a handle.

    config.get(cfg, "database.host")
    """

    return config_value.get(key)


def string(config_value: Config, key: str) -> str:
    """Read a string configuration value from a handle.

    config.string(cfg, "database.host")
    """

    return config_value.string(key)


def int(config_value: Config, key: str) -> builtins.int:
    """Read an integer configuration value from a handle.

    config.int(cfg, "port")
    """

    return config_value.int(key)


def bool(config_value: Config, key: str) -> builtins.bool:
    """Read a boolean configuration value from a handle.

    config.bool(cfg, "debug")
    """

    return config_value.bool(key)


def enable(config_value: Config, feature: str) -> Config:
    """Enable a feature flag and return the handle.

    config.enable(cfg, "tier1")
    """

    config_value.enable(feature)
    return config_value


def disable(config_value: Config, feature: str) -> Config:
    """Disable a feature flag and return the handle.

    config.disable(cfg, "tier1")
    """

    config_value.disable(feature)
    return config_value


def enabled(config_value: Config, feature: str) -> bool:
    """Return True when a feature is enabled on a handle.

    config.enabled(cfg, "tier1")
    """

    return config_value.enabled(feature)


def enabled_features(config_value: Config) -> list[str]:
    """Return all enabled features from a handle.

    config.enabled_features(cfg)
    """

    return config_value.enabled_features()
