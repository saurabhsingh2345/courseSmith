#!/usr/bin/env python3
"""Lay the Liam narration over an install take and deliver the film.

    python3 assemble.py vscode-install vscode

Same idea as the week-3 assembler: the narration is spoken at its natural pace
and the footage between two beat anchors is sped up to fit the words, so picture
and voice change together. Anchors come from the take's own marker log, not from
reading a contact sheet.

Voice is ElevenLabs Liam, not the program's Adam, so the content-addressed clip
store is keyed on the voice id as well as the text — otherwise identical
sentences would come back in Adam's voice from the shared cache.
"""
from __future__ import annotations
import hashlib, json, os, subprocess, sys
import numpy as np

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import adam_lay as A  # noqa: E402

LIAM = "TX3LPaxmHKxFdv7VOQHJ"
A.VOICE = LIAM
A.STORE = os.path.join(HERE, "_voice_liam")
os.makedirs(A.STORE, exist_ok=True)
A._store_path = lambda text: os.path.join(
    A.STORE, hashlib.sha1((LIAM + "|" + text).encode("utf-8")).hexdigest() + ".mp3")

OUT = os.path.abspath(os.path.join(HERE, "..", "..", "..", "videos", "install"))
WORK = os.path.join(HERE, "_work")
SR = 48000
GAP = 0.40        # between sentences inside a beat
BRIDGE = 0.55     # between beats
MIN_SPEED = 0.92
MAX_SPEED = 8.0
SOFT_SPEED = 1.35   # past this the picture visibly outruns the words, so the
                    # tail of the beat is dropped instead of ramping harder
W, H = 2560, 1440   # a 1.5x downscale of the 4K panel: sharper than 1080p


def dur(p):
    return float(subprocess.run(["ffprobe", "-v", "error", "-show_entries",
                                 "format=duration", "-of", "csv=p=0", p],
                                capture_output=True, text=True).stdout.strip())


def speak_beat(text, tag, raw):
    parts, sents = [], A.sentences(text)
    for i, s in enumerate(sents):
        p = os.path.join(raw, f"{tag}_{i:02d}.mp3")
        A.speak(s, p)
        parts.append(A.load(p))
        if i < len(sents) - 1:
            parts.append(np.zeros(int(A.SR * GAP), np.float32))
    return np.concatenate(parts) if parts else np.zeros(1, np.float32)


def ramp(src, t0, t1, speed, dst, want):
    speed = max(MIN_SPEED, min(MAX_SPEED, speed))
    have = (t1 - t0) / speed
    vf = f"setpts=PTS/{speed:.4f},scale={W}:{H}:flags=lanczos"
    if want > have + 0.15:                    # hold the last frame, never crawl
        vf += f",tpad=stop_mode=clone:stop_duration={want - have:.3f}"
    subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-ss", f"{t0}", "-to", f"{t1}",
                    "-i", src, "-an", "-vf", vf, "-r", "30",
                    "-c:v", "h264_videotoolbox", "-b:v", "24M", "-g", "15",
                    "-keyint_min", "15", "-pix_fmt", "yuv420p", dst], check=True)


def main():
    film, takename = sys.argv[1], sys.argv[2]
    spec = json.load(open(os.path.join(HERE, "narration", f"{film}.json")))
    tk = json.load(open(os.path.join(HERE, "takes", f"{takename}.json")))
    src = os.path.join(HERE, "takes", f"{takename}.mp4")
    anchors, total = tk["anchors"], tk["duration"]
    segs = spec["segments"]

    work = os.path.join(WORK, film); raw = os.path.join(work, "raw")
    os.makedirs(raw, exist_ok=True)

    # 1. speak it all, at its own pace
    audio, lens = [], []
    for i, sg in enumerate(segs):
        a = speak_beat(sg["say"], f"{film}_{i:02d}", raw)
        audio.append(a); lens.append(len(a) / A.SR)
        print(f"  {sg['beat']:14}{lens[-1]:6.1f}s of voice", flush=True)

    # 2. cut each shot to the words, at 1:1
    #
    #    This used to ramp the footage to fill the beat, up to 1.35x. Playing a
    #    25-second beat 35% fast while the voice runs at natural pace puts the
    #    picture SIX SECONDS ahead of the words by the end of it — which is
    #    exactly what he saw: "it happens on the screen and after 2 3 second the
    #    voice speaks it". So nothing is ever sped up now. Each beat carries an
    #    explicit `in`/`out` window over the take, the window is trimmed to the
    #    length of its narration, and the dead cursor-dwell in between is simply
    #    dropped. Real time equals screen time, so nothing can drift.
    parts = []
    for i, sg in enumerate(segs):
        f0 = float(sg["in"])
        out = float(sg["out"])
        want = lens[i] + (BRIDGE if i + 1 < len(segs) else 0.0)
        f1 = min(out, f0 + want)
        have = f1 - f0
        dst = os.path.join(work, f"v{i:02d}.mp4")
        ramp(src, f0, f1, 1.0, dst, want)
        parts.append(dst)
        note = "held %.1fs" % (want - have) if have < want - 0.15 else ""
        print(f"  {sg['beat']:14}{f0:7.1f}-{f1:6.1f} = {have:5.1f}s  "
              f"words {want:5.1f}s  x1.00 {note}", flush=True)

    # 3. one voice track on the same clock
    track = []
    for i, a in enumerate(audio):
        track.append(a)
        if i < len(audio) - 1:
            track.append(np.zeros(int(A.SR * BRIDGE), np.float32))
    a = np.concatenate(track)
    a = a * (0.92 / (float(np.max(np.abs(a))) or 1.0))
    wav = os.path.join(work, "voice.wav"); A.write_wav(a, wav)

    lst = os.path.join(work, "parts.txt")
    open(lst, "w").write("".join(f"file '{p}'\n" for p in parts))
    silent = os.path.join(work, "video.mp4")
    subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-f", "concat", "-safe", "0",
                    "-i", lst, "-c", "copy", silent], check=True)

    # 4. house level: two-pass loudnorm to -16 LUFS, resample BEFORE the limiter
    #    (loudnorm emits 192k and the AAC encoder's own resampler rings on a
    #    limited signal), and the limiter last with level=disabled so it cannot
    #    auto-level the output back up to full scale.
    meas = subprocess.run(["ffmpeg", "-hide_banner", "-nostats", "-vn", "-i", wav,
                           "-af", "loudnorm=I=-16:TP=-1.5:LRA=11:print_format=json",
                           "-f", "null", "-"], capture_output=True, text=True).stderr
    m = json.loads(meas[meas.rindex("{"):meas.rindex("}") + 1])
    af = ("loudnorm=I=-16:TP=-1.5:LRA=11"
          f":measured_I={m['input_i']}:measured_TP={m['input_tp']}"
          f":measured_LRA={m['input_lra']}:measured_thresh={m['input_thresh']}"
          f":offset={m['target_offset']}:linear=true,"
          f"aresample={SR},alimiter=limit=0.794:level=disabled")

    os.makedirs(OUT, exist_ok=True)
    final = os.path.join(OUT, f"{film}_liam.mp4")
    subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-i", silent, "-i", wav,
                    "-af", af, "-map", "0:v", "-map", "1:a",
                    "-c:v", "copy", "-c:a", "aac", "-b:a", "192k", "-ar", str(SR),
                    "-movflags", "+faststart", final], check=True)

    st = subprocess.run(["ffmpeg", "-hide_banner", "-nostats", "-i", final,
                         "-af", "astats=metadata=1", "-f", "null", "-"],
                        capture_output=True, text=True).stderr
    pk = max(float(l.split(":")[-1]) for l in st.splitlines() if "Peak level dB:" in l)
    print(f"\n{film}: {dur(final)/60:.2f} min  peak {pk:.2f} dBFS  ->  {final}")


if __name__ == "__main__":
    main()
