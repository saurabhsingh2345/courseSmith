# L85 — "Week 3 Day 4: Multi-Agent Team Build — Live Trading Dashboard with Claude Opus" (13:31)

**The payoff, and the best footage in the course.** A seven-agent team builds the
whole platform in about half an hour, and it works on the first run. Also the
lecture with the most quietly useful lesson about permissions.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | Claude | **He is being asked to approve things, and he explains why he did not YOLO.** No `--dangerously-skip-permissions`: *"it's good that I get to approve these things… I want to know what kinds of commands are being run, particularly in this kind of chaotic setup."* If he had wanted YOLO, he'd have done it **in the sandbox from Day 2** |
| 1:00 | Claude | **It is running serially, and that surprises him.** DB engineer finishes SQL → back-end engineer starts. Shift up/down now flips team lead / back-end / front-end. *"It's a bit more serial than I was expecting"* — he expected everything to fire at once |
| 1:50 | talk | **And the honest read on it:** serial *"is a more structured, organised way of doing it that probably leads to better outcomes"*. He could have prompted for more parallelism — *"but you'd probably get worse outcomes"* |
| 2:30 | Claude | Back-end engineer **complete: 121 tests passing**, marked done, **shut down**. LLM engineer unblocks. Muses that back-end and LLM engineer could have been one agent, but separate contexts have value |
| 3:20 | Claude | **Watching for the skill.** He wants to see whether the LLM engineer actually uses the **Cerebras skill** from Week 2 to route through OpenRouter. *"We'll be able to tell from whether it knows how to assign it to Cerebras"* |
| 4:00 | Claude | **Four agents live at once** — team lead, DevOps, front-end, LLM. *"Tons going on."* Tool calls streaming; **front-end engineer at 93,000 tokens** |
| 4:40 | Claude | **The permissions lesson, and it's the best 60 seconds of practical advice in Day 4.** Two `chmod` requests → **he presses 1 (this time only), not 2 (always)**: *"I'm happy for it to change ownership so I can execute them, but I want to approve that every time."* A test command → **2**. *"Making good decisions about whether to approve only in this case or always approve is important"* |
| 6:00 | Claude | **Docker wasn't running — so the agent launched Docker itself.** *"That was my mistake, I should have opened Docker. It just launched Docker and started it. Good for it."* He flags it as **something Opus 4.6 can do that older models could not** |
| 7:00 | Claude | LLM and front-end done, waiting on DevOps, then the integration tester. **He names his worry in advance**: *"if the integration tester finds problems, how is it going to delegate back to the agents that have completed?"* |
| 8:00 | Claude | **`Ctrl+T` toggles the task list**, shift up/down flips teammates. Two shortcuts, stated plainly |
| 8:40 | Claude | **11 minutes of integration testing**, then: **five API response mismatches causing a front-end crash.** Team lead: *"good catch, yes, fix."* **But it hands the fix to the integration tester, not the front-end engineer** — *"I was hoping the original front-end developer would fix it."* **He deliberately does not intervene**: *"I'd rather just let it be, let it complete, see where it gets"* |
| 9:30 | Claude | **Done. Six agents, six tasks, teammates shut down gracefully, team cleaned up.** A minor non-blocking CSS overlap noted. `scripts/start_mac.sh` exists |
| 10:00 | browser | **THE PAYOFF.** `localhost:8000`. **A live trading workstation**: market-data watch list on the left, **flashing prices**, an AI assistant panel, a portfolio, **$10,000 cash** |
| 10:40 | browser | Buys **JP Morgan** → appears in the portfolio, **heat map cell goes from zero to negative to positive** as fake prices move. Buys **Google** → *"look at that, the cell just moved across"*. **Portfolio P&L chart live** |
| 11:40 | browser | **The AI assistant.** *"Please add SPY to the watch list"* → added. *"Please buy one share of V"* → **executed at market price, appears on the heat map.** Box size = portfolio weight in dollars |
| 12:30 | browser | *"Please give me some trading advice"* → *cash-heavy, small position in Google, consider rebalancing.* Then **"please do this for me"** → it **sold the JP Morgan shares**. His verdict: *"I don't think that was a great move… but there we go, it's working."* Two honest quibbles: **the price chart doesn't fill in**, and P&L looked static until it moved |
| 13:00 | terminal | Closes and reopens — **state persists**. `scripts/stop_mac.sh`. `git status` → `git add .` → **he reviews the staged list on camera for anything that shouldn't be there**: no node_modules, `.dockerignore` present, **`.env.example` but no `.env`** |
| 13:20 | terminal | `git commit -m "agent teams v1"` → `git push origin agent-teams` → **`git checkout main` and it all vanishes.** *"Thanks to the power of git… bam, gone"* |
| 13:31 | end | |

## For our version

- **100% footage. This is the showcase lecture of the entire program.** The
  running dashboard with flashing prices is the single most sellable frame we
  will ever shoot — **hold on it, wide, uncut, with the numbers moving.**
- **Shoot the app at a real resolution and let it breathe.** No zoom-outs to fit
  a slide beside it. If a beat needs a label, lower-third it over the footage.
- **Keep the 1-vs-2 permissions beat verbatim in spirit.** It is concrete, it is
  the difference between a safe and an unsafe setup, and every other lecture
  only gestures at it.
- **Keep the Docker self-launch.** It is the best unscripted moment in Day 4 and
  it costs nothing — just don't have Docker running when we roll. **And never
  film Docker Desktop's container/image list** (real client work).
- **Keep the integration-tester-fixes-it-itself observation AND the decision not
  to intervene.** Watching an instructor let a suboptimal pattern play out is
  rarer and more useful than watching one save the demo.
- **`/usage` bookend.** L84 recorded 0 / 8 / 2. Put the closing number on screen
  here as the same lower-third. **This is half of the cost comparison that L87
  and L91 depend on** — if we skip it, the GSD comparison has no baseline.
- **The git-review-before-commit beat stays** and matters more for us: it is where
  a student learns to check no `.env` went up. **Our staged list must be clean of
  anything identifying** — check before rolling, not after.
- **Our result will differ from his and that is fine.** Narrate *our* screen. Do
  not pre-write "the price chart is broken" — write the result narration after
  the shoot. (This is exactly the mismatch that bit us on W1 L17.)
- **Footage: 100%.**
