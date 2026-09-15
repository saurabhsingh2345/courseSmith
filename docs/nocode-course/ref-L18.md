# L18 — "Day 3: Building a Kanban Board with GitHub Copilot in VS Code" (14:44)

**100% screen recording**, and the longest lecture in Section 1. Build two of four.
Its real subject is not Copilot — it is **how to make an agent debug properly**.

## The reset ritual (0:00-2:00)

The pattern repeated before each of the four builds, worth learning once:

1. Back in Cursor on the `instant` project, close the Kanban project
2. `pwd` → `/Users/ed/projects/instant`, then **`cd ..`**
3. **`mv kanban cursor_kanban`** — renames the finished build so it's kept as a record
4. Proves it's gone: `cd kanban` → "no such file or directory"
5. **`git clone`** the same repo again → a fresh `kanban` with only `agents.md`

## Setting up Copilot (2:00-6:30)

| approx | on screen | what happens |
|---|---|---|
| 2:00 | github.com | Copilot on the landing page; **plans and pricing** shows a **free tier** with a monthly request allowance — "you can do this without paying a cent" |
| 2:40 | slide/site | Three ways to use it: CLI (next week), a web front end, or **VS Code** — the popular way, and what he uses. Notes **Cursor is itself a fork of VS Code**, so it will look very familiar |
| 3:20 | code.visualstudio.com | Download VS Code, or Check for Updates if you have it |
| 4:00 | VS Code, Extensions | **Cmd Shift X** / Ctrl Shift X, or View > Extensions. Searches **GitHub Copilot**, verifies the publisher is **verified github.com**, notes **69 million downloads**. Install |
| 5:00 | VS Code | File > New Window > Open > `projects` > `kanban` > Open. **KANBAN in block capitals** confirms it |
| 5:30 | bottom-left avatar | **Sign in to GitHub** — required |
| 6:00 | the chat panel | **Cmd Shift I** / Ctrl Shift I, or View > Chat. Panel appears **on the right**. Tour: instruction box, **plan mode** dropdown, model selector on **Auto** (free-tier models listed) |
| 6:20 | usage hover | Hovering shows a **usage bar against your allowance** |

## The build (6:30-9:00)

- Plan mode, prompt: **"please plan the task at hand"** — it picks up `agents.md`
- Plan completes. Same joke: "seems like it might be interesting, but I'm not going
  to read it." Presses **Start implementation** (declines the run-in-background option)
- **Permission prompts arrive constantly** — allow the terminal command, allow
  `no` + enter. He approves one at a time, then escalates: **"always allow — allow
  all commands in this session"**, framing it as choosing your own risk level
- **It gets stuck.** Coded a lot, fixed a few bugs, **started the server in the
  wrong directory**, got errors, corrected the directory, started it again, then
  **sat waiting, not realising the server was up.** He has to **cancel** it
- **The surprise**: he runs `npm run dev` in `frontend` himself and **the app is
  actually finished and good** — it only floundered at the very last step. Reveals
  the model behind it was **Claude Haiku 4.5**, a small model, and "I think it's
  done a fabulous job" — he prefers this layout to Cursor's, and it has **no error
  badge**
- Tests the four requirements: drag and drop **great**; **delete is broken**;
  rename a column **works**

## The debugging lesson (9:00-14:44) — the reason this lecture matters

1. Presses **Keep** on the changed files, stops the server, gives feedback:
   everything works except the delete card feature
2. It reports the fix is done. **He calls out the anti-pattern by name** — the
   "classic move by coding agents": **guess what the problem is, put in a fix,
   claim victory.** Two things wrong with it: it shouldn't guess, it should prove;
   and it shouldn't claim fixed without testing and demonstrating
3. States the correct sequence: **reproduce the issue → prove the root cause → fix
   it → demonstrate it's fixed** — and adds that if you're new to development,
   **you have to be the boss of that process and push back even when it looks right**
4. Tests it. **No change at all.** "If it hasn't proven the problem, it probably
   doesn't work." A new error badge has now appeared too
5. Re-prompts with the discipline spelled out: *that didn't fix it — first
   reproduce the problem, prove you've reproduced it, find the root cause, fix it,
   and prove you fixed it*
6. It claims reproduce → fix → test. **This time delete actually works**
7. Closes on why he's glad it went wrong: **don't believe it when it says it found
   and fixed the problem**, and **give it the four-step instruction explicitly**

## For our version

- **This is the most valuable lecture in Day 3 and it is entirely unscripted luck.**
  We cannot plan a failure — but we can plan to *keep* one. Our house rule already
  says shoot from frame one and cut nothing that goes wrong; here it is the lesson.
- The **four-step debugging instruction is the single most reusable artefact in
  Week 1.** Make it a card the viewer can pause on, and bring it back every time a
  build breaks later in the course.
- **The reset ritual should be one animated graphic**, not four repetitions of
  terminal typing. Show `mv` + `git clone` once properly, then use a 3-second
  version of it before builds two, three and four.
- 14:44 of pure footage means our version needs almost no graphics here — this is
  where the section's footage average gets pulled up.
