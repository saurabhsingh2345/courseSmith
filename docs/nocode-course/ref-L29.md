# L29 — "Day 5: Web Apps 101: Front-End, Back-End, APIs & Docker Setup" (11:20)

**Slides, then a live install.** He states outright that these slides are lifted
from his MLOps course. Explicitly aimed at the non-technical half: **"for people
that are already pros, please put me on 2X, just zip through this part."**

This is the **most important lecture in Section 1 for a no-code audience.**

## Front-end vs back-end (0:40-5:00)

- The projects are **web applications** — you reach them through a browser
- **Front end** runs in the user's browser (Chrome, Safari, Edge): **HTML** (the
  page), **CSS** (appearance), **JavaScript** (interactivity)
- **Back end** runs on a server: it served the page in the first place, and the
  browser can send it messages to run business logic — **database access, calling
  an LLM, calling other APIs**. **Your secrets live here**, e.g. the `.env` file
- **APIs are the connection.** The front end makes an API call; results come back;
  that's how the two halves collaborate
- Charmingly pre-empts the objection: "all you pros are thinking of the ways I'm
  oversimplifying — feel free to explain it better than me in the Q&A"

**Where each project sat**, which is what makes it concrete:

| project | shape |
|---|---|
| Day 1 shooter | **all front end** |
| Day 3 Kanban | **all front end**, all JavaScript |
| Day 4 website + twin | had a back end, **but in JavaScript** — that's the part calling OpenRouter |
| **Today** | **JavaScript front end + Python back end** — "the most common back-end language for people working with LLMs" |

## Front-end frameworks, briefly (5:00-7:00)

Vanilla HTML/CSS/JS with simple libraries like jQuery (day one) → component
frameworks that update themselves: **React, Vue, Angular, Svelte** → the **single
page app (SPA)** pattern, loaded in one request with components making their own
API calls → **JavaScript or TypeScript** → and higher-level frameworks like
**Next.js**, made by **Vercel**, packaging routing, data fetching and
client-or-server rendering.

**The honest aside, and it is the best thing here:** he is "a bit of a horror with
React and Next.js" and coding agents have been a lifesaver. **But the caveat** —
LLMs produce sites that all look the same, **"a bit like slop"**, often with **that
purple hue**, "the same three icons, that very standard LLM-generated look." Where
great UX people still stand out is knowing **how to organise and communicate
information for the user**. So push back on the first draft. **"That's where you
can add the value."**

## Docker (7:00-11:20)

"I imagine 60 to 70% of you know it back to front. Let me quickly demystify."

- **Docker gives you a computer within your computer** — a ring-fenced set of
  resources isolated from the outside world. A lightweight alternative to virtual
  machines
- Why people love it: **isolation** ("you can't damage the outside world" — he
  notes this matters for agentic coding) and **portability** — build once, deploy
  elsewhere, "usually just works"
- **The three words**:
  - **Dockerfile** — a file, "like a recipe", instructions for installing and
    configuring the box within the box
  - **Image** — a snapshot or blueprint built from the Dockerfile, ready for
    primetime
  - **Container** — a live environment created from an image. **One image can
    create many containers**
- **Install, live**: **docker.com** → **Docker Desktop** → accept all defaults. On
  a PC it will prompt for **WSL** — say yes. May need a restart for the path
- **Docker Desktop tour**: **Containers** and **Images** in the left rail (blank
  for you, "I have a bunch"), plus **volumes** — ring-fenced storage containers use

## For our version

- **This is where our animation earns the whole section.** Front-end/back-end/API
  as one diagram with a request travelling across it; the four projects mapped onto
  that diagram in turn; Dockerfile → image → container as a build-then-clone
  sequence. He shows static slides for all of it.
- **The "LLM slop look" caveat is a gift to us specifically**, because our own
  course is about design as the differentiator. Show three generated sites side by
  side with the same purple hue and the same three icons — then say what to do
  about it.
- Keep the **"put me on 2x"** honesty. It costs nothing and it buys trust.
- Docker install is real footage — shoot it, ramp the download.
