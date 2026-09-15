# DEEPENING BRIEF — Week 3

You are adding narration to lectures that ALREADY HAVE narration, to use footage
that is currently running with nobody talking over it. You are not rewriting.

## The job

For each lecture assigned to you:

1. `python3 fit.py <lec>` — prints, per segment, `footage Xs / words Ys /
   room for N more words`. That N is your budget for that segment. Hit it
   within ~10%. Segments with room for fewer than 25 words: leave alone.
2. Read the existing `narration/<lec>.json`. Every `say` is already correct and
   already paid for — **do not reword, reorder or delete a single existing
   sentence.** You only INSERT new sentences.
3. **Look at the frames** in `frames/<lec>/` before writing a word for a
   segment. Directory names are `MMmSS_<beat>`. Use the Read tool on the jpgs
   covering the segment you are writing. This is mandatory: narration that
   describes a screen nobody looked at is the one defect that has cost this
   project whole lectures.
4. Write the deepened file to `narration/deep/<lec>.json` — the SAME structure
   (`lecture`, `cut_seconds`, `segments` with `anchor`, `beat`, `say`), same
   anchors, same beats, same order. Only `say` strings change, and only by
   gaining sentences.
5. Re-run `python3 fit.py <lec>` mentally against your new word counts; do not
   overshoot a segment. Overshoot causes a freeze frame.

## Where new sentences go

Insert INSIDE the segment, usually after the first one or two sentences (which
name what appeared) and before the closing line (which hands off to the next
beat). Never append everything at the end.

## What the new sentences must be

Real teaching, tied to this picture:

- **The why under the what.** The existing line says what the agent did; yours
  says why that is the right or wrong move, and what it costs.
- **Read the screen.** Numbers, filenames, counts, error text that is visible in
  the frames — say them out loud.
- **The question the student is asking right now**, then the answer.
- **The honest note.** If the agent chose clumsily, or the output is thinner
  than it looks, say so.
- **What transfers.** The rule the student takes to their own repo.

## What they must NOT be

- Anything not true of the frame it sits over. No invented files, counts,
  errors, or UI.
- Padding: "as you can see", "let's go ahead and", "what we're doing here is",
  restating the previous sentence, announcing what you are about to say.
- Emojis, exclamation marks, "amazing", "powerful", "seamless", hype.
- Forward references to lectures or tools that are not in the week. The week is
  Claude Code, sub-agents, hooks, plugins, the container sandbox, the Agent SDK
  orchestrator, agent teams, `gh`/GitHub CI, headless `claude -p`, and two built
  products (Control Tower, Signal Desk). **Cursor, Copilot, Antigravity, Cowork,
  Claude Code on web/phone, GSD, Gastown and Sprites are NOT in this week** —
  never mention them.
- Second person plural marketing voice. It is "you", level, direct.

## Voice

Vary sentence length: a long explaining sentence, then a short one that lands
it. No throat-clearing. Read `docs/nocode-course/w3/narration-rules.md` once
before you start.

## Done means

`narration/deep/<lec>.json` exists, is valid JSON, parses with the same segment
count and anchors as the original, and every segment is at or just under its
room budget. Report per lecture: words added, and any segment you left alone.
