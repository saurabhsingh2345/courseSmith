# Week 3 build plan — "Vibe Engineering as an expert"

Section 3, lectures **65-94** (95 is a cross-sell article, cut like L5).
**30 films, 5h 11m.** Reference transcripts pulled 2026-08-28.

## The four rules for Week 3, stated once

Given by him on 2026-08-28. These override anything the reference does.

1. **BEAT-BY-BEAT CLONE.** Same beats, same order, same substance. Our wording,
   our footage, our project. Never compress his coverage.
2. **FOOTAGE HEAVY. VERY LITTLE ANIMATION.** Week 3 is a products week and it is
   meant to be the best part of the program. Target **85-95% real screen
   recording** per lecture. A graphic only earns its place when it shows
   something the screen genuinely cannot — the chaos/control spine, the context
   stack, the eight stages. Recall graphics we already built in W1 rather than
   drawing new ones.
3. **THE NARRATION DESCRIBES WHAT IS ON SCREEN.** Not "random saying". Every
   line is about the frame it sits over. Result narration is written **after**
   the shoot. See [[coursesmith-narration-matches-picture]].
4. **CLAUDE AND CURSOR ONLY.** Everywhere the reference reaches for Codex or
   ChatGPT, we use **Cursor**. Verified substitution:

   | his | ours |
   |---|---|
   | `codex exec "<prompt>"` | `cursor-agent -p -f "<prompt>"` |

   `cursor-agent` is installed at `~/.local/bin/cursor-agent`, logged in, and
   `-p` is its non-interactive print mode with full tool access — the exact
   equivalent of `codex exec`. **`-f` is required** or it refuses to run in an
   untrusted directory (which is itself a teaching beat about permissions).

   This is a *better* substitution than it looks: the whole point of his beat is
   "a different LLM, from a different vendor, reviewing your work". Claude Code
   reviewed by Cursor keeps that intact using two tools we already pay for.

**And the project gets bigger.** He asked for a better, more ambitious capstone
than his FiNALLY trading workstation. Same architecture discipline — one
`plan.md` every agent converges on, hard boundaries between areas — but a
larger product.

## The days

| day | lectures | runtime | spine |
|---|---|---|---|
| D1 | 65-71 | 72:26 | Claude Code pro features: slash commands, sub-agents, hooks, plugins & marketplaces |
| D2 | 72-77 | 60:53 | sandboxing, Claude Code on web/mobile, GitHub integration, Sprites.dev |
| D3 | 78-82 | 45:47 | large codebases, Claude Agent SDK, Cowork, OpenClaw |
| D4 | 83-88 | 67:38 | **Agent Teams**, GSD, the multi-agent dashboard build |
| D5 | 89-94 | 64:35 | **Gastown** swarms, the three-way orchestrator verdict, final workstation, wrap |

## Verified about the tooling (checked 2026-08-28, not taken from his lecture)

- **Gastown** is real — Mayor orchestrates, Polecats execute in parallel, Witness
  and Deacon monitor, Refinery merges. MCP mailboxes, scales to 20-30 agents.
- **Sprites.dev** is real — Fly.io stateful sandboxes, Claude Code preinstalled,
  **$30 trial credits, ~$0.02/hour**. Cheap enough not to be a blocker.
- **`cursor-agent`** installed and authenticated. See rule 4.
- `claude` 2.1.251, model `opus[1m]`. The kickbacks.ai ad injection is removed and
  locked — see [[coursesmith-no-pii-in-renders]]. **Re-verify before every Claude
  Code take**, because the ad renders into the status line of every session.

## Beat sheets — ALL 30 WRITTEN

`ref-L65.md` … `ref-L94.md`. Every one carries: a beat table (timecode · what is
on screen · what happens) and a **For our version** block naming the
substitutions, the PII hazards, the footage target and what must not be
pre-scripted.

Footage targets per lecture, from the beat sheets:

| lecture | footage | note |
|---|---|---|
| 65-71 (D1) | 70-95% | L67/L68/L70/L71 are all-terminal |
| 72-77 (D2) | 75-95% | L74/L76 are pure setup footage |
| 78 | 55% | fourth consecutive recap — compress it |
| 79 | 10% | **the one slides lecture** — the "vegetables" |
| 80 | 85% | empty folder → playable game |
| 81 | 90% | synthetic receipts only |
| 82 | 90% | needs a Telegram token |
| 83 | 20% | **the second slides lecture** — agent-teams theory |
| 84-85 | **100%** | the seven-agent build and the live dashboard |
| 86-88 | 85-95% | GSD, the 5-hour grind, the side-by-side |
| 89 | 65% | Gastown glossary |
| 90 | **100%** | the tmux grid of eight Claude Codes |
| 91-92 | 80-95% | the three-way verdict, the capstone |
| 93 | 35% | deploy + best-practices summary |
| 94 | 60% | montage of our own footage |

**Week 3 average lands around 78% footage** against the rule-2 target. The four
slide-heavy lectures (79, 83, 89, 93) are the ones where the *content* is advice
or vocabulary; everywhere else we are at or above 85%.

## Still open, needs him

Nothing here blocks Day 1. Ordered by when it bites.

1. **Telegram bot token** (L82, Day 3). BotFather, one throwaway bot. Never
   appears on camera — same shrink-the-window trick he uses.
2. **A market-data API key** (L92, Day 5) — the only genuine *improvement*
   available. Free tier is enough for 60 delayed tickers. **Default without it:
   our own simulator, which looks identical on camera and costs nothing.**
3. **Token budget for Day 4 + Day 5.** GSD alone cost him ~10x an agent-teams run
   and five hours of wall clock; Gastown runs eight parallel Opus agents. These
   two days are the most expensive thing in the program by a wide margin. **Set
   the cap before we start and I will put `/usage` on screen at both ends** — the
   three-way comparison in L91 depends on having our own numbers, not his.
4. **Claude plan must cover Claude Code on web and mobile** (L73, L75).
5. **Your sign-off for L94** — where students should reach you, and whether we
   ask for ratings. His close is personal to him and cannot be cloned.

### Decided, not open

- **L92 "Codex Wins"** — resolved by rule 4. Cursor is the fourth contender, and
  **the verdict is written after the four builds are on screen, never before.**
  The lecture is retitled so the winner is not in the title.
- **Deployment** (L93) — Fly.io free tier, neutral app name, torn down after the
  shoot unless you say keep it.

## Standing risks carried into the shoot

- **`cursor-agent` inside a sprite (L91)** is the one beat that can fail for
  reasons outside our control. **Verify before the shoot day, not during.**
  Fallback: run it locally against the same repo and say why on camera.
- **GSD and Gastown are fast-moving third-party projects.** Re-verify install and
  command names on the day. If a flow changed, we shoot the new flow and say the
  docs moved — that is better footage than matching him.
- **Agent teams may have graduated out of experimental** (L83/L84). If so the
  settings-flag beat becomes *"you no longer need this"*, which is also better.
- **Never film**: Docker Desktop lists, Cursor Settings, VS Code Workspace Trust,
  the default zsh prompt, an uncropped Telegram window, `sprite list` or any URL
  bar carrying a personal name. Shoot from `/Users/Shared/projects`.
- **Result narration is written after the shoot, every time.** Our builds will
  not produce his defects. This is the W1 L17 lesson and it applies hardest in
  L85, L88, L91 and L92, where the whole lecture is a live outcome.

## How to re-pull the reference transcripts

They come from the caption files, not the player, so this takes ~2 minutes for
all 30 and needs no navigation per lecture. In the Udemy tab, with the course
open, define `__grab(lectureId)` to hit
`/api-2.0/users/me/subscribed-courses/7053781/lectures/<id>/?fields[lecture]=asset&fields[asset]=captions&fields[caption]=id,locale_id,url,source`,
pick the `en_US` caption, fetch its `url`, and strip `WEBVTT`, cue numbers,
`-->` lines and consecutive duplicates. Loop it into `window.__all` **in
batches** — 30 in one call exceeds the 45s CDP timeout (though the work does
complete). To read them out, render into the DOM and use `get_page_text`; the
javascript_tool result truncates at ~1k chars, `get_page_text` does not.

Lecture ids for all 95 are in `lecture-ids.txt`.
