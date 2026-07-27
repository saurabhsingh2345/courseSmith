import { useEffect, useState, type ReactNode } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api, type UpdateCourseRequest } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { ErrorNote } from "../components/ErrorNote";

/** Edit course metadata (name, description, archetype/animation/palette,
 *  brand colours) and manage its lessons. Writes go through PUT /api/courses. */
export function CourseEditorPage() {
  const { slug = "" } = useParams();
  const navigate = useNavigate();
  const { data, loading, error, reload } = useFetch(() => api.course(slug), [slug]);
  const { data: catalog } = useFetch(() => api.archetypes(), []);

  const [form, setForm] = useState<UpdateCourseRequest>({});
  const [saving, setSaving] = useState(false);
  const [saveMsg, setSaveMsg] = useState<string | null>(null);
  const [saveErr, setSaveErr] = useState<string | null>(null);
  const [newLessonTitle, setNewLessonTitle] = useState("");

  // Seed the form once the course loads.
  useEffect(() => {
    if (!data) return;
    setForm({
      name: data.name,
      description: data.description,
      archetype: data.meta.archetype,
      animation_style: data.meta.animation_style,
      color_palette: data.meta.color_palette,
      colors: { ...data.meta.colors },
    });
  }, [data]);

  const set = <K extends keyof UpdateCourseRequest>(key: K, value: UpdateCourseRequest[K]) =>
    setForm((f) => ({ ...f, [key]: value }));

  const save = async () => {
    setSaving(true);
    setSaveMsg(null);
    setSaveErr(null);
    try {
      await api.updateCourse(slug, form);
      setSaveMsg("Saved");
      reload();
    } catch (err) {
      setSaveErr(err instanceof Error ? err.message : String(err));
    } finally {
      setSaving(false);
    }
  };

  const addLesson = async () => {
    if (!newLessonTitle.trim()) return;
    try {
      const lesson = await api.createLesson(slug, { title: newLessonTitle.trim() });
      setNewLessonTitle("");
      navigate(`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(lesson.id)}/edit`);
    } catch (err) {
      setSaveErr(err instanceof Error ? err.message : String(err));
    }
  };

  const removeLesson = async (id: string) => {
    if (!confirm(`Delete lesson ${id}? This removes its source and generated files.`)) return;
    try {
      await api.deleteLesson(slug, id);
      reload();
    } catch (err) {
      setSaveErr(err instanceof Error ? err.message : String(err));
    }
  };

  const removeCourse = async () => {
    if (!confirm(`Delete course "${data?.name ?? slug}" and all its lessons?`)) return;
    try {
      await api.deleteCourse(slug);
      navigate("/");
    } catch (err) {
      setSaveErr(err instanceof Error ? err.message : String(err));
    }
  };

  return (
    <div className="mx-auto max-w-4xl p-6">
      <div className="mb-1 text-[12px] text-ink-500">
        <Link className="hover:underline" to={`/c/${encodeURIComponent(slug)}`}>
          {slug}
        </Link>{" "}
        / <span className="text-ink-400">edit</span>
      </div>
      <h1 className="mb-4 text-lg font-semibold text-ink-100">Course settings</h1>

      {error && <ErrorNote error={error} onRetry={reload} />}
      {loading && !data && <div className="text-ink-500">Loading…</div>}

      {data && (
        <div className="grid gap-6 md:grid-cols-2">
          {/* Left: metadata */}
          <div className="space-y-4">
            <Field label="Name">
              <input
                value={form.name ?? ""}
                onChange={(e) => set("name", e.target.value)}
                className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100"
              />
            </Field>
            <Field label="Description">
              <textarea
                value={form.description ?? ""}
                onChange={(e) => set("description", e.target.value)}
                rows={3}
                className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-200"
              />
            </Field>
            <Field label="Archetype">
              <select
                value={form.archetype ?? ""}
                onChange={(e) => set("archetype", e.target.value)}
                className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100"
              >
                <option value="">— none —</option>
                {(catalog?.archetypes ?? []).map((a) => (
                  <option key={a.name} value={a.name}>
                    {a.name} — {a.description}
                  </option>
                ))}
              </select>
            </Field>
            <Field label="Animation style">
              <select
                value={form.animation_style ?? ""}
                onChange={(e) => set("animation_style", e.target.value)}
                className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100"
              >
                <option value="">— archetype default —</option>
                {(catalog?.animation_styles ?? []).map((s) => (
                  <option key={s} value={s}>
                    {s}
                  </option>
                ))}
              </select>
            </Field>
            <Field label="Color palette">
              <select
                value={form.color_palette ?? ""}
                onChange={(e) => set("color_palette", e.target.value)}
                className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100"
              >
                <option value="">— custom colours below —</option>
                {(catalog?.palettes ?? []).map((p) => (
                  <option key={p.name} value={p.name}>
                    {p.name}
                  </option>
                ))}
              </select>
            </Field>

            <Field label="Brand colours">
              <div className="flex gap-4">
                {(["primary", "accent", "background"] as const).map((k) => (
                  <label key={k} className="flex flex-col items-center gap-1 text-[11px] text-ink-500">
                    {k}
                    <input
                      type="color"
                      value={form.colors?.[k] || "#000000"}
                      onChange={(e) =>
                        set("colors", { ...(form.colors ?? {}), [k]: e.target.value })
                      }
                      className="h-8 w-12 rounded border border-ink-800 bg-transparent"
                    />
                  </label>
                ))}
              </div>
            </Field>

            {saveErr && <ErrorNote error={saveErr} />}
            <div className="flex items-center gap-3">
              <button
                onClick={() => void save()}
                disabled={saving}
                className="rounded bg-sky-600 px-3 py-1 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
              >
                {saving ? "Saving…" : "Save"}
              </button>
              {saveMsg && <span className="text-[12px] text-emerald-400">{saveMsg}</span>}
              <button
                onClick={() => void removeCourse()}
                className="ml-auto rounded border border-red-500/40 bg-red-500/10 px-3 py-1 text-[13px] text-red-300 hover:bg-red-500/20"
              >
                Delete course
              </button>
            </div>
          </div>

          {/* Right: lessons */}
          <div>
            <h2 className="mb-2 text-[11px] uppercase tracking-wide text-ink-500">
              Lessons ({data.lessons.length})
            </h2>
            <div className="mb-3 flex gap-2">
              <input
                value={newLessonTitle}
                onChange={(e) => setNewLessonTitle(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") void addLesson();
                }}
                placeholder="New lesson title"
                className="flex-1 rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100 placeholder:text-ink-500"
              />
              <button
                onClick={() => void addLesson()}
                disabled={!newLessonTitle.trim()}
                className="rounded border border-ink-700 bg-ink-800 px-3 py-1 text-ink-200 hover:bg-ink-700 disabled:opacity-40"
              >
                Add
              </button>
            </div>
            <div className="space-y-2">
              {data.lessons.map((l) => (
                <div
                  key={l.id}
                  className="flex items-center justify-between rounded-lg border border-ink-800 bg-ink-900 px-3 py-2"
                >
                  <div className="min-w-0">
                    <div className="truncate text-ink-100">{l.title}</div>
                    <div className="font-mono text-[11px] text-ink-500">{l.id}</div>
                  </div>
                  <div className="flex shrink-0 gap-2">
                    <Link
                      to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(l.id)}/edit`}
                      className="rounded border border-ink-700 bg-ink-800 px-2 py-0.5 text-[12px] text-ink-200 hover:bg-ink-700"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => void removeLesson(l.id)}
                      className="rounded border border-red-500/30 px-2 py-0.5 text-[12px] text-red-300 hover:bg-red-500/10"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1 block text-[11px] uppercase tracking-wide text-ink-500">{label}</span>
      {children}
    </label>
  );
}
