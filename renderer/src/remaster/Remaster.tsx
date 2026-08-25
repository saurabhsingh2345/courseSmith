// The remastered cut.
//
// Beat lengths come from `timing.json`, which the narration builder writes after
// measuring how long each line actually takes to speak. Picture is laid out
// against those numbers, so a rewrite of one line reflows the film instead of
// pushing everything after it out of sync.

import React from 'react';
import {AbsoluteFill, Audio, Sequence, staticFile, useCurrentFrame, useVideoConfig} from 'remotion';
import {BEATS, Beat, Scene} from './edl';
import timing from './timing.json';
import {Ground, Rail, C, ramp} from './kit';
import {Run} from './Run';
import {
  Title,
  Chapter,
  Statement,
  Swap,
  Cards,
  Steps,
  Spotlight,
  Define,
  Scale,
  Address,
  PromptBuild,
  Devices,
  Quiz,
  Roadmap,
  Outro,
} from './scenes';

type Timing = {id: string; frames: number; start: number; sweep?: number[]};
const T = timing as unknown as {beats: Timing[]; total: number; audio: string};

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
    case 'swap':
      return <Swap panels={scene.panels} />;
    case 'cards':
      return <Cards {...scene} sweep={sweep} />;
    case 'steps':
      return <Steps {...scene} sweep={sweep} />;
    case 'spotlight':
      return <Spotlight {...scene} />;
    case 'define':
      return <Define {...scene} />;
    case 'scale':
      return <Scale {...scene} />;
    case 'address':
      return <Address />;
    case 'promptbuild':
      return <PromptBuild {...scene} />;
    case 'devices':
      return <Devices title={scene.title} />;
    case 'quiz':
      return <Quiz {...scene} />;
    case 'roadmap':
      return <Roadmap {...scene} />;
    case 'outro':
      return <Outro {...scene} />;
    case 'run':
      return <Run spec={scene} />;
    default:
      return null;
  }
};

/** A four-frame dip to black on the cut into a chapter card, so it lands. */
const CutIn: React.FC<{children: React.ReactNode}> = ({children}) => {
  const frame = useCurrentFrame();
  const a = ramp(frame, 0, 5);
  return (
    <AbsoluteFill style={{opacity: a}}>
      {children}
    </AbsoluteFill>
  );
};

export const Remaster: React.FC = () => {
  const frame = useCurrentFrame();
  const {durationInFrames} = useVideoConfig();

  // The chapter label riding in the bottom-left comes from whichever beat we
  // are inside, so it never has to be maintained separately from the cut.
  const current = T.beats.find(
    (b) => frame >= b.start && frame < b.start + b.frames
  );
  const chapter = current ? byId.get(current.id)?.chapter ?? '' : '';
  // Nothing is drawn over the recording except the highlight boxes — no ground,
  // no progress rail, no caption. It plays as it was captured.
  const onFootage = current ? byId.get(current.id)?.scene.k === 'run' : false;

  return (
    <AbsoluteFill style={{backgroundColor: C.ink0}}>
      {onFootage ? null : <Ground />}
      {T.beats.map((t) => {
        const beat = byId.get(t.id) as Beat;
        if (!beat) return null;
        const isCard = beat.scene.k === 'chapter' || beat.scene.k === 'title';
        const inner = <Paint scene={beat.scene} frames={t.frames} sweep={t.sweep} />;
        return (
          <Sequence key={t.id} from={t.start} durationInFrames={t.frames} name={t.id}>
            {isCard ? <CutIn>{inner}</CutIn> : inner}
          </Sequence>
        );
      })}
      {onFootage ? null : <Rail progress={frame / durationInFrames} chapter={chapter} />}
      <Audio src={staticFile(T.audio)} />
    </AbsoluteFill>
  );
};
