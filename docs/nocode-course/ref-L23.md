# L23 — "Day 4: Five Principles for Successful Vibe Coding: Be the Boss" (10:35)

**Slides only.** The ethics-and-discipline lecture of Week 1, and the one that most
directly contradicts the day's own YOLO project — deliberately.

## The five principles, under the heading "BE THE BOSS"

1. **Spend time on your `agents.md`.** Concise, and covering three things: **spec**
   what must be done, describe the **style** it must be done in, and define
   **what success means** so it is unambiguous how to judge the result
2. **Start simple.** Not just a simple first section — **your entire first
   `agents.md` should be the basic MVP.** Run that whole iteration through,
   including testing at each point, make sure it works, *then* decide to go one
   step more complex. "If you try to boil the ocean with a super complicated
   application spelt out in your agents.md, you might go off the rails and get stuck"
3. **Work incrementally**, constantly testing and validating the success criteria at
   each step. Disciplined enough that at every point you know you hit the milestone
   and can go back if you didn't
4. **Don't get lazy.** The trap arrives *after* early success — it built a shooter,
   it built a Kanban, "oh wow, this just works, I'll just YOLO." Your defences come
   down, you go several steps forward, and only then find that **several steps ago
   it made a terrible assumption that is complete nonsense.** When it declares it
   found and fixed a bug "with lots of confident bold text" — **challenge, demand
   the evidence, demand the root cause, demand the test results**
5. **Expect the fiasco.** There will be moments you're blown away, and there will be
   chaos. His own escalation, told as a story: it makes a mistake, you put "never
   ever make this mistake" in `agents.md`, it makes it again; you make it **write a
   performance report promising never to do it again and read that report every
   time** — and it still does it. "I've had moments when I have been just so angry."
   The lesson: **that is part of the process.** You're hitting the model's limits;
   detect it, manage it, handle it with style

## The contradiction, named out loud (6:00)

He raises the objection himself: these five principles fly in the face of YOLO.
"And you're right. And you could probably tell **I'm not a huge fan of YOLO**."
There are situations for it, but always with a sceptical eye. **Trust but verify** —
"even with the really strong models, you need to be the boss."

## Words for each audience (6:30-8:30)

**If you're new or junior:**
- **Use this as a learning opportunity.** Don't lose sight of it — challenge and
  question the LLM so you understand what's happening and still make the journey
  to senior. Otherwise "you just never acquire that skillset, which will only lead
  to frustration later"
- **Be sceptical.** The mental model: a **super-eager assistant** that researches
  and does things, but is sometimes misguided, sometimes overconfident, **uses
  band-aids instead of proper solutions, and guesses a lot**

**If you're senior:**
- **This tooling suits you best** — "it's harder for the junior engineers, sorry
  junior engineers, because you can feel bamboozled and you can't tell if it's
  making it up. As a senior engineer, you can tell"
- **Keep enjoying what you do.** The honest trade: the joy of hammering out a line
  of code is being taken away, but in exchange **he can now build entire systems he
  was terrible at** — front end, Terraform, DevOps. "It's given me extra powers I
  didn't have before, and building faster is fun"

## The balanced verdict (8:30-10:35)

- **Genuine 10x**: the Next.js Kanban view — "days, maybe a week, versus minutes"
- **Incremental only**: larger projects, legacy projects, complex backend changes
- **Actively negative**: his real example — building a **streamable HTTP MCP
  server**, too new to be in the training data, "Claude was messing it up for me
  again and again, and it took me longer than if I'd just written it myself"
- **The warning about estimates**: you have early success, you cut your estimates
  to a tenth, then you hit a stumbling block and everything is pushed out
- **Net**: always value-add, always faster with the top agents — but the size of the
  boost depends on the task. **Small greenfield MVP → a very significant
  multiplier. Most other cases → assistance, not the 10x of the hype**

## For our version

- **"Be the boss" is the thesis of our whole course** for a no-code audience.
  Make the five principles one animated card set that we can call back to
  by number for the rest of the 16 hours.
- **Principle 4 pairs exactly with L18's four-step debugging rule.** Show them
  together — the principle and the exact words to type.
- The junior/senior split needs care: our audience skews toward the junior end, so
  give that half more room and drop the "sorry junior engineers" framing.
- **Keep the MCP-server failure.** A named case where the agent made him slower is
  the most credible thing in the lecture and it costs 20 seconds.
