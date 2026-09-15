# L27 — "Day 4: Tutorials, Code Reviews with Opus, and Cross-Model Collaboration" (8:20)

**100% screen recording.** The day's redemption arc: having broken every rule, he
comes back and applies them.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | recap of off-camera work | He gave feedback on the prompts and the chat UI. **It got stuck** — it changed the UI, the colours came out wrong, he told it to fix it, no difference, **back and forth four or five times.** "This is the kind of thing that can get very frustrating" |
| 0:40 | **the unsticking technique** | Instead of pushing harder: **"don't do it that way. Remove what you did and build a completely different chat user interface. Try again."** It did, and it looked great. **The general rule: when you're stuck in a rut and it keeps making the same mistake, changing the approach is easier than diagnosing what's going wrong** |
| 1:30 | `npm run dev`, the site | Digital twin now: **good formatting, polished, a thinking indicator, then a proper streamed response.** Notes it's slow because it's all on his local machine. **The scroll bug is fixed**, and **the whole LinkedIn profile is now in the prompt**, so the answers are more rigorous. "Plenty of room for improvement, but for a few minutes of vibe coding, a great start" |
| 3:00 | **"and now for my master stroke"** | Ctrl C. New prompt: *write me a comprehensive tutorial in Markdown suitable for a **complete beginner in front-end coding**, walking me through what you've done. Include a summary of the technology, a high-level walkthrough, a detailed code review with code samples, and end with **five suggestions for how the code could be improved, based on a self-review**.* — **"I am basically having it appraise itself"** |
| 4:00 | `tutorial.md` | Right-click > Open Preview. *Professional Website and Digital Twin — Beginner Tutorial*: summary, walkthrough, code structure, code samples, styling, front end, back end, how to run it locally. **"It's pretty minimal, but it's a start"** — and the instruction: **iterate on it, say "build this out, give me more", until it's pitched at the level you need**. It is both a way to learn and **a way to challenge and test your agent** |
| 5:30 | **cross-model code review** | The Cursor feature he wants to show: **switch models mid-conversation.** New prompt: *please do a comprehensive code review of this project and write the results to `review.md`, including any remedial actions needed. **Don't actually change any code.*** Sends it **not to Codex but to Claude Opus 4.5** |
| 6:20 | the thread | **The whole context, conversation and work goes to Opus** — "what we're seeing in the thread is not Codex anymore, it's Opus." A different trained model, a different perspective, on the same project |
| 6:50 | `review.md` | **"I just love Opus."** Super comprehensive — dependency analysis, action plans, prioritised remedial actions. **The critical one: the `.env` file** ("this is probably a real problem"), plus other good advice |
| 7:30 | the loop closes | What to do next: **go back to Codex and say "read this document and implement the remedial actions, or say if you disagree."** Each LLM works on a piece of the puzzle. **Foreshadows sub-agents and multi-agents doing this in real time in later weeks** |
| 7:50 | wrap | The recipe: **have a model write a tutorial, drill down until you understand it, have a different model do a code review, then remediate.** "That is a great way to get good quality out of vibing on YOLO" |
| 8:00 | close | If it's not working: **be patient, simplify, simplify, simplify** — strip back to a simple site, get it working, add pieces. If you love it, ask the model to write a deploy to-do list for something like **Vercel** — "it would actually deploy it for you too, but I recommend just get it to write you the to-do list". **"We broke all of the rules, but we didn't break them totally, because at the end we came back and looked in at the code and checked it."** **27% complete.** Tomorrow, another blue day |

## For our version

- **This is the best lecture in Day 4 and arguably in Week 1**, because it turns
  "I can't read code" into a workflow: **have it teach you, then have a rival model
  audit it.** Two prompts, both reusable forever. Make them cards.
- **The unsticking technique deserves its own named beat** — "throw it away and ask
  for a different approach" is the single most useful thing a beginner can learn
  when a build stalls, and he gives it 40 seconds in passing.
- **Cross-model review is a genuine wow moment** and it is cheap to shoot: same
  conversation, different brain, better answer. Hold on the model switch.
- Note the **callback discipline** — Opus flags the `.env` problem, which is
  exactly what he set up in L25. Our cut should make that rhyme explicit.
- Ends Day 4 with the grid + **27%**. Fourth reuse of that graphic.
