// The film.
//
// Beat lengths come from timing.json, which the voice builder writes after
// measuring how long each line actually takes to speak. Nothing here decides how
// long anything is on screen.

import React from 'react';
import {
  AbsoluteFill,
  Audio,
  Sequence,
  staticFile,
  useCurrentFrame,
  useVideoConfig,
} from 'remotion';
import {BEATS, Beat, Scene} from './edl';
import timing from './timing.json';
import {Ground, Rail, C, ramp} from './kit';
import {Screen} from './Screen';
import {Chat} from './Chat';
import {
  Title,
  Chapter,
  Statement,
  BeforeAfter,
  Numbers,
  Checklist,
  Define,
  Anatomy,
  Compare,
  Ladder,
  Tiles,
  Loop,
  Quiz,
  Timeline,
  Outro,
} from './scenes';

type T = {id: string; frames: number; start: number; sweep?: number[]};
const TT = timing as unknown as {beats: T[]; total: number; audio: string};

const byId = new Map(BEATS.map((b) => [b.id, b] as const));

const Paint: React.FC<{scene: Scene; frames: number; sweep?: number[]}> = ({
  scene,
  frames,
  sweep,
}) => {
  switch (scene.k) {
    case 'title':
      return <Title {...scene} />;
    case 'chapter':
      return <Chapter n={scene.n} title={scene.title} />;
    case 'statement':
      return <Statement {...scene} />;
    case 'beforeafter':
      return <BeforeAfter {...scene} />;
    case 'numbers':
      return <Numbers {...scene} />;
    case 'checklist':
      return <Checklist {...scene} sweep={sweep} />;
    case 'define':
      return <Define {...scene} />;
    case 'anatomy':
      return <Anatomy {...scene} sweep={sweep} />;
    case 'compare':
      return <Compare {...scene} />;
    case 'ladder':
      return <Ladder {...scene} />;
    case 'tiles':
      return <Tiles {...scene} sweep={sweep} />;
    case 'loop':
      return <Loop {...scene} />;
    case 'quiz':
      return <Quiz {...scene} />;
    case 'timeline':
      return <Timeline {...scene} />;
    case 'outro':
      return <Outro {...scene} />;
    case 'chat':
      return <Chat spec={scene} durationInFrames={frames} />;
    case 'screen':
      return <Screen spec={scene.spec} durationInFrames={frames} />;
    default:
      return null;
  }
};

/** A five-frame fade up on the cut into a chapter card, so it lands. */
const CutIn: React.FC<{children: React.ReactNode}> = ({children}) => {
  const frame = useCurrentFrame();
  return <AbsoluteFill style={{opacity: ramp(frame, 0, 5)}}>{children}</AbsoluteFill>;
};

export const Film: React.FC = () => {
  const frame = useCurrentFrame();
  const {durationInFrames} = useVideoConfig();

  // The chapter label riding at the bottom comes from whichever beat we are
  // inside, so it never has to be maintained separately from the cut.
  const current = TT.beats.find((b) => frame >= b.start && frame < b.start + b.frames);
  const beat = current ? byId.get(current.id) : undefined;

  return (
    <AbsoluteFill style={{backgroundColor: C.ink0}}>
      <Ground />
      {TT.beats.map((t) => {
        const b = byId.get(t.id) as Beat;
        if (!b) return null;
        const card = b.scene.k === 'chapter' || b.scene.k === 'title';
        const inner = <Paint scene={b.scene} frames={t.frames} sweep={t.sweep} />;
        return (
          <Sequence key={t.id} from={t.start} durationInFrames={t.frames} name={t.id}>
            {card ? <CutIn>{inner}</CutIn> : inner}
          </Sequence>
        );
      })}
      <Rail progress={frame / durationInFrames} chapter={beat?.chapter ?? ''} />
      <Audio src={staticFile(TT.audio)} />
    </AbsoluteFill>
  );
};
