#!/usr/bin/env python3
"""Re-write the chapter marks on finished modules, stream-copy, in place.

    /usr/bin/python3 rechapter.py 10 11 13
    /usr/bin/python3 rechapter.py --all

The marks are recomputed from plan.json and the lecture durations exactly as
build.py computes them, so this is idempotent and safe to run over everything.
Exists because the first fourteen modules were built before build.py dropped
the source lectures' own chapters (see the -map_chapters note there).
"""
from __future__ import annotations
import json, os, subprocess, sys
from build import TITLES, TITLE, CHAP, END, XF, OUT, VIDS, WORK, dur, plan

def marks(m):
    N = len(m["lectures"])
    pieces = [TITLE]
    for l in m["lectures"]:
        pieces += [CHAP, dur(os.path.join(VIDS, f"{l['stem']}_adam.mp4"))]
    pieces.append(END)
    total = sum(pieces) - XF * (len(pieces) - 1)
    out, run = [], pieces[0]
    for i in range(1, len(pieces)):
        off = run - XF
        if i % 2 == 1 and i < len(pieces) - 1:      # a chapter card
            out.append((off, m["lectures"][i // 2]["title"]))
        run = run + pieces[i] - XF
    return out, total

def main():
    mods = plan()
    keys = sys.argv[1:]
    want = mods if "--all" in keys else [m for m in mods if str(m["n"]) in keys or f"{m['n']:02d}" in keys]
    os.makedirs(WORK, exist_ok=True)
    for m in want:
        n = m["n"]
        dst = os.path.join(OUT, f"module{n:02d}.mp4")
        if not os.path.exists(dst):
            print(f"M{n:02d}  not built yet"); continue
        mk, total = marks(m)
        got = dur(dst)
        lines, _ = TITLES[n]
        meta = os.path.join(WORK, f"m{n:02d}.rechap.txt")
        with open(meta, "w") as fh:
            fh.write(";FFMETADATA1\n")
            fh.write(f"title=Module {n:02d} - {' '.join(lines).replace('*', '')}\n")
            fh.write(f"album=No-Code AI Coding - Week {m['week']}\ntrack={n}\n")
            ends = [t for t, _ in mk][1:] + [got]
            for (t0, title), t1 in zip(mk, ends):
                fh.write(f"[CHAPTER]\nTIMEBASE=1/1000\nSTART={int(t0*1000)}\nEND={int(t1*1000)}\ntitle={title}\n")
        tmp = dst + ".rechap.mp4"
        subprocess.run(["ffmpeg", "-nostdin", "-y", "-loglevel", "error", "-i", dst, "-i", meta,
                        "-map", "0", "-map_metadata", "1", "-map_chapters", "1",
                        "-c", "copy", "-movflags", "+faststart", tmp], check=True)
        os.replace(tmp, dst); os.remove(meta)
        print(f"M{n:02d}  {len(mk)} chapters  " + "  ".join(f"{t:.1f}" for t, _ in mk))

if __name__ == "__main__":
    main()
