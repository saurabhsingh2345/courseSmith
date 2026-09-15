// nocode07 — the eight stages, and what it costs (his L7, 10:31)
//
// His spine kept whole: Yegge's eight stages, the mapping onto the three weeks,
// his stages-five-and-six verdict, then the costs half. His stages are a row of
// static generated pictures; ours is one ladder that grows rung by rung and is
// then lit up week by week — the graphic we call back to at the top of every
// week. Ours are slides 13-16 (what the ladder actually measures, the trade it
// makes, that you move up and down it daily, and the jump-to-the-top mistake)
// and 26-28 (what a token is, why reading costs more than writing, and the
// three habits that keep the bill small) — he charges for it without ever
// saying what is being charged.

import React from 'react';
import {Deck, Slide, deckFrames, C} from './kit';

export const DURS = [
  14.884, 16.122, 13.477, 15.836, 16.205, 16.073, 18.944, 17.204, 17.09, 15.887,
  18.693, 11.094,
  21.337,
  15.133, 17.911, 19.49, 18.213,
  26.008,
  13.239, 14.206, 13.269, 11.335, 15.842, 15.49, 5.813, 12.591, 14.448, 19.953,
  16.737, 19.774, 18.321, 16.851, 16.34, 16.467, 18.003, 15.015,
];
export const FILES = [
  ...Array.from({length: 12}, (_, i) => String(i + 1).padStart(2, '0')),
  'X1',
  ...Array.from({length: 4}, (_, i) => String(i + 13).padStart(2, '0')),
  'X2',
  ...Array.from({length: 18}, (_, i) => String(i + 17).padStart(2, '0')),
].map((n) => `${n}.mp3`);
export const GAP = 1.36;
export const L7_FRAMES = deckFrames(DURS, GAP);

const RUNGS = [
  'chat window, you retype it',
  'sidebar, asks every time',
  'sidebar, no asking',
  'agent takes the window',
  'the command line',
  'three to five at once',
  'ten or more, you steer',
  'an agent runs the agents',
];

const rung = (upto: number, pick?: number, kicker?: string, lines?: string[]): Slide => ({
  k: 'ladder', kicker, lines: lines ?? [`Stage *${String(upto).padStart(2, '0')}*`],
  items: RUNGS, upto, pick,
});

const SLIDES: Slide[] = [
  {k: 'head', kicker: 'day one, and the map of everything after it',
   lines: ['Eight', '*stages*'], size: 130},
  {k: 'head', kicker: 'where the ladder comes from', lines: ['Steve *Yegge*'], size: 104,
   stamp: 'more on gastown later'},
  {k: 'head', kicker: 'and the one thing every rung measures',
   lines: ['How much do you', 'hand over *unread*?'], size: 62},

  rung(1, 0, 'you and a chat window', ['Stage *01*']),
  {k: 'shot', src: 'nocode07/shots/04.mp4', kicker: 'it moves into your editor',
   lines: ['Stage *02*'], size: 88},
  {k: 'shot', src: 'nocode07/shots/05.mp4', kicker: 'one setting changes',
   lines: ['Stage *03*'], size: 88},
  rung(4, 3, 'the code leaves the screen', ['Stage *04*']),
  rung(5, 4, 'no windows at all', ['Stage *05*']),
  rung(6, 5, 'one becomes several', ['Stage *06*']),
  rung(7, 6, 'and several becomes many', ['Stage *07*']),
  rung(8, 7, 'you stop coordinating', ['Stage *08*']),
  {k: 'ladder', kicker: 'and there it is, end to end',
   lines: ['The whole *ladder*'], items: RUNGS, upto: 8},

  // ---- ours: he never separates the working style from the product ----
  {k: 'shot', src: 'nocode07/shots/12.mp4', kicker: 'one clarification before we map it',
   lines: ['A stage is not', 'a *product*'], size: 66},
  {k: 'rows', kicker: 'what it is really tracking', lines: ['Three things *move*'], size: 66,
   rows: [{t: 'How much of the output you actually read'},
          {t: 'How many things are running at once'},
          {t: 'Where your attention is pointed', hot: true}]},
  {k: 'rows', kicker: 'and every rung pays the same price',
   lines: ['Review, for', '*throughput*'], size: 66,
   rows: [{t: 'Higher is faster, with less certainty about what you have'},
          {t: 'Lower is slower, and you know exactly what landed'},
          {t: 'Neither one is simply correct', hot: true}]},
  {k: 'rows', kicker: 'which nobody says out loud', lines: ['Nobody *lives*', 'on one rung'], size: 66,
   rows: [{t: 'Delicate and load bearing — drop to two and read every change'},
          {t: 'Mechanical across forty files — go to five and let it run'},
          {t: 'It is a choice per task, not an identity', hot: true}]},
  {k: 'rows', kicker: 'and the mistake almost everyone makes',
   lines: ['Skipping to', 'the *top*'], size: 74,
   rows: [{t: 'Ten agents in week one goes badly, and specifically'},
          {t: 'You cannot supervise what you cannot evaluate'},
          {t: 'Each rung needs the judgement built on the one below', hot: true}]},
  // ---- ours: the obvious question his lecture leaves hanging ----
  {k: 'shot', src: 'nocode07/shots/17.mp4', kicker: 'so what replaces reading every line?',
   lines: ['You read', '*behaviour*'], size: 82},
  {k: 'head', kicker: 'so where does this program sit',
   lines: ['Stage one is', '*assumed*'], size: 76},
  {k: 'ladder', kicker: 'week one — where you are now',
   lines: ['Two, three, *four*'], items: RUNGS, picks: [1, 2, 3], upto: 8},
  {k: 'ladder', kicker: 'week two — a full week on one rung',
   lines: ['The *command line*'], items: RUNGS, picks: [4], upto: 8, tone: C.blue},
  {k: 'ladder', kicker: 'week three — where we finish',
   lines: ['Six, seven, *eight*'], items: RUNGS, picks: [5, 6, 7], upto: 8, tone: C.red},
  {k: 'ladder', kicker: 'and the honest verdict for real software today',
   lines: ['*Five* and *six*', 'are the sweet spot'], items: RUNGS, picks: [4, 5], upto: 8},
  {k: 'rows', kicker: 'seven and eight', lines: ['Shown, not', '*prescribed*'], size: 74,
   rows: [{t: 'They exist, they are moving fast, you should see them'},
          {t: 'Against production code this year — an open question'},
          {t: 'Ask again in six months', hot: true}]},

  {k: 'head', kicker: 'second half, and the practical one',
   lines: ['What does', 'it *cost*?'], size: 116},
  {k: 'rows', kicker: 'the reassuring part, and it is true',
   lines: ['*Zero* is a', 'real path'], size: 76,
   rows: [{t: 'The whole program is completable without spending a penny'},
          {t: 'Every day has a free route through it'},
          {t: 'We point it out as we go', hot: true}]},
  {k: 'rows', kicker: 'and the other side, equally true',
   lines: ['It can also', 'get *expensive*'], size: 66,
   rows: [{t: 'Twenty agents, one hour, a large codebase'},
          {t: 'Nothing on screen looks expensive while it happens', hot: true}]},
  {k: 'rows', kicker: 'so let us say what is actually charged',
   lines: ['You pay for', '*tokens*'], size: 76,
   rows: [{t: 'A token is roughly three quarters of a word'},
          {t: 'In: your question, your files, the whole conversation so far'},
          {t: 'Out: the answer'},
          {t: 'Both directions are billed', hot: true}]},
  {k: 'rows', kicker: 'and here is what surprises everybody',
   lines: ['*Reading* costs', 'more than writing'], size: 62,
   rows: [{t: 'A long chat resends its entire history every turn'},
          {t: 'Two hours of conversation is paid for again on every message',
           hot: true}]},
  {k: 'rows', kicker: 'three habits, none of which cost you anything',
   lines: ['Keeping it *small*'], size: 76, numbered: true,
   rows: [{t: 'Start a fresh conversation when the topic changes'},
          {t: 'Point at the files that matter, not the whole repository'},
          {t: 'Cheap model for mechanical work, strong one for thinking',
           hot: true}]},
  {k: 'cols', kicker: 'the recommended path through the three weeks',
   lines: ['Where the money *goes*'],
   cols: [{head: 'Week 1', items: ['Inside an editor', 'The free trial covers it']},
          {head: 'Weeks 2-3', items: ['On the command line', 'An entry monthly plan']},
          {head: 'Beyond', items: ['Optional entirely', 'For intensity, not completion']}]},
  {k: 'rows', kicker: 'and if you would rather spend nothing at all',
   lines: ['That path *exists*'], size: 76,
   rows: [{t: 'Free models cover a large part of what we do'},
          {t: 'Each day names them in its resources'},
          {t: 'Paid products get shown — seeing is not buying', hot: true}]},
  {k: 'rows', kicker: 'one caution about numbers', lines: ['Prices go *stale*'], size: 84,
   rows: [{t: 'Tiers, names and free allowances change constantly'},
          {t: 'They vary by region and by whatever offer is running'},
          {t: 'If we disagree with the website, believe the website', hot: true}]},
  {k: 'rows', kicker: 'and the rule that matters most',
   lines: ['Your spend', 'is *yours*'], size: 82,
   rows: [{t: 'Nobody else is watching that number for you'},
          {t: 'Set a limit wherever the provider allows one'},
          {t: 'Look at usage in week one, not at the end of the month', hot: true}]},
  {k: 'rows', kicker: 'a moment of perspective', lines: ['A laptop, and a', '*subscription*'], size: 62,
   rows: [{t: 'A thousand pound machine shocks nobody'},
          {t: 'Twenty a month somehow does — one arrives again every month'},
          {t: 'Underneath: trillions of operations, and somebody pays for the electricity',
           hot: true}]},
  {k: 'head', kicker: 'so, to close', lines: ['You are in charge', 'of your *charges*'],
   size: 62, stamp: 'tomorrow: under the hood'},
];

export const Nocode07: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode07/vo" files={FILES} gap={GAP} />
);
