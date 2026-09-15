#!/usr/bin/env python3
"""Cut nocode01 to his lecture-one shape: thirteen segments, 613 seconds.

Every screen segment is retimed so its picture length equals its narration
length exactly, so the cut lands on his beat boundaries rather than drifting.
Waiting is ramped (segment M runs a little over five times), never padded.
"""
import json, os, subprocess, sys

W = os.path.dirname(os.path.abspath(__file__))
T = os.path.join(W, "takes")
OUT = os.path.join(W, "cut")
os.makedirs(OUT, exist_ok=True)

def run(cmd):
    r = subprocess.run(cmd, capture_output=True, text=True)
    if r.returncode != 0:
        sys.stderr.write(" ".join(cmd) + "\n" + r.stderr[-2500:] + "\n")
        raise SystemExit(1)
    return r.stdout.strip()

def dur(p):
    return float(run(["ffprobe", "-v", "error", "-show_entries", "format=duration",
                      "-of", "csv=p=0", p]))

# name, source take, in, out, target seconds, optional crop
SEGS = [
    ("A", "A_slides.mp4",   None, None, 93.994, None),          # rendered, has audio
    ("B", "BC_browser.mp4",  1.2,  27.0, 26.177, None),
    ("C", "BC_browser.mp4", 27.0,  70.0, 40.004, None),
    ("D", "InsertD.mp4",    None, None, 30.900, None),          # rendered, has audio
    ("E", "E_drag.mp4",      4.0,  12.0,  7.364, "crop=2120:1192:860:480"),
    ("F", "InsertF.mp4",    None, None, 60.400, None),          # rendered, has audio
    ("G", "GH_cursor.mp4",   6.0,  37.4, 29.071, None),
    ("H1", "GH_cursor.mp4", 37.4,  54.9, 14.500, None),
    ("H2", "H2_folder.mp4",  0.5,  40.5, 35.732, None),
    ("I", "InsertI.mp4",    None, None, 43.674, None),          # rendered, has audio
    ("J", "JKLM_build.mp4",  1.5, 140.0, 122.213, None),
    ("K", "JKLM_build.mp4", 140.0, 174.0, 32.193, None),
    ("L", "JKLM_build.mp4", 174.0, 194.0, 18.864, None),
    ("M", "JKLM_build.mp4", 194.0, 486.5, 57.054, None),
]

RENDERED = {"A", "D", "F", "I"}

def build_video():
    parts = []
    for name, src, ss, to, target, crop in SEGS:
        p = os.path.join(T, src)
        outp = os.path.join(OUT, f"v_{name}.mp4")
        vf = []
        if crop:
            vf.append(crop)
        if name not in RENDERED:
            srclen = (to - ss)
            vf.append(f"setpts=PTS*{target / srclen:.6f}")
        vf.append("scale=1920:1080:flags=lanczos")
        vf.append("fps=30")
        cmd = ["ffmpeg", "-y", "-loglevel", "error"]
        if ss is not None:
            cmd += ["-ss", str(ss), "-to", str(to)]
        cmd += ["-i", p, "-an", "-vf", ",".join(vf),
                "-t", f"{target:.3f}",
                "-c:v", "libx264", "-preset", "medium", "-crf", "17",
                "-pix_fmt", "yuv420p", "-g", "60", outp]
        run(cmd)
        d = dur(outp)
        print(f"  video {name:2s} target {target:7.3f}  got {d:7.3f}", flush=True)
        parts.append(outp)
    return parts

VO_FOR = {
    "B": "B_browser.mp3", "C": "C_download.mp3", "E": "E_drag.mp3",
    "G": "G_welcome.mp3", "H1": None, "H2": None,
    "J": "J_panes.mp3", "K": "K_model.mp3", "L": "L_prompt.mp3", "M": "M_build.mp3",
}

def build_audio():
    """One narration piece per segment, padded to that segment's exact length."""
    parts = []
    pending_h = None
    for name, src, ss, to, target, crop in SEGS:
        outp = os.path.join(OUT, f"a_{name}.wav")
        if name in RENDERED:
            run(["ffmpeg", "-y", "-loglevel", "error", "-i", os.path.join(T, src),
                 "-vn", "-ac", "1", "-ar", "48000",
                 "-af", f"apad,atrim=0:{target:.3f}", outp])
        elif name in ("H1", "H2"):
            # H is one line of narration across two takes; lay it under H1 and
            # let the tail run into H2.
            if name == "H1":
                pending_h = target
                run(["ffmpeg", "-y", "-loglevel", "error",
                     "-i", os.path.join(W, "vo_screen", "H_folder.mp3"),
                     "-ac", "1", "-ar", "48000",
                     "-af", f"apad,atrim=0:{target:.3f}", outp])
            else:
                run(["ffmpeg", "-y", "-loglevel", "error",
                     "-i", os.path.join(W, "vo_screen", "H_folder.mp3"),
                     "-ac", "1", "-ar", "48000",
                     "-af", f"atrim=start={pending_h:.3f},asetpts=N/SR/TB,"
                            f"apad,atrim=0:{target:.3f}", outp])
        else:
            run(["ffmpeg", "-y", "-loglevel", "error",
                 "-i", os.path.join(W, "vo_screen", VO_FOR[name]),
                 "-ac", "1", "-ar", "48000",
                 "-af", f"apad,atrim=0:{target:.3f}", outp])
        print(f"  audio {name:2s} {dur(outp):7.3f}", flush=True)
        parts.append(outp)
    return parts

if __name__ == "__main__":
    print("video:")
    v = build_video()
    print("audio:")
    a = build_audio()

    with open(os.path.join(OUT, "v.txt"), "w") as fh:
        for p in v:
            fh.write(f"file '{p}'\n")
    with open(os.path.join(OUT, "a.txt"), "w") as fh:
        for p in a:
            fh.write(f"file '{p}'\n")

    run(["ffmpeg", "-y", "-loglevel", "error", "-f", "concat", "-safe", "0",
         "-i", os.path.join(OUT, "v.txt"), "-c", "copy", os.path.join(OUT, "video.mp4")])
    run(["ffmpeg", "-y", "-loglevel", "error", "-f", "concat", "-safe", "0",
         "-i", os.path.join(OUT, "a.txt"), "-c", "copy", os.path.join(OUT, "audio.wav")])

    print("video", dur(os.path.join(OUT, "video.mp4")))
    print("audio", dur(os.path.join(OUT, "audio.wav")))

    run(["ffmpeg", "-y", "-loglevel", "error",
         "-i", os.path.join(OUT, "video.mp4"),
         "-i", os.path.join(OUT, "audio.wav"),
         "-map", "0:v:0", "-map", "1:a:0",
         "-c:v", "copy",
         "-af", "loudnorm=I=-16:TP=-1.5:LRA=11",
         "-c:a", "aac", "-b:a", "192k",
         os.path.join(OUT, "nocode01.mp4")])
    final = os.path.join(OUT, "nocode01.mp4")
    print("FINAL", dur(final), os.path.getsize(final) // 1024 // 1024, "MB")
