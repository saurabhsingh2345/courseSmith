#!/usr/bin/env python3
"""Does the footage reveal things before the narration describes them?

    python3 align.py videos/nocode/vids/w3-36_adam.mp4

`assemble.py`'s contract is explicit: a narration segment's `anchor` is "a time
in the CUT - the moment the picture this sentence describes appears". So the
contract is testable. Find the big visual events, and for each anchor ask where
its nearest big event actually landed. An event BEFORE its anchor means the
screen gave the answer away while the voice was still setting the question up,
which is the thing that makes these lectures feel out of step.

Reported only when an anchor has NO event of its own: the words land on a
picture that finished printing seconds ago. An anchor whose change arrives with
it is the contract holding, and an event after its anchor is the house style -
the narration saying "watch what it does" before it does it. Neither is a defect.

A long segment describes several things, so events in the middle of one are
expected and are not judged here; only the moment a segment BEGINS is.

Event strength is the count of changed cells on a 320x180 luma, the same
measure deadzones.py uses; a big event is a terminal printing a screenful or a
pane appearing, not a cursor blinking.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys

import numpy as np

HERE = os.path.dirname(os.path.abspath(__file__))
NARR = os.path.abspath(os.path.join(HERE, "..", "w3", "narration"))
FPS = 4                 # finer than deadzones.py: we want the moment, not the span
W, H = 320, 180
CELL = 6
BIG = 1200              # changed cells that count as "the picture turned over"
LOOK_BACK = 25.0        # how far before an anchor to hunt for its picture
AT = (3.0, 6.0)         # an event this close to an anchor counts as landing ON it
LEAD_MIN = 2.5          # under this and it is timing jitter, not a lead


def events(path: str) -> tuple[np.ndarray, np.ndarray, float]:
    p = subprocess.run(
        ["ffmpeg", "-nostdin", "-v", "error", "-i", path,
         "-vf", f"fps={FPS},scale={W}:{H},format=gray", "-f", "rawvideo", "-"],
        capture_output=True)
    buf = np.frombuffer(p.stdout, dtype=np.uint8)
    n = len(buf) // (W * H)
    fr = buf[:n * W * H].reshape(n, H * W).astype(np.int16)
    moved = (np.abs(np.diff(fr, axis=0)) > CELL).sum(axis=1)
    t = (np.arange(len(moved)) + 1) / FPS
    return t, moved, n / FPS


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("files", nargs="+")
    ap.add_argument("--big", type=int, default=BIG)
    args = ap.parse_args()

    print("lecture\tanchor\tevent_at\tlead\tstrength\tsay")
    for f in args.files:
        lec = os.path.basename(f).replace("_adam.mp4", "").replace(".mp4", "")
        jf = os.path.join(NARR, f"{lec}.json")
        if not os.path.exists(jf):
            print(f"{lec}\tNO NARRATION", flush=True)
            continue
        segs = json.load(open(jf))["segments"]
        try:
            t, moved, _ = events(f)
        except Exception as e:                       # noqa: BLE001
            print(f"{lec}\tERROR\t{e}", flush=True)
            continue

        for s in segs:
            a = float(s["anchor"])
            # First question: does this anchor have a picture of its own? An
            # event inside the AT window means the words and the change arrive
            # together, which is the contract holding.
            on = (t >= a - AT[0]) & (t <= a + AT[1]) & (moved >= args.big)
            if on.any():
                continue
            # It does not. So the sentence is spoken over a picture that is
            # already finished - find when it finished.
            before = np.where((t >= a - LOOK_BACK) & (t < a - AT[0])
                              & (moved >= args.big))[0]
            if not len(before):
                continue
            at = float(t[before[-1]])       # the LAST one before the words
            lead = a - at
            if lead >= LEAD_MIN:
                say = " ".join(s["say"].split())[:70]
                print(f"{lec}\t{a:.1f}\t{at:.1f}\t{lead:.1f}\t"
                      f"{int(moved[before[-1]])}\t{say}", flush=True)
    return 0


if __name__ == "__main__":
    sys.exit(main())
