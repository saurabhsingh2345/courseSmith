# Research: world-class captions + context-aware graphics (FREE/OSS, 2025–2026)

_Web-research report, 2026-07-18._

## AREA 1 — Captions

### @remotion/captions (the backbone — zero new cost, already on Remotion)
- Standard `Caption` type `{text, startMs, endMs, timestampMs, confidence}` — maps 1:1 from
  whisperX words.
- **`createTikTokStyleCaptions()`** (v4.0.216+): captions[] + `combineTokensWithinMilliseconds`
  → `pages[]` each with `text`, `startMs`, `durationMs`, `tokens[]` (absolute fromMs/toMs).
  - 200–500ms = word-by-word; **900–1400ms = 3–5-word pages (education sweet spot)**.
  - Gotcha: tokens include leading whitespace → container needs `white-space: pre`.
- Active word: compare frame-derived ms vs token fromMs/toMs; spring scale/color/pill.
- Official **template-tiktok** repo: steal its Page/Word components and restyle.

### Standalone OSS caption burners — reference only
Captacity (MIT, MoviePy — style-parameter reference), subsai (GPLv3 — avoid), auto-subs
(NLE-oriented), auto-subtitle (no animation). "clip-crafter": nothing real.

### LLM keyword emphasis = confirmed OSS gap → build it
One LLM pass per section: input caption page texts → JSON `{word_index, emphasis:
"highlight"|"scale"|"emoji", emoji?}`. Constrain: ≤1 emphasized word/page, emoji on <20%
of pages. ~1 prompt of work with existing plumbing.

### ASS/libass vs React: don't switch
ASS karaoke (`\k`,`\t`) wins only without a compositor. Remotion gives springs, gradient
text, pills, design-token integration. ASS = second styling language, dated shaping.

### Style best practices (education-tuned)
- Chunking: MrBeast ≈2 words/page, Hormozi 1–3; for education 3–5 words/page; word-by-word
  only at key moments (constant flashing fatigues over 10-min lessons).
- Active word: spring scale 1.08–1.15 OR color shift — not both + position (too noisy).
- ≤1 keyword/page in accent hue, never stopwords. Mixed case reads better than ALL CAPS
  for education. Thick stroke/shadow (8–12px at 1080×1920, scale for 16:9).
- Caption block ≈10–15% of frame height; 16:9 → lower third, bottom margin ≥5%.
- Creators report 15–40% higher average view duration with animated word captions.

## AREA 2 — Context-aware graphics

### Local image generation on Apple Silicon
| Model | License | Verdict |
|---|---|---|
| **FLUX.2-klein-4B** (Jan 2026) | **Apache 2.0** | Best pick: ~8GB, seconds-class on M-series (klein-9B is non-commercial — avoid) |
| FLUX.1-schnell | Apache 2.0 | Good; ~50-80s/1024² on M-series |
| SDXL/-Turbo | OpenRAIL++ | Aging aesthetics; Turbo ok for texture/b-roll |
| SD 3.5 Medium/Turbo | Community license (free <$1M revenue) | License gate to check |
| Qwen-Image | Apache 2.0 | Best open text-in-image rendering; heavy/slow on Mac |
| HiDream-I1 | MIT | Too heavy for Mac throughput |
| Sana | code Apache, **outputs non-commercial** | Deal-breaker |

Runtime: **ComfyUI headless on MPS** (`/prompt` JSON API). PyTorch nightly; **FP8 fails on
MPS — use FP16/BF16 or GGUF**. ComfyUI-MLX ext claims 50–70% speedup. Feasible for ~1
themed illustration per lesson section generated async (30–90s each). Not reliable
unsupervised: generate 2–4 candidates, VLM picks best; fixed style-prefix from design
tokens + fixed seeds; avoid text-in-image.

### Better automated diagramming (highest ROI)
- **D2 (terrastruct, MPL-2.0, pure Go — embeddable as a library in the pipeline binary!)**:
  `--sketch` = tasteful hand-drawn; built-in themes + dark themes; deterministic; layout
  dagre/ELK free (TALA is paid). **Replace LLM-freehand-SVG with LLM→D2 source: compile
  errors become self-repair feedback, halving vision-QA burden.**
- **Kroki** (MIT, docker): one HTTP API over 30+ renderers (D2, PlantUML, GraphViz, Mermaid,
  Excalidraw, C4, Vega-Lite, TikZ…). Single sidecar → every DSL at once.
- **Mermaid v11 new kinds**: architecture-beta (icons!), block-beta, xychart-beta,
  sankey-beta, mindmap, timeline, kanban, radar, treemap, fishbone; `look: handDrawn`
  native sketch aesthetic.
- Penrose (MIT, beautiful math diagrams — only for math-heavy content); Typst CeTZ/fletcher
  (equations); mingrammer/diagrams (cloud-arch icons). PlantUML/Graphviz: dated, skip.

### napkin.ai-style text→infographic: OSS gap
Only OpenNapkinAI (early demo). Practical path: LLM → D2 sketch / mermaid handDrawn +
Iconify icons composited in Remotion — already 80% there.

### Free asset APIs
- **Iconify** (MIT, self-hostable): ~300k icons/214 sets, on-demand SVG with color/size
  params; filter per-set licenses. Best single asset win.
- **Pexels API**: photos + VIDEOS (b-roll), 20k req/mo, no attribution required (link to
  Pexels in product expected). Best b-roll option.
- Pixabay (must download/rehost — fine for a render pipeline). Unsplash (hotlink +
  attribution + app approval — most onerous, use Pexels first).
- **unDraw**: bundle once, recolor via accent token. svgl for brand logos (trademark care).

### Local VLM judges (replace gpt-4o-mini vision)
- **Qwen2.5-VL-7B** (Apache 2.0, via MLX on Mac): best local quality/size, real OCR of
  diagram labels. (3B is non-commercial license — use 7B.)
- Moondream 2 (fast, weak), Moondream 3 preview (9B MoE, pointing/grounding — license TBC),
  SmolVLM2 (cheap pre-filter).
- Moving diagrams to compile-validated D2/mermaid reduces what the VLM must catch.

## Recommended stack
1. Captions: createTikTokStyleCaptions + restyled template-tiktok components + own LLM
   emphasis pass.
2. Diagrams: LLM→D2 sketch (Go-native) primary; mermaid v11 expanded kinds secondary;
   optional Kroki sidecar.
3. Illustrations: FLUX.2-klein-4B via headless ComfyUI (FP16) + candidate-gen + VLM pick.
4. Assets: self-hosted Iconify + Pexels (rehost) + bundled unDraw.
5. QA: Qwen2.5-VL-7B via MLX.
