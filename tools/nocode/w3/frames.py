#!/usr/bin/env python3
"""Sample a lecture cut so the narration can be written to what is actually there.

    python3 frames.py w3-03              # a frame every 20s, plus one per beat
    python3 frames.py w3-03 --every 10

The house rule for this program is that narration describes the frame it sits
over. That is unenforceable if the narration is written from a plan instead of
from the footage, which is exactly how a lecture ended up describing a failure
that never happened. So: look at the frames first, then write.

Beat markers from the capture manifest are used as sample points too, because a
beat boundary is where the picture changes and therefore where a sentence has to
change with it.
"""

from __future__ import annotations

import glob
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
CUTS = os.path.join(HERE, "cuts")
CAPS = os.path.join(HERE, "captures")
OUT = os.path.join(HERE, "frames")


PAD_IN = 0.6      # must match cut.py
PAD_OUT = 1.2


def beats_for(lec: str) -> list[tuple[float, str]]:
    """Offsets of this lecture's beats, in seconds from the start of its cut.

    Walks every capture that contains the lecture, in capture order, and carries
    an elapsed offset across them — a lecture shot in two passes has its second
    pass appended to the first, so its later beats do not start from zero. Getting
    this wrong would put every sampled frame in the wrong place for exactly the
    lectures that needed a depth pass most.
    """
    out: list[tuple[float, str]] = []
    elapsed = 0.0
    for man_path in sorted(glob.glob(os.path.join(CAPS, "capture-*.json"))):
        man = json.load(open(man_path))
        ms = sorted((m for m in man["markers"] if m["label"].startswith(lec + "/")),
                    key=lambda m: m["t"])
        if not ms:
            continue
        t0 = ms[0]["t"]
        for m in ms:
            out.append((round(elapsed + (m["t"] - t0) + PAD_IN, 1),
                        # A beat label may itself contain a slash
                        # (`w3-29/build/t01`); it becomes a filename below, so
                        # it cannot keep one — ffmpeg was handed a path into a
                        # directory that does not exist and died on the first
                        # such beat, after the jpgs had already been cleared.
                        m["label"].split("/", 1)[1].replace("/", "-")))
        # This capture contributes from its first beat to the next lecture's
        # first marker, or to the end of its footage.
        later = [x["t"] for x in man["markers"] if x["t"] > ms[-1]["t"]]
        end = min(later) if later else max(s["t1"] for s in man["segments"])
        elapsed += (end - t0) + PAD_IN + PAD_OUT
    return out


def main() -> None:
    if len(sys.argv) < 2:
        sys.exit(__doc__)
    lec = sys.argv[1]
    every = 20.0
    if "--every" in sys.argv:
        every = float(sys.argv[sys.argv.index("--every") + 1])

    src = os.path.join(CUTS, f"{lec}.mp4")
    if not os.path.exists(src):
        sys.exit(f"no cut yet: {src}")
    dur = float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", src], capture_output=True, text=True).stdout.strip())

    dst = os.path.join(OUT, lec)
    os.makedirs(dst, exist_ok=True)
    for f in glob.glob(os.path.join(dst, "*.jpg")):
        os.remove(f)

    points: dict[float, str] = {}
    for t, name in beats_for(lec):
        if t < dur:
            points[round(t + 2.0, 1)] = name      # 2s in, past the transition
    t = 5.0
    while t < dur:
        points.setdefault(round(t, 1), "")
        t += every

    for t in sorted(points):
        tag = f"{int(t//60):02d}m{int(t%60):02d}"
        name = points[t]
        out = os.path.join(dst, f"{tag}{'_' + name if name else ''}.jpg")
        subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-ss", f"{t}",
                        "-i", src, "-frames:v", "1", "-vf", "scale=1280:-1",
                        "-q:v", "4", out], check=True)
    print(f"{lec}: {dur/60:.1f} min, {len(points)} frames -> {dst}")
    for t in sorted(points):
        if points[t]:
            print(f"   {int(t//60):02d}:{int(t%60):02d}  {points[t]}")


if __name__ == "__main__":
    main()
