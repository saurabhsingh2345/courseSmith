// nocode02 — "The agent builds a shooter" (his L2), REBUILT 2026-09-15.
//
// 6:54 of screen recording, 55% pixel-frozen, with a 44-second held frame at
// 2:19. The material here is good — a game being played and an agent being
// driven — so thirteen of fifteen segments keep their own footage once the
// waits are ramped out. Only N07 and N08 needed cards; N08 had no live picture
// anywhere in its take, which is where that 44-second hold came from.
//
// See tools/nocode/scripts/l02_v2.py to regenerate.

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';
import RAW from './l02.slides.json';
import DURS_JSON from '../../public/nocode02v2/vo/durs.json';
import CUES_JSON from '../../public/nocode02v2/vo/cues.json';

export const DURS: number[] = DURS_JSON;
export const CUES: number[][] = CUES_JSON;
const SLIDES = RAW as unknown as Slide[];
export const FILES = DURS.map((_, i) => `s${String(i).padStart(2, '0')}.mp3`);

export const GAP = 0.4;
export const L2_FRAMES = deckFrames(DURS, GAP);

export const Nocode02v2: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode02v2/vo" files={FILES}
        gap={GAP} cues={CUES} />
);
