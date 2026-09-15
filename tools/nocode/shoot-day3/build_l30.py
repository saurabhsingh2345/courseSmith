#!/usr/bin/env python3
"""Generate renderer/src/nocode/l30.tsx — the full-stack setup (his L30, 11:45).

His is 100% screen recording. Ours is mixed, and two of his beats cannot be
filmed here at all:

  * **The Copilot settings and usage page** shows his name, avatar, plan and
    spend. A2-A4 teach the same thing — find your usage page, look at it before
    you start, set a budget — as a graphic.
  * **VS Code's Welcome screen** lists Recent folders, and his include
    `~/Desktop/enfec_subs` and `~/Desktop/self`. A take that caught it was
    deleted; every take since closes that tab before rolling.

What IS filmed is the part that matters: a genuine `git clone` from the neutral
org, the project's shape, and a slow walk down agents.md — which is the
centrepiece of the lecture and the artefact the rest of Day 5 runs on.

Continuity note: A9/A9b say the inherited board was picked NOT because it won.
L21 shipped with "the honest headline is not that one of these tools won", and an
earlier draft here said "the one that came out best", which would have quietly
contradicted it two lectures later.
"""
import json, os, subprocess, sys

R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
VO = sys.argv[1] if len(sys.argv) > 1 else "/tmp/vo30"
TARGET = None           # his L29 is 11:20; we run long on purpose
GAP_FIXED = 0.24

TAKES = {
    'C': 'L30_A_clone_s02',      # cd, pwd, the real clone, ls, .gitignore
    'B': 'L30_B_brief_s01',      # the tree, and the walk down agents.md
}

GROUPS = {
    'C_clone': ('C', [(2.0, 66.0)]),      # the whole clone take
    'B_tree':  ('B', [(2.0, 24.0)]),      # the project shape
    'B_brief': ('B', [(24.0, 124.0)]),    # the walk down the brief
}

CUT = []
H = 'head'


def G(k, **kw):
    return ('G', dict(k=k, **kw))


SPEC = [
 ('A1', G(H, kicker='day five', lines=['Today we follow', 'the *rules*'], size=104, trans='fade')),
 ('A2', G('gauge', kicker='before anything else, the boring one',
          lines=['Find your *usage* page'], size=76,
          pct=38, limit=100, label='requests used this month', limitLabel='allowance',
          caption='look at it before you start, not when it runs out')),
 ('A3', G('myth', kicker='because of how running out actually feels',
          lines=['It gets *worse*,', 'not broken'], size=62,
          wrong='you hit a wall and get a clear message',
          right='the good model stops answering and quality quietly drops',
          note='you will blame the model for an hour before you check')),
 ('A4', G(H, kicker='and if you are on a paid plan', lines=['Set a *budget*'], size=112)),
 ('A5', G(H, kicker='now, the project', lines=['Why we are *not*', 'starting from nothing'], size=62, trans='push')),
 ('A6', G('rows', kicker='every build this week began with an empty folder',
          lines=['Which is their *best* case'], size=68,
          rows=[{'t': 'No existing decisions'}, {'t': 'No existing style'},
                {'t': 'Nothing to be careful around'},
                {'t': 'They scaffold beautifully', 'hot': True}])),
 ('A7', G('rows', kicker='and real work almost never looks like that',
          lines=['There is *always*', 'something already there'], size=58,
          rows=[{'t': "Somebody else's code"},
                {'t': 'A half-finished thing from last year'},
                {'t': 'A system nobody has touched since its author left', 'hot': True}])),
 ('A8', G('myth', kicker='so this project mixes both, deliberately',
          lines=['Inherit, then *build*'], size=76,
          wrong='start clean, where the agent is strongest',
          right='inherit working code and grow a real application around it',
          note='harder, and much closer to the job you actually have')),
 ('A9', ('F', 'B_tree')),
 ('A9b', G('rows', kicker='and before anyone asks why that one',
           lines=['*Not* because it won'], size=88,
           rows=[{'t': 'We said at the end of day three that none of them won'},
                 {'t': 'Any of the four would have done this job'},
                 {'t': 'I picked the one easiest to read on camera', 'hot': True},
                 {'t': 'Which is how real decisions usually get made'}])),
 ('A10', G(H, kicker='and from here on, pretend you did not write it',
           lines=['Another team built this.', 'You were *handed* it'], size=54)),
 ('A11', G('myth', kicker='and be clear about what it actually is',
           lines=['A convincing *demo*'], size=80,
           wrong='an application that needs finishing',
           right='a front end with nothing underneath it — reload and it forgets',
           note='no back end, no database, nothing saved anywhere')),
 ('A12', G('rows', kicker='so, the mission', lines=['Turn it into', 'something *real*'], size=68,
           rows=[{'t': 'A proper front end and back end'},
                 {'t': 'A database, so it remembers'},
                 {'t': 'An API between them'},
                 {'t': 'A sign-in, so it belongs to somebody', 'hot': True}])),
 ('A13', G(H, kicker='and one more, which for once is the right feature',
           lines=['An assistant that can', '*change* the board'], size=58, accent='#f5c518')),
 ('A14', G(H, kicker='small, but real', lines=['Ten steps.', 'Testing after *each* one'], size=68, trans='rise')),
 ('C1', G(H, kicker='so let us get it onto the machine', lines=['*Clone* it'], size=124, trans='push')),
 ('C2', ('F', 'C_clone')),
 ('C3', ('F', 'C_clone')),
 ('C4', ('Fs', 'C_clone', 'if you have never done it before', ['It ought to be harder.', 'It is *not*'], 58)),
 ('C5', ('F', 'B_tree')),
 ('C6', ('F', 'B_tree')),
 ('C7', G('hierarchy', kicker='and three of them are empty',
          lines=['Rooms with the doors', 'already *labelled*'], size=54,
          nodes=[{'t': 'pm', 'depth': 0, 'kind': 'dir'},
                 {'t': 'frontend', 'depth': 1, 'kind': 'dir'},
                 {'t': 'backend', 'depth': 1, 'kind': 'dir'},
                 {'t': 'scripts', 'depth': 1, 'kind': 'dir'},
                 {'t': 'docs', 'depth': 1, 'kind': 'dir'},
                 {'t': 'agents.md', 'depth': 1, 'kind': 'agents'}],
          active=1, caption='only the first one currently does anything')),
 ('C8', G('myth', kicker='and that is deliberate',
          lines=['An empty folder', 'is an *instruction*'], size=62,
          wrong='make folders when you need them',
          right='the back end goes HERE, not wherever it feels like',
          note='costs nothing, and prevents an argument later')),
 ('C9', ('F', 'C_clone')),
 ('C10', ('F', 'C_clone')),
 ('C11', G('myth', kicker='so the secret exists and can never be uploaded',
           lines=['Not something you', '*remember* to do'], size=58,
           wrong='be careful never to commit the .env file',
           right='list it in .gitignore once, and git cannot see it',
           note='set up at the start, and then impossible to get wrong')),
 ('C12', G(H, kicker='which is why I am showing you an empty project',
           lines=['Every decision here is', 'invisible *later*'], size=54)),
 ('B1', G(H, kicker='and now the file this week has been about', lines=['The *brief*'], size=132, trans='fade')),
 ('B2', ('F', 'B_brief')),
 ('B3', ('F', 'B_brief')),
 ('B4', ('F', 'B_brief')),
 ('B5', G(H, kicker='and the line underneath it',
          lines=['Nine words that save', 'somebody a *rewrite*'], size=54, accent='#f5c518')),
 ('B6', ('F', 'B_brief')),
 ('B7', G('myth', kicker='to the half of you worried about this section',
          lines=['You do *not* need', 'to know this'], size=62,
          wrong='you must understand the stack to specify it',
          right='ask the AI what it recommends, and take that',
          note='that is a real answer and nobody will think less of you')),
 ('B8', ('F', 'B_brief')),
 ('B9', G(H, kicker='and each of those is a fence',
          lines=['A decision *written down*', 'is one it cannot re-make'], size=52)),
 ('B10', G(H, kicker='now the section I want you to copy', lines=['*Starting point*'], size=124, trans='rise')),
 ('B11', ('F', 'B_brief')),
 ('B12', G('rows', kicker='without it, the agent has to guess',
           lines=['Is this *sacred*,', 'finished, or a mistake?'], size=58,
           rows=[{'t': 'It finds a working board in the folder'},
                 {'t': 'And no statement about what that board is'},
                 {'t': 'With the paragraph, there is no guessing at all', 'hot': True}])),
 ('B13', G('myth', kicker='and a tiny thing worth more than it looks',
           lines=['I changed *one* heading'], size=76,
           wrong='## Current state',
           right='## Starting point',
           note='same information, expiry date removed')),
 ('B14', G('rows', kicker='because in three hours',
           lines=['*Current state* will', 'be quietly false'], size=58,
           rows=[{'t': 'The state will be completely different'},
                 {'t': 'The heading will still say "current"'},
                 {'t': '"Starting point" can never go stale', 'hot': True}])),
 ('B15', G(H, kicker='that is the level of care worth taking',
           lines=['Which of your words will', 'still be *true* tomorrow'], size=52)),
 ('B16', ('F', 'B_brief')),
 ('B17', ('F', 'B_brief')),
 ('B18', G('rows', kicker='and you should recognise that one',
           lines=['*Promoted* into the file'], size=76, numbered=True,
           rows=[{'t': 'Reproduce the problem'},
                 {'t': 'Prove the root cause'},
                 {'t': 'Fix it'},
                 {'t': 'Demonstrate the fix', 'hot': True}],
           stamp='two days ago you typed this by hand, in a panic')),
 ('B19', G('myth', kicker='and one line that is in nobody else\'s template',
           lines=['Permission to *push back*'], size=68,
           wrong='follow the brief exactly',
           right='if an instruction here is a bad idea, say so rather than following it into a corner',
           note='remember this one')),
 ('B20', G(H, kicker='it is going to matter',
           lines=['And not in the way', 'I *expected*'], size=68)),
 ('B21', ('F', 'B_brief')),
 ('B22', G('myth', kicker='and that last clause is load-bearing',
           lines=['The plan *survives*', 'the conversation'], size=58,
           wrong='keep the plan updated',
           right='keep it updated INCLUDING any design decisions you make',
           note='we prove this later today by throwing the conversation away')),
 ('D1', G(H, kicker='two things in the panel', lines=['Both you have seen.', 'Neither *explained*'], size=54, trans='push')),
 ('D2', G('ladder', kicker='the first is the mode',
          lines=['One dropdown *apart*'], size=80,
          items=['a version that only talks to you',
                 'a version that reads, writes and runs things'],
          pick=1, tone='#e53935')),
 ('D3', G('rows', kicker='and next to the model selector',
          lines=['*Manage models*'], size=92,
          rows=[{'t': 'Point it at a model running on your own machine'},
                {'t': 'No account, no network, no cost'},
                {'t': 'It answers the question I am asked most', 'hot': True}])),
 ('D4', G('myth', kicker='so, does any of this work without a subscription',
          lines=['Yes. Slower, and *on you*'], size=62,
          wrong='you need a paid plan to do any of this',
          right='a local model works — it needs a strong machine and setup effort',
          note='behind the big paid models, but real')),
 ('D5', ('F', 'B_brief')),
 ('D6', G('compact', kicker='we drew this on day two',
          lines=['Watch it *fill up*'], size=88,
          blocks=24, keep=10, at=0.6,
          summary='everything before this gets compressed',
          freedLabel='and your instructions were at the start',
          caption='today you watch it fill over a real build — and we act on it')),
 ('D7', G(H, kicker='this is the last lecture of setting up',
          lines=['Everything after this', 'is *building*'], size=62)),
 ('D8', G('rows', kicker='and let me say plainly what we are demonstrating',
          lines=['Not that an AI', 'can *write code*'], size=62,
          rows=[{'t': 'You have watched that all week'},
                {'t': 'Somebody who cannot write this code can still direct it'},
                {'t': 'Check it, and end up with something real', 'hot': True}])),
 ('D9', G(H, kicker='the whole promise of the course',
          lines=['It survives the next', 'four lectures, or it *does not*'], size=52, trans='rise')),
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
        jobs.append(['30', i, TAKES[take], round(ss, 2), round(to, 2), round(win, 3)])
    json.dump(jobs, open('/tmp/l30_jobs.json', 'w'), indent=1)

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

    tsx = f'''// nocode30 — the full-stack setup (his L30, 11:45)
//
// Mixed cut. Two of his beats cannot be filmed here: the Copilot settings/usage
// page shows his name, plan and spend, and VS Code's Welcome screen lists
// `~/Desktop/enfec_subs` and `~/Desktop/self` in Recent. A take that caught the
// Welcome screen was deleted; every take since closes that tab before rolling.
//
// What is filmed is what matters — a genuine `git clone` from the neutral org,
// the project's shape, and a slow walk down agents.md, which is the artefact the
// whole of Day 5 runs on.
//
// Continuity: A9b says the inherited board was picked NOT because it won. L21
// shipped with "the honest headline is not that one of these tools won", and an
// earlier draft of A9 said "the one that came out best" — which would have
// contradicted it two lectures later.

import React from 'react';
import {{Deck, Slide, deckFrames}} from './kit';

export const DURS = [
{chr(10).join('  ' + ', '.join(f'{d:.3f}' for d in durs[i:i+8]) + ',' for i in range(0, len(durs), 8))}
];
export const FILES = [
{chr(10).join('  ' + ', '.join(f"'{n}'" for n in names[i:i+8]) + ',' for i in range(0, len(names), 8))}
].map((n) => `${{n}}.mp3`);
export const GAP = {gap:.3f};
export const L30_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode30/shots/${{n}}.mp4`;

const SLIDES: Slide[] = [
{chr(10).join(lines)}
];

export const Nocode30: React.FC = () => (
  <Deck slides={{SLIDES}} durs={{DURS}} voDir="nocode30/vo" files={{FILES}} gap={{GAP}} />
);
'''
    out = os.path.join(R, 'renderer/src/nocode/l30.tsx')
    open(out, 'w').write(tsx)
    print("wrote", out)
    print("wrote /tmp/l30_jobs.json")


if __name__ == '__main__':
    main()
