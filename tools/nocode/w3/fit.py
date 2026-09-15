#!/usr/bin/env python3
"""Check narration against footage WITHOUT rendering.

    python3 fit.py w3-02

assemble.py is the source of truth, but it costs a TTS round trip and a re-encode
per segment, and it cannot be run while a capture is recording. This estimates
from word count alone (Adam at the house rate measures 198 words per minute, pauses included)
so narration can be balanced first and rendered once.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
# Measured, not guessed: 537 words rendered to 162.8s of finished lecture,
# inter-sentence pauses included. A 150 wpm guess under-read the available room
# by about a third and produced lectures shorter than the footage could carry.
WPM = 198.0
MIN_SPEED = 0.92


def main() -> None:
    lec = sys.argv[1]
    d = json.load(open(os.path.join(HERE, "narration", f"{lec}.json")))
    # Prefer the real cut, fall back to the length the narration file records.
    # The fallback is what lets a lecture be written while the next capture is
    # still rolling - `beats.py` knows the length from the manifest, and cutting
    # cannot happen until the camera stops.
    cut = os.path.join(HERE, "cuts", f"{lec}.mp4")
    if os.path.exists(cut):
        total = float(subprocess.run(
            ["ffprobe", "-v", "error", "-show_entries", "format=duration",
             "-of", "csv=p=0", cut],
            capture_output=True, text=True).stdout.strip())
    else:
        total = float(d["cut_seconds"])
        print("  (no cut yet - using cut_seconds from the narration file)")

    segs = d["segments"]
    say_total = 0.0
    print(f"{lec}: {total/60:.1f} min of footage")
    for i, s in enumerate(segs):
        f0 = float(s["anchor"])
        f1 = float(segs[i + 1]["anchor"]) if i + 1 < len(segs) else total
        have = f1 - f0
        want = len(s["say"].split()) / WPM * 60 + 0.55
        say_total += want
        room = have / MIN_SPEED
        flag = "  FREEZE — trim %.0f words" % ((want - room) * WPM / 60) if want > room else \
               ("  room for %.0f more words" % ((room - want) * WPM / 60) if room - want > 4 else "")
        print(f"  seg {i}  footage {have:6.1f}s   words {want:6.1f}s{flag}")
    print(f"  lecture would run about {say_total/60:.2f} min "
          f"(ceiling {total/MIN_SPEED/60:.2f} min)")


if __name__ == "__main__":
    main()
