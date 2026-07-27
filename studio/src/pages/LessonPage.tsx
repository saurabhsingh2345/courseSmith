import { useState, type ReactNode } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type ArtifactFile } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { useScreenShortcuts } from "../state/ShortcutContext";
import {
  useRun,
  stageKey,
  lessonMatches,
  type LiveStageStatus,
} from "../state/RunContext";
import { StatusChip, type ChipStatus } from "../components/StatusChip";
import { ErrorNote } from "../components/ErrorNote";
import { formatBytes, formatMs } from "../lib/format";

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

export function LessonPage() {
  const { slug = "", id = "" } = useParams();
  const { run, liveStages, refreshTick, startRun } = useRun();
  const { data, loading, error, reload } = useFetch(
    () => api.lesson(slug, id),
    [slug, id, refreshTick],
  );

  const [fromStage, setFromStage] = useState("");
  const [force, setForce] = useState(false);
  const [note, setNote] = useState("");
  const [noteMsg, setNoteMsg] = useState<string | null>(null);

  useScreenShortcuts([{ keys: "r", label: "reload" }], (e) => {
    if (e.key === "r") reload();
  });

  const doRun = () => {
    if (run.running) return;
    void startRun({ course: slug, lesson: id, stage: fromStage || undefined, force });
  };

  const submitNote = async () => {
    if (!note.trim()) return;
    setNoteMsg(null);
    try {
      await api.feedback({ course: slug, lesson: id, note: note.trim() });
      setNote("");
      setNoteMsg("Saved to review-notes.yaml");
    } catch (err) {
      setNoteMsg(err instanceof Error ? err.message : String(err));
    }
  };

  return (
    <div className="mx-auto max-w-4xl p-6">
      <div className="mb-4 flex items-center gap-2 text-ink-500">
        <Link to="/" className="hover:text-ink-200">
          Courses
        </Link>
        <span>/</span>
        <Link to={`/c/${encodeURIComponent(slug)}`} className="hover:text-ink-200">
          {slug}
        </Link>
        <span>/</span>
        <span className="text-ink-200">{data?.id ?? id}</span>
      </div>

      {error && <ErrorNote error={error} onRetry={reload} />}
      {loading && !data && <div className="text-ink-500">Loading…</div>}

      {data && (
        <>
          <div className="mb-3 flex items-center gap-3">
            <h1 className="text-lg font-semibold text-ink-100">{data.title}</h1>
            <Link
              to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(id)}/preview`}
              className="rounded border border-ink-700 bg-ink-800 px-2 py-0.5 text-[12px] text-sky-300 hover:bg-ink-700"
            >
              Preview →
            </Link>
            <Link
              to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(id)}/quiz`}
              className="rounded border border-ink-700 bg-ink-800 px-2 py-0.5 text-[12px] text-sky-300 hover:bg-ink-700"
            >
              Edit quiz →
            </Link>
            <Link
              to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(id)}/edit`}
              className="rounded border border-ink-700 bg-ink-800 px-2 py-0.5 text-[12px] text-sky-300 hover:bg-ink-700"
            >
              Edit source →
            </Link>
            <Link
              to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(id)}/strategy`}
              className="rounded border border-ink-700 bg-ink-800 px-2 py-0.5 text-[12px] text-sky-300 hover:bg-ink-700"
            >
              Quiz strategy →
            </Link>
            <Link
              to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(id)}/results`}
              className="rounded border border-ink-700 bg-ink-800 px-2 py-0.5 text-[12px] text-sky-300 hover:bg-ink-700"
            >
              Results →
            </Link>
          </div>

          {/* Stage chips */}
          <div className="mb-5 flex flex-wrap gap-1.5">
            {data.stage_order.map((stage) => {
              const status: ChipStatus =
                liveStatusFor(liveStages, slug, id, stage) ??
                ((data.stages[stage] ?? "pending") as ChipStatus);
              return <StatusChip key={stage} status={status} label={stage} title={`${stage}: ${status}`} />;
            })}
          </div>

          {/* Run controls */}
          <div className="mb-6 flex flex-wrap items-center gap-3 rounded-lg border border-ink-800 bg-ink-900 p-3">
            <label className="flex items-center gap-1.5 text-ink-400">
              from
              <select
                value={fromStage}
                onChange={(e) => setFromStage(e.target.value)}
                className="rounded border border-ink-700 bg-ink-800 px-2 py-1 text-ink-200"
              >
                <option value="">all stages</option>
                {data.stage_order.map((s) => (
                  <option key={s} value={s}>
                    {s}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex items-center gap-1.5 text-ink-400">
              <input
                type="checkbox"
                checked={force}
                onChange={(e) => setForce(e.target.checked)}
              />
              force
            </label>
            <button
              onClick={doRun}
              disabled={run.running}
              className="rounded bg-sky-600 px-3 py-1 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
            >
              {run.running ? "Running…" : "Run"}
            </button>
          </div>

          {/* The finished video + its per-section clips, front and centre. */}
          <VideoDownloads artifacts={data.artifacts} />

          {/* Artifacts */}
          <Section title={`Artifacts (${data.artifacts.length})`}>
            {data.artifacts.length === 0 ? (
              <div className="text-ink-500">None yet.</div>
            ) : (
              <ul className="divide-y divide-ink-850 rounded-lg border border-ink-800">
                {data.artifacts.map((a) => (
                  <li key={a.name} className="flex items-center justify-between px-3 py-1.5">
                    <a
                      href={a.url}
                      target="_blank"
                      rel="noreferrer"
                      className="font-mono text-[12px] text-sky-300 hover:underline"
                    >
                      {a.name}
                    </a>
                    <span className="text-[11px] text-ink-500">{formatBytes(a.size)}</span>
                  </li>
                ))}
              </ul>
            )}
          </Section>

          {/* Script */}
          {data.script && (
            <Section title="Script">
              <div className="space-y-3">
                {data.script.sections.map((sec) => (
                  <div key={sec.id} className="rounded-lg border border-ink-800 bg-ink-900 p-3">
                    <div className="mb-1 flex items-center gap-2 text-[11px] text-ink-500">
                      <span className="font-mono">{sec.id}</span>
                      {sec.duration_est_sec !== undefined && (
                        <span>~{formatMs(sec.duration_est_sec * 1000)}</span>
                      )}
                    </div>
                    <p className="whitespace-pre-wrap text-ink-300">{sec.narration}</p>
                  </div>
                ))}
              </div>
            </Section>
          )}

          {/* Chapters */}
          {data.chapters && data.chapters.length > 0 && (
            <Section title="Chapters">
              <ul className="rounded-lg border border-ink-800">
                {data.chapters.map((ch) => (
                  <li
                    key={ch.id}
                    className="flex items-center justify-between border-b border-ink-850 px-3 py-1.5 last:border-0"
                  >
                    <span className="text-ink-200">{ch.title}</span>
                    <span className="font-mono text-[11px] text-ink-500">
                      {formatMs(ch.start_ms)}–{formatMs(ch.end_ms)}
                    </span>
                  </li>
                ))}
              </ul>
            </Section>
          )}

          {/* Feedback */}
          <Section title="Review note">
            <textarea
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Leave a note — appended to review-notes.yaml for the next revision pass."
              rows={3}
              className="w-full rounded-lg border border-ink-800 bg-ink-900 p-2 text-ink-200 placeholder:text-ink-500"
            />
            <div className="mt-2 flex items-center gap-3">
              <button
                onClick={() => void submitNote()}
                disabled={!note.trim()}
                className="rounded border border-ink-700 bg-ink-800 px-3 py-1 text-ink-200 hover:bg-ink-700 disabled:opacity-40"
              >
                Save note
              </button>
              {noteMsg && <span className="text-[11px] text-ink-400">{noteMsg}</span>}
            </div>
          </Section>
        </>
      )}
    </div>
  );
}

/**
 * The lesson video and its per-section clips as first-class downloads: a
 * player for final.mp4 with a download link, then one row per section chunk.
 *
 * The saved name is the server's `download_name`, not the URL's last segment —
 * on disk every lesson's video is `final.mp4`, which is a fine pipeline
 * contract and a terrible thing to have six of in one folder.
 */
function VideoDownloads({ artifacts }: { artifacts: ArtifactFile[] }) {
  const final = artifacts.find((a) => a.name === "final.mp4");
  const sections = artifacts
    .filter((a) => a.name.startsWith("sections/") && a.name.endsWith(".mp4"))
    .sort((a, b) => a.name.localeCompare(b.name));
  if (!final && sections.length === 0) return null;

  return (
    <Section title="Video">
      {final && (
        <div className="mb-3 overflow-hidden rounded-lg border border-ink-800">
          <video src={final.url} controls preload="metadata" className="max-h-80 w-full bg-black" />
          <div className="flex items-center justify-between bg-ink-900 px-3 py-2">
            <span className="text-ink-300">Full lesson</span>
            <a
              href={final.url}
              download={final.download_name}
              className="rounded bg-sky-600 px-3 py-1 font-medium text-white hover:bg-sky-500"
            >
              Download ({formatBytes(final.size)})
            </a>
          </div>
        </div>
      )}
      {sections.length > 0 && (
        <ul className="divide-y divide-ink-850 rounded-lg border border-ink-800">
          {sections.map((a) => {
            const base = a.name.replace(/^sections\//, "").replace(/\.mp4$/, "");
            const title = base.replace(/^\d+-/, "").replace(/-/g, " ");
            return (
              <li key={a.name} className="flex items-center justify-between px-3 py-1.5">
                <span className="text-ink-300">
                  <span className="mr-2 font-mono text-[11px] text-ink-500">
                    {base.slice(0, 2)}
                  </span>
                  {title}
                </span>
                <a
                  href={a.url}
                  download={a.download_name}
                  className="text-[12px] text-sky-300 hover:underline"
                >
                  download ({formatBytes(a.size)})
                </a>
              </li>
            );
          })}
        </ul>
      )}
    </Section>
  );
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="mb-6">
      <h2 className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-ink-500">
        {title}
      </h2>
      {children}
    </section>
  );
}
