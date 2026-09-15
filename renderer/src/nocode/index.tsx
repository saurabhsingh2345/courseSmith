import {Composition, registerRoot} from 'remotion';
import {Nocode01v2, L1_FRAMES} from './l01';
import {Nocode02v2, L2_FRAMES} from './l02';
import {Nocode03, L3_FRAMES} from './l03';
import {Nocode04, L4_FRAMES} from './l04';
import {Nocode06, L6_FRAMES} from './l06';
import {Nocode07, L7_FRAMES} from './l07';
import {Nocode08, L8_FRAMES} from './l08';
import {Nocode09, L9_FRAMES} from './l09';
import {Nocode10, L10_FRAMES} from './l10';
import {Nocode11, L11_FRAMES} from './l11';
import {Nocode12, L12_FRAMES} from './l12';
import {Nocode13, L13_FRAMES} from './l13';
import {Nocode14, L14_FRAMES} from './l14';
import {Nocode15, L15_FRAMES} from './l15';
import {Nocode16, L16_FRAMES} from './l16';
import {Nocode17, L17_FRAMES} from './l17';
import {Nocode18, L18_FRAMES} from './l18';
import {Nocode19, L19_FRAMES} from './l19';
import {Nocode20, L20_FRAMES} from './l20';
import {Nocode21, L21_FRAMES} from './l21';
import {Nocode29, L29_FRAMES} from './l29';
import {Nocode30, L30_FRAMES} from './l30';
import {Nocode31, L31_FRAMES} from './l31';
import {Nocode32, L32_FRAMES} from './l32';
import {Nocode33, L33_FRAMES} from './l33';
import {Nocode34, L34_FRAMES} from './l34';
import {DemoBrowser, DemoBrowserBare, DemoTemplates, GALLERY} from './demo';
import {INSERT_DEFAULT, InsertProps, W3Insert, insertFrames}
  from './w3insert';

const Root: React.FC = () => (
  <>
    <Composition id="Nocode01v2" component={Nocode01v2}
      durationInFrames={L1_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode02v2" component={Nocode02v2}
      durationInFrames={L2_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode03" component={Nocode03}
      durationInFrames={L3_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode04" component={Nocode04}
      durationInFrames={L4_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode06" component={Nocode06}
      durationInFrames={L6_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode07" component={Nocode07}
      durationInFrames={L7_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode08" component={Nocode08}
      durationInFrames={L8_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="DemoBrowser" component={DemoBrowser}
      durationInFrames={150} fps={30} width={1920} height={1080} />
    <Composition id="DemoBrowserBare" component={DemoBrowserBare}
      durationInFrames={150} fps={30} width={1920} height={1080} />
    <Composition id="DemoTemplates" component={DemoTemplates}
      durationInFrames={GALLERY.length * 5 * 30} fps={30} width={1920} height={1080} />
    <Composition id="Nocode09" component={Nocode09}
      durationInFrames={L9_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode10" component={Nocode10}
      durationInFrames={L10_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode11" component={Nocode11}
      durationInFrames={L11_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode12" component={Nocode12}
      durationInFrames={L12_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode13" component={Nocode13}
      durationInFrames={L13_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode14" component={Nocode14}
      durationInFrames={L14_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode15" component={Nocode15}
      durationInFrames={L15_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode16" component={Nocode16}
      durationInFrames={L16_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode17" component={Nocode17}
      durationInFrames={L17_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode18" component={Nocode18}
      durationInFrames={L18_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode19" component={Nocode19}
      durationInFrames={L19_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode20" component={Nocode20}
      durationInFrames={L20_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode21" component={Nocode21}
      durationInFrames={L21_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode29" component={Nocode29}
      durationInFrames={L29_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode30" component={Nocode30}
      durationInFrames={L30_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode31" component={Nocode31}
      durationInFrames={L31_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode32" component={Nocode32}
      durationInFrames={L32_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode33" component={Nocode33}
      durationInFrames={L33_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="Nocode34" component={Nocode34}
      durationInFrames={L34_FRAMES} fps={30} width={1920} height={1080} />
    <Composition id="W3Insert" component={W3Insert}
      durationInFrames={insertFrames(INSERT_DEFAULT.secs)}
      fps={30} width={1920} height={1080}
      defaultProps={INSERT_DEFAULT}
      calculateMetadata={({props}: {props: InsertProps}) => ({
        durationInFrames: insertFrames(props.secs),
      })} />
  </>
);
registerRoot(Root);
