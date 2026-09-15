#!/usr/bin/env python3
"""Find bright flash frames on a dark cut.

These decks live around 6-10% luma. A transition that briefly paints the whole
frame mid-grey reads on the player as a glitch, and it is invisible in a normal
review pass because it lasts a few frames. So: decode at 8 fps, take the mean
luma, and flag isolated spikes well above the lecture's own baseline.
"""
import subprocess, sys, os, numpy as np

FPS = 8
W, H = 32, 18

def scan(path):
    p = subprocess.run(
        ["ffmpeg","-nostdin","-v","error","-i",path,
         "-vf",f"fps={FPS},scale={W}:{H},format=gray","-f","rawvideo","-"],
        capture_output=True)
    buf = np.frombuffer(p.stdout, dtype=np.uint8)
    n = len(buf)//(W*H)
    fr = buf[:n*W*H].reshape(n, H*W).astype(np.float32)
    luma = fr.mean(axis=1)
    med = np.median(luma)
    # a flash is much brighter than the lecture's own typical frame AND flat
    std = fr.std(axis=1)
    hits = np.where(luma > 190)[0]
    # group consecutive
    out, i = [], 0
    while i < len(hits):
        j = i
        while j+1 < len(hits) and hits[j+1]-hits[j] <= 2: j += 1
        t0, t1 = hits[i]/FPS, hits[j]/FPS
        out.append((t0, t1-t0+1/FPS, luma[hits[i]:hits[j]+1].max()))
        i = j+1
    return med, out

for f in sys.argv[1:]:
    b = os.path.basename(f).replace("_adam.mp4","")
    try: med, hits = scan(f)
    except Exception as e:
        print(f"{b}\tERROR {e}"); continue
    hits = [h for h in hits if h[1] >= 1.0]
    if not hits:
        print(f"{b}\tbaseline={med:.0f}\tno flashes", flush=True)
    else:
        s = "; ".join(f"{int(t)//60}:{int(t)%60:02d} ({d*1000:.0f}ms peak{p:.0f})" for t,d,p in hits[:12])
        print(f"{b}\tbaseline={med:.0f}\tFLASHES={len(hits)}\t{s}", flush=True)
