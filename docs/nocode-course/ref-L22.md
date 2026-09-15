# L22 — "Day 4: YOLO Mode: Choosing the Right LLM for Agentic Coding in IDEs" (9:06)

**Slides only.** Opens Day 4 — a **blue day**, meaning a projects day. Half recap,
half the most practical model-selection advice in Week 1.

## The recap — four IDEs (0:00-4:00)

| product | what it actually is |
|---|---|
| **Cursor** | Made by **AnySphere** (not "Cursor" — "you don't hear that so often"). A **fork of VS Code**, which Microsoft open-sourced |
| **GitHub Copilot** | Not a fork — **an extension inside the real VS Code** |
| **Codex** | Also used as a **VS Code extension**, though that is **not the most common way** — the CLI is, and that's next week |
| **Antigravity** | Google's own **fork of VS Code**, a separate application |

## The recap — four models, and which one each build actually got

This is the honest accounting he promised in L20:

| model | maker | class | context | notes |
|---|---|---|---|---|
| **Composer** | AnySphere | fast frontier | 200k | Cursor's own proprietary model. **Almost certainly what Auto gave him** — and Cursor deliberately **won't tell you** which model Auto picked, so they can switch freely |
| **Claude Haiku 4.5** | Anthropic | fast frontier | 200k | Smallest and fastest of the new Anthropic models. Probably what Copilot's Auto used — "might have been Sonnet 4.5" |
| **GPT 5.2 Codex** | OpenAI | top frontier | **272k** | A version specially trained for use in Codex. "Confusing because Codex is in both places" |
| **Gemini 3 Pro** | Google | top frontier | **1M** | Flash is the fast sibling; both have the million-token window. Perhaps not as strong as GPT 5.2 Codex, but the context length gives other benefits — "neck and neck" |

**The framing that matters**: two separate decisions. Which **IDE** you're
comfortable in — "they're much of a muchness, basically agentic platforms that let
you repeatedly call an LLM in a loop to achieve a goal with tools." And **which
LLM** — "that is the big decision, and that is where you want to spend your time
and choose your money wisely." Most IDEs can reach most models; **Composer is
Cursor-only**.

## The rules of thumb (4:00-9:06) — the payload

1. **Always favour a smart model over a fast one.** "I would rather wait and let it
   give an accurate, high-quality answer than let it rattle something off and find
   it's broken." Favouring intelligence over speed is usually **faster overall**
2. **Set a budget, then buy the most intelligent model that budget supports.**
   "Trying to save money with a cheaper model usually ends up costing more because
   of the extra iterations and the time drain"
3. **Smaller and free open-source models need much more precise prompts** — "think
   like you're setting a detailed spec" — and **significantly more oversight**
4. **Do not YOLO with fast frontier or low-end open-source models.** And explicitly
   *not* mainly for safety reasons: the real risk is **time and productivity** —
   "you'll let it go, come back later, and just have a pile of nonsense"
5. **With fast frontier models, work in baby steps** — approve as you go, **eyes on
   the diffs**
6. **With top frontier models — GPT 5.2 Codex, Gemini 3 Pro, and "of course Claude
   Opus 4.5, my favourite of all" — you can YOLO and get great results**

## For our version

- **The IDE-vs-model split is the clearest single idea in Day 4** and it fixes the
  most common beginner confusion. Animate it as two dials: pick a chassis, pick an
  engine — and show that the engine is the one that changes the outcome.
- The four-model table is the honest footnote to Day 3's comparison. Ours must
  carry it, with **our own models named and dated on screen**.
- **Rule 4 is the counter-intuitive one worth dwelling on**: don't YOLO with cheap
  models — not for safety, for wasted time. That reframe is worth a beat of its own.
- All slides, so this is a graphics-heavy video in our cut. Pair it with L23
  (also slides) and put them either side of footage-heavy material so the section
  doesn't develop a dead patch.
