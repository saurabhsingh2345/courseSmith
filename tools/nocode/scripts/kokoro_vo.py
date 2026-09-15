#!/usr/bin/env python3
"""Speak a list of lines with Kokoro (house voice) and write mp3s.

Usage:  kokoro_vo.py <manifest.json> <outdir>

manifest.json: [{"file": "01.mp3", "text": "..."} , ...]
Writes each as mp3, prints "<file> <duration>".
"""
import json, os, subprocess, sys, wave
import numpy as np

SR = 24000
VOICE = 'af_heart'
SPEED = float(os.environ.get('KVO_SPEED', '0.85'))
SENT_GAP = float(os.environ.get('KVO_GAP', '0.40'))      # floor between sentences, matches the house pace
HEAD = 0.12
TAIL = 0.35


def speak(text, path):
    body = json.dumps({'model': 'kokoro', 'voice': VOICE, 'input': text,
                       'response_format': 'wav', 'speed': SPEED})
    subprocess.run(['curl', '-s', '-X', 'POST',
                    'http://localhost:8880/v1/audio/speech',
                    '-H', 'Content-Type: application/json',
                    '-d', body, '-o', path], check=True)
    if os.path.getsize(path) < 1200:
        raise SystemExit(f'Kokoro returned nothing for {text[:70]!r}')


def read_wav(path):
    with wave.open(path, 'rb') as w:
        n, sr, ch = w.getnframes(), w.getframerate(), w.getnchannels()
        raw = w.readframes(n)
    a = np.frombuffer(raw, dtype='<i2').astype(np.float32) / 32768.0
    if ch > 1:
        a = a.reshape(-1, ch).mean(1)
    return a, sr


def trim(a, thresh=0.006):
    idx = np.where(np.abs(a) > thresh)[0]
    if len(idx) == 0:
        return a
    return a[max(0, idx[0] - 400):min(len(a), idx[-1] + 900)]


def split_sentences(text):
    out, cur = [], ''
    for ch in text:
        cur += ch
        if ch in '.!?':
            out.append(cur.strip()); cur = ''
    if cur.strip():
        out.append(cur.strip())
    return [s for s in out if s]


def render(text, tmp, sr_out=SR):
    """Speak sentence by sentence so the gaps are ours, not the model's."""
    parts = []
    sents = split_sentences(text)
    for i, s in enumerate(sents):
        speak(s, tmp)
        a, sr = read_wav(tmp)
        a = trim(a)
        parts.append(a)
        if i < len(sents) - 1:
            parts.append(np.zeros(int(sr * SENT_GAP), dtype=np.float32))
    a = np.concatenate(parts) if parts else np.zeros(1, dtype=np.float32)
    a = np.concatenate([np.zeros(int(sr_out * HEAD), np.float32), a,
                        np.zeros(int(sr_out * TAIL), np.float32)])
    peak = float(np.max(np.abs(a))) or 1.0
    a = a * (0.90 / peak)
    return a, sr_out


def write_mp3(a, sr, path):
    raw = (np.clip(a, -1, 1) * 32767).astype('<i2').tobytes()
    subprocess.run(['ffmpeg', '-y', '-loglevel', 'error',
                    '-f', 's16le', '-ar', str(sr), '-ac', '1', '-i', 'pipe:0',
                    '-c:a', 'libmp3lame', '-b:a', '192k', path],
                   input=raw, check=True)


if __name__ == '__main__':
    manifest, outdir = sys.argv[1], sys.argv[2]
    os.makedirs(outdir, exist_ok=True)
    items = json.load(open(manifest))
    tmp = os.path.join(outdir, '_tmp.wav')
    total = 0.0
    report = []
    for it in items:
        a, sr = render(it['text'], tmp)
        dest = os.path.join(outdir, it['file'])
        write_mp3(a, sr, dest)
        d = len(a) / sr
        total += d
        report.append({'file': it['file'], 'dur': round(d, 3)})
        print(f"{it['file']} {d:.3f}", flush=True)
    if os.path.exists(tmp):
        os.remove(tmp)
    json.dump(report, open(os.path.join(outdir, 'durations.json'), 'w'), indent=2)
    print(f"TOTAL {total:.2f}s")
