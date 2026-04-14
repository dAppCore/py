"""Math helpers for Tier 1-friendly statistics and nearest-neighbour search.

from core import math
from core.math import kdtree, knn

scores = [0.2, 0.4, 0.9]
average = math.mean(scores)
tree = kdtree.build([[0.0, 0.0], [1.0, 1.0]], metric="euclidean")
"""

from __future__ import annotations

import bisect
import math as mathlib
import statistics
from typing import Any, Iterable, Sequence

from . import kdtree, knn
from ._shared import Number, _float_values
from .kdtree import KDTree


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
