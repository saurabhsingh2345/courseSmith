# L17 — "Day 3: Building a Kanban App with the Cursor AI Agent in YOLO Mode" (9:41)

**100% screen recording.** The first of four identical builds — same `agents.md`,
four different products. This is the template lecture for L18-L20.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | Cursor, agent panel | **Drag the panel border left** to make the agent area big — "we are going all in with our agent" |
| 0:20 | model selector | Leaves it on **Auto** — "let Cursor decide, go into the unknown". Notes many plans only offer Auto. (He later reveals Auto gave him **Composer**, Cursor's cheaper in-house model) |
| 0:40 | agent dropdown | Switches to **Plan** mode — the panel changes colour. Notes almost all these tools have a plan mode |
| 1:00 | prompt | Types only **"go ahead and plan"** — deliberately says nothing about *what*, because `agents.md` is loaded into context by default |
| 1:20 | the context meter | **Points at the circular context-usage indicator: 6.4% consumed.** Also shows "active rules" listing `agents.md` as loaded. He returns to this meter twice more (10.7%, then 28.8%) |
| 2:00 | the generated plan | **`Kanban MVP Implementation Plan`** — numbered phases (phase 4 drag and drop, phase 5 add cards…up to phase 8), an architecture overview with diagrams, suggested layout, out-of-scope, execution order |
| 2:40 | the plan, unread | **The honest joke**: the right thing is to read it carefully and give feedback — "but that's for boring people. I'm not even going to read it, but you probably should" |
| 3:00 | Build pressed | Off it goes in YOLO. `.gitignore` appears, then a `frontend` directory. "Watch it happening and enjoy the sensation of having your AI agent at work for you" |
| 4:00 | agent running, ~5 min in | "I'm enthralled." Narrates what he's seeing: **finding a problem and fixing it, almost debating with itself**, reasoning visible. Context now 28.8% |
| 5:00 | a test fails on camera | `delete test fail` — it runs tests, sees them fail, comes back and fixes them. **"This is the best way I can show you what an agent is — an LLM in a loop with tools to achieve a goal, and you can see all of those happening right here"** |
| 5:40 | it finishes | Claims tests pass and **launches the app itself**. Notes that on other runs it only printed instructions instead |
| 6:00 | the fallback | For anyone whose agent didn't launch it: terminal, `cd frontend`, **`npm run dev`** |
| 6:20 | localhost:3000 | **The Kanban appears** — Backlog / To Do / In Progress / Review / Done, nice highlighting. A Node error badge he waves off |
| 6:50 | testing it live | **Drag a card** between columns — works. **Add a card** ("this is a card / this is its description") — works. **Delete a card** — works. **Rename a column** to "not bad" — works. "It does appear to be meeting our requirements" |
| 8:00 | the defects | Honest about what's broken: the Next.js **error badge**, drag and drop is **janky**, and **you can't reorder within a column** |
| 8:20 | feedback prompt | Types real iterative feedback — mostly working nicely, but there's one error at the bottom of the screen, drag and drop is janky, can't reorder within a column, can it be more slick — **"also it would be nice to have more yellow and purple on the screen"** |
| 9:00 | second result | It "declares victory — but we'll be the judge of that." **Colours: fixed and good.** Reorder within a column: **now works.** The error badge: **still there** — he thinks it fixed it mid-run then reintroduced it, "I don't know how that happened" |
| 9:41 | wrap | "Good enough for me. It's done a fine job." Recaps the four things covered: writing a good `agents.md`, YOLO mode, giving feedback, iterating |

## For our version

- **This is the shape to copy for all four builds** — same file, same prompt, then
  test the four requirements on camera one by one. It makes the comparison fair and
  it gives every build the same beats.
- **The context meter is the single best teaching prop in Day 3.** He glances at it
  three times; we should make it a recurring on-screen callout tied back to L11's
  context diagram — same idea, now with a real number climbing.
- **Keep the failures.** The janky drag, the error badge that came back, "I'm not
  going to read the plan but you should" — this is what stops the lesson being an
  advert. Our house rule already says make the viewer feel the limitation.
- The live agent narration (**"it's debating with itself"**, a test failing and
  being fixed) is our best chance to prove L10's definition with footage instead of
  a diagram. Ramp the waiting at 12-14x and hold on the moments.
