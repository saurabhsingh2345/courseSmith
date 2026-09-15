import React from 'react';
import {Composition, registerRoot} from 'remotion';
import {Film} from './Film';
import timing from './timing.json';

const Root: React.FC = () => (
  <>
    <Composition
      id="Genghis"
      component={Film}
      durationInFrames={timing.totalFrames}
      fps={timing.fps}
      width={1920}
      height={1080}
    />
  </>
);

registerRoot(Root);
