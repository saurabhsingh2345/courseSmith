#!/usr/bin/env python3
"""Push overflowing sentences forward until every segment fits its footage.

    python3 balance.py w3-17            # report only
    python3 balance.py w3-17 --write
    python3 balance.py --all --write

`fit.py` says which segments outrun their footage; `assemble.py` responds to one
by holding on the last frame, which reads as a freeze. Both are avoidable
without cutting the writing: a segment that is long by two sentences is almost
always followed by one with room, because the deepening pass appends reflective
asides at the end of a beat and those belong to either side of the boundary.

Whole sentences only, and never across more than one boundary at a time, so the
narration keeps describing the picture it sits over. What will not fit anywhere
is reported, not silently dropped.
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
NARR = os.path.join(HERE, "narration")
CUTS = os.path.join(HERE, "cuts")

WPM = 198.0
MIN_SPEED = 0.92
PAD = 0.55


def secs(words: int) -> float:
    return words / WPM * 60 + PAD


def room(seconds: float) -> float:
    return seconds / MIN_SPEED


# A sentence opening with one of these, left as the last thing in a segment,
# is almost always the first half of a paragraph the trim removed.
ORPHAN = re.compile(
    r"^(And|But|So|Then|There is|There are|Notice|Read|Look|Watch|Which|Now|"
    r"It is worth|Here is|One more|Two things|Three things)\b", re.I)


def sentences(text: str) -> list[str]:
    parts = re.split(r"(?<=[.!?])\s+", text.strip())
    return [p for p in parts if p]


def cut_len(lec: str) -> float:
    out = subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", os.path.join(CUTS, f"{lec}.mp4")],
        capture_output=True, text=True).stdout.strip()
    return float(out)


def balance(lec: str, write: bool, trim: bool = False) -> tuple[int, float, float, int]:
    path = os.path.join(NARR, f"{lec}.json")
    d = json.load(open(path))
    segs = d["segments"]
    total = cut_len(lec)

    bounds = []
    for i, s in enumerate(segs):
        f0 = float(s["anchor"])
        f1 = float(segs[i + 1]["anchor"]) if i + 1 < len(segs) else total
        bounds.append(f1 - f0)

    moved = 0
    for i in range(len(segs) - 1):
        while True:
            have = room(bounds[i])
            want = secs(len(segs[i]["say"].split()))
            if want <= have:
                break
            sent = sentences(segs[i]["say"])
            if len(sent) < 2:
                break
            tail = sent[-1]
            nxt_have = room(bounds[i + 1])
            nxt_want = secs(len(segs[i + 1]["say"].split()))
            if nxt_want + secs(len(tail.split())) > nxt_have:
                break
            segs[i]["say"] = " ".join(sent[:-1])
            segs[i + 1]["say"] = tail + " " + segs[i + 1]["say"]
            moved += 1

    # Second pass: a segment that is still long, followed by one with spare
    # room, is fixed by letting the boundary drift a few seconds later rather
    # than by cutting the writing. More than about ten seconds and the beat
    # would start over the wrong picture, so that is the ceiling.
    nudged = 0.0
    for i in range(len(segs) - 1):
        want = secs(len(segs[i]["say"].split()))
        need = want - room(bounds[i])
        if need <= 0:
            continue
        nxt_spare = room(bounds[i + 1]) - secs(len(segs[i + 1]["say"].split()))
        give = min(need * MIN_SPEED, max(0.0, nxt_spare * MIN_SPEED), 12.0)
        if give < 0.5:
            continue
        segs[i + 1]["anchor"] = round(float(segs[i + 1]["anchor"]) + give, 1)
        bounds[i] += give
        bounds[i + 1] -= give
        nudged += give

    # Last resort: a segment that still will not fit, with nowhere to push and
    # no room to drift, loses trailing sentences. It is the deepening pass's own
    # additions that end up here, so this cuts the newest writing rather than
    # the beat it was attached to - but it is reported either way, because a
    # sentence silently deleted is worse than a freeze you were told about.
    trimmed = 0
    if trim:
        for i, s in enumerate(segs):
            cut_here = 0
            while secs(len(s["say"].split())) > room(bounds[i]):
                sent = sentences(s["say"])
                if len(sent) < 2:
                    break
                s["say"] = " ".join(sent[:-1])
                trimmed += 1
                cut_here += 1
            # A paragraph cut in half leaves its opening clause hanging - "And
            # notice the order those three come in, because it is not
            # arbitrary." with nothing after it. Whatever it was introducing is
            # gone, so the introduction goes too.
            while cut_here:
                sent = sentences(s["say"])
                if len(sent) < 2 or not ORPHAN.match(sent[-1]):
                    break
                s["say"] = " ".join(sent[:-1])
                trimmed += 1

    over = 0.0
    for i, s in enumerate(segs):
        want = secs(len(s["say"].split()))
        have = room(bounds[i])
        if want > have:
            print(f"  {lec} seg {i}: still {want - have:4.1f}s long "
                  f"({int((want - have) * WPM / 60)} words)")
            over += want - have

    if write and (moved or nudged or trimmed):
        json.dump(d, open(path, "w"), indent=2)
    return moved, over, nudged, trimmed


def main() -> None:
    write = "--write" in sys.argv
    if "--all" in sys.argv:
        lecs = sorted(f[:-5] for f in os.listdir(NARR) if f.endswith(".json"))
    else:
        lecs = [a for a in sys.argv[1:] if a.startswith("w3-")]
    for lec in lecs:
        if not os.path.exists(os.path.join(CUTS, f"{lec}.mp4")):
            continue
        moved, over, nudged, trimmed = balance(lec, write, "--trim" in sys.argv)
        flag = "" if over < 1.0 else f"  <- {over:.0f}s still overflowing"
        print(f"{lec}: moved {moved} sentence(s), "
              f"nudged {nudged:.0f}s, trimmed {trimmed}{flag}")


if __name__ == "__main__":
    main()
