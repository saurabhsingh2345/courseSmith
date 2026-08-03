# Build Without Code — production notes

Thirteen lessons. See `docs/no-code-track.md` for the engine work behind them.

## What records today

Lessons carrying a `[CAPTURE: tool=…]` marker record a **real session** of that
tool. Phase 1 supports terminal tools only, and the recording runs on the host
with your real credentials — `coursesmith doctor` reports what is missing.

| lesson | capture | tool needed |
|---|---|---|
| 05 | an agent reading a codebase and editing files | `claude` |
| 09 | a project going live | `vercel` |
| 09 | the repository it deployed from | `gh` |
| 11 | an agent finding and fixing a real bug | `claude` |
| 13 | the capstone, end to end | `claude`, `vercel` |

Install what you need: `brew install vhs` (the recorder itself), then the
tools — `npm i -g vercel`, `brew install gh`.

## Web captures

The browser driver is built: `rod` drives a checked-in **take** script and
saves real frames, with the origin recorded as provenance and a focus box per
shot so the scene pushes in on the thing being talked about rather than on the
middle of a screenshot.

A take can capture **stills**, **video**, or both in one run:

- `shot` saves a frame, optionally with a `focus` selector so the scene pushes
  in on the prompt box rather than the middle of the page. Right for "here is
  where that lives" — better than four seconds of footage, and it cannot flake.
- `record` / `mark` / `stop` records the screen. Right for what is inherently
  temporal: an app assembling itself, a deploy going green. Marks are measured
  against the recorder's own clock, so the engine can compress the waiting and
  still land exactly on the moment the app appears.

Reach for a shot by default and a recording when time is the subject.

**Sign in once per site**, then every capture runs headless:

```sh
coursesmith footage login lovable
```

The session lives in a browser profile under your home directory. Nothing goes
near the repository and no automation ever types a password.

### Takes ready to switch on

`takes/first-build.yaml` drives lesson 3's flagship shot. **Its selectors are
unverified** — they were written from the shape these pages usually have, not
from Lovable's real markup — so it is deliberately *not* wired into
`lessons/03-your-first-build/lesson.md` yet. A course that fails to build for
everybody without a Lovable account is worse than one lesson using drawn frames.

To turn it on, once the selectors are checked against the live site:

```
[CAPTURE: tool=lovable, take=first-build; the empty prompt box, and the same box with one sentence in it]
```

### Still to write

| lesson | the shot | why it must be real |
|---|---|---|
| 06 | a Figma frame becoming a working screen | the handoff is the surprising part |
| 07 | a Supabase table filling with rows | data stops being abstract when you see it land |
| 12 | a pricing page with the real numbers on it | the honesty lesson cannot use figures we drew |

## Desktop captures (Cursor, Figma)

The one recording with **you** in it. These apps have nothing to select, so the
engine frames the window, records, crops and times — and you drive the app.

```sh
coursesmith footage shoot no-code/05
```

It shows one beat at a time and stamps a mark when you press Enter. Take your
time: the engine compresses the waiting afterwards, so pausing to get a beat
right costs nothing in the finished video.

`takes/cursor-agent-edit.yaml` is lesson 5's alternative shot — the same agent
work, in an editor rather than a terminal, for viewers whose picture of "coding"
is a window with files in it. Unlike the web takes, this one needs no selectors,
so it works as written.

First time only: tick your terminal under **Screen Recording** *and*
**Accessibility** in System Settings → Privacy & Security. `coursesmith doctor`
checks both.

`run` deliberately refuses desktop captures and points at `footage shoot` — it
blocks on a keypress, and a batch build that stopped waiting for somebody looks
exactly like a hang.

## Freshness

These products redesign themselves and a course full of last quarter's frames
is wrong with nothing to tell you. That is what `footage list` is for:

```sh
coursesmith footage list no-code --stale-after 90
```

It reports every clip, when it was shot, what version of the tool it caught,
and which have aged past the window. Re-shooting is re-running the take.

## Fixtures

A capture may name a starting project with `fixture=<name>`, seeded from
`courses/no-code/fixtures/<name>/`. The recording runs in a throwaway copy, so
a fixture is never modified by a take.
