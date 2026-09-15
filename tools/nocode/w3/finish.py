#!/usr/bin/env python3
"""Buy and render the whole week, lecture by lecture, waiting out the top-ups.

    python3 finish.py            # everything in ORDER.md order
    python3 finish.py --plan

`render.py` stops at the first lecture it cannot afford, which is right when the
allowance is fixed for the month. It is wrong when the allowance is being topped
up while we work: the correct behaviour is to wait, buy one lecture, render it,
and go on - so the encoder is busy during the wait instead of after it.

A lecture is bought and rendered as a unit. Nothing is ever half-bought: a
lecture that runs out of voice halfway is silent from that point and the only
way to find out is to watch it.
"""
from __future__ import annotations
import json, os, subprocess, sys, time, urllib.request

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE); sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import adam_lay as A
import quota

NARR, CUTS = os.path.join(HERE, "narration"), os.path.join(HERE, "cuts")
OUT = os.path.abspath(os.path.join(HERE, "..", "..", "..", "videos", "nocode", "vids"))

ORDER = ["w3-01","w3-02","w3-03","w3-04","w3-05","w3-06",
         "w3-07","w3-08","w3-12","w3-13","w3-36","w3-37","w3-38",
         "w3-15","w3-16","w3-17","w3-18","w3-19","w3-20",
         "w3-23","w3-24","w3-27","w3-28","w3-29","w3-30","w3-31","w3-32",
         "w3-33","w3-34","w3-35","w3-39","w3-40","w3-41","w3-25","w3-26"]

WAIT = 240          # between allowance checks while a top-up lands
MAX_WAIT = 90 * 60  # give up on one lecture after this and move to the next


def allowance() -> int:
    req = urllib.request.Request("https://api.elevenlabs.io/v1/user/subscription",
                                 headers={"xi-api-key": A.KEY})
    s = json.load(urllib.request.urlopen(req, timeout=30))
    return s["character_limit"] - s["character_count"]


def cost(lec: str, have: set[str]) -> int:
    d = json.load(open(os.path.join(NARR, f"{lec}.json")))
    return sum(len(s) for sg in d["segments"]
               for s in A.sentences(sg["say"]) if s not in have)


def main() -> None:
    plan = "--plan" in sys.argv
    want = [a for a in sys.argv[1:] if a.startswith("w3-")] or ORDER
    done, failed, deferred = [], [], []

    for lec in want:
        cut = os.path.join(CUTS, f"{lec}.mp4")
        nar = os.path.join(NARR, f"{lec}.json")
        if not (os.path.exists(cut) and os.path.exists(nar)):
            print(f"{lec}: no cut or no narration", flush=True); failed.append(lec); continue

        c = cost(lec, quota.spoken())
        if plan:
            print(f"{lec}: {c} characters", flush=True); continue

        waited = 0
        while c > 0:
            left = allowance()
            if left >= c:
                break
            print(f"{lec}: needs {c}, allowance {left} - waiting for top-up "
                  f"({waited // 60} min so far)", flush=True)
            if waited >= MAX_WAIT:
                break
            time.sleep(WAIT); waited += WAIT

        if c > 0 and allowance() < c:
            print(f"{lec}: DEFERRED - still short after {waited // 60} min", flush=True)
            deferred.append(lec); continue

        if c > 0:
            print(f"=== {lec}: buying {c} characters", flush=True)
            if subprocess.run([sys.executable, "prefetch.py", lec], cwd=HERE).returncode:
                failed.append(lec); continue

        print(f"=== {lec}: rendering", flush=True)
        if subprocess.run([sys.executable, "assemble.py", lec], cwd=HERE).returncode:
            print(f"{lec}: ASSEMBLE FAILED", flush=True); failed.append(lec); continue
        done.append(lec)
        print(f"--- {lec} delivered  ({len(done)}/{len(want)})", flush=True)

    print(f"\ndelivered {len(done)}")
    if deferred: print("deferred (voice):", " ".join(deferred))
    if failed:   print("failed:", " ".join(failed))


if __name__ == "__main__":
    main()
