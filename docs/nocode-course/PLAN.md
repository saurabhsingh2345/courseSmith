# Section 1 build plan — 1:1 against the reference

Decided with him 2026-08-27. Supersedes the "compress his lessons" approach used
on `nocode01`.

## The problem this fixes

`nocode01` (11:37) covered his L1 + L2 (**17:07**). That is 1.47x compression.
Repeated across 95 lectures it lands the course at ~10-11h instead of 16.5h, and
the content drift compounds lecture over lecture until ours is a different course.

**New rule: our runtime tracks his, lecture group by lecture group.** We never
compress his coverage. Where his screen goes flat we add graphics *on top of* the
same content rather than cutting the content out.

## The arithmetic

| | his | ours |
|---|---|---|
| Section 1 | 34 lectures, **5h 30m** | 14 videos, **5h 25m** |
| whole course | 95 lectures, **16h 18m** | ~41 videos, **~16h 10m** |

Target length **~25 min** per video. Section 1 = **14 videos** (one already shipped),
not 7-8. 8 x 25min would be 3h20m against his 5h30m.

## Section 1 video map

`nocode01` is already shipped and covers L1+L2. The remaining 32 lectures split
into 13 videos:

| ours | his lectures | his runtime | topic |
|---|---|---|---|
| nocode01 | L1, L2 | 17:07 (ours 11:37) | SHIPPED - install Cursor, one naive prompt, a 3D game |
| nocode02 | L3, L4, (L5 cut) | 24:08 | Karpathy's "alien tool" tweet, who this is for, the 3-week roadmap |
| nocode03 | L7, L8, L9 | 25:21 | the 8 stages of AI coding; how LLMs work |
| nocode04 | L10, L11 | 21:28 | tools/loops/agents; context engineering + agents.md |
| nocode05 | L12, L13 | 24:56 | agents.md deep dive; YOLO -> Ralph loops |
| nocode06 | L14, L15, L16 | 25:29 | Artificial Analysis; hands-on Cursor/Copilot/Codex; agents.md + settings |
| nocode07 | L17, L18 | 24:25 | Kanban in Cursor YOLO; Kanban in Copilot |
| nocode08 | L19, L20, L21 | 25:15 | Kanban in Codex; Kanban in Antigravity; the verdict |
| nocode09 | L22, L23 | 19:41 | picking the LLM; five principles, "be the boss" |
| nocode10 | L24, L25 | 24:10 | OpenRouter setup; Next.js site with Codex |
| nocode11 | L26, L27, L28 | 26:44 | AI digital twin chat; code review with Opus; Karpathy + MVP rules |
| nocode12 | L29, L30 | 23:05 | web apps 101; full-stack with Copilot + FastAPI |
| nocode13 | L31, L32 | 23:17 | planning + scaffolding; Kanban with Docker + FastAPI |
| nocode14 | L33, L34 | 21:36 | debugging drag and drop; week 1 wrap |

Note L6 sits inside nocode02's block by topic ("Vibe Coding, Agentic Coders &
Coding Agents") - decide at script time whether it lands at the end of nocode02
or the head of nocode03.

## House rules for this section

1. **Screen recording carries it.** He runs ~85/15 footage-to-graphics across the
   section. `nocode01` was 73/27. Push ours toward his ratio - more real screen,
   graphics only where his screen goes flat.
2. **Same content, our words.** Every demo he runs, we run. Every concept he
   teaches, we teach, in the same order. Wording is ours; never transcribe him.
3. **His speaking rate is 182 wpm** (measured on L2: 1,257 words in 414s). Our
   house voice runs 151-174 wpm. At our rate the same word count runs ~15% longer,
   so budget words from the target runtime, not from his script length.
4. **Not every lecture has footage.** L3 is 100% slides and talking head - zero
   screen recording. The 85/15 ratio is a *section* average, not a per-video one.
5. **His bio is not ours.** L4 spends ~4 min on his personal story (JP Morgan,
   Times Square billboard, flying a plane). We have no presenter. That time gets
   replaced with an animated course roadmap and a look at the four projects,
   holding the same runtime.

## Reference material on disk

- `curriculum.md` - all 95 lectures, exact durations, section and day blocks
- `lecture-ids.txt` - Udemy lecture ids for direct URLs
- `transcripts/L##.txt` - his full narration, pulled from the player transcript
- `ref-L##.md` - per-lecture beat sheets: what is on screen, what he demos

Player URL form:
`https://www.udemy.com/course/ai-coder-from-vibe-coder-to-agentic-engineer/learn/lecture/<id>`

---

# What the full watch changed — 2026-08-27

All 34 lectures of Section 1 are now watched and written up in `ref-L01.md` …
`ref-L34.md`. Three findings revise the plan above.

## 1. The 85/15 footage target was wrong

85/15 came from **lecture 1 alone**. Measured across the whole section, his real
split is **about 57% footage / 43% slides**, and it is wildly uneven by day:

| day | lectures | runtime | footage share |
|---|---|---|---|
| **D1** the landscape | L1-8 | 59:16 | **~26%** |
| **D2** foundations | L9-14 | 64:39 | **~10%** |
| **D3** the four tools | L15-21 | 67:56 | **~92%** |
| **D4** YOLO project | L22-27 | 60:00 | **~55%** |
| **D5** commercial MVP | L28-34 | 78:33 | **~90%** |
| **section** | 34 | **5h 30m** | **~57%** |

So `nocode01` at 73/27 was never the problem — the problem was runtime, not ratio.

**Revised rule: match his per-day shape, not one section-wide number.** Build days
run at 90%+ footage and need almost no graphics. Theory days (D1, D2) are where he
is weakest — **75 minutes of near-static slides across L3-L9 and L11-L13** — and
that is exactly where our animation should be spent. Do not manufacture footage on
a theory day; beat him with motion instead.

## 2. Build these graphics once, reuse them all course

He returns to the same handful of frames repeatedly, always static. Each of these
should be authored once, properly, and recalled:

- **The curriculum grid with the current day filled in + a percentage-complete
  beat.** Used 5 times in Section 1 alone (L4, L8, L14, L21, L27, L34 — at 7%, 13%,
  20%, 27%, 33%). Highest-leverage graphic in the course
- **The layered context stack** (L11) — system prompt / tools / memory / conversation
  / `agents.md`. Recalled in L12 and again at L33's context reset
- **The eight stages ladder** (L7) and **the six workflows** (L13). He says they are
  "somewhat analogous" and then draws them separately — ours should be one system
- **The three surfaces**: IDE / plugin / CLI (L6)
- **The four-column tool comparison** (L21) — assemble one column per build across
  D3 rather than revealing it as a new slide at the end
- **"Be the boss" — the five principles** (L23, repeated in L28)
- **The four-step debugging rule** (L18) — reproduce → prove root cause → fix →
  demonstrate. Recurs in L25, L30, L33
- **Front end / back end / API, and Dockerfile → image → container** (L29)

## 3. The pattern that names Week 1

Three separate incidents, all unscripted, all the same failure:

- **L32** — an 80% coverage target met by writing worthless tests for ten minutes
- **L33** — half an hour stuck because its own broken tests said the bug wasn't
  fixed, when it had already fixed it
- **L34** — mocking the AI call twice after being told to test against the real API

Plus **L18**'s guess-a-fix-and-claim-victory. Our Section 1 should **name this
pattern once, explicitly** — an agent satisfying the letter of an instruction while
missing its purpose — and let the four examples land under it. He never names it.

## Assets we must build before shooting

- **A starter Kanban repo** under our own account, containing one `agents.md`
  (his: `github.com/ed-donner/kanban`) — used by all four D3 builds
- **A `pm` starter repo** with the D3 Codex front end inherited, empty `backend/`
  and `scripts/`, `.env`, `.gitignore`, `agents.md` and `docs/plan.md`
- **Our own `agents.md`** — same seven sections as his (see `ref-L16.md`), our wording
- **Our own 10-part `plan.md`** (see `ref-L31.md`)
- **A neutral persona + profile PDF** for the digital-twin project. **His D4 build
  puts a real name, real email and a real LinkedIn profile on screen.** This is the
  highest PII risk in Section 1 — see `coursesmith-no-pii-in-renders`
- Accounts: OpenRouter, GitHub, Docker Desktop, and whichever IDEs we shoot

## Copyright note

L3, L6, L28 and L34 all read **Karpathy posts** aloud, and **L24 reads the Jellyfin
AI policy at length**. Ours must paraphrase and attribute — at most a short quoted
line, never a full reproduction.

## Two lectures that do not transfer as-is

- **L5** (3:42) — a cross-sell for his other five Udemy courses. **Cut entirely.**
- **L4** (0:00-3:50) — his personal biography. Replaced by an animated course
  roadmap and a look at the four projects, holding the same runtime.
