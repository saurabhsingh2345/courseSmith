#!/usr/bin/env python3
"""Build the wrap lecture's footage out of the week's own footage.

    python3 montage.py

w3-26 is specified as a montage of what we actually shot — no bullet slides, no
new recording. So its "cut" is assembled here from the delivered lecture cuts,
and from then on it goes through the ordinary pipeline: frames, narration, fit,
assemble.

Every clip is re-encoded to identical parameters before concatenation. A concat
demuxer stream-copy across sources that differ in even one encoder setting
produces a file that plays for the first clip and then stalls, which is the kind
of defect that only shows up after the render.
"""

from __future__ import annotations

import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
CUTS = os.path.join(HERE, "cuts")
WORK = os.path.join(HERE, "_assemble", "_montage")

# (source lecture, start second, seconds). Chosen mid-beat, never across a
# terminal transition — the shell prompt reverts to a username-bearing default
# for a second or two after Claude Code exits, and a montage is exactly where an
# unwatched second gets shipped.
SHOTS = [
    # Day 1 — the workshop
    ("w3-04", 150, 12),   # sub-agents: the roster
    ("w3-05", 120, 11),   # hooks: a rule that fires
    ("w3-06", 200, 11),   # plugins: packaged and installed
    # Day 2 — the boundary, and the agent with no screen
    ("w3-07", 210, 11),   # sandboxing: the permission surface
    ("w3-08", 240, 12),   # the container sandbox
    ("w3-12", 300, 11),   # a big codebase
    ("w3-13", 150, 11),   # driving Claude from outside
    ("w3-36",  71, 11),   # claude -p: no screen at all
    ("w3-37", 221, 11),   # agents chained in a pipeline
    ("w3-38",  73, 11),   # a stack of documents, not a codebase
    # Day 3 — the team
    ("w3-17", 500, 13),   # the team appears
    ("w3-18", 120, 12),   # the forensics
    ("w3-19", 300, 12),   # the tests, judged
    ("w3-20",  60, 13),   # Control Tower alive
    # Day 4 — the orchestrator, and a second product
    ("w3-23", 330, 12),   # our own orchestrator
    ("w3-24", 240, 11),   # the verdict, grounded in the repo
    ("w3-27", 122, 11),   # a second brief, nobody in the chair
    ("w3-28", 709, 11),   # the pipeline grows a dependency
    ("w3-29", 291, 13),   # the pool runs unattended
    ("w3-30",  47, 11),   # what a program built
    ("w3-31",  33, 11),   # Signal Desk alive
    ("w3-32", 367, 12),   # the fix pass reads the rule
    # Day 5 — the loop, the failure, the verdict
    ("w3-33", 268, 11),   # a green build on somebody else's computer
    ("w3-34", 198, 11),   # an issue handed to an agent
    ("w3-35", 105, 11),   # the pull request, merged
    ("w3-39", 165, 11),   # diagnosing from evidence
    ("w3-40", 391, 12),   # the exception, written down properly
    ("w3-41",  79, 12),   # two products, honestly compared
    # the close
    ("w3-25",  95, 11),   # the image builds
    ("w3-25", 130, 15),   # and it runs from the container
]


def dur(p: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def main() -> None:
    os.makedirs(WORK, exist_ok=True)
    parts = []
    for i, (lec, start, length) in enumerate(SHOTS):
        src = os.path.join(CUTS, f"{lec}.mp4")
        if not os.path.exists(src):
            sys.exit(f"missing {src}")
        have = dur(src)
        if start + length > have:
            start = max(0.0, have - length - 1)
            print(f"  {lec}: clamped start to {start:.0f}s (cut is {have:.0f}s)")
        out = os.path.join(WORK, f"{i:02d}_{lec}.mp4")
        subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-ss", str(start),
                        "-i", src, "-t", str(length), "-an",
                        "-c:v", "h264_videotoolbox", "-b:v", "26M", "-g", "15",
                        "-keyint_min", "15", "-pix_fmt", "yuv420p", out],
                       check=True)
        parts.append(out)
        print(f"  {lec} {start:>5.0f}s +{length}s")

    lst = os.path.join(WORK, "parts.txt")
    with open(lst, "w") as fh:
        for p in parts:
            fh.write(f"file '{p}'\n")
    final = os.path.join(CUTS, "w3-26.mp4")
    subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-f", "concat",
                    "-safe", "0", "-i", lst, "-c", "copy", final], check=True)
    print(f"\nw3-26: {dur(final)/60:.2f} min  ->  {final}")


if __name__ == "__main__":
    main()
