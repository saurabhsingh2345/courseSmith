#!/usr/bin/env python3
"""Offline IRT item calibration (workstream D).

The coursesmith-tutor /irt/calibrate endpoint returns neutral-prior stubs. This
script does the real offline calibration the endpoint delegates to: given pooled
per-student responses to a shared set of questions, it estimates each item's
difficulty and discrimination, which feed the quiz-strategy difficulty
distribution (workstream G).

It is a dependency-free 1PL-style estimate:
  - difficulty  = logit(1 - p_correct)  — an item everyone gets right is easy
    (low difficulty); one everyone misses is hard (high difficulty).
  - discrimination = the point-biserial correlation between getting the item
    right and the student's overall score, scaled into [0, 2]. Items that only
    strong students get right discriminate well; items uncorrelated with ability
    (or negatively — a bad item) score low.

For a full 2PL/3PL fit swap in py-irt (see README / pyproject.toml); this keeps
the calibration shape real and testable until then.

Usage:
    python3 calibrate_irt.py responses.json   # -> calibrated items JSON
    python3 calibrate_irt.py --selftest

Input JSON: {"students": [{"q1": true, "q2": false, "q3": true}, ...]}
            each object is one student's answers keyed by question id.
Output JSON: {"calibrated": [{"question_id":..,"difficulty":..,
              "discrimination":..,"p_correct":..,"n":..}, ...], "note":..}
"""
from __future__ import annotations

import json
import math
import sys
from typing import Dict, List, Sequence


def _logit(p: float) -> float:
    p = min(max(p, 1e-3), 1 - 1e-3)  # clamp away from the asymptotes
    return math.log(p / (1 - p))


def _pearson(xs: Sequence[float], ys: Sequence[float]) -> float:
    n = len(xs)
    if n < 2:
        return 0.0
    mx = sum(xs) / n
    my = sum(ys) / n
    num = sum((x - mx) * (y - my) for x, y in zip(xs, ys))
    dx = math.sqrt(sum((x - mx) ** 2 for x in xs))
    dy = math.sqrt(sum((y - my) ** 2 for y in ys))
    if dx == 0 or dy == 0:
        return 0.0
    return num / (dx * dy)


def calibrate_irt(students: Sequence[Dict[str, bool]]) -> dict:
    """Estimate per-item difficulty + discrimination from pooled responses."""
    rows = [s for s in students if s]
    if not rows:
        raise ValueError("no student responses to calibrate")

    # Total score per student (proportion of their answered items correct).
    scores = [sum(1 for v in s.values() if v) / max(1, len(s)) for s in rows]

    # Preserve first-seen question order for stable output.
    order: List[str] = []
    seen = set()
    for s in rows:
        for qid in s:
            if qid not in seen:
                seen.add(qid)
                order.append(qid)

    items = []
    for qid in order:
        corr = [(1.0 if s[qid] else 0.0) for s in rows if qid in s]
        stud = [scores[i] for i, s in enumerate(rows) if qid in s]
        n = len(corr)
        p_correct = sum(corr) / n if n else 0.0
        difficulty = _logit(1 - p_correct)  # harder item -> higher difficulty
        # point-biserial correlation, scaled to [0, 2]; negative -> 0 (bad item).
        disc = max(0.0, _pearson(corr, stud)) * 2.0
        items.append({
            "question_id": qid,
            "difficulty": round(max(-2.0, min(2.0, difficulty)), 3),
            "discrimination": round(min(2.0, disc), 3),
            "p_correct": round(p_correct, 3),
            "n": n,
        })

    return {
        "calibrated": items,
        "note": "1PL difficulty + point-biserial discrimination (dependency-free). Swap in py-irt for a full 2PL/3PL fit.",
    }


def _selftest() -> int:
    # An easy item (most right) must have lower difficulty than a hard one.
    # A discriminating item tracks student ability; a coin-flip item does not.
    students = []
    # 10 strong students, 10 weak students.
    for _ in range(10):
        students.append({"easy": True, "hard": True, "disc": True, "noise": True})
    for _ in range(10):
        students.append({"easy": True, "hard": False, "disc": False, "noise": True})
    res = calibrate_irt(students)
    by = {it["question_id"]: it for it in res["calibrated"]}

    assert by["easy"]["difficulty"] < by["hard"]["difficulty"], "easy item should be less difficult"
    assert by["disc"]["discrimination"] > by["noise"]["discrimination"], \
        "an ability-tracking item should out-discriminate a constant one"
    assert by["noise"]["discrimination"] == 0.0, "a constant item cannot discriminate"

    try:
        calibrate_irt([])
        assert False, "expected ValueError on empty input"
    except ValueError:
        pass

    print("calibrate_irt selftest OK:", json.dumps(res["calibrated"]))
    return 0


def main(argv: List[str]) -> int:
    if len(argv) == 2 and argv[1] == "--selftest":
        return _selftest()
    if len(argv) != 2:
        print(__doc__, file=sys.stderr)
        return 2
    with open(argv[1]) as f:
        data = json.load(f)
    print(json.dumps(calibrate_irt(data.get("students", []))))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
