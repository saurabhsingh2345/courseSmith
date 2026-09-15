// nocode04 — the roadmap (his L4, 8:15)
//
// His first 3:50 is autobiography: employer, acquisition, Times Square, the
// aeroplane, the student count, "message me". None of that is ours and none of
// it teaches anything, so it is gone. What remains is the part that actually
// orients a learner: breadth and depth, the three kinds of day, the three-week
// map, and the four products you end up owning. ~4:23 rather than 8:15.

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';

export const DURS = [
  9.32, 11.85, 22.05, 11.86, 17.10, 8.65, 13.53, 14.08,
  13.89, 6.98, 13.38, 24.84, 24.12, 7.33, 21.50, 15.91,
];
export const FILES = Array.from({length: 16}, (_, i) => `${String(i + 1).padStart(2, '0')}.mp3`);
export const GAP = 1.8;
export const L4_FRAMES = deckFrames(DURS, GAP);

const SLIDES: Slide[] = [
  {k: 'head', kicker: 'the three weeks', lines: ['Where you *are*', 'and what is next'], size: 76},
  {k: 'head', kicker: 'two ideas, held together', lines: ['*Breadth*', 'and depth'], size: 104},
  {k: 'rows', kicker: 'breadth — so nothing is unfamiliar', lines: ['Many *tools*'], size: 84,
   rows: [{t: 'Cursor'}, {t: 'GitHub Copilot'}, {t: 'OpenAI Codex'},
          {t: 'Google Antigravity'}, {t: 'Claude Code, OpenCode, and others'}]},
  {k: 'head', kicker: 'and none of it is gated', lines: ['Free to', '*follow along*'], size: 82,
   stamp: 'watch the paid parts'},
  {k: 'rows', kicker: 'depth — so you are actually good at something',
   lines: ['One thing, *properly*'], size: 58,
   rows: [{t: 'Coding in the command line'},
          {t: 'Claude Code above all', hot: true},
          {t: 'Weeks two and three live there almost entirely'}]},
  {k: 'head', kicker: 'the structure', lines: ['Three kinds', 'of *day*'], size: 100},
  {k: 'rows', kicker: 'the first kind', lines: ['Core *skills*'], size: 90,
   rows: [{t: 'How the models actually work'}, {t: 'What context really is'},
          {t: 'Why agents fail the way they do'},
          {t: 'Today is one of these', hot: true}]},
  {k: 'rows', kicker: 'the second kind', lines: ['*Platforms*'], size: 90,
   rows: [{t: 'Open the real products and drive them'},
          {t: 'Lots of screen, lots of typing'},
          {t: 'Things going wrong, on camera', hot: true}]},
  {k: 'rows', kicker: 'the third kind', lines: ['*Projects*'], size: 90,
   rows: [{t: 'Something real, built end to end'},
          {t: 'Every one commercial — except today’s', hot: true}]},
  {k: 'ladder', kicker: 'three weeks, five days each', lines: ['The *shape*'],
   items: ['Week one — vibe coding', 'Week two — vibe engineering',
           'Week three — as an expert'], pick: 0},
  {k: 'rows', kicker: 'week one · where we are now', lines: ['Vibe *coding*'], size: 84,
   rows: [{t: 'Meet the tools'}, {t: 'Learn how to talk to an agent'},
          {t: 'Empty folder to working software'},
          {t: 'Ends on a commercial product, not a toy', hot: true}]},
  {k: 'rows', kicker: 'week two · the term is simon willison’s',
   lines: ['Vibe *engineering*'], size: 62,
   rows: [{t: 'Down into the command line'},
          {t: 'Slash commands, checkpoints, MCP'},
          {t: 'Skills, plugins, long autonomous loops'},
          {t: 'A Jira ticket in — a GitHub push out', hot: true}]},
  {k: 'rows', kicker: 'week three', lines: ['As an *expert*'], size: 90,
   rows: [{t: 'Sub-agents, teams of agents, swarms'},
          {t: 'Ten at once, in sandboxes, while you make coffee'},
          {t: 'Large existing codebases — the honest rough edges'},
          {t: 'And a capstone', hot: true}]},
  {k: 'head', kicker: 'what you will actually have', lines: ['*Four* products'], size: 104},
  {k: 'rows', lines: ['Yours, and *real*'], size: 76, numbered: true,
   rows: [{t: 'A personal site with an AI version of you in it'},
          {t: 'A Kanban platform you can talk to in plain English'},
          {t: 'A SaaS that drafts legal documents and hands you a PDF'},
          {t: 'A live trading workstation with a reasoning assistant'}]},
  {k: 'head', kicker: 'and then you can answer that post',
   lines: ['You will have', 'read the *manual*'], size: 76,
   stamp: 'tomorrow — how it works'},
];

export const Nocode04: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode04/vo" files={FILES} gap={GAP} />
);
