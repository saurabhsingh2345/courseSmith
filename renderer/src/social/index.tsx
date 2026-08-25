import {Composition, registerRoot} from 'remotion';
import {Film} from './Film';
import timing from './timing.json';

const T = timing as unknown as {total: number};

registerRoot(() => (
  <Composition
    id="Lumen"
    component={Film}
    durationInFrames={T.total}
    fps={30}
    width={1920}
    height={1080}
  />
));
