# Week 3 — state

**Thirty-five lectures. Shot, cut, written, voiced and rendered.** The upload
list is `DELIVERY.md`; the reasoning behind its order is `ORDER.md`.

## One command for the state of the week

```
python3 week.py --live
```

Prints every lecture: minutes of footage it was cut from, what the narration
runs, what that footage could carry at natural pace, the **room** left between
those two, whether the voice is paid for, and the length of the delivered file.

## The captures, and what each one is

Delivered week: `capture-A1` `A2` `A2b` `A2d` (Days 1+3) · `B` (Day 2) ·
`C` + `C2` (Day 4) · `D` `D2` `D3` (Day 5).

Top-up: **E** (58.5 min, signal-desk, w3-27..w3-31) · **E3** (7.9, w3-32 first
half re-shot) · **E2** (6.4, w3-32 second half) · **G** (17.4, GitHub,
w3-33..w3-35) · **H** (16.2, headless, w3-36..w3-38) · **I** (25.8, failure and
verdict, w3-39..w3-41). Scripts are `shoot_e.py`, `shoot_g.py`, `shoot_h.py`,
`shoot_i.py`, sharing `w3kit.py`.

**E's own w3-32 filmed a terminal that never received a keystroke.** After
Chrome quit, VS Code came forward but the keyboard went to the explorer tree, so
four beats' worth of commands vanished into type-ahead file search with no error
anywhere. `w3kit.close_browser()` now forces `Terminal: Focus Terminal`, and
`focus_shoot_window` calls `activate` before `set frontmost` because System
Events loses to an app that has just taken focus. Both cost a lecture each.

## The voice: bought, in full, once

ElevenLabs Adam is metered monthly and the key is shared with the second
machine — it moves while you are not looking. Two things follow and both are
done:

1. **`adam_lay.speak()` is content-addressed.** Every sentence is filed under a
   hash of its exact text in `tools/nocode/scripts/_voice/`, so a sentence is
   paid for once ever, wherever it later moves and whichever lecture it ends up
   in. Before this, inserting one sentence into a segment renumbered every clip
   after it and re-bought all of them.
2. **Buying is separated from rendering.** `prefetch.py` speaks a lecture into
   the store and stops — network and disk only, no ffmpeg — so it is safe to run
   while a capture is rolling.

**All 272,179 characters of the week are in the store.** A render of anything,
now or in October, costs nothing.

```
python3 quota.py --live         # prices any list against the live balance
python3 prefetch.py <lecs>      # buy, render later
python3 finish.py               # render the week in DELIVERY.md order
```

## PII is swept, not spot-checked

**`autoredact.py` is the tool that closes this.** It OCRs every frame of every
cut at four-second intervals, looks for the username, machine name and company,
re-samples every second around anything it finds, and blurs the line it sits on
for exactly the seconds it is legible. Fifteen regions across nine lectures.

```
python3 autoredact.py --plan        # report boxes, change nothing
python3 autoredact.py               # blur everything that needs it
python3 autoredact.py w3-33         # or just one
```

Three things about it are load-bearing:

* **It blurs onto the cut as it stands**, not from `.orig.mp4`. Rebuilding from
  the original is right for one hand-tuned pass and wrong for a second one,
  because the second pass's argv does not contain the first pass's boxes and
  they are silently lost. `redact.py --onto-current` is the flag.
* **`enable` goes on the blur, not only on the overlay.** Ungated, ffmpeg
  blurs every frame of the cut and throws all but the redacted seconds away:
  nine boxes over a ten-minute 4K lecture took twenty minutes instead of five.
* **A hit is widened to the whole line and padded either side in time**, because
  a terminal scrolls between samples and a box that fits the word at *t* is off
  the word at *t*+0.4.

What it found was not what anybody had been looking for. `gh` printing an
account name in an issue header was known and hand-redacted. The sweep found the
shell echoing **absolute home paths** (w3-07, w3-15, w3-23), an `ls -la` whose
owner column is his username nine times over (w3-13), and the prompt reverting
to `<user>@<machine>` after Claude Code exits (w3-08, w3-16, w3-27, w3-28,
w3-32, w3-33, w3-39, w3-40). **w3-07, w3-13, w3-15 and w3-16 had already
shipped with it in frame.**

`scanpii.py` still exists and is the cheaper check — it reports *that* a cut is
dirty. `autoredact.py` is what fixes it. Run both after any new capture.

## Numbers were checked against the frames, not against the repository

The top-up narration for `w3-27` to `w3-32` was written from the repository's
own artefacts — `plan.md`, `REVIEW.md`, `FIXED.md`, `run.log` — because the cuts
did not exist yet, and the repository moved during the shoot. Every numeric
claim has now been read off the frame it sits over. Five were wrong:

| lecture | was | frame says |
|---|---|---|
| w3-30 | 22 JavaScript files, ~1900 lines | **2310 total**, 26 files |
| w3-27, w3-29, w3-32, w3-40 | nineteen hundred lines | two thousand three hundred |
| w3-32 | "no installed modules, no lock file" | `node_modules` and `package-lock.json` are in the explorer |
| w3-32 | sort: −23, −20, −17, +17 | −23.21, −20.31, **+17.06, −16.87** |

Confirmed correct and left alone: 27 tests / 115 ms in w3-30, 28 tests / 125 ms
in w3-32, 366 lines of orchestrator in w3-28, the seven job times summing to
just under 24 minutes against 17:53 of wall clock in w3-29, 104 packages
audited, 17 of 30 instruments visible in w3-31, and the ALP alert firing at
408.91.

**w3-32's second half is E2, shot later than E3**, so the file tree in it
carries Day 5 artefacts (`ci.yml`, `VERDICT.md`, `risky.txt`). Nothing claims
otherwise any more, but do not write a claim about that tree without looking at
it.

## Per-lecture loop

```
cut.py <capture>  ->  frames.py <lec> --every 75  ->  LOOK at the frames
  ->  write narration/<lec>.json  ->  fit.py  ->  balance.py --write --trim
  ->  autoredact.py <lec>  ->  prefetch.py <lec>  ->  assemble.py <lec>
```

`assemble.py` copies to `videos/nocode/vids/` itself — that copy is delivery.
`autoredact.py` must run **before** `assemble.py`, because it rewrites the cut.

## What broke before, and what prevents it now

1. **Clipping.** `loudnorm` always emits 192 kHz, so the AAC encoder resampled
   internally and overshot the limiter by +3.6 dB. `assemble.py` has
   `aresample=48000` between loudnorm and the limiter.
2. **PII in frame.** `autoredact.py`, above. `tails.py` for the last frame.
3. **A question card answered "2"** put a false claim on camera. Every build
   prompt carries `NO_ASK`.
4. **`wait_idle` is blind to a plain shell process.** Use `wait_process()` /
   `wait_for()` from `w3kit`.
5. **A `| tail -N` beat shows nothing until it exits**, so a thirty-minute build
   piped through `tail` films an empty screen and then dumps. Stream it, or
   `tee`.
6. **The peak guard read only positive samples.** `peak_of()` reads
   `Peak level dB`.
7. **Blocked lectures leave dangling promises.** Grep the narration for forward
   references after any change: `tomorrow`, `next we`, `day four`, and each cut
   tool's name. All four places that have to move together are listed at the
   bottom of `ORDER.md`.
8. **A deepening pass can outrun the picture.** `balance.py` moves whole
   sentences forward, then drifts a boundary by up to ten seconds, and only
   trims as a last resort.
9. **`cut.py` stream-copies by default.** The capture is `g=15` and `PAD_IN` is
   0.6s, so the seek error is inside the pad: 54 minutes of footage cut in two
   seconds instead of 54 minutes. `--reencode` restores the old behaviour.

## Tools verified 2026-09-01, rather than assumed

| tool | state |
|---|---|
| Claude Code (Max), Agent SDK, `gh`, Docker | work |
| Cursor CLI **and** IDE | **signed out** — the IDE returns *Authentication error* on an agent turn |
| Antigravity | **signed out** — onboarding wants a Google login |
| Copilot | not installed, no subscription |

## Two cuts that exist and are not lectures

`cuts/w3-09.mp4` and `cuts/w3-11.mp4` are the stubs of two blocked lectures.
They have no narration and are not in `DELIVERY.md`. **w3-11 contains a personal
GitHub URL** on screen for 75 seconds; it is blurred now, but the file is not
material and is only kept so the capture index stays honest.

## Non-negotiable

Build with sonnet. No name, email, company, `/status`, Cursor Settings,
`docker ps` or `docker images` in frame. GitHub work goes in the **heftyradius**
org, never a personal account, because `gh` prints `owner/repo` on nearly every
command. Shoot from `/Users/Shared/projects`. Never drive Chrome "Browser 1".
`shootmode.py on` before, `teardown.sh` after.
