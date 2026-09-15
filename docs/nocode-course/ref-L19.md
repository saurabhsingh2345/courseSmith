# L19 — "Day 3: OpenAI Codex VS Code Extension: Zero-Shot Kanban App Build" (10:22)

**100% screen recording.** Build three of four, and the one he declares the winner.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | Cursor terminal | The reset ritual again: `cd ..`, **`mv kanban copilot_kanban`**, **`git clone`** → fresh `kanban` with one file |
| 1:00 | openai.com/codex | Codex is **first and foremost a CLI tool** like Claude Code — "next week is when we get deep into CLI tools". Shows the CLI screenshots. But there is **also a VS Code extension**, and since this week is about IDEs, that's the route |
| 2:00 | VS Code, Extensions | **Cmd Shift X**. Clears the old Copilot search, searches **Codex**. Top result, from **OpenAI**, verified tick, **3-and-a-bit million downloads**. Install; may need a restart |
| 2:50 | VS Code | The **Codex icon** appears — and its sidebar opens **on the left**, "so you know something's different" |
| 3:10 | File > New Window | Open `projects` > `kanban` > Open. Then a small comedy beat: **Copilot's panel is still there** — "x out of that, go away Copilot, we're done with you, you're dead to us." Cmd + to zoom, drags the panel wider |
| 4:00 | Codex settings | **Sign in with your OpenAI account.** Believes Codex needs a **paid ChatGPT subscription plan** — invites you to check whether free models are offered, and says **you can skip this one**, "it's not important that you have accounts on all these platforms" |
| 4:40 | the dropdowns | **Agent full access** selected — "again, the YOLO mode". Model: **GPT 5.2 Codex**, "the strongest, most powerful model we've got" — with the caveat that there will be newer ones by the time you watch |
| 5:20 | reasoning effort | Sets it to **high**. And the nuance that matters: **more thinking is not always better** — it can overthink, "agonise over stuff in a way that feels counterproductive, like it should have just gone for it". Depends on the project; "there's an art and a science to that" |
| 6:00 | prompt | **There is no plan mode in Codex**, but `agents.md` already tells it to plan first. So he types only **"please go ahead"**. It identifies documentation to read, reads `agents.md`, thinks, and goes |
| 6:30 | after the wait | **"That was a good 15 minutes. I know it was just a second for you, but it was 15 minutes for me."** Points at the context indicator: **16% used, 42,000 of 258,000 tokens**. Its report: dev server running via `npm run dev`, how to stop it, and **74 files changed** |
| 7:30 | localhost:3000 | **"Oh, wow."** The app is called **Kanban Studio** — "Focus. One board, five columns, zero clutter." He tests all four: move a card, **reorder**, **remove**, **rename a column** — all work |
| 8:20 | reaction | **"I've got to hand it to it. Definitely best so far. It looks beautiful."** Names it as **zero-shot** — instructions given, left alone, no questions asked |
| 8:50 | **the confession** | Pre-empts "you haven't shown us how the code works, how can I check it's good?" — **"I don't know how to write this kind of front-end code either. I'm a lousy React developer."** The point: **you can act like the manager and give feedback.** He knew `npm run dev`, but you can ask the agent that too |
| 9:50 | the caveat | "The devil's in the detail." Fine for a toy project; much harder with a bigger system and more advanced requirements — **"that's why this is still day three of a three-week course"** |
| 10:22 | end | "I think this is really awesome. I hope you like it too" |

## For our version

- **The confession is the beat of the lecture and it is perfect for a no-code
  audience.** An instructor with 400k students admitting he can't write the
  front-end code his agent just produced is the whole thesis of our course in one
  sentence. Give it room; don't bury it.
- The **reasoning-effort nuance** ("more thinking is not always better") is a real,
  non-obvious insight and it deserves a small graphic — effort dial vs quality,
  with a peak rather than a slope.
- **Cut the 15-minute wait to a hard ramp.** He says "just a second for you";
  ours should show it as a visible 12-14x ramp with the file count climbing, per
  our house rules.
- Note for continuity: he compares fairly and names the models — **Cursor got
  Composer, Copilot got Claude Haiku 4.5, Codex got GPT 5.2 Codex on high.**
  Our version must state which model each build actually used or the comparison
  is dishonest.
