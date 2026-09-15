#!/usr/bin/env python3
"""Lay narration over a lecture cut, ramping the footage to fit.

    python3 assemble.py w3-03

Reads narration/w3-NN.json:

    {"lecture": "w3-03",
     "segments": [
        {"anchor": 0,   "say": "..."},
        {"anchor": 47,  "say": "..."},
        {"anchor": 132, "say": "..."}]}

`anchor` is a time in the CUT — the moment the picture this sentence describes
appears. The narration is spoken at natural pace and the footage between two
anchors is then sped up or slowed to match how long the words actually took.

That ordering is the point. Ed's lectures work because the words and the picture
change together; ours only do if the footage bends to the narration rather than
the narration being padded to the footage. Waiting is ramped hard (an agent
thinking for four minutes becomes twenty seconds) and nothing is ever cut to
black.
"""

from __future__ import annotations

import json
import math
import os
import subprocess
import sys

import numpy as np

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import adam_lay as A  # noqa: E402

CUTS = os.path.join(HERE, "cuts")
NARR = os.path.join(HERE, "narration")
WORK = os.path.join(HERE, "_assemble")
OUT = os.path.abspath(os.path.join(HERE, "..", "..", "..",
                                   "videos", "nocode", "vids"))

SR = 48000          # loudnorm emits 192k; the AAC encoder must not resample
GAP = 0.42          # between sentences inside one segment
BRIDGE = 0.55       # silence between segments
MIN_SPEED = 0.92    # slower than this and the cursor crawls
MAX_SPEED = 16.0    # an agent thinking is ramped hard, never cut


def dur(p: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def speak_segment(text: str, tag: str, raw: str) -> np.ndarray:
    parts = []
    for i, sent in enumerate(A.sentences(text)):
        p = os.path.join(raw, f"{tag}_{i:02d}.mp3")
        A.speak(sent, p)
        parts.append(A.load(p))
        if i < len(A.sentences(text)) - 1:
            parts.append(np.zeros(int(A.SR * GAP), np.float32))
    return np.concatenate(parts) if parts else np.zeros(1, np.float32)


def ramp(src: str, t0: float, t1: float, speed: float, dst: str,
         want: float = 0.0) -> None:
    """One interval of footage at one speed, padded if the words outrun it.

    Slowing footage below MIN_SPEED to fill time makes the cursor crawl and looks
    broken. When a segment has more narration than footage, the clip is held on
    its last frame instead — a still frame under a sentence reads as emphasis;
    a 0.3x crawl reads as a fault. Never `-shortest`: that would silently drop
    the tail of the voice track.
    """
    speed = max(MIN_SPEED, min(MAX_SPEED, speed))
    vf = f"setpts=PTS/{speed:.4f}"
    have = (t1 - t0) / speed
    if want and want > have + 0.15:
        vf += f",tpad=stop_mode=clone:stop_duration={want - have:.3f}"
    subprocess.run([
        "ffmpeg", "-y", "-loglevel", "error", "-ss", f"{t0}", "-to", f"{t1}",
        "-i", src, "-an", "-vf", vf,
        "-c:v", "h264_videotoolbox", "-b:v", "26M",
        "-g", "15", "-keyint_min", "15", "-pix_fmt", "yuv420p", dst], check=True)


def main() -> None:
    if len(sys.argv) < 2:
        sys.exit(__doc__)
    lec = sys.argv[1]
    src = os.path.join(CUTS, f"{lec}.mp4")
    spec = json.load(open(os.path.join(NARR, f"{lec}.json")))
    segs = spec["segments"]
    total = dur(src)

    work = os.path.join(WORK, lec)
    raw = os.path.join(work, "raw")
    os.makedirs(raw, exist_ok=True)

    # 1. speak everything at natural pace and measure it
    audio, lens = [], []
    for i, sg in enumerate(segs):
        a = speak_segment(sg["say"], f"{lec}_{i:02d}", raw)
        audio.append(a)
        lens.append(len(a) / A.SR)
        print(f"  seg {i:02d}  {lens[-1]:6.1f}s  {sg['say'][:58]}…", flush=True)

    # 2. bend each footage interval to the words that sit over it
    parts = []
    for i, sg in enumerate(segs):
        f0 = float(sg["anchor"])
        f1 = float(segs[i + 1]["anchor"]) if i + 1 < len(segs) else total
        want = lens[i] + (BRIDGE if i + 1 < len(segs) else 0.0)
        have = max(f1 - f0, 0.05)
        speed = have / want
        dst = os.path.join(work, f"v{i:02d}.mp4")
        ramp(src, f0, f1, speed, dst, want=want)
        parts.append(dst)
        note = ("ramped" if speed > 1.6 else
                "HELD ON LAST FRAME — not enough footage" if speed < MIN_SPEED
                else "held" if speed < 1.05 else "")
        print(f"  cut {i:02d}  {have:6.1f}s of footage over {want:5.1f}s of words"
              f"  x{max(MIN_SPEED, min(MAX_SPEED, speed)):.2f} {note}", flush=True)

    # 3. one audio track on the same clock
    track = []
    for i, a in enumerate(audio):
        track.append(a)
        if i < len(audio) - 1:
            track.append(np.zeros(int(A.SR * BRIDGE), np.float32))
    a = np.concatenate(track)
    a = a * (0.92 / (float(np.max(np.abs(a))) or 1.0))
    wav = os.path.join(work, "voice.wav")
    A.write_wav(a, wav)

    lst = os.path.join(work, "parts.txt")
    with open(lst, "w") as fh:
        for p in parts:
            fh.write(f"file '{p}'\n")
    silent = os.path.join(work, "video.mp4")
    subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-f", "concat",
                    "-safe", "0", "-i", lst, "-c", "copy", silent], check=True)

    # 4. house level: a REAL two-pass loudnorm to -16 LUFS, then a limiter.
    #
    #    Three things bite here. Single-pass loudnorm only approximates the target.
    #    `alimiter` defaults to level=true, which auto-levels the output back
    #    up to full scale — measured at exactly 0.0 dBFS, i.e. clipping, which is
    #    the opposite of what a limiter is for. It must be level=disabled.
    #    And loudnorm emits 192 kHz whatever went in. Nothing downstream asks for
    #    48 kHz, so the AAC encoder resamples it itself, and its internal
    #    resampler rings hard on a limited signal: w3-17 left a limiter set to
    #    -5.6 dBFS and came back out of the encoder at -2.0, a 3.6 dB overshoot,
    #    which is what pushed the earlier renders past full scale. Resampling
    #    inside the chain, BEFORE the limiter, costs 0.1 dB instead. So the
    #    limiter sits last and is the real ceiling; the loop below is a net.
    meas = subprocess.run(
        ["ffmpeg", "-hide_banner", "-nostats", "-i", wav,
         "-af", "loudnorm=I=-16:TP=-1.5:LRA=11:print_format=json",
         "-f", "null", "-"], capture_output=True, text=True).stderr
    m = json.loads(meas[meas.rindex("{"):meas.rindex("}") + 1])
    def chain(ceiling):
        return ("loudnorm=I=-16:TP=-1.5:LRA=11"
                f":measured_I={m['input_i']}:measured_TP={m['input_tp']}"
                f":measured_LRA={m['input_lra']}:measured_thresh={m['input_thresh']}"
                f":offset={m['target_offset']}:linear=true,"
                f"aresample={SR},"
                f"alimiter=limit={ceiling:.4f}:level=disabled")

    def peak_of(path, stream_arg=()):
        """Absolute sample peak of the encoded audio, as a linear 0..1 value.

        Read `Peak level dB`, NOT `Max level`. `Max level` is the largest
        POSITIVE sample, so a waveform whose biggest excursion happens to be
        negative walks straight past it: w3-20 measured `Max level 0.853`
        (-1.4 dB) while the file was actually at -0.1 dB, and volumedetect —
        which takes an absolute value — disagreed by 1.25 dB. The guard was
        half-blind, and it is the guard's whole job not to be.
        """
        st = subprocess.run(
            ["ffmpeg", "-hide_banner", "-nostats", "-i", path, *stream_arg,
             "-af", "astats=metadata=1", "-f", "null", "-"],
            capture_output=True, text=True).stderr
        db = [float(l.split(":")[-1]) for l in st.splitlines()
              if "Peak level dB:" in l]
        if not db:
            return 1.0                     # unmeasurable: treat as a breach
        return 10 ** (max(db) / 20.0)

    os.makedirs(OUT, exist_ok=True)
    final = os.path.join(OUT, f"{lec}_adam.mp4")

    # Master through a LOSSLESS intermediate, then verify the AAC. If the encoder
    # overshoots the house ceiling we lower the LIMITER, never pre-gain the input:
    # a `volume=` in front of a loudnorm that still carries the original
    # `measured_*` makes it over-correct, and w3-17 came out at +2.4 dBFS that way
    # — hotter than with no correction at all.
    CEIL = 0.891                      # -1.0 dBFS
    ceiling = 0.841                   # -1.5 dBFS, matching the house TP
    mastered = os.path.join(work, "mastered.wav")
    for _ in range(3):
        subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-i", wav,
                        "-af", chain(ceiling), "-c:a", "pcm_s16le",
                        mastered], check=True)
        subprocess.run([
            "ffmpeg", "-y", "-loglevel", "error", "-i", silent, "-i", mastered,
            "-map", "0:v:0", "-map", "1:a:0",
            "-c:v", "copy", "-c:a", "aac", "-b:a", "192k", final], check=True)
        peak = peak_of(final, ("-map", "a"))
        if peak <= CEIL:
            break
        ceiling *= (CEIL / peak) * 0.97
        print(f"    encoded peak {20*math.log10(peak):+.2f} dBFS — "
              f"limiter down to {20*math.log10(ceiling):.2f} dBFS", flush=True)
    else:
        print(f"    !! still {20*math.log10(peak):+.2f} dBFS after 3 tries",
              flush=True)

    lvl = subprocess.run(
        ["ffmpeg", "-hide_banner", "-nostats", "-i", final,
         "-af", "ebur128=peak=true", "-f", "null", "-"],
        capture_output=True, text=True).stderr.strip().splitlines()
    for line in lvl[-14:]:
        if line.strip().startswith(("I:", "Peak:", "LRA:")):
            print("   ", line.strip(), flush=True)

    print(f"\n{lec}: {dur(final)/60:.2f} min  ->  {final}", flush=True)


if __name__ == "__main__":
    main()
