// nocode18 — the Copilot build (his L18, 14:44)
//
// Build two of four from the same brief. 57 of 97 slides are real screen
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
import {Deck, Slide, deckFrames} from './kit';

export const DURS = [
  11.353, 19.006, 11.065, 9.049, 8.897, 13.575, 12.337, 15.185,
  14.276, 15.613, 13.632, 12.262, 15.351, 8.432, 10.560, 13.262,
  14.961, 12.617, 13.157, 13.334, 29.995, 16.267, 15.666, 7.847,
  10.672, 14.588, 9.698, 6.801, 16.291, 11.658, 10.441, 13.195,
  11.457, 6.566, 5.339, 10.912, 14.248, 18.984, 6.857, 14.389,
  17.297, 14.400, 9.816, 13.737, 17.119, 11.440, 12.714, 11.175,
  10.123, 18.207, 7.709, 15.897, 10.617, 14.330, 11.052, 16.514,
  11.157, 14.317, 8.463, 15.136, 3.757, 14.639, 13.759, 14.723,
  11.497, 7.335, 8.195, 11.379, 5.487, 17.179, 8.129, 12.469,
  17.043, 12.545, 11.085, 17.410, 5.692, 16.018, 8.492, 16.002,
  7.080, 23.344, 5.198, 16.913, 8.575, 14.161, 16.855, 18.318,
  10.827, 13.901, 16.076, 5.447, 18.088, 8.342, 12.634, 11.507,
  13.415,
];
export const FILES = [
  '01', 'M1', 'M2', 'M3', '02', 'M4', 'M5', 'M6',
  'M7', 'M8', 'M9', 'M10', '03', '04', '05', 'M11',
  'M12', '06', 'M13', 'M14', 'T1', 'M15', 'M16', '07',
  '08', 'M18', 'M19', 'M20', 'M21', 'M22', 'M23', 'M24',
  'M25', 'R2', 'X1', 'X2', 'X3', 'R3', 'M26', 'M27',
  'M28', 'M29', 'M30', 'R4', 'R5', 'M31', 'M32', 'M33',
  'M34', 'R6', 'M35', 'M36', 'M37', 'M38', 'M39', 'M40',
  'M41', 'M42', 'M43', 'M44', 'R8', 'R9', 'M46', 'M47',
  'M45', 'R10', 'R11', 'R12', 'R13', 'R14', 'R15', 'Z1',
  'Z2', 'Z3', 'M48', 'R16', 'T2', 'T3', 'T4', 'T5',
  'T6', 'T7', 'T8', 'R17', 'H1', 'H2', 'H3', 'H4',
  'H5', 'H6', 'H7', 'Z4', 'Z5', 'Z6', 'Z7', 'Z8',
  'Z9',
].map((n) => `${n}.mp3`);
export const GAP = 0.240;
export const L18_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode18/shots/${n}.mp4`;

const SLIDES: Slide[] = [
  {k: 'shot', src: sh('00'), kicker: 'build two of four', lines: ['The one you', '*already have*'], size: 90},   // 01  P 0.0-11.5s
  {k: 'term', kicker: 'first, the ritual', lines: ['Rename it. *Clone it again*'], size: 62, title: 'cursor kanban, kept as a record', term: [{t: 'mv kanban kanban-cursor', kind: 'cmd'}, {t: 'git clone .../kanban.git', kind: 'cmd'}, {t: 'Cloning into kanban...', kind: 'out'}, {t: 'agents.md   README.md', kind: 'ok'}]},   // M1
  {k: 'shot', src: sh('02')},   // M2  P 11.5-22.7s
  {k: 'hierarchy', kicker: 'four builds, four records', lines: ['One brief. *Four folders*'], size: 62, nodes: [{t: 'projects', depth: 0, kind: 'dir'}, {t: 'kanban-cursor', depth: 1, kind: 'dir'}, {t: 'kanban-copilot', depth: 1, kind: 'dir'}, {t: 'kanban-claude', depth: 1, kind: 'dir'}, {t: 'kanban-antigravity', depth: 1, kind: 'dir'}, {t: 'kanban', depth: 1, kind: 'dir'}, {t: 'agents.md', depth: 2, kind: 'agents'}], active: 6, caption: 'the fifth one is always the fresh clone'},   // M3
  {k: 'nest', kicker: 'where it actually lives', lines: ['Not an app. *An extension*'], size: 62, outer: 'Visual Studio Code', inner: 'Copilot', caption: 'there is no Copilot in your dock'},   // 02
  {k: 'shot', src: sh('05')},   // M4  P 22.7-36.4s
  {k: 'shot', src: sh('06')},   // M5  P 70.0-80.2s
  {k: 'nest', kicker: 'and here is the good news', lines: ['Cursor *is* VS Code'], size: 68, outer: 'Visual Studio Code', inner: 'Cursor', ring: ['same file tree', 'same tabs', 'same shortcuts', 'same panel'], caption: 'built from its source, not merely inspired by it'},   // M6
  {k: 'shot', src: sh('08')},   // M7  P 80.2-91.9s
  {k: 'rows', kicker: 'and it happens again today', lines: ['*Three* of four', 'are the same editor'], size: 64, rows: [{t: 'Cursor — a fork of VS Code'}, {t: 'Copilot — an extension inside VS Code'}, {t: 'Antigravity — a fork of VS Code'}, {t: 'Learn the room once, and you know three of them', hot: true}]},   // M8
  {k: 'gauge', kicker: 'and the money', lines: ['A *real* free tier'], size: 68, pct: 4, limit: 100, label: 'this build: under 4 credits', limitLabel: 'monthly allowance', caption: 'not a seven day trial'},   // M9
  {k: 'shot', src: sh('11')},   // M10  P 91.9-102.1s
  {k: 'matrix', kicker: 'hold this in your head all day', lines: ['Only *one thing* changes'], size: 58, cols: [{head: 'Cursor', sub: 'frontier, chosen'}, {head: 'Copilot', sub: 'free tier, Auto'}, {head: 'Claude Code', sub: 'next'}, {head: 'Antigravity', sub: 'after that'}], rows: [{label: 'the brief', cells: ['agents.md', 'agents.md', '', '']}, {label: 'the prompt', cells: ['go ahead and plan', 'go ahead and plan', '', '']}, {label: 'asked me', cells: ['never', 'eight times', '', '']}, {label: 'five checks', cells: ['all five', 'all five', '', '']}], upto: 1, dateline: '28 august 2026'},   // 03
  {k: 'hierarchy', kicker: 'so, a fresh clone', lines: ['*One file.* Nothing else'], size: 68, nodes: [{t: 'kanban', depth: 0, kind: 'dir'}, {t: 'agents.md', depth: 1, kind: 'agents'}, {t: 'README.md', depth: 1, kind: 'file'}], active: 1},   // 04
  {k: 'shot', src: sh('14')},   // 05  P 102.1-110.8s
  {k: 'shot', src: sh('15')},   // M11  P 110.8-121.8s
  {k: 'ladder', kicker: 'the dropdown that matters', lines: ['Chatbot, or *agent*'], size: 68, items: ['Ask — it talks about your code', 'Edit — it changes the file you point at', 'Agent — it reads, writes and runs things'], pick: 2, tone: '#f5c518'},   // M12
  {k: 'myth', kicker: 'one difference, named early', lines: ['Where the *planning* lives'], size: 62, wrong: 'a Plan mode you switch into, as a separate thing', right: 'planning folded into the agent, driven by the brief', note: 'same output, different furniture'},   // 06
  {k: 'shot', src: sh('18')},   // M13  P 121.8-132.6s
  {k: 'shot', src: sh('19')},   // M14  P 132.6-143.6s
  {k: 'shot', src: sh('20')},   // T1  R 232.0-258.9s
  {k: 'shot', src: sh('21')},   // M15  R 258.9-273.6s
  {k: 'meter', kicker: 'read the grey line', lines: ['*3.6 credits*, and it told me'], size: 68, pct: 4, fill: 'spent on this whole build', rest: 'left this month', caption: 'the only place it is completely straight with you'},   // M16
  {k: 'shot', src: sh('23')},   // 07  P 36.4-44.5s
  {k: 'shot', src: sh('24')},   // 08  P 44.5-55.3s
  {k: 'shot', src: sh('25')},   // M18  P 55.3-70.0s
  {k: 'shot', src: sh('26')},   // M19  P 143.6-151.7s
  {k: 'shot', src: sh('27')},   // M20  P 151.7-157.4s
  {k: 'shot', src: sh('28')},   // M21  P 157.4-170.8s
  {k: 'myth', kicker: 'read the last line of each phase', lines: ['A list, or a *commitment*'], size: 62, wrong: 'Add unit tests', right: 'Success: core state behaviour is covered and tests pass', note: 'one you can hold it to later'},   // M22
  {k: 'editor', kicker: 'and it wrote them because we asked', lines: ['*Nine words*, two lectures ago'], size: 58, file: 'agents.md', start: 44, rows: [{t: '## Strategy', kind: 'h2'}, {t: '', kind: 'text'}, {t: '1. Write a plan first, broken into phases, with', kind: 'important'}, {t: '   **success criteria for each phase** that can', kind: 'important'}, {t: '   be checked off.', kind: 'important'}], focus: [2, 4]},   // M23
  {k: 'shot', src: sh('31')},   // M24  P 170.8-181.6s
  {k: 'shot', src: sh('32')},   // M25  P 192.0-214.5s
  {k: 'shot', src: sh('33')},   // R2  P 214.5-227.7s
  {k: 'shot', src: sh('34')},   // X1  P 227.7-238.4s
  {k: 'shot', src: sh('35')},   // X2  P 238.4-259.9s
  {k: 'ladder', kicker: 'the trade you are making', lines: ['Interruption *is* visibility'], size: 64, items: ['ask once per session — fast, and you see nothing', 'ask once per command — slow, and you see everything', 'never ask — fastest, and you find out afterwards'], pick: 1, tone: '#3b82f6'},   // X3
  {k: 'shot', src: sh('37')},   // R3  P 259.9-296.9s
  {k: 'shot', src: sh('38')},   // M26  P 296.9-310.6s
  {k: 'stack', kicker: 'what that command really does', lines: ['You are running *strangers* code'], size: 58, layers: [{t: 'your app', sub: 'the part you asked for', h: 46, tone: 'yellow'}, {t: 'the framework', sub: 'thousands of files you did not write', h: 66}, {t: 'its dependencies', sub: 'and their dependencies', h: 86}, {t: 'published by people you cannot name', sub: 'downloaded on demand', h: 60, tone: 'red'}], foot: 'normal, universal, and the largest hole in the boat'},   // M27
  {k: 'rows', kicker: 'and it has actually happened', lines: ['*Supply chain*, plainly'], size: 68, rows: [{t: 'One popular package gets hostile code added to it'}, {t: 'Everybody who installs it runs that code'}, {t: 'Not theoretical — it has hit packages with millions of users', hot: true}]},   // M28
  {k: 'myth', kicker: 'so where should a tool stop you', lines: ['*This* is the right line'], size: 64, wrong: 'stop me before every file edit, which I will switch off by lunchtime', right: 'stop me on the one command that downloads and runs strangers code', note: 'and do not let a blanket yes override that one'},   // M29
  {k: 'shot', src: sh('42')},   // M30  P 310.6-330.0s
  {k: 'shot', src: sh('43')},   // R4  R 0.0-14.1s
  {k: 'shot', src: sh('44')},   // R5  R 14.1-31.5s
  {k: 'loops', kicker: 'the honest rhythm of a build', lines: ['Work. *Ask.* Yes. Work'], size: 68, inner: ['work', 'ask', 'yes'], passes: 8, label: 'eight times', caption: 'nobody puts this bit in a demo'},   // M31
  {k: 'shot', src: sh('46')},   // M32  R 31.5-44.5s
  {k: 'shot', src: sh('47')},   // M33  R 44.5-56.0s
  {k: 'head', kicker: 'and if it wanders', lines: ['*That* is your moment.', 'Not later'], size: 76, accent: '#e53935'},   // M34
  {k: 'shot', src: sh('49')},   // R6  R 56.0-74.6s
  {k: 'shot', src: sh('50')},   // M35  R 74.6-82.6s
  {k: 'rows', kicker: 'because nobody explains this word', lines: ['What a *test* is'], size: 76, rows: [{t: 'A small piece of code that checks another piece of code'}, {t: 'Move a card from here to there. Now ask: is it there'}, {t: 'Write it once, run it a thousand times in two seconds', hot: true}]},   // M36
  {k: 'shot', src: sh('52')},   // M37  R 82.6-92.0s
  {k: 'shot', src: sh('53')},   // M38  R 159.5-174.1s
  {k: 'shot', src: sh('54')},   // M39  R 174.1-185.5s
  {k: 'shot', src: sh('55')},   // M40  R 185.5-202.3s
  {k: 'loops', kicker: 'and this is the new part', lines: ['A loop with *nobody in it*'], size: 68, inner: ['write the app', 'write the checks', 'run them', 'fix what failed'], passes: 3, label: 'no human', caption: 'a year ago this ended at "here is some code, try it"'},   // M41
  {k: 'shot', src: sh('57')},   // M42  R 273.6-286.5s
  {k: 'shot', src: sh('58')},   // M43  R 286.5-294.2s
  {k: 'myth', kicker: 'two very different claims', lines: ['Consistent, or *correct*'], size: 64, wrong: 'it wrote the exam, sat it, marked it, and passed', right: 'someone who is going to use it touched it and it worked', note: 'only one of those is worth anything to you'},   // M44
  {k: 'shot', src: sh('60')},   // R8  R 294.2-297.8s
  {k: 'shot', src: sh('61')},   // R9  R 297.8-311.0s
  {k: 'shot', src: sh('62')},   // M46  S 0.0-12.0s
  {k: 'editor', kicker: 'and nothing asked for a name', lines: ['It took *this* literally'], size: 58, file: 'agents.md', start: 19, rows: [{t: 'The priority is a **slick, professional, genuinely', kind: 'important'}, {t: 'good-looking interface** over a small set of', kind: 'important'}, {t: 'features. A viewer should want to use it.', kind: 'important'}], focus: [0, 2], caption: 'it read that as an instruction, not a sentiment'},   // M47
  {k: 'rows', kicker: 'so we do it ourselves', lines: ['The *same five*, in order'], size: 62, numbered: true, rows: [{t: 'Drag a card between columns'}, {t: 'Reorder a card inside a column'}, {t: 'Delete a card'}, {t: 'Rename a column'}, {t: 'Add a card, with a description'}]},   // M45
  {k: 'shot', src: sh('65')},   // R10  S 12.0-28.0s
  {k: 'shot', src: sh('66')},   // R11  S 28.0-42.0s
  {k: 'shot', src: sh('67')},   // R12  S 42.0-58.0s
  {k: 'shot', src: sh('68')},   // R13  S 58.0-72.0s
  {k: 'shot', src: sh('69')},   // R14  S 72.0-96.0s
  {k: 'shot', src: sh('70')},   // R15  S 96.0-117.5s
  {k: 'myth', kicker: 'one line in the brief, visible in the result', lines: ['Obvious, over *clever*'], size: 64, wrong: 'install a library that specialises in dragging', right: 'use the dragging the browser has had for years', note: 'fewer moving parts, easier for the next person'},   // Z1
  {k: 'shot', src: sh('72')},   // Z2  S 12.0-30.0s
  {k: 'shot', src: sh('73'), kicker: 'the whole lesson of the week', lines: ['The brief *is*', 'the product'], size: 84},   // Z3  P 181.6-192.0s
  {k: 'meter', kicker: 'what it cost, in both currencies', lines: ['*Eleven minutes.* Four credits'], size: 64, pct: 4, fill: 'of the monthly allowance', rest: 'still there', caption: 'most of the eleven minutes was me clicking yes'},   // M48
  {k: 'shot', src: sh('75'), kicker: 'and now an honest detour', lines: ['The lecture I built', 'this from *broke*'], size: 64},   // R16  R 311.0-326.7s
  {k: 'shot', src: sh('76'), kicker: 'which makes this the real subject', lines: ['*Debugging*'], size: 112},   // T2  R 202.3-208.3s
  {k: 'flow', kicker: 'the move it makes almost every time', lines: ['Guess. Patch. *Declare victory*'], size: 58, nodes: [{t: 'guess the cause', sub: 'without looking', tone: 'red'}, {t: 'write a fix', sub: 'for the guess'}, {t: '"fixed!"', sub: 'never once run', tone: 'red'}], caption: 'once you have seen it you will never unsee it'},   // T3
  {k: 'myth', kicker: 'two separate faults in that', lines: ['Proving, not *guessing*'], size: 64, wrong: 'it guessed, and it claimed', right: 'it should have proved, and it should have shown you', note: 'those are two different failures, not one'},   // T4
  {k: 'rows', kicker: 'write this one down', lines: ['The *four steps*'], size: 88, numbered: true, rows: [{t: 'Reproduce the problem'}, {t: 'Prove you have reproduced it'}, {t: 'Find the root cause'}, {t: 'Fix it — and demonstrate the fix', hot: true}], stamp: 'give it these four, in this order'},   // T5
  {k: 'rows', kicker: 'and notice the balance', lines: ['*Three* of four', 'are evidence'], size: 68, rows: [{t: 'Reproduce — evidence'}, {t: 'Prove — evidence'}, {t: 'Fix — the only line about code'}, {t: 'Demonstrate — evidence', hot: true}]},   // T6
  {k: 'shot', src: sh('81'), kicker: 'you can run this without reading code', lines: ['You do not need to read the fix', 'to ask whether it was *shown*'], size: 50},   // T7  R 208.3-232.0s
  {k: 'shot', src: sh('82'), kicker: 'remember what it always looks like', lines: ['Confident and detailed', 'is not *evidence*'], size: 58},   // T8  R 326.7-331.6s
  {k: 'rows', kicker: 'so, ours behaved. Keep the instruction anyway', lines: ['For the day it *does not*'], size: 64, rows: [{t: 'I am not going to stage a fault to make the lesson land'}, {t: 'Do not read our clean result as the normal one'}, {t: 'Have the four steps in your hand before you need them', hot: true}]},   // R17
  {k: 'shot', src: sh('84'), kicker: 'the part these comparisons leave out', lines: ['What this *did not*', 'show you'], size: 70},   // H1  S 96.0-112.0s
  {k: 'shot', src: sh('85')},   // H2  S 96.0-112.0s
  {k: 'rows', kicker: 'and this is the one that bites', lines: ['Requirements that *fight*'], size: 68, rows: [{t: 'Every line of our brief agrees with every other line'}, {t: 'A real request contains two things that cannot both be true'}, {t: 'And nobody notices until something has been built', hot: true}]},   // H3
  {k: 'stack', kicker: 'and it did not show you a second day', lines: ['*Persistence* is a different job'], size: 58, layers: [{t: 'no persistence', sub: 'what we built. Close the tab, it resets', h: 44, tone: 'yellow'}, {t: 'a database', sub: 'the moment you want it saved', h: 62}, {t: 'accounts and logins', sub: 'because whose board is it', h: 62}, {t: 'two people editing one card', sub: 'and now you have a real system', h: 72, tone: 'red'}], foot: 'not a bigger version of today — a different job'},   // H4
  {k: 'head', kicker: 'and the thing that actually breaks these', lines: ['Someone *changes their mind*', 'and nobody wrote it down'], size: 54},   // H5
  {k: 'editor', kicker: 'which is what the file is for', lines: ['Not a prompt. *A record*'], size: 62, file: 'agents.md', start: 1, rows: [{t: '# Kanban — project brief', kind: 'h1'}, {t: '', kind: 'text'}, {t: 'the written-down version of what was agreed', kind: 'important'}, {t: '', kind: 'text'}, {t: 'the only thing from today still useful in a month', kind: 'important'}], focus: [2, 4]},   // H6
  {k: 'gauge', kicker: 'one practical warning', lines: ['You will *feel* the limit', 'before you read it'], size: 54, pct: 97, limit: 100, label: 'requests used this month', limitLabel: 'allowance', caption: 'if a tool suddenly gets worse, check your usage before you blame it'},   // H7
  {k: 'head', kicker: 'two down', lines: ['Side by *side*'], size: 110, trans: 'fade'},   // Z4
  {k: 'matrix', kicker: 'same brief, same prompt, same checks', lines: ['What actually *differed*'], size: 58, cols: [{head: 'Cursor', sub: 'frontier, chosen'}, {head: 'Copilot', sub: 'free tier, Auto'}, {head: 'Claude Code', sub: 'next'}, {head: 'Antigravity', sub: 'after that'}], rows: [{label: 'the brief', cells: ['agents.md', 'agents.md', '', '']}, {label: 'the prompt', cells: ['go ahead and plan', 'go ahead and plan', '', '']}, {label: 'asked me', cells: ['never', 'eight times', '', '']}, {label: 'five checks', cells: ['all five', 'all five', '', '']}], upto: 2, dateline: '28 august 2026'},   // Z5
  {k: 'matrix', kicker: 'and the result', lines: ['*Both* cleared all five'], size: 62, cols: [{head: 'Cursor', sub: 'frontier, chosen'}, {head: 'Copilot', sub: 'free tier, Auto'}, {head: 'Claude Code', sub: 'next'}, {head: 'Antigravity', sub: 'after that'}], rows: [{label: 'the brief', cells: ['agents.md', 'agents.md', '', '']}, {label: 'the prompt', cells: ['go ahead and plan', 'go ahead and plan', '', '']}, {label: 'asked me', cells: ['never', 'eight times', '', '']}, {label: 'five checks', cells: ['all five', 'all five', '', '']}], upto: 2, dateline: '28 august 2026'},   // Z6
  {k: 'shot', src: sh('94'), kicker: 'which was not what I expected', lines: ['On a job this size,', 'it *barely matters*'], size: 58},   // Z7  S 100.0-117.5s
  {k: 'shot', src: sh('95'), kicker: 'and the next one changes shape', lines: ['A command line tool', 'wearing an *editor*'], size: 54},   // Z8  R 331.6-342.0s
  {k: 'rows', kicker: 'before you go', lines: ['*Reproduce. Prove.*', '*Fix. Demonstrate*'], size: 62, rows: [{t: 'Our build behaved today'}, {t: 'The day it does not, that is the difference'}, {t: 'between an afternoon and a week', hot: true}]},   // Z9
];

export const Nocode18: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode18/vo" files={FILES} gap={GAP} />
);
