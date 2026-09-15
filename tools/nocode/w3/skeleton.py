#!/usr/bin/env python3
"""Emit a narration skeleton for a lecture, with one segment per beat.

    python3 skeleton.py w3-02

Writes narration/w3-02.json with the anchors already filled in from the capture
markers and an empty `say` for each. Fill the sentences in against the frames
that frames.py sampled — never before.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import frames  # noqa: E402

CUTS = os.path.join(HERE, "cuts")
NARR = os.path.join(HERE, "narration")


def main() -> None:
    if len(sys.argv) < 2:
        sys.exit(__doc__)
    lec = sys.argv[1]
    src = os.path.join(CUTS, f"{lec}.mp4")
    if not os.path.exists(src):
        sys.exit(f"no cut yet: {src}")
    total = float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", src], capture_output=True, text=True).stdout.strip())

    beats = [(t, n) for t, n in frames.beats_for(lec) if t < total]
    if not beats:
        beats = [(0.0, "open")]

    os.makedirs(NARR, exist_ok=True)
    dst = os.path.join(NARR, f"{lec}.json")
    if os.path.exists(dst):
        sys.exit(f"{dst} already exists — not overwriting your writing")

    json.dump({"lecture": lec, "cut_seconds": round(total, 1),
               "segments": [{"anchor": t, "beat": n, "say": ""} for t, n in beats]},
              open(dst, "w"), indent=2)
    print(f"{lec}: {total/60:.1f} min, {len(beats)} beat(s) -> {dst}")
    for t, n in beats:
        print(f"   {int(t//60):02d}:{int(t%60):02d}  {n}")


if __name__ == "__main__":
    main()
