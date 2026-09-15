# Reference beat sheet — "Day 1: Welcome to the Course" (10:13)

Source: AI Coder (Ed Donner), Udemy, lecture 1 of 95. Watched 2026-08-26 by
frame-sampling the player at ~35s intervals. This records STRUCTURE ONLY —
shot types, holds, topic order. Narration and footage are ours.

## The cut

| in | out | hold | shot | on screen |
|---|---|---|---|---|
| 0:00 | 0:30 | 30s | title | split: illustrated sci-fi poster L, presenter R. Course title + "Week 1 Day 1" |
| 0:30 | 1:00 | 30s | slide | "BEFORE WE START" + stop-sign art. Lists objectives/intro/curriculum/logistics |
| 1:00 | 1:30 | 30s | slide | "many products, today we start with Cursor" + inset product shot |
| 1:30 | 2:30 | 60s | footage | cursor.com marketing page, scrolled. Presenter shrinks to PiP top-right |
| 2:30 | 3:10 | 40s | footage | WINDOWS installer, "Select Additional Tasks" checkboxes |
| 3:10 | 4:00 | 50s | footage | Cursor first launch, sign-in flow in embedded browser |
| 4:00 | 4:30 | 30s | footage | macOS Cursor welcome: Open project / Connect via SSH / recents |
| 4:30 | 5:15 | 45s | footage | macOS Finder open-folder dialog, picking the project folder |
| 5:15 | 6:00 | 45s | footage | WINDOWS open-folder dialog — same step, other platform |
| 6:00 | 8:30 | 150s | footage | empty project + New Chat panel. NEARLY STATIC. PiP drifts top-right -> bottom-right |
| 8:30 | 9:00 | 30s | footage | model picker open: Auto, MAX, Composer 1, Opus 4.5, Sonnet 4.5, GPT-5.2 Codex, Gemini 3 Flash |
| 9:00 | 9:30 | 30s | footage | one-sentence prompt for a 3D arena shooter sent; agent lists files, plans |
| 9:30 | 10:13 | 43s | footage | agent exploring, writing index.html. ENDS MID-BUILD |

## Numbers

- 11 distinct shots in 613s -> mean hold **56s**. Very long takes.
- graphics 90s (**15%**) / footage 523s (**85%**)
- ALL graphics are in the first 90 seconds. Nothing cuts away after 1:30.
- presenter on camera 100% of runtime (split-screen, then PiP)
- ends on a cliffhanger; the payoff is lecture 2

## What to keep

1. **The spine.** download -> install -> sign in -> open a folder -> pick a model
   -> one sentence -> agent builds. Something real on screen inside ten minutes.
2. ~~Both platforms.~~ **NOT ours.** He cuts Windows and macOS for the same
   step; we have no Windows machine. Say "download the build for your machine"
   and move on. Decision locked 2026-08-26 - do not raise again.
3. **The cliffhanger.** Do not finish the build in lesson 1.
4. **The naive one-sentence prompt.** No spec. Matches our house rule already.

## What to fix

1. **6:00-8:30 is 150 seconds of an almost-static screen.** That is the single
   weakest stretch and it is where our graphics go.
2. **56s mean hold is too long for us** with no presenter to carry it. Our films
   run far shorter holds and cut on the idea.
3. **No mid-lesson graphics at all.** Ours are the differentiator - bridges,
   the model-picker explained as a diagram, the agent loop drawn once.
4. **No talking head in our version.** Narration + footage + graphics.
   That is the format we already ship and it removes the on-camera dependency
   for all 95 lectures.

## Non-negotiable

**No name, email, company or client identifier may reach a rendered frame.**
Neutral account, neutral folder names, fullscreen shoot, still-audit before
render, and blur anything that cannot be cut. See the memory
`coursesmith-no-pii-in-renders`.

> **Superseded 2026-08-27** by `ref-L01.md`, which adds the full narration pass.
