import {Composition, registerRoot} from 'remotion';
import {Nocode01Slides, TOTAL_FRAMES} from './slides';
import {InsertD, InsertF, InsertI, D_FRAMES, F_FRAMES, I_FRAMES} from './inserts';

const Root: React.FC = () => (
  <>
    <Composition id="Nocode01Slides" component={Nocode01Slides}
      durationInFrames={TOTAL_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="InsertD" component={InsertD}
      durationInFrames={D_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="InsertF" component={InsertF}
      durationInFrames={F_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="InsertI" component={InsertI}
      durationInFrames={I_FRAMES} fps={30} width={1920} height={1080} />
  </>
);

registerRoot(Root);
