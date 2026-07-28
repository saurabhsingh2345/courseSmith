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
      width={1700}
      height={1800}
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
