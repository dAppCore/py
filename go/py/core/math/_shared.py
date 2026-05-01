from __future__ import annotations

import math as mathlib
from typing import Any, Iterable, Sequence


Number = int | float


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
