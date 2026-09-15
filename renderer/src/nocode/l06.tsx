// nocode06 — the vocabulary and the landscape (his L6, 8:58)
//
// His spine kept whole: the vibe-coding post, the emotional year, the inflection
// point, the three confusable terms, the three shapes of tool, and signal vs
// noise. The X slides are ours — he lists the three shapes and never says which
// one to actually reach for, or that the model matters more than the wrapper.

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';

export const DURS = [
  11.897, 8.585, 16.087, 10.373, 14.265, 19.487, 15.829, 16.777, 9.614, 8.794,
  16.351, 13.017, 9.889, 15.113, 16.373, 11.906, 16.099, 16.796, 13.618, 21.84,
  23.111, 15.555,
  10.856, 17.34, 17.027, 19.874, 11.241, 19.314, 20.272,
  7.42, 12.489, 17.907,
];
export const FILES = [
  ...Array.from({length: 22}, (_, i) => `${String(i + 1).padStart(2, '0')}`),
  'X1', 'X2', 'X3', 'X4', 'X5', 'X6', 'X7',
  '23', '24', '25',
].map((n) => `${n}.mp3`);
export const GAP = 2.0;
export const L6_FRAMES = deckFrames(DURS, GAP);

const SLIDES: Slide[] = [
  {k: 'head', kicker: 'the vocabulary', lines: ['What these words', 'actually *mean*'], size: 76},
  {k: 'head', kicker: 'a second post, from a good while earlier',
   lines: ['Where *vibe coding*', 'got its name'], size: 72},
  {k: 'quote', kicker: 'what he wrote then', who: 'Andrej Karpathy',
   text: 'Give in fully to the vibes, and forget that the code even exists'},
  {k: 'head', kicker: 'and then came the year that followed',
   lines: ['An emotional', '*rollercoaster*'], size: 78},
  {k: 'rows', kicker: 'the first three stages', lines: ['*Surprise*, first'], size: 78,
   rows: [{t: 'It could do far more than anyone expected'},
          {t: 'Denial — a parlour trick, surely'},
          {t: 'Astonishment — and then they tried it themselves'}]},
  {k: 'rows', kicker: 'and then it got ugly', lines: ['*Frustration*'], size: 92,
   rows: [{t: 'It did not work as cleanly as the demos'},
          {t: 'And then, genuinely, anger'},
          {t: 'Performance reports, so it read back its own mistakes', hot: true}]},
  {k: 'rows', kicker: 'and finally', lines: ['*Acceptance*'], size: 92,
   rows: [{t: 'Learn where it is strong'}, {t: 'Learn where it is weak'},
          {t: 'This program gets you there without the year of pain', hot: true}]},
  {k: 'head', kicker: 'and then something actually changed',
   lines: ['The *inflection*', 'point'], size: 88, stamp: 'late last year'},
  {k: 'rows', kicker: 'which matters to you, practically',
   lines: ['Worth *another* look'], size: 62,
   rows: [{t: 'The repeated, maddening mistakes largely stopped'},
          {t: 'If you gave up a year ago, that thing no longer exists', hot: true}]},

  {k: 'head', kicker: 'three terms, constantly confused',
   lines: ['Let us *pin* them', 'down'], size: 78},
  {k: 'rows', lines: ['Vibe *coder*'], size: 90,
   rows: [{t: 'Karpathy’s original term'},
          {t: 'Sometimes: a fairly amateur approach'},
          {t: 'Sometimes: this entire field'},
          {t: 'Both usages are common — that is the problem', hot: true}]},
  {k: 'rows', lines: ['Vibe *engineer*'], size: 84,
   rows: [{t: 'Simon Willison’s term'},
          {t: 'The professional side of the same thing'},
          {t: 'Real software, to a real standard, built with agents'},
          {t: 'This is week two', hot: true}]},
  {k: 'rows', lines: ['Agentic *coder*'], size: 84,
   rows: [{t: 'Where the expert lands'},
          {t: 'An agent as a genuine collaborator'},
          {t: 'Or several at once — that is week three', hot: true}]},
  {k: 'rows', kicker: 'and now the trap', lines: ['Same words,', '*two* directions'], size: 62,
   rows: [{t: 'A person who codes WITH agents — that is us', hot: true},
          {t: 'Or a person who BUILDS agents — a different job entirely'}]},
  {k: 'rows', kicker: 'and it gets worse', lines: ['Person, or *product*?'], size: 62,
   rows: [{t: 'The platforms themselves get called coding agents'},
          {t: 'There is no fixing this'},
          {t: 'When it matters, just ask which one they mean', hot: true}]},

  {k: 'head', kicker: 'the tools', lines: ['Only *three*', 'shapes'], size: 100},
  {k: 'shot', src: 'nocode06/shots/16.mp4', kicker: 'shape one',
   lines: ['The *IDE*'], size: 100},
  {k: 'shot', src: 'nocode06/shots/17.mp4', kicker: 'shape two',
   lines: ['The *plugin*'], size: 96},
  {k: 'rows', kicker: 'shape three', lines: ['The *command line*'], size: 62,
   rows: [{t: 'A terminal. Text, no windows, no buttons'},
          {t: 'It looks like something from 1985'},
          {t: 'If you have never used one, it looks hostile'}]},
  {k: 'rows', kicker: 'and here is the surprise', lines: ['The *third* one', 'won'], size: 84,
   rows: [{t: 'It started almost by accident'},
          {t: 'It became Claude Code'},
          {t: 'Now essentially every company ships one'},
          {t: 'By week two you will understand exactly why', hot: true}]},
  {k: 'cols', kicker: 'the landscape, sorted into the three',
   lines: ['Where each one *lives*'],
   cols: [{head: 'IDE', items: ['Cursor', 'Codex', 'Antigravity', 'Windsurf']},
          {head: 'Plugin', items: ['GitHub Copilot', 'inside VS Code']},
          {head: 'CLI', items: ['Claude Code', 'Codex, Gemini', 'OpenCode, Amp']}]},
  {k: 'shot', src: 'nocode06/shots/21.mp4', kicker: 'one honest footnote',
   lines: ['The boxes *leak*'], size: 84},
  {k: 'head', kicker: 'but which should you reach for?',
   lines: ['The honest', '*answer*'], size: 104},
  {k: 'shot', src: 'nocode06/shots/23.mp4',
   lines: ['Reach for the *IDE*'], size: 62},
  {k: 'shot', src: 'nocode06/shots/24.mp4',
   lines: ['Reach for the *plugin*'], size: 58},
  {k: 'rows', lines: ['Reach for the *CLI*'], size: 62,
   rows: [{t: 'When you want to hand over real work'},
          {t: 'Trivial to script, and to run several of at once'},
          {t: 'Leave it running for an hour while you do something else'},
          {t: 'Which is why weeks two and three live there', hot: true}]},
  {k: 'head', kicker: 'and now the part that matters most',
   lines: ['Chassis,', 'and *engine*'], size: 96},
  {k: 'rows', lines: ['Spend on the *model*'], size: 62,
   rows: [{t: 'A weak model in a beautiful IDE gives you nonsense'},
          {t: 'A frontier model in an ugly terminal gives you excellence'},
          {t: 'Pay for the engine, not for the wrapper', hot: true}]},
  {k: 'rows', kicker: 'and about all these brand names',
   lines: ['Anchor on the', '*shapes*'], size: 72,
   rows: [{t: 'The brands will change — some will be gone in a year'},
          {t: 'The three shapes will not'},
          {t: 'Every new tool is an example of something you already know', hot: true}]},

  {k: 'head', kicker: 'the last thing, and the most useful habit',
   lines: ['Signal,', 'and *noise*'], size: 96},
  {k: 'rows', lines: ['A *tsunami*'], size: 104,
   rows: [{t: 'New tools, new frameworks, new claims — every single day'},
          {t: 'Almost none of it will matter in six weeks'}]},
  {k: 'rows', kicker: 'and occasionally', lines: ['Something *lands*'], size: 84,
   rows: [{t: 'Claude Code was one of those'},
          {t: 'Telling the two apart, quickly, is a real skill'},
          {t: 'Focus on what actually holds up'},
          {t: 'Tomorrow — under the hood', hot: true}]},
];

export const Nocode06: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode06/vo" files={FILES} gap={GAP} />
);
