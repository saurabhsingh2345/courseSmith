#!/usr/bin/env python3
"""One continuous recording plus a log of exactly when each beat began.

The narration for these films is written first and the footage is bent to fit it
(assemble.py), so every beat needs an `anchor`: the second in the finished clip
where the picture that sentence describes appears. Reading those off a contact
sheet is guesswork that has cost a whole voice cycle before now, so they are
logged as the shoot happens instead.

Start times are derived BACKWARDS from the kill time and the probed duration:
ffmpeg takes an unpredictable moment to open the capture device, so an offset
measured from the spawn call drifts by a second or more, and a second is a
visible mis-cut.
"""
from __future__ import annotations
import json, os, subprocess, sys, time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S  # noqa: E402  (probes the 16:9 capture device at import)

TAKES = os.path.join(HERE, "takes")
os.makedirs(TAKES, exist_ok=True)


def dur(p):
    r = subprocess.run(["ffprobe", "-v", "error", "-show_entries",
                        "format=duration", "-of", "csv=p=0", p],
                       capture_output=True, text=True).stdout.strip()
    return float(r) if r else 0.0


class Take:
    def __init__(self, name, allow):
        self.name, self.allow = name, tuple(allow)
        self.path = os.path.join(TAKES, f"{name}.mp4")
        self.marks = []
        self.rec = S.Rec(name, guard=self.allow)
        self.rec.path = self.path

    def start(self, settle=2.5):
        self.rec.start(settle=settle)
        print(f"[take {self.name}] rolling", flush=True)

    def mark(self, beat):
        self.marks.append([beat, time.time()])
        print(f"  . {beat} @ {time.strftime('%H:%M:%S')}", flush=True)

    def hold(self, secs, beat=None):
        """Sit still on the current picture. Footage is what the words need."""
        if beat:
            self.mark(beat)
        time.sleep(secs)

    def stop(self):
        ok = self.rec.stop()
        kill = time.time()
        d = dur(self.path)
        t0 = kill - d
        out = {"take": self.name, "ok": bool(ok), "duration": d,
               "anchors": {b: round(max(t - t0, 0.0), 2) for b, t in self.marks}}
        json.dump(out, open(os.path.join(TAKES, f"{self.name}.json"), "w"), indent=1)
        print(f"[take {self.name}] {d:.1f}s  " +
              "  ".join(f"{b}={a}" for b, a in out["anchors"].items()), flush=True)
        return ok
