// Lumen — one page, five views, no build step.
//
// Views render into #view from the seeded data in data.js. State that a viewer
// can change (likes, follows, new comments, a posted photo) lives in `S` and is
// re-rendered from scratch, which is fine at this size and keeps every view a
// pure function of state.

const I = {
  home:    'M3 10.6 12 3.5l9 7.1V20a1 1 0 0 1-1 1h-5v-6.4H9V21H4a1 1 0 0 1-1-1z',
  compass: null, plus: null, chat: null, heart: null,
};

/** Line icons at a single 1.7px stroke so a row of them reads as one set. */
function icon(name, size = 24) {
  const s = (d, extra = '') =>
    `<svg width="${size}" height="${size}" viewBox="0 0 24 24" fill="none"
      stroke="currentColor" stroke-width="1.7" stroke-linecap="round"
      stroke-linejoin="round">${d}${extra}</svg>`;
  switch (name) {
    case 'home':    return s('<path d="M3.2 10.9 12 4l8.8 6.9V20a1 1 0 0 1-1 1h-4.6v-6.3H8.8V21H4.2a1 1 0 0 1-1-1z"/>');
    case 'search':  return s('<circle cx="11" cy="11" r="7"/><path d="m16.3 16.3 4 4"/>');
    case 'compass': return s('<circle cx="12" cy="12" r="9"/><path d="m15.6 8.4-2.1 5.1-5.1 2.1 2.1-5.1z"/>');
    case 'plus':    return s('<rect x="3.5" y="3.5" width="17" height="17" rx="4.5"/><path d="M12 8.4v7.2M8.4 12h7.2"/>');
    case 'chat':    return s('<path d="M20.5 11.6c0 4.1-3.8 7.4-8.5 7.4a9.7 9.7 0 0 1-2.6-.35L5 20.5l1.2-3.3A7.1 7.1 0 0 1 3.5 11.6C3.5 7.5 7.3 4.2 12 4.2s8.5 3.3 8.5 7.4z"/>');
    case 'heart':   return s('<path d="M12 20.2s-7.6-4.6-7.6-9.4a4.3 4.3 0 0 1 7.6-2.7 4.3 4.3 0 0 1 7.6 2.7c0 4.8-7.6 9.4-7.6 9.4z"/>');
    case 'heartf':  return `<svg width="${size}" height="${size}" viewBox="0 0 24 24" fill="currentColor"><path d="M12 20.6S3.9 15.7 3.9 10.6A4.7 4.7 0 0 1 12 7.5a4.7 4.7 0 0 1 8.1 3.1c0 5.1-8.1 10-8.1 10z"/></svg>`;
    case 'cmt':     return s('<path d="M20.5 11.8c0 4-3.8 7.2-8.5 7.2a9.9 9.9 0 0 1-2.6-.34L5 20.4l1.2-3.2A7 7 0 0 1 3.5 11.8C3.5 7.8 7.3 4.6 12 4.6s8.5 3.2 8.5 7.2z"/>');
    case 'share':   return s('<path d="M21 3.4 10.4 13.9M21 3.4l-6.8 17.3-3.8-7.8L2.7 9.2z"/>');
    case 'save':    return s('<path d="M6.5 3.6h11a1 1 0 0 1 1 1v15.6L12 16l-6.5 4.2V4.6a1 1 0 0 1 1-1z"/>');
    case 'savef':   return `<svg width="${size}" height="${size}" viewBox="0 0 24 24" fill="currentColor"><path d="M6.5 3.6h11a1 1 0 0 1 1 1v15.6L12 16l-6.5 4.2V4.6a1 1 0 0 1 1-1z"/></svg>`;
    case 'dots':    return s('<circle cx="5.5" cy="12" r="1.3" fill="currentColor" stroke="none"/><circle cx="12" cy="12" r="1.3" fill="currentColor" stroke="none"/><circle cx="18.5" cy="12" r="1.3" fill="currentColor" stroke="none"/>');
    case 'user':    return s('<circle cx="12" cy="8.6" r="3.9"/><path d="M4.6 20.4a7.6 7.6 0 0 1 14.8 0"/>');
    case 'grid':    return s('<rect x="3.6" y="3.6" width="6.2" height="6.2" rx="1.2"/><rect x="14.2" y="3.6" width="6.2" height="6.2" rx="1.2"/><rect x="3.6" y="14.2" width="6.2" height="6.2" rx="1.2"/><rect x="14.2" y="14.2" width="6.2" height="6.2" rx="1.2"/>');
    case 'tag':     return s('<path d="M20.4 13.1 13.1 20.4a1.6 1.6 0 0 1-2.3 0l-7.2-7.2V3.6h9.6l7.2 7.2a1.6 1.6 0 0 1 0 2.3z"/><circle cx="7.9" cy="7.9" r="1.4"/>');
    case 'img':     return s('<rect x="3.4" y="4.6" width="17.2" height="14.8" rx="2.6"/><circle cx="8.8" cy="10" r="1.6"/><path d="m3.9 17.4 4.6-4.3 4.1 3.6 3.1-2.7 4.8 4"/>', '');
    case 'mark':    return `<svg width="${size}" height="${size}" viewBox="0 0 28 28" fill="none"><circle cx="14" cy="14" r="12.2" stroke="#0D0F12" stroke-width="2.2"/><circle cx="14" cy="14" r="5.1" fill="#0D0F12"/><path d="M14 1.8v6.9M14 19.3v6.9M1.8 14h6.9M19.3 14h6.9" stroke="#0D0F12" stroke-width="2.2" stroke-linecap="round"/></svg>`;
    default: return '';
  }
}

const S = {
  view: 'home',
  likes: {},          // id -> bool
  saved: {},
  follows: {},        // handle -> bool
  extra: {},          // id -> [[handle, text]]
  open: null,         // post id shown in the lightbox
  tab: 'grid',
  chip: 'For you',
  thread: 'juniper',
  draft: null,        // {img, text, place, filter}
  posted: [],         // photo numbers posted this session, newest first
  toast: null,
};

const nfmt = (n) => n.toLocaleString('en-US');
const liked = (id) => (S.likes[id] ?? false);
const likeCount = (p) => p.likes + (liked(p.id) ? 1 : 0);
const following = (h) => (S.follows[h] ?? person(h).follows);
const el = (html) => { const d = document.createElement('div'); d.innerHTML = html.trim(); return d.firstElementChild; };

// ------------------------------------------------------------------ chrome

function sidebar() {
  const items = [
    ['home', 'Home', 'home'], ['explore', 'Explore', 'compass'],
    ['create', 'Create', 'plus'], ['messages', 'Messages', 'chat'],
    ['profile', 'Profile', 'user'],
  ];
  return `
  <aside class="side">
    <div class="brand">${icon('mark', 28)}<b>Lumen</b></div>
    <nav class="nav">
      ${items.map(([k, label, ic]) => `
        <button data-go="${k}" class="${S.view === k ? 'on' : ''}">
          ${icon(ic)}<span>${label}</span>
          ${k === 'messages' ? '<i class="dot"></i>' : ''}
        </button>`).join('')}
    </nav>
    <div class="spacer"></div>
    <button class="me-row">
      <img src="${avatar(ME.h)}" alt="">
      <span class="who"><b>${ME.name}</b><span>@${ME.h}</span></span>
    </button>
  </aside>`;
}

// ------------------------------------------------------------------ home

function storiesRow() {
  const seen = new Set(['tomas.p', 'wren', 'okan']);
  const who = ['maren.kova', 'kiro.wav', 'juniper', 'ada.frames', 'elsa.grn', 'atlas.reads'];
  return `<div class="stories">
    <div class="story mine">
      <div class="ring"><img src="${avatar(ME.h)}" alt=""><span class="plus">+</span></div>
      <span>Your story</span>
    </div>
    ${who.map((h) => `
      <div class="story ${seen.has(h) ? 'seen' : ''}">
        <div class="ring"><img src="${avatar(h)}" alt=""></div>
        <span>${h}</span>
      </div>`).join('')}
  </div>`;
}

function postCard(p, fresh) {
  const u = person(p.by);
  const cs = [...p.comments, ...(S.extra[p.id] ?? [])];
  const shown = cs.slice(-3);
  return `
  <article class="post ${fresh ? 'fresh' : ''}" data-post="${p.id}">
    <div class="post-top">
      <img class="av" src="${avatar(p.by)}" alt="">
      <span class="who"><b>${p.by}</b><span>${p.where}</span></span>
      <button class="more">${icon('dots', 20)}</button>
    </div>
    <div class="frame" data-open="${p.id}"><img src="${photo(p.img)}" alt=""></div>
    <div class="acts">
      <button data-like="${p.id}" class="${liked(p.id) ? 'liked' : ''}">${icon(liked(p.id) ? 'heartf' : 'heart')}</button>
      <button data-open="${p.id}">${icon('cmt')}</button>
      <button>${icon('share')}</button>
      <button class="save" data-save="${p.id}">${icon(S.saved[p.id] ? 'savef' : 'save')}</button>
    </div>
    <div class="post-body">
      <div class="likes">${nfmt(likeCount(p))} likes</div>
      <div class="cap"><b>${p.by}</b>${p.text}</div>
      ${p.tags?.length ? `<div class="tags">${p.tags.map((t) => `<span>#${t}</span>`).join('')}</div>` : ''}
      ${cs.length > 3 ? `<button class="more-c" data-open="${p.id}">View all ${cs.length} comments</button>` : ''}
      <div class="cmts">${shown.map(([h, t]) => `<div class="c"><b>${h}</b>${t}</div>`).join('')}</div>
      <div class="more-c" style="pointer-events:none">${p.when} ago</div>
    </div>
    <form class="add-c" data-cmt="${p.id}">
      <input placeholder="Add a comment…" autocomplete="off">
      <button class="send" type="submit">Post</button>
    </form>
  </article>`;
}

function homeView() {
  const mine = S.posted.map((n, i) => ({
    id: 900 + i, by: ME.h, img: n, where: 'Just now', likes: 0, when: 'now',
    text: S.draft?.text || 'First one from the new build.', tags: ['lumen'], comments: [],
  }));
  const feed = [...mine, ...POSTS];
  return `
  <div class="main">
    <div class="col">
      ${storiesRow()}
      ${feed.map((p, i) => postCard(p, i === 0 && mine.length > 0)).join('')}
    </div>
    <div class="rail">
      <div class="rail-card">
        <h3>Activity</h3>
        ${NOTIFS.map((n) => `
          <div class="notif">
            <img src="${avatar(n.h)}" alt="">
            <p><b>${n.h}</b> ${n.t} <em>· ${n.when}</em></p>
          </div>`).join('')}
      </div>
      <div class="rail-card">
        <h3>Suggested for you</h3>
        ${['noor.builds', 'kiro.wav', 'tomas.p', 'wren', 'lune.mp4'].map((h) => {
          const u = person(h), f = following(h);
          return `<div class="sug">
            <img src="${avatar(h)}" alt="">
            <span class="who"><b>${h}</b><span>${u.city}</span></span>
            <button class="btn-follow ${f ? 'yes' : 'no'}" data-follow="${h}">${f ? 'Following' : 'Follow'}</button>
          </div>`;
        }).join('')}
      </div>
      <div class="foot">Lumen · About · Help · Privacy<br>© 2026 Lumen</div>
    </div>
  </div>`;
}

// ------------------------------------------------------------------ explore

function exploreView() {
  const chips = ['For you', 'Portraits', 'Streets', 'Film', 'Architecture', 'Nature'];
  return `
  <div class="main">
    <div class="col" style="width:934px">
      <div class="head"><h1>Explore</h1><span class="sub">Updated 6 minutes ago</span></div>
      <div class="chips">${chips.map((c) => `<button class="chip ${S.chip === c ? 'on' : ''}" data-chip="${c}">${c}</button>`).join('')}</div>
      <div class="grid">
        ${EXPLORE.map((n, i) => {
          const p = POSTS.find((q) => q.img === n);
          const l = p ? likeCount(p) : 400 + n * 137;
          return `<div class="cell ${i === 4 || i === 12 ? 'tall' : ''}" ${p ? `data-open="${p.id}"` : ''}>
            <img src="${photo(n)}" alt="">
            <div class="veil">
              <span>${icon('heartf', 18)} ${nfmt(l)}</span>
              <span>${icon('cmt', 18)} ${p ? p.comments.length : (n % 7) + 2}</span>
            </div>
          </div>`;
        }).join('')}
      </div>
    </div>
  </div>`;
}

// ------------------------------------------------------------------ profile

function profileView() {
  const grid = [...S.posted, ...MY_POSTS];
  return `
  <div class="main">
    <div class="col" style="width:934px">
      <div class="prof">
        <img class="big" src="${avatar(ME.h)}" alt="">
        <div class="meta">
          <div class="row1">
            <h2>${ME.h}</h2>
            <button class="btn ghost">Edit profile</button>
            <button class="btn ghost">Share</button>
          </div>
          <div class="stats">
            <div><b>${grid.length}</b> posts</div>
            <div><b>2,418</b> followers</div>
            <div><b>312</b> following</div>
          </div>
          <div class="name">${ME.name}</div>
          <div class="bio">${ME.bio}<br>Building this app while learning how to build apps.</div>
        </div>
      </div>
      <div class="tabs">
        <button class="on">${icon('grid', 15)} Posts</button>
        <button>${icon('save', 15)} Saved</button>
        <button>${icon('tag', 15)} Tagged</button>
      </div>
      <div class="grid" style="margin-top:16px">
        ${grid.map((n, i) => `<div class="cell ${i === 0 && S.posted.length ? 'fresh' : ''}">
          <img src="${photo(n)}" alt="">
          <div class="veil"><span>${icon('heartf', 18)} ${nfmt(210 + n * 91)}</span><span>${icon('cmt', 18)} ${(n % 6) + 1}</span></div>
        </div>`).join('')}
      </div>
    </div>
  </div>`;
}

// ------------------------------------------------------------------ messages

function messagesView() {
  const t = person(S.thread);
  return `
  <div class="main">
    <div class="col" style="width:934px">
      <div class="head"><h1>Messages</h1><span class="sub">2 unread</span></div>
      <div class="dm">
        <div class="dm-list">
          ${THREADS.map((th) => `
            <button class="t ${S.thread === th.h ? 'on' : ''}" data-thread="${th.h}" style="width:100%;text-align:left">
              <img src="${avatar(th.h)}" alt="">
              <span class="who"><b>${th.h}</b><span>${th.last}</span></span>
              <span class="when">${th.when}</span>
            </button>`).join('')}
        </div>
        <div class="dm-pane">
          <div class="dm-head"><img src="${avatar(S.thread)}" alt=""><b>${S.thread}</b>
            <span style="margin-left:auto;color:var(--ink-3)">${icon('dots', 20)}</span></div>
          <div class="dm-body">
            ${MESSAGES.map((m) => `<div class="bub ${m.me ? 'me' : 'them'}">${m.t}</div>`).join('')}
          </div>
          <form class="dm-foot" data-dm="1">
            <input placeholder="Message ${S.thread}…" autocomplete="off">
          </form>
        </div>
      </div>
    </div>
  </div>`;
}

// ------------------------------------------------------------------ composer

const FILTERS = [
  ['None', 'none'],
  ['Ferrier', 'saturate(1.18) contrast(1.06)'],
  ['Bergen', 'grayscale(1) contrast(1.1)'],
  ['Kodachrome', 'saturate(1.4) hue-rotate(-8deg) contrast(1.08)'],
  ['Faded', 'saturate(.78) brightness(1.06) contrast(.94)'],
];

function createView() {
  const d = S.draft;
  const f = FILTERS.find(([n]) => n === (d?.filter ?? 'None'));
  return `
  <div class="main">
    <div class="col" style="width:934px">
      <div class="head"><h1>New post</h1><span class="sub">Step ${d ? 2 : 1} of 2</span></div>
      <div class="comp">
        <div class="stage">
          ${d
            ? `<img src="${photo(d.img)}" style="filter:${f[1]}" alt="">`
            : `<button class="drop" data-pick="1">${icon('img', 46)}<b>Drag a photo here</b><p>or click to choose one from your library</p></button>`}
        </div>
        <div class="panel">
          <div class="field">
            <label>Caption</label>
            <textarea rows="4" data-field="text" placeholder="Say something about it…">${d?.text ?? ''}</textarea>
          </div>
          <div class="field">
            <label>Location</label>
            <input type="text" data-field="place" value="${d?.place ?? ''}" placeholder="Add a place">
          </div>
          <div class="field">
            <label>Look</label>
            <div class="filters">
              ${FILTERS.map(([n, css]) => `
                <button class="f ${(d?.filter ?? 'None') === n ? 'on' : ''}" data-filter="${n}" ${d ? '' : 'disabled'}>
                  <span class="sw"><img src="${photo(d?.img ?? 6)}" style="filter:${css};opacity:${d ? 1 : .35}" alt=""></span>
                  <span>${n}</span>
                </button>`).join('')}
            </div>
          </div>
          <button class="go" data-share="1" ${d ? '' : 'disabled'}>Share</button>
        </div>
      </div>
    </div>
  </div>`;
}

// ------------------------------------------------------------------ lightbox

function lightbox() {
  const p = POSTS.find((q) => q.id === S.open);
  if (!p) return '';
  const cs = [...p.comments, ...(S.extra[p.id] ?? [])];
  return `
  <div class="lb" data-close="1">
    <button class="x" data-close="1">✕</button>
    <div class="sheet" data-stop="1">
      <div class="pic"><img src="${photo(p.img)}" alt=""></div>
      <div class="side-pane">
        <div class="post-top" style="border-bottom:1px solid var(--line)">
          <img class="av" src="${avatar(p.by)}" alt="">
          <span class="who"><b>${p.by}</b><span>${p.where}</span></span>
          <button class="more">${icon('dots', 20)}</button>
        </div>
        <div class="list">
          <div class="cmts" style="margin-top:14px">
            <div class="c"><b>${p.by}</b>${p.text}</div>
            ${cs.map(([h, t]) => `<div class="c"><b>${h}</b>${t}</div>`).join('')}
          </div>
        </div>
        <div class="acts" style="border-top:1px solid var(--line)">
          <button data-like="${p.id}" class="${liked(p.id) ? 'liked' : ''}">${icon(liked(p.id) ? 'heartf' : 'heart')}</button>
          <button>${icon('cmt')}</button><button>${icon('share')}</button>
          <button class="save" data-save="${p.id}">${icon(S.saved[p.id] ? 'savef' : 'save')}</button>
        </div>
        <div class="post-body" style="padding-top:0">
          <div class="likes">${nfmt(likeCount(p))} likes</div>
          <div class="more-c" style="pointer-events:none">${p.when} ago</div>
        </div>
        <form class="add-c" data-cmt="${p.id}">
          <input placeholder="Add a comment…" autocomplete="off">
          <button class="send" type="submit">Post</button>
        </form>
      </div>
    </div>
  </div>`;
}

// ------------------------------------------------------------------ render

const VIEWS = {home: homeView, explore: exploreView, profile: profileView, messages: messagesView, create: createView};

function render() {
  const scroll = window.scrollY;
  document.getElementById('app').innerHTML = sidebar() + VIEWS[S.view]();
  document.getElementById('overlay').innerHTML =
    lightbox() + (S.toast ? `<div class="toast">${S.toast}</div>` : '');
  window.scrollTo(0, S.keepScroll ? scroll : 0);
  S.keepScroll = false;
}

function toast(t) {
  S.toast = t;
  render();
  setTimeout(() => { S.toast = null; render(); }, 2400);
}

document.addEventListener('click', (e) => {
  const hit = (a) => e.target.closest(`[${a}]`)?.getAttribute(a);

  const go = hit('data-go');
  if (go) { S.view = go; S.open = null; return render(); }

  const like = hit('data-like');
  if (like) { S.likes[like] = !liked(like); S.keepScroll = true; return render(); }

  const save = hit('data-save');
  if (save) { S.saved[save] = !S.saved[save]; S.keepScroll = true; return render(); }

  const fol = hit('data-follow');
  if (fol) { S.follows[fol] = !following(fol); S.keepScroll = true; return render(); }

  const chip = hit('data-chip');
  if (chip) { S.chip = chip; return render(); }

  const th = hit('data-thread');
  if (th) { S.thread = th; return render(); }

  const open = hit('data-open');
  if (open) { S.open = Number(open); return render(); }

  if (hit('data-close') && !e.target.closest('[data-stop]')) { S.open = null; return render(); }

  if (hit('data-pick')) { S.draft = {img: 6, text: '', place: '', filter: 'None'}; return render(); }

  const fl = hit('data-filter');
  if (fl && S.draft) { S.draft.filter = fl; return render(); }

  if (hit('data-share') && S.draft) {
    S.posted.unshift(S.draft.img);
    S.view = 'home';
    return toast('Posted to your profile');
  }
});

document.addEventListener('input', (e) => {
  const f = e.target.getAttribute('data-field');
  if (f && S.draft) S.draft[f] = e.target.value;
  const box = e.target.closest('.add-c');
  if (box) box.querySelector('.send').classList.toggle('on', e.target.value.trim().length > 0);
});

document.addEventListener('submit', (e) => {
  e.preventDefault();
  const id = e.target.getAttribute('data-cmt');
  if (id) {
    const input = e.target.querySelector('input');
    const t = input.value.trim();
    if (!t) return;
    (S.extra[id] ??= []).push([ME.h, t]);
    S.keepScroll = true;
    return render();
  }
  if (e.target.getAttribute('data-dm')) {
    const input = e.target.querySelector('input');
    if (input.value.trim()) { MESSAGES.push({me: true, t: input.value.trim()}); }
    return render();
  }
});

render();
