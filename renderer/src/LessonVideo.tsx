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
import {StageAirContext,
  StageAlignContext,
  StageJustifyContext, StageCaptionsContext} from './components/Stage';
import {SceneCamera} from './components/SceneCamera';
import {SectionTransition, type CutStyle} from './components/SectionTransition';
import {FootageScene} from './components/FootageScene';
import {TerminalScene} from './components/TerminalScene';
import {TitleCard} from './components/TitleCard';
import {VSCodeScene} from './components/VSCodeScene';
import {WorkspaceScene} from './components/WorkspaceScene';
import {WhiteboardScene} from './components/WhiteboardScene';
import {FlowScene} from './components/FlowScene';
import {IllustrationScene} from './components/IllustrationScene';
import {SpineScene} from './components/SpineScene';
import {CastScene} from './components/CastScene';
import {StoryScene} from './components/StoryScene';
import {DataScene} from './components/DataScene';
import {QuizScene} from './components/QuizScene';
import {CompareScene} from './components/CompareScene';
import {AnatomyScene} from './components/AnatomyScene';
import {TimelineScene} from './components/TimelineScene';
import {CanvasScene} from './components/CanvasScene';
import {PromptLoopScene} from './components/PromptLoopScene';
import {MockupScene} from './components/MockupScene';
import {StackScene} from './components/StackScene';
import {SpecScene} from './components/SpecScene';
import {ShowcaseScene} from './components/ShowcaseScene';
import {BreakdownScene} from './components/BreakdownScene';
import {MetricScene} from './components/MetricScene';
import {OccupancyScene} from './components/OccupancyScene';
import {RankingScene} from './components/RankingScene';
import {JournalScene} from './components/JournalScene';
import {MultiplexScene} from './components/MultiplexScene';
import {ForkScene} from './components/ForkScene';
import {CapabilitiesScene} from './components/CapabilitiesScene';
import {BudgetScene} from './components/BudgetScene';
import {LatencyScene} from './components/LatencyScene';
import {MultiplyScene} from './components/MultiplyScene';
import {RatioScene} from './components/RatioScene';
import {TableScene} from './components/TableScene';
import {ToggleScene} from './components/ToggleScene';
import {GaugeScene} from './components/GaugeScene';
import {VerdictScene} from './components/VerdictScene';
import {DecisionScene} from './components/DecisionScene';
import {MythScene} from './components/MythScene';
import {RundownScene} from './components/RundownScene';
import {AnalogyScene} from './components/AnalogyScene';
import {TraceScene} from './components/TraceScene';
import {CostingScene} from './components/CostingScene';
import {ConstellationScene} from './components/ConstellationScene';
import {ChapterScene} from './components/ChapterScene';
import {CycleScene} from './components/CycleScene';
import {ScaleScene} from './components/ScaleScene';
import {ObjectiveScene} from './components/ObjectiveScene';
import {PrereqScene} from './components/PrereqScene';
import {RecapScene} from './components/RecapScene';
import {PitfallScene} from './components/PitfallScene';
import {CheckpointScene} from './components/CheckpointScene';
import {SyllabusScene} from './components/SyllabusScene';
import {OutcomeScene} from './components/OutcomeScene';
import {BridgeScene} from './components/BridgeScene';
import {DrillScene} from './components/DrillScene';
import {LabCardScene} from './components/LabCardScene';
import {MissionScene} from './components/MissionScene';
import {MachineScene} from './components/MachineScene';
import {BlueprintScene} from './components/BlueprintScene';
import {RelayScene} from './components/RelayScene';
import {LayersScene} from './components/LayersScene';
import {PipelineScene} from './components/PipelineScene';
import {RadixScene} from './components/RadixScene';
import {CarryScene} from './components/CarryScene';
import {BitfieldScene} from './components/BitfieldScene';
import {EncodeScene} from './components/EncodeScene';
import {GatesScene} from './components/GatesScene';
import {LadderScene} from './components/LadderScene';
import {RegionsScene} from './components/RegionsScene';
import {LookupScene} from './components/LookupScene';
import {StatesScene} from './components/StatesScene';
import {SchedulerScene} from './components/SchedulerScene';
import {ShellScene} from './components/ShellScene';
import {JourneyScene} from './components/JourneyScene';
import {HandshakeScene} from './components/HandshakeScene';
import {StepperScene} from './components/StepperScene';
import {GrowthScene} from './components/GrowthScene';
import {CallStackScene} from './components/CallStackScene';
import {HistoryScene} from './components/HistoryScene';
import {VersusScene} from './components/VersusScene';
import {ErasScene} from './components/ErasScene';
import {CardsScene} from './components/CardsScene';
import {DuelScene} from './components/DuelScene';
import {SpotlightScene} from './components/SpotlightScene';
import {SceneChrome, Watermark} from './components/SceneChrome';
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
      return (
        <TerminalScene
          theme={theme}
          assetBase={assetBase}
          durationInFrames={durationInFrames}
          props={scene.props}
        />
      );
    case 'footage':
      return (
        <FootageScene
          theme={theme}
          assetBase={assetBase}
          sceneStartMs={scene.startMs}
          durationInFrames={durationInFrames}
          props={scene.props}
        />
      );
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
    case 'spine':
      return (
        <SpineScene
          theme={theme}
          sceneStartMs={scene.startMs}
          durationInFrames={durationInFrames}
          props={scene.props}
        />
      );
    case 'cast':
      return <CastScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'story':
      return <StoryScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'data':
      return <DataScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'quiz':
      return <QuizScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'compare':
      return <CompareScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'anatomy':
      return <AnatomyScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'timeline':
      return <TimelineScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'canvas':
      return <CanvasScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'promptloop':
      return <PromptLoopScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'mockup':
      return <MockupScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'stack':
      return <StackScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'spec':
      return <SpecScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'showcase':
      return <ShowcaseScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'breakdown':
      return <BreakdownScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'metric':
      return <MetricScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'occupancy':
      return <OccupancyScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'ranking':
      return <RankingScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'journal':
      return <JournalScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'multiplex':
      return <MultiplexScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'fork':
      return <ForkScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'capabilities':
      return <CapabilitiesScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'budget':
      return <BudgetScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'latency':
      return <LatencyScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'multiply':
      return <MultiplyScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'ratio':
      return <RatioScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'table':
      return <TableScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'toggle':
      return <ToggleScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'gauge':
      return <GaugeScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'verdict':
      return <VerdictScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'decision':
      return <DecisionScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'myth':
      return <MythScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'rundown':
      return <RundownScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'analogy':
      return <AnalogyScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'trace':
      return <TraceScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'costing':
      return <CostingScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'constellation':
      return <ConstellationScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'chapter':
      return <ChapterScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'cycle':
      return <CycleScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'scale':
      return <ScaleScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'objective':
      return <ObjectiveScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'prereq':
      return <PrereqScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'recap':
      return <RecapScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'pitfall':
      return <PitfallScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'checkpoint':
      return <CheckpointScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    // The v7 foundations scenes, all on the standard contract.
    case 'syllabus':
      return <SyllabusScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'outcome':
      return <OutcomeScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'bridge':
      return <BridgeScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'drill':
      return <DrillScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'labcard':
      return <LabCardScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'mission':
      return <MissionScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'machine':
      return <MachineScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'blueprint':
      return <BlueprintScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'relay':
      return <RelayScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'layers':
      return <LayersScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'pipeline':
      return <PipelineScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'radix':
      return <RadixScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'carry':
      return <CarryScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'bitfield':
      return <BitfieldScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'encode':
      return <EncodeScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'gates':
      return <GatesScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'ladder':
      return <LadderScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'regions':
      return <RegionsScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'lookup':
      return <LookupScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'states':
      return <StatesScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'scheduler':
      return <SchedulerScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'shell':
      return <ShellScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'journey':
      return <JourneyScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'handshake':
      return <HandshakeScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'stepper':
      return <StepperScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'growth':
      return <GrowthScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'callstack':
      return <CallStackScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'history':
      return <HistoryScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'versus':
      return <VersusScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'eras':
      return <ErasScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'cards':
      return <CardsScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'duel':
      return <DuelScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    case 'spotlight':
      return <SpotlightScene theme={theme} sceneStartMs={scene.startMs} props={scene.props} />;
    default:
      // THROW, do not return null.
      //
      // Returning null was the old behaviour and it is the worst possible one: a
      // scene type this switch does not know drew nothing, so the render
      // SUCCEEDED and produced a blank frame with the voiceover playing over it.
      // Every expensive stage had already run — the plan, the review rounds, the
      // TTS, the alignment — and the only signal that anything was wrong was
      // watching the finished video.
      //
      // It happened for real: five templates were registered in the catalog with
      // their Go half complete and no component here yet, and the first clip cut
      // with one came out white and silent-looking. The pipeline had no way to
      // notice, because "renders nothing" and "renders correctly" are the same
      // outcome to everything downstream of this function.
      //
      // Failing here costs the render and nothing else, and the message names the
      // two things that are actually wrong.
      throw new Error(
        `LessonVideo has no component for scene type "${scene.type}". ` +
          `Either the template's <Name>Scene.tsx has not been written yet, or it exists ` +
          `and was never added to this switch. Refusing to render a blank frame — ` +
          `a scene that draws nothing is indistinguishable from one that drew correctly, ` +
          `so this would otherwise ship as a video with a voiceover over an empty screen.`,
      );
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
const surfaceFor = (scenes: Scene[], skin: ResolvedTheme['skin']): Surface => {
  // A skin that stands on nothing overrides the per-template backdrop outright.
  // The broadcast look is *defined* by an unlit stage — letting a flow scene
  // still bring its blueprint grid along would be picking the skin for the
  // colours and then undoing the composition it exists for.
  if (skin === 'broadcast') return 'void';
  if (skin === 'minimal') return 'clean';
  // The showroom look is a flat sheet with objects sitting on it. Every backdrop
  // this function can return other than 'clean' adds something behind the
  // subject — a grid, a pool of light, a drifting field — and all of them were
  // designed for a dark stage, where a backdrop is what stops the frame being a
  // void. On paper there is no void to fill, and anything drifting behind a card
  // shows THROUGH the shadow that is seating it, which takes the depth back out.
  if (skin === 'showroom') return 'sheet';
  // The title card is the same card whatever follows it, so it does not get a
  // vote — otherwise every snippet would be "mixed" and land on the default.
  const kinds = new Set(scenes.map((s) => s.type).filter((t) => t !== 'title'));
  if (kinds.size !== 1) return 'default';
  switch ([...kinds][0]) {
    case 'whiteboard':
      return 'paper';
    case 'flow':
    // A trace is a system diagram in everything but the axis it varies
    // along, so it stands on the same drawing-office grid the flow does.
    case 'trace':
    // A node-and-edge map earns the drawing-office grid for the same reason
    // the flow does: it makes the radial arrangement read as measured.
    case 'constellation':
    // A ring of stages is a mechanism, and a mechanism drawn on squared paper
    // reads as engineered rather than as decorative.
    case 'cycle':
      return 'blueprint';
    case 'cast':
    case 'story':
    // A break between two stretches of teaching is a held moment, and one pool
    // of light is what a held moment looks like. It is also the only scene in
    // the catalog whose subject is a 380px numeral, which wants a stage that
    // falls away at the edges rather than a field competing with it.
    case 'chapter':
      return 'spotlight';
    case 'data':
    case 'quiz':
    case 'compare':
    case 'anatomy':
    case 'timeline':
    // A figure at 250pt is the only thing that should be bright on the frame.
    // Anything drifting behind it competes with the one mark the clip is about.
    case 'metric':
    // A grid of up to twelve hundred cells is already the busiest frame in the
    // catalog. A field behind it reads as more cells.
    case 'occupancy':
    // Rows travelling across a drifting field read as two things moving. The
    // one movement on this frame has to be the re-sort.
    case 'ranking':
    // A monospace panel on a lit field is a code listing with a glow behind it.
    // The cursor is the only thing here that should catch the eye.
    case 'journal':
    // The frame is mostly empty on purpose — that imbalance is the argument. A
    // field filling the gap would take the emptiness away.
    case 'multiplex':
    // A row of near-identical cells with one lit. A field behind them would put
    // a second brightness gradient across a row whose whole job is being even.
    case 'fork':
    // The boundary already carries a glow, and it is the only thing on the frame
    // that should. A field behind it would blur the line the picture is about.
    case 'capabilities':
    // One bar and one figure. A drifting field behind a bar reads as a second
    // fill, which is the one thing this picture cannot afford.
    case 'budget':
    // A dashed decade grid is a set of thin marks, and a field behind them reads
    // as another gridline. The scale has to stay unambiguous.
    case 'latency':
    // Two figures and a row of glyphs. The product is the brightest thing here by
    // design, and a field behind it would take that away.
    case 'multiply':
    // Two aligned bars whose difference in length IS the content. A gradient
    // behind them would make one end of each bar read as longer than it is.
    case 'ratio':
    // The sheet's fidelity is the setup, and a field behind it would make it look
    // like a designed object rather than something a manufacturer printed.
    case 'table':
    // The knob already carries a glow and it is the one thing that moves. A
    // drifting field would put a second motion on a frame about one decision.
    case 'toggle':
    // A dashed rule and a filling bar are thin marks. A field behind them
    // competes with the one line the whole clip is about.
    case 'gauge':
    // Two columns of type and a closing line. Nothing here wants a field
    // behind it competing for the little ink on the frame.
    case 'verdict':
    // An axis and one line of type. A drifting glow behind the segments
    // would read as a fourth band.
    case 'decision':
    // Two lines of type and a rule through one of them. There is nothing
    // here a patterned field would not fight.
    case 'myth':
    // A row of cards with ghosted numerals behind them. A dot field would
    // read through the card fills as texture on the numbers.
    case 'rundown':
    // Two columns of plates joined by hairlines. A field behind them reads
    // through the untinted rows as noise on the connectors.
    case 'analogy':
    // A sheet of thin bars and a counting figure. Nothing repeating behind
    // it survives contact with twelve-pixel bars.
    case 'costing':
    // Nested frames only read as containment if there is nothing else on the
    // stage with edges. Any repeating field behind them competes with the one
    // thing the picture is doing, and a drifting glow reads as a sixth world.
    case 'scale':
    // The v7 foundations scenes, together. Every one of them is a technical
    // diagram — bit rows, block diagrams, lanes over a time axis, curves on a
    // grid — and each carries its own thin marks and its own one glow. A
    // drifting field behind a technical drawing is the difference between a
    // figure in a textbook and a poster about one, and this family is the
    // textbook.
    case 'syllabus':
    case 'outcome':
    case 'bridge':
    case 'drill':
    case 'labcard':
    case 'mission':
    case 'machine':
    case 'blueprint':
    case 'relay':
    case 'layers':
    case 'pipeline':
    case 'radix':
    case 'carry':
    case 'bitfield':
    case 'encode':
    case 'gates':
    case 'ladder':
    case 'regions':
    case 'lookup':
    case 'states':
    case 'scheduler':
    case 'shell':
    case 'journey':
    case 'handshake':
    case 'stepper':
    case 'growth':
    case 'callstack':
    case 'history':
    case 'versus':
    case 'eras':
    // A row of shadowed cards is already several objects sitting off the stage,
    // and the shadow is what seats them. A drifting field behind them shows
    // through those shadows and takes the depth back out.
    case 'cards':
      return 'clean';
    default:
      return 'default';
  }
};

/**
 * Read a scene prop that is only sometimes a string.
 *
 * Scene props are `Record<string, any>` off the wire, and the chrome slots are
 * optional by design — a template that has nothing to put in the eyebrow omits
 * the key. Coercing whatever turns up would put "[object Object]" in the corner
 * of the frame, so anything that is not a non-empty string reads as absent.
 */
const stringProp = (v: unknown): string | undefined =>
  typeof v === 'string' && v.trim() !== '' ? v : undefined;

/** How this video cuts between beats — same derivation as the surface. */
const cutStyleFor = (scenes: Scene[]): CutStyle => {
  const kinds = new Set(scenes.map((s) => s.type).filter((t) => t !== 'title'));
  if (kinds.size !== 1) return 'rise';
  switch ([...kinds][0]) {
    case 'illustration':
    case 'cast':
      return 'push';
    case 'story':
      return 'cut';
    default:
      return 'rise';
  }
};

export const LessonVideo: React.FC<LessonVideoProps> = (props) => {
  const {assetBase, audioFile, sfxFile, scenes, captions, captionEmphasis, motion} = props;
  const theme = useMemo(() => resolveTheme(props.theme), [props.theme]);
  const surface = useMemo(() => surfaceFor(scenes, theme.skin), [scenes, theme.skin]);
  const cutStyle = useMemo(() => cutStyleFor(scenes), [scenes]);
  // Only the broadcast skin carries standing chrome. `minimal` is defined by
  // its absence, and `default` never had any.
  const chrome = theme.skin === 'broadcast';
  // Scenes cross-dissolve. Every sequence but the last is extended past its
  // scene end by `cross` frames, so the outgoing scene is still mounted and
  // fading while the incoming one fades in on top of it (later siblings paint
  // last). Scene *content* still receives its true duration — only the mount
  // window is stretched, so nothing that paces itself off durationInFrames
  // (the execution viz, the walkthrough, the memory layout) is affected.
  const cross = secondsToFrames(FPS, resolveMotion(motion).timing.normal);
  // Whether a caption card will be drawn over the frame — the same condition
  // CaptionTrack is mounted on below, read once here so it cannot disagree with
  // the space every scene reserves for it. Reserving that band unconditionally
  // was giving up 11% of every frame in every default-configured video to a card
  // that has not rendered since captions became opt-in.
  const hasCaptions = Boolean(captions && captions.length > 0);
  return (
    <StageCaptionsContext.Provider value={hasCaptions}>
    <StageAirContext.Provider value={theme.air}>
    {/* The editorial skin is the only one that composes off the centre line;
        every other skin keeps the placement each scene has always had. */}
    <StageAlignContext.Provider value={theme.skin === 'editorial' ? 'flex-start' : 'center'}>
    {/* flex-start, NOT space-between. space-between was tried first and is
        wrong: it distributes the SCENE'S OWN children, so a budget bar, its
        label and its figure got pushed to opposite ends of the frame with dead
        gaps between them — the composition came apart to fill the space. The
        block stays tight and gets anchored to the top margin instead, which is
        what a left axis wants: a dense corner of type and picture, and the
        emptiness gathered in one place rather than ringed around it. */}
    <StageJustifyContext.Provider value={theme.skin === 'editorial' ? 'flex-start' : undefined}>
      <AbsoluteFill style={{fontFamily: theme.fontBody}}>
        <SceneBackground theme={theme} surface={surface} />
        {audioFile ? <Audio src={staticFile(`${assetBase ?? ''}/${audioFile}`)} /> : null}
      {/* The keystroke track, mixed under the voice. Its level is baked into
          the file (keysound.go) rather than set here, so the one number that
          decides whether this is pleasant or unbearable lives beside the
          synthesis that produced it. */}
      {sfxFile ? <Audio src={staticFile(`${assetBase ?? ''}/${sfxFile}`)} /> : null}
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
              cutStyle={cutStyle}
            >
              {/* The camera wraps the scene's own content and nothing else.
                  Callouts and the standing chrome sit OUTSIDE it deliberately:
                  a callout points at a fixed place on the frame and a watermark
                  is furniture, and drifting either one would look like a bug
                  rather than like a camera. */}
              <SceneCamera durationInFrames={duration} motion={motion} index={i}>
                {sceneContent(scene, props, theme, duration)}
              </SceneCamera>
              {/* Inside the transition so a callout fades out with its own
                  scene instead of hanging fully opaque over the next one. */}
              <CalloutLayer
                theme={theme}
                callouts={scene.callouts ?? []}
                sceneStartMs={scene.startMs}
              />
              {/* Standing chrome, for the skins that carry it. The eyebrow
                  falls back to the scene's kicker, which every existing
                  template already sets — so the whole catalog inherits the
                  treatment without a template being touched. */}
              {chrome ? (
                <SceneChrome
                  theme={theme}
                  eyebrow={stringProp(scene.props.eyebrow) ?? stringProp(scene.props.kicker)}
                  takeaway={stringProp(scene.props.takeaway)}
                />
              ) : null}
            </SectionTransition>
          </Sequence>
        );
      })}
      {/* On-screen captions are opt-in; scene graphs built with
          style.captions off carry no caption words (null/empty). */}
      {hasCaptions ? (
        <CaptionTrack theme={theme} captions={captions} emphasis={captionEmphasis} />
      ) : null}
      {/* Outside the sequences, so it does not cross-dissolve at every cut. A
          watermark that fades with the beats is one the eye keeps noticing. */}
        {chrome ? <Watermark theme={theme} /> : null}
      </AbsoluteFill>
    </StageJustifyContext.Provider>
    </StageAlignContext.Provider>
    </StageAirContext.Provider>
    </StageCaptionsContext.Provider>
  );
};
