#!/usr/bin/env python3
"""Write docs/nocode-course/upload/MODULES.md from plan.json and what is built.

    /usr/bin/python3 manifest.py
"""
from __future__ import annotations

import json
import os
import subprocess

from spine import WEEK_TITLE
from titles import M as TITLES

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
OUT = os.path.join(ROOT, "videos", "nocode", "modules")
DOC = os.path.join(ROOT, "docs", "nocode-course", "upload", "MODULES.md")


def dur(p: str) -> float | None:
    if not os.path.exists(p):
        return None
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def mmss(s: float) -> str:
    return f"{int(s // 60)}:{int(s % 60):02d}"


def main() -> None:
    mods = json.load(open(os.path.join(HERE, "plan.json")))
    lines = ["# Module manifest — the course in ~25 minute chunks", ""]
    built = [dur(os.path.join(OUT, f"module{m['n']:02d}.mp4")) for m in mods]
    have = [d for d in built if d is not None]
    tot = sum(have)
    lines += [
        f"**{len(mods)} modules"
        + (f", {int(tot // 3600)}h {int(tot % 3600 // 60):02d}m" if len(have) == len(mods)
           else f" ({len(have)} built so far)")
        + ".** Files are in `videos/nocode/modules/`. Upload in the `#` order below.",
        "",
        "Each module is a title card, then its lectures with a chapter card between",
        "them, then a card naming the next module. Every join is a half-second",
        "dissolve. The MP4s carry chapter marks at each lecture, so a player that",
        "shows chapters shows the lectures. Built by `tools/nocode/modules/build.py`",
        "from the 98 delivered lectures, which are unchanged. Name/company leaks",
        "found by OCR are blurred in the module (`redactions.json`).",
        "",
    ]
    week = None
    for m, d in zip(mods, built):
        if m["week"] != week:
            week = m["week"]
            wk = [x for x in mods if x["week"] == week]
            lines += ["", f"## Week {week} — {WEEK_TITLE[week]}", "",
                      f"{len(wk)} modules", "",
                      "| # | file | title | inside | length |",
                      "|---|---|---|---|---|"]
        title = " ".join(TITLES[m["n"]][0]).replace("*", "")
        inside = "<br>".join(f"{i + 1}. {l['title']}" for i, l in enumerate(m["lectures"]))
        length = mmss(d) if d is not None else f"({mmss(m['secs'])} planned)"
        lines.append(f"| {m['n']} | `module{m['n']:02d}.mp4` | {m['blocktitle']} — {title} "
                     f"| {inside} | {length} |")
    lines += ["", "## Lecture → module", "", "| lecture | module |", "|---|---|"]
    for m in mods:
        for l in m["lectures"]:
            lines.append(f"| `{l['stem']}_adam.mp4` | {m['n']} |")
    with open(DOC, "w") as fh:
        fh.write("\n".join(lines) + "\n")
    print(f"-> {DOC}  ({len(have)}/{len(mods)} built)")


if __name__ == "__main__":
    main()
