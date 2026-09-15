# L88 — "Week 3 Day 4: GSD vs Claude Agent Teams — Side-by-Side UI Comparison & Wrap-Up" (8:18)

**The verdict lecture.** The GSD build runs, gets tested the same way, and then
**both apps go on screen at once**. Closes Day 4.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | terminal | `scripts/start_mac.sh`. It runs |
| 0:20 | browser | **It works.** And a nice detail: **the portfolio isn't empty — it remembered the previous run's data**, because the database directory is local. *"This looks really cool. What do you think? I think this is probably a better display"* |
| 1:00 | browser | *"It's amazing that it looks so similar, actually."* Buys **3 Apple** → onto the list and the portfolio. Buys **10 Meta** → **cash drops, portfolio jumps, Meta becomes a big slice** — **but the heat map isn't colouring.** *"The heat map looks a bit inferior, but generally speaking it is working"* |
| 2:20 | browser | **The AI assistant.** *"Hi there"* → hello. *"Add IBM to the watch list"* → **thinking… still thinking… added.** *"We're not getting the trademark speed from Cerebras"* |
| 3:10 | browser | The compound instruction: **"buy one share of IBM and sell five shares of Meta"** → **it executes both.** IBM up to 1, Meta down to 5, portfolio updates. *"It is working"* |
| 3:50 | browser | **The same defect in both builds.** The price chart is blank — *"it's funny that we have the same problem in the other UI as well"*. Heat map still not colouring. Otherwise: *"pretty impressive, the chat is working really nicely"* |
| 4:30 | editor | **Did it use our skill?** He digs into the code: **it is calling Cerebras, but not with our code** — it went via **LiteLLM** and a different structured-output approach. His read: *"it's read our skill but taken it in a different direction… this sometimes happens when it does what we suggest, hits a bug, and the way it fixes the bug is to rewrite and do things differently."* **Stayed true to the intent, not the implementation** |
| 5:30 | terminal | Git check. Commits have been landing all along. **Spots a mistake: `node_modules` from the test directory is not git-ignored.** *"That seems like a bit of a mistake"* |
| 6:00 | browser | **"Wait a second — I couldn't resist."** **Both UIs side by side.** Neither squashes well horizontally, but there they are |
| 6:20 | browser | **Agent teams (30 minutes):** the janky bit, **and a real bug — it can't get prices for anything not on the watch list.** But *"the UI is really nice… very professional, and it hasn't fallen prey to that problem where LLMs tend to give everything purple backgrounds"* |
| 7:00 | browser | **GSD (5 hours):** looks very similar, **heat map not coloured**, less dramatic — **but it doesn't have the watch-list bug**, it handles new tickers, adds to the watch list and still prices them. *"Maybe it did a little bit better from a quality point of view"* |
| 7:30 | talk | **The verdict, and he hands it to the students.** *"It's your call — put it in the Q&A which one you think is better. But for me, I'm going with the first one"* — despite the defects, **because it got the whole thing done in half an hour**, and the defects are quick to fix. *"Two beautiful products"* |
| 7:50 | talk | **Day 4 wrap.** *"For you it's been an hour. For me it's been a whole day of toiling with GSD."* **He understands why people love it** — on a really big project, in YOLO mode, left alone, it is *"very, very diligent… clearly very thorough indeed"*. But **a lot of time and a lot of tokens.** The two ends of the spectrum: **dynamic agent teams vs regimented GSD orchestration**, two fabulous results. **93% of the course** — the rest is tomorrow |
| 8:18 | end | |

## For our version

- **The side-by-side is the Day 4 closing frame** and it must be shot properly:
  two browser windows, same size, same data, same time of day. **He apologises for
  the squash — we can just do it right** with a wider capture. This is one of the
  few places where beating him costs nothing.
- **Our verdict has to be our own.** He picks agent teams. **We report whatever our
  two runs actually produce, with our own numbers**, and we say the call out loud
  the way he does — with a reason, not a shrug. If ours lands the other way, that
  is a *better* lecture, not a broken clone.
- **The "same bug in both builds" observation is the sharpest thing here.** Two
  completely different orchestrators, same brief, **same missing chart**. That
  points at the brief, not the tool. **Name it: an ambiguous spec produces the
  same gap whatever machinery you point at it.** He gets close; we should land it.
- **Keep the skill-drift finding.** "It read our skill and took it in a different
  direction" is the honest limit of skills as a control surface, and it is a
  direct callback to Week 2's Cerebras skill. **Show the actual code on screen.**
- **Keep the `node_modules` catch.** Two lectures in a row where reviewing the git
  staging area catches something wrong is the point, not a coincidence.
- **Keep "no purple backgrounds."** It is a throwaway line and it is the most
  quoted-back kind of observation in a course like this.
- **Footage: ~95%.** No slide is needed; if we build the cost/time graphic it lives
  in L87, not here.
