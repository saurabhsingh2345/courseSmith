# L24 — "Day 4: Responsible YOLO Coding: Setting Up OpenRouter for AI Projects" (13:36)

Two halves. **~7 min of slides/web reading** — the strongest counterweight in the
course — then **~6 min of live setup** on OpenRouter.

## Part 1 — the counterpoints (0:00-7:00)

He deliberately front-loads the case against, before the fun.

**The Anthropic study.** A blog post on **how AI assistance affects coding-skill
formation** — which he calls "clearly kind of anti-Anthropic, very balanced." They
studied junior coders using AI assistance against those not using it on the same
task, then quizzed both on the underlying library:

- **AI-assisted group: 50%** · **unassisted group: 67%**
- Productivity gain was small for this task type, but the **detrimental effect on
  learning was significant.** "A sobering point we should all keep in mind"

**The Jellyfin policy.** Jellyfin is a popular open-source media-streaming project.
He reads their AI-assisted-contribution policy on screen, at length, because he
rates the writing:

- Opening: LLMs are powerful and flexible for experienced and new developers, **but
  there are trade-offs** — a "precipitous rise" in AI contributions alongside real
  criticism
- **LLM output is prohibited in direct communication** — issues, comments, feature
  requests, PR descriptions, forum posts — **except translations**
- The stated aim: let knowledgeable developers use these as legitimate aids
  **without a flood of slop contributions that violate their core ethos**
- The code rules: contributions **concise and focused** (touching unrelated files →
  rejected); **broken into small manageable commits**; formatting and quality
  upheld; **do not commit LLM meta files** like `.claude` or `agents.md`; you must
  **review the output and explain it yourself** — "if you can't explain what the
  LLM did, we are not interested in the change"; test it; take feedback; final
  discretion with reviewers
- **The golden rule, read verbatim**: don't let an LLM loose on the codebase with a
  vague vibe prompt and commit the results as is — lazy, always poor quality,
  "we are not at all interested in such slop. Make an effort or please do not bother"
- Notably: they say **it is not their place to judge whether something was
  AI-generated**, and it doesn't matter — **what matters is the quality**

**His reading of it** — the part worth keeping: there is an **asymmetry**. It is so
easy to generate tons of LLM code and submit it, and **the weight then falls on
senior people** to digest it and separate signal from noise "through endless
readmes with lots of emojis." **That's not fair, and it needs to shift back.**
The people generating the code — "you and me for the next two and a half weeks" —
must produce concise code they understand front to back. **"We own the code."**

Then the pivot: with those trade-offs in mind and that mantra, he is still a
massive fan. Reminds you everything is optional, it's a choose-your-own-adventure,
and **if you hit problems, simplify**.

## The project brief

**A personal portfolio website with a digital twin** — a chatbot that answers
questions about you and your career. Built by pure YOLO vibe coding: "we do
nothing, we just set it off and let it go."

## Part 2 — OpenRouter setup, live (7:00-13:36)

| approx | on screen | what happens |
|---|---|---|
| 7:00 | slide | Why a middleman: previously you needed separate accounts, top-ups and API keys with OpenAI, Anthropic and Google. **OpenRouter routes to any of them from one account** — free models and paid models, small fee on top-ups. If you already use OpenAI directly, just swap it in |
| 8:00 | **openrouter.ai** | Sign up (Google credentials or whatever you like) |
| 8:40 | avatar > Keys | **Create API key.** Name it anything. Optionally set a **monthly dollar limit** and an **expiration** |
| 9:30 | the key | Copy it. **The warning he says he gives every time**: the key must be exact, including the **`sk-or-v1`** prefix — secret key, open router, v1. "If you get this wrong by a single digit it won't work... somehow there's always a steady flow of people that have problems with their keys" |
| 10:30 | good practice on camera | **He deletes the key he just showed** — "because you've all seen my key" |
| 11:00 | Settings > Privacy and guardrails | To use **free** models you must enable **"free endpoints that may train on inputs"** and **"free endpoints that may publish your prompts"** — i.e. **free costs you privacy**, stated plainly |
| 12:20 | avatar > Credits | For a stronger model, add credits. Notes many providers have a $5 minimum; **OpenRouter's was $2 last he checked** — "super convenient". Completely optional |
| 13:36 | end | Key on the clipboard, "time for us to start coding" |

## For our version

- **Part 1 is the most quotable material in Section 1** and it is all reading text
  off a screen. This is where our animation wins outright: the 50%-vs-67% result as
  a single chart, the Jellyfin rules as a checklist that builds, and **the
  asymmetry as a visual** — one person generating, one person reviewing.
- **Copyright caution.** He reads long passages of the Jellyfin policy verbatim.
  Ours must **paraphrase and attribute**, quoting at most a short line.
- **Keep "we own the code" as a recurring card.** It is the honest counterweight
  to a course that otherwise spends 16 hours making generation look easy.
- **Keep the key-deletion moment.** Showing a secret on camera and then revoking it
  is exactly the habit we want to model — and it doubles as our own PII rule made
  visible. See `coursesmith-no-pii-in-renders`.
- The free-models privacy trade-off must be stated as plainly as he states it.
