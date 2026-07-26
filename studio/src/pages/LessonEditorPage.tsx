import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { ErrorNote } from "../components/ErrorNote";
import { DownloadButton } from "../components/DownloadButton";

type PreviewTab = "preview" | "quiz" | "artifacts";

/** Split-pane lesson editor: lesson.md source on the left, a live view of the
 *  last generated output (video/quiz/artifacts) on the right. */
export function LessonEditorPage() {
  const { slug = "", id = "" } = useParams();
  const { data, loading, error, reload } = useFetch(() => api.lesson(slug, id), [slug, id]);

  const [source, setSource] = useState("");
  const [dirty, setDirty] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveErr, setSaveErr] = useState<string | null>(null);
  const [tab, setTab] = useState<PreviewTab>("preview");

  useEffect(() => {
    if (data) {
      setSource(data.source);
      setDirty(false);
    }
  }, [data]);

  const save = async () => {
    setSaving(true);
    setSaveErr(null);
    try {
      await api.updateLesson(slug, id, { source });
      setDirty(false);
      reload();
    } catch (err) {
      setSaveErr(err instanceof Error ? err.message : String(err));
    } finally {
      setSaving(false);
    }
  };

  const video = (data?.artifacts ?? []).find((a) => /\.(mp4|mov|webm)$/i.test(a.name));
  const quiz = data?.quiz;

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex items-center gap-3 border-b border-ink-800 px-4 py-2">
        <div className="min-w-0 text-[12px] text-ink-500">
          <Link className="hover:underline" to={`/c/${encodeURIComponent(slug)}/edit`}>
            {slug}
          </Link>{" "}
          / <span className="text-ink-300">{data?.title ?? id}</span>
        </div>
        <div className="ml-auto flex items-center gap-2">
          {dirty && <span className="text-[11px] text-amber-400">unsaved</span>}
          <button
            onClick={() => void save()}
            disabled={!dirty || saving}
            className="rounded bg-sky-600 px-3 py-1 text-[13px] font-medium text-white hover:bg-sky-500 disabled:opacity-40"
          >
            {saving ? "Saving…" : "Save"}
          </button>
          <Link
            to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(id)}`}
            className="rounded border border-ink-700 bg-ink-800 px-3 py-1 text-[13px] text-ink-200 hover:bg-ink-700"
          >
            Workbench →
          </Link>
        </div>
      </div>

      {error && <div className="p-4"><ErrorNote error={error} onRetry={reload} /></div>}
      {loading && !data && <div className="p-4 text-ink-500">Loading…</div>}

      {data && (
        <div className="grid min-h-0 flex-1 grid-cols-1 md:grid-cols-2">
          {/* Left: source */}
          <div className="flex min-h-0 flex-col border-r border-ink-800">
            <textarea
              value={source}
              onChange={(e) => {
                setSource(e.target.value);
                setDirty(true);
              }}
              spellCheck={false}
              className="min-h-0 flex-1 resize-none bg-ink-950 p-4 font-mono text-[13px] leading-relaxed text-ink-200 outline-none"
            />
            {saveErr && <div className="p-2"><ErrorNote error={saveErr} /></div>}
          </div>

          {/* Right: preview */}
          <div className="flex min-h-0 flex-col overflow-auto">
            <div className="flex gap-1 border-b border-ink-800 px-3 py-2">
              {(["preview", "quiz", "artifacts"] as PreviewTab[]).map((t) => (
                <button
                  key={t}
                  onClick={() => setTab(t)}
                  className={
                    "rounded px-3 py-1 text-[12px] capitalize " +
                    (tab === t ? "bg-ink-800 text-ink-100" : "text-ink-400 hover:text-ink-200")
                  }
                >
                  {t}
                </button>
              ))}
            </div>

            <div className="min-h-0 flex-1 overflow-auto p-4">
              {tab === "preview" &&
                (video ? (
                  <video src={video.url} controls className="w-full rounded bg-black" preload="metadata" />
                ) : (
                  <EmptyHint slug={slug} id={id} what="No rendered video yet" />
                ))}

              {tab === "quiz" &&
                (quiz && quiz.questions.length > 0 ? (
                  <div className="space-y-3">
                    {quiz.questions.map((q, i) => (
                      <div key={q.id} className="rounded-lg border border-ink-800 bg-ink-900 p-3">
                        <div className="mb-1 flex items-center gap-2">
                          <span className="rounded bg-ink-800 px-1.5 py-0.5 text-[10px] uppercase text-ink-400">
                            {q.type}
                          </span>
                          <span className="text-[13px] text-ink-100">
                            Q{i + 1}. {q.prompt}
                          </span>
                        </div>
                        <ul className="ml-4 list-disc text-[12px] text-ink-400">
                          {q.options.map((o, oi) => (
                            <li key={oi} className={oi === q.answer_index ? "text-emerald-400" : ""}>
                              {o}
                            </li>
                          ))}
                        </ul>
                      </div>
                    ))}
                  </div>
                ) : (
                  <EmptyHint slug={slug} id={id} what="No quiz generated yet" />
                ))}

              {tab === "artifacts" &&
                ((data.artifacts ?? []).length > 0 ? (
                  <div className="space-y-2">
                    {data.artifacts.map((a) => (
                      <div
                        key={a.name}
                        className="flex items-center justify-between gap-2 rounded border border-ink-800 bg-ink-900 px-3 py-1.5"
                      >
                        <span className="truncate font-mono text-[12px] text-ink-300">{a.name}</span>
                        <DownloadButton url={a.url} filename={a.name.split("/").pop()} size={a.size} />
                      </div>
                    ))}
                  </div>
                ) : (
                  <EmptyHint slug={slug} id={id} what="No artifacts yet" />
                ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function EmptyHint({ slug, id, what }: { slug: string; id: string; what: string }) {
  return (
    <div className="text-[13px] text-ink-500">
      {what}. Run the pipeline from the{" "}
      <Link
        className="text-sky-300 hover:underline"
        to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(id)}`}
      >
        workbench
      </Link>{" "}
      or the{" "}
      <Link className="text-sky-300 hover:underline" to="/generation">
        Generation
      </Link>{" "}
      page.
    </div>
  );
}
