# What was checked before upload, and what is still wrong

Measured 2026-09-07 over all 98 delivered lectures in `videos/nocode/vids/`.
**Updated 2026-09-08:** a second animation pass over Week 3 took the remaining
dead picture to zero - the Week 3 section below is rewritten; everything else
still stands as measured on the 7th.
Harness: `tools/nocode/qa/` (README there explains each script and its traps).

## The delivery folder

**98 files, 15h 21m, and nothing else.** `nocode01.mp4` and `nocode02.mp4` (older
non-Adam cuts) were moved to `videos/nocode/_archive/not-delivered/`, so every
`.mp4` in `vids/` is a deliverable. Upload order and titles: `MANIFEST.md`.
Filename numbers are production numbers with gaps in them - do not sort by name.

## Fixed today

| what | where | how |
|---|---|---|
| Week 2 + Week 1 Day 4 were missing | 37 lectures | moved in from `~/Downloads/vids/` |
| 17 files encoded audio at 96 kHz | nocode11-21, 29-34 | audio re-encoded to 48 kHz, video stream copied, originals in `_archive/pre-48k/` |
| username on screen (`ls -la` owner column) | nocode45 10:13-10:50, nocode54 9:01-9:08 | blurred for exactly those seconds, pre-redact copies in `_archive/` |
| no title metadata | 63 Week 1+2 files | title / track / album written, `-c copy` |
| fullscreen banner + volume HUD baked in | w3-01 0:00-0:15 | blacked out |
| macOS permission dialog mid-frame | w3-08, 45s still exposed after cards | blurred for exactly those seconds |
| 25 of 35 Week 3 lectures >70% frozen picture | all 35 | animated cutaways, 30-45% of each runtime |
| 55 dead spans of 12s or more still left after that | 15 Week 3 lectures | second pass, 2026-09-08: +1,500s of cutaways, **now zero** |
| w3-41's video stream ran 0.5s short of its audio | the finale | re-laid; it had been cut before the variable-frame-rate fix |

## The Week 3 result, measured

Logs: `tools/nocode/qa/holds_w3.log` (the 2026-09-02 review), `holds_w3_after.log`
(after the first animation pass, 09-07) and `second_pass.json` (after the second,
09-08). Read the **sharp** row as the result - the coarse detector's 64x36
mean-luma view cannot see a few lines of terminal text move, and terminal text is
most of what these lectures contain.

| | original | after pass 1 | after pass 2 |
|---|---|---|---|
| **dead spans of 12s or more (sharp)** | 382 | 55 | **0** |
| **lectures with no dead span at all** | 0 | 21 of 35 | **35 of 35** |
| mean % of runtime on an unchanged picture (coarse) | 76% | 31% | 25% |
| lectures above 70% coarse | 28 | 1 | 1 |

Runtime is identical on all 35 to the frame, so no narration moved. The one
lecture still above 70% coarse is **`w3-31` at 77%**, and it is the same false
positive as before: that footage is the live Signal Desk console, which the coarse
metric cannot see moving. Its sharp count is zero.

### What the second pass changed

Fifteen lectures were re-planned and re-laid. The other twenty were already at
zero dead spans and were not touched.

| lecture | coarse: orig → pass 1 → pass 2 | animated | cutaways |
|---|---|---|---|
| `w3-29` | 99% → 46% → **1%** | 89% | 73 |
| `w3-38` | 72% → 24% → **0%** | 71% | 18 |
| `w3-37` | 97% → 11% → **0%** | 69% | 17 |
| `w3-35` | 79% → 12% → **0%** | 53% | 7 |
| `w3-30` | 80% → 29% → **6%** | 60% | 12 |
| `w3-04` | 62% → 8% → **8%** | 37% | 14 |
| `w3-36` | 100% → 32% → **9%** | 66% | 9 |
| `w3-01` | 66% → 10% → **10%** | 55% | 9 |
| `w3-33` | 81% → 20% → **11%** | 46% | 11 |
| `w3-12` | 75% → 15% → **12%** | 44% | 20 |
| `w3-25` | 89% → 34% → **25%** | 45% | 7 |
| `w3-28` | 75% → 34% → **27%** | 41% | 21 |
| `w3-17` | 81% → 43% → **29%** | 44% | 23 |
| `w3-08` | 90% → 38% → **30%** | 48% | 17 |
| `w3-23` | 83% → 39% → **39%** | 32% | 11 |

`w3-29` is the headline: the worst file in the week, 1,474 seconds with a single
896-second hold, now 1,310 seconds of cutaways over 89% of its runtime and no
dead picture anywhere in it.

**Two of the 55 spans were not uncovered footage - they were our own cards.**
`w3-23` 444-456 and `w3-35` 15-32 were a 14-second headline and a 9-second editor
panel held still, and the detector was right to call them dead picture. Both are
split into shorter cards now. That is also why no new slide in this pass runs
longer than ten seconds.

Several first-pass cards were also sitting fifteen to twenty-five seconds after
the words they illustrate, because the first pass authored into the next dead span
rather than the one the sentence lands in. `w3-37`'s "almost structured",
`w3-04`'s context meter and `w3-29`'s "present, not supervising" all moved to
where the narration says them, and the windows they left took the material that
belongs there.

## Verified clean

- **Decode:** zero errors in all 98 files.
- **Loudness:** -15.8 to -16.5 LUFS, true peak -0.8 to -1.7 dBTP. Nothing clips.
- **Dead air:** the only lectures with silences over 3s are `nocode02` (4 gaps,
  44s, worst 16.3s) and `nocode08` (9 gaps of about 3s). **Neither is a defect** -
  nocode02's are the 3D shooter being played on screen, score and health bars
  moving, and nocode08's are slide pauses. The 2026-09-02 review counted them as
  dead air; the frames say otherwise. Nothing else in the course has a gap over 3s.
- **PII:** OCR at one frame every 4 seconds across the 37 new lectures - two hits,
  both fixed above. Week 3 was swept on 2026-09-01 (`autoredact.py`).
- **Week 2 pacing** is the best in the course: 29 of 30 lectures spend under 8% of
  runtime on an unchanged picture, 15-27 picture changes a minute.

## Still wrong, and honestly

1. **`nocode01` 6:22-9:15 is an empty light-theme Cursor window** - 80 seconds of
   it is a single unchanging frame, and the narration describes three panes and an
   agent that are not visible. It is the first lecture in the course. **It needs a
   reshoot**; it cannot be repaired in the edit, because the narration text for
   those segments no longer exists on disk (the voice store is content-addressed by
   hash, and `tools/nocode/scripts/vo_screen/` is gone), so cards written over it
   would risk contradicting the voice. Everything else in Week 1 is fine.
2. **Light-theme footage** in nocode01 (baseline luma 224) and nocode02 (233)
   against 9-27 everywhere else. Dimming was tested and rejected: those frames are
   nearly uniform white with 9 levels of contrast, so darkening them only greys out
   already-faint UI text. This is a reshoot item too, not an edit.
3. **Resolution jumps at the Week 3 boundary** - Week 1 and 2 are 1080p, Week 3 is
   4K. Both play fine; a platform that transcodes will normalise it.
4. **Week 1 deck films hold a slide for 12-25 seconds** (`nocode03`, `nocode08`,
   `nocode07` most often). This is deck pacing, not a defect, and the sharp
   detector finds nothing dead longer than 26s in any of them. Left alone
   deliberately - "fixing" it means re-rendering thirty films.
5. **Titles in `MANIFEST.md` for Weeks 1 and 2 are the reference course's own
   lecture titles.** They were already in the shipped metadata for lectures 22-64,
   so they are consistent, but they are not ours. **Rewrite them on the platform
   before publishing.** Week 3's titles are original.
6. **Promises made on camera - two of the four are now kept.** `nocode04` @4:07
   lists four projects: the digital twin (nocode26) and the legal-doc SaaS
   (nocode59-64) both exist now. `nocode08` @2:04 draws a fifteen-day grid and the
   course now ships exactly fifteen days, five per week. **Still broken:**
   `nocode07` @0:19 badges "MORE ON GASTOWN LATER" and Gastown was cut from Week 3;
   `nocode04` says "live market data" and Signal Desk runs on a simulator. Both are
   single sentences over otherwise good footage - a pickup, not a reshoot.

## Two flagged items deliberately left alone

**`w3-04` 2:45-3:25 shows the local agent registry** (`claude-code-guide`,
`statusline-setup`, `vercel:*`) inside the line
`Error: Agent type 'simulator' not found. Available agents: ...`. That error IS
the lesson - the narration over it explains that the registry is read at session
start, so a definition on disk is not dispatchable until the session restarts.
Blurring it would hide the thing the voice is telling you to read. The names are
installed plugin agents, not personal information: no name, email, company or
home path is in the frame. Animation cards already cover 2:56-3:13 of it.

**`w3-33` 3:17-3:27 has the command palette open** on "View: Toggle Panel
Visibility". Cosmetic, no information in it, and the cards cover 3:04-3:16.
Cutting it would move the narration.

## The four showcase films

`~/Desktop/showcase/`, each a continuous run of consecutive lectures behind a
six-second title card in the house style. Nothing is cut or re-ordered inside them.

| file | length | lectures |
|---|---|---|
| `showcase-01-claude-code-from-zero.mp4` | 31:23 | nocode35-37 |
| `showcase-02-jira-ticket-to-pull-request.mp4` | 33:41 | nocode54-56 |
| `showcase-03-expert-week-sub-agents-hooks-plugins.mp4` | 30:05 | w3-01..06 |
| `showcase-04-agents-md-and-the-ralph-loop.mp4` | 25:11 | nocode12-13 |

Plus two **session reels** built by `tools/nocode/showcase/reel.py`, which cuts
inside a lecture and cross-dissolves every join instead of hard-cutting:

| file | length | lectures |
|---|---|---|
| `showcase-05-day-one-from-nothing-installed.mp4` | 26:23 | nocode01-04, with nocode01's empty-editor stretch removed |
| `showcase-06-session-two-how-the-thing-actually-works.mp4` | 24:52 | nocode09-11, ending inside 11 on its "next lecture" card |

`nocode01` is cut at 379.3s and resumes at 502.4s. Both points are the middle of
a sentence pause AND a segment boundary of the original assembly, so a whole
beat comes out rather than half of two. `~/Desktop/showcase/README.txt` says
which reel suits which viewer.
