// Seeded content for Lumen.
//
// Everything the app shows comes from here, so a screenshot taken twice looks
// the same twice. No dates computed at load time, no shuffling, no random ids —
// a feed that reorders itself between two captures is unusable as footage.

const PEOPLE = [
  {h: 'maren.kova',   name: 'Maren Kovach',    av: 1,  bio: 'Light, water, and the odd rooftop.',           city: 'Lisbon',      follows: true  },
  {h: 'atlas.reads',  name: 'Atlas Ferreira',  av: 2,  bio: 'Film only. Mostly out of focus.',              city: 'São Paulo',   follows: true  },
  {h: 'noor.builds',  name: 'Noor Haddad',     av: 3,  bio: 'Concrete, glass, and long shadows.',           city: 'Amman',       follows: false },
  {h: 'juniper',      name: 'Juniper Wallace', av: 4,  bio: 'Plants I have not killed yet.',                city: 'Portland',    follows: true  },
  {h: 'kiro.wav',     name: 'Kiro Nakamura',   av: 5,  bio: 'Night walks. 35mm.',                           city: 'Osaka',       follows: false },
  {h: 'elsa.grn',     name: 'Elsa Grinde',     av: 6,  bio: 'Cold water swimmer. Warm coffee drinker.',     city: 'Bergen',      follows: true  },
  {h: 'tomas.p',      name: 'Tomás Prieto',    av: 7,  bio: 'Markets, mornings, motion blur.',              city: 'Mexico City', follows: false },
  {h: 'ada.frames',   name: 'Ada Okonkwo',     av: 8,  bio: 'Portraits of people mid-sentence.',            city: 'Lagos',       follows: true  },
  {h: 'wren',         name: 'Wren Adeyemi',    av: 9,  bio: 'Birds, buses, bad weather.',                   city: 'Bristol',     follows: false },
  {h: 'silje.o',      name: 'Silje Ødegård',   av: 10, bio: 'Two lenses, one bag.',                         city: 'Oslo',        follows: true  },
  {h: 'rafi.dev',     name: 'Rafi Chowdhury',  av: 11, bio: 'Rooftops at golden hour.',                     city: 'Dhaka',       follows: false },
  {h: 'mira.stills',  name: 'Mira Lindqvist',  av: 12, bio: 'Still life, moving city.',                     city: 'Helsinki',    follows: true  },
  {h: 'okan',         name: 'Okan Demir',      av: 13, bio: 'Ferries and fog.',                             city: 'Istanbul',    follows: false },
  {h: 'lune.mp4',     name: 'Lune Bertrand',   av: 14, bio: 'Shoots at 4am. Sleeps at 4pm.',                city: 'Marseille',   follows: false },
  {h: 'devi.k',       name: 'Devi Krishnan',   av: 15, bio: 'Colour first, subject second.',                city: 'Kochi',       follows: true  },
];

const ME = {h: 'sam.builds', name: 'Sam Ellery', av: 16, bio: 'Learning to build things.', city: 'Remote'};

const POSTS = [
  {id: 1,  by: 'maren.kova',  img: 1,  where: 'Alfama, Lisbon',       likes: 1284, when: '2h',
   text: 'Six flights up for this. Worth every one of them.', tags: ['rooftops', 'goldenhour'],
   comments: [['atlas.reads', 'the light on that wall'], ['elsa.grn', 'ok this is the one'], ['juniper', 'six flights is nothing, you are fine']]},
  {id: 2,  by: 'kiro.wav',    img: 2,  where: 'Dotonbori, Osaka',     likes: 3902, when: '4h',
   text: 'Rain makes better neon than neon does.', tags: ['35mm', 'night'],
   comments: [['lune.mp4', 'reflections are doing a lot of work here'], ['okan', 'what stock is this?']]},
  {id: 3,  by: 'juniper',     img: 3,  where: 'Portland, OR',         likes: 742,  when: '7h',
   text: 'Repotted everything. Two survivors, one casualty.', tags: ['plants'],
   comments: [['mira.stills', 'RIP to the third one'], ['maren.kova', 'the pot on the left though']]},
  {id: 4,  by: 'ada.frames',  img: 4,  where: 'Yaba, Lagos',          likes: 5610, when: '11h',
   text: 'He was telling me about his brother when I pressed the shutter. That is the whole photo.',
   tags: ['portrait'],
   comments: [['atlas.reads', 'mid-sentence is the only way'], ['silje.o', 'incredible'], ['devi.k', 'the hands!']]},
  {id: 5,  by: 'elsa.grn',    img: 5,  where: 'Bergen, Norway',       likes: 2211, when: '14h',
   text: 'Four degrees. In and out in ninety seconds.', tags: ['coldwater'],
   comments: [['wren', 'absolutely not'], ['juniper', 'ninety seconds too long']]},
  {id: 6,  by: 'atlas.reads', img: 6,  where: 'Vila Madalena',        likes: 1893, when: '18h',
   text: 'Expired roll from 2019. Shot it anyway.', tags: ['film', 'expired'],
   comments: [['kiro.wav', 'the colour shift is free grading'], ['mira.stills', 'love a gamble']]},
  {id: 7,  by: 'silje.o',     img: 7,  where: 'Grünerløkka, Oslo',    likes: 968,  when: '21h',
   text: 'Packed two lenses. Used one.', tags: ['streets'],
   comments: [['tomas.p', 'always the way']]},
  {id: 8,  by: 'devi.k',      img: 8,  where: 'Fort Kochi',           likes: 4127, when: '1d',
   text: 'Found the wall first. Waited for someone to walk past it.', tags: ['colour', 'patience'],
   comments: [['ada.frames', 'the wait is the craft'], ['noor.builds', 'that blue']]},
  {id: 9,  by: 'mira.stills', img: 9,  where: 'Kallio, Helsinki',     likes: 1355, when: '1d',
   text: 'Breakfast, arranged slightly too carefully.', tags: ['stilllife'],
   comments: [['elsa.grn', 'slightly?'], ['juniper', 'the napkin fold is sending me']]},
  {id: 10, by: 'tomas.p',     img: 10, where: 'Mercado de Jamaica',   likes: 2740, when: '1d',
   text: 'Flowers before six in the morning are a completely different colour.', tags: ['markets'],
   comments: [['devi.k', 'that is real'], ['rafi.dev', 'need to get up earlier']]},
  {id: 11, by: 'noor.builds', img: 11, where: 'Amman, Jordan',        likes: 1602, when: '2d',
   text: 'One building, four hours of shadow.', tags: ['architecture'],
   comments: [['silje.o', 'the diagonal is perfect']]},
  {id: 12, by: 'rafi.dev',    img: 12, where: 'Gulshan, Dhaka',       likes: 3318, when: '2d',
   text: 'The whole city goes warm for about eleven minutes.', tags: ['goldenhour', 'rooftops'],
   comments: [['maren.kova', 'eleven minutes is generous'], ['okan', 'chasing this next week']]},
];

// The grid on Explore uses everything, including the posts the feed does not show.
const EXPLORE = [13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 1, 4, 8, 12, 2, 10];

const MY_POSTS = [20, 15, 22, 17, 13, 24, 19, 14, 21];

const THREADS = [
  {h: 'juniper',     last: 'sent you a photo',                 when: '12m', unread: true },
  {h: 'atlas.reads', last: 'ok but did you keep the negative?', when: '1h',  unread: true },
  {h: 'elsa.grn',    last: 'Saturday, 6am, no excuses',         when: '3h',  unread: false},
  {h: 'maren.kova',  last: 'You: on my way up now',             when: '1d',  unread: false},
  {h: 'devi.k',      last: 'Reacted 🔥 to your post',            when: '2d',  unread: false},
];

const MESSAGES = [
  {me: false, t: 'did you ever get that rooftop shot developed'},
  {me: true,  t: 'came back yesterday. one frame out of twelve is usable'},
  {me: false, t: 'thats a good ratio for expired film honestly'},
  {me: true,  t: 'posting the good one tonight'},
  {me: false, t: 'ok but did you keep the negative?'},
];

const NOTIFS = [
  {h: 'ada.frames',  t: 'liked your photo',            when: '4m'},
  {h: 'juniper',     t: 'started following you',        when: '22m'},
  {h: 'kiro.wav',    t: 'commented: "the grain here"',  when: '1h'},
  {h: 'elsa.grn',    t: 'liked your photo',            when: '3h'},
  {h: 'silje.o',     t: 'mentioned you in a comment',   when: '5h'},
];

const person = (h) => (h === ME.h ? ME : PEOPLE.find((p) => p.h === h));
const avatar = (h) => `img/avatars/a${person(h).av}.jpg`;
const photo  = (n) => `img/posts/p${n}.jpg`;
