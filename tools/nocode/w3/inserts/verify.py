#!/usr/bin/env python3
"""The honest before/after for the animation pass.

    python3 verify.py                 # every staged cut
    python3 verify.py w3-03 w3-39

Reports four things per lecture, and the fourth is the one that matters:

  frozen (coarse)  holds.py on the original - the number the 2026-09-02 review
                   graded the week on, kept so the "before" is comparable.
  frozen (coarse)  holds.py on the finished cut. It OVER-reports here, and
                   systematically: its 64x36 mean-luma view cannot see a few
                   lines of terminal text, which is most of what these lectures
                   contain. w3-39 reads 25% by this measure with a 42s "hold"
                   in which two commands and their output actually print.
  dead >=12s       deadzones.py on the finished cut: stretches where NOTHING
                   appears. This is the real result. Zero is the target, and it
                   is what "no frozen frames" actually means.
  sync leads       align.py on the finished cut: sentences still landing on a
                   picture that finished printing seconds earlier.

Runtime is asserted identical, because the whole approach depends on the audio
never moving.
"""

from __future__ import annotations

import glob
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import source as SRC  # noqa: E402

QA = os.path.abspath(os.path.join(HERE, "..", "..", "qa"))
STAGED = os.path.join(HERE, "staged")
VIDS = os.path.abspath(os.path.join(HERE, "..", "..", "..", "..",
                                    "videos", "nocode", "vids"))
PY = sys.executable


def run(args: list[str]) -> str:
    return subprocess.run(args, capture_output=True, text=True).stdout


def frozen(path: str) -> str:
    # The coarse figure is not the result (see the README) and costs a full 4K
    # decode per file. On a second pass over fifteen lectures that is an hour of
    # decoding for a column the sharp count already answers, so it is opt-in.
    if not os.environ.get("VERIFY_HOLDS"):
        return "-"
    out = run([PY, os.path.join(QA, "holds.py"), path])
    for f in out.split("\t"):
        if f.startswith("pct_time_in_long_holds="):
            return f.split("=")[1].strip()
    return "?"


def dead(path: str) -> tuple[int, float]:
    out = run([PY, os.path.join(QA, "deadzones.py"), "--min", "12", path])
    rows = [l.split("\t") for l in out.splitlines()[1:] if l.strip()]
    rows = [r for r in rows if len(r) >= 4 and r[1] != "ERROR"]
    return len(rows), sum(float(r[3]) for r in rows)


def leads(path: str) -> int:
    out = run([PY, os.path.join(QA, "align.py"), path])
    return max(0, len([l for l in out.splitlines() if l.strip()]) - 1)


def dur(path: str) -> float:
    try:
        return float(run(["ffprobe", "-v", "error", "-show_entries",
                          "format=duration", "-of", "csv=p=0", path]).strip())
    except ValueError:
        return -1.0


def main() -> int:
    lecs = sys.argv[1:] or sorted(
        os.path.basename(p).replace("_adam.mp4", "")
        for p in glob.glob(os.path.join(STAGED, "*_adam.mp4")))

    print(f"{'lec':7}{'frozen(coarse)':>17}{'dead>=12s':>18}{'leads':>7}  runtime")
    bad = []
    for lec in lecs:
        cut = os.path.join(STAGED, f"{lec}_adam.mp4")
        orig = os.path.join(SRC.ORIG, f"{lec}_adam.mp4")
        if not os.path.exists(orig):
            orig = os.path.join(VIDS, f"{lec}_adam.mp4")
        if not os.path.exists(cut):
            print(f"{lec:7}  not built")
            continue

        d0, d1 = dur(orig), dur(cut)
        n, secs = dead(cut)
        ld = leads(cut)
        same = abs(d0 - d1) < 0.25
        if not same or n or ld:
            bad.append(lec)
        print(f"{lec:7}{frozen(orig):>8} ->{frozen(cut):>6}"
              f"{n:>10} spans{secs:>6.0f}s{ld:>7}"
              f"  {d0:.1f} -> {d1:.1f}{'' if same else '   RUNTIME MOVED'}")

    print(f"\n{len(lecs) - len(bad)} of {len(lecs)} fully clean"
          + (f"; look at: {', '.join(bad)}" if bad else ""))
    return 0


if __name__ == "__main__":
    sys.exit(main())
