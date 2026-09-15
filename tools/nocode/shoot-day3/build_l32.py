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
    'P': 'L32_A_p34_s01',       # 687s — parts 3+4: the two errors, the build, the tests
    'Q': 'L32_B_p34b_s01',      # 312s — the run completing, and the commits appearing
    'K': 'L32_C_checks_s01',    # 187s — test-it-yourself, the criteria audit, the why
    'W': 'L32_D_web_s01',       # 107s — sign in, log out, sign in again
    'G': 'L32_E_git_s01',       # 100s — the history, the swept commit, the three commands
    'F': 'L32_F_fix_s01',       # 330s — the blank page, proved and fixed
}

GROUPS = {
    'P_perm': ('P', [(150.0, 250.0)]),    # the start script refused
    'P_port': ('P', [(250.0, 360.0)]),    # address already in use
    'P_p3':   ('P', [(360.0, 520.0)]),    # serving the front end
    'P_p4':   ('P', [(520.0, 640.0)]),    # the sign-in, and the test it wrote
    'Q_done': ('Q', [(40.0, 220.0)]),     # the run finishing and committing
    'K_test': ('K', [(30.0, 110.0)]),     # how to test it yourself
    'K_crit': ('K', [(110.0, 185.0)]),    # the criteria audit, and the why
    'W_form': ('W', [(2.0, 26.0)]),       # the sign-in screen
    'W_in':   ('W', [(34.0, 58.0)]),      # signing in, and the board
    'W_out':  ('W', [(68.0, 92.0)]),     # out and back in
    'G_log':  ('G', [(2.0, 40.0)]),       # the history, and what is inside a commit
    'G_fix':  ('G', [(40.0, 98.0)]),      # status, add, commit
    'F_blank':('F', [(4.0, 60.0)]),       # describing it
    'F_dig':  ('F', [(60.0, 92.0)]),     # proving the cause
    'F_ok':   ('F', [(300.0, 327.0)]),    # confirmed the way a browser would
}

CUT = []
H = 'head'


def G(k, **kw):
    return ('G', dict(k=k, **kw))


SPEC = [
 ('A1', G(H, kicker='last time ended on a promise', lines=['Three things', 'go *wrong*'], size=92, trans='fade')),
 ('A2', ('Fs', 'P_perm', 'the first one', ['Permission *denied*'], 104)),
 ('A3', G('myth', kicker='and it sounds worse than it is', lines=['A *flag* on a file'], size=88,
          wrong='the script is broken, or my machine is misconfigured',
          right='a file carries a flag saying whether it may be run, and nobody set it',
          note='the file was written a minute ago')),
 ('A4', ('F', 'P_perm')),
 ('A5', G('rows', kicker='and this is the pattern to internalise',
          lines=['Hand over the *error*'], size=88,
          rows=[{'t': 'You do not need to know the fix'},
                {'t': 'You need to hand it the error, exactly as it appeared'},
                {'t': 'It has read a million of these', 'hot': True}])),
 ('A6', ('Fs', 'P_port', 'second thing', ['Address already *in use*'], 84)),
 ('A7', G('myth', kicker='a port is a numbered door', lines=['One program, one *doorway*'], size=68,
          wrong='address already in use — something is wrong with the code',
          right='something else is standing in that doorway and never left',
          note='it is a problem with the room, not the building')),
 ('A8', ('F', 'P_port')),
 ('A9', G(H, kicker='two errors, two recoveries, no help from me',
          lines=['The third one is', 'a different *animal*'], size=62, trans='push')),
 ('B1', G(H, kicker='before I let it move on', lines=['The prompt', 'to *steal*'], size=104)),
 ('B2', ('Fs', 'K_test', 'here it is', ['Tell me how I can', '*test* this myself'], 76)),
 ('B3', G('myth', kicker='and look at what that sentence does',
          lines=['Not *is it done*'], size=96,
          wrong='is part four finished?',
          right='tell me how I can check part four myself',
          note='the verification stops living inside the thing being verified')),
 ('B4', G('rows', kicker='and for this audience it matters more, not less',
          lines=['A currency you can *spend*'], size=68,
          rows=[{'t': 'I cannot read its tests'},
                {'t': 'I can open a browser and see whether a page loads'},
                {'t': 'So I want the answer in the second form', 'hot': True}])),
 ('B5', ('F', 'K_test')),
 ('B6', ('Fs', 'K_crit', 'and then the check nobody does', ['Are the *success criteria*', 'actually met?'], 60)),
 ('B7', G('rows', kicker='and it only works because the plan wrote them down',
          lines=['A vague plan', 'cannot be *audited*'], size=62,
          rows=[{'t': 'A plan with a checklist can be'},
                {'t': 'And the audit costs one sentence', 'hot': True}])),
 ('B8', G(H, kicker='there is a second version I use constantly',
          lines=['Ask it to *explain*', 'a decision'], size=72)),
 ('B9', G('myth', kicker='and the wording changes the answer', lines=['Explain, not *justify*'], size=84,
          wrong='why did you do it that way?',
          right='explain the reasoning behind that choice',
          note='one invites a defence, the other invites the reasoning')),
 ('B10', G('rows', kicker='because it makes dozens of choices you never see',
           lines=['Most are *fine*'], size=96,
           rows=[{'t': 'Which library'}, {'t': 'Which structure'}, {'t': 'Which shortcut'},
                 {'t': 'And some are it doing what works rather than what you asked', 'hot': True}])),
 ('B11', ('F', 'K_crit')),
 ('B12', G(H, kicker='so when a file appears you did not expect', lines=['*Ask* why'], size=124)),
 ('B13', G('rows', kicker='and you can do all of that without reading code',
           lines=['Interrogating the', '*reasoning*'], size=68,
           rows=[{'t': 'Not auditing the implementation'},
                 {'t': 'Which is a job you can actually do', 'hot': True}])),
 ('C1', ('Fs', 'P_p3', 'part three', ['Serve the board', 'from the *back end*'], 72)),
 ('C2', G('rows', kicker='and it looks like nothing, which is why it is a whole part',
          lines=['Two things that', 'do not *know* each other'], size=54,
          rows=[{'t': 'A front end that draws a board'},
                {'t': 'A back end that answers on a port'},
                {'t': 'Part three is the introduction', 'hot': True}])),
 ('C3', G(H, kicker='and after this', lines=['*One* address', 'serves the whole thing'], size=68)),
 ('C4', ('F', 'P_p3')),
 ('C5', G(H, kicker='and that phrase, out of the box', lines=['Worth unpacking', '*once*'], size=76)),
 ('C6', G('rows', kicker='we said it on day three', lines=['The same *object*'], size=96,
          rows=[{'t': 'A box with the application and everything it needs, sealed together'},
                {'t': 'It behaves the same here, on your machine, and on a server'},
                {'t': 'So this is not a rehearsal of production. It is production.', 'hot': True}])),
 ('C7', ('F', 'P_p3')),
 ('D1', ('Fs', 'Q_done', 'now the diffs', ['Green *added*', 'Red *removed*'], 92)),
 ('D2', ('F', 'Q_done')),
 ('D3', G('myth', kicker='and I am not going to pretend', lines=['Read *every line*?'], size=92,
          wrong='you should review every single line before accepting it',
          right='nobody shipping at this pace does that, including me',
          note='a rule you will not follow is not a rule')),
 ('D4', G('rows', kicker='so here is the one I actually use',
          lines=['Bulk, then *slow down*'], size=76,
          rows=[{'t': 'Accept in bulk while the work is ordinary'},
                {'t': 'Look properly when it touches something you would not want wrong', 'hot': True}])),
 ('D5', G('rows', kicker='and that second list is short and specific',
          lines=['Where the *afternoons* go'], size=68, numbered=True,
          rows=[{'t': 'Money'}, {'t': 'Anything with a password in it'},
                {'t': 'Anything that deletes'},
                {'t': 'The first time we plug in the AI', 'hot': True}])),
 ('D6', G('pct', kicker='because risk is not spread evenly across a diff',
          pct=90, sub='plumbing that either works or obviously does not — '
                      'the other ten percent is where the afternoons go')),
 ('D7', G(H, kicker='and the phrase I keep coming back to',
          lines=['Watch it like a *hawk*.', 'Not read it like a lawyer'], size=54, accent='#f5c518')),
 ('E1', ('Fs', 'P_p4', 'part four', ['Sign in. One user.', 'A way *out*'], 80)),
 ('E2', ('F', 'P_p4')),
 ('E3', G('rows', kicker='and it writes a test for it, unprompted',
          lines=['A *good* test'], size=100,
          rows=[{'t': 'Wrong password, refused'},
                {'t': 'Right password, and the board'},
                {'t': 'Two things that matter, and nothing else', 'hot': True}])),
 ('E4', ('Fs', 'Q_done', 'and it reports back', ['Part four complete.', 'And it *checked*'], 68)),
 ('X1', ('F', 'W_form')),
 ('X2', G(H, kicker='and this is what I get', lines=['*Nothing*'], size=150, trans='fade')),
 ('X3', G(H, kicker='so let us slow this right down',
          lines=['How I found it matters', 'more than *what* it was'], size=52)),
 ('X4', G('rows', kicker='because its checks passed', lines=['By its *measurement*,', 'part four worked'], size=58,
          rows=[{'t': 'It asked for the page from the command line'},
                {'t': 'It got back the sign-in form, complete and correct'},
                {'t': 'Every check it could run, it ran, and they passed', 'hot': True}])),
 ('X5', G('myth', kicker='and here is the gap', lines=['It cannot *open* a browser'], size=72,
          wrong='it tested the page, so the page works',
          right='it tested the page the only way it can, and the fault only appears in a browser',
          note='the one place it cannot look is the one place it shows')),
 ('X6', G(H, kicker='which is the shape of nearly every bad afternoon',
          lines=['Not that it *lied*.', 'It measured the wrong thing'], size=52, accent='#f5c518')),
 ('X7', ('Fs', 'F_blank', 'so I describe it the way day three taught', ['Did. Expected.', '*Happened*'], 84)),
 ('X8', G('rows', kicker='and then the two facts that make it strange',
          lines=['Where the views', '*disagree*'], size=68, numbered=True,
          rows=[{'t': 'Blank in the browser'},
                {'t': 'Correct from the command line'},
                {'t': 'And asked to unpack it as a browser would, the command line fails too', 'hot': True}])),
 ('X9', G(H, kicker='that last one is the whole gift',
          lines=['I do not know the bug.', 'I know *where* it lives'], size=56)),
 ('X10', G('rows', kicker='and two clauses are there because of what just happened',
           lines=['Prove it. Then check', 'it the *right* way'], size=58,
           rows=[{'t': 'Prove the cause, do not guess at it'},
                 {'t': 'Confirm the fix the way a browser would, not the way the command line does', 'hot': True}])),
 ('X11', ('Fs', 'F_dig', 'and here is what it was', ['*Small*, as usual'], 116)),
 ('X12', G('flow', kicker='the back end fetches the page and passes it on',
           lines=['Unpacked on the way *in*'], size=68,
           nodes=[{'t': 'the front end', 'sub': 'sends it compressed'},
                 {'t': 'the library between', 'sub': 'unpacks it on the way in'},
                 {'t': 'our code', 'sub': 'copies the ORIGINAL labels across'},
                 {'t': 'your browser', 'sub': 'reads the label and gives up'}],
          caption='unpacked parcel, sealed label')),
 ('X13', G('myth', kicker='so the parcel is open and the label still says sealed',
           lines=['The browser gives up.', '*Silently*'], size=58,
           wrong='the label says compressed, so unpack it',
           right='it is already unpacked, and unpacking it again fails',
           note='no error on screen. just nothing.')),
 ('X14', G(H, kicker='and the command line did not care',
           lines=['It was never asked', 'to *unpack* anything'], size=60)),
 ('X15', G(H, kicker='one label', lines=['An application,', 'or a blank *rectangle*'], size=62, accent='#f5c518')),
 ('X16', ('F', 'F_ok')),
 ('X17', G('myth', kicker='and the rule I would carve into the desk',
           lines=['Two different *claims*'], size=88,
           wrong='the check passed, so it works',
           right='its check passing is not the same as it working',
           note='only one of those is yours to make')),
 ('E5', ('F', 'W_in')),
 ('E6', ('F', 'W_out')),
 ('E7', G('rows', kicker='and be precise about what that proves',
          lines=['*Less* than it looks'], size=100,
          rows=[{'t': 'The sign-in works'},
                {'t': 'It remembers me between visits'}])),
 ('E8', G('myth', kicker='and here is what it does not prove',
          lines=['There is no *database* yet'], size=68,
          wrong='the board survived, so the board is saved',
          right='the board is still the one the front end draws for itself',
          note='move a card and it would not survive a reload')),
 ('E9', G(H, kicker='that is part five, and it is next',
          lines=['A working demo is the', 'easiest place to *overbelieve*'], size=52)),
 ('F1', G(H, kicker='now the checkpoints', lines=['Not the beat I expected', 'to be *recording*'], size=60, trans='rise')),
 ('F2', G('term', kicker='last time I made a point of it', lines=['And I wrote it *again*'], size=84,
          title='in the prompt, in plain English',
          term=[{'t': 'commit to git after each part,', 'kind': 'cmd'},
                {'t': 'with a clear message', 'kind': 'cmd'}])),
 ('F3', ('Fs', 'G_log', 'so did it?', ['It *did*'], 132)),
 ('F4', G(H, kicker='but do not stop at the fact that commits exist',
          lines=['Look at what is', '*inside* them'], size=68)),
 ('F5', ('F', 'G_log')),
 ('F6', G('rows', kicker='and none of that is part three',
          lines=['That is part *two*'], size=104,
          rows=[{'t': 'The Dockerfile'}, {'t': 'The back end'}, {'t': 'Both of the scripts'},
                {'t': 'Swept in, because part two never got a checkpoint of its own', 'hot': True}])),
 ('F7', G('myth', kicker='does it break anything? no',
          lines=['The work is safe.', 'The *return point* is gone'], size=54,
          wrong='everything is committed, so everything is fine',
          right='there is no longer a point to go back to at the end of part two',
          note='two parts welded into one')),
 ('F8', ('Fs', 'G_fix', 'and one more, which I would have missed', ['In *no commit* at all'], 92)),
 ('F9', G(H, kicker='small. part of the back end. everything runs',
          lines=['A checkpoint I trusted', 'is not *complete*'], size=54)),
 ('F10', G('myth', kicker='so the lesson is not that it ignored me',
           lines=['It *followed* the instruction'], size=68,
           wrong='it skipped the checkpoints I asked for',
           right='it made them, and the history still has a gap in it',
           note='and the only reason I know is that I read it')),
 ('F10b', G(H, kicker='the same shape I flagged last lecture',
            lines=['One lecture *later*', 'than I expected'], size=60)),
 ('F10c', ('F', 'G_fix')),
 ('F10d', G('rows', kicker='and one clarification that tripped me up for years',
            lines=['On this machine.', '*Nowhere* else'], size=68,
            rows=[{'t': 'Nothing has been uploaded'},
                  {'t': 'There is no website involved'},
                  {'t': 'It is a save point in a folder on this laptop', 'hot': True}])),
 ('F11', G(H, kicker='somewhere else is a separate decision',
           lines=['I am not protecting', 'against *fire*'], size=64)),
 ('F12', G('rows', kicker='everything so far has been repeatable',
           lines=['Twenty minutes', 'from the *plan*'], size=68,
           rows=[{'t': 'If I lost all of it, I could make it again'}])),
 ('F13', G('myth', kicker='and the database is not', lines=['*Before* the risk'], size=104,
           wrong='commit when you have finished something',
           right='commit before the part that could hurt',
           note='not after you find out it was risky')),
 ('G1', G('rows', kicker='so where we are', lines=['Four of *ten*'], size=112, numbered=True,
          rows=[{'t': 'The box builds'}, {'t': 'The back end answers'},
                {'t': 'The board is served from it'},
                {'t': 'And you have to sign in to see it'}])),
 ('G2', G(H, kicker='and nothing clever has happened yet',
          lines=['By *design*'], size=132)),
 ('G3', ('Fs', 'Q_done', 'next, the database', ['The only part that', '*stops* and asks'], 68)),
 ('G4', G(H, kicker='and one thing to carry with you',
          lines=['All of it was', '*recovery*'], size=88, trans='rise')),
 ('G5', G('myth', kicker='a permission flag, a busy port, a blank page',
          lines=['Notice, not *write*'], size=96,
          wrong='I needed to know how to fix those',
          right='I needed to notice they were wrong',
          note='not one of them required me to write code')),
 ('G6', G(H, kicker='that is the job', lines=['And harder to *fake*', 'than programming'], size=58)),
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
        jobs.append(['32', i, TAKES[take], round(ss, 2), round(to, 2), round(win, 3)])
    json.dump(jobs, open('/tmp/l32_jobs.json', 'w'), indent=1)

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
export const L32_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode32/shots/${{n}}.mp4`;

const SLIDES: Slide[] = [
{chr(10).join(lines)}
];

export const Nocode32: React.FC = () => (
  <Deck slides={{SLIDES}} durs={{DURS}} voDir="nocode32/vo" files={{FILES}} gap={{GAP}} />
);
'''
    out = os.path.join(R, 'renderer/src/nocode/l32.tsx')
    open(out, 'w').write(tsx)
    print("wrote", out)
    print("wrote /tmp/l32_jobs.json")


if __name__ == '__main__':
    main()
