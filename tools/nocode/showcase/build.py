#!/usr/bin/env python3
"""Cut a continuous 25-35 minute showcase out of consecutive finished lectures.

    python3 build.py --list
    python3 build.py 01            # one showcase, into ~/Desktop/showcase/
    python3 build.py --all

Each showcase is a run of consecutive lectures from the delivery folder, joined
end to end behind one six-second title card rendered in the house style
(`W3Insert` in renderer/src/nocode, same kit as every lecture). The lectures are
untouched: Week 1/2 are 1080p stereo, Week 3 is 4K mono, so every input is
normalised to 1920x1080 / 30fps / 48 kHz stereo and the join is one re-encode.
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
CARD_SECS = 6

SHOWCASES = {
    "01": dict(
        slug="claude-code-from-zero",
        kicker="showcase  ·  week 2  ·  day 1",
        lines=["Claude Code,", "from *zero*"],
        stamp="welcome to pro week  ·  the rise of claude code  ·  init, context and testing",
        lectures=["nocode35", "nocode36", "nocode37"]),
    "02": dict(
        slug="jira-ticket-to-pull-request",
        kicker="showcase  ·  week 2  ·  day 4",
        lines=["From a *ticket*", "to a pull request"],
        stamp="jira mcp  ·  github mcp  ·  an agent that ships the issue on its own",
        lectures=["nocode54", "nocode55", "nocode56"]),
    "03": dict(
        slug="expert-week-sub-agents-hooks-plugins",
        kicker="showcase  ·  week 3  ·  day 1",
        lines=["The *expert* week", "begins"],
        stamp="slash commands  ·  sub-agents  ·  hooks  ·  plugins and marketplaces",
        lectures=["w3-01", "w3-02", "w3-03", "w3-04", "w3-05", "w3-06"]),
    "04": dict(
        slug="agents-md-and-the-ralph-loop",
        kicker="showcase  ·  week 1  ·  day 2",
        lines=["Context is the", "*whole* game"],
        stamp="mastering agents.md  ·  from yolo mode to ralph loops",
        lectures=["nocode12", "nocode13"]),
}


def dur(p: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def card(key: str, sc: dict) -> str:
    """The six-second title card, 1920x1080, silent."""
    os.makedirs(OUT, exist_ok=True)
    props = os.path.join(OUT, f".card_{key}.props.json")
    dst = os.path.join(OUT, f".card_{key}.mp4")
    with open(props, "w") as fh:
        json.dump({"slides": [{"k": "head", "kicker": sc["kicker"],
                               "lines": sc["lines"], "size": 120,
                               "stamp": sc["stamp"], "trans": "fade"}],
                   "secs": [CARD_SECS]}, fh)
    subprocess.run(["npx", "remotion", "render", "src/nocode/index.tsx",
                    "W3Insert", dst, f"--props={props}", "--log=error"],
                   cwd=RENDERER, check=True)
    return dst


def build(key: str) -> str:
    sc = SHOWCASES[key]
    files = [os.path.join(VIDS, f"{l}_adam.mp4") for l in sc["lectures"]]
    missing = [f for f in files if not os.path.exists(f)]
    if missing:
        raise SystemExit("missing:\n  " + "\n  ".join(missing))
    total = sum(dur(f) for f in files) + CARD_SECS
    print(f"{key} {sc['slug']}: {len(files)} lectures, "
          f"{int(total // 60)}:{int(total % 60):02d}", flush=True)

    c = card(key, sc)
    inputs = [c] + files
    steps = []
    for i, _ in enumerate(inputs):
        steps.append(f"[{i}:v]scale=1920:1080:flags=lanczos,fps=30,setsar=1,"
                     f"format=yuv420p,setpts=PTS-STARTPTS[v{i}]")
    # the card is silent: give it six seconds of silence on the same clock
    steps.append(f"anullsrc=r=48000:cl=stereo,atrim=0:{CARD_SECS},asetpts=PTS-STARTPTS[a0]")
    for i in range(1, len(inputs)):
        steps.append(f"[{i}:a]aresample=48000,aformat=channel_layouts=stereo,"
                     f"asetpts=PTS-STARTPTS[a{i}]")
    steps.append("".join(f"[v{i}][a{i}]" for i in range(len(inputs)))
                 + f"concat=n={len(inputs)}:v=1:a=1[v][a]")

    dst = os.path.join(OUT, f"showcase-{key}-{sc['slug']}.mp4")
    cmd = ["ffmpeg", "-nostdin", "-y", "-loglevel", "error"]
    for p in inputs:
        cmd += ["-i", p]
    cmd += ["-filter_complex", ";".join(steps), "-map", "[v]", "-map", "[a]",
            "-c:v", "libx264", "-crf", "20", "-preset", "fast",
            "-pix_fmt", "yuv420p", "-c:a", "aac", "-b:a", "192k",
            "-movflags", "+faststart", dst]
    subprocess.run(cmd, check=True)
    for p in (c, c.replace(".mp4", ".props.json")):
        os.path.exists(p) and os.remove(p)
    got = dur(dst)
    print(f"  -> {dst}  {int(got // 60)}:{int(got % 60):02d}"
          f"  ({os.path.getsize(dst) / 1e6:.0f} MB)"
          + ("" if abs(got - total) < 1.0 else "   RUNTIME OFF"), flush=True)
    return dst


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("keys", nargs="*")
    ap.add_argument("--all", action="store_true")
    ap.add_argument("--list", action="store_true")
    ap.add_argument("--src", default=None,
                    help="read lectures from here instead of the delivery folder. "
                         "A staged Week 3 cut and its delivered copy are the same "
                         "file - deliver.py MOVES it - so a showcase built from "
                         "staged is byte-identical to one built after delivery.")
    a = ap.parse_args()
    global VIDS
    if a.src:
        VIDS = os.path.abspath(a.src)
        print(f"source: {VIDS}")
    keys = sorted(SHOWCASES) if a.all else a.keys
    if a.list or not keys:
        for k, sc in SHOWCASES.items():
            files = [os.path.join(VIDS, f"{l}_adam.mp4") for l in sc["lectures"]]
            t = sum(dur(f) for f in files if os.path.exists(f)) + CARD_SECS
            print(f"{k}  {sc['slug']:42} {', '.join(sc['lectures'])}  "
                  f"{int(t // 60)}:{int(t % 60):02d}")
        return 0
    for k in keys:
        build(k)
    return 0


if __name__ == "__main__":
    sys.exit(main())
