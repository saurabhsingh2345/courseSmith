import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type Question } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { ErrorNote } from "../components/ErrorNote";

const TYPE_COLOR: Record<string, string> = {
  recall: "bg-sky-500/20 text-sky-200",
  application: "bg-emerald-500/20 text-emerald-200",
  debugging: "bg-amber-500/20 text-amber-200",
  prediction: "bg-fuchsia-500/20 text-fuchsia-200",
};

/**
 * Greedy interleave mirroring internal/pipeline/quiz_strategy.go: repeatedly
 * place the most-common remaining type that differs from the last placed type,
 * minimizing adjacent same-type pairs. Deterministic for a given input.
 */
export function interleave(questions: Question[]): Question[] {
  const buckets = new Map<string, Question[]>();
  const order: string[] = [];
  for (const q of questions) {
    if (!buckets.has(q.type)) order.push(q.type);
    buckets.get(q.type)?.push(q) ?? buckets.set(q.type, [q]);
  }
  const out: Question[] = [];
  let last = "";
  while (out.length < questions.length) {
    let best = "";
    for (const t of order) {
      const n = buckets.get(t)?.length ?? 0;
      if (n === 0 || t === last) continue;
      if (best === "" || n > (buckets.get(best)?.length ?? 0)) best = t;
    }
    if (best === "") {
      for (const t of order) if ((buckets.get(t)?.length ?? 0) > 0) { best = t; break; }
    }
    out.push(buckets.get(best)!.shift()!);
    last = best;
  }
  return out;
}

/** easy≈30%, hard≈25%, medium the rest — mirrors difficultyTargets(). */
function difficultyTargets(n: number) {
  const easy = Math.floor((n * 3) / 10);
  const hard = Math.floor((n * 25) / 100);
  return { easy, medium: n - easy - hard, hard };
}

export function adjacentRepeats(qs: Question[]): number {
  let r = 0;
  for (let i = 1; i < qs.length; i++) if (qs[i].type === qs[i - 1].type) r++;
  return r;
}

/** Interleaving / spacing / difficulty analysis for a lesson's quiz. */
export function QuizStrategyPage() {
  const { slug = "", id = "" } = useParams();
  const { data, loading, error, reload } = useFetch(() => api.quiz(slug, id), [slug, id]);
  const [mode, setMode] = useState<"authored" | "interleaved">("interleaved");

  const questions = useMemo<Question[]>(
    () => data?.merged?.questions ?? data?.generated?.questions ?? [],
    [data],
  );
  const ordered = useMemo(
    () => (mode === "interleaved" ? interleave(questions) : questions),
    [mode, questions],
  );

  const typeCounts = useMemo(() => {
    const m: Record<string, number> = {};
    for (const q of questions) m[q.type] = (m[q.type] ?? 0) + 1;
    return m;
  }, [questions]);

  const targets = difficultyTargets(questions.length);
  const repeats = adjacentRepeats(ordered);

  return (
    <div className="mx-auto max-w-3xl p-6">
      <div className="mb-1 text-[12px] text-ink-500">
        <Link className="hover:underline" to={`/c/${encodeURIComponent(slug)}`}>
          {slug}
        </Link>{" "}
        / <span className="text-ink-400">{id}</span> / quiz strategy
      </div>
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-lg font-semibold text-ink-100">Quiz strategy</h1>
        <Link
          to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(id)}/quiz`}
          className="rounded border border-ink-700 bg-ink-800 px-3 py-1 text-[13px] text-ink-200 hover:bg-ink-700"
        >
          Edit questions →
        </Link>
      </div>

      {error && <ErrorNote error={error} onRetry={reload} />}
      {loading && !data && <div className="text-ink-500">Loading…</div>}

      {data && questions.length === 0 && (
        <div className="rounded-lg border border-ink-800 bg-ink-900 p-6 text-ink-500">
          No quiz generated for this lesson yet.
        </div>
      )}

      {questions.length > 0 && (
        <>
          <div className="mb-5 grid gap-3 sm:grid-cols-3">
            <Stat label="Questions" value={String(questions.length)} />
            <Stat
              label="Adjacent same-type"
              value={String(repeats)}
              tone={repeats === 0 ? "good" : "warn"}
            />
            <Stat
              label="Difficulty target"
              value={`${targets.easy}/${targets.medium}/${targets.hard}`}
              hint="easy / medium / hard"
            />
          </div>

          <div className="mb-3">
            <h2 className="mb-2 text-[11px] uppercase tracking-wide text-ink-500">Type mix</h2>
            <div className="flex h-3 overflow-hidden rounded bg-ink-800">
              {Object.entries(typeCounts).map(([t, n]) => (
                <div
                  key={t}
                  className={TYPE_COLOR[t]?.split(" ")[0] ?? "bg-ink-600"}
                  style={{ width: `${(n / questions.length) * 100}%` }}
                  title={`${t}: ${n}`}
                />
              ))}
            </div>
            <div className="mt-1 flex flex-wrap gap-2 text-[11px] text-ink-500">
              {Object.entries(typeCounts).map(([t, n]) => (
                <span key={t}>
                  {t}: {n}
                </span>
              ))}
            </div>
          </div>

          <div className="mb-3 flex gap-2">
            <ModeChip label="Interleaved" active={mode === "interleaved"} onClick={() => setMode("interleaved")} />
            <ModeChip label="As authored" active={mode === "authored"} onClick={() => setMode("authored")} />
          </div>

          <ol className="space-y-2">
            {ordered.map((q, i) => (
              <li
                key={q.id}
                className="flex items-start gap-3 rounded-lg border border-ink-800 bg-ink-900 p-3"
              >
                <span className="mt-0.5 w-5 shrink-0 text-right font-mono text-[12px] text-ink-500">
                  {i + 1}
                </span>
                <span
                  className={
                    "mt-0.5 shrink-0 rounded px-1.5 py-0.5 text-[10px] uppercase " +
                    (TYPE_COLOR[q.type] ?? "bg-ink-700 text-ink-300")
                  }
                >
                  {q.type}
                </span>
                <span className="text-[13px] text-ink-200">{q.prompt}</span>
              </li>
            ))}
          </ol>
        </>
      )}
    </div>
  );
}

function Stat({
  label,
  value,
  hint,
  tone,
}: {
  label: string;
  value: string;
  hint?: string;
  tone?: "good" | "warn";
}) {
  const color = tone === "good" ? "text-emerald-400" : tone === "warn" ? "text-amber-400" : "text-ink-100";
  return (
    <div className="rounded-lg border border-ink-800 bg-ink-900 p-3">
      <div className="text-[11px] uppercase tracking-wide text-ink-500">{label}</div>
      <div className={"text-xl font-semibold " + color}>{value}</div>
      {hint && <div className="text-[11px] text-ink-500">{hint}</div>}
    </div>
  );
}

function ModeChip({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={
        "rounded-full border px-3 py-1 text-[12px] " +
        (active
          ? "border-sky-500/50 bg-sky-500/10 text-sky-200"
          : "border-ink-700 bg-ink-800 text-ink-300 hover:bg-ink-700")
      }
    >
      {label}
    </button>
  );
}
