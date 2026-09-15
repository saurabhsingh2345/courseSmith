#!/usr/bin/env python3
"""Extend every redaction window until the blurred region actually changes.

    /usr/bin/python3 extend.py            # rewrites redactions.json, prints deltas

OCR finds a name in SOME of the frames it is on screen, so a window built from
hits ends while the text is still there - the verifier caught three of those.
Text does not leave a screen by fading; it leaves when the region scrolls,
clears or gets covered. So from each window's last hit, step forward in
STEP-second probes while the region still looks like it did (mean abs
difference under THRESH on the greyscale crop), and stop at the first frame
that differs. Same backwards from the first hit. Capped at CAP seconds so a
static screenshot cannot turn a two-second blur into a ten-minute one.
"""
from __future__ import annotations
import json, os, subprocess, sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
VIDS = os.path.join(ROOT, "videos", "nocode", "vids")
STEP, CAP, THRESH = 0.5, 45.0, 8.0
PAD = 1.5   # must match redact_plan.PAD


def crop(path, t, x, y, w, h):
    r = subprocess.run(["ffmpeg", "-nostdin", "-v", "error", "-ss", f"{t:.3f}", "-i", path,
                        "-frames:v", "1", "-vf", f"crop={w}:{h}:{x}:{y},format=gray",
                        "-f", "rawvideo", "-"], capture_output=True)
    return r.stdout if len(r.stdout) == w * h else None


def diff(a, b):
    return sum(abs(p - q) for p, q in zip(a, b)) / len(a)


def dur(p):
    return float(subprocess.run(["ffprobe", "-v", "error", "-show_entries", "format=duration",
                                 "-of", "csv=p=0", p], capture_output=True, text=True).stdout)


def extend(path, t0, t1, x, y, w, h, end):
    # The window is hit±PAD, so the last frame OCR actually SAW the text is at
    # t1-PAD, not t1 - by t1 the region may already show what replaced it, and
    # comparing against that walks straight through the change (nocode01 went
    # 7s -> 55s that way). Reference the hit, walk from the hit.
    last, first = t1 - PAD, t0 + PAD
    ref = crop(path, last, x, y, w, h)
    n1 = t1
    if ref:
        t = last
        while t + STEP <= min(end, last + CAP):
            f = crop(path, t + STEP, x, y, w, h)
            if f is None or diff(f, ref) > THRESH:
                break
            t += STEP
        n1 = max(t1, t + 0.4)
    ref = crop(path, first, x, y, w, h)
    n0 = t0
    if ref:
        t = first
        while t - STEP >= max(0.0, first - CAP):
            f = crop(path, t - STEP, x, y, w, h)
            if f is None or diff(f, ref) > THRESH:
                break
            t -= STEP
        n0 = min(t0, t - 0.4)
    return round(max(0.0, n0), 2), round(min(end, n1), 2)


def main():
    p = os.path.join(HERE, "redactions.json")
    red = json.load(open(p))
    changed = {}
    for stem, boxes in red.items():
        path = os.path.join(VIDS, f"{stem}_adam.mp4")
        end = dur(path)
        out = []
        for t0, t1, x, y, w, h in boxes:
            n0, n1 = extend(path, t0, t1, x, y, w, h, end)
            if abs(n0 - t0) > 0.3 or abs(n1 - t1) > 0.3:
                changed.setdefault(stem, []).append((t0, t1, n0, n1))
            out.append([n0, n1, x, y, w, h])
        red[stem] = out
    json.dump(red, open(p, "w"), indent=1)
    for stem, ch in changed.items():
        for t0, t1, n0, n1 in ch:
            print(f"{stem:10} {t0:7.1f}-{t1:6.1f}  ->  {n0:7.1f}-{n1:6.1f}   (+{(n1-n0)-(t1-t0):.1f}s)")
    print(f"{sum(len(v) for v in changed.values())} window(s) extended in {len(changed)} lecture(s)")

if __name__ == "__main__":
    main()
