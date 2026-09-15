// The cut.
//
// Two kinds of beat, exactly as in the remaster this rig comes from.
//
// A `run` is a continuous stretch of the screen recording made on the twenty
// sixth of August, played at real speed. Its length is `to - from`, and its
// narration is a list of cues pinned to the footage's own clock, so the words
// land on what is happening rather than the picture being cut to fit a
// sentence.
//
// Everything else is a graphic scene whose length is however long its line
// takes to say. Those live between the runs, never inside one.
//
// Narration is written to be spoken. No semicolons, no parentheses, no dashes
// mid clause, because Kokoro reads all three as a full stop. Small numbers are
// spelled out.
//
// Every number quoted in this script was measured on the day. The build took
// six and a half minutes and produced five files totalling forty eight
// kilobytes. The blue change took thirty one seconds and touched one file.

export type IconName =
  | 'web' | 'mobile' | 'server' | 'loop' | 'sprout' | 'keys'
  | 'python' | 'bolt' | 'braces' | 'star' | 'folder';

export type Rect = {x: number; y: number; w: number; h: number};

export type Spot = {
  rect: Rect;
  label?: string;
  side?: 'top' | 'bottom' | 'left' | 'right';
  at: number;
  until?: number;
};

export type Cue = {at: number; say: string};

export type Scene =
  | {k: 'run'; from: number; to: number; rate?: number; spots?: Spot[]}
  | {k: 'title'; title: string; kicker: string; sub: string}
  | {k: 'chapter'; n: string; title: string}
  | {k: 'statement'; kicker?: string; text: string; accent?: string[]; body?: string; size?: number}
  | {k: 'swap'; panels: {icon: IconName; who: string; what: string}[]}
  | {
      k: 'cards';
      title: string;
      accent?: string[];
      items: {name: string; note: string; icon?: IconName}[];
      active?: number;
      footer?: string;
      sweepAt?: string[];
    }
  | {k: 'steps'; title: string; steps: {name: string; note: string}[]; sweepAt?: string[]}
  | {
      k: 'spotlight';
      title: string;
      accent?: string[];
      name: string;
      tagline: string;
      icon: IconName;
      facts: string[];
    }
  | {k: 'define'; word: string; pos: string; meaning: string; also: string}
  | {k: 'scale'; title: string; rungs: string[]; pick: number; left: string; right: string; note: string}
  | {k: 'address'}
  | {k: 'promptbuild'; title: string; chunks: {text: string; tag: string}[]}
  | {k: 'devices'; title: string}
  | {k: 'quiz'; n: string; question: string; options: string[]; answer: number; revealAt: number}
  | {k: 'roadmap'; title: string; nodes: {name: string; state: 'done' | 'here' | 'next'}[]}
  | {k: 'outro'; lines: string[]; sign: string};

export type Beat = {
  id: string;
  chapter: string;
  say?: string;
  cues?: Cue[];
  lead?: number;
  pad?: number;
  /** Extra stillness after the line, for a beat that must be read. */
  hold?: number;
  min?: number;
  scene: Scene;
};

// The recording is the whole screen at three thousand four hundred and fifty
// six by two thousand two hundred and thirty four. The Claude window sits under
// the menu bar; the browser window later fills the same frame.
// Measured off a real frame at 3456 wide, not estimated.
// The folder chip sitting in the composer bar.
const FOLDER_CHIP: Rect = {x: 1300, y: 1986, w: 288, h: 54};
// The line you type into.
const INPUT: Rect = {x: 1158, y: 2055, w: 1544, h: 92};
// The column the conversation is written down.
const TALK: Rect = {x: 1050, y: 360, w: 1680, h: 780};

export const BEATS: Beat[] = [
  // ============================================================ COLD OPEN ===
  {
    id: 'hook',
    chapter: '',
    scene: {k: 'run', from: 802, to: 836},
    cues: [
      {
        at: 0.9,
        say: 'This is a social photo app. A feed, stories along the top, likes, comments, profiles, a search box, a dark mode. It is running in a browser on a laptop.',
      },
      {
        at: 12.0,
        say: 'Watch the number under that photo. One hundred and ninety two becomes one hundred and ninety three. That is a real click on real software, not a video of one.',
      },
      {
        at: 23.5,
        say: 'Nobody wrote a line of its code. It was made by typing one ordinary English sentence, and then a second one.',
      },
    ],
  },
  {
    id: 'title',
    chapter: 'Open',
    say: 'This is a complete walk through of that. Every minute of it, including the waiting.',
    scene: {
      k: 'title',
      kicker: 'Building with Claude',
      title: 'Ask for an app. Get an app.',
      sub: 'One folder, one sentence, six and a half minutes. We are going to do the whole thing start to finish and hide none of it.',
    },
  },
  {
    id: 'promise',
    chapter: 'Open',
    say: 'I want to be exact about what I am claiming, because a lot of demonstrations are not. I do not mean a picture of an app. I do not mean a design mock up. I mean files on a disk, opened in a browser, that respond when you click them.',
    hold: 0.8,
    scene: {
      k: 'statement',
      kicker: 'To be clear',
      text: 'Files on a disk. Opened in a browser. They respond when you click them.',
      accent: ['respond when you click them.'],
      size: 84,
    },
  },

  // ============================================================ CHAPTER 1 ===
  {id: 'ch1', chapter: 'The idea', pad: 0.35, say: 'Start with the one idea underneath all of this.',
   scene: {k: 'chapter', n: '01', title: 'The one idea'}},
  {
    id: 'idea',
    chapter: 'The idea',
    say: 'For as long as software has existed, the way to make it was to learn a language a computer understands, and then say what you wanted in that language. That is the part that takes years. The wanting was never the hard bit. The translating was.',
    hold: 1.0,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'braces', who: 'The old way', what: 'Learn a language the machine understands. Translate your idea into it. This is the part that takes years.'},
        {icon: 'bolt', who: 'What changed', what: 'Describe the thing you want in the language you already speak. The translating is done for you.'},
      ],
    },
  },
  {
    id: 'loop',
    chapter: 'The idea',
    say: 'And the whole thing is a loop with three steps. You describe what you want. It builds. You look at the result and say what is wrong or what is missing. Then it goes round again. That loop is the entire skill. Everything else in this video is detail.',
    hold: 1.4,
    scene: {
      k: 'steps',
      title: 'The whole skill, in three steps',
      steps: [
        {name: 'Describe', note: 'Say what you want in a plain sentence. Do not try to be clever about it.'},
        {name: 'It builds', note: 'Files appear on your machine. You watch it happen.'},
        {name: 'React', note: 'Say what is wrong or missing. Go round again.'},
      ],
      sweepAt: ['You describe what you want', 'It builds', 'say what is wrong'],
    },
  },

  // ============================================================ CHAPTER 2 ===
  {
    id: 'whofor',
    chapter: 'The idea',
    say: 'Three sorts of people end up watching something like this. Someone with an idea and no notion of where to begin. Someone who tried to learn to code once and stopped somewhere around the second month, which is the ordinary place to stop. And someone who can already build things and simply wants to go faster. The loop is the same for all three of you.',
    hold: 1.8,
    scene: {
      k: 'cards',
      title: 'Who this is for',
      items: [
        {name: 'You have an idea', note: 'And no notion of where to begin.', icon: 'sprout'},
        {name: 'You tried once', note: 'And stopped around the second month. That is the normal place to stop.', icon: 'loop'},
        {name: 'You already build', note: 'You just want to go faster.', icon: 'bolt'},
      ],
      sweepAt: ['an idea and no notion', 'stopped somewhere around the second month', 'already build things'],
    },
  },
  {id: 'ch2', chapter: 'What you need', pad: 0.35, say: 'So what do you actually need in front of you.',
   scene: {k: 'chapter', n: '02', title: 'What you need'}},
  {
    id: 'three',
    chapter: 'What you need',
    say: 'There are three places you can talk to Claude, and the difference between them matters more than anything else in this video, so I am going to be slow about it.',
    scene: {
      k: 'cards',
      title: 'Three places you can talk to Claude',
      items: [
        {name: 'The website', note: 'claude dot ai in a browser tab.', icon: 'web'},
        {name: 'The desktop app', note: 'A real application on your machine.', icon: 'folder'},
        {name: 'The terminal', note: 'For people who already live there.', icon: 'server'},
      ],
    },
  },
  {
    id: 'thedifference',
    chapter: 'What you need',
    say: 'Here is the difference. In a browser tab, Claude can write you something and show it to you, but it cannot reach your hard disk. It has no hands. The desktop app can be pointed at a folder, and then it writes real files into that folder, the same as you would.',
    hold: 1.2,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'web', who: 'In a browser tab', what: 'It can make you something and show it. It cannot touch your disk. Nothing appears in any folder.'},
        {icon: 'folder', who: 'In the desktop app', what: 'Point it at a folder. Real files land in that folder, and they are still there tomorrow.'},
      ],
    },
  },
  {
    id: 'whichone',
    chapter: 'What you need',
    say: 'For what we are doing today, that settles it. We want files we can keep, so we are using the desktop app. If all you have is the website, everything about how you ask still applies. You just get your result on a web page instead of in a folder.',
    scene: {
      k: 'cards',
      title: 'Three places you can talk to Claude',
      items: [
        {name: 'The website', note: 'Fine for learning. Nothing lands on your disk.', icon: 'web'},
        {name: 'The desktop app', note: 'What we are using. Real files, in a folder you choose.', icon: 'folder'},
        {name: 'The terminal', note: 'Same powers. Steeper room to stand in.', icon: 'server'},
      ],
      active: 1,
      footer: 'Everything about how you ask is the same in all three.',
    },
  },

  // ============================================================ CHAPTER 3 ===
  {
    id: 'setup',
    chapter: 'What you need',
    say: 'Setting it up is four steps and none of them are interesting. Download the app. Sign in. Open the part of it meant for building things rather than chatting. Then pick your folder. If you have used any application on a computer before, none of this will surprise you.',
    hold: 1.4,
    scene: {
      k: 'steps',
      title: 'Setting up, once',
      steps: [
        {name: 'Download the app', note: 'From the same place you signed up.'},
        {name: 'Sign in', note: 'The same account as the website.'},
        {name: 'Open the building side', note: 'Not the chat side. Look for code.'},
        {name: 'Pick a folder', note: 'The part that actually matters.'},
      ],
      sweepAt: ['Download the app', 'Sign in', 'Open the part of it meant for building', 'pick your folder'],
    },
  },
  {
    id: 'model',
    chapter: 'What you need',
    say: 'There is usually a choice of model, and a choice of how hard it should think. Turn both up for a first build. A stronger model asks you fewer questions and makes better decisions on your behalf, and on a first attempt that is exactly what you want, because you do not yet know enough to answer the questions well.',
    scene: {
      k: 'scale',
      title: 'How hard should it think',
      rungs: ['fastest', 'balanced', 'hardest'],
      pick: 2,
      left: 'quick answers',
      right: 'better decisions',
      note: 'On your first build, turn it up. You do not yet know enough to correct it cheaply.',
    },
  },
  {id: 'ch3', chapter: 'The folder', pad: 0.35, say: 'Step one. Give it somewhere to work.',
   scene: {k: 'chapter', n: '03', title: 'Give it a folder'}},
  {
    id: 'folderwhy',
    chapter: 'The folder',
    say: 'Before you ask for anything, you tell the app which folder on your machine it is allowed to work in. This is the single most important habit in this whole video, and it takes about four seconds.',
    scene: {
      k: 'statement',
      text: 'You choose the folder. Not the other way round.',
      accent: ['You choose the folder.'],
      size: 88,
    },
  },
  {
    id: 'folderhow',
    chapter: 'The folder',
    scene: {
      k: 'run',
      from: 0,
      to: 34,
      spots: [
        {rect: FOLDER_CHIP, label: 'the folder it is pointed at', side: 'top', at: 9.0, until: 30.0},
      ],
    },
    cues: [
      {at: 0.8, say: 'Make an empty folder somewhere you will find it again. I made one on the desktop and called it instagram clone.'},
      {at: 11.0, say: 'This is the building side of the app, not the chat side. It is the one that can reach your disk.'},
      {at: 19.0, say: 'And that is the folder name sitting at the bottom of the window. From here on it is always visible, so you can never be confused about where the files are going.'},
    ],
  },
  {
    id: 'trust',
    chapter: 'The folder',
    say: 'The first time you point at a folder you will be asked to confirm it. The words are blunt on purpose. Claude may read, write, or execute files in this directory. Read that sentence properly. An empty folder you just made is a fine place to say yes. Your documents folder is not.',
    hold: 1.4,
    scene: {
      k: 'cards',
      title: 'Where to point it',
      items: [
        {name: 'A new empty folder', note: 'Yes. Nothing in it to lose.', icon: 'sprout'},
        {name: 'A project you already have', note: 'Only once you trust the loop.', icon: 'braces'},
        {name: 'Documents or Desktop itself', note: 'No. Far too much reach.', icon: 'keys'},
      ],
      active: 0,
      footer: 'Read the confirmation. It means exactly what it says.',
    },
  },

  // ============================================================ CHAPTER 4 ===
  {id: 'ch4', chapter: 'The sentence', pad: 0.35, say: 'Step two. The sentence.',
   scene: {k: 'chapter', n: '04', title: 'The first sentence'}},
  {
    id: 'vocab',
    chapter: 'The sentence',
    say: 'One piece of vocabulary before we go on, because you will hear it everywhere and it is less impressive than it sounds. A prompt is just the thing you typed. That is all the word means. It is not a special format and there is nothing to memorise.',
    hold: 1.6,
    scene: {
      k: 'define',
      word: 'prompt',
      pos: 'noun',
      meaning: 'The thing you typed. That is the whole definition.',
      also: 'Not a format. Not a spell. Not something you can get wrong by phrasing it like a person.',
    },
  },
  {
    id: 'fear',
    chapter: 'The sentence',
    say: 'This is where most people freeze. They think there is a correct way to phrase it, and that if they get the phrasing wrong they will get something useless back. So they sit there editing a sentence for ten minutes. Do not do that.',
    scene: {
      k: 'statement',
      kicker: 'The thing that stops people',
      text: 'There is no magic phrasing. You are not casting a spell.',
      accent: ['You are not casting a spell.'],
      size: 80,
    },
  },
  {
    id: 'whatityped',
    chapter: 'The sentence',
    say: 'Here is what I actually typed. Make a basic instagram clone, just html css and javascript. Eleven words. No capital letters where they belong, no full stop at the end. That is the entire specification for the thing you saw at the start.',
    hold: 1.2,
    scene: {
      k: 'promptbuild',
      title: 'The whole specification',
      chunks: [
        {text: 'make a basic instagram clone', tag: 'what you want'},
        {text: 'just html css and javascript', tag: 'how simple to keep it'},
      ],
    },
  },
  {
    id: 'thetyping',
    chapter: 'The sentence',
    scene: {k: 'run', from: 4, to: 42},
    cues: [
      {at: 1.0, say: 'There it is going in, at the speed I actually typed it. Nothing hidden and nothing edited out.'},
      {at: 12.0, say: 'Eleven words, all lower case, no full stop. If you were waiting for the clever part, this is it. There is not one.'},
      {at: 25.0, say: 'And that is the send. From this moment I did not touch the machine again until it had finished.'},
    ],
  },
  {
    id: 'twohalves',
    chapter: 'The sentence',
    say: 'That sentence has two halves and they do different jobs. The first half says what you want. The second half says how complicated you are willing to let it be. The second half is the one beginners leave out, and leaving it out is what produces something you cannot understand.',
    hold: 1.5,
    scene: {
      k: 'steps',
      title: 'Both halves matter',
      steps: [
        {name: 'What you want', note: 'A basic instagram clone. Name the thing. Do not describe every feature.'},
        {name: 'How simple to keep it', note: 'Just html css and javascript. This is the half people forget.'},
      ],
      sweepAt: ['what you want', 'how complicated'],
    },
  },
  {
    id: 'whysimple',
    chapter: 'The sentence',
    say: 'I want to show you what those five words were worth. When I asked without them, on the same day, on the same machine, it built something very different. Fifty eight files. Four thousand eight hundred lines. A login screen, a database, private messages, a reels page.',
    hold: 1.6,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'braces', who: 'Without those five words', what: 'Fifty eight files. Four thousand eight hundred lines. A database, accounts, messages. Sixteen minutes.'},
        {icon: 'bolt', who: 'With them', what: 'Five files. Forty eight kilobytes. Opens by double clicking. Six and a half minutes.'},
      ],
    },
  },

  // ============================================================ CHAPTER 5 ===
  {
    id: 'nospec',
    chapter: 'The sentence',
    say: 'The instinct almost everybody has is to write a specification. A list of every screen and every button, because that feels like being thorough. It is the wrong instinct. A long list gives it a hundred small instructions to satisfy and no sense of the shape of the thing.',
    hold: 1.2,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'braces', who: 'What people write', what: 'A list of nineteen features, each described in half a sentence, with nothing saying what the thing is.'},
        {icon: 'star', who: 'What works better', what: 'Name the thing. Say how simple to keep it. Let it fill in what obviously belongs.'},
      ],
    },
  },
  {
    id: 'nameit',
    chapter: 'The sentence',
    say: 'Naming something that already exists is the strongest shortcut you have. The word instagram carried a feed, stories along the top, likes, comments, profiles and a grid, and I never typed any of those words. Compare that to describing all six from scratch and getting the emphasis wrong.',
    hold: 1.6,
    scene: {
      k: 'spotlight',
      title: 'One word did all this',
      name: 'instagram',
      tagline: 'Six features I never had to describe',
      icon: 'star',
      facts: [
        'A feed of posts, newest first.',
        'Stories in a row along the top.',
        'Likes, comments, profiles, a grid.',
      ],
    },
  },
  {
    id: 'examples',
    chapter: 'The sentence',
    say: 'So here are three sentences in the shape that works, for three completely different things. Notice how short they are, and notice that each one ends by saying how simple to keep it.',
    hold: 2.2,
    scene: {
      k: 'cards',
      title: 'The shape that works',
      items: [
        {name: 'make a basic recipe box, just html css and javascript', note: 'A thing you name. A limit you set.', icon: 'web'},
        {name: 'make a simple habit tracker, plain html and javascript, no frameworks', note: 'Same shape, different thing.', icon: 'loop'},
        {name: 'make a small invoice generator, one html file I can open', note: 'The limit can be about the file, not the language.', icon: 'folder'},
      ],
      footer: 'Name the thing. Then say how simple to keep it.',
    },
  },
  {
    id: 'vague',
    chapter: 'The sentence',
    say: 'And here is what to avoid. Make me something cool. Build an app. Help me with my business. Not because they are rude, but because they contain no thing. There is nothing in them to build, so you will get questions back, and answering questions is slower than saying what you want.',
    scene: {
      k: 'cards',
      title: 'What not to say',
      items: [
        {name: 'make me something cool', note: 'There is no thing in this sentence.', icon: 'keys'},
        {name: 'build an app', note: 'An app that does what.', icon: 'keys'},
        {name: 'help me with my business', note: 'You will get questions, not software.', icon: 'keys'},
      ],
      footer: 'None of these are rude. They are just empty.',
    },
  },
  {id: 'ch5', chapter: 'The wait', pad: 0.35, say: 'Step three. The part nobody shows you.',
   scene: {k: 'chapter', n: '05', title: 'The waiting'}},
  {
    id: 'waitwhy',
    chapter: 'The wait',
    scene: {
      k: 'run',
      from: 36,
      to: 132,
      spots: [{rect: TALK, label: 'the first file lands here', side: 'bottom', at: 40.0, until: 62.0}],
    },
    cues: [
      {at: 0.8, say: 'Most videos cut this part out, and cutting it out is exactly why people think something has broken when it happens to them. So we are going to sit in it at real speed.'},
      {at: 14.0, say: 'First it reads the folder and finds it empty. That matters, because an empty folder means it can decide the shape of the whole thing rather than fitting in around what is already there.'},
      {at: 33.0, say: 'Then it commits to an approach and says so. Vanilla html, css and javascript, exactly as asked, and self contained.'},
      {at: 47.0, say: 'And there is the first file, forty five seconds after I pressed send. From here the folder is no longer empty and there is something on your disk that was not there before.'},
      {at: 68.0, say: 'Watch the left hand column rather than the words. Each of those lines is a real operation finishing.'},
    ],
  },
  {
    id: 'whatitdoes',
    chapter: 'The wait',
    say: 'While you wait it is doing four things over and over. It writes a file. It reads back what it wrote. It runs something to check the result. And it fixes what it finds. That last one is the part that matters, and we will see it happen.',
    hold: 1.4,
    scene: {
      k: 'steps',
      title: 'What is happening while nothing seems to happen',
      steps: [
        {name: 'Write', note: 'Put a file on your disk.'},
        {name: 'Read back', note: 'Look at what it actually wrote, not what it meant to write.'},
        {name: 'Check', note: 'Run it. Open it. See whether it works.'},
        {name: 'Fix', note: 'Correct what it finds, without being asked.'},
      ],
      sweepAt: ['It writes a file', 'reads back', 'runs something to check', 'fixes what it finds'],
    },
  },
  {
    id: 'howlong',
    chapter: 'The wait',
    say: 'The whole build took six and a half minutes. That is the honest number for this app on this day. If yours takes twelve, nothing is wrong. If it takes forty five seconds, nothing is wrong either. The variation is normal and it is not a sign of quality.',
    hold: 1.0,
    scene: {
      k: 'scale',
      title: 'How long a first build takes',
      rungs: ['30 seconds', '2 minutes', '6 minutes', '15 minutes', 'half an hour'],
      pick: 2,
      left: 'a small page',
      right: 'a whole product',
      note: 'Ours was six and a half minutes for five files. Yours will differ, and that is fine.',
    },
  },
  {
    id: 'dontpoke',
    chapter: 'The wait',
    say: 'One rule while you wait. Leave the folder alone. I broke this rule on the day, out of impatience, and moved the folder while it was still working. It noticed within seconds, told me the directory had been wiped, and started rebuilding everything from memory.',
    hold: 1.5,
    scene: {
      k: 'cards',
      title: 'While it is working',
      items: [
        {name: 'Do not edit the files', note: 'You will be writing over each other.', icon: 'braces'},
        {name: 'Do not move the folder', note: 'I did this. It noticed in seconds.', icon: 'folder'},
        {name: 'Do go and do something else', note: 'This is the correct move.', icon: 'loop'},
      ],
      active: 2,
    },
  },

  // ============================================================ CHAPTER 6 ===
  {
    id: 'permission',
    chapter: 'The wait',
    say: 'You may be asked to approve things while it works. Modern versions handle the safe ones for you and stop on anything that looks risky. If it does stop and ask, read what it wants to do. Writing a file inside your folder is ordinary. Anything reaching outside that folder deserves a proper look.',
    hold: 1.4,
    scene: {
      k: 'cards',
      title: 'When it asks permission',
      items: [
        {name: 'Writing inside your folder', note: 'Ordinary. This is the job.', icon: 'folder'},
        {name: 'Installing something', note: 'Normal, but know it is happening.', icon: 'server'},
        {name: 'Reaching outside the folder', note: 'Stop and read it properly.', icon: 'keys'},
      ],
      active: 2,
    },
  },
  {
    id: 'midbuild',
    chapter: 'The wait',
    scene: {k: 'run', from: 132, to: 505, rate: 3},
    cues: [
      {at: 1.0, say: 'Now I am going to speed the middle of it up three times, because the honest truth is that nothing surprising happens here and you should know that in advance.'},
      {at: 16.0, say: 'It writes a file. It reads back what it actually wrote rather than what it meant to write. It runs something to check the result. Then it fixes whatever it finds and carries on.'},
      {at: 40.0, say: 'Six minutes of your life go past in this stretch. If you were sitting in front of it you would have made a cup of tea by now, and that is genuinely the correct response.'},
      {at: 66.0, say: 'The only thing worth holding on to is that every one of those lines is a real file landing in a real folder. None of this is a preview or a simulation of working.'},
      {at: 95.0, say: 'And it is speeding up here, but the machine was not. This is what patience looks like compressed.'},
    ],
  },
  {id: 'ch6', chapter: 'What it made', pad: 0.35, say: 'It finishes. Now read what you were given.',
   scene: {k: 'chapter', n: '06', title: 'Reading the result'}},
  {
    id: 'done',
    chapter: 'What it made',
    scene: {k: 'run', from: 500, to: 566},
    cues: [
      {at: 1.0, say: 'When it finishes it does not simply stop and go quiet. It tells you what it made, file by file, and what each one of them is for.'},
      {at: 16.0, say: 'Five files. That is the entire application. You could read the whole list out loud in about ten seconds.'},
      {at: 27.0, say: 'And before it handed anything over it had already opened the app in a browser and clicked through it. Liking, commenting, the profile page, creating a post, the phone layout. It checked its own work without being asked to.'},
      {at: 49.0, say: 'That paragraph near the bottom is it telling you what it decided on your behalf, and what it cannot do. It is the most useful thing on this screen and almost nobody reads it.'},
    ],
  },
  {
    id: 'thefiles',
    chapter: 'What it made',
    say: 'Five files. Forty eight kilobytes altogether, which is smaller than a single photograph. The page itself, the styling, the behaviour, a tiny server you do not strictly need, and a read me explaining how to start it again.',
    hold: 1.8,
    scene: {
      k: 'cards',
      title: 'Everything it made',
      items: [
        {name: 'index.html', note: 'The page. Header, three views, the modals.', icon: 'web'},
        {name: 'styles.css', note: 'Every colour and size. Twelve kilobytes.', icon: 'star'},
        {name: 'app.js', note: 'All the behaviour. Twenty seven kilobytes, the biggest by far.', icon: 'braces'},
        {name: 'server.js', note: 'Thirty lines. Only needed if you do not open the file directly.', icon: 'server'},
      ],
      footer: 'Plus a README explaining how to run it again tomorrow.',
      sweepAt: ['The page itself', 'the styling', 'the behaviour', 'a tiny server'],
    },
  },
  {
    id: 'offline',
    chapter: 'What it made',
    say: 'One detail I did not ask for and would not have thought of. There are no photographs in this app at all. Every picture you saw is drawn by the code itself from the letters of the username, which means the whole thing works with the internet switched off.',
    hold: 1.6,
    scene: {
      k: 'spotlight',
      title: 'A decision nobody asked for',
      name: 'No photographs. Anywhere.',
      tagline: 'Every image is drawn from the letters of the name',
      icon: 'sprout',
      facts: [
        'Zero network requests. It works on a plane.',
        'Each post gets its own colours from its own name.',
        'The same name always draws the same picture.',
      ],
    },
  },
  {
    id: 'readit',
    chapter: 'What it made',
    say: 'Here is the habit worth building. Read that summary every time, even when you do not understand the words. You are not checking the code. You are checking whether it built the thing you meant, and that is a question you are already qualified to answer.',
    hold: 1.2,
    scene: {
      k: 'steps',
      title: 'What to look for in the summary',
      steps: [
        {name: 'Did it make the thing I meant', note: 'Not is the code good. Is it the right thing.'},
        {name: 'What did it decide for me', note: 'There is always something. Usually sensible.'},
        {name: 'What did it say it cannot do', note: 'This is the most valuable paragraph.'},
      ],
      sweepAt: ['built the thing you meant', 'already qualified'],
    },
  },

  // ============================================================ CHAPTER 7 ===
  {
    id: 'features',
    chapter: 'What it made',
    say: 'And look at what eleven words actually bought. A feed with likes and comments and saving. Stories you can open. An explore grid. Profiles you can follow and edit. A search box. A notifications panel. A dark mode. A phone layout. None of those were in my sentence.',
    hold: 2.4,
    scene: {
      k: 'cards',
      title: 'Things I never asked for',
      items: [
        {name: 'Stories, and a viewer', note: 'Tap a face, it opens. Arrow keys work.', icon: 'star'},
        {name: 'Search and notifications', note: 'Both there. Neither mentioned.', icon: 'bolt'},
        {name: 'A dark mode', note: 'Included, and it survived the colour change.', icon: 'loop'},
        {name: 'A phone layout', note: 'Bottom tab bar, edge to edge cards.', icon: 'mobile'},
      ],
      sweepAt: ['Stories you can open', 'A search box', 'A dark mode', 'A phone layout'],
    },
  },
  {
    id: 'phone',
    chapter: 'What it made',
    say: 'It even built the phone version. Narrow the window and the sidebar becomes a row of buttons along the bottom, and the photos run edge to edge the way they do in a real app. I did not ask for that either. It is simply what an instagram is.',
    hold: 1.8,
    scene: {k: 'devices', title: 'The phone layout, unasked for'},
  },
  {
    id: 'openhow',
    chapter: 'What it made',
    say: 'Opening it is the least dramatic part. Go to the folder and double click the page file, and it opens in your browser like anything else. It also wrote a small server for the situations where that is not enough, and told me plainly that I probably would not need it.',
    scene: {
      k: 'steps',
      title: 'Opening what you built',
      steps: [
        {name: 'Open the folder', note: 'The one you chose at the start.'},
        {name: 'Double click index.html', note: 'That is genuinely all.'},
        {name: 'It opens in your browser', note: 'Same browser as everything else.'},
      ],
      sweepAt: ['double click the page file', 'opens in your browser'],
    },
  },
  {id: 'ch7', chapter: 'Changing it', pad: 0.35, say: 'Step four. You want it different.',
   scene: {k: 'chapter', n: '07', title: 'Changing your mind'}},
  {
    id: 'secondprompt',
    chapter: 'Changing it',
    scene: {
      k: 'run',
      from: 552,
      to: 592,
      spots: [{rect: INPUT, label: 'five words', side: 'top', at: 14.0, until: 30.0}],
    },
    cues: [
      {at: 0.8, say: 'This is the half of the loop that turns a demonstration into a tool. You do not start again, and you do not go back to the beginning. You just say the next thing.'},
      {at: 16.0, say: 'So I typed five words. Make the theme color blue. No explanation of which parts, and no list of the things I meant by it. Five words into the same box as before.'},
    ],
  },
  {
    id: 'interpreted',
    chapter: 'Changing it',
    say: 'And look at what came back first, because this is the most interesting moment in the whole recording. It said the accent is already blue, so I will read that as make the whole thing blue. It worked out what I meant rather than what I said.',
    hold: 1.6,
    scene: {
      k: 'statement',
      kicker: 'The moment worth rewinding',
      text: 'It read what I meant, not what I typed.',
      accent: ['not what I typed.'],
      size: 90,
    },
  },
  {
    id: 'blueread',
    chapter: 'Changing it',
    scene: {k: 'run', from: 592, to: 652, rate: 3},
    cues: [
      {at: 1.0, say: 'Sped up again while it works, because you have already seen what working looks like.'},
      {at: 10.0, say: 'This one is quick. It is not building anything new, it is finding the places where colour is decided and changing them.'},
    ],
  },
  {
    id: 'blueanswer',
    chapter: 'Changing it',
    scene: {k: 'run', from: 652, to: 718},
    cues: [
      {at: 1.2, say: 'And here is what it said, which is the most interesting thing in the whole recording. It worked out that the accent colour was already blue, so taking me literally would have changed almost nothing.'},
      {at: 20.0, say: 'So it decides what I must have meant instead, says so plainly before touching anything, and only then makes the change. That sentence is the difference between a tool and an assistant.'},
      {at: 40.0, say: 'It lists exactly what it changed. The backgrounds, the cards, the borders, and the rings around the stories, which went from orange and pink to cyan and violet.'},
      {at: 55.0, say: 'And it tells you the one thing it deliberately did not change, and why.'},
    ],
  },
  {
    id: 'thirtyone',
    chapter: 'Changing it',
    say: 'Thirty one seconds later it was done, and it had changed exactly one file. Not five. One. The colours of the entire application live in about twelve lines at the top of the styling file, because six minutes earlier it had chosen to build it that way.',
    hold: 1.8,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'star', who: 'The change', what: 'Thirty one seconds. One file. Around twelve values, all in one block at the top.'},
        {icon: 'bolt', who: 'Why it was that easy', what: 'Because of a decision it made six minutes earlier, before anyone mentioned colour.'},
      ],
    },
  },
  {
    id: 'refused',
    chapter: 'Changing it',
    say: 'It also refused part of my instruction, and gave a reason. It left the pictures in their original colours, because those stand in for photographs, and making them blue would have turned the whole feed into one colour. I asked for blue. It gave me blue and kept the app looking like an app.',
    hold: 1.6,
    scene: {
      k: 'cards',
      title: 'What it changed, and what it would not',
      items: [
        {name: 'Backgrounds and cards', note: 'Pale blue in light. Navy in dark.', icon: 'star'},
        {name: 'The rings on the stories', note: 'Orange and pink became cyan and violet.', icon: 'loop'},
        {name: 'The photographs', note: 'Left alone, on purpose, and it said why.', icon: 'sprout'},
      ],
      active: 2,
      footer: 'Being told no, with a reason, is better than being obeyed.',
    },
  },

  // ============================================================ CHAPTER 8 ===
  {id: 'chw', chapter: 'When it goes wrong', pad: 0.35, say: 'Now the chapter that most tutorials leave out.',
   scene: {k: 'chapter', n: '08', title: 'When it goes wrong'}},
  {
    id: 'willgowrong',
    chapter: 'When it goes wrong',
    say: 'It will go wrong. Not always, not even often, but often enough that you need to know what to do, because the instinct almost everyone has is to start again from nothing, and starting again from nothing throws away the good parts along with the bad.',
    scene: {
      k: 'statement',
      kicker: 'The one nobody films',
      text: 'Do not start again. Say what you can see.',
      accent: ['Say what you can see.'],
      size: 88,
    },
  },
  {
    id: 'threewrongs',
    chapter: 'When it goes wrong',
    say: 'There are really only three ways it goes wrong, and each has one sentence that fixes it. It built the wrong thing. It broke something that used to work. Or it stopped in the middle and left you with half a room.',
    hold: 2.4,
    scene: {
      k: 'cards',
      title: 'Three ways it goes wrong',
      items: [
        {name: 'Wrong thing', note: 'Say: this is not what I meant, I wanted a something.', icon: 'keys'},
        {name: 'Broke what worked', note: 'Say: the likes stopped working after that change.', icon: 'loop'},
        {name: 'Stopped halfway', note: 'Say: carry on, you did not finish the profile page.', icon: 'bolt'},
      ],
      sweepAt: ['built the wrong thing', 'broke something that used to work', 'stopped in the middle'],
    },
  },
  {
    id: 'describe',
    chapter: 'When it goes wrong',
    say: 'And the rule underneath all three is the same. Describe what you see, not what you think is causing it. The button does nothing when I click it is a useful sentence. I think the javascript is broken is a guess, and if your guess is wrong you have just sent it in the wrong direction.',
    hold: 1.8,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'star', who: 'Say this', what: 'The like button does nothing when I click it. It worked before the colour change.'},
        {icon: 'keys', who: 'Not this', what: 'I think the javascript is broken. A wrong guess sends it in the wrong direction.'},
      ],
    },
  },
  {
    id: 'openbrowser',
    chapter: 'Using it',
    scene: {k: 'run', from: 700, to: 770, rate: 2},
    cues: [
      {at: 1.0, say: 'So let us go and use the thing. This is the folder being opened and the page being loaded, at double speed because it is not interesting.'},
      {at: 14.0, say: 'No installer, no account, no sign up screen. It is a page on your disk and a browser, which is the same arrangement as every website you have ever opened.'},
    ],
  },
  {id: 'ch8', chapter: 'Using it', pad: 0.35, say: 'Step five. Open it yourself.',
   scene: {k: 'chapter', n: '08', title: 'Using the thing'}},
  {
    id: 'openit',
    chapter: 'Using it',
    scene: {k: 'run', from: 780, to: 857},
    cues: [
      {at: 1.5, say: 'This is the app in an ordinary browser window. Not a preview inside the tool, and not a screenshot of one. A browser, the same one you read your email in.'},
      {at: 16.0, say: 'Stories along the top with the blue rings it made. A post with a caption and a location under the name. Real comments underneath it. All of that came from eleven words.'},
      {at: 32.0, say: 'The heart fills and the count moves by one. That is your own click, on your own machine, on software that did not exist eight minutes earlier.'},
      {at: 46.0, say: 'And that is the dark mode, which I never once asked for. It was simply included, and it is blue, because blue is what I asked for.'},
      {at: 58.0, say: 'The explore grid, built out of the same generated pictures.'},
      {at: 66.0, say: 'And the profile page, with a follower count and an empty posts tab sitting there waiting for you to fill it.'},
    ],
  },
  {
    id: 'yours',
    chapter: 'Using it',
    say: 'Everything you just watched is sitting in that folder on the desktop. Close the app, close the browser, turn the laptop off. Tomorrow those five files are still there and they still work, because they do not depend on anything that could be taken away.',
    hold: 1.2,
    scene: {
      k: 'statement',
      text: 'Turn the laptop off. Tomorrow it is still there.',
      accent: ['Tomorrow it is still there.'],
      size: 88,
    },
  },
  {
    id: 'privacy',
    chapter: 'Using it',
    say: 'A word about what leaves your machine, because people ask and they are right to. The files are on your disk. The app runs in your browser. Anything you type into it stays in that browser and is not sent anywhere. What did travel is your sentence, which went off to be answered, and that is the whole of it.',
    hold: 1.6,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'folder', who: 'Stays on your machine', what: 'The files. The app. Everything you type into the app itself.'},
        {icon: 'web', who: 'Left your machine', what: 'Your sentence, sent to be answered. That is all that travelled.'},
      ],
    },
  },
  {
    id: 'whatyouhave',
    chapter: 'Using it',
    say: 'Be clear eyed about what this is and is not. It is real software running on your machine that you can show someone. It is not on the internet, there are no other users, and nothing you type into it leaves the browser. Those are the next lessons, not this one.',
    hold: 1.8,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'bolt', who: 'What you have', what: 'Real, working software on your machine. Yours to keep, change, and show people.'},
        {icon: 'web', who: 'What you do not have yet', what: 'A website other people can visit. Real accounts. Anything saved anywhere but this browser.'},
      ],
    },
  },

  // ============================================================ CHAPTER 9 ===
  {
    id: 'moreasks',
    chapter: 'Changing it',
    say: 'And the same five word move works for everything else you might want. Make the photos bigger. Put the menu on the right. Add a button that clears everything. Make it work on my phone. Each of those is one sentence and one wait, and none of them require you to know how it was built.',
    hold: 2.0,
    scene: {
      k: 'cards',
      title: 'Other sentences that just work',
      items: [
        {name: 'make the photos bigger', note: 'Size and spacing.', icon: 'web'},
        {name: 'put the menu on the right', note: 'Layout.', icon: 'loop'},
        {name: 'add a button that clears everything', note: 'New behaviour.', icon: 'bolt'},
        {name: 'make it work on my phone', note: 'A whole second layout.', icon: 'mobile'},
      ],
      footer: 'One sentence. One wait. No knowledge of how it was built.',
    },
  },
  {
    id: 'whattobuild',
    chapter: 'Changing it',
    say: 'And if you are wondering what to point this at once the novelty wears off, the answer is almost always something small and annoying that you already do by hand. The list you keep in a notes app. The thing you recalculate in a spreadsheet every month. Those are better first projects than a clone of anything.',
    hold: 2.0,
    scene: {
      k: 'cards',
      title: 'Better first projects than a clone',
      items: [
        {name: 'The list in your notes app', note: 'You already maintain it. Make it a real thing.', icon: 'folder'},
        {name: 'The monthly spreadsheet', note: 'The one you rebuild every month by hand.', icon: 'braces'},
        {name: 'The thing you look up constantly', note: 'Put it behind one page and one search box.', icon: 'web'},
      ],
      footer: 'Small and annoying beats big and impressive, every time.',
    },
  },
  {id: 'ch9', chapter: 'Next', pad: 0.35, say: 'Last part. What to do the moment this video ends.',
   scene: {k: 'chapter', n: '09', title: 'What to do next'}},
  {
    id: 'recap',
    chapter: 'Next',
    say: 'So let us put the whole thing back together, because it is genuinely short. You made an empty folder. You pointed the app at it and read the warning. You typed one sentence naming a thing and setting a limit. You waited without touching anything. You read what it told you. Then you changed one thing by saying so.',
    hold: 2.4,
    scene: {
      k: 'steps',
      title: 'The entire video, in six lines',
      steps: [
        {name: 'An empty folder', note: 'Made on purpose.'},
        {name: 'Point and read the warning', note: 'It means what it says.'},
        {name: 'One sentence', note: 'Name a thing. Set a limit.'},
        {name: 'Wait, and touch nothing', note: 'Six and a half minutes.'},
        {name: 'Read what it tells you', note: 'Especially the limits.'},
        {name: 'Change one thing', note: 'Thirty one seconds.'},
      ],
      sweepAt: ['made an empty folder', 'pointed the app at it', 'typed one sentence', 'waited without touching', 'read what it told you', 'changed one thing'],
    },
  },
  {
    id: 'checkpoint',
    chapter: 'Next',
    say: 'Quick check that the important idea landed, because if only one thing survives this video it should be this one.',
    scene: {
      k: 'quiz',
      n: '01',
      question: 'You want files that stay on your computer. Where do you work?',
      options: ['A browser tab on the website', 'The desktop app, pointed at a folder', 'Either one, it makes no difference'],
      answer: 1,
      revealAt: 150,
    },
  },
  {
    id: 'answer',
    chapter: 'Next',
    say: 'The desktop app, pointed at a folder you chose. The website is genuinely useful and you should use it. It just cannot put anything on your disk, and today we wanted things on the disk.',
    scene: {
      k: 'statement',
      text: 'The website shows you things. The desktop app leaves you things.',
      accent: ['leaves you things.'],
      size: 82,
    },
  },
  {
    id: 'donext',
    chapter: 'Next',
    say: 'So here is your actual homework, and it is small on purpose. Make an empty folder. Point the app at it. Ask for something you would use yourself, in one sentence, and add the words just html css and javascript on the end. Then change one thing about it.',
    hold: 2.0,
    scene: {
      k: 'steps',
      title: 'Do this today. It takes ten minutes.',
      steps: [
        {name: 'Make an empty folder', note: 'Anywhere you will find it again.'},
        {name: 'Point the app at it', note: 'Read the confirmation before you agree.'},
        {name: 'Ask in one sentence', note: 'End with just html css and javascript.'},
        {name: 'Then change one thing', note: 'This is the step that teaches you the most.'},
      ],
      sweepAt: ['Make an empty folder', 'Point the app at it', 'Ask for something you would use', 'change one thing'],
    },
  },
  {
    id: 'habits',
    chapter: 'Next',
    say: 'Three habits will carry you further than any clever phrasing. Always point at a folder you made on purpose. Always read the summary at the end, even the parts you do not follow. And always change one thing straight away, because the change is where you learn that the loop is real.',
    hold: 2.2,
    scene: {
      k: 'steps',
      title: 'Three habits worth more than clever wording',
      steps: [
        {name: 'Point at a folder you made', note: 'On purpose, and empty.'},
        {name: 'Read the summary', note: 'Especially the part about what it cannot do.'},
        {name: 'Change one thing at once', note: 'This is where it stops being a demonstration.'},
      ],
      sweepAt: ['point at a folder', 'read the summary', 'change one thing'],
    },
  },
  {
    id: 'road',
    chapter: 'Next',
    say: 'Here is where this sits. You have just done the first one. The next lessons put it on the internet, give it real accounts, and make it remember things properly. None of those are harder than what you just watched. They are just later.',
    hold: 1.6,
    scene: {
      k: 'roadmap',
      title: 'Where this goes',
      nodes: [
        {name: 'Build it locally', state: 'done'},
        {name: 'Change it by asking', state: 'done'},
        {name: 'Put it on the internet', state: 'here'},
        {name: 'Real accounts', state: 'next'},
        {name: 'Store things properly', state: 'next'},
      ],
    },
  },
  {
    id: 'lastword',
    chapter: 'Next',
    say: 'One last thought before you go. The reason this matters is not that it saves a developer some typing. It is that the number of people who can make software just went from the few who learned a language to anyone who can describe what they want clearly. You are already in the second group. You always were.',
    hold: 2.4,
    scene: {
      k: 'statement',
      kicker: 'Why this actually matters',
      text: 'You are already in the group that can make software. You always were.',
      accent: ['You always were.'],
      size: 78,
    },
  },
  {
    id: 'outro',
    chapter: 'Next',
    say: 'Five files. Forty eight kilobytes. Six and a half minutes, and then thirty one seconds more to change its mind about the colour. Go and make an empty folder.',
    hold: 2.6,
    scene: {
      k: 'outro',
      lines: [
        'One sentence. Six and a half minutes.',
        'Five files that are still there tomorrow.',
        'Go and make an empty folder.',
      ],
      sign: 'Building with Claude',
    },
  },
];
