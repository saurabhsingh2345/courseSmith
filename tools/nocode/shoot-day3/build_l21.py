#!/usr/bin/env python3
"""Generate renderer/src/nocode/l21.tsx and the cut jobs for L21 (the verdict).

His L21 is 4:16 and it is the payload of the whole day: the takeaway is that the
IDE-based agents are all quite similar, and it arrives at 2:30 of a 4:16 lecture
as a fresh slide the viewer is asked to take on trust.

Ours does it differently, per ref-L21's own suggestion. The `matrix` template
(added to anim.tsx for this batch) has been filling in one column per build
across the day — upto 1 in L18, upto 2 in L18's close, upto 3 in L19 — so by the
time the fourth column lands here the grid is a record of what the viewer just
watched rather than a new claim. W5, W6 and V5 walk that grid.

Footage comes from L20's iteration take (the feedback loop that opens the
lecture, which is the single most useful habit in the week) plus the four
finished boards. Everything else is graphics, because the back half of this
lecture is argument, not screen.
"""
import json, os, subprocess, sys

R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
VO = sys.argv[1] if len(sys.argv) > 1 else "/tmp/vo21"
TARGET = None           # his L21 is 4:16; we run long on purpose
GAP_FIXED = 0.24

TAKES = {
    'I': 'L21_A_iterate_s01',    # typing real feedback, and the three-file fix
    'B': 'L20_S_checks_s01',     # the fourth finished board, before the fix
    'V': 'L21_B_verify_s01',     # the board after the fix, and the checks again
}

GROUPS = {
    'I_feed': ('I', [(6.0, 68.0)]),      # typing the feedback, and its first reply
    'I_fix':  ('I', [(68.0, 200.0)]),    # it working, and the three-file diff
    'V_after': ('V', [(0.0, 40.0)]),     # the board after the fix — full width
    'V_edge':  ('V', [(10.0, 30.0)]),    # held on the left edge, still touching
    'V_checks': ('V', [(40.0, 77.0)]),   # drag and rename still working
    'B_all':  ('B', [(0.0, 117.5)]),     # the fourth board, before the fix
}

CUT = []
H = 'head'


def G(k, **kw):
    return ('G', dict(k=k, **kw))


COLS = [
    {'head': 'Cursor',      'sub': 'Opus, chosen'},
    {'head': 'Copilot',     'sub': 'free tier, Auto'},
    {'head': 'Claude Code', 'sub': 'Opus, high effort'},
    {'head': 'Antigravity', 'sub': 'Gemini'},
]
ROWS = [
    {'label': 'the brief',   'cells': ['agents.md', 'agents.md', 'agents.md', 'agents.md']},
    {'label': 'the prompt',  'cells': ['go ahead and plan'] * 4},
    {'label': 'asked me',    'cells': ['never', 'eight times', 'once, first', 'never']},
    {'label': 'the plan',    'cells': ['in the panel', 'in the panel', 'in a file', 'in the panel']},
    {'label': 'five checks', 'cells': ['all five'] * 4},
]

SPEC = [
 ('W1', ('Fs', 'I_feed', 'first, the most useful habit of the week',
         ['Working in a *loop*'], 92)),
 ('W2', ('F', 'I_feed')),
 ('W3', G('rows', kicker='and look at how that feedback is phrased',
          lines=['Three parts. All *three* work'], size=58, numbered=True,
          rows=[{'t': 'The exact thing that is wrong'},
                {'t': 'Why it is wrong'},
                {'t': 'The standard you are holding it to', 'hot': True}],
          stamp='not "this is bad". Not "make it better"')),
 ('W4', ('F', 'I_fix')),
 ('X1', ('F', 'I_fix')),
 ('X2', ('Fs', 'V_after', 'and here is the board after', ['The dead space', 'is *gone*'], 62)),
 ('X3', ('Fs', 'V_edge', 'now look at the left edge', ['It did *one* of', 'the two things'], 58)),
 ('X4', G('rows', kicker='and this is not me catching it out',
          lines=['This is the loop *working*'], size=62,
          rows=[{'t': 'It made a change'},
                {'t': 'I went and looked'},
                {'t': 'Looking found the half that was not finished'},
                {'t': 'If I had taken its word, that edge would have shipped', 'hot': True}])),
 ('X5', G('myth', kicker='and notice why I could catch it at all',
          lines=['*Two* named faults'], size=76,
          wrong='the layout looks a bit off, can you tidy it up',
          right='no left padding on the first column, and the columns do not fill the width',
          note='two named faults means two things you can count off')),
 ('X6', G(H, kicker='vague feedback gets a vague answer',
          lines=['Say *make it better*', 'and you have nothing to check'], size=54)),
 ('X7', G(H, kicker='so it goes back a second time, with one line',
          lines=['Not one perfect instruction.', 'A *short loop*'], size=54, trans='rise')),
 ('V1', ('F', 'I_fix')),
 ('V2', G('rows', kicker='and the caveat comes first, not last',
          lines=['They did *not* run', 'the same model'], size=58,
          rows=[{'t': 'One had a frontier model I chose by hand'},
                {'t': 'One had whatever a free plan handed it'},
                {'t': 'Two more were different again'},
                {'t': 'So a difference could be the harness, or just the model', 'hot': True}])),
 ('V3', G('myth', kicker='which is why I keep naming it',
          lines=['The question to *ask*'], size=68,
          wrong='which of these tools is best',
          right='which model was running, and on what date',
          note='any comparison that will not tell you is not a comparison')),
 ('W5', G('matrix', kicker='four builds, one brief', lines=['The *grid*, finished'], size=62,
          cols=COLS, rows=ROWS, upto=4, dateline='28 august 2026')),
 ('V4', G('rows', kicker='and here is the honest headline',
          lines=['Not that one *won*'], size=72,
          rows=[{'t': 'A sidebar you can drag wider'},
                {'t': 'A place to type, and a history above it'},
                {'t': 'Usually a planning step and an execution step'},
                {'t': 'And a plain text file in your repo that drives all of it', 'hot': True}])),
 ('W6', G('matrix', kicker='read the rows, not the columns',
          lines=['Four rows *identical*'], size=62,
          cols=COLS, rows=ROWS, upto=4, dateline='the brief did not change. the prompt did not change')),
 ('W7', G('nest', kicker='and three of the four are one editor',
          lines=['Same *room*, four doors'], size=62,
          outer='Visual Studio Code', inner='Cursor · Copilot · Antigravity',
          ring=['same file tree', 'same tabs', 'same shortcuts', 'same panel'],
          caption='two forks and an extension — that is not an accident')),
 ('W8', ('F', 'B_all')),
 ('V5', G('rows', kicker='so what actually differed',
          lines=['*Furniture.* Permissions.', 'Model'], size=58,
          rows=[{'t': 'Where the panel sits and what the buttons are called'},
                {'t': 'How often it stops to ask you'},
                {'t': 'And which model was behind it'},
                {'t': 'That is the whole difference. It is smaller than the marketing', 'hot': True}])),
 ('W9', G('timeline', kicker='and this is the convergence that matters',
          lines=['They *agreed* on the file'], size=62,
          items=[{'when': 'a year ago', 't': 'every tool, its own config file', 'sub': 'its own name, its own place'},
                 {'when': 'this year', 't': 'agents.md, near-universally', 'sub': 'read without being told'},
                 {'when': 'today', 't': 'the last holdout adopted it', 'sub': 'since the course this is built from', 'hot': True}],
          caption='faster than anybody expected')),
 ('W10', G(H, kicker='which means the work you did is portable',
           lines=['Change tools next month.', 'The *file* comes with you'], size=54)),
 ('W11', G(H, kicker='so, choosing', lines=['The *boring* answer', 'is the right one'], size=68)),
 ('W12', G('rows', kicker='and every one of these beats a benchmark',
           lines=['Pick on *this*'], size=76, numbered=True,
           rows=[{'t': 'The one your team already uses — so you can ask someone'},
                 {'t': 'The one your employer already pays for'},
                 {'t': 'The one you can log into right now'},
                 {'t': 'The keyboard you already know', 'hot': True}])),
 ('V6', ('F', 'B_all')),
 ('W13', G('rows', kicker='and do not agonise, because it transfers',
           lines=['One *afternoon* to move'], size=68,
           rows=[{'t': 'Writing a brief'},
                 {'t': 'Reading a plan and its success checks'},
                 {'t': 'Watching a permission prompt'},
                 {'t': 'Checking a claim instead of believing it', 'hot': True}])),
 ('V7', G(H, kicker='which is the real reason we built it four times',
          lines=['The tool is the *least*', 'interesting variable'], size=58, trans='rise')),
 ('W14', G(H, kicker='so the course turns here', lines=['Products, done.', 'Now *technique*'], size=80, trans='push')),
 ('V8', G('rows', kicker='and none of these expire',
          lines=['What the rest of the', 'three weeks *is*'], size=58,
          rows=[{'t': 'A brief that does not need a follow-up'},
                {'t': 'Making an agent prove a fix instead of announcing one'},
                {'t': 'Working in a loop, not one long hopeful prompt'},
                {'t': 'Knowing when to stop', 'hot': True}])),
 ('W15', ('F', 'B_all')),
 ('W16', G(H, kicker='and one will launch before you finish this course',
           lines=['That is not a joke.', 'It is a *scheduling* observation'], size=52)),
 ('W17', G('grid', kicker='tomorrow — the first project, and it is the brave one',
           lines=['Day *four*'], done=3, now=3, trans='push')),
]


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
          f"film {film:.1f}s = {int(film//60)}:{int(film%60):02d}  (his 4:16)")

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
        jobs.append(['21', i, TAKES[take], round(ss, 2), round(to, 2), round(win, 3)])
    json.dump(jobs, open('/tmp/l21_jobs.json', 'w'), indent=1)

    rates.sort()
    print(f"footage slides {len(jobs)} of {len(SPEC)} = {len(jobs)/len(SPEC):.0%}")
    print(f"  rate range {rates[0][0]:.2f} ({rates[0][1]}) .. {rates[-1][0]:.2f} ({rates[-1][1]})")
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

    tsx = f'''// nocode21 — the verdict (his L21, 4:16)
//
// The payload of the whole day, and the reason `matrix` exists as a template.
// His version puts the takeaway on a fresh slide at 2:30 of a 4:16 lecture and
// asks the viewer to take it on trust. Ours fills the grid in one column per
// build across the day — upto 1 and 2 in L18, 3 in L19, 4 here — so the verdict
// reads as a summary of what they watched.
//
// The verdict is deliberately soft, per ref-L21: the four harnesses are
// structurally the same thing, the models differed so a difference proves
// little, and the choice should be made on which tool your team already uses.
// V2 and V3 put that caveat FIRST rather than burying it.
//
// One thing is newer than his lecture and better: W9's timeline. He describes
// Antigravity as the holdout that never adopted `agents.md`. It has adopted it
// since — verified in the shipped bundle and on camera in L20 — so the
// convergence, not the holdout, is the story.

import React from 'react';
import {{Deck, Slide, deckFrames}} from './kit';

export const DURS = [
{chr(10).join('  ' + ', '.join(f'{d:.3f}' for d in durs[i:i+8]) + ',' for i in range(0, len(durs), 8))}
];
export const FILES = [
{chr(10).join('  ' + ', '.join(f"'{n}'" for n in names[i:i+8]) + ',' for i in range(0, len(names), 8))}
].map((n) => `${{n}}.mp3`);
export const GAP = {gap:.3f};
export const L21_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode21/shots/${{n}}.mp4`;

const SLIDES: Slide[] = [
{chr(10).join(lines)}
];

export const Nocode21: React.FC = () => (
  <Deck slides={{SLIDES}} durs={{DURS}} voDir="nocode21/vo" files={{FILES}} gap={{GAP}} />
);
'''
    out = os.path.join(R, 'renderer/src/nocode/l21.tsx')
    open(out, 'w').write(tsx)
    print("wrote", out)
    print("wrote /tmp/l21_jobs.json")


if __name__ == '__main__':
    main()
