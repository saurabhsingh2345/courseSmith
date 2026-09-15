// nocode19 — the Claude Code build (his L19, 10:22)
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
import {Deck, Slide, deckFrames} from './kit';

export const DURS = [
  6.075, 11.700, 11.781, 24.054, 8.188, 13.505, 13.958, 12.897,
  15.110, 16.508, 8.321, 12.440, 13.246, 14.733, 8.154, 11.400,
  10.468, 15.994, 17.995, 7.259, 4.470, 16.219, 16.058, 10.876,
  17.578, 17.011, 5.698, 16.431, 13.630, 2.896, 10.727, 15.246,
  11.990, 14.989, 8.221, 8.937, 9.691, 10.240, 15.367, 8.867,
  13.500, 11.692, 11.903, 11.117, 19.674, 6.616, 11.916, 12.491,
  11.072, 12.354, 12.498, 11.888, 9.010, 11.675, 13.460, 10.780,
  16.056, 6.919, 4.722, 15.334, 8.472, 13.212, 15.882, 6.292,
  8.177, 15.112, 16.341, 20.764, 10.091,
];
export const FILES = [
  'C1', 'N1', 'N2', 'C2', 'N3', 'N4', 'N5', 'N6',
  'C3', 'N7', 'N8', 'N9', 'N10', 'N11', 'N12', 'C4',
  'C5', 'C6', 'C7', 'N13', 'C8', 'S1', 'S2', 'S3',
  'S4', 'S5', 'S6', 'S7', 'S8', 'S9', 'N14', 'N15',
  'N16', 'N17', 'N18', 'F1', 'F2', 'F3', 'F4', 'F5',
  'F6', 'F7', 'F8', 'F9', 'G1', 'G2', 'G3', 'G4',
  'G5', 'D4', 'G6', 'D5', 'O1', 'O2', 'O3', 'O4',
  'O5', 'N19', 'O6', 'N20', 'C9', 'C10', 'C11', 'C12',
  'N21', 'N22', 'O7', 'C13', 'N23',
].map((n) => `${n}.mp3`);
export const GAP = 0.240;
export const L19_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode19/shots/${n}.mp4`;

const SLIDES: Slide[] = [
  {k: 'shot', src: sh('00'), kicker: 'build three of four', lines: ['A command line tool', 'wearing an *editor*'], size: 62},   // C1  A 0.0-4.7s
  {k: 'term', kicker: 'the ritual, for the last time', lines: ['Rename. *Clone.* Repeat'], size: 62, title: 'copilot kanban, kept as a record', term: [{t: 'mv kanban kanban-copilot', kind: 'cmd'}, {t: 'git clone .../kanban.git', kind: 'cmd'}, {t: 'agents.md   README.md', kind: 'ok'}]},   // N1
  {k: 'shot', src: sh('02')},   // N2  A 4.7-13.7s
  {k: 'rows', kicker: 'and an honest swap', lines: ['Why *this one*', 'and not the other'], size: 58, rows: [{t: 'The reference course uses a tool behind a paid subscription'}, {t: 'This one has a free path in, and more of you already have it'}, {t: 'It makes exactly the same point about a CLI agent in an editor', hot: true}, {t: 'If you do have the other one, run the same brief through it'}]},   // C2
  {k: 'shot', src: sh('04')},   // N3  A 13.7-20.0s
  {k: 'myth', kicker: 'what this actually is', lines: ['Not an editor. *A program*'], size: 62, wrong: 'an editor with an agent bolted into the side of it', right: 'a command line program you happen to be watching through an editor', note: 'the thing doing the work does not need the editor at all'},   // N4
  {k: 'shot', src: sh('06')},   // N5  A 20.0-33.9s
  {k: 'shot', src: sh('07'), kicker: 'so today is the friendly version', lines: ['Next week we take', 'the *training wheels* off'], size: 56},   // N6  A 33.9-46.8s
  {k: 'shot', src: sh('08')},   // C3  A 46.8-61.9s
  {k: 'shot', src: sh('09')},   // N7  A 61.9-78.3s
  {k: 'rows', kicker: 'three tools in', lines: ['The furniture is', '*the same furniture*'], size: 62, rows: [{t: 'A panel on one side that you can drag wider'}, {t: 'A box to type in, and a history above it'}, {t: 'A mode selector and a permission dial'}, {t: 'And a plain text file in the repo that drives all of it', hot: true}]},   // N8
  {k: 'rows', kicker: 'on accounts', lines: ['You can *start* here', 'for nothing'], size: 64, rows: [{t: 'The reference course tells you to skip its version of this lecture'}, {t: 'Because the tool it uses needs a paid plan'}, {t: 'This one has a free tier, and if you already pay, you have it', hot: true}]},   // N9
  {k: 'ladder', kicker: 'and you can predict the options by now', lines: ['Four modes, *one dial*'], size: 64, items: ['Plan — talk, do not touch anything', 'Manual — ask me before every action', 'Edit automatically — change files, ask before commands', 'Auto — full access, ask nothing'], pick: 3, tone: '#e53935'},   // N10
  {k: 'shot', src: sh('13')},   // N11  A 78.3-93.0s
  {k: 'rows', kicker: 'naming it, with the date on the frame', lines: ['*Claude Opus*, high effort'], size: 68, rows: [{t: 'A one-million-token context window'}, {t: 'Thinking budget turned up, deliberately'}, {t: 'The same model family that ran the Cursor build', hot: true}], stamp: '28 august 2026 — check this yourself, it will have moved'},   // N12
  {k: 'shot', src: sh('15')},   // C4  A 93.0-104.4s
  {k: 'shot', src: sh('16')},   // C5  A 104.4-115.0s
  {k: 'curve', kicker: 'and the obvious answer is wrong', lines: ['More thinking is *not* better'], size: 62, bands: [{from: 0.0, to: 0.35, label: 'thinking helps', tone: '#3b82f6'}, {from: 0.35, to: 0.62, label: 'best value', tone: '#f5c518'}, {from: 0.62, to: 1.0, label: 'agonising', tone: '#e53935'}], mark: 0.48, yLabel: 'quality', xLabel: 'thinking budget', caption: 'the curve has a peak in it, not a slope'},   // C6
  {k: 'rows', kicker: 'so where the peak sits depends on the job', lines: ['Turn it *up* or leave it'], size: 64, rows: [{t: 'A gnarly bug in code you cannot read — turn it up'}, {t: 'A small, well specified build like this — let it go'}, {t: 'You will learn the difference from how long you sit waiting', hot: true}]},   // C7
  {k: 'shot', src: sh('19')},   // N13  A 115.0-122.3s
  {k: 'shot', src: sh('20')},   // C8  A 122.3-126.9s
  {k: 'shot', src: sh('21')},   // S1  A 126.9-143.1s
  {k: 'shot', src: sh('22')},   // S2  A 143.1-159.1s
  {k: 'shot', src: sh('23')},   // S3  A 159.1-170.0s
  {k: 'myth', kicker: 'and this is the sentence that earned my trust', lines: ['Checked, not *remembered*'], size: 62, wrong: 'these are the current versions, as far as I recall', right: 'versions verified live against the registry today, not from memory', note: 'a model\'s memory of what is current is frozen at its training date'},   // S4
  {k: 'rows', kicker: 'it also told me what it decided', lines: ['A judgement, *declared*'], size: 64, rows: [{t: 'Picked the stable drag library over the pre-release one'}, {t: 'And over the popular one that is no longer maintained'}, {t: 'Kept our palette rather than inventing its own', hot: true}, {t: 'None of that was asked for. All of it was said out loud'}]},   // S5
  {k: 'shot', src: sh('26')},   // S6  A 575.0-590.0s
  {k: 'shot', src: sh('27')},   // S7  A 590.0-632.1s
  {k: 'rows', kicker: 'same file, same four words, three readings', lines: ['What *plan first* meant'], size: 62, numbered: true, rows: [{t: 'Cursor — plan, then carry straight on'}, {t: 'Copilot — plan, then carry straight on'}, {t: 'Claude Code — plan, then stop and check with me', hot: true}], stamp: 'and the cheapest moment to change your mind is before it starts'},   // S8
  {k: 'shot', src: sh('29')},   // S9  A 632.1-640.0s
  {k: 'shot', src: sh('30')},   // N14  A 170.0-330.0s
  {k: 'rows', kicker: 'and do not let the speed-up hide it', lines: ['The loop is *ten minutes* long'], size: 62, rows: [{t: 'This is the point where you go and do something else'}, {t: 'That is not a flaw, it is the shape of the work now'}, {t: 'You write the brief carefully BECAUSE the loop is slow', hot: true}]},   // N15
  {k: 'shot', src: sh('32')},   // N16  A 655.0-691.3s
  {k: 'compact', kicker: 'what happens when the memory fills', lines: ['It throws away *the beginning*'], size: 58, blocks: 24, keep: 9, at: 0.62, summary: 'everything before this is compressed into a summary', freedLabel: 'and your original instructions were at the start', caption: 'which is why a long session starts making decisions you ruled out'},   // N17
  {k: 'editor', kicker: 'and the strongest argument for the file', lines: ['A message is forgotten.', 'A *file* is re-read'], size: 54, file: 'agents.md', start: 1, rows: [{t: '# Kanban — project brief', kind: 'h1'}, {t: '', kind: 'text'}, {t: 'read again at the start of every turn', kind: 'important'}, {t: 'not once, at the beginning, and then hoped for', kind: 'important'}], focus: [2, 3]},   // N18
  {k: 'shot', src: sh('35')},   // F1  A 691.3-718.6s
  {k: 'shot', src: sh('36')},   // F2  A 718.6-748.1s
  {k: 'rows', kicker: 'remember these from last lecture', lines: ['The *four steps*'], size: 84, numbered: true, rows: [{t: 'Reproduce the problem'}, {t: 'Prove you have reproduced it'}, {t: 'Find the root cause'}, {t: 'Fix it — and demonstrate the fix', hot: true}]},   // F3
  {k: 'shot', src: sh('38')},   // F4  A 748.1-780.0s
  {k: 'flow', kicker: 'and this is what proving a cause looks like', lines: ['A *chain*, not a guess'], size: 58, nodes: [{t: 'commit mid-drag', sub: 'the move is applied early'}, {t: 'layout re-renders', sub: 'under the cursor', tone: 'yellow'}, {t: 'keyboard sensor re-collides', sub: 'with the shifted layout'}, {t: 'card bounces back', sub: 'the symptom you saw', tone: 'red'}], caption: 'every link is checkable — that is the difference from a guess'},   // F5
  {k: 'shot', src: sh('40')},   // F6  A 964.4-1005.2s
  {k: 'rows', kicker: 'against the card from last lecture', lines: ['*Four* for four'], size: 76, numbered: true, rows: [{t: 'Reproduced it — in a real browser'}, {t: 'Proved the cause — a chain, not a location'}, {t: 'Fixed it — commit the move at the end, not mid-drag'}, {t: 'Demonstrated it — re-ran the failing check', hot: true}], stamp: 'nobody asked it to'},   // F7
  {k: 'shot', src: sh('42'), kicker: 'so the instruction is not a workaround', lines: ['It is a description of', 'what *good* looks like'], size: 54},   // F8  A 1005.2-1041.3s
  {k: 'shot', src: sh('43')},   // F9  A 1041.3-1075.0s
  {k: 'shot', src: sh('44')},   // G1  A 1550.0-1576.7s
  {k: 'rows', kicker: 'and that is the whole app', lines: ['*962 lines*, four files'], size: 72, rows: [{t: 'The production build passes'}, {t: 'The type check passes, the linter is clean'}, {t: '27 unit tests pass, and 10 browser tests in real Chromium'}, {t: '38 success criteria, ticked off against the file it wrote', hot: true}], stamp: 'every gate named before it started'},   // G2
  {k: 'shot', src: sh('46')},   // G3  A 1576.7-1593.0s
  {k: 'shot', src: sh('47')},   // G4  A 1593.0-1610.1s
  {k: 'shot', src: sh('48')},   // G5  A 1610.1-1625.3s
  {k: 'myth', kicker: 'the other half of the discipline', lines: ['Broken, or *by design*'], size: 64, wrong: 'two arrow keys do the same thing — that is a bug, fix it', right: 'the library walks the cards in a line, not a grid — documented', note: 'five minutes of checking before you demand a fix'},   // D4
  {k: 'shot', src: sh('50'), kicker: 'that is step two doing its job', lines: ['A tool willing to say', '*I was wrong* about that'], size: 54},   // G6  A 1625.3-1642.4s
  {k: 'rows', kicker: 'and both failures cost the same afternoon', lines: ['Two ways to *waste a day*'], size: 62, rows: [{t: 'Believing a fix that was never demonstrated'}, {t: 'Demanding a fix for something that was never broken', hot: true}]},   // D5
  {k: 'shot', src: sh('52')},   // O1  S 0.0-18.1s
  {k: 'shot', src: sh('53')},   // O2  S 18.1-41.3s
  {k: 'myth', kicker: 'one design decision differs, and it is better', lines: ['Where the *button* goes'], size: 64, wrong: 'one Add button at the top, then asking which column you meant', right: 'Add a card inside each column, where the card is going to land', note: 'the brief chose neither — two tools read one sentence differently'},   // O3
  {k: 'rows', kicker: 'which is worth noticing', lines: ['The brief did *not* say'], size: 68, rows: [{t: 'It said: add a card. Four words'}, {t: 'Everything about where, and how, was decided for you'}, {t: 'Leave the decisions you do not care about unspecified', hot: true}]},   // O4
  {k: 'shot', src: sh('56')},   // O5  S 41.3-73.1s
  {k: 'shot', src: sh('57')},   // N19  S 73.1-87.1s
  {k: 'rows', kicker: 'third time today', lines: ['The *same five*, in order'], size: 62, numbered: true, rows: [{t: 'Drag a card between columns'}, {t: 'Reorder a card inside a column'}, {t: 'Delete a card'}, {t: 'Rename a column'}, {t: 'Add a card, with a description'}]},   // O6
  {k: 'shot', src: sh('59')},   // N20  S 87.1-117.5s
  {k: 'shot', src: sh('60')},   // C9  A 1642.4-1654.1s
  {k: 'shot', src: sh('61'), kicker: 'and now the honest part', lines: ['I could not build', 'this *by hand*'], size: 62},   // C10  A 1654.1-1672.2s
  {k: 'rows', kicker: 'which is the whole point of the course', lines: ['*Four* jobs. None of them', 'require reading code'], size: 54, numbered: true, rows: [{t: 'Decide what gets built'}, {t: 'State it precisely enough to be acted on'}, {t: 'Check whether the thing in front of you does it'}, {t: 'Send it back when it does not', hot: true}]},   // C11
  {k: 'head', kicker: 'every one of those is a judgement', lines: ['Not one of them is', '*reading the code*'], size: 62},   // C12
  {k: 'shot', src: sh('64')},   // N21  A 1672.2-1683.5s
  {k: 'rows', kicker: 'and notice what nobody did', lines: ['Nobody wrote *any* of it'], size: 64, rows: [{t: 'No code written by a person'}, {t: 'No bug debugged by a person'}, {t: 'No library, folder layout or state decision made by a person'}, {t: 'All of it decided by something reading a file we wrote', hot: true}]},   // N22
  {k: 'matrix', kicker: 'three down, one to go', lines: ['The grid so *far*'], size: 58, cols: [{head: 'Cursor', sub: 'Opus, chosen'}, {head: 'Copilot', sub: 'free tier, Auto'}, {head: 'Claude Code', sub: 'Opus, high effort'}, {head: 'Antigravity', sub: 'next'}], rows: [{label: 'the brief', cells: ['agents.md', 'agents.md', 'agents.md', '']}, {label: 'the prompt', cells: ['go ahead and plan', 'go ahead and plan', 'go ahead and plan', '']}, {label: 'asked me', cells: ['never', 'eight times', 'once, before starting', '']}, {label: 'the plan', cells: ['in the panel', 'in the panel', 'written to a file', '']}, {label: 'five checks', cells: ['all five', 'all five', 'all five', '']}], upto: 3, dateline: '28 august 2026'},   // O7
  {k: 'rows', kicker: 'and the caveat, which is the useful part', lines: ['This was an *easy* job'], size: 64, rows: [{t: 'Small, clean, and described in one page'}, {t: 'All four tools will look impressive on it'}, {t: 'A real system has requirements that contradict each other'}, {t: 'It is day three of three weeks for a reason', hot: true}]},   // C13
  {k: 'shot', src: sh('68'), kicker: 'one to go, and it is the odd one out', lines: ['The *holdout*'], size: 104},   // N23  A 1683.5-1697.0s
];

export const Nocode19: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode19/vo" files={FILES} gap={GAP} />
);
