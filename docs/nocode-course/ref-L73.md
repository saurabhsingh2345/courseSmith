# L73 — "Week 3 Day 2: Remote Execution & Cloud Sandboxes with Claude Code on the Web" (9:59)

**Slides only.** The map of three approaches, before any of them are used. Dense
but entirely conceptual — L74 and L75 do the doing.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | slide | The naming mess, stated plainly: **remote execution / Claude Code on the Web / Codex on the Web** all mean roughly "the agent doesn't run on your computer" |
| 0:40 | slide | What it buys: delegate to **a different execution engine**, not a different agent · ten things running at once · **already sandboxed** · organised through GitHub · **work from your phone**, then pick it up at your desk |
| 1:40 | slide | Honest caveat: **hugely evolving space**, offerings differ, "you may have a different set than I have". Claude Code was first; Codex now does it too |
| 2:30 | slide | **What we are NOT doing**: Docker / VS Code **dev containers** — the old box-in-a-box way. It still works, everything here goes further |
| 3:10 | **approach 1 — native sandbox** | Built into Claude Code, triggered by **`/sandbox`**. Auto-approves inside the sandbox. Runs **locally**. Better than Docker because it is **implemented at the OS level** — lightweight and fast. Notes **Cursor has this too** through its UI |
| 4:30 | slide | **The catch**: Mac and Linux are fine; **Windows needs WSL**. If you don't already know WSL, use the other two approaches instead |
| 5:20 | **approach 2 — managed cloud sandbox** | "Claude Code on the Web". Anthropic runs Claude Code on their machines. Four ways in: an **`&` prefix** on any instruction · **`claude --remote`** from the terminal · **a GitHub issue tagged `@claude`** · and **Teleport**, grabbing a running cloud session and making it your active one. **`/tasks`** lists what is running |
| 8:00 | slide | Requires the **Claude GitHub app**; everything must be checked in, because the cloud clones the repo |
| 8:40 | **approach 3 — third-party sandbox** | **Sprites.dev**, by the Fly.io people. Fast, simple, flexible, built for coding agents. **And it works with any coding agent**, not just Claude |
| 9:59 | end | |

## For our version

- **Three approaches is the graphic of Day 2** — local sandbox / Anthropic-managed
  cloud / third-party — and it is one of the few places in Week 3 where animation
  genuinely beats footage. Build it once, **recall it at the top of L74, L75, L76**.
- **Keep "we are not doing Docker".** Saying what you are deliberately skipping,
  and why, is rare and it orients people who know Docker already.
- **Keep the Windows/WSL caveat.** Our program is macOS-only by decision, but this
  is a genuine fork for a chunk of the audience and costs ten seconds.
- **Cursor's sandbox gets a mention here** — and since Cursor is our second tool
  all week, we can show it for real rather than mention it in passing. We already
  filmed Cursor's four Run Modes in W1 L16; **recall that** and note it is the same
  idea one layer up.
- **Footage share here will be low (~30%) because the lecture is a map.** That is
  the exception; L74-L77 pay it back at 90%+.
