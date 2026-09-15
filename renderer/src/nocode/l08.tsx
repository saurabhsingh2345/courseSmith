// nocode08 — wrapping up day one (his L8, 3:48)
//
// His outro spine: the noise warning as a callback to L6, "your models are
// stronger than mine", the divergence caveat, the curriculum grid with day one
// filled in, tomorrow's trailer, and the percentage beat. His 1:50-3:00 block is
// "you get me as well" — Q&A, email, LinkedIn — which does not transfer, so it
// is replaced with the thing a learner actually needs at that moment: what to
// do when your result diverges from ours, and how to describe a failure. The
// grid and the progress beat are built to be recalled every lesson from here.

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';

// slide 07 is the curriculum grid: its line is only 4.8s, so the window is held
// open to 9s and the grid gets the room his outro gives it.
export const DURS = [
  8.706, 18.5, 18.151, 19.953, 23.969, 15.384, 9.0, 21.321, 14.158, 11.383,
];
export const FILES = Array.from({length: 10}, (_, i) =>
  `${String(i + 1).padStart(2, '0')}.mp3`);
export const GAP = 2.2;
export const L8_FRAMES = deckFrames(DURS, GAP);

const SLIDES: Slide[] = [
  {k: 'head', kicker: 'that is day one', lines: ['Four things', 'to *carry*'], size: 96},
  {k: 'rows', kicker: 'the most useful habit in this program',
   lines: ['Ignore the *noise*'], size: 82,
   rows: [{t: 'Something is announced weekly that sounds decisive'},
          {t: 'Almost none of it will matter in six weeks'},
          {t: 'Watch the few sources that hold up, let the rest arrive filtered',
           hot: true}]},
  {k: 'rows', kicker: 'and this one is in your favour',
   lines: ['Your models are', '*stronger*'], size: 70,
   rows: [{t: 'Newer than the ones these lessons were built against'},
          {t: 'So your results will differ, and often be better'},
          {t: 'Take joy in that rather than reading it as a mismatch', hot: true}]},
  {k: 'rows', kicker: 'the honest caveat', lines: ['Every project', '*diverges*'], size: 76,
   rows: [{t: 'No piece of code here produces one matching result'},
          {t: 'You are not typing along — you are instructing something creative'},
          {t: 'It builds something slightly different every single time', hot: true}]},
  {k: 'rows', kicker: 'so when yours does not match ours',
   lines: ['Check *behaviour*,', 'not pixels'], size: 62,
   rows: [{t: 'Does it run. Does it do what was asked. Does it break where ours broke'},
          {t: 'Something extra is not a failure — that is your version'},
          {t: 'Read it, decide if you like it, then say what to change', hot: true}]},
  {k: 'rows', kicker: 'and when it genuinely goes wrong',
   lines: ['Do not start *over*'], size: 80,
   rows: [{t: 'Say exactly what you saw, and paste the error in full'},
          {t: 'Ask it to explain before it repairs'},
          {t: 'Describing a problem precisely is most of the skill', hot: true}]},
  {k: 'grid', kicker: 'three weeks, five days each', lines: ['Where you *are*'],
   done: 1, now: 0},
  {k: 'rows', kicker: 'and what day one actually gave you',
   lines: ['More than it *felt* like'], size: 58,
   rows: [{t: 'A working 3D game you shipped without writing a line'},
          {t: 'Vibe coder, vibe engineer, agentic coder — pinned down'},
          {t: 'The three shapes, and which to reach for'},
          {t: 'Eight stages, mapped to the three weeks, and what it costs',
           hot: true}]},
  {k: 'rows', kicker: 'tomorrow — under the hood', lines: ['What is it', '*actually* doing?'], size: 62,
   rows: [{t: 'What a language model does when it answers you'},
          {t: 'What makes something an agent rather than a chat window'},
          {t: 'And a file called agents.md, at the heart of all of it', hot: true}]},
  {k: 'pct', kicker: 'day one of fifteen', pct: 7, sub: 'see you tomorrow'},
];

export const Nocode08: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode08/vo" files={FILES} gap={GAP} />
);
