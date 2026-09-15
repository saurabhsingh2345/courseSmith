// The cut for lesson two.
//
// Source is the twenty ninth of August shoot, already trimmed: the private
// document that was opened by mistake is physically removed from the media, so
// no timestamp in this file can reach it.
//
// Landmarks in source seconds:
//     3  the lesson one app opened from a file
//    37  switching to Claude
//   105  typing "put this on the internet so my friends can open it"
//   129  sent
//   510  live URL exists
//   515  typing "but i want everyone to see the same posts"
//   542  sent
//  1567  shared feed finished
//  1604  picking a username on the live page
//  1635  liking a post
//  1643  reloading
//  1676  193 likes - it held
//  1682  telling Claude it worked
//
// Narration is written to be spoken. No semicolons, no parentheses, no dashes
// mid clause. Small numbers spelled out.

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
  hold?: number;
  min?: number;
  scene: Scene;
};

// Measured off a real frame at 3456 wide.
const INPUT: Rect = {x: 1158, y: 2055, w: 1544, h: 92};
const URLBAR: Rect = {x: 120, y: 150, w: 2400, h: 90};

export const BEATS: Beat[] = [
  // ============================================================ COLD OPEN ===
  {
    id: 'hook',
    chapter: '',
    scene: {k: 'run', from: 1630, to: 1682},
    cues: [
      {at: 1.0, say: 'That is a web address at the top, not a file on somebody’s laptop. Anyone with that link can open this.'},
      {at: 11.0, say: 'Watch the number. One hundred and ninety two.'},
      {at: 20.0, say: 'A like goes in, and now the page is reloaded from scratch, which is the test that matters.'},
      {at: 33.0, say: 'One hundred and ninety three, and it stayed. That like is not in this browser. It is in the page itself, and everybody else opening that link sees it too.'},
    ],
  },
  {
    id: 'title',
    chapter: 'Open',
    say: 'This is lesson two, and it is the one that turns a thing on your laptop into a thing you can send somebody.',
    scene: {
      k: 'title',
      kicker: 'Building with Claude · Lesson two',
      title: 'Put it on the internet',
      sub: 'Two sentences take the app from lesson one and give it a web address, and then make everybody share one feed.',
    },
  },

  // ============================================================ CHAPTER 1 ===
  {id: 'ch1', chapter: 'Where we left off', pad: 0.35, say: 'First, a look at where lesson one left us.',
   scene: {k: 'chapter', n: '01', title: 'Where we left off'}},
  {
    id: 'lastlesson',
    chapter: 'Where we left off',
    scene: {k: 'run', from: 0, to: 44},
    cues: [
      {at: 1.0, say: 'Here is the app from lesson one, opened the way lesson one ended. Straight from the folder, no server, no setup.'},
      {at: 12.0, say: 'And look at the address bar, because this is the whole problem in one line. It says file, then a path on this laptop. That is not a website. There is no way to send that to anybody.'},
      {at: 26.0, say: 'It also says one hundred and ninety two likes, and in lesson one we left it at one hundred and ninety three. The like is gone, and that is not a bug. We will come back to why.'},
    ],
  },
  {
    id: 'whatwehave',
    chapter: 'Where we left off',
    say: 'So this is the honest position at the end of lesson one. Real software, working, on one machine, that nobody else on earth can open. Everything in this lesson is about closing that gap.',
    hold: 1.4,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'folder', who: 'What you have', what: 'Files on your disk. They work. They are yours. They open when you double click them.'},
        {icon: 'web', who: 'What you do not have', what: 'An address. Nobody else can reach it, not even you from your phone.'},
      ],
    },
  },
  {
    id: 'twothings',
    chapter: 'Where we left off',
    say: 'And there are two separate things missing, which people usually run together. One is a place on the internet where the page lives. The other is somewhere the posts and likes are kept that is not just your own browser. We are going to do them one at a time, in that order, because that is the order the problems show up in.',
    hold: 1.8,
    scene: {
      k: 'steps',
      title: 'Two different problems',
      steps: [
        {name: 'A place to live', note: 'An address anybody can type. This one is easy.'},
        {name: 'A shared memory', note: 'Everybody seeing the same posts. This one is the real work.'},
      ],
      sweepAt: ['a place on the internet', 'somewhere the posts and likes are kept'],
    },
  },

  // ============================================================ CHAPTER 2 ===
  {id: 'ch2', chapter: 'Getting an address', pad: 0.35, say: 'Problem one. Give it an address.',
   scene: {k: 'chapter', n: '02', title: 'Getting an address'}},
  {
    id: 'ask1',
    chapter: 'Getting an address',
    scene: {
      k: 'run',
      from: 100,
      to: 140,
      spots: [{rect: INPUT, label: 'the whole request', side: 'top', at: 12.0, until: 30.0}],
    },
    cues: [
      {at: 1.0, say: 'Back to the same window as lesson one, pointed at the same folder. Nothing new to install and nothing to set up.'},
      {at: 12.0, say: 'And here is the sentence. Put this on the internet so my friends can open it. That is the entire instruction.'},
      {at: 26.0, say: 'Notice what is not in it. I did not name a hosting company. I did not say how. I said what I wanted and why I wanted it, and the why is doing more work than you would think.'},
    ],
  },
  {
    id: 'thewhy',
    chapter: 'Getting an address',
    say: 'Those last five words, so my friends can open it, are the most useful thing in the sentence. They tell it what the finished thing is for. A request with a purpose attached gets you a decision. A request without one gets you a question back.',
    hold: 1.6,
    scene: {
      k: 'promptbuild',
      title: 'The sentence, in two halves',
      chunks: [
        {text: 'put this on the internet', tag: 'what you want'},
        {text: 'so my friends can open it', tag: 'what it is for'},
      ],
    },
  },
  {
    id: 'watchit',
    chapter: 'Getting an address',
    scene: {k: 'run', from: 140, to: 500, rate: 6},
    cues: [
      {at: 0.8, say: 'Six times speed through the waiting, because nothing is happening and there is no reason to charge you for it.'},
      {at: 13.0, say: 'It reads the actual files first, not just the folder listing, because it needs to know what kind of thing it is putting online before it can choose where to put it.'},
      {at: 32.0, say: 'Then it makes a decision I did not ask for. It bundles the three files into one document, because the place it has chosen will not fetch separate files alongside the page.'},
      {at: 50.0, say: 'A constraint of the destination, shaping the build, handled without a single question to me.'},
    ],
  },
  {
    id: 'live',
    chapter: 'Getting an address',
    scene: {k: 'run', from: 500, to: 560},
    cues: [
      {at: 1.5, say: 'Six minutes after the sentence, it is live, and there is the address.'},
      {at: 12.0, say: 'It also tells you what it changed and why. The dark mode button was fighting with the page it now lives inside, so it moved where that setting is stored.'},
      {at: 28.0, say: 'And then it says the thing that sets up the rest of this video. Each friend gets their own copy of the posts and likes. Their likes are private to them and will not affect anybody else.'},
      {at: 46.0, say: 'Read that twice. It has done exactly what I asked and it is telling me, unprompted, why what I asked for is not what I actually wanted.'},
    ],
  },
  {
    id: 'private',
    chapter: 'Getting an address',
    say: 'One practical detail before we move on. A page like this starts private, which means the link works for you and nobody else until you say otherwise. There is a share control on the page itself. That default is the right way round, and it is worth knowing rather than discovering.',
    hold: 1.2,
    scene: {
      k: 'cards',
      title: 'Before you send the link',
      items: [
        {name: 'It starts private', note: 'The address exists, but only you can open it.', icon: 'keys'},
        {name: 'You turn sharing on', note: 'One control, on the page itself.', icon: 'web'},
        {name: 'Then the link works', note: 'For anyone you send it to.', icon: 'star'},
      ],
      active: 1,
      footer: 'Private by default is the correct default. Now you know it is there.',
    },
  },

  // ============================================================ CHAPTER 3 ===
  {
    id: 'whatisaddress',
    chapter: 'Getting an address',
    say: 'And it is worth being precise about the word address, because it is doing a lot of quiet work. An address is not a copy of your files. It is a place other people can point their browser at, where a copy of your page is kept running for them.',
    hold: 1.6,
    scene: {
      k: 'define',
      word: 'address',
      pos: 'noun',
      meaning: 'A place other people can point a browser at.',
      also: 'Not a copy you send them. A place that answers when they knock.',
    },
  },
  {
    id: 'whoseesit',
    chapter: 'Getting an address',
    say: 'There are really three levels of who can see a thing, and they are worth knowing apart. Only you. Anyone who has the link. Or anyone at all, including search engines. Most beginners think there are two, and the middle one is the one you will use most.',
    hold: 1.8,
    scene: {
      k: 'scale',
      title: 'Who can see it',
      rungs: ['only you', 'anyone with the link', 'anyone at all'],
      pick: 1,
      left: 'private',
      right: 'public',
      note: 'Anyone with the link is where almost everything you make should sit.',
    },
  },
  {id: 'ch3', chapter: 'The catch', pad: 0.35, say: 'Which brings us to the catch.',
   scene: {k: 'chapter', n: '03', title: 'The catch'}},
  {
    id: 'thecatch',
    chapter: 'The catch',
    say: 'The page is on the internet. The posts are not. Every person who opens that link gets their own private copy of the feed, kept inside their own browser, and nothing any of them does is ever seen by anybody else.',
    hold: 1.8,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'web', who: 'Shared', what: 'The page itself. The layout, the pictures, the code. Everyone gets the same one.'},
        {icon: 'keys', who: 'Not shared', what: 'Every post, like and comment. Each viewer has a private copy nobody else can see.'},
      ],
    },
  },
  {
    id: 'whythat',
    chapter: 'The catch',
    say: 'And this is exactly why the like from lesson one disappeared at the start of this video. That like was saved against the address we were using then. Open the same app from a different address and it is a different drawer, with nothing in it.',
    hold: 1.6,
    scene: {
      k: 'cards',
      title: 'Why the like vanished',
      items: [
        {name: 'Lesson one', note: 'Liked at one address. Saved in that drawer.', icon: 'folder'},
        {name: 'Today', note: 'Same app, different address. Different drawer.', icon: 'web'},
        {name: 'The drawer is empty', note: 'Not broken. Just somewhere else.', icon: 'keys'},
      ],
      sweepAt: ['saved against the address', 'a different address', 'nothing in it'],
    },
  },
  {
    id: 'testyourself',
    chapter: 'The catch',
    say: 'Here is the test that shows it, and it takes ten seconds. Open your own link on your phone. Everything you did on the laptop is missing. Not lost, just somewhere else, and that is the moment this stops being an abstract idea.',
    scene: {
      k: 'devices',
      title: 'Same link, two devices, two separate memories',
    },
  },

  // ============================================================ CHAPTER 4 ===
  {
    id: 'statekinds',
    chapter: 'The catch',
    say: 'It helps to name the thing that is missing. Every app holds information, and it is held in one of three places. Inside the page while you look at it. Inside your own browser between visits. Or somewhere central that everybody reads from. Lesson one used the middle one. We want the third.',
    hold: 2.2,
    scene: {
      k: 'steps',
      title: 'Three places information can live',
      steps: [
        {name: 'In the page, right now', note: 'Gone the moment you refresh.'},
        {name: 'In your own browser', note: 'Survives a refresh. Yours alone. This is where lesson one stopped.'},
        {name: 'Somewhere central', note: 'Everybody reads the same copy. This is what we want.'},
      ],
      sweepAt: ['Inside the page while you look at it', 'Inside your own browser', 'somewhere central'],
    },
  },
  {
    id: 'whennotshare',
    chapter: 'The catch',
    say: 'And before we go and get it, one honest caution. Shared is not automatically better. A private notebook should stay private. A personal habit tracker probably should not be visible to everybody who has the link. Ask what the thing is for before you make it shared.',
    hold: 1.8,
    scene: {
      k: 'cards',
      title: 'Should it actually be shared',
      items: [
        {name: 'A feed, a board, a guest book', note: 'Yes. Being shared is the point.', icon: 'web'},
        {name: 'A personal tracker or notebook', note: 'No. Private is the feature.', icon: 'keys'},
        {name: 'Anything with real personal data', note: 'Not until you understand who can reach it.', icon: 'sprout'},
      ],
      active: 0,
      footer: 'Shared is a choice, not an upgrade.',
    },
  },
  {id: 'ch4', chapter: 'Asking properly', pad: 0.35, say: 'So we ask for the real thing.',
   scene: {k: 'chapter', n: '04', title: 'Asking for the real thing'}},
  {
    id: 'ask2',
    chapter: 'Asking properly',
    scene: {
      k: 'run',
      from: 512,
      to: 552,
      spots: [{rect: INPUT, label: 'the whole request, again', side: 'top', at: 16.0, until: 32.0}],
    },
    cues: [
      {at: 1.0, say: 'No new session, no starting over. Same conversation, straight underneath the last answer.'},
      {at: 14.0, say: 'But i want everyone to see the same posts, not their own copy. Twelve words, and I have spelled nothing correctly.'},
      {at: 30.0, say: 'I have not said database. I have not said server or account or backend. I described the outcome I wanted in the words I would use to a friend.'},
    ],
  },
  {
    id: 'nojargon',
    chapter: 'Asking properly',
    say: 'This is worth stopping on, because it is the single most transferable habit in this course. You are allowed to describe the result and let it choose the machinery. The words database and backend are answers, not questions, and if you guess the answer wrong you have just sent it down the wrong road.',
    hold: 1.8,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'star', who: 'Say the outcome', what: 'I want everyone to see the same posts. Now it can pick the right machinery.'},
        {icon: 'keys', who: 'Not the machinery', what: 'Add a database with a posts table. You have guessed, and you might be wrong.'},
      ],
    },
  },
  {
    id: 'itasks',
    chapter: 'Asking properly',
    say: 'And then it does something worth waiting for. It stops, and asks a question. When a friend opens the shared page, what should they be able to do. It has spotted a genuine fork in the road and it will not guess which way to go.',
    hold: 1.6,
    scene: {
      k: 'statement',
      kicker: 'The good sign',
      text: 'It stopped and asked, instead of guessing.',
      accent: ['instead of guessing.'],
      size: 86,
    },
  },
  {
    id: 'whyask',
    chapter: 'Asking properly',
    say: 'A question at this point is not the tool being slow. It is the tool having noticed that your sentence had two reasonable meanings and that picking the wrong one wastes the next fifteen minutes. Answer it in a few words and move on.',
    scene: {
      k: 'cards',
      title: 'When it asks you something',
      items: [
        {name: 'Answer briefly', note: 'A few words. Do not write an essay.', icon: 'bolt'},
        {name: 'Do not over think it', note: 'You can change your mind later by saying so.', icon: 'loop'},
        {name: 'Be glad it asked', note: 'The alternative is fifteen minutes in the wrong direction.', icon: 'star'},
      ],
      active: 2,
    },
  },

  // ============================================================ CHAPTER 5 ===
  {id: 'ch5', chapter: 'The long build', pad: 0.35, say: 'Now the longest stretch of work in this course so far.',
   scene: {k: 'chapter', n: '05', title: 'The long build'}},
  {
    id: 'longbuild',
    chapter: 'The long build',
    scene: {k: 'run', from: 546, to: 1284, rate: 8},
    cues: [
      {at: 0.8, say: 'Eight times speed. In real time this stretch is twelve minutes, and almost none of it is worth your attention at full pace.'},
      {at: 14.0, say: 'It reads the whole application first. Its own words are that it wants to restructure it rather than patch the bundle, which is the difference between a repair and a mess.'},
      {at: 36.0, say: 'Then it separates two ideas that were tangled together. What belongs to everybody, and what belongs to you alone. Nearly all of this build is that one distinction, applied carefully.'},
      {at: 60.0, say: 'Notice how little of it is writing. It is reading, running, checking, and going back. Then it hits a real bug, and that part we are going to watch properly.'},
    ],
  },
  {
    id: 'thebug',
    chapter: 'The long build',
    scene: {k: 'run', from: 1284, to: 1404, rate: 1.5},
    cues: [
      {at: 1.0, say: 'Here it is. Something is duplicating chunks of the file, and it goes and finds out why rather than trying something else and hoping.'},
      {at: 16.0, say: 'The cause is genuinely obscure. A character sequence inside the code was being read as an instruction by the tool doing the copying, so parts of the file were being repeated.'},
      {at: 34.0, say: 'You do not need to understand that. What you need to notice is that it was found, named, and fixed, and that you were not asked to help.'},
      {at: 47.0, say: 'Then a second one, and this is the better story. It realises it has just been checking a stale copy of the file rather than the real one. It caught itself being wrong about its own test.'},
    ],
  },
  {
    id: 'checking',
    chapter: 'The long build',
    say: 'That is the part I would tattoo on a beginner. It did not just test the thing. It questioned whether the test was testing anything, found that it was not, and started again. If you ever wonder what to look for when reading these summaries, that is it.',
    hold: 1.6,
    scene: {
      k: 'steps',
      title: 'What good checking looks like',
      steps: [
        {name: 'Run it', note: 'Not read it. Run it.'},
        {name: 'Ask whether the test was real', note: 'This is the step almost everybody skips.'},
        {name: 'Fix, then run it again', note: 'Until the result means something.'},
      ],
      sweepAt: ['did not just test the thing', 'whether the test was testing anything', 'started again'],
    },
  },

  // ============================================================ CHAPTER 6 ===
  {
    id: 'ratio',
    chapter: 'The long build',
    say: 'One number worth holding on to from this stretch. Of everything it did in those twelve minutes, only a small part was writing. The rest was reading what exists, running it, looking at the result, and correcting. If you ever try to judge whether one of these tools is any good, that ratio is what to judge it on.',
    hold: 1.8,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'braces', who: 'What people imagine', what: 'It writes code very quickly, and that is the whole trick.'},
        {icon: 'loop', who: 'What actually happens', what: 'Mostly reading, running, checking and correcting. The writing is the small part.'},
      ],
    },
  },
  {id: 'ch6', chapter: 'What it built', pad: 0.35, say: 'It finishes. Look at what the twelve words bought.',
   scene: {k: 'chapter', n: '06', title: 'What twelve words bought'}},
  {
    id: 'finish',
    chapter: 'What it built',
    scene: {k: 'run', from: 1500, to: 1570},
    cues: [
      {at: 1.5, say: 'It hands back a table, and this table is the whole design of the thing in about twenty words.'},
      {at: 13.0, say: 'Shared by everyone. Posts, captions, comments, likes, profiles. Yours alone. Your username, your theme, your saved posts, who you follow.'},
      {at: 30.0, say: 'And a line explaining one of those choices. Dark mode stays personal, because otherwise one friend hitting the moon button flips everybody else’s screen.'},
      {at: 46.0, say: 'Nobody asked for that distinction. It is the kind of judgement you would expect from somebody who has built a shared thing before and been bitten by it.'},
    ],
  },
  {
    id: 'sharedtable',
    chapter: 'What it built',
    say: 'That split is the entire lesson of this video, so here it is on its own. Anything that is a fact about the world is shared. Anything that is a fact about you is not. Getting that line in the right place is most of what building a shared thing is.',
    hold: 2.2,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'web', who: 'Shared by everyone', what: 'Posts. Captions. Comments. Likes. Profiles. Facts about the world.'},
        {icon: 'mobile', who: 'Yours alone', what: 'Your name. Your theme. Your saved posts. Who you follow. Facts about you.'},
      ],
    },
  },
  {
    id: 'details',
    chapter: 'What it built',
    say: 'And underneath, a set of decisions I never mentioned. Likes had to stop being a number and become a list of who, so the count can be shared while whether you liked it stays yours. Photographs get shrunk on the way in. A post is refused outright if it would make the page too heavy for everybody else.',
    hold: 2.4,
    scene: {
      k: 'cards',
      title: 'Decisions nobody asked for',
      items: [
        {name: 'Likes became a list', note: 'Who liked it, not how many. The count is shared, the tick is yours.', icon: 'star'},
        {name: 'Photos get shrunk', note: 'A phone snapshot is several megabytes. Everyone would carry it.', icon: 'mobile'},
        {name: 'Oversized posts refused', note: 'Rather than quietly breaking the page for all of you.', icon: 'keys'},
        {name: 'Faces drawn from names', note: 'So a stranger who comments still has a face, at no cost.', icon: 'sprout'},
      ],
      sweepAt: ['Likes had to stop being a number', 'Photographs get shrunk', 'refused outright'],
    },
  },
  {
    id: 'notasked',
    chapter: 'What it built',
    say: 'None of that was in my twelve words. All of it follows from them. This is the thing people mean when they say the sentence does more than it looks like, and it is why a short honest request beats a long confused one.',
    scene: {
      k: 'statement',
      text: 'Twelve badly spelled words. All of that followed from them.',
      accent: ['All of that followed from them.'],
      size: 80,
    },
  },

  // ============================================================ CHAPTER 7 ===
  {
    id: 'conflict',
    chapter: 'What it built',
    say: 'There is also a limitation it chose, deliberately, and told me about. If two people do something at the exact same moment, the last one wins and the other person’s action can be dropped. It decided that was acceptable for friends messing about and said so rather than hiding it.',
    hold: 1.8,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'bolt', who: 'What it built', what: 'Simple and fast. Last action wins if two land at the same instant.'},
        {icon: 'server', who: 'What it skipped', what: 'Merging simultaneous edits. Real work, rarely worth it for six friends.'},
      ],
    },
  },
  {id: 'ch7', chapter: 'What it could not check', pad: 0.35, say: 'And then the most valuable paragraph in the whole session.',
   scene: {k: 'chapter', n: '07', title: 'The honest paragraph'}},
  {
    id: 'couldnot',
    chapter: 'What it could not check',
    say: 'It lists what it tested and found working. Then it says, in plain words, what it could not test. It could not reach the live page, because the browser it has access to is not signed in as you. So one link in the chain is unverified, and it tells you which one.',
    hold: 1.8,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'bolt', who: 'Verified', what: 'Identity, likes, comments, posting, reload, and the page rebuilding itself. No errors.'},
        {icon: 'keys', who: 'Could not verify', what: 'The live save, against the real address. Its browser is not signed in as you.'},
      ],
    },
  },
  {
    id: 'whyhonest',
    chapter: 'What it could not check',
    say: 'A tool that says I could not check this part is more useful than one that says everything is fine, because you now know exactly where to look if something is wrong. And it did not stop there. It told me how to run that test myself, in one sentence.',
    hold: 1.4,
    scene: {
      k: 'statement',
      kicker: 'The habit to build',
      text: 'Trust the one that tells you what it could not check.',
      accent: ['what it could not check.'],
      size: 82,
    },
  },
  {
    id: 'thetest',
    chapter: 'What it could not check',
    scene: {k: 'run', from: 1596, to: 1690},
    cues: [
      {at: 1.5, say: 'So we do the part it cannot. Open the real link, in a real browser, signed in as a real person.'},
      {at: 14.0, say: 'And immediately something new. It asks who you are. Everyone opening this link shares one feed, and this is the name your posts and likes show up under.'},
      {at: 30.0, say: 'That screen did not exist twenty minutes ago and I never asked for it. It is what the question earlier was about.'},
      {at: 44.0, say: 'Now the actual test. One hundred and ninety two likes. A click.'},
      {at: 56.0, say: 'And a full reload, which throws away everything the browser was holding in memory. If the like only ever lived in this browser, it dies here.'},
      {at: 76.0, say: 'One hundred and ninety three. It held. The page went away and came back carrying the like with it.'},
    ],
  },

  // ============================================================ CHAPTER 8 ===
  {id: 'ch8', chapter: 'Closing the loop', pad: 0.35, say: 'And then the part almost nobody does.',
   scene: {k: 'chapter', n: '08', title: 'Closing the loop'}},
  {
    id: 'tellit',
    chapter: 'Closing the loop',
    scene: {k: 'run', from: 1680, to: 1723},
    cues: [
      {at: 1.0, say: 'Tell it what happened. It works, I liked a post and it was still there after a refresh.'},
      {at: 13.0, say: 'And it goes and looks. Not takes my word for it. It reads the live page and reports what it found there.'},
      {at: 24.0, say: 'A viewer picked the name alex. They liked that post. It was recorded as a name on a list, on top of the original count, and written into the activity feed.'},
    ],
  },
  {
    id: 'loopvalue',
    chapter: 'Closing the loop',
    say: 'Telling it the result costs you eight seconds and it is the highest value thing you will do all session. It turns a guess into a fact, and it means the next change is built on something true rather than something assumed.',
    hold: 1.4,
    scene: {
      k: 'steps',
      title: 'The loop, completed',
      steps: [
        {name: 'It says what it could not check', note: 'Honestly, and specifically.'},
        {name: 'You go and check it', note: 'Ten seconds. You are the one with the login.'},
        {name: 'You tell it what happened', note: 'Eight seconds. Now it is a fact.'},
      ],
      sweepAt: ['Tell it the result', 'a guess into a fact', 'the next change'],
    },
  },
  {
    id: 'warning',
    chapter: 'Closing the loop',
    say: 'And then it gives you a warning that nobody asked for, and it is the most important sentence in this video. The live page is now the source of truth, not the copy on your laptop. If it rebuilds and republishes carelessly, it would wipe everybody’s posts back to the starting content.',
    hold: 2.2,
    scene: {
      k: 'statement',
      kicker: 'Read this one twice',
      text: 'The live page is now the truth. Do not let it republish blind.',
      accent: ['Do not let it republish blind.'],
      size: 78,
    },
  },
  {
    id: 'whatthatmeans',
    chapter: 'Closing the loop',
    say: 'In ordinary language. The moment other people start using your thing, their stuff lives in it, and every change you make afterwards has to carry their stuff forward. That is the actual difference between a project and a product, and you have just crossed it.',
    hold: 1.8,
    scene: {
      k: 'swap',
      panels: [
        {icon: 'sprout', who: 'Before today', what: 'Break it, rebuild it, start over. Nobody is affected but you.'},
        {icon: 'web', who: 'From today', what: 'Other people’s posts live in it. Every change has to carry them forward.'},
      ],
    },
  },

  // ============================================================ CHAPTER 9 ===
  {
    id: 'safety',
    chapter: 'Closing the loop',
    say: 'Two things to be careful about now that the thing is reachable. Anything you type into it is visible to everyone with the link, so do not put anything private in there. And the page carries its own content inside it, so a hundred large photographs will make it slow for everybody, which is exactly why it refuses the oversized ones.',
    hold: 2.0,
    scene: {
      k: 'cards',
      title: 'Two things to watch now',
      items: [
        {name: 'Everyone with the link sees everything', note: 'Nothing you type in there is private. Behave accordingly.', icon: 'keys'},
        {name: 'The page carries its own contents', note: 'Heavy photos make it slow for everybody, not just you.', icon: 'mobile'},
      ],
      footer: 'Both of these are consequences of how simple it is. That is the trade.',
    },
  },
  {
    id: 'nextup',
    chapter: 'Closing the loop',
    say: 'And this is the point where the approach we used today runs out of road. It is genuinely excellent for a handful of friends. It is not how a real product is built, and the next lesson is about the arrangement that is.',
    hold: 1.6,
    scene: {
      k: 'spotlight',
      title: 'Where today’s approach stops',
      name: 'Good for six friends',
      tagline: 'Not how a real product is built',
      icon: 'sprout',
      facts: [
        'Everything lives inside one page.',
        'Simultaneous edits can collide.',
        'Fine for a group chat. Not for a business.',
      ],
    },
  },
  {id: 'ch9', chapter: 'Next', pad: 0.35, say: 'So where does that leave you.',
   scene: {k: 'chapter', n: '09', title: 'Where that leaves you'}},
  {
    id: 'recap',
    chapter: 'Next',
    say: 'Two sentences. That is the entire input for this lesson. Put this on the internet so my friends can open it. Then, but i want everyone to see the same posts, not their own copy. Everything else in these thirty minutes followed from those.',
    hold: 2.4,
    scene: {
      k: 'steps',
      title: 'The whole lesson, in two sentences',
      steps: [
        {name: 'put this on the internet so my friends can open it', note: 'Six minutes. A real address.'},
        {name: 'but i want everyone to see the same posts', note: 'Seventeen minutes. One shared feed.'},
      ],
      sweepAt: ['Put this on the internet', 'but i want everyone'],
    },
  },
  {
    id: 'check',
    chapter: 'Next',
    say: 'One check before we finish, because if you only remember one thing from this lesson it should be this one.',
    scene: {
      k: 'quiz',
      n: '01',
      question: 'Your page is online and a friend likes a post. Where does that like live?',
      options: ['In their browser, private to them', 'In the page itself, shared with everyone', 'It depends on how you asked'],
      answer: 2,
      revealAt: 150,
    },
  },
  {
    id: 'answer',
    chapter: 'Next',
    say: 'It depends on how you asked. After the first sentence it was private to them. After the second it was shared. Same app, same link, same afternoon. The difference was one sentence describing what you actually wanted.',
    scene: {
      k: 'statement',
      text: 'Same app. Same link. The difference was one sentence.',
      accent: ['one sentence.'],
      size: 84,
    },
  },
  {
    id: 'homework',
    chapter: 'Next',
    say: 'Your turn, and it is small. Take whatever you built after lesson one and ask for an address for it. Open that address on your phone. Then decide, honestly, whether the thing you are building needs everybody to share one copy or not, because that answer changes everything that comes after.',
    hold: 2.0,
    scene: {
      k: 'steps',
      title: 'Do this today',
      steps: [
        {name: 'Ask for an address', note: 'One sentence. Say what it is for.'},
        {name: 'Open it on your phone', note: 'This is the moment it becomes real.'},
        {name: 'Decide if it should be shared', note: 'Be honest. Not everything should be.'},
      ],
      sweepAt: ['ask for an address', 'Open that address on your phone', 'needs everybody to share'],
    },
  },
  {
    id: 'road',
    chapter: 'Next',
    say: 'And here is the road. Lesson one built it. This one gave it an address and a shared memory. Next we put it on a proper host of its own, with a real database behind it, which is the arrangement almost every application you use is built on.',
    hold: 1.8,
    scene: {
      k: 'roadmap',
      title: 'Where this goes',
      nodes: [
        {name: 'Build it locally', state: 'done'},
        {name: 'Change it by asking', state: 'done'},
        {name: 'Give it an address', state: 'done'},
        {name: 'One shared feed', state: 'here'},
        {name: 'A real host and database', state: 'next'},
      ],
    },
  },
  {
    id: 'outro',
    chapter: 'Next',
    say: 'Two sentences, twenty three minutes of machine time, and a link you can send to somebody. Go and send it to somebody.',
    hold: 2.6,
    scene: {
      k: 'outro',
      lines: [
        'Two sentences. One address.',
        'One feed that everybody shares.',
        'Go and send the link to somebody.',
      ],
      sign: 'Building with Claude · Lesson two',
    },
  },
];
