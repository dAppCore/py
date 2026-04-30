"""Locale and translation helpers.

from core import i18n

messages = i18n.new()
"""

from __future__ import annotations

import builtins
from typing import Any, Protocol


class Translator(Protocol):
    def translate(self, message_id: str, *args: Any) -> Any: ...
    def set_language(self, lang: str) -> None: ...
    def language(self) -> str: ...
    def available_languages(self) -> list[str]: ...


class I18n:
    """Locale collection plus optional translator dispatch.

    messages = i18n.I18n()
    """

    def __init__(self) -> None:
        self._locales: list[Any] = []
        self._locale = ""
        self._translator: Translator | None = None

    def add_locales(self, *mounts: Any) -> None:
        """Append locale mounts.

        messages.add_locales("locales")
        """

        self._locales.extend(mounts)

    def locales(self) -> list[Any]:
        """Return collected locale mounts.

        messages.locales()
        """

        return builtins.list(self._locales)

    def set_translator(self, translator: Translator | None) -> None:
        """Register a translator implementation.

        messages.set_translator(translator)
        """

        self._translator = translator
        if translator is not None and self._locale:
            translator.set_language(self._locale)

    def translator(self) -> Translator | None:
        """Return the registered translator.

        messages.translator()
        """

        return self._translator

    def translate(self, message_id: str, *args: Any) -> Any:
        """Translate a message or return the key as-is.

        messages.translate("hello")
        """

        if self._translator is None:
            return message_id
        return self._translator.translate(message_id, *args)

    def set_language(self, lang: str) -> None:
        """Set the active language.

        messages.set_language("de")
        """

        if lang == "":
            return
        self._locale = lang
        if self._translator is not None:
            self._translator.set_language(lang)

    def language(self) -> str:
        """Return the active language or `en`.

        messages.language()
        """

        if self._locale:
            return self._locale
        if self._translator is not None:
            value = self._translator.language()
            if value:
                return value
        return "en"

    def available_languages(self) -> list[str]:
        """Return available language codes.

        messages.available_languages()
        """

        if self._translator is None:
            return ["en"]
        return builtins.list(self._translator.available_languages())


def new() -> I18n:
    """Create an I18n handle.

    i18n.new()
    """

    return I18n()


def add_locales(i18n_value: I18n, *mounts: Any) -> I18n:
    """Append locale mounts and return the handle.

    i18n.add_locales(messages, "locales")
    """

    i18n_value.add_locales(*mounts)
    return i18n_value


def locales(i18n_value: I18n) -> list[Any]:
    """Return collected locale mounts.

    i18n.locales(messages)
    """

    return i18n_value.locales()


def set_translator(i18n_value: I18n, translator: Translator | None) -> I18n:
    """Register a translator and return the handle.

    i18n.set_translator(messages, translator)
    """

    i18n_value.set_translator(translator)
    return i18n_value


def translator(i18n_value: I18n) -> Translator | None:
    """Return the registered translator.

    i18n.translator(messages)
    """

    return i18n_value.translator()


def translate(i18n_value: I18n, message_id: str, *args: Any) -> Any:
    """Translate a message or return the key.

    i18n.translate(messages, "hello")
    """

    return i18n_value.translate(message_id, *args)


def set_language(i18n_value: I18n, lang: str) -> I18n:
    """Set the active language and return the handle.

    i18n.set_language(messages, "de")
    """

    i18n_value.set_language(lang)
    return i18n_value


def language(i18n_value: I18n) -> str:
    """Return the active language.

    i18n.language(messages)
    """

    return i18n_value.language()


def available_languages(i18n_value: I18n) -> list[str]:
    """Return available language codes.

    i18n.available_languages(messages)
    """

    return i18n_value.available_languages()


__all__ = [
    "I18n",
    "Translator",
    "add_locales",
    "available_languages",
    "language",
    "locales",
    "new",
    "set_language",
    "set_translator",
    "translate",
    "translator",
]
