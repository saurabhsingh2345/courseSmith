#!/usr/bin/env python3
"""Generate renderer/src/nocode/l33.tsx — the database, the routes, and the wiring.

The reference's centrepiece is a half-hour rut: his agent looped on a drag-and-
drop bug it had ALREADY fixed, because it trusted its own broken tests over the
application. That did not happen to us, and it is not staged.

What we got instead is the same lesson from the other side, and it is stronger
because it is causal rather than anecdotal. In L32 the agent verified a page with
the one tool that could not see the fault and reported success on a blank screen.
One sentence was added to this part's prompt — confirm it in a real browser, not
just in tests — and it did, and it left a card called "browser test card" sitting
in the backlog as evidence. G3-G10 are that card: last lecture's failure was a
missing instruction, not a lazy tool, and naming the check fixed it permanently.

G11-G20 then do the thing the rut existed to argue for anyway. Drag a card, log
out, hard reload, sign back in, and find it where you left it — a move retrieved
from a database by a session that did not exist when the move was made.
"""
import json, os, subprocess, sys

R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
VO = sys.argv[1] if len(sys.argv) > 1 else "/tmp/vo33"
TARGET = None           # his L29 is 11:20; we run long on purpose
GAP_FIXED = 0.24

TAKES = {
    'D': 'L33_A_part5_s01',     # 297s — the design, written for review only
    'I': 'L33_B_p567_s01',      # 1013s — part 5 implemented, and the integrity fix
    'W': 'L33_C_p67_s01',       # 937s — parts 6 and 7
    'C': 'L33_D_check_s01',     # 128s — going and looking
}

GROUPS = {
    'D_ask':  ('D', [(6.0, 40.0)]),        # the prompt, and the stop it obeys
    'D_doc':  ('D', [(60.0, 200.0)]),      # the design document appearing
    'D_read': ('D', [(240.0, 294.0)]),     # reading it back
    'I_push': ('I', [(710.0, 770.0)]),     # the answer to the constraint question
    'I_stop': ('I', [(960.0, 1000.0)]),    # and it stops, and asks
    # Every range is sized so no clip ramps past about 8x — a 14x ramp on the
    # part 7 footage read as a flicker rather than as work being done.
    'W_p6':   ('W', [(30.0, 130.0)]),      # the routes
    'W_p7':   ('W', [(340.0, 420.0)]),     # the wiring, which is the long one
    'W_done': ('W', [(880.0, 925.0)]),     # and it reports back
    'C_board':('C', [(2.0, 16.0)]),        # the board out of the database
    'C_eviD': ('C', [(16.0, 32.0)]),       # the card it left behind
    # The three check clips are held close to real time on purpose. The whole
    # argument of the lecture is that you go and LOOK, and a drag ramped to 6x
    # reads as a glitch rather than as a card being moved by a person.
    'C_drag': ('C', [(34.0, 50.0)]),       # moving one
    'C_out':  ('C', [(58.0, 74.0)]),       # logged out, hard reloaded
    'C_back': ('C', [(84.0, 112.0)]),      # and it is still there
}

CUT = []
H = 'head'


def G(k, **kw):
    return ('G', dict(k=k, **kw))


SPEC = [
 ('A1', G(H, kicker='part five is the only part that stops and asks',
          lines=['And I want to explain', '*why*'], size=62, trans='fade')),
 ('A2', G('rows', kicker='everything before this was reversible',
          lines=['Throw it away.', 'Lose *twenty minutes*'], size=62,
          rows=[{'t': 'If part three had come out wrong, I would have asked again'}])),
 ('A3', G('myth', kicker='the database is different', lines=['Not because it is *hard*'], size=88,
          wrong='put the checkpoint before the difficult part',
          right='put it before the part everything else is shaped by',
          note='difficulty is not the variable')),
 ('A4', G('rows', kicker='because of what sits on top of it',
          lines=['You do not lose *part five*'], size=76,
          rows=[{'t': 'Six, seven, eight, nine and ten all assume this shape'},
                {'t': 'Get it wrong and you lose everything built on it', 'hot': True}])),
 ('A5', G(H, kicker='so the question is not how difficult is this',
          lines=['How much of what comes next', '*depends* on it'], size=52, accent='#f5c518')),
 ('A6', ('Fs', 'D_ask', 'and that is answerable', ['Without knowing anything', 'about *databases*'], 58)),
 ('B1', G(H, kicker='now what it came back with', lines=['No *jargon*.', 'I promise'], size=88, trans='rise')),
 ('B2', G('myth', kicker='and it starts with one idea', lines=['A table is a *spreadsheet*'], size=76,
          wrong='databases are a specialist subject',
          right='rows are records, columns are fields, and that is the whole idea',
          note='the rest is bookkeeping about how spreadsheets refer to each other')),
 ('B3', G('rows', kicker='it designed four of them', lines=['*Four* spreadsheets'], size=104, numbered=True,
          rows=[{'t': 'Users'}, {'t': 'Boards'}, {'t': 'Columns'}, {'t': 'Cards'}])),
 ('B4', G('rows', kicker='users is one row per person',
          lines=['One row.', 'Built for *thousands*'], size=68,
          rows=[{'t': 'The whole app has one login'},
                {'t': 'And the table is shaped as though it had many, deliberately', 'hot': True}])),
 ('B5', ('Fs', 'D_doc', 'boards carries the user it belongs to', ['One *number* is the', 'whole relationship'], 56)),
 ('B6', ('F', 'D_doc')),
 ('B7', ('F', 'D_doc')),
 ('B8', G(H, kicker='and now the sentence that makes it all make sense',
          lines=['Moving a card is', '*one number* changing'], size=58, accent='#f5c518')),
 ('B9', G('flow', kicker='that is genuinely it', lines=['All that *dragging*'], size=104,
          nodes=[{'t': 'you drag a card', 'sub': 'across the screen'},
                 {'t': 'its column changes', 'sub': 'one field, one row'},
                 {'t': 'positions are tidied', 'sub': 'so the order still makes sense'},
                 {'t': 'done', 'sub': 'and it is right tomorrow'}],
          caption='underneath, one field changing value')),
 ('C1', G(H, kicker='now the part I want to be honest about',
          lines=['The *counterweight*'], size=112, trans='push')),
 ('C2', G('myth', kicker='i had a design in my head before i asked',
          lines=['One row. The *whole* board'], size=68,
          wrong='store the entire board as one blob of text against the user',
          right='separate tables for columns and cards',
          note='mine would have worked. mine is also the hacky one.')),
 ('C3', G(H, kicker='and I knew it was the hacky one',
          lines=['I would have done it', '*anyway*'], size=68)),
 ('C4', G(H, kicker='its version is better than mine',
          lines=['Not slightly.', '*Properly*'], size=104)),
 ('C5', G('rows', kicker='and here is the test of whether I understand why',
          lines=['I can *explain* it'], size=100,
          rows=[{'t': 'With my blob, finding one column means reading the whole board and picking through it'},
                {'t': 'With its version, that is a question you can just ask', 'hot': True}])),
 ('C6', G('rows', kicker='and the second reason, which is the bigger one',
          lines=['A second board.', 'A second *user*'], size=68,
          rows=[{'t': 'With mine, unpick the format and rewrite everything that touches it'},
                {'t': 'With its version, it is another row', 'hot': True}])),
 ('C7', ('Fs', 'D_read', 'so I approved its design over my own', ['And *sit* with this'], 116)),
 ('C8', G('myth', kicker='both of these are true at once',
          lines=['Which is *which*, today'], size=92,
          wrong='it is better than me, or it is worse than me',
          right='it is better at some things and worse at others',
          note='the whole skill is telling them apart')),
 ('D1', G(H, kicker='i did not approve it silently', lines=['Copy this *even when*', 'you are not sure'], size=56)),
 ('D2', ('Fs', 'D_read', 'one thing looked awkward to me', ['No two columns in the', '*same position*'], 60)),
 ('D3', G('myth', kicker='which sounds obviously correct', lines=['And might *not* be'], size=104,
          wrong='no two things may ever share a position — obviously',
          right='reordering usually has a moment where two briefly do',
          note='before everything settles')),
 ('D4', G(H, kicker='and here is the important part',
          lines=['I do not *know*.', 'Genuinely'], size=96)),
 ('D5', G('rows', kicker='so I did not tell it to change it',
          lines=['*Described* it. Asked.'], size=88, numbered=True,
          rows=[{'t': 'Here is what worries me'}, {'t': 'Here is why'},
                {'t': 'If it causes that, change it. If not, leave it.', 'hot': True}])),
 ('D6', G(H, kicker='that shape is worth memorising',
          lines=['*You* decide'], size=140, accent='#f5c518')),
 ('D7', G('rows', kicker='and it costs nothing either way',
          lines=['One *sentence*'], size=112,
          rows=[{'t': 'If I am wrong, I have lost a sentence and learned something'},
                {'t': 'If I am right, I have caught it before a single row existed', 'hot': True}])),
 ('D8', ('Fs', 'I_push', 'and here is what came back', ['Better than I *deserved*'], 88)),
 ('D9', G('myth', kicker='yes — the reorder would have collided',
          lines=['The database would', 'have *refused* it'], size=58,
          wrong='two cards can never share a position, so nothing can go wrong',
          right='during a move between columns, two would have — briefly',
          note='and the whole move would have failed')),
 ('D10', G('rows', kicker='but it did not remove the rule, which is what I would have done',
           lines=['Keep the rule.', 'Change the *move*'], size=68,
           rows=[{'t': 'Positions are tidied on the way through'},
                 {'t': 'So the collision never happens at all', 'hot': True}])),
 ('D11', ('F', 'I_push')),
 ('D12', G('rows', kicker='and look at what it cost me',
           lines=['I could not have told you', 'the *fix*'], size=58,
           rows=[{'t': 'I did not know whether there was a problem'},
                 {'t': 'I had a feeling one rule looked uncomfortable'},
                 {'t': 'And I said so out loud', 'hot': True}])),
 ('D13', G(H, kicker='that is the entire skill', lines=['Not expertise.', '*Refusing* to let it past'], size=56)),
 ('D14', ('Fs', 'I_stop', 'one more thing', ['It finished, and it', '*stopped*'], 84)),
 ('D15', G(H, kicker='because the plan told it to, four hours ago, in writing',
           lines=['And it *remembered*'], size=112)),
 ('D16', G('myth', kicker='which is the return on writing a plan at all',
           lines=['Built into the thing', 'being *followed*'], size=58,
           wrong='I have to remember to stop it at the right moment',
           right='the stop is written down, so it happens without me',
           note='four hours later, unprompted')),
 ('E1', ('Fs', 'W_p6', 'part six', ['The shortest explanation', 'of the *day*'], 62)),
 ('E2', G('rows', kicker='a route is a question it knows how to answer',
          lines=['An address, and a *rule*'], size=76,
          rows=[{'t': 'What is on my board'},
                {'t': 'Move this card to that column'},
                {'t': 'Rename this column'}])),
 ('E3', ('F', 'W_p6')),
 ('E4', G(H, kicker='and that is genuinely all a back end is',
          lines=['A list of questions,', 'and the *answers*'], size=58)),
 ('E5', ('Fs', 'W_p6', 'at the end of part six', ['Nothing on screen', 'has *changed*'], 68)),
 ('E6', G(H, kicker='which is worth pausing on',
          lines=['The first part today', 'with *no* visible result'], size=54)),
 ('E7', G('myth', kicker='and it is a trap for how you judge progress',
          lines=['The screen is not', 'the *measure*'], size=62,
          wrong='nothing changed, so nothing happened',
          right='every route exists and nobody is asking them yet',
          note='that is what the checklist in the plan is for')),
 ('F1', ('Fs', 'W_p7', 'part seven', ['The *biggest* change', 'of the day'], 76)),
 ('F2', G('rows', kicker='up to now the board has been drawing itself',
          lines=['Nothing was ever *saved*'], size=76,
          rows=[{'t': 'The cards are written into the front end'},
                {'t': 'The same five columns every time you reload', 'hot': True}])),
 ('F3', G('flow', kicker='after part seven', lines=['It *asks*'], size=140,
          nodes=[{'t': 'the board loads', 'sub': 'by asking the back end'},
                 {'t': 'you move a card', 'sub': 'and it says so'},
                 {'t': 'the back end writes it', 'sub': 'to the database'},
                 {'t': 'and it is still there', 'sub': 'tomorrow'}])),
 ('F4', G(H, kicker='which is the difference between', lines=['A *picture* of an app,', 'and an app'], size=58)),
 ('F5', ('F', 'W_p7')),
 ('G1', ('Fs', 'W_done', 'and it worked', ['First time.', 'No *drama*'], 96)),
 ('G2', ('F', 'C_board')),
 ('G3', ('Fs', 'C_eviD', 'there is a card in the backlog', ['I did not *make* that'], 96)),
 ('G4', G(H, kicker='it made it while checking its own work, and left it',
          lines=['The most useful *litter*', 'I have seen all week'], size=52)),
 ('G5', G('myth', kicker='because last lecture it did not open a browser',
          lines=['That check *passed*.', 'The page was blank'], size=56,
          wrong='it checked and said it worked, so it lied',
          right='it ran the only check available from where it was standing',
          note='and that check could not see the fault')),
 ('G6', G('term', kicker='and one sentence changed between then and now',
          lines=['*That* is the difference'], size=88,
          title='added to this part\'s prompt',
          term=[{'t': 'confirm it in a real browser,', 'kind': 'cmd'},
                {'t': 'not just in tests', 'kind': 'cmd'},
                {'t': '  and it did', 'kind': 'ok'}])),
 ('G7', G(H, kicker='and there are two lessons available here',
          lines=['Only *one* of them', 'is true'], size=64, trans='push')),
 ('G8', G('myth', kicker='the false one', lines=['It was not being *lazy*'], size=100,
          wrong='it cut corners and I told it off',
          right='it did exactly what it had been asked, both times',
          note='there was nothing to tell off')),
 ('G9', G('rows', kicker='the true one', lines=['*How* to check', 'was missing'], size=76,
          rows=[{'t': 'So it chose, and it chose the cheapest check available'},
                {'t': 'Once the instruction named the check, it ran that one', 'hot': True}])),
 ('G10', G(H, kicker='so when it feels like carelessness',
           lines=['Look at your *instruction*', 'before the tool'], size=52, accent='#f5c518')),
 ('G11', G('myth', kicker='and it has shown me evidence, and I am still going to look',
           lines=['Not because I *distrust* it'], size=68,
           wrong='I am checking because I do not believe it',
           right='I am checking because the check is cheap and being wrong is not',
           note='those are different reasons and only one of them scales')),
 ('G12', G('rows', kicker='here is the arithmetic',
           lines=['Thirty seconds, against', '*three parts*'], size=58,
           rows=[{'t': 'Checking costs me thirty seconds'},
                 {'t': 'Being wrong costs three more parts built on a board that does not save', 'hot': True}])),
 ('G13', ('F', 'C_drag')),
 ('G14', G(H, kicker='and then the part that actually proves it',
           lines=['More than *reloading*', 'the page'], size=62)),
 ('G15', G('rows', kicker='because a reload proves almost nothing',
           lines=['End the *session*'], size=112, numbered=True,
           rows=[{'t': 'Log out, so the session is gone'},
                 {'t': 'Hard reload, so nothing comes from the browser memory'}])),
 ('G16', ('F', 'C_out')),
 ('G17', ('F', 'C_back')),
 ('G18', G(H, kicker='parts five, six and seven in one frame',
           lines=['A session that did not', '*exist* when I moved it'], size=54, accent='#f5c518')),
 ('G19', G('myth', kicker='and the rule I would keep from today',
           lines=['*Use* the thing'], size=132,
           wrong='when it says it works, read its report more carefully',
           right='when it says it works, go and use the thing',
           note='the report is not the application')),
 ('G20', G(H, kicker='thirty seconds, with your own hands',
           lines=['And it requires *no*', 'expertise whatsoever'], size=54)),
 ('Z1', G(H, kicker='one last thing', lines=['The bridge into', '*next week*'], size=96, trans='rise')),
 ('Z2', G(H, kicker='this conversation has run since the first prompt this morning',
          lines=['All of it, in *one* thread'], size=62)),
 ('Z3', G('myth', kicker='and there is a limit to what it holds at once',
          lines=['The desk has an *edge*'], size=96,
          wrong='it remembers everything you have said',
          right='there is a stack of paper on a desk, and the desk has an edge',
          note='we drew this on day two')),
 ('Z4', G('rows', kicker='and what happens at the edge is not an error',
          lines=['*Quietly*'], size=140,
          rows=[{'t': 'It starts summarising the older parts'},
                {'t': 'And then it starts dropping them', 'hot': True}])),
 ('Z5', G(H, kicker='this tool does not show you a meter',
          lines=['It is *handling* it', 'for us'], size=68)),
 ('Z6', G(H, kicker='next week you will see exactly how full it is',
          lines=['And we are going to', '*obsess* over it'], size=58)),
 ('Z7', G('rows', kicker='because the symptom is not a crash',
          lines=['It gets slightly *worse*'], size=88,
          rows=[{'t': 'It forgets a decision you made four hours ago'},
                {'t': 'It re-solves something that was already settled', 'hot': True}])),
 ('Z8', G('myth', kicker='and you will read that wrong',
          lines=['Not a *bad day*'], size=112,
          wrong='the model is having a bad day today',
          right='the beginning of the conversation has fallen off the desk',
          note='quietly, and without telling you')),
 ('Z9', G(H, kicker='so here is the practice', lines=['Stop it. Start a', '*new* one'], size=76)),
 ('Z10', G('myth', kicker='and I will be honest that it feels wrong every time',
           lines=['It is the *right* move'], size=96,
           wrong='you are throwing away everything it knows',
           right='you are moving what matters somewhere it cannot be forgotten',
           note='it feels like starting again with a stranger')),
 ('Z11', G(H, kicker='but not before the one thing that makes it safe',
           lines=['The whole technique,', 'in a *sentence*'], size=54)),
 ('Z12', G('term', kicker='and the second half is load-bearing',
           lines=['*Including* the decisions'], size=84,
           title='before you start the new conversation',
           term=[{'t': 'please confirm plan.md is up to date', 'kind': 'cmd'},
                 {'t': 'with all the latest, including any', 'kind': 'cmd'},
                 {'t': 'design decisions that you made', 'kind': 'cmd'},
                 {'t': '  the checklist already says WHAT got built', 'kind': 'out'},
                 {'t': '  nothing says WHY', 'kind': 'warn'}])),
 ('Z13', G('rows', kicker='because think about what is not written down anywhere',
           lines=['*Why* rows, not a blob'], size=88,
           rows=[{'t': 'That we chose it on purpose'},
                 {'t': 'That I nearly did it differently', 'hot': True}])),
 ('Z14', G(H, kicker='if that only lives in the conversation',
           lines=['It *dies* with', 'the conversation'], size=62)),
 ('Z15', G('myth', kicker='so the plan stops being a to-do list',
           lines=['The project memory'], size=100,
           wrong='the plan is a checklist of what to build',
           right='the plan is where the reasoning lives',
           note='in a file, outside the thing that forgets')),
 ('Z16', G(H, kicker='then it reads the plan, and knows where things stand',
           lines=['Not because it remembers.', 'You left a *note*'], size=52)),
 ('Z17', G(H, kicker='and that is what this week has actually been teaching',
           lines=['Hiding behind a', '*Kanban board*'], size=62, trans='fade')),
 ('Z18', G('myth', kicker='not how to prompt', lines=['The *hardest* part'], size=124,
           wrong='the skill is knowing what to say to it',
           right='the skill is leaving things where the next one can pick them up',
           note='it was always the hardest part of working with people')),
 ('Z19', G(H, kicker='next time, the last three parts',
           lines=['And a model that gets', 'to move your *cards*'], size=54)),
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
        jobs.append(['33', i, TAKES[take], round(ss, 2), round(to, 2), round(win, 3)])
    json.dump(jobs, open('/tmp/l33_jobs.json', 'w'), indent=1)

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

    tsx = f'''// nocode33 — parts 5 to 7: the database, the routes, and the wiring
//
// His half-hour rut did not happen to us and is not staged. The replacement is
// causal rather than anecdotal: L32's blank page came from a missing instruction
// about HOW to check, one sentence fixed it, and the agent left evidence of the
// new check sitting on the board. Then the thirty seconds nobody can outsource —
// drag, log out, hard reload, sign in, and the card is still where you put it.

import React from 'react';
import {{Deck, Slide, deckFrames}} from './kit';

export const DURS = [
{chr(10).join('  ' + ', '.join(f'{d:.3f}' for d in durs[i:i+8]) + ',' for i in range(0, len(durs), 8))}
];
export const FILES = [
{chr(10).join('  ' + ', '.join(f"'{n}'" for n in names[i:i+8]) + ',' for i in range(0, len(names), 8))}
].map((n) => `${{n}}.mp3`);
export const GAP = {gap:.3f};
export const L33_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode33/shots/${{n}}.mp4`;

const SLIDES: Slide[] = [
{chr(10).join(lines)}
];

export const Nocode33: React.FC = () => (
  <Deck slides={{SLIDES}} durs={{DURS}} voDir="nocode33/vo" files={{FILES}} gap={{GAP}} />
);
'''
    out = os.path.join(R, 'renderer/src/nocode/l33.tsx')
    open(out, 'w').write(tsx)
    print("wrote", out)
    print("wrote /tmp/l33_jobs.json")


if __name__ == '__main__':
    main()
