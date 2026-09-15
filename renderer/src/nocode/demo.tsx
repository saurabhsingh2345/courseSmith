// Preview compositions — not part of any lecture. Used to eyeball a template
// on its own before it goes into a film.
import React from 'react';
import {AbsoluteFill, Sequence} from 'remotion';
import {C, Slide, FPS} from './kit';
import {BrowserSlide, BrowserWindow} from './chrome';

export const DemoBrowser: React.FC = () => (
  <BrowserSlide img="demo/localhost.png" url="localhost:4599" tab="Arena"
    kicker="what you just built, running" lines={['On your *own* machine']} scale={0.74} />
);

export const DemoBrowserBare: React.FC = () => (
  <AbsoluteFill style={{background: C.bg, alignItems: 'center', justifyContent: 'center'}}>
    <BrowserWindow img="demo/localhost.png" url="localhost:4599" scale={0.86} />
  </AbsoluteFill>
);

// ---- one slide per template, 5s each, for the contact sheet ---------------
export const GALLERY: Slide[] = [
  {k: 'tokens', kicker: 'a model does not read words',
   lines: ['It reads *tokens*'], size: 62,
   chips: ['Un', 'question', 'ably', ' the', ' capital', ' of', ' France'],
   hot: [0, 2], caption: 'the chips disagree with the word boundaries — that is the point'},
  {k: 'probs', kicker: 'and it never picks a word',
   lines: ['It picks a', '*probability*'], size: 62,
   items: [{t: 'four', p: 94}, {t: 'lots', p: 3}, {t: '4', p: 2},
           {t: 'bananas', p: 0.001}],
   caption: '"two plus two is ..."'},
  {k: 'chat', kicker: 'the illusion of memory',
   lines: ['Nothing is', '*remembered*'], size: 58,
   turns: [{who: 'you', t: 'I am learning to code with agents'},
           {who: 'model', t: 'Great — where would you like to start?'},
           {who: 'you', t: 'What am I learning?'}],
   resend: 2, caption: 'every turn re-sends the whole conversation'},
  {k: 'flow', kicker: 'the part everyone gets wrong',
   lines: ['The model runs', '*nothing*'], size: 58,
   nodes: [{t: 'Model', sub: 'emits tokens', tone: '#f5c518'},
           {t: 'Our code', sub: 'reads them', tone: '#3b82f6'},
           {t: 'The tool', sub: 'actually runs'},
           {t: 'Back in', sub: 'as more input'}],
   loop: true, caption: 'and then it goes round again — that is the loop'},
  {k: 'timeline', kicker: 'the definition kept moving',
   lines: ['What *is* an agent?'], size: 58,
   items: [{when: 'the openai era', t: 'Does work *independently*',
            sub: 'Books the restaurant while you watch'},
           {when: 'early 2025', t: 'An LLM controls the *workflow*',
            sub: 'Hugging Face, then Building Effective Agents'},
           {when: 'late 2025', t: 'Runs *tools* in a *loop*',
            sub: 'Simon Willison — the one that stuck', hot: true}]},
  {k: 'nest', kicker: 'two words people swap by mistake',
   lines: ['The model, and the', '*thing around it*'], size: 54,
   outer: 'the application', inner: 'the model',
   ring: ['memory of the chat', 'web search', 'files and tools', 'the interface']},
  {k: 'tree', kicker: 'two coins, one is heads',
   lines: ['Is the other', '*tails*?'], size: 62,
   leaves: [{t: 'HH', keep: true, note: 'other is heads'},
            {t: 'HT', keep: true, note: 'other is tails'},
            {t: 'TH', keep: true, note: 'other is tails'},
            {t: 'TT', keep: false, note: 'ruled out'}],
   caption: 'two of the three survivors are tails — two thirds, not a half'},
  {k: 'term', kicker: 'and it can say it wants to do something',
   lines: ['A *tool* call'], size: 66,
   term: [{t: 'what is the square root of pi?', kind: 'cmd'},
          {t: 'Python: import math; math.sqrt(math.pi)', kind: 'warn'},
          {t: '1.7724538509055159', kind: 'ok'}],
   title: 'the model, and then our code'},
  {k: 'meter', kicker: 'and it is not free',
   lines: ['The *context* window'], size: 62,
   pct: 68, fill: '68% used', rest: 'room left',
   caption: 'every turn re-sends everything, so it fills up as you talk'},
  {k: 'browser', kicker: 'and here is the thing itself',
   lines: ['Running *locally*'], size: 58,
   img: 'demo/localhost.png', url: 'localhost:4599', tab: 'Arena', scale: 0.66},
  {k: 'gauge', kicker: 'the question every context beat is asking',
   lines: ['Does it *fit*?'], size: 62, trans: 'rise',
   pct: 118, limit: 76, label: 'this conversation', limitLabel: 'the window',
   caption: 'the single most repeated device in the reference clips'},
  {k: 'myth', kicker: 'replacing a wrong model, not filling a gap',
   lines: ['It never *searched*', 'anything'], size: 54, trans: 'push',
   wrong: 'The model went and looked it up',
   right: 'It emitted *tokens*. Our code did the looking',
   note: 'Nothing else changes until you believe this one.'},
  {k: 'stepper', kicker: 'one token at a time, input re-read every step',
   lines: ['This is *inference*'], size: 62, trans: 'fade',
   seed: 'What is the capital of France?',
   steps: [' The', ' capital', ' of', ' France', ' is', ' Paris', '.'],
   caption: 'the whole input goes back in on every single step'},
  {k: 'occupancy', kicker: 'all of it, visible at once',
   lines: ['The window, *filling*'], size: 62, trans: 'rise',
   used: 64, legend: 'spoken for',
   caption: 'a bar hides how much is left — a population does not'},
];

const HOLD = 5;
export const DemoTemplates: React.FC = () => (
  <AbsoluteFill style={{background: C.bg}}>
    {GALLERY.map((s, i) => (
      <Sequence key={i} from={i * HOLD * FPS} durationInFrames={HOLD * FPS}>
        <OneSlide s={s} />
      </Sequence>
    ))}
  </AbsoluteFill>
);

// re-export the private renderer through a tiny shim
import {SlideView} from './kit';
const OneSlide: React.FC<{s: Slide}> = ({s}) => <SlideView s={s} />;
