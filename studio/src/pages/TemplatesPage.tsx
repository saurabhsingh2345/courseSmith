import { api } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { ErrorNote } from "../components/ErrorNote";

/** Browse the archetype catalog: course shapes, animation philosophies, and
 *  brand palettes exposed by the Go pipeline registry. Read-only reference the
 *  course editor's selectors draw from. */
export function TemplatesPage() {
  const { data, loading, error, reload } = useFetch(() => api.archetypes(), []);

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="mb-1 text-lg font-semibold text-ink-100">Templates</h1>
      <p className="mb-5 text-[13px] text-ink-500">
        Archetypes, animation philosophies, and palettes. Pick these per course in{" "}
        <span className="text-ink-400">Course settings</span>.
      </p>

      {error && <ErrorNote error={error} onRetry={reload} />}
      {loading && !data && <div className="text-ink-500">Loading…</div>}

      {data && (
        <div className="space-y-8">
          <section>
            <h2 className="mb-3 text-[11px] uppercase tracking-wide text-ink-500">Archetypes</h2>
            <div className="grid gap-3 sm:grid-cols-2">
              {data.archetypes.map((a) => (
                <div key={a.name} className="rounded-lg border border-ink-800 bg-ink-900 p-4">
                  <div className="flex items-center justify-between">
                    <span className="font-medium text-ink-100">{a.name}</span>
                    <span className="rounded bg-ink-800 px-1.5 py-0.5 text-[10px] text-ink-400">
                      {a.default_animation} · {a.pace_hint} wpm
                    </span>
                  </div>
                  <div className="mt-1 text-[13px] text-ink-300">{a.description}</div>
                  <div className="mt-2 border-l-2 border-ink-700 pl-2 text-[12px] italic text-ink-500">
                    {a.prompt_hint}
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section>
            <h2 className="mb-3 text-[11px] uppercase tracking-wide text-ink-500">Animation styles</h2>
            <div className="flex flex-wrap gap-2">
              {data.animation_styles.map((s) => (
                <span
                  key={s}
                  className="rounded-full border border-ink-700 bg-ink-800 px-3 py-1 text-[13px] text-ink-200"
                >
                  {s}
                </span>
              ))}
            </div>
          </section>

          <section>
            <h2 className="mb-3 text-[11px] uppercase tracking-wide text-ink-500">Palettes</h2>
            <div className="grid gap-3 sm:grid-cols-2">
              {data.palettes.map((p) => (
                <div
                  key={p.name}
                  className="flex items-center gap-3 rounded-lg border border-ink-800 bg-ink-900 p-3"
                >
                  <div className="flex gap-1">
                    {[p.colors.primary, p.colors.accent, p.colors.background].map((c, i) => (
                      <span
                        key={i}
                        className="h-8 w-8 rounded border border-ink-700"
                        style={{ background: c }}
                        title={c}
                      />
                    ))}
                  </div>
                  <div>
                    <div className="text-ink-100">{p.name}</div>
                    <div className="font-mono text-[11px] text-ink-500">
                      {p.colors.primary} · {p.colors.accent}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </section>
        </div>
      )}
    </div>
  );
}
