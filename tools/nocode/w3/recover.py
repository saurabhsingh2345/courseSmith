#!/usr/bin/env python3
"""Rebuild a capture manifest from the marker log after an interrupted take.

    python3 recover.py C

rig.py flushes every marker to captures/capture-<X>-markers.jsonl the moment it
is made, precisely so a take that never reached Session.stop() is still cuttable.
The segment's end time is taken from its mtime (when ffmpeg stopped writing) and
its start derived backwards from the probed duration — the same arithmetic
Session.stop() uses.
"""
from __future__ import annotations
import glob, json, os, subprocess, sys

HERE = os.path.dirname(os.path.abspath(__file__))
CAPS = os.path.join(HERE, "captures")
TAKES = os.path.join(HERE, "..", "scripts", "takes")


def main() -> None:
    cap = sys.argv[1]
    jl = os.path.join(CAPS, f"capture-{cap}-markers.jsonl")
    markers = [json.loads(l) for l in open(jl) if l.strip()]
    segs = []
    for p in sorted(glob.glob(os.path.join(TAKES, f"w3{cap}_s*.mp4"))):
        d = float(subprocess.run(["ffprobe", "-v", "error", "-show_entries",
                                  "format=duration", "-of", "csv=p=0", p],
                                 capture_output=True, text=True).stdout.strip())
        end = os.path.getmtime(p)
        segs.append({"path": os.path.abspath(p), "dur": d,
                     "t0": end - d, "t1": end})
    placed = []
    for m in markers:
        hit = {"seg": None, "offset": None, "orphan": True}
        for i, sg in enumerate(segs):
            if sg["t0"] <= m["t"] <= sg["t1"]:
                hit = {"seg": i, "offset": round(m["t"] - sg["t0"], 3)}
                break
        placed.append({**m, **hit})
    out = os.path.join(CAPS, f"capture-{cap}.json")
    json.dump({"capture": cap, "t_start": markers[0]["t"],
               "segments": segs, "markers": placed}, open(out, "w"), indent=2)
    orph = sum(1 for m in placed if m.get("orphan"))
    print(f"recovered {len(placed)} markers, {len(segs)} segment(s), "
          f"{sum(s['dur'] for s in segs)/60:.1f} min, {orph} orphan(s) -> {out}")


if __name__ == "__main__":
    main()
