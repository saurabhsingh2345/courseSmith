# L12 — "Day 2: Mastering agents.md: Context Window Strategy for Coding Agents" (11:52)

**Slides only.** Practical companion to L11. He shows a real `agents.md` on a slide
rather than in an editor.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | slide | "I warned you this was going to be all me talking." Invites disagreement in the Q&A — "nothing I like more than good debates" |
| 0:40 | slide: markdown primer | `.md` is markdown. One hash = h1, two = h2, three = h3; hyphens for bullets, numbered lists. "A text file with some useful extra characters" |
| 1:40 | slide | **Why markdown specifically**: LLMs were trained on enormous amounts of it, so they love reading and writing it. Write in natural language — concise, crisp, assertive, low ambiguity, **high signal per word** |
| 2:30 | slide | It is squashed into the context window, so **this is precious space. Every word counts** |
| 3:00 | slide: the hierarchy | One in the **project root** (yesterday: `instant`). But also one in **any subdirectory, at any nesting depth**. A nested one is read only when the agent works on files there; it then walks **backwards through parent directories** collecting the rest. **Inner overrides outer** — that is how you write a specific rule that beats a general one. This is the promised answer to "it's not included every single time" |
| 5:00 | slide | Naming: `agents.md` for Cursor / Codex / Copilot, **`claude.md`** for Claude Code, **`gemini.md`** for Antigravity. Cursor also has **rules**, but "the trend is everyone converging on agents.md" as the de facto standard |
| 6:00 | **his real agents.md** | What he puts in one: overall **project goals and success criteria**; a **checklist** it is forced to tick off; links to other documents it can optionally load; **coding standards**. Be concise, be specific, and **correct for problems you've hit with the agent before** |
| 7:30 | the same file, annotated | His actual gripes, verbatim in spirit: agents **overcomplicate everything** → "simpler is better"; comments only when necessary; **short readmes** — "LLM-generated readmes are just the worst"; **no emojis**; **IMPORTANT in block capitals works**, "who knew"; avoid over-defensive programming; avoid `isinstance` checks; only handle exceptions when necessary — "they love to put tries around everything"; `uv run` never `python3` |
| 9:00 | slide | Backticks: one for inline code, three for a block |
| 9:20 | slide | **Focus on positives.** LLMs are strangely incoherent at remembering what *not* to do. Say what it should do; use "never" sparingly. He then admits his own file breaks this rule |
| 10:00 | slide: 2025 vs 2026 | **The 2025 school**: your success comes from sweating `agents.md` — one at the root, more in subdirectories, supporting planning and success-criteria files, continually pruned and rewritten, agent stopped and restarted with fresh context. Hard work, and it was the way |
| 11:00 | slide | **The 2026 mindset**: let it hang out, give up the reins, focus on the end goal, use skills, loops, agents, sub-agents, swarms; let it self-correct |
| 11:30 | slide | **The confession**: "I am still of the 2025 school of thought." For toy projects like yesterday's shooter he lets go; for larger work he is "all over it". Expects the field to move his way through 2026 |

## For our version

- **Show a real file in a real editor, not on a slide.** This is our easiest win in
  Day 2: open `agents.md` in the editor, scroll it, highlight lines as they're
  discussed. He shows a static image; we can shoot it.
- The **nested-hierarchy rule is the one genuinely confusing idea** here and a
  static slide cannot carry it. Animate the directory tree, walk the agent into a
  subfolder, show which files get pulled and in what order, show inner beating outer.
- His do-and-don't list is worth keeping almost line for line — it is the most
  concrete, immediately usable content in the whole day. Reword, keep the specifics.
- Keep the **2025-vs-2026 split and the confession**. An instructor saying "the
  field has moved past my habit and I haven't" is the credibility beat of the day.
