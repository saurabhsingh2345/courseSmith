# L11 — "Day 2: Context Engineering: System Prompts, Context Windows & agents.md" (13:00)

**Slides only.** The longest lecture in Day 2 and the conceptual centre of the
whole week — everything in weeks 2 and 3 leans on it.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | slide | "Possibly the single most popular expression in our world last year." The output is based **entirely** on the input; the model is stateless; that input is **the context**. Getting it right is everything, because it is all there is |
| 1:20 | slide | **Prompt engineering evolved into context engineering** — not just the prompt, but the tools and everything else surrounding it |
| 2:10 | **the context diagram** | The layered stack he builds up and reuses all course: |
| | | **1. System prompt** — the most general framing. What role is this LLM playing, what's its job, what tone, what approach. Sits at the very beginning |
| | | **2. Tool descriptions** — more tools means more space consumed. Notes some people call these part of the system prompt; "it's a definitional thing, it doesn't matter" |
| | | **3. Memory** — information persisting across conversations; long- or medium-term |
| | | **4. The conversation so far** — *everything*: first message, its reply, follow-ups, the model's **reasoning tokens**, any code it generated, tools it called and **the output of those tool calls** |
| 6:00 | diagram, extra line | **`agents.md`** gets its own row. Persists like memory, but special: project facts, coding standards, objectives. **`claude.md`** for Claude Code, **`gemini.md`** for Antigravity — all "agents.md files" generically. Then the self-interruption: *"I can hear you correcting me — it's not every single time, it's more complicated, we'll get to that"* |
| 7:40 | slide: context window | A hard token limit on how far back the model can look. **Exceed it and it fails — that is a break, an error** |
| 8:40 | slide | **The more important point.** Too much context degrades quality *well before* the limit — not speed, **accuracy**. You lose coherence, it starts forgetting. Best results come at the very start of a conversation when context is nearly empty. "**Less is more**" |
| 10:00 | slide: compacting | Built into Claude Code: near the limit it summarises the history and replaces it, freeing space. **Why people fear it** — you are trusting the model to judge what mattered; when it drops something you said, the agent appears to forget, or repeats a mistake |
| 11:20 | slide | His own habit: stop the agent, rewrite `agents.md` by hand, restart fresh rather than let it compact. Then he undercuts himself — **"compacting has got a lot better; as of 2026 there is less reason to fear it. Don't be like me. Trust the compactor"** |
| 12:00 | slide: window sizes | **GPT 5.2 — 400,000 tokens. Claude Sonnet 4.5 / Opus 4.5 — 200,000. Gemini (Antigravity) — 1,000,000.** Then: he doesn't obsess over max size, because performance degrades with fill across all of them |
| 13:00 | end | |

## For our version

- **The layered context diagram is the single most valuable graphic in Week 1.**
  Build it once as an animated stack that fills up layer by layer, then reuse it
  in L12, and again in weeks 2 and 3 whenever compaction or `claude.md` comes up.
- The degradation point deserves its own visual: a quality curve falling as the
  window fills, with the hard limit as a cliff at the end. He only says it.
- **"Don't be like me, trust the compactor" is the best line in the lecture** —
  an expert admitting his own habit is now superstition. Keep that shape.
- Context-window numbers go stale fast. State them as "as of early 2026" on screen
  so the frame doesn't age into being wrong.
