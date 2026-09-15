#!/usr/bin/env python3
"""Where exactly is the picture dead? The animation map.

`holds.py` answers "how much of this lecture is frozen" - a number per lecture.
This answers "which seconds", because an animated insert has to land on a span
where nothing is happening and nowhere else. It uses a FINER detector
than holds.py on purpose - see the note above the constants.

    python3 deadzones.py --min 12 videos/nocode/vids/w3-*_adam.mp4

Prints one TSV row per dead span: lecture, start, end, duration. A span is
reported from the first frame of stillness to the last, in seconds of the
delivered master, so it can be handed straight to an ffmpeg overlay window.
"""
from __future__ import annotations

import argparse
import os
import subprocess
import sys

import numpy as np

FPS = 2

# NOT holds.py's detector, deliberately.
#
# holds.py asks "would a student say the picture is moving?" and answers it on a
# 64x36 luma with a mean-difference threshold. That is the right question for
# grading a lecture and its numbers stand.
#
# It is the wrong question here. This script decides which seconds may be COVERED
# by an animated card, and covering a frame that has text appearing on it would
# hide footage the narration is reading out loud. Checked against w3-36, whose
# 64x36 reading is one unbroken 148-second hold: three commands and two answers
# land inside that window. A few lines of terminal text in a 4K frame move the
# mean of a 64x36 luma by well under 1.4, so the coarse detector cannot see them.
#
# So: a much finer grid, and count the CELLS that moved rather than averaging the
# whole frame. A couple of words of new text lights up a handful of cells hard;
# encoder noise moves many cells barely. Counting cells over a per-cell floor
# separates those two, which averaging cannot.
W, H = 320, 180
CELL = 6          # per-cell luma delta that counts as ink, not noise
CELLS = 40        # this many changed cells = something appeared on screen


def spans(path: str, min_len: float):
    p = subprocess.run(
        ["ffmpeg", "-nostdin", "-v", "error", "-i", path,
         "-vf", f"fps={FPS},scale={W}:{H},format=gray", "-f", "rawvideo", "-"],
        capture_output=True)
    buf = np.frombuffer(p.stdout, dtype=np.uint8)
    n = len(buf) // (W * H)
    if n < 2:
        raise RuntimeError("no frames decoded")
    fr = buf[:n * W * H].reshape(n, H * W).astype(np.int16)
    moved = (np.abs(np.diff(fr, axis=0)) > CELL).sum(axis=1)
    same = moved < CELLS

    out, start = [], None
    for i, s in enumerate(same):
        if s and start is None:
            start = i
        elif not s and start is not None:
            out.append((start, i))
            start = None
    if start is not None:
        out.append((start, len(same)))

    keep = []
    for a, b in out:
        # frame i of `same` compares frame i to i+1, so the still run covers
        # frames a..b inclusive - one more frame than the diff run.
        t0, t1 = a / FPS, (b + 1) / FPS
        if t1 - t0 >= min_len:
            keep.append((t0, t1, t1 - t0))
    return keep, n / FPS


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--min", type=float, default=12.0,
                    help="shortest dead span worth reporting, seconds")
    ap.add_argument("files", nargs="+")
    args = ap.parse_args()

    print("lecture\tstart\tend\tdur\tdur_pct_of_lecture")
    for f in args.files:
        b = os.path.basename(f).replace("_adam.mp4", "").replace(".mp4", "")
        try:
            keep, total = spans(f, args.min)
        except Exception as e:                      # noqa: BLE001
            print(f"{b}\tERROR\t{e}", flush=True)
            continue
        for t0, t1, d in keep:
            print(f"{b}\t{t0:.1f}\t{t1:.1f}\t{d:.1f}\t{d / total * 100:.0f}%",
                  flush=True)
    return 0


if __name__ == "__main__":
    sys.exit(main())
