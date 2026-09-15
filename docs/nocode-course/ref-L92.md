# L92 — "Week 3 Day 5: Codex Wins — Final Trader Workstation with Live Market Data" (14:25)

**The capstone.** Four builds compared side by side, a surprise winner, then the
real one: **live market data, 60 tickers, three coding agents in four tmux panes,
and a light/dark toggle.** The biggest single lecture in Week 3.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | browser | **The four-way parade begins.** **GSD** first — the five-hour build. Buys 3 Apple. **He finally works out the chart**: clicking a ticker changes the title, *"maybe it is working, we just have to leave it a bit"*. Buys GOOG through the chat — executed |
| 1:40 | browser | **Claude agent teams** — *"my favourite one to work with"*. The flashing watch list. **And he solves his own bug from L85**: *"I guess I just didn't get this the first time — you click here and this updates, and you have to wait a while for the chart"*. Buys 1 JPM through the chat |
| 3:00 | browser | **Gastown.** *"They all look quite similar… not a big surprise, it was Opus 4.6 building them all, just orchestrated differently"*, with the first one having *"maybe overworked it"* over five hours |
| 4:00 | browser | **Codex.** *"Wow, it does look different."* Same dark look, same flashing, **but the charts are clearly improved** — *"really professional, really sleek"*. Buys 1 Meta — **the portfolio updates fast** |
| 5:00 | browser | **THE DECIDING MOMENT, found live on camera.** *"Are they all doing that? Yes they are — oh, but not this one."* **The agent-teams build's portfolio has gone static.** Nor is GSD updating. **Only Codex keeps the portfolio live.** *"My favourite one — maybe not so good after all"* |
| 5:50 | talk | **The verdict, and he is visibly surprised by it.** *"To my surprise, the Codex implementation is perhaps my personal winner. A late-breaking surprise."* And the fairness note: **it is the only one that did NOT get the Cerebras skill** — *"although it seems to figure everything out"* — **and it took 15 minutes, the fastest.** He throws it to the students: *"you should tell me what you think"* |
| 7:00 | browser | **LIVE MARKET DATA.** The screen is green and barely moving — **because it is real, and it is after hours.** *"This is the real stock price of JP Morgan right now."* Occasional after-hours ticks |
| 8:00 | talk | **And it did not work first time.** The provider key went in, but **Claude's code had a wrong hard-coded key**. **Codex found it, fixed it, and then volunteered the explanation** — *"hey, you realise it's after trading hours, so it's going to be a bit static"*. **Codex was exactly right.** *"A modern Bloomberg-esque terminal with a position heat map, an AI chat, and real live market data"* |
| 9:00 | talk | **His overall call, stated carefully:** *"whilst Claude Code is my favourite platform — I love the tooling — I will say that on this occasion at least, Codex emerges as the leader of the pack"* |
| 9:40 | terminal | **"But wait, there's more."** `sprite list` → `sprite -s finally-worker console` → remoted into the cloud sandbox. **`tmux attach -t quad`** → **four terminals at once** |
| 10:30 | terminal | **The four panes, and the working setup of the whole course.** Top-left **Codex in YOLO with sub-agents on**. Top-right **Claude in dangerously-skip-permissions, able to run agent teams**. Bottom-left **a second Codex**. Bottom-right **git**. `Ctrl+B` + arrows to move. *"Three different heads that we're talking to, that can each have their own gangs of agents"* |
| 11:20 | talk | **How he actually worked for two hours.** **Market data and back end with Codex, UI with Claude**, tracking everything in markdown files, **having them occasionally review each other's work**. *"It's been quite hard work managing everything, but I've always been the boss. I've not dug into the code"* |
| 12:00 | talk | **THE MOST IMPORTANT SPEECH IN THE COURSE.** The 10× feeling, then the roadblock — *"the agents just keep hitting the same problem and can't get their way through it… you can feel like, ah, I'm wasting all my time"*. **But even counting the block, two hours is still two or three times faster than writing it himself.** *"Never be surprised when that initial 10× gets pulled back by some bugs. It's still a significant net positive."* And the method for the block: **reproduce it, write a markdown file explaining it, then fix, then prove you fixed it — and be ready to find you didn't** |
| 13:20 | browser | **THE REVEAL.** The FINally trader workstation with live market data. **60 tickers across tech, financials, healthcare, consumer, industrials and energy**, live updating. Live Apple ticker. Portfolio P&L. Heat map. LLM chat. *"Let's agree this is sensational."* Clicks Netflix → the main chart, *"having a bit of a journey downwards"* |
| 14:00 | browser | **The one more trick: a light/dark toggle Claude added.** Flips to light — *"it just looks so fresh and professional and clean"* — and back |
| 14:10 | talk | **The valuation line.** *"If you'd shown me this a couple of years ago and asked how long it would take, I'd probably have said a month or two, and I would have charged a fair amount of money for this."* Still a demo — **no user login, not deployed** — but both are small steps from a container |
| 14:25 | end | Take the repo, or build your own in a different direction with the same wow factor |

## For our version

- **NEEDED FROM YOU (only real blocker in Week 3): a market-data API key.** He uses
  a paid live feed. **Default is our simulator, which is free and looks identical
  on camera** — but the "these are real prices right now" beat is genuinely
  better. **A free tier is enough** for 60 delayed tickers. **Say the word and I'll
  wire it; otherwise we ship simulated and say so on camera.**
- **SUBSTITUTION — the title itself.** His verdict is "Codex wins". **Ours runs
  Cursor as the fourth contender, so our verdict must be our own and must be
  honest**, whichever way it falls. **Do not pre-write it.** The lecture title
  becomes something like *"The Final Trader Workstation — and a Surprise Winner"*,
  and we name the winner only after the four builds are on screen together.
- **The four-way parade must be shot in one session, same data, same window size.**
  This is the frame the whole program has been building toward.
- **The static-portfolio discovery is the best moment in the course precisely
  because it is unscripted** — he finds a defect in his own favourite, live. **We
  cannot fake this.** Shoot the parade honestly and narrate what our screens
  actually do. (W1 L17 taught us this the hard way.)
- **The 10× / roadblock speech is the emotional thesis of the whole program.** Give
  it room. **This is the one place in Week 3 where a slow, plain graphic is
  correct** — a curve that rises, stalls hard, and still ends well above the line.
  One frame, held, while he talks.
- **The four-pane tmux shot is the "what does an expert's screen look like" frame.**
  Ours is **Cursor + Claude + a second agent + git**. Hold it wide.
- **The light/dark toggle is a 15-second closer that lands enormously.** Keep it,
  and make sure our build has it — if it doesn't, ask for it on camera.
- **PII: the deployed URL and the sprite name carry his name (`finally-ed`).** Ours
  must be neutral. Check every URL bar and every `sprite list` output before we
  roll.
- **Footage: ~95%.** One graphic, the 10× curve.
