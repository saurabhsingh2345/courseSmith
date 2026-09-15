// nocode01 — "Welcome: a 3D game in an hour" (his L1), REBUILT 2026-09-15.
//
// The shipped cut was 10:12 and 81% pixel-frozen, with one held frame lasting
// 140 seconds. See tools/nocode/scripts/l01_v2.py for the why and the how; in
// short, the lecture was ten minutes assembled out of about four minutes of
// moving picture, and the rest was the last frame of a clip.
//
// Here the footage is densified (waits ramped, every live frame kept, each clip
// cut to exactly the words over it) and the narration that has no footage
// anywhere in the take — most of it the spoken aside about which projects are
// worth building — runs over animated cards instead of a frozen IDE.
//
// Regenerate:
//   /usr/bin/python3 tools/nocode/scripts/l01_v2.py --manifest
//   /usr/bin/python3 tools/nocode/scripts/adam_deck.py \
//       tools/nocode/narration/l01_v2.json renderer/public/nocode01v2/vo
//   /usr/bin/python3 tools/nocode/scripts/l01_v2.py --slides

import React from 'react';
import {Deck, Slide, deckFrames} from './kit';
import RAW from './l01.slides.json';
import DURS_JSON from '../../public/nocode01v2/vo/durs.json';
import CUES_JSON from '../../public/nocode01v2/vo/cues.json';

export const DURS: number[] = DURS_JSON;
export const CUES: number[][] = CUES_JSON;
const SLIDES = RAW as unknown as Slide[];
export const FILES = DURS.map((_, i) => `s${String(i).padStart(2, '0')}.mp3`);

export const GAP = 0.4;
export const L1_FRAMES = deckFrames(DURS, GAP);

export const Nocode01v2: React.FC = () => (
  <Deck slides={SLIDES} durs={DURS} voDir="nocode01v2/vo" files={FILES}
        gap={GAP} cues={CUES} />
);
