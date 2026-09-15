// nocode14 — beyond the hype (his L14, 7:13) — and the end of Day 2
//
// Two halves, same as his: the anti-hype argument, then Artificial Analysis.
// His spine kept whole — the honest range, the order-of-magnitude case and the
// cases where it slowed him down, the project being the variable, "a multiplier
// but not a 10x", the two camps, the Intelligence Index, the November 2025 step
// change, his on-the-record prediction that the curve will not keep steepening,
// the coding and tool-use boards, and the Day 2 close at 13%.
//
// The numbers are OURS and they are current — checked on artificialanalysis.ai
// on 2026-08-28, not copied from his lecture, which cites a superseded
// generation. His own lecture tells you to expect exactly that, so reading the
// board as it stands on our day is the faithful thing to do rather than the
// unfaithful one. Every frame carrying a number is dated on screen.
//
// This is the one lecture in Day 2 where he has real footage. Ours draws the
// same charts rather than filming his site: a capture would age the moment it
// rendered, and a chart we draw can be dated, sourced and re-cut in a minute
// when the board moves. `board` is the new template — a ranked index with the
// spread called out, because the spread is the actual teaching point and his
// screen recording buries it.
//
// Ours: X1-X5 — how to misread a leaderboard. He shows you the site and never
// says to read the price column alongside the intelligence one, or that the
// benchmark measures the model while what you ship is the model plus your
// tools plus your file plus your workflow.

import React from 'react';
import {Deck, Slide, deckFrames, C} from './kit';

export const DURS = [
  15.504, 18.337, 12.421, 16.916, 15.933, 12.333, 15.144, 12.72, 11.549,
  13.615, 11.098, 14.09, 16.93, 17.207, 12.07, 16.155, 14.46, 11.816,
  18.438, 9.963, 16.968, 13.861, 4.773, 14.336, 20.53, 13.0, 12.734,
  7.269, 8.863, 15.061, 13.256,
];
export const FILES = [
  '01', '02', '03', '04', '05', '06', '07', '08', '09', '10',
  '11', '12', '13', '14', '15', '16', '17', '18', '19', '20',
  '21', '22', 'X1', 'X2', 'X2b', 'X3', 'X4', 'X5', '23', '24',
  '25',
].map((n) => `${n}.mp3`);
export const GAP = 0.19;
export const L14_FRAMES = deckFrames(DURS, GAP);

// artificialanalysis.ai, Intelligence Index, read 2026-08-28
const INDEX = [
  {t: 'Claude Opus 5 (max)', score: 63},
  {t: 'Claude Fable 5', score: 62},
  {t: 'GPT-5.6 Sol (max)', score: 61},
  {t: 'Grok 4.6 (high)', score: 61},
];

const SLIDES: Slide[] = [
  // ------------------------------------------------------------- the honest range
  {k: 'head', kicker: 'a sobering one to finish the day on',
   lines: ['Some of the hype', 'is *not* warranted'], size: 76, trans: 'fade',
   stamp: 'better from me than on a deadline'},
  {k: 'windows', kicker: 'one job — a boilerplate front end', trans: 'rise',
   lines: ['The *honest* range'], size: 68, unit: '',
   items: [{t: 'A coding agent', tokens: 20, note: 'minutes'},
           {t: 'A good front-end engineer', tokens: 480, note: 'a solid day'},
           {t: 'Me, and it is not my field', tokens: 14400, note: 'several weeks'}],
   caption: 'forms, a table, some routing — nothing clever'},
  {k: 'head', kicker: 'and it is real',
   lines: ['An order of', '*magnitude*'], size: 116, trans: 'push',
   stamp: 'this is the result people write posts about'},
  {k: 'rows', kicker: 'but that is one end of a range', trans: 'rise',
   lines: ['The other end does', 'not get *written* about'], size: 54,
   rows: [{t: 'Tasks where it is only incrementally faster'},
          {t: 'And a smaller number where it slowed me down'},
          {t: 'Subtly wrong, and I lost more time finding it than writing it', hot: true}]},
  {k: 'myth', kicker: 'and the variable is not the one people name', trans: 'rise',
   lines: ['The *project* decides'], size: 66,
   wrong: 'pick a better model and it gets faster',
   right: 'innovative work on a big codebase stays incremental',
   note: 'greenfield, from nothing, full of boilerplate — days turn into minutes'},
  {k: 'head', kicker: 'so, in aggregate',
   lines: ['A multiplier.', 'Not a *10x*'], size: 116, trans: 'push',
   stamp: 'anybody selling 10x across all your work has not measured all of theirs'},
  {k: 'cols', kicker: 'and there are two camps now', trans: 'fade',
   lines: ['Both of them', 'are *wrong*'], size: 66,
   cols: [{head: 'the over-hyped', items: ['engineering is finished',
                                            'ship it all tonight']},
          {head: 'the anti-AI movement', items: ['burnt by the first camp',
                                                  'decided it is a con']}]},
  {k: 'head', kicker: 'which is on those of us actually using them',
   lines: ['Set the record', '*straight*'], size: 106, trans: 'push',
   stamp: 'on the strengths, and on the limits'},

  // ------------------------------------------------------------ the one bookmark
  {k: 'head', kicker: 'if you bookmark one site out of this program',
   lines: ['artificialanalysis', '*.ai*'], size: 96, trans: 'fade',
   stamp: 'independent · updated constantly · every model that matters'},
  {k: 'head', kicker: 'and the disclaimer matters more here than anywhere',
   lines: ['You will *not* see', 'what I am seeing'], size: 68, trans: 'push',
   stamp: 'this field moves in weeks — take the method, not the numbers',
   stampColor: C.red},
  {k: 'head', kicker: 'the headline number',
   lines: ['The *Intelligence*', 'Index'], size: 92, trans: 'fade',
   stamp: 'one score, so you need not hold eight leaderboards in your head'},
  {k: 'board', kicker: 'checked while making this lecture', trans: 'rise',
   lines: ['The top of the', 'index *today*'], size: 56,
   items: INDEX, max: 70, dateline: 'artificialanalysis.ai · 28 august 2026'},
  {k: 'board', kicker: 'and notice the spread, not the order', trans: 'none',
   lines: ['*Two points* separate', 'the top four'], size: 54,
   items: INDEX, max: 70, spread: true,
   dateline: 'artificialanalysis.ai · 28 august 2026',
   caption: 'closer together than the marketing around them suggests'},
  {k: 'rows', kicker: 'so the choice is rarely about the ranking', trans: 'rise',
   lines: ['Price, speed, and the', 'room you *work* in'], size: 54,
   rows: [{t: 'My own default is Claude, and the index is only part of why'},
          {t: 'The other part is that I use it inside Claude Code'},
          {t: 'The model and the harness around it are not separable', hot: true}]},

  // --------------------------------------------------------------- the charts
  {k: 'head', kicker: 'but this is where i would spend your time',
   lines: ['The *history*', 'chart'], size: 100, trans: 'fade',
   stamp: 'the shape of the change, not today’s snapshot'},
  {k: 'curve', kicker: 'and there is a visible step in it', trans: 'rise',
   lines: ['*November 2025*'], size: 74, yLabel: 'frontier score',
   xLabel: 'time', rising: true, step: 0.55, stepLabel: 'nov 2025',
   caption: 'roughly where trust in coding agents changed'},
  {k: 'head', kicker: 'so if a colleague tried these and wrote them off',
   lines: ['Ask them *when*'], size: 128, trans: 'push',
   stamp: 'a lot of the strongest objections are scar tissue from before that step'},
  {k: 'curve', kicker: 'and then the slightly spooky one', trans: 'fade',
   lines: ['Up, to the right,', 'and *steepening*'], size: 58,
   yLabel: 'frontier intelligence', xLabel: 'time', rising: true, accel: true,
   caption: 'long enough that it stops looking like a trend and starts looking like a law'},
  {k: 'head', kicker: 'so here is a prediction, on the record',
   lines: ['That curve will *not*', 'keep steepening'], size: 66, trans: 'push',
   stamp: 'reasoning was a one-time unlock, not a permanent gradient',
   stampColor: C.red},
  {k: 'head', kicker: 'and i would rather be wrong than unfalsifiable',
   lines: ['Check it in', '*six months*'], size: 100, trans: 'fade'},

  // ------------------------------------------------- the boards that matter to us
  {k: 'myth', kicker: 'and scroll further down, because we care about two others',
   lines: ['Coding. And *tool use*'], size: 58, trans: 'rise',
   wrong: 'the headline index is the one to read',
   right: 'reasoning and reliable tool-calling are different skills',
   note: 'in agentic coding, the second one is what you feel all day'},
  {k: 'head', kicker: 'and that is where the surprises live',
   lines: ['A different *order*', 'on every board'], size: 70, trans: 'push',
   stamp: 'worth knowing before you commit a project to one of them'},

  // ------------------------------------------- ours: how to misread a leaderboard
  {k: 'head', kicker: 'two things before we leave it',
   lines: ['A leaderboard is easy', 'to *misread*'], size: 68, trans: 'fade'},
  {k: 'board', kicker: 'one — read across, not down', trans: 'rise',
   lines: ['Two points down,', 'a *third* of the price'], size: 52,
   items: INDEX, max: 70, cost: true,
   dateline: 'read the price and speed columns alongside this one',
   caption: 'if you only ever read the top row you will never find it'},
  {k: 'occupancy', kicker: 'and speed compounds, in a way it does not in chat',
   lines: ['One call, or *forty*'], size: 66, trans: 'rise',
   used: 2.5, cols: 40, rows: 1, legend: 'a chat answer is one call',
   caption: 'half the speed is not half as annoying — it is a different working day'},
  {k: 'stack', kicker: 'two — and this is the one that catches people', trans: 'fade',
   lines: ['The benchmark measures', 'one *layer*'], size: 52,
   layers: [{t: 'The model', sub: 'what the benchmark scores', h: 1.2, tone: C.red},
            {t: 'Your tools', sub: 'and how well described they are', h: 1.1},
            {t: 'Your agents file', sub: 'the part you write', h: 1.1, tone: C.blue},
            {t: 'Your workflow', sub: 'which of the six rungs', h: 1.2}],
   hot: [0], foot: 'and what you ship is the whole of it'},
  {k: 'myth', kicker: 'which is why the ranking is not the whole story', trans: 'rise',
   lines: ['Setup *beats* rank'], size: 68,
   wrong: 'the top model wins',
   right: 'a middling model in a good environment beats it',
   note: 'the last three lectures were the parts you control. this chart is the part you do not'},
  {k: 'head', kicker: 'so read it like a spec sheet for a component',
   lines: ['It tells you what is', '*possible*'], size: 68, trans: 'push',
   stamp: 'it does not tell you what you will build'},

  // ---------------------------------------------------------------- day two ends
  {k: 'head', kicker: 'and that is day two',
   lines: ['You have put up with', 'a lot of *talking*'], size: 66, trans: 'fade',
   stamp: 'and none of it was optional'},
  {k: 'rows', kicker: 'here is what you walked out with', trans: 'rise',
   lines: ['The whole', '*foundation*'], size: 76,
   rows: [{t: 'What a model is, and what an agent adds to it'},
          {t: 'What the context is, and why it degrades before it fills'},
          {t: 'What belongs in your agents file'},
          {t: 'And the six workflows you get to choose between', hot: true}]},
  {k: 'grid', kicker: 'tomorrow — tools, tools, tools',
   lines: ['Cursor. Copilot.', 'Codex. *Antigravity*'], done: 2, now: 2,
   trans: 'push'},
];

export const Nocode14: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode14/vo" files={FILES} gap={GAP} />
);
