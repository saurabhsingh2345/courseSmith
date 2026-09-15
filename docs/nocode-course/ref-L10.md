# L10 — "Day 2: Tools, Loops, and the Definition of AI Agents" (8:28)

Continues straight on from L9's four tricks. **One real screen demo** in the
middle; the rest is slides.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | slide | **Trick 3 — tools.** The output tokens don't have to be an answer; they can say the model *wants to do something* — search the web, use a calculator, run Python. You run it, then call the LLM again with the result |
| 1:30 | slide | **The grounding point he insists on**: the LLM never searches anything. It is a statistical engine emitting tokens. **Our software interprets those tokens and runs the tool.** He returns to this repeatedly |
| 2:30 | **ChatGPT, live** | Pastes a prefix — *to use Python code to respond to the next question, just reply `Python:` and a Python expression* — then asks **"what is the square root of pi?"** |
| 3:10 | ChatGPT reply | It declines to answer directly and **emits a one-line Python expression, imports included**. "A super clear example of what it means to have an LLM use a tool." Tool calling is a clever input plus interpreting the output |
| 4:20 | slide | **Trick 4 — the loop.** What's better than calling an LLM once? Calling it again. Produce output, ask if it's finished, if not call again, repeat until the goal is met. "So simple it's super elegant" |
| 5:00 | slide | Those four tricks are what get you from ChatGPT 2022 to Cursor Agent and Claude Code today |
| 5:30 | slide: definitions over time | The meme that an AI agent could be whatever you wanted. **(1) OpenAI-era**: systems that do work for you independently — GPT Agent, formerly Operator, booking a restaurant while you watch. Or simply "an LLM that can act" |
| 6:40 | slide | **(2) Early 2025, Hugging Face**: systems where **an LLM controls the workflow**. Anthropic doubled down in the post **Building Effective Agents**. Prevailing definition for a long time |
| 7:20 | slide | **(3) Late 2025, Simon Willison — the one that stuck**: *an LLM that **runs tools in a loop to achieve a goal***. Ties straight back to tricks 3 and 4 |
| 7:50 | callback | Proves it against yesterday's build: we gave the Cursor agent a **goal** (a first-person shooter in a web page), it was clearly in a **loop** (files appearing one after another), and it had **tools** (write files, run code, generate code). "That gave you first-hand experience of an AI agent" |
| 8:28 | end | |

## For our version

- **The definition-history beat is the spine** and it lands because it converges
  on our own lesson-1 footage. Ours can do this better than his: we still have
  `~/arena` and the nocode01 take, so the callback can be **actual footage of our
  agent looping**, not a description of it.
- Run the square-root-of-pi demo live. It is short, visual, and it is the only
  moment in Day 2's first half with a real screen.
- The "the LLM never actually runs anything, our software does" point is the most
  commonly misunderstood idea in the lecture. Give it a dedicated animation:
  tokens out → our code reads them → our code calls the tool → result back in.
