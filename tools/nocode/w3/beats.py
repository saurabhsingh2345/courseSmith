#!/usr/bin/env python3
"""Print each lecture's beats as offsets into its own cut.

    python3 beats.py G
    python3 beats.py E,E2 w3-32

The capture manifest stores markers in wall-clock time. `cut.py` turns a lecture
into a file that starts at its first marker minus PAD_IN, so every anchor a
narration file needs is `marker - first_marker + PAD_IN`. Working that out by
hand for nine lectures is how an anchor ends up eleven seconds off the thing it
describes.

This prints anchors that can be pasted straight into a narration file, and it
works before the cut exists - which is the point, because it means the writing
can start while the next capture is still rolling.
"""

from __future__ import annotations

import json
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
CAPS = os.path.join(HERE, "captures")
PAD_IN = 0.6


def load(spec: str) -> list[dict]:
    out = []
    for name in spec.split(","):
        p = os.path.join(CAPS, f"capture-{name.strip()}.json")
        if not os.path.exists(p):
            sys.exit(f"no manifest: {p}")
        out.append(json.load(open(p)))
    return out


def main() -> None:
    if len(sys.argv) < 2:
        sys.exit(__doc__)
    mans = load(sys.argv[1])
    want = [a for a in sys.argv[2:] if a.startswith("w3-")]

    # Same rule as cut.py: a lecture owns every marker whose label starts with
    # its id, in capture order, and its footage runs to the first marker
    # belonging to a different lecture.
    spans: dict[str, list] = {}
    for man in mans:
        mk = sorted([m for m in man["markers"] if m["label"].startswith("w3-")],
                    key=lambda m: m["t"])
        hard = max(s["t1"] for s in man["segments"]) if man["segments"] else 0.0
        i = 0
        while i < len(mk):
            lec = mk[i]["label"].split("/", 1)[0]
            j = i
            while j < len(mk) and mk[j]["label"].split("/", 1)[0] == lec:
                j += 1
            t0 = mk[i]["t"]
            t1 = mk[j]["t"] if j < len(mk) else hard
            spans.setdefault(lec, []).append((t0, t1, mk[i:j]))
            i = j

    for lec in sorted(spans):
        if want and lec not in want:
            continue
        base = 0.0
        total = sum(t1 - t0 for t0, t1, _ in spans[lec])
        print(f"\n== {lec}   {total/60:.1f} min   cut_seconds {total + PAD_IN:.1f}")
        for t0, t1, beats in spans[lec]:
            for m in beats:
                off = base + (m["t"] - t0) + PAD_IN
                print(f'  {{"anchor": {off:.1f}, "beat": '
                      f'"{m["label"].split("/", 1)[1]}", "say": ""}},'
                      f'   # {m.get("note", "")}')
            base += t1 - t0


if __name__ == "__main__":
    main()
