# L76 — "Week 3 Day 2: Third-Party Cloud Sandboxes: Running Claude Code on Sprites.dev" (10:26)

**Entirely screen recording.** The `@claude` result lands, then the third approach.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | GitHub | **The `@claude` issue finished in ~7 minutes** — faster than writing the document took. Ticked every to-do, claims **90+ comprehensive tests** |
| 0:50 | GitHub | It pushed a branch. **Create PR** → **21 files changed, 1,600 lines**. He skims: *"I don't have time to go through this carefully, but I'm pleased to see… not unwieldy functions"* → **create → merge → confirm** |
| 2:00 | slide | **The blue/purple/yellow recap.** Blue = native `/sandbox`, done. Purple = Anthropic's cloud, five ways, done. Yellow = third-party, next |
| 3:00 | talk | **Marks the third approach optional** — you may be perfectly happy with Anthropic's. It matters most **if you are not using Claude** |
| 3:40 | browser | **sprites.dev**. "Stateful sandbox environments with checkpoint and restore"… "hardware-isolated execution environment… a persistent Linux computer". **$30 of trial credits, ~500 sprites free** |
| 4:40 | signup | **It wants a credit card** — he flags it honestly: *"which I know for some of you is like, uh-uh"*. Then **confusingly bounces to fly.io's landing page**; you have to navigate back to sprites.dev yourself |
| 5:30 | terminal | The install command **contains your key**, so **he can't show his screen** — you copy yours from the dashboard. Then **`sprite login`**, browser popup |
| 6:20 | terminal | **`sprite create finally-worker`** → **"Created in 0.6 seconds"**. Prompt changes to `sprite@sprite` — you are on the remote box |
| 7:00 | terminal | `ls` — empty. **`git clone`** the repo, `ls`, `cd finally` — the whole project is there, in the cloud |
| 7:50 | terminal | **`claude`** — and **it comes pre-installed on every box**. Dark mode, then the login: it needs a browser, **on a remote machine**. They handled it — press 1 and **it opens the browser on YOUR computer**, authorize, back to the remote |
| 9:00 | terminal | **"Warning: Claude Code running in bypass permissions mode"** — Sprites **automatically configures YOLO**, because the box is disposable and isolated |
| 9:30 | terminal | Pro detail: it launched an **older Claude Code on Opus 4.5**. Quit and relaunch → **auto-updates to the latest**. "That's only because it's a fresh install on this box" |
| 10:00 | terminal | Sets it working: read the planning docs, code-review the market data back end, run the tests, write `market-data-review.md` |
| 10:26 | end | |

## For our version

- **`sprite create` in 0.6 seconds, then `claude` already installed, is the shot.**
  Two commands and you have a disposable Linux box running a coding agent. Pure
  footage, no graphic can improve it.
- **Keep the credit card warning and the fly.io redirect.** Both are small
  frictions a viewer will hit and neither is in the marketing.
- **PII: the install command embeds your API key** — he could not show his. Ours
  must not either. Type it off-camera or blur; the beat is "paste the one from
  your dashboard".
- **The remote-login-opens-your-local-browser beat is genuinely surprising** and
  worth holding.
- **Auto-YOLO on a disposable box is the whole argument of Day 2** — it connects
  straight back to W1 L16's Run Mode ladder. **Recall that graphic for a few
  seconds**; the rest stays footage.
- **Cost is real but small**: ~$0.02/hour, $30 trial. Worth stating on screen.
