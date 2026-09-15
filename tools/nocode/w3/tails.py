#!/usr/bin/env python3
"""Pull the last frame of every cut so the tails can be looked at in one go.

    python3 tails.py                 # every cut without a tail frame yet
    python3 tails.py w3-31 w3-32     # just these
    python3 tails.py --force

After Claude Code exits, the shell prompt reverts to `<user>@<machine>` for a
second or two before the driver resets it. That was in the tail of four
delivered lectures before anybody looked. The rule is that the last two seconds
of every cut get looked at before it is rendered, and this is the thing that
makes looking cheap.
"""

from __future__ import annotations

import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
CUTS = os.path.join(HERE, "cuts")
OUT = os.path.join(HERE, "frames")


def tail(lec: str, force: bool) -> str | None:
    src = os.path.join(CUTS, f"{lec}.mp4")
    dst = os.path.join(OUT, f"tail_{lec}.jpg")
    if os.path.exists(dst) and not force:
        return None
    r = subprocess.run(
        ["ffmpeg", "-y", "-loglevel", "error", "-sseof", "-1.5", "-i", src,
         "-frames:v", "1", "-vf", "scale=1280:-1", "-q:v", "3", dst],
        capture_output=True)
    return dst if r.returncode == 0 else None


def main() -> None:
    force = "--force" in sys.argv
    want = [a for a in sys.argv[1:] if a.startswith("w3-")]
    if not want:
        want = sorted({f[:-4] for f in os.listdir(CUTS)
                       if f.endswith(".mp4") and ".orig" not in f})
    made = []
    for lec in want:
        p = tail(lec, force)
        if p:
            made.append(os.path.basename(p))
    print(f"{len(made)} tail frame(s) written to frames/")
    for m in made:
        print("  ", m)
    print("LOOK at every one of them before rendering.")


if __name__ == "__main__":
    main()
