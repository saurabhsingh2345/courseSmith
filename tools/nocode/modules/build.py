#!/usr/bin/env python3
"""Assemble one ~25 minute course module out of consecutive finished lectures.

    /usr/bin/python3 build.py --list
    /usr/bin/python3 build.py 04
    /usr/bin/python3 build.py 04 --cards      # render only the card strip
    /usr/bin/python3 build.py --all

A module is: a title card, then for each lecture a chapter card and the lecture
itself, then a card naming what comes next. Every join is a half-second
cross-dissolve, and the whole thing carries MP4 chapter marks so a player shows
the lectures inside it.

Three things worth knowing.

**One Remotion render per module, not one per card.** `W3Insert` takes
`slides` and `secs` as props and sizes itself from them, so the title card,
every chapter card and the end card come out as a single strip and the pieces
are sliced back out of it with `-ss`/`-t`. Forty-two renders instead of a
hundred and eighty.

**`-ss` and `-t` go BEFORE `-i`.** After `-i`, `-t` is an output option and the
last one silently caps the whole module at one piece's length - `reel.py` has
the same warning on it for the same reason.

**The lectures are never modified.** Redaction boxes from `redactions.json` are
applied inside this one re-encode, in the lecture's own pixels before the scale
to 1080p, so Week 3's 4K boxes and Week 1's 1080p boxes both land correctly and
nothing is ever written back to the delivery folder.
"""
from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys

from titles import M as TITLES

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
# Normally the delivery folder. `NOCODE_VIDS` points it at a rework folder so a
# rebuilt lecture can be cut into a module and reviewed before anything in
# vids/ is touched - the delivered files are the course, and a module is only
# ever a re-encode that READS them.
VIDS = os.environ.get("NOCODE_VIDS") or os.path.join(ROOT, "videos", "nocode", "vids")
RENDERER = os.path.join(ROOT, "renderer")
# Also overridable: a rework build must not overwrite a shipped module.
OUT = os.environ.get("NOCODE_MODULES") or os.path.join(ROOT, "videos", "nocode", "modules")
WORK = os.path.join(OUT, ".work")

TITLE, CHAP, END, XF, FADE = 5.5, 2.6, 4.5, 0.5, 1.2
WEEK_NEXT = {1: ["Week *two*", "next"], 2: ["Week *three*", "next"],
             3: ["That is the", "*course*"]}


def plan() -> list[dict]:
    with open(os.path.join(HERE, "plan.json")) as fh:
        return json.load(fh)


def redactions() -> dict:
    # Every window in redactions.json was measured against a particular cut. A
    # rebuilt lecture has different timings, so those windows would blur the
    # wrong seconds - point this at a file that matches the sources being cut.
    p = os.environ.get("NOCODE_REDACTIONS") or os.path.join(HERE, "redactions.json")
    return json.load(open(p)) if os.path.exists(p) else {}


def dur(p: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def slides(m: dict, mods: list[dict]) -> tuple[list[dict], list[float]]:
    n, N = m["n"], len(m["lectures"])
    lines, stamp = TITLES[n]
    week, block = m["week"], m["blocktitle"]
    out = [{"k": "head", "kicker": f"week {week}  ·  {block.lower()}  ·  module {n:02d}",
            "lines": lines, "size": 120, "stamp": stamp, "trans": "fade"}]
    secs = [TITLE]

    for i, lec in enumerate(m["lectures"]):
        if N == 1:
            # A one-row contents list is not a contents list. Name the lecture.
            out.append({"k": "head", "kicker": f"module {n:02d}  ·  one lecture",
                        "lines": [lec["title"]], "size": 76, "trans": "fade"})
        else:
            out.append({"k": "chapter",
                        "kicker": f"module {n:02d}  ·  chapter {i + 1} of {N}",
                        "idx": i,
                        "items": [l["title"] for l in m["lectures"]],
                        "trans": "fade"})
        secs.append(CHAP)

    nxt = next((x for x in mods if x["n"] == n + 1), None)
    if nxt and nxt["week"] == week:
        nl, ns = TITLES[nxt["n"]]
        out.append({"k": "head", "kicker": f"next up  ·  module {nxt['n']:02d}",
                    "lines": nl, "size": 96, "stamp": ns, "trans": "fade"})
    else:
        out.append({"k": "head", "kicker": f"week {week} complete",
                    "lines": WEEK_NEXT[week], "size": 104,
                    "stamp": f"module {n:02d} of 42  ·  end of week {week}",
                    "trans": "fade"})
    secs.append(END)
    return out, secs


def strip(m: dict, mods: list[dict]) -> str:
    os.makedirs(WORK, exist_ok=True)
    sl, secs = slides(m, mods)
    props = os.path.join(WORK, f"m{m['n']:02d}.props.json")
    dst = os.path.join(WORK, f"m{m['n']:02d}.cards.mp4")
    with open(props, "w") as fh:
        json.dump({"slides": sl, "secs": secs}, fh, indent=1)
    subprocess.run(["npx", "remotion", "render", "src/nocode/index.tsx",
                    "W3Insert", dst, f"--props={props}", "--log=error"],
                   cwd=RENDERER, check=True)
    return dst


def blur_chain(stem: str, red: dict, src: str) -> tuple[str, str]:
    """The redaction filter for one lecture, in its own pixels. Returns
    (filter prefix ending in the named output, that name)."""
    boxes = red.get(stem, [])
    if not boxes:
        return "", src
    parts, last = [], src
    for i, (t0, t1, x, y, w, h) in enumerate(boxes):
        x, y, w, h = int(x), int(y), int(w), int(h)
        # `enable` must gate the BLUR as well as the overlay, or ffmpeg blurs
        # every frame and throws all but the window away - the same trap
        # w3/redact.py documents, where nine boxes cost twenty minutes.
        win = f"between(t,{t0},{t1})"
        tag = f"{src}r{i}"
        # boxblur refuses a radius that is not under half the plane's smaller
        # side, and the chroma plane is half the box. A 40px-tall band gets a
        # luma radius of 12 but a chroma radius of 9 at most; power 2, applied
        # twice, is what makes a small radius enough to destroy the text.
        m = min(w, h)
        lr = max(1, min(12, m // 2 - 1))
        cr = max(1, min(6, m // 4 - 1))
        bb = f"boxblur=lr={lr}:lp=2:cr={cr}:cp=2:enable='{win}'"
        parts.append(
            f"[{last}]split[{tag}k][{tag}c];"
            f"[{tag}c]crop={w}:{h}:{x}:{y},{bb},{bb}[{tag}b];"
            f"[{tag}k][{tag}b]overlay={x}:{y}:enable='{win}'[{tag}o]")
        last = f"{tag}o"
    return ";".join(parts) + ";", last


def build(m: dict, mods: list[dict], red: dict) -> str:
    os.makedirs(OUT, exist_ok=True)
    n, N = m["n"], len(m["lectures"])
    cards = strip(m, mods)

    # (path, ss, length, kind, stem)
    pieces = [(cards, 0.0, TITLE, "card", None)]
    for i, lec in enumerate(m["lectures"]):
        p = os.path.join(VIDS, f"{lec['stem']}_adam.mp4")
        if not os.path.exists(p):
            raise SystemExit(f"missing {p}")
        pieces.append((cards, TITLE + i * CHAP, CHAP, "card", None))
        pieces.append((p, 0.0, dur(p), "lec", lec["stem"]))
    pieces.append((cards, TITLE + N * CHAP, END, "card", None))

    total = sum(p[2] for p in pieces) - XF * (len(pieces) - 1)

    cmd = ["ffmpeg", "-nostdin", "-y", "-loglevel", "error"]
    for path, ss, ln, _, _ in pieces:
        cmd += ["-ss", f"{ss:.3f}", "-t", f"{ln:.3f}", "-i", path]

    steps, nred = [], 0
    for i, (_, _, ln, kind, stem) in enumerate(pieces):
        pre, vin = ("", f"{i}:v")
        if kind == "lec":
            pre, vin = blur_chain(stem, red, f"{i}:v")
            nred += len(red.get(stem, []))
        steps.append(f"{pre}[{vin}]scale=1920:1080:flags=lanczos,fps=30,setsar=1,"
                     f"format=yuv420p,setpts=PTS-STARTPTS[v{i}]")
        if kind == "card":
            steps.append(f"anullsrc=r=48000:cl=stereo,atrim=0:{ln},"
                         f"asetpts=PTS-STARTPTS[a{i}]")
        else:
            steps.append(f"[{i}:a]aresample=48000,"
                         f"aformat=channel_layouts=stereo,asetpts=PTS-STARTPTS[a{i}]")

    # xfade offsets are cumulative and each transition eats XF, so the offset
    # is the running length of everything already joined, minus XF.
    vprev, aprev, run, marks = "v0", "a0", pieces[0][2], []
    for i in range(1, len(pieces)):
        off = run - XF
        # A chapter mark lands on the chapter CARD, so jumping to a chapter
        # shows its intro rather than dropping in two seconds into the voice.
        if pieces[i][3] == "card" and i + 1 < len(pieces) and pieces[i + 1][3] == "lec":
            marks.append((off, pieces[i + 1][4]))
        steps.append(f"[{vprev}][v{i}]xfade=transition=fade:duration={XF}:"
                     f"offset={off:.3f}[vx{i}]")
        steps.append(f"[{aprev}][a{i}]acrossfade=d={XF}:c1=tri:c2=tri[ax{i}]")
        vprev, aprev = f"vx{i}", f"ax{i}"
        run = run + pieces[i][2] - XF

    steps.append(f"[{vprev}]fade=t=out:st={total - FADE:.3f}:d={FADE}[vend]")
    steps.append(f"[{aprev}]afade=t=out:st={total - FADE:.3f}:d={FADE}[aend]")

    raw = os.path.join(WORK, f"m{n:02d}.raw.mp4")
    cmd += ["-filter_complex", ";".join(steps),
            "-map", "[vend]", "-map", "[aend]",
            # Some delivered lectures carry their own chapter marks (the Day
            # 4/5 files built on the second Mac), and ffmpeg copies chapters
            # from the first input that has any. Drop them here so the marks
            # written below are the only ones on the module.
            "-map_chapters", "-1",
            "-c:v", "libx264", "-crf", "20", "-preset", "fast",
            "-pix_fmt", "yuv420p", "-c:a", "aac", "-b:a", "192k",
            "-movflags", "+faststart", raw]
    print(f"M{n:02d}  {N} lec  {len(pieces)} pieces  "
          f"{int(total // 60)}:{int(total % 60):02d}"
          + (f"  {nred} redaction box(es)" if nred else ""), flush=True)
    subprocess.run(cmd, check=True)

    # Chapters go on in a stream copy, so the encode above stays one pass.
    lines, _ = TITLES[n]
    name = " ".join(lines).replace("*", "")
    meta = os.path.join(WORK, f"m{n:02d}.meta.txt")
    ends = [t for t, _ in marks][1:] + [total]
    with open(meta, "w") as fh:
        fh.write(";FFMETADATA1\n")
        fh.write(f"title=Module {n:02d} - {name}\n")
        fh.write(f"album=No-Code AI Coding - Week {m['week']}\n")
        fh.write(f"track={n}\n")
        by = {l["stem"]: l["title"] for l in m["lectures"]}
        for (t0, stem), t1 in zip(marks, ends):
            fh.write("[CHAPTER]\nTIMEBASE=1/1000\n"
                     f"START={int(t0 * 1000)}\nEND={int(t1 * 1000)}\n"
                     f"title={by[stem]}\n")

    dst = os.path.join(OUT, f"module{n:02d}.mp4")
    subprocess.run(["ffmpeg", "-nostdin", "-y", "-loglevel", "error",
                    "-i", raw, "-i", meta, "-map_metadata", "1",
                    "-map_chapters", "1",
                    "-c", "copy", "-movflags", "+faststart", dst], check=True)
    for p in (raw, meta, cards, cards.replace(".cards.mp4", ".props.json")):
        os.path.exists(p) and os.remove(p)

    got = dur(dst)
    print(f"  -> {os.path.basename(dst)}  {int(got // 60)}:{int(got % 60):02d}"
          f"  ({os.path.getsize(dst) / 1e6:.0f} MB)  {len(marks)} chapters"
          + ("" if abs(got - total) < 1.5 else f"   RUNTIME OFF (wanted {total:.1f})"),
          flush=True)
    return dst


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("keys", nargs="*")
    ap.add_argument("--all", action="store_true")
    ap.add_argument("--list", action="store_true")
    ap.add_argument("--cards", action="store_true",
                    help="render only the card strip, leave it in .work/")
    a = ap.parse_args()
    mods = plan()
    red = redactions()
    if a.list:
        for m in mods:
            lines, _ = TITLES[m["n"]]
            print(f"M{m['n']:02d}  week {m['week']}  {m['blocktitle']:24} "
                  f"{int(m['secs']//60):3d}:{int(m['secs']%60):02d}  "
                  f"{' '.join(lines).replace('*','')}")
        return 0
    want = [m for m in mods if a.all or str(m["n"]) in a.keys
            or f"{m['n']:02d}" in a.keys]
    if not want:
        print("nothing selected; try --list")
        return 1
    for m in want:
        if a.cards:
            print(strip(m, mods))
        else:
            build(m, mods, red)
    return 0


if __name__ == "__main__":
    sys.exit(main())
