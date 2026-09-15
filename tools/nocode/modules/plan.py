#!/usr/bin/env python3
"""Pack the 98 lectures into ~25 minute modules and write plan.json.

    /usr/bin/python3 plan.py            # print the plan
    /usr/bin/python3 plan.py --write    # print it and write plan.json

A module never straddles a `block` (see spine.py): for Weeks 1-2 that is the
teaching Day, for Week 3 a thematic run. Inside a block the split is a DP over
contiguous partitions that minimises squared deviation from the block's own
average, subject to no module exceeding MAX once the cards are counted - so a
64-minute day comes out as three balanced modules rather than 30 + 30 + 4.

Runtime here is the FINISHED runtime: the title card, one chapter card per
lecture and the end card, less the half second every dissolve eats.
"""
from __future__ import annotations

import json
import math
import os
import subprocess
import sys

from spine import SPINE, WEEK_TITLE

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
VIDS = os.path.join(ROOT, "videos", "nocode", "vids")

TARGET = 1500.0   # 25 minutes
MAX    = 1800.0   # the hard ceiling he set
TITLE  = 5.5      # module title card
CHAP   = 2.6      # one chapter card per lecture
END    = 4.5      # end card
XF     = 0.5      # every join is a dissolve, and eats this much


def dur(stem: str) -> float:
    p = os.path.join(VIDS, f"{stem}_adam.mp4")
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def overhead(n: int) -> float:
    """Cards, less what the dissolves eat. n = lectures in the module."""
    # pieces = title + n*(chapter, lecture) + end  ->  2n+2 pieces, 2n+1 joins
    return TITLE + n * CHAP + END - XF * (2 * n + 1)


def runtime(secs: list[float]) -> float:
    return sum(secs) + overhead(len(secs))


def split(secs: list[float], k: int) -> list[list[int]]:
    """Contiguous partition into exactly k parts, balanced, none over MAX."""
    n = len(secs)
    avg = runtime(secs) / k
    INF = float("inf")
    # best[i][j] = cost of partitioning the first i lectures into j parts
    best = [[INF] * (k + 1) for _ in range(n + 1)]
    back = [[-1] * (k + 1) for _ in range(n + 1)]
    best[0][0] = 0.0
    for j in range(1, k + 1):
        for i in range(1, n + 1):
            for a in range(j - 1, i):
                if best[a][j - 1] == INF:
                    continue
                r = runtime(secs[a:i])
                if r > MAX:
                    continue
                c = best[a][j - 1] + (r - avg) ** 2
                if c < best[i][j]:
                    best[i][j] = c
                    back[i][j] = a
    if best[n][k] == INF:
        return []
    out, i = [], n
    for j in range(k, 0, -1):
        a = back[i][j]
        out.append(list(range(a, i)))
        i = a
    return out[::-1]


def build_plan() -> list[dict]:
    rows = [dict(stem=s[0], week=s[1], block=s[2], blocktitle=s[3], title=s[4],
                 secs=dur(s[0])) for s in SPINE]
    blocks: list[tuple[str, list[dict]]] = []
    for r in rows:
        if not blocks or blocks[-1][0] != r["block"]:
            blocks.append((r["block"], []))
        blocks[-1][1].append(r)

    mods, num = [], 0
    for key, lecs in blocks:
        secs = [l["secs"] for l in lecs]
        k = max(1, math.ceil(runtime(secs) / MAX))
        parts = []
        while k <= len(lecs):
            parts = split(secs, k)
            if parts:
                break
            k += 1
        if not parts:                       # one lecture longer than MAX alone
            parts = [[i] for i in range(len(lecs))]
        for p in parts:
            num += 1
            picked = [lecs[i] for i in p]
            mods.append(dict(
                n=num, week=picked[0]["week"], block=key,
                blocktitle=picked[0]["blocktitle"],
                lectures=[dict(stem=l["stem"], title=l["title"], secs=l["secs"])
                          for l in picked],
                secs=runtime([l["secs"] for l in picked])))
    return mods


def hhmm(s: float) -> str:
    return f"{int(s // 60):3d}:{int(s % 60):02d}"


def main() -> int:
    mods = build_plan()
    week = None
    for m in mods:
        if m["week"] != week:
            week = m["week"]
            print(f"\n=== WEEK {week} — {WEEK_TITLE[week]} ===")
        flag = "  <-- OVER" if m["secs"] > MAX else ""
        print(f"M{m['n']:02d}  {m['blocktitle']:24} {hhmm(m['secs'])}  "
              f"{len(m['lectures'])} lec  "
              f"{', '.join(l['stem'] for l in m['lectures'])}{flag}")
    tot = sum(m["secs"] for m in mods)
    src = sum(l["secs"] for m in mods for l in m["lectures"])
    print(f"\n{len(mods)} modules   finished {int(tot//3600)}h{int(tot%3600//60):02d}m"
          f"   from {int(src//3600)}h{int(src%3600//60):02d}m of lectures"
          f"   (+{tot - src:.0f}s of cards)")
    print(f"shortest {hhmm(min(m['secs'] for m in mods))}"
          f"   longest {hhmm(max(m['secs'] for m in mods))}"
          f"   mean {hhmm(tot / len(mods))}")
    if "--write" in sys.argv:
        with open(os.path.join(HERE, "plan.json"), "w") as fh:
            json.dump(mods, fh, indent=1)
        print(f"-> {os.path.join(HERE, 'plan.json')}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
