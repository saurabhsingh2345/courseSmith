#!/usr/bin/env python3
"""Write title / track / album metadata into the delivered lectures.

    /usr/bin/python3 tag.py --plan        # what would change
    /usr/bin/python3 tag.py w1            # Week 1 + 2 (nocode*)
    /usr/bin/python3 tag.py w3            # Week 3, AFTER deliver.py has run

The platform shows whatever title you type into it; these tags are what a player
and a file browser show, and they are what makes a folder of 98 mp4s navigable.
Container metadata only - `-c copy`, so no stream is re-encoded and no quality
is lost. The file is rewritten to a temp name and moved into place, so an
interrupted run cannot leave a truncated lecture.
"""
from __future__ import annotations
import json, os, subprocess, sys

HERE=os.path.dirname(os.path.abspath(__file__))
ROOT=os.path.abspath(os.path.join(HERE,"..","..",".."))
VIDS=os.path.join(ROOT,"videos","nocode","vids")
MAN=os.path.join(ROOT,"docs","nocode-course","upload","manifest.json")
ALBUM="AI Coder — from no-code to agentic engineer"

def rows():
    return json.load(open(MAN))

def tag(path: str, title: str, track: int, week: int, plan: bool) -> bool:
    cur=subprocess.run(["ffprobe","-v","error","-show_entries","format_tags=title",
                        "-of","csv=p=0",path],capture_output=True,text=True).stdout.strip()
    want=f"{track:02d} — {title}"
    if cur==want:
        return False
    if plan:
        print(f"  {os.path.basename(path):24} {cur!r} -> {want!r}")
        return True
    tmp=path+".tag.mp4"
    r=subprocess.run(["ffmpeg","-nostdin","-y","-loglevel","error","-i",path,
                      "-map","0","-c","copy","-movflags","+faststart",
                      "-metadata",f"title={want}",
                      "-metadata",f"album={ALBUM}",
                      "-metadata",f"track={track}",
                      "-metadata",f"comment=Section {week}",
                      tmp])
    if r.returncode or not os.path.exists(tmp):
        print(f"  {os.path.basename(path)} FAILED"); os.path.exists(tmp) and os.remove(tmp)
        return False
    os.replace(tmp,path)
    print(f"  {track:02d} {os.path.basename(path):24} {title[:60]}")
    return True

def main() -> int:
    plan="--plan" in sys.argv
    which=[a for a in sys.argv[1:] if not a.startswith("--")] or ["w1","w3"]
    n=0
    for r in rows():
        is_w3=r["file"].startswith("w3-")
        if is_w3 and "w3" not in which: continue
        if not is_w3 and "w1" not in which: continue
        p=os.path.join(VIDS,r["file"])
        if not os.path.exists(p):
            print(f"  missing {r['file']}"); continue
        n+=tag(p,r["title"],r["pos"],r["week"],plan)
    print(f"{n} file(s) {'would change' if plan else 'tagged'}")
    return 0

if __name__=="__main__":
    sys.exit(main())
