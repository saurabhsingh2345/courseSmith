// nocode10 — tools, the loop, and what an agent actually is (his L10, 8:28)
//
// His spine kept whole: trick three (tools), the grounding point that the model
// runs nothing, the square-root-of-pi demo, trick four (the loop), the three
// definitions of "agent" over time, and the callback proving yesterday's build
// was one.
//
// The definition history is the spine and it lands because it converges on our
// own footage — so the callback uses the real take of our agent looping in the
// editor, not a description of it.
//
// Ours are the X slides: how a model is told which tools exist and why a vague
// description makes it pick wrong (X1-X4), the stopping problem and why a human
// is still in the loop (X5-X8), what a loop actually costs and the ceiling that
// puts on it (X9-X12), and why the protocols are just plumbing (X13-X14). He
// describes the loop without ever saying how it ends or what it costs.

import React from 'react';
import {Deck, Slide, deckFrames, C} from './kit';

// same 0.12-style tight gap logic as l09: adam_deck bakes 0.22s lead and 0.60s
// tail into every clip, so 0.26s of gap still leaves ~1.08s of silence.
export const DURS = [
  14.941, 5.552, 14.281, 7.588, 13.574, 13.186, 10.634,
  12.904, 13.2, 14.818, 8.042,
  6.429, 13.402, 5.781, 10.67, 15.644,
  11.616, 14.146,
  5.121, 11.29, 10.88,
  6.769, 12.973, 13.902, 15.739, 12.121, 16.398, 14.38, 9.448,
  19.542, 5.534, 12.961, 15.073, 15.518, 9.526, 11.295, 11.996,
  9.17, 12.866, 13.279, 11.224, 13.864,
];
export const FILES = [
  '01', '02', '03', '04', '05', '06', '07',
  'X1', 'X2', 'X3', 'X4',
  '08', '09', '10', '11', '12',
  'X13', 'X14',
  '13', '14', '15',
  'X5', 'X6', 'X7', 'X8', 'X9', 'X10', 'X11', 'X12',
  '16', '17', '18', '19', '20', '21', '22', '23', '24', '25', '26', '27', '28',
].map((n) => `${n}.mp3`);
export const GAP = 0.26;
export const L10_FRAMES = deckFrames(DURS, GAP);

const T = 'rise' as const;
const DEFS = [
  {when: 'the openai era', t: 'Goes off and works *independently*',
   sub: 'You give it a task, it disappears, it comes back'},
  {when: 'early last year', t: 'The model controls the *workflow*',
   sub: 'Not a program that calls a model — a program the model steers'},
  {when: 'and the one that stuck', t: 'Runs *tools* in a *loop* for a *goal*',
   sub: 'Fourteen words, and every one of them is something you now know',
   hot: true},
];

const SLIDES: Slide[] = [
  {k: 'rows', kicker: 'two tricks down, two to go',
   lines: ['These two are', 'the *bridge*'], size: 68, trans: 'fade',
   rows: [{t: 'So far: something that produces text and remembers nothing'},
          {t: 'Yesterday: something that built you a game'},
          {t: 'Those do not obviously connect', hot: true}]},
  {k: 'head', kicker: 'trick three', lines: ['*Tools*'], size: 140, trans: 'push'},
  {k: 'rows', kicker: 'the tokens do not have to be an answer',
   lines: ['They can be a', '*request*'], size: 76, trans: T,
   rows: [{t: 'Search the web. Run this calculation'},
          {t: 'Read that file. Write this one'},
          {t: 'In a format agreed beforehand', hot: true}]},
  {k: 'head', kicker: 'and now the sentence to hold on to',
   lines: ['The most', '*misunderstood*', 'idea in the field'], size: 62, trans: 'push'},
  {k: 'myth', kicker: 'it is worth being blunt about this', trans: T,
   lines: ['It runs *nothing*'], size: 74,
   wrong: 'The model went and searched the web',
   right: 'It emitted *tokens*. That is all it can do.',
   note: 'It does not search. It does not run. It does not read one of your files.'},
  {k: 'rows', kicker: 'what actually happens',
   lines: ['*Our code* does', 'the doing'], size: 66, trans: T,
   rows: [{t: 'It reads the tokens and notices they are a request'},
          {t: 'It goes and does the thing itself — an ordinary function call'},
          {t: 'Then hands the result back as more input', hot: true}]},
  {k: 'flow', kicker: 'so the shape is this', lines: ['Out, read, *run*,', 'back in'],
   size: 56, trans: T,
   nodes: [{t: 'Model', sub: 'emits tokens', tone: '#f5c518'},
           {t: 'Our code', sub: 'reads them', tone: '#3b82f6'},
           {t: 'The tool', sub: 'actually runs'},
           {t: 'Back in', sub: 'as more input'}],
   loop: true, caption: 'and around it goes'},

  {k: 'head', kicker: 'but how does it know which tools exist?',
   lines: ['It is *told*.', 'In text.'], size: 82, trans: 'push'},
  {k: 'term', kicker: 'the application writes a list, and it goes in like everything else',
   lines: ['The *menu*'], size: 84, trans: T, title: 'what the model is handed',
   term: [{t: 'read_file — reads a file. args: path', kind: 'out'},
          {t: 'write_file — writes a file. args: path, contents', kind: 'out'},
          {t: 'run — runs a shell command. args: cmd', kind: 'out'},
          {t: 'search — searches the web. args: query', kind: 'out'}]},
  {k: 'rows', kicker: 'which explains a failure you will meet',
   lines: ['A bad menu, bad', '*choices*'], size: 66, trans: T,
   rows: [{t: 'A vague description and it picks the wrong tool'},
          {t: 'Or invents arguments that merely sound right'},
          {t: 'Not carelessness — it is choosing from a menu you wrote', hot: true}]},
  {k: 'rows', kicker: 'and it is why tools matter as much as models',
   lines: ['Same engine,', 'better *menu*'], size: 66, trans: T,
   rows: [{t: 'Dramatically better behaviour, same model underneath', hot: true}]},

  {k: 'head', kicker: 'let me show you it with nothing clever involved',
   lines: ['You can do this', 'in a *minute*'], size: 66, trans: 'push'},
  // the demo is real: a logged-out ChatGPT session, recorded page-only and
  // mounted in our own browser frame. The account avatar is blurred.
  {k: 'browser', kicker: 'first you tell it the rules',
   lines: ['A protocol, invented', 'on the *spot*'], size: 50, trans: T,
   src: 'nocode10/shots/gpt_rule.mp4', url: 'chatgpt.com', tab: 'ChatGPT',
   aspect: '1920 / 992', scale: 0.82},
  {k: 'browser', kicker: 'then ask it something it cannot do in its head',
   lines: ['"The square root', 'of *pi*"'], size: 50, trans: T,
   src: 'nocode10/shots/gpt_ask.mp4', url: 'chatgpt.com', tab: 'ChatGPT',
   aspect: '1920 / 992', scale: 0.82},
  {k: 'browser', kicker: 'and it does not answer',
   lines: ['It asks for a *tool*'], size: 50, trans: T,
   src: 'nocode10/shots/gpt_answer.mp4', url: 'chatgpt.com', tab: 'ChatGPT',
   aspect: '1920 / 992', scale: 0.82},
  {k: 'rows', kicker: 'and that is a tool call. all of it.',
   lines: ['Clever input, and', 'code that *reads* it'], size: 54, trans: T,
   rows: [{t: 'The model asked. It did not run anything'},
          {t: 'You ran the line and pasted the number back'},
          {t: 'Everything built on top is that, with better plumbing', hot: true}]},

  {k: 'rows', kicker: 'one last thing on the plumbing',
   lines: ['The *acronyms*'], size: 90, trans: 'push',
   rows: [{t: 'There are now standards for describing tools to a model'},
          {t: 'And for connecting it to outside systems'},
          {t: 'They are genuinely useful', hot: true}]},
  {k: 'rows', kicker: 'but every one of them is what you just watched',
   lines: ['The rest is', '*convenience*'], size: 68, trans: T,
   rows: [{t: 'A description of what is available'},
          {t: 'And code that reads the tokens and does the work'},
          {t: 'Understand the one-line version and you understand all of them',
           hot: true}]},

  {k: 'head', kicker: 'trick four', lines: ['The *loop*'], size: 140, trans: 'push'},
  {k: 'head', kicker: 'and it is almost embarrassingly simple',
   lines: ['What beats calling', 'it *once*?'], size: 62, trans: T,
   stamp: 'calling it again'},
  {k: 'flow', kicker: 'produce, check, repeat', lines: ['Until the goal', 'is *met*'],
   size: 62, trans: T,
   nodes: [{t: 'Produce', sub: 'do the next thing', tone: '#f5c518'},
           {t: 'Check', sub: 'is the goal met?'},
           {t: 'Not yet', sub: 'feed it all back'}],
   loop: true,
   caption: 'the difference between something that answers and something that finishes'},

  {k: 'head', kicker: 'but there is a question buried in it',
   lines: ['How does it know', 'it is *finished*?'], size: 60, trans: 'push'},
  {k: 'rows', kicker: 'the answer is that it decides',
   lines: ['With *tokens*,', 'like everything else'], size: 56, trans: T,
   rows: [{t: 'On every pass it is asked whether the goal is met'},
          {t: 'So the loop’s judgement is exactly the model’s judgement'},
          {t: 'And no better', hot: true}]},
  {k: 'rows', kicker: 'so a loop goes wrong in exactly two ways',
   lines: ['Too *early*,', 'or *never*'], size: 74, trans: T,
   rows: [{t: 'It stops, believing it is done when it is not'},
          {t: 'Or it will not stop, finding one more thing every time'},
          {t: 'Burning time and money on work nobody asked for', hot: true}]},
  {k: 'rows', kicker: 'which is the honest reason you are still here',
   lines: ['You are the *check*'], size: 76, trans: T,
   rows: [{t: 'The loop does not have one it can trust'},
          {t: 'Yesterday’s approval prompts were exactly this'},
          {t: 'A stopping point somebody built in on purpose', hot: true}]},

  {k: 'head', kicker: 'and there is a cost consequence',
   lines: ['Every pass re-sends', '*everything*'], size: 58, trans: 'push'},
  {k: 'occupancy', kicker: 'instructions, tools, files, and the whole history so far',
   lines: ['Pass *twenty*'], size: 84, trans: T,
   used: 86, legend: 're-read on this pass',
   caption: 'twenty passes is not twenty short calls — the last one re-reads all nineteen'},
  {k: 'gauge', kicker: 'and that is the ceiling on all of it',
   lines: ['A *race*'], size: 104, trans: T,
   pct: 112, limit: 74, label: 'what the loop has accumulated', limitLabel: 'the window',
   caption: 'between finishing the job and filling up'},
  {k: 'rows', kicker: 'which is why the whole of tomorrow is context',
   lines: ['The resource the', 'loop *spends*'], size: 62, trans: T,
   rows: [{t: 'Not because context is an advanced topic'},
          {t: 'Because it is what the loop is actually spending', hot: true}]},

  {k: 'rows', kicker: 'and now the four fit together',
   lines: ['2022, to *yesterday*'], size: 72, trans: 'push', numbered: true,
   rows: [{t: 'It predicts text'},
          {t: 'It is handed everything, so it appears to remember'},
          {t: 'It can ask for things, and our code does them'},
          {t: 'And it goes round until the goal is met', hot: true}]},

  {k: 'head', kicker: 'which brings us to a harder question',
   lines: ['What *is* an', 'AI agent?'], size: 86, trans: 'push'},
  {k: 'rows', kicker: 'for about two years',
   lines: ['A *marketing* word'], size: 82, trans: T,
   rows: [{t: 'It meant whatever the person saying it wanted it to mean'},
          {t: 'But it has actually converged — and the path is worth two minutes',
           hot: true}]},
  {k: 'timeline', kicker: 'first, the early one', lines: ['*Independently*'], size: 74,
   trans: T, items: DEFS, upto: 1},
  {k: 'timeline', kicker: 'then, from early last year', lines: ['It steers the *workflow*'],
   size: 62, trans: 'none', items: DEFS, upto: 2},
  {k: 'timeline', kicker: 'and the one that stuck', lines: ['Tools. Loop. *Goal*.'],
   size: 66, trans: 'none', items: DEFS, upto: 3},
  {k: 'rows', kicker: 'and look what it is made of',
   lines: ['The definition that', 'won is *today*'], size: 58, trans: 'push', numbered: true,
   rows: [{t: 'Tools — trick three'},
          {t: 'A loop — trick four'},
          {t: 'A goal — which is what you typed', hot: true}]},

  {k: 'head', kicker: 'so let us test it against something you watched',
   lines: ['One badly spelled', '*sentence*'], size: 62, trans: 'push',
   stamp: 'that was the goal'},
  {k: 'shot', src: 'nocode10/shots/loop.mp4', kicker: 'was it in a loop?',
   lines: ['Step after *step*'], size: 72},
  {k: 'shot', src: 'nocode10/shots/tools.mp4', kicker: 'did it use tools?',
   lines: ['It wrote to your *disk*'], size: 62},
  {k: 'rows', kicker: 'so by the definition the industry took two years to agree on',
   lines: ['You used one', '*yesterday*'], size: 68, trans: T,
   rows: [{t: 'Goal, loop, tools — all three, on your own screen'},
          {t: 'You used an agent before anybody defined it'},
          {t: 'Which is the right order to learn it in', hot: true}]},
  {k: 'rows', kicker: 'one honest caveat, and it will save you an argument',
   lines: ['This will move', '*again*'], size: 74, trans: T,
   rows: [{t: 'It has moved three times in three years'},
          {t: 'There is no reason to think it has stopped', hot: true}]},
  {k: 'rows', kicker: 'so hold one loosely and one tightly',
   lines: ['The *mechanism*', 'will not move'], size: 62, trans: 'push',
   rows: [{t: 'Tools and a loop are what is happening under there'},
          {t: 'Whatever the word ends up meaning'},
          {t: 'Tomorrow: context, and a file called agents.md', hot: true}]},
];

export const Nocode10: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode10/vo" files={FILES} gap={GAP} />
);
