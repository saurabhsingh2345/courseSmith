# tools/tutor — adaptive-learning models (workstream D)

The Go microservice (`cmd/coursesmith-tutor`) serves BKT + a simplified
spaced-repetition scheduler with real math, and stubs IRT calibration. The
heavier statistical models live here as a Python sidecar, mirroring how
`tools/align` hosts whisperX.

## Shipped (dependency-free) + the library upgrade path

The offline fitters are implemented in plain Python (stdlib only, so they run
with a stock `python3` — no venv needed). Each has a documented seam to swap in
the heavier statistical library for production-scale fitting.

| Model | Shipped now | Production upgrade |
| --- | --- | --- |
| BKT (offline fit) | `fit_bkt.py` — coarse max-likelihood grid search over the same forward recurrence the Go engine uses; fits init/learn/slip/guess from pooled response sequences. | [pyBKT](https://github.com/CAHLR/pyBKT) EM fit. |
| IRT calibration | `calibrate_irt.py` — 1PL difficulty via `logit(1 - p_correct)` + point-biserial discrimination scaled to [0, 2]. | [py-irt](https://github.com/nd-ball/py-irt) 2PL/3PL. |
| FSRS scheduling | the Go service's simplified scheduler (`internal/adaptive/engine.go`). | [fsrs](https://github.com/open-spaced-repetition/py-fsrs) FSRS-5 weights, once review history exists. |

## Layout

```
tools/tutor/
  pyproject.toml        # uv-managed; `uv sync --extra models` for pyBKT/py-irt/fsrs
  fit_bkt.py            # logs → per-concept BKT params JSON   (--selftest)
  calibrate_irt.py      # logs → per-question difficulty/discrimination JSON (--selftest)
  test_tutor.py         # stdlib unittest for both fitters
  README.md             # this file
```

## Usage

```bash
python3 fit_bkt.py responses.json        # {"sequences": [[true,false,...], ...]} → params
python3 calibrate_irt.py responses.json  # {"students": [{"q1": true, ...}, ...]} → items
python3 -m unittest discover tools/tutor # run the tests (no pytest needed)
```

## Data flow

1. Lesson pages POST quiz responses to the running service
   (`/bkt/estimate`, `/fsrs/schedule` — real online math in `internal/adaptive`).
2. Periodically, pooled responses are fit offline here (`fit_bkt.py`,
   `calibrate_irt.py`) and the resulting parameters are fed back to the service
   and to the quiz-strategy stage.

No student data is persisted yet — the online service runs with sensible
defaults and the offline fitters run on demand, so the whole integration shape
is real and testable before a student database exists.
