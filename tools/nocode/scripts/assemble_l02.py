#!/usr/bin/env python3
"""Cut nocode02 to his lecture-two shape: 6:54, all screen recording."""
import os, subprocess, sys
W = os.path.dirname(os.path.abspath(__file__))
T, OUT = os.path.join(W, "takes"), os.path.join(W, "cut2")
os.makedirs(OUT, exist_ok=True)

def run(c):
    r = subprocess.run(c, capture_output=True, text=True)
    if r.returncode: sys.stderr.write(" ".join(c)+"\n"+r.stderr[-1500:]+"\n"); raise SystemExit(1)
    return r.stdout.strip()
def dur(p):
    return float(run(["ffprobe","-v","error","-show_entries","format=duration","-of","csv=p=0",p]))

# name, take, in, out, target, vo
S = [
 ("N01","L2A_files.mp4",   1.5,  24.0, 20.0, "N01_done.mp3"),
 ("N02","L2A_files.mp4",  24.0,  50.0, 18.0, "N02_files.mp3"),
 ("N03","L2B_serve.mp4",   2.0,  24.0, 17.0, "N03_run.mp3"),
 ("N04","L2C_play.mp4",    1.5,  18.0, 14.0, "N04_title.mp3"),
 ("N05","L2C_play.mp4",   26.0,  56.0, 26.0, "N05_play1.mp3"),
 ("N06","L2C_play.mp4",   56.0, 100.0, 30.0, "N06_play2.mp3"),
 ("N07","L2D_iter1.mp4",   2.0,  42.0, 36.0, "N07_iter1.mp3"),
 ("N08","L2D_iter1.mp4",  42.0, 120.0, 22.0, "N08_edit1.mp3"),
 ("N09","L2E_test1.mp4",   8.0,  40.0, 22.0, "N09_test1.mp3"),
 ("N10","L2F_iter2.mp4",   2.0,  40.0, 34.0, "N10_iter2.mp3"),
 ("N11","L2G_test2.mp4",  10.0,  38.0, 22.0, "N11_test2.mp3"),
 ("N12","L2H_wrap.mp4",    1.5,  58.0, 54.0, "N12_wrap.mp3"),
 ("N13","L2F_iter2.mp4",  45.0, 228.0, 40.0, "N13_close.mp3"),
 ("N14","L2G_test2.mp4",  16.0,  52.0, 34.0, "N14_close.mp3"),
 ("N15","L2C_play.mp4",   72.0, 104.0, 25.0, "N15_close.mp3"),
]

print("video:")
vids, auds = [], []
for name, take, ss, to, tgt, vo in S:
    o = os.path.join(OUT, f"v_{name}.mp4")
    rate = tgt / (to - ss)
    run(["ffmpeg","-y","-loglevel","error","-ss",str(ss),"-to",str(to),
         "-i",os.path.join(T,take),"-an",
         "-vf",f"setpts=PTS*{rate:.6f},scale=1920:1080:flags=lanczos,fps=30",
         "-t",f"{tgt:.3f}","-c:v","libx264","-preset","medium","-crf","17",
         "-pix_fmt","yuv420p","-g","60",o])
    print(f"  {name} src {to-ss:6.1f}s -> {dur(o):6.2f}s (x{1/rate:.2f})", flush=True)
    vids.append(o)
    a = os.path.join(OUT, f"a_{name}.wav")
    run(["ffmpeg","-y","-loglevel","error","-i",os.path.join(W,"vo_l2",vo),
         "-ac","1","-ar","48000","-af",f"adelay=250|250,apad,atrim=0:{tgt:.3f}",a])
    auds.append(a)

for lst, parts, out, args in (("v.txt", vids, "video.mp4", ["-c","copy"]),
                              ("a.txt", auds, "audio.wav", ["-c","copy"])):
    with open(os.path.join(OUT,lst),"w") as fh:
        for p in parts: fh.write(f"file '{p}'\n")
    run(["ffmpeg","-y","-loglevel","error","-f","concat","-safe","0",
         "-i",os.path.join(OUT,lst)]+args+[os.path.join(OUT,out)])

v, a = dur(os.path.join(OUT,"video.mp4")), dur(os.path.join(OUT,"audio.wav"))
print(f"\nvideo {v:.2f}s  audio {a:.2f}s  (his 414)")
run(["ffmpeg","-y","-loglevel","error","-i",os.path.join(OUT,"video.mp4"),
     "-i",os.path.join(OUT,"audio.wav"),"-map","0:v:0","-map","1:a:0","-c:v","copy",
     "-af","alimiter=limit=0.955:level=false,loudnorm=I=-16:TP=-1.5:LRA=11,aresample=48000",
     "-c:a","aac","-b:a","192k","-movflags","+faststart",
     os.path.join(OUT,"nocode02.mp4")])
print("FINAL", dur(os.path.join(OUT,"nocode02.mp4")))
