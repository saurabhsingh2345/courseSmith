#!/usr/bin/env python3
"""Turn one long capture into per-lecture source clips.

    python3 cut.py A              # cut every lecture found in capture A
    python3 cut.py A w3-03        # just this one
    python3 cut.py A --list       # show what the markers say, cut nothing
    python3 cut.py A,A1b          # merge a depth pass into the same lectures

Several captures can feed one lecture. Asking an agent to do something takes
twenty seconds where the reference instructor types for ten minutes, so a first
pass through a lecture produces far less footage than the lecture needs. A
second capture re-using the same `w3-NN/` marker prefix is appended in capture
order, which is how a thin lecture gets its depth without a reshoot.

A lecture owns every marker whose label starts `w3-NN/`. Its footage runs from
its first marker to the first marker belonging to a DIFFERENT lecture — so the
shot plan only has to mark where each beat begins, never where it ends.

Work is done in wall-clock time, then mapped onto segments, because a capture
pauses whenever he touches another app: one lecture can straddle two segments
with a gap in between, and the gap must not appear in the cut.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
CAPS = os.path.join(HERE, "captures")
CUTS = os.path.join(HERE, "cuts")
PAD_IN = 0.6    # keep a breath before a beat starts
PAD_OUT = 1.2   # and after it ends, so a cut never lands on a keystroke
COPY = "--reencode" not in sys.argv


def load(cap: str) -> list[dict]:
    """One or more manifests, in the order given."""
    out = []
    for name in cap.split(","):
        p = os.path.join(CAPS, f"capture-{name.strip()}.json")
        if not os.path.exists(p):
            sys.exit(f"no manifest: {p}")
        out.append(json.load(open(p)))
    return out


def lecture_of(label: str) -> str | None:
    head = label.split("/", 1)[0].strip()
    return head if head.startswith("w3-") else None


def ranges(man: dict) -> list:
    ms = [m for m in man["markers"] if lecture_of(m["label"])]
    ms.sort(key=lambda m: m["t"])
    segs = man["segments"]
    if not segs:
        return []
    hard_end = max(s["t1"] for s in segs)

    out, i = [], 0
    while i < len(ms):
        lec = lecture_of(ms[i]["label"])
        j = i
        beats = []
        while j < len(ms) and lecture_of(ms[j]["label"]) == lec:
            beats.append(ms[j]["label"])
            j += 1
        t0 = ms[i]["t"]
        t1 = ms[j]["t"] if j < len(ms) else hard_end
        out.append((lec, t0, t1, beats))
        i = j

    # A lecture shot in two passes appears twice; merge on the id.
    merged: dict[str, list] = {}
    for lec, t0, t1, beats in out:
        if lec in merged:
            merged[lec].append((t0, t1, beats))
        else:
            merged[lec] = [(t0, t1, beats)]
    return [(lec, spans) for lec, spans in merged.items()]


def slices(man: dict, t0: float, t1: float) -> list[tuple[str, float, float]]:
    """Map a wall-clock window onto (segment path, ss, to) pieces."""
    out = []
    for sg in man["segments"]:
        a, b = max(t0, sg["t0"]), min(t1, sg["t1"])
        if b - a < 0.35:
            continue
        out.append((sg["path"], round(a - sg["t0"], 3), round(b - sg["t0"], 3)))
    return out


def cut_one(mans: list, lec: str, spans: list) -> str | None:
    os.makedirs(CUTS, exist_ok=True)
    pieces, work = [], os.path.join(CUTS, "_work")
    os.makedirs(work, exist_ok=True)
    n = 0
    for man, t0, t1, _ in spans:
        for path, ss, to in slices(man, t0 - PAD_IN, t1 + PAD_OUT):
            n += 1
            dst = os.path.join(work, f"{lec}_p{n:02d}.mp4")
            # --copy is the default and is 400x faster: the capture is written
            # with g=15, so a keyframe every half second, and the seek error a
            # stream copy can introduce is smaller than PAD_IN. Re-encoding an
            # hour of capture costs an hour of the same hardware encoder the
            # next take needs. Re-encode only when something downstream needs
            # frame-exact boundaries.
            enc = (["-c", "copy"] if COPY else
                   ["-c:v", "h264_videotoolbox", "-b:v", "40M",
                    "-g", "15", "-keyint_min", "15", "-pix_fmt", "yuv420p"])
            subprocess.run(["ffmpeg", "-y", "-loglevel", "error",
                            "-ss", f"{ss}", "-to", f"{to}", "-i", path,
                            *enc, "-an", dst], check=True)
            pieces.append(dst)
    if not pieces:
        print(f"  {lec}: NO FOOTAGE (markers fell entirely in a pause)")
        return None

    dst = os.path.join(CUTS, f"{lec}.mp4")
    if len(pieces) == 1:
        os.replace(pieces[0], dst)
    else:
        lst = os.path.join(work, f"{lec}.txt")
        with open(lst, "w") as fh:
            for p in pieces:
                fh.write(f"file '{p}'\n")
        subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-f", "concat",
                        "-safe", "0", "-i", lst, "-c", "copy", dst], check=True)
        for p in pieces:
            os.remove(p)
    d = subprocess.run(["ffprobe", "-v", "error", "-show_entries", "format=duration",
                        "-of", "csv=p=0", dst], capture_output=True, text=True).stdout.strip()
    print(f"  {lec}: {float(d)/60:.1f} min  ->  {dst}")
    return dst


def main() -> None:
    if len(sys.argv) < 2:
        sys.exit(__doc__)
    cap = sys.argv[1]
    want = [a for a in sys.argv[2:] if not a.startswith("-")]
    mans = load(cap)
    # Collect every capture's spans for each lecture, in capture order.
    merged: dict[str, list] = {}
    for man in mans:
        for lec, spans in ranges(man):
            merged.setdefault(lec, []).extend(
                (man, t0, t1, beats) for t0, t1, beats in spans)
    if not merged:
        sys.exit("no lecture markers in these captures")

    if "--list" in sys.argv:
        for lec in sorted(merged):
            spans = merged[lec]
            tot = sum(t1 - t0 for _, t0, t1, _ in spans)
            print(f"{lec}  {tot/60:5.1f} min  {len(spans)} span(s)")
            for _, _, _, beats in spans:
                for b in beats:
                    print("      ", b)
        return

    for lec in sorted(merged):
        if want and lec not in want:
            continue
        cut_one(mans, lec, merged[lec])


if __name__ == "__main__":
    main()
