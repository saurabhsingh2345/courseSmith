#!/usr/bin/env python3
"""Buy the voice for a lecture now, and render it whenever.

    python3 prefetch.py --plan w3-17 w3-08
    python3 prefetch.py w3-17 w3-08
    python3 prefetch.py --upgrades

Two problems this solves, and they are different problems.

**The allowance is shared.** It is one key, on two machines, against a monthly
character limit. Watching it drop by eight thousand characters in an hour with
nothing running locally is enough of a reason not to leave the buying until the
rendering. Every sentence spoken here lands in the content store keyed by a hash
of its exact text, permanently, so the render afterwards costs nothing whatever
the allowance has done in the meantime.

**Rendering competes with the camera.** `assemble.py` re-encodes every segment
with h264_videotoolbox, which is the same hardware encoder a rolling capture is
using. This only makes network calls and writes mp3s, so it is safe to run while
a capture is recording - which is exactly when there is time to run it.
"""

from __future__ import annotations

import json
import os
import sys
import urllib.request

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import adam_lay as A  # noqa: E402
import quota  # noqa: E402

NARR = os.path.join(HERE, "narration")
CUTS = os.path.join(HERE, "cuts")

# The twenty lectures that already ship, whose narration was deepened. Cheapest
# characters in the project: most of their sentences are already paid for.
UPGRADES = ["w3-01", "w3-02", "w3-03", "w3-04", "w3-05", "w3-06", "w3-07",
            "w3-08", "w3-12", "w3-13", "w3-15", "w3-16", "w3-17", "w3-18",
            "w3-19", "w3-20", "w3-23", "w3-24", "w3-25", "w3-26"]


def allowance() -> int:
    req = urllib.request.Request(
        "https://api.elevenlabs.io/v1/user/subscription",
        headers={"xi-api-key": A.KEY})
    sub = json.load(urllib.request.urlopen(req, timeout=30))
    return sub["character_limit"] - sub["character_count"]


def main() -> None:
    plan = "--plan" in sys.argv
    lecs = [a for a in sys.argv[1:] if a.startswith("w3-")]
    if "--upgrades" in sys.argv or not lecs:
        lecs = UPGRADES

    have = quota.spoken()
    left = allowance()
    print(f"allowance {left}\n")

    todo = []
    for lec in lecs:
        p = os.path.join(NARR, f"{lec}.json")
        if not os.path.exists(p):
            print(f"{lec}: no narration")
            continue
        d = json.load(open(p))
        new = [s for sg in d["segments"] for s in A.sentences(sg["say"])
               if s not in have]
        cost = sum(len(s) for s in new)
        print(f"{lec}  {len(new):4d} new sentence(s)  {cost:6d} chars")
        todo.append((lec, new, cost))

    total = sum(c for _, _, c in todo)
    print(f"\ntotal {total} chars, leaving about {left - total}")
    if plan:
        return
    if total > left:
        sys.exit("NOT ENOUGH ALLOWANCE - narrow the list rather than "
                 "half-buying a lecture")

    spent = 0
    raw = os.path.join(HERE, "_assemble", "_prefetch")
    os.makedirs(raw, exist_ok=True)
    for lec, new, _ in todo:
        for i, sent in enumerate(new):
            A.speak(sent, os.path.join(raw, f"{lec}_{i:04d}.mp3"))
            spent += len(sent)
            print(f"  {lec} {i + 1}/{len(new)}  {spent}/{total}", flush=True)
    print(f"\nbought {spent} chars; allowance now about {allowance()}")
    print("every one of those sentences is now in the content store - "
          "assemble.py will not pay for it again")


if __name__ == "__main__":
    main()
