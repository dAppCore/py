"""Named action helpers mirroring Core's capability map.

from core import action

actions = action.new_registry()
action.register(actions, "echo", lambda ctx, values: values["text"])
"""

from __future__ import annotations

import builtins
import inspect
from typing import Any, Callable


ActionHandler = Callable[..., Any]


class Action:
    """Named callable with enable/disable and existence checks.

    item = action.Action("echo", handler)
    """

    def __init__(
        self,
        name: str = "",
        handler: ActionHandler | None = None,
        description: str = "",
        schema: dict[str, Any] | None = None,
    ) -> None:
        self.name = name
        self.handler = handler
        self.description = description
        self.schema = {} if schema is None else dict(schema)
        self.enabled = True

    def run(self, values: dict[str, Any] | None = None, context: Any = None) -> Any:
        """Run the action with Core-shaped options.

        item.run({"text": "hello"})
        """

        if not self.exists():
            raise RuntimeError(f"action not registered: {self.name or '<nil>'}")
        if not self.enabled:
            raise RuntimeError(f"action disabled: {self.name}")
        return _invoke_handler(self.handler, context, _values(values))

    def exists(self) -> bool:
        """Return True when a handler is present.

        item.exists()
        """

        return self.handler is not None


class ActionRegistry:
    """Named registry of actions in insertion order.

    actions = action.ActionRegistry()
    """

    def __init__(self) -> None:
        self._actions: dict[str, Action] = {}
        self._order: list[str] = []

    def register(
        self,
        name: str,
        handler: ActionHandler | None,
        description: str = "",
        schema: dict[str, Any] | None = None,
    ) -> Action:
        """Register or replace a named action.

        actions.register("echo", handler)
        """

        if name not in self._actions:
            self._order.append(name)
        item = Action(name, handler, description, schema)
        self._actions[name] = item
        return item

    def get(self, name: str) -> Action:
        """Return a registered action or a placeholder.

        actions.get("echo")
        """

        return self._actions.get(name, Action(name))

    def names(self) -> list[str]:
        """Return registered action names in insertion order.

        actions.names()
        """

        return builtins.list(self._order)

    def run(self, name: str, values: dict[str, Any] | None = None, context: Any = None) -> Any:
        """Run a named action.

        actions.run("echo", {"text": "hello"})
        """

        return self.get(name).run(values, context)

    def disable(self, name: str) -> Action:
        """Disable a named action.

        actions.disable("echo")
        """

        item = self.get(name)
        if not item.exists():
            raise KeyError(name)
        item.enabled = False
        return item

    def enable(self, name: str) -> Action:
        """Enable a named action.

        actions.enable("echo")
        """

        item = self.get(name)
        if not item.exists():
            raise KeyError(name)
        item.enabled = True
        return item


def new(
    name: str = "",
    handler: ActionHandler | None = None,
    description: str = "",
    schema: dict[str, Any] | None = None,
) -> Action:
    """Create an Action handle.

    action.new("echo", handler)
    """

    return Action(name, handler, description, schema)


def new_registry() -> ActionRegistry:
    """Create an ActionRegistry handle.

    action.new_registry()
    """

    return ActionRegistry()


def register(
    registry_value: ActionRegistry,
    name: str,
    handler: ActionHandler | None,
    description: str = "",
    schema: dict[str, Any] | None = None,
) -> Action:
    """Register or replace a named action.

    action.register(actions, "echo", handler)
    """

    return registry_value.register(name, handler, description, schema)


def get(registry_value: ActionRegistry, name: str) -> Action:
    """Return a named action or a placeholder.

    action.get(actions, "echo")
    """

    return registry_value.get(name)


def names(registry_value: ActionRegistry) -> list[str]:
    """Return registered action names.

    action.names(actions)
    """

    return registry_value.names()


def run(action_value: Action | ActionRegistry, *arguments: Any, **kwargs: Any) -> Any:
    """Run an action handle or a named action on a registry.

    action.run(item, {"text": "hello"})
    action.run(actions, "echo", {"text": "hello"})
    """

    context = kwargs.get("context")
    if isinstance(action_value, ActionRegistry):
        if not arguments:
            raise TypeError("action.run expected an action name")
        name = str(arguments[0])
        values = arguments[1] if len(arguments) > 1 else None
        return action_value.run(name, values, context)
    values = arguments[0] if arguments else None
    return action_value.run(values, context)


def exists(action_value: Action) -> bool:
    """Return True when an action has a handler.

    action.exists(item)
    """

    return action_value.exists()


def disable(registry_value: ActionRegistry, name: str) -> Action:
    """Disable a named action.

    action.disable(actions, "echo")
    """

    return registry_value.disable(name)


def enable(registry_value: ActionRegistry, name: str) -> Action:
    """Enable a named action.

    action.enable(actions, "echo")
    """

    return registry_value.enable(name)


def _values(values: dict[str, Any] | None) -> dict[str, Any]:
    if values is None:
        return {}
    return dict(values)


def _invoke_handler(handler: ActionHandler | None, context: Any, values: dict[str, Any]) -> Any:
    if handler is None:
        raise RuntimeError("action handler is not set")

    try:
        signature = inspect.signature(handler)
    except (TypeError, ValueError):
        return handler(context, values)

    parameters = [
        parameter
        for parameter in signature.parameters.values()
        if parameter.kind in (inspect.Parameter.POSITIONAL_ONLY, inspect.Parameter.POSITIONAL_OR_KEYWORD)
    ]
    if any(parameter.kind == inspect.Parameter.VAR_POSITIONAL for parameter in signature.parameters.values()):
        return handler(context, values)
    if builtins.len(parameters) == 0:
        return handler()
    if builtins.len(parameters) == 1:
        return handler(values)
    return handler(context, values)


__all__ = [
    "Action",
    "ActionRegistry",
    "disable",
    "enable",
    "exists",
    "get",
    "names",
    "new",
    "new_registry",
    "register",
    "run",
]
