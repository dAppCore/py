"""Service registry helpers with concrete lifecycle examples.

from core import service

registry = service.ServiceRegistry()
registry.register("brain", service.Service(name="brain"))
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Callable


@dataclass(slots=True)
class Service:
    """Service DTO with optional lifecycle hooks.

    service.Service(name="brain")
    """

    name: str
    instance: Any = None
    on_start: Callable[[], Any] | None = None
    on_stop: Callable[[], Any] | None = None
    on_reload: Callable[[], Any] | None = None


class ServiceRegistry:
    """Ordered service registry.

    registry = service.ServiceRegistry()
    """

    def __init__(self) -> None:
        self._services: dict[str, Service] = {}

    def register(self, name: str, service_value: Service | Any) -> None:
        """Register a service by name.

        registry.register("brain", service.Service(name="brain"))
        """

        if isinstance(service_value, Service):
            service_object = service_value
            service_object.name = name
        else:
            service_object = Service(name=name, instance=service_value)
        self._services[name] = service_object

    def get(self, name: str) -> Any:
        """Return the service instance or DTO.

        registry.get("brain")
        """

        service_object = self._services[name]
        return service_object.instance if service_object.instance is not None else service_object

    def names(self) -> list[str]:
        """Return registered service names.

        registry.names()
        """

        return list(self._services.keys())

    def start_all(self) -> list[Any]:
        """Run `on_start` hooks in registration order.

        registry.start_all()
        """

        results: list[Any] = []
        for service_object in self._services.values():
            if service_object.on_start is not None:
                results.append(service_object.on_start())
        return results

    def stop_all(self) -> list[Any]:
        """Run `on_stop` hooks in registration order.

        registry.stop_all()
        """

        results: list[Any] = []
        for service_object in self._services.values():
            if service_object.on_stop is not None:
                results.append(service_object.on_stop())
        return results
