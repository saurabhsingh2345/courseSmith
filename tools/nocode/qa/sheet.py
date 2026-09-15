#!/usr/bin/env python3
"""Contact sheets for reviewing a lecture cut.

Samples a video at even intervals, tiles the frames 2 wide with a burned-in
timestamp, and writes one or more PNG sheets. Timestamps matter more than the
picture quality here: the whole point is to be able to say "at 4:12 the screen
shows X while the narration says Y".
"""
import os, subprocess, sys, math
from PIL import Image, ImageDraw, ImageFont

FONT = "/System/Library/Fonts/Supplemental/Arial.ttf"
TW, TH = 768, 432          # tile size
COLS, ROWS = 2, 4          # 8 frames per sheet
PER = COLS * ROWS

def dur(path):
    out = subprocess.run(["ffprobe","-v","error","-show_entries","format=duration",
                          "-of","csv=p=0",path],capture_output=True,text=True).stdout.strip()
    return float(out)

def grab(path, t, dest):
    subprocess.run(["ffmpeg","-nostdin","-v","error","-ss",f"{t:.2f}","-i",path,
                    "-frames:v","1","-vf",f"scale={TW}:{TH}:force_original_aspect_ratio=decrease,"
                    f"pad={TW}:{TH}:-1:-1:color=black","-y",dest],check=False)

def mmss(t):
    return f"{int(t)//60}:{int(t)%60:02d}"

def sheets(path, out_prefix, n):
    d = dur(path)
    # keep off the very first and last frame; those are fades on most cuts
    times = [d*(i+0.5)/n for i in range(n)]
    tmp = out_prefix + "_f"
    frames = []
    for i, t in enumerate(times):
        f = f"{tmp}{i:03d}.png"
        grab(path, t, f)
        if os.path.exists(f) and os.path.getsize(f) > 0:
            frames.append((t, f))
    made = []
    font = ImageFont.truetype(FONT, 30)
    for s in range(math.ceil(len(frames)/PER)):
        chunk = frames[s*PER:(s+1)*PER]
        if not chunk: break
        sheet = Image.new("RGB", (TW*COLS, TH*ROWS), (24,24,24))
        drw = ImageDraw.Draw(sheet)
        for k,(t,f) in enumerate(chunk):
            im = Image.open(f).convert("RGB")
            x, y = (k % COLS)*TW, (k // COLS)*TH
            sheet.paste(im, (x,y))
            lab = mmss(t)
            drw.rectangle([x+4,y+4,x+4+14*len(lab)+16,y+46], fill=(0,0,0))
            drw.text((x+12,y+8), lab, fill=(255,220,0), font=font)
            drw.rectangle([x,y,x+TW-1,y+TH-1], outline=(90,90,90), width=2)
        p = f"{out_prefix}_s{s+1}.png"
        sheet.save(p, optimize=True)
        made.append(p)
    for _,f in frames: os.remove(f)
    return made, d

if __name__ == "__main__":
    vid = sys.argv[1]; pre = sys.argv[2]; n = int(sys.argv[3]) if len(sys.argv)>3 else 16
    m, d = sheets(vid, pre, n)
    print(f"{os.path.basename(vid)} {d:.1f}s -> {' '.join(os.path.basename(x) for x in m)}")
