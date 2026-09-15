#!/usr/bin/env python3
"""Audit every delivered lecture in one command.

    python3 audit.py                 # every w3-* in the delivery folder
    python3 audit.py --all           # the whole no-code course
    python3 audit.py w3-17 w3-31

Four things have shipped broken before, each of them silently, and each of them
is checked here:

  * **clipping** - `loudnorm` emits 192 kHz and the AAC encoder resampled it,
    overshooting the limiter by 3.6 dB. Peak is read as `Peak level dB`, not
    astats' `Max level`, which reports only the largest POSITIVE sample and let
    a file through at -0.1 dB reading -1.4.
  * **loudness drift** - the house target is -16 LUFS integrated. A lecture two
    LU off is audible against the one before it.
  * **a freeze at the tail** - narration outrunning the picture makes
    `assemble.py` hold on the last frame, which reads as a crash.
  * **the prompt reverting to `<user>@<machine>`** in the last seconds, which is
    a name and a machine name on screen. That cannot be detected, only looked
    at, so the tail frame is written out for every file and the report says so.

Nothing here modifies anything.
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
VIDS = os.path.abspath(os.path.join(HERE, "..", "..", "..",
                                    "videos", "nocode", "vids"))
TAILS = os.path.join(HERE, "frames", "delivered")

PEAK_FAIL = -0.5     # dBFS; anything above this is on the way to clipping
LUFS_TARGET = -16.0
LUFS_TOL = 1.5
FREEZE_TAIL = 2.5    # seconds of stillness at the end that count as a freeze


def ff(*args: str) -> str:
    return subprocess.run(["ffmpeg", "-hide_banner", "-nostats", *args],
                          capture_output=True, text=True).stderr


def duration(path: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", path], capture_output=True, text=True).stdout.strip())


def levels(path: str) -> tuple[float, float]:
    """Peak dBFS and integrated LUFS, in one pass each."""
    out = ff("-i", path, "-map", "a", "-af", "astats=measure_overall=Peak_level",
             "-f", "null", "-")
    m = re.findall(r"Peak level dB:\s*(-?\d+\.?\d*)", out)
    peak = max(float(x) for x in m) if m else float("nan")

    out = ff("-i", path, "-map", "a",
             "-af", "loudnorm=I=-16:TP=-1.5:LRA=11:print_format=json",
             "-f", "null", "-")
    try:
        j = json.loads(out[out.rindex("{"):out.rindex("}") + 1])
        lufs = float(j["input_i"])
    except Exception:
        lufs = float("nan")
    return peak, lufs


def stillness(path: str, total: float, look: float = 20.0) -> float:
    """Seconds of stillness the file ends on, 0 if it ends moving."""
    start = max(0.0, total - look)
    out = ff("-ss", f"{start:.2f}", "-i", path,
             "-vf", "freezedetect=n=-58dB:d=1.2", "-map", "0:v", "-f", "null", "-")
    starts = [float(x) for x in re.findall(r"freeze_start:\s*(\d+\.?\d*)", out)]
    ends = [float(x) for x in re.findall(r"freeze_end:\s*(\d+\.?\d*)", out)]
    if not starts:
        return 0.0
    if len(ends) >= len(starts):
        return 0.0          # every freeze ended before the file did
    return (total - start) - starts[-1]


def frozen_tail(path: str, total: float, cut: str | None = None) -> float:
    """Seconds the DELIVERED file holds a frame that the FOOTAGE does not.

    A raw `freezedetect` on the tail is not the test, and reporting it as one
    made twenty of thirty-five lectures look broken when four had a defect of
    under a second. Most of this week ends on a terminal that has finished
    printing: the last twenty seconds are genuinely one still image, and a
    lecture whose closing sentence sits over a settled screen is correct, not
    frozen. The defect being looked for is different - narration outrunning the
    footage, so `assemble.py` runs out of picture and holds the last frame - and
    it is only visible as stillness in the OUTPUT that is not in the SOURCE.
    """
    still = stillness(path, total)
    if still <= 0.0 or cut is None or not os.path.exists(cut):
        return still
    src_still = stillness(cut, duration(cut))
    # Whatever the source was already still for is the footage, not a hold.
    return max(0.0, still - src_still)


def tail_frame(path: str, name: str) -> str:
    os.makedirs(TAILS, exist_ok=True)
    dst = os.path.join(TAILS, f"{name}.jpg")
    subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-sseof", "-1.2",
                    "-i", path, "-frames:v", "1", "-vf", "scale=1280:-1",
                    "-q:v", "3", dst], capture_output=True)
    return dst


def main() -> None:
    want = [a for a in sys.argv[1:] if not a.startswith("-")]
    files = sorted(f for f in os.listdir(VIDS) if f.endswith(".mp4"))
    if want:
        files = [f for f in files if any(w in f for w in want)]
    elif "--all" not in sys.argv:
        files = [f for f in files if f.startswith("w3-")]

    bad = 0
    print(f"{'file':26s} {'min':>6s} {'peak':>7s} {'LUFS':>7s}  notes")
    for f in files:
        p = os.path.join(VIDS, f)
        total = duration(p)
        peak, lufs = levels(p)
        # The cut this lecture was assembled from, so stillness that is in the
        # footage is not reported as a hold.
        cut = os.path.join(HERE, "cuts", f[:-len("_adam.mp4")] + ".mp4") \
            if f.endswith("_adam.mp4") else None
        freeze = frozen_tail(p, total, cut)
        notes = []
        if peak > PEAK_FAIL:
            notes.append(f"PEAK {peak:+.1f} dBFS - re-render, never re-gain")
        if abs(lufs - LUFS_TARGET) > LUFS_TOL:
            notes.append(f"LOUDNESS {lufs:.1f} LUFS off house -16")
        if freeze > FREEZE_TAIL:
            notes.append(f"ENDS ON A FROZEN FRAME for {freeze:.1f}s")
        if notes:
            bad += 1
        tail_frame(p, f[:-4])
        print(f"{f:26s} {total/60:6.2f} {peak:7.1f} {lufs:7.1f}  "
              + ("; ".join(notes) if notes else "ok"))

    print(f"\n{len(files)} file(s), {bad} with a problem")
    print(f"tail frames in {os.path.relpath(TAILS, HERE)}/ - LOOK at them, "
          "the username in a reverted shell prompt is not detectable")


if __name__ == "__main__":
    main()
