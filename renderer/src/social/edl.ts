// The cut.
//
// One beat, one idea, one line of narration. A beat's length on screen is
// `lead + however long the line actually takes to say + pad + hold`, measured by
// the voice builder rather than guessed here, so rewriting a sentence reflows
// the film instead of pushing everything after it out of sync.
//
// There are no fixed-length footage runs in this cut, because there is no
// footage: the app and the assistant are both components, so every beat is free
// to be exactly as long as its sentence. That is the main structural difference
// from the lesson this follows, and it is why nothing here has to be padded to
// fill a gap or clipped to fit one.
//
// Narration is written to be *spoken*. No semicolons, no parentheses, no dashes
// mid-clause — Kokoro reads all three as a full stop and the line arrives in
// pieces. Small numbers are spelled out for the same reason.

import type {ChatStep} from './Chat';
import type {ScreenSpec} from './Screen';
import type {GlyphName} from './kit';

export type Scene =
  | {k: 'title'; title: string; kicker: string; sub: string}
  | {k: 'chapter'; n: string; title: string}
  | {k: 'statement'; kicker?: string; text: string; accent?: string[]; serif?: string[]; body?: string; size?: number}
  | {k: 'beforeafter'; title: string; left: {label: string; items: string[]}; right: {label: string; items: string[]}}
  | {k: 'numbers'; title: string; stats: {value: string; label: string; note?: string}[]}
  | {k: 'checklist'; title: string; items: {name: string; note: string}[]; sweepAt?: string[]}
  | {k: 'define'; word: string; pos: string; meaning: string; also: string}
  | {k: 'anatomy'; title: string; chunks: {text: string; tag: string}[]; sweepAt?: string[]}
  | {k: 'compare'; title: string; good: {label: string; items: string[]}; bad: {label: string; items: string[]}}
  | {k: 'ladder'; title: string; rungs: {name: string; note: string}[]; pick: number}
  | {k: 'tiles'; title: string; items: {name: string; note: string; icon: GlyphName}[]; sweepAt?: string[]}
  | {k: 'loop'; title: string; steps: {name: string; note: string}[]}
  | {k: 'quiz'; n: string; question: string; options: string[]; answer: number; revealAt: number}
  | {k: 'timeline'; title: string; nodes: {name: string; state: 'done' | 'here' | 'next'}[]}
  | {k: 'outro'; lines: string[]; sign: string}
  | {k: 'chat'; title: string; steps: ChatStep[]; zoom?: number}
  | {k: 'screen'; spec: ScreenSpec};

export type Beat = {
  id: string;
  chapter: string;
  say: string;
  lead?: number;
  pad?: number;
  /** Extra stillness after the line, for a beat that needs to be read. */
  hold?: number;
  min?: number;
  scene: Scene;
};

// Lumen's own logical pixels, for callouts. The window is 1440 by 900, the
// sidebar is 244 wide, the feed column sits at x 363 and is 618 across, and the
// right hand rail starts at 1021.
const FEED_X = 363;
const RAIL_X = 1021;
// Explore and profile use a wider single column, starting further left.
const WIDE_X = 375;

export const BEATS: Beat[] = [
  // ===================================================================== open
  {
    id: 'hook',
    chapter: 'Open',
    say: 'In about half an hour you are going to have a working social app. Photos, a feed, profiles, comments, the whole thing. And you are not going to write a single line of code to get it. I want to be clear that I do not mean a mockup, and I do not mean a drawing of an app. I mean software, running in a browser, that you can click.',
    hold: 0.6,
    scene: {
      k: 'statement',
      kicker: 'Before we start',
      text: 'You are about to build a social network without writing any code.',
      accent: ['code.'],
      serif: ['without'],
      size: 88,
    },
  },
  {
    id: 'title',
    chapter: 'Open',
    say: 'This is lesson one. Building real software by describing it.',
    scene: {
      k: 'title',
      kicker: 'No-code with Claude · Lesson one',
      title: 'Build an app without writing code',
      sub: 'We will design a photo sharing app, describe it in plain English, and have it running in a browser before the end of this video.',
    },
  },

  // ================================================================ chapter 1
  {id: 'ch1', chapter: 'What this is', say: 'First, what this actually is.', pad: 0.35,
   scene: {k: 'chapter', n: '01', title: 'What this actually is'}},
  {
    id: 'who',
    chapter: 'What this is',
    say: 'Three kinds of people end up watching this. Someone with an idea and no idea where to start. Someone who tried to learn to code once and quit somewhere around month two, which is the normal place to quit. And someone who can already build things, who just wants to go faster. All three of you are in the right place, because the loop I am about to show you is the same loop for all three.',
    hold: 1.6,
    scene: {
      k: 'tiles',
      title: 'Who this is for',
      items: [
        {name: 'The idea', note: 'You know what you want. You have no idea where to start.', icon: 'sparkle'},
        {name: 'The false start', note: 'You tried to learn. You quit around month two.', icon: 'clock'},
        {name: 'The builder', note: 'You can already do this. You want to go faster.', icon: 'wand'},
      ],
      sweepAt: ['an idea and no idea', 'quit somewhere around month two', 'already build things'],
    },
  },
  {
    id: 'whatchanged',
    chapter: 'What this is',
    say: 'Not long ago, building the app we are about to build meant learning several separate skills first. You needed a language. You needed to understand how a page talks to a server. You needed a lot of patience, and you needed months before anything at all worked. That has genuinely changed, and I do not mean that in a marketing way. The skill now is describing what you want clearly, and reading what comes back honestly.',
    hold: 1.6,
    scene: {
      k: 'beforeafter',
      title: 'What actually changed',
      left: {
        label: 'The old way in',
        items: [
          'Learn a language first',
          'Learn how the pieces connect',
          'Months before anything runs',
          'Give up somewhere in month two',
        ],
      },
      right: {
        label: 'What you do now',
        items: [
          'Describe what you want',
          'Look at what it made',
          'Say what to change',
          'Have it running this afternoon',
        ],
      },
    },
  },
  {
    id: 'notmagic',
    chapter: 'What this is',
    say: 'One thing so you are not confused later. It is not searching the internet for an app that looks like yours and handing you a copy. It has read an enormous amount of code, and it is writing new code, for you, in the moment. That is why it can build the specific thing you described. It is also why it sometimes gets things wrong in ways a search engine never would, and we will deal with that properly in chapter eight.',
    hold: 1.0,
    scene: {
      k: 'statement',
      kicker: 'How it actually works',
      text: 'It is not finding you an app. It is writing one.',
      accent: ['writing'],
      body: 'Which is why it can build your exact idea, and also why it can be confidently wrong. Both come from the same place.',
      size: 88,
    },
  },
  {
    id: 'loop',
    chapter: 'What this is',
    say: 'The whole lesson is really one loop, repeated. You describe something. You look at what you got. You say what is wrong with it. And then you go round again. Everything else we cover today is just doing those three steps well, and the third one is where almost all the skill lives.',
    hold: 1.4,
    scene: {
      k: 'loop',
      title: 'The whole job is one loop',
      steps: [
        {name: 'Describe', note: 'Say what you want in plain language.'},
        {name: 'Look', note: 'Open it. Actually use it.'},
        {name: 'Change', note: 'Name the one thing that is wrong.'},
      ],
    },
  },
  {
    id: 'honest',
    chapter: 'What this is',
    say: 'Let me be straight with you about the limits, because a lot of videos on this subject will not be. You are not going to build the next big social network this afternoon. What you can build is something real, that works, that you can show to another person, and that you can keep changing. That is a much bigger deal than it sounds, because the gap between nothing and something that runs is the hard one.',
    hold: 1.2,
    scene: {
      k: 'statement',
      kicker: 'The honest version',
      text: 'Not a company. A real, working thing you can keep changing.',
      accent: ['working'],
      serif: ['real,'],
      body: 'The gap between nothing and something that runs is the hard one. That is the gap we are crossing today.',
      size: 82,
    },
  },
  {
    id: 'numbers',
    chapter: 'What this is',
    say: 'To set your expectations properly, here is what today actually cost. Everything you are about to see was built in one sitting, from four plain English requests, and the app it produced is a bit over eight hundred lines of code. You will never need to read those lines. But it is worth knowing they are real files, sitting on a real computer, and that you could open them if you ever wanted to.',
    hold: 1.4,
    scene: {
      k: 'numbers',
      title: 'What one sitting actually produces',
      stats: [
        {value: 'One', label: 'sitting', note: 'Start to a running app, with no prior setup.'},
        {value: 'Four', label: 'requests', note: 'Plain English. No syntax, no commands.'},
        {value: '800+', label: 'lines of code', note: 'Real files, which you never have to read.'},
      ],
    },
  },

  // ================================================================ chapter 2
  {id: 'ch2', chapter: 'Setting up', say: 'What you need in front of you.', pad: 0.35,
   scene: {k: 'chapter', n: '02', title: 'What you actually need'}},
  {
    id: 'needs',
    chapter: 'Setting up',
    say: 'Four things, and you probably have three of them already. A computer, and it does not matter which one. Claude, which is the assistant that does the building. A browser, to look at what you made. And a rough idea of what you want. That last one is the only item people skip, and it is the one that matters most, which is why we are going to spend a whole chapter on it.',
    hold: 2.7,
    scene: {
      k: 'checklist',
      title: 'Four things, and you have most of them',
      items: [
        {name: 'A computer', note: 'Mac, Windows, Linux. It genuinely does not matter.'},
        {name: 'Claude', note: 'The assistant that writes and runs the files.'},
        {name: 'A browser', note: 'To open what you built and actually use it.'},
        {name: 'A rough idea', note: 'The one people skip. The one that matters most.'},
      ],
      sweepAt: ['A computer', 'Claude', 'A browser', 'a rough idea'],
    },
  },
  {
    id: 'install',
    chapter: 'Setting up',
    say: 'Getting Claude in front of you takes about two minutes. Go to claude dot a i in a browser and make an account, or download the desktop app if you would rather have it sitting in your dock. Sign in. And that is the setup finished. There is no project to create, nothing to configure, and no settings you need to change before you start.',
    hold: 1.6,
    scene: {
      k: 'checklist',
      title: 'Two minutes of setup',
      items: [
        {name: 'Open claude.ai', note: 'Or install the desktop app. Either is fine.'},
        {name: 'Sign in', note: 'A free account is enough for this lesson.'},
        {name: 'Start typing', note: 'No project to create. No settings to change.'},
      ],
      sweepAt: ['Go to claude dot a i', 'Sign in', 'that is the setup finished'],
    },
  },
  {
    id: 'chat-hello',
    chapter: 'Setting up',
    say: 'If you have never used it for something like this, spend one message finding out what it will actually do. Ask it directly, in exactly those words. The answer is genuinely useful, and more importantly it gets you past the blank page feeling, which is the real obstacle on day one.',
    hold: 3.3,
    scene: {
      k: 'chat',
      title: 'A first message',
      steps: [
        {k: 'you', text: 'Can you build me a working web app if I do not know how to code?'},
        {k: 'wait'},
        {k: 'say', text: 'Yes. You describe what you want in plain language, I write the files and run them, and you look at the result and tell me what to change.'},
        {k: 'say', text: 'You will not need to read any code. You will need to be specific about what you want, and honest about what looks wrong.'},
        {k: 'done', text: 'What would you like to build?'},
      ],
    },
  },
  {
    id: 'nothing-else',
    chapter: 'Setting up',
    say: 'Now notice what is not on that list. No account with a hosting company. No credit card. No editor to install, no terminal to learn, and no thirty gigabyte download that fails at ninety percent. If a beginner tutorial opens by asking you to install five things, you are watching the wrong tutorial for today, and you should close it.',
    hold: 1.2,
    scene: {
      k: 'statement',
      kicker: 'Not on the list',
      text: 'No editor. No terminal. No card. Nothing to install.',
      accent: ['Nothing'],
      body: 'If a beginner tutorial opens with five installs, it is not a beginner tutorial. Close it.',
      size: 86,
    },
  },

  // ================================================================ chapter 3
  {id: 'ch3', chapter: 'How to ask', say: 'Now the actual skill.', pad: 0.35,
   scene: {k: 'chapter', n: '03', title: 'How to ask for a thing'}},
  {
    id: 'define',
    chapter: 'How to ask',
    say: 'You are going to hear the word prompt a great deal, so it is worth pinning down. A prompt is just what you asked for, written down. That is genuinely all it is. There is no special syntax to memorise, and there are no magic words that unlock a better answer. A clear request from somebody who knows what they want beats a clever one every single time.',
    hold: 1.6,
    scene: {
      k: 'define',
      word: 'prompt',
      pos: 'noun',
      meaning: 'What you asked for, written down. Nothing more technical than that.',
      also: 'There is no syntax and there are no magic words. Clear beats clever, every time.',
    },
  },
  {
    id: 'anatomy',
    chapter: 'How to ask',
    say: 'A request that works usually has four parts, whether or not you thought about them. What the thing is. What is in it. Who it is for. And how it should feel. Most people say the first one and stop, and then wonder why what came back feels generic. Watch what happens when I ask for a photo app and say all four of those out loud.',
    hold: 3.1,
    scene: {
      k: 'anatomy',
      title: 'Four parts of a request that works',
      chunks: [
        {text: 'Build me a photo sharing app', tag: 'what it is'},
        {text: 'with a feed, profiles and comments', tag: 'what is in it'},
        {text: 'for people who post film photos', tag: 'who it is for'},
        {text: 'calm, black and white, photos do the talking', tag: 'how it feels'},
      ],
      sweepAt: ['What the thing is', 'What is in it', 'Who it is for', 'how it should feel'],
    },
  },
  {
    id: 'compare',
    chapter: 'How to ask',
    say: 'Here is the same idea as a before and after. On the left is what most people type the first time. On the right is what gets you something you actually like. The difference is not length, and it is not vocabulary. It is that the right hand column made a decision, and the left hand column left every single decision to somebody else.',
    hold: 2.9,
    scene: {
      k: 'compare',
      title: 'The difference is a decision, not a vocabulary',
      bad: {
        label: 'leaves it open',
        items: [
          '"Make me a social media app"',
          '"Make it look good"',
          '"Add the usual features"',
          '"Fix the design"',
        ],
      },
      good: {
        label: 'makes a decision',
        items: [
          '"A photo feed, one column, newest first"',
          '"Black and white chrome. Photos bring the colour"',
          '"Feed, profile, upload. Nothing else yet"',
          '"The captions are too small to read"',
        ],
      },
    },
  },
  {
    id: 'context',
    chapter: 'How to ask',
    say: 'Something that trips almost everybody up on day one. It remembers this conversation. You do not have to repeat yourself, and you should not start a fresh chat for every small change. Say make the captions bigger, and it knows which captions you mean, because it is the one that made them. Treat it like a conversation, not like a search box.',
    hold: 1.2,
    scene: {
      k: 'statement',
      kicker: 'It remembers',
      text: 'A conversation, not a search box.',
      accent: ['conversation,'],
      body: 'You never have to describe the whole app again. Say "make the captions bigger" and it knows which captions, because it made them.',
      size: 96,
    },
  },
  {
    id: 'oneatatime',
    chapter: 'How to ask',
    say: 'And change one thing at a time. If you ask for six changes in a single message you will get all six, and if two of them come back wrong you will have no idea which two. One change, look at it, then the next change. It feels slower for about five minutes, and it is much faster after that.',
    hold: 1.2,
    scene: {
      k: 'statement',
      kicker: 'The discipline',
      text: 'One change. Look at it. Then the next one.',
      accent: ['One'],
      body: 'Six changes in one message means six things to untangle when two of them are wrong. Slower for five minutes, faster for the rest of the day.',
      size: 92,
    },
  },
  {
    id: 'chat-vague',
    chapter: 'How to ask',
    say: 'Before I show you the good version, here is the bad one, because the contrast is the lesson. This is what I get when I type the thing most people type. Look at how much of that reply is it asking me questions back, and look at what it built. It is not wrong. It is just nobody in particular, because I told it nothing about who it was for.',
    hold: 2.8,
    scene: {
      k: 'chat',
      title: 'The version that leaves it open',
      steps: [
        {k: 'you', text: 'Make me a social media app. Make it look good.'},
        {k: 'wait'},
        {k: 'say', text: 'I can do that. A few questions first, because social media app covers a lot of ground. Photos, text posts, or video?'},
        {k: 'say', text: 'And who is it for? Otherwise I will default to something like a generic timeline, which is probably not what you have in mind.'},
        {k: 'done', text: 'I will start with a blue and white timeline of text posts and we can go from there.'},
      ],
    },
  },
  {
    id: 'chat-ask',
    chapter: 'How to ask',
    say: 'So let us actually do it. I am going to type that request in, exactly as we built it up, and then I am going to stop talking and let you watch what happens. Notice what I am not asking for. I am not asking for files, or folders, or a framework, or a language. I am describing the thing I want to exist, and I am letting it choose how.',
    hold: 3.9,
    scene: {
      k: 'chat',
      title: 'Lumen — the first request',
      steps: [
        {k: 'you', text: 'Build me a photo sharing app called Lumen. A feed, profiles and comments, for people who post film photos. Keep the chrome calm and black and white so the photos do the talking.'},
        {k: 'wait'},
        {k: 'say', text: 'Good brief. I will start with the shell, the feed and one profile, and leave everything else out until you have looked at it.'},
        {k: 'say', text: 'For the look: paper white background, near black text, hairline borders, and no accent colour anywhere in the chrome. The only saturated thing on the page will be the like button.'},
        {k: 'files', items: ['index.html', 'css/style.css', 'js/data.js', 'js/app.js']},
        {k: 'tool', text: 'Created 4 files'},
        {k: 'tool', text: 'Started a local server on port 8000'},
        {k: 'done', text: 'Lumen is running at localhost:8000. Open it and tell me what is wrong.'},
      ],
    },
  },
  {
    id: 'patience',
    chapter: 'How to ask',
    say: 'That took about forty seconds, and I want to name what just happened, because it is easy to miss. It made a pile of decisions I never specified. It picked the layout, it chose the type sizes, it invented the placeholder people. Every one of those is a decision I am now allowed to overrule, and none of them cost me anything to get wrong.',
    hold: 1.2,
    scene: {
      k: 'statement',
      kicker: 'What just happened',
      text: 'It made a hundred decisions. You can overrule any of them.',
      accent: ['overrule'],
      body: 'A wrong decision costs one sentence to fix. That is the actual change here, and it is why guessing is now cheap.',
      size: 84,
    },
  },

  // ================================================================ chapter 4
  {id: 'ch4', chapter: 'Scope', say: 'Before we look, one word about size.', pad: 0.35,
   scene: {k: 'chapter', n: '04', title: 'Asking for the right size of thing'}},
  {
    id: 'ladder',
    chapter: 'Scope',
    say: 'This is the mistake I see more than any other. People ask for the finished product on the very first try, get something half working and confusing, and conclude that the whole approach does not work. Ask for the bottom rung. Get that working. Then climb. We are starting today with one screen that does one thing, and that is entirely on purpose.',
    hold: 2.9,
    scene: {
      k: 'ladder',
      title: 'Ask for the bottom rung first',
      rungs: [
        {name: 'One screen', note: 'A feed you can scroll. Nothing else works yet.'},
        {name: 'A few screens', note: 'Profile, explore, and a way to post.'},
        {name: 'It remembers', note: 'Your posts are still there tomorrow.'},
        {name: 'Other people', note: 'Accounts, logging in, real users.'},
        {name: 'The world', note: 'A real address, on the internet, at scale.'},
      ],
      pick: 0,
    },
  },
  {
    id: 'tiles',
    chapter: 'Scope',
    say: 'Concretely, here is everything Lumen is going to do by the end of this video. A feed of photos. A profile with a grid on it. An explore page. Somewhere to write comments. A screen to post a new photo with a filter on it. And messages. Six things. Notice there is no login on that list, and no database, and both of those absences are deliberate.',
    hold: 3.3,
    scene: {
      k: 'tiles',
      title: 'Six things, by the end of this video',
      items: [
        {name: 'A feed', note: 'Photos, newest first, one column.', icon: 'feed'},
        {name: 'A profile', note: 'Your posts as a grid, with a count.', icon: 'grid'},
        {name: 'Explore', note: 'Everyone else, in a wall of images.', icon: 'eye'},
        {name: 'Comments', note: 'Write one. It appears under the photo.', icon: 'chat'},
        {name: 'Posting', note: 'Pick a photo, add a filter, share it.', icon: 'camera'},
        {name: 'Messages', note: 'A thread list and a conversation.', icon: 'bell'},
      ],
      sweepAt: ['A feed of photos', 'A profile with a grid', 'An explore page', 'write comments', 'post a new photo', 'And messages'],
    },
  },
  {
    id: 'name',
    chapter: 'Scope',
    say: 'And give the thing a name early. I called this one Lumen, which took me about four seconds and does not matter at all, except that it does. A named thing gets treated like a real project by both of you. An unnamed thing stays an experiment, and experiments are easy to abandon.',
    hold: 1.0,
    scene: {
      k: 'statement',
      kicker: 'Small thing, real effect',
      text: 'Name it. An unnamed thing stays an experiment.',
      accent: ['Name'],
      body: 'It took four seconds and it does not matter, except that a named project gets finished more often than an unnamed one.',
      size: 92,
    },
  },

  // ================================================================ chapter 5
  {id: 'ch5', chapter: 'The feed', say: 'Right. Let us look at what it built.', pad: 0.35,
   scene: {k: 'chapter', n: '05', title: 'Looking at what it built'}},
  {
    id: 'screen-open',
    chapter: 'The feed',
    say: 'This is Lumen, in a browser, on my machine. Before anything else, look at the address bar. Localhost, colon, eight thousand. Localhost is a name that always means this computer, the one you are sitting at. So nobody else can see this yet. It is yours, it is completely private, and nothing you do in here can break anything.',
    hold: 3.5,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'home'},
        start: 1,
        url: 'localhost:8000',
        urlSpot: {at: 40, label: 'this computer only'},
      },
    },
  },
  {
    id: 'screen-parts',
    chapter: 'The feed',
    say: 'Three parts, and you already know all three from every app you have ever used. Navigation down the left hand side. The thing itself in the middle. And a column on the right for everything that is not the main event. I did not ask for that layout anywhere in my request. It is simply what a feed looks like, and it knew that.',
    hold: 3.3,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'home'},
        start: 1,
        spots: [
          {rect: {x: 0, y: 0, w: 244, h: 900}, label: 'get around', at: 24, side: 'right'},
          {rect: {x: FEED_X, y: 26, w: 618, h: 848}, label: 'the feed', at: 52, side: 'top'},
          {rect: {x: RAIL_X, y: 30, w: 300, h: 660}, label: 'everything else', at: 82, side: 'left'},
        ],
      },
    },
  },
  {
    id: 'screen-nav',
    chapter: 'The feed',
    say: 'Close in on the navigation and you can see how much thought went into something I never mentioned. Five items, each with a drawn icon, the current one filled in darker. There is a red dot on messages to say there is something unread. And my own name and picture sit at the very bottom, which is where every app puts them.',
    hold: 2.9,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'home'},
        start: 1,
        push: {to: 2.1, at: [0.09, 0.32]},
        spots: [{rect: {x: 16, y: 223, w: 211, h: 44}, label: 'unread', at: 56, side: 'right'}],
      },
    },
  },
  {
    id: 'screen-rail',
    chapter: 'The feed',
    say: 'The right hand column is worth a look too, because it is doing two jobs. The top card is activity, who liked what and who followed you. The bottom card is suggestions, people you might want to follow, each with a follow button next to them. Neither of those was in my request. Both of them are in every photo app ever made.',
    hold: 3.1,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'home'},
        start: 1,
        push: {to: 1.45, at: [0.82, 0.40]},
        spots: [{rect: {x: RAIL_X, y: 368, w: 300, h: 319}, label: 'a follow button, unasked for', at: 62, side: 'left'}],
      },
    },
  },
  {
    id: 'screen-stories',
    chapter: 'The feed',
    say: 'Now a detail worth stopping on properly. I never mentioned stories, those circles along the top. It added them because that is what photo apps have. This cuts both ways, and you need to hold both halves in your head. Sometimes you get something genuinely useful for free. And sometimes you get something you did not want at all, and it is your job to notice and say so.',
    hold: 3.1,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'home'},
        start: 1,
        push: {to: 1.7, at: [0.47, 0.10]},
        spots: [{rect: {x: FEED_X, y: 30, w: 618, h: 123}, label: 'I never asked for these', at: 48, side: 'bottom'}],
      },
    },
  },
  {
    id: 'screen-scroll',
    chapter: 'The feed',
    say: 'And it scrolls, with real posts in it. These are placeholder people with placeholder captions, which is exactly right at this stage. You want enough content to see whether the design holds up under real material, and you absolutely do not want to be writing captions before you know whether you like the layout they are sitting in.',
    hold: 2.7,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'home'},
        start: 1,
        scroll: [0, 1150],
      },
    },
  },
  {
    id: 'screen-post',
    chapter: 'The feed',
    say: 'Close in on a single post and you can see the decision it made. The photo is the biggest thing on the screen by a very long way. The name is small. The caption is small. The buttons are thin grey outlines that almost disappear. Nothing on this card competes with the picture, and that is the black and white chrome I asked for, doing its job.',
    hold: 3.1,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'home'},
        scroll: 1150,
        start: 1,
        push: {to: 1.5, at: [0.47, 0.48]},
      },
    },
  },
  {
    id: 'screen-like',
    chapter: 'The feed',
    say: 'And here is the one piece of colour in the entire interface. Press the heart and it fills in red, and the count above it goes up by one. One saturated thing on a monochrome page, so the only coloured pixel that is not a photograph is the one that means you liked something. That was its idea, and it is a better idea than the one I gave it.',
    hold: 3.1,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'home', likes: {'2': true}},
        scroll: 1150,
        start: 1,
        push: {to: 1.8, at: [0.47, 0.85]},
        cursor: [{at: 8, to: [394, 1925]}, {at: 36, to: [394, 1925], click: true}],
      },
    },
  },
  {
    id: 'screen-comment',
    chapter: 'The feed',
    say: 'Comments work as well. I type into the box underneath the photo, I press post, and it appears in the list with my own name on it. Nothing is being saved anywhere yet, so this comment lives until I refresh the page and then it is gone. But the mechanism is real, and making it survive a refresh is a later chapter.',
    hold: 3.1,
    scene: {
      k: 'screen',
      spec: {
        state: {
          view: 'home',
          extra: {'2': [['sam.builds', 'the reflections are doing all the work here']]},
        },
        scroll: 1560,
        start: 1,
        push: {to: 1.75, at: [0.47, 0.55]},
      },
    },
  },
  {
    id: 'screen-lightbox',
    chapter: 'The feed',
    say: 'And click the photograph itself and it opens up properly. Big on the left, the whole conversation on the right. I did not ask for this screen anywhere. It is another thing it added because photo apps have one, and this time I am very glad it did, because it is the best looking screen in the app.',
    hold: 3.1,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'home', open: 1},
        start: 1,
      },
    },
  },
  {
    id: 'readit',
    chapter: 'The feed',
    say: 'Now, the habit that separates people who get good at this from people who give up in a week. Do not just look at the screenshot. Use the thing. Click every button on it. Scroll all the way to the bottom. Actively try to break it. Every problem you find now is one sentence of work to fix, and every problem you miss ends up in the finished app.',
    hold: 1.6,
    scene: {
      k: 'statement',
      kicker: 'The habit',
      text: 'Do not review the screenshot. Use the app.',
      accent: ['Use'],
      body: 'Click everything. Scroll to the end. Try to break it. Every problem you find is one sentence of work. Every one you miss is in the finished thing.',
      size: 90,
    },
  },

  // ================================================================ chapter 6
  {id: 'ch6', chapter: 'Checkpoint', say: 'Quick check before we carry on.', pad: 0.35,
   scene: {k: 'chapter', n: '06', title: 'Two questions'}},
  {
    id: 'quiz1-q',
    chapter: 'Checkpoint',
    say: 'First one. The address bar said localhost, colon, eight thousand. What does that tell you about who can see your app right now? Have a think before I answer it.',
    hold: 4.1,
    scene: {
      k: 'quiz',
      n: '01',
      question: 'The address said localhost. Who can see your app?',
      options: [
        'Anyone with the link',
        'Only this computer',
        'Everyone once you save it',
        'Nobody until you pay',
      ],
      answer: 1,
      revealAt: 100000,
    },
  },
  {
    id: 'quiz1-a',
    chapter: 'Checkpoint',
    say: 'Only this computer. Localhost is a name that always means the machine you are sitting at, and it never means anywhere else. Putting your app somewhere other people can actually reach it is a completely separate job, and it is a later lesson.',
    hold: 1.4,
    scene: {
      k: 'quiz',
      n: '01',
      question: 'The address said localhost. Who can see your app?',
      options: [
        'Anyone with the link',
        'Only this computer',
        'Everyone once you save it',
        'Nobody until you pay',
      ],
      answer: 1,
      revealAt: 0,
    },
  },
  {
    id: 'quiz2-q',
    chapter: 'Checkpoint',
    say: 'Second one, and this is the one that actually matters for the rest of your life. Which of these four requests is going to get you something you like?',
    hold: 4.3,
    scene: {
      k: 'quiz',
      n: '02',
      question: 'Which request gets you something you like?',
      options: [
        '"Make the design better"',
        '"Add all the normal features"',
        '"The captions are too small to read"',
        '"Make it more modern"',
      ],
      answer: 2,
      revealAt: 100000,
    },
  },
  {
    id: 'quiz2-a',
    chapter: 'Checkpoint',
    say: 'The captions are too small to read. It is the only one of the four that names a specific thing and says what is wrong with it. Better, and normal, and modern are all opinions, and it does not have access to yours. Point at one thing, and describe the problem with it.',
    hold: 1.6,
    scene: {
      k: 'quiz',
      n: '02',
      question: 'Which request gets you something you like?',
      options: [
        '"Make the design better"',
        '"Add all the normal features"',
        '"The captions are too small to read"',
        '"Make it more modern"',
      ],
      answer: 2,
      revealAt: 0,
    },
  },

  // ================================================================ chapter 7
  {id: 'ch7', chapter: 'Making it yours', say: 'Now we make it ours.', pad: 0.35,
   scene: {k: 'chapter', n: '07', title: 'Round two, and round three'}},
  {
    id: 'chat-refine',
    chapter: 'Making yours',
    say: 'Let us start with the smallest possible change, so you can see the loop turn once quickly. One sentence, one fix. And notice that I did not describe the app again, and I did not say which file to edit. I said the captions are too small, and it knew exactly what I meant, because we are in the middle of a conversation about them.',
    hold: 3.5,
    scene: {
      k: 'chat',
      title: 'Lumen — one small change',
      steps: [
        {k: 'you', text: 'The captions under the photos are too small to read comfortably.'},
        {k: 'wait'},
        {k: 'say', text: 'Taking the caption and the comments up two steps, and the like count with them so the block still reads as a group.'},
        {k: 'tool', text: 'Edited css/style.css'},
        {k: 'done', text: 'Reload it. If it is still tight I can open up the line spacing too.'},
      ],
    },
  },
  {
    id: 'chat-explore',
    chapter: 'Making yours',
    say: 'Round two, and now I want the other screens. Watch how I ask for this one. I am not saying add an explore page. I am saying what should be on it and how it ought to behave, because add an explore page would get me somebody else\'s idea of an explore page, and I have my own.',
    hold: 3.9,
    scene: {
      k: 'chat',
      title: 'Lumen — round two',
      steps: [
        {k: 'you', text: 'Add an Explore page. A tight grid of everyone\'s photos, three across, with almost no gap, so it reads as one wall of images rather than a list. Filter chips along the top. Hovering a photo shows its likes and comments.'},
        {k: 'wait'},
        {k: 'say', text: 'Three across with a five pixel gap, and I will run two of the cells at double height so the wall does not read as a spreadsheet.'},
        {k: 'tool', text: 'Edited css/style.css and js/app.js'},
        {k: 'done', text: 'Explore is in the sidebar. The chips do not filter anything yet, they only change which one is selected. Say the word and I will wire them up.'},
      ],
    },
  },
  {
    id: 'screen-explore',
    chapter: 'Making yours',
    say: 'And there it is. Look at how little space there is between those photographs. That was the entire point of saying one wall of images instead of saying a grid. And two of the cells are twice as tall as the others, which stops the whole thing reading like a spreadsheet. I did not ask for that. It made a judgement call, and the judgement call was right.',
    hold: 3.3,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'explore'},
        start: 1,
        scroll: [0, 300],
        spots: [{rect: {x: 688, y: 460, w: 308, h: 619}, label: 'twice as tall, its idea', at: 70, side: 'left'}],
      },
    },
  },
  {
    id: 'screen-chips',
    chapter: 'Making yours',
    say: 'And here is something honest that most videos would cut. Those filter chips along the top do not filter anything at all yet. It told me so when it built them, in the last line of its reply, and I left them exactly as they were. Half finished is the normal state of a project, and knowing which half is which is the genuinely useful skill.',
    hold: 3.3,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'explore', chip: 'Portraits'},
        start: 1,
        push: {to: 1.8, at: [0.50, 0.12]},
        spots: [{rect: {x: WIDE_X, y: 91, w: 545, h: 38}, label: 'these do nothing yet', at: 56, side: 'bottom'}],
      },
    },
  },
  {
    id: 'chat-more',
    chapter: 'Making yours',
    say: 'Round three, and I am going to ask for three things at once now, which contradicts what I told you earlier about one change at a time. The reason is that the first two rounds both went well, so I trust it with more. That is a judgement you earn during a session. Start strict, and loosen up once it has proved itself.',
    hold: 3.9,
    scene: {
      k: 'chat',
      title: 'Lumen — round three',
      steps: [
        {k: 'you', text: 'Three more things. A profile page with my photos as a grid and a follower count. A page to post a photo where I pick a filter and write a caption, and pressing share puts it at the top of my feed. And messages, with a thread list and a conversation.'},
        {k: 'wait'},
        {k: 'say', text: 'For the filters I will show each look as a live thumbnail of your actual photo, rather than as a name, so you are choosing from the real thing.'},
        {k: 'files', items: ['js/app.js', 'css/style.css']},
        {k: 'tool', text: 'Edited 2 files, added 3 views'},
        {k: 'done', text: 'All three are in. Posting is not saved anywhere yet, so a refresh clears it. That is the next rung.'},
      ],
    },
  },
  {
    id: 'screen-profile',
    chapter: 'Making yours',
    say: 'The profile. A big picture, a name, three counts, and then the grid. If that arrangement feels extremely familiar, it is because almost every profile page on the internet is arranged exactly this way. And there is no prize for being different here. Save your originality for the parts of your app that people actually care about.',
    hold: 3.1,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'profile'},
        start: 1,
        push: {to: 1.32, at: [0.5, 0.2]},
      },
    },
  },
  {
    id: 'screen-create',
    chapter: 'Making yours',
    say: 'This is the screen I like the most, and it is the one where the filters live. Five different looks, and each little square is a preview of my actual photograph rather than a word. You can see the difference between them at a glance, without clicking anything. That was entirely its idea, not mine, and it is better than the thing I asked for.',
    hold: 3.5,
    scene: {
      k: 'screen',
      spec: {
        state: {
          view: 'create',
          draft: {img: 6, text: 'Expired roll from 2019. Shot it anyway.', place: 'Vila Madalena', filter: 'Kodachrome'},
        },
        start: 1,
        push: {to: 1.55, at: [0.78, 0.46]},
        spots: [{rect: {x: 925, y: 373, w: 384, h: 86}, at: 66}],
      },
    },
  },
  {
    id: 'screen-shared',
    chapter: 'Making yours',
    say: 'Press share, and it lands at the top of my feed with my caption on it. That is the loop closing for the first time. I described a thing, it built the thing, and now I am using the thing. Ten minutes ago this application did not exist anywhere in the world.',
    hold: 3.5,
    scene: {
      k: 'screen',
      spec: {
        state: {
          view: 'home',
          posted: [6],
          draft: {img: 6, text: 'Expired roll from 2019. Shot it anyway.', place: 'Vila Madalena', filter: 'Kodachrome'},
        },
        start: 1,
        push: {to: 1.4, at: [0.47, 0.24]},
        spots: [{rect: {x: FEED_X, y: 179, w: 618, h: 200}, label: 'mine, thirty seconds old', at: 62, side: 'bottom'}],
      },
    },
  },
  {
    id: 'screen-messages',
    chapter: 'Making yours',
    say: 'And messages. A list of conversations on the left, the thread itself on the right, my messages in black and theirs in grey. This is the screen where you can most clearly watch it borrowing a convention, and borrowing conventions is completely correct. Nobody has ever wanted a chat app with a creative layout.',
    hold: 3.1,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'messages', thread: 'atlas.reads'},
        start: 1,
        push: {to: 1.25, at: [0.58, 0.48]},
      },
    },
  },
  {
    id: 'screen-follow',
    chapter: 'Making yours',
    say: 'And the small interactions all work. Press follow on somebody in that suggestions list and the button changes to following, in grey, because it is no longer the thing you want to press. Nobody specified that. It is just what a follow button does, and it knew, and that is the part that still surprises me.',
    hold: 3.1,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'home', follows: {'noor.builds': true}},
        start: 1,
        push: {to: 2.1, at: [0.88, 0.47]},
        cursor: [{at: 10, to: [1282, 441]}, {at: 40, to: [1282, 441], click: true}],
      },
    },
  },
  {
    id: 'taste',
    chapter: 'Making yours',
    say: 'Here is the part that nobody tells you. Every judgement call in those three rounds, the tight grid, the live filter previews, the single red heart, came from somebody having an opinion about photo apps. It cannot want things on your behalf. Taste is the half of this job that does not get automated, and it is the half worth practising.',
    hold: 1.6,
    scene: {
      k: 'statement',
      kicker: 'The part that stays yours',
      text: 'It can build anything. It cannot want anything.',
      accent: ['want'],
      body: 'Every good decision in this app started as somebody having an opinion. That is the half of the work that does not get automated.',
      size: 92,
    },
  },

  // ================================================================ chapter 8
  {id: 'ch8', chapter: 'When it breaks', say: 'Now the bit the tutorials skip.', pad: 0.35,
   scene: {k: 'chapter', n: '08', title: 'When it breaks, and it will'}},
  {
    id: 'breaks',
    chapter: 'When it breaks',
    say: 'Everything I have shown you so far worked. That is not what a real session looks like, and I would be doing you a disservice if I stopped here. Things come back wrong, or blank, or half finished. The difference between people who get somewhere with this and people who quit is entirely in what they do in the next thirty seconds.',
    hold: 1.4,
    scene: {
      k: 'statement',
      kicker: 'Expect this',
      text: 'It will come back wrong. That is the normal case, not the failure case.',
      accent: ['normal'],
      serif: ['wrong.'],
      size: 82,
    },
  },
  {
    id: 'chat-bug',
    chapter: 'When it breaks',
    say: 'This genuinely happened while I was making this video. The whole middle of the page came back empty. Just the sidebar, and nothing else at all. Watch how I report it. Where I was, what I expected to see, and what I got instead. I am not guessing at the cause, because diagnosing it is not my job.',
    hold: 3.9,
    scene: {
      k: 'chat',
      title: 'Lumen — the page is blank',
      steps: [
        {k: 'you', text: 'The sidebar renders but the whole middle of the page is empty. I expected the feed. Odd detail: if I scroll down a long way the feed is actually there, far below the fold.'},
        {k: 'wait'},
        {k: 'say', text: 'That last detail is the answer. The feed is not missing, it is being pushed below the sidebar, so the two are stacking vertically instead of sitting side by side.'},
        {k: 'say', text: 'The container that should lay them out in a row never got the class that does it. It is one word, in the HTML.'},
        {k: 'tool', text: 'Edited index.html, 1 line'},
        {k: 'done', text: 'Fixed. Reload it.'},
      ],
    },
  },
  {
    id: 'screen-fixed',
    chapter: 'When it breaks',
    say: 'One line, and there it is. Now I want you to notice what I did not do. I did not read any code. I did not know what a container was, and I still do not really. I never used the word bug. I said where I was, what I expected, and what I got. Three facts. And the odd detail about scrolling is the thing that actually solved it.',
    hold: 3.3,
    scene: {
      k: 'screen',
      spec: {
        state: {view: 'home'},
        start: 1,
        push: {to: 1.25, at: [0.42, 0.35]},
      },
    },
  },
  {
    id: 'compare-bug',
    chapter: 'When it breaks',
    say: 'So here is the shape of a report that gets fixed on the first attempt, sitting next to the shape of one that turns into four rounds of guessing. The left hand column is a feeling. The right hand column is evidence. And the single most valuable thing you can add is the strange little detail you almost did not bother mentioning.',
    hold: 3.1,
    scene: {
      k: 'compare',
      title: 'How to say something is broken',
      bad: {
        label: 'a feeling',
        items: [
          '"It is broken"',
          '"That did not work"',
          '"Still wrong"',
          '"I think the database is bad"',
        ],
      },
      good: {
        label: 'evidence',
        items: [
          '"On the profile page, the grid is empty"',
          '"I expected nine photos. I see none"',
          '"It worked before I asked for messages"',
          '"Odd detail: it appears if I scroll far down"',
        ],
      },
    },
  },
  {
    id: 'nonsense',
    chapter: 'When it breaks',
    say: 'One more failure mode worth naming out loud, because it is the one that catches beginners hardest. Sometimes it will confidently describe something it did not actually do, or refer to a button that is not on the screen. This is the thing you genuinely have to watch for. And the defence is exactly the habit from chapter five. Open it. Click the button. If the button is not there, say so plainly and it will fix it.',
    hold: 1.6,
    scene: {
      k: 'statement',
      kicker: 'Watch for this',
      text: 'It can describe a button that is not there.',
      accent: ['not'],
      body: 'Confidently, and in detail. The defence is the same habit: open it, click the thing, and say plainly when the thing is missing.',
      size: 88,
    },
  },
  {
    id: 'stuck',
    chapter: 'When it breaks',
    say: 'And if two rounds go by and it is still wrong, stop describing and start over. Ask for that one screen on its own, in a completely fresh conversation. Going round the same loop a third time almost never works, because the conversation is now full of the wrong idea. Starting again from a clean brief almost always does.',
    hold: 1.4,
    scene: {
      k: 'statement',
      kicker: 'If two rounds fail',
      text: 'Stop. Start over with a smaller ask.',
      accent: ['smaller'],
      body: 'A third round on the same broken thing rarely works, because the conversation is full of the wrong idea. One screen, fresh chat, almost always does.',
      size: 92,
    },
  },

  // ================================================================ chapter 9
  {id: 'ch9', chapter: 'Next', say: 'Where this goes.', pad: 0.35,
   scene: {k: 'chapter', n: '09', title: 'Where this goes next'}},
  {
    id: 'timeline',
    chapter: 'Next',
    say: 'Here is the whole road, and where you are standing on it after today. You have an app running on your own machine that you built by describing it out loud. The next thing is making it remember, so your posts survive a refresh. Then accounts, so it is not only you. Then a real address, so other people can reach it. Each of those is one lesson, and each one is the same loop you already know.',
    hold: 3.9,
    scene: {
      k: 'timeline',
      title: 'Where you are',
      nodes: [
        {name: 'Describing what you want', state: 'done'},
        {name: 'An app that runs', state: 'done'},
        {name: 'Six screens, working', state: 'here'},
        {name: 'It remembers your posts', state: 'next'},
        {name: 'Accounts and logging in', state: 'next'},
        {name: 'A real address', state: 'next'},
      ],
    },
  },
  {
    id: 'real',
    chapter: 'Next',
    say: 'One honest caveat before you go off and show somebody. Refresh Lumen and the photo you posted disappears, because nothing is being written down anywhere yet. That is not a bug, and it is not a limitation of this whole approach. It is simply the next rung on the ladder, and it is genuinely the next lesson.',
    hold: 1.6,
    scene: {
      k: 'statement',
      kicker: 'The catch',
      text: 'Refresh it and your post is gone. Nothing is saved yet.',
      accent: ['Nothing'],
      body: 'Not a bug. The next rung. Making it remember is the whole of the next lesson.',
      size: 84,
    },
  },
  {
    id: 'cost',
    chapter: 'Next',
    say: 'A practical note on money, since nobody ever mentions it. A session about the size of this one sits comfortably inside a free account. Long sessions on a large app will not, and at that point there is a paid tier. I am not going to tell you which to choose. I am telling you that you can do every single thing in this lesson without paying anything.',
    hold: 1.6,
    scene: {
      k: 'statement',
      kicker: 'On money',
      text: 'Everything in this lesson fits in a free account.',
      accent: ['free'],
      body: 'Bigger apps and longer sessions eventually need the paid tier. Nothing you have watched today did.',
      size: 88,
    },
  },
  {
    id: 'share',
    chapter: 'Next',
    say: 'And one last thing before the homework. Show it to somebody. Not because it is finished, because it is very obviously not finished. Show it to them because the moment you watch another person scroll a thing that you described into existence is the moment this stops feeling like a trick and starts feeling like something you can do.',
    hold: 1.6,
    scene: {
      k: 'statement',
      kicker: 'Do this today',
      text: 'Show it to somebody. It stops feeling like a trick.',
      accent: ['somebody.'],
      body: 'Not because it is finished. Because watching another person use a thing you described changes how possible it feels.',
      size: 88,
    },
  },
  {
    id: 'beyond',
    chapter: 'Next',
    say: 'And one last thought, because this is bigger than photo apps. The loop you learned today is not really about social networks. It works on a tool that renames three hundred files, or a page that tracks something you care about, or a small game for one specific child. Anything you can describe, you can now ask for. Photo apps were just a good excuse.',
    hold: 3.5,
    scene: {
      k: 'tiles',
      title: 'The loop is not about photo apps',
      items: [
        {name: 'A tool', note: 'Rename three hundred files. Once. Correctly.', icon: 'stack'},
        {name: 'A tracker', note: 'One page, one number you care about.', icon: 'db'},
        {name: 'A small game', note: 'For one specific child, with their name in it.', icon: 'sparkle'},
      ],
      sweepAt: ['a tool that renames', 'a page that tracks', 'a small game'],
    },
  },
  {
    id: 'homework',
    chapter: 'Next',
    say: 'So, before the next lesson, do this. Open Claude and ask for one screen of something you actually want. Not an app, one screen. Then open it and say the single thing that is most wrong with it. That is the entire skill, start to finish, and you can practise it in about ten minutes.',
    hold: 3.1,
    scene: {
      k: 'checklist',
      title: 'Ten minutes of homework',
      items: [
        {name: 'Ask for one screen', note: 'Of something you want. Not an app. One screen.'},
        {name: 'Open it', note: 'In a browser. Click everything on it.'},
        {name: 'Say one thing', note: 'The single most wrong thing. Then go again.'},
      ],
      sweepAt: ['ask for one screen', 'Then open it', 'say the single thing'],
    },
  },
  {
    id: 'outro',
    chapter: 'Next',
    say: 'You built something today. Describe it, look at it, change it. That is the whole trick, and it works on very nearly anything. See you in the next one.',
    hold: 3.1,
    scene: {
      k: 'outro',
      lines: ['You built something today.', 'Describe it. Look at it. Change it.'],
      sign: 'No-code with Claude · Lesson two next',
    },
  },
];
