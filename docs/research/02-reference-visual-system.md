# The reference visual system: skins, and ten templates

Source material: four explainer clips supplied as look-and-feel references
(three by *Kai*, one by *PawelCodeStuff*). Frames were sampled at eight points
across each; the analysis below is of the visual grammar, not the content.

## What the references actually are

Two related but distinct systems.

**System A — Kai.** Near-black stage. Standing chrome on every frame: a tiny
mono uppercase eyebrow top-centre (`NUMBER 1 / 3 · CAPACITY`), one very large
condensed all-caps headline, a mono takeaway line bottom-left, a watermark
bottom-right. One small, precise diagram floating in the middle third with a lot
of air around it. Three accents only, used semantically rather than decoratively
— blue for the rival, gold for the measured quantity, red for the limit.

**System B — PawelCodeStuff.** Charcoal, a single violet accent, sentence case,
no headline furniture and no chrome at all. The diagram *is* the frame.

## The three things that carry the look

Most of the gap was never templates. In order of leverage:

1. **Chrome and type.** The eyebrow / headline / takeaway / watermark set,
   identical across every beat. This is where nearly all the cohesion comes from.
2. **Semantic colour.** Three fixed accents that mean the same thing in every
   clip, so a viewer reads the argument before the narrator finishes the
   sentence. Deliberately *not* branding — see `videoskin.go`.
3. **Composition.** Small diagram, lots of air. The opposite of the catalog's
   stage-filling default, and right for a different kind of clip — hence a skin
   setting rather than a global change.

## Skins

`style.skin` in course config: `default` (unchanged), `broadcast` (System A),
`minimal` (System B). Independent of `style.mode` (light/dark) — every skin
derives in both polarities and around the whole hue circle.

Skins are **additive by construction**. `deriveVideoTheme` runs exactly as it
always did and a skin then overrides the tokens it disagrees with.
`TestDefaultSkinIsUnchanged` and `TestDefaultSkinAddsNoJSONKeys` hold the line;
all 28 pre-existing visual baselines pass with **zero** pixels differing.

Implementation: `internal/pipeline/videoskin.go` (tokens, semantic accents,
air), `renderer/src/components/SceneChrome.tsx` (chrome),
`SceneHeader.tsx` (per-skin type), `Stage.tsx` (air as a scale — see below),
`SceneBackground.tsx` (the `void` surface).

**Air is a scale, not padding.** Padding does nothing: nearly every scene sizes
its content against the `STAGE_W` constant at module scope, so a fatter padding
leaves a fixed-width card exactly as wide and merely overflows the box. Scaling
reaches fixed and fluid widths alike and preserves each scene's internal
proportions. It arrives via `StageAirContext` so no scene had to be edited.

## The ten templates

Each is named for what fills the frame, with the reference beat it came from.

| Template | What fills the frame | The rule it enforces |
|---|---|---|
| `metric` | One figure per beat, counting up, unit + label + note; optional recap row | Every number needs a unit and a label; not every figure may be `neutral` |
| `gauge` | Bar filling toward a dashed ceiling; overrun drawn past it in the limit colour | The ceiling is established first; no bar past 4× it |
| `verdict` | Two asymmetric columns, then the call alone at headline size | At least one *narrated* condition where the call is wrong |
| `decision` | An axis split into bands, each carrying its own answer | Bounds ascend and the last band is open-ended — the partition is total |
| `myth` | A claim struck through in place and replaced by what is true | The correction may not be a bare negation of the claim |
| `analogy` | Metaphor column ↔ reality column, walked pair by pair | Nothing maps to nothing; the analogy must say where it breaks |
| `rundown` | Numbered card row, lit one at a time | A promise naming N must be backed by exactly N cards |
| `trace` | Actors → queue → one shared value, drained a step at a time | Every step states the value after it; the value must actually move |
| `costing` | Line items stacking into a running total | The stated total equals the sum, ±2%; one line must be a hidden cost |
| `constellation` | One idea centred, properties radiating, lit one spoke at a time | Every spoke carries the relation word that joins it to the centre |

All ten are complete: Go template, prompt, Remotion component, tests, Root
composition, visual baseline and gallery preview. The catalog is 28 templates.

(It is 32 now: `canvas` and `promptloop` followed, and then `chapter`, `cycle`
and `scale` — the v2 batch, written for the shape a course has rather than for
one question. See §14 of `whatwehave.md`.)

Boundaries against the existing catalog: `data` is charts and maps, `metric` is
a single number. `compare` *introduces* two things, `verdict` *judges* — every
reference clip ends with the latter. `flow` is static layered boxes with
traffic; `trace` is state changing over time. `stack` already covers the
layer-cake frames, so no template was added for them.

## Adding one: the fourteen touchpoints

Every template touches exactly these. `metric`, `gauge` and `verdict` are three
worked examples to copy from.

**Go**
1. `internal/pipeline/snippet_<name>.go` — `init()` registration, the spec and
   beat types, `Normalize`, `Validate`, `Scenes`.
2. `snippet.go` — one field on `SnippetPlan`, one on `SnippetBeat`.
3. `snippet_templates.go` — one `beatFields` bool, one `case` in
   `rejectForeignBeatFields`.
4. `snippet_normalize.go` — one `planFields` bool, one strip in
   `stripPlanFields`.
5. `scenegraph.go` — one `Scene<Name>` constant.
6. `internal/pipeline/snippet_<name>_test.go` — the rules, especially the one
   the template exists for.

**Prompt**
7. `prompts/snippet_<name>.tmpl` — `system` and `user` blocks. Must state the
   length requirement, the beat vocabulary, and the enforced shape.

**Renderer**
8. `renderer/src/components/<Name>Scene.tsx`.
9. `types.ts` — add to the `SceneType` union.
10. `LessonVideo.tsx` — import, `case` in `sceneContent`, and usually a line in
    `surfaceFor`.
11. `Root.tsx` — a `<name>VizProps` fixture and a `<Composition>`.

**Baselines**
12. `test/visual_regression.mjs` — one or two `TARGETS` entries, on the frame
    that holds the most states at once.
13. `test/template_previews.mjs` — one `SOURCES` entry.
14. `node test/visual_regression.mjs --update && node test/template_previews.mjs`.

`internal/studio`'s `TestEveryTemplateHasAPreview` fails if step 13/14 is
skipped, so the catalog cannot grow a template with no gallery card.

### Two things worth knowing before writing a test

- Removing beats to isolate a rule usually trips the **shared** beat-count floor
  first and proves nothing. Re-point the beats instead of deleting them.
- Fixtures for any template that distinguishes roles (`gauge`, `verdict`) must
  carry `accentQuantity` / `accentLimit` explicitly. `resolveTheme` falls both
  back to the brand accent, and a bar that overruns its ceiling would render the
  same gold as one that clears it — the one thing the picture must never do.
