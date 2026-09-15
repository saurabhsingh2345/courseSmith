# L16 — "Day 3: Exploring the agents.md File and Cursor Settings for Vibe Coding" (9:49)

**Almost entirely screen recording.** He reads his hand-written `agents.md` line by
line inside Cursor. This is the most directly copyable lecture in Week 1.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | Cursor | **Shortcut keys.** Left sidebar / file tree: **Cmd B** (Ctrl B on PC). Right agent panel: **Cmd Option B** (Ctrl Option B). Toggles them on camera a few times |
| 0:40 | file tree | The joke: the repo has **exactly one file**. "It was a lot of palaver for one file — I could have just given you this file" |
| 1:00 | `agents.md` open | Points out the markdown structure — headings, bulleted and numbered lists. Then **right-click > Open Preview** to see it rendered: "this is really looking at it the way the agent will see it" |
| 1:40 | the file | **"I wrote this. I didn't generate it. I hand wrote it."** Then walks it section by section |

## His `agents.md`, section by section — copy this structure

**Business requirements**
- An **MVP / prototype** of a Kanban-style project-management **web app**
- **One board only.** "Always start simple"
- **Five fixed columns**, renameable
- Each card has **title and details only**
- **Drag and drop** to move cards between columns — "that's not so simple, but we
  need something that's going to be cool"
- Add a card, delete a card, **and no more than that**. No archive, no search, no filter
- **Priority: a slick, professional, gorgeous UI**, very simple features
- **Opens with dummy data**

He then names the technique out loud: be precise, don't be ambiguous, be assertive,
generally favour positives, make it obvious what must be built. Pre-empts the
objection — "but real projects are more complicated" — with "patience, we've got
three weeks".

**Technical details**
- A **Next.js** app, in a subdirectory called **`frontend`**
- **No persistence.** No user login
- Popular libraries, as simple as possible, elegant UI
- On the apparent duplication: **"there's no harm in repetition with these things,
  particularly for the really important points"**

**Colour scheme** — his own deck palette, pasted in. Explicitly optional: use your
own or delete it and let the agent choose.

**Strategy**
- Write a plan **with success criteria per phase to be checked off**
- Include project scaffolding, **rigorous unit testing**
- Then execute the plan
- Then **extensive integration testing with Playwright or similar**, fixing defects
- **Only complete when the MVP is finished and tested, server running, ready for the user**
- Notes this is increasingly the agents' default behaviour anyway, but he keeps it

**Coding standards** — down to three, from many more he used to need:
1. **Latest library versions, idiomatic as of today.** He deliberately does *not*
   hard-code the date — "they've got access to that"
2. **Keep it simple, never over-engineer, always simplify.** No unnecessary
   defensive programming, no extra features. "I still see it frequently
   over-complicating and I feel like you can't stress this too much"
3. **Be concise, keep readmes minimal, IMPORTANT: no emojis ever** — they get in
   the way and "actually cause some breaks sometimes with Windows PCs"

**The reassurance** (~7:30): do you have to learn to write these yourself? "Well,
yes, but honestly you can just start with what I've got here." The habit that
matters is being **specific and simple**. You don't need the technical details, the
colour scheme, or the strategy — the agent can figure those out. And the technique
he rates highest: **start minimal, run it, delete everything it built, then rewrite
the file to account for where it went wrong.** Iterate that way.

## Cursor settings (8:30-9:49)

| approx | on screen | what happens |
|---|---|---|
| 8:30 | Cursor settings | Settings menu, or **Cmd Shift J** (Ctrl Shift J on PC). Go to **Agents**, scroll to **Auto Run** |
| 9:00 | the auto-run choice | Three options framed as a risk decision: **run everything unsandboxed** (YOLO, as used for `instant`), **auto run in sandbox**, or **ask every time**. His advice: if you're unsure, stay sandboxed or ask-every-time, watch what it creates, approve each step, and move to YOLO when comfortable. **"But I'm here to take some risks, so I'm going with that from the get-go"** |
| 9:49 | end | |

## For our version

- **The whole file is the deliverable.** Write our own with the same seven sections
  and the same discipline, and put it in our starter repo so viewers can clone it.
  Reword every line; keep the structure and the specifics.
- **Open Preview is the shot that teaches it.** Raw markdown next to rendered
  markdown, side by side, is what makes "this is how the agent reads it" land.
- The **auto-run risk choice** is a genuine decision point for a beginner audience
  and deserves a graphic: three settings, what each one lets through, when to move
  up. He shows a settings pane; we can show the consequence.
- Keep his **"I hand wrote it"** beat. In a course about generating everything,
  the one file he refuses to generate is a real teaching point.
