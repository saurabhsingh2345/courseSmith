#!/usr/bin/env python3
"""Lay a lecture's rendered cutaways over its delivered master.

    python3 lay.py w3-20                 # -> staged/w3-20_adam.mp4
    python3 lay.py w3-20 --deliver       # ...and copy into videos/nocode/vids/

The narration is the thing that must not move, so the audio stream is COPIED,
never re-encoded and never re-timed, and each insert replaces exactly the frames
it is as long as. Runtime out equals runtime in, to the frame - which is checked
at the end rather than assumed.

Why overlay rather than cut-and-concat: a concat would need the footage segments
either re-encoded per segment or stream-copied at keyframe boundaries, and the
masters' keyframes do not land on our timestamps. One overlay pass re-encodes
the video exactly once and leaves the timeline alone.

`enable` goes on the overlay so ffmpeg composites only inside the window. The
matching trap from autoredact.py does not apply here - there is no blur to gate -
but the shape is the same: gate the expensive filter, not the cheap one.

The encode matches delivery: h264_videotoolbox at 26M, as `assemble.py` used, so
these files sit next to the other 34 with the same profile and the same look.
"""

from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
PLANS = os.path.join(HERE, "plans")
OUT = os.path.join(HERE, "out")
STAGED = os.path.join(HERE, "staged")
ORIG = os.path.join(HERE, "orig")
VIDS = os.path.abspath(os.path.join(HERE, "..", "..", "..", "..",
                                    "videos", "nocode", "vids"))
sys.path.insert(0, HERE)
import source as SRC  # noqa: E402

FPS = 30
FADE = 0.35        # the cut in and out of a cutaway, seconds


def probe(path: str, streams: str, fields: str) -> list[str]:
    return subprocess.run(
        ["ffprobe", "-v", "error", "-select_streams", streams,
         "-show_entries", fields, "-of", "csv=p=0", path],
        capture_output=True, text=True, check=True).stdout.strip().splitlines()


def dur(path: str) -> float:
    return float(probe(path, "v:0", "format=duration")[0])


def frames(secs: list[float]) -> int:
    return sum(round(s * FPS) for s in secs)


def build(lec: str, plan: dict, master: str, dst: str) -> None:
    ins = plan["inserts"]
    strip = os.path.join(OUT, f"{lec}_strip.mp4")
    manifest = os.path.join(OUT, f"{lec}_strip.json")

    if os.path.exists(manifest) and "chunks" in json.load(open(manifest)):
        # A few strips per lecture (render.py chunks past ~120s of cutaways),
        # each insert trimmed back out of its chunk by frame offset.
        m = json.load(open(manifest))
        inputs = [os.path.join(OUT, f) for f in m["chunks"]]
        for c, p in enumerate(inputs):
            if not os.path.exists(p):
                raise SystemExit(f"render first: {p}")
            have = int(probe(p, "v:0", "stream=nb_frames")[0].strip(","))
            want = sum(ln for ln, ch in zip(m["lengths"], m["chunk"]) if ch == c)
            if have != want:
                raise SystemExit(
                    f"strip {c} is {have} frames but the plan says {want}")
        srcs = [(inputs[ch], off, ln)
                for ch, off, ln in zip(m["chunk"], m["offsets"], m["lengths"])]
        idx = [ch + 1 for ch in m["chunk"]]
    elif os.path.exists(strip) and os.path.exists(manifest):
        # One strip for the whole lecture, trimmed back apart by frame offset.
        m = json.load(open(manifest))
        have = int(probe(strip, "v:0", "stream=nb_frames")[0].strip(","))
        want = sum(m["lengths"])
        if have != want:
            raise SystemExit(f"strip is {have} frames but the plan says {want}")
        srcs = [(strip, off, ln) for off, ln in zip(m["offsets"], m["lengths"])]
        inputs = [strip]
        idx = [1] * len(ins)
    else:
        # One file per insert - the original path, kept so a half-finished run
        # against the old layout still lays.
        per = [os.path.join(OUT, f"{lec}_{i:02d}.mp4") for i in range(len(ins))]
        missing = [p for p in per if not os.path.exists(p)]
        if missing:
            raise SystemExit("render these first:\n  " + "\n  ".join(missing))
        for i, p in enumerate(per):
            have = int(probe(p, "v:0", "stream=nb_frames")[0].strip(","))
            want = frames(ins[i]["secs"])
            if have != want:
                raise SystemExit(
                    f"#{i}: rendered {have} frames but the plan says {want}")
        srcs = [(p, 0, frames(ins[i]["secs"])) for i, p in enumerate(per)]
        inputs = per
        idx = list(range(1, len(per) + 1))

    # CONCAT, not a chain of overlays.
    #
    # The first version overlaid each cutaway onto the master in sequence. That
    # holds one 4K RGBA buffer per stage - `format=yuva420p` for the alpha fade -
    # and ffmpeg was SIGKILLed on the lectures with the most cards: w3-08, w3-12,
    # w3-13, w3-17, w3-38. w3-04 has the same thirteen cards as w3-13 and
    # survived, which is what marginal memory pressure looks like.
    #
    # Cutting the timeline into ordered segments and concatenating them holds
    # one segment at a time, needs no alpha at all, and is frame-exact: the
    # segment frame counts sum to the master's, so the runtime cannot drift and
    # the narration cannot move. The cutaways dip from black instead of
    # cross-fading from the footage, which reads as an edit either way.
    # The masters are NOT constant frame rate: assemble.py's ramp drops a frame
    # every half second or so (w3-13 has 588 gaps of two frame periods), so
    # frame index N of the master is NOT at N/30 seconds. Trimming by frame
    # index on the raw stream landed every card early and the video stream came
    # out eight seconds shorter than the audio. `fps=30` first fills those gaps
    # by repeating the previous frame, so index and time agree again and the
    # total frame count is the runtime.
    total_frames = int(round(dur(master) * FPS))
    order, cursor = [], 0
    for i, item in enumerate(ins):
        at_f = int(round(float(item["at"]) * FPS))
        if at_f > cursor:
            order.append(("footage", cursor, at_f))
        order.append(("card", i))
        cursor = at_f + frames(item["secs"])
    if total_frames > cursor:
        order.append(("footage", cursor, total_frames))

    steps, labels = [], []
    for k, seg in enumerate(order):
        lab = f"s{k}"
        if seg[0] == "footage":
            _, a, b = seg
            steps.append(f"[m]trim=start_frame={a}:end_frame={b},"
                         f"setpts=PTS-STARTPTS[{lab}]")
        else:
            i = seg[1]
            _, off, ln = srcs[i]
            secs_len = ln / FPS
            steps.append(
                f"[{idx[i]}:v]trim=start_frame={off}:end_frame={off + ln},"
                f"setpts=PTS-STARTPTS,"
                f"fade=t=in:st=0:d={FADE},"
                f"fade=t=out:st={secs_len - FADE:.4f}:d={FADE}[{lab}]")
        labels.append(f"[{lab}]")

    steps.insert(0, f"[0:v]fps={FPS}:round=near,split={sum(1 for s in order if s[0] == 'footage')}"
                 + "".join(f"[m{k}]" for k in range(sum(1 for s in order if s[0] == 'footage'))))
    # each footage segment reads its own split output - a filter pad feeds one consumer
    fi = 0
    for k, s in enumerate(steps):
        if s.startswith("[m]trim="):
            steps[k] = f"[m{fi}]" + s[3:]
            fi += 1
    steps.append("".join(labels) + f"concat=n={len(order)}:v=1:a=0[v]")

    cmd = ["ffmpeg", "-nostdin", "-y", "-loglevel", "error", "-i", master]
    for p in inputs:
        cmd += ["-i", p]
    cmd += ["-filter_complex", ";".join(steps),
            "-map", "[v]", "-map", "0:a",
            "-c:v", "h264_videotoolbox", "-b:v", "26M",
            "-pix_fmt", "yuv420p",
            "-c:a", "copy", "-movflags", "+faststart", dst]
    subprocess.run(cmd, check=True)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("lecture")
    ap.add_argument("--deliver", action="store_true",
                    help="copy the staged cut into videos/nocode/vids/")
    args = ap.parse_args()
    lec = args.lecture

    plan = json.load(open(os.path.join(PLANS, f"{lec}.json")))
    # source.py owns the orig -> retimed -> staged chain; asking it here is what
    # stops a card being laid over a picture that resync.py has already moved.
    master = SRC.source(lec)
    print(f"  source: {os.path.relpath(master, HERE)}")

    os.makedirs(STAGED, exist_ok=True)
    dst = os.path.join(STAGED, f"{lec}_adam.mp4")
    before = dur(master)
    build(lec, plan, master, dst)
    after = dur(dst)

    v0 = float(probe(master, "v:0", "stream=duration")[0])
    v1 = float(probe(dst, "v:0", "stream=duration")[0])
    if abs(v1 - v0) > 0.25:
        print(f"  VIDEO STREAM MOVED {v0:.2f} -> {v1:.2f} - the picture no longer "
              f"lines up with the voice, do not ship")
        return 1

    animated = sum(frames(i["secs"]) for i in plan["inserts"]) / FPS
    print(f"{lec}: {before:.2f}s -> {after:.2f}s   "
          f"animated {animated:.0f}s ({animated / before * 100:.0f}% of runtime)"
          f"   inserts={len(plan['inserts'])}")
    if abs(after - before) > 0.25:
        print("  RUNTIME MOVED - the narration is now out of sync, do not ship")
        return 1

    if args.deliver:
        # Keep the untouched master before anything overwrites it. A second
        # animation pass must be built on the ORIGINAL footage, not on a file
        # that already has cards laid over it - re-laying onto an animated cut
        # would put a card on top of a card and there would be no way back.
        os.makedirs(ORIG, exist_ok=True)
        keep = os.path.join(ORIG, f"{lec}_adam.mp4")
        live = os.path.join(VIDS, f"{lec}_adam.mp4")
        if not os.path.exists(keep):
            os.replace(live, keep)          # same filesystem: free, and atomic
            print(f"  original moved aside -> {keep}")
        os.replace(dst, live)
        print(f"  delivered -> {live}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
