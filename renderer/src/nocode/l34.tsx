// nocode34 — the assistant, and the week one wrap
//
// The payoff nearly went out false. The API key never reached the container and
// the card move was decided by a regular expression, so the only thing that
// worked was the only thing that never needed the model. P6-P14 are finding
// that; both were fixed on camera before the demo was shot. The demo names the
// card indirectly on purpose — a pattern cannot resolve "the card about the
// flaky checkout tests", and a model reading the board can.

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';

export const DURS = [
  6.705, 8.631, 5.199, 5.920, 9.312, 10.286, 5.523, 8.624,
  2.763, 11.365, 8.796, 10.761, 7.662, 9.406, 3.432, 4.777,
  9.030, 9.880, 7.953, 4.889, 8.060, 11.436, 7.084, 7.379,
  8.175, 8.982, 11.182, 11.178, 7.951, 9.298, 11.133, 10.455,
  3.652, 7.647, 8.278, 6.644, 12.061, 11.945, 11.864, 9.178,
  9.380, 5.468, 9.768, 11.405, 4.757, 3.347, 8.468, 8.362,
  10.216, 10.993, 12.887, 10.577, 7.617, 4.781, 8.868, 12.395,
  10.181, 6.346, 6.822, 4.311, 5.476, 5.724, 4.956, 5.439,
  4.853, 8.588, 4.818, 12.357, 9.243, 9.100, 8.149, 4.240,
  8.052, 9.756, 9.527, 10.264, 10.085, 8.273, 4.009,
];
export const FILES = [
  'A1', 'A2', 'A3', 'A4', 'A5', 'A6', 'A7', 'A8',
  'B1', 'B2', 'B3', 'B4', 'B5', 'B6', 'C1', 'C2',
  'C3', 'C4', 'P1', 'P2', 'P3', 'P4', 'P5', 'P6',
  'P7', 'P8', 'P9', 'P10', 'P11', 'P12', 'P13', 'P14',
  'P15', 'P16', 'P17', 'P18', 'P19', 'P20', 'P21', 'P22',
  'P23', 'H1', 'H2', 'H3', 'H4', 'H5', 'H6', 'H7',
  'H8', 'H9', 'H10', 'H11', 'H12', 'H13', 'H14', 'H15',
  'H16', 'W1', 'W2', 'W3', 'W4', 'W5', 'W6', 'W7',
  'W8', 'W9', 'W10', 'W11', 'W12', 'W13', 'W14', 'W15',
  'W16', 'W17', 'W18', 'W19', 'W20', 'W21', 'W22',
].map((n) => `${n}.mp3`);
export const GAP = 0.240;
export const L34_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode34/shots/${n}.mp4`;

const SLIDES: Slide[] = [
  {k: 'head', kicker: 'last three parts', lines: ['And why they are *three*', 'and not one'], size: 62, trans: 'fade'},   // A1
  {k: 'shot', src: sh('01'), kicker: 'part eight does nothing you can see', lines: ['One call. One *answer*'], size: 96},   // A2  E 6.0-60.0s
  {k: 'rows', kicker: 'that is the whole part', lines: ['No interface.', 'No *cleverness*'], size: 76, rows: [{t: 'No board involved at all'}]},   // A3
  {k: 'head', kicker: 'and it is the most valuable of the three', lines: ['It separates *two* questions'], size: 72},   // A4
  {k: 'rows', kicker: 'that people spend hours confusing', lines: ['*Reach* it, or', 'is it any *good*'], size: 62, numbered: true, rows: [{t: 'Is the key right, is the name right, is the account funded, does the network allow it'}, {t: 'And separately — is our feature any good', hot: true}]},   // A5
  {k: 'myth', kicker: 'because if you build it all at once', lines: ['They look *identical*', 'from outside'], size: 58, wrong: 'it is not working, so something in my feature is wrong', right: 'you are debugging two questions at the same time', note: 'and neither of them will tell you which one it is'},   // A6
  {k: 'shot', src: sh('06'), kicker: 'so part eight answers the first one alone', lines: ['Two minutes, and never', 'in *question* again'], size: 54},   // A7  E 210.0-238.0s
  {k: 'head', kicker: 'and steal that habit well beyond AI', lines: ['Prove you can *reach* it', 'before you build on it'], size: 52, accent: '#f5c518'},   // A8
  {k: 'shot', src: sh('08'), kicker: 'part nine', lines: ['Gives it the *board*'], size: 124},   // B1  B 50.0-67.2s
  {k: 'myth', kicker: 'because on its own it knows nothing about you', lines: ['It has never *seen*', 'your cards'], size: 62, wrong: 'ask it what is in progress and it will tell you', right: 'it will produce something plausible and completely invented', note: 'and it will sound entirely confident doing it'},   // B2
  {k: 'flow', kicker: 'so before every question we send', lines: ['We *attach* the board'], size: 96, nodes: [{t: 'here are the columns', sub: 'and what is in them'}, {t: 'here are the cards', sub: 'titles and descriptions'}, {t: 'now, the user asks this', sub: 'your sentence, at the end'}, {t: 'and it answers', sub: 'from what it was handed'}]},   // B3
  {k: 'shot', src: sh('11'), kicker: 'and that is the entire trick', lines: ['Behind *every* AI feature', 'that seemed to know things'], size: 50},   // B4  B 67.2-130.0s
  {k: 'myth', kicker: 'there is no memory, and no learning', lines: ['A *briefing note*'], size: 124, wrong: 'it learned about my project', right: 'a very good writer was handed a note a fraction of a second before answering', note: 'and the note is thrown away afterwards'},   // B5
  {k: 'rows', kicker: 'and once you see it that way', lines: ['Failures become', '*predictable*'], size: 76, rows: [{t: 'A lot of AI products stop being mysterious'}, {t: 'It did not know because nobody told it', hot: true}]},   // B6
  {k: 'shot', src: sh('14'), kicker: 'and part ten', lines: ['Turns talking', 'into *doing*'], size: 96},   // C1  B 736.0-752.0s
  {k: 'head', kicker: 'up to part nine it can describe your board', lines: ['After part ten', 'it can *change* it'], size: 62},   // C2
  {k: 'head', kicker: 'which is worth being careful about', lines: ['Something that makes things up', 'gets to touch the *data*'], size: 48, accent: '#f5c518'},   // C3
  {k: 'myth', kicker: 'and this is exactly the case I flagged with the diffs', lines: ['*This* is that'], size: 132, wrong: 'accept everything in bulk, it has been fine all day', right: 'slow down when it touches something you would not want wrong', note: 'the first time we plug in the AI — I said it then, and here it is'},   // C4
  {k: 'shot', src: sh('18')},   // P1  D 2.0-14.0s
  {k: 'shot', src: sh('19')},   // P2  D 14.0-26.0s
  {k: 'head', kicker: 'except it does not', lines: ['It returns an *error*'], size: 116},   // P3
  {k: 'shot', src: sh('21'), kicker: 'and the reason is almost funny', lines: ['The key is on *my* machine'], size: 84},   // P4  R 20.0-80.0s
  {k: 'myth', kicker: 'and the application does not run on my machine', lines: ['Nobody handed the key', 'to the *box*'], size: 58, wrong: 'the key is in a file in the project, so the app has it', right: 'the app runs in the container, and the container was never given it', note: 'every question needing the model failed, on the first one I asked'},   // P5
  {k: 'head', kicker: 'now that is the small problem', lines: ['I nearly put something', 'in front of you that', 'was not *true*'], size: 54, trans: 'push'},   // P6
  {k: 'shot', src: sh('24'), kicker: 'i went and read how it decides to move a card', lines: ['Not the *model*.', 'A pattern'], size: 68},   // P7  R 90.0-140.0s
  {k: 'myth', kicker: 'somewhere in there is a rule', lines: ['The model is *never asked*'], size: 76, wrong: 'you asked in English, so a language model understood you', right: 'if the sentence looks roughly like move something to somewhere, pull out the two things', note: 'written the way this was written in nineteen ninety'},   // P8
  {k: 'rows', kicker: 'and see how close I came to not noticing', lines: ['The two faults *hid*', 'each other'], size: 62, rows: [{t: 'With the key missing, every real question was failing'}, {t: 'And moving a card worked perfectly'}, {t: 'Because moving a card never needed the model', hot: true}]},   // P9
  {k: 'head', kicker: 'so the demo I was about to record', lines: ['Would have been a', '*regular expression*'], size: 58},   // P10
  {k: 'myth', kicker: 'and nothing was hidden from me', lines: ['It did not *occur* to it'], size: 92, wrong: 'it concealed how the move actually worked', right: 'it is all in the code, in a file I could open', note: 'it was never asked to mention it, so it did not'},   // P11
  {k: 'head', kicker: 'which is the sharpest version of the whole week', lines: ['Whether what you asked is', 'what you *wanted*'], size: 52, accent: '#f5c518'},   // P12
  {k: 'rows', kicker: 'and notice how it was caught', lines: ['*Using* it', 'found the reading'], size: 72, rows: [{t: 'Not by reading the code — I read that part afterwards'}, {t: 'By using the thing and finding one bit broken'}, {t: 'Which made me look at the bit that worked', hot: true}]},   // P13
  {k: 'shot', src: sh('31'), kicker: 'so I sent it both', lines: ['Slower and *real*, over', 'instant and fake'], size: 54},   // P14  R 600.0-660.0s
  {k: 'shot', src: sh('32')},   // P15  D 64.0-80.0s
  {k: 'rows', kicker: 'and it does, accurately', lines: ['*Reading* the board'], size: 104, rows: [{t: 'Naming the columns and what is in them'}, {t: 'Which means it is not guessing', hot: true}]},   // P16
  {k: 'shot', src: sh('34'), kicker: 'and now the one I have been building towards', lines: ['Not by *dragging* it.', 'By asking'], size: 62},   // P17  D 104.0-150.0s
  {k: 'shot', src: sh('35')},   // P18  D 150.0-176.0s
  {k: 'flow', kicker: 'and this time I can tell you what happened, because I checked', lines: ['The *whole* round trip'], size: 88, nodes: [{t: 'the board was described', sub: 'to a language model, with my sentence'}, {t: 'it decided', sub: 'which card, and which column I meant'}, {t: 'it said so', sub: 'in a form our back end understands'}, {t: 'and the database changed', sub: 'for real'}], caption: 'I never named the card. It worked out which one I meant.'},   // P19
  {k: 'shot', src: sh('37')},   // P20  D 196.0-230.0s
  {k: 'rows', kicker: 'so look at what this is now', lines: ['Reorganise your work', 'by *asking*'], size: 62, numbered: true, rows: [{t: 'A board that persists, in a database, in a container'}, {t: 'A front end, a back end, and an API between them'}, {t: 'And an assistant that can change it', hot: true}]},   // P21
  {k: 'head', kicker: 'that is not a demo', lines: ['The shape of a product', 'people *pay* for'], size: 58},   // P22
  {k: 'myth', kicker: 'and built, I want to say clearly', lines: ['Both found by *using* it'], size: 88, wrong: 'and it all worked first time', right: 'with one thing in it that was fake until forty minutes ago', note: 'and a key that was never plugged in'},   // P23
  {k: 'head', kicker: 'and now the part that matters more than the demo', lines: ['Anybody can show you', 'a thing *working*'], size: 54, trans: 'rise'},   // H1
  {k: 'head', kicker: 'this is not finished', lines: ['The list is only useful', 'while you are *impressed*'], size: 52},   // H2
  {k: 'rows', kicker: 'so, quickly', lines: ['What is *wrong* with it'], size: 96, numbered: true, rows: [{t: 'One user, hard-coded'}, {t: 'No sign it is thinking, so it looks frozen'}, {t: 'One board'}, {t: 'Nothing deployed — it exists only on this machine'}]},   // H3
  {k: 'head', kicker: 'and then the one to sit with', lines: ['Because it is the kind', 'of thing that *ships*'], size: 56},   // H4
  {k: 'shot', src: sh('45'), kicker: 'the sign-in', lines: ['Let me show you *something*'], size: 92},   // H5  H 2.0-22.0s
  {k: 'rows', kicker: 'your browser downloads these to make the page work', lines: ['You can read *any* of them'], size: 76, rows: [{t: 'That is normal. Every website does it.'}]},   // H6
  {k: 'shot', src: sh('47')},   // H7  H 22.0-44.0s
  {k: 'shot', src: sh('48'), kicker: 'and in one of them', lines: ['The *password*.', 'Written out'], size: 92},   // H8  H 44.0-92.0s
  {k: 'myth', kicker: 'which means it is not a lock', lines: ['A *sign* on a door'], size: 116, wrong: 'you have to sign in, so the board is protected', right: 'anybody who opens the page finds the password in thirty seconds', note: 'and I have been calling it a sign-in all afternoon'},   // H9
  {k: 'head', kicker: 'now, to be fair, nobody did anything wrong', lines: ['My *shortcut*.', 'Not its mistake'], size: 68},   // H10
  {k: 'rows', kicker: 'but this is exactly how it happens outside a classroom', lines: ['Nobody goes *back*'], size: 100, rows: [{t: 'The shortcut goes in for perfectly good reasons'}, {t: 'Somebody shows the thing to somebody else'}, {t: 'And nobody asks which shortcuts were load-bearing', hot: true}]},   // H11
  {k: 'myth', kicker: 'so the lesson is not go and learn about authentication', lines: ['*Which kind* is this one'], size: 76, wrong: 'we will do that properly later', right: 'this must never leave this laptop', note: 'know which of your simplifications is which'},   // H12
  {k: 'head', kicker: 'and I could only tell you that', lines: ['Because I *looked* at what', 'my app was sending'], size: 52},   // H13
  {k: 'head', kicker: 'so what would I do about all of it', lines: ['Not fix it by *hand*'], size: 112},   // H14
  {k: 'rows', kicker: 'a fresh conversation, and a different agent', lines: ['Without telling it', 'what I *think*'], size: 68, rows: [{t: 'Review this project and tell me what is wrong with it'}]},   // H15
  {k: 'myth', kicker: 'and it would find the password in four seconds', lines: ['One *prompt*'], size: 140, wrong: 'a second opinion is a luxury for big projects', right: 'something with no stake in what the first one wrote, for one prompt', note: 'the most useful and least used move available to you'},   // H16
  {k: 'head', kicker: 'so that is week one', lines: ['What happened to *you*,', 'not what we built'], size: 54, trans: 'push'},   // W1
  {k: 'myth', kicker: 'because the app is not the point', lines: ['There are two *hundred*', 'of them'], size: 68, wrong: 'you have built a Kanban board', right: 'you have a method, and it is the same wherever you point it', note: 'and most of the two hundred are better than ours'},   // W2
  {k: 'head', kicker: 'five moves', lines: ['Whatever you point', 'it *at*'], size: 104},   // W3
  {k: 'rows', kicker: 'one', lines: ['Write the plan *yourself*'], size: 104, rows: [{t: 'In pieces sized so that if one fails, you know where to look'}]},   // W4
  {k: 'rows', kicker: 'two', lines: ['Ask for *questions* first'], size: 96, rows: [{t: 'And treat silence as a pass, not as agreement'}]},   // W5
  {k: 'rows', kicker: 'three', lines: ['Check in *after* every part'], size: 88, rows: [{t: 'Before the risky thing, rather than after it'}]},   // W6
  {k: 'rows', kicker: 'four', lines: ['Describe faults *precisely*'], size: 84, rows: [{t: 'What you did. What you expected. What happened instead.'}]},   // W7
  {k: 'rows', kicker: 'five', lines: ['Go and *look*'], size: 140, rows: [{t: 'Because its check passing is not the same as it working', hot: true}]},   // W8
  {k: 'myth', kicker: 'and none of those needed you to read a line of code', lines: ['Every one needed', '*attention*'], size: 68, wrong: 'to direct this work you need to understand the code', right: 'you need to notice when the shape of it stops making sense', note: 'that is the trade this whole course is built on'},   // W9
  {k: 'head', kicker: 'now the honest accounting', lines: ['I have been *promising* it', 'all along'], size: 56},   // W10
  {k: 'rows', kicker: 'what we actually saw', lines: ['All of it *true*'], size: 112, rows: [{t: 'Day one — tireless, and it was'}, {t: 'Day three — jumping to conclusions, and it did'}, {t: 'Today — verified with the one tool that could not see the fault'}]},   // W11
  {k: 'rows', kicker: 'and in the same afternoon', lines: ['Also *true*'], size: 124, rows: [{t: 'A database design better than the one in my head'}, {t: 'And it caught a fault in its own scheme when I asked one question', hot: true}]},   // W12
  {k: 'myth', kicker: 'so if you have come out of this week with either extreme', lines: ['I have taught you *badly*'], size: 76, wrong: 'these things are magic', right: 'these things are useless', note: 'both of those are my failure, in opposite directions'},   // W13
  {k: 'head', kicker: 'extremely capable, and no judgement about when to stop', lines: ['*You* are the judgement'], size: 96, accent: '#f5c518'},   // W14
  {k: 'head', kicker: 'next week we change tools', lines: ['And I want your expectations', 'set *properly*'], size: 52},   // W15
  {k: 'head', kicker: 'everything you have learned transfers', lines: ['*All* of it'], size: 150},   // W16
  {k: 'rows', kicker: 'what changes is what you are shown', lines: ['It stops being *polite*'], size: 88, rows: [{t: 'You will see how full the context is'}, {t: 'You will see what it is about to run, before it runs', hot: true}]},   // W17
  {k: 'myth', kicker: 'which is less comfortable and much better', lines: ['Why we did it in', 'this *order*'], size: 62, wrong: 'we should have started with the tool that shows you everything', right: 'you now know what those numbers mean', note: 'because you have felt what happens when nobody is watching them'},   // W18
  {k: 'quote', kicker: 'one more thing to take with you, and it is not mine', text: 'Keep the AI on a tight leash — the temptation is always to hand it something huge and hope.', who: 'the idea is Karpathy\'s; the phrasing is mine'},   // W19
  {k: 'rows', kicker: 'and everything today was that leash', lines: ['*Ten* parts, not one prompt'], size: 76, numbered: true, rows: [{t: 'A stop before the database'}, {t: 'A browser opened by hand when the tests said everything was fine'}, {t: 'And a demo pulled apart because it was too good to be true', hot: true}]},   // W20
  {k: 'myth', kicker: 'and it is not a lack of ambition', lines: ['The *only* way found', 'so far'], size: 68, wrong: 'working in small steps means thinking small', right: 'it is how you get ambitious work out of these and still know whether it worked', note: 'nobody has found another one'},   // W21
  {k: 'head', kicker: 'well done for getting through the first week', lines: ['See you in the *second*'], size: 88, trans: 'fade'},   // W22
];

export const Nocode34: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode34/vo" files={FILES} gap={GAP} />
);
