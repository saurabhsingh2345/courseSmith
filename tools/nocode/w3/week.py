#!/usr/bin/env python3
"""One table for the whole week: footage, writing, voice, delivery.

    python3 week.py
    python3 week.py --live      # also ask what the voice allowance is

Every other tool here answers one question about one lecture. This answers the
only question that matters at the end of a day: what is finished, what is
written but unvoiced, what is shot but unwritten, and how much runtime is still
sitting in footage nobody has narrated yet.

Columns:
  cut       minutes of footage the lecture was cut from
  written   what the narration would run at the house rate
  ceiling   what that footage could carry at natural pace (cut / 0.92)
  room      ceiling minus written - runtime still available for writing alone
  voice     paid = every sentence is in the content store, so a render is free
  out       the delivered file exists in videos/nocode/vids/
"""

from __future__ import annotations

import glob
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import adam_lay as A  # noqa: E402
import quota  # noqa: E402

CUTS = os.path.join(HERE, "cuts")
NARR = os.path.join(HERE, "narration")
VIDS = os.path.abspath(os.path.join(HERE, "..", "..", "..",
                                    "videos", "nocode", "vids"))
WPM = 198.0
MIN_SPEED = 0.92


def dur(path: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", path], capture_output=True, text=True).stdout.strip())


def main() -> None:
    lecs = sorted({os.path.basename(f)[:-4] for f in glob.glob(os.path.join(CUTS, "*.mp4"))
                   if ".orig" not in f and ".redacted" not in f}
                  | {os.path.basename(f)[:-5] for f in glob.glob(os.path.join(NARR, "w3-*.json"))})
    paid = quota.spoken()

    tot_cut = tot_written = tot_ceiling = tot_out = 0.0
    unpaid = 0
    print(f"{'lecture':9s} {'cut':>7s} {'written':>8s} {'ceiling':>8s} "
          f"{'room':>6s}  {'voice':7s} {'out':>6s}")
    for lec in lecs:
        cut = os.path.join(CUTS, f"{lec}.mp4")
        nar = os.path.join(NARR, f"{lec}.json")
        c = dur(cut) / 60 if os.path.exists(cut) else 0.0
        if os.path.exists(nar):
            d = json.load(open(nar))
            words = sum(len(s["say"].split()) for s in d["segments"])
            w = words / WPM
            new = sum(len(s) for sg in d["segments"]
                      for s in A.sentences(sg["say"]) if s not in paid)
        else:
            w, new = 0.0, -1
        ceil = c / MIN_SPEED
        out = os.path.join(VIDS, f"{lec}_adam.mp4")
        o = dur(out) / 60 if os.path.exists(out) else 0.0

        tot_cut += c
        tot_written += w
        tot_ceiling += ceil
        tot_out += o
        if new > 0:
            unpaid += new

        voice = "-" if new < 0 else ("paid" if new == 0 else f"{new}")
        print(f"{lec:9s} {c:7.1f} {w:8.1f} {ceil:8.1f} {ceil - w:6.1f}  "
              f"{voice:7s} {o:6.1f}")

    print(f"\n{'TOTAL':9s} {tot_cut:7.1f} {tot_written:8.1f} {tot_ceiling:8.1f} "
          f"{tot_ceiling - tot_written:6.1f}  {unpaid:<7d} {tot_out:6.1f}")
    print(f"\nfootage:   {tot_cut/60:.2f} h cut")
    print(f"written:   {tot_written/60:.2f} h of narration")
    print(f"ceiling:   {tot_ceiling/60:.2f} h - what the footage could carry")
    print(f"delivered: {tot_out/60:.2f} h in videos/nocode/vids/")
    print(f"\nunbought voice for what is written: {unpaid} characters")

    if "--live" in sys.argv:
        import urllib.request
        req = urllib.request.Request(
            "https://api.elevenlabs.io/v1/user/subscription",
            headers={"xi-api-key": A.KEY})
        sub = json.load(urllib.request.urlopen(req, timeout=30))
        left = sub["character_limit"] - sub["character_count"]
        print(f"allowance left:                     {left} characters")


if __name__ == "__main__":
    main()
