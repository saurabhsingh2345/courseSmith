#!/usr/bin/env python3
"""Turn a piiscan log into redactions.json: merged blur boxes per lecture.

    /usr/bin/python3 redact_plan.py scan.log            # report
    /usr/bin/python3 redact_plan.py scan.log --write

The scan reads one frame a second and tesseract misses a good share of them -
`nocode01`'s leak is on screen for seven seconds and OCR found four of them -
so a box per hit would leave holes in the middle of a visible name. Two
widenings close them:

**Spatially, to the line and its neighbours.** A terminal scrolls between
samples, so a box that fits the word at t is off the word at t+1. The band is
the safe unit, and a blurred band inside a moving terminal reads as motion
rather than as a redaction. Widening also makes a scrolling line chain into one
box instead of a ladder of near-misses.

**In time, across gaps up to GAP.** Two hits at the same place five seconds
apart are one appearance with OCR failures in between, so the span between them
is blurred too. Verified against nocode01: hits at 289, 290, 291 and 295 come
out as 288.2-295.8, and the text is on screen 288.75-295.75.
"""
from __future__ import annotations

import json
import os
import re
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
VIDS = os.path.join(ROOT, "videos", "nocode", "vids")

PAD = 1.5     # seconds either side of a sample; 0.8 left two edge misses in nocode18
GAP = 5.0     # merge two boxes at one place if they are this close in time
XPAD = 12     # pixels either side of the word


def frame_size(stem: str) -> tuple[int, int]:
    out = subprocess.run(
        ["ffprobe", "-v", "error", "-select_streams", "v:0",
         "-show_entries", "stream=width,height", "-of", "csv=p=0",
         os.path.join(VIDS, f"{stem}_adam.mp4")],
        capture_output=True, text=True).stdout.strip().split(",")
    return int(out[0]), int(out[1])


def parse(path: str) -> dict[str, list[tuple]]:
    hits: dict[str, list[tuple]] = {}
    cur = None
    head = re.compile(r"^# (\S+?)_adam\.mp4\s+\d+ frames")
    hit = re.compile(r"^\s*([\d.]+)\s+(\S.*?)\s\s+(\d+) (\d+) (\d+) (\d+)\s*$")
    done = re.compile(r"^# (\S+?)_adam\.mp4: \d+ hits")
    finished = set()
    for ln in open(path):
        m = head.match(ln)
        if m:
            cur = m.group(1)
            hits.setdefault(cur, [])
            continue
        m = done.match(ln)
        if m:
            finished.add(m.group(1))
            continue
        m = hit.match(ln)
        if m and cur:
            t, word, x, y, w, h = m.groups()
            hits[cur].append((float(t), word, int(x), int(y), int(w), int(h)))
    # A file still being scanned has no closing line; leave it out entirely so
    # a partial box list is never mistaken for a finished one.
    return {k: v for k, v in hits.items() if k in finished}


def overlap(a: tuple, b: tuple) -> bool:
    """Do two (t0,t1,x0,y0,x1,y1) boxes touch in both time and space?"""
    if a[0] - GAP > b[1] or b[0] - GAP > a[1]:
        return False
    return not (a[4] < b[2] or b[4] < a[2] or a[5] < b[3] or b[5] < a[3])


def merge(boxes: list[tuple]) -> list[tuple]:
    boxes = sorted(boxes)
    changed = True
    while changed:
        changed = False
        out: list[tuple] = []
        for b in boxes:
            for i, o in enumerate(out):
                if overlap(o, b):
                    out[i] = (min(o[0], b[0]), max(o[1], b[1]),
                              min(o[2], b[2]), min(o[3], b[3]),
                              max(o[4], b[4]), max(o[5], b[5]))
                    changed = True
                    break
            else:
                out.append(b)
        boxes = out
    return sorted(boxes)


def plan(hits: dict[str, list[tuple]]) -> dict[str, list[list[float]]]:
    out: dict[str, list[list[float]]] = {}
    for stem, hs in sorted(hits.items()):
        if not hs:
            continue
        W, H = frame_size(stem)
        raw = []
        for t, _w, x, y, w, h in hs:
            raw.append((round(t - PAD, 2), round(t + PAD, 2),
                        max(0, x - XPAD), max(0, y - h),
                        min(W, x + w + XPAD), min(H, y + 2 * h)))
        boxes = []
        for t0, t1, x0, y0, x1, y1 in merge(raw):
            # crop needs even geometry on a yuv420p plane
            x0, y0 = int(x0) & ~1, int(y0) & ~1
            bw = min(W - x0, int(x1 - x0 + 1)) & ~1
            bh = min(H - y0, max(16, int(y1 - y0 + 1))) & ~1
            boxes.append([round(max(0.0, t0), 2), round(t1, 2), x0, y0, bw, bh])
        out[stem] = boxes
    return out


def main() -> int:
    log = sys.argv[1]
    hits = parse(log)
    p = plan(hits)
    tot_b = tot_s = 0
    for stem, boxes in p.items():
        W, H = frame_size(stem)
        secs = sum(b[1] - b[0] for b in boxes)
        tot_b += len(boxes)
        tot_s += secs
        print(f"\n{stem}  ({W}x{H})  {len(boxes)} box(es), {secs:.0f}s blurred")
        for t0, t1, x, y, w, h in boxes:
            area = 100.0 * w * h / (W * H)
            print(f"   {t0:8.1f} -> {t1:7.1f}  ({t1-t0:5.1f}s)  "
                  f"{w:4}x{h:<4} at {x:4},{y:<4}  {area:4.1f}% of frame"
                  + ("   <-- LARGE" if area > 12 else ""))
    print(f"\n{len(p)} lecture(s), {tot_b} boxes, {tot_s:.0f}s of blur total")
    if "--write" in sys.argv:
        dst = os.path.join(HERE, "redactions.json")
        with open(dst, "w") as fh:
            json.dump(p, fh, indent=1)
        print(f"-> {dst}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
