#!/usr/bin/env python3
"""OCR every cut looking for the things that must never be in a frame.

    python3 scanpii.py                 # every cut, a frame every 8s
    python3 scanpii.py w3-41 --every 3

The house rule is that no username, machine name, company or personal account
appears in any rendered frame. Two known ways it gets in:

  * After Claude Code exits, the shell prompt reverts to `<user>@<machine>` for
    a second or two. `tails.py` catches that at the END of a cut; this catches
    it anywhere.
  * `gh` prints the account name in issue and PR headers.

Reports `lecture  time  the matching line`, which is exactly what redact.py
needs as a --box time window.
"""
from __future__ import annotations
import glob, os, re, subprocess, sys, tempfile

HERE = os.path.dirname(os.path.abspath(__file__))
CUTS = os.path.join(HERE, "cuts")

NEEDLES = [r"enfecsolutions", r"enfecs", r"macbook", r"saurabh", r"enfec\b",
           r"\bsingh\b", r"/Users/enfec"]
RX = re.compile("|".join(NEEDLES), re.I)


def dur(p):
    return float(subprocess.run(["ffprobe","-v","error","-show_entries",
        "format=duration","-of","csv=p=0",p],capture_output=True,text=True).stdout.strip())


def main():
    every = 8.0
    if "--every" in sys.argv:
        every = float(sys.argv[sys.argv.index("--every") + 1])
    want = [a for a in sys.argv[1:] if a.startswith("w3-")]
    cuts = ([os.path.join(CUTS, f"{l}.mp4") for l in want] if want
            else sorted(glob.glob(os.path.join(CUTS, "w3-*.mp4"))))
    cuts = [c for c in cuts if ".orig." not in c]

    hits = 0
    with tempfile.TemporaryDirectory() as tmp:
        jpg = os.path.join(tmp, "f.jpg")
        txt = os.path.join(tmp, "o")
        for c in cuts:
            lec = os.path.basename(c)[:-4]
            total = dur(c)
            t, found = 2.0, 0
            while t < total:
                subprocess.run(["ffmpeg","-y","-loglevel","error","-ss",f"{t}",
                                "-i",c,"-frames:v","1","-vf","scale=1600:-1",
                                "-q:v","3",jpg], check=False)
                if os.path.exists(jpg):
                    subprocess.run(["tesseract",jpg,txt,"--psm","6"],
                                   capture_output=True)
                    try:
                        body = open(txt + ".txt", errors="ignore").read()
                    except FileNotFoundError:
                        body = ""
                    for line in body.splitlines():
                        if RX.search(line):
                            print(f"{lec}  {int(t//60):02d}:{int(t%60):02d}  "
                                  f"{line.strip()[:100]}", flush=True)
                            found += 1; hits += 1
                            break
                t += every
            print(f"-- {lec}: {found} hit(s) over {total/60:.1f} min", flush=True)
    print(f"\nTOTAL {hits} frame(s) with something that must not ship")


if __name__ == "__main__":
    main()
