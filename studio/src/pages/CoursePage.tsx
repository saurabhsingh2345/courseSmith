import { Link, useParams } from "react-router-dom";
import { api, type StageStatus } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { useScreenShortcuts } from "../state/ShortcutContext";
import {
  useRun,
  stageKey,
  lessonMatches,
  type LiveStageStatus,
} from "../state/RunContext";
import { chipDot, type ChipStatus } from "../components/StatusChip";
import { ErrorNote } from "../components/ErrorNote";

/** Live overlay for a cell, tolerant of short vs. full lesson ids. */
function liveStatusFor(
  live: Record<string, LiveStageStatus>,
  course: string,
  lessonId: string,
  stage: string,
): LiveStageStatus | undefined {
  const exact = live[stageKey(course, lessonId, stage)];
  if (exact) return exact;
  for (const [k, v] of Object.entries(live)) {
    const [c, l, s] = k.split("|");
    if (c === course && s === stage && lessonMatches(l, lessonId)) return v;
  }
  return undefined;
}

export function CoursePage() {
  const { slug = "" } = useParams();
  const { run, liveStages, refreshTick, startRun } = useRun();
  const { data, loading, error, reload } = useFetch(
    () => api.course(slug),
    [slug, refreshTick],
  );

  useScreenShortcuts([{ keys: "r", label: "reload" }], (e) => {
    if (e.key === "r") reload();
  });

  const runLesson = (lesson: string, stage?: string) => {
    if (run.running) return;
    void startRun({ course: slug, lesson, stage });
  };

  return (
    <div className="mx-auto max-w-6xl p-6">
      <div className="mb-4 flex items-center gap-2 text-ink-500">
        <Link to="/" className="hover:text-ink-200">
          Courses
        </Link>
        <span>/</span>
        <span className="text-ink-200">{data?.name ?? slug}</span>
        <Link
          to={`/c/${encodeURIComponent(slug)}/edit`}
          className="ml-auto rounded border border-ink-700 bg-ink-800 px-3 py-1 text-[13px] text-ink-200 hover:bg-ink-700"
        >
          Settings
        </Link>
      </div>

      {error && <ErrorNote error={error} onRetry={reload} />}
      {loading && !data && <div className="text-ink-500">Loading…</div>}

      {data && (
        <>
          {data.description && <p className="mb-4 max-w-2xl text-ink-400">{data.description}</p>}
          {data.lessons.length === 0 ? (
            <div className="text-ink-500">No lessons in this course.</div>
          ) : (
            <div className="overflow-x-auto rounded-lg border border-ink-800">
              <table className="w-full border-collapse text-left">
                <thead>
                  <tr className="border-b border-ink-800 bg-ink-900 text-[11px] uppercase tracking-wide text-ink-500">
                    <th className="sticky left-0 z-10 bg-ink-900 px-3 py-2 font-medium">Lesson</th>
                    {data.stage_order.map((s) => (
                      <th key={s} className="px-2 py-2 font-medium" title={s}>
                        {s}
                      </th>
                    ))}
                    <th className="px-3 py-2" />
                  </tr>
                </thead>
                <tbody>
                  {data.lessons.map((lesson) => (
                    <tr key={lesson.id} className="border-b border-ink-850 hover:bg-ink-900/50">
                      <td className="sticky left-0 z-10 bg-ink-950 px-3 py-2">
                        <Link
                          to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(lesson.id)}`}
                          className="text-ink-100 hover:text-sky-300"
                        >
                          <span className="text-ink-500">{lesson.id}</span> {lesson.title}
                        </Link>
                      </td>
                      {data.stage_order.map((stage) => {
                        const staticStatus = (lesson.stages[stage] ?? "pending") as StageStatus;
                        const live = liveStatusFor(liveStages, slug, lesson.id, stage);
                        const status: ChipStatus = live ?? staticStatus;
                        return (
                          <td key={stage} className="px-2 py-2 text-center">
                            <button
                              onClick={() => runLesson(lesson.id, stage)}
                              disabled={run.running}
                              title={`${stage}: ${status} — click to run this stage`}
                              className="inline-flex h-4 w-4 items-center justify-center rounded-full disabled:cursor-not-allowed"
                            >
                              <span className={`h-2 w-2 rounded-full ${chipDot[status]}`} />
                            </button>
                          </td>
                        );
                      })}
                      <td className="px-3 py-2 text-right">
                        <button
                          onClick={() => runLesson(lesson.id)}
                          disabled={run.running}
                          className="rounded border border-ink-700 bg-ink-800 px-2 py-0.5 text-[11px] text-ink-200 hover:bg-ink-700 disabled:opacity-40"
                        >
                          Run
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
          <div className="mt-3 flex flex-wrap gap-4 text-[11px] text-ink-500">
            {(["done", "stale", "pending", "running", "failed", "skipped"] as ChipStatus[]).map(
              (s) => (
                <span key={s} className="flex items-center gap-1">
                  <span className={`h-2 w-2 rounded-full ${chipDot[s]}`} /> {s}
                </span>
              ),
            )}
          </div>
        </>
      )}
    </div>
  );
}
