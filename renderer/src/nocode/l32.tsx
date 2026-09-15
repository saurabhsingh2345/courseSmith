// nocode31 — planning and scaffolding (his L31, 11:43)
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
import {Deck, Slide, deckFrames} from './kit';

export const DURS = [
  12.084, 6.754, 12.044, 11.166, 11.941, 6.915, 11.551, 16.458,
  11.716, 4.482, 2.675, 11.193, 11.755, 11.127, 5.677, 10.194,
  7.021, 6.633, 12.059, 8.994, 12.952, 9.378, 6.206, 12.900,
  7.662, 8.621, 5.510, 11.452, 9.064, 6.451, 9.606, 10.869,
  11.126, 8.404, 12.522, 10.104, 4.790, 10.613, 11.047, 3.342,
  3.541, 6.854, 6.102, 10.386, 6.814, 9.359, 8.950, 7.902,
  11.445, 10.816, 4.134, 14.912, 10.842, 7.272, 6.822, 4.464,
  9.733, 3.686, 3.923, 8.425, 11.208, 8.777, 5.220, 13.826,
  10.043, 4.892, 7.570, 9.909, 13.702, 8.354, 10.452, 11.673,
  10.407, 12.644, 12.108, 11.731, 8.103, 9.582, 8.487, 9.044,
  7.477, 8.720, 5.153, 6.460,
];
export const FILES = [
  'A1', 'A2', 'A3', 'A4', 'A5', 'A6', 'A7', 'A8',
  'A9', 'B1', 'B2', 'B3', 'B4', 'B5', 'B6', 'B7',
  'B8', 'B9', 'B10', 'B11', 'B12', 'B13', 'C1', 'C2',
  'C3', 'C4', 'C5', 'C6', 'C7', 'D1', 'D2', 'D3',
  'D4', 'D5', 'D6', 'D7', 'E1', 'E2', 'E3', 'E4',
  'X1', 'X2', 'X3', 'X4', 'X5', 'X6', 'X7', 'X8',
  'X9', 'X10', 'X11', 'X12', 'X13', 'X14', 'X15', 'X16',
  'X17', 'E5', 'E6', 'E7', 'E8', 'E9', 'F1', 'F2',
  'F3', 'F4', 'F5', 'F6', 'F7', 'F8', 'F9', 'F10',
  'F10b', 'F10c', 'F10d', 'F11', 'F12', 'F13', 'G1', 'G2',
  'G3', 'G4', 'G5', 'G6',
].map((n) => `${n}.mp3`);
export const GAP = 0.240;
export const L32_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode32/shots/${n}.mp4`;

const SLIDES: Slide[] = [
  {k: 'head', kicker: 'last time ended on a promise', lines: ['Three things', 'go *wrong*'], size: 92, trans: 'fade'},   // A1
  {k: 'shot', src: sh('01'), kicker: 'the first one', lines: ['Permission *denied*'], size: 104},   // A2  P 150.0-188.0s
  {k: 'myth', kicker: 'and it sounds worse than it is', lines: ['A *flag* on a file'], size: 88, wrong: 'the script is broken, or my machine is misconfigured', right: 'a file carries a flag saying whether it may be run, and nobody set it', note: 'the file was written a minute ago'},   // A3
  {k: 'shot', src: sh('03')},   // A4  P 188.0-250.0s
  {k: 'rows', kicker: 'and this is the pattern to internalise', lines: ['Hand over the *error*'], size: 88, rows: [{t: 'You do not need to know the fix'}, {t: 'You need to hand it the error, exactly as it appeared'}, {t: 'It has read a million of these', hot: true}]},   // A5
  {k: 'shot', src: sh('05'), kicker: 'second thing', lines: ['Address already *in use*'], size: 84},   // A6  P 250.0-283.0s
  {k: 'myth', kicker: 'a port is a numbered door', lines: ['One program, one *doorway*'], size: 68, wrong: 'address already in use — something is wrong with the code', right: 'something else is standing in that doorway and never left', note: 'it is a problem with the room, not the building'},   // A7
  {k: 'shot', src: sh('07')},   // A8  P 283.0-360.0s
  {k: 'head', kicker: 'two errors, two recoveries, no help from me', lines: ['The third one is', 'a different *animal*'], size: 62, trans: 'push'},   // A9
  {k: 'head', kicker: 'before I let it move on', lines: ['The prompt', 'to *steal*'], size: 104},   // B1
  {k: 'shot', src: sh('10'), kicker: 'here it is', lines: ['Tell me how I can', '*test* this myself'], size: 76},   // B2  K 30.0-46.3s
  {k: 'myth', kicker: 'and look at what that sentence does', lines: ['Not *is it done*'], size: 96, wrong: 'is part four finished?', right: 'tell me how I can check part four myself', note: 'the verification stops living inside the thing being verified'},   // B3
  {k: 'rows', kicker: 'and for this audience it matters more, not less', lines: ['A currency you can *spend*'], size: 68, rows: [{t: 'I cannot read its tests'}, {t: 'I can open a browser and see whether a page loads'}, {t: 'So I want the answer in the second form', hot: true}]},   // B4
  {k: 'shot', src: sh('13')},   // B5  K 46.3-110.0s
  {k: 'shot', src: sh('14'), kicker: 'and then the check nobody does', lines: ['Are the *success criteria*', 'actually met?'], size: 60},   // B6  K 110.0-139.3s
  {k: 'rows', kicker: 'and it only works because the plan wrote them down', lines: ['A vague plan', 'cannot be *audited*'], size: 62, rows: [{t: 'A plan with a checklist can be'}, {t: 'And the audit costs one sentence', hot: true}]},   // B7
  {k: 'head', kicker: 'there is a second version I use constantly', lines: ['Ask it to *explain*', 'a decision'], size: 72},   // B8
  {k: 'myth', kicker: 'and the wording changes the answer', lines: ['Explain, not *justify*'], size: 84, wrong: 'why did you do it that way?', right: 'explain the reasoning behind that choice', note: 'one invites a defence, the other invites the reasoning'},   // B9
  {k: 'rows', kicker: 'because it makes dozens of choices you never see', lines: ['Most are *fine*'], size: 96, rows: [{t: 'Which library'}, {t: 'Which structure'}, {t: 'Which shortcut'}, {t: 'And some are it doing what works rather than what you asked', hot: true}]},   // B10
  {k: 'shot', src: sh('19')},   // B11  K 139.3-185.0s
  {k: 'head', kicker: 'so when a file appears you did not expect', lines: ['*Ask* why'], size: 124},   // B12
  {k: 'rows', kicker: 'and you can do all of that without reading code', lines: ['Interrogating the', '*reasoning*'], size: 68, rows: [{t: 'Not auditing the implementation'}, {t: 'Which is a job you can actually do', hot: true}]},   // B13
  {k: 'shot', src: sh('22'), kicker: 'part three', lines: ['Serve the board', 'from the *back end*'], size: 72},   // C1  P 360.0-401.9s
  {k: 'rows', kicker: 'and it looks like nothing, which is why it is a whole part', lines: ['Two things that', 'do not *know* each other'], size: 54, rows: [{t: 'A front end that draws a board'}, {t: 'A back end that answers on a port'}, {t: 'Part three is the introduction', hot: true}]},   // C2
  {k: 'head', kicker: 'and after this', lines: ['*One* address', 'serves the whole thing'], size: 68},   // C3
  {k: 'shot', src: sh('25')},   // C4  P 401.9-459.5s
  {k: 'head', kicker: 'and that phrase, out of the box', lines: ['Worth unpacking', '*once*'], size: 76},   // C5
  {k: 'rows', kicker: 'we said it on day three', lines: ['The same *object*'], size: 96, rows: [{t: 'A box with the application and everything it needs, sealed together'}, {t: 'It behaves the same here, on your machine, and on a server'}, {t: 'So this is not a rehearsal of production. It is production.', hot: true}]},   // C6
  {k: 'shot', src: sh('28')},   // C7  P 459.5-520.0s
  {k: 'shot', src: sh('29'), kicker: 'now the diffs', lines: ['Green *added*', 'Red *removed*'], size: 92},   // D1  Q 40.0-83.3s
  {k: 'shot', src: sh('30')},   // D2  Q 83.3-146.9s
  {k: 'myth', kicker: 'and I am not going to pretend', lines: ['Read *every line*?'], size: 92, wrong: 'you should review every single line before accepting it', right: 'nobody shipping at this pace does that, including me', note: 'a rule you will not follow is not a rule'},   // D3
  {k: 'rows', kicker: 'so here is the one I actually use', lines: ['Bulk, then *slow down*'], size: 76, rows: [{t: 'Accept in bulk while the work is ordinary'}, {t: 'Look properly when it touches something you would not want wrong', hot: true}]},   // D4
  {k: 'rows', kicker: 'and that second list is short and specific', lines: ['Where the *afternoons* go'], size: 68, numbered: true, rows: [{t: 'Money'}, {t: 'Anything with a password in it'}, {t: 'Anything that deletes'}, {t: 'The first time we plug in the AI', hot: true}]},   // D5
  {k: 'pct', kicker: 'because risk is not spread evenly across a diff', pct: 90, sub: 'plumbing that either works or obviously does not — the other ten percent is where the afternoons go'},   // D6
  {k: 'head', kicker: 'and the phrase I keep coming back to', lines: ['Watch it like a *hawk*.', 'Not read it like a lawyer'], size: 54, accent: '#f5c518'},   // D7
  {k: 'shot', src: sh('36'), kicker: 'part four', lines: ['Sign in. One user.', 'A way *out*'], size: 80},   // E1  P 520.0-558.0s
  {k: 'shot', src: sh('37')},   // E2  P 558.0-640.0s
  {k: 'rows', kicker: 'and it writes a test for it, unprompted', lines: ['A *good* test'], size: 100, rows: [{t: 'Wrong password, refused'}, {t: 'Right password, and the board'}, {t: 'Two things that matter, and nothing else', hot: true}]},   // E3
  {k: 'shot', src: sh('39'), kicker: 'and it reports back', lines: ['Part four complete.', 'And it *checked*'], size: 68},   // E4  Q 146.9-170.1s
  {k: 'shot', src: sh('40')},   // X1  W 2.0-26.0s
  {k: 'head', kicker: 'and this is what I get', lines: ['*Nothing*'], size: 150, trans: 'fade'},   // X2
  {k: 'head', kicker: 'so let us slow this right down', lines: ['How I found it matters', 'more than *what* it was'], size: 52},   // X3
  {k: 'rows', kicker: 'because its checks passed', lines: ['By its *measurement*,', 'part four worked'], size: 58, rows: [{t: 'It asked for the page from the command line'}, {t: 'It got back the sign-in form, complete and correct'}, {t: 'Every check it could run, it ran, and they passed', hot: true}]},   // X4
  {k: 'myth', kicker: 'and here is the gap', lines: ['It cannot *open* a browser'], size: 72, wrong: 'it tested the page, so the page works', right: 'it tested the page the only way it can, and the fault only appears in a browser', note: 'the one place it cannot look is the one place it shows'},   // X5
  {k: 'head', kicker: 'which is the shape of nearly every bad afternoon', lines: ['Not that it *lied*.', 'It measured the wrong thing'], size: 52, accent: '#f5c518'},   // X6
  {k: 'shot', src: sh('46'), kicker: 'so I describe it the way day three taught', lines: ['Did. Expected.', '*Happened*'], size: 84},   // X7  F 4.0-60.0s
  {k: 'rows', kicker: 'and then the two facts that make it strange', lines: ['Where the views', '*disagree*'], size: 68, numbered: true, rows: [{t: 'Blank in the browser'}, {t: 'Correct from the command line'}, {t: 'And asked to unpack it as a browser would, the command line fails too', hot: true}]},   // X8
  {k: 'head', kicker: 'that last one is the whole gift', lines: ['I do not know the bug.', 'I know *where* it lives'], size: 56},   // X9
  {k: 'rows', kicker: 'and two clauses are there because of what just happened', lines: ['Prove it. Then check', 'it the *right* way'], size: 58, rows: [{t: 'Prove the cause, do not guess at it'}, {t: 'Confirm the fix the way a browser would, not the way the command line does', hot: true}]},   // X10
  {k: 'shot', src: sh('50'), kicker: 'and here is what it was', lines: ['*Small*, as usual'], size: 116},   // X11  F 60.0-92.0s
  {k: 'flow', kicker: 'the back end fetches the page and passes it on', lines: ['Unpacked on the way *in*'], size: 68, nodes: [{t: 'the front end', sub: 'sends it compressed'}, {t: 'the library between', sub: 'unpacks it on the way in'}, {t: 'our code', sub: 'copies the ORIGINAL labels across'}, {t: 'your browser', sub: 'reads the label and gives up'}], caption: 'unpacked parcel, sealed label'},   // X12
  {k: 'myth', kicker: 'so the parcel is open and the label still says sealed', lines: ['The browser gives up.', '*Silently*'], size: 58, wrong: 'the label says compressed, so unpack it', right: 'it is already unpacked, and unpacking it again fails', note: 'no error on screen. just nothing.'},   // X13
  {k: 'head', kicker: 'and the command line did not care', lines: ['It was never asked', 'to *unpack* anything'], size: 60},   // X14
  {k: 'head', kicker: 'one label', lines: ['An application,', 'or a blank *rectangle*'], size: 62, accent: '#f5c518'},   // X15
  {k: 'shot', src: sh('55')},   // X16  F 300.0-327.0s
  {k: 'myth', kicker: 'and the rule I would carve into the desk', lines: ['Two different *claims*'], size: 88, wrong: 'the check passed, so it works', right: 'its check passing is not the same as it working', note: 'only one of those is yours to make'},   // X17
  {k: 'shot', src: sh('57')},   // E5  W 34.0-58.0s
  {k: 'shot', src: sh('58')},   // E6  W 68.0-92.0s
  {k: 'rows', kicker: 'and be precise about what that proves', lines: ['*Less* than it looks'], size: 100, rows: [{t: 'The sign-in works'}, {t: 'It remembers me between visits'}]},   // E7
  {k: 'myth', kicker: 'and here is what it does not prove', lines: ['There is no *database* yet'], size: 68, wrong: 'the board survived, so the board is saved', right: 'the board is still the one the front end draws for itself', note: 'move a card and it would not survive a reload'},   // E8
  {k: 'head', kicker: 'that is part five, and it is next', lines: ['A working demo is the', 'easiest place to *overbelieve*'], size: 52},   // E9
  {k: 'head', kicker: 'now the checkpoints', lines: ['Not the beat I expected', 'to be *recording*'], size: 60, trans: 'rise'},   // F1
  {k: 'term', kicker: 'last time I made a point of it', lines: ['And I wrote it *again*'], size: 84, title: 'in the prompt, in plain English', term: [{t: 'commit to git after each part,', kind: 'cmd'}, {t: 'with a clear message', kind: 'cmd'}]},   // F2
  {k: 'shot', src: sh('64'), kicker: 'so did it?', lines: ['It *did*'], size: 132},   // F3  G 2.0-23.6s
  {k: 'head', kicker: 'but do not stop at the fact that commits exist', lines: ['Look at what is', '*inside* them'], size: 68},   // F4
  {k: 'shot', src: sh('66')},   // F5  G 23.6-40.0s
  {k: 'rows', kicker: 'and none of that is part three', lines: ['That is part *two*'], size: 104, rows: [{t: 'The Dockerfile'}, {t: 'The back end'}, {t: 'Both of the scripts'}, {t: 'Swept in, because part two never got a checkpoint of its own', hot: true}]},   // F6
  {k: 'myth', kicker: 'does it break anything? no', lines: ['The work is safe.', 'The *return point* is gone'], size: 54, wrong: 'everything is committed, so everything is fine', right: 'there is no longer a point to go back to at the end of part two', note: 'two parts welded into one'},   // F7
  {k: 'shot', src: sh('69'), kicker: 'and one more, which I would have missed', lines: ['In *no commit* at all'], size: 92},   // F8  G 40.0-63.2s
  {k: 'head', kicker: 'small. part of the back end. everything runs', lines: ['A checkpoint I trusted', 'is not *complete*'], size: 54},   // F9
  {k: 'myth', kicker: 'so the lesson is not that it ignored me', lines: ['It *followed* the instruction'], size: 68, wrong: 'it skipped the checkpoints I asked for', right: 'it made them, and the history still has a gap in it', note: 'and the only reason I know is that I read it'},   // F10
  {k: 'head', kicker: 'the same shape I flagged last lecture', lines: ['One lecture *later*', 'than I expected'], size: 60},   // F10b
  {k: 'shot', src: sh('73')},   // F10c  G 63.2-98.0s
  {k: 'rows', kicker: 'and one clarification that tripped me up for years', lines: ['On this machine.', '*Nowhere* else'], size: 68, rows: [{t: 'Nothing has been uploaded'}, {t: 'There is no website involved'}, {t: 'It is a save point in a folder on this laptop', hot: true}]},   // F10d
  {k: 'head', kicker: 'somewhere else is a separate decision', lines: ['I am not protecting', 'against *fire*'], size: 64},   // F11
  {k: 'rows', kicker: 'everything so far has been repeatable', lines: ['Twenty minutes', 'from the *plan*'], size: 68, rows: [{t: 'If I lost all of it, I could make it again'}]},   // F12
  {k: 'myth', kicker: 'and the database is not', lines: ['*Before* the risk'], size: 104, wrong: 'commit when you have finished something', right: 'commit before the part that could hurt', note: 'not after you find out it was risky'},   // F13
  {k: 'rows', kicker: 'so where we are', lines: ['Four of *ten*'], size: 112, numbered: true, rows: [{t: 'The box builds'}, {t: 'The back end answers'}, {t: 'The board is served from it'}, {t: 'And you have to sign in to see it'}]},   // G1
  {k: 'head', kicker: 'and nothing clever has happened yet', lines: ['By *design*'], size: 132},   // G2
  {k: 'shot', src: sh('80'), kicker: 'next, the database', lines: ['The only part that', '*stops* and asks'], size: 68},   // G3  Q 170.1-220.0s
  {k: 'head', kicker: 'and one thing to carry with you', lines: ['All of it was', '*recovery*'], size: 88, trans: 'rise'},   // G4
  {k: 'myth', kicker: 'a permission flag, a busy port, a blank page', lines: ['Notice, not *write*'], size: 96, wrong: 'I needed to know how to fix those', right: 'I needed to notice they were wrong', note: 'not one of them required me to write code'},   // G5
  {k: 'head', kicker: 'that is the job', lines: ['And harder to *fake*', 'than programming'], size: 58},   // G6
];

export const Nocode32: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode32/vo" files={FILES} gap={GAP} />
);
