# L33 — "Day 5: Building a Kanban App with GitHub Copilot: Debugging Drag and Drop" (9:49)

**100% screen recording.** Parts 5 to 7, and the lecture where an agent gets
genuinely, expensively stuck.

## Part 5 — the database, and deferring to the agent (0:00-2:00)

Part 5 pauses for **sign-off**, as the plan demanded. He opens the schema doc:
**users** table, **board** (one per user, with a title stored for future multi-board
expansion), **columns**, **cards**, an **ordering strategy** and a **migration
approach**.

**The moment worth keeping**: he had assumed it would store everything as a **JSON
blob** — *"but this is possibly a better design. Maybe mine was not the most
sensible way of doing it… I probably would have done it differently, but my way
might have been a bit more hacky. This is perhaps a better way."* **He approves the
agent's design over his own.**

## Part 6 — back-end routes (2:00-2:40)

Finishes, he checks nothing is broken, front and back not yet wired. Checkpoints:
`git status` → `git add .` → `git commit -m "part 6 done"`.

## Part 7 — wiring the halves together (2:40-4:30)

He flags it as **the most significant change so far** and it takes **15 minutes**:
"tests were repeatedly failing and it rebuilt them. It had to do a lot of rewriting
of the front end."

- `start-mac` → **it does a production build**, "which is what I like to see"
- localhost:8000 → sign in → **"let's move this over here. Well, that's not a good
  sign, is it?"** — drag and drop is **janky**; some cards move, some snap back
- **But persistence works.** He moves a card, closes the browser entirely, reopens,
  logs back in — **the card is still in the second column.** "That's pretty
  impressive. This is working nicely"
- Checkpoints the mixed state honestly: `git commit -m "part 7 built, some drag and
  drop bugs"`

## **The half-hour rut** (4:30-7:30) — the centrepiece

His debugging prompt, phrased exactly the way L18 taught: *the persistence is
working, but drag and drop only works occasionally. Most of the time I drag a card,
the next column highlights, but when I release, the card goes back to its original
position. **Please test thoroughly and fix. Reproduce the problem, fix it, and
confirm it is fixed.*** — "always make sure that you're very prescriptive about
debugging."

Then the outcome nobody would script:

- **It runs for half an hour**, stuck in a loop, trying the same shapes again and
  again. **"I can see it keeps saying it's likely because of this — I keep seeing it
  jumping to conclusions, which is such a classic move by the LLM"**
- He **stops it** with the stop button and **tries the app himself**
- **It had already fixed the bug.** Somewhere in that half hour drag and drop
  started working — but **its own tests for drag and drop were failing**, so it kept
  concluding it hadn't succeeded
- **"It shows that human involvement was needed to check and confirm that actually
  it's working fine."**

He then demonstrates it working: drag, drop, empty a column, close the tab, reopen
localhost:8000, sign back in — **the state is exactly as left**. Checkpoint:
`git commit -am "part 7 working"`.

## The context reset (7:30-9:49)

The setup for Week 2, and a real technique:

- **Context has been filling up across parts 1-7.** Copilot doesn't show a meter the
  way Cursor does — **"it's handling it for us"** — but it is summarising and
  dropping messages. **"This is something we are going to obsess over when we use
  Claude Code next week. And we should obsess over it just as much nonetheless"**
- The practice, which he admits **"feels kind of galling to do"**: **stop the chat
  and start a fresh one**, letting it re-read `plan.md` to learn where things stand
- First, make the plan carry the state: *please confirm `plan.md` is up to date with
  all the latest, **including any design decisions that you made.** Let me know when
  ready.* Then Keep, `git add .`, `git commit -m "final updates after part 7"`
- Then **right-click > New Chat**. "Starting all over again"

## For our version

- **The half-hour rut is the best unscripted footage in Section 1.** An agent that
  fixed the bug and could not tell — because it trusted its own broken tests over
  reality — is the most precise argument for staying in the loop that the week
  produces. Do not compress it into a sentence; ramp the wait and hold on the reveal.
- **Approving the agent's database design over his own** is the counterweight that
  stops the day being anti-AI. Keep both beats in the same video.
- **The context reset is the bridge into Week 2** and it deserves a callback to
  L11's context diagram — the stack we drew, now visibly full, then emptied, with
  `plan.md` as the thing that survives.
- "Include any design decisions that you made" is the load-bearing half of that
  prompt. Card it.
