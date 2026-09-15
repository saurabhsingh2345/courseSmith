#!/usr/bin/env python3
"""Cut the day-one takes into exact slide windows.

Each clip's picture length is retimed to equal its slide's window (narration +
inter-slide gap) so the swap costs no timing anywhere else in the film.
"""
import os, subprocess, sys
R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
T = f"{R}/tools/nocode/scripts/takes"

# lecture, slide index, source take, in, out, window seconds
JOBS = [
 ("06", 16, "A_ide",    0.6, 20.0, 18.10),
 ("06", 17, "D_plugin", 3.0, 24.0, 18.80),
 ("06", 21, "D_plugin",55.0, 76.0, 17.55),
 ("06", 23, "A_ide",   22.0, 44.0, 19.34),
 ("06", 24, "D_plugin",26.0, 48.0, 19.03),
 ("07",  4, "B_ask",   22.0, 44.0, 17.56),
 ("07",  5, "C_yolo",   6.0, 26.0, 17.43),
 ("07", 12, "C_yolo",  40.0, 66.0, 22.70),
 ("07", 17, "C_yolo",  70.0,100.0, 27.37),
]

def dur(p):
    return float(subprocess.run(["ffprobe","-v","error","-show_entries",
        "format=duration","-of","csv=p=0",p],capture_output=True,text=True).stdout.strip())

for lec, idx, take, ss, to, win in JOBS:
    src = f"{T}/{take}.mp4"
    out = f"{R}/renderer/public/nocode{lec}/shots/{idx:02d}.mp4"
    os.makedirs(os.path.dirname(out), exist_ok=True)
    rate = win / (to - ss)
    cmd = ["ffmpeg","-y","-loglevel","error","-ss",str(ss),"-to",str(to),"-i",src,
           "-an","-vf",f"setpts=PTS*{rate:.6f},scale=1920:1080:flags=lanczos,fps=30",
           "-t",f"{win:.3f}","-c:v","libx264","-preset","medium","-crf","18",
           "-pix_fmt","yuv420p","-g","30",out]
    r = subprocess.run(cmd, capture_output=True, text=True)
    if r.returncode:
        print(f"FAIL {lec}/{idx}: {r.stderr[-300:]}"); continue
    print(f"  nocode{lec}/shots/{idx:02d}.mp4  {dur(out):5.2f}s  (want {win:.2f})  "
          f"{take} {ss:.0f}-{to:.0f}s @ {rate:.2f}x")
