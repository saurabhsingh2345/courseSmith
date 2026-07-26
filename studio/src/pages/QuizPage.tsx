import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { useRun } from "../state/RunContext";
import { useScreenShortcuts } from "../state/ShortcutContext";
import { ErrorNote } from "../components/ErrorNote";
import {
  computeOverrides,
  diffQuestion,
  overriddenFields,
  toEdited,
  type EditedQuestion,
  type QuizQuestion,
} from "../lib/quizDiff";

/** Seed the editor from the generated quiz with any saved overrides layered on. */
function seedEdited(
  generated: QuizQuestion[],
  overrides: { id: string; drop?: boolean; prompt?: string; options?: string[]; answer_index?: number; explanation?: string }[],
): EditedQuestion[] {
  const byId = new Map(overrides.map((o) => [o.id, o]));
  return generated.map((g) => {
    const base = toEdited(g);
    const o = byId.get(g.id);
    if (o) {
      if (o.drop) base.drop = true;
      if (o.prompt !== undefined) base.prompt = o.prompt;
      if (o.options !== undefined) base.options = [...o.options];
      if (o.answer_index !== undefined) base.answer_index = o.answer_index;
      if (o.explanation !== undefined) base.explanation = o.explanation;
    }
    return base;
  });
}

export function QuizPage() {
  const { slug = "", id = "" } = useParams();
  const { run, refreshTick, startRun } = useRun();
  const { data, loading, error, reload } = useFetch(
    () => api.quiz(slug, id),
    [slug, id, refreshTick],
  );

  const generated = useMemo<QuizQuestion[]>(
    () => (data?.generated?.questions ?? []) as QuizQuestion[],
    [data],
  );
  const savedOverrides = useMemo(() => data?.overrides?.questions ?? [], [data]);

  const [edited, setEdited] = useState<EditedQuestion[]>([]);
  const [dirty, setDirty] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveMsg, setSaveMsg] = useState<string | null>(null);

  useEffect(() => {
    if (!data) return;
    setEdited(seedEdited(generated, savedOverrides));
    setDirty(false);
    setSaveMsg(null);
  }, [data, generated, savedOverrides]);

  useScreenShortcuts([{ keys: "r", label: "reload" }], (e) => {
    if (e.key === "r" && !dirty) reload();
  });

  const pending = useMemo(
    () => computeOverrides(generated, edited).questions,
    [generated, edited],
  );

  const patch = (idx: number, partial: Partial<EditedQuestion>) => {
    setEdited((prev) => prev.map((e, i) => (i === idx ? { ...e, ...partial } : e)));
    setDirty(true);
    setSaveMsg(null);
  };

  const setOption = (idx: number, optIdx: number, value: string) => {
    setEdited((prev) =>
      prev.map((e, i) =>
        i === idx ? { ...e, options: e.options.map((o, j) => (j === optIdx ? value : o)) } : e,
      ),
    );
    setDirty(true);
    setSaveMsg(null);
  };

  const addOption = (idx: number) => {
    setEdited((prev) => prev.map((e, i) => (i === idx ? { ...e, options: [...e.options, ""] } : e)));
    setDirty(true);
  };

  const removeOption = (idx: number, optIdx: number) => {
    setEdited((prev) =>
      prev.map((e, i) => {
        if (i !== idx) return e;
        const options = e.options.filter((_, j) => j !== optIdx);
        let answer = e.answer_index;
        if (optIdx === answer) answer = 0;
        else if (optIdx < answer) answer -= 1;
        return { ...e, options, answer_index: Math.max(0, Math.min(answer, options.length - 1)) };
      }),
    );
    setDirty(true);
  };

  const resetQuestion = (idx: number) => {
    const gen = generated[idx];
    if (gen) patch(idx, toEdited(gen));
  };

  const resetAll = () => {
    setEdited(generated.map((g) => toEdited(g)));
    setDirty(true);
    setSaveMsg(null);
  };

  const save = async () => {
    setSaving(true);
    setSaveMsg(null);
    try {
      const payload = computeOverrides(generated, edited);
      await api.putQuizOverrides(slug, id, payload);
      setDirty(false);
      setSaveMsg(
        `Saved — ${payload.questions.length} override${payload.questions.length === 1 ? "" : "s"}`,
      );
      reload();
    } catch (err) {
      setSaveMsg(err instanceof Error ? err.message : String(err));
    } finally {
      setSaving(false);
    }
  };

  const regenerate = () => {
    if (run.running) return;
    // Force a fresh generation; the backend re-runs the quiz stage.
    void startRun({ course: slug, lesson: id, stage: "quiz", force: true });
  };

  return (
    <div className="mx-auto max-w-3xl p-6">
      <div className="mb-4 flex items-center gap-2 text-ink-500">
        <Link to="/" className="hover:text-ink-200">
          Courses
        </Link>
        <span>/</span>
        <Link to={`/c/${encodeURIComponent(slug)}`} className="hover:text-ink-200">
          {slug}
        </Link>
        <span>/</span>
        <Link
          to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(id)}`}
          className="hover:text-ink-200"
        >
          {id}
        </Link>
        <span>/</span>
        <span className="text-ink-200">quiz</span>
      </div>

      {error && <ErrorNote error={error} onRetry={reload} />}
      {loading && !data && <div className="text-ink-500">Loading…</div>}

      {data && generated.length === 0 && (
        <div className="rounded-lg border border-ink-800 bg-ink-900 p-4 text-ink-400">
          No generated quiz for this lesson yet.{" "}
          <button
            onClick={regenerate}
            disabled={run.running}
            className="text-sky-300 hover:underline disabled:opacity-40"
          >
            Generate it
          </button>
          .
        </div>
      )}

      {data && generated.length > 0 && (
        <>
          <div className="mb-4 flex items-center gap-2">
            <h1 className="text-lg font-semibold text-ink-100">
              {data.generated?.title ?? "Quiz"}
            </h1>
            <span className="text-ink-500">
              {generated.length} question{generated.length === 1 ? "" : "s"}
            </span>
          </div>

          <div className="space-y-4">
            {edited.map((q, i) => {
              const gen = generated[i];
              const changed = gen ? overriddenFields(diffQuestion(gen, q) ?? { id: q.id }) : [];
              return (
                <div
                  key={q.id}
                  className={`rounded-lg border p-4 ${
                    q.drop
                      ? "border-ink-800 bg-ink-900/40 opacity-60"
                      : "border-ink-800 bg-ink-900"
                  }`}
                >
                  <div className="mb-2 flex items-center gap-2">
                    <span className="font-mono text-[11px] text-ink-500">
                      {gen?.type ?? "question"} · {q.id}
                    </span>
                    {changed.length > 0 && (
                      <span className="rounded bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium text-amber-400">
                        edited: {changed.join(", ")}
                      </span>
                    )}
                    <div className="ml-auto flex items-center gap-3 text-[11px]">
                      <label className="flex items-center gap-1 text-ink-400">
                        <input
                          type="checkbox"
                          checked={q.drop}
                          onChange={(e) => patch(i, { drop: e.target.checked })}
                        />
                        drop
                      </label>
                      {changed.length > 0 && (
                        <button
                          onClick={() => resetQuestion(i)}
                          className="text-ink-400 hover:text-ink-200"
                        >
                          reset
                        </button>
                      )}
                    </div>
                  </div>

                  <textarea
                    value={q.prompt}
                    onChange={(e) => patch(i, { prompt: e.target.value })}
                    rows={2}
                    disabled={q.drop}
                    className="mb-3 w-full rounded border border-ink-800 bg-ink-950 p-2 text-ink-100 disabled:opacity-50"
                  />

                  <div className="space-y-1.5">
                    {q.options.map((opt, oi) => (
                      <div key={oi} className="flex items-center gap-2">
                        <input
                          type="radio"
                          name={`ans-${q.id}`}
                          checked={q.answer_index === oi}
                          onChange={() => patch(i, { answer_index: oi })}
                          disabled={q.drop}
                          title="Mark as correct answer"
                        />
                        <input
                          type="text"
                          value={opt}
                          onChange={(e) => setOption(i, oi, e.target.value)}
                          disabled={q.drop}
                          className={`flex-1 rounded border bg-ink-950 px-2 py-1 disabled:opacity-50 ${
                            q.answer_index === oi
                              ? "border-emerald-500/40 text-emerald-300"
                              : "border-ink-800 text-ink-200"
                          }`}
                        />
                        <button
                          onClick={() => removeOption(i, oi)}
                          disabled={q.drop || q.options.length <= 2}
                          title={q.options.length <= 2 ? "Need at least 2 options" : "Remove option"}
                          className="px-1 text-ink-500 hover:text-red-400 disabled:opacity-30"
                        >
                          ✕
                        </button>
                      </div>
                    ))}
                    <button
                      onClick={() => addOption(i)}
                      disabled={q.drop}
                      className="text-[11px] text-sky-300 hover:underline disabled:opacity-40"
                    >
                      + add option
                    </button>
                  </div>

                  <div className="mt-3">
                    <div className="mb-1 text-[11px] uppercase tracking-wide text-ink-500">
                      Explanation
                    </div>
                    <textarea
                      value={q.explanation}
                      onChange={(e) => patch(i, { explanation: e.target.value })}
                      rows={2}
                      disabled={q.drop}
                      className="w-full rounded border border-ink-800 bg-ink-950 p-2 text-ink-300 disabled:opacity-50"
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </>
      )}

      {/* Sticky action bar */}
      {data && generated.length > 0 && (
        <div className="sticky bottom-0 z-10 mt-4 flex items-center gap-3 border-t border-ink-800 bg-ink-950/95 py-3 backdrop-blur">
          <button
            onClick={() => void save()}
            disabled={saving || !dirty}
            className="rounded bg-sky-600 px-3 py-1 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
          >
            {saving ? "Saving…" : "Save overrides"}
          </button>
          <button
            onClick={resetAll}
            disabled={pending.length === 0 && !dirty}
            className="rounded border border-ink-700 bg-ink-800 px-3 py-1 text-ink-200 hover:bg-ink-700 disabled:opacity-40"
          >
            Reset to generated
          </button>
          <button
            onClick={regenerate}
            disabled={run.running}
            className="rounded border border-ink-700 bg-ink-800 px-3 py-1 text-ink-200 hover:bg-ink-700 disabled:opacity-40"
            title="Force-regenerate the quiz from source (discards nothing; overrides still apply on top)"
          >
            Regenerate
          </button>
          <span className="ml-auto text-[11px] text-ink-500">
            {dirty ? "unsaved changes · " : ""}
            {pending.length} override{pending.length === 1 ? "" : "s"} pending
          </span>
          {saveMsg && <span className="text-[11px] text-ink-400">{saveMsg}</span>}
        </div>
      )}
    </div>
  );
}
