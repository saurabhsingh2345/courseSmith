import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { useScreenShortcuts } from "../state/ShortcutContext";
import { ErrorNote } from "../components/ErrorNote";

export function CoursesPage() {
  const { data, loading, error, reload } = useFetch(() => api.courses(), []);
  const navigate = useNavigate();

  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [creating, setCreating] = useState(false);
  const [createErr, setCreateErr] = useState<string | null>(null);

  useScreenShortcuts(
    [
      { keys: "n", label: "new course" },
      { keys: "r", label: "reload" },
    ],
    (e) => {
      if (e.key === "r") reload();
      else if (e.key === "n") setShowForm(true);
    },
  );

  const create = async () => {
    if (!name.trim()) return;
    setCreating(true);
    setCreateErr(null);
    try {
      const course = await api.createCourse({
        name: name.trim(),
        description: description.trim() || undefined,
      });
      navigate(`/c/${encodeURIComponent(course.slug)}`);
    } catch (err) {
      setCreateErr(err instanceof Error ? err.message : String(err));
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="mx-auto max-w-4xl p-6">
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-lg font-semibold text-ink-100">Courses</h1>
        <button
          onClick={() => setShowForm((v) => !v)}
          className="rounded border border-ink-700 bg-ink-800 px-3 py-1 text-ink-200 hover:bg-ink-700"
        >
          {showForm ? "Cancel" : "+ New course"}
        </button>
      </div>

      {showForm && (
        <div className="mb-5 rounded-lg border border-ink-800 bg-ink-900 p-4">
          <div className="mb-3">
            <label className="mb-1 block text-[11px] uppercase tracking-wide text-ink-500">
              Name
            </label>
            <input
              autoFocus
              value={name}
              onChange={(e) => setName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") void create();
              }}
              placeholder="e.g. Bash for Beginners"
              className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100 placeholder:text-ink-600"
            />
            {name.trim() && (
              <div className="mt-1 text-[11px] text-ink-500">
                slug: <span className="font-mono text-ink-400">{slugPreview(name)}</span>
              </div>
            )}
          </div>
          <div className="mb-3">
            <label className="mb-1 block text-[11px] uppercase tracking-wide text-ink-500">
              Description <span className="text-ink-600">(optional)</span>
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              placeholder="One or two sentences about the course."
              className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-200 placeholder:text-ink-600"
            />
          </div>
          {createErr && <ErrorNote error={createErr} />}
          <div className="flex items-center gap-3">
            <button
              onClick={() => void create()}
              disabled={!name.trim() || creating}
              className="rounded bg-sky-600 px-3 py-1 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
            >
              {creating ? "Creating…" : "Create course"}
            </button>
            <span className="text-[11px] text-ink-500">
              Scaffolds course.yaml + an example lesson, then opens it.
            </span>
          </div>
        </div>
      )}

      {error && <ErrorNote error={error} onRetry={reload} />}
      {loading && !data && <div className="text-ink-500">Loading…</div>}
      <div className="grid gap-3 sm:grid-cols-2">
        {(data ?? []).map((c) => (
          <Link
            key={c.slug}
            to={`/c/${encodeURIComponent(c.slug)}`}
            className="rounded-lg border border-ink-800 bg-ink-900 p-4 transition-colors hover:border-ink-600 hover:bg-ink-850"
          >
            <div className="font-medium text-ink-100">{c.name}</div>
            <div className="mt-1 line-clamp-2 text-ink-400">{c.description}</div>
            <div className="mt-2 text-[11px] text-ink-500">
              {c.lesson_count} lesson{c.lesson_count === 1 ? "" : "s"}
            </div>
          </Link>
        ))}
      </div>
      {data && data.length === 0 && !loading && !showForm && (
        <div className="text-ink-500">
          No courses yet.{" "}
          <button onClick={() => setShowForm(true)} className="text-sky-300 hover:underline">
            Create one
          </button>
          .
        </div>
      )}
    </div>
  );
}

/** Mirror the backend slugify so the user sees the resulting slug live. */
function slugPreview(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}
