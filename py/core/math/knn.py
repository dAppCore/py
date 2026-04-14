"""KNN helpers exposed on the RFC import path.

from core.math.knn import search

neighbours = search([[0.0, 0.0], [1.0, 1.0]], [0.8, 0.8], k=1)
"""

from __future__ import annotations

from typing import Any, Sequence

from ._shared import Number, _metric, _point, _search


def search(
    points: Sequence[Sequence[Number]],
    query: Sequence[Number],
    k: int = 1,
    metric: str = "euclidean",
) -> list[dict[str, Any]]:
    """Return the `k` nearest points without building a tree handle.

    search([[0.0, 0.0], [1.0, 1.0]], [0.8, 0.8], k=1)
    """

    return _search([_point(point) for point in points], _point(query), k, _metric(metric))


__all__ = ["search"]
