#!/usr/bin/env python3
"""Speak the cut, then hand the renderer the timings it produced.

Each beat's line is synthesised on its own, trimmed, and measured. The beat's
length on screen is then lead + speech + pad (floored at `min`), so picture is
laid out against real speech durations rather than against a guess. The same
pass builds the finished audio bed: voice, a very quiet pad ducked under it, and
a swell into every chapter card.
"""

import json
import os
import subprocess
import sys
import wave

import numpy as np

HERE = os.path.dirname(os.path.abspath(__file__))
RENDERER = os.path.abspath(os.path.join(HERE, '..', '..'))
PUBLIC = os.path.join(RENDERER, 'public', 'nocode1')
WORK = os.path.join(HERE, '.voice')
os.makedirs(WORK, exist_ok=True)
os.makedirs(PUBLIC, exist_ok=True)

SR = 48000
FPS = 30
VOICE = 'af_heart'
SPEED = 0.85          # matches the pace of the original narration
LEAD = 0.42           # default silence before a line
PAD = 0.55            # default silence after a line
VOICE_RMS = 0.085     # target speech level before the master pass


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
        raise SystemExit(f'Kokoro returned nothing for: {text[:60]!r}')


def read_wav(path):
    with wave.open(path, 'rb') as w:
        n, sr, ch = w.getnframes(), w.getframerate(), w.getnchannels()
        raw = w.readframes(n)
    a = np.frombuffer(raw, dtype='<i2').astype(np.float32) / 32768.0
    if ch > 1:
        a = a.reshape(-1, ch).mean(1)
    if sr != SR:
        # linear resample is plenty for a mono speech clip
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
    env = np.abs(a)
    env = np.convolve(env, np.ones(win) / win, mode='same')
    loud = np.where(env > thresh)[0]
    if len(loud) == 0:
        return a
    k = int(keep * SR)
    return a[max(0, loud[0] - k): min(len(a), loud[-1] + k)]


# --------------------------------------------------------------------------
# Score: a pad that stays out of the way, and a swell into each chapter.

def swell(dur_s, seed):
    """Filtered noise rising into a soft peak — the sound of a scene change."""
    n = int(dur_s * SR)
    rng = np.random.default_rng(seed)
    x = rng.normal(0, 1, n).astype(np.float32)
    # one-pole lowpass, sweeping upward, so it opens as it rises
    y = np.empty_like(x)
    acc = 0.0
    cut = np.linspace(0.004, 0.10, n)
    for i in range(n):
        acc += cut[i] * (x[i] - acc)
        y[i] = acc
    env = np.linspace(0, 1, n) ** 2.6
    y *= env
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


# --------------------------------------------------------------------------

def main():
    edl_path = os.path.join(HERE, 'edl.json')
    beats = json.load(open(edl_path))
    print(f'{len(beats)} beats')

    clips, plan = [], []
    for b in beats:
        scene = b['scene'] or {}
        lead = b.get('lead', LEAD)
        pad = b.get('pad', PAD)

        # A run's length is the footage it plays, full stop. Its narration is a
        # set of cues pinned to offsets inside that stretch, so the picture is
        # never cut to fit a sentence — the sentences are placed to fit the
        # picture.
        if scene.get('k') == 'run':
            dur = (scene['to'] - scene['from']) / float(scene.get('rate') or 1)
            frames = int(round(dur * FPS))
            pieces = []
            for i, cue in enumerate(b.get('cues') or []):
                wav = os.path.join(WORK, f"{b['id']}-{i}.wav")
                if not os.path.exists(wav):
                    speak(cue['say'], wav)
                a = trim(read_wav(wav))
                end = cue['at'] + len(a) / SR
                if end > dur:
                    print(f"  ! {b['id']} cue {i} runs {end - dur:.1f}s past the footage")
                pieces.append((cue['at'], a))
            clips.append(pieces)
            plan.append({'id': b['id'], 'lead': 0.0, 'speech': 0.0,
                         'frames': frames, 'sweep': None, 'run': True})
            continue

        say = (b.get('say') or '').strip()
        if say:
            wav = os.path.join(WORK, f"{b['id']}.wav")
            if not os.path.exists(wav):
                speak(say, wav)
            a = trim(read_wav(wav))
        else:
            a = np.zeros(0, dtype=np.float32)
        speech = len(a) / SR
        hold = b.get('hold', 0.0)
        dur = max(b.get('min', 0.0),
                  lead + speech + pad + hold if say else b.get('min', 2.0))
        frames = max(int(round(dur * FPS)), 12)

        # Cards that the narration walks through light up as their phrase is
        # spoken. Kokoro's delivery is even enough that the character offset of
        # a phrase inside the line predicts its time to well under the length of
        # a highlight, so no second alignment pass is needed.
        sweep = None
        anchors = scene.get('sweepAt')
        if anchors and say and speech > 0:
            low = say.lower()
            sweep = []
            cursor = 0
            for phrase in anchors:
                at = low.find(phrase.lower(), cursor)
                if at < 0:
                    at = low.find(phrase.lower())
                if at < 0:
                    at = cursor
                cursor = max(cursor, at + 1)
                sweep.append(int(round((lead + (at / len(say)) * speech) * FPS)))

        clips.append([(lead, a)] if len(a) else [])
        plan.append({'id': b['id'], 'lead': lead, 'speech': speech,
                     'frames': frames, 'sweep': sweep, 'run': False})

    # lay the beats end to end
    start = 0
    for p in plan:
        p['start'] = start
        start += p['frames']
    total = start
    total_s = total / FPS
    print(f'total {total} frames = {total_s/60:.2f} min')

    # ---- voice track
    voice = np.zeros(int(total_s * SR) + SR, dtype=np.float32)
    for p, pieces in zip(plan, clips):
        for at, a in pieces:
            if len(a):
                add(voice, a, p['start'] / FPS + at)

    speech_mask = np.abs(voice) > 0.004
    rms = float(np.sqrt(np.mean(voice[speech_mask] ** 2))) if speech_mask.any() else 0.05
    voice *= VOICE_RMS / max(rms, 1e-6)
    print(f'voice rms {rms:.4f} -> {VOICE_RMS}')

    # ---- score
    music = np.zeros_like(voice)
    ids = {b['id']: b for b in beats}
    for p in plan:
        sc = ids[p['id']]['scene']
        if sc.get('k') in ('chapter', 'title', 'outro'):
            at = p['start'] / FPS
            add(music, swell(1.4, sum(map(ord, p['id'])) % 9999), at - 1.25)
            add(music, thump(), at)

    write_wav(os.path.join(WORK, 'voice.wav'), voice)
    write_wav(os.path.join(WORK, 'music.wav'), music)

    # ---- mix: shape the voice, duck the bed under it, master to -16 LUFS
    out = os.path.join(PUBLIC, 'narration.wav')
    run([
        'ffmpeg', '-v', 'error',
        '-i', os.path.join(WORK, 'voice.wav'),
        '-i', os.path.join(WORK, 'music.wav'),
        '-filter_complex',
        # voice: tame the peaks, lift presence a touch, roll off rumble
        '[0:a]highpass=f=85,'
        'acompressor=threshold=0.12:ratio=3:attack=8:release=180:makeup=1.6,'
        'treble=g=2.4:f=3800:width_type=q:w=0.9,'
        'alimiter=limit=0.90[vv];'
        # a filter output can only be consumed once, and the voice is needed
        # twice: in the mix, and as the sidechain key that ducks the bed.
        '[vv]asplit=2[v][vkey];'
        # bed: dull it right down, then duck it whenever the voice is present
        '[1:a]highpass=f=38,volume=0.9[m];'
        '[m][vkey]sidechaincompress=threshold=0.03:ratio=6:attack=20:release=420[md];'
        '[v][md]amix=inputs=2:weights=1 1:normalize=0,'
        'loudnorm=I=-16:TP=-1.5:LRA=11[out]',
        '-map', '[out]', '-ac', '1', '-ar', str(SR), '-c:a', 'pcm_s16le',
        '-y', out,
    ])
    print(f'wrote {out}')

    timing = {
        'total': total,
        'audio': 'nocode1/narration.wav',
        'beats': [
            {k: v for k, v in
             {'id': p['id'], 'frames': p['frames'], 'start': p['start'],
              'sweep': p['sweep']}.items() if v is not None}
            for p in plan
        ],
    }
    tpath = os.path.join(HERE, 'timing.json')
    json.dump(timing, open(tpath, 'w'), indent=2)
    print(f'wrote {tpath}')

    mins = int(total_s // 60)
    print(f'\nRUNTIME {mins}:{total_s - mins*60:04.1f}')
    for p in plan:
        print(f"  {p['start']/FPS:7.2f}  {p['frames']/FPS:5.2f}s  {p['id']}")


if __name__ == '__main__':
    sys.exit(main())
