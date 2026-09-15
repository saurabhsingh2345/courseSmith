# L89 — "Week 3 Day 5: Gastown — Orchestrating Swarms of Claude Code Agents" (13:03)

**Opens the final day.** A recap, the Yegge chart one last time, then the whole
Gastown vocabulary and install. Half slides, half terminal.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | slide | *"I can't believe it. It's the last day."* Day 4 recap: agent teams (seven or eight agents, fast, exciting, nerve-wracking) then GSD (five hours, checks and double-checks). **Both zero-shot.** And a revision of his own verdict: *"the more I think about it, the second one didn't have any major defects, where the first one definitely did"* |
| 1:40 | slide | Recalls **the sub-agents vs agent-teams table** and **the GSD flow** — plan, execute, verify |
| 2:20 | browser | **Both apps on screen again.** Heat map switching, the chat, watch list, positions, portfolio P&L. *"These charts are terrific."* And he takes the purple jab back: *"maybe I'm being unfair about the purple background — after all, that is one of our brand colours"*. **Both zero-shot, no debugging, huge amounts of code.** *"Absolutely extraordinary"* |
| 3:40 | slide | **The Yegge eight stages, final appearance.** Stage 6 was where we'd been. **Stage 7 we never actually did** — we never brought up ten agents and hand-managed them; **we jumped straight to stage 8**, *"which is the right way to do it"* |
| 4:30 | slide | **And the reason the chart exists**: Yegge drew it to introduce **Gastown**, his own product, which takes stage 8 to an extreme. **The spectrum**: GSD (disciplined) — agent teams (out there) — **Gastown "way over there by a mile"** |
| 5:10 | browser | **The Gastown repo, and an unusually strong disclaimer.** *"This next segment is completely optional. I recommend you watch me do it… but be warned, it's a bit chaotic, it's a bit of madness. If it doesn't work out for you, just move on"* |
| 5:50 | browser | **What it is.** A **workspace manager** — *"really it's just a fancy set of fabrics built around Claude Code or any coding agent"*; it even calls itself a new IDE construct. **Solves manual coordination of many agents** — Steve says **20 or 30**. A hierarchy, **mail between agents**, structure for *"controlled chaos"*. He'll run six to eight |
| 7:00 | browser | **The lingo warning.** *"Take a paper and pen, write them down, get your glossary ready."* **Beads** — Steve's own infrastructure underneath, gives issues an ID, the **beads ledger** |
| 7:40 | browser | **The vocabulary, one at a time.** **Town workspace** = `~/gt`, outside your projects folder. **Rig** = a project — his is **`fin`** (and an older broken one called `finally` still hanging around). **Crew** = a workspace in the rig — his is **`ed`**. **Polcat** = a worker agent, assigned tasks. **Convoy** = a bundle of beads slung to one polcat |
| 9:40 | browser | **Mayor.** *"As far as you're concerned, feels like a Claude Code."* **All of these are Claude Codes** — tons of them running — **but you only talk to the mayor.** There's a way to flip and see the others. Coordinator, mailbox, *"tons of stuff, and I'm scraping the surface… my full understanding is fairly surface level, as you can tell"* |
| 11:00 | browser | Install: **`brew install gastown`** on Mac, one command on PC — **better in WSL with tmux** if you're on Windows. And again: *"you don't need to do this. You can just watch me"* |
| 11:40 | terminal | **iTerm2.** `gt` init in `~/gt` with git → done. `cd` in. **`gt rig add fin`** + a GitHub repo called `fin` → set up |
| 12:20 | terminal | **`gt crew add ed --rig fin`** → `cd fin/crew/ed` — **the repo is cloned here, containing a single file: `spec`** |
| 12:40 | terminal | **`gt mayor attach`** → **Claude Code launches.** It looks around: *"no hooked work, no mail, inbox is empty, waiting for instructions"* |
| 13:03 | end | |

## For our version

- **The glossary is the whole risk of this lecture.** Five invented words in three
  minutes, over a scrolling readme, is exactly the "dragging" failure mode. **Build
  one graphic and one only for Day 5: the town.** Rig contains crew contains
  polcats; mayor to one side; refinery and witness in the corner; beads as tickets
  flowing along the arrows. **Then reuse that one frame every time a word recurs**
  in L90 and L91 rather than re-explaining.
- **Keep the "completely optional, this is madness" disclaimer verbatim in spirit.**
  It is the honest framing and it protects the student from a bad afternoon.
- **Keep "they are all Claude Codes."** That single sentence is what makes Gastown
  comprehensible, and it makes Day 5 a Claude lecture rather than a third-party
  tour.
- **Keep his admission that his understanding is surface level.** An instructor
  saying so, on camera, about a tool he is demoing, is worth more than polish.
- **The Yegge chart returns for the last time — recall it, do not redraw it.** Light
  stages 7 and 8 and make the point he makes: *we skipped 7.*
- **Shoot the install and setup as one continuous terminal take.** Four commands,
  clean prompt (`PROMPT='%~ %# '`), from `/Users/Shared/projects`. **Our rig gets a
  neutral name** and **no stale second rig on screen** — his `finally` leftover
  causes confusion for the next two lectures.
- **Footage: ~65%** — the recap and glossary are unavoidably slides; everything from
  11:40 is pure terminal.
