#!/usr/bin/env python3
"""Week 3 capture rig — one long continuous session, many cuts.

Week 3 is not shot lecture by lecture. Each day is ONE continuous capture and
every lecture is cut out of it afterwards. That only works if we can find the
boundaries again, so this wraps stage.SegRec with a marker log.

Call mark("w3-03/open") whenever a beat starts. On stop() we write a manifest
that maps every marker to a segment and an offset inside it, which cut.py turns
into per-lecture source clips.

Segment start times are derived BACKWARDS from the kill time and the probed
duration, not from the spawn time: ffmpeg takes an unpredictable moment to
open the capture device, so spawn-time offsets drift by a second or more and a
second is a visible mis-cut.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys
import time

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S  # noqa: E402

HERE = os.path.dirname(os.path.abspath(__file__))
OUT = os.path.join(HERE, "captures")
os.makedirs(OUT, exist_ok=True)


def _dur(path: str) -> float:
    r = subprocess.run(["ffprobe", "-v", "error", "-show_entries", "format=duration",
                        "-of", "csv=p=0", path], capture_output=True, text=True).stdout.strip()
    try:
        return float(r)
    except ValueError:
        return 0.0


# The day rig records at 40 Mbit, which is right for a five-minute take and
# wrong for a three-hour one: it wants 54 GB and there are 57 free. 30 Mbit at
# 3840x2160 is still far above what UI text needs (g=15 keeps every half-second
# an I-frame), and it brings a two-hour session down to about 27 GB.
BITRATE = "30M"


class _SegRec(S.SegRec):
    """SegRec that records at Week 3 bitrate and remembers when each segment stopped."""

    def __init__(self, *a, **kw):
        super().__init__(*a, **kw)
        self.seg_ends: dict[str, float] = {}

    def _spawn(self):
        p = self._seg_path()
        self.p = subprocess.Popen([
            "ffmpeg", "-y", "-loglevel", "error",
            "-f", "avfoundation", "-capture_cursor", "1",
            "-framerate", "30", "-i", S.CAP_INDEX,
            "-c:v", "h264_videotoolbox", "-b:v", BITRATE,
            "-g", "15", "-keyint_min", "15",
            # Fragmented output, so a killed capture is still playable. A plain
            # MP4 writes its index at the very end: kill ffmpeg and the whole
            # file is unreadable — which is how 11 minutes of finished footage
            # was lost. With these flags the file is valid at every moment.
            "-movflags", "+frag_keyframe+empty_moov+default_base_moof",
            "-pix_fmt", "yuv420p", p,
        ], stdin=subprocess.PIPE, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        self.segs.append(p)

    def _kill(self):
        p = self.p.args[-1] if self.p else None
        super()._kill()
        if p:
            self.seg_ends[p] = time.time()


class Session:
    """One capture day.

        s = Session("A", allow=("Code", "Claude"))
        s.start()
        s.mark("w3-02/repo")
        ...
        s.stop()
    """

    def __init__(self, name: str, allow=("Code",), poll: float = 0.35):
        self.name = name
        self.rec = _SegRec(f"w3{name}", allow=allow, poll=poll)
        self.markers: list[dict] = []
        self.t_start = 0.0
        self.log_path = os.path.join(OUT, f"capture-{name}-markers.jsonl")
        self.manifest_path = os.path.join(OUT, f"capture-{name}.json")

    def start(self, settle: float = 2.5, need_gb: float = 30.0):
        import shutil as _sh
        free = _sh.disk_usage("/").free / 1e9
        if free < need_gb:
            raise SystemExit(
                f"only {free:.0f} GB free, this session wants ~{need_gb:.0f} GB. "
                "Cut and purge the previous capture first — footage is the one "
                "thing here that cannot be regenerated, so nothing is deleted "
                "automatically.")
        self.t_start = time.time()
        open(self.log_path, "a").close()
        self.rec.start(settle=settle)
        print(f"[capture {self.name}] rolling — allow={self.rec.allow}", flush=True)

    def mark(self, label: str, note: str = ""):
        """Label the beat that starts NOW.

        Written through to disk immediately: a crash three hours into a capture
        must not cost the index to the footage that survived.
        """
        m = {"t": time.time(), "label": label, "note": note}
        self.markers.append(m)
        with open(self.log_path, "a") as fh:
            fh.write(json.dumps(m) + "\n")
        el = m["t"] - self.t_start
        print(f"  · {el/60:5.1f}m  {label}" + (f"  ({note})" if note else ""), flush=True)

    def stop(self):
        kept = self.rec.stop()
        segs = []
        for p in kept:
            d = _dur(p)
            end = self.rec.seg_ends.get(p, time.time())
            segs.append({"path": p, "dur": d, "t0": end - d, "t1": end})
        segs.sort(key=lambda s: s["t0"])

        # Place every marker on the timeline of the segment that was rolling.
        placed = []
        for m in self.markers:
            hit = None
            for idx, sg in enumerate(segs):
                if sg["t0"] <= m["t"] <= sg["t1"]:
                    hit = {"seg": idx, "offset": round(m["t"] - sg["t0"], 3)}
                    break
            if hit is None:
                # Marker fell in a pause (he was in another app). Snap it to the
                # start of the next segment so the cut still lands somewhere real.
                nxt = next((i for i, sg in enumerate(segs) if sg["t0"] > m["t"]), None)
                hit = {"seg": nxt, "offset": 0.0, "orphan": True} if nxt is not None else \
                      {"seg": None, "offset": None, "orphan": True}
            placed.append({**m, **hit})

        man = {"capture": self.name, "t_start": self.t_start,
               "segments": segs, "markers": placed}
        with open(self.manifest_path, "w") as fh:
            json.dump(man, fh, indent=2)

        total = sum(s["dur"] for s in segs)
        orphans = sum(1 for m in placed if m.get("orphan"))
        print(f"[capture {self.name}] {len(segs)} segment(s), {total/60:.1f} min, "
              f"{len(placed)} marker(s), {orphans} orphan(s)", flush=True)
        print("  manifest:", self.manifest_path, flush=True)
        if orphans:
            print("  !! orphan markers landed in a pause — check before cutting", flush=True)
        return man


if __name__ == "__main__":
    print("import this; see capture-A.md for the shot plan")
