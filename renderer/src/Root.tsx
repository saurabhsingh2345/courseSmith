import {Composition, staticFile} from 'remotion';
import {LessonVideo} from './LessonVideo';
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
