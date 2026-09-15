# L20 — "Day 3: Building a Kanban App with Antigravity IDE and Gemini 3 Pro" (10:37)

**100% screen recording.** Build four of four — and the only one whose *setup*
genuinely differs, because Antigravity rejects the `agents.md` convention.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | the Codex result, still up | **The fairness caveat, delivered properly**: Codex did so well because it ran the frontier model. Cursor's Auto gave him **Composer**, Cursor's cheaper in-house model. Copilot gave him **Claude Haiku 4.5**. Codex ran **GPT 5.2 Codex on high reasoning**. **"Strong model means strong results"** |
| 1:00 | Cursor terminal | Reset ritual: `cd ..`, **`mv kanban codex_kanban`**, **`git clone`**. He notes he keeps coming back to Cursor's terminal just because it's a useful terminal to have open |
| 1:40 | **antigravity.google** | "Cool URL." *Experience liftoff with the next generation IDE.* Download for your platform; Mac drag-to-Applications, PC wizard — **take all the defaults** |
| 2:40 | Antigravity, first launch | **It's VS Code again** — another fork, "that's why it's similar, it's the same thing". File > New Window if you don't see the welcome screen |
| 3:10 | top right | **Google Auth sign-in**, usually done during download |
| 3:30 | Open Folder | `projects` > `kanban` > Open. Familiar layout; **agent chat on the right**. It pitches **Gemini 3 Flash** — "frontier level intelligence, blazing fast, more generous quotas" |
| 4:00 | the panel | **Plan mode** available (left on). Model list — and notably **Google offers the Anthropic models too**, plus **GPT-OSS**, the open-source one |
| 4:30 | Antigravity settings | **Agent Autofix lints** — a lint error is code failing basic structural checks; he turns it **on**, "good practice". Then the YOLO control: **Request review** (the safe default), **Agent decides** ("rather nice"), or **Always proceed** — he picks **Always proceed**, "I'm going to be YOLOing all the way today. You should only do that if you're comfortable with it" |
| 5:40 | **the divergence** | **Antigravity has not adopted `agents.md`.** So: select all in `agents.md` and copy (**Cmd A, Cmd C**) |
| 6:10 | file tree | Right-click > **New Folder** > **`.agent`** — a special folder Antigravity expects, with subfolders for different things. Inside it, the important one is **`rules`**. Inside `rules`, a new markdown file — **`strategy.md`** |
| 6:50 | strategy.md | The editor shows a **special header with an activation mode**: **Always on** (his choice), Manual, or Model decision. Pastes in the whole of `agents.md`, **saves (Cmd S — "make sure the dot disappears")**, then **deletes `agents.md`** to avoid confusion |
| 8:00 | prompt | **"please go ahead"**. Confirms the model is **Gemini 3 Pro on high** — "the strongest version" |
| 8:20 | the result | A **Kanban MVP walkthrough** summary containing **an actual screen recording of the app working** — "Is that for real? Isn't that crazy?" — plus run instructions and an Accept All prompt |
| 8:50 | **the reveal** | While running, it **drove real browsers itself to test the screens** — that's where the recording came from — and it used **Playwright** for browser-automation tests. Some tests failed and it fixed them; **one it apparently never fixed** |
| 9:30 | the app | "Fresh and clean, I think it looks very good." Drag between columns **works**, rename a column **works**. But **Add Card is weak** — a janky localhost JavaScript prompt, and **no way to add a description** |
| 10:37 | end | "That adding a new card is the only thing that feels a little bit simplistic. Let's see if we can't improve that" — straight into L21 |

## For our version

- **The `.agent/rules/strategy.md` conversion is the one genuinely new mechanic in
  Day 3** and it is fiddly on screen. Animate it: the same content, two homes, one
  standard and one holdout — then shoot the actual file creation.
- **The fairness caveat must open our version too.** Four builds compared while
  three ran different-strength models is only an honest comparison if you say so
  out loud, up front.
- **The agent driving its own browser and recording itself is the most striking
  footage in Day 3.** He almost throws it away as an aside. Ours should hold on it
  and name what happened — an agent writing its own acceptance tests.
- The **activation-mode choice** (always on / manual / model decides) is a neat
  callback to L11's "it's not included every single time". Link them.
