#!/usr/bin/env python3
"""How long does this lecture leave the same picture on screen?

Decodes at 2 fps into tiny greyscale frames and measures the run length of
"nothing changed". A slide deck that holds 30s per card and a screen recording
that sits on an unanswered prompt look identical to a student: nothing is
happening. The metric is deliberately crude - mean absolute difference over a
64x36 luma - because that is what survives video compression noise.
"""
import subprocess, sys, os, numpy as np

FPS = 2
W, H = 64, 36
THRESH = 1.4      # mean abs luma delta below this = "same picture"

def holds(path):
    p = subprocess.run(
        ["ffmpeg","-nostdin","-v","error","-i",path,
         "-vf",f"fps={FPS},scale={W}:{H},format=gray","-f","rawvideo","-"],
        capture_output=True)
    buf = np.frombuffer(p.stdout, dtype=np.uint8)
    n = len(buf)//(W*H)
    fr = buf[:n*W*H].reshape(n, H*W).astype(np.int16)
    d = np.abs(np.diff(fr, axis=0)).mean(axis=1)
    same = d < THRESH
    runs, cur = [], 1
    for s in same:
        if s: cur += 1
        else:
            runs.append(cur); cur = 1
    runs.append(cur)
    secs = np.array(runs)/FPS
    total = n/FPS
    return dict(
        dur=total,
        longest=secs.max(),
        n_holds_over_15=int((secs>15).sum()),
        n_holds_over_30=int((secs>30).sum()),
        frac_in_holds_over_15=float(secs[secs>15].sum()/total),
        changes_per_min=float(len(runs)/(total/60)),
    )

if __name__ == "__main__":
    for f in sys.argv[1:]:
        b = os.path.basename(f).replace("_adam.mp4","")
        try:
            r = holds(f)
        except Exception as e:
            print(f"{b}\tERROR {e}"); continue
        print(f"{b}\tdur={r['dur']:.0f}\tlongest_hold={r['longest']:.0f}s"
              f"\tholds>15s={r['n_holds_over_15']}\tholds>30s={r['n_holds_over_30']}"
              f"\tpct_time_in_long_holds={r['frac_in_holds_over_15']*100:.0f}%"
              f"\tcuts/min={r['changes_per_min']:.1f}", flush=True)
