# L87 — "Week 3 Day 4: Building a Trading Platform with Claude GSD — A 5-Hour Deep Dive" (10:21)

**The grind.** GSD interviews him, plans ten phases, and then takes **five hours
and roughly ten times the tokens** to build what agent teams built in thirty
minutes. The most valuable *negative* result in the course.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | Claude | Answers "what do you want to build?" with **"build the entire project, everything as described in `planning/plan.md`"** — the same brief the agent team got |
| 0:40 | Claude | **The interview, and it is the best argument for GSD.** It asks focused questions: **how polished must the front end be?** → *production quality, not a demo.* **Should the LLM chat work without an API key, mock mode?** → *no, it should need a key.* **Docker files and start/stop scripts, or just app code?** → *include Docker.* Submit answers. *"I love this — the way it's interactive and comes back with distinct questions"* |
| 2:00 | Claude | **His framing: "more controlled, on rails, but still with a lot happening in parallel."** Approves and lets it run |
| 2:30 | Claude | **~20 minutes later.** It produces **V1 requirements** covering database, APIs, front end, LLM chat, Docker, end-to-end tests, and **asks him to review the document** |
| 3:10 | editor | **`.planning/requirements.md`** — *"a bit confusing since we called our own one `planning`"*. Scrolls it: **IDs for each part, a self-written checklist, an explicit out-of-scope section**, and a **V2 deferred list** (enhanced visualisation, social/discovery, Terraform). *"Super impressive… a great opportunity for us to come in and read it and review it."* **Approve** |
| 4:30 | Claude | It moves straight on to the roadmap **without him running the next command** — *"you don't need to run the commands one by one; if you interact with it this way it will just drive"*. **Main context window visibly filling** |
| 5:10 | Claude | **The roadmap: ten phases.** Database foundation · portfolio · trade execution · watch list · app assembly · LLM chat integration · front-end foundation · watch lists · portfolio/trading · chat interface · packaging and testing. **His verdict, and it is a real critique of L85:** *"this is a better approach than the Claude agents — I like the way the UI is deferred to a bit later. I thought it was odd that Claude teams began with the UI."* Chat interface near the end **makes total sense**. Approve |
| 6:20 | Claude | Skips `discuss` — *"it depends on the size of your project"* — and goes straight to **`plan phase 1`** |
| 6:50 | Claude | **`/usage` — 8% → 18%.** *"Almost double the amount just to get to planning phase one that the entire Claude teams used before."* **A token hog, and slow — an hour in already.** Wonders aloud whether **ten phases was over-boiling it** |
| 7:40 | Claude | Executes phase 1. **Context bar in the yellow** — he says he'll watch how it handles compaction. Then plans **phases 2 and 3 in parallel** on a hunch that the roadmap said they could run together. **It can** |
| 8:20 | talk | **The reveal.** *"I have some slightly unexpected news… the project has finished, 100% complete. The thing you maybe weren't expecting is that it's actually **five hours later** for me."* **It ground and thrashed.** Some parallelism, but *"basically it was very serial and everything got checked and double-checked in an agonising way"* |
| 9:00 | talk | **His diagnosis: over-storied.** He should have condensed the roadmap phases. **Completion alone took half an hour** — it marked the project complete, then had to update its status, which meant **re-running tests that had already passed**, which **hit an issue**, which sent it back to thrash again. *"It says it churned for one minute — but that's not true"* |
| 9:40 | Claude | **`/usage`, the honest accounting.** Session at 24%, past a day boundary so it reset; the weekly went **9% → 17%**, and the other **8% → 9%**. **His summary: roughly 10× the tokens and 10× the time of agent teams.** But: *"very thorough, very disciplined, every single step, tons of documentation, hundreds of tool calls"* |
| 10:00 | talk | **And the reframe that saves it:** *"I could have left this going overnight and woken up the next morning and it would have been working for me all through the night."* The question for the break: **will it work first time, and will it look different?** |
| 10:21 | end | |

## For our version

- **This lecture is the reason Day 4 is worth cloning.** Two orchestrators, same
  brief, same model, one costs 10× — **that number is the whole point and it must
  be on screen, not just in the narration.** Build one graphic for Day 4 and make
  it this: teams vs GSD, time and tokens, side by side, drawn from **our own**
  `/usage` numbers.
- **The five-hour gap is a hard scheduling constraint on us.** Plan the shoot so
  the GSD run starts and we come back the next day. **Budget the tokens before we
  start** — this is the single most expensive thing in the program. Decide the
  cap first and say the cap on camera.
- **His roadmap critique of his own earlier run is gold — keep it.** "I thought it
  was odd that Claude teams began with the UI" is an instructor comparing two of
  his own results honestly. That is the tone of the whole week.
- **The requirements-doc review is the one calm beat.** Hold on the file, scroll
  slowly, let the out-of-scope and V2-deferred sections land. **Human review of a
  machine-written spec is L79's advice made visible.**
- **Keep "over-storied."** Ten phases was too many, he says so, and the student
  learns that the knob exists. If our run is faster because we use fewer phases,
  **say that we changed it and why** — that is a better lesson than matching him.
- **Write the result narration after the run.** Our timing and token numbers will
  not be his. Nothing about this lecture's outcome can be scripted in advance.
- **Footage: ~90%.** One graphic, at the end, and only if we have both numbers.
