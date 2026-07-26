import { useState } from "react";
import { useTutorAPI } from "../hooks/useTutorAPI";
import type { BKTEstimate, BKTParams, FSRSRating, FSRSResult } from "../api/tutor";
import { ErrorNote } from "../components/ErrorNote";

const DEFAULTS: BKTParams = { p_init: 0.2, p_learn: 0.15, p_slip: 0.1, p_guess: 0.2 };
const RATINGS: FSRSRating[] = ["again", "hard", "good", "easy"];

/**
 * Tune BKT/FSRS parameters and simulate a learner. Mastery is computed by the
 * coursesmith-tutor service over increasing prefixes of a response sequence, so
 * the bars show p(known) after each answer. The service is optional — when it's
 * unreachable the page surfaces a friendly note instead of failing.
 */
export function AdaptiveConfigPage() {
  const { estimateMastery, scheduleReview, error, loading } = useTutorAPI();
  const [params, setParams] = useState<BKTParams>(DEFAULTS);
  const [responses, setResponses] = useState<boolean[]>([true, true, false, true, true]);
  const [series, setSeries] = useState<number[]>([]);
  const [estimate, setEstimate] = useState<BKTEstimate | null>(null);
  const [fsrs, setFsrs] = useState<FSRSResult | null>(null);

  const setParam = (k: keyof BKTParams, v: number) => setParams((p) => ({ ...p, [k]: v }));

  const simulate = async () => {
    if (responses.length === 0) return;
    const prefixes = responses.map((_, i) => responses.slice(0, i + 1));
    const results = await Promise.all(prefixes.map((p) => estimateMastery(p, params)));
    const known = results.map((r) => (r ? r.p_known : 0));
    setSeries(known);
    setEstimate(results[results.length - 1] ?? null);
  };

  const schedule = async (rating: FSRSRating) => {
    const res = await scheduleReview(rating);
    if (res) setFsrs(res);
  };

  return (
    <div className="mx-auto max-w-3xl p-6">
      <h1 className="mb-1 text-lg font-semibold text-ink-100">Adaptive config</h1>
      <p className="mb-5 text-[13px] text-ink-500">
        BKT mastery + FSRS scheduling from the tutor service.
      </p>

      {error && <ErrorNote error={`Tutor service: ${error}`} />}

      <section className="mb-6 rounded-lg border border-ink-800 bg-ink-900 p-4">
        <h2 className="mb-3 text-[11px] uppercase tracking-wide text-ink-500">BKT parameters</h2>
        <div className="grid gap-4 sm:grid-cols-2">
          {(Object.keys(DEFAULTS) as (keyof BKTParams)[]).map((k) => (
            <label key={k} className="block">
              <div className="mb-1 flex justify-between text-[12px] text-ink-300">
                <span className="font-mono">{k}</span>
                <span className="text-ink-400">{params[k].toFixed(2)}</span>
              </div>
              <input
                type="range"
                min={0}
                max={1}
                step={0.01}
                value={params[k]}
                onChange={(e) => setParam(k, Number(e.target.value))}
                className="w-full accent-sky-500"
              />
            </label>
          ))}
        </div>
      </section>

      <section className="mb-6 rounded-lg border border-ink-800 bg-ink-900 p-4">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-[11px] uppercase tracking-wide text-ink-500">Response sequence</h2>
          <div className="flex gap-2">
            <button
              onClick={() => setResponses((r) => [...r, true])}
              className="rounded border border-emerald-500/40 bg-emerald-500/10 px-2 py-0.5 text-[12px] text-emerald-300 hover:bg-emerald-500/20"
            >
              + correct
            </button>
            <button
              onClick={() => setResponses((r) => [...r, false])}
              className="rounded border border-red-500/40 bg-red-500/10 px-2 py-0.5 text-[12px] text-red-300 hover:bg-red-500/20"
            >
              + incorrect
            </button>
            <button
              onClick={() => {
                setResponses([]);
                setSeries([]);
                setEstimate(null);
              }}
              className="rounded border border-ink-700 px-2 py-0.5 text-[12px] text-ink-400 hover:bg-ink-800"
            >
              clear
            </button>
          </div>
        </div>

        <div className="mb-4 flex flex-wrap gap-1">
          {responses.map((c, i) => (
            <span
              key={i}
              className={
                "flex h-6 w-6 items-center justify-center rounded text-[12px] " +
                (c ? "bg-emerald-500/20 text-emerald-300" : "bg-red-500/20 text-red-300")
              }
            >
              {c ? "✓" : "✗"}
            </span>
          ))}
          {responses.length === 0 && <span className="text-[12px] text-ink-600">No responses.</span>}
        </div>

        <button
          onClick={() => void simulate()}
          disabled={loading || responses.length === 0}
          className="rounded bg-sky-600 px-3 py-1 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
        >
          {loading ? "Simulating…" : "Simulate mastery"}
        </button>

        {series.length > 0 && (
          <div className="mt-4">
            <div className="mb-1 text-[11px] uppercase tracking-wide text-ink-500">
              p(known) over time
            </div>
            <div className="flex h-28 items-end gap-1">
              {series.map((v, i) => (
                <div key={i} className="flex flex-1 flex-col items-center gap-1">
                  <div className="flex w-full flex-1 items-end">
                    <div
                      className="w-full rounded-t bg-sky-500"
                      style={{ height: `${Math.max(2, v * 100)}%` }}
                      title={v.toFixed(2)}
                    />
                  </div>
                  <span className="text-[10px] text-ink-600">{Math.round(v * 100)}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {estimate && (
          <div className="mt-4 rounded border border-ink-800 bg-ink-950 p-3 text-[13px]">
            <div className="text-ink-200">
              Mastery <span className="font-semibold text-sky-300">{Math.round(estimate.p_known * 100)}%</span>
              {estimate.mastered && <span className="ml-2 text-emerald-400">· mastered</span>}
              <span className="ml-2 text-ink-500">difficulty: {estimate.difficulty_hint}</span>
            </div>
            <div className="mt-1 text-ink-400">{estimate.recommendation}</div>
          </div>
        )}
      </section>

      <section className="rounded-lg border border-ink-800 bg-ink-900 p-4">
        <h2 className="mb-3 text-[11px] uppercase tracking-wide text-ink-500">FSRS review scheduling</h2>
        <div className="flex flex-wrap gap-2">
          {RATINGS.map((r) => (
            <button
              key={r}
              onClick={() => void schedule(r)}
              className="rounded border border-ink-700 bg-ink-800 px-3 py-1 text-[13px] capitalize text-ink-200 hover:bg-ink-700"
            >
              {r}
            </button>
          ))}
        </div>
        {fsrs && (
          <div className="mt-3 text-[13px] text-ink-300">
            Next review in <span className="font-semibold text-sky-300">{fsrs.interval_days}</span> day
            {fsrs.interval_days === 1 ? "" : "s"} · stability {fsrs.stability.toFixed(1)} · {fsrs.note}
          </div>
        )}
      </section>
    </div>
  );
}
