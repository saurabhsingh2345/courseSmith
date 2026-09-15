#!/usr/bin/env python3
"""Generate renderer/src/nocode/l32.tsx — parts 3 and 4, and two real recoveries.

L31 ended on "now watch what happens next — two things go wrong", and both of
them did, on camera, without being arranged: a start script the machine refused
to run (the executable flag was never set) and a container that could not bind
port 8000 because something stale was still holding it. A1-A9 are those two, and
they are the honest answer to what a good stretch actually looks like.

The checkpoint section is NOT the reference's. His agent skipped the commits he
had asked for; ours made them. So F reads the history instead of complaining
about it, and finds two things that are true of ours: part 2 never got a
checkpoint of its own — its files were swept into part 3's commit, so there is no
longer a point to return to at the end of part 2 — and one back-end file is in no
commit at all. Same lesson as L31's D10-D12, arriving one lecture later and from
a direction I did not plan: you find out what is true by reading what is there.

The 80% coverage trap (l32_trap.json, written for this slot) is NOT here. It did
not happen during parts 3 and 4 — the tests it wrote were short and tested the
two things that matter. The trap is held for whichever lecture it actually occurs
in rather than staged, and E2-E3 plant it instead.
"""
import json, os, subprocess, sys

R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
VO = sys.argv[1] if len(sys.argv) > 1 else "/tmp/vo32"
TARGET = None           # his L29 is 11:20; we run long on purpose
GAP_FIXED = 0.24

TAKES = {
    'P': 'L32_A_p34_s01',       # parts 3+4: the two errors, the build, the tests
    'Q': 'L32_B_p34b_s01',      # the run completing, and the commits appearing
    'K': 'L32_C_checks_s01',    # test-it-yourself, the criteria audit, the why
    'W': 'L32_D_web_s01',       # sign in, log out, log back in, board intact
    'G': 'L32_E_git_s01',       # the history, the swept commit, the three commands
}

GROUPS = {}
CUT = []
H = 'head'
def main():
    all_durs = json.load(open(os.path.join(VO, 'durs.json')))
    if len(all_durs) != len(SPEC):
        raise SystemExit(f"durs has {len(all_durs)} entries, SPEC has {len(SPEC)}")
    bad = [c for c in CUT if c not in [s[0] for s in SPEC]]
    if bad:
        raise SystemExit(f"CUT names not in SPEC: {bad}")

    keep = [(nm, sp, d) for (nm, sp), d in zip(SPEC, all_durs) if nm not in CUT]
    spec = [(nm, sp) for nm, sp, _ in keep]
    durs = [d for _, _, d in keep]
    names = [nm for nm, _, _ in keep]
    print(f"cut {len(CUT)} voiced slides, keeping {len(spec)}")
    SPEC[:] = spec

    speech = sum(durs)
    gap = GAP_FIXED if TARGET is None else max(0.0, (TARGET - speech) / (len(durs) - 1))
    film = speech + gap * (len(durs) - 1)
    print(f"speech {speech:.1f}s  slides {len(durs)}  GAP {gap:.3f}  "
          f"film {film:.1f}s = {int(film//60)}:{int(film%60):02d}  (his 11:20)")

    windows = [d + gap for d in durs[:-1]] + [durs[-1]]

    # share each group's range across its slides, weighted by window
    members = {}
    for i, (nm, spec) in enumerate(SPEC):
        if spec[0] in ('F', 'Fs'):
            members.setdefault(spec[1], []).append(i)
    spans = {}
    for gname, idxs in members.items():
        take, ranges = GROUPS[gname]
        total_win = sum(windows[i] for i in idxs)
        total_src = sum(b - a for a, b in ranges)
        # lay the slides end to end along the concatenated ranges, so a hole in
        # the middle of a take is stepped over rather than sampled
        cursor = 0.0
        for i in idxs:
            length = total_src * (windows[i] / total_win)
            start, remaining = cursor, length
            # map `start` into the range list
            acc = 0.0
            for a, b in ranges:
                span = b - a
                if start < acc + span:
                    off = start - acc
                    # a slide must not straddle a hole: if it would, push it
                    # wholly into the next range instead of spanning the gap
                    if off + remaining > span and (b - (a + off)) < remaining * 0.6:
                        acc += span
                        start = acc
                        continue
                    ss = a + off
                    to = min(b, ss + remaining)
                    spans[i] = (take, ss, to)
                    break
                acc += span
            else:
                a, b = ranges[-1]
                spans[i] = (take, max(a, b - length), b)
            cursor += length
    for i, (nm, spec) in enumerate(SPEC):
        if spec[0] in ('Fx', 'Fxs'):
            spans[i] = (spec[1], spec[2], spec[3])

    jobs, rates = [], []
    for i in sorted(spans):
        take, ss, to = spans[i]
        win = windows[i]
        rate = win / (to - ss)
        rates.append((rate, names[i]))
        jobs.append(['31', i, TAKES[take], round(ss, 2), round(to, 2), round(win, 3)])
    json.dump(jobs, open('/tmp/l31_jobs.json', 'w'), indent=1)

    rates.sort()
    print(f"footage slides {len(jobs)} of {len(SPEC)} = {len(jobs)/len(SPEC):.0%}")
    # A slides-only lecture has no rates at all. L29 is the first of those to go
    # through this generator, and the unguarded print crashed on an empty list.
    if rates:
        print(f"  rate range {rates[0][0]:.2f} ({rates[0][1]}) "
              f".. {rates[-1][0]:.2f} ({rates[-1][1]})")
    else:
        print("  no footage — slides only")
    # The band is deliberately loose on the top end. A rate above 1 means the
    # clip is slowed, and the P take's long stretches are a STATIC plan sitting
    # on screen — slowing a frame that barely changes is imperceptible. What
    # actually matters is the bottom end: speeding a take up past ~3x turns
    # readable text into a blur.
    # N14 is the deliberate 12-14x ramp over the seven-minute wait, per the
    # house rules — it is meant to be far out of band.
    hot = [r for r in rates
           if r[0] > 1.55 or r[0] < 0.30]
    if hot:
        print("  OUT OF BAND:", [(f'{r:.2f}', n) for r, n in hot])

    # ---- emit the tsx
    def js(v, ind=0):
        pad = ' ' * ind
        if isinstance(v, str):
            return "'" + v.replace("\\", "\\\\").replace("'", "\\'").replace("\n", "\\n") + "'"
        if isinstance(v, bool):
            return 'true' if v else 'false'
        if isinstance(v, (int, float)):
            return repr(v)
        if isinstance(v, list):
            return '[' + ', '.join(js(x) for x in v) + ']'
        if isinstance(v, dict):
            return '{' + ', '.join(f"{k}: {js(x)}" for k, x in v.items()) + '}'
        raise TypeError(v)

    lines = []
    for i, (nm, spec) in enumerate(SPEC):
        if spec[0] in ('F', 'Fx', 'Fs', 'Fxs'):
            take, ss, to = spans[i]
            if spec[0] == 'Fs':
                kick, hl, sz = spec[2], spec[3], spec[4]
            elif spec[0] == 'Fxs':
                kick, hl, sz = spec[4], spec[5], spec[6]
            else:
                kick = hl = sz = None
            extra = ''
            if hl:
                extra = (f", kicker: {js(kick)}, lines: {js(hl)}, size: {sz}")
            lines.append(f"  {{k: 'shot', src: sh('{i:02d}'){extra}}},"
                         f"   // {nm}  {take} {ss:.1f}-{to:.1f}s")
        else:
            body = ', '.join(f"{k}: {js(v)}" for k, v in spec[1].items())
            lines.append(f"  {{{body}}},   // {nm}")

    tsx = f'''// nocode31 — planning and scaffolding (his L31, 11:43)
//
// Two of his set pieces did not happen to us, and the narration says so rather
// than staging them — the same call L18 made when our Copilot build succeeded
// where his broke:
//   * his agent asked three questions off the opening prompt; ours asked none,
//     because the brief was clearer. C1-C6 make that the honest lesson: the
//     point was to find out whether the brief was clear, and silence is a pass.
//   * his agent silently swapped `uv` for a bare requirements.txt; ours produced
//     pyproject.toml correctly. D10-D12 keep the transferable shape — you find
//     out about ignored instructions from what APPEARS in the project.

import React from 'react';
import {{Deck, Slide, deckFrames}} from './kit';

export const DURS = [
{chr(10).join('  ' + ', '.join(f'{d:.3f}' for d in durs[i:i+8]) + ',' for i in range(0, len(durs), 8))}
];
export const FILES = [
{chr(10).join('  ' + ', '.join(f"'{n}'" for n in names[i:i+8]) + ',' for i in range(0, len(names), 8))}
].map((n) => `${{n}}.mp3`);
export const GAP = {gap:.3f};
export const L31_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode31/shots/${{n}}.mp4`;

const SLIDES: Slide[] = [
{chr(10).join(lines)}
];

export const Nocode31: React.FC = () => (
  <Deck slides={{SLIDES}} durs={{DURS}} voDir="nocode31/vo" files={{FILES}} gap={{GAP}} />
);
'''
    out = os.path.join(R, 'renderer/src/nocode/l31.tsx')
    open(out, 'w').write(tsx)
    print("wrote", out)
    print("wrote /tmp/l31_jobs.json")


if __name__ == '__main__':
    main()
