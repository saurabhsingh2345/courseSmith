# L9 — "Day 2: How LLMs Work: Tokens, Memory, and Reasoning Explained" (11:02)

**Slides only, zero footage.** First lecture of Day 2, which he flags up front as
a "purple day" — a theory day.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | curriculum grid, purple stripe | "You survived day one." Two warnings, delivered as **bad news**: today is mostly him talking, and it is foundational. If you know it, **put me on 2x** |
| 1:00 | slide: what is an LLM | Predicts what text follows an input. "**Autocomplete on steroids**" — an enormous pattern-matching statistical engine trained on internet-scale data |
| 2:00 | slide: tokens | Not words, **tokens** — a chunk of a few letters, often a word, often part of one. Output is not the next token but **the probability of every possible next token**. Worked example: "two plus two is" → very high probability on `four`, tiny probability on `bananas` |
| 3:20 | slide: inference | One token at a time, whole input re-passed each step. Walks "what is the capital of France?" through → `the` → `capital` → `of France is` → `Paris`. That process is **inference**. Points to his YouTube playlist in resources |
| 4:30 | slide: LLM vs AI application | The distinction he wants held: **GPT is the LLM, ChatGPT is the application** wrapped around it (memory, web search, other functionality). Other examples: the Cursor agent, **Duolingo Max**'s chat-to-learn feature, **Atlassian Rovo** |
| 5:40 | slide | Since ChatGPT in 2022 we've all been surprised that predictive text on steroids looks intelligent. **There are four tricks behind it.** Two are covered here, two in L10 |
| 6:20 | **trick 1 — the illusion of memory** | An LLM is **completely stateless**. Say "I'm Ed", then "who am I" as a fresh call → "I don't know". But ChatGPT remembers, because **the entire conversation so far is passed in every single time**. That is the whole trick |
| 8:30 | **trick 2 — reasoning / thinking** | Started as "please think step by step" giving better outcomes. Models were then trained to emit their working before the answer. **The two-coins demo**: toss two coins, one is heads, chance the other is tails? Without reasoning it answers *a half*; with reasoning it answers **two thirds** and says "this is probably a trick question" first |
| 11:02 | end | |

## For our version

- Everything here is a diagram waiting to happen and he shows it as static slides.
  **Three animations carry this lecture**: tokens streaming with a probability
  bar per candidate; the conversation re-sent in full on every turn (the memory
  illusion, which is genuinely hard to see in a still); the two-coins tree.
- **The two-coins demo is the one thing to run live on screen**, not draw. It is
  the cheapest real footage in a lecture that otherwise has none.
- Keep the "put me on 2x" honesty — it is the same register as our own
  "what this did not show you" beat.
