#!/usr/bin/env python3
"""Generate renderer/src/nocode/l29.tsx — Web Apps 101 (his L29, 11:20).

**No footage.** His lecture ends with a live Docker install and a tour of his own
Docker Desktop. We cannot film ours: this machine holds 25 images, 53 volumes and
7 running containers from real client work, and the Containers/Images lists would
publish those project names. So the Docker half is animated instead, and C13/C14
were rewritten to describe a fresh install rather than imply we are looking at
mine.

That makes this a slides-only lecture like L03 and L28, which is also what the
reference is for its first two thirds.

ref-L29 calls this "the most important lecture in Section 1 for a no-code
audience", and the two beats to get right are:
  * **secrets live on the back end** (A9-A11) — the front end runs on somebody
    else's computer, so a key in it is a public notice;
  * **the LLM slop look** (B10-B14) — the purple gradient, three thin-line icons,
    enormous headline, rounded "get started" button. Naming it is the setup for
    "name the fault and name the standard", which is the week's whole method.
"""
import json, os, subprocess, sys

R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
VO = sys.argv[1] if len(sys.argv) > 1 else "/tmp/vo29"
TARGET = None           # his L29 is 11:20; we run long on purpose
GAP_FIXED = 0.24

TAKES = {}
GROUPS = {}
CUT = []
H = 'head'


def G(k, **kw):
    return ('G', dict(k=k, **kw))


SPEC = [
 ('A1', G(H, kicker='before we build anything', lines=['The *vocabulary*'], size=124, trans='fade')),
 ('A2', G(H, kicker='and I mean this', lines=['If you know this,', 'put me on *2x*'], size=68)),
 ('A3', G('myth', kicker='what all of this actually is', lines=['A *web* application'], size=76,
          wrong='an app you install from a store',
          right='an address you open in a browser',
          note='no installer, no app store — that is the whole definition')),
 ('A4', G(H, kicker='and every one of them is', lines=['*Two halves*', 'talking to each other'], size=62, trans='rise')),
 ('A5', G('nest', kicker='the half you can see', lines=['The *front end*'], size=84,
          outer='your browser, on your computer', inner='the front end',
          ring=['the page', 'how it looks', 'what happens when you click'])),
 ('A6', G('rows', kicker='and the three names you will see forever',
          lines=['*HTML. CSS. JavaScript*'], size=68, numbered=True,
          rows=[{'t': 'HTML — the page itself'},
                {'t': 'CSS — how the page looks'},
                {'t': 'JavaScript — what happens when you drag a card', 'hot': True}])),
 ('A7', G('nest', kicker='the half you cannot', lines=['The *back end*'], size=88,
          outer='a server, somewhere else', inner='the back end',
          caption='it sent you the page in the first place')),
 ('A8', G('rows', kicker='and it does the serious things',
          lines=['What the *back end* is for'], size=64,
          rows=[{'t': 'Reading and writing the database'},
                {'t': 'Calling a language model'},
                {'t': "Talking to other companies' services"},
                {'t': 'Anything that must be remembered, or paid for, or protected', 'hot': True}])),
 ('A9', G(H, kicker='the most important sentence in this lecture',
          lines=['Your secrets live in', 'the *back end*'], size=68, accent='#e53935')),
 ('A10', G('myth', kicker='because of where the front end runs',
           lines=['Not a secret. A *notice*'], size=72,
           wrong='put the key in the front end, it is only used once',
           right='the front end runs on their computer — anyone can read it',
           note='you find out when the bill arrives')),
 ('A11', G('hierarchy', kicker='which is what this file is for',
           lines=['*.env*, and never checked in'], size=76,
           nodes=[{'t': 'pm', 'depth': 0, 'kind': 'dir'},
                  {'t': '.env', 'depth': 1, 'kind': 'agents'},
                  {'t': '.gitignore', 'depth': 1, 'kind': 'file'},
                  {'t': 'frontend', 'depth': 1, 'kind': 'dir'},
                  {'t': 'backend', 'depth': 1, 'kind': 'dir'}],
           active=1, caption='listed inside .gitignore, so git never sees it')),
 ('A12', G(H, kicker='two halves, two machines', lines=['So how do they *talk*'], size=96)),
 ('A13', G('myth', kicker='and the word is less impressive than it sounds',
           lines=['An *API*'], size=132,
           wrong='some complicated protocol you have to learn',
           right='a fixed list of messages the back end agrees to answer',
           note='the front end sends one. the back end answers.')),
 ('A14', G('flow', kicker='so, dragging one card', lines=['The *round trip*'], size=76,
           nodes=[{'t': 'you drag a card', 'sub': 'in the browser'},
                  {'t': '"this card moved"', 'sub': 'the API call', 'tone': 'yellow'},
                  {'t': 'written to the database', 'sub': 'on the server'},
                  {'t': '"done"', 'sub': 'and the board is right tomorrow', 'tone': 'yellow'}],
           caption='that round trip is the whole reason anything is remembered')),
 ('D4', G(H, kicker='and it explains something you have felt',
          lines=['Every call is a *journey*'], size=80)),
 ('D5', G('rows', kicker='usually hundredths of a second — but real',
          lines=['Why *spinners* exist'], size=76,
          rows=[{'t': 'The two halves are genuinely apart'},
                {'t': 'The distance is real, and sometimes slow'},
                {'t': 'A good app tells you it is thinking', 'hot': True}])),
 ('D6', G('myth', kicker='and this is a decision somebody makes',
          lines=['Responsive, or *frozen*'], size=72,
          wrong='wait for the server, then update the button',
          right='say "saving" instantly, let the journey happen behind it',
          note='if nobody decides this, the app feels broken while working perfectly')),
 ('A15', G(H, kicker='now put the week on that diagram', lines=['Where each project *sat*'], size=76, trans='push')),
 ('D1', G('nest', kicker='day one', lines=['The shooter:', '*all* front end'], size=68,
          outer='your browser', inner='the whole game',
          caption='no server, no account, nothing saved')),
 ('D2', G('rows', kicker='which is not a criticism',
          lines=['Front end only is *fine*'], size=72,
          rows=[{'t': 'A calculator'}, {'t': 'A drawing tool'}, {'t': 'A game'},
                {'t': 'Nothing remembered, nothing secret — you need no back end', 'hot': True}])),
 ('D3', G('myth', kicker='so here is the line, and carry it with you',
          lines=['When you need the *other half*'], size=68,
          wrong='every app needs a server',
          right='you need one to REMEMBER, or to keep a SECRET',
          note='that tells you how big a job is before you start it')),
 ('A16', G('rows', kicker='day three, four times over',
           lines=['The board: *all* front end'], size=68,
           rows=[{'t': 'No back end at all'},
                 {'t': 'Which is why it forgot everything on reload'},
                 {'t': 'It was never storing anything — there was nowhere to store it', 'hot': True}])),
 ('A17', G('rows', kicker='yesterday', lines=['A back end — in *JavaScript*'], size=68,
           rows=[{'t': 'Both halves in the same language'},
                 {'t': 'A completely normal way to build'},
                 {'t': 'Today we do it differently, and it is worth saying why'}])),
 ('A18', G('myth', kicker='today', lines=['JavaScript front.', '*Python* back'], size=68,
           wrong='python is the better language',
           right='when the other end is a language model, python is where that world lives',
           note='and the help you find online will assume it')),
 ('B1', G(H, kicker='a short history, because the names never get explained',
          lines=['The *front end*,', 'in order'], size=64, trans='rise')),
 ('B2', G('rows', kicker='first', lines=['Hand-written *pages*'], size=84,
          rows=[{'t': 'You wrote the page, how it looked, and a bit of behaviour'},
                {'t': 'A small library if you wanted something fancy'},
                {'t': 'Still how much of the web works, and nothing wrong with it'}])),
 ('B3', G('myth', kicker='then, one idea changed it', lines=['*Components*'], size=124,
          wrong='write a page, then update bits of it by hand',
          right='write a card, write a column — and they redraw themselves',
          note='when the data changes, the component follows')),
 ('B4', G('rows', kicker='four names, roughly one job',
          lines=['React. Vue.', 'Angular. *Svelte*'], size=62,
          rows=[{'t': 'They do broadly the same thing'},
                {'t': 'Ours uses React'},
                {'t': 'Because the front end we inherited uses React', 'hot': True},
                {'t': 'Which is usually how you pick one'}])),
 ('B5', G('myth', kicker='and out of that, a pattern with a bad name',
          lines=['The *single page* app'], size=72,
          wrong='every click loads a whole new page',
          right='loaded once — then pieces fetch what they need and update in place',
          note='SPA. that is all it means.')),
 ('B6', G(H, kicker='which is why a good web app', lines=['Feels like a *program*,', 'not a document'], size=62)),
 ('B7', G('stack', kicker='and on top sit the higher-level frameworks',
          lines=['*Next.js*'], size=112,
          layers=[{'t': 'your pages', 'sub': 'the part you write', 'h': 44, 'tone': 'yellow'},
                  {'t': 'routing', 'sub': 'which address shows what', 'h': 56},
                  {'t': 'data fetching', 'sub': 'getting what the page needs', 'h': 56},
                  {'t': 'server or browser rendering', 'sub': 'decided per page', 'h': 62}],
          foot='the decisions nobody wants to make twice')),
 ('B8', G(H, kicker='and now the honest part', lines=['I am not *strong* at this'], size=88, trans='push')),
 ('B9', G('rows', kicker='so I will not pretend otherwise',
          lines=['In front of an app', 'I did not *hand write*'], size=58,
          rows=[{'t': 'Building front ends by hand is not something I am good at'},
                {'t': 'It is exactly the work coding agents made possible for me'},
                {'t': 'Which is the whole premise of this course', 'hot': True}])),
 ('B10', G(H, kicker='but there is a catch, and you have seen it',
           lines=['Ask for a web page,', 'get *the same* web page'], size=58)),
 ('B11', G('rows', kicker='you will not be able to unsee this',
           lines=['The *slop* look'], size=104,
           rows=[{'t': 'A dark background with a purple gradient sliding across it'},
                 {'t': 'Three feature cards, three thin-line icons, same icon set'},
                 {'t': 'An enormous headline, a smaller subheading'},
                 {'t': 'A rounded button that says "Get started"', 'hot': True}])),
 ('B12', G('myth', kicker='and the problem is not that it is ugly',
           lines=['It is *anonymous*'], size=88,
           wrong='this looks bad',
           right='this looks like nobody decided anything',
           note='it looks like every other thing generated the same way')),
 ('B13', G('rows', kicker='and this is where you add value — code or no code',
           lines=['Not a *coding* skill'], size=80,
           rows=[{'t': 'What the person needs to see first'},
                 {'t': 'What can safely be shown later'},
                 {'t': 'What should not be on the screen at all'},
                 {'t': 'A thinking skill. The model does not have it.', 'hot': True}])),
 ('B14', G('myth', kicker='so push back on the first draft',
           lines=['Name the fault.', 'Name the *standard*'], size=62,
           wrong='make it better',
           right='this is the same purple gradient every generated site has — our brief has a palette, use it',
           note='we established yesterday what vague feedback buys you')),
 ('C1', G(H, kicker='one more word, and it frightens people off', lines=['*Docker*'], size=140, trans='fade')),
 ('C2', G('nest', kicker='and it is simpler than its reputation',
          lines=['A computer *inside*', 'your computer'], size=62,
          outer='your machine', inner='a small, sealed one',
          ring=['its own copy of everything', 'cannot see your files', 'cannot touch the rest'])),
 ('C3', G(H, kicker='two reasons people use it', lines=['And *both* matter', 'to you specifically'], size=62)),
 ('C4', G('rows', kicker='one — isolation', lines=['What happens in the box,', 'stays in the *box*'], size=58,
          rows=[{'t': 'It cannot install something odd on your laptop'},
                {'t': 'It cannot overwrite your files'},
                {'t': 'It cannot fight with a Python you installed two years ago', 'hot': True}])),
 ('C5', G(H, kicker='and notice how directly that speaks to this week',
          lines=['A sealed box is the', 'honest answer to the *nervousness*'], size=52, accent='#f5c518')),
 ('C6', G('rows', kicker='two — portability', lines=['Build once.', 'Runs *anywhere*'], size=72,
          rows=[{'t': 'The box is described by a file, so it can be rebuilt anywhere'},
                {'t': 'Your machine, a colleague\'s, a rented server'},
                {'t': '"It usually just works" — rarely true of software', 'hot': True}])),
 ('C7', G(H, kicker='three words, then we are done', lines=['Dockerfile.', 'Image. *Container*'], size=62, trans='rise')),
 ('C8', G('editor', kicker='one — the recipe', lines=['A *Dockerfile*'], size=104,
          file='Dockerfile', start=1,
          rows=[{'t': 'FROM python:3.12', 'kind': 'code'},
                {'t': 'COPY . /app', 'kind': 'code'},
                {'t': 'RUN uv sync', 'kind': 'code'},
                {'t': 'CMD ["uvicorn", "main:app"]', 'kind': 'code'}],
          caption='you can read that without knowing anything. it is a list of instructions.')),
 ('C9', G('myth', kicker='two — follow the recipe', lines=['An *image*'], size=132,
          wrong='the running thing',
          right='the finished snapshot — the cake, baked and frozen',
          note='it does not do anything on its own. it sits there, ready.')),
 ('C10', G('flow', kicker='three — start it', lines=['A *container*'], size=124,
           nodes=[{'t': 'image', 'sub': 'one snapshot', 'tone': 'yellow'},
                  {'t': 'container', 'sub': 'running'},
                  {'t': 'container', 'sub': 'running'},
                  {'t': 'container', 'sub': 'running'}],
           caption='one image can start as many as you like — and they do not know about each other')),
 ('C11', G(H, kicker='recipe, cake, cake being eaten', lines=['That is *Docker*'], size=112)),
 ('C12', G('rows', kicker='and installing it is the dull part',
           lines=['Take *every* default'], size=80,
           rows=[{'t': 'docker.com — download Docker Desktop for your machine'},
                 {'t': 'On Windows it asks about WSL. The answer is yes.'},
                 {'t': 'It may want a restart. Give it one.'}])),
 ('C13', G('rows', kicker='and three things in the left rail are the whole tour',
           lines=['Containers. Images.', '*Volumes*'], size=58, numbered=True,
           rows=[{'t': 'Containers — what is alive right now'},
                 {'t': 'Images — what is built and ready to start'},
                 {'t': 'Volumes — sealed storage that survives the container', 'hot': True}],
           stamp='a container thrown away forgets everything inside it')),
 ('C14', G(H, kicker='on a fresh install all three are empty',
           lines=['They will not', '*stay* empty'], size=76)),
 ('D7', G('rows', kicker='and two thirds of you knew all of that',
           lines=['That is genuinely', 'the *vocabulary*'], size=62,
           rows=[{'t': 'Front end — in the browser'},
                 {'t': 'Back end — on a server, where the secrets live'},
                 {'t': 'API — the fixed list of messages between them'},
                 {'t': 'Container — a sealed box to run it in', 'hot': True}])),
 ('D8', G(H, kicker='none of it is difficult',
          lines=['The people who know it', 'forgot they had to *learn* it'], size=54, trans='rise')),
 ('C15', G(H, kicker='next, we start building', lines=['And by the *rules*', 'this time'], size=72, trans='push')),
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
        jobs.append(['29', i, TAKES[take], round(ss, 2), round(to, 2), round(win, 3)])
    json.dump(jobs, open('/tmp/l29_jobs.json', 'w'), indent=1)

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

    tsx = f'''// nocode29 — Web Apps 101 (his L29, 11:20)
//
// Slides only. His lecture ends with a live Docker install and a tour of his own
// Docker Desktop; we cannot film ours, because this machine holds 25 images, 53
// volumes and 7 running containers from real client work and the Containers and
// Images lists would publish those project names. The Docker half is animated
// instead, and C13/C14 describe a fresh install rather than implying we are
// looking at mine.
//
// ref-L29 calls this the most important lecture in Section 1 for a no-code
// audience. Two beats carry it: secrets live on the back end (A9-A11), because
// the front end runs on somebody else's computer; and the LLM slop look
// (B10-B14) — the purple gradient, three thin-line icons, the enormous headline,
// the rounded "Get started" button — which sets up "name the fault and name the
// standard".

import React from 'react';
import {{Deck, Slide, deckFrames}} from './kit';

export const DURS = [
{chr(10).join('  ' + ', '.join(f'{d:.3f}' for d in durs[i:i+8]) + ',' for i in range(0, len(durs), 8))}
];
export const FILES = [
{chr(10).join('  ' + ', '.join(f"'{n}'" for n in names[i:i+8]) + ',' for i in range(0, len(names), 8))}
].map((n) => `${{n}}.mp3`);
export const GAP = {gap:.3f};
export const L29_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode29/shots/${{n}}.mp4`;

const SLIDES: Slide[] = [
{chr(10).join(lines)}
];

export const Nocode29: React.FC = () => (
  <Deck slides={{SLIDES}} durs={{DURS}} voDir="nocode29/vo" files={{FILES}} gap={{GAP}} />
);
'''
    out = os.path.join(R, 'renderer/src/nocode/l29.tsx')
    open(out, 'w').write(tsx)
    print("wrote", out)
    print("wrote /tmp/l29_jobs.json")


if __name__ == '__main__':
    main()
