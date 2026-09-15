# L2 — "Day 1: Building a First-Person Shooter Game with Cursor AI Agent" (6:54)

Picks up mid-build from L1. **Pure screen recording — no slides at all.**

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | Cursor, agent done | Agent created **three** files. Immediately hedges: "every time you run this it does something a bit different" — yours may be one `index.html` |
| 0:30 | clicking each file | Opens each of the three: a long file, and a stylesheet. Scrolls the agent's reasoning column |
| 1:00 | Finder, `instant` dir | Shows the three files on disk. **Double-clicks `index.html`** to launch it in the browser |
| 1:20 | the game, title screen | "Arrow keys to move and turn, space to shoot" — **Neon Arena** first-person shooter. Start match |
| 1:40 | playing | He plays, gets hit, **dies**. "I'm not much good at this" — but the point lands: a game from nothing, no work |
| 2:10 | playing again | Replays to **win**, on camera, to "redeem myself". Victory screen |
| 2:40 | agent panel | Iteration 1, verbatim: *"That's great. Please add some detail to the opponent so that it looks more like an enemy."* — "we'll let it be creative" |
| 3:10 | agent editing | Fast-forwards through the edit |
| 3:30 | game reloaded | Double-click `index.html` again. Enemy is more detailed. He wins again |
| 4:00 | agent panel | Iteration 2, verbatim: *"Please add a heads up display HUD and also make the difficulty harder."* Notes planning-then-executing will be covered properly later |
| 4:30 | game reloaded | Harder, with a HUD. "It achieved that goal" |
| 5:00 | talking over the app | Wraps the exercise: no line of code written, iterating hand-in-hand. **Recovery advice** — if stuck, delete the whole `instant` dir, start over, try a different model, try different prompts |
| 5:30 | the teaser | Same project built with **Ralph Loops + Claude Code**, **zero-shot** — one prompt, no feedback. A technique the course covers later |
| 5:50 | the good game | Double-clicks its `index.html`. Entry screen, "FIGHT". Full-size: **gun model, enemy energy, health and kills HUD, minimap bottom-right, health pickups** |
| 6:30 | still playing | Openly distracted by how good it is. "Try not to get too distracted by this" |
| 6:54 | wrap | "That concludes the instant gratification... from this point on, it's all business" |

## Numbers

- **0% graphics.** Every second is screen recording
- Two build-and-test iterations plus one zero-shot showcase
- The emotional beat is the **dies-then-wins** sequence and the **Ralph Loop reveal**

## For our version

- Covered by `nocode01` — ours ends on a working game rather than his teaser.
- **The Ralph Loop reveal is the single best hook in Day 1** and we did not use it.
  It plants week 2 and pays off later. Consider opening `nocode02` with it.
- His two iteration prompts are the model for ours: short, conversational, each
  producing a *visible* change. Matches our existing house rule.
