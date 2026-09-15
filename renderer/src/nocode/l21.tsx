// nocode21 — the verdict (his L21, 4:16)
//
// The payload of the whole day, and the reason `matrix` exists as a template.
// His version puts the takeaway on a fresh slide at 2:30 of a 4:16 lecture and
// asks the viewer to take it on trust. Ours fills the grid in one column per
// build across the day — upto 1 and 2 in L18, 3 in L19, 4 here — so the verdict
// reads as a summary of what they watched.
//
// The verdict is deliberately soft, per ref-L21: the four harnesses are
// structurally the same thing, the models differed so a difference proves
// little, and the choice should be made on which tool your team already uses.
// V2 and V3 put that caveat FIRST rather than burying it.
//
// One thing is newer than his lecture and better: W9's timeline. He describes
// Antigravity as the holdout that never adopted `agents.md`. It has adopted it
// since — verified in the shipped bundle and on camera in L20 — so the
// convergence, not the holdout, is the story.

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';

export const DURS = [
  8.474, 11.363, 13.471, 11.116, 12.596, 8.868, 8.900, 13.386,
  12.162, 8.398, 11.305, 7.963, 21.212, 4.464, 6.131, 18.212,
  15.217, 10.213, 11.933, 18.420, 18.376, 8.561, 5.806, 15.248,
  19.136, 15.280, 8.262, 8.489, 21.923, 13.733, 8.714, 9.230,
];
export const FILES = [
  'W1', 'W2', 'W3', 'W4', 'X1', 'X2', 'X3', 'X4',
  'X5', 'X6', 'X7', 'V1', 'V2', 'V3', 'W5', 'V4',
  'W6', 'W7', 'W8', 'V5', 'W9', 'W10', 'W11', 'W12',
  'V6', 'W13', 'V7', 'W14', 'V8', 'W15', 'W16', 'W17',
].map((n) => `${n}.mp3`);
export const GAP = 0.240;
export const L21_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode21/shots/${n}.mp4`;

const SLIDES: Slide[] = [
  {k: 'shot', src: sh('00'), kicker: 'first, the most useful habit of the week', lines: ['Working in a *loop*'], size: 92},   // W1  I 6.0-32.6s
  {k: 'shot', src: sh('01')},   // W2  I 32.6-68.0s
  {k: 'rows', kicker: 'and look at how that feedback is phrased', lines: ['Three parts. All *three* work'], size: 58, numbered: true, rows: [{t: 'The exact thing that is wrong'}, {t: 'Why it is wrong'}, {t: 'The standard you are holding it to', hot: true}], stamp: 'not "this is bad". Not "make it better"'},   // W3
  {k: 'shot', src: sh('03')},   // W4  I 68.0-114.3s
  {k: 'shot', src: sh('04')},   // X1  I 114.3-166.6s
  {k: 'shot', src: sh('05'), kicker: 'and here is the board after', lines: ['The dead space', 'is *gone*'], size: 62},   // X2  V 0.0-40.0s
  {k: 'shot', src: sh('06'), kicker: 'now look at the left edge', lines: ['It did *one* of', 'the two things'], size: 58},   // X3  V 10.0-30.0s
  {k: 'rows', kicker: 'and this is not me catching it out', lines: ['This is the loop *working*'], size: 62, rows: [{t: 'It made a change'}, {t: 'I went and looked'}, {t: 'Looking found the half that was not finished'}, {t: 'If I had taken its word, that edge would have shipped', hot: true}]},   // X4
  {k: 'myth', kicker: 'and notice why I could catch it at all', lines: ['*Two* named faults'], size: 76, wrong: 'the layout looks a bit off, can you tidy it up', right: 'no left padding on the first column, and the columns do not fill the width', note: 'two named faults means two things you can count off'},   // X5
  {k: 'head', kicker: 'vague feedback gets a vague answer', lines: ['Say *make it better*', 'and you have nothing to check'], size: 54},   // X6
  {k: 'head', kicker: 'so it goes back a second time, with one line', lines: ['Not one perfect instruction.', 'A *short loop*'], size: 54, trans: 'rise'},   // X7
  {k: 'shot', src: sh('11')},   // V1  I 166.6-200.0s
  {k: 'rows', kicker: 'and the caveat comes first, not last', lines: ['They did *not* run', 'the same model'], size: 58, rows: [{t: 'One had a frontier model I chose by hand'}, {t: 'One had whatever a free plan handed it'}, {t: 'Two more were different again'}, {t: 'So a difference could be the harness, or just the model', hot: true}]},   // V2
  {k: 'myth', kicker: 'which is why I keep naming it', lines: ['The question to *ask*'], size: 68, wrong: 'which of these tools is best', right: 'which model was running, and on what date', note: 'any comparison that will not tell you is not a comparison'},   // V3
  {k: 'matrix', kicker: 'four builds, one brief', lines: ['The *grid*, finished'], size: 62, cols: [{head: 'Cursor', sub: 'Opus, chosen'}, {head: 'Copilot', sub: 'free tier, Auto'}, {head: 'Claude Code', sub: 'Opus, high effort'}, {head: 'Antigravity', sub: 'Gemini'}], rows: [{label: 'the brief', cells: ['agents.md', 'agents.md', 'agents.md', 'agents.md']}, {label: 'the prompt', cells: ['go ahead and plan', 'go ahead and plan', 'go ahead and plan', 'go ahead and plan']}, {label: 'asked me', cells: ['never', 'eight times', 'once, first', 'never']}, {label: 'the plan', cells: ['in the panel', 'in the panel', 'in a file', 'in the panel']}, {label: 'five checks', cells: ['all five', 'all five', 'all five', 'all five']}], upto: 4, dateline: '28 august 2026'},   // W5
  {k: 'rows', kicker: 'and here is the honest headline', lines: ['Not that one *won*'], size: 72, rows: [{t: 'A sidebar you can drag wider'}, {t: 'A place to type, and a history above it'}, {t: 'Usually a planning step and an execution step'}, {t: 'And a plain text file in your repo that drives all of it', hot: true}]},   // V4
  {k: 'matrix', kicker: 'read the rows, not the columns', lines: ['Four rows *identical*'], size: 62, cols: [{head: 'Cursor', sub: 'Opus, chosen'}, {head: 'Copilot', sub: 'free tier, Auto'}, {head: 'Claude Code', sub: 'Opus, high effort'}, {head: 'Antigravity', sub: 'Gemini'}], rows: [{label: 'the brief', cells: ['agents.md', 'agents.md', 'agents.md', 'agents.md']}, {label: 'the prompt', cells: ['go ahead and plan', 'go ahead and plan', 'go ahead and plan', 'go ahead and plan']}, {label: 'asked me', cells: ['never', 'eight times', 'once, first', 'never']}, {label: 'the plan', cells: ['in the panel', 'in the panel', 'in a file', 'in the panel']}, {label: 'five checks', cells: ['all five', 'all five', 'all five', 'all five']}], upto: 4, dateline: 'the brief did not change. the prompt did not change'},   // W6
  {k: 'nest', kicker: 'and three of the four are one editor', lines: ['Same *room*, four doors'], size: 62, outer: 'Visual Studio Code', inner: 'Cursor · Copilot · Antigravity', ring: ['same file tree', 'same tabs', 'same shortcuts', 'same panel'], caption: 'two forks and an extension — that is not an accident'},   // W7
  {k: 'shot', src: sh('18')},   // W8  B 0.0-31.4s
  {k: 'rows', kicker: 'so what actually differed', lines: ['*Furniture.* Permissions.', 'Model'], size: 58, rows: [{t: 'Where the panel sits and what the buttons are called'}, {t: 'How often it stops to ask you'}, {t: 'And which model was behind it'}, {t: 'That is the whole difference. It is smaller than the marketing', hot: true}]},   // V5
  {k: 'timeline', kicker: 'and this is the convergence that matters', lines: ['They *agreed* on the file'], size: 62, items: [{when: 'a year ago', t: 'every tool, its own config file', sub: 'its own name, its own place'}, {when: 'this year', t: 'agents.md, near-universally', sub: 'read without being told'}, {when: 'today', t: 'the last holdout adopted it', sub: 'since the course this is built from', hot: true}], caption: 'faster than anybody expected'},   // W9
  {k: 'head', kicker: 'which means the work you did is portable', lines: ['Change tools next month.', 'The *file* comes with you'], size: 54},   // W10
  {k: 'head', kicker: 'so, choosing', lines: ['The *boring* answer', 'is the right one'], size: 68},   // W11
  {k: 'rows', kicker: 'and every one of these beats a benchmark', lines: ['Pick on *this*'], size: 76, numbered: true, rows: [{t: 'The one your team already uses — so you can ask someone'}, {t: 'The one your employer already pays for'}, {t: 'The one you can log into right now'}, {t: 'The keyboard you already know', hot: true}]},   // W12
  {k: 'shot', src: sh('24')},   // V6  B 31.4-81.4s
  {k: 'rows', kicker: 'and do not agonise, because it transfers', lines: ['One *afternoon* to move'], size: 68, rows: [{t: 'Writing a brief'}, {t: 'Reading a plan and its success checks'}, {t: 'Watching a permission prompt'}, {t: 'Checking a claim instead of believing it', hot: true}]},   // W13
  {k: 'head', kicker: 'which is the real reason we built it four times', lines: ['The tool is the *least*', 'interesting variable'], size: 58, trans: 'rise'},   // V7
  {k: 'head', kicker: 'so the course turns here', lines: ['Products, done.', 'Now *technique*'], size: 80, trans: 'push'},   // W14
  {k: 'rows', kicker: 'and none of these expire', lines: ['What the rest of the', 'three weeks *is*'], size: 58, rows: [{t: 'A brief that does not need a follow-up'}, {t: 'Making an agent prove a fix instead of announcing one'}, {t: 'Working in a loop, not one long hopeful prompt'}, {t: 'Knowing when to stop', hot: true}]},   // V8
  {k: 'shot', src: sh('29')},   // W15  B 81.4-117.5s
  {k: 'head', kicker: 'and one will launch before you finish this course', lines: ['That is not a joke.', 'It is a *scheduling* observation'], size: 52},   // W16
  {k: 'grid', kicker: 'tomorrow — the first project, and it is the brave one', lines: ['Day *four*'], done: 3, now: 3, trans: 'push'},   // W17
];

export const Nocode21: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode21/vo" files={FILES} gap={GAP} />
);
