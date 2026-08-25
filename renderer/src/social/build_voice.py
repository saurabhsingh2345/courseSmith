#!/usr/bin/env python3
"""Speak the cut, then hand the film the timings it produced.

Every beat here is a graphic scene or a component — there is no captured footage
in this cut — so every beat's length is simply lead + measured speech + pad, plus
an optional `hold` for a beat that needs to sit still and be read after the line
lands. Picture is laid out against those measurements, so rewriting one sentence
reflows the film rather than desyncing everything after it.

Reads  src/social/edl.json   (written by dump-edl.mjs)
Writes public/social/narration.wav  and  src/social/timing.json
"""

import json
import os
import subprocess
import sys
import wave

import numpy as np

HERE = os.path.dirname(os.path.abspath(__file__))
RENDERER = os.path.abspath(os.path.join(HERE, '..', '..'))
PUBLIC = os.path.join(RENDERER, 'public', 'social')
WORK = os.path.join(HERE, '.voice')
os.makedirs(WORK, exist_ok=True)
os.makedirs(PUBLIC, exist_ok=True)

SR = 48000
FPS = 30
VOICE = 'af_heart'
SPEED = 0.85          # the house pace for a beginner lesson
LEAD = 0.40           # silence before a line
PAD = 0.58            # silence after a line
VOICE_RMS = 0.085


def run(cmd, **kw):
    return subprocess.run(cmd, check=True, capture_output=True, **kw)


def speak(text, path):
    body = json.dumps({
        'model': 'kokoro', 'voice': VOICE, 'input': text,
        'response_format': 'wav', 'speed': SPEED,
    })
    run(['curl', '-s', '-X', 'POST', 'http://localhost:8880/v1/audio/speech',
         '-H', 'Content-Type: application/json', '-d', body, '-o', path])
    if os.path.getsize(path) < 1200:
        raise SystemExit(f'Kokoro returned nothing for: {text[:70]!r}')


def read_wav(path):
    with wave.open(path, 'rb') as w:
        n, sr, ch = w.getnframes(), w.getframerate(), w.getnchannels()
        raw = w.readframes(n)
    a = np.frombuffer(raw, dtype='<i2').astype(np.float32) / 32768.0
    if ch > 1:
        a = a.reshape(-1, ch).mean(1)
    if sr != SR:
        a = np.interp(np.linspace(0, len(a), int(len(a) * SR / sr), endpoint=False),
                      np.arange(len(a)), a).astype(np.float32)
    return a


def write_wav(path, a):
    a = np.clip(a, -1.0, 1.0)
    with wave.open(path, 'wb') as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(SR)
        w.writeframes((a * 32767.0).astype('<i2').tobytes())


def trim(a, thresh=0.006, keep=0.05):
    """Drop the silence Kokoro pads on either side, keeping a little air."""
    win = int(0.01 * SR)
    if len(a) < win * 3:
        return a
    env = np.convolve(np.abs(a), np.ones(win) / win, mode='same')
    loud = np.where(env > thresh)[0]
    if len(loud) == 0:
        return a
    k = int(keep * SR)
    return a[max(0, loud[0] - k): min(len(a), loud[-1] + k)]


# --------------------------------------------------------------- score

def swell(dur_s, seed):
    """Filtered noise rising into a soft peak — the sound of a scene change."""
    n = int(dur_s * SR)
    rng = np.random.default_rng(seed)
    x = rng.normal(0, 1, n).astype(np.float32)
    y = np.empty_like(x)
    acc = 0.0
    cut = np.linspace(0.004, 0.10, n)
    for i in range(n):
        acc += cut[i] * (x[i] - acc)
        y[i] = acc
    y *= np.linspace(0, 1, n) ** 2.6
    y /= (np.max(np.abs(y)) + 1e-9)
    return (y * 0.055).astype(np.float32)


def thump():
    """A short low sine drop under a chapter number landing."""
    n = int(0.55 * SR)
    t = np.arange(n, dtype=np.float32) / SR
    f = 92.0 * np.exp(-t * 5.5) + 38.0
    y = np.sin(2 * np.pi * np.cumsum(f) / SR) * np.exp(-t * 7.0)
    return (y * 0.11).astype(np.float32)


def add(track, clip, at_s):
    o = int(at_s * SR)
    if o < 0:
        clip = clip[-o:]
        o = 0
    n = min(len(clip), len(track) - o)
    if n > 0:
        track[o:o + n] += clip[:n]


# --------------------------------------------------------------- main

def main():
    beats = json.load(open(os.path.join(HERE, 'edl.json')))
    print(f'{len(beats)} beats')

    words = sum(len((b.get('say') or '').split()) for b in beats)
    print(f'{words} words of narration')

    clips, plan = [], []
    for b in beats:
        scene = b['scene'] or {}
        lead = b.get('lead', LEAD)
        pad = b.get('pad', PAD)
        hold = b.get('hold', 0.0)

        say = (b.get('say') or '').strip()
        wav = os.path.join(WORK, f"{b['id']}.wav")
        if say and not os.path.exists(wav):
            speak(say, wav)
        a = trim(read_wav(wav)) if say else np.zeros(0, dtype=np.float32)
        speech = len(a) / SR

        dur = max(b.get('min', 0.0), lead + speech + pad + hold) if say else b.get('min', 2.0)
        frames = max(int(round(dur * FPS)), 12)

        # A scene the narration walks through lights each item up as its phrase
        # is spoken. Kokoro's delivery is even enough that a phrase's character
        # offset inside the line predicts its time to well inside the length of
        # a highlight, so no second alignment pass is needed.
        sweep = None
        anchors = scene.get('sweepAt')
        if anchors and say and speech > 0:
            low = say.lower()
            sweep, cursor = [], 0
            for phrase in anchors:
                at = low.find(phrase.lower(), cursor)
                if at < 0:
                    at = low.find(phrase.lower())
                if at < 0:
                    print(f"  ! {b['id']}: sweep phrase not in line: {phrase!r}")
                    at = cursor
                cursor = max(cursor, at + 1)
                sweep.append(int(round((lead + (at / len(say)) * speech) * FPS)))

        clips.append((lead, a))
        plan.append({'id': b['id'], 'speech': speech, 'frames': frames, 'sweep': sweep})

    start = 0
    for p in plan:
        p['start'] = start
        start += p['frames']
    total = start
    total_s = total / FPS

    voice = np.zeros(int(total_s * SR) + SR, dtype=np.float32)
    for p, (at, a) in zip(plan, clips):
        if len(a):
            add(voice, a, p['start'] / FPS + at)

    mask = np.abs(voice) > 0.004
    rms = float(np.sqrt(np.mean(voice[mask] ** 2))) if mask.any() else 0.05
    voice *= VOICE_RMS / max(rms, 1e-6)

    music = np.zeros_like(voice)
    kinds = {b['id']: (b['scene'] or {}).get('k') for b in beats}
    for p in plan:
        if kinds[p['id']] in ('chapter', 'title', 'outro'):
            at = p['start'] / FPS
            add(music, swell(1.4, sum(map(ord, p['id'])) % 9999), at - 1.25)
            add(music, thump(), at)

    write_wav(os.path.join(WORK, 'voice.wav'), voice)
    write_wav(os.path.join(WORK, 'music.wav'), music)

    out = os.path.join(PUBLIC, 'narration.wav')
    run([
        'ffmpeg', '-v', 'error',
        '-i', os.path.join(WORK, 'voice.wav'),
        '-i', os.path.join(WORK, 'music.wav'),
        '-filter_complex',
        '[0:a]highpass=f=85,'
        'acompressor=threshold=0.12:ratio=3:attack=8:release=180:makeup=1.6,'
        'treble=g=2.4:f=3800:width_type=q:w=0.9,'
        'alimiter=limit=0.90[vv];'
        # a filter output can only be consumed once, and the voice is needed
        # twice: in the mix, and as the key that ducks the bed under it
        '[vv]asplit=2[v][vkey];'
        '[1:a]highpass=f=38,volume=0.9[m];'
        '[m][vkey]sidechaincompress=threshold=0.03:ratio=6:attack=20:release=420[md];'
        '[v][md]amix=inputs=2:weights=1 1:normalize=0,'
        'loudnorm=I=-16:TP=-1.5:LRA=11[out]',
        '-map', '[out]', '-ac', '1', '-ar', str(SR), '-c:a', 'pcm_s16le',
        '-y', out,
    ])

    timing = {
        'total': total,
        'audio': 'social/narration.wav',
        'beats': [
            {k: v for k, v in
             {'id': p['id'], 'frames': p['frames'], 'start': p['start'],
              'sweep': p['sweep']}.items() if v is not None}
            for p in plan
        ],
    }
    json.dump(timing, open(os.path.join(HERE, 'timing.json'), 'w'), indent=2)

    mins = int(total_s // 60)
    print(f'\nRUNTIME {mins}:{total_s - mins * 60:04.1f}   ({words} words, '
          f'{words / (total_s / 60):.0f} wpm on screen)')
    for p in plan:
        print(f"  {p['start']/FPS:7.2f}  {p['frames']/FPS:5.2f}s  {p['id']}")


if __name__ == '__main__':
    sys.exit(main())
