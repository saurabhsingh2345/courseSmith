# Quality gates (workstream I)

Regression harness for the parts of courseSmith that unit tests can't cover:
rendered pixels, animation timing, and learning-science invariants. These sit
on top of the per-package unit tests (`go test ./...`, studio `vitest`, renderer
`tsc`) and catch regressions those miss.

| Gate | Checks | Status |
| --- | --- | --- |
| `internal/pipeline/quiz_strategy_test.go` | Interleaving minimizes adjacent same-type pairs; difficulty split sums to N and skews medium; the stage emits a valid permutation | **working** — pure Go, deterministic, no deps |
| `visual_regression.mjs` | Renders each Remotion composition to a PNG and pixel-diffs against a committed baseline under `baselines/` | **working** — pixelmatch + committed baselines |
| `animation_timing.mjs` | Renders frames of a composition and asserts animation boundaries (derived from the motion tokens) land within a frame tolerance | **working** — samples a pixel signal per frame |
| `learning_science.mjs` | Validates a generated `quiz_sequence.json` against the interleaving + difficulty-distribution rules | **working** — pure JSON checks, no browser |
| studio `test:a11y` | Runs axe over the base component library (ARIA/roles/labels) | **working** — vitest-axe + jsdom |

## Running

```bash
# Learning science (fast, deterministic)
go test ./internal/pipeline -run 'Interleav|Difficulty|QuizStrategy' -v
node test/learning_science.mjs courses/python-basics/lessons/01-what-is-python/generated

# Visual regression (needs a browser; ensureBrowser downloads a headless shell)
node test/visual_regression.mjs            # compare against baselines/*.png
node test/visual_regression.mjs --update   # re-record baselines after an intended change

# Animation timing
node test/animation_timing.mjs

# Accessibility (from studio/)
cd studio && npm run test:a11y
```

## Baselines

`baselines/*.png` are the committed golden frames — regenerate them with
`--update` whenever a composition intentionally changes, and eyeball the diff
before committing. The transient `*.actual.png` / `*.diff.png` a compare run
writes are gitignored.

The renderer scripts reuse the `@remotion/bundler` + `@remotion/renderer` path
proven for still renders, plus `pixelmatch` + `pngjs` (renderer dev-deps). They
skip cleanly (exit 0) where a browser or the render deps aren't available, so
they can be wired into CI ahead of every runner being able to render. CI wires
everything in `.github/workflows/{visual-regression,accessibility,learning-science}.yml`
(and the umbrella `quality-gates.yml`).
