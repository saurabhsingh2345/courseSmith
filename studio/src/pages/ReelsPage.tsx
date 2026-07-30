import { useMemo, useState } from "react";
import {
  api,
  type CreateReelSegment,
  type ReelSegmentInfo,
  type ReelSummary,
  type SnippetTemplateInfo,
} from "../api/client";
import { useFetch } from "../lib/useFetch";
import { ErrorNote } from "../components/ErrorNote";

// ReelsPage: a longer video built from several templates on one timeline.
//
// The page is two halves and the split matters. Building a reel is choosing an
// ORDER — which look carries which part of the argument — so the builder is a
// list you add to and reorder, not a grid you pick one thing from. That is the
// difference from the snippets page, where the gallery is the primary control
// because a snippet is one decision.
//
// Editing an existing reel is the other half, and it is the reason the page
// exists rather than a CLI being enough. After watching a cut, what you want is
// to change one part — and every edit here costs the same as a rebuild, so the
// page batches: change several segments, then run once. A page that re-ran on
// every keystroke would spend an hour of render time on a minute of editing.

const REEL_COURSE = "reels";

/**
 * A segmented control, matching the snippets page.
 *
 * The runtime picker was the only one of these, written inline. Three of them
 * inline is three places to get the dark-shell colours wrong — and the base
 * Button's "secondary" variant resolves to the light-mode surface token, so it
 * renders as a white pill here and cannot be used.
 */
function Segmented<T extends string | number>({
  value,
  options,
  onChange,
}: {
  value: T;
  options: { value: T; label: string; hint?: string }[];
  onChange: (v: T) => void;
}) {
  return (
    <div className="flex rounded-lg border border-ink-800 p-0.5">
      {options.map((o) => (
        <button
          key={String(o.value)}
          type="button"
          onClick={() => onChange(o.value)}
          title={o.hint}
          className={[
            "rounded-md px-3 py-1.5 text-[13px] transition-colors",
            value === o.value ? "bg-ink-800 text-ink-100" : "text-ink-400 hover:text-ink-200",
          ].join(" ")}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}

/** Small monospace chip, matching the snippets page. */
function Chip({ children, tone }: { children: React.ReactNode; tone?: "muted" | "warn" }) {
  const color =
    tone === "warn" ? "bg-amber-500/15 text-amber-300" : "bg-ink-800 text-ink-400";
  return (
    <span className={`rounded px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wide ${color}`}>
      {children}
    </span>
  );
}

/**
 * The template picker, grouped the way the catalog is.
 *
 * A native select rather than the snippets page's card gallery: here you are
 * choosing a template for the *fourth segment of nine*, so the choice has to sit
 * inline next to its prompt. A grid would push the thing you are editing off the
 * screen every time you opened it.
 */
function TemplateSelect({
  templates,
  value,
  onChange,
  id,
}: {
  templates: SnippetTemplateInfo[];
  value: string;
  onChange: (v: string) => void;
  id?: string;
}) {
  // Grouped by first-seen category, which is the order Go sent them in — the
  // same contract the snippets gallery relies on.
  const groups = useMemo(() => {
    const byCat = new Map<string, { title: string; items: SnippetTemplateInfo[] }>();
    for (const t of templates) {
      const key = t.category ?? "";
      if (!byCat.has(key)) byCat.set(key, { title: t.category_title || key, items: [] });
      byCat.get(key)!.items.push(t);
    }
    return [...byCat.values()];
  }, [templates]);

  return (
    <select
      id={id}
      className="rounded-md border border-ink-800 bg-ink-950 px-2 py-1.5 text-[13px] text-ink-100 focus:border-brand focus:outline-none"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    >
      {groups.map((g) => (
        <optgroup key={g.title} label={g.title}>
          {g.items.map((t) => (
            <option key={t.name} value={t.name}>
              {t.title} ({t.name}){t.since ? ` · ${t.since}` : ""}
            </option>
          ))}
        </optgroup>
      ))}
    </select>
  );
}

/** One row of the builder: a template and what it should cover. */
function SegmentRow({
  index,
  segment,
  templates,
  onChange,
  onRemove,
  onMove,
  canMoveUp,
  canMoveDown,
}: {
  index: number;
  segment: CreateReelSegment;
  templates: SnippetTemplateInfo[];
  onChange: (next: CreateReelSegment) => void;
  onRemove: () => void;
  onMove: (dir: -1 | 1) => void;
  canMoveUp: boolean;
  canMoveDown: boolean;
}) {
  return (
    <div className="flex items-start gap-3 rounded-lg border border-ink-800 bg-ink-900 p-3">
      <span className="mt-2 w-5 shrink-0 text-center font-mono text-[12px] text-ink-500">
        {index + 1}
      </span>
      <div className="flex min-w-0 flex-1 flex-col gap-2">
        <TemplateSelect
          templates={templates}
          value={segment.template}
          onChange={(template) => onChange({ ...segment, template })}
        />
        <textarea
          className="min-h-[52px] w-full resize-y rounded-md border border-ink-800 bg-ink-950 p-2 text-[13px] text-ink-100 placeholder:text-ink-500 focus:border-brand focus:outline-none"
          placeholder="What should this part cover?"
          value={segment.prompt}
          onChange={(e) => onChange({ ...segment, prompt: e.target.value })}
        />
        {/* The material is shown because it is the field that decides whether
            this segment is true. The writer plans from it, so a figure that is
            wrong here is wrong in the finished video — and left invisible, a
            fact nobody can check is a fact nobody will. Smaller and quieter
            than the prompt: read it, correct it, do not have to write it. */}
        <textarea
          className="min-h-[38px] w-full resize-y rounded-md border border-ink-800/60 bg-ink-950 p-2 font-mono text-[11.5px] leading-snug text-ink-300 placeholder:text-ink-600 focus:border-brand focus:outline-none"
          placeholder="Facts this part is built from — figures, names, thresholds"
          value={segment.material ?? ""}
          onChange={(e) => onChange({ ...segment, material: e.target.value })}
          aria-label={`Material for segment ${index + 1}`}
        />
      </div>
      {/* Order is the argument, so reordering is a first-class control rather
          than something you do by deleting and re-adding. */}
      <div className="flex shrink-0 flex-col gap-1">
        <button
          type="button"
          className="rounded px-1.5 text-ink-500 hover:bg-ink-800 hover:text-ink-200 disabled:opacity-30"
          onClick={() => onMove(-1)}
          disabled={!canMoveUp}
          aria-label={`Move segment ${index + 1} earlier`}
        >
          ↑
        </button>
        <button
          type="button"
          className="rounded px-1.5 text-ink-500 hover:bg-ink-800 hover:text-ink-200 disabled:opacity-30"
          onClick={() => onMove(1)}
          disabled={!canMoveDown}
          aria-label={`Move segment ${index + 1} later`}
        >
          ↓
        </button>
        <button
          type="button"
          className="rounded px-1.5 text-ink-500 hover:bg-ink-800 hover:text-ink-200"
          onClick={onRemove}
          aria-label={`Remove segment ${index + 1}`}
        >
          ✕
        </button>
      </div>
    </div>
  );
}

/**
 * The editor for an existing reel.
 *
 * Edits are staged locally and applied together. Every change here moves the
 * narration — swapping a template re-plans the segment, skipping one takes its
 * words out of the read — so each one costs a full rebuild. Batching them turns
 * four edits into one run instead of four, and the banner says so plainly
 * rather than letting the cost be discovered.
 */
function ReelEditor({ id, onClose }: { id: string; onClose: () => void }) {
  const detail = useFetch(() => api.reel(id), [id]);
  const templates = useFetch(() => api.snippetTemplates(), []);
  const [pending, setPending] = useState<Record<string, Partial<ReelSegmentInfo>>>({});
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const segments = detail.data?.segment_list ?? [];
  const dirty = Object.keys(pending).length > 0;

  const stage = (segId: string, patch: Partial<ReelSegmentInfo>) =>
    setPending((p) => ({ ...p, [segId]: { ...p[segId], ...patch } }));

  const valueOf = (seg: ReelSegmentInfo) => ({ ...seg, ...pending[seg.id] });

  const applyAndRun = async () => {
    setBusy(true);
    setError(null);
    try {
      for (const [segId, patch] of Object.entries(pending)) {
        await api.patchReelSegment(id, segId, {
          template: patch.template,
          prompt: patch.prompt,
          skip: patch.skip,
        });
      }
      setPending({});
      await api.runReel(id);
      detail.reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mt-6 rounded-xl border border-ink-800 bg-ink-900 p-5">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <h2 className="truncate text-[15px] font-semibold text-ink-100">
            {detail.data?.title || id}
          </h2>
          <p className="mt-0.5 text-[12px] text-ink-500">
            {segments.length} segments · {segments.filter((s) => s.skip).length} skipped
          </p>
        </div>
        <button type="button" className="rounded px-2 py-1 text-ink-500 hover:bg-ink-800" onClick={onClose}>
          Close
        </button>
      </div>

      {detail.error && <ErrorNote error={detail.error} onRetry={detail.reload} />}
      {error && <div className="mt-3 rounded-lg border border-danger/40 bg-danger/10 p-3 text-danger">{error}</div>}

      {detail.data?.ready && detail.data.video_url && (
        <video className="mt-4 w-full rounded-lg border border-ink-800" controls src={detail.data.video_url} />
      )}

      <div className="mt-4 flex flex-col gap-2">
        {segments.map((seg, i) => {
          const v = valueOf(seg);
          const changed = Boolean(pending[seg.id]);
          return (
            <div
              key={seg.id}
              className={[
                "rounded-lg border p-3",
                v.skip ? "border-ink-800 bg-ink-950 opacity-60" : "border-ink-800 bg-ink-950",
                changed ? "ring-1 ring-brand/50" : "",
              ].join(" ")}
            >
              <div className="mb-2 flex items-center gap-2">
                <span className="font-mono text-[12px] text-ink-500">{i + 1}</span>
                <span className="font-mono text-[11px] text-ink-500">{seg.id}</span>
                {v.skip && <Chip tone="warn">not in the cut</Chip>}
                {changed && <Chip tone="warn">edited</Chip>}
                <div className="ml-auto flex gap-2">
                  <button
                    type="button"
                    className="rounded px-2 py-0.5 text-[12px] text-ink-400 hover:bg-ink-800 hover:text-ink-100"
                    onClick={() => stage(seg.id, { skip: !v.skip })}
                  >
                    {v.skip ? "Put back" : "Drop"}
                  </button>
                </div>
              </div>
              {templates.data && (
                <TemplateSelect
                  templates={templates.data}
                  value={v.template}
                  onChange={(template) => stage(seg.id, { template })}
                />
              )}
              <textarea
                className="mt-2 min-h-[52px] w-full resize-y rounded-md border border-ink-800 bg-ink-900 p-2 text-[13px] text-ink-100 focus:border-brand focus:outline-none"
                value={v.prompt}
                onChange={(e) => stage(seg.id, { prompt: e.target.value })}
              />
            </div>
          );
        })}
      </div>

      {/* The cost, stated before it is paid. */}
      <div className="mt-4 flex items-center gap-3">
        <button
          type="button"
          className="rounded bg-sky-600 px-4 py-1.5 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
          onClick={applyAndRun}
          disabled={!dirty || busy}
        >
          {busy ? "Applying…" : `Apply ${Object.keys(pending).length || ""} and re-run`}
        </button>
        {dirty && (
          <p className="text-[12px] text-ink-500">
            Every change here moves the narration, so this re-plans, re-voices and re-renders the
            whole reel. Make all your changes first.
          </p>
        )}
      </div>
    </div>
  );
}

export function ReelsPage() {
  const reels = useFetch(() => api.reels(), []);
  const templates = useFetch(() => api.snippetTemplates(), []);

  const [title, setTitle] = useState("");
  const [brief, setBrief] = useState("");
  const [segments, setSegments] = useState<CreateReelSegment[]>([]);
  const [planOnly, setPlanOnly] = useState(true);
  const [mode, setMode] = useState<"dark" | "light">("dark");
  const [captions, setCaptions] = useState<"off" | "on">("off");
  const [skin, setSkin] = useState<"default" | "broadcast" | "minimal">("default");
  const [busy, setBusy] = useState(false);
  const [casting, setCasting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [open, setOpen] = useState<string | null>(null);

  const firstTemplate = templates.data?.[0]?.name ?? "metric";

  const addSegment = () =>
    setSegments((s) => [...s, { template: firstTemplate, prompt: "" }]);

  const move = (i: number, dir: -1 | 1) =>
    setSegments((s) => {
      const next = [...s];
      const j = i + dir;
      if (j < 0 || j >= next.length) return s;
      [next[i], next[j]] = [next[j], next[i]];
      return next;
    });

  /**
   * Ask the model for a structure.
   *
   * The proposal fills the builder rather than creating a reel, which is the
   * whole point: casting is a suggestion you read and change, not a commitment.
   * Everything it returns is editable in place before anything is spent, and
   * that is also why the endpoint writes nothing.
   */
  const cast = async () => {
    const text = brief.trim();
    if (!text) return;
    setCasting(true);
    setError(null);
    try {
      const proposal = await api.castReel({
        brief: text,
        title: title.trim() || undefined,
        segments: 5,
      });
      setSegments(proposal.segments);
      if (!title.trim()) setTitle(proposal.title);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setCasting(false);
    }
  };

  const canCreate = segments.length >= 2 && segments.every((s) => s.prompt.trim() !== "");

  const create = async () => {
    setBusy(true);
    setError(null);
    try {
      const created = await api.createReel({
        title: title.trim() || undefined,
        brief: brief.trim() || undefined,
        segments,
        mode,
        captions,
        skin,
        plan_only: planOnly,
      });
      setSegments([]);
      setTitle("");
      setBrief("");
      setOpen(created.id);
      reels.reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="text-lg font-semibold text-ink-100">Make a reel</h1>
      <p className="mt-1 text-ink-400">
        One video cut from several templates, narrated as one continuous read. Use a snippet for one
        idea in one look; use a reel when the piece is long enough that one look would not hold it.
      </p>

      <section className="mt-5">
        <label className="mb-2 block text-[11px] uppercase tracking-wide text-ink-500" htmlFor="reel-title">
          Title
        </label>
        <input
          id="reel-title"
          className="w-full rounded-lg border border-ink-800 bg-ink-950 p-3 text-[14px] text-ink-100 placeholder:text-ink-500 focus:border-brand focus:outline-none"
          placeholder="What decides whether a model runs on your machine"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />

        <label className="mb-2 mt-4 block text-[11px] uppercase tracking-wide text-ink-500" htmlFor="reel-brief">
          What is the whole piece about?{" "}
          <span className="normal-case tracking-normal text-ink-500">
            — optional today; it is what the caster will read
          </span>
        </label>
        <textarea
          id="reel-brief"
          className="min-h-[64px] w-full resize-y rounded-lg border border-ink-800 bg-ink-950 p-3 text-[14px] text-ink-100 placeholder:text-ink-500 focus:border-brand focus:outline-none"
          value={brief}
          onChange={(e) => setBrief(e.target.value)}
        />
        <div className="mt-2 flex items-center gap-3">
          <button
            type="button"
            className="rounded border border-ink-700 px-3 py-1.5 text-ink-200 hover:bg-ink-800 disabled:opacity-40"
            onClick={cast}
            disabled={!brief.trim() || casting}
          >
            {casting ? "Casting…" : "Cast it for me"}
          </button>
          <span className="text-[12px] text-ink-500">
            Proposes the segments below. Nothing is created — change anything before you build.
          </span>
        </div>
      </section>

      <section className="mt-6">
        <div className="mb-2 flex items-center justify-between">
          <h2 className="text-[11px] uppercase tracking-wide text-ink-500">
            The segments, in order
          </h2>
          <span className="text-[12px] text-ink-500">{segments.length} · at least 2</span>
        </div>
        {templates.error && <ErrorNote error={templates.error} onRetry={templates.reload} />}
        <div className="flex flex-col gap-2">
          {segments.map((seg, i) => (
            <SegmentRow
              key={i}
              index={i}
              segment={seg}
              templates={templates.data ?? []}
              onChange={(next) => setSegments((s) => s.map((v, j) => (j === i ? next : v)))}
              onRemove={() => setSegments((s) => s.filter((_, j) => j !== i))}
              onMove={(dir) => move(i, dir)}
              canMoveUp={i > 0}
              canMoveDown={i < segments.length - 1}
            />
          ))}
        </div>
        <button
          type="button"
          className="mt-2 rounded border border-ink-700 px-3 py-1.5 text-ink-300 hover:bg-ink-800"
          onClick={addSegment}
          disabled={!templates.data}
        >
          + Add a segment
        </button>
      </section>

      {error && <div className="mt-4 rounded-lg border border-danger/40 bg-danger/10 p-3 text-danger">{error}</div>}

      {/* The look. Identical to the snippets page, and applied to the whole
          reel: a piece that changed polarity or caption style between segments
          would read as several videos stitched together, which is the one thing
          a reel is built not to be. */}
      <section className="mt-6">
        <h2 className="mb-2 text-[11px] uppercase tracking-wide text-ink-500">How should it look?</h2>
        <div className="flex flex-wrap items-center gap-3">
          <Segmented
            value={mode}
            onChange={setMode}
            options={[
              { value: "dark", label: "Dark", hint: "The default editorial look" },
              { value: "light", label: "Light", hint: "A paper-white video, same branding" },
            ]}
          />
          <Segmented
            value={captions}
            onChange={setCaptions}
            options={[
              { value: "off", label: "No captions", hint: "The .vtt sidecar is still written" },
              { value: "on", label: "Captions", hint: "Burn the karaoke caption track into the video" },
            ]}
          />
          <Segmented
            value={skin}
            onChange={setSkin}
            options={[
              { value: "default", label: "Default", hint: "The catalog's own look" },
              { value: "broadcast", label: "Broadcast", hint: "Near-black stage, large type, standing chrome" },
              { value: "minimal", label: "Minimal", hint: "Flat, one accent, no furniture" },
            ]}
          />
        </div>
      </section>

      <section className="mt-6 flex items-center gap-3">
        <button
          type="button"
          className="rounded bg-sky-600 px-4 py-1.5 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
          onClick={create}
          disabled={!canCreate || busy}
        >
          {busy ? "Starting…" : planOnly ? "Plan it" : "Make the reel"}
        </button>
        {/* Plan-only defaults ON here and OFF for snippets, deliberately:
            planning a reel is one call per segment and rendering is minutes, so
            seeing the plan before paying for it is the sane default. */}
        <label className="flex items-center gap-2 text-[13px] text-ink-400">
          <input type="checkbox" checked={planOnly} onChange={(e) => setPlanOnly(e.target.checked)} />
          Plan only — stop before the voice and the render
        </label>
      </section>

      {open && <ReelEditor id={open} onClose={() => setOpen(null)} />}

      <section className="mt-8">
        <h2 className="mb-2 text-[11px] uppercase tracking-wide text-ink-500">Your reels</h2>
        {reels.error && <ErrorNote error={reels.error} onRetry={reels.reload} />}
        {reels.data?.length === 0 && <p className="text-ink-500">No reels yet.</p>}
        <div className="flex flex-col gap-2">
          {(reels.data ?? []).map((r: ReelSummary) => (
            <div key={r.id} className="flex items-center gap-3 rounded-lg border border-ink-800 bg-ink-900 p-3">
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <span className="truncate font-medium text-ink-100">{r.title || r.id}</span>
                  <Chip>{r.segments} segments</Chip>
                  {r.skipped > 0 && <Chip tone="warn">{r.skipped} skipped</Chip>}
                  {r.ready ? (
                    <span className="text-[11px] text-success">ready</span>
                  ) : (
                    <span className="text-[11px] text-ink-500">not rendered</span>
                  )}
                </div>
                <div className="truncate text-[12px] text-ink-500">{r.brief}</div>
              </div>
              {r.ready && r.video_url && (
                <a className="text-[13px] text-sky-400 hover:underline" href={r.video_url} download>
                  Download
                </a>
              )}
              <button
                type="button"
                className="rounded border border-ink-700 px-3 py-1.5 text-ink-300 hover:bg-ink-800"
                onClick={() => setOpen(r.id)}
              >
                Edit
              </button>
              <button
                type="button"
                className="rounded px-2 py-1 text-ink-500 hover:bg-ink-800 hover:text-ink-200"
                aria-label={`Delete ${r.title || r.id}`}
                onClick={async () => {
                  await api.deleteReel(r.id);
                  if (open === r.id) setOpen(null);
                  reels.reload();
                }}
              >
                ✕
              </button>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}

export { REEL_COURSE };
