#!/usr/bin/env python3
"""Speak a slide deck with Adam at natural pace and report each slide's length.

For the animated lectures there is no pre-existing picture to fit, so the voice
leads: each slide's window becomes exactly as long as its line needs.
Usage: adam_deck.py <manifest.json> <outdir>
"""
import json, os, sys
import numpy as np
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import adam_lay as A

GAP = 0.42          # between sentences inside a slide
LEAD = 0.22
TAIL = 0.60

def _clauses(sent, min_chars=18):
    """Where each clause of one sentence starts, as a fraction of the sentence.

    Always starts with 0.0 (the sentence itself). A fragment shorter than
    `min_chars` is not worth a reveal of its own - "Build, then review, then
    test." should move twice, not four times - so short pieces are folded into
    the clause before them.
    """
    marks, i = [0], 0
    while i < len(sent):
        nxt = min((p for p in (sent.find(s, i + 1) for s in (", ", "; ", " — "))
                   if p != -1), default=-1)
        if nxt == -1:
            break
        if nxt - marks[-1] >= min_chars and len(sent) - nxt >= min_chars:
            marks.append(nxt + 2)
        i = nxt + 1
    return [m / len(sent) for m in marks]


def build(manifest, outdir):
    outdir = os.path.abspath(outdir)
    raw = os.path.join(outdir, "raw"); os.makedirs(raw, exist_ok=True)
    os.makedirs(outdir, exist_ok=True)
    items = json.load(open(manifest))
    durs, cues = [], []
    for it in items:
        tag = it["file"].replace(".mp3", "")
        clips = []
        for i, s in enumerate(A.sentences(it["text"])):
            p = os.path.join(raw, f"{tag}_{i:02d}.mp3")
            A.speak(s, p)
            clips.append(A.load(p))
        # When each phrase STARTS, relative to the slide. The deck reveals
        # element i on cue i, so the picture moves on the words that describe
        # it instead of cascading through everything in the first second and
        # then sitting still for the rest of the slide.
        #
        # Cues are per CLAUSE, not per sentence. A slide whose narration is one
        # long sentence would otherwise get exactly one cue and then hold still
        # for ten seconds, which is the defect we are removing - and a long
        # sentence is precisely the case with several things to point at. A
        # clause boundary is where a reader breathes, so it is where the next
        # row wants to land. There is no word-level alignment here, so the
        # split inside a sentence is estimated by character position across the
        # clip's measured length; that is accurate to a fraction of a second at
        # speaking pace, and the reveal is a fade, not a cut.
        at, cue = LEAD, []
        for c, sent in zip(clips, A.sentences(it["text"])):
            span = len(c) / A.SR
            for frac in _clauses(sent):
                cue.append(round(at + frac * span, 3))
            at += span + GAP
        cues.append(cue)
        parts = [np.zeros(int(A.SR * LEAD), np.float32)]
        for i, c in enumerate(clips):
            parts.append(c)
            if i < len(clips) - 1:
                parts.append(np.zeros(int(A.SR * GAP), np.float32))
        parts.append(np.zeros(int(A.SR * TAIL), np.float32))
        a = np.concatenate(parts)
        peak = float(np.max(np.abs(a))) or 1.0
        a = a * (0.92 / peak)
        wav = os.path.join(outdir, tag + ".wav")
        A.write_wav(a, wav)
        mp3 = os.path.join(outdir, tag + ".mp3")
        A.run(["ffmpeg", "-y", "-loglevel", "error", "-i", wav,
               "-c:a", "libmp3lame", "-b:a", "192k", mp3])
        d = len(a) / A.SR
        durs.append(round(d, 3))
        print(f"  {tag}  {d:6.2f}s  ({len(clips)} sentences)", flush=True)
    json.dump(durs, open(os.path.join(outdir, "durs.json"), "w"))
    json.dump(cues, open(os.path.join(outdir, "cues.json"), "w"))
    long = [(i, d) for i, d in enumerate(durs) if d > 10.0]
    print(f"\n  slides {len(durs)}  speech total {sum(durs):.1f}s")
    if long:
        # A slide over ten seconds goes still before it ends however well it
        # animates - that is the defect this pass exists to remove. Split it.
        print(f"  {len(long)} slide(s) over 10s, longest {max(d for _, d in long):.1f}s:")
        for i, d in long:
            print(f"    slide {i:2d} ({items[i]['file']})  {d:5.1f}s  "
                  f"{len(cues[i])} sentence(s)")
    return durs

if __name__ == "__main__":
    build(sys.argv[1], sys.argv[2])
