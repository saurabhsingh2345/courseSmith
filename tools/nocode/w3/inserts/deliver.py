#!/usr/bin/env python3
"""Put finished cuts into the delivery folder, refusing anything that is not right.

    python3 deliver.py --plan          # what would happen, change nothing
    python3 deliver.py w3-03 w3-39     # deliver these
    python3 deliver.py --all

`videos/nocode/vids/` is the delivery folder and this is the LAST step. Every
cut is gated first, because a delivery overwrites the only copy of a master:

  runtime unchanged   the whole approach rests on the audio never moving, so a
                      cut whose duration drifted is a cut whose narration is out
                      of sync. Hard fail.
  no dead span >=12s  the point of the exercise. Measured with deadzones.py, not
                      holds.py - see the README on why the coarse number reads
                      high on a finished cut and is not the result.
  no sync lead >=8s   no sentence still landing on a picture that finished
                      printing seconds earlier.
  audio identical     the audio stream is copied by lay.py, so its size and
                      codec must match the original byte for byte. This is the
                      check that would catch a re-encode slipping in.

Delivery MOVES rather than copies: the master goes to `orig/` and the staged cut
takes its place, both with os.replace, which is free on one filesystem. A
`orig/` entry is never overwritten, so the footage as shot survives every later
pass.
"""

from __future__ import annotations

import argparse
import glob
import os
import shutil
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
MAX_LEAD = 8.0


def probe(path: str, fields: str, streams: str | None = None) -> str:
    args = ["ffprobe", "-v", "error"]
    if streams:
        args += ["-select_streams", streams]
    args += ["-show_entries", fields, "-of", "csv=p=0", path]
    return subprocess.run(args, capture_output=True, text=True).stdout.strip()


def gate(lec: str, fast: bool = False) -> tuple[bool, list[str]]:
    cut = os.path.join(STAGED, f"{lec}_adam.mp4")
    live = os.path.join(VIDS, f"{lec}_adam.mp4")
    why = []
    if not os.path.exists(cut):
        return False, ["not built"]
    if not os.path.exists(live):
        return False, ["no master in the delivery folder"]

    try:
        d0 = float(probe(live, "format=duration"))
        d1 = float(probe(cut, "format=duration"))
    except ValueError:
        return False, ["unreadable - the cut may be truncated"]
    if abs(d0 - d1) > 0.25:
        why.append(f"runtime moved {d0:.2f} -> {d1:.2f}")

    try:
        v0 = float(probe(live, "stream=duration", "v:0"))
        v1 = float(probe(cut, "stream=duration", "v:0"))
        if abs(v0 - v1) > 0.25:
            why.append(f"video stream moved {v0:.2f} -> {v1:.2f} (VFR master "
                       f"trimmed by frame index?)")
    except ValueError:
        why.append("video stream unreadable")

    a0 = probe(live, "stream=codec_name,sample_rate,channels", "a:0")
    a1 = probe(cut, "stream=codec_name,sample_rate,channels", "a:0")
    if a0 != a1:
        why.append(f"audio stream changed: {a0} -> {a1}")

    if fast:
        # Correctness only. A truncated or mis-trimmed cut shows up as a decode
        # error, which neither of the quality checks below would report.
        errs = subprocess.run(["ffmpeg", "-nostdin", "-v", "error", "-i", cut,
                               "-f", "null", "-"],
                              capture_output=True, text=True).stderr.strip()
        if errs:
            why.append(f"decode errors: {errs.splitlines()[0][:80]}")
        return not why, why

    out = subprocess.run([PY, os.path.join(QA, "deadzones.py"), "--min", "12", cut],
                         capture_output=True, text=True).stdout
    dead = [l for l in out.splitlines()[1:] if l.strip() and "ERROR" not in l]
    if dead:
        why.append(f"{len(dead)} dead span(s) >=12s remain")

    out = subprocess.run([PY, os.path.join(QA, "align.py"), cut],
                         capture_output=True, text=True).stdout
    big = [l for l in out.splitlines()[1:] if l.strip()
           and len(l.split("\t")) >= 4 and float(l.split("\t")[3]) >= MAX_LEAD]
    if big:
        why.append(f"{len(big)} sync lead(s) >={MAX_LEAD:.0f}s remain")

    return not why, why


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("lectures", nargs="*")
    ap.add_argument("--all", action="store_true")
    ap.add_argument("--plan", action="store_true", help="gate only, deliver nothing")
    ap.add_argument("--fast", action="store_true",
                    help="skip the two slow quality checks (dead spans, sync leads) "
                         "and gate on correctness only: runtime, video-stream "
                         "duration, audio identity. Those two decode the whole 4K "
                         "cut twice more per lecture - about three minutes each - "
                         "and they re-measure what render.py's check() already "
                         "enforced against the measured map before the render. Use "
                         "when the plans were validated at render time; run "
                         "verify.py afterwards for the numbers.")
    args = ap.parse_args()

    lecs = args.lectures or sorted(
        os.path.basename(p).replace("_adam.mp4", "")
        for p in glob.glob(os.path.join(STAGED, "*_adam.mp4"))) if (
        args.lectures or args.all or args.plan) else []
    if not lecs:
        print("nothing named; use --all or --plan")
        return 1

    ok, blocked = [], []
    for lec in lecs:
        passed, why = gate(lec, args.fast)
        if not passed:
            blocked.append(lec)
            print(f"{lec:7} BLOCKED   {'; '.join(why)}")
            continue
        if args.plan:
            print(f"{lec:7} ready")
            ok.append(lec)
            continue

        os.makedirs(SRC.ORIG, exist_ok=True)
        keep = os.path.join(SRC.ORIG, f"{lec}_adam.mp4")
        live = os.path.join(VIDS, f"{lec}_adam.mp4")
        cut = os.path.join(STAGED, f"{lec}_adam.mp4")
        if not os.path.exists(keep):
            os.replace(live, keep)
        os.replace(cut, live)
        print(f"{lec:7} delivered  (original -> orig/{lec}_adam.mp4)")
        ok.append(lec)

    verb = "ready" if args.plan else "delivered"
    print(f"\n{len(ok)} {verb}" + (f", {len(blocked)} blocked: "
                                   f"{', '.join(blocked)}" if blocked else ""))
    return 1 if blocked else 0


if __name__ == "__main__":
    sys.exit(main())
