#!/usr/bin/env python3
"""A demo reel that CUTS inside lectures and cross-dissolves every join.

    /usr/bin/python3 reel.py --list
    /usr/bin/python3 reel.py 05

`build.py` joins whole lectures with hard cuts, which is right when every
lecture is worth showing end to end. This is for the opening of the course,
where one stretch of `nocode01` is an empty editor (see QA.md) and a demo should
not contain it.

Two rules the cuts follow:

1. **Cut on a narration boundary, never mid-explanation.** A cut point is the
   middle of a silence at least 0.45s long, so the join lands between sentences
   and the voice does not jump. `nocode01`'s removed stretch runs from one
   segment boundary of the original assembly to the next, so a whole beat comes
   out rather than half of two.
2. **Dissolve, do not hard-cut.** 0.6s of cross-fade on picture and voice
   together. A hard cut between two different screen recordings reads as a
   glitch; a dissolve reads as an edit.

The delivered lectures are never touched - this only reads them.
"""
from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
VIDS = os.path.join(ROOT, "videos", "nocode", "vids")
RENDERER = os.path.join(ROOT, "renderer")
OUT = os.path.expanduser("~/Desktop/showcase")
CARD = 6.0
XF = 0.6          # cross-dissolve, seconds
FADE_OUT = 1.5    # fade to black at the end

REELS = {
    "05": dict(
        slug="day-one-from-nothing-installed",
        kicker="showcase  ·  week 1  ·  day 1",
        lines=["Day one,", "from *nothing*"],
        stamp="install cursor  ·  build a 3d game  ·  what agentic coding actually is",
        pieces=[
            # nocode01 minus its empty-editor stretch. 379.3 and 502.4 are both
            # inside a sentence pause AND on a segment boundary of the original
            # assembly: 379.3 closes the "call it Instant" beat, 502.4 opens the
            # model picker. What comes out is the one beat that has nothing on
            # screen; what stays runs setup -> pick the model -> send the
            # request -> watch it write the file.
            ("nocode01", 0.0, 379.3),
            ("nocode01", 502.4, None),
            ("nocode02", 0.0, None),
            ("nocode03", 0.0, None),
            ("nocode04", 0.0, None),
        ]),
    "06": dict(
        slug="session-two-how-the-thing-actually-works",
        kicker="showcase  ·  week 1  ·  day 2",
        lines=["What is actually", "*happening* in there"],
        stamp="tokens, memory and reasoning  ·  tools and loops  ·  the context window",
        pieces=[
            ("nocode09", 0.0, None),
            ("nocode10", 0.0, None),
            # Stops inside lecture 11 rather than running it out, because 25
            # minutes is the brief. 318.28 is the middle of a 1.3s pause and the
            # frame under it is the card badged "NOT EXACTLY - AND THAT IS THE
            # NEXT LECTURE", so the reel ends on its own invitation to carry on.
            ("nocode11", 0.0, 318.28),
        ]),
}


def dur(p: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def card(key: str, r: dict) -> str:
    os.makedirs(OUT, exist_ok=True)
    props = os.path.join(OUT, f".reel_{key}.props.json")
    dst = os.path.join(OUT, f".reel_{key}.mp4")
    with open(props, "w") as fh:
        json.dump({"slides": [{"k": "head", "kicker": r["kicker"],
                               "lines": r["lines"], "size": 120,
                               "stamp": r["stamp"], "trans": "fade"}],
                   "secs": [CARD]}, fh)
    subprocess.run(["npx", "remotion", "render", "src/nocode/index.tsx",
                    "W3Insert", dst, f"--props={props}", "--log=error"],
                   cwd=RENDERER, check=True)
    return dst


def build(key: str) -> str:
    r = REELS[key]
    segs = []            # (path, ss, length)
    for name, a, b in r["pieces"]:
        p = os.path.join(VIDS, f"{name}_adam.mp4")
        if not os.path.exists(p):
            raise SystemExit(f"missing {p}")
        end = dur(p) if b is None else b
        segs.append((p, a, end - a))
        print(f"  {name:10} {a:7.1f} -> {end:7.1f}   {end - a:6.1f}s")

    c = card(key, r)
    ins = [(c, 0.0, CARD)] + segs
    total = sum(s[2] for s in ins) - XF * (len(ins) - 1)
    print(f"{key} {r['slug']}: {len(segs)} pieces, "
          f"{int(total // 60)}:{int(total % 60):02d}", flush=True)

    cmd = ["ffmpeg", "-nostdin", "-y", "-loglevel", "error"]
    for p, ss, ln in ins:
        # -ss AND -t both go BEFORE -i. After -i, `-t` is an OUTPUT option and
        # the last one silently caps the whole reel at one piece's length - the
        # first build of this came out 4:23 instead of 26:23 for exactly that.
        cmd += ["-ss", f"{ss:.3f}", "-t", f"{ln:.3f}", "-i", p]
    steps = []
    for i, _ in enumerate(ins):
        steps.append(f"[{i}:v]scale=1920:1080:flags=lanczos,fps=30,setsar=1,"
                     f"format=yuv420p,setpts=PTS-STARTPTS[v{i}]")
    # the card is silent; give it its own six seconds on the same clock
    steps.append(f"anullsrc=r=48000:cl=stereo,atrim=0:{CARD},asetpts=PTS-STARTPTS[a0]")
    for i in range(1, len(ins)):
        steps.append(f"[{i}:a]aresample=48000,aformat=channel_layouts=stereo,"
                     f"asetpts=PTS-STARTPTS[a{i}]")

    # xfade offsets are cumulative and each transition eats XF seconds, so the
    # offset is the running length of everything already joined minus XF.
    vprev, aprev, run = "v0", "a0", ins[0][2]
    for i in range(1, len(ins)):
        off = run - XF
        vo, ao = f"vx{i}", f"ax{i}"
        steps.append(f"[{vprev}][v{i}]xfade=transition=fade:duration={XF}:"
                     f"offset={off:.3f}[{vo}]")
        steps.append(f"[{aprev}][a{i}]acrossfade=d={XF}:c1=tri:c2=tri[{ao}]")
        vprev, aprev = vo, ao
        run = run + ins[i][2] - XF

    # Fade the last second and a half. A reel that stops dead reads as a file
    # that got truncated; a fade reads as an ending.
    steps.append(f"[{vprev}]fade=t=out:st={total - FADE_OUT:.3f}:d={FADE_OUT}[vend]")
    steps.append(f"[{aprev}]afade=t=out:st={total - FADE_OUT:.3f}:d={FADE_OUT}[aend]")

    dst = os.path.join(OUT, f"showcase-{key}-{r['slug']}.mp4")
    cmd += ["-filter_complex", ";".join(steps),
            "-map", "[vend]", "-map", "[aend]",
            "-c:v", "libx264", "-crf", "20", "-preset", "fast",
            "-pix_fmt", "yuv420p", "-c:a", "aac", "-b:a", "192k",
            "-movflags", "+faststart", dst]
    subprocess.run(cmd, check=True)
    for p in (c, c.replace(".mp4", ".props.json")):
        os.path.exists(p) and os.remove(p)
    got = dur(dst)
    print(f"  -> {dst}  {int(got // 60)}:{int(got % 60):02d}"
          f"  ({os.path.getsize(dst) / 1e6:.0f} MB)"
          + ("" if abs(got - total) < 1.5 else f"   RUNTIME OFF (wanted {total:.1f})"),
          flush=True)
    return dst


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("keys", nargs="*")
    ap.add_argument("--list", action="store_true")
    a = ap.parse_args()
    if a.list or not a.keys:
        for k, r in REELS.items():
            print(f"{k}  {r['slug']}  {len(r['pieces'])} pieces")
        return 0
    for k in a.keys:
        build(k)
    return 0


if __name__ == "__main__":
    sys.exit(main())
