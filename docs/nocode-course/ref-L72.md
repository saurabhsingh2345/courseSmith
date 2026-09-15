# L72 — "Week 3 Day 2: Claude Code Sandboxing and Cloud Execution Deep Dive" (7:46)

**Slides only.** A recap of Day 1, two plugin files he owed us, then the framing
for sandboxing. One of only two pure-slides lectures in Week 3.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | slide | Today is lighter than yesterday but "really juicy stuff". Another yellow day |
| 0:30 | **the Day 1 list** | Anticipates the reaction — **"a bit overwhelmed… a bit similar to Andrej's original tweet"** — and defuses it item by item: slash commands are **"really replaced by skills almost always now"** · focus on sub-agents, teams come later · **hooks are rarely used**, look them up when you hit a problem · plugins are handy for teams, but **most of the time you'll build skills** |
| 2:00 | slide | The second objection: **"this is so Claude Code focused"**. His answer: **skills are ubiquitous now**, the de facto standard Anthropic created, used everywhere. Plugins are still Claude-only "as of now" |
| 3:00 | the plugin structure | The two files he owed us. **`.mcp.json`** — MCP servers bundled inside a plugin. **`.lsp.json`** — Language Server Protocol, so Claude Code can validate a language it doesn't already know |
| 4:20 | talk | Admits LSP is obscure and says **why** he included it: **it was named in Karpathy's tweet**, and he wants you to reach the end of the program and understand every word of that tweet |
| 5:00 | slide | The homework that matters: **actually go and make one of each** — a slash command, a sub-agent, a hook, a plugin — so you know where the docs are |
| 5:40 | **sub-agent pros and cons** (recalled) | Puts the L71 slide back up. The primary reason is **freeing context** — and that same delegation is the source of the drawbacks |
| 6:20 | slide | **Sandboxing framing.** Ring-fence resources: let Claude do its worst, but only inside this directory, only through these ten domains |
| 7:00 | slide | And the second benefit — **it makes you willing to YOLO**, because the risk is contained |
| 7:20 | slide | **Approval fatigue**, and it is the sharpest point in the lecture: pressing 1, 1, 1, 1 without reading is *"basically YOLOing, but you're giving yourself a false sense of security"*. **Unsandboxed may be less safe than you think** |
| 7:46 | end | Teases "the blue box" — remote execution |

## For our version

- **Approval fatigue is the beat to keep whole.** It reframes the whole
  permissions ladder we drew in W1 L16: the cautious setting is only safer if you
  are actually reading. That is a genuinely uncomfortable idea and it is true.
- **The LSP aside is worth keeping** precisely because he explains *why* he's
  telling us — it closes the Karpathy loop the program opened in W1.
- **This is a slides lecture, so it is the exception to the footage rule.** Even
  so: recall the W1 `ladder` for the permissions options and the `stack` for
  context rather than drawing new art. See WEEK3.md rule 2.
- His "skills replaced slash commands" line is a **correction to his own Day 1** —
  keep it, it is honest and it saves the viewer effort.
