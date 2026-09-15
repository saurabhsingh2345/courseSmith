#!/usr/bin/env python3
"""Fit narration to footage using the REAL voice, not a words-per-minute guess.

    python3 realbalance.py                 # report every lecture
    python3 realbalance.py --write         # move, then trim, until nothing freezes
    python3 realbalance.py w3-04 --write

`fit.py` and `balance.py` estimate at 198 words per minute. That average is right
across a lecture and wrong inside a segment: a sentence full of short words and
commas runs long, and the deepening pass wrote a lot of those. Several lectures
came out of the render with HELD ON LAST FRAME, which is a freeze on screen.

Every sentence is already in the content store as an mp3 named for the sha1 of
its own text, so the exact duration is on disk and costs nothing to read. This
measures it, adds assemble.py's own 0.42s inter-sentence gap, and compares
against footage / 0.92 - the same ceiling assemble uses.

Overflow moves to the next segment first, exactly like balance.py, and is only
trimmed when there is nowhere for it to go.
"""
from __future__ import annotations
import glob, hashlib, json, os, subprocess, sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE); sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import adam_lay as A

NARR, CUTS = os.path.join(HERE, "narration"), os.path.join(HERE, "cuts")
STORE = os.path.join(HERE, "..", "scripts", "_voice")
GAP, MIN_SPEED, WPM = 0.42, 0.92, 198.0

# An mp3's container duration counts encoder padding that the decoder throws
# away, so ffprobe reads about 4% longer than the samples assemble.py actually
# concatenates. Measured against four segments of w3-04's render log:
# 118.8/123.2, 32.4/33.9, 125.4/130.6, 21.1/21.9 - all within half a percent
# of the same figure. Without this the trimmer cuts writing that would have fit.
DECODE = 0.961

_cache: dict[str, float] = {}


def clip_seconds(sent: str) -> float:
    """Measured if we own the clip, estimated if we do not yet."""
    if sent in _cache:
        return _cache[sent]
    p = os.path.join(STORE, hashlib.sha1(sent.encode("utf-8")).hexdigest() + ".mp3")
    if os.path.exists(p):
        out = subprocess.run(["ffprobe", "-v", "error", "-show_entries",
                              "format=duration", "-of", "csv=p=0", p],
                             capture_output=True, text=True).stdout.strip()
        v = float(out) if out else len(sent.split()) / WPM * 60
    else:
        v = len(sent.split()) / WPM * 60
    _cache[sent] = v
    return v


def seg_seconds(say: str) -> float:
    s = A.sentences(say)
    return sum(clip_seconds(x) for x in s) * DECODE + GAP * max(0, len(s) - 1)


def dur(p: str) -> float:
    return float(subprocess.run(["ffprobe", "-v", "error", "-show_entries",
        "format=duration", "-of", "csv=p=0", p],
        capture_output=True, text=True).stdout.strip())


def process(lec: str, write: bool) -> tuple[int, int, float]:
    path = os.path.join(NARR, f"{lec}.json")
    cut = os.path.join(CUTS, f"{lec}.mp4")
    if not (os.path.exists(path) and os.path.exists(cut)):
        return 0, 0, 0.0
    d = json.load(open(path))
    segs = d["segments"]
    total = dur(cut)

    def have(i: int) -> float:
        f0 = float(segs[i]["anchor"])
        f1 = float(segs[i + 1]["anchor"]) if i + 1 < len(segs) else total
        return (f1 - f0) / MIN_SPEED

    moved = trimmed = 0
    # forward pass: push whole trailing sentences into the next segment
    for i in range(len(segs) - 1):
        while seg_seconds(segs[i]["say"]) > have(i):
            s = A.sentences(segs[i]["say"])
            if len(s) < 2:
                break
            nxt = A.sentences(segs[i + 1]["say"])
            if seg_seconds(" ".join([s[-1]] + nxt)) > have(i + 1):
                break
            segs[i]["say"] = " ".join(s[:-1])
            segs[i + 1]["say"] = " ".join([s[-1]] + nxt)
            moved += 1
    # last resort: drop trailing sentences that fit nowhere
    for i in range(len(segs)):
        while seg_seconds(segs[i]["say"]) > have(i):
            s = A.sentences(segs[i]["say"])
            if len(s) < 2:
                break
            segs[i]["say"] = " ".join(s[:-1])
            trimmed += 1

    over = sum(max(0.0, seg_seconds(s["say"]) - have(i))
               for i, s in enumerate(segs))
    if write and (moved or trimmed):
        json.dump(d, open(path, "w"), indent=2, ensure_ascii=False)
    return moved, trimmed, over


def main() -> None:
    write = "--write" in sys.argv
    want = [a for a in sys.argv[1:] if a.startswith("w3-")]
    lecs = want or sorted(os.path.basename(p)[:-5]
                          for p in glob.glob(os.path.join(NARR, "w3-*.json")))
    tm = tt = 0
    for lec in lecs:
        m, t, over = process(lec, write)
        if m or t or over > 0.5:
            print(f"{lec}: moved {m}, trimmed {t}"
                  f"{f', STILL {over:.1f}s over' if over > 0.5 else ''}", flush=True)
        tm += m; tt += t
    print(f"\nmoved {tm}, trimmed {tt}")


if __name__ == "__main__":
    main()
