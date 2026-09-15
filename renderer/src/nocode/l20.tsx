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
  10.133, 15.314, 13.510, 7.121, 9.289, 9.619, 6.044, 17.163,
  14.513, 12.638, 5.014, 16.481, 12.254, 11.761, 5.903, 16.071,
  9.423, 12.334, 21.363, 16.679, 12.812, 10.385, 22.268, 5.040,
  3.431, 4.800, 9.713, 17.614, 12.384, 14.234, 14.779, 13.601,
  11.636, 15.267, 13.016, 13.956, 12.510, 11.745, 13.697, 11.388,
  20.511, 17.188, 8.730, 4.856, 14.785, 12.069, 11.693, 15.918,
  11.580, 12.151, 16.926, 11.457, 9.264, 19.021, 9.737, 9.799,
  10.631, 6.375, 11.040, 7.189,
];
export const FILES = [
  'P1', 'P2', 'P3', 'P4', 'Q1', 'Q2', 'Q3', 'Q4',
  'Q5', 'Q6', 'Q7', 'Q8', 'Q9', 'Q10', 'Q11', 'Q12',
  'Q13', 'Q14', 'Q15', 'Q16', 'G2', 'P5', 'P6', 'P7',
  'G9', 'P8', 'P9', 'P10', 'P11', 'P12', 'P13', 'G8',
  'P14', 'P15', 'P16', 'R1', 'R2', 'R3', 'R4', 'R5',
  'R6', 'R7', 'R8', 'R9', 'R10', 'R11', 'R12', 'R13',
  'R14', 'R15', 'R16', 'K1', 'K2', 'K3', 'K4', 'K5',
  'K6', 'K7', 'K8', 'P17',
].map((n) => `${n}.mp3`);
export const GAP = 0.240;
export const L20_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode20/shots/${n}.mp4`;

const SLIDES: Slide[] = [
  {k: 'shot', src: sh('00'), kicker: 'build four of four', lines: ['The *holdout*'], size: 112},   // P1  B 6.0-12.1s
  {k: 'rows', kicker: 'and the caveat comes first, not last', lines: ['They did *not* run', 'the same model'], size: 58, rows: [{t: 'Cursor — a frontier model I chose by hand'}, {t: 'Copilot — whatever the free plan handed it'}, {t: 'Claude Code — the same model as Cursor'}, {t: 'This one — Google\'s default, which is fast, not strongest', hot: true}]},   // P2
  {k: 'head', kicker: 'the least surprising sentence in the course', lines: ['Strong model,', '*strong* results'], size: 76},   // P3
  {k: 'myth', kicker: 'so here is the question to ask anyone', lines: ['Ask *which model*'], size: 72, wrong: 'which of these tools is the best one', right: 'which model was running, and on what date', note: 'a comparison that will not tell you is not a comparison'},   // P4
  {k: 'head', kicker: 'and it has the best address in the industry', lines: ['antigravity', '*.google*'], size: 96, trans: 'rise'},   // Q1
  {k: 'rows', kicker: 'installing it', lines: ['Take *all* the defaults'], size: 80, rows: [{t: 'Mac — drag it into Applications'}, {t: 'Windows — click next a few times'}, {t: 'There is nothing to configure and nothing to choose', hot: true}]},   // Q2
  {k: 'shot', src: sh('06')},   // Q3  B 12.1-15.9s
  {k: 'nest', kicker: 'and this is the third time today', lines: ['It is *VS Code*. Again'], size: 68, outer: 'Visual Studio Code', inner: 'Antigravity', ring: ['same file tree', 'same tabs', 'same shortcuts', 'same panel'], caption: 'a different company, a different model, the same room'},   // Q4
  {k: 'rows', kicker: 'so count what that is worth', lines: ['*Three* of four', 'are one editor'], size: 64, rows: [{t: 'Cursor — a fork of VS Code'}, {t: 'Antigravity — a fork of VS Code'}, {t: 'Copilot — an extension inside VS Code'}, {t: 'Learn the room once and you have learned three of them', hot: true}]},   // Q5
  {k: 'head', kicker: 'and I am not filming mine', lines: ['A sign-in screen is', 'somebody\'s *name and email*'], size: 54},   // Q6
  {k: 'shot', src: sh('10')},   // Q7  B 15.9-19.0s
  {k: 'shot', src: sh('11')},   // Q8  B 46.0-54.7s
  {k: 'rows', kicker: 'in google\'s own editor, offered as equals', lines: ['Its *competitors* models'], size: 64, rows: [{t: 'Gemini 3.7 Flash — where it starts you'}, {t: 'Claude Sonnet 4.6 and Claude Opus 4.6 — Anthropic'}, {t: 'GPT-OSS 120B — open weights, run it yourself'}], stamp: '28 august 2026 — check it yourself, it will have moved'},   // Q9
  {k: 'myth', kicker: 'which is the most useful thing to understand', lines: ['They have *come apart*'], size: 68, wrong: 'you choose a tool, and the tool is the thing', right: 'the harness is what you choose. the model is what does the work', note: 'and the harness vendors have accepted that'},   // Q10
  {k: 'shot', src: sh('14')},   // Q11  B 54.7-57.9s
  {k: 'rows', kicker: 'and nobody explains this word either', lines: ['A *lint* is not a bug'], size: 72, rows: [{t: 'Something imported and never used'}, {t: 'A name written two different ways'}, {t: 'Nothing that stops the app running'}, {t: 'All of it mess that makes the next change harder', hot: true}]},   // Q12
  {k: 'shot', src: sh('16'), kicker: 'so turn it on', lines: ['It costs *nothing*', 'and happens without you'], size: 54},   // Q13  B 57.9-63.0s
  {k: 'ladder', kicker: 'the permission dial, for the fourth time', lines: ['This one has *two* positions'], size: 62, items: ['Request review — ask me before you act', 'Always proceed — do not ask me at all'], pick: 1, tone: '#e53935'},   // Q14
  {k: 'myth', kicker: 'and one option has gone missing', lines: ['I do not *know* why'], size: 72, wrong: 'the course I built this from had a third setting: the model decides', right: 'this build has two. I am telling you rather than guessing', note: 'a course that pretends to know everything is worse than one that says where it stops'},   // Q15
  {k: 'shot', src: sh('19'), kicker: 'you should hear me say this every time', lines: ['A throwaway project,', 'an *empty folder*'], size: 54},   // Q16  A 0.0-233.8s
  {k: 'hierarchy', kicker: 'what the other three did without being asked', lines: ['*One file*, in the root'], size: 64, nodes: [{t: 'kanban', depth: 0, kind: 'dir'}, {t: 'agents.md', depth: 1, kind: 'agents'}, {t: 'README.md', depth: 1, kind: 'file'}], active: 1, caption: 'nobody told any of them it was there'},   // G2
  {k: 'head', kicker: 'and now the section you are not getting', lines: ['Let me tell you', 'what I *planned*'], size: 64, trans: 'push'},   // P5
  {k: 'hierarchy', kicker: 'nine steps of fiddly file management', lines: ['The conversion', 'his lecture *teaches*'], size: 58, nodes: [{t: 'kanban', depth: 0, kind: 'dir'}, {t: '.agent', depth: 1, kind: 'dir'}, {t: 'rules', depth: 2, kind: 'dir'}, {t: 'strategy.md', depth: 3, kind: 'agents'}, {t: 'agents.md', depth: 1, kind: 'file'}], active: 3, caption: 'select, copy, three folders, paste, set a header, save, then delete the original'},   // P6
  {k: 'head', kicker: 'to get the same words', lines: ['To the *same model*'], size: 104},   // P7
  {k: 'shot', src: sh('24')},   // G9  B 19.0-21.1s
  {k: 'shot', src: sh('25')},   // P8  B 21.1-24.1s
  {k: 'shot', src: sh('26'), kicker: 'it went and found the file', lines: ['*AGENTS.md*'], size: 120},   // P9  B 24.1-30.0s
  {k: 'rows', kicker: 'and I checked twice, because courses get this wrong', lines: ['Not a *guess*'], size: 80, rows: [{t: 'The app ships a settings page for this file, by name'}, {t: 'It writes a line into its own log when it picks one up'}, {t: 'It reads it deliberately', hot: true}]},   // P10
  {k: 'timeline', kicker: 'four companies. two of them direct competitors', lines: ['*One* convention'], size: 80, items: [{when: 'a year ago', t: 'every tool, its own config file', sub: 'its own name, its own folder'}, {when: 'this year', t: 'agents.md, near-universally', sub: 'read without being told'}, {when: 'today', t: 'the last holdout reads it too', sub: 'since the course this is built from', hot: true}], caption: 'about twelve months, and quietly'},   // P11
  {k: 'head', kicker: 'so the file is not an investment in one product', lines: ['It is what *survives*', 'you changing your mind'], size: 54, trans: 'rise'},   // P12
  {k: 'hierarchy', kicker: 'its own folder still exists, and still has a use', lines: ['For rules that live *only here*'], size: 58, nodes: [{t: 'kanban', depth: 0, kind: 'dir'}, {t: 'agents.md', depth: 1, kind: 'agents'}, {t: '.agent', depth: 1, kind: 'dir'}, {t: 'rules', depth: 2, kind: 'dir'}], active: 1, caption: 'agents.md is the brief all four read — .agent/rules is for anything only this tool needs'},   // P13
  {k: 'head', kicker: 'which will outlive this particular argument', lines: ['The convention is not the point.', 'The *brief* is the point'], size: 52},   // G8
  {k: 'shot', src: sh('32')},   // P14  B 30.0-36.9s
  {k: 'shot', src: sh('33')},   // P15  B 36.9-46.0s
  {k: 'rows', kicker: 'same nine words. four readings', lines: ['What *plan first* meant'], size: 64, numbered: true, rows: [{t: 'Cursor — plan, then carry straight on'}, {t: 'Copilot — plan, then carry straight on'}, {t: 'Claude Code — plan, then stop and check'}, {t: 'Antigravity — plan to a file, then stop and check', hot: true}]},   // P16
  {k: 'shot', src: sh('35')},   // R1  A 233.8-430.0s
  {k: 'shot', src: sh('36')},   // R2  A 600.0-623.5s
  {k: 'myth', kicker: 'two settings, because two different questions', lines: ['*Ask*, and *show*'], size: 76, wrong: 'one switch called permissions that means everything', right: 'one dial for stopping before it acts, another for showing you after', note: 'and I like that they are separate'},   // R3
  {k: 'shot', src: sh('38')},   // R4  A 623.5-649.1s
  {k: 'shot', src: sh('39')},   // R5  A 649.1-670.5s
  {k: 'myth', kicker: 'and this is where his build fell down', lines: ['A prompt, or a *form*'], size: 72, wrong: 'a bare browser pop-up asking for a name, nowhere for a description', right: 'a dialog with a required title and an optional description', note: 'the brief said a card has both — it read that as a requirement'},   // R6
  {k: 'rows', kicker: 'and nothing asked for any of this', lines: ['It inferred the *taste*'], size: 68, rows: [{t: 'The five hex codes out of our brief, by name'}, {t: 'No emoji anywhere — it drew the icons instead'}, {t: 'The brief mentioned a palette and never mentioned emoji', hot: true}]},   // R7
  {k: 'rows', kicker: 'and the gates it set itself', lines: ['*16 of 16*'], size: 104, rows: [{t: 'Sixteen tests out of sixteen passing'}, {t: 'No lint warnings at all'}, {t: 'Production build compiled cleanly'}, {t: 'A server running, with the address', hot: true}]},   // R8
  {k: 'shot', src: sh('43'), kicker: 'and now the most useful part', lines: ['The *honest* part'], size: 96},   // R9  A 670.5-679.9s
  {k: 'rows', kicker: 'this is the tool\'s headline feature', lines: ['It drives a *real browser*', 'and records itself'], size: 54, rows: [{t: 'The agent tests the app by actually using it'}, {t: 'And hands you a video of it working'}, {t: 'In the course I built this from it is the striking moment of the day'}]},   // R10
  {k: 'shot', src: sh('45')},   // R11  A 679.9-702.5s
  {k: 'shot', src: sh('46')},   // R12  A 702.5-724.5s
  {k: 'rows', kicker: 'look at the shape of what it did', lines: ['Four things, *in order*'], size: 68, numbered: true, rows: [{t: 'Named the failure'}, {t: 'Named the cause — a 404 fetching the driver'}, {t: 'Named the substitute it used instead'}, {t: 'Named what the substitute does not cover', hot: true}], stamp: 'it did not report success and let me find out'},   // R13
  {k: 'shot', src: sh('48'), kicker: 'you will see far more of this', lines: ['How a thing *fails*', 'is worth more'], size: 58},   // R14  A 724.5-746.2s
  {k: 'shot', src: sh('49'), kicker: 'so I will not describe it as though I had', lines: ['I cannot *show* you', 'the feature'], size: 58},   // R15  A 746.2-769.0s
  {k: 'shot', src: sh('50')},   // R16  S 0.0-19.4s
  {k: 'shot', src: sh('51')},   // K1  S 19.4-32.6s
  {k: 'rows', kicker: 'fourth time today', lines: ['The *same five*, in order'], size: 62, numbered: true, rows: [{t: 'Drag a card between columns'}, {t: 'Reorder a card inside a column'}, {t: 'Delete a card'}, {t: 'Rename a column'}, {t: 'Add a card, with a description'}]},   // K2
  {k: 'shot', src: sh('53')},   // K3  S 32.6-54.3s
  {k: 'shot', src: sh('54')},   // K4  S 54.3-65.6s
  {k: 'shot', src: sh('55')},   // K5  S 65.6-76.9s
  {k: 'shot', src: sh('56')},   // K6  S 76.9-89.2s
  {k: 'shot', src: sh('57'), kicker: 'demonstrated four times in one afternoon', lines: ['Rather than *asserted* once'], size: 62},   // K7  S 89.2-96.7s
  {k: 'shot', src: sh('58')},   // K8  S 96.7-109.4s
  {k: 'shot', src: sh('59'), kicker: 'and it is short', lines: ['Next: the *verdict*'], size: 96},   // P17  S 109.4-117.5s
];

export const Nocode20: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode20/vo" files={FILES} gap={GAP} />
);
