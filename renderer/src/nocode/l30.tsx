// nocode30 — the full-stack setup (his L30, 11:45)
//
// Mixed cut. Two of his beats cannot be filmed here: the Copilot settings/usage
// page shows his name, plan and spend, and VS Code's Welcome screen lists
// `~/Desktop/enfec_subs` and `~/Desktop/self` in Recent. A take that caught the
// Welcome screen was deleted; every take since closes that tab before rolling.
//
// What is filmed is what matters — a genuine `git clone` from the neutral org,
// the project's shape, and a slow walk down agents.md, which is the artefact the
// whole of Day 5 runs on.
//
// Continuity: A9b says the inherited board was picked NOT because it won. L21
// shipped with "the honest headline is not that one of these tools won", and an
// earlier draft of A9 said "the one that came out best" — which would have
// contradicted it two lectures later.

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';

export const DURS = [
  9.353, 12.657, 13.712, 10.549, 5.986, 11.158, 12.366, 15.025,
  10.888, 15.890, 13.993, 14.268, 12.403, 15.101, 10.231, 10.915,
  10.505, 9.528, 5.739, 9.043, 5.925, 9.445, 11.357, 11.075,
  10.161, 13.491, 9.977, 10.611, 6.299, 15.668, 15.434, 11.580,
  6.025, 8.630, 16.991, 9.625, 8.177, 8.300, 11.086, 8.217,
  14.005, 8.094, 10.428, 13.177, 11.910, 9.421, 6.814, 10.520,
  10.551, 6.507, 12.017, 14.088, 15.835, 7.851, 13.677, 9.752,
  16.373, 5.189,
];
export const FILES = [
  'A1', 'A2', 'A3', 'A4', 'A5', 'A6', 'A7', 'A8',
  'A9', 'A9b', 'A10', 'A11', 'A12', 'A13', 'A14', 'C1',
  'C2', 'C3', 'C4', 'C5', 'C6', 'C7', 'C8', 'C9',
  'C10', 'C11', 'C12', 'B1', 'B2', 'B3', 'B4', 'B5',
  'B6', 'B7', 'B8', 'B9', 'B10', 'B11', 'B12', 'B13',
  'B14', 'B15', 'B16', 'B17', 'B18', 'B19', 'B20', 'B21',
  'B22', 'D1', 'D2', 'D3', 'D4', 'D5', 'D6', 'D7',
  'D8', 'D9',
].map((n) => `${n}.mp3`);
export const GAP = 0.240;
export const L30_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode30/shots/${n}.mp4`;

const SLIDES: Slide[] = [
  {k: 'head', kicker: 'day five', lines: ['Today we follow', 'the *rules*'], size: 104, trans: 'fade'},   // A1
  {k: 'gauge', kicker: 'before anything else, the boring one', lines: ['Find your *usage* page'], size: 76, pct: 38, limit: 100, label: 'requests used this month', limitLabel: 'allowance', caption: 'look at it before you start, not when it runs out'},   // A2
  {k: 'myth', kicker: 'because of how running out actually feels', lines: ['It gets *worse*,', 'not broken'], size: 62, wrong: 'you hit a wall and get a clear message', right: 'the good model stops answering and quality quietly drops', note: 'you will blame the model for an hour before you check'},   // A3
  {k: 'head', kicker: 'and if you are on a paid plan', lines: ['Set a *budget*'], size: 112},   // A4
  {k: 'head', kicker: 'now, the project', lines: ['Why we are *not*', 'starting from nothing'], size: 62, trans: 'push'},   // A5
  {k: 'rows', kicker: 'every build this week began with an empty folder', lines: ['Which is their *best* case'], size: 68, rows: [{t: 'No existing decisions'}, {t: 'No existing style'}, {t: 'Nothing to be careful around'}, {t: 'They scaffold beautifully', hot: true}]},   // A6
  {k: 'rows', kicker: 'and real work almost never looks like that', lines: ['There is *always*', 'something already there'], size: 58, rows: [{t: 'Somebody else\'s code'}, {t: 'A half-finished thing from last year'}, {t: 'A system nobody has touched since its author left', hot: true}]},   // A7
  {k: 'myth', kicker: 'so this project mixes both, deliberately', lines: ['Inherit, then *build*'], size: 76, wrong: 'start clean, where the agent is strongest', right: 'inherit working code and grow a real application around it', note: 'harder, and much closer to the job you actually have'},   // A8
  {k: 'shot', src: sh('08')},   // A9  B 2.0-11.2s
  {k: 'rows', kicker: 'and before anyone asks why that one', lines: ['*Not* because it won'], size: 88, rows: [{t: 'We said at the end of day three that none of them won'}, {t: 'Any of the four would have done this job'}, {t: 'I picked the one easiest to read on camera', hot: true}, {t: 'Which is how real decisions usually get made'}]},   // A9b
  {k: 'head', kicker: 'and from here on, pretend you did not write it', lines: ['Another team built this.', 'You were *handed* it'], size: 54},   // A10
  {k: 'myth', kicker: 'and be clear about what it actually is', lines: ['A convincing *demo*'], size: 80, wrong: 'an application that needs finishing', right: 'a front end with nothing underneath it — reload and it forgets', note: 'no back end, no database, nothing saved anywhere'},   // A11
  {k: 'rows', kicker: 'so, the mission', lines: ['Turn it into', 'something *real*'], size: 68, rows: [{t: 'A proper front end and back end'}, {t: 'A database, so it remembers'}, {t: 'An API between them'}, {t: 'A sign-in, so it belongs to somebody', hot: true}]},   // A12
  {k: 'head', kicker: 'and one more, which for once is the right feature', lines: ['An assistant that can', '*change* the board'], size: 58, accent: '#f5c518'},   // A13
  {k: 'head', kicker: 'small, but real', lines: ['Ten steps.', 'Testing after *each* one'], size: 68, trans: 'rise'},   // A14
  {k: 'head', kicker: 'so let us get it onto the machine', lines: ['*Clone* it'], size: 124, trans: 'push'},   // C1
  {k: 'shot', src: sh('16')},   // C2  C 2.0-16.3s
  {k: 'shot', src: sh('17')},   // C3  C 16.3-29.2s
  {k: 'shot', src: sh('18'), kicker: 'if you have never done it before', lines: ['It ought to be harder.', 'It is *not*'], size: 58},   // C4  C 29.2-37.2s
  {k: 'shot', src: sh('19')},   // C5  B 11.2-18.9s
  {k: 'shot', src: sh('20')},   // C6  B 18.9-24.0s
  {k: 'hierarchy', kicker: 'and three of them are empty', lines: ['Rooms with the doors', 'already *labelled*'], size: 54, nodes: [{t: 'pm', depth: 0, kind: 'dir'}, {t: 'frontend', depth: 1, kind: 'dir'}, {t: 'backend', depth: 1, kind: 'dir'}, {t: 'scripts', depth: 1, kind: 'dir'}, {t: 'docs', depth: 1, kind: 'dir'}, {t: 'agents.md', depth: 1, kind: 'agents'}], active: 1, caption: 'only the first one currently does anything'},   // C7
  {k: 'myth', kicker: 'and that is deliberate', lines: ['An empty folder', 'is an *instruction*'], size: 62, wrong: 'make folders when you need them', right: 'the back end goes HERE, not wherever it feels like', note: 'costs nothing, and prevents an argument later'},   // C8
  {k: 'shot', src: sh('23')},   // C9  C 37.2-52.2s
  {k: 'shot', src: sh('24')},   // C10  C 52.2-66.0s
  {k: 'myth', kicker: 'so the secret exists and can never be uploaded', lines: ['Not something you', '*remember* to do'], size: 58, wrong: 'be careful never to commit the .env file', right: 'list it in .gitignore once, and git cannot see it', note: 'set up at the start, and then impossible to get wrong'},   // C11
  {k: 'head', kicker: 'which is why I am showing you an empty project', lines: ['Every decision here is', 'invisible *later*'], size: 54},   // C12
  {k: 'head', kicker: 'and now the file this week has been about', lines: ['The *brief*'], size: 132, trans: 'fade'},   // B1
  {k: 'shot', src: sh('28')},   // B2  B 24.0-29.8s
  {k: 'shot', src: sh('29')},   // B3  B 29.8-43.8s
  {k: 'shot', src: sh('30')},   // B4  B 43.8-57.7s
  {k: 'head', kicker: 'and the line underneath it', lines: ['Nine words that save', 'somebody a *rewrite*'], size: 54, accent: '#f5c518'},   // B5
  {k: 'shot', src: sh('32')},   // B6  B 57.7-63.2s
  {k: 'myth', kicker: 'to the half of you worried about this section', lines: ['You do *not* need', 'to know this'], size: 62, wrong: 'you must understand the stack to specify it', right: 'ask the AI what it recommends, and take that', note: 'that is a real answer and nobody will think less of you'},   // B7
  {k: 'shot', src: sh('34')},   // B8  B 63.2-78.5s
  {k: 'head', kicker: 'and each of those is a fence', lines: ['A decision *written down*', 'is one it cannot re-make'], size: 52},   // B9
  {k: 'head', kicker: 'now the section I want you to copy', lines: ['*Starting point*'], size: 124, trans: 'rise'},   // B10
  {k: 'shot', src: sh('37')},   // B11  B 78.5-86.0s
  {k: 'rows', kicker: 'without it, the agent has to guess', lines: ['Is this *sacred*,', 'finished, or a mistake?'], size: 58, rows: [{t: 'It finds a working board in the folder'}, {t: 'And no statement about what that board is'}, {t: 'With the paragraph, there is no guessing at all', hot: true}]},   // B12
  {k: 'myth', kicker: 'and a tiny thing worth more than it looks', lines: ['I changed *one* heading'], size: 76, wrong: '## Current state', right: '## Starting point', note: 'same information, expiry date removed'},   // B13
  {k: 'rows', kicker: 'because in three hours', lines: ['*Current state* will', 'be quietly false'], size: 58, rows: [{t: 'The state will be completely different'}, {t: 'The heading will still say "current"'}, {t: '"Starting point" can never go stale', hot: true}]},   // B14
  {k: 'head', kicker: 'that is the level of care worth taking', lines: ['Which of your words will', 'still be *true* tomorrow'], size: 52},   // B15
  {k: 'shot', src: sh('42')},   // B16  B 86.0-95.5s
  {k: 'shot', src: sh('43')},   // B17  B 95.5-107.3s
  {k: 'rows', kicker: 'and you should recognise that one', lines: ['*Promoted* into the file'], size: 76, numbered: true, rows: [{t: 'Reproduce the problem'}, {t: 'Prove the root cause'}, {t: 'Fix it'}, {t: 'Demonstrate the fix', hot: true}], stamp: 'two days ago you typed this by hand, in a panic'},   // B18
  {k: 'myth', kicker: 'and one line that is in nobody else\'s template', lines: ['Permission to *push back*'], size: 68, wrong: 'follow the brief exactly', right: 'if an instruction here is a bad idea, say so rather than following it into a corner', note: 'remember this one'},   // B19
  {k: 'head', kicker: 'it is going to matter', lines: ['And not in the way', 'I *expected*'], size: 68},   // B20
  {k: 'shot', src: sh('47')},   // B21  B 107.3-116.8s
  {k: 'myth', kicker: 'and that last clause is load-bearing', lines: ['The plan *survives*', 'the conversation'], size: 58, wrong: 'keep the plan updated', right: 'keep it updated INCLUDING any design decisions you make', note: 'we prove this later today by throwing the conversation away'},   // B22
  {k: 'head', kicker: 'two things in the panel', lines: ['Both you have seen.', 'Neither *explained*'], size: 54, trans: 'push'},   // D1
  {k: 'ladder', kicker: 'the first is the mode', lines: ['One dropdown *apart*'], size: 80, items: ['a version that only talks to you', 'a version that reads, writes and runs things'], pick: 1, tone: '#e53935'},   // D2
  {k: 'rows', kicker: 'and next to the model selector', lines: ['*Manage models*'], size: 92, rows: [{t: 'Point it at a model running on your own machine'}, {t: 'No account, no network, no cost'}, {t: 'It answers the question I am asked most', hot: true}]},   // D3
  {k: 'myth', kicker: 'so, does any of this work without a subscription', lines: ['Yes. Slower, and *on you*'], size: 62, wrong: 'you need a paid plan to do any of this', right: 'a local model works — it needs a strong machine and setup effort', note: 'behind the big paid models, but real'},   // D4
  {k: 'shot', src: sh('53')},   // D5  B 116.8-124.0s
  {k: 'compact', kicker: 'we drew this on day two', lines: ['Watch it *fill up*'], size: 88, blocks: 24, keep: 10, at: 0.6, summary: 'everything before this gets compressed', freedLabel: 'and your instructions were at the start', caption: 'today you watch it fill over a real build — and we act on it'},   // D6
  {k: 'head', kicker: 'this is the last lecture of setting up', lines: ['Everything after this', 'is *building*'], size: 62},   // D7
  {k: 'rows', kicker: 'and let me say plainly what we are demonstrating', lines: ['Not that an AI', 'can *write code*'], size: 62, rows: [{t: 'You have watched that all week'}, {t: 'Somebody who cannot write this code can still direct it'}, {t: 'Check it, and end up with something real', hot: true}]},   // D8
  {k: 'head', kicker: 'the whole promise of the course', lines: ['It survives the next', 'four lectures, or it *does not*'], size: 52, trans: 'rise'},   // D9
];

export const Nocode30: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode30/vo" files={FILES} gap={GAP} />
);
