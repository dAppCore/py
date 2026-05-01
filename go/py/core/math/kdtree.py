"""KDTree-style nearest-neighbour helpers.

from core.math.kdtree import build

tree = build([[0.0, 0.0], [1.0, 1.0]], metric="euclidean")
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Sequence

from ._shared import Number, _metric, _point, _search


@dataclass(slots=True)
class KDTree:
    """In-memory nearest-neighbour index for vector points.

    tree = build([[0.0, 0.0], [1.0, 1.0]])
    """

    points: list[tuple[float, ...]]
    metric: str = "euclidean"

    def nearest(self, query: Sequence[Number], k: int = 1) -> list[dict[str, Any]]:
        """Return the `k` nearest points to the query.

        tree.nearest([0.8, 0.8], k=2)
        """

        return _search(self.points, _point(query), k, self.metric)


def build(points: Sequence[Sequence[Number]], metric: str = "euclidean") -> KDTree:
    """Build a KDTree-like handle.

    build([[0.0, 0.0], [1.0, 1.0]], metric="euclidean")
    """

    return KDTree(points=[_point(point) for point in points], metric=_metric(metric))


__all__ = ["KDTree", "build"]
