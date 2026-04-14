"""Signal-processing helpers for Tier 1-friendly filtering and transforms.

from core.math import signal

smoothed = signal.moving_average([1, 3, 6, 10], window=2)
delta = signal.difference([1, 3, 6, 10])
"""

from __future__ import annotations

from ._shared import Number, _float_values


def moving_average(values: list[Number] | tuple[Number, ...], window: int = 1) -> list[float]:
    """Return a trailing moving average for each sample.

    signal.moving_average([1, 3, 6, 10], window=2)
    """

    if window <= 0:
        raise ValueError("window must be positive")

    items = _float_values(values, allow_empty=True)
    if not items:
        return []

    result: list[float] = []
    total = 0.0
    for index, value in enumerate(items):
        total += value
        if index >= window:
            total -= items[index - window]

        sample_count = min(index + 1, window)
        result.append(total / sample_count)
    return result


def difference(values: list[Number] | tuple[Number, ...], lag: int = 1) -> list[float]:
    """Return the finite difference transform for a sequence.

    signal.difference([1, 3, 6, 10])
    """

    if lag <= 0:
        raise ValueError("lag must be positive")

    items = _float_values(values, allow_empty=True)
    if lag >= len(items):
        return []
    return [items[index] - items[index - lag] for index in range(lag, len(items))]


__all__ = ["difference", "moving_average"]
