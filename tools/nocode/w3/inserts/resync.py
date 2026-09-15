#!/usr/bin/env python3
"""Push a reveal onto the word that describes it. Video only.

    python3 resync.py w3-03            # -> retimed/w3-03.mp4
    python3 resync.py w3-03 --plan     # print the time map, encode nothing

`../../qa/align.py` finds where the screen gave the answer away before the voice
got there. This fixes it the only way that is available once a lecture is
delivered: **the audio never moves**, so the picture is re-timed around it.

A fix is one line in the lecture's plan:

    "resync": [{"reveal": 54.8, "land": 72.0, "why": "..."}]

`reveal` is when the picture currently turns over and `land` is the anchor of the
sentence that describes it. Everything before is stretched so the pre-reveal
frame stays up until the words arrive, and everything after is compressed to
give the time back - so the file's duration and every later anchor are unchanged,
which is what keeps the narration in sync.

Stretching a stretch of terminal that was about to change anyway is invisible.
Compressing the stretch afterwards is nearly always invisible too, because these
lectures are 10-97% unchanged picture and the slack is sitting right there. What
is NOT free is compressing a passage where something is being read or typed, so
`--plan` prints every segment's speed and refuses anything outside SPEED_CAP:
a fix that needs a 3x sprint to pay for itself is a fix in the wrong place.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
PLANS = os.path.join(HERE, "plans")
OUT = os.path.join(HERE, "retimed")
ORIG = os.path.join(HERE, "orig")
VIDS = os.path.abspath(os.path.join(HERE, "..", "..", "..", "..",
                                    "videos", "nocode", "vids"))
SPEED_CAP = (0.18, 2.6)     # slower than this crawls; faster than this reads as a glitch
MIN_SEG = 0.5


def dur(path: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", path], capture_output=True, text=True,
        check=True).stdout.strip())


def timemap(fixes: list[dict], total: float) -> list[tuple[float, float, float, float]]:
    """Output/source control points -> a list of (out0, out1, src0, src1).

    Monotonic by construction: control points are sorted and each fix must move
    its reveal strictly later, so source time never runs backwards.
    """
    pts = [(0.0, 0.0)]
    for f in sorted(fixes, key=lambda f: float(f["land"])):
        out, src = float(f["land"]), float(f["reveal"])
        if src >= out:
            raise SystemExit(
                f"reveal {src} is not before land {out} - nothing to fix")
        if out <= pts[-1][0] or src <= pts[-1][1]:
            raise SystemExit(f"fix at {out} is out of order or overlaps the previous")
        pts.append((out, src))
    pts.append((total, total))

    segs = []
    for (o0, s0), (o1, s1) in zip(pts, pts[1:]):
        if o1 - o0 < MIN_SEG or s1 - s0 < MIN_SEG:
            raise SystemExit(f"segment {o0:.1f}-{o1:.1f} is too short to retime")
        segs.append((o0, o1, s0, s1))
    return segs


def build(segs, src: str, dst: str) -> None:
    steps, labels = [], []
    for i, (o0, o1, s0, s1) in enumerate(segs):
        speed = (s1 - s0) / (o1 - o0)          # source seconds per output second
        lab = f"v{i}"
        steps.append(
            f"[0:v]trim=start={s0:.4f}:end={s1:.4f},"
            f"setpts=(PTS-STARTPTS)/{speed:.6f}[{lab}]")
        labels.append(f"[{lab}]")
    steps.append("".join(labels) + f"concat=n={len(segs)}:v=1:a=0,fps=30[v]")
    subprocess.run(
        ["ffmpeg", "-nostdin", "-y", "-loglevel", "error", "-i", src,
         "-filter_complex", ";".join(steps),
         "-map", "[v]", "-map", "0:a",
         "-c:v", "h264_videotoolbox", "-b:v", "26M", "-pix_fmt", "yuv420p",
         "-c:a", "copy", "-movflags", "+faststart", dst], check=True)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("lecture")
    ap.add_argument("--plan", action="store_true", help="print the map, encode nothing")
    args = ap.parse_args()
    lec = args.lecture

    plan = json.load(open(os.path.join(PLANS, f"{lec}.json")))
    fixes = plan.get("resync") or []
    src = os.path.join(ORIG, f"{lec}_adam.mp4")
    if not os.path.exists(src):
        src = os.path.join(VIDS, f"{lec}_adam.mp4")
    total = dur(src)

    if not fixes:
        print(f"{lec}: no resync fixes in the plan")
        return 0

    segs = timemap(fixes, total)
    bad = []
    print(f"{lec}  {total:.2f}s  {len(fixes)} fix(es)")
    for o0, o1, s0, s1 in segs:
        speed = (s1 - s0) / (o1 - o0)
        flag = ""
        if not SPEED_CAP[0] <= speed <= SPEED_CAP[1]:
            flag = "  OUTSIDE THE CAP"
            bad.append((o0, o1, speed))
        kind = "hold " if speed < 0.9 else ("catch" if speed > 1.1 else "     ")
        print(f"  out {o0:7.1f}-{o1:7.1f}  <- src {s0:7.1f}-{s1:7.1f}  "
              f"{kind} x{speed:.2f}{flag}")
    if bad:
        print("\nrefusing: move the fix, or spread it over two, so the picture "
              "never has to sprint to catch up")
        return 1
    if args.plan:
        return 0

    os.makedirs(OUT, exist_ok=True)
    dst = os.path.join(OUT, f"{lec}.mp4")
    build(segs, src, dst)
    after = dur(dst)
    print(f"  {total:.2f}s -> {after:.2f}s")
    if abs(after - total) > 0.25:
        print("  RUNTIME MOVED - the narration is now out of sync, do not ship")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
