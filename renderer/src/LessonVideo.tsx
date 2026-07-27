import {useMemo} from 'react';
import {AbsoluteFill, Audio, Sequence, staticFile} from 'remotion';
import {CaptionTrack} from './components/CaptionTrack';
import {CalloutLayer} from './components/CalloutLayer';
import {CodeScene} from './components/CodeScene';
import {DiagramScene} from './components/DiagramScene';
import {D3Diagram} from './components/D3Diagram';
import {PointsScene} from './components/PointsScene';
import {PythonExecutionViz} from './components/PythonExecutionViz';
import {MemoryLayout} from './components/MemoryLayout';
import {SceneBackground, type Surface} from './components/SceneBackground';
import {SectionTransition} from './components/SectionTransition';
import {TerminalScene} from './components/TerminalScene';
import {TitleCard} from './components/TitleCard';
import {VSCodeScene} from './components/VSCodeScene';
import {WorkspaceScene} from './components/WorkspaceScene';
import {WhiteboardScene} from './components/WhiteboardScene';
import {FlowScene} from './components/FlowScene';
import {IllustrationScene} from './components/IllustrationScene';
import {CastScene} from './components/CastScene';
import {StoryScene} from './components/StoryScene';
import {DataScene} from './components/DataScene';
import {FPS, LessonVideoProps, Scene, msToFrame} from './types';
import {ResolvedTheme, resolveTheme} from './theme/theme';
import {resolveMotion, secondsToFrames} from './theme/motion';

const sceneContent = (
  scene: Scene,
  props: LessonVideoProps,
  theme: ResolvedTheme,
  durationInFrames: number,
) => {
  const {assetBase, motion} = props;
  switch (scene.type) {
    case 'title':
      return <TitleCard theme={theme} motion={motion} props={scene.props} />;
    case 'code':
      // A code scene with an execution trace becomes the Python-Tutor viz —
      // or, when props.view === 'memory', the stack/heap MemoryLayout view of
      // that same trace; otherwise it stays a self-typing code scene.
      if (scene.props.trace) {
        return scene.props.view === 'memory' ? (
          <MemoryLayout
            theme={theme}
            motion={motion}
            durationInFrames={durationInFrames}
            props={scene.props}
          />
        ) : (
          <PythonExecutionViz
            theme={theme}
            motion={motion}
            durationInFrames={durationInFrames}
            props={scene.props}
          />
        );
      }
      return <CodeScene theme={theme} props={scene.props} />;
    case 'diagram':
      // A "d3" diagram is a structured graph laid out and animated by D3;
      // otherwise it's a generated SVG revealed group-by-group.
      return scene.props.kind === 'd3' ? (
        <D3Diagram theme={theme} motion={motion} assetBase={assetBase} props={scene.props} />
      ) : (
        <DiagramScene theme={theme} assetBase={assetBase} props={scene.props} />
      );
    case 'terminal':
      return <TerminalScene theme={theme} assetBase={assetBase} props={scene.props} />;
    case 'points':
      return (
        <PointsScene theme={theme} motion={motion} sceneStartMs={scene.startMs} props={scene.props} />
      );
    case 'walkthrough':
      return (
        <VSCodeScene
          theme={theme}
          sceneStartMs={scene.startMs}
          durationInFrames={durationInFrames}
          props={scene.props}
        />
      );
    case 'workspace':
      return <WorkspaceScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'whiteboard':
      return <WhiteboardScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'flow':
      return <FlowScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'illustration':
      return <IllustrationScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'cast':
      return <CastScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'story':
      return <StoryScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'data':
      return <DataScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    default:
      return null;
  }
};

/**
 * Which backdrop this video stands on, from the scenes in it.
 *
 * A snippet is one template start to finish, so its scenes agree and it gets
 * that template's surface. A lesson mixes title, code, diagram, terminal and
 * points, and mixed content gets the neutral one — a course that switched
 * backdrop every time it showed a diagram would be a course that flickers.
 *
 * Derived rather than passed, so a new template gets a considered backdrop by
 * adding one line here instead of by remembering to thread a prop from Go.
 */
const surfaceFor = (scenes: Scene[]): Surface => {
  // The title card is the same card whatever follows it, so it does not get a
  // vote — otherwise every snippet would be "mixed" and land on the default.
  const kinds = new Set(scenes.map((s) => s.type).filter((t) => t !== 'title'));
  if (kinds.size !== 1) return 'default';
  switch ([...kinds][0]) {
    case 'whiteboard':
      return 'paper';
    case 'flow':
      return 'blueprint';
    case 'cast':
    case 'story':
      return 'spotlight';
    case 'data':
      return 'clean';
    default:
      return 'default';
  }
};

export const LessonVideo: React.FC<LessonVideoProps> = (props) => {
  const {assetBase, audioFile, scenes, captions, captionEmphasis, motion} = props;
  const theme = useMemo(() => resolveTheme(props.theme), [props.theme]);
  const surface = useMemo(() => surfaceFor(scenes), [scenes]);
  // Scenes cross-dissolve. Every sequence but the last is extended past its
  // scene end by `cross` frames, so the outgoing scene is still mounted and
  // fading while the incoming one fades in on top of it (later siblings paint
  // last). Scene *content* still receives its true duration — only the mount
  // window is stretched, so nothing that paces itself off durationInFrames
  // (the execution viz, the walkthrough, the memory layout) is affected.
  const cross = secondsToFrames(FPS, resolveMotion(motion).timing.normal);
  return (
    <AbsoluteFill style={{fontFamily: theme.fontBody}}>
      <SceneBackground theme={theme} surface={surface} />
      {audioFile ? <Audio src={staticFile(`${assetBase ?? ''}/${audioFile}`)} /> : null}
      {scenes.map((scene, i) => {
        const from = msToFrame(scene.startMs);
        const duration = Math.max(1, msToFrame(scene.endMs) - from);
        const isLast = i === scenes.length - 1;
        return (
          <Sequence
            key={i}
            from={from}
            durationInFrames={isLast ? duration : duration + cross}
            name={`${scene.type}-${i}`}
          >
            <SectionTransition
              durationInFrames={duration}
              crossFrames={cross}
              motion={motion}
              isLast={isLast}
            >
              {sceneContent(scene, props, theme, duration)}
              {/* Inside the transition so a callout fades out with its own
                  scene instead of hanging fully opaque over the next one. */}
              <CalloutLayer
                theme={theme}
                callouts={scene.callouts ?? []}
                sceneStartMs={scene.startMs}
              />
            </SectionTransition>
          </Sequence>
        );
      })}
      {/* On-screen captions are opt-in; scene graphs built with
          style.captions off carry no caption words (null/empty). */}
      {captions && captions.length > 0 ? (
        <CaptionTrack theme={theme} captions={captions} emphasis={captionEmphasis} />
      ) : null}
    </AbsoluteFill>
  );
};
