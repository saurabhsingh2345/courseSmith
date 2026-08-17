import {Composition, staticFile} from 'remotion';
import {LessonVideo} from './LessonVideo';
import {FigureSheet, CastSheet} from './components/FigureSheet';
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
          // One of each shape, so the baseline covers the whole vocabulary —
          // a shape with no composition is a shape whose regression hides.
          {label: 'You', icon: 'cursor', shape: 'circle', atMs: 300},
          {label: 'The internet', icon: 'cloud', shape: 'cloud', atMs: 2400, from: 0},
          {label: 'CDN edge', icon: 'globe', atMs: 4600, from: 1},
          {label: 'Load balancer', icon: 'network', atMs: 6800, from: 2},
          {label: 'App server', icon: 'terminal', atMs: 9000, from: 3},
          {label: 'Cache it!', icon: 'highlighter', shape: 'sticky', atMs: 11200, from: 4},
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
          {id: 'gw', label: 'API gateway', kind: 'service', icon: 'server', rank: 1, order: 0, atMs: 2200},
          {id: 'counter', label: 'Rate counter', kind: 'cache', icon: 'cache', rank: 2, order: 0, atMs: 4100},
          {id: 'queue', label: 'Work queue', kind: 'queue', icon: 'queue', rank: 2, order: 1, atMs: 6000},
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

// A demo of the spine scene: one shot of each of the twelve layouts, in the
// order a real clip tends to use them.
//
// Twelve rather than a representative three, because this template's whole claim
// is that a single clip can open, explain, turn and close — a fixture showing
// three `state` shots would prove the opposite of what it exists to show. It is
// also the only way to check the rail, which is the one element that has to be
// continuous ACROSS the cut and therefore cannot be judged from one frame, and
// the only place the segment count is exercised at all.
const SPINE_SHOT_MS = 3000;
const spineShots: {shot: string; headline: string; props: Record<string, unknown>}[] = [
  {
    shot: 'open',
    headline: 'You are still clicking',
    props: {
      note: 'PART ONE',
      emphasis: 'clicking',
      caption: 'An hour of it produces exactly one result, and then it is gone.',
      objects: [{figure: 'cursor', label: 'Clicking'}],
    },
  },
  {
    shot: 'chapter',
    headline: 'Part one: the instruction',
    props: {
      note: 'CHAPTER',
      ordinal: 1,
      emphasis: 'instruction',
      caption: 'Three things to get straight before you build anything.',
      objects: [
        {figure: 'prompt', label: 'Writing one'},
        {figure: 'guardrail', label: 'Keeping it honest'},
        {figure: 'automation', label: 'Running it twice'},
      ],
    },
  },
  {
    shot: 'state',
    headline: 'A prompt is an instruction you keep',
    props: {
      emphasis: 'you keep',
      caption: 'Write it once and it runs again tomorrow, unchanged.',
      objects: [{figure: 'prompt', label: 'The prompt'}],
    },
  },
  {
    shot: 'pair',
    headline: 'One is gone, one stays',
    props: {
      emphasis: 'stays',
      caption: 'The difference is not effort. It is whether anything survives the work.',
      objects: [
        {figure: 'cursor', label: 'A click', detail: 'Gone the moment it finishes.'},
        {figure: 'notebook', label: 'A prompt', detail: 'Runs again tomorrow, unchanged.'},
      ],
    },
  },
  {
    shot: 'row',
    headline: 'Three things you get back',
    props: {
      objects: [
        {figure: 'clock', label: 'Time', detail: 'The second run is free.'},
        {figure: 'recycle', label: 'Repeatability', detail: 'Same input, same output.'},
        {figure: 'share', label: 'Handover', detail: 'Somebody else can run it.'},
      ],
    },
  },
  {
    shot: 'orbit',
    headline: 'Everything hangs off one habit',
    props: {
      emphasis: 'one habit',
      caption: 'Write the instruction down. The rest follows from having it.',
      objects: [
        {figure: 'brain', label: 'The habit'},
        {figure: 'code', label: 'Code'},
        {figure: 'checklist', label: 'Reviews'},
        {figure: 'chart', label: 'Reports'},
        {figure: 'envelope', label: 'Email'},
      ],
    },
  },
  {
    shot: 'steps',
    headline: 'How the loop actually runs',
    props: {
      objects: [
        {figure: 'pencil', label: 'Write it', detail: 'One sentence is enough to start.'},
        {figure: 'terminal', label: 'Run it', detail: 'Watch what it actually does.'},
        {figure: 'highlighter', label: 'Fix it', detail: 'Change the words, not the output.'},
      ],
    },
  },
  {
    shot: 'recap',
    headline: 'That is the loop, done',
    props: {
      emphasis: 'done',
      caption: 'Nothing here needed a line of code.',
      objects: [
        {figure: 'blueprint', label: 'Planned', detail: 'You knew what you wanted.'},
        {figure: 'blocks', label: 'Assembled', detail: 'Out of parts that existed.'},
        {figure: 'deploy', label: 'Shipped', detail: 'And other people can use it.'},
      ],
    },
  },
  {
    shot: 'aside',
    headline: 'If you have used a spreadsheet, you already know this',
    props: {
      emphasis: 'already know',
      caption: 'A formula is an instruction you keep too. This is the same habit, aimed somewhere else.',
      objects: [{figure: 'spreadsheet'}],
    },
  },
  {
    shot: 'focus',
    headline: 'Then stop touching it',
    props: {
      emphasis: 'stop',
      caption: 'A prompt you keep editing is a prompt you have not finished thinking about.',
      objects: [{figure: 'lock'}],
    },
  },
  {
    shot: 'quote',
    headline: 'The work you can hand over is the only work that scales',
    props: {
      emphasis: 'hand over',
      caption: 'Everything else is a thing only you can do, forever.',
      objects: [],
    },
  },
  {
    shot: 'close',
    headline: 'Write your first one today',
    props: {
      note: 'NEXT',
      emphasis: 'today',
      caption: 'Open the next lesson and write one',
      objects: [{figure: 'rocket'}],
    },
  },
];

const spineVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: SPINE_SHOT_MS * spineShots.length,
  scenes: spineShots.map((s, i) => ({
    type: 'spine' as const,
    startMs: i * SPINE_SHOT_MS,
    endMs: (i + 1) * SPINE_SHOT_MS,
    props: {
      shot: s.shot,
      headline: s.headline,
      index: i,
      total: spineShots.length,
      ...s.props,
    },
  })),
  captions: [],
};

// The same clip in light mode. Every scene that quietly assumed a dark stage
// looks fine until this frame is rendered — and this one leans on `surface` for
// its tiles and on the accent as *type*, which are the two tokens that flip.
const spineLightProps: LessonVideoProps = {
  ...spineVizProps,
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    mode: 'light',
    bgTop: '#fafbfc',
    bgBottom: '#ebeef4',
    surface: '#ffffff',
    surfaceBorder: '#d3dce3',
    text: '#13222f',
    textMuted: '#4b6071',
    mass: '#8ea2b4',
    ink: '#1c354a',
    accentText: '#8d6e00',
    grain: 0.01,
  },
};

// The same illustration clip in light mode, with the tokens Go derives for
// style.mode: light. It has its own baseline because light mode is the branch
// nobody's default config exercises — every scene that quietly assumed a dark
// stage looks fine until this frame is rendered.
const illustrationLightProps: LessonVideoProps = {
  ...illustrationVizProps,
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    mode: 'light',
    bgTop: '#fafbfc',
    bgBottom: '#ebeef4',
    surface: '#ffffff',
    surfaceBorder: '#d3dce3',
    text: '#13222f',
    textMuted: '#4b6071',
    mass: '#8ea2b4',
    ink: '#1c354a',
    accentText: '#8d6e00',
    grain: 0.01,
  },
  // Captions on, so the panel that used to be a hardcoded near-black slab is
  // actually in frame.
  captions: [
    {word: 'Your', startMs: 200, endMs: 500},
    {word: 'app', startMs: 500, endMs: 800},
    {word: 'is', startMs: 800, endMs: 1000},
    {word: 'doing', startMs: 1000, endMs: 1400},
    {word: 'the', startMs: 1400, endMs: 1600},
    {word: 'same', startMs: 1600, endMs: 2000},
    {word: 'work', startMs: 2000, endMs: 2400},
    {word: 'twice', startMs: 2400, endMs: 3000},
  ],
};

// A demo of the cast scene, so `remotion still CastViz` renders it standalone.
// Four shots, because the pose *change* is the template — a single frame can
// only show that a pose exists, not that the character moves between them.
const castVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 16000,
  scenes: [
    {
      type: 'cast',
      startMs: 0,
      endMs: 4000,
      props: {
        headline: 'Nobody reads a five hundred line diff',
        caption: 'It sits for three days and gets a thumbs up nobody means.',
        pose: 'idle',
        prevPose: 'idle',
        expression: 'concerned',
        flip: false,
      },
    },
    {
      type: 'cast',
      startMs: 4000,
      endMs: 8000,
      props: {
        headline: 'The size is the problem',
        caption: 'Reviewers lose the thread by the second file.',
        pose: 'shrug',
        prevPose: 'defeated',
        expression: 'thinking',
        prop: 'stack',
        flip: true,
      },
    },
    {
      type: 'cast',
      startMs: 8000,
      endMs: 12000,
      props: {
        headline: 'Ship it in slices',
        caption: 'Four small reviews beat one big one on every measure.',
        pose: 'point',
        prevPose: 'think',
        expression: 'neutral',
        prop: 'chart',
        flip: false,
      },
    },
    {
      type: 'cast',
      startMs: 12000,
      endMs: 16000,
      props: {
        headline: 'Faster, and actually reviewed',
        caption: 'Real comments instead of a rubber stamp.',
        pose: 'confident',
        prevPose: 'point',
        expression: 'happy',
        flip: true,
      },
    },
  ],
  captions: [],
};

// A demo of the story scene, so `remotion still StoryViz` renders it standalone.
// Six shots covering every staging and five of the six camera moves, because
// this template's whole claim is that consecutive shots differ — a single frame
// can only prove one of them composes.
const storyVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 30000,
  scenes: [
    {
      type: 'story',
      startMs: 0,
      endMs: 5000,
      props: {
        headline: 'Your query reads every single row',
        caption: 'Ten million rows, one at a time.',
        staging: 'duo',
        camera: 'push',
        pose: 'idle',
        prevPose: 'idle',
        expression: 'concerned',
        prop: 'stack',
        durationMs: 5000,
      },
    },
    {
      type: 'story',
      startMs: 5000,
      endMs: 10000,
      props: {
        headline: "Scanning is fine until it isn't",
        caption: 'It gets slower exactly as fast as you grow.',
        staging: 'object',
        camera: 'hold',
        prevPose: 'defeated',
        prop: 'clock',
        durationMs: 5000,
      },
    },
    {
      type: 'story',
      startMs: 10000,
      endMs: 15000,
      props: {
        headline: 'An index is a sorted copy',
        caption: 'One column, kept in order, nothing more.',
        staging: 'hero',
        camera: 'push',
        pose: 'point',
        prevPose: 'defeated',
        expression: 'neutral',
        durationMs: 5000,
      },
    },
    {
      type: 'story',
      startMs: 15000,
      endMs: 20000,
      props: {
        headline: 'Sorted means you can skip',
        caption: 'Halve the search space, then halve it again.',
        staging: 'pair',
        camera: 'pan',
        prevPose: 'point',
        prop: 'chart',
        propB: 'network',
        durationMs: 5000,
      },
    },
    {
      type: 'story',
      startMs: 20000,
      endMs: 25000,
      props: {
        headline: 'Every write pays for it',
        caption: 'The order has to be maintained on the way in.',
        staging: 'duo',
        camera: 'drift',
        pose: 'shrug',
        prevPose: 'point',
        expression: 'thinking',
        prop: 'gears',
        durationMs: 5000,
      },
    },
    {
      type: 'story',
      startMs: 25000,
      endMs: 30000,
      props: {
        headline: 'One line, a thousand times faster',
        caption: 'You are not searching harder. You are searching less.',
        staging: 'empty',
        camera: 'pull',
        prevPose: 'think',
        durationMs: 5000,
      },
    },
  ],
  captions: [],
};

// A demo of the data scene, so `remotion still DataViz` renders it standalone.
//
// Every kind in the vocabulary, six seconds each, rather than a hand-picked
// few. Thirteen kinds share one context object and one idea of what a
// highlight looks like (DataScene.tsx), and the way that stops being true is
// one kind quietly growing its own layout — which nothing catches except
// looking at all of them next to each other.
const dataChartDemos: {kind: string; title: string; unit: string; series?: string[]; points: {label: string; value?: number; values?: number[]}[]}[] = [
  {
    kind: 'bars',
    title: 'Where the time actually goes',
    unit: 'ms',
    points: [
      {label: 'Database', value: 412},
      {label: 'Template render', value: 96},
      {label: 'Auth check', value: 41},
      {label: 'Serialization', value: 28},
      {label: 'Routing', value: 9},
    ],
  },
  {
    kind: 'stackedbars',
    title: 'What each request spends',
    unit: 'ms',
    series: ['Database', 'Render', 'Network'],
    points: [
      {label: 'Search', values: [310, 84, 40]},
      {label: 'Checkout', values: [120, 210, 66]},
      {label: 'Home', values: [40, 66, 22]},
      {label: 'Profile', values: [88, 40, 18]},
    ],
  },
  {
    kind: 'groupedbars',
    title: 'Before and after the cache',
    unit: 'ms',
    series: ['Before', 'After'],
    points: [
      {label: 'Search', values: [310, 96]},
      {label: 'Checkout', values: [420, 180]},
      {label: 'Home', values: [140, 44]},
    ],
  },
  {
    kind: 'line',
    title: 'Build time, release by release',
    unit: 's',
    points: [
      {label: 'v1.0', value: 42},
      {label: 'v1.4', value: 61},
      {label: 'v2.0', value: 128},
      {label: 'v2.3', value: 96},
      {label: 'v3.0', value: 38},
    ],
  },
  {
    kind: 'area',
    title: 'Storage used, month by month',
    unit: 'GB',
    points: [
      {label: 'Jan', value: 120},
      {label: 'Mar', value: 180},
      {label: 'May', value: 340},
      {label: 'Jul', value: 520},
      {label: 'Sep', value: 610},
    ],
  },
  {
    kind: 'scatter',
    title: 'Team size against shipping',
    unit: '',
    series: ['Team size', 'Deploys per week'],
    points: [
      {label: 'Payments', values: [6, 22]},
      {label: 'Search', values: [14, 9]},
      {label: 'Growth', values: [4, 31]},
      {label: 'Platform', values: [11, 14]},
      {label: 'Mobile', values: [8, 6]},
    ],
  },
  {
    kind: 'donut',
    title: 'What the bundle is made of',
    unit: 'kB',
    points: [
      {label: 'Dependencies', value: 540},
      {label: 'Application', value: 180},
      {label: 'Polyfills', value: 96},
      {label: 'Styles', value: 44},
    ],
  },
  {
    kind: 'waffle',
    title: 'Who finishes the tutorial',
    unit: '%',
    points: [
      {label: 'Finished', value: 34},
      {label: 'Stopped early', value: 47},
      {label: 'Never started', value: 19},
    ],
  },
  {
    kind: 'gauge',
    title: 'How full each disk is',
    unit: '%',
    points: [
      {label: 'Primary', value: 88},
      {label: 'Replica', value: 61},
      {label: 'Archive', value: 24},
    ],
  },
  {
    kind: 'treemap',
    title: 'Every megabyte we ship',
    unit: 'kB',
    points: [
      {label: 'React', value: 142},
      {label: 'Charts', value: 96},
      {label: 'Icons', value: 62},
      {label: 'Date maths', value: 44},
      {label: 'Analytics', value: 28},
      {label: 'Our code', value: 210},
    ],
  },
  {
    kind: 'funnel',
    title: 'From visit to purchase',
    unit: '',
    points: [
      {label: 'Visited', value: 10000},
      {label: 'Signed up', value: 3200},
      {label: 'Added to cart', value: 940},
      {label: 'Paid', value: 310},
    ],
  },
  {
    kind: 'kpi',
    title: 'The quarter in three numbers',
    unit: '',
    points: [
      {label: 'Deploys per day', value: 41},
      {label: 'Median review, hours', value: 3.5},
      {label: 'Rollbacks', value: 2},
    ],
  },
  {
    kind: 'map',
    title: 'Where the cables land',
    unit: '',
    points: [
      {label: 'United States of America', value: 88},
      {label: 'United Kingdom', value: 61},
      {label: 'Japan', value: 47},
      {label: 'Brazil', value: 26},
      {label: 'India', value: 35},
      {label: 'Australia', value: 18},
    ],
  },
];

const DATA_DEMO_MS = 6000;

/**
 * The quiz template. The frame chosen for its baseline is an `explain` beat,
 * because that is the only state where every part of the scene is on screen at
 * once: the question, the revealed answer, a dimmed distractor, and the
 * explanation under the option it belongs to.
 */
const quizVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 26000,
  scenes: [
    {
      type: 'quiz',
      startMs: 0,
      endMs: 26000,
      props: {
        title: 'What does len() really count?',
        question: 'What does len() return for [[1, 2], [3, 4, 5]]?',
        options: ['2', '5', '7', 'TypeError'],
        answer: 0,
        why: [
          'len() counts the top level, and there are two lists inside it.',
          'This counts every number instead — what you would get after flattening.',
          'This adds the two inner counts together, which no list operation does.',
          'Nested lists are perfectly valid, so nothing raises here.',
        ],
        steps: [
          {startMs: 0, endMs: 5000, show: 'ask'},
          {startMs: 5000, endMs: 11000, show: 'think'},
          {startMs: 11000, endMs: 16000, show: 'reveal'},
          {startMs: 16000, endMs: 21000, show: 'explain', option: 1},
          {startMs: 21000, endMs: 26000, show: 'explain', option: 2},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The compare template, baselined on its verdict beat — the only state where
 * the winner is marked, the loser has receded and the verdict line is up, so a
 * regression in any of the three shows.
 */
const compareVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 22000,
  scenes: [
    {
      type: 'compare',
      startMs: 0,
      endMs: 22000,
      props: {
        title: 'Two ways to build a list',
        language: 'python',
        left: {
          label: 'A for loop',
          code: 'out = []\nfor x in xs:\n    out.append(x * 2)',
          note: '3 lines, one mutation',
        },
        right: {
          label: 'A comprehension',
          code: 'out = [x * 2 for x in xs]',
          note: '1 line, no mutation',
        },
        winner: 'right',
        verdict: 'When the loop only builds a list, the comprehension says so in one line.',
        steps: [
          {startMs: 0, endMs: 5000, show: 'left'},
          {startMs: 5000, endMs: 10000, show: 'right'},
          {startMs: 10000, endMs: 16000, show: 'both'},
          {startMs: 16000, endMs: 22000, show: 'verdict'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The anatomy template, baselined on a part beat — the state that exercises the
 * lit run, the dimmed remainder, the callout line and the note at once.
 */
const anatomyVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 24000,
  scenes: [
    {
      type: 'anatomy',
      startMs: 0,
      endMs: 24000,
      props: {
        title: 'Every part of a function signature',
        subject: 'def greet(name, excited=False) -> str:',
        // Spans are what Go emits, resolved from each part's quoted text.
        parts: [
          {label: 'the keyword', note: 'Tells Python a function is being defined here.', start: 0, end: 3},
          {label: 'the name', note: 'What you call it later, and what a traceback shows.', start: 4, end: 9},
          {label: 'the parameters', note: 'One required, one with a default that makes it optional.', start: 10, end: 29},
          {label: 'the return type', note: 'A hint for readers and tools; Python does not enforce it.', start: 31, end: 37},
        ],
        steps: [
          {startMs: 0, endMs: 4000, whole: true},
          {startMs: 4000, endMs: 8000, part: 0},
          {startMs: 8000, endMs: 12000, part: 1},
          {startMs: 12000, endMs: 16000, part: 2},
          {startMs: 16000, endMs: 20000, part: 3},
          {startMs: 20000, endMs: 24000, whole: true},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The timeline template, baselined mid-walk: the spine part-filled, the current
 * stop enlarged with its note up, and the stops still ahead visible but faded —
 * which is the state that proves the future is drawn rather than revealed.
 */
const timelineVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 28000,
  scenes: [
    {
      type: 'timeline',
      startMs: 0,
      endMs: 28000,
      props: {
        title: 'What happens when you press enter',
        milestones: [
          {mark: '0ms', title: 'You press enter', note: 'The browser has a string and nothing else yet.', figure: 'cursor'},
          {mark: '2ms', title: 'DNS lookup', note: 'The name becomes an address, usually straight from a cache.', figure: 'search'},
          {mark: '30ms', title: 'TCP and TLS', note: 'A connection opens and both sides agree how to encrypt it.', figure: 'lock'},
          {mark: '80ms', title: 'The server answers', note: 'Your request finally reaches something that can reply.', figure: 'server'},
          {mark: '140ms', title: 'First paint', note: 'Enough HTML has arrived for the browser to draw something.', figure: 'monitor'},
        ],
        steps: [
          {startMs: 0, endMs: 4000, at: 0},
          {startMs: 4000, endMs: 9000, at: 1},
          {startMs: 9000, endMs: 14000, at: 2},
          {startMs: 14000, endMs: 19000, at: 3},
          {startMs: 19000, endMs: 24000, at: 4},
          {startMs: 24000, endMs: 28000, whole: true},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The canvas template. Four cards, walked one at a time and then fired.
 *
 * The baseline frame is taken during the run rather than while building,
 * because that is the state no other frame can stand in for: the token over a
 * card, the ticks it has already left behind, the wire live behind it and dark
 * ahead. A frame on a build beat proves the layout and nothing about the payoff.
 */
const canvasVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 30000,
  scenes: [
    {
      type: 'canvas',
      startMs: 0,
      endMs: 30000,
      props: {
        title: 'From form to spreadsheet to Slack',
        payload: 'New signup',
        nodes: [
          {
            app: 'Typeform',
            title: 'Someone submits the form',
            kind: 'trigger',
            icon: 'zap',
            note: 'Nothing runs until this happens — the whole chain waits here.',
          },
          {
            app: 'Make',
            title: 'Check the plan field',
            kind: 'filter',
            icon: 'filter',
            note: 'Free-tier signups stop here; only paid ones carry on.',
          },
          {
            app: 'Sheets',
            title: 'Append a row',
            kind: 'action',
            icon: 'database',
            note: 'One row per signup, written the moment it arrives.',
          },
          {
            app: 'Slack',
            title: 'Post to the sales channel',
            kind: 'output',
            icon: 'message',
            note: 'The team sees it before the customer has closed the tab.',
          },
        ],
        steps: [
          {startMs: 0, endMs: 5000, at: 0},
          {startMs: 5000, endMs: 10000, at: 1},
          {startMs: 10000, endMs: 15000, at: 2},
          {startMs: 15000, endMs: 20000, at: 3},
          {startMs: 20000, endMs: 30000, run: true},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The promptloop template. Two rounds: an attempt that falls short, a prompt
 * that names the specific gap, and an attempt that closes it.
 *
 * The baseline frame is on the *second* answer, which is the only state that
 * shows the template's whole argument at once: three turns in the thread, the
 * attempt counter past one, and the goal bar reaching further than it did.
 */
const promptLoopVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 32000,
  scenes: [
    {
      type: 'promptloop',
      startMs: 0,
      endMs: 32000,
      props: {
        title: 'Prompting your way to a landing page',
        goal: 'A landing page with a working signup form',
        turns: [
          {
            who: 'you',
            text: 'Build me a landing page for a note-taking app with a signup form.',
            startMs: 0,
            endMs: 8000,
          },
          {
            who: 'ai',
            text: 'Built a hero, a feature grid and a form.',
            startMs: 8000,
            endMs: 16000,
            attempt: 1,
            status: 'partial',
            changes: ['Hero and headline in place', 'Form has no validation', 'Nothing happens on submit'],
          },
          {
            who: 'you',
            text: 'The form does nothing. Validate the email and show a success message.',
            startMs: 16000,
            endMs: 24000,
          },
          {
            who: 'ai',
            text: 'Added validation and a confirmation state.',
            startMs: 24000,
            endMs: 32000,
            attempt: 2,
            status: 'ok',
            changes: ['Email checked before submit', 'Success message replaces the form'],
          },
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The mockup template. A five-block signup page, built downward.
 *
 * Five is deliberately the ceiling rather than a typical page: it is the only
 * fixture that exercises the fit, where the stack is taller than the viewport
 * and every block scales rather than the footer falling off the bottom.
 */
const mockupVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 34000,
  scenes: [
    {
      type: 'mockup',
      startMs: 0,
      endMs: 34000,
      props: {
        title: 'A signup page, one block at a time',
        device: 'browser',
        screen: 'Signup',
        blocks: [
          {kind: 'header', label: 'Nav bar', text: 'Notely', note: 'A logo and two links — anything more is a reason to leave.'},
          {
            kind: 'hero',
            label: 'Hero',
            text: 'Notes that find themselves',
            note: 'One promise, big, above the fold, before anyone scrolls.',
          },
          {kind: 'grid', label: 'Feature row', note: 'Three reasons, no paragraphs — nobody reads paragraphs here.'},
          {kind: 'input', label: 'Email field', text: 'you@work.com', note: 'One field. Every extra one costs you signups.'},
          {kind: 'button', label: 'Signup button', text: 'Start free', note: 'The verb says what happens next, not the word submit.'},
        ],
        steps: [
          {startMs: 0, endMs: 5500, at: 0},
          {startMs: 5500, endMs: 11000, at: 1},
          {startMs: 11000, endMs: 17000, at: 2},
          {startMs: 17000, endMs: 23000, at: 3},
          {startMs: 23000, endMs: 29000, at: 4},
          {startMs: 29000, endMs: 34000, whole: true},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The stack template at its widest: four tiers, three of them holding two tools
 * so the side-by-side "these are alternatives" reading is exercised.
 */
const stackVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 30000,
  scenes: [
    {
      type: 'stack',
      startMs: 0,
      endMs: 30000,
      props: {
        title: 'The four tools behind a no-code job board',
        layers: [
          {
            name: 'Frontend',
            role: 'What the person actually looks at',
            tools: [
              {name: 'Softr', icon: 'monitor', note: 'Fastest if your data is Airtable'},
              {name: 'Framer', icon: 'layers', note: 'Better when design matters more'},
            ],
          },
          {
            name: 'Automation',
            role: 'The glue between everything else',
            tools: [
              {name: 'Make', icon: 'shuffle', note: 'Visual, cheap, forgiving'},
              {name: 'n8n', icon: 'network', note: 'Self-host it when volume grows'},
            ],
          },
          {
            name: 'Data',
            role: 'Where the records actually live',
            tools: [{name: 'Airtable', icon: 'database', note: 'A database that looks like a spreadsheet'}],
          },
          {
            name: 'AI',
            role: 'The judgement you would otherwise do yourself',
            tools: [
              {name: 'OpenAI', icon: 'brain', note: 'Summarising and tagging listings'},
              {name: 'Claude', icon: 'sparkles', note: 'Longer documents, fewer mistakes'},
            ],
          },
        ],
        steps: [
          {startMs: 0, endMs: 5500, at: 0},
          {startMs: 5500, endMs: 11500, at: 1},
          {startMs: 11500, endMs: 17500, at: 2},
          {startMs: 17500, endMs: 24000, at: 3},
          {startMs: 24000, endMs: 30000, whole: true},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The spec template, with one criterion deliberately missed — the case the
 * template exists for and the only one where the crossed box, the struck text
 * and a verdict short of the total are all on screen.
 */
const specVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 32000,
  scenes: [
    {
      type: 'spec',
      startMs: 0,
      endMs: 32000,
      props: {
        title: 'Write the test before you write the prompt',
        goal: 'A signup form that actually converts',
        constraints: ['No backend', 'Ship today'],
        criteria: [
          {text: 'One field, nothing else', note: 'Every extra field costs you signups, so the count is the spec.'},
          {
            text: 'Invalid email caught before submit',
            note: 'Caught in the browser, not after a round trip.',
          },
          {
            text: 'Success message replaces the form',
            status: 'missed',
            note: 'Nobody should be left wondering whether it worked.',
          },
          {text: 'Readable on a phone', note: 'Most of this traffic will never see a laptop.'},
        ],
        steps: [
          {startMs: 0, endMs: 5000, at: 0},
          {startMs: 5000, endMs: 10500, at: 1},
          {startMs: 10500, endMs: 16000, at: 2},
          {startMs: 16000, endMs: 21500, at: 3},
          {startMs: 21500, endMs: 32000, check: true},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The showcase template at full width: four decision cells, three strengths,
 * two limits, and the hand-off. Eight beats — the shape that needed the beat
 * ceiling raised past seven.
 */
/**
 * The metric template on the example that motivated it: the memory arithmetic
 * behind running a large model locally, four figures and a recap.
 *
 * Rendered under the broadcast skin, because this template was designed for it
 * — a figure at 250pt on an unlit stage is the composition, and showing it on
 * the default backdrop would be a preview of a clip nobody would cut that way.
 * It is also the catalog's only skinned baseline, which makes it the guard on
 * the skin itself: the chrome, the semantic accents and the stage scale all
 * regress here or nowhere.
 */
const metricVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    mode: 'dark',
    skin: 'broadcast',
    air: 0.06,
    watermark: '<coursesmith>',
    bgTop: '#0a0c0d',
    bgBottom: '#060708',
    surface: '#16181b',
    surfaceBorder: '#2e3338',
    text: '#fafafa',
    textMuted: '#989fa4',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
    mass: '#dee1e3',
    ink: '#090d11',
    accentText: '#ffd43b',
    grain: 0.02,
  },
  audioFile: '',
  durationMs: 40000,
  scenes: [
    {
      type: 'metric',
      startMs: 0,
      endMs: 40000,
      props: {
        title: 'What a 70B model actually costs to run',
        figures: [
          {
            value: '70',
            unit: 'B params',
            label: 'The model you want to run',
            note: 'Every parameter has to be in memory before a token comes out',
            role: 'quantity',
            countsUp: true,
          },
          {
            value: '140',
            unit: 'GB',
            label: 'Memory it needs at 16-bit',
            note: 'Two bytes a parameter, before you fit one conversation in',
            role: 'quantity',
            countsUp: true,
          },
          {
            value: '24',
            unit: 'GB',
            label: 'What a 4090 actually has',
            note: 'The card everyone recommends holds a sixth of it',
            role: 'limit',
            countsUp: true,
          },
          {
            value: '6',
            unit: 'cards',
            label: 'What it would take',
            note: 'At which point the power supply is the cheap part',
            role: 'limit',
            countsUp: true,
          },
        ],
        steps: [
          {startMs: 0, endMs: 8000, show: 'state', at: 0},
          {startMs: 8000, endMs: 16000, show: 'state', at: 1},
          {startMs: 16000, endMs: 24000, show: 'state', at: 2},
          {startMs: 24000, endMs: 32000, show: 'state', at: 3},
          {startMs: 32000, endMs: 40000, show: 'recap'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The gauge template on the example that motivated it: which models fit in one
 * card's memory. Rendered on the default backdrop rather than the broadcast
 * skin, deliberately — `metric` is the skinned baseline, and having this one
 * unskinned proves the new templates read on the look the catalog already had.
 */
const gaugeVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    // The semantic accents as Go derives them (videoskin.go). Carried
    // explicitly because this template is *about* the distinction between the
    // quantity and the limit: without them resolveTheme falls both back to the
    // brand accent, and a bar that overruns its ceiling renders the same gold
    // as one that clears it — which is the one thing the picture must never do.
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 40000,
  scenes: [
    {
      type: 'gauge',
      startMs: 0,
      endMs: 40000,
      props: {
        title: 'Which models actually fit in 24GB',
        unit: 'GB',
        ceiling: {
          value: 24,
          label: 'What a 4090 holds',
          note: 'And a couple of gigabytes of that is already spoken for',
          // 24 / (26 * 1.08)
          frac: 0.8547,
        },
        bars: [
          {
            label: '7B at 16-bit',
            value: 14,
            note: 'Comfortable, with room for a long conversation',
            fits: true,
            frac: 0.4986,
          },
          {
            label: '13B at 16-bit',
            value: 26,
            note: 'Two gigabytes over, which is the same as not fitting',
            fits: false,
            frac: 0.9259,
          },
          {
            label: '13B quantised to 4-bit',
            value: 8,
            note: 'The same model, a third of the memory, slightly worse answers',
            fits: true,
            frac: 0.2849,
          },
        ],
        steps: [
          {startMs: 0, endMs: 8000, show: 'ceiling'},
          {startMs: 8000, endMs: 16000, show: 'bar', at: 0},
          {startMs: 16000, endMs: 24000, show: 'bar', at: 1},
          {startMs: 24000, endMs: 32000, show: 'bar', at: 2},
          {startMs: 32000, endMs: 40000, show: 'verdict'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The occupancy template on the clip that motivated it: a mixture-of-experts
 * model where sixteen of eight hundred and ninety-six experts run per token.
 *
 * The population is deliberately near the top of the template's range, because
 * the cell geometry is the part most likely to break: at 896 cells the squares
 * are a few pixels across and the gap is derived rather than fixed. A fixture at
 * twenty cells would prove nothing about the case the template exists for.
 *
 * The skin is `broadcast` because this batch assumes it — the grid is composed
 * for a near-black stage with air around it, and the baseline should be of the
 * thing people will actually render.
 */
const occupancyVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    // Carried explicitly for the same reason as the gauge: the band's role is
    // the whole argument, and without the semantic accents an "active" band
    // renders the same brand gold as a "limit" one.
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 30000,
  scenes: [
    {
      type: 'occupancy',
      startMs: 0,
      endMs: 30000,
      props: {
        title: '16 of 896 experts do the work',
        emphasis: '16',
        emphasisRole: 'quantity',
        total: 896,
        unit: 'expert',
        label: "the model's experts",
        // As occupancyGridShape derives them: round(sqrt(896 * 16/9)) = 40.
        cols: 40,
        rows: 23,
        bands: [
          {
            count: 16,
            from: 0,
            label: 'Active this token',
            note: 'Every other one sits idle while still holding memory',
            role: 'quantity',
          },
          {
            count: 120,
            from: 16,
            label: 'Warm in cache',
            note: 'Recently used, so still resident and still costing you',
            role: 'neutral',
          },
        ],
        steps: [
          {startMs: 0, endMs: 9000, show: 'grid'},
          {startMs: 9000, endMs: 17000, show: 'fill', at: 0},
          {startMs: 17000, endMs: 24000, show: 'fill', at: 1},
          {startMs: 24000, endMs: 30000, show: 'read'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The ranking template on the clip that motivated it: one write into a sorted
 * set, and the board re-sorting around it.
 *
 * Two arrivals rather than one, because the interesting case is the SECOND —
 * the board is already full, so a row landing inside it pushes the bottom row
 * off, and the exit is the half of the picture a single arrival never exercises.
 */
const rankingVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 32000,
  scenes: [
    {
      type: 'ranking',
      startMs: 0,
      endMs: 32000,
      props: {
        title: 'One write, and the board re-sorts',
        emphasis: 'One write',
        emphasisRole: 'quantity',
        metric: 'score',
        entries: [
          {label: 'shadowwolf', value: 9842, role: 'neutral', arrival: false},
          {label: 'neon_blade', value: 9610, role: 'neutral', arrival: false},
          {label: 'kira_07', value: 9354, role: 'neutral', arrival: false},
          {label: 'mochi', value: 9128, role: 'neutral', arrival: false},
          {label: 'vortex', value: 8901, role: 'neutral', arrival: false},
          {
            label: 'phoenix',
            value: 9501,
            note: 'The write and the re-sort are one operation, not two',
            role: 'quantity',
            arrival: true,
          },
          {
            label: 'ghoststep',
            value: 9990,
            note: 'Straight to the top, and mochi drops off the board',
            role: 'quantity',
            arrival: true,
          },
        ],
        steps: [
          {startMs: 0, endMs: 9000, show: 'board', order: [0, 1, 2, 3, 4]},
          {startMs: 9000, endMs: 17000, show: 'insert', at: 0, entered: 5, order: [0, 1, 5, 2, 3]},
          {startMs: 17000, endMs: 26000, show: 'insert', at: 1, entered: 6, order: [6, 0, 1, 5, 2]},
          {startMs: 26000, endMs: 32000, show: 'read', order: [6, 0, 1, 5, 2]},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The journal template on the clip that motivated it: an append-only file that
 * rebuilds a dataset when it is read back.
 *
 * The fixture stops the replay on the DELETE rather than on a write, because the
 * delete is the line that makes the point — an append-only log records the
 * removal as another entry, and a viewer who has only seen SETs replay has not
 * understood why the file is the source of truth.
 */
const journalVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 44000,
  scenes: [
    {
      type: 'journal',
      startMs: 0,
      endMs: 44000,
      props: {
        title: 'The file that rebuilds your database',
        emphasis: 'rebuilds',
        emphasisRole: 'quantity',
        file: 'appendonly.aof',
        writeLabel: 'appending',
        replayLabel: 'replaying — top to bottom',
        entries: [
          {text: 'SET user:42 alice', note: 'The first write anyone made', role: 'neutral'},
          {text: 'SET cart:42 [items]', role: 'neutral'},
          {text: 'INCR visits', role: 'neutral'},
          {text: 'DEL cart:42', note: 'The delete is a record too, not an erasure', role: 'limit'},
          {text: 'SET score:42 1200', role: 'neutral'},
        ],
        steps: [
          {startMs: 0, endMs: 7000, show: 'file', written: 0},
          {startMs: 7000, endMs: 13000, show: 'append', at: 0, written: 1},
          {startMs: 13000, endMs: 19000, show: 'append', at: 1, written: 2},
          {startMs: 19000, endMs: 25000, show: 'append', at: 2, written: 3},
          {startMs: 25000, endMs: 31000, show: 'append', at: 3, written: 4},
          {startMs: 31000, endMs: 36000, show: 'replay', at: 0, written: 4},
          {startMs: 36000, endMs: 41000, show: 'replay', at: 3, written: 4},
          {startMs: 41000, endMs: 44000, show: 'read', written: 4},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The multiplex template on the clip that motivated it: one thread serving a
 * pool of sockets.
 *
 * The fixture's second round wakes three sources rather than one, because that
 * is the state the template exists for and the one a single-round fixture would
 * never exercise — one ready socket draws the same picture as polling.
 */
const multiplexVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 30000,
  scenes: [
    {
      type: 'multiplex',
      startMs: 0,
      endMs: 30000,
      props: {
        title: 'One thread, a hundred thousand clients',
        emphasis: 'One thread',
        emphasisRole: 'quantity',
        sourceKind: 'sockets',
        worker: 'epoll',
        workerNote: '1 thread',
        sources: [
          {label: '#00428'},
          {label: '#00429'},
          {label: '#00430'},
          {label: '#00431'},
          {label: '#00432'},
          {label: '#00433'},
          {label: '#00434'},
          {label: '#00435'},
        ],
        steps: [
          {startMs: 0, endMs: 8000, show: 'pool'},
          {
            startMs: 8000,
            endMs: 15000,
            show: 'round',
            at: 0,
            ready: [1],
            note: 'One socket has data, so one gets handled',
            role: 'neutral',
          },
          {
            startMs: 15000,
            endMs: 24000,
            show: 'round',
            at: 1,
            ready: [4, 6, 7],
            note: 'Three woke together, and one thread took all three',
            role: 'quantity',
          },
          {startMs: 24000, endMs: 30000, show: 'read'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The fork template on the clip that motivated it: a background snapshot that
 * never pauses the database.
 *
 * Six pages and one write, deliberately. The fixture has to prove the half of
 * the picture that is easy to lose — five pages still shared, at full strength,
 * beside the one that diverged.
 */
const forkVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 26000,
  scenes: [
    {
      type: 'fork',
      startMs: 0,
      endMs: 26000,
      props: {
        title: 'A snapshot that never pauses the database',
        emphasis: 'never pauses',
        emphasisRole: 'quantity',
        origin: 'redis-server',
        originNote: 'pid 4271',
        parent: 'parent',
        child: 'child',
        pages: [
          {label: 'user:42'},
          {label: 'cart:42'},
          {label: 'session:7'},
          {label: 'score:42'},
          {label: 'keys G-M'},
          {label: 'keys N-Z'},
        ],
        steps: [
          {startMs: 0, endMs: 9000, show: 'shared', copied: {}},
          {
            startMs: 9000,
            endMs: 19000,
            show: 'write',
            at: 1,
            by: 'parent',
            note: 'The child still sees the old value, and nothing else moved',
            role: 'quantity',
            copied: {'1': 'parent'},
          },
          {
            startMs: 19000,
            endMs: 26000,
            show: 'read',
            note: 'Five of the six pages are still one page',
            copied: {'1': 'parent'},
          },
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The capabilities template on the clip that motivated it: a WebAssembly module
 * that cannot open a file unless the host passes one in.
 *
 * Four capabilities and one grant, deliberately. The fixture has to hold both
 * halves of the rule at once — something across the line and something still
 * refused — because a frame with all four denied is a wall and a frame with all
 * four granted is the host, and neither is what the template draws.
 */
const capabilitiesVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 28000,
  scenes: [
    {
      type: 'capabilities',
      startMs: 0,
      endMs: 28000,
      props: {
        title: 'It cannot open a file unless you let it',
        emphasis: 'unless you let it',
        emphasisRole: 'quantity',
        subject: 'WASM module',
        subjectNote: 'app.wasm',
        boundary: 'zero default access',
        granter: 'the host',
        items: [
          {label: 'files', note: 'Handed in as one directory, not the disk', role: 'quantity'},
          {label: 'network', note: 'So a bad module cannot phone home, even if it wants to', role: 'limit'},
          {label: 'the clock', note: 'Denied, which is why timing attacks get harder', role: 'limit'},
          {label: 'random', note: 'Still shut unless the host passes a source', role: 'neutral'},
        ],
        steps: [
          {startMs: 0, endMs: 10000, show: 'sealed', granted: []},
          {startMs: 10000, endMs: 20000, show: 'grant', at: 0, granted: [0]},
          {startMs: 20000, endMs: 28000, show: 'read', granted: [0]},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The budget template on the clip that motivated it: what is actually left of a
 * card's memory once the weights, the runtime and the cache are paid for.
 *
 * The numbers are chosen so the remainder is small but POSITIVE. A busting
 * budget is the more dramatic frame and the easier one to draw; a budget that
 * only just survives is the case where the segment widths and the remainder gap
 * all have to be right, so it is the one worth holding a baseline on.
 */
const budgetVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 36000,
  scenes: [
    {
      type: 'budget',
      startMs: 0,
      endMs: 36000,
      props: {
        title: 'What is really left of 24GB',
        emphasis: 'really left',
        emphasisRole: 'quantity',
        pot: 24,
        unit: 'GB',
        potLabel: 'what a 4090 holds',
        remainderLabel: 'left for your context',
        remainder: 1.5,
        remainderFrac: 0.0625,
        claims: [
          {
            amount: 14,
            label: 'the model weights',
            note: 'A 7B at 16-bit, before anything runs',
            role: 'neutral',
            frac: 0.5833,
          },
          {
            amount: 2.5,
            label: 'CUDA and the driver',
            note: 'Gone before your code starts',
            role: 'neutral',
            frac: 0.1042,
          },
          {
            amount: 6,
            label: 'the KV cache at 8k',
            note: 'And it grows with every token you add',
            role: 'limit',
            frac: 0.25,
          },
        ],
        steps: [
          {startMs: 0, endMs: 7000, show: 'pot', taken: [], left: 24},
          {startMs: 7000, endMs: 14000, show: 'claim', at: 0, taken: [0], left: 10},
          {startMs: 14000, endMs: 21000, show: 'claim', at: 1, taken: [0, 1], left: 7.5},
          {startMs: 21000, endMs: 29000, show: 'claim', at: 2, taken: [0, 1, 2], left: 1.5},
          {startMs: 29000, endMs: 36000, show: 'remainder', taken: [0, 1, 2], left: 1.5},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The latency template on the case that motivated it: an in-memory read, an
 * indexed query and the same query without its index.
 *
 * The span is deliberately wide — 0.12ms to 6.5s, five decades — because the
 * axis derivation is the part most likely to be wrong, and a two-decade fixture
 * would not exercise the tick padding at either end.
 */
const latencyVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 34000,
  scenes: [
    {
      type: 'latency',
      startMs: 0,
      endMs: 34000,
      props: {
        title: 'Not the same kind of slow',
        emphasis: 'kind of slow',
        emphasisRole: 'limit',
        // As latencyAxis derives them: floor(log10(0.12)) = -1, ceil(log10(6479)) = 4.
        ticks: [
          {label: '100µs', frac: 0},
          {label: '1ms', frac: 0.2},
          {label: '10ms', frac: 0.4},
          {label: '100ms', frac: 0.6},
          {label: '1s', frac: 0.8},
          {label: '10s', frac: 1},
        ],
        operations: [
          {
            label: 'a Redis GET',
            value: '0.1ms',
            note: 'Memory, one hop, no parsing',
            role: 'quantity',
            frac: 0.0159,
          },
          {
            label: 'an indexed SQL query',
            value: '12ms',
            note: 'A hundred times longer, and that is the good case',
            role: 'neutral',
            frac: 0.4157,
          },
          {
            label: 'the same query unindexed',
            value: '6.5s',
            note: 'You could have done fifty thousand GETs in that time',
            role: 'limit',
            frac: 0.9527,
          },
        ],
        steps: [
          {startMs: 0, endMs: 7000, show: 'axis', placed: []},
          {startMs: 7000, endMs: 14000, show: 'place', at: 0, placed: [0]},
          {startMs: 14000, endMs: 21000, show: 'place', at: 1, placed: [0, 1]},
          {startMs: 21000, endMs: 28000, show: 'place', at: 2, placed: [0, 1, 2]},
          {startMs: 28000, endMs: 34000, show: 'read', placed: [0, 1, 2]},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The multiply template on the frame that motivated it: one GPU node's power
 * draw against a full rack's.
 *
 * 14.5 x 8 = 116, and those are the fixture's real numbers rather than round
 * ones — the arithmetic validator is the whole reason this template exists, so
 * the baseline should be of a case where the product is not obvious by eye.
 */
const multiplyVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 28000,
  scenes: [
    {
      type: 'multiply',
      startMs: 0,
      endMs: 28000,
      props: {
        title: 'One node is fine. Eight is a substation.',
        emphasis: 'a substation',
        emphasisRole: 'limit',
        unitValue: 14.5,
        unit: 'kW',
        unitLabel: 'one B200 node',
        unitNote: 'About what a domestic oven pulls',
        count: 8,
        countLabel: 'nodes in one rack',
        total: 116,
        totalLabel: 'before cooling',
        totalNote: 'More than most office floors are wired for',
        caveat: 'And cooling adds roughly half again',
        role: 'limit',
        steps: [
          {startMs: 0, endMs: 8000, show: 'unit'},
          {startMs: 8000, endMs: 15000, show: 'count'},
          {startMs: 15000, endMs: 23000, show: 'total'},
          {startMs: 23000, endMs: 28000, show: 'caveat'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The ratio template on the frame that motivated it: the DGX Spark's memory
 * bandwidth against a Mac Studio's.
 *
 * 270 out of 800 is 0.3375, which the clip calls "a third" — the case the
 * tolerance exists for. A fixture with round numbers would never exercise the
 * gap between the arithmetic and the phrase a person actually says.
 */
const ratioVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 28000,
  scenes: [
    {
      type: 'ratio',
      startMs: 0,
      endMs: 28000,
      props: {
        // Deliberately NOT "a third of the Mac": the phrase is set at 104px below
        // the bars, and repeating it in the headline made the frame stutter — the
        // same words twice reads as a rendering fault rather than as emphasis.
        // The headline sets up the omission; the phrase is the payoff.
        title: 'The spec sheet leaves this out',
        emphasis: 'leaves this out',
        emphasisRole: 'limit',
        unit: 'GB/s',
        reference: {label: 'Mac Studio M3 Ultra', value: 800, role: 'rival'},
        subject: {label: 'DGX Spark', value: 270, role: 'limit', frac: 0.3375},
        phrase: 'a third',
        note: 'So every token comes out slower, whatever the spec sheet leads with',
        steps: [
          {startMs: 0, endMs: 8000, show: 'reference'},
          {startMs: 8000, endMs: 15000, show: 'subject'},
          {startMs: 15000, endMs: 22000, show: 'fraction'},
          {startMs: 22000, endMs: 28000, show: 'read'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The table template on the frame that motivated it: a GPU spec sheet where the
 * line that decides whether a model fits is the fifth of six.
 *
 * The rows are in the order the real product page prints them, which is the
 * subject — the fixture would prove nothing if the buried row had been moved to
 * make it easier to bury.
 */
const tableVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 26000,
  scenes: [
    {
      type: 'table',
      startMs: 0,
      endMs: 26000,
      props: {
        title: 'The line the spec sheet buries',
        emphasis: 'buries',
        emphasisRole: 'limit',
        source: 'RTX 5090, from the product page',
        rows: [
          {label: 'CUDA Cores', value: '21,760'},
          {label: 'Tensor Cores', value: '680'},
          {label: 'Boost Clock', value: '2.52 GHz'},
          {label: 'Memory Bandwidth', value: '1,792 GB/s'},
          {label: 'Memory Capacity', value: '32 GB'},
          {label: 'TDP', value: '575 W'},
        ],
        at: 4,
        note: 'Every other number is irrelevant if the model does not fit in this one',
        role: 'limit',
        steps: [
          {startMs: 0, endMs: 10000, show: 'sheet'},
          {startMs: 10000, endMs: 19000, show: 'focus'},
          {startMs: 19000, endMs: 26000, show: 'read'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The toggle template on the clip that motivated it: whether WebAssembly is
 * replacing JavaScript.
 *
 * The fixture is captured on its SECOND qualifier, because that is the state
 * where the switch has already receded into a header and the asterisks own the
 * frame — the layout change is the part most likely to break, and the answer
 * beat alone would never exercise it.
 */

// The v5 course-scaffolding fixtures. All five run under the `editorial` skin —
// the batch and the skin arrived together, and these are the catalog's first
// baselines of an off-centre composition.
const V5_THEME = {
  primary: '#306998',
  accent: '#ffd43b',
  background: '#ffffff',
  courseName: 'Coursesmith',
  skin: 'editorial' as const,
  accentQuantity: '#f5ca47',
  accentLimit: '#ec5b51',
  accentRival: '#518cec',
};

const objectiveVizProps: LessonVideoProps = {
  theme: V5_THEME,
  audioFile: '',
  durationMs: 28000,
  scenes: [{type: 'objective', startMs: 0, endMs: 28000, props: {
    title: "What you'll be able to do",
    emphasis: 'be able to do', emphasisRole: 'quantity',
    audience: 'you have shipped an API client',
    outcomes: [
      {action: 'Write a retry that gives up', evidence: 'your client stops after four attempts instead of hanging'},
      {action: 'Size a backoff to a real timeout', evidence: "total retry time lands under the caller's deadline"},
    ],
    steps: [
      {startMs: 0, endMs: 8000, show: 'frame', lit: []},
      {startMs: 8000, endMs: 16000, show: 'outcome', at: 0, lit: [0]},
      {startMs: 16000, endMs: 23000, show: 'outcome', at: 1, lit: [0, 1]},
      {startMs: 23000, endMs: 28000, show: 'contract', lit: [0, 1]},
    ],
  }}],
  captions: [],
};

const prereqVizProps: LessonVideoProps = {
  theme: V5_THEME,
  audioFile: '',
  durationMs: 26000,
  scenes: [{type: 'prereq', startMs: 0, endMs: 26000, props: {
    title: 'What this one stands on',
    emphasis: 'stands on', emphasisRole: 'limit',
    assumptions: [
      {item: 'Reading a stack trace', source: 'taught', where: 'lesson 2, when the client first failed', skippable: false},
      {item: 'Running the suite locally', source: 'external', where: 'any Node setup guide', skippable: true,
       breaks: 'you can watch, but you cannot try the checkpoint'},
    ],
    steps: [
      {startMs: 0, endMs: 9000, show: 'assume', at: 0, lit: [0]},
      {startMs: 9000, endMs: 18000, show: 'assume', at: 1, lit: [0, 1]},
      {startMs: 18000, endMs: 26000, show: 'floor', lit: [0, 1]},
    ],
  }}],
  captions: [],
};

const recapVizProps: LessonVideoProps = {
  theme: V5_THEME,
  audioFile: '',
  durationMs: 27000,
  scenes: [{type: 'recap', startMs: 0, endMs: 27000, props: {
    title: 'Where we got to',
    emphasis: 'got to', emphasisRole: 'quantity',
    thread: 'all of it has been about what to do when the consumer is slower',
    claims: [
      {claim: 'A queue absorbs a burst, not a trend', from: 'lesson 2'},
      {claim: 'An unbounded queue turns latency into memory', from: 'lesson 4'},
    ],
    steps: [
      {startMs: 0, endMs: 9000, show: 'claim', at: 0, lit: [0]},
      {startMs: 9000, endMs: 18000, show: 'claim', at: 1, lit: [0, 1]},
      {startMs: 18000, endMs: 27000, show: 'standing', lit: [0, 1]},
    ],
  }}],
  captions: [],
};

const pitfallVizProps: LessonVideoProps = {
  theme: V5_THEME,
  audioFile: '',
  durationMs: 32000,
  scenes: [{type: 'pitfall', startMs: 0, endMs: 32000, props: {
    title: 'The retry that never gives up',
    emphasis: 'never gives up', emphasisRole: 'limit',
    mistake: 'Retrying every error, including the ones that will never succeed',
    symptom: 'one dead endpoint and your p99 pins to the timeout ceiling',
    why: 'The retry looked like resilience, and in the happy path it is.',
    fix: 'Retry only what is retryable, and cap the total wait',
    steps: [
      {startMs: 0, endMs: 8000, show: 'mistake'},
      {startMs: 8000, endMs: 16000, show: 'symptom'},
      {startMs: 16000, endMs: 23000, show: 'why'},
      {startMs: 23000, endMs: 32000, show: 'fix'},
    ],
  }}],
  captions: [],
};

const checkpointVizProps: LessonVideoProps = {
  theme: V5_THEME,
  audioFile: '',
  durationMs: 30000,
  scenes: [{type: 'checkpoint', startMs: 0, endMs: 30000, props: {
    title: 'Make it give up',
    emphasis: 'give up', emphasisRole: 'limit',
    task: 'Write a retry that stops after four tries',
    list: ['Point the client at an endpoint that always fails', 'Add a counter and break at four', 'Log the total elapsed time'],
    done: 'the log shows four attempts and the call returns',
    stuck: 'if it hangs, your timeout is longer than your budget',
    steps: [
      {startMs: 0, endMs: 7000, show: 'task', ticked: []},
      {startMs: 7000, endMs: 13000, show: 'step', at: 0, ticked: [0]},
      {startMs: 13000, endMs: 19000, show: 'step', at: 1, ticked: [0, 1]},
      {startMs: 19000, endMs: 24000, show: 'step', at: 2, ticked: [0, 1, 2]},
      {startMs: 24000, endMs: 30000, show: 'done', ticked: [0, 1, 2]},
    ],
  }}],
  captions: [],
};

const toggleVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    skin: 'broadcast',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 30000,
  scenes: [
    {
      type: 'toggle',
      startMs: 0,
      endMs: 30000,
      props: {
        title: 'No. But that is not the interesting part.',
        emphasis: 'not the interesting part',
        emphasisRole: 'quantity',
        question: 'Is WebAssembly replacing JavaScript?',
        from: 'yes',
        to: 'no',
        qualifiers: [
          {
            label: 'in the browser',
            note: 'It cannot touch the DOM without going through JavaScript to get there',
            role: 'limit',
          },
          {
            label: 'outside it',
            note: 'On the edge it replaced containers instead, which nobody was predicting',
            role: 'quantity',
          },
        ],
        steps: [
          {startMs: 0, endMs: 9000, show: 'answer', raised: []},
          {startMs: 9000, endMs: 17000, show: 'qualify', at: 0, raised: [0]},
          {startMs: 17000, endMs: 25000, show: 'qualify', at: 1, raised: [0, 1]},
          {startMs: 25000, endMs: 30000, show: 'settle', raised: [0, 1]},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The verdict template on the example that motivated it: whether to self-host a
 * database. Three conditions it holds on, two it breaks on, and a call.
 */
const verdictVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    // Carried explicitly for the same reason as the gauge: this template is
    // *about* the difference between the ground the call holds on and the
    // ground it breaks on, and without the semantic accents both columns
    // render in the same brand gold.
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 42000,
  scenes: [
    {
      type: 'verdict',
      startMs: 0,
      endMs: 42000,
      props: {
        title: 'Should you actually self-host your database?',
        subject: 'Self-hosting Postgres',
        call: 'Rent it until you have a platform team',
        holds: [
          'Under about fifty gigabytes of data',
          'When downtime costs you real money',
          'If nobody runs backups weekly',
        ],
        breaks: [
          'When compliance forbids a managed provider',
          'Past a few terabytes, where the bill inverts',
        ],
        steps: [
          {startMs: 0, endMs: 6000, show: 'subject'},
          {startMs: 6000, endMs: 12000, show: 'holds', at: 0},
          {startMs: 12000, endMs: 18000, show: 'holds', at: 1},
          {startMs: 18000, endMs: 24000, show: 'holds', at: 2},
          {startMs: 24000, endMs: 30000, show: 'breaks', at: 0},
          {startMs: 30000, endMs: 36000, show: 'breaks', at: 1},
          {startMs: 36000, endMs: 42000, show: 'call'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The decision template on the example that motivated it: which GPU to buy,
 * decided by one question. Three tiers, each in a different role colour so the
 * axis reads as a gradient of consequence.
 */
const decisionVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 40000,
  scenes: [
    {
      type: 'decision',
      startMs: 0,
      endMs: 40000,
      props: {
        title: 'Which GPU should you actually buy?',
        question: 'How big is your model?',
        unit: 'GB',
        tiers: [
          {
            band: 'Under 8GB',
            answer: 'A used 3060 is enough',
            note: 'Nothing bigger buys you anything at this size',
            role: 'quantity',
          },
          {
            band: '8 to 24GB',
            answer: 'Buy the 4090',
            note: 'The last size where one consumer card still does it',
            role: 'rival',
          },
          {
            band: 'Over 24GB',
            answer: 'Rent it by the hour',
            note: 'Two cards and a power supply costs more than a year of renting',
            role: 'limit',
          },
        ],
        steps: [
          {startMs: 0, endMs: 8000, show: 'question'},
          {startMs: 8000, endMs: 16000, show: 'tier', at: 0},
          {startMs: 16000, endMs: 24000, show: 'tier', at: 1},
          {startMs: 24000, endMs: 32000, show: 'tier', at: 2},
          {startMs: 32000, endMs: 40000, show: 'rule'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The myth template on the example that motivated it: the belief that Redis is
 * only a cache. The strike frame and the evidence frames are the two states
 * worth watching.
 */
const mythVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 42000,
  scenes: [
    {
      type: 'myth',
      startMs: 0,
      endMs: 42000,
      props: {
        title: 'What everyone gets wrong about Redis',
        claim: 'Redis is just a cache',
        truth: 'Redis is a data structure server',
        why: 'Because the first thing anyone uses it for is caching, and it is very good at that',
        evidence: [
          'Sorted sets give you a leaderboard',
          'Streams give you a durable log',
          'Lua scripts run atomically',
        ],
        steps: [
          {startMs: 0, endMs: 7000, show: 'claim'},
          {startMs: 7000, endMs: 14000, show: 'strike'},
          {startMs: 14000, endMs: 21000, show: 'evidence', at: 0},
          {startMs: 21000, endMs: 28000, show: 'evidence', at: 1},
          {startMs: 28000, endMs: 35000, show: 'evidence', at: 2},
          {startMs: 35000, endMs: 42000, show: 'why'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The rundown template on the example that motivated it: the three numbers that
 * decide whether a model runs locally. The promise announces three and there are
 * exactly three cards, which is the agreement this template is built to keep.
 */
const rundownVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 36000,
  scenes: [
    {
      type: 'rundown',
      startMs: 0,
      endMs: 36000,
      props: {
        title: 'The three numbers that decide everything',
        promise: 'Three numbers decide everything',
        items: [
          {
            label: 'Memory capacity',
            detail: 'How much the model needs before a single token comes out',
            icon: 'database',
          },
          {
            label: 'Memory bandwidth',
            detail: 'The hidden boss — it, not compute, sets your tokens per second',
            icon: 'zap',
          },
          {
            label: 'Compute',
            detail: 'The number every spec sheet leads with, and the one that matters least',
            icon: 'gear',
          },
        ],
        steps: [
          {startMs: 0, endMs: 8000, show: 'promise'},
          {startMs: 8000, endMs: 16000, show: 'item', at: 0},
          {startMs: 16000, endMs: 24000, show: 'item', at: 1},
          {startMs: 24000, endMs: 30000, show: 'item', at: 2},
          {startMs: 30000, endMs: 36000, show: 'all'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The analogy template on the image the strongest reference clip is built from:
 * a librarian in a library standing in for a machine running a model.
 */
const analogyVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 42000,
  scenes: [
    {
      type: 'analogy',
      startMs: 0,
      endMs: 42000,
      props: {
        title: 'A librarian and a library',
        familiar: 'A library',
        familiarIcon: 'book',
        real: 'Running a model',
        realIcon: 'server',
        pairs: [
          {
            from: 'The size of the room',
            to: 'Memory capacity',
            note: 'A book that will not fit in the room cannot be read at all',
          },
          {
            from: 'The walk to the shelf',
            to: 'Memory bandwidth',
            note: 'Most of the day goes on walking, not on reading',
          },
          {
            from: 'How fast they read',
            to: 'Compute',
            note: 'Rarely the bottleneck, and the only number on the box',
          },
        ],
        breaks: 'A librarian can skim; a machine reads every word, every time',
        steps: [
          {startMs: 0, endMs: 8000, show: 'picture'},
          {startMs: 8000, endMs: 16000, show: 'pair', at: 0},
          {startMs: 16000, endMs: 24000, show: 'pair', at: 1},
          {startMs: 24000, endMs: 32000, show: 'pair', at: 2},
          {startMs: 32000, endMs: 42000, show: 'breaks'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The trace template on the classic race: two users buying the last item. The
 * value goes 1 -> 0 -> 0, and the second decrement changing nothing is the bug.
 */
const traceVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 44000,
  scenes: [
    {
      type: 'trace',
      startMs: 0,
      endMs: 44000,
      props: {
        title: 'Two users, one item left',
        actors: ['User A', 'User B'],
        resource: 'Inventory',
        start: '1',
        ops: [
          {by: 0, op: 'read inv', becomes: '1', note: 'A reads one in stock and decides to sell', changes: false},
          {by: 1, op: 'read inv', becomes: '1', note: 'B reads the same one, before A has written anything', changes: false},
          {by: 0, op: 'write 0', becomes: '0', note: 'A takes the item and writes zero', changes: true},
          {by: 1, op: 'write 0', becomes: '0', note: 'B writes zero too — from the value it read a moment ago', changes: false},
        ],
        outcome: 'Two customers, one item, both charged',
        broken: true,
        steps: [
          {startMs: 0, endMs: 7000, show: 'setup'},
          {startMs: 7000, endMs: 13000, show: 'queue'},
          {startMs: 13000, endMs: 20000, show: 'step', at: 0},
          {startMs: 20000, endMs: 27000, show: 'step', at: 1},
          {startMs: 27000, endMs: 33000, show: 'step', at: 2},
          {startMs: 33000, endMs: 39000, show: 'step', at: 3},
          {startMs: 39000, endMs: 44000, show: 'outcome'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The costing template on a bill whose surprise is in the small lines: the card
 * is the obvious cost, the power and the cooling are the ones nobody budgets.
 */
const costingVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 40000,
  scenes: [
    {
      type: 'costing',
      startMs: 0,
      endMs: 40000,
      props: {
        title: 'What a GPU box really costs in year one',
        subject: 'A self-hosted GPU box, year one',
        unit: '$',
        lines: [
          {label: 'The card', amount: 1800, note: 'The number everyone quotes', running: 1800, frac: 1},
          {label: 'The rest of the box', amount: 1100, note: 'Power supply, board, and the case it fits in', running: 2900, frac: 0.6111},
          {
            label: 'Electricity',
            amount: 620,
            note: 'Four hundred watts, running most of the day, for a year',
            hidden: true,
            running: 3520,
            frac: 0.3444,
          },
          {
            label: 'The noise fix',
            amount: 340,
            note: 'Nobody keeps it in the room they work in for long',
            hidden: true,
            running: 3860,
            frac: 0.1889,
          },
        ],
        total: 3860,
        verdict: 'Twice the sticker price, and the card was the cheap half',
        steps: [
          {startMs: 0, endMs: 7000, show: 'setup'},
          {startMs: 7000, endMs: 14000, show: 'line', at: 0},
          {startMs: 14000, endMs: 21000, show: 'line', at: 1},
          {startMs: 21000, endMs: 28000, show: 'line', at: 2},
          {startMs: 28000, endMs: 34000, show: 'line', at: 3},
          {startMs: 34000, endMs: 40000, show: 'total'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The constellation template on the frame the Redis reference clip closes with:
 * the name in the middle and the four properties that define it around it.
 */
const constellationVizProps: LessonVideoProps = {
  theme: {
    primary: '#306998',
    accent: '#ffd43b',
    background: '#ffffff',
    courseName: 'Coursesmith',
    accentQuantity: '#f0c74c',
    accentLimit: '#ec5b51',
    accentRival: '#518cec',
  },
  audioFile: '',
  durationMs: 40000,
  scenes: [
    {
      type: 'constellation',
      startMs: 0,
      endMs: 40000,
      props: {
        title: 'Everything that makes Redis Redis',
        centre: 'Redis',
        centreIcon: 'database',
        spokes: [
          {rel: 'is', label: 'In-memory', note: 'Every value lives in RAM, which is why it answers in microseconds', icon: 'zap', angle: -90.0},
          {rel: 'is', label: 'Single-threaded', note: 'One command at a time, so nothing ever races another', icon: 'clock', angle: 0.0},
          {rel: 'gives you', label: 'Data structures', note: 'Lists, sets and sorted sets, not just opaque blobs', icon: 'layers', angle: 90.0},
          {rel: 'survives', label: 'Restarts', note: 'Snapshots and an append-only log, if you ask for them', icon: 'shield', angle: 180.0},
        ],
        steps: [
          {startMs: 0, endMs: 7000, show: 'centre'},
          {startMs: 7000, endMs: 14000, show: 'spoke', at: 0},
          {startMs: 14000, endMs: 21000, show: 'spoke', at: 1},
          {startMs: 21000, endMs: 28000, show: 'spoke', at: 2},
          {startMs: 28000, endMs: 34000, show: 'spoke', at: 3},
          {startMs: 34000, endMs: 40000, show: 'whole'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The theme tokens Go derives for a dark-mode course from the default brand
 * colours. Copied from `videoThemeForConfig` output rather than hand-picked, so
 * a fixture cannot quietly disagree with what the pipeline actually emits — and
 * carrying the three semantic accents explicitly matters here: `resolveTheme`
 * falls them all back to the brand accent, which would render a path's walked
 * segment and its unwalked one in the same colour.
 */
const darkTokens = {
  primary: '#306998',
  accent: '#ffd43b',
  background: '#ffffff',
  courseName: 'Coursesmith',
  mode: 'dark' as const,
  bgTop: '#101d28',
  bgBottom: '#070c15',
  surface: '#1b2732',
  surfaceBorder: '#324452',
  text: '#f2f5f8',
  textMuted: '#a6b4bf',
  accentQuantity: '#f5ca47',
  accentLimit: '#ec5b51',
  accentRival: '#518cec',
  mass: '#dee6ed',
  ink: '#071018',
  accentText: '#ffd43b',
  grain: 0.04,
};

/** The same course in light mode, from the same source. */
const lightTokens = {
  ...darkTokens,
  mode: 'light' as const,
  bgTop: '#f0f4f7',
  bgBottom: '#e9ecef',
  surface: '#ffffff',
  surfaceBorder: '#c8d2da',
  text: '#13222f',
  textMuted: '#4b6071',
  // Quantity rotates toward amber on paper: gold walked down to AA lands on a
  // khaki that reads as mud rather than as a chosen colour.
  accentQuantity: '#a45c09',
  accentLimit: '#c12115',
  accentRival: '#1557c1',
  mass: '#7c96ab',
  ink: '#1c354a',
  accentText: '#826600',
  grain: 0.01,
};

/**
 * The chapter template on a break three parts into a five-part course: two
 * stops ticked off behind, this one lit, two still faint ahead.
 */
const chapterVizProps: LessonVideoProps = {
  theme: darkTokens,
  audioFile: '',
  durationMs: 30000,
  scenes: [
    {
      type: 'chapter',
      startMs: 0,
      endMs: 30000,
      props: {
        title: 'Part three: loops',
        path: 'The Python course',
        at: 2,
        ordinal: 3,
        total: 5,
        stops: [
          {label: 'Printing', icon: 'terminal', note: 'Getting Python to say something back to you', state: 'done'},
          {label: 'Variables', icon: 'box', note: 'Names for the things you want to keep', state: 'done'},
          {label: 'Loops', icon: 'refresh', note: 'Doing the same work without writing it twice', state: 'here'},
          {label: 'Functions', icon: 'puzzle', note: 'Wrapping work up so you can call it by name', state: 'ahead'},
          {label: 'Files', icon: 'folder', note: 'Reading and writing what outlives the program', state: 'ahead'},
        ],
        steps: [
          {startMs: 0, endMs: 8000, show: 'path'},
          {startMs: 8000, endMs: 15000, show: 'done', at: 0},
          {startMs: 15000, endMs: 22000, show: 'done', at: 1},
          {startMs: 22000, endMs: 30000, show: 'here'},
        ],
      },
    },
  ],
  captions: [],
};

const chapterLightProps: LessonVideoProps = {...chapterVizProps, theme: lightTokens};

/**
 * The cycle template on the debugging loop: four stages, and a return that says
 * what is smaller next lap.
 */
const cycleVizProps: LessonVideoProps = {
  theme: darkTokens,
  audioFile: '',
  durationMs: 62000,
  scenes: [
    {
      type: 'cycle',
      startMs: 0,
      endMs: 62000,
      props: {
        title: 'The debugging loop',
        name: 'The debugging loop',
        changes: 'The failing case gets smaller each lap',
        stages: [
          {label: 'Reproduce', icon: 'repeat', note: 'Get it to fail on demand, or you are only guessing', angle: -90},
          {label: 'Isolate', icon: 'search', note: 'Cut away everything that still fails without it', angle: 0},
          {label: 'Fix', icon: 'wrench', note: 'Change one thing, and only one', angle: 90},
          {label: 'Verify', icon: 'check', note: 'Run the case that failed, then run everything else', angle: 180},
        ],
        steps: [
          {startMs: 0, endMs: 9000, show: 'ring'},
          {startMs: 9000, endMs: 20000, show: 'stage', at: 0},
          {startMs: 20000, endMs: 30000, show: 'stage', at: 1},
          {startMs: 30000, endMs: 40000, show: 'stage', at: 2},
          {startMs: 40000, endMs: 50000, show: 'stage', at: 3},
          {startMs: 50000, endMs: 62000, show: 'again'},
        ],
      },
    },
  ],
  captions: [],
};

const cycleLightProps: LessonVideoProps = {...cycleVizProps, theme: lightTokens};

/**
 * The scale template on four rungs of data, forty million times apart end to
 * end — which is exactly the span no bar chart can draw.
 */
const scaleVizProps: LessonVideoProps = {
  theme: darkTokens,
  audioFile: '',
  durationMs: 52000,
  scenes: [
    {
      type: 'scale',
      startMs: 0,
      endMs: 52000,
      props: {
        title: 'How much is a terabyte, really',
        unit: 'MB',
        span: '40 billion',
        levels: [
          {label: 'This sentence', value: 0.0001, display: '100 bytes', icon: 'file', note: 'A hundred characters, and nothing else'},
          {label: 'A phone photo', value: 4, display: '4 MB', icon: 'star', note: 'Forty thousand of those sentences', times: '40000'},
          {label: 'A feature film', value: 4000, display: '4 GB', icon: 'play', note: 'A thousand photos, played back at speed', times: '1000'},
          {label: 'A small library', value: 4000000, display: '4 TB', icon: 'city', note: 'A thousand films, on one drive you can hold', times: '1000'},
        ],
        steps: [
          {startMs: 0, endMs: 9000, show: 'level', at: 0},
          {startMs: 9000, endMs: 19000, show: 'level', at: 1},
          {startMs: 19000, endMs: 29000, show: 'level', at: 2},
          {startMs: 29000, endMs: 41000, show: 'level', at: 3},
          {startMs: 41000, endMs: 52000, show: 'whole'},
        ],
      },
    },
  ],
  captions: [],
};

const scaleLightProps: LessonVideoProps = {...scaleVizProps, theme: lightTokens};

const showcaseVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 40000,
  scenes: [
    {
      type: 'showcase',
      startMs: 0,
      endMs: 40000,
      props: {
        title: 'Airtable, and when not to use it',
        name: 'Airtable',
        category: 'Database',
        tagline: 'A database that looks and feels like a spreadsheet',
        icon: 'database',
        facts: [
          {label: 'Best for', value: 'Small structured datasets'},
          {label: 'Price', value: 'Free to 1,000 rows'},
          {label: 'Lock-in', value: 'CSV out, formulas stay'},
          {label: 'Learning curve', value: 'An afternoon'},
        ],
        strengths: [
          'Non-technical people can edit it',
          'Views and filters without queries',
          'Connects to almost everything',
        ],
        limits: ['Slows badly past fifty thousand rows', 'Per-seat pricing punishes big teams'],
        steps: [
          {startMs: 0, endMs: 5000, show: 'intro'},
          {startMs: 5000, endMs: 9500, show: 'fact', at: 0},
          {startMs: 9500, endMs: 14000, show: 'fact', at: 1},
          {startMs: 14000, endMs: 18500, show: 'fact', at: 2},
          {startMs: 18500, endMs: 23000, show: 'fact', at: 3},
          {startMs: 23000, endMs: 28500, show: 'strengths'},
          {startMs: 28500, endMs: 34500, show: 'limits'},
          {startMs: 34500, endMs: 40000, show: 'handoff'},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The breakdown template on the example that motivated it: building a whole
 * website with no-code tools, four phases deep, with item beats pulled out of
 * the two stages where the choice actually matters. Seven beats.
 */
const breakdownVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 38000,
  scenes: [
    {
      type: 'breakdown',
      startMs: 0,
      endMs: 38000,
      props: {
        title: 'Building a whole website with no code',
        phases: [
          {
            title: 'Design',
            detail: 'Decide what it looks like before anything gets built',
            items: [
              {name: 'Figma', note: 'Fastest if you already know it', icon: 'layers'},
              {name: 'Canva', note: 'Templates when you are not a designer', icon: 'star'},
            ],
          },
          {
            title: 'Wireframe',
            detail: 'Block out the page before anyone argues about colour',
            items: [
              {name: 'Whimsical', note: 'Fast, ugly, on purpose', icon: 'box'},
              {name: 'Excalidraw', note: 'Hand-drawn feel, free forever', icon: 'idea'},
            ],
          },
          {
            title: 'Front end',
            detail: 'Turn the design into pages people can actually open',
            items: [
              {name: 'Webflow', note: 'Most control, steepest curve', icon: 'monitor'},
              {name: 'Framer', note: 'Fastest from design to live', icon: 'zap'},
              {name: 'Softr', note: 'Best if your data is Airtable', icon: 'layers'},
            ],
          },
          {
            title: 'Back end',
            detail: 'Where the data lives and what happens on submit',
            items: [
              {name: 'Airtable', note: 'A database anyone can edit', icon: 'database'},
              {name: 'Make', note: 'Glue between the form and everything', icon: 'shuffle'},
            ],
          },
        ],
        steps: [
          {startMs: 0, endMs: 5000, show: 'phase', at: 0},
          {startMs: 5000, endMs: 10000, show: 'item', at: 0, item: 0},
          {startMs: 10000, endMs: 15000, show: 'phase', at: 1},
          {startMs: 15000, endMs: 20000, show: 'phase', at: 2},
          {startMs: 20000, endMs: 26000, show: 'item', at: 2, item: 0},
          {startMs: 26000, endMs: 32000, show: 'phase', at: 3},
          {startMs: 32000, endMs: 38000, show: 'whole'},
        ],
      },
    },
  ],
  captions: [],
};

const dataVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: dataChartDemos.length * DATA_DEMO_MS,
  scenes: dataChartDemos.map((demo, i) => {
    const startMs = i * DATA_DEMO_MS;
    const endMs = startMs + DATA_DEMO_MS;
    // The last two points are lit halfway through, so every kind is seen both
    // at rest and under a highlight.
    const litFrom = startMs + DATA_DEMO_MS / 2;
    return {
      type: 'data',
      startMs,
      endMs,
      props: {
        title: demo.title,
        kind: demo.kind,
        unit: demo.unit,
        ...(demo.series ? {series: demo.series} : {}),
        points: demo.points.map((p) => ({
          label: p.label,
          value: p.values ? p.values.reduce((a, b) => a + b, 0) : (p.value ?? 0),
          ...(p.values ? {values: p.values} : {}),
        })),
        highlight: [
          {startMs: litFrom, endMs, labels: demo.points.slice(-2).map((p) => p.label)},
        ],
        captions: [
          {startMs, endMs: litFrom, text: `Kind: ${demo.kind}. Nothing highlighted yet.`},
          {startMs: litFrom, endMs, text: `Kind: ${demo.kind}. The last two points are lit.`},
        ],
      },
    };
  }),
  captions: [],
};

// A demo of the workspace scene, so `remotion still WorkspaceViz` renders it
// standalone. The output below is what this program really prints — the plan
// stage runs the file set through the sandbox, so a fixture that lied about it
// would be lying about the one thing this template guarantees.
const workspaceVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 24000,
  scenes: [
    {
      type: 'workspace',
      startMs: 0,
      endMs: 24000,
      props: {
        title: 'Share one function across files',
        project: 'greet',
        entry: 'main.py',
        command: 'python3 main.py',
        output: 'Hello, Ada!\nHello, Alan!',
        files: [
          {path: 'greet.py', code: 'def hello(who):\n    return f"Hello, {who}!"\n'},
          {
            path: 'main.py',
            code: 'from greet import hello\n\nfor name in ["Ada", "Alan"]:\n    print(hello(name))\n',
          },
        ],
        steps: [
          {startMs: 0, endMs: 5000, file: 'greet.py', through: 0, focus: 'tree', run: false,
           caption: 'Start with the piece you want to reuse.'},
          {startMs: 5000, endMs: 10000, file: 'greet.py', through: 0, focus: 'code', run: false},
          {startMs: 10000, endMs: 14000, file: 'main.py', through: 1, focus: 'tabs', run: false,
           caption: 'One line, and the function is here.'},
          {startMs: 14000, endMs: 19000, file: 'main.py', through: 0, focus: 'code', run: false},
          {startMs: 19000, endMs: 24000, file: 'main.py', through: 0, focus: 'terminal', run: true,
           caption: 'The output is real — this actually ran.'},
        ],
      },
    },
  ],
  captions: [],
};

// A demo of the VS Code walkthrough scene, so `remotion still VSCodeViz`
// renders it standalone. There is a light-mode twin below: the editor carries
// its own palette rather than the design system's, so it is the one scene where
// "does light mode work" cannot be inferred from any other composition.
const vscodeVizProps: LessonVideoProps = {
  theme: {primary: '#306998', accent: '#ffd43b', background: '#ffffff', courseName: 'Coursesmith'},
  audioFile: '',
  durationMs: 18000,
  scenes: [
    {
      type: 'walkthrough',
      startMs: 0,
      endMs: 18000,
      props: {
        title: 'Build the greeting step by step',
        file: 'greeting.py',
        project: 'python-basics',
        language: 'python',
        files: ['greeting.py', 'variables.py', 'math_ops.py'],
        // The keystroke schedule Go produces for step 0's code over this
        // window, pasted verbatim.
        //
        // Present so the baseline exercises the *real* typing path. Without it
        // this composition silently used the renderer's fallback estimate, so
        // the whole schedule-consuming branch — and the click track that shares
        // these numbers — shipped with no picture of it. Regenerate with
        // KeystrokeSchedule(code, 5000 * typingPortionOfWindow) if the step 0
        // code or its window changes; the gap at index 12 is the newline pause,
        // which is the shape worth eyeballing.
        keystrokes: [
          86, 154, 257, 367, 468, 540, 642, 704, 780, 860, 936, 1062, 1385, 1462, 1537, 1629,
          1724, 1839, 1915, 1997, 2096, 2172, 2286, 2370, 2470, 2577, 2712, 2794, 2904, 2970,
          3119, 3186, 3256, 3331, 3413, 3500,
        ],
        steps: [
          {code: 'name = "Ada"\nprint("Hello, " + name)', atMs: 0},
          {code: 'name = "Ada"\n\ndef greet(who):\n    return "Hello, " + who\n\nprint(greet(name))', atMs: 5000},
          {code: 'name = "Ada"\n\ndef greet(who, excited=False):\n    suffix = "!" if excited else "."\n    return "Hello, " + who + suffix\n\nprint(greet(name, excited=True))', atMs: 9000},
          // The run step. Without one in the demo the terminal drawer — the
          // half of this template that shows the code actually doing something
          // — had no composition and no baseline, so a regression in it could
          // only be found by watching a real clip.
          {
            code: 'name = "Ada"\n\ndef greet(who, excited=False):\n    suffix = "!" if excited else "."\n    return "Hello, " + who + suffix\n\nprint(greet(name, excited=True))',
            atMs: 13000,
            run: true,
            command: 'python3 greeting.py',
            output: 'Hello, Ada!',
          },
        ],
      },
    },
  ],
  captions: [],
};

const vscodeLightProps: LessonVideoProps = {
  ...vscodeVizProps,
  theme: {...illustrationLightProps.theme},
};

/**
 * The opening gesture, which nothing else has a picture of.
 *
 * The main VS Code fixture sets neither `intro` nor `typeAtMs`, so the whole
 * opening — the window scaling up, the pointer travelling to the file, the
 * click, the tab sliding in — is switched off in one case and happens at
 * negative frames in the other. Every part of it could break and every baseline
 * would still pass, while the snippet path renders it on every clip.
 *
 * `keystrokes` is deliberately absent: the frame this composition baselines is
 * before the first character, so the schedule is immaterial here, and pasting a
 * second copy of it would be a second thing to regenerate.
 */
const vscodeIntroProps: LessonVideoProps = {
  ...vscodeVizProps,
  scenes: vscodeVizProps.scenes.map((s) =>
    s.type === 'walkthrough'
      ? {...s, props: {...s.props, intro: true, typeAtMs: 1400, keystrokes: undefined}}
      : s,
  ),
};



// footageVizProps renders the `footage` template's own contribution — the
// browser chrome, the address of the origin the driver recorded, and the
// capture credit — around a deliberately neutral placeholder frame.
//
// It is the one preview in the gallery that cannot show real content, and that
// is a property of the template rather than a shortcut: what `footage` produces
// is *your* recording, so a card showing somebody else's clip would misrepresent
// what you get. What the template actually contributes is the frame and the
// credit, and that is exactly what this shows.
const footageVizProps: LessonVideoProps = {
  theme: {primary: '#5b5bd6', accent: '#f5c26b', background: '#0b0d12', courseName: 'Coursesmith'},
  audioFile: '',
  captions: [],
  durationMs: 4000,
  scenes: [
    {
      type: 'footage',
      startMs: 0,
      endMs: 4000,
      props: {
        title: 'https://lovable.dev',
        origin: 'https://lovable.dev',
        frames: [{mark: 'app-built', path: 'footage-placeholder.png'}],
        provenance: {
          tool: 'Lovable',
          realMs: 96000,
          shownMs: 21000,
        },
      },
    },
  ],
};

// ---------------------------------------------------------------------------
// The v7 `foundations` batch: thirty templates built for a CS-foundations
// course. Every fixture below is a real frame from that course — a real base
// conversion, a real truth table, a real UTF-8 encoding — because the gallery
// card is the only thing a creator sees before picking a template, and a card
// full of placeholder words teaches nothing about what the template is for.
//
// All thirty run under the `editorial` skin, the way the v5 batch does: the
// batch and the gallery arrived together, so these are its baselines.
const V7_THEME = {
  primary: '#306998',
  accent: '#ffd43b',
  background: '#ffffff',
  courseName: 'Coursesmith',
  skin: 'editorial' as const,
  accentQuantity: '#f5ca47',
  accentLimit: '#ec5b51',
  accentRival: '#518cec',
};

const syllabusVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 32000,
  scenes: [{type: 'syllabus', startMs: 0, endMs: 32000, props: {
    title: 'Eight modules, and where you are',
    emphasis: 'where you are', emphasisRole: 'quantity',
    modules: [
      {label: 'Computer Fundamentals', sub: 'what a computer actually is'},
      {label: 'Binary & Digital Systems', sub: 'every number is a switch'},
      {label: 'Computer Architecture', sub: 'CPU, RAM, and the bus between'},
      {label: 'Memory & Storage', sub: 'the hierarchy under your files'},
      {label: 'Operating Systems', sub: 'who decides what runs next'},
      {label: 'Networking & Internet', sub: 'a packet leaves your laptop'},
      {label: 'Algorithms & Complexity', sub: 'why one loop beats two'},
      {label: 'Git & Professional Development', sub: 'the history of your work'},
    ],
    current: 4,
    steps: [
      {startMs: 0, endMs: 6000, show: 'map', ticked: []},
      {startMs: 6000, endMs: 11000, show: 'stop', at: 0, ticked: []},
      {startMs: 11000, endMs: 16000, show: 'stop', at: 1, ticked: []},
      {startMs: 16000, endMs: 21000, show: 'stop', at: 2, ticked: []},
      {startMs: 21000, endMs: 26000, show: 'stop', at: 3, ticked: []},
      {startMs: 26000, endMs: 32000, show: 'here', ticked: [0, 1, 2, 3]},
    ],
  }}],
  captions: [],
};

const outcomeVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 30000,
  scenes: [{type: 'outcome', startMs: 0, endMs: 30000, props: {
    title: 'Three things you can do after this',
    emphasis: 'Three things', emphasisRole: 'quantity',
    lesson: 'Binary & Digital Systems',
    abilities: [
      {skill: 'Convert 172 to binary by hand', payoff: 'subnet masks stop being magic numbers'},
      {skill: 'Read a hex dump without a table', payoff: 'a corrupted file tells you where it broke'},
      {skill: 'Size a value to the right integer type', payoff: 'the overflow bug never ships'},
    ],
    steps: [
      {startMs: 0, endMs: 6000, show: 'promise', lit: []},
      {startMs: 6000, endMs: 13000, show: 'ability', at: 0, lit: [0]},
      {startMs: 13000, endMs: 19000, show: 'ability', at: 1, lit: [0, 1]},
      {startMs: 19000, endMs: 25000, show: 'ability', at: 2, lit: [0, 1, 2]},
      {startMs: 25000, endMs: 30000, show: 'contract', lit: [0, 1, 2]},
    ],
  }}],
  captions: [],
};

const bridgeVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 28000,
  scenes: [{type: 'bridge', startMs: 0, endMs: 28000, props: {
    title: 'From bits to the machine that moves them',
    emphasis: 'the machine that moves them', emphasisRole: 'rival',
    from: 'Binary & Digital Systems',
    to: 'Computer Architecture',
    established: [
      'A byte is eight switches, nothing more',
      'Hex is four bits written as one digit',
      'Every instruction is a number too',
    ],
    gap: 'So what reads those numbers, and in what order?',
    steps: [
      {startMs: 0, endMs: 6000, show: 'back', carried: [], gapOpen: false, arrived: false},
      {startMs: 6000, endMs: 11000, show: 'carry', at: 0, carried: [0], gapOpen: false, arrived: false},
      {startMs: 11000, endMs: 15500, show: 'carry', at: 1, carried: [0, 1], gapOpen: false, arrived: false},
      {startMs: 15500, endMs: 20000, show: 'carry', at: 2, carried: [0, 1, 2], gapOpen: false, arrived: false},
      {startMs: 20000, endMs: 24000, show: 'gap', carried: [0, 1, 2], gapOpen: true, arrived: false},
      {startMs: 24000, endMs: 28000, show: 'ahead', carried: [0, 1, 2], gapOpen: true, arrived: true},
    ],
  }}],
  captions: [],
};

const drillVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 28000,
  scenes: [{type: 'drill', startMs: 0, endMs: 28000, props: {
    title: 'How many bytes does a 32-bit address bus reach?',
    emphasis: '32-bit address bus', emphasisRole: 'quantity',
    question: 'How much memory can a 32-bit address bus address?',
    options: ['32 bytes', '4 gigabytes', '32 gigabytes', '4 terabytes'],
    answer: 1,
    why: 'Two to the thirty-second addresses, one byte each, is 4 GiB',
    steps: [
      {startMs: 0, endMs: 7000, show: 'ask', struck: [], revealed: false, whyOn: false},
      {startMs: 7000, endMs: 12000, show: 'eliminate', at: 0, struck: [0], revealed: false, whyOn: false},
      {startMs: 12000, endMs: 16500, show: 'eliminate', at: 2, struck: [0, 2], revealed: false, whyOn: false},
      {startMs: 16500, endMs: 20000, show: 'eliminate', at: 3, struck: [0, 2, 3], revealed: false, whyOn: false},
      {startMs: 20000, endMs: 24000, show: 'reveal', struck: [0, 2, 3], revealed: true, whyOn: false},
      {startMs: 24000, endMs: 28000, show: 'why', struck: [0, 2, 3], revealed: true, whyOn: true},
    ],
  }}],
  captions: [],
};

const labcardVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 32000,
  scenes: [{type: 'labcard', startMs: 0, endMs: 32000, props: {
    title: 'Watch a packet leave your laptop',
    emphasis: 'leave your laptop', emphasisRole: 'rival',
    task: 'Trace the route from your laptop to a server',
    tools: [{name: 'Terminal'}, {name: 'traceroute'}, {name: 'dig'}],
    stepList: [
      {n: 1, text: 'Run dig example.com and note the A record'},
      {n: 2, text: 'Run traceroute to that IP address'},
      {n: 3, text: 'Count the hops before the first timeout'},
      {n: 4, text: 'Re-run it over your phone hotspot'},
    ],
    expect: 'The first hop is your router, on 192.168.x.x',
    steps: [
      {startMs: 0, endMs: 6000, show: 'task', reached: []},
      {startMs: 6000, endMs: 12000, show: 'step', at: 0, reached: [0]},
      {startMs: 12000, endMs: 18000, show: 'step', at: 1, reached: [0, 1]},
      {startMs: 18000, endMs: 23000, show: 'step', at: 2, reached: [0, 1, 2]},
      {startMs: 23000, endMs: 28000, show: 'step', at: 3, reached: [0, 1, 2, 3]},
      {startMs: 28000, endMs: 32000, show: 'expect', reached: [0, 1, 2, 3]},
    ],
  }}],
  captions: [],
};

const missionVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 32000,
  scenes: [{type: 'mission', startMs: 0, endMs: 32000, props: {
    title: 'Build the tool that reads your disk',
    emphasis: 'reads your disk', emphasisRole: 'rival',
    goal: 'Write a script that reports every drive and its free space',
    specs: [
      'Lists every mounted volume',
      'Prints sizes in GiB, not bytes',
      'Flags any volume over 90 percent full',
      'Exits non-zero when a flag fires',
    ],
    deliverable: 'a CLI script and its README',
    done: 'a full disk makes the script exit 1',
    steps: [
      {startMs: 0, endMs: 6000, show: 'brief', landed: [], deliverableOn: false, doneOn: false},
      {startMs: 6000, endMs: 11000, show: 'spec', at: 0, landed: [0], deliverableOn: false, doneOn: false},
      {startMs: 11000, endMs: 15500, show: 'spec', at: 1, landed: [0, 1], deliverableOn: false, doneOn: false},
      {startMs: 15500, endMs: 20000, show: 'spec', at: 2, landed: [0, 1, 2], deliverableOn: false, doneOn: false},
      {startMs: 20000, endMs: 24000, show: 'spec', at: 3, landed: [0, 1, 2, 3], deliverableOn: false, doneOn: false},
      {startMs: 24000, endMs: 28000, show: 'deliverable', landed: [0, 1, 2, 3], deliverableOn: true, doneOn: false},
      {startMs: 28000, endMs: 32000, show: 'done', landed: [0, 1, 2, 3], deliverableOn: true, doneOn: true},
    ],
  }}],
  captions: [],
};

const machineVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 34000,
  scenes: [{type: 'machine', startMs: 0, endMs: 34000, props: {
    title: 'Open the box: five parts, five jobs',
    emphasis: 'five parts, five jobs', emphasisRole: 'quantity',
    chassis: 'a desktop PC',
    parts: [
      {label: 'CPU', job: 'fetches, decodes and executes instructions', size: 'large'},
      {label: 'RAM', job: 'holds what the CPU is working on right now', size: 'large'},
      {label: 'SSD', job: 'keeps the files when the power goes off', size: 'medium'},
      {label: 'GPU', job: 'does the same arithmetic on thousands of pixels', size: 'medium'},
      {label: 'Power supply', job: 'turns wall AC into the 12V the board wants', size: 'small'},
    ],
    steps: [
      {startMs: 0, endMs: 6000, show: 'whole', visited: []},
      {startMs: 6000, endMs: 11000, show: 'part', at: 0, visited: [0]},
      {startMs: 11000, endMs: 16000, show: 'part', at: 1, visited: [0, 1]},
      {startMs: 16000, endMs: 21000, show: 'part', at: 2, visited: [0, 1, 2]},
      {startMs: 21000, endMs: 26000, show: 'part', at: 3, visited: [0, 1, 2, 3]},
      {startMs: 26000, endMs: 30000, show: 'part', at: 4, visited: [0, 1, 2, 3, 4]},
      {startMs: 30000, endMs: 34000, show: 'fit', visited: [0, 1, 2, 3, 4]},
    ],
  }}],
  captions: [],
};

const blueprintVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 34000,
  scenes: [{type: 'blueprint', startMs: 0, endMs: 34000, props: {
    title: 'The bus is the whole architecture',
    emphasis: 'the whole architecture', emphasisRole: 'rival',
    blocks: [
      {id: 'cpu', label: 'CPU', role: 'unit'},
      {id: 'ram', label: 'RAM', role: 'store'},
      {id: 'ssd', label: 'SSD', role: 'store'},
      {id: 'io', label: 'I/O controller', role: 'io'},
    ],
    wires: [
      {from: 0, to: 1, label: 'address bus'},
      {from: 1, to: 0, label: 'data bus'},
      {from: 0, to: 3, label: 'control bus'},
      {from: 3, to: 2, label: 'SATA'},
    ],
    steps: [
      {startMs: 0, endMs: 6000, show: 'board', lit: []},
      {startMs: 6000, endMs: 10000, show: 'block', at: 0, lit: []},
      {startMs: 10000, endMs: 15000, show: 'path', at: 0, lit: [0]},
      {startMs: 15000, endMs: 20000, show: 'path', at: 1, lit: [0, 1]},
      {startMs: 20000, endMs: 25000, show: 'path', at: 2, lit: [0, 1, 2]},
      {startMs: 25000, endMs: 30000, show: 'path', at: 3, lit: [0, 1, 2, 3]},
      {startMs: 30000, endMs: 34000, show: 'whole', lit: [0, 1, 2, 3]},
    ],
  }}],
  captions: [],
};

const relayVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 38000,
  scenes: [{type: 'relay', startMs: 0, endMs: 38000, props: {
    title: 'Power on to login prompt, in six legs',
    emphasis: 'six legs', emphasisRole: 'quantity',
    stages: [
      {label: 'Power on', does: 'holds the CPU in reset until the rails settle', hands: 'the reset vector'},
      {label: 'Firmware', does: 'runs POST, then finds a bootable device', hands: 'the boot sector'},
      {label: 'Bootloader', does: 'loads the kernel image and the initramfs', hands: 'a kernel in RAM'},
      {label: 'Kernel', does: 'mounts the root filesystem and starts drivers', hands: 'process number one'},
      {label: 'Init', does: 'starts services in dependency order', hands: 'a login prompt'},
      {label: 'Login shell', does: 'authenticates you and starts your session', hands: ''},
    ],
    steps: [
      {startMs: 0, endMs: 6000, show: 'line', lit: []},
      {startMs: 6000, endMs: 11000, show: 'ignite', at: 0, lit: [0]},
      {startMs: 11000, endMs: 15500, show: 'ignite', at: 1, from: 0, lit: [0, 1]},
      {startMs: 15500, endMs: 20000, show: 'ignite', at: 2, from: 1, lit: [0, 1, 2]},
      {startMs: 20000, endMs: 25000, show: 'ignite', at: 3, from: 2, lit: [0, 1, 2, 3]},
      {startMs: 25000, endMs: 29500, show: 'ignite', at: 4, from: 3, lit: [0, 1, 2, 3, 4]},
      {startMs: 29500, endMs: 34000, show: 'ignite', at: 5, from: 4, lit: [0, 1, 2, 3, 4, 5]},
      {startMs: 34000, endMs: 38000, show: 'chain', lit: [0, 1, 2, 3, 4, 5]},
    ],
  }}],
  captions: [],
};

const layersVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 36000,
  scenes: [{type: 'layers', startMs: 0, endMs: 36000, props: {
    title: 'Nothing above the syscall line touches hardware',
    emphasis: 'the syscall line', emphasisRole: 'limit',
    strata: [
      {label: 'Application', holds: 'the code you actually wrote', above: true},
      {label: 'Standard library', holds: 'printf, malloc, fopen', above: true},
      {label: 'Kernel', holds: 'the scheduler and the filesystem', above: false},
      {label: 'Device drivers', holds: 'one module per piece of hardware', above: false},
      {label: 'Hardware', holds: 'the CPU, the disk, the NIC', above: false},
    ],
    boundary: 1,
    boundaryLabel: 'the syscall line',
    steps: [
      {startMs: 0, endMs: 5000, show: 'stack', lit: [], crossed: false},
      {startMs: 5000, endMs: 9500, show: 'stratum', at: 0, lit: [0], crossed: false},
      {startMs: 9500, endMs: 14000, show: 'stratum', at: 1, lit: [0, 1], crossed: false},
      {startMs: 14000, endMs: 19000, show: 'cross', lit: [0, 1], crossed: true},
      {startMs: 19000, endMs: 23500, show: 'stratum', at: 2, lit: [0, 1, 2], crossed: true},
      {startMs: 23500, endMs: 28000, show: 'stratum', at: 3, lit: [0, 1, 2, 3], crossed: true},
      {startMs: 28000, endMs: 32000, show: 'stratum', at: 4, lit: [0, 1, 2, 3, 4], crossed: true},
      {startMs: 32000, endMs: 36000, show: 'whole', lit: [0, 1, 2, 3, 4], crossed: true},
    ],
  }}],
  captions: [],
};

// The pipeline grid is simulated in Go, one occupancy row per tick. The rows
// below are that simulation's output for five instructions through IF/ID/EX/
// MEM/WB with one load-use bubble — hand-run so the fixture is a real trace
// rather than a plausible-looking one. -1 is an empty stage.
const pipelineVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 34000,
  scenes: [{type: 'pipeline', startMs: 0, endMs: 34000, props: {
    title: 'Five stages, one instruction finishing every tick',
    emphasis: 'every tick', emphasisRole: 'quantity',
    stages: [{name: 'IF'}, {name: 'ID'}, {name: 'EX'}, {name: 'MEM'}, {name: 'WB'}],
    items: [{label: 'lw'}, {label: 'add'}, {label: 'sub'}, {label: 'and'}, {label: 'beq'}],
    stall: 'the load result is not back before add needs it',
    sequentialTicks: 25,
    pipelinedTicks: 9,
    steps: [
      {startMs: 0, endMs: 5000, show: 'empty', occ: [-1, -1, -1, -1, -1], bubble: -1, tick: 0, retired: 0, inFlight: []},
      {startMs: 5000, endMs: 9000, show: 'fill', occ: [0, -1, -1, -1, -1], bubble: -1, tick: 1, retired: 0, inFlight: [0]},
      {startMs: 9000, endMs: 13000, show: 'fill', occ: [1, 0, -1, -1, -1], bubble: -1, tick: 2, retired: 0, inFlight: [0, 1]},
      {startMs: 13000, endMs: 17000, show: 'fill', occ: [2, 1, 0, -1, -1], bubble: -1, tick: 3, retired: 0, inFlight: [0, 1, 2]},
      {startMs: 17000, endMs: 22000, show: 'stall', occ: [2, 1, -1, 0, -1], bubble: 2, tick: 4, retired: 0, inFlight: [0, 1, 2]},
      {startMs: 22000, endMs: 26000, show: 'fill', occ: [3, 2, 1, -1, 0], bubble: -1, tick: 5, retired: 0, inFlight: [0, 1, 2, 3]},
      {startMs: 26000, endMs: 30000, show: 'fill', occ: [4, 3, 2, 1, -1], bubble: -1, tick: 6, retired: 1, inFlight: [1, 2, 3, 4]},
      {startMs: 30000, endMs: 34000, show: 'flow', occ: [4, 3, 2, 1, -1], bubble: -1, tick: 6, retired: 1, inFlight: [1, 2, 3, 4]},
    ],
  }}],
  captions: [],
};

const radixVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 26000,
  scenes: [{type: 'radix', startMs: 0, endMs: 26000, props: {
    title: '172 is 10101100 is AC',
    emphasis: '10101100', emphasisRole: 'quantity',
    decimal: 172,
    story: 'the first octet of 172.16.0.0, a private range',
    hex: 'AC',
    cells: [
      {bit: '1', weight: 128},
      {bit: '0', weight: 64},
      {bit: '1', weight: 32},
      {bit: '0', weight: 16},
      {bit: '1', weight: 8},
      {bit: '1', weight: 4},
      {bit: '0', weight: 2},
      {bit: '0', weight: 1},
    ],
    sumSteps: [
      {at: 0, weight: 128, total: 128},
      {at: 2, weight: 32, total: 160},
      {at: 4, weight: 8, total: 168},
      {at: 5, weight: 4, total: 172},
    ],
    nibbles: [
      {bits: '1010', hex: 'A'},
      {bits: '1100', hex: 'C'},
    ],
    steps: [
      {startMs: 0, endMs: 5000, show: 'value'},
      {startMs: 5000, endMs: 10000, show: 'weights'},
      {startMs: 10000, endMs: 17000, show: 'sum'},
      {startMs: 17000, endMs: 22000, show: 'hex'},
      {startMs: 22000, endMs: 26000, show: 'same'},
    ],
  }}],
  captions: [],
};

// 1011 + 110 = 10001, which is 11 + 6 = 17. The columns below are indexed by
// significance — entry 0 is the RIGHTMOST — and carry out of column 1 rides
// all the way to column 3, which is the hop this template exists to draw.
const carryVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 34000,
  scenes: [{type: 'carry', startMs: 0, endMs: 34000, props: {
    title: 'Eleven plus six, one column at a time',
    emphasis: 'one column at a time', emphasisRole: 'limit',
    a: '1011',
    b: '110',
    sum: '10001',
    aDecimal: 11,
    bDecimal: 6,
    sumDecimal: 17,
    columns: [
      {a: '1', b: '0', carryIn: 0, digit: '1', carryOut: 0},
      {a: '1', b: '1', carryIn: 0, digit: '0', carryOut: 1},
      {a: '0', b: '1', carryIn: 1, digit: '0', carryOut: 1},
      {a: '1', b: '0', carryIn: 1, digit: '0', carryOut: 1},
      {a: '0', b: '0', carryIn: 1, digit: '1', carryOut: 0},
    ],
    steps: [
      {startMs: 0, endMs: 5000, show: 'problem', done: []},
      {startMs: 5000, endMs: 9000, show: 'column', at: 0, done: [0]},
      {startMs: 9000, endMs: 13500, show: 'column', at: 1, done: [0, 1]},
      {startMs: 13500, endMs: 18000, show: 'column', at: 2, done: [0, 1, 2]},
      {startMs: 18000, endMs: 22500, show: 'column', at: 3, done: [0, 1, 2, 3]},
      {startMs: 22500, endMs: 26500, show: 'column', at: 4, done: [0, 1, 2, 3, 4]},
      {startMs: 26500, endMs: 30500, show: 'carrychain', done: [0, 1, 2, 3, 4]},
      {startMs: 30500, endMs: 34000, show: 'answer', done: [0, 1, 2, 3, 4]},
    ],
  }}],
  captions: [],
};

// 0xC0C80000 — the IEEE-754 single-precision encoding of -6.25. Sign, then an
// eight-bit exponent, then a twenty-three bit mantissa: 1 + 8 + 23 tiles the
// row exactly, which is what the validator's interval arithmetic checks.
const bitfieldVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 30000,
  scenes: [{type: 'bitfield', startMs: 0, endMs: 30000, props: {
    title: 'These 32 bits are three fields, not one number',
    emphasis: '32 bits', emphasisRole: 'quantity',
    bits: '11000000110010000000000000000000',
    cells: [
      {bit: '1', index: 0},
      {bit: '1', index: 1},
      {bit: '0', index: 2},
      {bit: '0', index: 3},
      {bit: '0', index: 4},
      {bit: '0', index: 5},
      {bit: '0', index: 6},
      {bit: '0', index: 7},
      {bit: '1', index: 8},
      {bit: '1', index: 9},
      {bit: '0', index: 10},
      {bit: '0', index: 11},
      {bit: '1', index: 12},
      {bit: '0', index: 13},
      {bit: '0', index: 14},
      {bit: '0', index: 15},
      {bit: '0', index: 16},
      {bit: '0', index: 17},
      {bit: '0', index: 18},
      {bit: '0', index: 19},
      {bit: '0', index: 20},
      {bit: '0', index: 21},
      {bit: '0', index: 22},
      {bit: '0', index: 23},
      {bit: '0', index: 24},
      {bit: '0', index: 25},
      {bit: '0', index: 26},
      {bit: '0', index: 27},
      {bit: '0', index: 28},
      {bit: '0', index: 29},
      {bit: '0', index: 30},
      {bit: '0', index: 31},
    ],
    fields: [
      {label: 'sign', from: 0, to: 0, means: 'one, so the number is negative', bits: '1', value: 1},
      {label: 'exponent', from: 1, to: 8, means: '129 less the bias of 127, so times four', bits: '10000001', value: 129},
      {label: 'mantissa', from: 9, to: 31, means: '1.5625, once the hidden leading one is put back', bits: '10010000000000000000000', value: 4718592},
    ],
    steps: [
      {startMs: 0, endMs: 5000, show: 'row', done: []},
      {startMs: 5000, endMs: 9500, show: 'split', done: []},
      {startMs: 9500, endMs: 14500, show: 'field', at: 0, done: [0]},
      {startMs: 14500, endMs: 20000, show: 'field', at: 1, done: [0, 1]},
      {startMs: 20000, endMs: 25500, show: 'field', at: 2, done: [0, 1, 2]},
      {startMs: 25500, endMs: 30000, show: 'read', done: [0, 1, 2]},
    ],
  }}],
  captions: [],
};

// U+1F600 really is F0 9F 98 80. Stripping each byte's marker prefix and
// concatenating the payloads gives 000 011111 011000 000000 — 0x1F600 — which
// is the arithmetic the byte boxes are drawn from.
const encodeVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 32000,
  scenes: [{type: 'encode', startMs: 0, endMs: 32000, props: {
    title: 'One emoji, four bytes',
    emphasis: 'four bytes', emphasisRole: 'quantity',
    glyph: '😀',
    codepoint: 'U+1F600',
    note: 'one character, four bytes, and every one of them is needed',
    bytes: [
      {hex: '0xF0', bits: '11110000', marker: '11110', payload: '000', lead: true},
      {hex: '0x9F', bits: '10011111', marker: '10', payload: '011111', lead: false},
      {hex: '0x98', bits: '10011000', marker: '10', payload: '011000', lead: false},
      {hex: '0x80', bits: '10000000', marker: '10', payload: '000000', lead: false},
    ],
    steps: [
      {startMs: 0, endMs: 5000, show: 'glyph', landed: 0},
      {startMs: 5000, endMs: 10000, show: 'codepoint', landed: 0},
      {startMs: 10000, endMs: 14500, show: 'bytes', at: 0, landed: 1},
      {startMs: 14500, endMs: 18500, show: 'bytes', at: 1, landed: 2},
      {startMs: 18500, endMs: 22500, show: 'bytes', at: 2, landed: 3},
      {startMs: 22500, endMs: 27000, show: 'bytes', at: 3, landed: 4},
      {startMs: 27000, endMs: 32000, show: 'note', landed: 4},
    ],
  }}],
  captions: [],
};

const gatesVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 30000,
  scenes: [{type: 'gates', startMs: 0, endMs: 30000, props: {
    title: 'XOR fires only when the inputs differ',
    emphasis: 'the inputs differ', emphasisRole: 'rival',
    gate: 'XOR',
    law: 'one when the inputs differ',
    inputs: ['A', 'B'],
    rows: [
      {in: [0, 0], out: 0},
      {in: [0, 1], out: 1},
      {in: [1, 0], out: 1},
      {in: [1, 1], out: 0},
    ],
    steps: [
      {startMs: 0, endMs: 5000, show: 'circuit', done: []},
      {startMs: 5000, endMs: 10000, show: 'row', at: 0, done: [0]},
      {startMs: 10000, endMs: 15000, show: 'row', at: 1, done: [0, 1]},
      {startMs: 15000, endMs: 20000, show: 'row', at: 2, done: [0, 1, 2]},
      {startMs: 20000, endMs: 25000, show: 'row', at: 3, done: [0, 1, 2, 3]},
      {startMs: 25000, endMs: 30000, show: 'law', done: [0, 1, 2, 3]},
    ],
  }}],
  captions: [],
};

// The latency numbers every programmer should know, on a log axis running from
// 0.3 ns to 10 ms. `logPos` and the decade ticks are computed in Go and
// rounded to six places; the values below are that computation's output.
const ladderVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 40000,
  scenes: [{type: 'ladder', startMs: 0, endMs: 40000, props: {
    title: 'Every rung down costs an order of magnitude',
    emphasis: 'an order of magnitude', emphasisRole: 'limit',
    rungs: [
      {label: 'registers', capacity: 'a few hundred bytes', latencyNs: 0.3, latency: '0.3 ns', logPos: 0},
      {label: 'L1 cache', capacity: '64 KB per core', latencyNs: 1, latency: '1 ns', logPos: 0.069505},
      {label: 'L2 cache', capacity: '1 MB per core', latencyNs: 7, latency: '7 ns', logPos: 0.181842},
      {label: 'main memory', capacity: '16 GB', latencyNs: 100, latency: '100 ns', logPos: 0.335361},
      {label: 'SSD', capacity: '1 TB', latencyNs: 150000, latency: '150 µs', logPos: 0.757552},
      {label: 'spinning disk', capacity: '4 TB', latencyNs: 10000000, latency: '10 ms', logPos: 1},
    ],
    ticks: [
      {pos: 0.069505, label: '1 ns'},
      {pos: 0.202433, label: '10 ns'},
      {pos: 0.335361, label: '100 ns'},
      {pos: 0.468289, label: '1 µs'},
      {pos: 0.601216, label: '10 µs'},
      {pos: 0.734144, label: '100 µs'},
      {pos: 0.867072, label: '1 ms'},
      {pos: 1, label: '10 ms'},
    ],
    ratio: '×33,333,333',
    steps: [
      {startMs: 0, endMs: 4500, show: 'ladder'},
      {startMs: 4500, endMs: 8500, show: 'rung', at: 0},
      {startMs: 8500, endMs: 12500, show: 'rung', at: 1},
      {startMs: 12500, endMs: 16500, show: 'rung', at: 2},
      {startMs: 16500, endMs: 21000, show: 'rung', at: 3},
      {startMs: 21000, endMs: 26000, show: 'miss', at: 3, to: 4, cost: '×1,500 slower'},
      {startMs: 26000, endMs: 30000, show: 'rung', at: 4},
      {startMs: 30000, endMs: 34500, show: 'rung', at: 5},
      {startMs: 34500, endMs: 40000, show: 'spread'},
    ],
  }}],
  captions: [],
};

const regionsVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 40000,
  scenes: [{type: 'regions', startMs: 0, endMs: 40000, props: {
    title: 'The heap and the stack spend the same gap',
    emphasis: 'the same gap', emphasisRole: 'limit',
    regions: [
      {label: 'code', role: 'code', note: 'the instructions, mapped read-only', grows: ''},
      {label: 'static data', role: 'static', note: 'globals, sized before the program starts', grows: ''},
      {label: 'the heap', role: 'heap', note: 'malloc hands out addresses from here upward', grows: 'up'},
      {label: 'free space', role: 'gap', note: 'unclaimed, and both fronts are spending it', grows: ''},
      {label: 'the stack', role: 'stack', note: 'one frame per call, growing downward', grows: 'down'},
    ],
    heapAt: 2,
    stackAt: 4,
    gapAt: 3,
    steps: [
      {startMs: 0, endMs: 4500, show: 'map', seen: [], grown: [], collided: false},
      {startMs: 4500, endMs: 8500, show: 'region', at: 0, seen: [0], grown: [], collided: false},
      {startMs: 8500, endMs: 12500, show: 'region', at: 1, seen: [0, 1], grown: [], collided: false},
      {startMs: 12500, endMs: 17000, show: 'region', at: 2, seen: [0, 1, 2], grown: [], collided: false},
      {startMs: 17000, endMs: 21000, show: 'region', at: 3, seen: [0, 1, 2, 3], grown: [], collided: false},
      {startMs: 21000, endMs: 25000, show: 'region', at: 4, seen: [0, 1, 2, 3, 4], grown: [], collided: false},
      {startMs: 25000, endMs: 29000, show: 'grow', at: 2, seen: [0, 1, 2, 3, 4], grown: [2], collided: false},
      {startMs: 29000, endMs: 33000, show: 'grow', at: 4, seen: [0, 1, 2, 3, 4], grown: [2, 4], collided: false},
      {startMs: 33000, endMs: 36500, show: 'collide', seen: [0, 1, 2, 3, 4], grown: [2, 4], collided: true},
      {startMs: 36500, endMs: 40000, show: 'whole', seen: [0, 1, 2, 3, 4], grown: [2, 4], collided: true},
    ],
  }}],
  captions: [],
};

const lookupVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 34000,
  scenes: [{type: 'lookup', startMs: 0, endMs: 34000, props: {
    title: 'Four questions to turn a name into an address',
    emphasis: 'Four questions', emphasisRole: 'quantity',
    key: 'www.example.com',
    answer: '93.184.216.34',
    hops: [
      {where: 'your resolver', gives: 'nothing cached, so it starts at the root', miss: 'a cold cache means the full walk'},
      {where: 'the root server', gives: 'go ask the .com servers', miss: 'the root knows no hostnames at all'},
      {where: 'the .com servers', gives: 'go ask ns1.example.com', miss: 'a TLD knows delegations, not addresses'},
      {where: 'the authoritative server', gives: 'www.example.com is 93.184.216.34', miss: ''},
    ],
    steps: [
      {startMs: 0, endMs: 5000, show: 'ask', visited: [], answered: false, cached: false},
      {startMs: 5000, endMs: 9500, show: 'hop', at: 0, visited: [0], answered: false, cached: false},
      {startMs: 9500, endMs: 14000, show: 'hop', at: 1, visited: [0, 1], answered: false, cached: false},
      {startMs: 14000, endMs: 18500, show: 'hop', at: 2, visited: [0, 1, 2], answered: false, cached: false},
      {startMs: 18500, endMs: 23500, show: 'hop', at: 3, visited: [0, 1, 2, 3], answered: false, cached: false},
      {startMs: 23500, endMs: 29000, show: 'hit', visited: [0, 1, 2, 3], answered: true, cached: false},
      {startMs: 29000, endMs: 34000, show: 'cache', visited: [0, 1, 2, 3], answered: true, cached: true},
    ],
  }}],
  captions: [],
};

// The token walk is validated arc by arc: a transition may only fire when it
// starts where the dot is standing. `from` is the node before the beat acts and
// `token` the node after, so the component animates the slide without deriving
// either. This route runs new → ready → running, gets pre-empted, blocks on a
// read, and exits — which is every arc in the machine.
const statesVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 40000,
  scenes: [{type: 'states', startMs: 0, endMs: 40000, props: {
    title: 'A process is only ever in one of five states',
    emphasis: 'one of five states', emphasisRole: 'quantity',
    nodes: [
      {id: 'new', label: 'new'},
      {id: 'ready', label: 'ready'},
      {id: 'running', label: 'running'},
      {id: 'blocked', label: 'blocked'},
      {id: 'terminated', label: 'terminated'},
    ],
    arcs: [
      {from: 0, to: 1, on: 'admitted by the scheduler'},
      {from: 1, to: 2, on: 'the scheduler picks it'},
      {from: 2, to: 3, on: 'it asks for disk'},
      {from: 3, to: 1, on: 'the read completes'},
      {from: 2, to: 1, on: 'its time slice expires'},
      {from: 2, to: 4, on: 'it calls exit'},
    ],
    steps: [
      {startMs: 0, endMs: 4500, show: 'machine', from: 0, token: 0, lit: []},
      {startMs: 4500, endMs: 8500, show: 'fire', at: 0, from: 0, token: 1, lit: [0]},
      {startMs: 8500, endMs: 12500, show: 'fire', at: 1, from: 1, token: 2, lit: [0, 1]},
      {startMs: 12500, endMs: 16500, show: 'fire', at: 4, from: 2, token: 1, lit: [0, 1, 4]},
      {startMs: 16500, endMs: 20500, show: 'fire', at: 1, from: 1, token: 2, lit: [0, 1, 4]},
      {startMs: 20500, endMs: 24500, show: 'fire', at: 2, from: 2, token: 3, lit: [0, 1, 2, 4]},
      {startMs: 24500, endMs: 28500, show: 'fire', at: 3, from: 3, token: 1, lit: [0, 1, 2, 3, 4]},
      {startMs: 28500, endMs: 32000, show: 'fire', at: 1, from: 1, token: 2, lit: [0, 1, 2, 3, 4]},
      {startMs: 32000, endMs: 36000, show: 'fire', at: 5, from: 2, token: 4, lit: [0, 1, 2, 3, 4, 5]},
      {startMs: 36000, endMs: 40000, show: 'walk', from: 4, token: 4, lit: [0, 1, 2, 3, 4, 5]},
    ],
  }}],
  captions: [],
};

const schedulerVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 36000,
  scenes: [{type: 'scheduler', startMs: 0, endMs: 36000, props: {
    title: 'Round Robin: nobody waits more than one quantum',
    emphasis: 'one quantum', emphasisRole: 'limit',
    policy: 'Round Robin',
    procs: [
      {label: 'P1', total: 6},
      {label: 'P2', total: 4},
      {label: 'P3', total: 2},
    ],
    slots: [
      {proc: 0, len: 2, start: 0},
      {proc: 1, len: 2, start: 2},
      {proc: 2, len: 2, start: 4},
      {proc: 0, len: 2, start: 6},
      {proc: 1, len: 2, start: 8},
      {proc: 0, len: 2, start: 10},
    ],
    units: 12,
    steps: [
      {startMs: 0, endMs: 4500, show: 'queue', laid: 0},
      {startMs: 4500, endMs: 8000, show: 'run', at: 0, laid: 1},
      {startMs: 8000, endMs: 11500, show: 'run', at: 1, laid: 2},
      {startMs: 11500, endMs: 16000, show: 'switch', at: 1, boundary: 2, laid: 2},
      {startMs: 16000, endMs: 19500, show: 'run', at: 2, laid: 3},
      {startMs: 19500, endMs: 23000, show: 'run', at: 3, laid: 4},
      {startMs: 23000, endMs: 26500, show: 'run', at: 4, laid: 5},
      {startMs: 26500, endMs: 30500, show: 'run', at: 5, laid: 6},
      {startMs: 30500, endMs: 36000, show: 'fair', laid: 6},
    ],
  }}],
  captions: [],
};

const shellVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 34000,
  scenes: [{type: 'shell', startMs: 0, endMs: 34000, props: {
    title: 'One command says the disk is 88 percent full',
    emphasis: '88 percent full', emphasisRole: 'limit',
    host: 'ubuntu',
    entries: [
      {
        cmd: 'ls -l /var/log',
        output: [
          '-rw-r--r--  1 root root  1.2M Aug  7 09:14 syslog',
          '-rw-r-----  1 root adm    48K Aug  7 09:02 auth.log',
          'drwxr-xr-x  2 root root  4.0K Aug  1 00:00 nginx',
        ],
        note: 'the first column is the permission bits',
      },
      {
        cmd: 'df -h /',
        output: [
          'Filesystem      Size  Used Avail Use% Mounted on',
          '/dev/nvme0n1p2  468G  391G   54G  88% /',
        ],
        note: 'eighty-eight percent is the number to watch',
      },
      {
        cmd: 'chmod 640 /var/log/auth.log',
        output: [],
        note: 'six-four-zero is rw-, then r--, then nothing',
      },
    ],
    steps: [
      {startMs: 0, endMs: 4000, show: 'prompt', typed: [], shown: []},
      {startMs: 4000, endMs: 8000, show: 'type', at: 0, typed: [0], shown: []},
      {startMs: 8000, endMs: 13000, show: 'output', at: 0, typed: [0], shown: [0]},
      {startMs: 13000, endMs: 17000, show: 'type', at: 1, typed: [0, 1], shown: [0]},
      {startMs: 17000, endMs: 23000, show: 'output', at: 1, typed: [0, 1], shown: [0, 1]},
      {startMs: 23000, endMs: 29000, show: 'type', at: 2, typed: [0, 1, 2], shown: [0, 1]},
      {startMs: 29000, endMs: 34000, show: 'recap', typed: [0, 1, 2], shown: [0, 1, 2]},
    ],
  }}],
  captions: [],
};

const journeyVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 30000,
  scenes: [{type: 'journey', startMs: 0, endMs: 30000, props: {
    title: 'Four stops between your keystroke and the page',
    emphasis: 'Four stops', emphasisRole: 'quantity',
    stops: [
      {label: 'your laptop', kind: 'device', adds: 'opens a socket and writes the request'},
      {label: 'the hall router', kind: 'router', adds: 'rewrites the source address and forwards'},
      {label: 'the DNS resolver', kind: 'dns', adds: 'turns example.com into 93.184.216.34'},
      {label: 'the web server', kind: 'server', adds: 'renders the page and writes it back'},
    ],
    return: 'the HTML for the page',
    steps: [
      {startMs: 0, endMs: 6000, show: 'map', reached: 0, legs: []},
      {startMs: 6000, endMs: 12000, show: 'hop', at: 1, reached: 1, legs: [1]},
      {startMs: 12000, endMs: 18000, show: 'hop', at: 2, reached: 2, legs: [1, 2]},
      {startMs: 18000, endMs: 24000, show: 'reach', at: 3, reached: 3, legs: [1, 2, 3]},
      {startMs: 24000, endMs: 30000, show: 'return', reached: 3, legs: [1, 2, 3]},
    ],
  }}],
  captions: [],
};

const handshakeVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 28000,
  scenes: [{type: 'handshake', startMs: 0, endMs: 28000, props: {
    title: 'Three packets before a single byte of data',
    emphasis: 'Three packets', emphasisRole: 'quantity',
    left: 'your browser',
    right: 'the server',
    msgs: [
      {dir: 'ltr', label: 'SYN, sequence 0', means: 'I want to talk, and here is my start'},
      {dir: 'rtl', label: 'SYN-ACK, sequence 0, ack 1', means: 'so do I, and I heard yours'},
      {dir: 'ltr', label: 'ACK, ack 1', means: 'I heard yours too, so we are open'},
    ],
    steps: [
      {startMs: 0, endMs: 5000, show: 'wire', delivered: []},
      {startMs: 5000, endMs: 11000, show: 'msg', at: 0, delivered: [0]},
      {startMs: 11000, endMs: 17000, show: 'msg', at: 1, delivered: [0, 1]},
      {startMs: 17000, endMs: 23000, show: 'msg', at: 2, delivered: [0, 1, 2]},
      {startMs: 23000, endMs: 28000, show: 'open', delivered: [0, 1, 2]},
    ],
  }}],
  captions: [],
};

// A real binary search for 31 in a sorted row of eight: mid lands on 17, which
// is too small, low moves past it, and the second probe is the answer. Two
// comparisons for eight cells — which is the whole point of the picture.
const stepperVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 32000,
  scenes: [{type: 'stepper', startMs: 0, endMs: 32000, props: {
    title: 'Eight cells, two comparisons, one answer',
    emphasis: 'two comparisons', emphasisRole: 'quantity',
    values: [3, 8, 12, 17, 23, 31, 42, 56],
    pointers: ['low', 'mid', 'high'],
    target: 31,
    steps: [
      {startMs: 0, endMs: 5000, show: 'array', values: [3, 8, 12, 17, 23, 31, 42, 56],
       ops: 0, ptr: {low: 0, mid: -1, high: 7}, touched: []},
      {startMs: 5000, endMs: 9000, show: 'point', values: [3, 8, 12, 17, 23, 31, 42, 56],
       ops: 0, ptr: {low: 0, mid: 3, high: 7}, touched: []},
      {startMs: 9000, endMs: 14000, show: 'compare', at: [3], values: [3, 8, 12, 17, 23, 31, 42, 56],
       ops: 1, ptr: {low: 0, mid: 3, high: 7}, touched: [3]},
      {startMs: 14000, endMs: 18000, show: 'point', values: [3, 8, 12, 17, 23, 31, 42, 56],
       ops: 1, ptr: {low: 4, mid: 3, high: 7}, touched: [3]},
      {startMs: 18000, endMs: 22000, show: 'point', values: [3, 8, 12, 17, 23, 31, 42, 56],
       ops: 1, ptr: {low: 4, mid: 5, high: 7}, touched: [3]},
      {startMs: 22000, endMs: 27000, show: 'compare', at: [5], values: [3, 8, 12, 17, 23, 31, 42, 56],
       ops: 2, ptr: {low: 4, mid: 5, high: 7}, touched: [3, 5]},
      {startMs: 27000, endMs: 32000, show: 'found', at: [5], values: [3, 8, 12, 17, 23, 31, 42, 56],
       ops: 2, ptr: {low: 4, mid: 5, high: 7}, touched: [3, 5]},
    ],
  }}],
  captions: [],
};

// The curves are sampled in Go: 24 points at n = 1..24, each cost divided by a
// ceiling of 40 and clamped to the top of the frame. O(n²) leaves the chart at
// n = 7, which is the shot. The probe readings are the real costs at a million.
const growthVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 32000,
  scenes: [{type: 'growth', startMs: 0, endMs: 32000, props: {
    title: 'Doubling the input is not always doubling the work',
    emphasis: 'not always doubling the work', emphasisRole: 'limit',
    curves: [
      {
        class: 'logn', label: 'binary search', notation: 'O(log n)', reading: '20',
        points: [0, 0.025, 0.0396, 0.05, 0.058, 0.0646, 0.0702, 0.075, 0.0792, 0.083, 0.0865, 0.0896,
                 0.0925, 0.0952, 0.0977, 0.1, 0.1022, 0.1042, 0.1062, 0.108, 0.1098, 0.1115, 0.1131, 0.1146],
      },
      {
        class: 'n', label: 'the linear scan', notation: 'O(n)', reading: '1,000,000',
        points: [0.025, 0.05, 0.075, 0.1, 0.125, 0.15, 0.175, 0.2, 0.225, 0.25, 0.275, 0.3,
                 0.325, 0.35, 0.375, 0.4, 0.425, 0.45, 0.475, 0.5, 0.525, 0.55, 0.575, 0.6],
      },
      {
        class: 'n2', label: 'the nested loop', notation: 'O(n²)', reading: '1,000,000,000,000',
        points: [0.025, 0.1, 0.225, 0.4, 0.625, 0.9, 1, 1, 1, 1, 1, 1,
                 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1],
      },
    ],
    probe: 1000000,
    probeLabel: '1,000,000',
    worst: 2,
    steps: [
      {startMs: 0, endMs: 5000, show: 'axes', drawn: []},
      {startMs: 5000, endMs: 10000, show: 'curve', at: 0, drawn: [0]},
      {startMs: 10000, endMs: 15000, show: 'curve', at: 1, drawn: [0, 1]},
      {startMs: 15000, endMs: 20500, show: 'curve', at: 2, drawn: [0, 1, 2]},
      {startMs: 20500, endMs: 27000, show: 'probe', drawn: [0, 1, 2]},
      {startMs: 27000, endMs: 32000, show: 'moral', drawn: [0, 1, 2]},
    ],
  }}],
  captions: [],
};

// factorial(4). Four plates go on in call order and come off in reverse, each
// handing its value down into the plate below — 1, then 2, then 6, then 24.
// `onStack` and `returned` are the Go simulation's output, not a guess.
const callstackVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 40000,
  scenes: [{type: 'callstack', startMs: 0, endMs: 40000, props: {
    title: 'factorial(4) is four calls before a single answer',
    emphasis: 'four calls', emphasisRole: 'quantity',
    fn: 'factorial',
    base: 'one is where it stops calling',
    answer: '24',
    frames: [
      {args: 'n=4', returns: '24', base: false},
      {args: 'n=3', returns: '6', base: false},
      {args: 'n=2', returns: '2', base: false},
      {args: 'n=1', returns: '1', base: true},
    ],
    steps: [
      {startMs: 0, endMs: 4000, show: 'call', at: 0, onStack: [0], returned: []},
      {startMs: 4000, endMs: 8000, show: 'call', at: 1, onStack: [0, 1], returned: []},
      {startMs: 8000, endMs: 12000, show: 'call', at: 2, onStack: [0, 1, 2], returned: []},
      {startMs: 12000, endMs: 16000, show: 'call', at: 3, onStack: [0, 1, 2, 3], returned: []},
      {startMs: 16000, endMs: 21000, show: 'base', at: 3, onStack: [0, 1, 2, 3], returned: []},
      {startMs: 21000, endMs: 25000, show: 'return', at: 3, value: '1', into: 2, onStack: [0, 1, 2], returned: [3]},
      {startMs: 25000, endMs: 29000, show: 'return', at: 2, value: '2', into: 1, onStack: [0, 1], returned: [2, 3]},
      {startMs: 29000, endMs: 33000, show: 'return', at: 1, value: '6', into: 0, onStack: [0], returned: [1, 2, 3]},
      {startMs: 33000, endMs: 36500, show: 'return', at: 0, value: '24', onStack: [], returned: [0, 1, 2, 3]},
      {startMs: 36500, endMs: 40000, show: 'empty', onStack: [], returned: [0, 1, 2, 3]},
    ],
  }}],
  captions: [],
};

const historyVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 38000,
  scenes: [{type: 'history', startMs: 0, endMs: 38000, props: {
    title: 'A branch is two edges leaving one commit',
    emphasis: 'two edges', emphasisRole: 'quantity',
    lanes: ['main', 'feature'],
    commits: [
      {col: 0, lane: 0, label: 'initial commit', parents: [], children: [1], merge: false},
      {col: 1, lane: 0, label: 'add the parser', parents: [0], children: [2, 3], merge: false},
      {col: 2, lane: 1, label: 'start the cache', parents: [1], children: [4], merge: false},
      {col: 3, lane: 0, label: 'fix the parser', parents: [1], children: [5], merge: false},
      {col: 4, lane: 1, label: 'cache lookups', parents: [2], children: [5], merge: false},
      {col: 5, lane: 0, label: 'merge the cache', parents: [3, 4], children: [], merge: true},
    ],
    edges: [
      {from: 0, to: 1, fromLane: 0, toLane: 0, curved: false},
      {from: 1, to: 2, fromLane: 0, toLane: 1, curved: true},
      {from: 1, to: 3, fromLane: 0, toLane: 0, curved: false},
      {from: 2, to: 4, fromLane: 1, toLane: 1, curved: false},
      {from: 3, to: 5, fromLane: 0, toLane: 0, curved: false},
      {from: 4, to: 5, fromLane: 1, toLane: 0, curved: true},
    ],
    steps: [
      {startMs: 0, endMs: 4500, show: 'graph', landed: [], head: -1},
      {startMs: 4500, endMs: 8500, show: 'commit', at: 0, landed: [0], head: 0},
      {startMs: 8500, endMs: 12500, show: 'commit', at: 1, landed: [0, 1], head: 1},
      {startMs: 12500, endMs: 16500, show: 'commit', at: 2, landed: [0, 1, 2], head: 2},
      {startMs: 16500, endMs: 20500, show: 'commit', at: 3, landed: [0, 1, 2, 3], head: 3},
      {startMs: 20500, endMs: 25500, show: 'branch', at: 1, kids: [2, 3], landed: [0, 1, 2, 3], head: 3},
      {startMs: 25500, endMs: 29500, show: 'commit', at: 4, landed: [0, 1, 2, 3, 4], head: 4},
      {startMs: 29500, endMs: 34000, show: 'merge', at: 5, landed: [0, 1, 2, 3, 4, 5], head: 5},
      {startMs: 34000, endMs: 38000, show: 'log', landed: [0, 1, 2, 3, 4, 5], head: 5},
    ],
  }}],
  captions: [],
};

const versusVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 34000,
  scenes: [{type: 'versus', startMs: 0, endMs: 34000, props: {
    title: 'TCP guarantees order; UDP guarantees nothing',
    emphasis: 'UDP guarantees nothing', emphasisRole: 'rival',
    left: 'TCP',
    right: 'UDP',
    rows: [
      {dim: 'ordering', leftVal: 'bytes arrive in order', rightVal: 'datagrams can arrive out of order', edge: 'left'},
      {dim: 'delivery', leftVal: 'retransmits until acknowledged', rightVal: 'sends once, no retries', edge: 'left'},
      {dim: 'setup cost', leftVal: 'three packets before data', rightVal: 'no handshake at all', edge: 'right'},
      {dim: 'head-of-line blocking', leftVal: 'one lost packet stalls everything', rightVal: 'a lost datagram stalls nothing', edge: 'right'},
      {dim: 'header overhead', leftVal: 'twenty bytes per segment', rightVal: 'eight bytes per datagram', edge: 'right'},
    ],
    verdict: 'Reach for TCP when every byte matters, UDP when late data is useless',
    leftWins: 2,
    rightWins: 3,
    evens: 0,
    sweep: false,
    steps: [
      {startMs: 0, endMs: 5000, show: 'face', landed: []},
      {startMs: 5000, endMs: 9500, show: 'row', at: 0, landed: [0]},
      {startMs: 9500, endMs: 14000, show: 'row', at: 1, landed: [0, 1]},
      {startMs: 14000, endMs: 18500, show: 'row', at: 2, landed: [0, 1, 2]},
      {startMs: 18500, endMs: 23000, show: 'row', at: 3, landed: [0, 1, 2, 3]},
      {startMs: 23000, endMs: 28000, show: 'row', at: 4, landed: [0, 1, 2, 3, 4]},
      {startMs: 28000, endMs: 34000, show: 'verdict', landed: [0, 1, 2, 3, 4]},
    ],
  }}],
  captions: [],
};

const erasVizProps: LessonVideoProps = {
  theme: V7_THEME,
  audioFile: '',
  durationMs: 38000,
  scenes: [{type: 'eras', startMs: 0, endMs: 38000, props: {
    title: 'Five generations, each handing the next a problem',
    emphasis: 'Five generations', emphasisRole: 'quantity',
    eras: [
      {label: 'vacuum tubes', when: '1940s', mark: 'ENIAC filled a room and drew 150 kilowatts', carry: 'the stored-program idea'},
      {label: 'transistors', when: '1950s', mark: 'the TX-0 ran without warming up first', carry: 'switching without heat'},
      {label: 'integrated circuits', when: '1960s', mark: 'the System/360 made a family of machines', carry: 'one instruction set, many machines'},
      {label: 'microprocessors', when: '1970s', mark: 'the Intel 4004 put a CPU on one chip', carry: 'a computer per person'},
      {label: 'the internet', when: '1990s', mark: 'TCP/IP made every machine reachable', carry: 'a computer per pocket'},
    ],
    threads: [
      {from: 0, to: 1, carry: 'the stored-program idea'},
      {from: 1, to: 2, carry: 'switching without heat'},
      {from: 2, to: 3, carry: 'one instruction set, many machines'},
      {from: 3, to: 4, carry: 'a computer per person'},
    ],
    carryNow: 'a computer per pocket',
    steps: [
      {startMs: 0, endMs: 4500, show: 'band', lit: []},
      {startMs: 4500, endMs: 9000, show: 'era', at: 0, lit: [0]},
      {startMs: 9000, endMs: 13500, show: 'era', at: 1, lit: [0, 1]},
      {startMs: 13500, endMs: 18000, show: 'era', at: 2, lit: [0, 1, 2]},
      {startMs: 18000, endMs: 22500, show: 'era', at: 3, lit: [0, 1, 2, 3]},
      {startMs: 22500, endMs: 27000, show: 'era', at: 4, lit: [0, 1, 2, 3, 4]},
      {startMs: 27000, endMs: 32500, show: 'thread', lit: [0, 1, 2, 3, 4]},
      {startMs: 32500, endMs: 38000, show: 'now', at: 4, lit: [0, 1, 2, 3, 4]},
    ],
  }}],
  captions: [],
};

/**
 * The showroom skin's palette, copied out of the Go derivation rather than
 * invented here.
 *
 * Every other fixture on this page passes the three course colours and lets
 * resolveTheme fill the rest, which works because those defaults are the dark
 * stage. A showroom fixture cannot do that: leave the tokens out and it renders
 * white cards on a near-black default, which is neither look. So the values are
 * the exact output of deriveVideoThemeSkinned(primary #306998, skin showroom) —
 * if that derivation moves, these go stale silently, and the baselines are what
 * will say so.
 */
const showroomTheme = {
  primary: '#306998',
  accent: '#2563eb',
  background: '#ffffff',
  courseName: 'Coursesmith',
  skin: 'showroom' as const,
  mode: 'light' as const,
  air: 0.07,
  bgTop: '#f6f7f7',
  bgBottom: '#f2f3f3',
  surface: '#ffffff',
  surfaceBorder: '#e3e6e8',
  text: '#141a1f',
  textMuted: '#5c6770',
  accentQuantity: '#a45c09',
  accentLimit: '#c12115',
  accentRival: '#1557c1',
  mass: '#7d91a1',
  ink: '#222f3a',
  shadow: '#243442',
  shadowStrength: 0.1,
  rim: '#ffffff',
  accentText: '#2563eb',
  grain: 0.005,
};

/**
 * The cards template on the clip that motivated it: three assistants side by
 * side, each wearing its own mark.
 *
 * Three cards rather than two, and the third one deliberately without a logo.
 * That is the common case in the wild and the fixture has to hold it: Simple
 * Icons carries Claude and Gemini and does not carry OpenAI's marks at all, so a
 * row about coding assistants CANNOT be all-logos however well the fetch works.
 * A baseline showing three perfect brand marks would be a picture of the lucky
 * case, and the mixed row — two marks and a glyph — is what the template
 * actually has to make look deliberate.
 *
 * The path data and the tints are both real, exactly as snippet_cards_art.go
 * resolves them: geometry out of the fetched document and the brand's own hex out
 * of its fill attribute. Claude's #D97757 and Gemini's #8E75B2 both clear the
 * readability threshold on a white card, so both get the wash lockup — which is
 * why this fixture does NOT cover the plate case, and a brand whose colour is too
 * pale to paint with is worth a second row here if one ever ships.
 */
const cardsVizProps: LessonVideoProps = {
  theme: showroomTheme,
  audioFile: '',
  durationMs: 40000,
  scenes: [
    {
      type: 'cards',
      startMs: 0,
      endMs: 40000,
      props: {
        title: 'Three assistants, one coding job',
        emphasis: 'one coding job',
        // No role, deliberately. The three semantic accents are claims about
        // meaning — the measured number, the ceiling, the alternative — and "one
        // coding job" is none of those, it is just the phrase the sentence turns
        // on. An unroled emphasis takes the brand accent, which makes no claim.
        relation: 'versus',
        ask: 'strongest at',
        closer: 'Claude for long refactors, Gemini for whole repos, ChatGPT for a quick answer',
        items: [
          {
            title: 'Claude',
            note: 'Holding a long argument about a codebase',
            role: 'quantity',
            icon: 'brain',
            tint: '#D97757',
            markFrom: 'simpleicons:claude',
            mark: 'm4.7144 15.9555 4.7174-2.6471.079-.2307-.079-.1275h-.2307l-.7893-.0486-2.6956-.0729-2.3375-.0971-2.2646-.1214-.5707-.1215-.5343-.7042.0546-.3522.4797-.3218.686.0608 1.5179.1032 2.2767.1578 1.6514.0972 2.4468.255h.3886l.0546-.1579-.1336-.0971-.1032-.0972L6.973 9.8356l-2.55-1.6879-1.3356-.9714-.7225-.4918-.3643-.4614-.1578-1.0078.6557-.7225.8803.0607.2246.0607.8925.686 1.9064 1.4754 2.4893 1.8336.3643.3035.1457-.1032.0182-.0728-.164-.2733-1.3539-2.4467-1.445-2.4893-.6435-1.032-.17-.6194c-.0607-.255-.1032-.4674-.1032-.7285L6.287.1335 6.6997 0l.9957.1336.419.3642.6192 1.4147 1.0018 2.2282 1.5543 3.0296.4553.8985.2429.8318.091.255h.1579v-.1457l.1275-1.706.2368-2.0947.2307-2.6957.0789-.7589.3764-.9107.7468-.4918.5828.2793.4797.686-.0668.4433-.2853 1.8517-.5586 2.9021-.3643 1.9429h.2125l.2429-.2429.9835-1.3053 1.6514-2.0643.7286-.8196.85-.9046.5464-.4311h1.0321l.759 1.1293-.34 1.1657-1.0625 1.3478-.8804 1.1414-1.2628 1.7-.7893 1.36.0729.1093.1882-.0183 2.8535-.607 1.5421-.2794 1.8396-.3157.8318.3886.091.3946-.3278.8075-1.967.4857-2.3072.4614-3.4364.8136-.0425.0304.0486.0607 1.5482.1457.6618.0364h1.621l3.0175.2247.7892.522.4736.6376-.079.4857-1.2142.6193-1.6393-.3886-3.825-.9107-1.3113-.3279h-.1822v.1093l1.0929 1.0686 2.0035 1.8092 2.5075 2.3314.1275.5768-.3218.4554-.34-.0486-2.2039-1.6575-.85-.7468-1.9246-1.621h-.1275v.17l.4432.6496 2.3436 3.5214.1214 1.0807-.17.3521-.6071.2125-.6679-.1214-1.3721-1.9246L14.38 17.959l-1.1414-1.9428-.1397.079-.674 7.2552-.3156.3703-.7286.2793-.6071-.4614-.3218-.7468.3218-1.4753.3886-1.9246.3157-1.53.2853-1.9004.17-.6314-.0121-.0425-.1397.0182-1.4328 1.9672-2.1796 2.9446-1.7243 1.8456-.4128.164-.7164-.3704.0667-.6618.4008-.5889 2.386-3.0357 1.4389-1.882.929-1.0868-.0062-.1579h-.0546l-6.3385 4.1164-1.1293.1457-.4857-.4554.0608-.7467.2307-.2429 1.9064-1.3114Z',
          },
          {
            title: 'Gemini',
            note: 'Reading a whole repository at once',
            role: 'rival',
            icon: 'sparkles',
            tint: '#8E75B2',
            markFrom: 'simpleicons:googlegemini',
            mark: 'M11.04 19.32Q12 21.51 12 24q0-2.49.93-4.68.96-2.19 2.58-3.81t3.81-2.55Q21.51 12 24 12q-2.49 0-4.68-.93a12.3 12.3 0 0 1-3.81-2.58 12.3 12.3 0 0 1-2.58-3.81Q12 2.49 12 0q0 2.49-.96 4.68-.93 2.19-2.55 3.81a12.3 12.3 0 0 1-3.81 2.58Q2.49 12 0 12q2.49 0 4.68.96 2.19.93 3.81 2.55t2.55 3.81',
          },
          {
            title: 'ChatGPT',
            note: 'Getting to a working answer fastest',
            role: 'neutral',
            icon: 'message',
          },
        ],
        steps: [
          {startMs: 0, endMs: 7000, show: 'row', lit: [0, 1, 2]},
          {startMs: 7000, endMs: 15500, show: 'card', at: 0, lit: [0]},
          {startMs: 15500, endMs: 24000, show: 'card', at: 1, lit: [1]},
          {startMs: 24000, endMs: 32000, show: 'card', at: 2, lit: [2]},
          {startMs: 32000, endMs: 40000, show: 'all', lit: [0, 1, 2]},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The duel template on the frame that motivated the whole batch: a free tier
 * against a paid one, with what the money buys drawn as the gap between two bars.
 *
 * The scores are the fixture's real content. 42 against 88 is a 46-point gap,
 * which is what the picture is for — and it is worth noting that the clip still
 * picks the PAID side here, so this fixture does not cover the more interesting
 * case where the verdict goes against the longer bar. That case is deliberately
 * allowed by the validator (see snippet_duel.go) and is worth a second fixture if
 * a real clip ever leans on it.
 *
 * ChatGPT gets no mark on purpose, same as in the cards fixture: OpenAI's logos
 * were pulled from Simple Icons on a trademark request, so a duel involving them
 * is the mixed case whether anybody wanted it or not.
 */
const duelVizProps: LessonVideoProps = {
  theme: showroomTheme,
  audioFile: '',
  durationMs: 34000,
  scenes: [
    {
      type: 'duel',
      startMs: 0,
      endMs: 34000,
      props: {
        title: 'Free ChatGPT or paid Gemini',
        emphasis: 'paid Gemini',
        emphasisRole: 'rival',
        axis: 'capability',
        pick: 1,
        verdict: 'Pay for one model if you use it daily; free tiers are for trying it out',
        sides: [
          {
            title: 'ChatGPT',
            tag: 'Free',
            note: 'The most mature chatbot, on its older model',
            // Zero, not 42, and the change is the point of the fixture rather
            // than a tweak to it. A measured zero used to render as an empty
            // track — identical to the frame before anything had been measured —
            // and the first real clip hit it immediately, because "monthly cost"
            // is an obvious axis and it puts a free tier at 0. The baseline now
            // holds the end-cap that says "measured, and it is nothing".
            score: 0,
            role: 'neutral',
            icon: 'message',
            // The favicon path, which no fixture covered until a real render
            // found the bug in it: a fetched favicon brings its own background
            // and it is almost always opaque white, so on the pale tint every
            // other tile gets it drew a white square inside a coloured square.
            // OpenAI is exactly the case that hits this — Simple Icons dropped
            // their marks, so the fetch falls through to the favicon service.
            //
            // Stood in for rather than embedded: what this has to exercise is
            // "an image with its own opaque background", and committing a real
            // brand's bitmap to the repo to prove that would be carrying a
            // trademark around for no extra coverage. The white field and the
            // dark ring are the two things that matter.
            markFrom: 'favicon:example.com',
            image:
              'data:image/svg+xml;base64,' +
              btoa(
                '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">' +
                  '<rect width="64" height="64" fill="#ffffff"/>' +
                  '<circle cx="32" cy="32" r="16" fill="none" stroke="#1b1b1b" stroke-width="7"/>' +
                  '</svg>',
              ),
          },
          {
            title: 'Gemini',
            tag: '$20 a month',
            note: 'Reads a whole repository in one pass',
            score: 88,
            role: 'rival',
            icon: 'sparkles',
            tint: '#8E75B2',
            markFrom: 'simpleicons:googlegemini',
            mark: 'M11.04 19.32Q12 21.51 12 24q0-2.49.93-4.68.96-2.19 2.58-3.81t3.81-2.55Q21.51 12 24 12q-2.49 0-4.68-.93a12.3 12.3 0 0 1-3.81-2.58 12.3 12.3 0 0 1-2.58-3.81Q12 2.49 12 0q0 2.49-.96 4.68-.93 2.19-2.55 3.81a12.3 12.3 0 0 1-3.81 2.58Q2.49 12 0 12q2.49 0 4.68.96 2.19.93 3.81 2.55t2.55 3.81',
          },
        ],
        steps: [
          {startMs: 0, endMs: 6000, show: 'pair', lit: [0, 1], bars: false},
          {startMs: 6000, endMs: 13000, show: 'card', at: 0, lit: [0], bars: false},
          {startMs: 13000, endMs: 20000, show: 'card', at: 1, lit: [1], bars: false},
          {startMs: 20000, endMs: 27000, show: 'bars', lit: [0, 1], bars: true},
          {startMs: 27000, endMs: 34000, show: 'call', lit: [1], bars: true},
        ],
      },
    },
  ],
  captions: [],
};

/**
 * The spotlight template: one tool held up with three claims beside it.
 *
 * Captured on the beat where the second of three rows has just landed, which is
 * the only state that shows all three of the things this template has to get
 * right at once — a settled row, a row mid-arrival, and the empty space the third
 * one has not reached yet.
 */
const spotlightVizProps: LessonVideoProps = {
  theme: showroomTheme,
  audioFile: '',
  durationMs: 30000,
  scenes: [
    {
      type: 'spotlight',
      startMs: 0,
      endMs: 30000,
      props: {
        title: 'What Claude Code is actually for',
        emphasis: 'actually for',
        card: {
          title: 'Claude Code',
          note: 'An agent that works in your terminal',
          role: 'quantity',
          icon: 'terminal',
          tint: '#D97757',
          markFrom: 'simpleicons:claude',
          mark: 'm4.7144 15.9555 4.7174-2.6471.079-.2307-.079-.1275h-.2307l-.7893-.0486-2.6956-.0729-2.3375-.0971-2.2646-.1214-.5707-.1215-.5343-.7042.0546-.3522.4797-.3218.686.0608 1.5179.1032 2.2767.1578 1.6514.0972 2.4468.255h.3886l.0546-.1579-.1336-.0971-.1032-.0972L6.973 9.8356l-2.55-1.6879-1.3356-.9714-.7225-.4918-.3643-.4614-.1578-1.0078.6557-.7225.8803.0607.2246.0607.8925.686 1.9064 1.4754 2.4893 1.8336.3643.3035.1457-.1032.0182-.0728-.164-.2733-1.3539-2.4467-1.445-2.4893-.6435-1.032-.17-.6194c-.0607-.255-.1032-.4674-.1032-.7285L6.287.1335 6.6997 0l.9957.1336.419.3642.6192 1.4147 1.0018 2.2282 1.5543 3.0296.4553.8985.2429.8318.091.255h.1579v-.1457l.1275-1.706.2368-2.0947.2307-2.6957.0789-.7589.3764-.9107.7468-.4918.5828.2793.4797.686-.0668.4433-.2853 1.8517-.5586 2.9021-.3643 1.9429h.2125l.2429-.2429.9835-1.3053 1.6514-2.0643.7286-.8196.85-.9046.5464-.4311h1.0321l.759 1.1293-.34 1.1657-1.0625 1.3478-.8804 1.1414-1.2628 1.7-.7893 1.36.0729.1093.1882-.0183 2.8535-.607 1.5421-.2794 1.8396-.3157.8318.3886.091.3946-.3278.8075-1.967.4857-2.3072.4614-3.4364.8136-.0425.0304.0486.0607 1.5482.1457.6618.0364h1.621l3.0175.2247.7892.522.4736.6376-.079.4857-1.2142.6193-1.6393-.3886-3.825-.9107-1.3113-.3279h-.1822v.1093l1.0929 1.0686 2.0035 1.8092 2.5075 2.3314.1275.5768-.3218.4554-.34-.0486-2.2039-1.6575-.85-.7468-1.9246-1.621h-.1275v.17l.4432.6496 2.3436 3.5214.1214 1.0807-.17.3521-.6071.2125-.6679-.1214-1.3721-1.9246L14.38 17.959l-1.1414-1.9428-.1397.079-.674 7.2552-.3156.3703-.7286.2793-.6071-.4614-.3218-.7468.3218-1.4753.3886-1.9246.3157-1.53.2853-1.9004.17-.6314-.0121-.0425-.1397.0182-1.4328 1.9672-2.1796 2.9446-1.7243 1.8456-.4128.164-.7164-.3704.0667-.6618.4008-.5889 2.386-3.0357 1.4389-1.882.929-1.0868-.0062-.1579h-.0546l-6.3385 4.1164-1.1293.1457-.4857-.4554.0608-.7467.2307-.2429 1.9064-1.3114Z',
        },
        points: [
          {text: 'Refactors across files you never opened', icon: 'code'},
          {text: 'Reads the whole repo before it edits', icon: 'search'},
          {text: 'Runs the tests and fixes what broke', icon: 'check'},
        ],
        steps: [
          {startMs: 0, endMs: 6000, show: 'card', at: -1, shown: 0},
          {startMs: 6000, endMs: 12000, show: 'point', at: 0, shown: 1},
          {startMs: 12000, endMs: 18000, show: 'point', at: 1, shown: 2},
          {startMs: 18000, endMs: 24000, show: 'point', at: 2, shown: 3},
          {startMs: 24000, endMs: 30000, show: 'all', at: -1, shown: 3},
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
      id="FootageViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(footageVizProps.durationMs)}
      defaultProps={footageVizProps}
    />
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
      width={1700}
      height={2280}
      durationInFrames={300}
    />
    <Composition
      id="CastSheet"
      component={CastSheet}
      fps={FPS}
      width={1600}
      height={1360}
      durationInFrames={300}
    />
    <Composition
      id="WorkspaceViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(workspaceVizProps.durationMs)}
      defaultProps={workspaceVizProps}
    />
    <Composition
      id="VSCodeLightViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(vscodeLightProps.durationMs)}
      defaultProps={vscodeLightProps}
    />
    <Composition
      id="TimelineViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(timelineVizProps.durationMs)}
      defaultProps={timelineVizProps}
    />
    <Composition
      id="CanvasViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(canvasVizProps.durationMs)}
      defaultProps={canvasVizProps}
    />
    <Composition
      id="PromptLoopViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(promptLoopVizProps.durationMs)}
      defaultProps={promptLoopVizProps}
    />
    <Composition
      id="MockupViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(mockupVizProps.durationMs)}
      defaultProps={mockupVizProps}
    />
    <Composition
      id="StackViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(stackVizProps.durationMs)}
      defaultProps={stackVizProps}
    />
    <Composition
      id="SpecViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(specVizProps.durationMs)}
      defaultProps={specVizProps}
    />
    <Composition
      id="ConstellationViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(constellationVizProps.durationMs)}
      defaultProps={constellationVizProps}
    />
    <Composition
      id="ChapterViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(chapterVizProps.durationMs)}
      defaultProps={chapterVizProps}
    />
    <Composition
      id="ChapterLightViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(chapterLightProps.durationMs)}
      defaultProps={chapterLightProps}
    />
    <Composition
      id="CycleViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(cycleVizProps.durationMs)}
      defaultProps={cycleVizProps}
    />
    <Composition
      id="CycleLightViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(cycleLightProps.durationMs)}
      defaultProps={cycleLightProps}
    />
    <Composition
      id="ScaleViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(scaleVizProps.durationMs)}
      defaultProps={scaleVizProps}
    />
    <Composition
      id="ScaleLightViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(scaleLightProps.durationMs)}
      defaultProps={scaleLightProps}
    />
    <Composition
      id="CostingViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(costingVizProps.durationMs)}
      defaultProps={costingVizProps}
    />
    <Composition
      id="TraceViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(traceVizProps.durationMs)}
      defaultProps={traceVizProps}
    />
    <Composition
      id="AnalogyViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(analogyVizProps.durationMs)}
      defaultProps={analogyVizProps}
    />
    <Composition
      id="RundownViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(rundownVizProps.durationMs)}
      defaultProps={rundownVizProps}
    />
    <Composition
      id="MythViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(mythVizProps.durationMs)}
      defaultProps={mythVizProps}
    />
    <Composition
      id="DecisionViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(decisionVizProps.durationMs)}
      defaultProps={decisionVizProps}
    />
    <Composition
      id="VerdictViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(verdictVizProps.durationMs)}
      defaultProps={verdictVizProps}
    />
    <Composition
      id="GaugeViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(gaugeVizProps.durationMs)}
      defaultProps={gaugeVizProps}
    />
    <Composition
      id="ToggleViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(toggleVizProps.durationMs)}
      defaultProps={toggleVizProps}
    />
    <Composition
      id="TableViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(tableVizProps.durationMs)}
      defaultProps={tableVizProps}
    />
    <Composition
      id="RatioViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(ratioVizProps.durationMs)}
      defaultProps={ratioVizProps}
    />
    <Composition
      id="MultiplyViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(multiplyVizProps.durationMs)}
      defaultProps={multiplyVizProps}
    />
    <Composition
      id="LatencyViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(latencyVizProps.durationMs)}
      defaultProps={latencyVizProps}
    />
    <Composition
      id="BudgetViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(budgetVizProps.durationMs)}
      defaultProps={budgetVizProps}
    />
    <Composition
      id="CapabilitiesViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(capabilitiesVizProps.durationMs)}
      defaultProps={capabilitiesVizProps}
    />
    <Composition
      id="ForkViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(forkVizProps.durationMs)}
      defaultProps={forkVizProps}
    />
    <Composition
      id="MultiplexViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(multiplexVizProps.durationMs)}
      defaultProps={multiplexVizProps}
    />
    <Composition
      id="JournalViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(journalVizProps.durationMs)}
      defaultProps={journalVizProps}
    />
    <Composition
      id="ObjectiveViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(objectiveVizProps.durationMs)}
      defaultProps={objectiveVizProps}
    />
    <Composition
      id="PrereqViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(prereqVizProps.durationMs)}
      defaultProps={prereqVizProps}
    />
    <Composition
      id="RecapViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(recapVizProps.durationMs)}
      defaultProps={recapVizProps}
    />
    <Composition
      id="PitfallViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(pitfallVizProps.durationMs)}
      defaultProps={pitfallVizProps}
    />
    <Composition
      id="CheckpointViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(checkpointVizProps.durationMs)}
      defaultProps={checkpointVizProps}
    />
    <Composition
      id="RankingViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(rankingVizProps.durationMs)}
      defaultProps={rankingVizProps}
    />
    <Composition
      id="OccupancyViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(occupancyVizProps.durationMs)}
      defaultProps={occupancyVizProps}
    />
    <Composition
      id="MetricViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(metricVizProps.durationMs)}
      defaultProps={metricVizProps}
    />
    <Composition
      id="ShowcaseViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(showcaseVizProps.durationMs)}
      defaultProps={showcaseVizProps}
    />
    <Composition
      id="BreakdownViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(breakdownVizProps.durationMs)}
      defaultProps={breakdownVizProps}
    />
    <Composition
      id="AnatomyViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(anatomyVizProps.durationMs)}
      defaultProps={anatomyVizProps}
    />
    <Composition
      id="CompareViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(compareVizProps.durationMs)}
      defaultProps={compareVizProps}
    />
    <Composition
      id="QuizViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(quizVizProps.durationMs)}
      defaultProps={quizVizProps}
    />
    <Composition
      id="VSCodeIntroViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(vscodeIntroProps.durationMs)}
      defaultProps={vscodeIntroProps}
    />
    <Composition
      id="SpineViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(spineVizProps.durationMs)}
      defaultProps={spineVizProps}
    />
    <Composition
      id="SpineLightViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(spineLightProps.durationMs)}
      defaultProps={spineLightProps}
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
      id="IllustrationLightViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(illustrationLightProps.durationMs)}
      defaultProps={illustrationLightProps}
    />
    <Composition
      id="CastViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(castVizProps.durationMs)}
      defaultProps={castVizProps}
    />
    <Composition
      id="StoryViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(storyVizProps.durationMs)}
      defaultProps={storyVizProps}
    />
    <Composition
      id="DataViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(dataVizProps.durationMs)}
      defaultProps={dataVizProps}
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
      id="SyllabusViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(syllabusVizProps.durationMs)}
      defaultProps={syllabusVizProps}
    />
    <Composition
      id="OutcomeViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(outcomeVizProps.durationMs)}
      defaultProps={outcomeVizProps}
    />
    <Composition
      id="BridgeViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(bridgeVizProps.durationMs)}
      defaultProps={bridgeVizProps}
    />
    <Composition
      id="DrillViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(drillVizProps.durationMs)}
      defaultProps={drillVizProps}
    />
    <Composition
      id="LabCardViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(labcardVizProps.durationMs)}
      defaultProps={labcardVizProps}
    />
    <Composition
      id="MissionViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(missionVizProps.durationMs)}
      defaultProps={missionVizProps}
    />
    <Composition
      id="MachineViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(machineVizProps.durationMs)}
      defaultProps={machineVizProps}
    />
    <Composition
      id="BlueprintViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(blueprintVizProps.durationMs)}
      defaultProps={blueprintVizProps}
    />
    <Composition
      id="RelayViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(relayVizProps.durationMs)}
      defaultProps={relayVizProps}
    />
    <Composition
      id="LayersViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(layersVizProps.durationMs)}
      defaultProps={layersVizProps}
    />
    <Composition
      id="PipelineViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(pipelineVizProps.durationMs)}
      defaultProps={pipelineVizProps}
    />
    <Composition
      id="RadixViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(radixVizProps.durationMs)}
      defaultProps={radixVizProps}
    />
    <Composition
      id="CarryViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(carryVizProps.durationMs)}
      defaultProps={carryVizProps}
    />
    <Composition
      id="BitfieldViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(bitfieldVizProps.durationMs)}
      defaultProps={bitfieldVizProps}
    />
    <Composition
      id="EncodeViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(encodeVizProps.durationMs)}
      defaultProps={encodeVizProps}
    />
    <Composition
      id="GatesViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(gatesVizProps.durationMs)}
      defaultProps={gatesVizProps}
    />
    <Composition
      id="LadderViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(ladderVizProps.durationMs)}
      defaultProps={ladderVizProps}
    />
    <Composition
      id="RegionsViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(regionsVizProps.durationMs)}
      defaultProps={regionsVizProps}
    />
    <Composition
      id="LookupViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(lookupVizProps.durationMs)}
      defaultProps={lookupVizProps}
    />
    <Composition
      id="StatesViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(statesVizProps.durationMs)}
      defaultProps={statesVizProps}
    />
    <Composition
      id="SchedulerViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(schedulerVizProps.durationMs)}
      defaultProps={schedulerVizProps}
    />
    <Composition
      id="ShellViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(shellVizProps.durationMs)}
      defaultProps={shellVizProps}
    />
    <Composition
      id="JourneyViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(journeyVizProps.durationMs)}
      defaultProps={journeyVizProps}
    />
    <Composition
      id="HandshakeViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(handshakeVizProps.durationMs)}
      defaultProps={handshakeVizProps}
    />
    <Composition
      id="StepperViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(stepperVizProps.durationMs)}
      defaultProps={stepperVizProps}
    />
    <Composition
      id="GrowthViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(growthVizProps.durationMs)}
      defaultProps={growthVizProps}
    />
    <Composition
      id="CallStackViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(callstackVizProps.durationMs)}
      defaultProps={callstackVizProps}
    />
    <Composition
      id="HistoryViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(historyVizProps.durationMs)}
      defaultProps={historyVizProps}
    />
    <Composition
      id="VersusViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(versusVizProps.durationMs)}
      defaultProps={versusVizProps}
    />
    <Composition
      id="ErasViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(erasVizProps.durationMs)}
      defaultProps={erasVizProps}
    />
    <Composition
      id="CardsViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(cardsVizProps.durationMs)}
      defaultProps={cardsVizProps}
    />
    <Composition
      id="DuelViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(duelVizProps.durationMs)}
      defaultProps={duelVizProps}
    />
    <Composition
      id="SpotlightViz"
      component={LessonVideo}
      fps={FPS}
      width={1920}
      height={1080}
      durationInFrames={msToFrame(spotlightVizProps.durationMs)}
      defaultProps={spotlightVizProps}
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
