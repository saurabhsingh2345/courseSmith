#!/usr/bin/env python3
"""Offline BKT parameter fitting (workstream D).

The coursesmith-tutor service does *online* mastery estimation with hand-set
BKT defaults (internal/adaptive/engine.go). This script does the *offline* other
half: given pooled per-student response sequences for one concept, it fits the
four BKT parameters (init/learn/slip/guess) that best explain the data, so the
service's defaults can be replaced with calibrated values.

It is a coarse maximum-likelihood grid search over the same forward recurrence
the Go engine uses — dependency-free so it runs anywhere with a stock Python.
For production-scale fitting swap in pyBKT (see README / pyproject.toml); this
keeps the integration shape real and testable until then.

Usage:
    python3 fit_bkt.py responses.json      # -> fitted params JSON on stdout
    python3 fit_bkt.py --selftest          # -> runs the built-in checks

Input JSON: {"sequences": [[true, false, true], [true, true], ...]}
            each inner list is one student's correct/incorrect answers, in order.
Output JSON: {"p_init":..,"p_learn":..,"p_slip":..,"p_guess":..,
              "log_likelihood":.., "sequences_fit":N}
"""
from __future__ import annotations

import json
import sys
from typing import List, Sequence, Tuple


def sequence_log_likelihood(seq: Sequence[bool], p_init: float, p_learn: float,
                            p_slip: float, p_guess: float) -> float:
    """Log P(response sequence | params) under standard BKT.

    Mirrors the forward recurrence in internal/adaptive/engine.go: at each step
    predict P(correct), accumulate its log, then apply the Bayesian update and
    the learning transition.
    """
    import math

    pk = p_init
    ll = 0.0
    for correct in seq:
        p_correct = pk * (1 - p_slip) + (1 - pk) * p_guess
        # Guard against log(0) for degenerate parameter corners.
        p_obs = p_correct if correct else (1 - p_correct)
        ll += math.log(max(p_obs, 1e-9))
        if correct:
            num = pk * (1 - p_slip)
            posterior = num / (num + (1 - pk) * p_guess)
        else:
            num = pk * p_slip
            posterior = num / (num + (1 - pk) * (1 - p_guess))
        pk = posterior + (1 - posterior) * p_learn
    return ll


def _grid(lo: float, hi: float, n: int) -> List[float]:
    if n <= 1:
        return [(lo + hi) / 2]
    step = (hi - lo) / (n - 1)
    return [lo + step * i for i in range(n)]


def fit_bkt(sequences: Sequence[Sequence[bool]], steps: int = 6) -> dict:
    """Coarse-to-fine grid search for the max-likelihood BKT parameters.

    Slip and guess are constrained below 0.5 (the identifiability convention:
    a knower answers right more often than a guesser).
    """
    seqs = [list(s) for s in sequences if len(s) > 0]
    if not seqs:
        raise ValueError("no non-empty response sequences to fit")

    # Search ranges; slip/guess capped at 0.45 for identifiability.
    ranges = {"p_init": (0.01, 0.99), "p_learn": (0.01, 0.6),
              "p_slip": (0.01, 0.45), "p_guess": (0.01, 0.45)}
    best = {k: (lo + hi) / 2 for k, (lo, hi) in ranges.items()}

    def total_ll(params: dict) -> float:
        return sum(sequence_log_likelihood(s, params["p_init"], params["p_learn"],
                                           params["p_slip"], params["p_guess"]) for s in seqs)

    best_ll = total_ll(best)
    span = {k: (hi - lo) / 2 for k, (lo, hi) in ranges.items()}

    # Coarse-to-fine: each round searches a 5-point grid per axis around the
    # current best, then halves the search radius. Cheap and deterministic.
    for _ in range(steps):
        improved = True
        while improved:
            improved = False
            for axis in ("p_init", "p_learn", "p_slip", "p_guess"):
                lo = max(ranges[axis][0], best[axis] - span[axis])
                hi = min(ranges[axis][1], best[axis] + span[axis])
                for v in _grid(lo, hi, 5):
                    cand = dict(best)
                    cand[axis] = v
                    ll = total_ll(cand)
                    if ll > best_ll + 1e-9:
                        best_ll, best, improved = ll, cand, True
        for k in span:
            span[k] /= 2

    return {
        "p_init": round(best["p_init"], 4),
        "p_learn": round(best["p_learn"], 4),
        "p_slip": round(best["p_slip"], 4),
        "p_guess": round(best["p_guess"], 4),
        "log_likelihood": round(best_ll, 4),
        "sequences_fit": len(seqs),
    }


def _selftest() -> int:
    import math

    # 1. Likelihood is a proper probability model: exp(sum) over both outcomes
    #    of a one-step sequence sums to 1 for fixed params.
    ll_t = sequence_log_likelihood([True], 0.3, 0.1, 0.1, 0.2)
    ll_f = sequence_log_likelihood([False], 0.3, 0.1, 0.1, 0.2)
    assert abs(math.exp(ll_t) + math.exp(ll_f) - 1.0) < 1e-9, "one-step probs must sum to 1"

    # 2. Recovery: data generated from mostly-correct learners should fit to a
    #    high learn/init and low slip, and beat a deliberately-bad param set.
    seqs = [[False, True, True, True, True, True]] * 20 + [[True, True, True, True]] * 20
    fit = fit_bkt(seqs)
    good_ll = sum(sequence_log_likelihood(s, fit["p_init"], fit["p_learn"],
                                          fit["p_slip"], fit["p_guess"]) for s in seqs)
    bad_ll = sum(sequence_log_likelihood(s, 0.5, 0.01, 0.45, 0.45) for s in seqs)
    assert good_ll > bad_ll, "fit should beat a bad baseline"
    assert 0 <= fit["p_slip"] <= 0.45 and 0 <= fit["p_guess"] <= 0.45

    # 3. Empty input is rejected.
    try:
        fit_bkt([])
        assert False, "expected ValueError on empty input"
    except ValueError:
        pass

    print("fit_bkt selftest OK:", json.dumps(fit))
    return 0


def main(argv: List[str]) -> int:
    if len(argv) == 2 and argv[1] == "--selftest":
        return _selftest()
    if len(argv) != 2:
        print(__doc__, file=sys.stderr)
        return 2
    with open(argv[1]) as f:
        data = json.load(f)
    sequences = data.get("sequences", [])
    print(json.dumps(fit_bkt(sequences)))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
