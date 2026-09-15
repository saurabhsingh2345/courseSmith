#!/usr/bin/env python3
"""Render lectures in priority order, stopping before the voice allowance runs out.

    python3 render.py --plan                 # what it would do, spending nothing
    python3 render.py                        # render until the allowance is gone
    python3 render.py w3-33 w3-34 w3-35      # just these, same guard
    python3 render.py --force w3-17          # ignore the guard for one lecture
    python3 render.py --buy                  # allow it to spend, if you mean it

The voice is metered monthly. A render that runs out halfway leaves a lecture
with the first two thirds spoken and the rest silent, and the only way to find
out is to watch it — so the allowance is checked BEFORE each lecture, against
the exact cost of the sentences that are not already in the content store.

Order matters and is not alphabetical: an upgraded lecture that already ships is
worth more than half of a new block, and half a block is worth less than none of
one, because a lecture that refers to a lecture nobody has is worse than a gap.
`ORDER.md` is the source of truth for what belongs to which block.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys
import urllib.request

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import adam_lay as A  # noqa: E402
import quota  # noqa: E402

CUTS = os.path.join(HERE, "cuts")
NARR = os.path.join(HERE, "narration")

# Blocks, in delivery order. A block is all-or-nothing on purpose: half of one
# is worse than none of it, because a lecture that refers to a lecture nobody
# has is worse than a gap.
BLOCKS = [
    ("upgrades", ["w3-01", "w3-02", "w3-03", "w3-04", "w3-05", "w3-06", "w3-07",
                  "w3-08", "w3-12", "w3-13", "w3-15", "w3-16", "w3-17", "w3-18",
                  "w3-19", "w3-20", "w3-23", "w3-24", "w3-25", "w3-26"]),
    ("github", ["w3-33", "w3-34", "w3-35"]),
    ("headless", ["w3-36", "w3-37", "w3-38"]),
    ("second product", ["w3-27", "w3-28", "w3-29", "w3-30", "w3-31", "w3-32"]),
    ("failure and verdict", ["w3-39", "w3-40", "w3-41"]),
]
PRIORITY = [lec for _, block in BLOCKS for lec in block]


def allowance() -> int:
    req = urllib.request.Request(
        "https://api.elevenlabs.io/v1/user/subscription",
        headers={"xi-api-key": A.KEY})
    sub = json.load(urllib.request.urlopen(req, timeout=30))
    return sub["character_limit"] - sub["character_count"]


def cost(lec: str, have: set[str]) -> int:
    p = os.path.join(NARR, f"{lec}.json")
    d = json.load(open(p))
    return sum(len(s) for sg in d["segments"]
               for s in A.sentences(sg["say"]) if s not in have)


def ready(lec: str) -> bool:
    return (os.path.exists(os.path.join(CUTS, f"{lec}.mp4"))
            and os.path.exists(os.path.join(NARR, f"{lec}.json")))


def main() -> None:
    plan_only = "--plan" in sys.argv
    force = "--force" in sys.argv
    buy = "--buy" in sys.argv
    want = [a for a in sys.argv[1:] if a.startswith("w3-")]
    order = want or PRIORITY

    left = allowance()
    print(f"allowance: {left} characters\n")
    have = quota.spoken()

    done, skipped = [], []
    for lec in order:
        if not ready(lec):
            skipped.append((lec, "no cut or no narration yet"))
            continue
        c = cost(lec, have)
        # Rendering does not buy. `prefetch.py` buys, deliberately and in one
        # place, because the allowance is shared with another machine and a
        # render that quietly spends it leaves a block half-voiced with nothing
        # to say so but watching it.
        if c > 0 and not buy and not force:
            skipped.append((lec, f"not prefetched - {c} characters unbought"))
            continue
        if c > left and not force:
            skipped.append((lec, f"needs {c}, only {left} left"))
            continue
        print(f"=== {lec}  ({c} new characters, {left} left)", flush=True)
        if not plan_only:
            r = subprocess.run([sys.executable, "assemble.py", lec], cwd=HERE)
            if r.returncode:
                skipped.append((lec, "assemble failed"))
                continue
            have = quota.spoken()
        left -= c
        done.append(lec)

    print(f"\n{'would render' if plan_only else 'rendered'}: "
          f"{len(done)} lecture(s)")
    for lec in done:
        print("  ", lec)
    if skipped:
        print("\nnot done:")
        for lec, why in skipped:
            print(f"   {lec:8s} {why}")
    print(f"\nallowance after: about {left}")


if __name__ == "__main__":
    main()
