#!/usr/bin/env python3
"""Speak the cut with Adam, then hand the renderer the timings it produced.

Each beat's line is synthesised alone, trimmed and measured, so its length on
screen is lead + measured speech + pad. Picture cannot drift from voice, and
rewriting one line reflows the film instead of desyncing everything after it.

The ElevenLabs key and the Adam voice id come from the rig's existing .env by
absolute path. That file is gitignored and holds a secret — it is read here,
never copied.

Cache is keyed on the TEXT, not the filename: reusing a filename after editing
a line has shipped the wrong take before.
"""

import hashlib
import json
import os
import subprocess
import sys
import time
import urllib.request
import wave

import numpy as np

HERE = os.path.dirname(os.path.abspath(__file__))
RENDERER = os.path.abspath(os.path.join(HERE, '..', '..'))
PUBLIC = os.path.join(RENDERER, 'public', 'genghis')
WORK = os.path.join(HERE, '.voice')
ENV = '/Users/enfecsolutions/Desktop/enfec_subs/courseSmith/tools/nocode/nc/No-Code/.env'

SR = 48000
FPS = 30
LEAD = 0.45           # default silence before a line
PAD = 0.75            # default silence after a line — documentary breathing room
MODEL = 'eleven_multilingual_v2'
# Slightly under natural pace: this is a narration, not a read-through.
SETTINGS = {'stability': 0.5, 'similarity_boost': 0.75,
            'style': 0.0, 'use_speaker_boost': True, 'speed': 0.94}

os.makedirs(WORK, exist_ok=True)
os.makedirs(PUBLIC, exist_ok=True)


def env():
    d = {}
    for line in open(ENV):
        if '=' in line and not line.strip().startswith('#'):
            k, v = line.split('=', 1)
            d[k.strip()] = v.strip().strip('"').strip("'")
    return d


def run(cmd):
    r = subprocess.run(cmd, capture_output=True, text=True)
    if r.returncode != 0:
        raise SystemExit(f'{cmd[0]} failed:\n{r.stderr[-800:]}')
    return r


def speak(text, key, voice, dest_mp3):
    """POST to ElevenLabs, retrying with backoff. Cached on a hash of the text."""
    stamp = dest_mp3 + '.txt'
    if os.path.exists(dest_mp3) and os.path.getsize(dest_mp3) > 2000 \
            and os.path.exists(stamp) and open(stamp).read() == text:
        return 'cached'
    body = json.dumps({'text': text, 'model_id': MODEL,
                       'voice_settings': SETTINGS}).encode()
    req = urllib.request.Request(
        f'https://api.elevenlabs.io/v1/text-to-speech/{voice}',
        data=body, method='POST',
        headers={'xi-api-key': key, 'Content-Type': 'application/json',
                 'Accept': 'audio/mpeg'})
    for attempt in range(5):
        try:
            with urllib.request.urlopen(req, timeout=180) as r:
                open(dest_mp3, 'wb').write(r.read())
            break
        except urllib.error.HTTPError as e:
            detail = e.read()[:300].decode('utf8', 'replace')
            if e.code == 422 and 'speed' in detail:
                # Older voice-settings schema on this account: drop `speed`
                # and let ffmpeg handle pace instead of failing the build.
                SETTINGS.pop('speed', None)
                return speak(text, key, voice, dest_mp3)
            if attempt == 4:
                raise SystemExit(f'ElevenLabs {e.code}: {detail}')
            print(f'    retry {attempt + 1} after {e.code}')
            time.sleep(2 ** attempt)
        except Exception as exc:
            if attempt == 4:
                raise
            print(f'    retry {attempt + 1} after {exc}')
            time.sleep(2 ** attempt)
    open(stamp, 'w').write(text)
    return 'spoken'


def to_wav(mp3, wav):
    run(['ffmpeg', '-y', '-v', 'error', '-i', mp3, '-ac', '1', '-ar', str(SR),
         '-c:a', 'pcm_s16le', wav])


def read_wav(path):
    with wave.open(path, 'rb') as w:
        a = np.frombuffer(w.readframes(w.getnframes()), dtype='<i2')
    return a.astype(np.float32) / 32768.0


def write_wav(path, mono, sr=SR):
    x = np.clip(mono, -1.0, 1.0)
    with wave.open(path, 'wb') as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(sr)
        w.writeframes((x * 32767).astype('<i2').tobytes())


def trim(a, thresh=0.0035, keep=0.05):
    """Strip leading/trailing near-silence, keeping a little air either side."""
    loud = np.abs(a) > thresh
    if not loud.any():
        return a
    i, j = np.argmax(loud), len(a) - np.argmax(loud[::-1])
    k = int(keep * SR)
    return a[max(0, i - k):min(len(a), j + k)]


def main():
    edl = json.load(open(os.path.join(HERE, 'edl.json')))
    e = env()
    key, voice = e['ELEVENLABS_API_KEY'], e['ELEVENLABS_VOICE_ID']

    beats, cursor = [], 0
    clips = {}
    for b in edl:
        text = (b.get('say') or '').strip()
        lead = float(b.get('lead', LEAD if text else 0))
        pad = float(b.get('pad', PAD if text else 0))
        speech = 0.0
        if text:
            h = hashlib.sha1(text.encode()).hexdigest()[:10]
            mp3 = os.path.join(WORK, f"{b['id']}-{h}.mp3")
            wav = mp3[:-4] + '.wav'
            state = speak(text, key, voice, mp3)
            if state == 'spoken' or not os.path.exists(wav):
                to_wav(mp3, wav)
            a = trim(read_wav(wav))
            speech = len(a) / SR
            clips[b['id']] = a
            print(f"  {state:7s} {speech:6.2f}s  {b['id']}")
        total = lead + speech + pad
        frames = max(int(round(total * FPS)), 30)
        beats.append({
            'id': b['id'],
            'startFrame': cursor,
            'durFrames': frames,
            'leadFrames': int(round(lead * FPS)),
            'speechFrames': int(round(speech * FPS)),
            'say': text,
        })
        cursor += frames

    total_s = cursor / FPS
    print(f"\n  {len(beats)} beats · {cursor} frames · "
          f"{int(total_s // 60)}:{total_s % 60:04.1f}")

    # Lay the spoken lines onto one narration track at their beat positions.
    track = np.zeros(int(cursor / FPS * SR) + SR, dtype=np.float32)
    for b in beats:
        a = clips.get(b['id'])
        if a is None:
            continue
        at = int((b['startFrame'] + b['leadFrames']) / FPS * SR)
        track[at:at + len(a)] += a
    # Normalise the voice to a predictable level; the master pass does loudness.
    rms = np.sqrt((track ** 2).mean())
    if rms > 0:
        track *= 0.085 / rms
    write_wav(os.path.join(PUBLIC, 'narration.wav'), track)
    json.dump({'fps': FPS, 'totalFrames': cursor, 'beats': beats},
              open(os.path.join(HERE, 'timing.json'), 'w'), indent=1)
    json.dump({'fps': FPS, 'totalFrames': cursor, 'beats': beats},
              open(os.path.join(PUBLIC, 'timing.json'), 'w'), indent=1)
    print('  wrote narration.wav + timing.json')


if __name__ == '__main__':
    main()
