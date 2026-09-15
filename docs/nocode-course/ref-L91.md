# L91 — "Week 3 Day 5: Gastown vs Claude Agent Teams vs GSD — Multi-Agent Orchestrators" (11:19)

**The failure, the fix, the verdict, and the fourth contender.** Gastown's first
run is a total blank page; he refuses to debug it and hands it back. Then the
three-way comparison lands, and he pivots to Codex sub-agents inside a sprite.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | talk | **"I was wrong to be too confident."** The page could not be displayed. **Total fail. Nothing there** |
| 0:30 | Claude | **And the technique that matters:** *"I didn't try and debug it, didn't diagnose it myself. I just said to the mayor: there's nothing there, please go ahead and fix it. Sling your tasks, get the polcats involved."* **Eight polcats fire again.** Second pass |
| 1:10 | talk | **The credit where it's due, restated.** It started from **absolute nothing** — the others had the market-data foundation and its docs. **A total build from scratch, massively parallel, by far the fastest of anything he's run** |
| 2:00 | talk | **The terminology finally clicks, and he explains the two he fumbled.** **Refinery** = the special worker that handles merge requests — *"the reason it was idle for a long time is there were no MRs to merge"*, then it kicks into gear. The **red refinery** was the dead leftover rig. **Witness** = the process that watches everything and notifies the mayor if anything goes wrong |
| 3:10 | browser | **Port 8002 — we do have a product.** *"It looks remarkably similar to our other products"* — and the honest reason: **the spec was precise, and it's always Claude Code, always Opus 4.6, building all of these** |
| 3:50 | browser | **This one says "select a ticker from the watch list"** — and clicking one populates the chart. *"Maybe that's what I was meant to do in the other one"* — **a quiet resolution of the blank-chart bug that appeared in both Day 4 builds** |
| 4:30 | browser | Buys **3 Apple** → heat map. Buys **Netflix** → heat map updates, positions listed. **And the honest caveat**: *"we can't really call it a zero shot this time — it took two iterations. But I didn't give it any particular feedback, I didn't help it troubleshoot. It figured all this out for itself"* |
| 5:20 | browser | **The chat.** *"Please buy three shares of JPM"* → **JP Morgan appears, green, updating everywhere.** *"The AI chat works. This is great"* — **including the market data simulator and interface, all built from scratch** |
| 6:00 | slide | **THE THREE-WAY VERDICT — the summary the whole week builds to.** A line from most controlled to most crazy: **GSD** — SDD family, clear spec, careful markdown files, plan → review → execute → validate → audit; *"reliable, disciplined, lends itself to bigger projects, human in the loop, measurable predictable outcomes… more reliable, more slow"* (**5 hours**). **Claude Agent Teams** in the middle — experimental, *"for sure it is faster"* (**~30 min**). **Gastown** at the far end — preset structure, ready for a lot of concurrent work, **~30 min but it did more, because it built the market data too**, *"the most chaotic, the fastest in terms of building the most in the same amount of time… pretty impressive, if a little bit bewildering"* |
| 8:00 | slide | **Portability.** GSD works with all the coding agents. **Gastown too** — Steve stresses it is not tied to Claude Code even though he built it around it. **Agent teams is a Claude Code feature** — *"but as it happens, very similar agent features are built into the others"* |
| 8:30 | talk | **His pick, stated plainly:** *"no surprise, I most favour the ones in the middle. I like Claude Agent Teams a lot. It was the one that gave me a lot of power without feeling like I was completely out of control"* |
| 9:00 | terminal | **The fourth contender, in a sprite.** `sprite list` → the remote fly.io sandboxes → **`sprite -s finally-worker console`** to remote in. `cd finally`, `git status` → **on branch `codex`** |
| 9:50 | terminal | **Codex is preinstalled in the sprite.** `codex --yolo` — *"the equivalent of dangerously skip permissions for Claude"*. **`/experimental` → toggle on sub-agents**, and the prompt: **spawn multiple agents to parallelise the work and win in efficiency**, then *"check out `plan.md` and do this"* |
| 10:30 | terminal | **And the gotcha he had to fix**: he **renamed `claude.md` to `agents.md`**, *"because Codex expects `agents.md`, not `claude.md`. That made it work" |
| 10:50 | terminal | `cat scripts/start_mac.sh` → port **8003**. `docker ps` → **not running, so he starts it.** *"It was built in parallel, and I got to tell you, it was super fast. Faster than any of them… maybe 15 minutes end to end"* |
| 11:10 | terminal | **`sprite -s finally-worker proxy 8003`** — maps the sprite's port to his local machine so he can open it in a browser |
| 11:19 | end | |

## For our version

- **The three-way summary slide is the single most important graphic in Week 3.**
  Build it properly: **one axis, three tools, and for each — time, tokens, what it
  started from, and what it produced.** Every number ours, from our own `/usage`
  bookends in L84/L85/L86/L87/L90. **This is the frame students will screenshot.**
- **SUBSTITUTION — this is the big one for Day 5.** He runs **Codex sub-agents in a
  sprite** as the fourth contender. Per the standing rule we run **Cursor**:
  `cursor-agent -p -f` in the sprite, with Cursor's own parallel-agent capability,
  on a branch named `cursor`. **Verify on shoot day that `cursor-agent` installs
  and runs headless inside a sprite** — if it does not, the fallback is running it
  locally against the same repo and saying why. **This is the one beat in Week 3
  that can fail for reasons outside our control; check it before the shoot, not
  during.**
- **The `claude.md` → `agents.md` rename becomes ours too** — Cursor reads
  `agents.md`. **Keep the beat**, it is the same lesson: the memory file has a
  name per agent and getting it wrong silently costs you the context.
- **Keep the total failure and the refusal to debug it.** *"I didn't diagnose it
  myself, I just told the mayor to fix it"* is the most advanced move in the
  course — trusting the orchestration layer with its own repair — and it is only
  credible **because we show it failing first.** Do not shoot a clean run.
- **Keep the "two iterations, so not zero-shot" honesty.** He could have glossed it.
- **Keep "it's always Claude Code, always Opus, so of course they look similar."**
  It is the deflating, correct explanation for why four orchestrators converge,
  and it stops the student mistaking the wrapper for the intelligence.
- **The sprite proxy command is a genuinely useful takeaway** — a remote sandbox's
  port on your local browser, one command. **Hold on it.**
- **Footage: ~80%** — one slide, the three-way verdict, and it earns it.
