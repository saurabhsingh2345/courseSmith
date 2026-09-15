# L25 — "Day 4: YOLO Mode: Building a Next.js Website with GPT Codex in Cursor" (10:34)

**100% screen recording.** The YOLO project build, in real time, including a
failure and a lazy fix.

## Setup (0:00-4:30)

| approx | on screen | what happens |
|---|---|---|
| 0:00 | Cursor | File > New Window > Open Project > `projects` > **New Folder `site`** > Open. **SITE** in block capitals confirms it |
| 0:40 | **`.env`** | Right-click > New File > **exactly `.env`**. Inside: **`OPENROUTER_API_KEY=`** then the pasted key. Repeated warnings that the spelling must be exact. **Cmd S to save — "the blob goes away"**. Notes the OpenAI variant is `OPENAI_API_KEY` |
| 2:20 | **`.gitignore`** | New File > `.gitignore`, containing `.env`. Explains what a gitignore is and why: even with no repo yet, if you ever make one **the key never gets committed.** "Always good best practice" |
| 3:20 | **LinkedIn** | His profile page > **Resources > Save to PDF**. Caveats that LinkedIn changes this and it may not be available to everyone — **fallback: paste the profile text into a text file, or use a PDF of your CV.** Moves the PDF into the project directory |
| 4:00 | Cursor settings | **Cmd Shift J**. Agents > set **Usage Summary** to Always (shows plan consumption). Scroll to **Auto Run** > **Run Everything Unsandboxed** — with the standard "if you're not comfortable, keep approving step by step". Models tab > enable **GPT 5.2 Codex High**: "more expensive, but higher quality" |

## The one prompt (4:30)

Drags the agent panel wide, turns **Auto off**, selects **GPT 5.2 Codex High**.
Names what he's doing wrong first — the right approach is an `agents.md` and
careful information — "**but we are on complete YOLO mode. Tomorrow we're going to
be more disciplined.**"

The prompt, in substance: *build me a professional website running locally; my
LinkedIn profile is in `linkedin.pdf`; make it stunning, **enterprise meets edgy**;
include about me, my career journey, links to a portfolio for the future; iterate
to make it as slick and professional as possible; let me know when complete;
use Next.js.*

## The build and the failure (5:30-8:00)

- It plans, then **creates a `web` subdirectory itself** — "good for it" (he had
  predicted trouble running in the root)
- Long wait: "**it may mean you need to go and get another coffee**"
- Reports a polished enterprise-meets-edgy Next.js site in `web`
- Terminal (**Ctrl backtick**), `cd web`, **`npm run dev`**. Server starts
- **localhost:3000 → "That's not much good."** It's broken, live on camera
- **The lazy fix, done deliberately**: he copies the raw error, pastes it into the
  agent with **no explanation at all**, and hits enter. **"We are vibing. We are
  YOLOing. We're not going to even bother explaining ourselves"**
- It returns in seconds with a tiny change and says **"the page should render"** —
  and he immediately flags it: **it didn't prove the problem and didn't prove the
  fix.** "What we should do is have an agents.md that insists: do not say you fixed
  something unless you've proven it." Then carries on anyway

## The result (8:00-10:34)

`npm run dev` again — and it works. His reaction: **"First impression, positive. I
like it a lot."** He never specified colours and it does read as enterprise-meets-edgy.

Walks the site full-screen: title, **About Me** summarised from the LinkedIn PDF,
**Career Journey** (the background darkens on scroll — "that's kind of cool"),
**Portfolio coming soon**, and a *Let's build the future / Book a conversation /
Connect on LinkedIn* section. Tests every link: **Career** scrolls to the journey,
**Portfolio** opens his real personal site on a separate page ("that's clever of
it"), **Let's talk** opens mail to **the real email address it pulled from his
LinkedIn profile**.

Verdict: a great website. **"The only thing it's missing is some functionality."**

## For our version

- **The `.env` + `.gitignore` sequence is the most important 3 minutes for a
  no-code audience in the whole section** — it is the first time they handle a
  secret. Shoot it slowly, and put a graphic over it explaining what a key is and
  why it must never be committed.
- **PII warning, ours specifically.** His build pulls a real name, real email and a
  real LinkedIn profile onto screen. **We must use a neutral persona** — a made-up
  profile PDF with no real contact details anywhere in frame. See
  `coursesmith-no-pii-in-renders`; this is the highest-risk lecture in Section 1.
- **Keep the breakage and keep the lazy fix.** Him paste-the-error-and-shrug, then
  immediately naming why that was bad practice, is the exact tension the day is
  about. Do not shoot a clean run.
- Ramp the build wait 12-14x per house rules rather than cutting it.
