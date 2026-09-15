#!/usr/bin/env python3
"""Speak a lecture with ElevenLabs Adam and lay it into fixed picture windows.

Adam reads faster than the Kokoro tracks these cuts were timed against, so each
line is spoken sentence by sentence and the slack is spread across the pauses
rather than time-stretching him. Squeezing only happens if a line genuinely
overruns its window, and it is reported when it does.

Usage as a library:
    from adam_lay import lay_track
    lay_track(items, outdir, tag)          # items: [{"file","text","target"}]
"""
import hashlib, json, os, shutil, subprocess, sys, time, urllib.request, wave
import numpy as np

W = os.path.dirname(os.path.abspath(__file__))
ENV = os.path.join(os.path.dirname(W), "nc", "No-Code", ".env")
SR = 48000
MODEL = "eleven_multilingual_v2"
LEAD = 0.16
MIN_GAP = 0.16
MAX_GAP = 1.60      # beyond this it reads as dead air; slack goes to the tail


def _env():
    d = {}
    for line in open(ENV):
        if "=" in line and not line.strip().startswith("#"):
            k, v = line.split("=", 1)
            d[k.strip()] = v.strip().strip('"').strip("'")
    return d


E = _env()
KEY, VOICE = E["ELEVENLABS_API_KEY"], E["ELEVENLABS_VOICE_ID"]


def run(cmd):
    r = subprocess.run(cmd, capture_output=True)
    if r.returncode:
        sys.stderr.write(" ".join(cmd) + "\n" + r.stderr.decode()[-1400:] + "\n")
        raise SystemExit(1)
    return r


def dur(p):
    return float(subprocess.run(["ffprobe", "-v", "error", "-show_entries",
                                 "format=duration", "-of", "csv=p=0", p],
                                capture_output=True, text=True).stdout.strip())


STORE = os.path.join(W, "_voice")


def _store_path(text):
    return os.path.join(
        STORE, hashlib.sha1(text.encode("utf-8")).hexdigest() + ".mp3")


def _publish(src, dest, text):
    """Put a clip where the caller asked for it, and file it by content too."""
    if os.path.abspath(src) != os.path.abspath(dest):
        shutil.copyfile(src, dest)
    open(dest + ".txt", "w", encoding="utf-8").write(text)
    keep = _store_path(text)
    if not os.path.exists(keep):
        os.makedirs(STORE, exist_ok=True)
        shutil.copyfile(dest, keep)


def speak(text, dest):
    # Two caches, and the second one is why this is not just a filename check.
    #
    # The sidecar exists because the cache used to be keyed on the FILENAME
    # alone, which meant editing a line after it had been voiced silently
    # shipped the old audio — the picture said one thing and the voice said
    # another, and nothing warned you.
    #
    # The content store exists because the sidecar is not enough either. Clips
    # are numbered `<lecture>_<segment>_<index>`, so inserting one sentence
    # renumbers every clip after it and every one of them is re-spoken at full
    # price for words that did not change. The allowance is metered monthly and
    # a deepening pass across twenty lectures can spend the lot. Filed by a hash
    # of the exact text, a sentence is paid for once, ever — wherever it moves
    # to, and whichever lecture it ends up in.
    side = dest + ".txt"
    if os.path.exists(dest) and os.path.getsize(dest) > 1500:
        try:
            if open(side, encoding="utf-8").read() == text:
                return
        except OSError:
            pass
        os.remove(dest)
    keep = _store_path(text)
    if os.path.exists(keep) and os.path.getsize(keep) > 1500:
        _publish(keep, dest, text)
        return
    body = json.dumps({
        "text": text, "model_id": MODEL,
        "voice_settings": {"stability": 0.45, "similarity_boost": 0.8,
                           "style": 0.0, "use_speaker_boost": True},
    }).encode()
    req = urllib.request.Request(
        f"https://api.elevenlabs.io/v1/text-to-speech/{VOICE}", data=body,
        method="POST", headers={"xi-api-key": KEY,
                                "Content-Type": "application/json",
                                "Accept": "audio/mpeg"})
    # the API resets the connection now and then on a long deck; a lost take
    # is not worth losing the whole run over
    data = None
    for attempt in range(5):
        try:
            with urllib.request.urlopen(req, timeout=120) as r:
                data = r.read()
            break
        except Exception as e:
            if attempt == 4:
                raise
            wait = 2 ** attempt
            sys.stderr.write(f"  retry {attempt + 1}/4 in {wait}s ({e})\n")
            sys.stderr.flush()
            time.sleep(wait)
    if not data or len(data) < 1500:
        raise SystemExit(f"nothing back for {text[:50]!r}")
    open(dest, "wb").write(data)
    _publish(dest, dest, text)


def load(path):
    wav = path + ".wav"
    run(["ffmpeg", "-y", "-loglevel", "error", "-i", path,
         "-ac", "1", "-ar", str(SR), wav])
    with wave.open(wav, "rb") as w:
        a = np.frombuffer(w.readframes(w.getnframes()), dtype="<i2")
    a = a.astype(np.float32) / 32768.0
    idx = np.where(np.abs(a) > 0.004)[0]
    if len(idx):
        a = a[max(0, idx[0] - 600):min(len(a), idx[-1] + 1800)]
    return a


def sentences(text):
    out, cur = [], ""
    for ch in text:
        cur += ch
        if ch in ".!?":
            out.append(cur.strip()); cur = ""
    if cur.strip():
        out.append(cur.strip())
    return [s for s in out if s]


def write_wav(a, path):
    with wave.open(path, "wb") as w:
        w.setnchannels(1); w.setsampwidth(2); w.setframerate(SR)
        w.writeframes((np.clip(a, -1, 1) * 32767).astype("<i2").tobytes())


def lay_one(text, target, tag, raw, fit):
    clips = []
    for i, s in enumerate(sentences(text)):
        p = os.path.join(raw, f"{tag}_{i:02d}.mp3")
        speak(s, p)
        clips.append(load(p))
    speech = sum(len(c) for c in clips) / SR
    ngap = max(1, len(clips) - 1)
    gap = (target - speech - LEAD) / ngap
    if gap > MAX_GAP:
        gap = MAX_GAP          # the picture carries the rest
    squeeze = 1.0
    if gap < MIN_GAP:
        want = target - LEAD - MIN_GAP * ngap
        squeeze = speech / max(0.3, want)
        gap = MIN_GAP
    parts = [np.zeros(int(SR * LEAD), np.float32)]
    for i, c in enumerate(clips):
        parts.append(c)
        if i < len(clips) - 1:
            parts.append(np.zeros(int(SR * gap), np.float32))
    tmp = os.path.join(fit, f"{tag}_pre.wav")
    write_wav(np.concatenate(parts), tmp)
    dest = os.path.join(fit, f"{tag}.wav")
    af = (f"atempo={squeeze:.6f}," if abs(squeeze - 1) > 0.005 else "") \
         + f"apad,atrim=0:{target:.3f}"
    run(["ffmpeg", "-y", "-loglevel", "error", "-i", tmp,
         "-ac", "1", "-ar", str(SR), "-af", af, dest])
    return dest, gap, squeeze, len(clips)


def lay_track(items, outdir, name="adam_track.wav"):
    outdir = os.path.abspath(outdir)
    raw = os.path.join(outdir, "raw"); fit = os.path.join(outdir, "fit")
    os.makedirs(raw, exist_ok=True); os.makedirs(fit, exist_ok=True)
    pieces = []
    for it in items:
        tag = it["file"].replace(".mp3", "")
        d, gap, sq, n = lay_one(it["text"], it["target"], tag, raw, fit)
        flag = "  <-- squeezed" if abs(sq - 1) > 0.005 else ""
        print(f"  {tag:14s} {it['target']:7.2f}s  {n:2d} sent  gap {gap:.2f}s{flag}",
              flush=True)
        pieces.append(d)
    lst = os.path.join(fit, "all.txt")
    with open(lst, "w") as fh:
        for p in pieces:
            fh.write(f"file '{p}'\n")
    track = os.path.join(outdir, name)
    run(["ffmpeg", "-y", "-loglevel", "error", "-f", "concat", "-safe", "0",
         "-i", lst, "-c", "copy", track])
    print(f"  track {dur(track):.2f}s -> {track}")
    return track


if __name__ == "__main__":
    items = json.load(open(sys.argv[1]))
    lay_track(items, sys.argv[2])
