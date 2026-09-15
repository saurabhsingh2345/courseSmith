#!/usr/bin/env python3
"""Which file is the current source for a lecture, and its dead-span map.

The chain is orig -> retimed -> staged -> delivered, and getting it wrong is the
one mistake in this pipeline that is not recoverable: cutaway windows planned
against the original picture are wrong the moment `resync.py` has moved that
picture, and a card laid onto an already-carded file cannot be lifted off again.

So there is exactly one place that answers "what am I working from".

    python3 source.py w3-03            # print the source and the map
    python3 source.py w3-03 --remap    # re-measure the dead spans of it
"""

from __future__ import annotations

import argparse
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ORIG = os.path.join(HERE, "orig")
RETIMED = os.path.join(HERE, "retimed")
MAPS = os.path.join(HERE, "maps")
QA = os.path.abspath(os.path.join(HERE, "..", "..", "qa"))
DEADZONES = os.path.join(QA, "deadzones.py")
ALIGN = os.path.join(QA, "align.py")
NARR = os.path.abspath(os.path.join(HERE, "..", "narration"))
VIDS = os.path.abspath(os.path.join(HERE, "..", "..", "..", "..",
                                    "videos", "nocode", "vids"))
MIN_SPAN = 8.0


def source(lec: str) -> str:
    """The picture the cutaways must be planned against."""
    for p in (os.path.join(RETIMED, f"{lec}.mp4"),
              os.path.join(ORIG, f"{lec}_adam.mp4"),
              os.path.join(VIDS, f"{lec}_adam.mp4")):
        if os.path.exists(p):
            return p
    raise SystemExit(f"no source for {lec}")


def map_path(lec: str) -> str:
    return os.path.join(MAPS, f"{lec}.tsv")


def align_path(lec: str) -> str:
    return os.path.join(MAPS, f"{lec}.align.tsv")


def anchors(lec: str) -> list[float]:
    """Every narration anchor: the second each sentence's picture is due."""
    import json
    p = os.path.join(NARR, f"{lec}.json")
    if not os.path.exists(p):
        return []
    return sorted(float(s["anchor"]) for s in json.load(open(p))["segments"])


def leads(lec: str) -> list[tuple[float, float]]:
    """(anchor, lead) for anchors whose picture had already finished printing.

    These are the ones a cutaway has to end on: cover the early reveal, and cut
    back so the picture arrives with the words instead of ahead of them.
    """
    p = align_path(lec)
    if not os.path.exists(p):
        return []
    out = []
    with open(p) as fh:
        for line in fh:
            f = line.rstrip("\n").split("\t")
            if len(f) >= 4 and f[0] == lec and f[1] not in ("anchor", "ERROR"):
                out.append((float(f[1]), float(f[3])))
    return sorted(out)


def remap(lec: str) -> str:
    os.makedirs(MAPS, exist_ok=True)
    dst = map_path(lec)
    with open(dst, "w") as fh:
        subprocess.run([sys.executable, DEADZONES, "--min", str(MIN_SPAN),
                        source(lec)], stdout=fh, check=True)
    with open(align_path(lec), "w") as fh:
        subprocess.run([sys.executable, ALIGN, source(lec)], stdout=fh, check=True)
    return dst


def spans(lec: str) -> list[tuple[float, float]]:
    p = map_path(lec)
    if not os.path.exists(p):
        return []
    out = []
    with open(p) as fh:
        for line in fh:
            f = line.rstrip("\n").split("\t")
            if len(f) >= 4 and f[1] not in ("ERROR", "start"):
                out.append((float(f[1]), float(f[2])))
    return sorted(out)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("lecture")
    ap.add_argument("--remap", action="store_true")
    args = ap.parse_args()
    if args.remap:
        remap(args.lecture)
    src = source(args.lecture)
    sp = spans(args.lecture)
    print(f"{args.lecture}\n  source: {os.path.relpath(src, HERE)}\n"
          f"  map:    {os.path.relpath(map_path(args.lecture), HERE)}  "
          f"({len(sp)} dead spans, {sum(b - a for a, b in sp):.0f}s)")
    ld = leads(args.lecture)
    an = anchors(args.lecture)
    print(f"  anchors: {', '.join(f'{a:.1f}' for a in an)}")
    if ld:
        print("  footage leads the voice at:")
        for a, lead in ld:
            print(f"    anchor {a:7.1f}  picture finished {lead:5.1f}s earlier"
                  f"   -> a cutaway must end just before {a:.1f}")
    for a, b in sp:
        hit = [x for x in an if a - 1 <= x <= b + 2]
        note = f"   anchor(s) {', '.join(f'{h:.1f}' for h in hit)}" if hit else ""
        print(f"    {a:8.1f} - {b:8.1f}   ({b - a:5.1f}s){note}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
