#!/usr/bin/env python3
"""Synthesise the score, mix it under Adam, and master the result.

The music is generated here rather than licensed. That keeps the whole film
clean: public-domain picture, an original score, and a voice we hold rights to.
It also means the score is written against the cut — the drum arrives with the
army and stops dead at the section on what the conquest cost — instead of a
library track being trimmed to length.

Voice: a drone on D with its fifth, a frame drum, and a bowed sustain that only
appears twice. Nothing that pretends to be Mongolian folk music; it is a
documentary bed with a pulse.
"""

import json
import os
import subprocess
import wave

import numpy as np

HERE = os.path.dirname(os.path.abspath(__file__))
RENDERER = os.path.abspath(os.path.join(HERE, '..', '..'))
PUBLIC = os.path.join(RENDERER, 'public', 'genghis')
SR = 48000

# Per beat: drone level, drum bpm (0 = silent), bow level, colour.
# `colour` shifts the drone's upper voice: 'open' is the fifth, 'dark' flattens
# it to a minor second against the root, which is what the cost section needs.
PLAN = {
    'open-steppe':  (0.50, 0,  0.00, 'open'),
    'open-turn':    (0.62, 0,  0.20, 'open'),
    'title':        (0.80, 0,  0.34, 'open'),
    'steppe-cold':  (0.44, 0,  0.00, 'open'),
    'steppe-eat':   (0.44, 0,  0.00, 'open'),
    'steppe-ride':  (0.48, 0,  0.10, 'open'),
    'born':         (0.52, 0,  0.22, 'open'),
    'hard-years':   (0.56, 0,  0.14, 'dark'),
    'merit':        (0.52, 44, 0.00, 'open'),
    'kurultai':     (0.88, 52, 0.40, 'open'),
    'machine-in':   (0.66, 60, 0.00, 'open'),
    'decimal':      (0.62, 66, 0.00, 'open'),
    'kit':          (0.62, 66, 0.00, 'open'),
    'yam':          (0.64, 78, 0.00, 'open'),
    'map-china':    (0.72, 72, 0.00, 'open'),
    'beijing-art':  (0.70, 60, 0.16, 'dark'),
    'map-west':     (0.76, 80, 0.00, 'open'),
    'map-khwarazm': (0.84, 88, 0.00, 'open'),
    'cost':         (0.78, 0,  0.26, 'dark'),
    'cost-stat':    (0.52, 0,  0.10, 'dark'),
    'death':        (0.42, 0,  0.00, 'dark'),
    'map-after':    (0.74, 68, 0.00, 'open'),
    'pax':          (0.58, 0,  0.30, 'open'),
    'ideas':        (0.54, 0,  0.24, 'open'),
    'plague':       (0.50, 0,  0.08, 'dark'),
    'close':        (0.60, 0,  0.26, 'open'),
    'end':          (0.44, 0,  0.14, 'open'),
}

ROOT = 36.71  # D1


def read_wav(path):
    with wave.open(path, 'rb') as w:
        ch, sr = w.getnchannels(), w.getframerate()
        a = np.frombuffer(w.readframes(w.getnframes()), dtype='<i2')
    a = a.astype(np.float32) / 32768.0
    if ch > 1:
        a = a.reshape(-1, ch).mean(1)
    return a, sr


def write_wav(path, x, sr=SR, channels=1):
    """Write mono, or dual-mono stereo.

    The premaster is written STEREO so the two-pass loudnorm measures the same
    channel layout that ships. Converting to stereo after the limiter (with
    `-ac 2`) applies a pan-law attenuation and the delivered film came out
    3 dB under target: bed.wav read -1.5 dBFS / -16 LUFS, the muxed file
    -4.4 dBFS / -17.1 LUFS.
    """
    x = np.clip(x, -1, 1)
    if channels == 2:
        x = np.repeat(x[:, None], 2, axis=1).reshape(-1)
    with wave.open(path, 'wb') as w:
        w.setnchannels(channels)
        w.setsampwidth(2)
        w.setframerate(sr)
        w.writeframes((x * 32767).astype('<i2').tobytes())


def box_smooth(x, seconds):
    """Moving average via a prefix sum — O(n) regardless of window length.

    np.convolve with a 2.5-second kernel over a fourteen-million-sample track
    is 1.7e12 operations and does not finish. There is no scipy in this repo to
    fall back on, and a prefix sum gives the identical result.
    """
    k = max(1, int(seconds * SR))
    c = np.concatenate(([0.0], np.cumsum(x, dtype=np.float64)))
    n = len(x)
    lo = np.clip(np.arange(n) - k // 2, 0, n)
    hi = np.clip(np.arange(n) + k - k // 2, 0, n)
    return ((c[hi] - c[lo]) / np.maximum(hi - lo, 1)).astype(np.float32)


def envelope(beats, key, n):
    """Sample-rate ramp of one plan parameter, smoothed across beat borders."""
    out = np.zeros(n, dtype=np.float32)
    idx = {'drone': 0, 'bpm': 1, 'bow': 2}[key]
    for b in beats:
        a = int(b['startFrame'] / 30 * SR)
        z = min(n, int((b['startFrame'] + b['durFrames']) / 30 * SR))
        out[a:z] = PLAN.get(b['id'], (0.5, 0, 0, 'open'))[idx]
    # Parameters glide between beats instead of stepping, which is the
    # difference between a score and a list of cues.
    return box_smooth(out, 2.5)


def main():
    timing = json.load(open(os.path.join(HERE, 'timing.json')))
    beats = timing['beats']
    n = int(timing['totalFrames'] / timing['fps'] * SR)
    t = np.arange(n, dtype=np.float32) / SR

    drone_lv = envelope(beats, 'drone', n)
    bpm = envelope(beats, 'bpm', n)
    bow_lv = envelope(beats, 'bow', n)

    # ------------------------------------------------------------------ drone
    # Root, octave and fifth, each slightly detuned against a twin so the pitch
    # beats slowly instead of sitting dead still.
    def voice(f, detune, amp, phase=0.0):
        lfo = 1.0 + 0.03 * np.sin(2 * np.pi * 0.07 * t + phase)
        return amp * (
            np.sin(2 * np.pi * f * t * lfo + phase)
            + 0.7 * np.sin(2 * np.pi * (f * (1 + detune)) * t + phase + 1.1)
        )

    dark = np.zeros(n, dtype=np.float32)
    for b in beats:
        if PLAN.get(b['id'], (0, 0, 0, 'open'))[3] == 'dark':
            a = int(b['startFrame'] / 30 * SR)
            z = min(n, int((b['startFrame'] + b['durFrames']) / 30 * SR))
            dark[a:z] = 1.0
    dark = box_smooth(dark, 2.0)

    drone = voice(ROOT, 0.0016, 0.55)
    drone += voice(ROOT * 2, 0.0011, 0.30, 0.6)
    # The upper voice: a fifth when open, a minor ninth when dark. Crossfaded,
    # so the harmony sours through the cost section rather than switching.
    fifth = voice(ROOT * 3, 0.0009, 0.20, 1.7)
    sour = voice(ROOT * 2.12, 0.0013, 0.20, 2.3)
    drone += fifth * (1 - dark) + sour * dark
    # No filter pass: every voice above is a sine, so the lowpass this used
    # to apply is already expressed in the harmonic amplitudes.
    drone *= drone_lv

    # -------------------------------------------------------------------- drum
    # A frame drum: a short pitched thump with a noise transient on the head.
    drum = np.zeros(n, dtype=np.float32)
    pos = 0.0
    rng = np.random.default_rng(7)
    while pos < n / SR:
        i = int(pos * SR)
        if i >= n:
            break
        cur = bpm[i]
        if cur < 8:
            pos += 0.25
            continue
        ln = int(0.42 * SR)
        seg = np.arange(min(ln, n - i), dtype=np.float32) / SR
        env = np.exp(-seg * 9.0)
        body = np.sin(2 * np.pi * 61.0 * seg) * env
        body += 0.45 * np.sin(2 * np.pi * 92.0 * seg) * np.exp(-seg * 15.0)
        noise = rng.normal(0, 1, len(seg)).astype(np.float32) * np.exp(-seg * 70.0) * 0.25
        hit = (body + noise) * 0.5
        # Every fourth beat is accented, so a pulse has a bar rather than being
        # a metronome.
        beat_no = int(pos * cur / 60.0)
        hit *= 1.0 if beat_no % 4 else 1.35
        drum[i:i + len(seg)] += hit
        pos += 60.0 / cur

    # --------------------------------------------------------------------- bow
    # A sustained bowed note an octave and a fifth up, with vibrato. Slow
    # attack, and only where the plan asks for it.
    vib = 1.0 + 0.006 * np.sin(2 * np.pi * 4.6 * t)
    bow = 0.5 * np.sin(2 * np.pi * ROOT * 6 * t * vib)
    bow += 0.25 * np.sin(2 * np.pi * ROOT * 9 * t * vib + 0.7)
    bow += 0.12 * np.sin(2 * np.pi * ROOT * 12 * t * vib + 1.4)
    bow = bow * bow_lv * 0.22

    score = drone * 0.5 + drum * 0.30 + bow
    # Keep the bed well under the voice; the master pass sets absolute level.
    peak = np.abs(score).max()
    if peak > 0:
        score *= 0.30 / peak

    # ------------------------------------------------------------------- duck
    voice_track, vsr = read_wav(os.path.join(PUBLIC, 'narration.wav'))
    if vsr != SR:
        raise SystemExit(f'narration.wav is {vsr} Hz, expected {SR}')
    if len(voice_track) < n:
        voice_track = np.pad(voice_track, (0, n - len(voice_track)))
    voice_track = voice_track[:n]

    # Envelope-follow the voice and pull the score down by ~5 dB under it.
    lvl = box_smooth(np.abs(voice_track), 0.25)
    lvl = lvl / max(lvl.max(), 1e-6)
    duck = 1.0 - 0.44 * np.clip(lvl * 3.2, 0, 1)
    # Smooth the gain so the duck breathes rather than pumping per syllable.
    duck = box_smooth(duck, 0.5)

    mix = voice_track + score * duck
    write_wav(os.path.join(HERE, '.voice', 'premaster.wav'), mix, channels=2)
    print(f'  premaster {len(mix) / SR:.1f}s  peak {20 * np.log10(np.abs(mix).max()):.2f} dBFS')

    # ------------------------------------------------------------------ master
    # Real two-pass loudnorm, then resample, then the limiter LAST. The
    # resample has to sit BEFORE the limiter: loudnorm always emits 192 kHz and
    # a resampler downstream of a limiter rings straight back over the ceiling.
    src = os.path.join(HERE, '.voice', 'premaster.wav')
    dst = os.path.join(PUBLIC, 'bed.wav')
    r = subprocess.run(
        ['ffmpeg', '-hide_banner', '-i', src, '-vn', '-af',
         'loudnorm=I=-16:TP=-1.5:LRA=11:print_format=json', '-f', 'null', '-'],
        capture_output=True, text=True)
    tail = r.stderr[r.stderr.rindex('{'):r.stderr.rindex('}') + 1]
    m = json.loads(tail)
    print('  measured', {k: m[k] for k in list(m)[:4]})
    subprocess.run(
        ['ffmpeg', '-y', '-v', 'error', '-i', src, '-af',
         f"loudnorm=I=-16:TP=-1.5:LRA=11:linear=true"
         f":measured_I={m['input_i']}:measured_TP={m['input_tp']}"
         f":measured_LRA={m['input_lra']}:measured_thresh={m['input_thresh']}"
         f",aresample=48000,alimiter=limit=0.841:level=disabled",
         # No -ac here: the premaster is already stereo, so there is no
         # conversion left to apply a pan-law to.
         '-ar', '48000', '-c:a', 'pcm_s16le', dst], check=True)
    v = subprocess.run(['ffmpeg', '-i', dst, '-af', 'volumedetect', '-f', 'null', '-'],
                       capture_output=True, text=True).stderr
    for line in v.splitlines():
        if 'max_volume' in line or 'mean_volume' in line:
            print('  ', line.split(']')[-1].strip())
    print('  wrote', dst)


if __name__ == '__main__':
    main()
