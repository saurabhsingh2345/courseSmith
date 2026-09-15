// nocode15 — the lab setup (his L15, 8:27) — Day 3 opens
//
// His shape kept: ~2.5 minutes of ground rules, then straight into the lab.
// Four principles, the node check, the terminal primer (pwd / cd), the real
// clone of our starter repo, and the project opening in the editor.
//
// **Codex is replaced by Claude Code throughout Day 3**, on his instruction.
// The four tools are Cursor, Copilot, Claude Code and Antigravity. That is a
// deliberate deviation from the reference and arguably the better line-up —
// Claude Code is what most professionals have actually settled on.
//
// Nine slides are real screen recordings, cut from one continuous 90-second
// lab take (`L15_E_lab`) plus the project-opening take. Shot on a dark Cursor
// theme — the fix for Day 1's bright-footage complaint — from
// /Users/Shared/projects so `pwd` prints no username, and with the zsh prompt
// neutralised to the path alone, because the default prints `user@host` on
// every line and here that is his name and a hostname carrying the company.
// The clone URL does show his GitHub handle; he approved that explicitly.
//
// nodejs.org is a graphic rather than a browser capture: driving Chrome risks
// his tabs and profile in frame for a beat that is just "download the installer".

import React from 'react';
import {Deck, Slide, deckFrames, C} from './kit';

export const DURS = [
  10.43, 10.692, 11.912, 16.935, 8.294, 12.417, 14.037, 14.69, 9.852,
  15.926, 14.059, 12.794, 11.732, 13.154, 12.277, 12.731, 9.288, 7.149,
  13.602, 9.165, 11.444, 10.674, 13.539, 13.879, 14.735, 12.887, 11.84,
  12.19, 11.984, 10.694, 10.754, 9.624, 10.274, 6.153, 13.925, 11.023,
  5.907, 9.92, 12.514, 12.408, 13.702, 10.269,
];
export const FILES = [
  '01', '02', '03', '04', 'A1', 'A2', 'A3', 'A4', 'A5', '05',
  '06', '07', '08', '09', 'A6', 'A7', 'A8', '10', '11', '12',
  '13', '14', '15', '16', '17', 'A13', '18', '19', '20', '21',
  '22', 'A9', 'A10', 'A11', 'A12', '23', 'X1', 'X2', 'X3', 'X4',
  'X5', '24',
].map((n) => `${n}.mp3`);
export const GAP = 0.38;
export const L15_FRAMES = deckFrames(DURS, GAP);

const TOOLS = [
  {t: 'Cursor', sub: 'VS Code with an agent bolted on', h: 1},
  {t: 'GitHub Copilot', sub: 'an extension, not an app', h: 1},
  {t: 'Claude Code', sub: 'lives in the terminal', h: 1, tone: C.yellow},
  {t: 'Antigravity', sub: "Google's, running Gemini", h: 1, tone: C.blue},
];

const SLIDES: Slide[] = [
  // ---------------------------------------------------------- the ground rules
  {k: 'grid', kicker: 'day three', lines: ['Less talking.', 'More *doing*'],
   done: 2, now: 2, trans: 'fade'},
  {k: 'head', kicker: 'and this is where people get stuck or get going',
   lines: ['Four ground', '*rules*'], size: 116, trans: 'push',
   stamp: 'and which one it is has little to do with ability'},
  {k: 'rows', kicker: 'rule one, and it is the one that matters', trans: 'rise',
   lines: ['Everything today', 'is *optional*'], size: 62,
   rows: [{t: 'Four tools. You do not need to install all four'},
          {t: 'Watch all of it, pick one'},
          {t: 'And stick with that one for the rest of the week', hot: true}]},
  {k: 'stack', kicker: 'the short version of which to pick', trans: 'fade',
   lines: ['Four *tools*'], size: 84, layers: TOOLS,
   foot: 'friendliest · already in your editor · what professionals use · the new one'},

  // ------------------------------------------------------- ours: the four, properly
  {k: 'head', kicker: 'because the differences matter more than the logos',
   lines: ['Let me introduce', 'them *properly*'], size: 78, trans: 'push'},
  {k: 'rows', kicker: 'one — cursor', trans: 'rise',
   lines: ['The most', '*approachable*'], size: 74,
   rows: [{t: 'A fork of VS Code with the agent panel built into the side'},
          {t: 'Where we built the game on day one', hot: true}]},
  {k: 'rows', kicker: 'two — github copilot', trans: 'rise',
   lines: ['An *extension*,', 'not an app'], size: 62,
   rows: [{t: 'Lives inside Visual Studio Code'},
          {t: 'Your route in if you do not want to move editors'},
          {t: 'And the one most companies have already paid for', hot: true}]},
  {k: 'rows', kicker: 'three — claude code', trans: 'rise',
   lines: ['A different *shape*'], size: 74,
   rows: [{t: 'Runs in the terminal, not in an editor window'},
          {t: 'Which sounds like a step backwards, and is not'},
          {t: 'Where a lot of professional developers have settled', hot: true}]},
  {k: 'rows', kicker: 'four — antigravity', trans: 'rise',
   lines: ['The *newest* of', 'the four'], size: 66,
   rows: [{t: "Google's entry, running Gemini"},
          {t: 'Most opinionated about how you should work'},
          {t: 'Worth seeing because it does not copy the others', hot: true}]},

  // ---------------------------------------------------------- rules two to four
  {k: 'myth', kicker: 'rule two', trans: 'fade',
   lines: ['Your results *will* vary'], size: 60,
   wrong: 'the same prompt gives the same app',
   right: 'different model, different tier, different day',
   note: 'that is not you doing it wrong — it is the nature of the thing'},
  {k: 'head', kicker: 'rule three, and write this one down',
   lines: ['*Simplify,* simplify,', 'simplify'], size: 82, trans: 'push'},
  {k: 'rows', kicker: 'when you are properly stuck', trans: 'rise',
   lines: ['A working *ugly* thing', 'beats a broken pretty one'], size: 50,
   rows: [{t: 'Cut the scope. Ask for less'},
          {t: 'Get something small working, then build out from there'},
          {t: 'And it is a far better place to give feedback from', hot: true}]},
  {k: 'flow', kicker: 'and if it is still stuck after that', trans: 'rise',
   lines: ['There is no prize for', 'finishing all *four*'], size: 54,
   nodes: [{t: 'give feedback', sub: 'tell it what is wrong'},
           {t: 'delete and retry', sub: 'start the prompt again'},
           {t: 'skip it', sub: 'move to the next tool', tone: C.yellow}]},
  {k: 'head', kicker: 'rule four',
   lines: ['Have some *fun*', 'with it'], size: 104, trans: 'fade',
   stamp: 'four tools, one brief, four different ways to fail'},

  // ------------------------------------------------- ours: why a kanban board
  {k: 'head', kicker: 'so why a kanban board?',
   lines: ['Deliberately', '*chosen*'], size: 116, trans: 'push'},
  {k: 'rows', kicker: 'it sits between a tutorial and a product', trans: 'rise',
   lines: ['Drag and drop is the', '*fiddly* part'], size: 56,
   rows: [{t: 'Layout, and state that survives being moved around'},
          {t: 'And drag and drop, which is genuinely awkward to get right'},
          {t: 'That last one is what separates the four builds', hot: true}]},
  {k: 'head', kicker: 'and small enough to finish in one sitting',
   lines: ['Four *complete*', 'results to compare'], size: 74, trans: 'fade',
   stamp: 'rather than four half-finished ones'},

  // ---------------------------------------------------------------- the setup
  {k: 'head', kicker: 'and the setup is the same for all four builds',
   lines: ['So we do it once,', '*properly*'], size: 80, trans: 'push'},
  {k: 'head', kicker: 'the one thing you need first',
   lines: ['nodejs', '*.org*'], size: 128, trans: 'fade',
   stamp: 'download the build for your machine and run the installer'},
  {k: 'shot', src: 'nocode15/shots/19.mp4', kicker: 'and to check it worked',
   lines: ['The *terminal*.', 'Command J'], size: 62},
  {k: 'head', kicker: 'and if you have never used one',
   lines: ['Nothing to be', '*afraid* of'], size: 96, trans: 'fade',
   stamp: 'it looks like the 1980s because it is'},
  {k: 'shot', src: 'nocode15/shots/21.mp4', kicker: 'first command',
   lines: ['*node --version*'], size: 70},
  {k: 'shot', src: 'nocode15/shots/22.mp4', kicker: 'second — print working directory',
   lines: ['*pwd*. Where am I?'], size: 66},
  {k: 'shot', src: 'nocode15/shots/23.mp4', kicker: 'third — change directory',
   lines: ['*cd ..* goes up', 'one level'], size: 60},
  {k: 'myth', kicker: 'one aside, because it catches people', trans: 'rise',
   lines: ['Slashes lean *differently*'], size: 54,
   wrong: 'Windows: C:\\Users\\you\\projects',
   right: 'Mac and Linux: /Users/you/projects',
   note: 'if you are on a PC and a path looks wrong, that is usually why'},
  {k: 'rows', kicker: 'two last terminal things', trans: 'rise',
   lines: ['Then we *go*'], size: 84,
   rows: [{t: 'A folder name with a space in it needs quotes round the path'},
          {t: 'And the up arrow brings back your last command', hot: true}]},

  // ------------------------------------------------------------- the clone
  {k: 'head', kicker: 'now the thing we came for',
   lines: ['There is a *starter*', 'repository'], size: 74, trans: 'push',
   stamp: 'the address is in the resources for this lecture'},
  {k: 'shot', src: 'nocode15/shots/27.mp4', kicker: 'git clone — make me a local copy',
   lines: ['You will use this', 'command *for life*'], size: 56},
  {k: 'shot', src: 'nocode15/shots/28.mp4', kicker: 'enumerating, counting, compressing',
   lines: ['*Four* objects'], size: 76},
  {k: 'shot', src: 'nocode15/shots/29.mp4', kicker: 'cd into it, and confirm where we are',
   lines: ['*pwd* again'], size: 76},
  {k: 'shot', src: 'nocode15/shots/30.mp4', kicker: 'ls — list what is here',
   lines: ['Two files.', 'That is *all*'], size: 66},

  // ------------------------------------------------- ours: what to check, and cost
  {k: 'rows', kicker: 'so here is what to check at the end of each build', trans: 'rise',
   lines: ['The *same five*,', 'every time'], size: 60,
   rows: [{t: 'Does the board appear with cards already in it?'},
          {t: 'Can you drag a card between columns?'},
          {t: 'Can you reorder cards inside a column — the harder half?'}]},
  {k: 'rows', kicker: 'and the last two', trans: 'none',
   lines: ['That is what makes it', 'a *comparison*'], size: 56,
   rows: [{t: 'Can you add a card, and delete a card?'},
          {t: 'And would you be willing to show it to someone?'},
          {t: 'Five checks — not four demonstrations', hot: true}]},
  {k: 'head', kicker: 'five checks, same five every time',
   lines: ['A *comparison*, not four', 'demonstrations'], size: 58, trans: 'push'},
  {k: 'head', kicker: 'one honest note before we start',
   lines: ['Four builds is', '*real money*'], size: 84, trans: 'fade',
   stamp: 'watching rather than building is a legitimate way to take this day'},
  {k: 'shot', src: 'nocode15/shots/35.mp4', kicker: 'open the folder as a project',
   lines: ['*KANBAN*'], size: 92},

  // ----------------------------------- ours: why four results, from one brief
  {k: 'head', kicker: 'one thing worth saying before we go on',
   lines: ['Four *different*', 'applications'], size: 80, trans: 'fade'},
  {k: 'myth', kicker: 'and the temptation is to rank them', trans: 'rise',
   lines: ['That is not quite', 'what is *happening*'], size: 56,
   wrong: 'the best app means the best model',
   right: 'only one thing changes between these builds',
   note: 'and it is not the file — the file is identical all four times'},
  {k: 'stack', kicker: 'remember where we finished yesterday', trans: 'fade',
   lines: ['Only the *harness*', 'changes'], size: 62,
   layers: [{t: 'The model', sub: 'roughly comparable across the four', h: 1.1},
            {t: 'The tools it has', sub: 'and how it is allowed to use them', h: 1.2,
             tone: C.yellow},
            {t: 'agents.md', sub: 'identical, all four times', h: 1.0, tone: C.blue},
            {t: 'How much you let it run', sub: 'the setting we are about to meet', h: 1.2,
             tone: C.yellow}],
   hot: [1, 3]},
  {k: 'rows', kicker: 'so watch for the differences that come from the harness',
   lines: ['Those are what you', 'are *choosing* between'], size: 54, trans: 'rise',
   rows: [{t: 'How it plans, and whether it shows you the plan'},
          {t: 'Whether it asks before it runs something'},
          {t: 'And what it does when a test fails', hot: true}]},
  {k: 'head', kicker: 'and if you hit an install problem on the way in',
   lines: ['Paste the error into', 'the *agent*'], size: 70, trans: 'push',
   stamp: 'using the tool to install the tool is entirely fair'},
  {k: 'head', kicker: 'next — we read that file line by line',
   lines: ['The most important', 'thing you write *all week*'], size: 62,
   trans: 'fade', stamp: 'and all four builds live or die by it'},
];

export const Nocode15: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode15/vo" files={FILES} gap={GAP} />
);
