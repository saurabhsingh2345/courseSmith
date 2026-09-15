#!/usr/bin/env python3
"""One row per lecture, six frames across, for gross-defect triage.

Enough resolution to spot an OS dialog, a bright intrusion, a blank frame or a
wasted half-frame; not enough to read terminal text. Used to decide which
lectures earn a full-resolution look.
"""
import subprocess, sys, os
from PIL import Image, ImageDraw, ImageFont
TW, TH, N = 430, 242, 6
F = ImageFont.truetype("/System/Library/Fonts/Supplemental/Arial.ttf", 20)
FS = ImageFont.truetype("/System/Library/Fonts/Supplemental/Arial.ttf", 15)

def row(path, sheet, y):
    d = float(subprocess.run(["ffprobe","-v","error","-show_entries","format=duration",
        "-of","csv=p=0",path],capture_output=True,text=True).stdout.strip())
    b = os.path.basename(path).replace("_adam.mp4","")
    drw = ImageDraw.Draw(sheet)
    for i in range(N):
        t = d*(i+0.5)/N
        o = f"probe/tri_{b}_{i}.png"
        subprocess.run(["ffmpeg","-nostdin","-v","error","-ss",f"{t:.2f}","-i",path,
            "-frames:v","1","-vf",f"scale={TW}:{TH}","-y",o],check=False)
        x = i*TW
        try: sheet.paste(Image.open(o), (x, y))
        except Exception: pass
        os.path.exists(o) and os.remove(o)
        drw.rectangle([x+2,y+2,x+60,y+22], fill=(0,0,0))
        drw.text((x+6,y+4), f"{int(t)//60}:{int(t)%60:02d}", fill=(255,220,0), font=FS)
        drw.rectangle([x,y,x+TW-1,y+TH-1], outline=(70,70,70))
    drw.rectangle([0,y,120,y+26], fill=(200,30,30))
    drw.text((6,y+3), b, fill=(255,255,255), font=F)

vids = sys.argv[2:]
per = 6
for s in range((len(vids)+per-1)//per):
    chunk = vids[s*per:(s+1)*per]
    sheet = Image.new("RGB",(TW*N, TH*len(chunk)),(15,15,15))
    for k,v in enumerate(chunk): row(v, sheet, k*TH)
    p = f"{sys.argv[1]}_{s+1}.png"; sheet.save(p, optimize=True); print(p, flush=True)
