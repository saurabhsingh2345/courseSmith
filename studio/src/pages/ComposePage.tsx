import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, type DraftDetail, type DraftMeta } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { useRun } from "../state/RunContext";
import { useScreenShortcuts } from "../state/ShortcutContext";
import { ErrorNote } from "../components/ErrorNote";

// ComposePage is the front door: one prompt, one button, and the lesson is
// drafting, filed into the chosen course, and generating its video. The
// review-first path still exists ("Draft only") — the draft parks unfiled
// below so it can be read and edited before committing.

export function ComposePage() {
  const drafts = useFetch(() => api.drafts(), []);
  const courses = useFetch(() => api.courses(), []);
  const { run, startRun } = useRun();
  const navigate = useNavigate();

  const [prompt, setPrompt] = useState("");
  const [scope, setScope] = useState("");
  const [busy, setBusy] = useState<"create" | "draft" | null>(null);
  const [phase, setPhase] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [open, setOpen] = useState<DraftDetail | null>(null);

  useScreenShortcuts([{ keys: "r", label: "reload" }], (e) => {
    if (e.key === "r") {
      drafts.reload();
    }
  });

  // The one-shot path: draft → file into the course → start the pipeline →
  // land on the lesson page, which shows live stage progress over SSE.
  const createAndGenerate = async () => {
    const text = prompt.trim();
    if (!text || !scope) return;
    setBusy("create");
    setError(null);
    try {
      setPhase("Drafting lesson…");
      const draft = await api.createDraft({ prompt: text, course: scope });
      setPhase("Filing into course…");
      const lesson = await api.assignDraft(draft.id, { course: scope });
      setPhase("Starting generation…");
      await startRun({ course: lesson.course, lesson: lesson.id });
      setPrompt("");
      navigate(
        `/c/${encodeURIComponent(lesson.course)}/l/${encodeURIComponent(lesson.id)}`,
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(null);
      setPhase("");
    }
  };

  // The review-first path: draft only, parked unfiled below.
  const draftOnly = async () => {
    const text = prompt.trim();
    if (!text) return;
    setBusy("draft");
    setError(null);
    try {
      const draft = await api.createDraft({ prompt: text, course: scope || undefined });
      setPrompt("");
      setOpen(draft);
      drafts.reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(null);
    }
  };

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="text-lg font-semibold text-ink-100">Make a lesson</h1>
      <p className="mt-1 text-ink-400">
        Say what you want taught, pick the course, and the video starts generating.
      </p>

      <div className="mt-4 rounded-lg border border-ink-800 bg-ink-900 p-4">
        <textarea
          autoFocus
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyDown={(e) => {
            // Enter creates; Shift+Enter is a newline.
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              void createAndGenerate();
            }
          }}
          rows={3}
          placeholder="e.g. make a lesson on loops in python"
          className="w-full resize-none rounded border border-ink-800 bg-ink-950 px-3 py-2 text-ink-100 placeholder:text-ink-600"
        />
        <div className="mt-3 flex flex-wrap items-center gap-3">
          <select
            value={scope}
            onChange={(e) => setScope(e.target.value)}
            className="rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-200"
          >
            <option value="">choose a course…</option>
            {(courses.data ?? []).map((c) => (
              <option key={c.slug} value={c.slug}>
                {c.name}
              </option>
            ))}
          </select>
          <button
            onClick={() => void createAndGenerate()}
            disabled={!prompt.trim() || !scope || busy !== null || run.running}
            title={
              run.running
                ? "A generation run is already in progress"
                : !scope
                  ? "Pick the course it goes in"
                  : undefined
            }
            className="rounded bg-sky-600 px-4 py-1.5 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
          >
            {busy === "create" ? phase || "Working…" : "Create lesson →"}
          </button>
          <button
            onClick={() => void draftOnly()}
            disabled={!prompt.trim() || busy !== null}
            className="rounded border border-ink-700 px-3 py-1.5 text-ink-300 hover:bg-ink-800 disabled:opacity-40"
          >
            {busy === "draft" ? "Drafting…" : "Draft only"}
          </button>
          <span className="text-[11px] text-ink-600">
            Draft only parks it below for review; nothing generates until you file it.
          </span>
        </div>
        {error && (
          <div className="mt-3">
            <ErrorNote error={error} />
          </div>
        )}
      </div>

      <h2 className="mt-8 text-ink-200">Unfiled drafts</h2>
      {drafts.error && <ErrorNote error={drafts.error} onRetry={drafts.reload} />}
      {drafts.loading && !drafts.data && <div className="mt-2 text-ink-500">Loading…</div>}
      {drafts.data && drafts.data.length === 0 && (
        <div className="mt-2 text-ink-500">
          Nothing unfiled. Drafts wait here until you file them into a course.
        </div>
      )}
      <div className="mt-3 grid gap-3">
        {(drafts.data ?? []).map((d) => (
          <DraftCard
            key={d.id}
            draft={d}
            courses={(courses.data ?? []).map((c) => ({ slug: c.slug, name: c.name }))}
            onOpen={async () => setOpen(await api.draft(d.id))}
            onChanged={() => {
              drafts.reload();
              courses.reload();
            }}
          />
        ))}
      </div>

      {open && (
        <DraftSheet
          draft={open}
          onClose={() => setOpen(null)}
          onChanged={() => {
            drafts.reload();
            courses.reload();
          }}
        />
      )}
    </div>
  );
}

type CourseOption = { slug: string; name: string };

function DraftCard({
  draft,
  courses,
  onOpen,
  onChanged,
}: {
  draft: DraftMeta;
  courses: CourseOption[];
  onOpen: () => void;
  onChanged: () => void;
}) {
  return (
    <div className="rounded-lg border border-ink-800 bg-ink-900 p-4">
      <div className="flex items-start justify-between gap-4">
        <button onClick={onOpen} className="min-w-0 text-left">
          <div className="font-medium text-ink-100 hover:underline">{draft.title}</div>
          {draft.summary && <div className="mt-1 text-ink-400">{draft.summary}</div>}
          <div className="mt-2 truncate text-[11px] text-ink-600">“{draft.prompt}”</div>
        </button>
        <FilePicker draft={draft} courses={courses} onChanged={onChanged} />
      </div>
    </div>
  );
}

/** The "which vault does this go in" control. */
function FilePicker({
  draft,
  courses,
  onChanged,
}: {
  draft: DraftMeta;
  courses: CourseOption[];
  onChanged: () => void;
}) {
  const navigate = useNavigate();
  const [busy, setBusy] = useState(false);
  const [newCourse, setNewCourse] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const file = async (req: { course?: string; new_course?: string }) => {
    setBusy(true);
    setError(null);
    try {
      const lesson = await api.assignDraft(draft.id, req);
      onChanged();
      navigate(`/c/${encodeURIComponent(lesson.course)}/l/${encodeURIComponent(lesson.id)}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
      setBusy(false);
    }
  };

  if (newCourse !== null) {
    return (
      <div className="w-64 shrink-0">
        <input
          autoFocus
          value={newCourse}
          onChange={(e) => setNewCourse(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && newCourse.trim()) void file({ new_course: newCourse.trim() });
            if (e.key === "Escape") setNewCourse(null);
          }}
          placeholder="New course name"
          className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1 text-ink-100 placeholder:text-ink-600"
        />
        <div className="mt-2 flex gap-2">
          <button
            onClick={() => void file({ new_course: newCourse.trim() })}
            disabled={!newCourse.trim() || busy}
            className="rounded bg-sky-600 px-2 py-1 text-white hover:bg-sky-500 disabled:opacity-40"
          >
            {busy ? "Filing…" : "Create & file"}
          </button>
          <button onClick={() => setNewCourse(null)} className="px-2 py-1 text-ink-400">
            Cancel
          </button>
        </div>
        {error && <div className="mt-2 text-[11px] text-rose-400">{error}</div>}
      </div>
    );
  }

  return (
    <div className="w-56 shrink-0 text-right">
      <select
        disabled={busy}
        value=""
        onChange={(e) => {
          const v = e.target.value;
          if (v === "__new") setNewCourse("");
          else if (v) void file({ course: v });
        }}
        className="w-full rounded border border-ink-700 bg-ink-800 px-2 py-1 text-ink-200 disabled:opacity-40"
      >
        <option value="">{busy ? "Filing…" : "File into…"}</option>
        {courses.map((c) => (
          <option key={c.slug} value={c.slug}>
            {c.name}
          </option>
        ))}
        <option value="__new">+ New course…</option>
      </select>
      {error && <div className="mt-2 text-[11px] text-rose-400">{error}</div>}
    </div>
  );
}

/** Read and edit a draft's lesson.md before filing it. */
function DraftSheet({
  draft,
  onClose,
  onChanged,
}: {
  draft: DraftDetail;
  onClose: () => void;
  onChanged: () => void;
}) {
  const [source, setSource] = useState(draft.source);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const dirty = source !== draft.source;

  const save = async () => {
    setSaving(true);
    setError(null);
    try {
      await api.updateDraft(draft.id, source);
      onChanged();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setSaving(false);
    }
  };

  const discard = async () => {
    await api.deleteDraft(draft.id);
    onChanged();
    onClose();
  };

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center bg-black/60 p-6">
      <div className="flex max-h-full w-full max-w-3xl flex-col rounded-lg border border-ink-800 bg-ink-900">
        <div className="flex items-center justify-between border-b border-ink-800 p-4">
          <div className="min-w-0">
            <div className="truncate font-medium text-ink-100">{draft.title}</div>
            <div className="truncate text-[11px] text-ink-600">“{draft.prompt}”</div>
          </div>
          <button onClick={onClose} className="px-2 text-ink-400 hover:text-ink-200">
            Close
          </button>
        </div>
        <textarea
          value={source}
          onChange={(e) => setSource(e.target.value)}
          spellCheck={false}
          className="min-h-[24rem] flex-1 resize-none bg-ink-950 p-4 font-mono text-[13px] leading-relaxed text-ink-200"
        />
        {error && (
          <div className="px-4 pt-3">
            <ErrorNote error={error} />
          </div>
        )}
        <div className="flex items-center gap-3 border-t border-ink-800 p-4">
          <button
            onClick={() => void save()}
            disabled={!dirty || saving}
            className="rounded bg-sky-600 px-3 py-1 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
          >
            {saving ? "Saving…" : "Save changes"}
          </button>
          <button onClick={() => void discard()} className="text-ink-400 hover:text-rose-400">
            Discard draft
          </button>
          <span className="ml-auto text-[11px] text-ink-600">
            Close and use “File into…” to put it in a course.
          </span>
        </div>
      </div>
    </div>
  );
}
