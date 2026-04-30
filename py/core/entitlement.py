"""Permission-result helpers mirroring Core's entitlement primitive.

from core import entitlement

grant = entitlement.new(True, False, 5, 4, 1, "")
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(slots=True)
class Entitlement:
    """Permission decision with optional quota information.

    grant = entitlement.Entitlement(allowed=True, limit=5, used=4, remaining=1)
    """

    allowed: bool = False
    unlimited: bool = False
    limit: int = 0
    used: int = 0
    remaining: int = 0
    reason: str = ""

    def near_limit(self, threshold: float) -> bool:
        """Return True when usage meets or exceeds the threshold.

        grant.near_limit(0.8)
        """

        if self.unlimited or self.limit == 0:
            return False
        return (self.used / self.limit) >= threshold

    def usage_percent(self) -> float:
        """Return current usage as a percentage of the limit.

        grant.usage_percent()
        """

        if self.limit == 0:
            return 0.0
        return (self.used / self.limit) * 100.0


def new(
    allowed: bool = False,
    unlimited: bool = False,
    limit: int = 0,
    used: int = 0,
    remaining: int | None = None,
    reason: str = "",
) -> Entitlement:
    """Create an Entitlement value.

    entitlement.new(True, False, 5, 4, 1, "")
    """

    remaining_value = limit - used if remaining is None else int(remaining)
    return Entitlement(
        allowed=bool(allowed),
        unlimited=bool(unlimited),
        limit=int(limit),
        used=int(used),
        remaining=remaining_value,
        reason=str(reason),
    )


def near_limit(entitlement_value: Entitlement, threshold: float) -> bool:
    """Return True when the entitlement is near its limit.

    entitlement.near_limit(grant, 0.8)
    """

    return entitlement_value.near_limit(threshold)


def usage_percent(entitlement_value: Entitlement) -> float:
    """Return current usage percentage for an entitlement.

    entitlement.usage_percent(grant)
    """

    return entitlement_value.usage_percent()


__all__ = ["Entitlement", "near_limit", "new", "usage_percent"]
