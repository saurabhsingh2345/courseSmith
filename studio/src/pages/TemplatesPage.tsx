// The archetype catalog, and what picking one actually does to a course.
//
// It used to be three stacked lists — archetypes, animation styles, palettes —
// which told you the names of things and nothing about them. The two questions
// somebody browsing this page has are "what does this one look like?" and "how
// do I use it?", and a list of names answers neither.
//
// So: pick an archetype on the left, see it resolved on the right. The palette
// as swatches, the motion language as something that actually moves, and an
// Apply control that writes the choice to a real course through the same
// endpoint the course editor uses.
//
// **There is no save button for the archetype itself, and that is deliberate.**
// `/api/archetypes` is a GET of a Go registry, and the motion values below are
// owned by internal/pipeline/motion.go and mirrored here under a drift test —
// the renderer and this page have to agree with Go or a preview lies about the
// video. A slider here that appeared to edit them would be a control with
// nothing behind it. What is editable is which archetype a *course* uses, which
// is what Apply does.

import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowRight, Check, Play } from "lucide-react";
import { api, type Archetype, type PaletteColors } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { ErrorNote } from "../components/ErrorNote";
import { Badge } from "../components/base/Badge";
import { Button } from "../components/base/Button";
import { Card } from "../components/base/Card";
import { Skeleton } from "../components/base/Skeleton";
import { motion, ms } from "../theme/motionTokens";
import { cn } from "../lib/cn";
import { studio, useStudioStore } from "../store/studioStore";

/** Words-per-minute, shown as the length it implies for a spoken minute. */
const paceNote = (wpm: number) => `${wpm} wpm · ~${Math.round(wpm)} words a minute of narration`;

// ---------------------------------------------------------------------------
// The motion demo
// ---------------------------------------------------------------------------

/**
 * Five blocks entering on the real stagger, easing and duration.
 *
 * Replay is a `key` bump rather than a state machine: remounting restarts the
 * CSS animations from zero, which is both the shortest way to express "play it
 * again" and the only one that cannot drift out of step with itself.
 */
function MotionDemo({ archetype }: { archetype: Archetype }) {
  const [run, setRun] = useState(0);
  const items = ["idea", "example", "detail", "recap", "check"];

  // Re-run whenever the archetype changes, so clicking through the list shows
  // each one moving rather than sitting still after the first.
  useEffect(() => setRun((r) => r + 1), [archetype.name]);

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <h3 className="text-[11px] font-semibold uppercase tracking-widest text-ink-500">Motion</h3>
        <Button variant="ghost" size="sm" onClick={() => setRun((r) => r + 1)}>
          <Play className="size-3.5" aria-hidden /> Replay
        </Button>
      </div>

      <div
        key={run}
        className="flex flex-wrap gap-2 rounded-[var(--radius-md)] border border-ink-800 bg-ink-950 p-4"
      >
        {items.map((label, i) => (
          <span
            key={label}
            className="rounded-[var(--radius-sm)] border border-ink-700 bg-ink-800 px-2.5 py-1 text-[12px] text-ink-200"
            style={{
              // The tokens, used as themselves — this is the same stagger and
              // easing the rendered video uses for a list of points.
              animation: `cs-rise ${motion.timing.normal}s ${motion.easing.entrance} both`,
              animationDelay: `${i * motion.stagger.items}s`,
            }}
          >
            {label}
          </span>
        ))}
      </div>

      <dl className="mt-3 grid grid-cols-2 gap-x-4 gap-y-1 text-[12px] sm:grid-cols-3">
        {[
          ["entrance", `${ms(motion.timing.normal)}ms`],
          ["stagger", `${ms(motion.stagger.items)}ms`],
          ["easing", motion.easing.entrance],
          ["fast", `${ms(motion.timing.fast)}ms`],
          ["slow", `${ms(motion.timing.slow)}ms`],
          ["words", `${ms(motion.stagger.words)}ms`],
        ].map(([k, v]) => (
          <div key={k} className="flex items-baseline justify-between gap-2 border-b border-ink-850 py-1">
            <dt className="text-ink-500">{k}</dt>
            <dd className="truncate font-mono text-[11px] text-ink-300" title={v}>
              {v}
            </dd>
          </div>
        ))}
      </dl>
      <p className="mt-2 text-[11px] text-ink-500">
        Shared across archetypes and owned by the Go pipeline — the studio mirrors these values so a
        preview moves exactly like the render.
      </p>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Apply
// ---------------------------------------------------------------------------

/** Write the archetype onto a course, through the endpoint the editor uses. */
function ApplyToCourse({ archetype }: { archetype: Archetype }) {
  const courses = useStudioStore((s) => s.courses);
  const status = useStudioStore((s) => s.coursesStatus);
  const [slug, setSlug] = useState<string>("");
  const [state, setState] = useState<"idle" | "saving" | "saved">("idle");
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    void studio.loadCourses();
  }, []);

  // A different archetype is a different unsaved intent.
  useEffect(() => setState("idle"), [archetype.name]);

  const apply = useCallback(async () => {
    if (!slug) return;
    setState("saving");
    setError(null);
    try {
      await api.updateCourse(slug, { archetype: archetype.name });
      setState("saved");
    } catch (e) {
      setState("idle");
      setError(e instanceof Error ? e.message : String(e));
    }
  }, [slug, archetype.name]);

  if (status === "ready" && courses.length === 0) {
    return (
      <p className="text-[12px] text-ink-500">
        No courses yet — create one to apply an archetype to it.
      </p>
    );
  }

  return (
    <div>
      <h3 className="mb-2 text-[11px] font-semibold uppercase tracking-widest text-ink-500">
        Apply to a course
      </h3>
      {error && <ErrorNote error={error} />}
      <div className="flex flex-col gap-2 sm:flex-row">
        {/* A native select, like every other picker in the studio — it is the
            control a phone renders as its own wheel, and consistency with the
            ten already here beats reaching for a primitive nothing else uses. */}
        <select
          aria-label="Course"
          value={slug}
          onChange={(e) => setSlug(e.target.value)}
          className="min-h-11 w-full rounded-[var(--radius-md)] border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100 sm:flex-1"
        >
          <option value="">{status === "loading" ? "Loading courses…" : "Choose a course"}</option>
          {courses.map((c) => (
            <option key={c.slug} value={c.slug}>
              {c.name}
            </option>
          ))}
        </select>
        <Button onClick={apply} disabled={!slug || state === "saving"}>
          {state === "saved" ? <Check className="size-4" aria-hidden /> : null}
          {state === "saving" ? "Applying…" : state === "saved" ? "Applied" : "Apply"}
        </Button>
      </div>
      {state === "saved" && (
        <div className="mt-2 flex items-center gap-2 text-[12px] text-success">
          <span>
            {archetype.name} is now the archetype for {courses.find((c) => c.slug === slug)?.name}.
          </span>
          <button
            type="button"
            onClick={() => navigate(`/c/${slug}/edit`)}
            className="inline-flex items-center gap-1 text-brand hover:underline"
          >
            Open settings <ArrowRight className="size-3" aria-hidden />
          </button>
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------

function Swatches({ palette }: { palette: PaletteColors }) {
  const entries: [string, string][] = [
    ["primary", palette.colors.primary],
    ["accent", palette.colors.accent],
    ["background", palette.colors.background],
  ];
  return (
    <div className="flex items-center gap-3">
      <div className="flex">
        {entries.map(([role, hex], i) => (
          <span
            key={role}
            title={`${role} ${hex}`}
            style={{ background: hex }}
            className={cn(
              "size-9 border border-ink-700",
              i === 0 && "rounded-l-[var(--radius-sm)]",
              i === entries.length - 1 && "rounded-r-[var(--radius-sm)]",
              i > 0 && "-ml-px",
            )}
          />
        ))}
      </div>
      <div className="min-w-0">
        <div className="truncate text-[13px] text-ink-100">{palette.name}</div>
        <div className="truncate font-mono text-[11px] text-ink-500">
          {palette.colors.primary} · {palette.colors.accent}
        </div>
      </div>
    </div>
  );
}

export function TemplatesPage() {
  const { data, loading, error, reload } = useFetch(() => api.archetypes(), []);
  const [selected, setSelected] = useState<string | null>(null);

  const archetypes = useMemo(() => data?.archetypes ?? [], [data]);
  const active = archetypes.find((a) => a.name === selected) ?? archetypes[0] ?? null;

  return (
    <div className="mx-auto max-w-6xl p-4 sm:p-6">
      <header className="mb-5">
        <h1 className="text-lg font-semibold text-ink-100">Templates</h1>
        <p className="mt-1 max-w-2xl text-[13px] text-ink-500">
          An archetype is the shape of a course — how fast it talks, how it moves, and the prompt
          each lesson is written against. Pick one to see it resolved, then apply it to a course.
        </p>
      </header>

      {error && <ErrorNote error={error} onRetry={reload} />}

      {loading && !data && (
        <div className="grid gap-4 lg:grid-cols-[minmax(0,18rem)_1fr]">
          <Skeleton className="h-64" />
          <Skeleton className="h-64" />
        </div>
      )}

      {data && active && (
        <div className="grid gap-4 lg:grid-cols-[minmax(0,18rem)_1fr] lg:items-start">
          {/* The list. A radiogroup, not a set of buttons — exactly one is
              chosen and arrow keys should move between them. */}
          <div role="radiogroup" aria-label="Archetypes" className="flex flex-col gap-1.5">
            {archetypes.map((a) => {
              const isActive = a.name === active.name;
              return (
                <button
                  key={a.name}
                  role="radio"
                  aria-checked={isActive}
                  onClick={() => setSelected(a.name)}
                  className={cn(
                    "min-h-11 rounded-[var(--radius-md)] border px-3 py-2 text-left",
                    "transition-colors duration-[var(--motion-fast)]",
                    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2 focus-visible:ring-offset-bg",
                    isActive
                      ? "border-brand bg-ink-800/60"
                      : "border-ink-800 bg-ink-900 hover:border-ink-700",
                  )}
                >
                  <div className="flex items-center justify-between gap-2">
                    <span
                      className={cn("truncate text-[13px]", isActive ? "text-ink-100" : "text-ink-200")}
                    >
                      {a.name}
                    </span>
                    <span className="shrink-0 font-mono text-[10px] text-ink-500">
                      {a.pace_hint}
                    </span>
                  </div>
                  <div className="mt-0.5 line-clamp-2 text-[12px] text-ink-500">{a.description}</div>
                </button>
              );
            })}
          </div>

          {/* The detail. */}
          <Card className="border-ink-800 bg-ink-900 p-4 sm:p-5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div className="min-w-0">
                <h2 className="text-base font-semibold text-ink-100">{active.name}</h2>
                <p className="mt-1 text-[13px] text-ink-300">{active.description}</p>
              </div>
              <div className="flex shrink-0 flex-wrap gap-1.5">
                <Badge variant="outline">{active.default_animation}</Badge>
                <Badge variant="secondary" title={paceNote(active.pace_hint)}>
                  {active.pace_hint} wpm
                </Badge>
              </div>
            </div>

            <blockquote className="mt-4 border-l-2 border-ink-700 pl-3 text-[12px] italic text-ink-400">
              {active.prompt_hint}
            </blockquote>

            <hr className="my-5 border-ink-800" />
            <MotionDemo archetype={active} />

            <hr className="my-5 border-ink-800" />
            <div>
              <h3 className="mb-2 text-[11px] font-semibold uppercase tracking-widest text-ink-500">
                Palettes
              </h3>
              <div className="grid gap-2 sm:grid-cols-2">
                {data.palettes.map((p) => (
                  <div
                    key={p.name}
                    className="rounded-[var(--radius-md)] border border-ink-800 bg-ink-950 p-2.5"
                  >
                    <Swatches palette={p} />
                  </div>
                ))}
              </div>
              <p className="mt-2 text-[11px] text-ink-500">
                Palettes are chosen per course, independently of the archetype.
              </p>
            </div>

            <hr className="my-5 border-ink-800" />
            <div>
              <h3 className="mb-2 text-[11px] font-semibold uppercase tracking-widest text-ink-500">
                Animation styles
              </h3>
              <div className="flex flex-wrap gap-1.5">
                {data.animation_styles.map((s) => (
                  <Badge
                    key={s}
                    variant={s === active.default_animation ? "default" : "outline"}
                    title={s === active.default_animation ? `${active.name}'s default` : undefined}
                  >
                    {s}
                  </Badge>
                ))}
              </div>
            </div>

            <hr className="my-5 border-ink-800" />
            <ApplyToCourse archetype={active} />
          </Card>
        </div>
      )}
    </div>
  );
}
