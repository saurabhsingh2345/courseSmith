// nocode33 — parts 5 to 7: the database, the routes, and the wiring
//
// His half-hour rut did not happen to us and is not staged. The replacement is
// causal rather than anecdotal: L32's blank page came from a missing instruction
// about HOW to check, one sentence fixed it, and the agent left evidence of the
// new check sitting on the board. Then the thirty seconds nobody can outsource —
// drag, log out, hard reload, sign in, and the card is still where you put it.

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';

export const DURS = [
  11.932, 7.953, 6.173, 9.900, 7.450, 5.501, 6.609, 10.492,
  5.829, 9.571, 10.028, 8.703, 7.541, 7.025, 9.512, 6.500,
  9.402, 8.861, 9.339, 8.443, 9.960, 9.464, 8.773, 6.748,
  7.061, 10.352, 6.623, 8.622, 9.378, 9.342, 5.445, 10.170,
  9.890, 7.378, 11.376, 8.249, 7.503, 5.679, 8.939, 4.312,
  7.887, 6.589, 6.711, 6.997, 6.739, 8.905, 5.762, 9.632,
  5.613, 4.029, 6.439, 6.413, 7.377, 8.231, 8.391, 7.803,
  7.147, 5.692, 8.138, 10.280, 10.126, 10.171, 7.772, 11.729,
  4.705, 7.617, 10.708, 13.424, 7.045, 9.828, 8.906, 3.870,
  9.208, 10.332, 7.435, 7.296, 11.230, 11.068, 8.863, 7.780,
  7.153, 5.549, 7.403, 11.878, 10.769, 7.609, 9.850, 5.434,
  11.334, 6.116,
];
export const FILES = [
  'A1', 'A2', 'A3', 'A4', 'A5', 'A6', 'B1', 'B2',
  'B3', 'B4', 'B5', 'B6', 'B7', 'B8', 'B9', 'C1',
  'C2', 'C3', 'C4', 'C5', 'C6', 'C7', 'C8', 'D1',
  'D2', 'D3', 'D4', 'D5', 'D6', 'D7', 'D8', 'D9',
  'D10', 'D11', 'D12', 'D13', 'D14', 'D15', 'D16', 'E1',
  'E2', 'E3', 'E4', 'E5', 'E6', 'E7', 'F1', 'F2',
  'F3', 'F4', 'F5', 'G1', 'G2', 'G3', 'G4', 'G5',
  'G6', 'G7', 'G8', 'G9', 'G10', 'G11', 'G12', 'G13',
  'G14', 'G15', 'G16', 'G17', 'G18', 'G19', 'G20', 'Z1',
  'Z2', 'Z3', 'Z4', 'Z5', 'Z6', 'Z7', 'Z8', 'Z9',
  'Z10', 'Z11', 'Z12', 'Z13', 'Z14', 'Z15', 'Z16', 'Z17',
  'Z18', 'Z19',
].map((n) => `${n}.mp3`);
export const GAP = 0.240;
export const L33_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode33/shots/${n}.mp4`;

const SLIDES: Slide[] = [
  {k: 'head', kicker: 'part five is the only part that stops and asks', lines: ['And I want to explain', '*why*'], size: 62, trans: 'fade'},   // A1
  {k: 'rows', kicker: 'everything before this was reversible', lines: ['Throw it away.', 'Lose *twenty minutes*'], size: 62, rows: [{t: 'If part three had come out wrong, I would have asked again'}]},   // A2
  {k: 'myth', kicker: 'the database is different', lines: ['Not because it is *hard*'], size: 88, wrong: 'put the checkpoint before the difficult part', right: 'put it before the part everything else is shaped by', note: 'difficulty is not the variable'},   // A3
  {k: 'rows', kicker: 'because of what sits on top of it', lines: ['You do not lose *part five*'], size: 76, rows: [{t: 'Six, seven, eight, nine and ten all assume this shape'}, {t: 'Get it wrong and you lose everything built on it', hot: true}]},   // A4
  {k: 'head', kicker: 'so the question is not how difficult is this', lines: ['How much of what comes next', '*depends* on it'], size: 52, accent: '#f5c518'},   // A5
  {k: 'shot', src: sh('05'), kicker: 'and that is answerable', lines: ['Without knowing anything', 'about *databases*'], size: 58},   // A6  D 6.0-40.0s
  {k: 'head', kicker: 'now what it came back with', lines: ['No *jargon*.', 'I promise'], size: 88, trans: 'rise'},   // B1
  {k: 'myth', kicker: 'and it starts with one idea', lines: ['A table is a *spreadsheet*'], size: 76, wrong: 'databases are a specialist subject', right: 'rows are records, columns are fields, and that is the whole idea', note: 'the rest is bookkeeping about how spreadsheets refer to each other'},   // B2
  {k: 'rows', kicker: 'it designed four of them', lines: ['*Four* spreadsheets'], size: 104, numbered: true, rows: [{t: 'Users'}, {t: 'Boards'}, {t: 'Columns'}, {t: 'Cards'}]},   // B3
  {k: 'rows', kicker: 'users is one row per person', lines: ['One row.', 'Built for *thousands*'], size: 68, rows: [{t: 'The whole app has one login'}, {t: 'And the table is shaped as though it had many, deliberately', hot: true}]},   // B4
  {k: 'shot', src: sh('10'), kicker: 'boards carries the user it belongs to', lines: ['One *number* is the', 'whole relationship'], size: 56},   // B5  D 60.0-113.3s
  {k: 'shot', src: sh('11')},   // B6  D 113.3-159.6s
  {k: 'shot', src: sh('12')},   // B7  D 159.6-200.0s
  {k: 'head', kicker: 'and now the sentence that makes it all make sense', lines: ['Moving a card is', '*one number* changing'], size: 58, accent: '#f5c518'},   // B8
  {k: 'flow', kicker: 'that is genuinely it', lines: ['All that *dragging*'], size: 104, nodes: [{t: 'you drag a card', sub: 'across the screen'}, {t: 'its column changes', sub: 'one field, one row'}, {t: 'positions are tidied', sub: 'so the order still makes sense'}, {t: 'done', sub: 'and it is right tomorrow'}], caption: 'underneath, one field changing value'},   // B9
  {k: 'head', kicker: 'now the part I want to be honest about', lines: ['The *counterweight*'], size: 112, trans: 'push'},   // C1
  {k: 'myth', kicker: 'i had a design in my head before i asked', lines: ['One row. The *whole* board'], size: 68, wrong: 'store the entire board as one blob of text against the user', right: 'separate tables for columns and cards', note: 'mine would have worked. mine is also the hacky one.'},   // C2
  {k: 'head', kicker: 'and I knew it was the hacky one', lines: ['I would have done it', '*anyway*'], size: 68},   // C3
  {k: 'head', kicker: 'its version is better than mine', lines: ['Not slightly.', '*Properly*'], size: 104},   // C4
  {k: 'rows', kicker: 'and here is the test of whether I understand why', lines: ['I can *explain* it'], size: 100, rows: [{t: 'With my blob, finding one column means reading the whole board and picking through it'}, {t: 'With its version, that is a question you can just ask', hot: true}]},   // C5
  {k: 'rows', kicker: 'and the second reason, which is the bigger one', lines: ['A second board.', 'A second *user*'], size: 68, rows: [{t: 'With mine, unpick the format and rewrite everything that touches it'}, {t: 'With its version, it is another row', hot: true}]},   // C6
  {k: 'shot', src: sh('21'), kicker: 'so I approved its design over my own', lines: ['And *sit* with this'], size: 116},   // C7  D 240.0-270.8s
  {k: 'myth', kicker: 'both of these are true at once', lines: ['Which is *which*, today'], size: 92, wrong: 'it is better than me, or it is worse than me', right: 'it is better at some things and worse at others', note: 'the whole skill is telling them apart'},   // C8
  {k: 'head', kicker: 'i did not approve it silently', lines: ['Copy this *even when*', 'you are not sure'], size: 56},   // D1
  {k: 'shot', src: sh('24'), kicker: 'one thing looked awkward to me', lines: ['No two columns in the', '*same position*'], size: 60},   // D2  D 270.8-294.0s
  {k: 'myth', kicker: 'which sounds obviously correct', lines: ['And might *not* be'], size: 104, wrong: 'no two things may ever share a position — obviously', right: 'reordering usually has a moment where two briefly do', note: 'before everything settles'},   // D3
  {k: 'head', kicker: 'and here is the important part', lines: ['I do not *know*.', 'Genuinely'], size: 96},   // D4
  {k: 'rows', kicker: 'so I did not tell it to change it', lines: ['*Described* it. Asked.'], size: 88, numbered: true, rows: [{t: 'Here is what worries me'}, {t: 'Here is why'}, {t: 'If it causes that, change it. If not, leave it.', hot: true}]},   // D5
  {k: 'head', kicker: 'that shape is worth memorising', lines: ['*You* decide'], size: 140, accent: '#f5c518'},   // D6
  {k: 'rows', kicker: 'and it costs nothing either way', lines: ['One *sentence*'], size: 112, rows: [{t: 'If I am wrong, I have lost a sentence and learned something'}, {t: 'If I am right, I have caught it before a single row existed', hot: true}]},   // D7
  {k: 'shot', src: sh('30'), kicker: 'and here is what came back', lines: ['Better than I *deserved*'], size: 88},   // D8  I 710.0-735.6s
  {k: 'myth', kicker: 'yes — the reorder would have collided', lines: ['The database would', 'have *refused* it'], size: 58, wrong: 'two cards can never share a position, so nothing can go wrong', right: 'during a move between columns, two would have — briefly', note: 'and the whole move would have failed'},   // D9
  {k: 'rows', kicker: 'but it did not remove the rule, which is what I would have done', lines: ['Keep the rule.', 'Change the *move*'], size: 68, rows: [{t: 'Positions are tidied on the way through'}, {t: 'So the collision never happens at all', hot: true}]},   // D10
  {k: 'shot', src: sh('33')},   // D11  I 735.6-770.0s
  {k: 'rows', kicker: 'and look at what it cost me', lines: ['I could not have told you', 'the *fix*'], size: 58, rows: [{t: 'I did not know whether there was a problem'}, {t: 'I had a feeling one rule looked uncomfortable'}, {t: 'And I said so out loud', hot: true}]},   // D12
  {k: 'head', kicker: 'that is the entire skill', lines: ['Not expertise.', '*Refusing* to let it past'], size: 56},   // D13
  {k: 'shot', src: sh('36'), kicker: 'one more thing', lines: ['It finished, and it', '*stopped*'], size: 84},   // D14  I 960.0-1000.0s
  {k: 'head', kicker: 'because the plan told it to, four hours ago, in writing', lines: ['And it *remembered*'], size: 112},   // D15
  {k: 'myth', kicker: 'which is the return on writing a plan at all', lines: ['Built into the thing', 'being *followed*'], size: 58, wrong: 'I have to remember to stop it at the right moment', right: 'the stop is written down, so it happens without me', note: 'four hours later, unprompted'},   // D16
  {k: 'shot', src: sh('39'), kicker: 'part six', lines: ['The shortest explanation', 'of the *day*'], size: 62},   // E1  W 30.0-54.4s
  {k: 'rows', kicker: 'a route is a question it knows how to answer', lines: ['An address, and a *rule*'], size: 76, rows: [{t: 'What is on my board'}, {t: 'Move this card to that column'}, {t: 'Rename this column'}]},   // E2
  {k: 'shot', src: sh('41')},   // E3  W 54.4-91.1s
  {k: 'head', kicker: 'and that is genuinely all a back end is', lines: ['A list of questions,', 'and the *answers*'], size: 58},   // E4
  {k: 'shot', src: sh('43'), kicker: 'at the end of part six', lines: ['Nothing on screen', 'has *changed*'], size: 68},   // E5  W 91.1-130.0s
  {k: 'head', kicker: 'which is worth pausing on', lines: ['The first part today', 'with *no* visible result'], size: 54},   // E6
  {k: 'myth', kicker: 'and it is a trap for how you judge progress', lines: ['The screen is not', 'the *measure*'], size: 62, wrong: 'nothing changed, so nothing happened', right: 'every route exists and nobody is asking them yet', note: 'that is what the checklist in the plan is for'},   // E7
  {k: 'shot', src: sh('46'), kicker: 'part seven', lines: ['The *biggest* change', 'of the day'], size: 76},   // F1  W 340.0-377.9s
  {k: 'rows', kicker: 'up to now the board has been drawing itself', lines: ['Nothing was ever *saved*'], size: 76, rows: [{t: 'The cards are written into the front end'}, {t: 'The same five columns every time you reload', hot: true}]},   // F2
  {k: 'flow', kicker: 'after part seven', lines: ['It *asks*'], size: 140, nodes: [{t: 'the board loads', sub: 'by asking the back end'}, {t: 'you move a card', sub: 'and it says so'}, {t: 'the back end writes it', sub: 'to the database'}, {t: 'and it is still there', sub: 'tomorrow'}]},   // F3
  {k: 'head', kicker: 'which is the difference between', lines: ['A *picture* of an app,', 'and an app'], size: 58},   // F4
  {k: 'shot', src: sh('50')},   // F5  W 377.9-420.0s
  {k: 'shot', src: sh('51'), kicker: 'and it worked', lines: ['First time.', 'No *drama*'], size: 96},   // G1  W 880.0-925.0s
  {k: 'shot', src: sh('52')},   // G2  C 2.0-16.0s
  {k: 'shot', src: sh('53'), kicker: 'there is a card in the backlog', lines: ['I did not *make* that'], size: 96},   // G3  C 16.0-32.0s
  {k: 'head', kicker: 'it made it while checking its own work, and left it', lines: ['The most useful *litter*', 'I have seen all week'], size: 52},   // G4
  {k: 'myth', kicker: 'because last lecture it did not open a browser', lines: ['That check *passed*.', 'The page was blank'], size: 56, wrong: 'it checked and said it worked, so it lied', right: 'it ran the only check available from where it was standing', note: 'and that check could not see the fault'},   // G5
  {k: 'term', kicker: 'and one sentence changed between then and now', lines: ['*That* is the difference'], size: 88, title: 'added to this part\'s prompt', term: [{t: 'confirm it in a real browser,', kind: 'cmd'}, {t: 'not just in tests', kind: 'cmd'}, {t: '  and it did', kind: 'ok'}]},   // G6
  {k: 'head', kicker: 'and there are two lessons available here', lines: ['Only *one* of them', 'is true'], size: 64, trans: 'push'},   // G7
  {k: 'myth', kicker: 'the false one', lines: ['It was not being *lazy*'], size: 100, wrong: 'it cut corners and I told it off', right: 'it did exactly what it had been asked, both times', note: 'there was nothing to tell off'},   // G8
  {k: 'rows', kicker: 'the true one', lines: ['*How* to check', 'was missing'], size: 76, rows: [{t: 'So it chose, and it chose the cheapest check available'}, {t: 'Once the instruction named the check, it ran that one', hot: true}]},   // G9
  {k: 'head', kicker: 'so when it feels like carelessness', lines: ['Look at your *instruction*', 'before the tool'], size: 52, accent: '#f5c518'},   // G10
  {k: 'myth', kicker: 'and it has shown me evidence, and I am still going to look', lines: ['Not because I *distrust* it'], size: 68, wrong: 'I am checking because I do not believe it', right: 'I am checking because the check is cheap and being wrong is not', note: 'those are different reasons and only one of them scales'},   // G11
  {k: 'rows', kicker: 'here is the arithmetic', lines: ['Thirty seconds, against', '*three parts*'], size: 58, rows: [{t: 'Checking costs me thirty seconds'}, {t: 'Being wrong costs three more parts built on a board that does not save', hot: true}]},   // G12
  {k: 'shot', src: sh('63')},   // G13  C 34.0-50.0s
  {k: 'head', kicker: 'and then the part that actually proves it', lines: ['More than *reloading*', 'the page'], size: 62},   // G14
  {k: 'rows', kicker: 'because a reload proves almost nothing', lines: ['End the *session*'], size: 112, numbered: true, rows: [{t: 'Log out, so the session is gone'}, {t: 'Hard reload, so nothing comes from the browser memory'}]},   // G15
  {k: 'shot', src: sh('66')},   // G16  C 58.0-74.0s
  {k: 'shot', src: sh('67')},   // G17  C 84.0-112.0s
  {k: 'head', kicker: 'parts five, six and seven in one frame', lines: ['A session that did not', '*exist* when I moved it'], size: 54, accent: '#f5c518'},   // G18
  {k: 'myth', kicker: 'and the rule I would keep from today', lines: ['*Use* the thing'], size: 132, wrong: 'when it says it works, read its report more carefully', right: 'when it says it works, go and use the thing', note: 'the report is not the application'},   // G19
  {k: 'head', kicker: 'thirty seconds, with your own hands', lines: ['And it requires *no*', 'expertise whatsoever'], size: 54},   // G20
  {k: 'head', kicker: 'one last thing', lines: ['The bridge into', '*next week*'], size: 96, trans: 'rise'},   // Z1
  {k: 'head', kicker: 'this conversation has run since the first prompt this morning', lines: ['All of it, in *one* thread'], size: 62},   // Z2
  {k: 'myth', kicker: 'and there is a limit to what it holds at once', lines: ['The desk has an *edge*'], size: 96, wrong: 'it remembers everything you have said', right: 'there is a stack of paper on a desk, and the desk has an edge', note: 'we drew this on day two'},   // Z3
  {k: 'rows', kicker: 'and what happens at the edge is not an error', lines: ['*Quietly*'], size: 140, rows: [{t: 'It starts summarising the older parts'}, {t: 'And then it starts dropping them', hot: true}]},   // Z4
  {k: 'head', kicker: 'this tool does not show you a meter', lines: ['It is *handling* it', 'for us'], size: 68},   // Z5
  {k: 'head', kicker: 'next week you will see exactly how full it is', lines: ['And we are going to', '*obsess* over it'], size: 58},   // Z6
  {k: 'rows', kicker: 'because the symptom is not a crash', lines: ['It gets slightly *worse*'], size: 88, rows: [{t: 'It forgets a decision you made four hours ago'}, {t: 'It re-solves something that was already settled', hot: true}]},   // Z7
  {k: 'myth', kicker: 'and you will read that wrong', lines: ['Not a *bad day*'], size: 112, wrong: 'the model is having a bad day today', right: 'the beginning of the conversation has fallen off the desk', note: 'quietly, and without telling you'},   // Z8
  {k: 'head', kicker: 'so here is the practice', lines: ['Stop it. Start a', '*new* one'], size: 76},   // Z9
  {k: 'myth', kicker: 'and I will be honest that it feels wrong every time', lines: ['It is the *right* move'], size: 96, wrong: 'you are throwing away everything it knows', right: 'you are moving what matters somewhere it cannot be forgotten', note: 'it feels like starting again with a stranger'},   // Z10
  {k: 'head', kicker: 'but not before the one thing that makes it safe', lines: ['The whole technique,', 'in a *sentence*'], size: 54},   // Z11
  {k: 'term', kicker: 'and the second half is load-bearing', lines: ['*Including* the decisions'], size: 84, title: 'before you start the new conversation', term: [{t: 'please confirm plan.md is up to date', kind: 'cmd'}, {t: 'with all the latest, including any', kind: 'cmd'}, {t: 'design decisions that you made', kind: 'cmd'}, {t: '  the checklist already says WHAT got built', kind: 'out'}, {t: '  nothing says WHY', kind: 'warn'}]},   // Z12
  {k: 'rows', kicker: 'because think about what is not written down anywhere', lines: ['*Why* rows, not a blob'], size: 88, rows: [{t: 'That we chose it on purpose'}, {t: 'That I nearly did it differently', hot: true}]},   // Z13
  {k: 'head', kicker: 'if that only lives in the conversation', lines: ['It *dies* with', 'the conversation'], size: 62},   // Z14
  {k: 'myth', kicker: 'so the plan stops being a to-do list', lines: ['The project memory'], size: 100, wrong: 'the plan is a checklist of what to build', right: 'the plan is where the reasoning lives', note: 'in a file, outside the thing that forgets'},   // Z15
  {k: 'head', kicker: 'then it reads the plan, and knows where things stand', lines: ['Not because it remembers.', 'You left a *note*'], size: 52},   // Z16
  {k: 'head', kicker: 'and that is what this week has actually been teaching', lines: ['Hiding behind a', '*Kanban board*'], size: 62, trans: 'fade'},   // Z17
  {k: 'myth', kicker: 'not how to prompt', lines: ['The *hardest* part'], size: 124, wrong: 'the skill is knowing what to say to it', right: 'the skill is leaving things where the next one can pick them up', note: 'it was always the hardest part of working with people'},   // Z18
  {k: 'head', kicker: 'next time, the last three parts', lines: ['And a model that gets', 'to move your *cards*'], size: 54},   // Z19
];

export const Nocode33: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode33/vo" files={FILES} gap={GAP} />
);
