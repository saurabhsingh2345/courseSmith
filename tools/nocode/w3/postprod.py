#!/usr/bin/env python3
"""After a capture: cut every lecture, then sample frames so narration can be written.

    python3 postprod.py A1
    python3 postprod.py A1,A1b       # with a depth pass merged in

Prints a length report at the end, because the thing most worth knowing straight
after a shoot is which lectures have enough footage and which do not. A lecture
that came out far shorter than its slot needs a depth pass, not a padded script.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
CUTS = os.path.join(HERE, "cuts")

# Target minutes per lecture, from WEEK3-PLAN.md.
TARGET = {
    "w3-01": 11, "w3-02": 12, "w3-03": 12, "w3-04": 13, "w3-05": 11, "w3-06": 13,
    "w3-07": 10, "w3-08": 13, "w3-09": 12, "w3-10": 12, "w3-11": 14,
    "w3-12": 14, "w3-13": 17, "w3-14": 15,
    "w3-15": 11, "w3-16": 11, "w3-17": 12, "w3-18": 11, "w3-19": 11, "w3-20": 12,
    "w3-21": 12, "w3-22": 10, "w3-23": 11, "w3-24": 12, "w3-25": 10, "w3-26": 10,
}


def dur(p: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def main() -> None:
    if len(sys.argv) < 2:
        sys.exit(__doc__)
    cap = sys.argv[1]
    subprocess.run([sys.executable, os.path.join(HERE, "cut.py"), cap], check=True)

    lecs = []
    for name in sorted(os.listdir(CUTS)):
        if name.endswith(".mp4"):
            lecs.append(name[:-4])

    print("\n=== footage against target ===")
    for lec in lecs:
        d = dur(os.path.join(CUTS, f"{lec}.mp4")) / 60
        t = TARGET.get(lec, 0)
        # Footage is ramped, so it only has to be within reach of the target,
        # not equal to it. Below about 60% there is nothing to ramp and the
        # assembly falls back to freeze frames.
        verdict = "ok" if not t or d >= t * 0.6 else "THIN — needs a depth pass"
        print(f"  {lec}  {d:5.1f} min   target {t:>2} min   {verdict}")
        subprocess.run([sys.executable, os.path.join(HERE, "frames.py"), lec],
                       check=False)


if __name__ == "__main__":
    main()
