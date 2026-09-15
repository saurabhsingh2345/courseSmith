// nocode31 — planning and scaffolding (his L31, 11:43)
//
// Two of his set pieces did not happen to us, and the narration says so rather
// than staging them — the same call L18 made when our Copilot build succeeded
// where his broke:
//   * his agent asked three questions off the opening prompt; ours asked none,
//     because the brief was clearer. C1-C6 make that the honest lesson: the
//     point was to find out whether the brief was clear, and silence is a pass.
//   * his agent silently swapped `uv` for a bare requirements.txt; ours produced
//     pyproject.toml correctly. D10-D12 keep the transferable shape — you find
//     out about ignored instructions from what APPEARS in the project.

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';

export const DURS = [
  11.806, 10.119, 13.609, 10.818, 11.928, 10.071, 8.930, 15.888,
  6.412, 13.023, 12.276, 7.469, 17.941, 8.019, 7.332, 12.459,
  8.482, 12.548, 11.937, 10.421, 12.982, 11.262, 9.883, 8.903,
  14.599, 11.849, 9.905, 10.448, 12.680, 11.590, 10.392, 5.603,
  9.789, 11.288, 7.166, 13.712, 11.329, 12.150, 16.745, 8.854,
  14.669, 9.992, 11.877, 14.071, 10.655, 14.168, 10.960, 6.302,
  12.358, 15.283, 13.685, 7.712, 11.045,
];
export const FILES = [
  'A1', 'A2', 'A3', 'A4', 'A5', 'A6', 'A7', 'A8',
  'A9', 'A10', 'A11', 'A12', 'A13', 'B1', 'B2', 'B3',
  'B4', 'B5', 'C1', 'C2', 'C3', 'C4', 'C5', 'C6',
  'C7', 'D1', 'D2', 'D3', 'D4', 'D5', 'B6', 'E7',
  'E8', 'E9', 'E10', 'E11', 'E12', 'D6', 'D7', 'D8',
  'D9', 'E1', 'E2', 'E3', 'E4', 'E5', 'E6', 'D10',
  'D11', 'D12', 'B7', 'D13', 'B8',
].map((n) => `${n}.mp3`);
export const GAP = 0.240;
export const L31_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode31/shots/${n}.mp4`;

const SLIDES: Slide[] = [
  {k: 'head', kicker: 'yesterday we broke every rule', lines: ['Today we *follow* them'], size: 96, trans: 'fade'},   // A1
  {k: 'rows', kicker: 'ten parts, in a file, written by me', lines: ['Not written by the *AI*'], size: 76, numbered: true, rows: [{t: 'Plan'}, {t: 'Scaffolding'}, {t: 'Serve the front end'}, {t: 'Sign in'}, {t: 'Database model'}, {t: 'Back end routes'}, {t: 'Wire them together'}, {t: 'Reach the AI'}, {t: 'Extend the plumbing'}, {t: 'The assistant', hot: true}]},   // A2
  {k: 'myth', kicker: 'and that was a choice', lines: ['Whose *judgement*?'], size: 88, wrong: 'ask the agent to write the plan — very common, and it will be reasonable', right: 'a plan is a set of decisions, and I have opinions about these', note: 'if you have opinions, write your own'},   // A3
  {k: 'head', kicker: 'and if you genuinely do not', lines: ['Ask for a draft.', 'Then *argue* with it'], size: 68},   // A4
  {k: 'shot', src: sh('04'), kicker: 'part one sounds like a joke', lines: ['*Planning* the plan'], size: 108},   // A5  Q 8.0-40.1s
  {k: 'rows', kicker: 'part two', lines: ['One page.', 'One *API call*'], size: 84, rows: [{t: 'The container'}, {t: 'The back end'}, {t: 'The start and stop scripts'}, {t: 'Nothing clever. A hello world that runs.', hot: true}]},   // A6
  {k: 'rows', kicker: 'parts three to five', lines: ['Serve. Sign in.', '*Design* the data'], size: 62, rows: [{t: 'Three — serve the board we inherited'}, {t: 'Four — the sign-in'}, {t: 'Five — the database design, and STOP for approval', hot: true}]},   // A7
  {k: 'rows', kicker: 'and six to ten', lines: ['Routes. Wire.', 'Then the *AI*'], size: 62, rows: [{t: 'Six — routes that read and change the board'}, {t: 'Seven — wire the halves so a move persists'}, {t: 'Eight — prove we can reach the AI at all'}, {t: 'Nine and ten — the board, a question, and the assistant', hot: true}]},   // A8
  {k: 'head', kicker: 'now the reason those particular ten', lines: ['The *sizing*', 'is the lesson'], size: 72, trans: 'rise'},   // A9
  {k: 'myth', kicker: 'and it is one rule', lines: ['Small enough to *dig into*'], size: 68, wrong: 'break it into the smallest possible pieces', right: 'each step small enough that if it fails, you know how to start looking', note: 'not trivial. not so big that failure leaves you staring.'},   // A10
  {k: 'rows', kicker: 'and it scales with you, not with the software', lines: ['The right step size', 'is a fact about *you*'], size: 54, rows: [{t: 'Know less — use smaller steps'}, {t: 'Know more — use bigger ones'}, {t: 'Nobody else can pick it for you', hot: true}]},   // A11
  {k: 'head', kicker: 'which answers the obvious objection', lines: ['Why not just tell it', 'to do *all ten*'], size: 62},   // A12
  {k: 'myth', kicker: 'because yesterday was a toy and this has a database', lines: ['It will go off *somewhere*'], size: 68, wrong: 'give it everything and let it run — it worked yesterday', right: 'ten steps at this size goes off the rails in the middle', note: 'and you will not know where, because you were not watching'},   // A13
  {k: 'head', kicker: 'before the first prompt, one technique', lines: ['*One sentence*'], size: 124, trans: 'push'},   // B1
  {k: 'shot', src: sh('14'), kicker: 'and here it is', lines: ['Let me know if you have questions.', 'Do *no work* yet'], size: 50},   // B2  Q 40.1-60.0s
  {k: 'rows', kicker: 'because of when it is cheapest to be wrong', lines: ['Questions cost *seconds*'], size: 76, rows: [{t: 'An agent about to misunderstand you reveals it in its questions'}, {t: 'The questions cost seconds'}, {t: 'The misunderstanding costs an hour', hot: true}]},   // B3
  {k: 'head', kicker: 'and nothing has been written yet', lines: ['You are still holding', 'the *steering wheel*'], size: 58},   // B4
  {k: 'myth', kicker: 'the other half comes later', lines: ['*Confident*, not done'], size: 84, wrong: 'is it done?', right: 'let me know when you are confident', note: 'done is a claim about the work. confident is a claim about itself.'},   // B5
  {k: 'shot', src: sh('18')},   // C1  Q 60.0-220.0s
  {k: 'head', kicker: 'which is not what I expected', lines: ['It asked *nothing*'], size: 120},   // C2
  {k: 'myth', kicker: 'and that is the technique working', lines: ['Silence is a *result*'], size: 88, wrong: 'no questions means it did not engage', right: 'no questions means the brief was clear', note: 'the point was never to collect questions'},   // C3
  {k: 'head', kicker: 'and it cost fifteen seconds to find out', lines: ['The cheapest information', 'you will buy *all day*'], size: 54},   // C4
  {k: 'rows', kicker: 'and it would have been just as useful the other way', lines: ['Three confused questions'], size: 76, rows: [{t: 'Would have told me the brief was vague'}, {t: 'And I would rather learn that now'}, {t: 'Than after it has built the wrong thing twice', hot: true}]},   // C5
  {k: 'head', kicker: 'so keep the habit even when it returns nothing', lines: ['You are testing whether', 'you were *clear*'], size: 54},   // C6
  {k: 'rows', kicker: 'and notice the small grey line under its answer', lines: ['It *names* the model'], size: 80, rows: [{t: 'And what it charged you'}, {t: 'This tool picks a model per task'}, {t: 'So that line will not always say the same thing'}, {t: 'Read it rather than assuming', hot: true}]},   // C7
  {k: 'shot', src: sh('25')},   // D1  P 10.0-127.7s
  {k: 'shot', src: sh('26')},   // D2  P 127.7-226.5s
  {k: 'head', kicker: 'and this is the moment with the most leverage today', lines: ['Right now the plan', 'is *words*'], size: 68, accent: '#f5c518'},   // D3
  {k: 'rows', kicker: 'so read it properly — the shape, not every sub-step', lines: ['What to *look* for'], size: 80, rows: [{t: 'Are the parts in a sensible order'}, {t: 'Does anything assume something that does not exist yet'}, {t: 'Is there a success criterion you could not check yourself', hot: true}]},   // D4
  {k: 'myth', kicker: 'and that last one is the real test', lines: ['A wish, or a *criterion*'], size: 68, wrong: 'the board works correctly', right: 'log out, log back in, and the card is still in the second column', note: 'if you cannot verify it yourself, you will be taking its word later'},   // D5
  {k: 'shot', src: sh('30'), kicker: 'and the most boring habit in software', lines: ['*Check in*', 'after every part'], size: 84},   // B6  P 226.5-330.0s
  {k: 'head', kicker: 'and I am naming the three commands', lines: ['Look. Include. *Save*'], size: 92, trans: 'rise'},   // E7
  {k: 'term', kicker: 'because nobody ever does', lines: ['The *three* commands'], size: 76, title: 'after every part', term: [{t: 'git status', kind: 'cmd'}, {t: '  what changed since the last save point', kind: 'out'}, {t: 'git add .', kind: 'cmd'}, {t: '  include all of it', kind: 'out'}, {t: 'git commit -m "part 2 complete"', kind: 'cmd'}, {t: '  the save point exists', kind: 'ok'}]},   // E8
  {k: 'rows', kicker: 'and the message matters more than people think', lines: ['You will *read* it later'], size: 76, rows: [{t: 'When you are hunting for the last version that worked'}, {t: '"fixes" tells you nothing'}, {t: '"part 7 built, some drag and drop bugs" tells you everything', hot: true}]},   // E9
  {k: 'head', kicker: 'and this is saved on your own machine', lines: ['It has not gone *anywhere*'], size: 76},   // E10
  {k: 'rows', kicker: 'and if any of that is unfamiliar', lines: ['*Ask* the agent'], size: 104, rows: [{t: 'Ask it what git is'}, {t: 'Ask it what a commit is'}, {t: 'Ask it to explain what it just did'}]},   // E11
  {k: 'head', kicker: 'which is the general point of this whole course', lines: ['The thing you direct is', 'also your best *teacher*'], size: 52},   // E12
  {k: 'shot', src: sh('37')},   // D6  S 10.0-114.1s
  {k: 'rows', kicker: 'because part two is where you find out', lines: ['Wrong *now*, with seven files'], size: 62, rows: [{t: 'Does the container build'}, {t: 'Do the scripts run on your machine'}, {t: 'Is the thing in the browser the thing in the box'}, {t: 'Better to know now than later with seventy files', hot: true}]},   // D7
  {k: 'myth', kicker: 'and here is the prompt to steal', lines: ['Not *is it done*'], size: 96, wrong: 'is it done?', right: 'tell me how I can test this myself, and let me know when you are confident', note: 'it hands you the verification instead of keeping it'},   // D8
  {k: 'shot', src: sh('40')},   // D9  S 200.0-283.3s
  {k: 'shot', src: sh('41')},   // E1  S 114.1-200.0s
  {k: 'rows', kicker: 'if you can read the command, read it', lines: ['Most are *ordinary*'], size: 88, rows: [{t: 'Make a folder'}, {t: 'Install a package'}, {t: 'Start a server'}, {t: 'You get a feel for the normal ones within an hour', hot: true}]},   // E2
  {k: 'rows', kicker: 'and if you cannot read it, three options', lines: ['None of them are *guessing*'], size: 68, numbered: true, rows: [{t: 'Ask it what the command does and why it needs to run'}, {t: 'Paste it into a different AI and ask that one'}, {t: 'Say no, and ask it to explain itself first', hot: true}]},   // E3
  {k: 'head', kicker: 'and that third one is underrated', lines: ['Refusing is the most', '*normal* thing a manager does'], size: 52},   // E4
  {k: 'myth', kicker: 'so let me reframe this for anyone it stresses', lines: ['Shown, in *advance*'], size: 88, wrong: 'you are being asked to audit software you cannot read', right: 'you are being shown everything before it happens, by something asking permission', note: 'better than most developers had ten years ago'},   // E5
  {k: 'head', kicker: 'it is a learning opportunity', lines: ['Wearing the costume', 'of an *interruption*'], size: 58},   // E6
  {k: 'head', kicker: 'one thing to watch for that did NOT happen to us', lines: ['And I am flagging it', '*anyway*'], size: 62, trans: 'push'},   // D10
  {k: 'myth', kicker: 'our brief named a package manager, and it used it', lines: ['His agent *quietly* swapped it'], size: 58, wrong: 'pyproject.toml, as the brief specified', right: 'a bare requirements.txt appeared, which nobody asked for', note: 'he only noticed because a FILE appeared that should not exist'},   // D11
  {k: 'rows', kicker: 'so carry the shape even though it went right here', lines: ['Look at what *shows up*'], size: 76, rows: [{t: 'Instructions do not always get followed'}, {t: 'And it will not tell you when one was not'}, {t: 'You find out from what appears in the project', hot: true}]},   // D12
  {k: 'head', kicker: 'so, checkpoints', lines: ['Not because *this* part', 'was risky'], size: 68},   // B7
  {k: 'head', kicker: 'but because the habit has to be automatic', lines: ['*Before* you reach', 'the part that is'], size: 64},   // D13
  {k: 'shot', src: sh('52'), kicker: 'now watch what happens next', lines: ['Two things go *wrong*'], size: 100},   // B8  S 283.3-345.0s
];

export const Nocode31: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode31/vo" files={FILES} gap={GAP} />
);
