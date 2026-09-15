// The cut for lesson three, the last in the course.
//
// House rules for this one, learned across the first two:
//   footage carries the film, graphics are glue
//   waiting is ramped hard, never shown at pace
//   nothing on screen but the app: every take was shot fullscreen
//
// Source landmarks, in seconds (VS Code frame, fullscreen overlay and the
// Chrome password dialog are all physically removed from this media):
//     26  cursor on the live page from lesson two
//     92  new session, folder chip
//    137  the sentence sent
//   1440  the database question
//   1519  Postgres provisioned and seeded
//   3300  build finished
//   3400  the live login page
//   3483  signing up for real
//   3551  liking and commenting
//   3597  asking it to read the row back
//   3739  the rows, in Postgres

export type IconName =
  | 'web' | 'mobile' | 'server' | 'loop' | 'sprout' | 'keys'
  | 'python' | 'bolt' | 'braces' | 'star' | 'folder';

export type Rect = {x: number; y: number; w: number; h: number};
export type Spot = {rect: Rect; label?: string; side?: 'top'|'bottom'|'left'|'right'; at: number; until?: number};
export type Cue = {at: number; say: string};

export type Scene =
  | {k: 'run'; from: number; to: number; rate?: number; spots?: Spot[]}
  | {k: 'title'; title: string; kicker: string; sub: string}
  | {k: 'chapter'; n: string; title: string}
  | {k: 'statement'; kicker?: string; text: string; accent?: string[]; body?: string; size?: number}
  | {k: 'swap'; panels: {icon: IconName; who: string; what: string}[]}
  | {k: 'cards'; title: string; accent?: string[]; items: {name: string; note: string; icon?: IconName}[]; active?: number; footer?: string; sweepAt?: string[]}
  | {k: 'steps'; title: string; steps: {name: string; note: string}[]; sweepAt?: string[]}
  | {k: 'spotlight'; title: string; accent?: string[]; name: string; tagline: string; icon: IconName; facts: string[]}
  | {k: 'define'; word: string; pos: string; meaning: string; also: string}
  | {k: 'scale'; title: string; rungs: string[]; pick: number; left: string; right: string; note: string}
  | {k: 'address'}
  | {k: 'promptbuild'; title: string; chunks: {text: string; tag: string}[]}
  | {k: 'devices'; title: string}
  | {k: 'quiz'; n: string; question: string; options: string[]; answer: number; revealAt: number}
  | {k: 'roadmap'; title: string; nodes: {name: string; state: 'done'|'here'|'next'}[]}
  | {k: 'outro'; lines: string[]; sign: string};

export type Beat = {
  id: string; chapter: string; say?: string; cues?: Cue[];
  lead?: number; pad?: number; hold?: number; min?: number; scene: Scene;
};

export const BEATS: Beat[] = [
  // ================================================================ HOOK ===
  {
    id: 'hook',
    chapter: '',
    scene: {k: 'run', from: 1028, to: 1042, win: {x: 610, y: 749, w: 1700, h: 956}},
    cues: [
      {at: 0.8, say: 'This is a three dimensional game, running in a browser, on an ordinary laptop.'},
      {at: 6.5, say: 'Nobody typed a line of it. It came from one sentence, badly spelled, written by somebody who cannot code.'},
    ],
  },
  {
    id: 'title',
    chapter: 'Open',
    say: 'That sentence took about eight minutes to become a working game. In the next ten minutes you are going to do the same thing, from a completely empty machine.',
    scene: {
      k: 'title',
      kicker: 'Building with AI · Day one',
      title: 'From nothing to a running game',
      sub: 'Install the tool, open a folder, choose a model, write one plain sentence, and watch an agent build the whole thing while you read.',
    },
  },
  {
    id: 'plan',
    chapter: 'Open',
    say: 'There are only four things to learn today. Where the tool comes from. What a project folder is and why it matters more than it sounds. How to choose a model. And what an agent actually does once you press enter. Everything after today is a variation on these four.',
    scene: {
      k: 'cards',
      title: 'Day one, in four moves',
      accent: ['four'],
      items: [
        {name: 'Install', note: 'Download the desktop app and drag it across. Two minutes.', icon: 'bolt'},
        {name: 'Open a folder', note: 'The one decision that shapes everything the agent can see.', icon: 'folder'},
        {name: 'Pick a model', note: 'Cheap and quick, or slow and careful. It is a real choice.', icon: 'braces'},
        {name: 'Say what you want', note: 'One sentence. No specification, no jargon.', icon: 'star'},
      ],
    },
  },

  // ========================================================== CHAPTER 1 ===
  {id: 'ch1', chapter: 'Install', pad: 0.3, say: 'Start where everybody starts. An empty machine and a download page.',
   scene: {k: 'chapter', n: '01', title: 'Getting the tool'}},
  {
    id: 'dlpage',
    chapter: 'Install',
    scene: {k: 'run', from: 5, to: 44},
    cues: [
      {at: 1.0, say: 'This is the download page. There are three ways in and they matter less than the page makes it look.'},
      {at: 8.0, say: 'The desktop app is the one you want on day one. It is a full editor with an agent living inside it.'},
      {at: 16.0, say: 'The terminal option is the same agent without the editor around it, and the web option runs it on somebody else\'s machine. Both are excellent and neither is where you start.'},
      {at: 28.0, say: 'Scroll down and it lists every build. Mac, Windows, Linux. You want the one for the machine you are sitting at, and the page has usually worked that out already.'},
    ],
  },
  {
    id: 'dlclick',
    chapter: 'Install',
    scene: {k: 'run', from: 60, to: 96},
    cues: [
      {at: 1.5, say: 'One click, and it starts coming down.'},
      {at: 7.0, say: 'It is about two hundred and sixty megabytes, so on a normal connection this is a thirty second wait rather than a coffee.'},
      {at: 16.0, say: 'And there it is, finished, sitting in the downloads list at the top of the browser.'},
      {at: 24.0, say: 'This is the least interesting thing that will happen today, which is rather the point. The hard part of building software used to start here. Now it starts after this.'},
    ],
  },
  {
    id: 'install',
    chapter: 'Install',
    scene: {k: 'run', from: 196, to: 244, win: {x: 1457, y: 57, w: 1244, h: 700}},
    cues: [
      {at: 1.0, say: 'Open the file you just downloaded and you get this. On a Mac, installing anything is this one gesture.'},
      {at: 8.0, say: 'The application on the left. Your applications folder on the right. Drag the first onto the second and it is installed.'},
      {at: 18.0, say: 'That is genuinely all. There is no installer with nine screens and no set of options to get wrong.'},
      {at: 26.0, say: 'If you are on Windows the file you downloaded is a setup program instead, and you click next a few times. Same outcome, one extra minute.'},
      {at: 38.0, say: 'Then open it, and sign in when it asks. It is free to start and you do not need to pay anything to follow along today.'},
    ],
  },

  // ========================================================== CHAPTER 2 ===
  {id: 'ch2', chapter: 'The folder', pad: 0.3, say: 'Now the first thing that actually matters.',
   scene: {k: 'chapter', n: '02', title: 'The folder is the project'}},
  {
    id: 'folderidea',
    chapter: 'The folder',
    say: 'Before you open anything, understand this, because it is the thing beginners get wrong for weeks. The agent is not looking at your whole computer. It looks at one folder. That folder is the project. Everything it can read, everything it can change, and everything it can break lives inside it, and nothing outside it is visible at all. So an empty folder is not a limitation. It is a clean workshop, and it is the safest place to start.',
    scene: {
      k: 'steps',
      title: 'What opening a folder actually does',
      steps: [
        {name: 'You choose a folder', note: 'Any folder. A brand new empty one is perfect.'},
        {name: 'It becomes the workspace', note: 'The agent reads and writes only inside it.'},
        {name: 'Nothing else is visible', note: 'Your documents, your other work, the rest of the disk. Invisible.'},
        {name: 'Files appear as it works', note: 'You watch the project grow in the sidebar, in real time.'},
      ],
      sweepAt: ['You choose a folder', 'It becomes the workspace', 'Nothing else is visible', 'Files appear as it works'],
    },
  },
  {
    id: 'welcome',
    chapter: 'The folder',
    scene: {k: 'run', from: 246, to: 274},
    cues: [
      {at: 1.0, say: 'First time you open it, this is the screen. Four ways to begin and a list of anything you have opened before.'},
      {at: 9.0, say: 'Clone repo pulls down somebody else\'s existing code. Connect via S S H works on a machine somewhere else. Ignore both today.'},
      {at: 17.0, say: 'Open project is the one. I made an empty folder called arena before we started, so it is sitting there in the recent list.'},
    ],
  },
  {
    id: 'openit',
    chapter: 'The folder',
    scene: {k: 'run', from: 274, to: 294},
    cues: [
      {at: 1.5, say: 'Click it, and the whole application rearranges itself around that folder.'},
      {at: 9.0, say: 'Notice the name in the title bar has changed to arena. That is the tell. You are now inside a project rather than floating outside one.'},
    ],
  },
  {
    id: 'workspace',
    chapter: 'The folder',
    scene: {k: 'run', from: 294, to: 330},
    cues: [
      {at: 1.0, say: 'And this is the room you will be working in. Three parts, and you can ignore most of it today.'},
      {at: 8.0, say: 'On the left, the files in your folder. It is empty, because the folder is empty. In a few minutes it will not be.'},
      {at: 16.0, say: 'The middle is the editor, where code appears. You are not going to type in it. You may never type in it.'},
      {at: 22.0, say: 'On the right is the agent, and that is the entire job today.'},
    ],
  },

  // ========================================================== CHAPTER 3 ===
  {id: 'ch3', chapter: 'The model', pad: 0.3, say: 'One box to look at before you write anything.',
   scene: {k: 'chapter', n: '03', title: 'Choosing a brain'}},
  {
    id: 'modelidea',
    chapter: 'The model',
    say: 'Underneath the agent is a model, and you get to choose which one. This is not a settings screen you should skip. The model is the difference between something that guesses and something that thinks, and on a first build the difference is very visible. The rule is simple. Fast models for small edits, where you already know what you want and you just want it done. Careful models for anything you cannot describe precisely, which on day one is everything.',
    scene: {
      k: 'cards',
      title: 'Fast or careful',
      accent: ['Fast', 'careful'],
      items: [
        {name: 'Auto', note: 'It picks for you. A fine default once you stop thinking about it.', icon: 'loop'},
        {name: 'Fast models', note: 'Seconds, cheap, good for a small change you can describe exactly.', icon: 'bolt'},
        {name: 'Careful models', note: 'Slower and pricier. They plan before they write and they check their work.', icon: 'star'},
      ],
      active: 2,
      footer: 'On day one, pick the careful one. You are about to ask for something you cannot specify.',
    },
  },
  {
    id: 'picker',
    chapter: 'The model',
    scene: {k: 'run', from: 338, to: 364},
    cues: [
      {at: 1.0, say: 'The box at the bottom of the agent panel is the model. Click it and you get three controls.'},
      {at: 9.0, say: 'Fast is a speed switch. Effort is how long it is allowed to think before it starts writing. And model is the one that matters.'},
    ],
  },
  {
    id: 'modellist',
    chapter: 'The model',
    scene: {k: 'run', from: 368, to: 400},
    cues: [
      {at: 1.5, say: 'Open that and you see what you are actually paying for. Several different companies, sitting in one list, and you can switch between them mid project.'},
      {at: 11.0, say: 'That is worth pausing on. You are not locked to one company\'s brain. If one of them writes something you do not like, you change the dropdown and ask again.'},
      {at: 21.0, say: 'I am taking the careful one, because I am about to ask for a whole game in a single sentence and I want it to think first.'},
    ],
  },

  // ========================================================== CHAPTER 4 ===
  {id: 'ch4', chapter: 'The sentence', pad: 0.3, say: 'Now the part everybody comes for.',
   scene: {k: 'chapter', n: '04', title: 'One sentence'}},
  {
    id: 'prompt',
    chapter: 'The sentence',
    scene: {k: 'run', from: 426, to: 443},
    cues: [
      {at: 1.0, say: 'Here is what I typed. Look at how bad it is.'},
      {at: 5.0, say: 'No capital letters. No punctuation to speak of. No technical words at all. This is what a real beginner types, and it is enough.'},
    ],
  },
  {
    id: 'agentidea',
    chapter: 'The sentence',
    say: 'What happens when you press enter is the whole reason this is different from a chat window. It does not answer you. It goes and does the work. It looks at your folder to see what is there. It decides on an approach. It writes real files onto your disk. And it keeps going, on its own, until the thing you asked for exists. That loop is what the word agent means, and everything else in this course is built on it.',
    scene: {
      k: 'steps',
      title: 'What happens after enter',
      steps: [
        {name: 'It looks', note: 'Reads the folder to find out what already exists.'},
        {name: 'It thinks', note: 'Picks an approach and plans the pieces before writing anything.'},
        {name: 'It writes', note: 'Creates real files on your disk, not a suggestion in a chat box.'},
        {name: 'It keeps going', note: 'Repeats until the thing is finished, without being asked again.'},
      ],
      sweepAt: ['It looks', 'It thinks', 'It writes', 'It keeps going'],
    },
  },
  {
    id: 'thinking',
    chapter: 'The sentence',
    scene: {k: 'run', from: 466, to: 515},
    cues: [
      {at: 1.5, say: 'Press enter and watch what it does first.'},
      {at: 6.0, say: 'It tells you the approach in a sentence. A browser based shooter using a three dimensional library called Three point J S.'},
      {at: 15.0, say: 'Then it runs one command to list the folder, because it wants to know what it is walking into.'},
      {at: 24.0, say: 'Then it thinks for forty nine seconds. That is not the software being slow. That is the careful model doing what you paid it to do, and it is the reason the next part goes smoothly.'},
      {at: 36.0, say: 'It comes back having decided on the whole shape. Waves of enemies, power ups, sound, all from a sentence that never mentioned any of them.'},
    ],
  },
  {
    id: 'building',
    chapter: 'The sentence',
    scene: {k: 'run', from: 686, to: 896, rate: 8},
    cues: [
      {at: 1.0, say: 'And then it just builds. I have sped this up eight times so you can watch it happen.'},
      {at: 7.0, say: 'A file appears in the sidebar on the left that was not there sixty seconds ago. Green lines are code being added. It is writing, checking, and correcting itself as it goes.'},
      {at: 16.0, say: 'In real time this took about eight minutes. You do not have to sit and watch it. This is the part where you go and make the coffee you did not need earlier.'},
    ],
  },
  {
    id: 'result',
    chapter: 'The sentence',
    scene: {k: 'run', from: 856, to: 900},
    cues: [
      {at: 1.5, say: 'When it stops, it writes you a plain summary of what it made.'},
      {at: 7.0, say: 'Read this bit properly, because this is the part that teaches you. It explains the controls, the four enemy types, and how the difficulty works.'},
      {at: 17.0, say: 'It also tells you it is one single file with no build step, and it names the exact settings you can change to make it harder or easier.'},
      {at: 28.0, say: 'You asked for a game. It also handed you the instructions for editing it. That is a very different relationship to the one you have with software you buy.'},
    ],
  },

  // ========================================================== CHAPTER 5 ===
  {
    id: 'browser',
    chapter: 'Play it',
    scene: {k: 'run', from: 946, to: 996},
    cues: [
      {at: 1.0, say: 'So let us actually play it. There is a browser built into the editor, so you never have to leave.'},
      {at: 9.0, say: 'It started a small local server while it was working, which is a fancy way of saying the game is being served from your own machine and nobody else can see it.'},
      {at: 20.0, say: 'Point the browser at it.'},
      {at: 27.0, say: 'And there is a title screen. Nobody asked for a title screen.'},
      {at: 36.0, say: 'It invented the name, the colours, the glow, and the little diagram showing you which keys to press.'},
    ],
  },
  {
    id: 'play',
    chapter: 'Play it',
    scene: {k: 'run', from: 1022, to: 1066, win: {x: 610, y: 749, w: 1700, h: 956}},
    cues: [
      {at: 1.0, say: 'Arrow keys to move, space to shoot, exactly as the sentence asked.'},
      {at: 8.0, say: 'Enemies come in waves. They get faster. Some of them shoot back.'},
      {at: 18.0, say: 'And eventually they get you, and it counts up your score, your waves and your kills.'},
      {at: 28.0, say: 'One sentence. Eight minutes. A finished game with a scoring system nobody specified.'},
    ],
  },
  {
    id: 'honest',
    chapter: 'Play it',
    say: 'Now the honest part, because a course that only shows you the good version is not worth your time. This worked because it is a small self contained thing with nobody else depending on it. Ask for the same thing inside a large existing codebase and it goes slower and gets more wrong. It also cost real money, a few cents of it, and on bigger jobs that becomes real. None of that makes today a trick. It makes it the easy end of a real skill, which is exactly where day one should be.',
    scene: {
      k: 'cards',
      title: 'What today did not show you',
      accent: ['not'],
      items: [
        {name: 'A big codebase', note: 'An empty folder is the friendly case. Existing code is harder.', icon: 'server'},
        {name: 'The cost', note: 'Careful models are metered. Small here, not always small.', icon: 'keys'},
        {name: 'Being wrong', note: 'It will confidently build the wrong thing. Reading the summary is how you catch it.', icon: 'loop'},
      ],
      footer: 'Every one of these gets its own lesson. None of them is a reason to stop.',
    },
  },
  {
    id: 'outro',
    chapter: 'Play it',
    say: 'So that is day one. You installed it, you opened a folder, you chose a model, and you wrote one bad sentence that turned into a working game. Do that yourself before the next lesson. Ask for something you actually want rather than the thing I asked for, and let it surprise you. Next time we stop accepting the first answer and start steering it.',
    scene: {
      k: 'outro',
      lines: [
        'Install it. Open an empty folder.',
        'Choose the careful model.',
        'Ask for one thing, in your own words.',
        'Next: how to steer it when the first answer is wrong.',
      ],
      sign: 'Day one complete',
    },
  },
];
