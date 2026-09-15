#!/usr/bin/env python3
"""Record one lecture's slide deck to mp4.

    python3 record_deck.py w3-01

The deck advances when each voice clip ends, so its length is simply the sum of
the clips plus the gap the driver leaves between them. That is measured here
rather than guessed, and a tail is added so the last slide is not cut on the
final syllable.

Chrome is launched with a throwaway profile: his normal window would put tabs,
bookmarks and a profile picture into the frame.
"""

from __future__ import annotations

import glob
import os
import subprocess
import sys
import time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import w3drive as W  # noqa: E402
from w3drive import S  # noqa: E402

PROGRAM = os.path.abspath(os.path.join(HERE, "..", "..", "..", "program"))
OUT = os.path.join(HERE, "decks")
PROFILE = "/Users/Shared/projects/.shoot-chrome"
CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
PORT = 8765
GAP = 0.28          # the driver's pause between slides
TAIL = 1.6
LEAD = 1.8          # ?wait= before the first slide starts


def part_dir(lec: str) -> str:
    hits = glob.glob(os.path.join(PROGRAM, f"part-{lec}-*"))
    if not hits:
        sys.exit(f"no part folder for {lec}")
    return hits[0]


def clip_len(p: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def main() -> None:
    if len(sys.argv) < 2:
        sys.exit(__doc__)
    lec = sys.argv[1]
    d = part_dir(lec)
    vo = sorted(glob.glob(os.path.join(d, "vo", "[0-9][0-9].mp3")))
    if not vo:
        sys.exit(f"no voice clips in {d}/vo — run generate_vo.py first")

    # The deck plays the voice at 0.9x, the house rate, so every clip is longer
    # on screen than on disk.
    length = sum(clip_len(p) / 0.9 for p in vo) + GAP * (len(vo) - 1) + TAIL
    print(f"{lec}: {len(vo)} slides, deck runs {length:.1f}s")

    os.makedirs(OUT, exist_ok=True)
    srv = subprocess.Popen([sys.executable, "-m", "http.server", str(PORT)],
                           cwd=PROGRAM, stdout=subprocess.DEVNULL,
                           stderr=subprocess.DEVNULL)
    time.sleep(1.5)
    url = (f"http://127.0.0.1:{PORT}/{os.path.basename(d)}/slides.html"
           f"?record=1&wait={int(LEAD*1000)}")
    # --app gives a window with no tab strip, address bar or bookmarks, and
    # --window-size fills the panel WITHOUT going fullscreen. Fullscreen would put
    # Chrome on its own Space, which is what broke the first VS Code take.
    ch = subprocess.Popen([CHROME, f"--user-data-dir={PROFILE}", "--no-first-run",
                           "--no-default-browser-check",
                           "--autoplay-policy=no-user-gesture-required",
                           "--window-position=0,0", "--window-size=1920,1080",
                           f"--app={url}"],
                          stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    time.sleep(3.0)

    raw = os.path.join(OUT, f"{lec}_raw.mp4")
    rec = subprocess.Popen([
        "ffmpeg", "-y", "-loglevel", "error", "-f", "avfoundation",
        "-capture_cursor", "0", "-framerate", "30", "-i", S.CAP_INDEX,
        "-c:v", "h264_videotoolbox", "-b:v", "30M", "-g", "15",
        "-keyint_min", "15", "-pix_fmt", "yuv420p", raw],
        stdin=subprocess.PIPE, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

    time.sleep(length + LEAD + 2.0)
    try:
        rec.stdin.write(b"q"); rec.stdin.flush(); rec.wait(timeout=15)
    except Exception:
        rec.kill()
    ch.terminate()
    srv.terminate()
    time.sleep(1.0)

    # Trim the lead-in; the deck's own audio is discarded because the narration
    # track is rebuilt cleanly at assembly time.
    dst = os.path.join(OUT, f"{lec}.mp4")
    subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-ss", f"{LEAD + 2.6}",
                    "-i", raw, "-an", "-c:v", "h264_videotoolbox", "-b:v", "26M",
                    "-g", "15", "-keyint_min", "15", "-pix_fmt", "yuv420p",
                    dst], check=True)
    os.remove(raw)
    print(f"  -> {dst}  {clip_len(dst):.1f}s")


if __name__ == "__main__":
    main()
