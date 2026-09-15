#!/usr/bin/env python3
"""Generate renderer/src/nocode/l31.tsx — planning and scaffolding (his L31, 11:43).

The disciplined counterpart to Day 4's YOLO, and the lecture that demonstrates
the method the whole week argues for.

**Two of his set pieces did not happen to us, and the narration says so.**
  * His agent asked three questions off the opening prompt. Ours asked none — it
    read both files and said the plan was clear enough to proceed. C1-C6 turn
    that into the honest version of the lesson: the point was never to collect
    questions, it was to find out whether the brief was clear, and silence is a
    pass. It cost fifteen seconds to learn.
  * His agent silently swapped `uv` for a bare `requirements.txt`. Ours produced
    `pyproject.toml` and used uv properly. D10-D12 keep the general shape of the
    lesson — instructions do not always get followed, and you find out by
    watching what APPEARS in the project, not by being told — while being clear
    it did not happen here.

Not staging either failure is the same call L18 made when our Copilot build
succeeded where his broke.
"""
import json, os, subprocess, sys

R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
VO = sys.argv[1] if len(sys.argv) > 1 else "/tmp/vo31"
TARGET = None           # his L29 is 11:20; we run long on purpose
GAP_FIXED = 0.24

TAKES = {
    'Q': 'L31_A_plan_s01',       # the opening prompt, and it answering with no questions
    'P': 'L31_B_part1_s01',      # part 1: enriching the plan
    'S': 'L31_C_part2_s01',      # part 2: docker, fastapi, the scripts
}

GROUPS = {
    'Q_ask':  ('Q', [(8.0, 60.0)]),      # typing the prompt
    'Q_ans':  ('Q', [(60.0, 220.0)]),    # its review, and "questions: none"
    'P_plan': ('P', [(10.0, 330.0)]),    # writing the enriched plan
    'S_scaf': ('S', [(10.0, 200.0)]),    # scaffolding
    'S_done': ('S', [(200.0, 345.0)]),   # what it built, and how to test it
}

CUT = []
H = 'head'


def G(k, **kw):
    return ('G', dict(k=k, **kw))


TEN = [{'t': 'Plan'}, {'t': 'Scaffolding'}, {'t': 'Serve the front end'},
       {'t': 'Sign in'}, {'t': 'Database model'}, {'t': 'Back end routes'},
       {'t': 'Wire them together'}, {'t': 'Reach the AI'},
       {'t': 'Extend the plumbing'}, {'t': 'The assistant', 'hot': True}]

SPEC = [
 ('A1', G(H, kicker='yesterday we broke every rule', lines=['Today we *follow* them'], size=96, trans='fade')),
 ('A2', G('rows', kicker='ten parts, in a file, written by me',
          lines=['Not written by the *AI*'], size=76, numbered=True, rows=TEN)),
 ('A3', G('myth', kicker='and that was a choice', lines=['Whose *judgement*?'], size=88,
          wrong='ask the agent to write the plan — very common, and it will be reasonable',
          right='a plan is a set of decisions, and I have opinions about these',
          note='if you have opinions, write your own')),
 ('A4', G(H, kicker='and if you genuinely do not',
          lines=['Ask for a draft.', 'Then *argue* with it'], size=68)),
 ('A5', ('Fs', 'Q_ask', 'part one sounds like a joke', ['*Planning* the plan'], 108)),
 ('A6', G('rows', kicker='part two', lines=['One page.', 'One *API call*'], size=84,
          rows=[{'t': 'The container'}, {'t': 'The back end'},
                {'t': 'The start and stop scripts'},
                {'t': 'Nothing clever. A hello world that runs.', 'hot': True}])),
 ('A7', G('rows', kicker='parts three to five', lines=['Serve. Sign in.', '*Design* the data'], size=62,
          rows=[{'t': 'Three — serve the board we inherited'},
                {'t': 'Four — the sign-in'},
                {'t': 'Five — the database design, and STOP for approval', 'hot': True}])),
 ('A8', G('rows', kicker='and six to ten', lines=['Routes. Wire.', 'Then the *AI*'], size=62,
          rows=[{'t': 'Six — routes that read and change the board'},
                {'t': 'Seven — wire the halves so a move persists'},
                {'t': 'Eight — prove we can reach the AI at all'},
                {'t': 'Nine and ten — the board, a question, and the assistant', 'hot': True}])),
 ('A9', G(H, kicker='now the reason those particular ten', lines=['The *sizing*', 'is the lesson'], size=72, trans='rise')),
 ('A10', G('myth', kicker='and it is one rule', lines=['Small enough to *dig into*'], size=68,
           wrong='break it into the smallest possible pieces',
           right='each step small enough that if it fails, you know how to start looking',
           note='not trivial. not so big that failure leaves you staring.')),
 ('A11', G('rows', kicker='and it scales with you, not with the software',
           lines=['The right step size', 'is a fact about *you*'], size=54,
           rows=[{'t': 'Know less — use smaller steps'},
                 {'t': 'Know more — use bigger ones'},
                 {'t': 'Nobody else can pick it for you', 'hot': True}])),
 ('A12', G(H, kicker='which answers the obvious objection',
           lines=['Why not just tell it', 'to do *all ten*'], size=62)),
 ('A13', G('myth', kicker='because yesterday was a toy and this has a database',
           lines=['It will go off *somewhere*'], size=68,
           wrong='give it everything and let it run — it worked yesterday',
           right='ten steps at this size goes off the rails in the middle',
           note='and you will not know where, because you were not watching')),
 ('B1', G(H, kicker='before the first prompt, one technique', lines=['*One sentence*'], size=124, trans='push')),
 ('B2', ('Fs', 'Q_ask', 'and here it is', ['Let me know if you have questions.', 'Do *no work* yet'], 50)),
 ('B3', G('rows', kicker='because of when it is cheapest to be wrong',
          lines=['Questions cost *seconds*'], size=76,
          rows=[{'t': 'An agent about to misunderstand you reveals it in its questions'},
                {'t': 'The questions cost seconds'},
                {'t': 'The misunderstanding costs an hour', 'hot': True}])),
 ('B4', G(H, kicker='and nothing has been written yet',
          lines=['You are still holding', 'the *steering wheel*'], size=58)),
 ('B5', G('myth', kicker='the other half comes later', lines=['*Confident*, not done'], size=84,
          wrong='is it done?',
          right='let me know when you are confident',
          note='done is a claim about the work. confident is a claim about itself.')),
 ('C1', ('F', 'Q_ans')),
 ('C2', G(H, kicker='which is not what I expected',
          lines=['It asked *nothing*'], size=120)),
 ('C3', G('myth', kicker='and that is the technique working',
          lines=['Silence is a *result*'], size=88,
          wrong='no questions means it did not engage',
          right='no questions means the brief was clear',
          note='the point was never to collect questions')),
 ('C4', G(H, kicker='and it cost fifteen seconds to find out',
          lines=['The cheapest information', 'you will buy *all day*'], size=54)),
 ('C5', G('rows', kicker='and it would have been just as useful the other way',
          lines=['Three confused questions'], size=76,
          rows=[{'t': 'Would have told me the brief was vague'},
                {'t': 'And I would rather learn that now'},
                {'t': 'Than after it has built the wrong thing twice', 'hot': True}])),
 ('C6', G(H, kicker='so keep the habit even when it returns nothing',
          lines=['You are testing whether', 'you were *clear*'], size=54)),
 ('C7', G('rows', kicker='and notice the small grey line under its answer',
          lines=['It *names* the model'], size=80,
          rows=[{'t': 'And what it charged you'},
                {'t': 'This tool picks a model per task'},
                {'t': 'So that line will not always say the same thing'},
                {'t': 'Read it rather than assuming', 'hot': True}])),
 ('D1', ('F', 'P_plan')),
 ('D2', ('F', 'P_plan')),
 ('D3', G(H, kicker='and this is the moment with the most leverage today',
          lines=['Right now the plan', 'is *words*'], size=68, accent='#f5c518')),
 ('D4', G('rows', kicker='so read it properly — the shape, not every sub-step',
          lines=['What to *look* for'], size=80,
          rows=[{'t': 'Are the parts in a sensible order'},
                {'t': 'Does anything assume something that does not exist yet'},
                {'t': 'Is there a success criterion you could not check yourself', 'hot': True}])),
 ('D5', G('myth', kicker='and that last one is the real test',
          lines=['A wish, or a *criterion*'], size=68,
          wrong='the board works correctly',
          right='log out, log back in, and the card is still in the second column',
          note='if you cannot verify it yourself, you will be taking its word later')),
 ('B6', ('Fs', 'P_plan', 'and the most boring habit in software', ['*Check in*', 'after every part'], 84)),
 ('E7', G(H, kicker='and I am naming the three commands', lines=['Look. Include. *Save*'], size=92, trans='rise')),
 ('E8', G('term', kicker='because nobody ever does', lines=['The *three* commands'], size=76,
          title='after every part',
          term=[{'t': 'git status', 'kind': 'cmd'},
                {'t': '  what changed since the last save point', 'kind': 'out'},
                {'t': 'git add .', 'kind': 'cmd'},
                {'t': '  include all of it', 'kind': 'out'},
                {'t': 'git commit -m "part 2 complete"', 'kind': 'cmd'},
                {'t': '  the save point exists', 'kind': 'ok'}])),
 ('E9', G('rows', kicker='and the message matters more than people think',
          lines=['You will *read* it later'], size=76,
          rows=[{'t': 'When you are hunting for the last version that worked'},
                {'t': '"fixes" tells you nothing'},
                {'t': '"part 7 built, some drag and drop bugs" tells you everything', 'hot': True}])),
 ('E10', G(H, kicker='and this is saved on your own machine',
           lines=['It has not gone *anywhere*'], size=76)),
 ('E11', G('rows', kicker='and if any of that is unfamiliar',
           lines=['*Ask* the agent'], size=104,
           rows=[{'t': 'Ask it what git is'},
                 {'t': 'Ask it what a commit is'},
                 {'t': 'Ask it to explain what it just did'}])),
 ('E12', G(H, kicker='which is the general point of this whole course',
           lines=['The thing you direct is', 'also your best *teacher*'], size=52)),
 ('D6', ('F', 'S_scaf')),
 ('D7', G('rows', kicker='because part two is where you find out',
          lines=['Wrong *now*, with seven files'], size=62,
          rows=[{'t': 'Does the container build'},
                {'t': 'Do the scripts run on your machine'},
                {'t': 'Is the thing in the browser the thing in the box'},
                {'t': 'Better to know now than later with seventy files', 'hot': True}])),
 ('D8', G('myth', kicker='and here is the prompt to steal',
          lines=['Not *is it done*'], size=96,
          wrong='is it done?',
          right='tell me how I can test this myself, and let me know when you are confident',
          note='it hands you the verification instead of keeping it')),
 ('D9', ('F', 'S_done')),
 ('E1', ('F', 'S_scaf')),
 ('E2', G('rows', kicker='if you can read the command, read it',
          lines=['Most are *ordinary*'], size=88,
          rows=[{'t': 'Make a folder'}, {'t': 'Install a package'},
                {'t': 'Start a server'},
                {'t': 'You get a feel for the normal ones within an hour', 'hot': True}])),
 ('E3', G('rows', kicker='and if you cannot read it, three options',
          lines=['None of them are *guessing*'], size=68, numbered=True,
          rows=[{'t': 'Ask it what the command does and why it needs to run'},
                {'t': 'Paste it into a different AI and ask that one'},
                {'t': 'Say no, and ask it to explain itself first', 'hot': True}])),
 ('E4', G(H, kicker='and that third one is underrated',
          lines=['Refusing is the most', '*normal* thing a manager does'], size=52)),
 ('E5', G('myth', kicker='so let me reframe this for anyone it stresses',
          lines=['Shown, in *advance*'], size=88,
          wrong='you are being asked to audit software you cannot read',
          right='you are being shown everything before it happens, by something asking permission',
          note='better than most developers had ten years ago')),
 ('E6', G(H, kicker='it is a learning opportunity',
          lines=['Wearing the costume', 'of an *interruption*'], size=58)),
 ('D10', G(H, kicker='one thing to watch for that did NOT happen to us',
           lines=['And I am flagging it', '*anyway*'], size=62, trans='push')),
 ('D11', G('myth', kicker='our brief named a package manager, and it used it',
           lines=['His agent *quietly* swapped it'], size=58,
           wrong='pyproject.toml, as the brief specified',
           right='a bare requirements.txt appeared, which nobody asked for',
           note='he only noticed because a FILE appeared that should not exist')),
 ('D12', G('rows', kicker='so carry the shape even though it went right here',
           lines=['Look at what *shows up*'], size=76,
           rows=[{'t': 'Instructions do not always get followed'},
                 {'t': 'And it will not tell you when one was not'},
                 {'t': 'You find out from what appears in the project', 'hot': True}])),
 ('B7', G(H, kicker='so, checkpoints', lines=['Not because *this* part', 'was risky'], size=68)),
 ('D13', G(H, kicker='but because the habit has to be automatic',
           lines=['*Before* you reach', 'the part that is'], size=64)),
 ('B8', ('Fs', 'S_done', 'now watch what happens next', ['Two things go *wrong*'], 100)),
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
