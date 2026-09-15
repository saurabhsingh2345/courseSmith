#!/usr/bin/env python3
"""Two-pass loudnorm to the house target: -16 LUFS, about -1.3 dBTP.

Usage: loudnorm.py <in.mp4> <out.mp4>

Why a script and not a shell one-liner. The house command appends the measured
values from pass one to the filter string, and doing that in the shell has bitten
this project before: a variable ending in `offset=N` swallows a character when
`:linear=true` is concatenated onto it, and ffmpeg then reports `offset` as an
invalid option. Building the string in python removes that whole class of bug.

`linear=true` on its own lands the true peak around -0.4 dBTP, which is hotter
than nocode09/10 and needed a limiter to match, so the limiter is not optional.
"""
import json, re, subprocess, sys


def measure(src):
    r = subprocess.run(
        # -vn matters: without it the measure pass decodes every video frame,
        # which on a 20-minute 1080p film is minutes of work for a number that
        # only depends on the audio.
        ["ffmpeg", "-hide_banner", "-vn", "-i", src, "-af",
         "loudnorm=I=-16:TP=-1.4:LRA=11:print_format=json", "-f", "null", "-"],
        capture_output=True, text=True)
    m = re.findall(r"\{[^{}]*input_i[^{}]*\}", r.stderr, re.S)
    if not m:
        sys.stderr.write(r.stderr[-2000:])
        raise SystemExit("loudnorm pass one produced no json")
    return json.loads(m[-1])


def main():
    src, dst = sys.argv[1], sys.argv[2]
    d = measure(src)
    print("pass one:", {k: d[k] for k in
                        ('input_i', 'input_tp', 'input_lra', 'input_thresh')})
    f = ("loudnorm=I=-16:TP=-1.4:LRA=11:linear=true"
         f":measured_I={d['input_i']}"
         f":measured_TP={d['input_tp']}"
         f":measured_LRA={d['input_lra']}"
         f":measured_thresh={d['input_thresh']}"
         f":offset={d['target_offset']}"
         # loudnorm outputs 192 kHz. Limiting at 192 kHz and letting the AAC
         # encoder resample afterwards puts peaks back ABOVE the ceiling the
         # limiter just enforced - that is the clipping this project chased
         # once already. Resample to 48 kHz first, then limit, so the limiter
         # is the last thing to touch the samples at their final rate.
         ",aresample=48000"
         ",alimiter=limit=0.82:level=disabled")
    cmd = ["ffmpeg", "-y", "-loglevel", "error", "-i", src, "-c:v", "copy",
           "-af", f, "-c:a", "aac", "-b:a", "192k", dst]
    subprocess.run(cmd, check=True)
    after = measure(dst)
    print("pass two:", {k: after[k] for k in ('input_i', 'input_tp', 'input_lra')})
    print("wrote", dst)


if __name__ == "__main__":
    main()
