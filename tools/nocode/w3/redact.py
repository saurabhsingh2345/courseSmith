#!/usr/bin/env python3
"""Blur a region of a lecture cut for a given time window.

    python3 redact.py w3-07 --box 165 225 620 1590 960 80 \
                            --box 335 405 620 1775 700 80

Boxes are t0 t1 x y w h, in seconds and in the cut's own 3840x2160 pixels.
The original cut is kept as <lecture>.orig.mp4 so a mistake is recoverable.

Why this exists: the agent runs shell commands with absolute paths, and those
echo his home directory — which carries his username and his company name — into
the frame. The house rule is that no name or company appears in any frame, and a
moving terminal cannot be cropped, so the region is blurred for exactly the
seconds it is visible.

boxblur radius must stay under 15 or it fails on the chroma plane; 12:6 twice is
stronger than 24 once and actually works.
"""

from __future__ import annotations

import os
import shutil
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
CUTS = os.path.join(HERE, "cuts")


def main() -> None:
    if len(sys.argv) < 2:
        sys.exit(__doc__)
    lec = sys.argv[1]
    # --onto-current blurs the cut as it stands instead of rebuilding it from
    # the untouched original. Rebuilding is right for a single hand-tuned pass
    # and wrong for a second one: it silently drops every box the first pass
    # applied, because those boxes are not in this invocation's argv. An
    # automated sweep runs more than once, so it needs the other behaviour.
    onto_current = "--onto-current" in sys.argv
    boxes = []
    a = [x for x in sys.argv[2:] if x != "--onto-current"]
    while a:
        if a[0] != "--box":
            sys.exit(f"unexpected {a[0]!r}")
        boxes.append(tuple(float(x) for x in a[1:7]))
        a = a[7:]
    if not boxes:
        sys.exit("no --box given")

    src = os.path.join(CUTS, f"{lec}.mp4")
    orig = os.path.join(CUTS, f"{lec}.orig.mp4")
    if not os.path.exists(orig):
        shutil.copy2(src, orig)
        print(f"kept original at {os.path.basename(orig)}")
    source = src if onto_current else orig

    parts, last = [], "0:v"
    for i, (t0, t1, x, y, w, h) in enumerate(boxes):
        x, y, w, h = int(x), int(y), int(w), int(h)
        # `enable` goes on the BLUR as well as the overlay. Without it ffmpeg
        # blurs every frame of the cut and throws all but the few seconds away:
        # nine boxes over a ten-minute 4K lecture took twenty minutes instead of
        # one. boxblur honours timeline editing, so gated it is a no-op outside
        # the window and the cost becomes proportional to the seconds redacted.
        win = f"between(t,{t0},{t1})"
        parts.append(
            f"[{last}]split[keep{i}][cut{i}];"
            f"[cut{i}]crop={w}:{h}:{x}:{y},"
            f"boxblur=12:6:enable='{win}',boxblur=12:6:enable='{win}'[bl{i}];"
            f"[keep{i}][bl{i}]overlay={x}:{y}:enable='{win}'[v{i}]"
        )
        last = f"v{i}"
    fc = ";".join(parts)

    dst = os.path.join(CUTS, f"{lec}.redacted.mp4")
    subprocess.run([
        "ffmpeg", "-y", "-loglevel", "error", "-i", source,
        "-filter_complex", fc, "-map", f"[{last}]", "-an",
        "-c:v", "h264_videotoolbox", "-b:v", "26M", "-g", "15",
        "-keyint_min", "15", "-pix_fmt", "yuv420p", dst], check=True)
    os.replace(dst, src)
    print(f"{lec}: {len(boxes)} region(s) blurred -> {src}")


if __name__ == "__main__":
    main()
