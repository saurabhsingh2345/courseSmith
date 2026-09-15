#!/usr/bin/env python3
"""Generate renderer/src/nocode/l18.tsx and the cut jobs for L18.

Why a generator instead of hand-writing the deck. L18 is 97 slides and about
73% of them are footage, which means 71 spans to pick out of three takes. Hand
picking those is where the timing errors came from on lesson one, so instead the
footage slides are declared as members of a GROUP — a contiguous range of one
take — and each group's range is divided across its slides in proportion to
their narration windows. Chronology is preserved automatically, the whole range
gets used, and nothing has to be re-derived if a narration length changes.

The five checks are the exception: those spans are pinned by hand, read off a
contact sheet, because a check that starts halfway through its own drag teaches
nothing.

Rate is reported per slide. A rate below 1 means the span is longer than its
window and gets sped up, which is exactly what the static stretches need — the
plan sitting still for ninety seconds becomes text visibly streaming past.
"""
import json, os, subprocess, sys

R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
VO = sys.argv[1] if len(sys.argv) > 1 else "/tmp/vo18"
TARGET = None           # his L18 is 14:44; we run long on purpose now
GAP_FIXED = 0.24        # set by taste, since there is no runtime to hit

TAKES = {
    'P': 'L18_P_build_s01',      # prompt, plan, implementation start, first stall
    'R': 'L18_R_build_s01',      # the build: approvals, diffs, phases, the report
    'S': 'L18_S_checks_s01',     # the five checks, by hand
}

# a group is (take, from, to) — a contiguous stretch of one take, shared out
# across every slide that names it, in order, weighted by narration length
# A group is (take, [(from, to), ...]) — one or more contiguous stretches of a
# single take, shared out across every slide that names it, in order, weighted by
# narration length.
#
# R_build has a HOLE in it, and the hole is not an aesthetic choice. Between
# roughly 95s and 155s the panel shows an npm error whose last line reads
# `/Users/<username>/.npm/_logs/...`, and a dev-server banner printing the
# machine's LAN address. Both break the no-PII rule, the text scrolls so a fixed
# mask box cannot cover it reliably, and there is plenty of other footage — so
# that window is simply never selected.
GROUPS = {
    'P_open':  ('P', [(0.0, 70.0)]),      # the prompt typed, and it reading the brief
    'P_plan':  ('P', [(70.0, 192.0)]),    # the six-phase plan on screen
    'P_stall': ('P', [(192.0, 330.0)]),   # implementation starts, then the stall
    'R_build': ('R', [(0.0, 92.0), (158.0, 232.0)]),   # PII hole at 92-158s
    'R_done':  ('R', [(232.0, 342.0)]),   # the final report, model + credits line
}

H = 'head'


def G(k, **kw):
    return ('G', dict(k=k, **kw))


CHECKS = ['loads with\ndata', 'drag between\ncolumns', 'reorder in\na column',
          'delete a\ncard', 'rename a\ncolumn', 'add with a\ndescription']

MATRIX_COLS = [
    {'head': 'Cursor',   'sub': 'frontier, chosen'},
    {'head': 'Copilot',  'sub': 'free tier, Auto'},
    {'head': 'Claude Code', 'sub': 'next'},
    {'head': 'Antigravity', 'sub': 'after that'},
]
MATRIX_ROWS = [
    {'label': 'the brief',   'cells': ['agents.md', 'agents.md', '', '']},
    {'label': 'the prompt',  'cells': ['go ahead and plan', 'go ahead and plan', '', '']},
    {'label': 'asked me',    'cells': ['never', 'eight times', '', '']},
    {'label': 'five checks', 'cells': ['all five', 'all five', '', '']},
]

# ---------------------------------------------------------------- the deck
# ('F', group) footage from a group | ('Fx', take, ss, to) pinned footage
# ('G', spec)  a graphic
SPEC = [
 ('01', ('Fs', 'P_open', 'build two of four', ['The one you','*already have*'], 90)),
 ('M1', G('term', kicker='first, the ritual', lines=['Rename it. *Clone it again*'], size=62,
          title='cursor kanban, kept as a record',
          term=[{'t': 'mv kanban kanban-cursor', 'kind': 'cmd'},
                {'t': 'git clone .../kanban.git', 'kind': 'cmd'},
                {'t': 'Cloning into kanban...', 'kind': 'out'},
                {'t': 'agents.md   README.md', 'kind': 'ok'}])),
 ('M2', ('F', 'P_open')),
 ('M3', G('hierarchy', kicker='four builds, four records', lines=['One brief. *Four folders*'], size=62,
          nodes=[{'t': 'projects', 'depth': 0, 'kind': 'dir'},
                 {'t': 'kanban-cursor', 'depth': 1, 'kind': 'dir'},
                 {'t': 'kanban-copilot', 'depth': 1, 'kind': 'dir'},
                 {'t': 'kanban-claude', 'depth': 1, 'kind': 'dir'},
                 {'t': 'kanban-antigravity', 'depth': 1, 'kind': 'dir'},
                 {'t': 'kanban', 'depth': 1, 'kind': 'dir'},
                 {'t': 'agents.md', 'depth': 2, 'kind': 'agents'}],
          active=6, caption='the fifth one is always the fresh clone')),
 ('02', G('nest', kicker='where it actually lives', lines=['Not an app. *An extension*'], size=62,
          outer='Visual Studio Code', inner='Copilot',
          caption='there is no Copilot in your dock')),
 ('M4', ('F', 'P_open')),
 ('M5', ('F', 'P_plan')),
 ('M6', G('nest', kicker='and here is the good news', lines=['Cursor *is* VS Code'], size=68,
          outer='Visual Studio Code', inner='Cursor',
          ring=['same file tree', 'same tabs', 'same shortcuts', 'same panel'],
          caption='built from its source, not merely inspired by it')),
 ('M7', ('F', 'P_plan')),
 ('M8', G('rows', kicker='and it happens again today', lines=['*Three* of four', 'are the same editor'], size=64,
          rows=[{'t': 'Cursor — a fork of VS Code'},
                {'t': 'Copilot — an extension inside VS Code'},
                {'t': 'Antigravity — a fork of VS Code'},
                {'t': 'Learn the room once, and you know three of them', 'hot': True}])),
 ('M9', G('gauge', kicker='and the money', lines=['A *real* free tier'], size=68,
          pct=4, limit=100, label='this build: under 4 credits',
          limitLabel='monthly allowance',
          caption='not a seven day trial')),
 ('M10', ('F', 'P_plan')),
 ('03', G('matrix', kicker='hold this in your head all day', lines=['Only *one thing* changes'], size=58,
          cols=MATRIX_COLS, rows=MATRIX_ROWS, upto=1,
          dateline='28 august 2026')),
 ('04', G('hierarchy', kicker='so, a fresh clone', lines=['*One file.* Nothing else'], size=68,
          nodes=[{'t': 'kanban', 'depth': 0, 'kind': 'dir'},
                 {'t': 'agents.md', 'depth': 1, 'kind': 'agents'},
                 {'t': 'README.md', 'depth': 1, 'kind': 'file'}],
          active=1)),
 ('05', ('F', 'P_plan')),
 ('M11', ('F', 'P_plan')),
 ('M12', G('ladder', kicker='the dropdown that matters', lines=['Chatbot, or *agent*'], size=68,
           items=['Ask — it talks about your code',
                  'Edit — it changes the file you point at',
                  'Agent — it reads, writes and runs things'],
           pick=2, tone='#f5c518')),
 ('06', G('myth', kicker='one difference, named early', lines=['Where the *planning* lives'], size=62,
          wrong='a Plan mode you switch into, as a separate thing',
          right='planning folded into the agent, driven by the brief',
          note='same output, different furniture')),
 ('M13', ('F', 'P_plan')),
 ('M14', ('F', 'P_plan')),
 ('T1', ('F', 'R_done')),
 ('M15', ('F', 'R_done')),
 ('M16', G('meter', kicker='read the grey line', lines=['*3.6 credits*, and it told me'], size=68,
           pct=4, fill='spent on this whole build', rest='left this month',
           caption='the only place it is completely straight with you')),
 ('07', ('F', 'P_open')),
 ('08', ('F', 'P_open')),
 ('M18', ('F', 'P_open')),
 ('M19', ('F', 'P_plan')),
 ('M20', ('F', 'P_plan')),
 ('M21', ('F', 'P_plan')),
 ('M22', G('myth', kicker='read the last line of each phase', lines=['A list, or a *commitment*'], size=62,
           wrong='Add unit tests',
           right='Success: core state behaviour is covered and tests pass',
           note='one you can hold it to later')),
 ('M23', G('editor', kicker='and it wrote them because we asked', lines=['*Nine words*, two lectures ago'], size=58,
           file='agents.md', start=44,
           rows=[{'t': '## Strategy', 'kind': 'h2'},
                 {'t': '', 'kind': 'text'},
                 {'t': '1. Write a plan first, broken into phases, with', 'kind': 'important'},
                 {'t': '   **success criteria for each phase** that can', 'kind': 'important'},
                 {'t': '   be checked off.', 'kind': 'important'}],
           focus=[2, 4])),
 ('M24', ('F', 'P_plan')),
 ('M25', ('F', 'P_stall')),
 ('R2', ('F', 'P_stall')),
 ('X1', ('F', 'P_stall')),
 ('X2', ('F', 'P_stall')),
 ('X3', G('ladder', kicker='the trade you are making', lines=['Interruption *is* visibility'], size=64,
          items=['ask once per session — fast, and you see nothing',
                 'ask once per command — slow, and you see everything',
                 'never ask — fastest, and you find out afterwards'],
          pick=1, tone='#3b82f6')),
 ('R3', ('F', 'P_stall')),
 ('M26', ('F', 'P_stall')),
 ('M27', G('stack', kicker='what that command really does', lines=['You are running *strangers* code'], size=58,
           layers=[{'t': 'your app', 'sub': 'the part you asked for', 'h': 46, 'tone': 'yellow'},
                   {'t': 'the framework', 'sub': 'thousands of files you did not write', 'h': 66},
                   {'t': 'its dependencies', 'sub': 'and their dependencies', 'h': 86},
                   {'t': 'published by people you cannot name', 'sub': 'downloaded on demand', 'h': 60, 'tone': 'red'}],
           foot='normal, universal, and the largest hole in the boat')),
 ('M28', G('rows', kicker='and it has actually happened', lines=['*Supply chain*, plainly'], size=68,
           rows=[{'t': 'One popular package gets hostile code added to it'},
                 {'t': 'Everybody who installs it runs that code'},
                 {'t': 'Not theoretical — it has hit packages with millions of users', 'hot': True}])),
 ('M29', G('myth', kicker='so where should a tool stop you', lines=['*This* is the right line'], size=64,
           wrong='stop me before every file edit, which I will switch off by lunchtime',
           right='stop me on the one command that downloads and runs strangers code',
           note='and do not let a blanket yes override that one')),
 ('M30', ('F', 'P_stall')),
 ('R4', ('F', 'R_build')),
 ('R5', ('F', 'R_build')),
 ('M31', G('loops', kicker='the honest rhythm of a build', lines=['Work. *Ask.* Yes. Work'], size=68,
           inner=['work', 'ask', 'yes'], passes=8,
           label='eight times', caption='nobody puts this bit in a demo')),
 ('M32', ('F', 'R_build')),
 ('M33', ('F', 'R_build')),
 ('M34', G(H, kicker='and if it wanders', lines=['*That* is your moment.', 'Not later'], size=76, accent='#e53935')),
 ('R6', ('F', 'R_build')),
 ('M35', ('F', 'R_build')),
 ('M36', G('rows', kicker='because nobody explains this word', lines=['What a *test* is'], size=76,
           rows=[{'t': 'A small piece of code that checks another piece of code'},
                 {'t': 'Move a card from here to there. Now ask: is it there'},
                 {'t': 'Write it once, run it a thousand times in two seconds', 'hot': True}])),
 ('M37', ('F', 'R_build')),
 ('M38', ('F', 'R_build')),
 ('M39', ('F', 'R_build')),
 ('M40', ('F', 'R_build')),
 ('M41', G('loops', kicker='and this is the new part', lines=['A loop with *nobody in it*'], size=68,
           inner=['write the app', 'write the checks', 'run them', 'fix what failed'],
           passes=3, label='no human',
           caption='a year ago this ended at "here is some code, try it"')),
 ('M42', ('F', 'R_done')),
 ('M43', ('F', 'R_done')),
 ('M44', G('myth', kicker='two very different claims', lines=['Consistent, or *correct*'], size=64,
           wrong='it wrote the exam, sat it, marked it, and passed',
           right='someone who is going to use it touched it and it worked',
           note='only one of those is worth anything to you')),
 ('R8', ('F', 'R_done')),
 ('R9', ('F', 'R_done')),
 ('M46', ('Fx', 'S', 0.0, 12.0)),
 ('M47', G('editor', kicker='and nothing asked for a name', lines=['It took *this* literally'], size=58,
           file='agents.md', start=19,
           rows=[{'t': 'The priority is a **slick, professional, genuinely', 'kind': 'important'},
                 {'t': 'good-looking interface** over a small set of', 'kind': 'important'},
                 {'t': 'features. A viewer should want to use it.', 'kind': 'important'}],
           focus=[0, 2], caption='it read that as an instruction, not a sentiment')),
 ('M45', G('rows', kicker='so we do it ourselves', lines=['The *same five*, in order'], size=62,
           numbered=True,
           rows=[{'t': 'Drag a card between columns'},
                 {'t': 'Reorder a card inside a column'},
                 {'t': 'Delete a card'},
                 {'t': 'Rename a column'},
                 {'t': 'Add a card, with a description'}])),
 ('R10', ('Fx', 'S', 12.0, 28.0)),
 ('R11', ('Fx', 'S', 28.0, 42.0)),
 ('R12', ('Fx', 'S', 42.0, 58.0)),
 ('R13', ('Fx', 'S', 58.0, 72.0)),
 ('R14', ('Fx', 'S', 72.0, 96.0)),
 ('R15', ('Fx', 'S', 96.0, 117.5)),
 ('Z1', G('myth', kicker='one line in the brief, visible in the result',
          lines=['Obvious, over *clever*'], size=64,
          wrong='install a library that specialises in dragging',
          right='use the dragging the browser has had for years',
          note='fewer moving parts, easier for the next person')),
 ('Z2', ('Fx', 'S', 12.0, 30.0)),
 ('Z3', ('Fs', 'P_plan', 'the whole lesson of the week', ['The brief *is*', 'the product'], 84)),
 ('M48', G('meter', kicker='what it cost, in both currencies', lines=['*Eleven minutes.* Four credits'], size=64,
           pct=4, fill='of the monthly allowance', rest='still there',
           caption='most of the eleven minutes was me clicking yes')),
 ('R16', ('Fs', 'R_done', 'and now an honest detour', ['The lecture I built', 'this from *broke*'], 64)),
 ('T2', ('Fs', 'R_build', 'which makes this the real subject', ['*Debugging*'], 112)),
 ('T3', G('flow', kicker='the move it makes almost every time',
          lines=['Guess. Patch. *Declare victory*'], size=58,
          nodes=[{'t': 'guess the cause', 'sub': 'without looking', 'tone': 'red'},
                 {'t': 'write a fix', 'sub': 'for the guess'},
                 {'t': '"fixed!"', 'sub': 'never once run', 'tone': 'red'}],
          caption='once you have seen it you will never unsee it')),
 ('T4', G('myth', kicker='two separate faults in that', lines=['Proving, not *guessing*'], size=64,
          wrong='it guessed, and it claimed',
          right='it should have proved, and it should have shown you',
          note='those are two different failures, not one')),
 ('T5', G('rows', kicker='write this one down', lines=['The *four steps*'], size=88, numbered=True,
          rows=[{'t': 'Reproduce the problem'},
                {'t': 'Prove you have reproduced it'},
                {'t': 'Find the root cause'},
                {'t': 'Fix it — and demonstrate the fix', 'hot': True}],
          stamp='give it these four, in this order')),
 ('T6', G('rows', kicker='and notice the balance', lines=['*Three* of four', 'are evidence'], size=68,
          rows=[{'t': 'Reproduce — evidence'},
                {'t': 'Prove — evidence'},
                {'t': 'Fix — the only line about code'},
                {'t': 'Demonstrate — evidence', 'hot': True}])),
 ('T7', ('Fs', 'R_build', 'you can run this without reading code', ['You do not need to read the fix', 'to ask whether it was *shown*'], 50)),
 ('T8', ('Fs', 'R_done', 'remember what it always looks like', ['Confident and detailed', 'is not *evidence*'], 58)),
 ('R17', G('rows', kicker='so, ours behaved. Keep the instruction anyway',
           lines=['For the day it *does not*'], size=64,
           rows=[{'t': 'I am not going to stage a fault to make the lesson land'},
                 {'t': 'Do not read our clean result as the normal one'},
                 {'t': 'Have the four steps in your hand before you need them', 'hot': True}])),
 ('H1', ('Fxs', 'S', 96.0, 112.0, 'the part these comparisons leave out', ['What this *did not*', 'show you'], 70)),
 ('H2', ('Fx', 'S', 96.0, 112.0)),
 ('H3', G('rows', kicker='and this is the one that bites', lines=['Requirements that *fight*'], size=68,
          rows=[{'t': 'Every line of our brief agrees with every other line'},
                {'t': 'A real request contains two things that cannot both be true'},
                {'t': 'And nobody notices until something has been built', 'hot': True}])),
 ('H4', G('stack', kicker='and it did not show you a second day',
          lines=['*Persistence* is a different job'], size=58,
          layers=[{'t': 'no persistence', 'sub': 'what we built. Close the tab, it resets', 'h': 44, 'tone': 'yellow'},
                  {'t': 'a database', 'sub': 'the moment you want it saved', 'h': 62},
                  {'t': 'accounts and logins', 'sub': 'because whose board is it', 'h': 62},
                  {'t': 'two people editing one card', 'sub': 'and now you have a real system', 'h': 72, 'tone': 'red'}],
           foot='not a bigger version of today — a different job')),
 ('H5', G(H, kicker='and the thing that actually breaks these', lines=['Someone *changes their mind*', 'and nobody wrote it down'], size=54)),
 ('H6', G('editor', kicker='which is what the file is for', lines=['Not a prompt. *A record*'], size=62,
          file='agents.md', start=1,
          rows=[{'t': '# Kanban — project brief', 'kind': 'h1'},
                {'t': '', 'kind': 'text'},
                {'t': 'the written-down version of what was agreed', 'kind': 'important'},
                {'t': '', 'kind': 'text'},
                {'t': 'the only thing from today still useful in a month', 'kind': 'important'}],
          focus=[2, 4])),
 ('H7', G('gauge', kicker='one practical warning', lines=['You will *feel* the limit', 'before you read it'], size=54,
          pct=97, limit=100, label='requests used this month',
          limitLabel='allowance',
          caption='if a tool suddenly gets worse, check your usage before you blame it')),
 ('Z4', G(H, kicker='two down', lines=['Side by *side*'], size=110, trans='fade')),
 ('Z5', G('matrix', kicker='same brief, same prompt, same checks',
          lines=['What actually *differed*'], size=58,
          cols=MATRIX_COLS, rows=MATRIX_ROWS, upto=2,
          dateline='28 august 2026')),
 ('Z6', G('matrix', kicker='and the result',
          lines=['*Both* cleared all five'], size=62,
          cols=MATRIX_COLS, rows=MATRIX_ROWS, upto=2,
          dateline='28 august 2026')),
 ('Z7', ('Fxs', 'S', 100.0, 117.5, 'which was not what I expected', ['On a job this size,', 'it *barely matters*'], 58)),
 ('Z8', ('Fs', 'R_done', 'and the next one changes shape', ['A command line tool', 'wearing an *editor*'], 54)),
 ('Z9', G('rows', kicker='before you go', lines=['*Reproduce. Prove.*', '*Fix. Demonstrate*'], size=62,
          rows=[{'t': 'Our build behaved today'},
                {'t': 'The day it does not, that is the difference'},
                {'t': 'between an afternoon and a week', 'hot': True}])),
]


# Slides voiced but CUT from the film.
#
# The deck was written to an estimate of 26 chars per second, taken from L17.
# Measured, this deck runs at 18 — L18 has more slides with shorter lines than
# L17 did, and adam_deck bakes 0.22s of lead, 0.60s of tail and 0.42s between
# sentences into every one of them, so per-slide overhead dominates. The voiced
# deck came out at 20:24 against his 14:44.
#
# The 1:1 runtime rule wins, so 31 slides are cut. The mp3s are already paid
# for, which is why the choice could be made on merit rather than on cost: what
# goes is duplicative lead-ins ('08', M20, M35, Z4), second halves of pairs that
# make the same point twice (M16 after M15, M43 before M44, M37 after M36, Z6
# after Z5), and my own expansions beyond what the reference lecture covers
# (Z1/Z2, most of the H block). Every beat the reference actually teaches stays,
# and the whole T block — the four-step debugging instruction, which is the most
# reusable artefact in the week — stays intact.
# Nothing is cut. He relaxed the 1:1 runtime rule on 2026-08-28 — see
# coursesmith-runtime-overshoot-ok. This deck measures 20:24 against his 14:44
# and every slide stays in, because the extra time IS the lecture: his build
# broke by accident and gave him the debugging lesson for free, ours worked, so
# ours has to earn the same lesson deliberately.
CUT = []


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
          f"film {film:.1f}s = {int(film//60)}:{int(film%60):02d}  (his 14:44)")

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
        jobs.append(['18', i, TAKES[take], round(ss, 2), round(to, 2), round(win, 3)])
    json.dump(jobs, open('/tmp/l18_jobs.json', 'w'), indent=1)

    rates.sort()
    print(f"footage slides {len(jobs)} of {len(SPEC)} = {len(jobs)/len(SPEC):.0%}")
    print(f"  rate range {rates[0][0]:.2f} ({rates[0][1]}) .. {rates[-1][0]:.2f} ({rates[-1][1]})")
    # The band is deliberately loose on the top end. A rate above 1 means the
    # clip is slowed, and the P take's long stretches are a STATIC plan sitting
    # on screen — slowing a frame that barely changes is imperceptible. What
    # actually matters is the bottom end: speeding a take up past ~3x turns
    # readable text into a blur.
    hot = [r for r in rates if r[0] > 1.55 or r[0] < 0.30]
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

    tsx = f'''// nocode18 — the Copilot build (his L18, 14:44)
//
// Build two of four from the same brief. {len(jobs)} of {len(SPEC)} slides are real screen
// recording, cut from three takes: the plan, the build, and the five checks run
// by hand afterwards.
//
// **Our result differs from his, and the lecture says so out loud.** In the
// course this is modelled on, Copilot floundered: it started the server in the
// wrong directory, sat waiting without realising it was up, and shipped a
// broken delete which it then "fixed" twice without ever reproducing the fault.
// That accident is where the four-step debugging instruction comes from, and it
// is the most reusable thing in the whole week.
//
// Ours did not break. On the free tier, with Auto routing to GPT-5.6 Luna and
// 3.6 credits spent, it planned in six phases with success criteria, asked
// permission eight times, wrote its own Vitest suite, drove a real browser with
// Playwright to check its own work, and passed all five checks when we ran them
// by hand. So the lecture teaches the four steps as a discipline to have ready
// rather than staging a fault to justify them — R16 and R17 say exactly that.
//
// The one place the tools genuinely diverge is permission: `chat.tools.autoApprove`
// was already on and Copilot STILL held `npx create-next-app`, because that
// category downloads and runs third-party code. M26-M30 unpack why that is the
// right line to draw, and it is the only real difference we found between the
// two builds.
//
// PII: VS Code shows the signed-in account's avatar — a photograph — in the
// activity bar, and Copilot cannot be used signed out, so it is in every frame.
// cut18.py masks it with a filled box in the activity bar's own #181818.

import React from 'react';
import {{Deck, Slide, deckFrames}} from './kit';

export const DURS = [
{chr(10).join('  ' + ', '.join(f'{d:.3f}' for d in durs[i:i+8]) + ',' for i in range(0, len(durs), 8))}
];
export const FILES = [
{chr(10).join('  ' + ', '.join(f"'{n}'" for n in names[i:i+8]) + ',' for i in range(0, len(names), 8))}
].map((n) => `${{n}}.mp3`);
export const GAP = {gap:.3f};
export const L18_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode18/shots/${{n}}.mp4`;

const SLIDES: Slide[] = [
{chr(10).join(lines)}
];

export const Nocode18: React.FC = () => (
  <Deck slides={{SLIDES}} durs={{DURS}} voDir="nocode18/vo" files={{FILES}} gap={{GAP}} />
);
'''
    out = os.path.join(R, 'renderer/src/nocode/l18.tsx')
    open(out, 'w').write(tsx)
    print("wrote", out)
    print("wrote /tmp/l18_jobs.json")


if __name__ == '__main__':
    main()
