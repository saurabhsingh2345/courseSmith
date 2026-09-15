#!/usr/bin/env python3
"""Cut the waiting out of a screen take, so the picture never stops.

The defect this exists for: `assemble_l01.py` retimes a whole take by ONE
factor to hit a narration length. Segment J runs 138.5s of source at 1.13x -
and 90 seconds of that source is an agent thinking, with the frame not moving
by a single pixel. The delivered lecture therefore holds one frame for 140
seconds while the voice keeps talking, which is what viewers report as "the
audio does not match the video".

The fix is not a different global rate. A take is live in bursts and dead in
between, so the two need different treatment: live picture plays at its own
pace, and each wait is ramped hard down to a beat. That is what a human editor
does with a build, and it is why the result reads as edited rather than padded.

    /usr/bin/python3 densify.py TAKE.mp4 OUT.mp4 [--hold 1.1] [--min-dead 2.5]
                                [--max-speed 24] [--live-rate 1.0] [--plan p.json]

Writes OUT.mp4 and, with --plan, the time map: for every kept piece, where it
came from and where it landed. Narration is written against that map, so a line
can be aimed at the second a command lands rather than at a guess.
"""
from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys

import numpy as np

FPS_SAMPLE = 4
W, H, CELL, CELLS = 320, 180, 6, 40


def dur(p: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def dead_spans(path: str, min_dead: float) -> list[tuple[float, float]]:
    """Same detector as qa/deadzones.py - count cells that moved, do not
    average - so a few lines of terminal text still count as a live picture."""
    p = subprocess.run(
        ["ffmpeg", "-nostdin", "-v", "error", "-i", path, "-vf",
         f"fps={FPS_SAMPLE},scale={W}:{H},format=gray", "-f", "rawvideo", "-"],
        capture_output=True)
    buf = np.frombuffer(p.stdout, dtype=np.uint8)
    n = len(buf) // (W * H)
    if n < 2:
        raise SystemExit(f"no frames decoded from {path}")
    fr = buf[:n * W * H].reshape(n, H * W).astype(np.int16)
    moved = (np.abs(np.diff(fr, axis=0)) > CELL).sum(axis=1)
    still = moved < CELLS

    out, s = [], None
    for i, v in enumerate(still):
        if v and s is None:
            s = i
        elif not v and s is not None:
            out.append((s, i))
            s = None
    if s is not None:
        out.append((s, len(still)))
    return [(a / FPS_SAMPLE, (b + 1) / FPS_SAMPLE) for a, b in out
            if (b + 1 - a) / FPS_SAMPLE >= min_dead]


def plan(path: str, min_dead: float, hold: float, max_speed: float,
         live_rate: float, lo: float = 0.0, hi: float | None = None) -> list[dict]:
    """Alternating live and ramped pieces covering [lo, hi) of the take.

    A take usually holds several segments of the lecture, each with its own
    narration, so the range matters: densifying the whole file would move the
    boundaries the narration was written against.
    """
    total = dur(path) if hi is None else hi
    dead = [(max(a, lo), min(b, total)) for a, b in dead_spans(path, min_dead)
            if b > lo and a < total]
    dead = [(a, b) for a, b in dead if b - a >= min_dead * 0.6]
    pieces, at = [], lo
    for a, b in dead:
        if a - at > 0.05:
            pieces.append({"from": at, "to": a, "kind": "live", "rate": live_rate})
        # A wait becomes a beat. Never slower than realtime, never so fast the
        # ramp reads as a glitch - past about 24x a long think is a single
        # flicker and the viewer cannot tell what was skipped.
        span = b - a
        rate = min(max_speed, max(1.0, span / hold))
        pieces.append({"from": a, "to": b, "kind": "ramp", "rate": rate})
        at = b
    if total - at > 0.05:
        pieces.append({"from": at, "to": total, "kind": "live", "rate": live_rate})

    out = 0.0
    for p in pieces:
        p["len"] = (p["to"] - p["from"]) / p["rate"]
        p["at"] = out
        out += p["len"]
    return pieces


def fit(path: str, lo: float, hi: float, need: float, min_dead: float,
        max_hold: float = 4.5, max_speed: float = 24.0) -> list[dict] | None:
    """Densify [lo, hi) to land on exactly `need` seconds.

    Keeping every live frame and letting the WAITS take up the slack is the
    right trade: the alternative is a shot that runs out of picture and freezes
    on its last frame, which is the defect this whole pass exists to remove.
    Each wait becomes the same length, so the cut still has a rhythm.

    Returns None when the range simply has not got `need` seconds in it even
    with every wait stretched - then the words belong on a card, not on
    footage, and the caller has to be told rather than handed a frozen shot.
    """
    pieces = plan(path, min_dead, 1.0, max_speed, 1.0, lo, hi)
    live_len = sum(p["to"] - p["from"] for p in pieces if p["kind"] == "live")
    ramps = [p for p in pieces if p["kind"] == "ramp"]

    # More live picture than words is not a problem: the caller takes the
    # seconds it needs off the front and the rest is simply unused. Only a
    # SHORTFALL can freeze a shot, so only that has to be solved.
    hold = 1.0 if live_len >= need or not ramps else (need - live_len) / len(ramps)
    hold = max(0.8, min(max_hold, hold))
    total = live_len + hold * len(ramps)

    # Still short, and there is no wait left to stretch. A few percent off the
    # playback rate of a screen recording is invisible and is what an editor
    # would reach for before dropping the line; beyond that it reads as slow
    # motion, so it is refused and the words go on a card instead.
    slow = 1.0
    if total < need - 0.05:
        slow = total / need
        if slow < 0.82:
            return None

    out, at = [], 0.0
    for p in pieces:
        span = p["to"] - p["from"]
        rate = slow if p["kind"] == "live" else \
            max(slow, min(max_speed, span / hold))
        q = dict(p, rate=rate, len=span / rate, at=at)
        at += q["len"]
        out.append(q)
    return out


def render(src: str, dst: str, pieces: list[dict], crf: int) -> None:
    # One filter graph rather than a temp file per piece: a 40-piece take would
    # otherwise be 40 encodes and 40 generation losses.
    steps, labels = [], []
    for i, p in enumerate(pieces):
        steps.append(
            f"[0:v]trim=start={p['from']:.3f}:end={p['to']:.3f},"
            f"setpts=(PTS-STARTPTS)/{p['rate']:.6f}[v{i}]")
        labels.append(f"[v{i}]")
    steps.append("".join(labels) + f"concat=n={len(pieces)}:v=1:a=0[cat]")
    steps.append("[cat]scale=1920:1080:flags=lanczos,fps=30,setsar=1,"
                 "format=yuv420p[out]")
    subprocess.run(
        ["ffmpeg", "-nostdin", "-y", "-loglevel", "error", "-i", src,
         "-filter_complex", ";".join(steps), "-map", "[out]", "-an",
         "-c:v", "libx264", "-preset", "medium", "-crf", str(crf),
         "-pix_fmt", "yuv420p", "-g", "60", dst], check=True)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("src")
    ap.add_argument("dst")
    ap.add_argument("--hold", type=float, default=1.1,
                    help="seconds a ramped wait becomes")
    ap.add_argument("--min-dead", type=float, default=2.5,
                    help="shortest wait worth ramping")
    ap.add_argument("--max-speed", type=float, default=24.0)
    ap.add_argument("--live-rate", type=float, default=1.0,
                    help=">1 tightens the live picture too")
    ap.add_argument("--crf", type=int, default=17)
    ap.add_argument("--plan")
    ap.add_argument("--dry", action="store_true")
    a = ap.parse_args()

    pieces = plan(a.src, a.min_dead, a.hold, a.max_speed, a.live_rate)
    src_len = dur(a.src)
    out_len = sum(p["len"] for p in pieces)
    live = sum(p["len"] for p in pieces if p["kind"] == "live")
    print(f"{os.path.basename(a.src)}  {src_len:.1f}s -> {out_len:.1f}s "
          f"({len(pieces)} pieces, {sum(1 for p in pieces if p['kind']=='ramp')} ramps)")
    print(f"  live picture {live:.1f}s = {100*live/out_len:.0f}% of the cut")
    for p in pieces:
        if p["kind"] == "ramp":
            print(f"    ramp {p['from']:7.1f}-{p['to']:7.1f} "
                  f"({p['to']-p['from']:5.1f}s) at {p['rate']:5.1f}x "
                  f"-> {p['len']:4.1f}s at {p['at']:6.1f}")
    if a.plan:
        json.dump({"src": a.src, "src_len": src_len, "out_len": out_len,
                   "pieces": pieces}, open(a.plan, "w"), indent=1)
        print(f"  plan -> {a.plan}")
    if a.dry:
        return 0
    render(a.src, a.dst, pieces, a.crf)
    got = dur(a.dst)
    print(f"  wrote {a.dst}  {got:.1f}s"
          + ("" if abs(got - out_len) < 0.5 else f"  WANTED {out_len:.1f}"))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
