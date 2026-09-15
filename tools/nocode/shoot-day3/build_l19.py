#!/usr/bin/env python3
"""Generate renderer/src/nocode/l19.tsx and the cut jobs for L19 (Claude Code).

Same generator design as build_l18.py — footage slides declare a GROUP (one or
more contiguous stretches of a take) and each group's source is shared out across
its slides in proportion to their narration windows, so chronology is preserved
and nothing has to be re-derived when a narration length changes.

Two L19-specific notes:

  * **There is a PII hole in the build take.** Claude Code ran `lsof` to find out
    what was holding port 3000, and lsof prints a USER column — his username is
    on camera. The window is excluded from the group ranges below rather than
    masked, because the panel scrolls and a fixed box cannot follow it.
  * **The model is the same family the Cursor build used** (`opus[1m]` at
    `effortLevel: high`, from his settings), which makes this the cleanest
    isolation of the harness in the whole day: same model, different wrapper.
    N12 says so on camera, with the date.
"""
import json, os, subprocess, sys

R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
VO = sys.argv[1] if len(sys.argv) > 1 else "/tmp/vo19"
TARGET = None           # his L19 is 10:22; we run long on purpose
GAP_FIXED = 0.24

TAKES = {
    'A': 'L19_A_build_s01',      # prompt, plan, the question, the whole build
    'S': 'L19_S_checks_s01',     # the five checks, by hand
}

# Ranges read off the take (1697.9s total).
#
# A_wait is the interesting one. Claude Code planned, then asked "start at Phase
# 0, or adjust anything first?" — and then sat waiting for an answer for about
# seven minutes. That is 420 seconds of a static screen. Rather than throw it
# away, a slice of it becomes the RAMP the house rules ask for: 160s of source in
# a ~13s window is a visible 12x, which is what "I know it was a second for you,
# it was fifteen minutes for me" should look like on screen.
#
# A_build has a HOLE at 855-950s. The agent ran `lsof` to find out what was
# holding port 3000, and lsof prints a USER column — his username is on camera
# at t=880, scrolled away by t=940. The panel scrolls, so a fixed mask box
# cannot follow it; the window is simply never selected.
GROUPS = {
    'A_open':  ('A', [(0.0, 20.0)]),        # the prompt being typed
    'A_plan':  ('A', [(20.0, 170.0)]),      # the plan streaming in and completing
    'A_wait':  ('A', [(170.0, 330.0)]),     # the long wait, for the ramp
    'A_ask':   ('A', [(575.0, 640.0)]),     # the question, my two words, and off it goes
    'A_build': ('A', [(655.0, 780.0), (950.0, 1075.0)]),  # PII hole at 855-950
    'A_done':  ('A', [(1550.0, 1697.0)]),   # the finish
    'S_all':   ('S', [(0.0, 117.5)]),       # the five checks, shared evenly
}

CUT = []                # nothing cut; we ship long by design

H = 'head'


def G(k, **kw):
    return ('G', dict(k=k, **kw))


MATRIX_COLS = [
    {'head': 'Cursor',      'sub': 'Opus, chosen'},
    {'head': 'Copilot',     'sub': 'free tier, Auto'},
    {'head': 'Claude Code', 'sub': 'Opus, high effort'},
    {'head': 'Antigravity', 'sub': 'next'},
]
MATRIX_ROWS = [
    {'label': 'the brief',   'cells': ['agents.md', 'agents.md', 'agents.md', '']},
    {'label': 'the prompt',  'cells': ['go ahead and plan'] * 3 + ['']},
    {'label': 'asked me',    'cells': ['never', 'eight times', 'once, before starting', '']},
    {'label': 'the plan',    'cells': ['in the panel', 'in the panel', 'written to a file', '']},
    {'label': 'five checks', 'cells': ['all five', 'all five', 'all five', '']},
]

FOUR_STEPS = [{'t': 'Reproduce the problem'},
              {'t': 'Prove you have reproduced it'},
              {'t': 'Find the root cause'},
              {'t': 'Fix it — and demonstrate the fix', 'hot': True}]

SPEC = [
 ('C1', ('Fs', 'A_open', 'build three of four', ['A command line tool', 'wearing an *editor*'], 62)),
 ('N1', G('term', kicker='the ritual, for the last time', lines=['Rename. *Clone.* Repeat'], size=62,
          title='copilot kanban, kept as a record',
          term=[{'t': 'mv kanban kanban-copilot', 'kind': 'cmd'},
                {'t': 'git clone .../kanban.git', 'kind': 'cmd'},
                {'t': 'agents.md   README.md', 'kind': 'ok'}])),
 ('N2', ('F', 'A_open')),
 ('C2', G('rows', kicker='and an honest swap', lines=['Why *this one*', 'and not the other'], size=58,
          rows=[{'t': 'The reference course uses a tool behind a paid subscription'},
                {'t': 'This one has a free path in, and more of you already have it'},
                {'t': 'It makes exactly the same point about a CLI agent in an editor', 'hot': True},
                {'t': 'If you do have the other one, run the same brief through it'}])),
 ('N3', ('F', 'A_open')),
 ('N4', G('myth', kicker='what this actually is', lines=['Not an editor. *A program*'], size=62,
          wrong='an editor with an agent bolted into the side of it',
          right='a command line program you happen to be watching through an editor',
          note='the thing doing the work does not need the editor at all')),
 ('N5', ('F', 'A_plan')),
 ('N6', ('Fs', 'A_plan', 'so today is the friendly version', ['Next week we take', 'the *training wheels* off'], 56)),
 ('C3', ('F', 'A_plan')),
 ('N7', ('F', 'A_plan')),
 ('N8', G('rows', kicker='three tools in', lines=['The furniture is', '*the same furniture*'], size=62,
          rows=[{'t': 'A panel on one side that you can drag wider'},
                {'t': 'A box to type in, and a history above it'},
                {'t': 'A mode selector and a permission dial'},
                {'t': 'And a plain text file in the repo that drives all of it', 'hot': True}])),
 ('N9', G('rows', kicker='on accounts', lines=['You can *start* here', 'for nothing'], size=64,
          rows=[{'t': 'The reference course tells you to skip its version of this lecture'},
                {'t': 'Because the tool it uses needs a paid plan'},
                {'t': 'This one has a free tier, and if you already pay, you have it', 'hot': True}])),
 ('N10', G('ladder', kicker='and you can predict the options by now',
           lines=['Four modes, *one dial*'], size=64,
           items=['Plan — talk, do not touch anything',
                  'Manual — ask me before every action',
                  'Edit automatically — change files, ask before commands',
                  'Auto — full access, ask nothing'],
           pick=3, tone='#e53935')),
 ('N11', ('F', 'A_plan')),
 ('N12', G('rows', kicker='naming it, with the date on the frame',
           lines=['*Claude Opus*, high effort'], size=68,
           rows=[{'t': 'A one-million-token context window'},
                 {'t': 'Thinking budget turned up, deliberately'},
                 {'t': 'The same model family that ran the Cursor build', 'hot': True}],
           stamp='28 august 2026 — check this yourself, it will have moved')),
 ('C4', ('F', 'A_plan')),
 ('C5', ('F', 'A_plan')),
 ('C6', G('curve', kicker='and the obvious answer is wrong',
          lines=['More thinking is *not* better'], size=62,
          bands=[{'from': 0.0, 'to': 0.35, 'label': 'thinking helps', 'tone': '#3b82f6'},
                 {'from': 0.35, 'to': 0.62, 'label': 'best value', 'tone': '#f5c518'},
                 {'from': 0.62, 'to': 1.0, 'label': 'agonising', 'tone': '#e53935'}],
          mark=0.48, yLabel='quality', xLabel='thinking budget',
          caption='the curve has a peak in it, not a slope')),
 ('C7', G('rows', kicker='so where the peak sits depends on the job',
          lines=['Turn it *up* or leave it'], size=64,
          rows=[{'t': 'A gnarly bug in code you cannot read — turn it up'},
                {'t': 'A small, well specified build like this — let it go'},
                {'t': 'You will learn the difference from how long you sit waiting', 'hot': True}])),
 ('N13', ('F', 'A_plan')),
 ('C8', ('F', 'A_plan')),
 ('S1', ('F', 'A_plan')),
 ('S2', ('F', 'A_plan')),
 ('S3', ('F', 'A_plan')),
 ('S4', G('myth', kicker='and this is the sentence that earned my trust',
          lines=['Checked, not *remembered*'], size=62,
          wrong='these are the current versions, as far as I recall',
          right='versions verified live against the registry today, not from memory',
          note='a model\'s memory of what is current is frozen at its training date')),
 ('S5', G('rows', kicker='it also told me what it decided',
          lines=['A judgement, *declared*'], size=64,
          rows=[{'t': 'Picked the stable drag library over the pre-release one'},
                {'t': 'And over the popular one that is no longer maintained'},
                {'t': 'Kept our palette rather than inventing its own', 'hot': True},
                {'t': 'None of that was asked for. All of it was said out loud'}])),
 ('S6', ('F', 'A_ask')),
 ('S7', ('F', 'A_ask')),
 ('S8', G('rows', kicker='same file, same four words, three readings',
          lines=['What *plan first* meant'], size=62, numbered=True,
          rows=[{'t': 'Cursor — plan, then carry straight on'},
                {'t': 'Copilot — plan, then carry straight on'},
                {'t': 'Claude Code — plan, then stop and check with me', 'hot': True}],
          stamp='and the cheapest moment to change your mind is before it starts')),
 ('S9', ('F', 'A_ask')),
 ('N14', ('F', 'A_wait')),
 ('N15', G('rows', kicker='and do not let the speed-up hide it',
           lines=['The loop is *ten minutes* long'], size=62,
           rows=[{'t': 'This is the point where you go and do something else'},
                 {'t': 'That is not a flaw, it is the shape of the work now'},
                 {'t': 'You write the brief carefully BECAUSE the loop is slow', 'hot': True}])),
 ('N16', ('F', 'A_build')),
 ('N17', G('compact', kicker='what happens when the memory fills',
           lines=['It throws away *the beginning*'], size=58,
           blocks=24, keep=9, at=0.62,
           summary='everything before this is compressed into a summary',
           freedLabel='and your original instructions were at the start',
           caption='which is why a long session starts making decisions you ruled out')),
 ('N18', G('editor', kicker='and the strongest argument for the file',
           lines=['A message is forgotten.', 'A *file* is re-read'], size=54,
           file='agents.md', start=1,
           rows=[{'t': '# Kanban — project brief', 'kind': 'h1'},
                 {'t': '', 'kind': 'text'},
                 {'t': 'read again at the start of every turn', 'kind': 'important'},
                 {'t': 'not once, at the beginning, and then hoped for', 'kind': 'important'}],
           focus=[2, 3])),
 ('F1', ('F', 'A_build')),
 ('F2', ('F', 'A_build')),
 ('F3', G('rows', kicker='remember these from last lecture', lines=['The *four steps*'], size=84,
          numbered=True, rows=FOUR_STEPS)),
 ('F4', ('F', 'A_build')),
 ('F5', G('flow', kicker='and this is what proving a cause looks like',
          lines=['A *chain*, not a guess'], size=58,
          nodes=[{'t': 'commit mid-drag', 'sub': 'the move is applied early'},
                 {'t': 'layout re-renders', 'sub': 'under the cursor', 'tone': 'yellow'},
                 {'t': 'keyboard sensor re-collides', 'sub': 'with the shifted layout'},
                 {'t': 'card bounces back', 'sub': 'the symptom you saw', 'tone': 'red'}],
          caption='every link is checkable — that is the difference from a guess')),
 ('F6', ('F', 'A_build')),
 ('F7', G('rows', kicker='against the card from last lecture',
          lines=['*Four* for four'], size=76, numbered=True,
          rows=[{'t': 'Reproduced it — in a real browser'},
                {'t': 'Proved the cause — a chain, not a location'},
                {'t': 'Fixed it — commit the move at the end, not mid-drag'},
                {'t': 'Demonstrated it — re-ran the failing check', 'hot': True}],
          stamp='nobody asked it to')),
 ('F8', ('Fs', 'A_build', 'so the instruction is not a workaround',
         ['It is a description of', 'what *good* looks like'], 54)),
 ('F9', ('F', 'A_build')),
 ('G1', ('F', 'A_done')),
 ('G2', G('rows', kicker='and that is the whole app',
          lines=['*962 lines*, four files'], size=72,
          rows=[{'t': 'The production build passes'},
                {'t': 'The type check passes, the linter is clean'},
                {'t': '27 unit tests pass, and 10 browser tests in real Chromium'},
                {'t': '38 success criteria, ticked off against the file it wrote', 'hot': True}],
          stamp='every gate named before it started')),
 ('G3', ('F', 'A_done')),
 ('G4', ('F', 'A_done')),
 ('G5', ('F', 'A_done')),
 ('D4', G('myth', kicker='the other half of the discipline',
          lines=['Broken, or *by design*'], size=64,
          wrong='two arrow keys do the same thing — that is a bug, fix it',
          right='the library walks the cards in a line, not a grid — documented',
          note='five minutes of checking before you demand a fix')),
 ('G6', ('Fs', 'A_done', 'that is step two doing its job',
          ['A tool willing to say', '*I was wrong* about that'], 54)),
 ('D5', G('rows', kicker='and both failures cost the same afternoon',
          lines=['Two ways to *waste a day*'], size=62,
          rows=[{'t': 'Believing a fix that was never demonstrated'},
                {'t': 'Demanding a fix for something that was never broken', 'hot': True}])),
 ('O1', ('F', 'S_all')),
 ('O2', ('F', 'S_all')),
 ('O3', G('myth', kicker='one design decision differs, and it is better',
          lines=['Where the *button* goes'], size=64,
          wrong='one Add button at the top, then asking which column you meant',
          right='Add a card inside each column, where the card is going to land',
          note='the brief chose neither — two tools read one sentence differently')),
 ('O4', G('rows', kicker='which is worth noticing', lines=['The brief did *not* say'], size=68,
          rows=[{'t': 'It said: add a card. Four words'},
                {'t': 'Everything about where, and how, was decided for you'},
                {'t': 'Leave the decisions you do not care about unspecified', 'hot': True}])),
 ('O5', ('F', 'S_all')),
 ('N19', ('F', 'S_all')),
 ('O6', G('rows', kicker='third time today', lines=['The *same five*, in order'], size=62,
          numbered=True,
          rows=[{'t': 'Drag a card between columns'},
                {'t': 'Reorder a card inside a column'},
                {'t': 'Delete a card'},
                {'t': 'Rename a column'},
                {'t': 'Add a card, with a description'}])),
 ('N20', ('F', 'S_all')),
 ('C9', ('F', 'A_done')),
 ('C10', ('Fs', 'A_done', 'and now the honest part',
          ['I could not build', 'this *by hand*'], 62)),
 ('C11', G('rows', kicker='which is the whole point of the course',
           lines=['*Four* jobs. None of them', 'require reading code'], size=54,
           numbered=True,
           rows=[{'t': 'Decide what gets built'},
                 {'t': 'State it precisely enough to be acted on'},
                 {'t': 'Check whether the thing in front of you does it'},
                 {'t': 'Send it back when it does not', 'hot': True}])),
 ('C12', G(H, kicker='every one of those is a judgement',
           lines=['Not one of them is', '*reading the code*'], size=62)),
 ('N21', ('F', 'A_done')),
 ('N22', G('rows', kicker='and notice what nobody did',
           lines=['Nobody wrote *any* of it'], size=64,
           rows=[{'t': 'No code written by a person'},
                 {'t': 'No bug debugged by a person'},
                 {'t': 'No library, folder layout or state decision made by a person'},
                 {'t': 'All of it decided by something reading a file we wrote', 'hot': True}])),
 ('O7', G('matrix', kicker='three down, one to go', lines=['The grid so *far*'], size=58,
          cols=MATRIX_COLS, rows=MATRIX_ROWS, upto=3, dateline='28 august 2026')),
 ('C13', G('rows', kicker='and the caveat, which is the useful part',
           lines=['This was an *easy* job'], size=64,
           rows=[{'t': 'Small, clean, and described in one page'},
                 {'t': 'All four tools will look impressive on it'},
                 {'t': 'A real system has requirements that contradict each other'},
                 {'t': 'It is day three of three weeks for a reason', 'hot': True}])),
 ('N23', ('Fs', 'A_done', 'one to go, and it is the odd one out',
          ['The *holdout*'], 104)),
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
          f"film {film:.1f}s = {int(film//60)}:{int(film%60):02d}  (his 10:22)")

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
        jobs.append(['19', i, TAKES[take], round(ss, 2), round(to, 2), round(win, 3)])
    json.dump(jobs, open('/tmp/l19_jobs.json', 'w'), indent=1)

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
           if (r[0] > 1.55 or r[0] < 0.30) and r[1] != 'N14']
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

    tsx = f'''// nocode19 — the Claude Code build (his L19, 10:22)
//
// Build three of four. Claude Code stands in for Codex, which he approved:
// Codex needs a paid ChatGPT plan, this has a free path in, more of our audience
// already has it, and it makes the identical point — a command line agent living
// inside the editor, which is the bridge into next week's CLI material.
//
// It ran **opus[1m] at effortLevel high**, from his own settings — the same model
// family the Cursor build used. That makes this the cleanest isolation of the
// harness in the whole day: same model, different wrapper. N12 says so on camera
// with the date on the frame.
//
// Four things this build did that the reference lecture's Codex run never shows,
// and all four are in the narration:
//   * it wrote the plan to a FILE (PLAN.md), eight phases each with a "gate",
//     rather than leaving it in a panel that scrolls away;
//   * "versions verified live against npm today, not from memory" — it checked
//     current library versions instead of trusting training data, unprompted;
//   * it STOPPED AND ASKED before executing, with permissions wide open on Auto,
//     because the brief says plan first and it read that as plan-then-check.
//     Neither Cursor nor Copilot did that off the same file;
//   * during its own browser verification it found a real bug and ran exactly
//     the four steps L18 teaches — reproduced it, proved a causal chain, fixed
//     it, re-ran the failing check — then separately tried a second fix, saw no
//     change, and STOPPED rather than claim it. F1-F9 and G1-G6 carry both.
//
// PII: the take has a HOLE at 855-950s. The agent ran `lsof` to see what held
// port 3000 and lsof prints a USER column, so his username is on camera at
// t=880. The panel scrolls, so the window is excluded from footage selection
// rather than masked. cut18.py masks the VS Code activity-bar avatar as usual,
// and skips that mask on the browser take.

import React from 'react';
import {{Deck, Slide, deckFrames}} from './kit';

export const DURS = [
{chr(10).join('  ' + ', '.join(f'{d:.3f}' for d in durs[i:i+8]) + ',' for i in range(0, len(durs), 8))}
];
export const FILES = [
{chr(10).join('  ' + ', '.join(f"'{n}'" for n in names[i:i+8]) + ',' for i in range(0, len(names), 8))}
].map((n) => `${{n}}.mp3`);
export const GAP = {gap:.3f};
export const L19_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode19/shots/${{n}}.mp4`;

const SLIDES: Slide[] = [
{chr(10).join(lines)}
];

export const Nocode19: React.FC = () => (
  <Deck slides={{SLIDES}} durs={{DURS}} voDir="nocode19/vo" files={{FILES}} gap={{GAP}} />
);
'''
    out = os.path.join(R, 'renderer/src/nocode/l19.tsx')
    open(out, 'w').write(tsx)
    print("wrote", out)
    print("wrote /tmp/l19_jobs.json")


if __name__ == '__main__':
    main()
