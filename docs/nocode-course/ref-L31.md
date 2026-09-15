# L31 — "Day 5: Building Step-by-Step with AI Copilot: Planning & Scaffolding" (11:43)

**100% screen recording.** The disciplined counterpart to Day 4's YOLO. This is the
lecture that demonstrates the method the whole week has been arguing for.

## The 10-part plan in `docs/plan.md` (0:00-3:30)

He wrote the high-level steps himself, and explains the choice: **you could have the
LLM write the ten steps, and that's very common — but he is opinionated and wanted
guardrails.** "If you have opinions on this, I recommend you do too."

| part | what it is |
|---|---|
| **1. Plan** | **Enrich this document** — plan each part in detail, sub-steps as **a checklist for the agent to tick off**, with **tests and success criteria** for each. Also create an `agents.md` inside `frontend/` describing the existing code. **Ensure the user checks and approves the plan** |
| **2. Scaffolding** | Docker infrastructure, FastAPI back end, start/stop scripts in `scripts/`. Must serve example static HTML **and make an API call** — a hello-world running locally |
| **3. Front end** | Serve the existing demo Kanban board |
| **4. Sign-in** | The fake user sign-in experience |
| **5. Database model** | Set up the schema, **document the approach, get user sign-off** |
| **6. Back end** | API routes to read and change the Kanban; test thoroughly; create the database if absent |
| **7. Wire them together** | Front end calls back end — move cards, log out, log in, **and it persists** |
| **8. AI connectivity** | Back end can make an AI call via OpenRouter — **test it with two plus two** |
| **9. Extend the plumbing** | Send the board plus a question, get an answer and potentially a change |
| **10. The widget** | A beautiful sidebar supporting full AI chat, **letting the LLM change the board** |

**The reasoning he gives for step size** is the transferable part: *"if any one of
these steps doesn't work, I have a really good sense of how to dig in and figure out
why. Each step should be something I know how to dig into."*

## Part 1 executed (3:30-6:30)

- Everything is checked in to GitHub first — **"one of the crucial points I'm going
  to be making is that when you work this way you're always checkpointing"**
- Model: **GPT 5.2 Codex** — "it made the Kanban front end, we should let it keep going"
- **The opening prompt, and he flags it as a technique**: *please review `agents.md`
  and the plan and proceed — and **let me know if you have any questions. Do not do
  any work yet.*** — "asking it to ask me questions is a great way to start"
- **It asks three sensible questions**: enrich `plan.md` with checklists, tests and
  success criteria? Create the frontend `agents.md` now or after? Any **minimum
  coverage target**? He answers yes, yes, and **80% unit coverage plus robust
  integration testing** — a decision that comes back to bite him
- Result: a detailed plan with **checkboxes**, a Dockerfile step, minimal readme
  notes, **`uv` used inside the container**, a health endpoint, listed success
  criteria per part. He reads it and approves — **"we're looking for any signs of
  anything we don't like"**

## Part 2 and the two lessons (6:30-11:43)

**Lesson 1 — it claims done without testing.** Part 2 reports as complete, but he
is suspicious: *"did you test part two yourself?"* → **"No, I did not run tests
yet."** He then prescribes: run tests thoroughly, bring up the server, check the
routes, bring it down, **"let me know when you are confident."** He names it: *"a
great example — it just wanted to move on."*

**Lesson 2 — it quietly ignored an instruction.** He spots a **`requirements.txt`**
appearing when the plan said **`uv`**. *"I'm suspicious. I am suspicious."* He lets
it run and comes back to it in L32.

**On approving commands**: he's been pressing allow while reading each one. His
advice for people who can't judge them — **ask ChatGPT to explain what's going on,
get a second pair of eyes**, or deny and ask the agent to explain itself first. And
the reframe: for a newcomer this is **"an amazing learning opportunity to inquire
and see what it takes to build this kind of software."**

**On step size**, addressing the obvious objection: *"it feels like we're going in
such small steps after we YOLO'd yesterday — why can't I just tell it to do all ten?
… **this is too big a deal. If we do all ten steps it will go off the rails.** You
could always try if you want to be bold — you can always go back."*

## For our version

- **The ten-part plan is the most copyable artefact in Day 5.** Show it as an
  animated checklist that ticks off across L31-L34 — the viewer should always know
  which part we're on.
- **"Ask me questions before you do any work" is a one-line technique that changes
  outcomes** and it belongs on a card. Same for **"let me know when you are
  confident"** instead of "is it done".
- **Keep both failures.** Claiming done without testing, and silently swapping `uv`
  for `requirements.txt`, are exactly the behaviours the week has been warning
  about — arriving on camera, unplanned.
- The "why not all ten steps at once" objection is what our audience will be
  thinking. Answer it as directly as he does.
