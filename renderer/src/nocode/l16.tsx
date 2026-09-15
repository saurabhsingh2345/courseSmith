// nocode16 — reading agents.md (his L16, 9:49)
//
// The most directly copyable lecture in Week 1. He reads his hand-written brief
// line by line inside the editor; so do we, from our own starter repo, which
// viewers can clone and change.
//
// Sixteen slides are real screen recordings, almost all cut from one continuous
// 150-second walk down the file (`L16_E_walk`), plus the shortcut toggles, the
// one-file reveal, and the side-by-side raw/rendered preview that `ref-L16.md`
// calls the shot that teaches it. Shot on a dark Cursor theme from a neutral
// path.
//
// The Auto Run beat is a GRAPHIC, not a screen recording, and deliberately so:
// Cursor's settings pane shows the account email and team name on every
// section, so filming it breaks the no-PII rule. `ref-L16.md` recommended
// showing the consequence rather than the pane anyway. Our version also has
// four options to his three — this build of Cursor splits allowlist into
// sandboxed and unsandboxed.

import React from 'react';
import {Deck, Slide, deckFrames, C} from './kit';

export const DURS = [
  17.563, 14.035, 8.7, 12.576, 12.088, 17.926, 13.898, 9.902, 13.972,
  11.403, 11.294, 12.607, 10.83, 10.783, 13.009, 14.273, 12.787, 11.919,
  15.67, 7.242, 10.776, 15.718, 10.82, 10.642, 11.301, 16.842, 8.97,
  16.784, 9.152, 15.926, 6.249, 13.858, 13.707, 11.596, 13.567, 12.753,
  12.964, 16.594, 12.901, 9.447, 11.344, 12.099, 9.523, 12.962, 11.944,
  10.221, 13.225, 13.14,
];
export const FILES = [
  '01', '02', '03', '04', '05', '06', '07', '08', '09', '10',
  '11', '12', '13', '14', '15', '16', '17', '18', '19', '20',
  '21', '22', '23', '24', '25', '26', '27', '28', '29', '30',
  'X1', 'X2', 'X3', 'X4', 'X5', '31', '32', '33', '34', 'A1',
  'A2', 'A3', 'A4', 'A5', 'A6', 'A7', 'A8', 'A9',
].map((n) => `${n}.mp3`);
export const GAP = 0;
export const L16_FRAMES = deckFrames(DURS, GAP);

const RUNMODES = ['allow-\nlist', 'allowlist\n+ sandbox', 'auto-review\n+ sandbox',
                  'run every-\nthing'];

const s = (n: string) => `nocode16/shots/${n}.mp4`;

const SLIDES: Slide[] = [
  // ------------------------------------------------------ one file, two shortcuts
  {k: 'shot', src: s('00'), kicker: 'a lot of ceremony for one file',
   lines: ['And it holds', '*one* file'], size: 68},
  {k: 'shot', src: s('01'), kicker: 'two shortcuts worth muscle memory',
   lines: ['*Cmd B*. *Opt Cmd E*'], size: 62},
  {k: 'head', kicker: 'and on a PC',
   lines: ['Control B.', 'Control *Option E*'], size: 78, trans: 'push',
   stamp: 'cluttered editor to just the file, in about a second'},
  {k: 'shot', src: s('03'), kicker: 'sixty-nine lines, seven sections',
   lines: ['I *wrote* this.', 'I did not generate it'], size: 56},
  {k: 'head', kicker: 'which may sound perverse in a program like this',
   lines: ['This file *is* the', 'instruction'], size: 74, trans: 'fade',
   stamp: 'ask a machine to write its own and you have hidden the problem'},
  {k: 'shot', src: s('05'), kicker: 'raw on the left, rendered on the right',
   lines: ['Closer to how the', 'model *receives* it'], size: 54},

  // -------------------------------------------------------- business requirements
  {k: 'shot', src: s('06'), kicker: 'section one — business requirements',
   lines: ['An *MVP*. Not a', 'platform'], size: 62},
  {k: 'rows', kicker: 'start simple and stay simple', trans: 'rise',
   lines: ['Every line here', 'says *less*'], size: 62,
   rows: [{t: 'One board. Not multiple boards, not a board switcher'},
          {t: 'Five columns, fixed in number, but renameable'},
          {t: 'A card has a title and a description. Nothing else', hot: true}]},
  {k: 'shot', src: s('08'), kicker: 'and notice how blunt that is',
   lines: ['*Nothing* else'], size: 96},
  {k: 'rows', kicker: 'because if I do not say it', trans: 'rise',
   lines: ['I will get all', 'of *this*'], size: 66,
   rows: [{t: 'Due dates. Labels. Assignees'},
          {t: 'A priority dropdown nobody asked for'},
          {t: 'And then I have to read it, test it, and delete it', hot: true}]},
  {k: 'shot', src: s('10'), kicker: 'and then the fence',
   lines: ['No archive. No search.', 'No *filters*'], size: 54},
  {k: 'head', kicker: 'that list of nos',
   lines: ['The most valuable', '*paragraph* in the file'], size: 66, trans: 'push',
   stamp: 'every one of them is a feature it would otherwise cheerfully build'},
  {k: 'shot', src: s('12'), kicker: 'small line, big effect',
   lines: ['Opens with', '*dummy data*'], size: 74},
  {k: 'head', kicker: 'and the priority, stated so there is no ambiguity',
   lines: ['A *good-looking* interface', 'over more features'], size: 58,
   trans: 'fade', stamp: 'a viewer should want to use it'},
  {k: 'rows', kicker: 'notice the technique running through all of it', trans: 'rise',
   lines: ['Say what must', 'be *built*'], size: 68,
   rows: [{t: 'Be precise. Do not be ambiguous'},
          {t: 'Be assertive rather than suggestive'},
          {t: 'And where it matters, say what must NOT be built', hot: true}]},

  // ----------------------------------------------------------- technical details
  {k: 'shot', src: s('15'), kicker: 'section two — technical details',
   lines: ['Next.js, in a folder', 'called *frontend*'], size: 54},
  {k: 'rows', kicker: 'and the rest of it', trans: 'rise',
   lines: ['Prefer the *obvious*', 'over the clever'], size: 58,
   rows: [{t: 'No persistence. State lives in memory'},
          {t: 'No accounts, no login, no authentication'},
          {t: 'Keep the implementation as simple as the requirements allow', hot: true}]},
  {k: 'head', kicker: 'and that last line is a repeat, on purpose',
   lines: ['No harm in *repeating*', 'yourself'], size: 66, trans: 'push',
   stamp: 'it is not an essay — nobody is marking you down for it'},
  {k: 'shot', src: s('18'), kicker: 'section three — and it is explicitly optional',
   lines: ['A *colour scheme*'], size: 80},

  // ------------------------------------------------------------------- strategy
  {k: 'head', kicker: 'section four — strategy',
   lines: ['It stops describing the', 'product, and starts on', 'the *work*'],
   size: 54, trans: 'fade'},
  {k: 'shot', src: s('20'), kicker: 'a plan first, in phases',
   lines: ['With *success criteria*', 'that get ticked off'], size: 54},
  {k: 'shot', src: s('21'), kicker: 'and the line that saves the most time',
   lines: ['Complete when it is', '*ready to use*'], size: 56},
  {k: 'myth', kicker: 'because without that line', trans: 'rise',
   lines: ['It stops at what', 'looks *finished*'], size: 58,
   wrong: '"done" — meaning the code exists',
   right: '"done" — meaning a person could use it',
   note: 'those are not the same claim, and it will make the first one'},

  // ----------------------------------------------------------- coding standards
  {k: 'head', kicker: 'section five — coding standards',
   lines: ['Down to *three*'], size: 110, trans: 'push',
   stamp: 'from a list that used to be much longer'},
  {k: 'shot', src: s('24'), kicker: 'one — and note there is no date in it',
   lines: ['Current versions.', '*Idiomatic today*'], size: 54},
  {k: 'shot', src: s('25'), kicker: 'two — the one I would keep if I kept only one',
   lines: ['Keep it simple.', 'Never *over-engineer*'], size: 52},
  {k: 'head', kicker: 'and I see it often enough that you cannot stress it too much',
   lines: ['The most common', '*failure mode*'], size: 74, trans: 'fade',
   stamp: 'in agent-written code, by a distance'},
  {k: 'shot', src: s('27'), kicker: 'three — and block capitals genuinely work',
   lines: ['IMPORTANT: never', 'use *emojis*'], size: 54},
  {k: 'head', kicker: 'and that one is not fussiness',
   lines: ['They *break* some', 'Windows terminals'], size: 66, trans: 'push',
   stamp: 'and an interface full of them looks like a toy'},
  {k: 'shot', src: s('29'), kicker: 'section six — a short working agreement',
   lines: ['Two tries, then', '*stop and explain*'], size: 56},

  // ------------------------------------------- ours: what is deliberately absent
  {k: 'head', kicker: 'two additions of my own',
   lines: ['What is *not* in it,', 'on purpose'], size: 70, trans: 'fade'},
  {k: 'rows', kicker: 'anything that changes daily', trans: 'rise',
   lines: ['Does not belong in a', 'file read *every call*'], size: 54,
   rows: [{t: 'No sprint, no ticket number, no who-is-on-what'},
          {t: 'Not the bug you are fixing this week'},
          {t: 'Say those in the conversation, and let them go', hot: true}]},
  {k: 'head', kicker: 'and the second thing',
   lines: ['A *stale* line is worse', 'than a missing one'], size: 62, trans: 'push',
   stamp: 'missing makes it guess. stale makes it confidently wrong'},
  {k: 'rows', kicker: 'so prune it', trans: 'rise',
   lines: ['If it only ever grows,', 'it is *drifting*'], size: 56,
   rows: [{t: 'Solve a problem properly in the code'},
          {t: 'Then delete the line that was working around it'},
          {t: 'Or it starts costing more than it saves', hot: true}]},
  {k: 'term', kicker: 'and a way to check any of this is landing', trans: 'rise',
   lines: ['*Ask* it'], size: 88, title: 'a fresh session', prompt: '',
   term: [{t: '> summarise, in your own words, the instructions', kind: 'cmd'},
          {t: '  you are working under', kind: 'cmd'},
          {t: 'Build a Kanban MVP. One board, five columns.', kind: 'out'},
          {t: 'Keep it simple, no extra features, no emojis.', kind: 'out'},
          {t: 'Stop after two failed attempts and explain.', kind: 'out'},
          {t: 'twenty seconds, once a month', kind: 'ok'}]},

  // ------------------------------------------------- do you have to write these
  {k: 'head', kicker: 'so, the obvious question',
   lines: ['Start with *this* one'], size: 104, trans: 'fade',
   stamp: 'clone it, change the requirements, keep the shape'},
  {k: 'rows', kicker: 'because the shape is the transferable part', trans: 'rise',
   lines: ['Four things, and', 'they *travel*'], size: 60,
   rows: [{t: 'Be specific. Be simple'},
          {t: 'Say what done looks like'},
          {t: 'And say what not to build', hot: true}]},
  {k: 'flow', kicker: 'and the technique i rate most highly', trans: 'rise',
   lines: ['Write it by *being*', '*annoyed* at it'], size: 56,
   nodes: [{t: 'a minimal file', sub: 'ten minutes'},
           {t: 'run it', sub: 'and read what it built'},
           {t: 'delete all of it', sub: 'yes, all'},
           {t: 'rewrite the file', sub: 'for where it went wrong', tone: C.yellow}],
   loop: true,
   caption: 'two or three rounds of that and you have a genuinely good brief'},

  // --------------------------------------------------------------- the auto run
  {k: 'head', kicker: 'one last thing before we build',
   lines: ['The most *consequential*', 'setting in the product'], size: 58,
   trans: 'push'},
  {k: 'head', kicker: 'and every one of the four tools has a version of it',
   lines: ['Does it *ask* you', 'first?'], size: 88, trans: 'fade'},
  {k: 'ladder', kicker: 'in cursor it is called run mode', trans: 'rise',
   lines: ['Four choices, cautious', 'to *reckless*'], size: 52, items: RUNMODES,
   upto: 0},
  {k: 'ladder', kicker: 'one — allowlist', trans: 'none',
   lines: ['Nothing runs unless', 'you *approved* it'], size: 54,
   items: RUNMODES, upto: 1, pick: 0},
  {k: 'ladder', kicker: 'two — allowlist, with sandbox', trans: 'none',
   lines: ['Same rule, but *confined*'], size: 56, items: RUNMODES,
   upto: 2, pick: 1},
  {k: 'ladder', kicker: 'three — auto-review with sandbox, and the default',
   lines: ['Most things run.', 'The rest get *checked*'], size: 50, trans: 'none',
   items: RUNMODES, upto: 3, pick: 2},
  {k: 'ladder', kicker: 'four — run everything, unsandboxed', trans: 'none',
   lines: ['No approvals.', 'No *confinement*'], size: 58, items: RUNMODES,
   pick: 3, tone: C.red},
  {k: 'myth', kicker: 'and here is how i would actually decide', trans: 'rise',
   lines: ['Not a *courage* test'], size: 70,
   wrong: 'how brave are you feeling?',
   right: 'how much would you mind if this went wrong?',
   note: 'framing it as bravery is how people end up regretting the setting'},
  {k: 'rows', kicker: 'so, three cases', trans: 'rise',
   lines: ['It depends what is', 'in *reach*'], size: 62,
   rows: [{t: 'A throwaway folder with nothing valuable — unsandboxed is fine'},
          {t: 'A real repository with your work in it — stay sandboxed'},
          {t: 'Anything with credentials in reach — allowlist, and read them', hot: true}]},
  {k: 'head', kicker: 'and my advice for today, if you are unsure',
   lines: ['Leave it on the', '*default*'], size: 96, trans: 'push',
   stamp: 'watch one build, then move up once you have seen it behave'},
];

export const Nocode16: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode16/vo" files={FILES} gap={GAP} />
);
