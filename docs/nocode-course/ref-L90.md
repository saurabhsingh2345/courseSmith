# L90 — "Week 3 Day 5: Gastown's Parallel Polecats — Claude Code Agents Build in Chaos" (11:15)

**All footage, all chaos.** One natural-language sentence to the mayor, and eight
Claude Codes build the entire project from nothing, in parallel. The most
visually spectacular lecture in the course.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | Claude | **The one prompt.** *"Build the entire project as specified in `spec.md`, file the issues, create a convoy and sling the work to polcats."* All the lingo in one sentence — *"that's all the terminology it needs to know how to do"*. It could be done with individual commands; natural language is easier |
| 1:00 | Claude | **Beads pour out.** Ten issues filed, `fin-` prefixed, *"a bit like a git issue or a Jira ticket"*. Still setup, no action. *"I'm excited for action"* |
| 1:50 | Claude | **"Creating a convoy and slinging phase one to polcats."** Work is about to start |
| 2:20 | tmux | **The reveal, and it is the shot of the lecture.** *"The crazy thing about Gastown is you're never quite sure what's going on. It looks like not much is happening… but it's not quiet."* **`Ctrl+B` then `W`** — and up comes **a grid of tmux panes, every one a separate Claude Code**: the mayor, **witness**, and polcats. Plus a **convoys screen**: some running, some idle waiting for work |
| 3:30 | Claude | *"Check on the polcats."* Status comes back per-worker: **backend foundation closed and merged, polcat `rust` already nuked, `fin-wmw` in the merge queue, polcat `chrome` pushed its branch and is waiting to submit its MR.** *"As long as the mayor knows what's going on, I guess that's okay by me"* |
| 4:20 | talk | **The crucial difference, and it must not be missed.** *"The challenge we've given it is a bigger, harder challenge than yesterday, because we haven't included the market data implementation. This is completely blank. We've really had to do everything from scratch."* **Gastown starts from nothing; teams and GSD both started from the market-data foundation** |
| 5:00 | Claude | **Phase two: all six beads slung at once.** `Ctrl+B W` again — *"plenty going on… I'm not entirely sure what all these different windows are doing, but I trust that stuff is happening. The thing about Gastown is you just go with it"* |
| 6:00 | dashboard | The web UI. Polcats by name: **chrome, fury, guzzle, nitro, rust, shiny** — *"all doing stuff as they work in parallel. This is definitely the most parallel that we've ever had."* **The contrast, stated cleanly**: GSD was slow, careful, one at a time, five hours; **this slings everything to a ton of workers at once** |
| 6:50 | talk | **The second thing he didn't tell it**: *"we also didn't tell it about OpenRouter and structured outputs. All I gave it was the specification. I don't know if this is all going to hang together"* |
| 7:20 | Claude | **The mystery error.** A daemon *"won't stay up"*. Rather than debug it, **he asks the mayor what it's for and what breaks without it** — answer: minimal, the witness still runs, we only lose multi-project watching, and there's only one project. **Asking the orchestrator to triage its own infrastructure is the technique here** |
| 8:20 | dashboard | Dashboard says something is stuck; it's stale. Everything is *"working feverishly"* — chrome, dust, fury, guzzle, nitro, refinery, rust, shiny, thunder. **The red one is the dead leftover rig** |
| 9:00 | Claude | **Status update from the mayor**, per polcat: front-end watches and charts done, Docker done, back end done, front end done, others in progress or being fixed. **Eight things in parallel.** His comparison: *"in the Claude agent team we had two things on the go at any one time; GSD felt very serial. This is all out, everything on the go"* |
| 10:00 | Claude | **The merge conflict beat, and it is the best technical moment in Day 5.** Eight polcats done coding, **five items in a merge queue**, and **the refinery** — *"merged three branches, resolved a conflict"*. He names it: *"a classic software engineering problem when you have eight different workers on the same codebase"* |
| 10:50 | Claude | Polcats nudge the refinery about new MRs. *"Everything is flowing."* Then: **finished, all merged into main**, in a repo that started **completely empty except for the spec** |
| 11:15 | end | *"Well, I'm gonna give it a try"* |

## For our version

- **100% footage, and shoot it for the spectacle.** The `Ctrl+B W` grid of live
  Claude Codes is the money shot of Week 3 Day 5. **Full screen, hold it, let the
  panes move.** Do not cut away for a diagram — the L89 town graphic already did
  that job.
- **The "started from absolutely nothing" beat has to land loud**, because L91's
  verdict depends on it. **Put it on screen as a lower-third** — teams and GSD got
  the market-data foundation, Gastown got a spec. Otherwise the comparison is
  unfair and the student won't notice.
- **Keep the merge-conflict resolution.** It is the one thing here that no other
  orchestrator in the course had to solve, and it is the honest answer to "what
  actually breaks when you run eight agents on one repo".
- **Keep "ask the mayor what the broken daemon is for."** Delegating triage of the
  orchestrator to the orchestrator is a genuinely new technique and it takes ten
  seconds of screen time.
- **Keep the going-with-the-flow tone.** He is bewildered and says so. Sanding that
  off would make our version dishonest — and the bewilderment *is* the lesson
  about where the top of the ladder actually is.
- **Clean rig before rolling** — no stale second rig, no red dead refinery. His
  leftover confuses two lectures; ours should not.
- **Watch the token spend.** Eight parallel Opus agents is the single most
  expensive thing in the program after GSD. **Set the cap before we start and put
  `/usage` on screen at the end** so Day 5 joins the same cost comparison.
- **Footage: 100%.**
