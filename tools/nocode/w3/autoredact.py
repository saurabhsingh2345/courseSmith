#!/usr/bin/env python3
"""Find every frame with a name, machine or company in it and blur exactly it.

    python3 autoredact.py --plan            # every cut, report boxes, change nothing
    python3 autoredact.py w3-07 w3-13       # blur these
    python3 autoredact.py                   # blur every cut that needs it

`scanpii.py` says a lecture has a problem and roughly when. That is enough to
know a take is unusable but not enough to fix it, and re-shooting an eight
minute lecture because a shell echoed an absolute path once is not a trade
anybody should make. So this does the other half: OCR with word boxes, turn the
matching words into pixel rectangles, merge the rectangles that belong to the
same moment on screen, and hand them to `redact.py`.

Two things it does deliberately rather than tightly:

* **A hit is widened to the whole line, and to the neighbouring line height.**
  A terminal scrolls between samples; a box that fits the word at t is off the
  word at t+0.4. The band is the safe unit, and a blurred band in a moving
  terminal reads as motion rather than as a redaction.
* **A window is padded either side of a hit**, because the scan sees one frame
  a second and the text was on screen before and after that frame.

**Speed is the whole design.** Seeking a 3840x2160 cut once per sample and
running tesseract serially costs about six seconds a frame, which is three
hours for this week. Instead each pass decodes the cut ONCE with an `fps`
filter, writing every sample it will ever need as a jpg, and then runs tesseract
over them `--jobs` at a time. Same frames, same OCR, about twenty times faster.

Coarse-to-fine, because OCR is still the expensive part: sample the whole cut at
`--coarse` (8s), then re-sample only the neighbourhoods of the coarse hits at
`--fine` (1s). `redact.py` is called with `--onto-current`, so each sweep blurs what it can
still read on top of what is already blurred. Rebuilding from `.orig.mp4` would
drop the previous sweep's boxes and the hand-tuned ones in `redact-w3.sh`.
"""

from __future__ import annotations

import csv
import glob
import io
import os
import re
import subprocess
import sys
import tempfile

HERE = os.path.dirname(os.path.abspath(__file__))
CUTS = os.path.join(HERE, "cuts")

# The same list scanpii.py uses, plus the org that must never be in frame.
NEEDLES = [r"enfecsolutions", r"enfecs", r"macbook", r"saurabh", r"enfec",
           r"singh", r"saurabhsingh2345"]
RX = re.compile("|".join(NEEDLES), re.I)

OCR_W = 1600          # OCR is done on a 1600-wide frame
FULL_W = 3840         # cuts are 3840x2160
SCALE = FULL_W / OCR_W

COARSE = 4.0          # 8 missed a line that was on screen for six seconds
FINE = 1.0
PAD_T = 1.2           # seconds of window padding either side of a fine hit
LINE_PAD = 1.4        # multiples of the word's own height, above and below
JOBS = 10


def idx(name: str) -> int:
    """f000123.jpg -> 122. ffmpeg numbers from 1; the sample list from 0."""
    return int(re.search(r"(\d+)", name).group(1)) - 1


def dur(p: str) -> float:
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def sample(cut: str, out: str, every: float, t0: float, t1: float) -> list[float]:
    """Decode once; write a jpg every `every` seconds between t0 and t1.

    Returns the timestamp of each written frame. `fps=1/every` puts the first
    sample half an interval in, which is what the returned times reflect.
    """
    os.makedirs(out, exist_ok=True)
    for f in os.listdir(out):
        os.remove(os.path.join(out, f))
    subprocess.run(
        ["ffmpeg", "-y", "-loglevel", "error", "-ss", f"{t0}", "-i", cut,
         "-t", f"{max(0.1, t1 - t0)}",
         "-vf", f"fps=1/{every},scale={OCR_W}:-1", "-q:v", "3",
         os.path.join(out, "f%06d.jpg")], check=False)
    n = len(os.listdir(out))
    return [t0 + i * every for i in range(n)]


def ocr_all(out: str) -> dict[str, str]:
    """Run tesseract over every jpg in `out`, JOBS at a time. Returns tsv text."""
    names = sorted(f for f in os.listdir(out) if f.endswith(".jpg"))
    if not names:
        return {}
    script = "\n".join(
        f'tesseract "{os.path.join(out, n)}" "{os.path.join(out, n[:-4])}" '
        f"--psm 6 tsv 2>/dev/null" for n in names)
    subprocess.run(["bash", "-c",
                    f"cat <<'EOF' | xargs -P {JOBS} -I CMD bash -c CMD\n"
                    + script + "\nEOF"], capture_output=True)
    res = {}
    for n in names:
        p = os.path.join(out, n[:-4] + ".tsv")
        if os.path.exists(p):
            res[n] = open(p, errors="ignore").read()
    return res


def bands(tsv: str) -> list[tuple[int, int, int, int]]:
    """Bands to blur on one frame, in FULL-resolution pixels."""
    raw = []
    for row in csv.DictReader(io.StringIO(tsv), delimiter="\t",
                              quoting=csv.QUOTE_NONE):
        txt = (row.get("text") or "").strip()
        if not txt or not RX.search(txt):
            continue
        try:
            y, h = int(row["top"]), int(row["height"])
        except (KeyError, ValueError):
            continue
        pad = max(int(h * LINE_PAD), 8)
        # The whole line, not the word: a scrolling terminal moves it sideways
        # too, and a band is what stays over the text between two samples.
        raw.append([0, max(0, y - pad), OCR_W, h + 2 * pad])
    raw.sort(key=lambda b: b[1])
    merged: list[list[int]] = []
    for x, y, w, h in raw:
        if merged and y <= merged[-1][1] + merged[-1][3]:
            top = min(merged[-1][1], y)
            bot = max(merged[-1][1] + merged[-1][3], y + h)
            merged[-1][1], merged[-1][3] = top, bot - top
        else:
            merged.append([x, y, w, h])
    return [(int(x * SCALE), int(y * SCALE), int(w * SCALE), int(h * SCALE))
            for x, y, w, h in merged]


def windows(cut: str, tmp: str, coarse: float, fine: float) -> list[tuple]:
    """(t0, t1, x, y, w, h) boxes covering everything that must not ship."""
    total = dur(cut)

    # Coarse pass over the whole cut: which neighbourhoods are worth a look.
    times = sample(cut, tmp, coarse, 0.0, total)
    tsvs = ocr_all(tmp)
    # Index off the FILENAME, never off position in the dict: a frame whose OCR
    # produced no tsv is missing from it, and every later time would then be
    # attributed to the wrong frame.
    rough = [times[idx(n)] for n in sorted(tsvs)
             if idx(n) < len(times) and RX.search(tsvs[n])]
    if not rough:
        return []

    spans: list[list[float]] = []
    for r in rough:
        lo, hi = max(0.0, r - coarse), min(total, r + coarse)
        if spans and lo <= spans[-1][1]:
            spans[-1][1] = hi
        else:
            spans.append([lo, hi])

    # Fine pass, only over those spans.
    found: list[tuple[float, tuple]] = []
    for lo, hi in spans:
        times = sample(cut, tmp, fine, lo, hi)
        tsvs = ocr_all(tmp)
        for n in sorted(tsvs):
            i = idx(n)
            if i >= len(times):
                continue
            for b in bands(tsvs[n]):
                found.append((times[i], b))
    if not found:
        return []

    # Group consecutive samples whose bands overlap into one timed box.
    out: list[list] = []
    for t, (x, y, w, h) in sorted(found, key=lambda f: f[0]):
        placed = False
        for o in out:
            same_band = not (y + h < o[3] or y > o[3] + o[5])
            if same_band and t - o[1] <= fine * 2.5:
                o[1] = t
                top, bot = min(o[3], y), max(o[3] + o[5], y + h)
                o[3], o[5] = top, bot - top
                placed = True
                break
        if not placed:
            out.append([t, t, x, y, w, h])
    return [(round(max(0.0, t0 - PAD_T), 2), round(min(total, t1 + PAD_T), 2),
             x, y, w, h) for t0, t1, x, y, w, h in out]


def main() -> None:
    plan = "--plan" in sys.argv
    coarse, fine = COARSE, FINE
    if "--coarse" in sys.argv:
        coarse = float(sys.argv[sys.argv.index("--coarse") + 1])
    if "--fine" in sys.argv:
        fine = float(sys.argv[sys.argv.index("--fine") + 1])
    want = [a for a in sys.argv[1:] if a.startswith("w3-")]
    cuts = ([os.path.join(CUTS, f"{l}.mp4") for l in want] if want
            else sorted(glob.glob(os.path.join(CUTS, "w3-*.mp4"))))
    cuts = [c for c in cuts if ".orig." not in c]

    total_boxes = 0
    with tempfile.TemporaryDirectory() as tmp:
        for c in cuts:
            lec = os.path.basename(c)[:-4]
            boxes = windows(c, tmp, coarse, fine)
            if not boxes:
                print(f"{lec}: clean", flush=True)
                continue
            total_boxes += len(boxes)
            secs = sum(t1 - t0 for t0, t1, *_ in boxes)
            print(f"{lec}: {len(boxes)} box(es), {secs:.1f}s of blur",
                  flush=True)
            for t0, t1, x, y, w, h in boxes:
                print(f"    {t0:8.2f} {t1:8.2f}  {x:5d} {y:5d} {w:5d} {h:4d}",
                      flush=True)
            if plan:
                continue
            argv = ["python3", os.path.join(HERE, "redact.py"), lec,
                    "--onto-current"]
            for t0, t1, x, y, w, h in boxes:
                argv += ["--box", str(t0), str(t1), str(x), str(y),
                         str(w), str(h)]
            subprocess.run(argv, check=True)
    print(f"\n{total_boxes} box(es){' planned' if plan else ' applied'}")


if __name__ == "__main__":
    main()
