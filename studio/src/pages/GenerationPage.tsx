import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { useGeneration } from "../hooks/useGeneration";
import { ErrorNote } from "../components/ErrorNote";
import { useScreenShortcuts } from "../state/ShortcutContext";

/**
 * Generation control panel: pick a course + lesson, start/cancel a pipeline
 * run, and watch progress and logs stream in over the existing SSE run channel.
 * Runs are singleton server-side, so this drives the same run the header RunBar
 * reflects.
 */
export function GenerationPage() {
  const { data: courses } = useFetch(() => api.courses(), []);
  const [slug, setSlug] = useState("");
  const [lessonId, setLessonId] = useState("");
  const [fromStage, setFromStage] = useState("");
  const [force, setForce] = useState(false);
  const [startErr, setStartErr] = useState<string | null>(null);

  const { data: detail } = useFetch(
    () => (slug ? api.course(slug) : Promise.resolve(null)),
    [slug],
  );
  const gen = useGeneration(detail?.stage_order);
  const logEndRef = useRef<HTMLDivElement>(null);

  useScreenShortcuts([{ keys: "g", label: "start" }], (e) => {
    if (e.key === "g") void start();
  });

  useEffect(() => {
    logEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [gen.logs.length]);

  const start = async () => {
    if (!slug || !lessonId || gen.run.running) return;
    setStartErr(null);
    try {
      await gen.start({ course: slug, lesson: lessonId, stage: fromStage || undefined, force });
    } catch (err) {
      setStartErr(err instanceof Error ? err.message : String(err));
    }
  };

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="mb-4 text-lg font-semibold text-ink-100">Generation</h1>

      <div className="mb-5 rounded-lg border border-ink-800 bg-ink-900 p-4">
        <div className="grid gap-3 sm:grid-cols-2">
          <label className="block">
            <span className="mb-1 block text-[11px] uppercase tracking-wide text-ink-500">Course</span>
            <select
              value={slug}
              onChange={(e) => {
                setSlug(e.target.value);
                setLessonId("");
              }}
              className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100"
            >
              <option value="">Select course…</option>
              {(courses ?? []).map((c) => (
                <option key={c.slug} value={c.slug}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>

          <label className="block">
            <span className="mb-1 block text-[11px] uppercase tracking-wide text-ink-500">Lesson</span>
            <select
              value={lessonId}
              onChange={(e) => setLessonId(e.target.value)}
              disabled={!detail}
              className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100 disabled:opacity-50"
            >
              <option value="">Select lesson…</option>
              {(detail?.lessons ?? []).map((l) => (
                <option key={l.id} value={l.id}>
                  {l.title}
                </option>
              ))}
            </select>
          </label>

          <label className="block">
            <span className="mb-1 block text-[11px] uppercase tracking-wide text-ink-500">
              From stage <span className="text-ink-500">(optional)</span>
            </span>
            <select
              value={fromStage}
              onChange={(e) => setFromStage(e.target.value)}
              disabled={!detail}
              className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100 disabled:opacity-50"
            >
              <option value="">Full pipeline</option>
              {(detail?.stage_order ?? []).map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </label>

          <label className="flex items-end gap-2 pb-1.5 text-[13px] text-ink-300">
            <input
              type="checkbox"
              checked={force}
              onChange={(e) => setForce(e.target.checked)}
              className="accent-sky-500"
            />
            Force (ignore up-to-date cache)
          </label>
        </div>

        {startErr && <div className="mt-3"><ErrorNote error={startErr} /></div>}

        <div className="mt-4 flex items-center gap-3">
          <button
            onClick={() => void start()}
            disabled={!slug || !lessonId || gen.run.running}
            className="rounded bg-sky-600 px-3 py-1 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
          >
            {gen.run.running ? "Running…" : "Start"}
          </button>
          {gen.run.running && (
            <button
              onClick={() => void gen.cancel()}
              className="rounded border border-red-500/40 bg-red-500/10 px-3 py-1 text-red-300 hover:bg-red-500/20"
            >
              Cancel
            </button>
          )}
          <span className="text-[11px] text-ink-500">events: {gen.connection}</span>
        </div>
      </div>

      {/* Progress */}
      {gen.run.running && (
        <div className="mb-5">
          <div className="mb-1 flex justify-between text-[12px] text-ink-400">
            <span>
              {gen.run.course}/{gen.run.lesson}
              {gen.activeStage ? ` · ${gen.activeStage}` : ""}
            </span>
            <span>{gen.percent}%</span>
          </div>
          <div className="h-2 overflow-hidden rounded bg-ink-800">
            <div
              className="h-full bg-sky-500 transition-all"
              style={{ width: `${gen.percent}%` }}
            />
          </div>
        </div>
      )}

      {gen.lastError && <ErrorNote error={gen.lastError} />}

      {/* Logs */}
      <div className="mb-2 text-[11px] uppercase tracking-wide text-ink-500">Logs</div>
      <div className="surface-dark h-[360px] overflow-auto rounded-lg border border-ink-800 bg-ink-950 p-3 font-mono text-[12px] text-ink-300">
        {gen.logs.length === 0 ? (
          <div className="text-ink-500">No output yet — start a run to stream logs.</div>
        ) : (
          gen.logs.map((l) => (
            <div key={l.seq} className="whitespace-pre-wrap">
              {l.stage ? <span className="text-ink-500">[{l.stage}] </span> : null}
              {l.line}
            </div>
          ))
        )}
        <div ref={logEndRef} />
      </div>

      {slug && lessonId && (
        <div className="mt-4 text-[12px] text-ink-500">
          When it finishes, browse outputs in the{" "}
          <Link
            className="text-sky-300 hover:underline"
            to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(lessonId)}/results`}
          >
            results gallery
          </Link>
          .
        </div>
      )}
    </div>
  );
}
