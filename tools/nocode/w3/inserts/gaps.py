#!/usr/bin/env python3
"""What a lecture's plan still leaves on a dead picture.

The first animation pass took Week 3 from 76% frozen to 31%, and what it left
behind is not spread evenly - four lectures own most of it. Authoring a second
pass needs the one thing neither `source.py` nor `brief.py` prints: the dead
spans MINUS the windows the existing plan already covers.

    python3 gaps.py w3-29              # the uncovered spans, with the words over them
    python3 gaps.py --all              # the whole week, one line per lecture
    python3 gaps.py w3-29 --min 8      # include shorter gaps

A gap here is measured against the SAME map the plan was authored from, so the
numbers line up with `render.py`'s `check()`: anything this prints is a window
`check()` will accept, and anything it does not print would be refused.

Runtime is unchanged by a lay, so these timestamps are also the timestamps of
the delivered file - which is why they agree with `qa/deadzones_w3_after.log`.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
import textwrap

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import source as SRC  # noqa: E402
import brief as BRIEF  # noqa: E402

PLANS = os.path.join(HERE, "plans")
NARR = os.path.abspath(os.path.join(HERE, "..", "narration"))
FPS = 30.0
PAD = 0.5   # never start a card on the first frame of a live-to-dead boundary


def plan_windows(lec: str) -> list[tuple[float, float]]:
    p = os.path.join(PLANS, f"{lec}.json")
    if not os.path.exists(p):
        return []
    plan = json.load(open(p))
    out = []
    for item in plan["inserts"]:
        at = float(item["at"])
        n = sum(round(float(s) * FPS) for s in item["secs"])
        out.append((at, at + n / FPS))
    return sorted(out)


def gaps(lec: str, floor: float) -> list[tuple[float, float]]:
    """Dead spans with the plan's windows cut out of them."""
    out = []
    covered = plan_windows(lec)
    for a, b in SRC.spans(lec):
        free = [(a, b)]
        for ca, cb in covered:
            nxt = []
            for fa, fb in free:
                if cb <= fa or ca >= fb:
                    nxt.append((fa, fb))
                    continue
                if ca > fa:
                    nxt.append((fa, min(ca, fb)))
                if cb < fb:
                    nxt.append((max(cb, fa), fb))
            free = nxt
        out += [(fa, fb) for fa, fb in free if fb - fa >= floor]
    return sorted(out)


def lectures() -> list[str]:
    return sorted(f[:-4] for f in os.listdir(os.path.join(HERE, "maps"))
                  if f.endswith(".tsv") and not f.endswith(".align.tsv"))


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("lecture", nargs="?")
    ap.add_argument("--all", action="store_true")
    ap.add_argument("--min", type=float, default=10.0)
    args = ap.parse_args()

    if args.all:
        tot_n = tot_s = 0
        print(f"{'lecture':8} {'gaps':>5} {'dead left':>10} {'runtime':>8} "
              f"{'left%':>6}  {'animated now':>12}")
        for lec in lectures():
            g = gaps(lec, args.min)
            if not g:
                continue
            secs = sum(b - a for a, b in g)
            dur = float(subprocess.run(
                ["ffprobe", "-v", "error", "-show_entries", "format=duration",
                 "-of", "csv=p=0", SRC.source(lec)], capture_output=True,
                text=True, check=True).stdout.strip())
            have = sum(b - a for a, b in plan_windows(lec))
            tot_n += len(g)
            tot_s += secs
            print(f"{lec:8} {len(g):5d} {secs:9.0f}s {dur:7.0f}s "
                  f"{secs / dur * 100:5.0f}% {have / dur * 100:11.0f}%")
        print(f"\n{tot_n} gaps, {tot_s:.0f}s ({tot_s / 60:.0f} min) still on a dead picture")
        return 0

    lec = args.lecture
    if not lec:
        ap.error("a lecture, or --all")
    segs = json.load(open(os.path.join(NARR, f"{lec}.json")))["segments"]
    total = float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", SRC.source(lec)], capture_output=True, text=True,
        check=True).stdout.strip())
    g = gaps(lec, args.min)
    have = sum(b - a for a, b in plan_windows(lec))
    print(f"{lec}   runtime {total:.0f}s   animated {have:.0f}s ({have / total * 100:.0f}%)"
          f"   uncovered dead {sum(b - a for a, b in g):.0f}s in {len(g)} gaps")
    an = SRC.anchors(lec)
    ld = dict(SRC.leads(lec))
    print()
    for a, b in g:
        hit = [x for x in an if a - 1 <= x <= b + 2]
        tag = "   <- anchor " + ", ".join(f"{h:.1f}" for h in hit) if hit else ""
        if any(h in ld for h in hit):
            tag += "  *** END A CARD JUST BEFORE IT ***"
        print(f"[{a:7.1f} - {b:7.1f}]  {b - a:5.1f}s{tag}")
        said = BRIEF.words_over(segs, a, b, total)
        for line in textwrap.wrap(said, 100) or ["(silence)"]:
            print(f"    {line}")
        print()
    return 0


if __name__ == "__main__":
    sys.exit(main())
