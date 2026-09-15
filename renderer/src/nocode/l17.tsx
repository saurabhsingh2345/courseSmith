// nocode17 — the Cursor build (his L17, 9:41)
//
// The first of four builds from the same brief. 27 of the 49 slides are real
// screen recordings, cut from four takes: the plan, the 11-minute build, the
// finished board, and the interaction checks.
//
// **Our result differs from his, and the lecture is honest about it.** His
// first pass produced janky drag-and-drop, a persistent Next.js error badge,
// and no in-column reordering, so his lecture is a feedback-and-iterate story.
// Ours worked on the first pass — five columns, our amber palette, working drag
// between and within columns, add and delete, plus 11 unit tests and 5
// Playwright tests the agent wrote and ran itself. So the lecture makes the
// honest point instead: this went well because the brief was good, and the
// five checks still get run by hand because the agent graded its own homework.
//
// Two real-world details kept rather than edited out: it could not use port
// 3000 (already busy on this machine) and said so, and the sandbox stopped to
// ask permission mid-build — which is exactly the setting L16 ends on.

import React from 'react';
import {Deck, Slide, deckFrames, C} from './kit';

export const DURS = [
  7.53, 10.255, 13.314, 17.2, 9.65, 8.823, 13.675, 7.248, 14.004,
  11.536, 15.623, 15.119, 12.498, 9.487, 11.855, 11.175, 12.685, 2.094,
  2.248, 13.13, 10.837, 10.479, 13.051, 6.803, 12.53, 12.131, 12.058,
  7.254, 15.01, 7.901, 13.409, 13.779, 11.441, 15.527, 16.928, 15.379,
  12.259, 9.512, 13.244, 13.09, 14.778, 16.791, 14.717, 12.345, 9.578,
  8.442, 17.905, 6.318,
];
export const FILES = [
  '01', '02', '03', '04', '05', '07', '08', 'A1', 'A2', 'A3',
  'A4', 'A5', 'A6', '10', '11', '12', '13', 'A9', '14', '15',
  '16', 'A10', 'A11', 'X1', 'X2', 'X3', 'X4', 'A13', 'A14', '18',
  '19', 'X5', '21', '22', '23', '24', '25', '26', '27', '28',
  '29', '30', '31', '32', '33', '34', '35', '36',
].map((n) => `${n}.mp3`);
export const GAP = 0.39;
export const L17_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode17/shots/${n}.mp4`;
const CHECKS = ['loads with\ndata', 'drag between\ncolumns',
                'reorder in\na column', 'add and\ndelete', 'would you\nshow it'];

const SLIDES: Slide[] = [
  {k: 'head', kicker: 'nothing left to talk about', lines: ['Let us *build*', 'the thing'],
   size: 106, trans: 'fade'},
  {k: 'head', kicker: 'first, give it some room',
   lines: ['We are going all in', 'on the *agent*'], size: 68, trans: 'push'},
  {k: 'myth', kicker: 'second, the model', trans: 'rise',
   lines: ['I am going to *name* it'], size: 66,
   wrong: 'leave it on Auto and let the tool decide',
   right: 'pick it, so you know what you are comparing',
   note: 'if each tool picks for itself, a difference tells you nothing'},
  {k: 'rows', kicker: 'and two reasons for that', trans: 'rise',
   lines: ['Claude *Opus 5*'], size: 88,
   rows: [{t: 'The strongest thing available on this account'},
          {t: 'Four tools compared near their best, not near their cheapest'},
          {t: 'And when the results differ, I know it was not the model', hot: true}]},
  {k: 'head', kicker: 'third, and this is the one people skip',
   lines: ['Switch it to', '*Plan* mode'], size: 96, trans: 'fade',
   stamp: 'the pill changes colour so you can see which mode you are in'},
  {k: 'shot', src: sh('07'), kicker: 'and now watch how little i type',
   lines: ['*Four words*'], size: 96},
  {k: 'shot', src: sh('08'), kicker: 'not one of them says what to build',
   lines: ['It does not', '*need* me to'], size: 70},
  {k: 'shot', src: sh('A1'), kicker: 'the plan is worth thirty seconds',
   lines: ['Not the vague bullet', 'list people *expect*'], size: 54},
  {k: 'shot', src: sh('A2'), kicker: 'the first heading',
   lines: ['Scope. Do this,', '*nothing else*'], size: 58},
  {k: 'shot', src: sh('A3'), kicker: 'the fence from our file, in its own words',
   lines: ['The list of *nos*,', 'doing its job'], size: 56},
  {k: 'shot', src: sh('A4'), kicker: 'then the stack',
   lines: ['Current. Obvious.', 'Not *exotic*'], size: 58},
  {k: 'shot', src: sh('A5'), kicker: 'and the one to notice',
   lines: ['A *library* for the', 'drag and drop'], size: 58},
  {k: 'shot', src: sh('A6'), kicker: 'neither of these was in the requirements',
   lines: ['Vitest. *Playwright*'], size: 74},
  {k: 'shot', src: sh('10'), kicker: 'reading the brief, thinking about structure',
   lines: ['A *document*,', 'not code'], size: 68},
  {k: 'shot', src: sh('11'), kicker: 'phases, architecture, out of scope, execution order',
   lines: ['The *plan*'], size: 96},
  {k: 'head', kicker: 'now the honest thing to do here',
   lines: ['Read it. *Disagree*', 'with it'], size: 76, trans: 'push',
   stamp: 'the highest-leverage thirty seconds in the whole process'},
  {k: 'head', kicker: 'and i am going to be straight with you',
   lines: ['I am *not* going', 'to read it'], size: 84, trans: 'fade',
   stamp: 'you probably should. this is what most people actually do'},
  {k: 'shot', src: sh('A9'), kicker: 'so', lines: ['*Build*'], size: 150},
  {k: 'shot', src: sh('14'), kicker: 'and now we watch',
   lines: ['The clearest look at', 'an *agent* all week'], size: 54},
  {k: 'shot', src: sh('15'), kicker: 'a gitignore, then a frontend directory',
   lines: ['None of it needed', 'a *decision* from me'], size: 54},
  {k: 'shot', src: sh('16'), kicker: 'watch the file tree on the left',
   lines: ['A project that did not', 'exist a *minute* ago'], size: 50},
  {k: 'shot', src: sh('A10'), kicker: 'and keep half an eye on two things',
   lines: ['The tree, and the', '*context* indicator'], size: 54},
  {k: 'shot', src: sh('A11'), kicker: 'every file, every command, every test result',
   lines: ['It *climbs* for the', 'whole build'], size: 56},

  // ------------------------------------------- ours: the quiet goal change
  {k: 'head', kicker: 'one thing to watch for while that runs',
   lines: ['The most common way', 'this goes *wrong*'], size: 62, trans: 'push'},
  {k: 'myth', kicker: 'if it hits something it cannot do', trans: 'rise',
   lines: ['It changes the *goal*'], size: 72,
   wrong: 'drag and drop is hard — I will say so',
   right: 'drag and drop is hard — here are two arrow buttons',
   note: 'and then it reports success, because from where it sits it delivered'},
  {k: 'head', kicker: 'and it is not lying',
   lines: ['The letter of it.', 'Not the *point* of it'], size: 66, trans: 'fade',
   stamp: 'the same failure we keep meeting this week'},
  {k: 'ladder', kicker: 'which is why these are written down before the build',
   lines: ['Five checks, fixed', 'in *advance*'], size: 54, items: CHECKS,
   trans: 'rise'},
  {k: 'shot', src: sh('A13'), kicker: 'and here it has stopped and asked',
   lines: ['*Permission*'], size: 104},
  {k: 'shot', src: sh('A14'), kicker: 'the sandbox setting from the last lecture',
   lines: ['Doing exactly what', 'I said it *would*'], size: 56},
  {k: 'shot', src: sh('18'), kicker: 'and here is the moment i wanted you to see',
   lines: ['Tests fail. It goes', 'back and *fixes* them'], size: 52},
  {k: 'shot', src: sh('19'), kicker: 'a model, in a loop, with tools',
   lines: ['The whole *definition*,', 'on screen'], size: 54},
  {k: 'head', kicker: 'one practical note on that indicator',
   lines: ['Past *seventy* percent', 'on a build this size'], size: 58, trans: 'push',
   stamp: 'stop, and start again with a tighter brief'},

  // ---------------------------------------------------------------- the result
  {k: 'shot', src: sh('21'), kicker: 'six of six tasks complete',
   lines: ['Let us be the', '*judge* of that'], size: 62},
  {k: 'head', kicker: 'one detail worth knowing',
   lines: ['It did *not* start on', 'port 3000'], size: 66, trans: 'fade',
   stamp: 'read the line rather than assuming the address'},
  {k: 'shot', src: sh('23'), kicker: 'backlog, ready, in progress, review, done',
   lines: ['That is *our* amber,', 'not a template'], size: 54},
  {k: 'shot', src: sh('24'), kicker: 'check one — does it load with data in it?',
   lines: ['Read what is on', 'the *cards*'], size: 60},
  {k: 'shot', src: sh('25'), kicker: 'check two — drag between columns',
   lines: ['Both counts *update*'], size: 70},
  {k: 'shot', src: sh('26'), kicker: 'check three — the harder half',
   lines: ['Reorder *inside*', 'a column'], size: 66},
  {k: 'shot', src: sh('27'), kicker: 'check four — and watch the column empty',
   lines: ['Somebody thought about', 'the *empty state*'], size: 50},
  {k: 'shot', src: sh('28'), kicker: 'check five — a judgement, not a test',
   lines: ['Would I *show* it', 'to someone?'], size: 60},

  // -------------------------------------------------- ours: why this went well
  {k: 'head', kicker: 'and now the honest part',
   lines: ['This first pass', '*worked*'], size: 104, trans: 'push',
   stamp: 'which is not what usually happens'},
  {k: 'myth', kicker: 'it says it tested itself', trans: 'rise',
   lines: ['Should you *believe* it?'], size: 62,
   wrong: '11 unit tests and 5 browser tests all pass',
   right: 'evidence the agent chose, and graded',
   note: 'much better than simply asserting it is done. still not your eyes'},
  {k: 'rows', kicker: 'which is why the order matters', trans: 'rise',
   lines: ['Its tests, then', '*yours*'], size: 74,
   rows: [{t: 'It ran its own checks and reported them'},
          {t: 'Then I ran the five, by hand, on camera'},
          {t: 'They agreed — and mine are the ones that count', hot: true}]},
  {k: 'head', kicker: 'and this did not go well by luck',
   lines: ['The *brief* was good'], size: 106, trans: 'fade',
   stamp: 'every awkward decision was made before the agent started'},
  {k: 'head', kicker: 'the last two lectures, cashed out',
   lines: ['Work in the file is work', 'it cannot *guess* at'], size: 56,
   trans: 'push', stamp: 'and guessing is where these things go wrong'},
  {k: 'head', kicker: 'so — one of four',
   lines: ['Three more tools', 'to *go*'], size: 96, trans: 'fade',
   stamp: 'and I do not know which of them will struggle'},
  {k: 'rows', kicker: 'to recap, because it went past quickly', trans: 'rise',
   lines: ['That loop is', 'the *job* now'], size: 68,
   rows: [{t: 'A good agents file, and a model picked deliberately'},
          {t: 'Plan before build, and read what comes back'},
          {t: 'Let it run — and let the sandbox ask'},
          {t: 'Then test against a list you wrote in advance', hot: true}]},
  {k: 'head', kicker: 'next — the same brief, the same five checks',
   lines: ['In *GitHub Copilot*'], size: 92, trans: 'push'},
];

export const Nocode17: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode17/vo" files={FILES} gap={GAP} />
);
