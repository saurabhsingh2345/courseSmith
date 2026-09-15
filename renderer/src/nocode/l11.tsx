// nocode11 — context engineering (his L11, 13:00)
//
// The conceptual centre of Week 1. His spine kept whole: the stateless model,
// prompt engineering becoming context engineering, the layered stack built up
// one row at a time, agents.md joining it, the hard window limit, the far more
// important degradation before that limit, compacting and the fear of it, the
// undercut of his own habit, and the window sizes.
//
// This is the lecture the new templates were built for. `stack` is the layered
// context diagram and it is the most reused graphic in the program — it comes
// back in L12, at every compaction beat, and in weeks two and three. `curve`
// draws the quality decline he only ever says out loud. `compact` folds a
// history into a summary, which is the one idea here no still frame can carry.
// `windows` draws the model sizes and then, honestly, the usable slice of them.
//
// Deliberately light on `rows` — four in fifty-five slides against L9's
// twenty-six in forty-eight. Twelve distinct panel kinds carry the middle.
//
// Ours are the X slides. He tells you the context degrades and never tells you
// how to see it happening: X1-X4 put `/context` on screen and read the fixed
// tax out loud. X5-X8 replace the habit he just talked you out of with three
// that cost nothing.

import React from 'react';
import {Deck, Slide, deckFrames, C} from './kit';

export const DURS = [
  13.707, 16.093, 16.907, 16.357, 13.665, 12.62, 14.68, 12.051, 18.502,
  16.561, 11.698, 12.201, 9.328, 16.07, 15.445, 12.133, 14.226, 7.812,
  16.579, 13.161, 17.614, 15.764, 8.722, 12.905, 13.553, 7.905, 11.513,
  12.204, 15.028, 14.233, 8.226, 12.565, 15.801, 14.481, 11.321, 8.921,
  14.736, 9.639, 11.97, 9.582, 14.056, 16.863, 9.009, 10.665, 13.887,
  13.742, 14.448, 7.84, 15.71, 15.02, 14.601, 13.061, 12.875, 14.574,
  13.467, 14.448, 13.897, 11.308,
];
export const FILES = [
  '01', '02', '03', '04', '05', '06', '07',
  '08', '09', '10', '11', '12', '13', '14', '15', '16', '17',
  '18', '19', '20', '21', '22',
  '23', '24', '25',
  '26', '27', '28', '29', '30', '31',
  'X1', 'X2', 'X3', 'X4',
  '32', '33', '34', '35', '36', '37', '38', '39',
  'X5', 'X6', 'X7', 'X8',
  '40', '41', '42', '43', '44',
  'Y1', 'Y2', 'Y3',
  '45', '46', '47',
].map((n) => `${n}.mp3`);
export const GAP = 0.25;
export const L11_FRAMES = deckFrames(DURS, GAP);

// the stack before agents.md has its own row, and after. Revealing the second
// array in full after the first reads as a new layer slotting into place.
const STACK4 = [
  {t: 'System prompt', sub: 'role, job, tone', h: 1.1},
  {t: 'Tool descriptions', sub: 'every tool, described in words', h: 1.2},
  {t: 'Memory', sub: 'carried between conversations', h: 1.0},
  {t: 'The conversation so far', sub: 'messages, reasoning, code,\ntool calls and their output', h: 4.4},
];
const STACK5 = [
  ...STACK4.slice(0, 3),
  {t: 'agents.md', sub: 'the one row you write', h: 1.2, tone: C.blue},
  STACK4[3],
];

const SLIDES: Slide[] = [
  // ---------------------------------------------------------------- the hook
  {k: 'head', kicker: 'day two, still mostly talking',
   lines: ['This will happen', 'to you *this week*'], size: 92, trans: 'fade'},
  {k: 'chat', kicker: 'two hours into a good session', trans: 'rise',
   lines: ['Same model.', 'Same project.'], size: 56,
   turns: [{who: 'you', t: 'tabs, never spaces — I have said this'},
           {who: 'model', t: 'understood, switching to tabs'},
           {who: 'you', t: 'now add the delete endpoint'},
           {who: 'model', t: 'done. quick check — tabs or spaces here?'}],
   caption: 'nothing changed about the model, and nothing changed about you'},
  {k: 'head', kicker: 'the most popular phrase of the year',
   lines: ['*Context*', 'engineering'], size: 122, trans: 'push',
   stamp: 'and for once the hype is earned'},
  {k: 'myth', kicker: 'start from where we finished', trans: 'rise',
   lines: ['It is *stateless*'], size: 66,
   wrong: 'it remembers our conversation',
   right: 'it starts from nothing, every single call',
   note: 'not a limitation nobody got round to fixing — it is what the thing is'},
  {k: 'flow', kicker: 'which leads somewhere strict', trans: 'rise',
   lines: ['Output comes from', 'the input. *Entirely*'], size: 58,
   nodes: [{t: 'the input', sub: 'all it will ever see'},
           {t: 'the model', sub: 'no state, no yesterday'},
           {t: 'the output', sub: 'nothing else fed it', tone: C.yellow}],
   caption: 'no hidden state, no side channel, nothing written down anywhere else'},
  {k: 'head', kicker: 'and because it is all there is',
   lines: ['It has a name.', 'The *context*'], size: 100, trans: 'push',
   stamp: 'getting it right is the job'},
  {k: 'meter', kicker: 'why the name changed', trans: 'rise',
   lines: ['Your prompt is the', '*small* part'], size: 62,
   pct: 4, fill: 'what you typed', rest: 'everything else it reads',
   caption: 'we called it prompt engineering and wrote articles about magic phrases'},

  // -------------------------------------------------- the stack, layer by layer
  {k: 'stack', kicker: 'everything it sees on one call', trans: 'fade',
   lines: ['Picture it as', 'a *stack*'], size: 60, layers: STACK4, upto: 0,
   foot: 'top to bottom, in order'},
  {k: 'stack', kicker: 'layer one — before anything you write',
   lines: ['The *system* prompt'], size: 60, layers: STACK4, upto: 1, hot: [0],
   foot: 'you did not write this. your tool did'},
  {k: 'stack', kicker: 'layer two',
   lines: ['*Tool* descriptions'], size: 60, layers: STACK4, upto: 2, hot: [1],
   foot: 'every tool costs you space before it has done anything'},
  {k: 'head', kicker: 'part of the system prompt, or its own layer?',
   lines: ['*Does not', 'matter*'], size: 118, trans: 'push',
   stamp: 'it is words, it is there, and it is not free'},
  {k: 'stack', kicker: 'layer three', trans: 'fade',
   lines: ['*Memory*'], size: 60, layers: STACK4, upto: 3, hot: [2],
   foot: 'the layer that makes it feel like it knows you'},
  {k: 'stack', kicker: 'layer four, and it is the big one',
   lines: ['The conversation', '*so far*'], size: 58, layers: STACK4, upto: 4, hot: [3],
   foot: '“so far” is doing an enormous amount of work'},
  {k: 'rows', kicker: 'what is actually in that layer', trans: 'rise',
   lines: ['All of it. Not the', 'parts you *remember*'], size: 54,
   rows: [{t: 'Your first message, and the reply'},
          {t: 'Every follow-up, and every answer'},
          {t: 'Its reasoning tokens — thinking you never saw on screen', hot: true}]},
  {k: 'term', kicker: 'and this is the one that catches people', trans: 'rise',
   lines: ['Every tool call —', 'and its *output*'], size: 54,
   title: 'agent session', prompt: '',
   term: [{t: '$ pytest tests/', kind: 'cmd'},
          {t: 'collected 214 items', kind: 'out'},
          {t: 'tests/test_board.py ......................  [ 31% ]', kind: 'out'},
          {t: 'tests/test_cards.py .......F..............  [ 68% ]', kind: 'out'},
          {t: 'tests/test_api.py ....................      [100% ]', kind: 'out'},
          {t: '… 200 more lines, all of it now context', kind: 'warn'}]},
  {k: 'occupancy', kicker: 'and it is re-sent on every turn from here', trans: 'rise',
   lines: ['One careless read', 'of a *big file*'], size: 56,
   used: 41, legend: 'one file, one test run',
   caption: 'more of your window than everything you have typed all day'},
  {k: 'stack', kicker: 'so look at the finished stack', trans: 'fade',
   lines: ['Almost none of it', 'is *yours*'], size: 58, layers: STACK4,
   foot: 'you typed forty words. it is reading sixty thousand tokens'},

  // ------------------------------------------------------------- agents.md
  {k: 'head', kicker: 'one more row, and you write this one',
   lines: ['agents', '*.md*'], size: 130, trans: 'push', accent: C.blue},
  {k: 'stack', kicker: 'memory, but a particular kind', trans: 'fade',
   lines: ['Facts about', '*this* project'], size: 58, layers: STACK5, hot: [3],
   foot: 'true on Tuesday, still true three weeks later'},
  {k: 'term', kicker: 'not a feature you switch on', trans: 'rise',
   lines: ['A text file.', 'In your *repo*'], size: 58, title: 'kanban', prompt: '',
   term: [{t: '$ ls', kind: 'cmd'},
          {t: 'agents.md   src/   tests/   README.md', kind: 'out'},
          {t: '$ git add agents.md', kind: 'cmd'},
          {t: 'committed like any other file', kind: 'ok'}]},
  {k: 'cols', kicker: 'same idea, different labels', trans: 'rise',
   lines: ['The filename depends', 'on the *tool*'], size: 54,
   cols: [{head: 'agents.md', items: ['Cursor', 'Codex', 'Copilot']},
          {head: 'claude.md', items: ['Claude Code']},
          {head: 'gemini.md', items: ['Antigravity']}]},
  {k: 'head', kicker: 'and some of you are already objecting',
   lines: ['“Every single', 'call, *forever*?”'], size: 88, trans: 'push',
   stamp: 'not exactly — and that is the next lecture'},

  // ------------------------------------------------------- the hard limit
  {k: 'head', kicker: 'so the stack is bigger than you thought',
   lines: ['How much', '*actually* fits?'], size: 96, trans: 'fade'},
  {k: 'gauge', kicker: 'the context window', trans: 'rise',
   lines: ['A hard limit,', 'set by the *model*'], size: 56,
   pct: 74, limit: 88, label: 'this conversation', limitLabel: 'the window',
   caption: 'measured in tokens — a ceiling on how far back it can look'},
  {k: 'gauge', kicker: 'and when you go past it', trans: 'none',
   lines: ['It does not', '*degrade*. It fails'], size: 56,
   pct: 104, limit: 88, label: 'this conversation', limitLabel: 'the window',
   over: true,
   caption: 'you get an error. the call does not go through'},

  // ------------------------------------------- the part that actually matters
  {k: 'head', kicker: 'if you take one thing from this lecture',
   lines: ['The limit is *not*', 'your problem'], size: 92, trans: 'push'},
  {k: 'curve', kicker: 'quality against how full it is', trans: 'fade',
   lines: ['It sags long before', 'the *wall*'], size: 58, cliff: true,
   caption: 'not at ninety percent. not up against it. well before'},
  {k: 'myth', kicker: 'and not in the way people assume', trans: 'rise',
   lines: ['What degrades'], size: 66,
   wrong: 'it gets slower',
   right: 'it gets less accurate',
   note: 'it loses the thread, and stops weighting the important parts properly'},
  {k: 'rows', kicker: 'which in practice looks like this', trans: 'rise',
   lines: ['You have *seen*', 'all three'], size: 60,
   rows: [{t: 'It forgets an instruction you gave it early on'},
          {t: 'It contradicts a decision you both made an hour ago'},
          {t: 'It reintroduces a bug it already fixed', hot: true}]},
  {k: 'curve', kicker: 'and here is the uncomfortable part', trans: 'fade',
   lines: ['Your best work is at', 'the *beginning*'], size: 54,
   cliff: true, mark: 0.12,
   bands: [{from: 0, to: 0.3, label: 'sharpest', tone: C.yellow},
           {from: 0.3, to: 0.68, label: 'drifting', tone: C.blue},
           {from: 0.68, to: 1, label: 'unreliable', tone: C.red}],
   caption: 'the long, invested, productive-feeling sessions are often its worst'},
  {k: 'head', kicker: 'three words, and they are annoying ones',
   lines: ['*Less* is more'], size: 132, trans: 'push',
   stamp: 'even when you are three hours deep'},

  // ------------------------------------------- ours: stop guessing, watch it
  {k: 'head', kicker: 'this is where most explanations stop',
   lines: ['You do not have', 'to *guess*'], size: 96, trans: 'fade'},
  {k: 'term', kicker: 'claude code prints it straight at you', trans: 'rise',
   lines: ['Type *slash context*'], size: 60, title: 'claude code', prompt: '',
   term: [{t: 'Context left until auto-compact: 23%', kind: 'warn'},
          {t: '> /context', kind: 'cmd'},
          {t: 'system prompt      11,200   5.6%', kind: 'out'},
          {t: 'tool descriptions  18,400   9.2%', kind: 'out'},
          {t: 'files read         74,900  37.5%', kind: 'out'},
          {t: 'conversation       49,100  24.6%', kind: 'out'},
          {t: 'free               46,400  23.2%', kind: 'ok'}]},
  {k: 'stack', kicker: 'expect a small shock the first time', trans: 'fade',
   lines: ['A fixed tax, paid', 'on turn *zero*'], size: 56, layers: STACK5, hot: [0, 1],
   foot: 'gone before you type a single character'},
  {k: 'head', kicker: 'once you can see the number',
   lines: ['Check the number', '*before* the prompt'], size: 78, trans: 'push',
   stamp: 'nine times out of ten, that is the answer'},

  // ------------------------------------------------------------- compacting
  {k: 'head', kicker: 'so what happens near the ceiling?',
   lines: ['It has a name.', '*Compacting*'], size: 100, trans: 'fade'},
  {k: 'compact', kicker: 'built into claude code', trans: 'rise',
   lines: ['The history folds', 'into a *summary*'], size: 56,
   blocks: 11, keep: 4, summary: 'summary of everything above',
   at: 3.0, freedLabel: 'room handed back',
   caption: 'the detail goes, the gist stays, and the conversation carries on'},
  {k: 'head', kicker: 'and people are frightened of it',
   lines: ['You are trusting it to', 'decide what *mattered*'], size: 72, trans: 'push',
   stamp: 'nobody stopped to ask you which parts were load-bearing'},
  {k: 'rows', kicker: 'from where you are sitting', trans: 'rise',
   lines: ['It looks like it', '*forgot*'], size: 66,
   rows: [{t: 'Something you said is summarised away'},
          {t: 'A correction you made goes with it'},
          {t: 'So it repeats a mistake you already fixed', hot: true}]},
  {k: 'head', kicker: 'and there is no warning when it happens',
   lines: ['Nothing tells you', 'it was *dropped*'], size: 80, trans: 'push',
   stamp: 'the session quietly gets worse, and you blame the model'},
  {k: 'flow', kicker: 'so a whole practice grew up around avoiding it', trans: 'rise',
   lines: ['Five steps, every', 'time it gets *close*'], size: 52,
   nodes: [{t: 'watch', sub: 'the number'}, {t: 'stop', sub: 'before it folds'},
           {t: 'write', sub: 'notes by hand'}, {t: 'kill', sub: 'the session'},
           {t: 'paste', sub: 'and begin again', tone: C.yellow}],
   caption: 'and it works. that is not the question'},
  {k: 'head', kicker: 'and I will be honest with you about it',
   lines: ['That habit is', '*superstition* now'], size: 82, trans: 'push',
   stamp: 'as of 2026', stampColor: C.red},
  {k: 'head', kicker: 'compacting has got a great deal better',
   lines: ['*Trust* the', 'compactor'], size: 128, trans: 'fade'},

  // --------------------------------------------- ours: what to do instead
  {k: 'head', kicker: 'so what is worth doing instead?',
   lines: ['Three habits.', '*Smaller and duller*'], size: 82, trans: 'push'},
  {k: 'myth', kicker: 'habit one', trans: 'rise',
   lines: ['One *task* per', 'conversation'], size: 60,
   wrong: 'the session is going well — keep it going',
   right: 'finish the thing, then start a new one',
   note: 'you are trading a warm context for an accurate one, and accurate wins'},
  {k: 'stack', kicker: 'habit two', trans: 'fade',
   lines: ['Scrolling back is a', '*missing line*'], size: 56, layers: STACK5, hot: [3],
   foot: 'write it once, instead of paying for it every turn'},
  {k: 'rows', kicker: 'habit three', trans: 'rise',
   lines: ['Do not argue with', 'it a *third* time'], size: 58,
   rows: [{t: 'The first wrong answer is in the context'},
          {t: 'So is the second one'},
          {t: 'It now reads them as examples of how this goes', hot: true}]},

  // ---------------------------------------------------------- the numbers
  {k: 'head', kicker: 'last thing, and I am dating it out loud',
   lines: ['The numbers.', '*Early 2026*'], size: 100, trans: 'push',
   stamp: 'these age faster than anything else here'},
  {k: 'windows', kicker: 'context window, by model', trans: 'fade',
   lines: ['What you are *sold*'], size: 62,
   items: [{t: 'Claude Haiku 4.5 — the cheap tier', tokens: 200000},
           {t: 'Gemini 3 Pro — Antigravity', tokens: 1000000},
           {t: 'Claude Opus 5', tokens: 1000000}]},
  {k: 'head', kicker: 'and they will be wrong soon',
   lines: ['Carry the *shape*,', 'not the digits'], size: 80, trans: 'push',
   stamp: 'the gap is no longer between vendors — it is between tiers'},
  {k: 'curve', kicker: 'because the curve does not care how wide it is',
   lines: ['A bigger room that', 'still gets *muddy*'], size: 54, trans: 'fade',
   cliff: true,
   bands: [{from: 0, to: 0.3, label: 'sharpest', tone: C.yellow},
           {from: 0.3, to: 0.68, label: 'drifting', tone: C.blue},
           {from: 0.68, to: 1, label: 'unreliable', tone: C.red}],
   caption: 'the shape is the same. only the axis got longer'},
  {k: 'windows', kicker: 'so this is the honest picture', trans: 'rise',
   lines: ['What stays', '*reliable*'], size: 62, usable: 0.42,
   items: [{t: 'Claude Haiku 4.5 — the cheap tier', tokens: 200000},
           {t: 'Gemini 3 Pro — Antigravity', tokens: 1000000},
           {t: 'Claude Opus 5', tokens: 1000000}],
   caption: 'a bigger window buys headroom, not permission to stop thinking'},

  // ------------------------------- ours: what that number means for your repo
  {k: 'tokens', kicker: 'a token count means nothing until you can picture it',
   lines: ['One line of code,', 'as it is *read*'], size: 54, trans: 'fade',
   chips: ['def', ' get', '_board', '(', 'board', '_id', ':', ' int', ')', ':'],
   hot: [1, 2, 5],
   caption: 'about three quarters of a word — and two to three tokens per line of code'},
  {k: 'meter', kicker: 'so convert the window into lines', trans: 'rise',
   lines: ['15,000 lines.', 'Your repo is *50,000*'], size: 54,
   pct: 30, fill: '200k window', rest: 'the rest of your codebase',
   caption: 'a serious project runs into the hundreds of thousands'},
  {k: 'head', kicker: 'and it never did',
   lines: ['Your window does not', 'hold your *project*'], size: 74, trans: 'push',
   stamp: 'which is exactly why that one file has to carry it'},

  // ---------------------------------------------------------------- close
  {k: 'head', kicker: 'and it is not a bag of tricks',
   lines: ['It is *one* idea'], size: 122, trans: 'fade'},
  {k: 'occupancy', kicker: 'that agent from the start of the lecture', trans: 'rise',
   lines: ['Brilliant at nine.', 'Useless by *four*'], size: 56,
   used: 94, legend: 'and nobody told you',
   caption: 'same model, same project, same person typing'},
  {k: 'head', kicker: 'next — the one part you fully control',
   lines: ['The *agents file*,', 'up close'], size: 84, trans: 'push',
   stamp: 'and the real rule about when it gets read'},
];

export const Nocode11: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode11/vo" files={FILES} gap={GAP} />
);
