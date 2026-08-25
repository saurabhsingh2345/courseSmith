import {Composition, registerRoot} from 'remotion';
import {Remaster} from './Remaster';
import timing from './timing.json';

const T = timing as unknown as {total: number};

const Root: React.FC = () => (
  <Composition
    id="Remaster"
    component={Remaster}
    durationInFrames={T.total}
    fps={30}
    width={1920}
    height={1080}
  />
);

registerRoot(Root);
