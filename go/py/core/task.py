"""Task composition helpers built from named actions.

from core import action, task

actions = action.new_registry()
tasks = task.new_registry()
"""

from __future__ import annotations

import builtins
from dataclasses import dataclass, field
from threading import Thread
from typing import Any

from . import action as action_module


@dataclass(slots=True)
class Step:
    """Single task step referencing a named action.

    step = task.Step(action="echo")
    """

    action: str
    with_values: dict[str, Any] = field(default_factory=dict)
    async_step: bool = False
    input: str = ""


@dataclass(slots=True)
class Task:
    """Named sequence of steps.

    plan = task.Task(name="deploy", steps=[step])
    """

    name: str = ""
    description: str = ""
    steps: list[Step] = field(default_factory=list)

    def run(
        self,
        actions: action_module.ActionRegistry,
        values: dict[str, Any] | None = None,
        context: Any = None,
    ) -> Any:
        """Run task steps in order.

        plan.run(actions, {"text": "hello"})
        """

        if not self.steps:
            raise RuntimeError(f"task has no steps: {self.name or '<nil>'}")

        runtime_values = {} if values is None else dict(values)
        previous: Any = None
        previous_ok = False

        for step in self.steps:
            step_values = dict(step.with_values)
            if not step_values:
                step_values = dict(runtime_values)
            if step.input == "previous" and previous_ok:
                step_values["_input"] = previous

            current_action = actions.get(step.action)
            if not current_action.exists():
                raise RuntimeError(f"action not found: {step.action}")

            if step.async_step:
                Thread(
                    target=_run_async,
                    args=(current_action, step_values, context),
                    daemon=True,
                ).start()
                continue

            previous = current_action.run(step_values, context)
            previous_ok = True

        return previous


class TaskRegistry:
    """Named registry of tasks in insertion order.

    items = task.TaskRegistry()
    """

    def __init__(self) -> None:
        self._tasks: dict[str, Task] = {}
        self._order: list[str] = []

    def register(self, name: str, steps: list[Step | dict[str, Any]], description: str = "") -> Task:
        """Register or replace a named task.

        items.register("deploy", steps)
        """

        if name not in self._tasks:
            self._order.append(name)
        plan = Task(name=name, description=description, steps=[_step(step) for step in steps])
        self._tasks[name] = plan
        return plan

    def get(self, name: str) -> Task:
        """Return a named task or a placeholder.

        items.get("deploy")
        """

        return self._tasks.get(name, Task(name=name))

    def names(self) -> list[str]:
        """Return task names in insertion order.

        items.names()
        """

        return builtins.list(self._order)


def new(
    name: str = "",
    steps: list[Step | dict[str, Any]] | None = None,
    description: str = "",
) -> Task:
    """Create a Task handle.

    task.new("deploy", [{"action": "echo"}])
    """

    return Task(name=name, description=description, steps=[] if steps is None else [_step(step) for step in steps])


def new_step(
    action: str,
    with_values: dict[str, Any] | None = None,
    async_step: bool = False,
    input: str = "",
) -> Step:
    """Create a Step value.

    task.new_step("echo", {"text": "hello"})
    """

    return Step(action=action, with_values={} if with_values is None else dict(with_values), async_step=async_step, input=input)


def new_registry() -> TaskRegistry:
    """Create a TaskRegistry handle.

    task.new_registry()
    """

    return TaskRegistry()


def register(
    registry_value: TaskRegistry,
    name: str,
    steps: list[Step | dict[str, Any]],
    description: str = "",
) -> Task:
    """Register or replace a named task.

    task.register(items, "deploy", steps)
    """

    return registry_value.register(name, steps, description)


def get(registry_value: TaskRegistry, name: str) -> Task:
    """Return a task or a placeholder.

    task.get(items, "deploy")
    """

    return registry_value.get(name)


def names(registry_value: TaskRegistry) -> list[str]:
    """Return registered task names.

    task.names(items)
    """

    return registry_value.names()


def run(
    task_value: Task | TaskRegistry,
    actions: action_module.ActionRegistry,
    *arguments: Any,
    **kwargs: Any,
) -> Any:
    """Run a task handle or a named task from a registry.

    task.run(plan, actions, {"text": "hello"})
    task.run(items, actions, "deploy", {"text": "hello"})
    """

    context = kwargs.get("context")
    if isinstance(task_value, TaskRegistry):
        if not arguments:
            raise TypeError("task.run expected a task name")
        name = str(arguments[0])
        values = arguments[1] if builtins.len(arguments) > 1 else None
        return task_value.get(name).run(actions, values, context)
    values = arguments[0] if arguments else None
    return task_value.run(actions, values, context)


def exists(task_value: Task) -> bool:
    """Return True when the task has at least one step.

    task.exists(plan)
    """

    return builtins.len(task_value.steps) > 0


def _step(value: Step | dict[str, Any]) -> Step:
    if isinstance(value, Step):
        return value
    return Step(
        action=str(value["action"]),
        with_values=dict(value.get("with_values", value.get("with", {}))),
        async_step=bool(value.get("async_step", value.get("async", False))),
        input=str(value.get("input", "")),
    )


def _run_async(current_action: action_module.Action, step_values: dict[str, Any], context: Any) -> None:
    try:
        current_action.run(step_values, context)
    except Exception:
        return


__all__ = [
    "Step",
    "Task",
    "TaskRegistry",
    "exists",
    "get",
    "names",
    "new",
    "new_registry",
    "new_step",
    "register",
    "run",
]
