"""Math helpers for Tier 1-friendly statistics and nearest-neighbour search.

from core import math

scores = [0.2, 0.4, 0.9]
average = math.mean(scores)
tree = math.kdtree.build([[0.0, 0.0], [1.0, 1.0]], metric="euclidean")
"""

from __future__ import annotations

import bisect
from dataclasses import dataclass
import math as mathlib
import statistics
from typing import Any, Iterable, Sequence


Number = int | float


def mean(values: Iterable[Number]) -> float:
    """Return the arithmetic mean of numeric values.

    math.mean([0.2, 0.4, 0.9])
    """

    return statistics.fmean(_float_values(values))


def median(values: Iterable[Number]) -> float:
    """Return the median of numeric values.

    math.median([0.2, 0.4, 0.9])
    """

    return float(statistics.median(_float_values(values)))


def variance(values: Iterable[Number]) -> float:
    """Return the population variance of numeric values.

    math.variance([0.2, 0.4, 0.9])
    """

    items = _float_values(values)
    average = statistics.fmean(items)
    return sum((value - average) ** 2 for value in items) / len(items)


def stdev(values: Iterable[Number]) -> float:
    """Return the population standard deviation of numeric values.

    math.stdev([0.2, 0.4, 0.9])
    """

    return mathlib.sqrt(variance(values))


def sort(values: Sequence[Any]) -> list[Any]:
    """Return a sorted copy of the values.

    math.sort([3, 1, 2])
    """

    return sorted(values)


def binary_search(values: Sequence[Any], target: Any) -> int:
    """Return the index of a sorted value or `-1`.

    math.binary_search([1, 2, 3], 2)
    """

    index = bisect.bisect_left(values, target)
    if index >= len(values) or values[index] != target:
        return -1
    return index


def epsilon_equal(left: Number, right: Number, epsilon: float = 1e-9) -> bool:
    """Return True when two numbers are within epsilon.

    math.epsilon_equal(0.1 + 0.2, 0.3)
    """

    return abs(float(left) - float(right)) <= epsilon


def normalize(values: Iterable[Number]) -> list[float]:
    """Scale values into the `[0, 1]` range.

    math.normalize([10, 20, 30])
    """

    items = _float_values(values, allow_empty=True)
    if not items:
        return []
    minimum = min(items)
    maximum = max(items)
    if minimum == maximum:
        return [0.0 for _ in items]
    scale = maximum - minimum
    return [(value - minimum) / scale for value in items]


def rescale(values: Iterable[Number], new_min: float, new_max: float) -> list[float]:
    """Scale values into a target numeric range.

    math.rescale([10, 20, 30], -1.0, 1.0)
    """

    items = _float_values(values, allow_empty=True)
    if not items:
        return []
    minimum = min(items)
    maximum = max(items)
    if minimum == maximum:
        return [float(new_min) for _ in items]
    input_scale = maximum - minimum
    output_scale = new_max - new_min
    return [new_min + (((value - minimum) / input_scale) * output_scale) for value in items]


@dataclass(slots=True)
class KDTree:
    """In-memory nearest-neighbour index for vector points.

    tree = math.kdtree.build([[0.0, 0.0], [1.0, 1.0]])
    """

    points: list[tuple[float, ...]]
    metric: str = "euclidean"

    def nearest(self, query: Sequence[Number], k: int = 1) -> list[dict[str, Any]]:
        """Return the `k` nearest points to the query.

        tree.nearest([0.8, 0.8], k=2)
        """

        return _search(self.points, _point(query), k, self.metric)


class _KDTreeModule:
    def build(self, points: Sequence[Sequence[Number]], metric: str = "euclidean") -> KDTree:
        """Build a KDTree-like handle.

        math.kdtree.build([[0.0, 0.0], [1.0, 1.0]], metric="euclidean")
        """

        return KDTree(points=[_point(point) for point in points], metric=_metric(metric))


class _KNNModule:
    def search(
        self,
        points: Sequence[Sequence[Number]],
        query: Sequence[Number],
        k: int = 1,
        metric: str = "euclidean",
    ) -> list[dict[str, Any]]:
        """Return the `k` nearest points without building a tree handle.

        math.knn.search([[0.0, 0.0], [1.0, 1.0]], [0.8, 0.8], k=1)
        """

        return _search([_point(point) for point in points], _point(query), k, _metric(metric))


kdtree = _KDTreeModule()
knn = _KNNModule()


def _float_values(values: Iterable[Number], *, allow_empty: bool = False) -> list[float]:
    items = [float(value) for value in values]
    if not items and not allow_empty:
        raise ValueError("values must not be empty")
    return items


def _point(values: Sequence[Number]) -> tuple[float, ...]:
    point = tuple(float(value) for value in values)
    if not point:
        raise ValueError("points must not be empty")
    return point


def _search(points: Sequence[tuple[float, ...]], query: tuple[float, ...], k: int, metric: str) -> list[dict[str, Any]]:
    if k <= 0:
        raise ValueError("k must be positive")
    metric_name = _metric(metric)

    neighbours = [
        {
            "index": index,
            "distance": _distance(metric_name, point, query),
            "point": list(point),
        }
        for index, point in enumerate(points)
    ]
    neighbours.sort(key=lambda item: (item["distance"], item["index"]))
    return neighbours[:k]


def _metric(metric: str) -> str:
    metric_name = metric.lower()
    if metric_name not in {"euclidean", "manhattan", "chebyshev", "cosine"}:
        raise ValueError(f"unknown metric: {metric}")
    return metric_name


def _distance(metric: str, left: Sequence[float], right: Sequence[float]) -> float:
    if len(left) != len(right):
        raise ValueError(f"point dimension mismatch: {len(left)} != {len(right)}")

    if metric == "euclidean":
        return mathlib.sqrt(sum((a - b) ** 2 for a, b in zip(left, right)))
    if metric == "manhattan":
        return sum(abs(a - b) for a, b in zip(left, right))
    if metric == "chebyshev":
        return max(abs(a - b) for a, b in zip(left, right))

    dot_product = sum(a * b for a, b in zip(left, right))
    left_norm = mathlib.sqrt(sum(a * a for a in left))
    right_norm = mathlib.sqrt(sum(b * b for b in right))
    if left_norm == 0 and right_norm == 0:
        return 0.0
    if left_norm == 0 or right_norm == 0:
        return 1.0
    return 1.0 - (dot_product / (left_norm * right_norm))


__all__ = [
    "KDTree",
    "binary_search",
    "epsilon_equal",
    "kdtree",
    "knn",
    "mean",
    "median",
    "normalize",
    "rescale",
    "sort",
    "stdev",
    "variance",
]
