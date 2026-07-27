import {Composition, staticFile} from 'remotion';
import {LessonVideo} from './LessonVideo';
import {FigureSheet} from './components/FigureSheet';
import {FPS, LessonVideoProps, CodeTrace, msToFrame} from './types';
import execTrace from './fixtures/execTrace.json';

// A tiny sample so the studio renders something before a lesson is loaded.
const sampleProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
  },
  audioFile: '',
  durationMs: 6000,
  scenes: [
    {
      type: 'title',
      startMs: 0,
      endMs: 3000,
      props: {
        heading: 'Coursesmith Renderer',
        subtitle: 'run `coursesmith preview <lesson>` to load a real lesson',
        intro: true,
        outcomes: ['Animated scenes', 'Self-typing code', 'Word-synced visuals'],
      },
    },
    {
      type: 'code',
      startMs: 3000,
      endMs: 6000,
      props: {
        title: 'Sample',
        code: 'print("Hello from coursesmith!")',
        language: 'python',
        output: 'Hello from coursesmith!\n',
      },
    },
  ],
  captions: [],
};

// A demo of the Python-Tutor execution viz (workstream C) with a real trace,
// so `remotion still ExecViz` / the studio showcase can render it standalone.
const execVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#f6f7f9', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 8000,
  scenes: [
    {
      type: 'code',
      startMs: 0,
      endMs: 8000,
      props: {
        title: 'Watching the code run',
        code: (execTrace as CodeTrace).code,
        language: 'python',
        trace: execTrace as CodeTrace,
      },
    },
  ],
  captions: [],
};

// A demo of the stack/heap MemoryLayout view (workstream C) over the same real
// trace, so `remotion still MemoryViz` renders it standalone.
const memoryVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#f6f7f9', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 8000,
  scenes: [
    {
      type: 'code',
      startMs: 0,
      endMs: 8000,
      props: {
        title: 'Stack & heap',
        code: (execTrace as CodeTrace).code,
        language: 'python',
        trace: execTrace as CodeTrace,
        view: 'memory',
      },
    },
  ],
  captions: [],
};

// A demo of an animated D3 node-link diagram (workstream A), loading a fixture
// spec from public/ so `remotion still D3Viz` renders it standalone.
const d3VizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 6000,
  scenes: [
    {type: 'diagram', startMs: 0, endMs: 6000, props: {src: 'd3demo.json', kind: 'd3', title: ''}},
  ],
  captions: [],
};

// A demo of the storyboard points scene (workstream: no dead screens), so
// `remotion still PointsViz` renders it standalone.
const pointsVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 8000,
  scenes: [
    {
      type: 'points',
      startMs: 0,
      endMs: 8000,
      props: {
        title: 'Why Beginners Start With Python',
        items: [
          {text: 'Reads like plain English', icon: 'book', atMs: 400},
          {text: 'One idea per line', icon: 'idea', atMs: 2200},
          {text: 'Instant feedback', icon: 'zap', atMs: 4000},
          {text: 'Scales to real projects', icon: 'rocket', atMs: 5800},
        ],
      },
    },
  ],
  captions: [],
};


// A demo of the whiteboard scene, so `remotion still WhiteboardViz` renders it
// standalone. Six items with links exercises the 3x2 grid, the connectors, and
// the accent settling back to chalk on everything but the newest box.
const whiteboardVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 14000,
  scenes: [
    {
      type: 'whiteboard',
      startMs: 0,
      endMs: 14000,
      props: {
        title: 'How a web request travels',
        items: [
          {label: 'Browser', icon: 'monitor', atMs: 300},
          {label: 'DNS lookup', icon: 'search', atMs: 2400, from: 0},
          {label: 'CDN edge', icon: 'globe', atMs: 4600, from: 1},
          {label: 'Load balancer', icon: 'layers', atMs: 6800, from: 2},
          {label: 'App server', icon: 'terminal', atMs: 9000, from: 3},
          {label: 'Database', icon: 'database', atMs: 11200, from: 4},
        ],
      },
    },
  ],
  captions: [],
};

// A demo of the flow scene, so `remotion still FlowViz` renders it standalone.
// Branching (the gateway feeds both a cache and a queue) exercises the layering,
// and the focus window at 9s exercises the dim/highlight path.
const flowVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 16000,
  scenes: [
    {
      type: 'flow',
      startMs: 0,
      endMs: 16000,
      props: {
        title: 'How a request gets rate limited',
        ranks: 4,
        nodes: [
          {id: 'client', label: 'Client', kind: 'client', icon: 'monitor', rank: 0, order: 0, atMs: 300},
          {id: 'gw', label: 'API gateway', kind: 'service', icon: 'gear', rank: 1, order: 0, atMs: 2200},
          {id: 'counter', label: 'Rate counter', kind: 'cache', icon: 'zap', rank: 2, order: 0, atMs: 4100},
          {id: 'queue', label: 'Work queue', kind: 'queue', icon: 'layers', rank: 2, order: 1, atMs: 6000},
          {id: 'db', label: 'Postgres', kind: 'store', icon: 'database', rank: 3, order: 0, atMs: 7900},
        ],
        edges: [
          {from: 0, to: 1, atMs: 2200},
          {from: 1, to: 2, atMs: 4100},
          {from: 1, to: 3, atMs: 6000},
          {from: 3, to: 4, atMs: 7900},
        ],
        focus: [{startMs: 10000, endMs: 16000, nodes: [0, 1, 2]}],
      },
    },
  ],
  captions: [],
};

// A demo of the illustration scene, so `remotion still IllustrationViz` renders
// it standalone. Three shots, because this template's whole shape is the cut
// between them: the side alternates, and each beat picks a different figure.
const illustrationVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 15000,
  scenes: [
    {
      type: 'illustration',
      startMs: 0,
      endMs: 5000,
      props: {
        headline: 'Your app is doing the same work twice',
        emphasis: 'twice',
        caption: 'Every request recomputes a result nobody asked to change.',
        figure: 'gears',
        flip: false,
      },
    },
    {
      type: 'illustration',
      startMs: 5000,
      endMs: 10000,
      props: {
        headline: 'A cache remembers the answer',
        emphasis: 'remembers',
        caption: 'Store it once, hand it back for as long as it stays true.',
        figure: 'lightbulb',
        flip: true,
      },
    },
    {
      type: 'illustration',
      startMs: 10000,
      endMs: 15000,
      props: {
        headline: 'Ninety percent fewer queries',
        emphasis: 'Ninety percent',
        caption: 'The database only sees the work that actually changed.',
        figure: 'chart',
        flip: false,
      },
    },
  ],
  captions: [],
};

// A demo of the VS Code walkthrough scene, so `remotion still VSCodeViz`
// renders it standalone.
const vscodeVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 12000,
  scenes: [
    {
      type: 'walkthrough',
      startMs: 0,
      endMs: 12000,
      props: {
        title: 'Build the greeting step by step',
        file: 'greeting.py',
        project: 'python-basics',
        language: 'python',
        files: ['greeting.py', 'variables.py', 'math_ops.py'],
        steps: [
          {code: 'name = "Ada"\nprint("Hello, " + name)', atMs: 0},
          {code: 'name = "Ada"\n\ndef greet(who):\n    return "Hello, " + who\n\nprint(greet(name))', atMs: 5000},
          {code: 'name = "Ada"\n\ndef greet(who, excited=False):\n    suffix = "!" if excited else "."\n    return "Hello, " + who + suffix\n\nprint(greet(name, excited=True))', atMs: 9000},
        ],
      },
    },
  ],
  captions: [],
};

export const RemotionRoot: React.FC = () => {
  return (
    <>
    <Composition
      id="VSCodeViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(vscodeVizProps.durationMs)}
      defaultProps={vscodeVizProps}
    />
    <Composition
      id="WhiteboardViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(whiteboardVizProps.durationMs)}
      defaultProps={whiteboardVizProps}
    />
    <Composition
      id="FlowViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(flowVizProps.durationMs)}
      defaultProps={flowVizProps}
    />
    {/* Development aid, not a scene: every figure in the vocabulary on one
        frame. See FigureSheet.tsx. */}
    <Composition
      id="FigureSheet"
      component={FigureSheet}
      fps={FPS}
      width={1400}
      height={1120}
      durationInFrames={300}
    />
    <Composition
      id="IllustrationViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(illustrationVizProps.durationMs)}
      defaultProps={illustrationVizProps}
    />
    <Composition
      id="PointsViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(pointsVizProps.durationMs)}
      defaultProps={pointsVizProps}
    />
    <Composition
      id="D3Viz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(d3VizProps.durationMs)}
      defaultProps={d3VizProps}
    />
    <Composition
      id="ExecViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(execVizProps.durationMs)}
      defaultProps={execVizProps}
    />
    <Composition
      id="MemoryViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(memoryVizProps.durationMs)}
      defaultProps={memoryVizProps}
    />
    <Composition
      id="LessonVideo"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(sampleProps.durationMs)}
      defaultProps={sampleProps}
      calculateMetadata={async ({props}) => {
        let resolved = props;
        // `coursesmith preview` stages a lesson under public/preview/; when
        // the studio opens with the sample props, load it instead.
        if (!props.assetBase) {
          try {
            const res = await fetch(staticFile('preview/lesson-video.json'));
            if (res.ok) {
              resolved = (await res.json()) as LessonVideoProps;
            }
          } catch {
            // No preview staged — keep the sample.
          }
        }
        return {
          durationInFrames: Math.max(1, msToFrame(resolved.durationMs)),
          props: resolved,
        };
      }}
    />
    </>
  );
};
