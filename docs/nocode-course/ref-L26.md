# L26 — "Day 4: Adding an AI Digital Twin Chat with OpenRouter & Vibe Coding" (7:49)

**100% screen recording.** Adds the AI feature to the site, then opens the hood.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | terminal | **Ctrl C** to stop the server, `clear` |
| 0:20 | slide-free aside | **Checkpoint before you go further.** The right way is **Git commits** (covered later). The **"poor person's version of Git"**: duplicate the whole `site` folder as a snapshot. Warning for anyone using Git: **there is already a repo under `web`** that must be removed so there's only one |
| 1:20 | agent prompt | *That's great. Please now add the ability to have an AI chat with a digital twin which can answer questions about my career. Please use OpenRouter. My OpenRouter API key is in the `.env` file in the project root. Please use the model named…* |
| 1:50 | **openrouter.ai/models** | Filters the model list by **free**. Looks for **GPT-OSS**, the open-source model — there is a **free version (occasional use only)** and a **paid version**. Picks the **paid one**: "still super cheap, but you can choose either." Copies the exact model name back into the prompt |
| 2:40 | prompt sent | *Make the changes. Make sure it works. Let me know when ready for me.* |
| 3:10 | result | "An AI digital twin is wired up and ready", using `.env`. His genuine suspense: **"What's your bet? Is this going to work?"** |
| 3:30 | `npm run dev` | Server holds. Browser to localhost:3000 |
| 4:00 | the site | **It scrolls straight down to the digital-twin section on load** — "I wonder if it meant to do that." The section doesn't look like a traditional chat: buttons and prompts rather than a chat box |
| 4:40 | **chatting live** | "Hi there" — **slightly janky, it bounces around** — then: *Hello, I'm the digital twin of Ed Donner. How can I assist you today?* Then **"What are you most proud of?"** and a genuinely good answer about building Nebula and watching people find roles that fit. **"What a nice answer. Nicely put"** |
| 5:40 | assessment | "This is showing you the good and the rough edges of vibe coding." Good: it works, it looks distinctive, "I've seen many digital twins and they don't tend to look this way." Rough: **the auto-scroll on load**, and **some buttons do nothing at all** |
| 6:20 | the reckoning | Names what they did: **by genuinely YOLO vibe coding, "we've broken all of the rules I said in the first half of today"** — no `agents.md`, no checking as we go. Fine for an MVP and for exploring, **but to take it further you must come back**: test thoroughly, fix what you're not satisfied with, iterate the prompts, and **read the code** |
| 6:50 | **opening the hood** | Navigates `web` > `src` > `app` > `api` > `chat` > **`route.ts`**. Pushes the agent panel aside, enlarges the code. **"This is the backend code that calls OpenRouter."** Reads through it |
| 7:30 | the finding | Finds **the system prompt that tells the model who he is** — and it is **simplistic, a very basic prompt.** The improvement he identifies himself: **send the whole LinkedIn profile** so the chatbot can answer robustly |
| 7:49 | end | |

## For our version

- **The checkpointing aside is genuinely important and he gives it 60 seconds.**
  For a no-code audience "duplicate the folder before you let it loose" is
  practical, immediately actionable advice. Give it a proper graphic.
- **"Opening the hood" is the beat that redeems the whole day.** After 20 minutes
  of not looking at code, he opens one file, reads it, and finds a real weakness.
  That is the answer to "but I don't know how to code" — you don't need to write
  it to *look* at it. Frame it exactly that way.
- **Same PII rule as L25** — the twin answers questions about a real person using a
  real profile. Ours uses a neutral persona throughout.
- The rough edges (auto-scroll, dead buttons) must survive our cut. A working demo
  with visible flaws teaches more than a polished one.
