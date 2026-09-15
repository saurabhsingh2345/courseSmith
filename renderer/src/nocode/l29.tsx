// nocode29 — Web Apps 101 (his L29, 11:20)
//
// Slides only. His lecture ends with a live Docker install and a tour of his own
// Docker Desktop; we cannot film ours, because this machine holds 25 images, 53
// volumes and 7 running containers from real client work and the Containers and
// Images lists would publish those project names. The Docker half is animated
// instead, and C13/C14 describe a fresh install rather than implying we are
// looking at mine.
//
// ref-L29 calls this the most important lecture in Section 1 for a no-code
// audience. Two beats carry it: secrets live on the back end (A9-A11), because
// the front end runs on somebody else's computer; and the LLM slop look
// (B10-B14) — the purple gradient, three thin-line icons, the enormous headline,
// the rounded "Get started" button — which sets up "name the fault and name the
// standard".

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';

export const DURS = [
  11.527, 11.684, 12.475, 7.690, 13.378, 10.368, 10.936, 13.254,
  7.718, 13.517, 10.383, 6.778, 12.674, 14.616, 8.800, 14.830,
  15.910, 3.907, 14.725, 14.917, 12.086, 13.414, 14.020, 12.974,
  6.639, 14.111, 13.816, 12.426, 13.849, 8.187, 13.412, 5.132,
  14.816, 8.640, 16.602, 13.235, 15.670, 13.658, 12.357, 9.445,
  4.101, 13.639, 12.777, 15.783, 10.185, 12.781, 10.963, 9.198,
  9.951, 15.870, 15.382, 8.700, 11.089, 7.361, 14.811,
];
export const FILES = [
  'A1', 'A2', 'A3', 'A4', 'A5', 'A6', 'A7', 'A8',
  'A9', 'A10', 'A11', 'A12', 'A13', 'A14', 'D4', 'D5',
  'D6', 'A15', 'D1', 'D2', 'D3', 'A16', 'A17', 'A18',
  'B1', 'B2', 'B3', 'B4', 'B5', 'B6', 'B7', 'B8',
  'B9', 'B10', 'B11', 'B12', 'B13', 'B14', 'C1', 'C2',
  'C3', 'C4', 'C5', 'C6', 'C7', 'C8', 'C9', 'C10',
  'C11', 'C12', 'C13', 'C14', 'D7', 'D8', 'C15',
].map((n) => `${n}.mp3`);
export const GAP = 0.240;
export const L29_FRAMES = deckFrames(DURS, GAP);

const sh = (n: string) => `nocode29/shots/${n}.mp4`;

const SLIDES: Slide[] = [
  {k: 'head', kicker: 'before we build anything', lines: ['The *vocabulary*'], size: 124, trans: 'fade'},   // A1
  {k: 'head', kicker: 'and I mean this', lines: ['If you know this,', 'put me on *2x*'], size: 68},   // A2
  {k: 'myth', kicker: 'what all of this actually is', lines: ['A *web* application'], size: 76, wrong: 'an app you install from a store', right: 'an address you open in a browser', note: 'no installer, no app store — that is the whole definition'},   // A3
  {k: 'head', kicker: 'and every one of them is', lines: ['*Two halves*', 'talking to each other'], size: 62, trans: 'rise'},   // A4
  {k: 'nest', kicker: 'the half you can see', lines: ['The *front end*'], size: 84, outer: 'your browser, on your computer', inner: 'the front end', ring: ['the page', 'how it looks', 'what happens when you click']},   // A5
  {k: 'rows', kicker: 'and the three names you will see forever', lines: ['*HTML. CSS. JavaScript*'], size: 68, numbered: true, rows: [{t: 'HTML — the page itself'}, {t: 'CSS — how the page looks'}, {t: 'JavaScript — what happens when you drag a card', hot: true}]},   // A6
  {k: 'nest', kicker: 'the half you cannot', lines: ['The *back end*'], size: 88, outer: 'a server, somewhere else', inner: 'the back end', caption: 'it sent you the page in the first place'},   // A7
  {k: 'rows', kicker: 'and it does the serious things', lines: ['What the *back end* is for'], size: 64, rows: [{t: 'Reading and writing the database'}, {t: 'Calling a language model'}, {t: 'Talking to other companies\' services'}, {t: 'Anything that must be remembered, or paid for, or protected', hot: true}]},   // A8
  {k: 'head', kicker: 'the most important sentence in this lecture', lines: ['Your secrets live in', 'the *back end*'], size: 68, accent: '#e53935'},   // A9
  {k: 'myth', kicker: 'because of where the front end runs', lines: ['Not a secret. A *notice*'], size: 72, wrong: 'put the key in the front end, it is only used once', right: 'the front end runs on their computer — anyone can read it', note: 'you find out when the bill arrives'},   // A10
  {k: 'hierarchy', kicker: 'which is what this file is for', lines: ['*.env*, and never checked in'], size: 76, nodes: [{t: 'pm', depth: 0, kind: 'dir'}, {t: '.env', depth: 1, kind: 'agents'}, {t: '.gitignore', depth: 1, kind: 'file'}, {t: 'frontend', depth: 1, kind: 'dir'}, {t: 'backend', depth: 1, kind: 'dir'}], active: 1, caption: 'listed inside .gitignore, so git never sees it'},   // A11
  {k: 'head', kicker: 'two halves, two machines', lines: ['So how do they *talk*'], size: 96},   // A12
  {k: 'myth', kicker: 'and the word is less impressive than it sounds', lines: ['An *API*'], size: 132, wrong: 'some complicated protocol you have to learn', right: 'a fixed list of messages the back end agrees to answer', note: 'the front end sends one. the back end answers.'},   // A13
  {k: 'flow', kicker: 'so, dragging one card', lines: ['The *round trip*'], size: 76, nodes: [{t: 'you drag a card', sub: 'in the browser'}, {t: '"this card moved"', sub: 'the API call', tone: 'yellow'}, {t: 'written to the database', sub: 'on the server'}, {t: '"done"', sub: 'and the board is right tomorrow', tone: 'yellow'}], caption: 'that round trip is the whole reason anything is remembered'},   // A14
  {k: 'head', kicker: 'and it explains something you have felt', lines: ['Every call is a *journey*'], size: 80},   // D4
  {k: 'rows', kicker: 'usually hundredths of a second — but real', lines: ['Why *spinners* exist'], size: 76, rows: [{t: 'The two halves are genuinely apart'}, {t: 'The distance is real, and sometimes slow'}, {t: 'A good app tells you it is thinking', hot: true}]},   // D5
  {k: 'myth', kicker: 'and this is a decision somebody makes', lines: ['Responsive, or *frozen*'], size: 72, wrong: 'wait for the server, then update the button', right: 'say "saving" instantly, let the journey happen behind it', note: 'if nobody decides this, the app feels broken while working perfectly'},   // D6
  {k: 'head', kicker: 'now put the week on that diagram', lines: ['Where each project *sat*'], size: 76, trans: 'push'},   // A15
  {k: 'nest', kicker: 'day one', lines: ['The shooter:', '*all* front end'], size: 68, outer: 'your browser', inner: 'the whole game', caption: 'no server, no account, nothing saved'},   // D1
  {k: 'rows', kicker: 'which is not a criticism', lines: ['Front end only is *fine*'], size: 72, rows: [{t: 'A calculator'}, {t: 'A drawing tool'}, {t: 'A game'}, {t: 'Nothing remembered, nothing secret — you need no back end', hot: true}]},   // D2
  {k: 'myth', kicker: 'so here is the line, and carry it with you', lines: ['When you need the *other half*'], size: 68, wrong: 'every app needs a server', right: 'you need one to REMEMBER, or to keep a SECRET', note: 'that tells you how big a job is before you start it'},   // D3
  {k: 'rows', kicker: 'day three, four times over', lines: ['The board: *all* front end'], size: 68, rows: [{t: 'No back end at all'}, {t: 'Which is why it forgot everything on reload'}, {t: 'It was never storing anything — there was nowhere to store it', hot: true}]},   // A16
  {k: 'rows', kicker: 'yesterday', lines: ['A back end — in *JavaScript*'], size: 68, rows: [{t: 'Both halves in the same language'}, {t: 'A completely normal way to build'}, {t: 'Today we do it differently, and it is worth saying why'}]},   // A17
  {k: 'myth', kicker: 'today', lines: ['JavaScript front.', '*Python* back'], size: 68, wrong: 'python is the better language', right: 'when the other end is a language model, python is where that world lives', note: 'and the help you find online will assume it'},   // A18
  {k: 'head', kicker: 'a short history, because the names never get explained', lines: ['The *front end*,', 'in order'], size: 64, trans: 'rise'},   // B1
  {k: 'rows', kicker: 'first', lines: ['Hand-written *pages*'], size: 84, rows: [{t: 'You wrote the page, how it looked, and a bit of behaviour'}, {t: 'A small library if you wanted something fancy'}, {t: 'Still how much of the web works, and nothing wrong with it'}]},   // B2
  {k: 'myth', kicker: 'then, one idea changed it', lines: ['*Components*'], size: 124, wrong: 'write a page, then update bits of it by hand', right: 'write a card, write a column — and they redraw themselves', note: 'when the data changes, the component follows'},   // B3
  {k: 'rows', kicker: 'four names, roughly one job', lines: ['React. Vue.', 'Angular. *Svelte*'], size: 62, rows: [{t: 'They do broadly the same thing'}, {t: 'Ours uses React'}, {t: 'Because the front end we inherited uses React', hot: true}, {t: 'Which is usually how you pick one'}]},   // B4
  {k: 'myth', kicker: 'and out of that, a pattern with a bad name', lines: ['The *single page* app'], size: 72, wrong: 'every click loads a whole new page', right: 'loaded once — then pieces fetch what they need and update in place', note: 'SPA. that is all it means.'},   // B5
  {k: 'head', kicker: 'which is why a good web app', lines: ['Feels like a *program*,', 'not a document'], size: 62},   // B6
  {k: 'stack', kicker: 'and on top sit the higher-level frameworks', lines: ['*Next.js*'], size: 112, layers: [{t: 'your pages', sub: 'the part you write', h: 44, tone: 'yellow'}, {t: 'routing', sub: 'which address shows what', h: 56}, {t: 'data fetching', sub: 'getting what the page needs', h: 56}, {t: 'server or browser rendering', sub: 'decided per page', h: 62}], foot: 'the decisions nobody wants to make twice'},   // B7
  {k: 'head', kicker: 'and now the honest part', lines: ['I am not *strong* at this'], size: 88, trans: 'push'},   // B8
  {k: 'rows', kicker: 'so I will not pretend otherwise', lines: ['In front of an app', 'I did not *hand write*'], size: 58, rows: [{t: 'Building front ends by hand is not something I am good at'}, {t: 'It is exactly the work coding agents made possible for me'}, {t: 'Which is the whole premise of this course', hot: true}]},   // B9
  {k: 'head', kicker: 'but there is a catch, and you have seen it', lines: ['Ask for a web page,', 'get *the same* web page'], size: 58},   // B10
  {k: 'rows', kicker: 'you will not be able to unsee this', lines: ['The *slop* look'], size: 104, rows: [{t: 'A dark background with a purple gradient sliding across it'}, {t: 'Three feature cards, three thin-line icons, same icon set'}, {t: 'An enormous headline, a smaller subheading'}, {t: 'A rounded button that says "Get started"', hot: true}]},   // B11
  {k: 'myth', kicker: 'and the problem is not that it is ugly', lines: ['It is *anonymous*'], size: 88, wrong: 'this looks bad', right: 'this looks like nobody decided anything', note: 'it looks like every other thing generated the same way'},   // B12
  {k: 'rows', kicker: 'and this is where you add value — code or no code', lines: ['Not a *coding* skill'], size: 80, rows: [{t: 'What the person needs to see first'}, {t: 'What can safely be shown later'}, {t: 'What should not be on the screen at all'}, {t: 'A thinking skill. The model does not have it.', hot: true}]},   // B13
  {k: 'myth', kicker: 'so push back on the first draft', lines: ['Name the fault.', 'Name the *standard*'], size: 62, wrong: 'make it better', right: 'this is the same purple gradient every generated site has — our brief has a palette, use it', note: 'we established yesterday what vague feedback buys you'},   // B14
  {k: 'head', kicker: 'one more word, and it frightens people off', lines: ['*Docker*'], size: 140, trans: 'fade'},   // C1
  {k: 'nest', kicker: 'and it is simpler than its reputation', lines: ['A computer *inside*', 'your computer'], size: 62, outer: 'your machine', inner: 'a small, sealed one', ring: ['its own copy of everything', 'cannot see your files', 'cannot touch the rest']},   // C2
  {k: 'head', kicker: 'two reasons people use it', lines: ['And *both* matter', 'to you specifically'], size: 62},   // C3
  {k: 'rows', kicker: 'one — isolation', lines: ['What happens in the box,', 'stays in the *box*'], size: 58, rows: [{t: 'It cannot install something odd on your laptop'}, {t: 'It cannot overwrite your files'}, {t: 'It cannot fight with a Python you installed two years ago', hot: true}]},   // C4
  {k: 'head', kicker: 'and notice how directly that speaks to this week', lines: ['A sealed box is the', 'honest answer to the *nervousness*'], size: 52, accent: '#f5c518'},   // C5
  {k: 'rows', kicker: 'two — portability', lines: ['Build once.', 'Runs *anywhere*'], size: 72, rows: [{t: 'The box is described by a file, so it can be rebuilt anywhere'}, {t: 'Your machine, a colleague\'s, a rented server'}, {t: '"It usually just works" — rarely true of software', hot: true}]},   // C6
  {k: 'head', kicker: 'three words, then we are done', lines: ['Dockerfile.', 'Image. *Container*'], size: 62, trans: 'rise'},   // C7
  {k: 'editor', kicker: 'one — the recipe', lines: ['A *Dockerfile*'], size: 104, file: 'Dockerfile', start: 1, rows: [{t: 'FROM python:3.12', kind: 'code'}, {t: 'COPY . /app', kind: 'code'}, {t: 'RUN uv sync', kind: 'code'}, {t: 'CMD ["uvicorn", "main:app"]', kind: 'code'}], caption: 'you can read that without knowing anything. it is a list of instructions.'},   // C8
  {k: 'myth', kicker: 'two — follow the recipe', lines: ['An *image*'], size: 132, wrong: 'the running thing', right: 'the finished snapshot — the cake, baked and frozen', note: 'it does not do anything on its own. it sits there, ready.'},   // C9
  {k: 'flow', kicker: 'three — start it', lines: ['A *container*'], size: 124, nodes: [{t: 'image', sub: 'one snapshot', tone: 'yellow'}, {t: 'container', sub: 'running'}, {t: 'container', sub: 'running'}, {t: 'container', sub: 'running'}], caption: 'one image can start as many as you like — and they do not know about each other'},   // C10
  {k: 'head', kicker: 'recipe, cake, cake being eaten', lines: ['That is *Docker*'], size: 112},   // C11
  {k: 'rows', kicker: 'and installing it is the dull part', lines: ['Take *every* default'], size: 80, rows: [{t: 'docker.com — download Docker Desktop for your machine'}, {t: 'On Windows it asks about WSL. The answer is yes.'}, {t: 'It may want a restart. Give it one.'}]},   // C12
  {k: 'rows', kicker: 'and three things in the left rail are the whole tour', lines: ['Containers. Images.', '*Volumes*'], size: 58, numbered: true, rows: [{t: 'Containers — what is alive right now'}, {t: 'Images — what is built and ready to start'}, {t: 'Volumes — sealed storage that survives the container', hot: true}], stamp: 'a container thrown away forgets everything inside it'},   // C13
  {k: 'head', kicker: 'on a fresh install all three are empty', lines: ['They will not', '*stay* empty'], size: 76},   // C14
  {k: 'rows', kicker: 'and two thirds of you knew all of that', lines: ['That is genuinely', 'the *vocabulary*'], size: 62, rows: [{t: 'Front end — in the browser'}, {t: 'Back end — on a server, where the secrets live'}, {t: 'API — the fixed list of messages between them'}, {t: 'Container — a sealed box to run it in', hot: true}]},   // D7
  {k: 'head', kicker: 'none of it is difficult', lines: ['The people who know it', 'forgot they had to *learn* it'], size: 54, trans: 'rise'},   // D8
  {k: 'head', kicker: 'next, we start building', lines: ['And by the *rules*', 'this time'], size: 72, trans: 'push'},   // C15
];

export const Nocode29: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode29/vo" files={FILES} gap={GAP} />
);
