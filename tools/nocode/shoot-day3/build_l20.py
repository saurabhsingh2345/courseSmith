#!/usr/bin/env python3
"""Generate renderer/src/nocode/l20.tsx and the cut jobs for L20 (Antigravity).

Build four of four. Antigravity IDE 1.107.0 — a VS Code fork, the third of the
four tools built on that editor.

**The reference lecture's middle section is obsolete and this one says so.**
ref-L20 treats Antigravity as the holdout that never adopted `agents.md`, and
spends its middle on a nine-step conversion: select the brief, copy it, create
`.agent/`, create `rules/` inside it, create `strategy.md` inside that, paste,
set an activation mode, save, delete the original. None of that is needed. Off
the same four words the other three got, it replied "based on the requirements in
AGENTS.md" and had found the file unaided. Verified twice — the shipped bundle
carries an `AGENTS.md` settings tab beside `GEMINI.md`, a `USE_AGENT_MD` flag,
and an `[InstructionsContextComputer] AGENTS.md files added:` log line. P5-P13
turn that into the day's real finding: four companies, two of them direct
competitors, one convention, in about twelve months.

Two more of his facts had moved, both corrected on camera with the date:
  * the model list starts on **Gemini 3.7 Flash High**, not "Gemini 3 Pro", and
    offers **Claude Sonnet 4.6, Claude Opus 4.6 and GPT-OSS 120B** alongside
    Google's own — which is Q8's point about the harness and the model coming
    apart;
  * **Auto Execution has two settings**, Always Proceed and Request Review. His
    third middle option, where the model decides when to interrupt you, is gone
    from this build. Q15 says that, and says it does not know why, rather than
    inventing a reason.

It also wrote its plan to a FILE (`implementation_plan.md`) and stopped for
approval **even with Auto Execution set to Always Proceed** — so three of the
four tools read "plan first" as plan-then-check. P14-P16.
"""
import json, os, subprocess, sys

R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
VO = sys.argv[1] if len(sys.argv) > 1 else "/tmp/vo20"
TARGET = None           # his L19 is 10:22; we run long on purpose
GAP_FIXED = 0.24

TAKES = {
    'A': 'L20_A_build_s01',      # the build executing, from Proceed onward
    'B': 'L20_B_plan_s01',       # scrolled back: the AGENTS.md reply, the plan file, the model list
    'S': 'L20_S_checks_s01',     # the five checks, by hand
}

# A_build starts AFTER the plan was approved: the first attempt's guard allowed
# only "Electron" (the AppleScript process name) while SegRec polls NSWorkspace,
# which calls this app "Antigravity IDE", so nothing recorded until that was
# found. The plan beat is therefore the B take — the conversation persists, so it
# is scrolled back to afterwards. Same screen, same words, filmed a few minutes
# later.
GROUPS = {
    # The B take is 75s, not 150s — its beats, read off the shoot log:
    #   0-8 open · 8-30 scrolled back to the AGENTS.md reply · 30-46 the plan
    #   46-63 the model list · 63-75 closed
    'B_reply': ('B', [(6.0, 30.0)]),      # "based on the requirements in AGENTS.md"
    'B_plan':  ('B', [(30.0, 46.0)]),     # the plan it wrote before touching anything
    'B_model': ('B', [(46.0, 63.0)]),     # the model dropdown, Claude and GPT-OSS in it
    'A_build': ('A', [(0.0, 430.0)]),     # it executing, phase by phase
    'A_done':  ('A', [(600.0, 769.0)]),   # the review screen, the summary, the note
    'S_all':   ('S', [(0.0, 117.5)]),     # the five checks, shared evenly
}

CUT = []
H = 'head'


def G(k, **kw):
    return ('G', dict(k=k, **kw))


SPEC = [
 ('P1', ('Fs', 'B_reply', 'build four of four', ['The *holdout*'], 112)),
 ('P2', G('rows', kicker='and the caveat comes first, not last',
          lines=['They did *not* run', 'the same model'], size=58,
          rows=[{'t': 'Cursor — a frontier model I chose by hand'},
                {'t': 'Copilot — whatever the free plan handed it'},
                {'t': 'Claude Code — the same model as Cursor'},
                {'t': 'This one — Google\'s default, which is fast, not strongest', 'hot': True}])),
 ('P3', G(H, kicker='the least surprising sentence in the course',
          lines=['Strong model,', '*strong* results'], size=76)),
 ('P4', G('myth', kicker='so here is the question to ask anyone',
          lines=['Ask *which model*'], size=72,
          wrong='which of these tools is the best one',
          right='which model was running, and on what date',
          note='a comparison that will not tell you is not a comparison')),
 ('Q1', G(H, kicker='and it has the best address in the industry',
          lines=['antigravity', '*.google*'], size=96, trans='rise')),
 ('Q2', G('rows', kicker='installing it', lines=['Take *all* the defaults'], size=80,
          rows=[{'t': 'Mac — drag it into Applications'},
                {'t': 'Windows — click next a few times'},
                {'t': 'There is nothing to configure and nothing to choose', 'hot': True}])),
 ('Q3', ('F', 'B_reply')),
 ('Q4', G('nest', kicker='and this is the third time today',
          lines=['It is *VS Code*. Again'], size=68,
          outer='Visual Studio Code', inner='Antigravity',
          ring=['same file tree', 'same tabs', 'same shortcuts', 'same panel'],
          caption='a different company, a different model, the same room')),
 ('Q5', G('rows', kicker='so count what that is worth',
          lines=['*Three* of four', 'are one editor'], size=64,
          rows=[{'t': 'Cursor — a fork of VS Code'},
                {'t': 'Antigravity — a fork of VS Code'},
                {'t': 'Copilot — an extension inside VS Code'},
                {'t': 'Learn the room once and you have learned three of them', 'hot': True}])),
 ('Q6', G(H, kicker='and I am not filming mine',
          lines=['A sign-in screen is', 'somebody\'s *name and email*'], size=54)),
 ('Q7', ('F', 'B_reply')),
 ('Q8', ('F', 'B_model')),
 ('Q9', G('rows', kicker='in google\'s own editor, offered as equals',
          lines=['Its *competitors* models'], size=64,
          rows=[{'t': 'Gemini 3.7 Flash — where it starts you'},
                {'t': 'Claude Sonnet 4.6 and Claude Opus 4.6 — Anthropic'},
                {'t': 'GPT-OSS 120B — open weights, run it yourself'}],
          stamp='28 august 2026 — check it yourself, it will have moved')),
 ('Q10', G('myth', kicker='which is the most useful thing to understand',
           lines=['They have *come apart*'], size=68,
           wrong='you choose a tool, and the tool is the thing',
           right='the harness is what you choose. the model is what does the work',
           note='and the harness vendors have accepted that')),
 ('Q11', ('F', 'B_model')),
 ('Q12', G('rows', kicker='and nobody explains this word either',
           lines=['A *lint* is not a bug'], size=72,
           rows=[{'t': 'Something imported and never used'},
                 {'t': 'A name written two different ways'},
                 {'t': 'Nothing that stops the app running'},
                 {'t': 'All of it mess that makes the next change harder', 'hot': True}])),
 ('Q13', ('Fs', 'B_model', 'so turn it on', ['It costs *nothing*', 'and happens without you'], 54)),
 ('Q14', G('ladder', kicker='the permission dial, for the fourth time',
           lines=['This one has *two* positions'], size=62,
           items=['Request review — ask me before you act',
                  'Always proceed — do not ask me at all'],
           pick=1, tone='#e53935')),
 ('Q15', G('myth', kicker='and one option has gone missing',
           lines=['I do not *know* why'], size=72,
           wrong='the course I built this from had a third setting: the model decides',
           right='this build has two. I am telling you rather than guessing',
           note='a course that pretends to know everything is worse than one that says where it stops')),
 ('Q16', ('Fs', 'A_build', 'you should hear me say this every time', ['A throwaway project,', 'an *empty folder*'], 54)),
 ('G2', G('hierarchy', kicker='what the other three did without being asked',
          lines=['*One file*, in the root'], size=64,
          nodes=[{'t': 'kanban', 'depth': 0, 'kind': 'dir'},
                 {'t': 'agents.md', 'depth': 1, 'kind': 'agents'},
                 {'t': 'README.md', 'depth': 1, 'kind': 'file'}],
          active=1, caption='nobody told any of them it was there')),
 ('P5', G(H, kicker='and now the section you are not getting',
          lines=['Let me tell you', 'what I *planned*'], size=64, trans='push')),
 ('P6', G('hierarchy', kicker='nine steps of fiddly file management',
          lines=['The conversion', 'his lecture *teaches*'], size=58,
          nodes=[{'t': 'kanban', 'depth': 0, 'kind': 'dir'},
                 {'t': '.agent', 'depth': 1, 'kind': 'dir'},
                 {'t': 'rules', 'depth': 2, 'kind': 'dir'},
                 {'t': 'strategy.md', 'depth': 3, 'kind': 'agents'},
                 {'t': 'agents.md', 'depth': 1, 'kind': 'file'}],
          active=3, caption='select, copy, three folders, paste, set a header, save, then delete the original')),
 ('P7', G(H, kicker='to get the same words', lines=['To the *same model*'], size=104)),
 ('G9', ('F', 'B_reply')),
 ('P8', ('F', 'B_reply')),
 ('P9', ('Fs', 'B_reply', 'it went and found the file', ['*AGENTS.md*'], 120)),
 ('P10', G('rows', kicker='and I checked twice, because courses get this wrong',
           lines=['Not a *guess*'], size=80,
           rows=[{'t': 'The app ships a settings page for this file, by name'},
                 {'t': 'It writes a line into its own log when it picks one up'},
                 {'t': 'It reads it deliberately', 'hot': True}])),
 ('P11', G('timeline', kicker='four companies. two of them direct competitors',
           lines=['*One* convention'], size=80,
           items=[{'when': 'a year ago', 't': 'every tool, its own config file',
                   'sub': 'its own name, its own folder'},
                  {'when': 'this year', 't': 'agents.md, near-universally',
                   'sub': 'read without being told'},
                  {'when': 'today', 't': 'the last holdout reads it too',
                   'sub': 'since the course this is built from', 'hot': True}],
           caption='about twelve months, and quietly')),
 ('P12', G(H, kicker='so the file is not an investment in one product',
           lines=['It is what *survives*', 'you changing your mind'], size=54, trans='rise')),
 ('P13', G('hierarchy', kicker='its own folder still exists, and still has a use',
           lines=['For rules that live *only here*'], size=58,
           nodes=[{'t': 'kanban', 'depth': 0, 'kind': 'dir'},
                  {'t': 'agents.md', 'depth': 1, 'kind': 'agents'},
                  {'t': '.agent', 'depth': 1, 'kind': 'dir'},
                  {'t': 'rules', 'depth': 2, 'kind': 'dir'}],
           active=1,
           caption='agents.md is the brief all four read — .agent/rules is for anything only this tool needs')),
 ('G8', G(H, kicker='which will outlive this particular argument',
          lines=['The convention is not the point.', 'The *brief* is the point'], size=52)),
 ('P14', ('F', 'B_plan')),
 ('P15', ('F', 'B_plan')),
 ('P16', G('rows', kicker='same nine words. four readings',
           lines=['What *plan first* meant'], size=64, numbered=True,
           rows=[{'t': 'Cursor — plan, then carry straight on'},
                 {'t': 'Copilot — plan, then carry straight on'},
                 {'t': 'Claude Code — plan, then stop and check'},
                 {'t': 'Antigravity — plan to a file, then stop and check', 'hot': True}])),
 ('R1', ('F', 'A_build')),
 ('R2', ('F', 'A_done')),
 ('R3', G('myth', kicker='two settings, because two different questions',
          lines=['*Ask*, and *show*'], size=76,
          wrong='one switch called permissions that means everything',
          right='one dial for stopping before it acts, another for showing you after',
          note='and I like that they are separate')),
 ('R4', ('F', 'A_done')),
 ('R5', ('F', 'A_done')),
 ('R6', G('myth', kicker='and this is where his build fell down',
          lines=['A prompt, or a *form*'], size=72,
          wrong='a bare browser pop-up asking for a name, nowhere for a description',
          right='a dialog with a required title and an optional description',
          note='the brief said a card has both — it read that as a requirement')),
 ('R7', G('rows', kicker='and nothing asked for any of this',
          lines=['It inferred the *taste*'], size=68,
          rows=[{'t': 'The five hex codes out of our brief, by name'},
                {'t': 'No emoji anywhere — it drew the icons instead'},
                {'t': 'The brief mentioned a palette and never mentioned emoji', 'hot': True}])),
 ('R8', G('rows', kicker='and the gates it set itself',
          lines=['*16 of 16*'], size=104,
          rows=[{'t': 'Sixteen tests out of sixteen passing'},
                {'t': 'No lint warnings at all'},
                {'t': 'Production build compiled cleanly'},
                {'t': 'A server running, with the address', 'hot': True}])),
 ('R9', ('Fs', 'A_done', 'and now the most useful part', ['The *honest* part'], 96)),
 ('R10', G('rows', kicker='this is the tool\'s headline feature',
           lines=['It drives a *real browser*', 'and records itself'], size=54,
           rows=[{'t': 'The agent tests the app by actually using it'},
                 {'t': 'And hands you a video of it working'},
                 {'t': 'In the course I built this from it is the striking moment of the day'}])),
 ('R11', ('F', 'A_done')),
 ('R12', ('F', 'A_done')),
 ('R13', G('rows', kicker='look at the shape of what it did',
           lines=['Four things, *in order*'], size=68, numbered=True,
           rows=[{'t': 'Named the failure'},
                 {'t': 'Named the cause — a 404 fetching the driver'},
                 {'t': 'Named the substitute it used instead'},
                 {'t': 'Named what the substitute does not cover', 'hot': True}],
           stamp='it did not report success and let me find out')),
 ('R14', ('Fs', 'A_done', 'you will see far more of this', ['How a thing *fails*', 'is worth more'], 58)),
 ('R15', ('Fs', 'A_done', 'so I will not describe it as though I had', ['I cannot *show* you', 'the feature'], 58)),
 ('R16', ('F', 'S_all')),
 ('K1', ('F', 'S_all')),
 ('K2', G('rows', kicker='fourth time today', lines=['The *same five*, in order'], size=62,
          numbered=True,
          rows=[{'t': 'Drag a card between columns'},
                {'t': 'Reorder a card inside a column'},
                {'t': 'Delete a card'},
                {'t': 'Rename a column'},
                {'t': 'Add a card, with a description'}])),
 ('K3', ('F', 'S_all')),
 ('K4', ('F', 'S_all')),
 ('K5', ('F', 'S_all')),
 ('K6', ('F', 'S_all')),
 ('K7', ('Fs', 'S_all', 'demonstrated four times in one afternoon', ['Rather than *asserted* once'], 62)),
 ('K8', ('F', 'S_all')),
 ('P17', ('Fs', 'S_all', 'and it is short', ['Next: the *verdict*'], 96)),
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
          f"film {film:.1f}s = {int(film//60)}:{int(film%60):02d}  (his 10:37)")

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
        jobs.append(['20', i, TAKES[take], round(ss, 2), round(to, 2), round(win, 3)])
    json.dump(jobs, open('/tmp/l20_jobs.json', 'w'), indent=1)

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
export const L20_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode20/shots/${{n}}.mp4`;

const SLIDES: Slide[] = [
{chr(10).join(lines)}
];

export const Nocode20: React.FC = () => (
  <Deck slides={{SLIDES}} durs={{DURS}} voDir="nocode20/vo" files={{FILES}} gap={{GAP}} />
);
'''
    out = os.path.join(R, 'renderer/src/nocode/l20.tsx')
    open(out, 'w').write(tsx)
    print("wrote", out)
    print("wrote /tmp/l20_jobs.json")


if __name__ == '__main__':
    main()
