"""DNS helpers for host and service lookups.

from core import dns

addresses = dns.lookup_host("localhost")
"""

from __future__ import annotations

import socket


def lookup_host(name: str) -> list[str]:
    """Return host addresses for a name.

    dns.lookup_host("localhost")
    """

    return _unique(item[4][0] for item in socket.getaddrinfo(name, None))


def lookup_ip(name: str) -> list[str]:
    """Return IP addresses for a host.

    dns.lookup_ip("localhost")
    """

    return lookup_host(name)


def reverse_lookup(address: str) -> list[str]:
    """Return reverse-DNS names for an address.

    dns.reverse_lookup("127.0.0.1")
    """

    hostname, aliases, _ = socket.gethostbyaddr(address)
    return _unique([hostname, *aliases])


def lookup_port(network: str, service: str) -> int:
    """Return the port number for a service name.

    dns.lookup_port("tcp", "http")
    """

    return socket.getservbyname(service, network.lower())


def _unique(values: list[str] | tuple[str, ...] | object) -> list[str]:
    seen: set[str] = set()
    result: list[str] = []
    for value in values:
        text = str(value)
        if text in seen:
            continue
        seen.add(text)
        result.append(text)
    return result


__all__ = ["lookup_host", "lookup_ip", "lookup_port", "reverse_lookup"]
