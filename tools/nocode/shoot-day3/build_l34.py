#!/usr/bin/env python3
"""Generate renderer/src/nocode/l34.tsx — the assistant, and the week one wrap.

The payoff nearly went out false, and the lecture is built around that rather
than around the demo.

Two faults hid each other. The API key never reached the container, so every
question that genuinely needed the model returned a 500. And moving a card was
decided by a REGULAR EXPRESSION that ran before the model was called at all. So
the one thing that worked was the one thing that never needed the AI — and the
demo would have been "type a sentence, watch a card move, tell the audience a
language model changed your data" while the model was unreachable and uninvolved.

P6-P14 are that discovery. Both were then fixed on camera: the key plumbed into
the container, the pattern matching removed so the model decides from the board
context it already receives. The demo move is phrased INDIRECTLY on purpose —
"the card about the flaky checkout tests" is not the card's title — because a
pattern cannot resolve that and a model reading the board can.

The honest section is also not the reference's. His complaint is that everything
ended up in one module; ours is properly separated, so that would be false. What
is true and worse: the sign-in is entirely client-side, so the password ships
inside a JavaScript file any visitor can download. H5-H12 film finding it.
"""
import json, os, subprocess, sys

R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
VO = sys.argv[1] if len(sys.argv) > 1 else "/tmp/vo34"
TARGET = None           # his L29 is 11:20; we run long on purpose
GAP_FIXED = 0.24

TAKES = {
    'E': 'L34_A_p8_s01',        # 265s — part 8, reaching the model at all
    'B': 'L34_B_p910_s01',      # 1247s — parts 9 and 10
    'R': 'L34_C_real_s01',      # 741s — the key, and taking the pattern out
    'D': 'L34_D_demo_s01',      # 234s — the assistant, working for real
    'H': 'L34_H_honest_s01',    # 98s — the password in the bundle
}

GROUPS = {
    'E_ask':  ('E', [(6.0, 60.0)]),        # one call, one answer
    'E_ok':   ('E', [(210.0, 238.0)]),     # and it comes back
    # Ranges are sized so nothing ramps past about 8x. The demo clips (D_*) are
    # deliberately the gentlest — the whole point of this lecture is that you
    # watch the thing happen, and a card moved at 10x reads as a glitch.
    'B_ctx':  ('B', [(50.0, 130.0)]),      # giving it the board
    'B_chat': ('B', [(736.0, 752.0)]),     # and the panel
    'D_open': ('D', [(2.0, 14.0)]),        # board and assistant side by side
    'D_err':  ('D', [(14.0, 26.0)]),       # the first ask
    'R_log':  ('R', [(20.0, 80.0)]),       # the key never reached the box
    'R_regex':('R', [(90.0, 140.0)]),      # and the move was a pattern
    'R_fix':  ('R', [(600.0, 660.0)]),     # both fixed
    'D_sum':  ('D', [(64.0, 80.0)]),       # summarise my project
    'D_move': ('D', [(104.0, 150.0)]),     # move it by asking
    'D_land': ('D', [(150.0, 176.0)]),     # and it moves
    'D_back': ('D', [(196.0, 230.0)]),     # and it survived the session
    'H_ls':   ('H', [(2.0, 22.0)]),        # what the browser downloads
    'H_find': ('H', [(22.0, 44.0)]),       # which file has it
    'H_show': ('H', [(44.0, 92.0)]),       # and there it is
}

CUT = []
H = 'head'


def G(k, **kw):
    return ('G', dict(k=k, **kw))


SPEC = [
 ('A1', G(H, kicker='last three parts', lines=['And why they are *three*', 'and not one'], size=62, trans='fade')),
 ('A2', ('Fs', 'E_ask', 'part eight does nothing you can see', ['One call. One *answer*'], 96)),
 ('A3', G('rows', kicker='that is the whole part',
          lines=['No interface.', 'No *cleverness*'], size=76,
          rows=[{'t': 'No board involved at all'}])),
 ('A4', G(H, kicker='and it is the most valuable of the three',
          lines=['It separates *two* questions'], size=72)),
 ('A5', G('rows', kicker='that people spend hours confusing',
          lines=['*Reach* it, or', 'is it any *good*'], size=62, numbered=True,
          rows=[{'t': 'Is the key right, is the name right, is the account funded, does the network allow it'},
                {'t': 'And separately — is our feature any good', 'hot': True}])),
 ('A6', G('myth', kicker='because if you build it all at once',
          lines=['They look *identical*', 'from outside'], size=58,
          wrong='it is not working, so something in my feature is wrong',
          right='you are debugging two questions at the same time',
          note='and neither of them will tell you which one it is')),
 ('A7', ('Fs', 'E_ok', 'so part eight answers the first one alone', ['Two minutes, and never', 'in *question* again'], 54)),
 ('A8', G(H, kicker='and steal that habit well beyond AI',
          lines=['Prove you can *reach* it', 'before you build on it'], size=52, accent='#f5c518')),
 ('B1', ('Fs', 'B_ctx', 'part nine', ['Gives it the *board*'], 124)),
 ('B2', G('myth', kicker='because on its own it knows nothing about you',
          lines=['It has never *seen*', 'your cards'], size=62,
          wrong='ask it what is in progress and it will tell you',
          right='it will produce something plausible and completely invented',
          note='and it will sound entirely confident doing it')),
 ('B3', G('flow', kicker='so before every question we send',
          lines=['We *attach* the board'], size=96,
          nodes=[{'t': 'here are the columns', 'sub': 'and what is in them'},
                 {'t': 'here are the cards', 'sub': 'titles and descriptions'},
                 {'t': 'now, the user asks this', 'sub': 'your sentence, at the end'},
                 {'t': 'and it answers', 'sub': 'from what it was handed'}])),
 ('B4', ('Fs', 'B_ctx', 'and that is the entire trick', ['Behind *every* AI feature', 'that seemed to know things'], 50)),
 ('B5', G('myth', kicker='there is no memory, and no learning',
          lines=['A *briefing note*'], size=124,
          wrong='it learned about my project',
          right='a very good writer was handed a note a fraction of a second before answering',
          note='and the note is thrown away afterwards')),
 ('B6', G('rows', kicker='and once you see it that way',
          lines=['Failures become', '*predictable*'], size=76,
          rows=[{'t': 'A lot of AI products stop being mysterious'},
                {'t': 'It did not know because nobody told it', 'hot': True}])),
 ('C1', ('Fs', 'B_chat', 'and part ten', ['Turns talking', 'into *doing*'], 96)),
 ('C2', G(H, kicker='up to part nine it can describe your board',
          lines=['After part ten', 'it can *change* it'], size=62)),
 ('C3', G(H, kicker='which is worth being careful about',
          lines=['Something that makes things up', 'gets to touch the *data*'], size=48, accent='#f5c518')),
 ('C4', G('myth', kicker='and this is exactly the case I flagged with the diffs',
          lines=['*This* is that'], size=132,
          wrong='accept everything in bulk, it has been fine all day',
          right='slow down when it touches something you would not want wrong',
          note='the first time we plug in the AI — I said it then, and here it is')),
 ('P1', ('F', 'D_open')),
 ('P2', ('F', 'D_err')),
 ('P3', G(H, kicker='except it does not', lines=['It returns an *error*'], size=116)),
 ('P4', ('Fs', 'R_log', 'and the reason is almost funny', ['The key is on *my* machine'], 84)),
 ('P5', G('myth', kicker='and the application does not run on my machine',
          lines=['Nobody handed the key', 'to the *box*'], size=58,
          wrong='the key is in a file in the project, so the app has it',
          right='the app runs in the container, and the container was never given it',
          note='every question needing the model failed, on the first one I asked')),
 ('P6', G(H, kicker='now that is the small problem',
          lines=['I nearly put something', 'in front of you that', 'was not *true*'], size=54, trans='push')),
 ('P7', ('Fs', 'R_regex', 'i went and read how it decides to move a card', ['Not the *model*.', 'A pattern'], 68)),
 ('P8', G('myth', kicker='somewhere in there is a rule',
          lines=['The model is *never asked*'], size=76,
          wrong='you asked in English, so a language model understood you',
          right='if the sentence looks roughly like move something to somewhere, pull out the two things',
          note='written the way this was written in nineteen ninety')),
 ('P9', G('rows', kicker='and see how close I came to not noticing',
          lines=['The two faults *hid*', 'each other'], size=62,
          rows=[{'t': 'With the key missing, every real question was failing'},
                {'t': 'And moving a card worked perfectly'},
                {'t': 'Because moving a card never needed the model', 'hot': True}])),
 ('P10', G(H, kicker='so the demo I was about to record',
           lines=['Would have been a', '*regular expression*'], size=58)),
 ('P11', G('myth', kicker='and nothing was hidden from me',
           lines=['It did not *occur* to it'], size=92,
           wrong='it concealed how the move actually worked',
           right='it is all in the code, in a file I could open',
           note='it was never asked to mention it, so it did not')),
 ('P12', G(H, kicker='which is the sharpest version of the whole week',
           lines=['Whether what you asked is', 'what you *wanted*'], size=52, accent='#f5c518')),
 ('P13', G('rows', kicker='and notice how it was caught',
           lines=['*Using* it', 'found the reading'], size=72,
           rows=[{'t': 'Not by reading the code — I read that part afterwards'},
                 {'t': 'By using the thing and finding one bit broken'},
                 {'t': 'Which made me look at the bit that worked', 'hot': True}])),
 ('P14', ('Fs', 'R_fix', 'so I sent it both', ['Slower and *real*, over', 'instant and fake'], 54)),
 ('P15', ('F', 'D_sum')),
 ('P16', G('rows', kicker='and it does, accurately',
           lines=['*Reading* the board'], size=104,
           rows=[{'t': 'Naming the columns and what is in them'},
                 {'t': 'Which means it is not guessing', 'hot': True}])),
 ('P17', ('Fs', 'D_move', 'and now the one I have been building towards', ['Not by *dragging* it.', 'By asking'], 62)),
 ('P18', ('F', 'D_land')),
 ('P19', G('flow', kicker='and this time I can tell you what happened, because I checked',
           lines=['The *whole* round trip'], size=88,
           nodes=[{'t': 'the board was described', 'sub': 'to a language model, with my sentence'},
                  {'t': 'it decided', 'sub': 'which card, and which column I meant'},
                  {'t': 'it said so', 'sub': 'in a form our back end understands'},
                  {'t': 'and the database changed', 'sub': 'for real'}],
           caption='I never named the card. It worked out which one I meant.')),
 ('P20', ('F', 'D_back')),
 ('P21', G('rows', kicker='so look at what this is now',
           lines=['Reorganise your work', 'by *asking*'], size=62, numbered=True,
           rows=[{'t': 'A board that persists, in a database, in a container'},
                 {'t': 'A front end, a back end, and an API between them'},
                 {'t': 'And an assistant that can change it', 'hot': True}])),
 ('P22', G(H, kicker='that is not a demo',
           lines=['The shape of a product', 'people *pay* for'], size=58)),
 ('P23', G('myth', kicker='and built, I want to say clearly',
           lines=['Both found by *using* it'], size=88,
           wrong='and it all worked first time',
           right='with one thing in it that was fake until forty minutes ago',
           note='and a key that was never plugged in')),
 ('H1', G(H, kicker='and now the part that matters more than the demo',
          lines=['Anybody can show you', 'a thing *working*'], size=54, trans='rise')),
 ('H2', G(H, kicker='this is not finished',
          lines=['The list is only useful', 'while you are *impressed*'], size=52)),
 ('H3', G('rows', kicker='so, quickly', lines=['What is *wrong* with it'], size=96, numbered=True,
          rows=[{'t': 'One user, hard-coded'},
                {'t': 'No sign it is thinking, so it looks frozen'},
                {'t': 'One board'},
                {'t': 'Nothing deployed — it exists only on this machine'}])),
 ('H4', G(H, kicker='and then the one to sit with',
          lines=['Because it is the kind', 'of thing that *ships*'], size=56)),
 ('H5', ('Fs', 'H_ls', 'the sign-in', ['Let me show you *something*'], 92)),
 ('H6', G('rows', kicker='your browser downloads these to make the page work',
          lines=['You can read *any* of them'], size=76,
          rows=[{'t': 'That is normal. Every website does it.'}])),
 ('H7', ('F', 'H_find')),
 ('H8', ('Fs', 'H_show', 'and in one of them', ['The *password*.', 'Written out'], 92)),
 ('H9', G('myth', kicker='which means it is not a lock',
          lines=['A *sign* on a door'], size=116,
          wrong='you have to sign in, so the board is protected',
          right='anybody who opens the page finds the password in thirty seconds',
          note='and I have been calling it a sign-in all afternoon')),
 ('H10', G(H, kicker='now, to be fair, nobody did anything wrong',
           lines=['My *shortcut*.', 'Not its mistake'], size=68)),
 ('H11', G('rows', kicker='but this is exactly how it happens outside a classroom',
           lines=['Nobody goes *back*'], size=100,
           rows=[{'t': 'The shortcut goes in for perfectly good reasons'},
                 {'t': 'Somebody shows the thing to somebody else'},
                 {'t': 'And nobody asks which shortcuts were load-bearing', 'hot': True}])),
 ('H12', G('myth', kicker='so the lesson is not go and learn about authentication',
           lines=['*Which kind* is this one'], size=76,
           wrong='we will do that properly later',
           right='this must never leave this laptop',
           note='know which of your simplifications is which')),
 ('H13', G(H, kicker='and I could only tell you that',
           lines=['Because I *looked* at what', 'my app was sending'], size=52)),
 ('H14', G(H, kicker='so what would I do about all of it',
           lines=['Not fix it by *hand*'], size=112)),
 ('H15', G('rows', kicker='a fresh conversation, and a different agent',
           lines=['Without telling it', 'what I *think*'], size=68,
           rows=[{'t': 'Review this project and tell me what is wrong with it'}])),
 ('H16', G('myth', kicker='and it would find the password in four seconds',
           lines=['One *prompt*'], size=140,
           wrong='a second opinion is a luxury for big projects',
           right='something with no stake in what the first one wrote, for one prompt',
           note='the most useful and least used move available to you')),
 ('W1', G(H, kicker='so that is week one', lines=['What happened to *you*,', 'not what we built'], size=54, trans='push')),
 ('W2', G('myth', kicker='because the app is not the point',
          lines=['There are two *hundred*', 'of them'], size=68,
          wrong='you have built a Kanban board',
          right='you have a method, and it is the same wherever you point it',
          note='and most of the two hundred are better than ours')),
 ('W3', G(H, kicker='five moves', lines=['Whatever you point', 'it *at*'], size=104)),
 ('W4', G('rows', kicker='one', lines=['Write the plan *yourself*'], size=104,
          rows=[{'t': 'In pieces sized so that if one fails, you know where to look'}])),
 ('W5', G('rows', kicker='two', lines=['Ask for *questions* first'], size=96,
          rows=[{'t': 'And treat silence as a pass, not as agreement'}])),
 ('W6', G('rows', kicker='three', lines=['Check in *after* every part'], size=88,
          rows=[{'t': 'Before the risky thing, rather than after it'}])),
 ('W7', G('rows', kicker='four', lines=['Describe faults *precisely*'], size=84,
          rows=[{'t': 'What you did. What you expected. What happened instead.'}])),
 ('W8', G('rows', kicker='five', lines=['Go and *look*'], size=140,
          rows=[{'t': 'Because its check passing is not the same as it working', 'hot': True}])),
 ('W9', G('myth', kicker='and none of those needed you to read a line of code',
          lines=['Every one needed', '*attention*'], size=68,
          wrong='to direct this work you need to understand the code',
          right='you need to notice when the shape of it stops making sense',
          note='that is the trade this whole course is built on')),
 ('W10', G(H, kicker='now the honest accounting', lines=['I have been *promising* it', 'all along'], size=56)),
 ('W11', G('rows', kicker='what we actually saw', lines=['All of it *true*'], size=112,
           rows=[{'t': 'Day one — tireless, and it was'},
                 {'t': 'Day three — jumping to conclusions, and it did'},
                 {'t': 'Today — verified with the one tool that could not see the fault'}])),
 ('W12', G('rows', kicker='and in the same afternoon',
           lines=['Also *true*'], size=124,
           rows=[{'t': 'A database design better than the one in my head'},
                 {'t': 'And it caught a fault in its own scheme when I asked one question', 'hot': True}])),
 ('W13', G('myth', kicker='so if you have come out of this week with either extreme',
           lines=['I have taught you *badly*'], size=76,
           wrong='these things are magic',
           right='these things are useless',
           note='both of those are my failure, in opposite directions')),
 ('W14', G(H, kicker='extremely capable, and no judgement about when to stop',
           lines=['*You* are the judgement'], size=96, accent='#f5c518')),
 ('W15', G(H, kicker='next week we change tools',
           lines=['And I want your expectations', 'set *properly*'], size=52)),
 ('W16', G(H, kicker='everything you have learned transfers',
           lines=['*All* of it'], size=150)),
 ('W17', G('rows', kicker='what changes is what you are shown',
           lines=['It stops being *polite*'], size=88,
           rows=[{'t': 'You will see how full the context is'},
                 {'t': 'You will see what it is about to run, before it runs', 'hot': True}])),
 ('W18', G('myth', kicker='which is less comfortable and much better',
           lines=['Why we did it in', 'this *order*'], size=62,
           wrong='we should have started with the tool that shows you everything',
           right='you now know what those numbers mean',
           note='because you have felt what happens when nobody is watching them')),
 ('W19', G('quote', kicker='one more thing to take with you, and it is not mine',
           text='Keep the AI on a tight leash — the temptation is always to hand it something huge and hope.',
           who="the idea is Karpathy's; the phrasing is mine")),
 ('W20', G('rows', kicker='and everything today was that leash',
           lines=['*Ten* parts, not one prompt'], size=76, numbered=True,
           rows=[{'t': 'A stop before the database'},
                 {'t': 'A browser opened by hand when the tests said everything was fine'},
                 {'t': 'And a demo pulled apart because it was too good to be true', 'hot': True}])),
 ('W21', G('myth', kicker='and it is not a lack of ambition',
           lines=['The *only* way found', 'so far'], size=68,
           wrong='working in small steps means thinking small',
           right='it is how you get ambitious work out of these and still know whether it worked',
           note='nobody has found another one')),
 ('W22', G(H, kicker='well done for getting through the first week',
           lines=['See you in the *second*'], size=88, trans='fade')),
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
        jobs.append(['34', i, TAKES[take], round(ss, 2), round(to, 2), round(win, 3)])
    json.dump(jobs, open('/tmp/l34_jobs.json', 'w'), indent=1)

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

    tsx = f'''// nocode34 — the assistant, and the week one wrap
//
// The payoff nearly went out false. The API key never reached the container and
// the card move was decided by a regular expression, so the only thing that
// worked was the only thing that never needed the model. P6-P14 are finding
// that; both were fixed on camera before the demo was shot. The demo names the
// card indirectly on purpose — a pattern cannot resolve "the card about the
// flaky checkout tests", and a model reading the board can.

import React from 'react';
import {{Deck, Slide, deckFrames}} from './kit';

export const DURS = [
{chr(10).join('  ' + ', '.join(f'{d:.3f}' for d in durs[i:i+8]) + ',' for i in range(0, len(durs), 8))}
];
export const FILES = [
{chr(10).join('  ' + ', '.join(f"'{n}'" for n in names[i:i+8]) + ',' for i in range(0, len(names), 8))}
].map((n) => `${{n}}.mp3`);
export const GAP = {gap:.3f};
export const L34_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode34/shots/${{n}}.mp4`;

const SLIDES: Slide[] = [
{chr(10).join(lines)}
];

export const Nocode34: React.FC = () => (
  <Deck slides={{SLIDES}} durs={{DURS}} voDir="nocode34/vo" files={{FILES}} gap={{GAP}} />
);
'''
    out = os.path.join(R, 'renderer/src/nocode/l34.tsx')
    open(out, 'w').write(tsx)
    print("wrote", out)
    print("wrote /tmp/l34_jobs.json")


if __name__ == '__main__':
    main()
