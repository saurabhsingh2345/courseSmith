# L80 — "Week 3 Day 3: Driving Claude Code Programmatically with Claude Agent SDK" (9:08)

**Sizzle #1 of three.** Almost pure footage — an empty folder becomes a playable
Space Invaders, driven by ~15 lines of Python. The best footage-to-slide ratio in
Week 3.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | talk | Names the segment **"the spicy part"** — demos of adjacent things not fully covered. *"You don't need to take all of this in."* Permission to just watch |
| 0:30 | talk | **The misconception, stated and killed.** The Claude Agent SDK is **not an agent framework**, despite the name. It is **a way to programmatically drive Claude Code** — all the agent power, reached from code instead of a terminal |
| 1:10 | VS Code | A project called **`space`** — *"a completely empty, totally empty directory with nothing in it whatsoever"*. The empty explorer is the shot |
| 1:30 | terminal | `uv init --bare` · `uv python pin 3.13` · `uv add python-dotenv requests claude-agent-sdk`. **Names the rename**: it used to be `claude-code-sdk`; it is now `claude-agent-sdk` — **agent, singular** |
| 2:30 | editor | Copies in a `.env` and writes a `.gitignore` *"even though we don't have a git repo — still good practice"* |
| 3:00 | editor | `main.py`. `load_dotenv(override=True)` — **override so `.env` beats a stale system env var, "otherwise trouble"** |
| 3:30 | editor | **Fights the autocomplete on camera** — *"stop, stop, stop filling in"*. He wants to type it himself |
| 4:00 | editor | The prompt: **"make a vanilla HTML + JS + CSS website for a game of Space Invaders"** — that's it, plus *write the code to files in the current directory including `index.html`*. Jokes that he promised only one frivolous build — *"well, sue me"* |
| 5:00 | editor | `tools = [...]` — read, write, edit and friends. `options = ClaudeAgentOptions(allowed_tools=tools, model=...)`. Scrolls the parameter list: **permission_mode, mcp_servers, model** |
| 6:00 | editor | Picks **Opus** as the model and immediately warns against it: *"you should not do this… use a cheap model. I'm splashing out for your entertainment"* — **especially in a tight loop** |
| 6:30 | editor | `async for message in query(prompt=prompt, options=options): print(message)` then `asyncio.run(main())`. **That is the whole program** |
| 7:00 | terminal | `uv run main.py`. Messages stream back, **a file appears in the empty explorer**. *"We are driving Claude Code using code"* |
| 8:00 | browser | Double-click `index.html` → **Space Invaders**. Enter to start. **It has sound. Arrow keys work. There's a score. There's the alien.** *"Absolutely brilliant"* |
| 8:45 | talk | The honest footnote: *"the astute amongst you will note this is probably in its training data"*. Still a working game from one sentence |
| 9:08 | end | |

## For our version

- **This is the footage-heaviest lecture in Week 3 — shoot it as close to 100% as
  we can get.** Empty folder → typing → stream → game. Four shots, no diagram
  needed until the very end.
- **The "it is not an agent framework" beat is the one idea in the lecture.** One
  graphic, early: terminal-you and code-you both pointing at the same Claude Code.
  Then never draw again.
- **Substitution: none needed.** The SDK is Anthropic's; the lecture is already
  Claude-only. This one is a straight clone.
- **Build something better than Space Invaders?** No — *keep* the game. It reads
  instantly on screen, it makes noise, and the callback to Day 1's frivolous build
  is the joke. Our "bigger project" energy goes into Day 4/5, not here.
- **Do the cheap-model warning as its own held frame.** He says it in passing;
  it is the single most expensive mistake a student can copy from this lecture.
- **Narration must describe the frame**: "the folder is empty", "the file list
  just grew by one", "that's the score in the top corner" — not a retelling of
  what the SDK is while a game is playing.
- **Footage: ~85%.**
