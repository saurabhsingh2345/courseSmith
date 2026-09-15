# L34 — "Day 5: AI Assistant Kanban App Complete: Copilot, OpenRouter & Week 1 Wrap" (11:47)

**Mostly screen recording**, closing with the week-one wrap. Parts 8, 9 and 10 —
the AI assistant — and then the assignment.

## Part 8 — AI connectivity, on a fresh chat (0:00-3:00)

The first prompt into the new chat: *please read `agents.md`, then read the plan,
and **let me know any questions before we start part 8**.*

**It asks three, and one of them catches a real problem:**
1. Should the two-plus-two integration test be **fully mocked**? → **"No mocking.
   I want it to go all the way to OpenRouter. Let's check we can get an answer from
   the model."** He adds that he'd noticed the mocking in the plan earlier and
   disliked it — **"so I'm super happy that it asked"**
2. `/api/ai` or `/api/chat`? → chat
3. Configurable or hard-coded? → hard-coded for now

Then: **"please note these answers in `plan.md` and proceed."**

**On the fresh context** — the payoff of L33's reset: *"there's something comforting
about knowing we've got a completely fresh chat. It's not completely empty — it's
got `agents.md` and the plan — and that means we're making really efficient use of
the space. It's not cluttered with all that stuff about worrying about drag and
drop, which would have hogged room. **We're positioning our agent for success.**"*

**The honest verdict on the reset**, both halves: *"I just experienced the good and
the bad of restarting the conversation. It seemed **faster and more accurate** —
almost like it decluttered its brain. **But it also got stuck** trying to run tests
because it had forgotten the right way to start the server, and it took five minutes
of thrashing around."* His fix: **have it update its own `plan.md` so that's clear
for the future.**

Checkpoint, commit, part 8 done.

## Part 9 — and the mocking problem again (3:00-4:00)

Part 9 reports done, **but it had written mocked tests again** and he had to prompt
it to actually test against OpenRouter. It claims it now has. **"We won't know for
sure until we've added part 10."** Commit.

## Part 10 — the AI assistant (4:00-7:00)

*"Do you think it's going to work?"*

- `start-mac`, localhost:8000, sign in — **the board appears with an AI assistant
  chat on the right**
- **"hi there"** → *hello, how can I help you with your board today* — **"there's an
  AI at the other end of this. We have an AI assistant app. People pay big money for
  this stuff"**
- **"Please summarise my project for me"** — fast, and accurate about the columns
- **The moment**: *"please move the **Gather customer signals** card from Backlog to
  Done"* → **it moves on screen. "Bam. Do you see that? It moved. The magic of AI"**
- Proves persistence properly: new tab, localhost:8000, sign in fresh — **the card
  is still in Done**

What they have: a Kanban board, **persisted to disk**, with a **database**, running
in a **Docker container**, with a **front end and a back end**, and an **AI
assistant that can reorganise the project.** "It is an app."

## **"This is not the end, this is the beginning"** (7:00-9:00)

The honest assessment, which is the best part of the lecture. What still needs doing:
resize the columns, show the chat thinking, stream results, add real users, add
multiple boards.

Then he opens the code and **criticises his own result**: the main Python module,
`main.py`, is **"a bit of a mess… this definitely is something of a disaster.
It's shoved everything into one Python module which has many different concerns.
It seems like a major gaffe that needs to be fixed."** What he'd do: **have another
agent do a code review** — "it would definitely point this out" — then restructure
before anything else. Which is exactly the L27 technique, applied to a real defect.

## The assignment and the wrap (9:00-11:47)

**The assignment**: take it further in whatever direction you want. Add users, fix
the UI, restructure the back end, move to a remote database like **Supabase**, or
ask it for **step-by-step deployment instructions** — once it's in a container it
deploys easily to **Vercel, AWS App Runner, GCP Cloud Run** — "it can even do the
deployment itself." Then message him and say what you built.

**The final Karpathy fragment**, held back from L28 on purpose: agent capabilities
**crossed some threshold of coherence around December**, causing a **phase shift in
software engineering** — and **the intelligence part now feels quite a bit ahead of
everything else around it**, the integrations and workflows and processes. Which is
precisely what weeks 2 and 3 are about.

**Week 1 recap**: a first-person shooter, a personal website with a digital twin,
and the MVP of a real commercial product. **Week 2** is the transition from vibe
coding to **vibe engineering** — Claude Code, in the CLI, "in anger". **"You are 33%
of the way through."**

## For our version

- **"This is not the end, this is the beginning" is the right ending for our video
  too** — and our house rules already require an honest "what this did not show
  you" beat before the outro. His is better than ours: he opens the actual file and
  calls his own output a disaster.
- **The AI moving a card on command is the single best demo moment in Section 1.**
  Hold on it. Everything in Week 1 has been building to a visible, unmistakable
  payoff and this is it.
- **The mocking catch (part 8) rhymes with the coverage trap (L32) and the drag-drop
  rut (L33)** — three separate instances of an agent satisfying the letter of an
  instruction. Our Day 5 cut should name that pattern once, explicitly, and let the
  three examples land under it.
- Ends on the grid at **33%** — the fifth and final reuse in Section 1.
