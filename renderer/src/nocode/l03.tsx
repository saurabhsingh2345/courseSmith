// nocode03 — "The Missing Manual for Agentic AI Coding" (his L3)
//
// His lecture is 415s of slides and a talking head. Ours keeps his spine —
// the Karpathy post, this program is the manual, who it is for, the anti-hype
// close — and spends the room his talking head occupied on the thing his
// lecture promises but never delivers: what every word on that list means.
//
// REBUILT 2026-09-15. The first cut was 22 slides averaging 17.5 seconds, one
// per narration paragraph, and every reveal on a slide fired inside its first
// 1.65s — so two thirds of this lecture was a still frame with a voice over
// it, which is what the team saw. The narration is unchanged to the sentence;
// it is re-segmented into 51 short slides by tools/nocode/scripts/l03_v2.py,
// which emits BOTH the voice manifest and the slide list below so the two
// cannot drift. Each row now lands on the sentence that says it, driven by
// cues.json from adam_deck.py.
//
// Regenerate with:
//   /usr/bin/python3 tools/nocode/scripts/l03_v2.py
//   /usr/bin/python3 tools/nocode/scripts/adam_deck.py \
//       tools/nocode/narration/l3_v2.json renderer/public/nocode03v2/vo

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';
import RAW from './l03.slides.json';
import DURS_JSON from '../../public/nocode03v2/vo/durs.json';
import CUES_JSON from '../../public/nocode03v2/vo/cues.json';

export const DURS: number[] = DURS_JSON;
export const CUES: number[][] = CUES_JSON;
const SLIDES = RAW as unknown as Slide[];
export const FILES = DURS.map((_, i) => `s${String(i).padStart(2, '0')}.mp3`);

export const GAP = 0.55;
export const L3_FRAMES = deckFrames(DURS, GAP);

export const Nocode03: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode03v2/vo" files={FILES}
        gap={GAP} cues={CUES} />
);
