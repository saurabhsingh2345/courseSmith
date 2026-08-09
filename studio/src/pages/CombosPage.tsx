import { useMemo, useState } from "react";
import {
  api,
  type CreateComboSegment,
  type ComboSegmentInfo,
  type ComboSummary,
  type SnippetTemplateInfo,
} from "../api/client";
import { ComboPools, ComboSkins, type Skin } from "../lib/families";
import { useFetch } from "../lib/useFetch";
import { ErrorNote } from "../components/ErrorNote";

// CombosPage: a longer video, directed from one line.
//
// The page is one control and one review, and that split is the whole design.
// You say what the video is about, how long it runs, what theme it is cut in and
// whether it carries captions — four decisions — and the director makes the rest:
// what the piece argues, how it divides into parts, which look carries each part
// and how long each one gets.
//
// What comes back is a PROPOSAL, not a video. Nothing is written until you press
// build, and everything in it is editable in place. That matters because the
// structure is the decision worth disagreeing with and it is cheapest to
// disagree with here — before a planning call has been spent on any part. A page
// that went straight from a prompt to a render would make "ask" and "commit" the
// same irreversible action, and the cost of being wrong is minutes of render.
//
// Editing an existing combo is the other half, and it is why this is a page
// rather than a CLI command. After watching a cut, what you want is to change
// one part — and every edit costs a rebuild, so the page batches: change several
// segments, then run once.

const COMBO_COURSE = "combos";

/** A segmented control. */
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
    <div className="flex flex-wrap rounded-lg border border-ink-800 p-0.5">
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

/** Small monospace chip. */
function Chip({ children, tone }: { children: React.ReactNode; tone?: "muted" | "warn" | "role" }) {
  const color =
    tone === "warn"
      ? "bg-amber-500/15 text-amber-300"
      : tone === "role"
        ? "bg-sky-500/15 text-sky-300"
        : "bg-ink-800 text-ink-400";
  return (
    <span className={`rounded px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wide ${color}`}>
      {children}
    </span>
  );
}

/**
 * The template picker, grouped the way the catalog is and narrowed to the theme.
 *
 * A native select rather than the snippets page's card gallery: here you are
 * choosing a template for the fourth segment of nine, so the choice has to sit
 * inline next to its prompt. A grid would push the thing you are editing off the
 * screen every time you opened it.
 *
 * Narrowed to the theme's pool because the server enforces the same rule. A
 * picker offering a look the server will reject teaches people that the picker
 * lies, and the error arrives after the click rather than before it.
 */
function TemplateSelect({
  templates,
  skin,
  value,
  onChange,
  id,
}: {
  templates: SnippetTemplateInfo[];
  skin: Skin;
  value: string;
  onChange: (v: string) => void;
  id?: string;
}) {
  const groups = useMemo(() => {
    const allowed = new Set(ComboPools[skin] ?? []);
    const byCat = new Map<string, { title: string; items: SnippetTemplateInfo[] }>();
    for (const t of templates) {
      if (!allowed.has(t.family ?? "")) continue;
      const key = t.category ?? "";
      if (!byCat.has(key)) byCat.set(key, { title: t.category_title || key, items: [] });
      byCat.get(key)!.items.push(t);
    }
    return [...byCat.values()];
  }, [templates, skin]);

  return (
    <select
      id={id}
      className="rounded-md border border-ink-800 bg-ink-950 px-2 py-1.5 text-[13px] text-ink-100 focus:border-brand focus:outline-none"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    >
      {/* A template from another theme is still shown when it is the current
          value, or swapping the theme on an existing piece would silently blank
          every picker rather than showing what is there. */}
      {!groups.some((g) => g.items.some((t) => t.name === value)) && value && (
        <option value={value}>{value} (not in this theme)</option>
      )}
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

/**
 * One proposed segment, as the review shows it.
 *
 * Everything the director decided is visible and editable: the look, what the
 * part establishes, the facts it will be built from, and the reason the look was
 * chosen. The reason is read-only — it is the machine's account of itself, and
 * editing it would only produce a note nobody wrote.
 */
function ProposedSegment({
  index,
  segment,
  templates,
  skin,
  onChange,
  onRemove,
  onMove,
  canMoveUp,
  canMoveDown,
}: {
  index: number;
  segment: CreateComboSegment;
  templates: SnippetTemplateInfo[];
  skin: Skin;
  onChange: (next: CreateComboSegment) => void;
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
        <div className="flex flex-wrap items-center gap-2">
          {segment.role && <Chip tone="role">{segment.role}</Chip>}
          {segment.heading && (
            <span className="truncate text-[13px] font-medium text-ink-200">{segment.heading}</span>
          )}
        </div>
        <TemplateSelect
          templates={templates}
          skin={skin}
          value={segment.template}
          onChange={(template) => onChange({ ...segment, template })}
        />
        <textarea
          className="min-h-[52px] w-full resize-y rounded-md border border-ink-800 bg-ink-950 p-2 text-[13px] text-ink-100 placeholder:text-ink-500 focus:border-brand focus:outline-none"
          placeholder="What should the viewer know after this part that they did not before?"
          value={segment.prompt}
          onChange={(e) => onChange({ ...segment, prompt: e.target.value })}
          aria-label={`What segment ${index + 1} establishes`}
        />
        {/* The material is shown because it is the field that decides whether
            this segment is TRUE. The writer plans from it, so a figure that is
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
        {segment.why && (
          <p className="text-[11.5px] leading-snug text-ink-500">Why this look: {segment.why}</p>
        )}
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
 * The editor for an existing combo.
 *
 * Edits are staged locally and applied together. Every change here moves the
 * narration — swapping a template re-plans the segment, skipping one takes its
 * words out of the read — so each one costs a full rebuild. Batching them turns
 * four edits into one run instead of four, and the banner says so plainly
 * rather than letting the cost be discovered.
 */
function ComboEditor({ id, onClose }: { id: string; onClose: () => void }) {
  const detail = useFetch(() => api.combo(id), [id]);
  const templates = useFetch(() => api.snippetTemplates(), []);
  const [pending, setPending] = useState<Record<string, Partial<ComboSegmentInfo>>>({});
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const segments = detail.data?.segment_list ?? [];

  // What each segment gave up, keyed by id, out of the plan the run wrote.
  //
  // The plan stage ships the closest draft when the correction rounds run out —
  // correct, since the clip still renders — but that used to leave no trace
  // anywhere a creator looks. A segment that came in under its word budget was
  // indistinguishable from one that passed, so the combos with three loose
  // segments read as finished.
  const compromises = useMemo(() => {
    const out = new Map<string, string[]>();
    const plan = detail.data?.plan;
    if (!plan) return out;
    try {
      // The plan arrives as raw JSON so the page does not have to hold a mirror
      // of every template's shape; only this one field is read from it.
      const parsed = JSON.parse(typeof plan === "string" ? plan : JSON.stringify(plan));
      for (const seg of parsed?.segments ?? []) {
        const list = seg?.plan?.compromises;
        if (Array.isArray(list) && list.length > 0) out.set(seg.id, list);
      }
    } catch {
      // A plan that does not parse is not worth failing the editor over — the
      // segments and their prompts are what this page is actually for.
    }
    return out;
  }, [detail.data?.plan]);
  const dirty = Object.keys(pending).length > 0;

  const stage = (segId: string, patch: Partial<ComboSegmentInfo>) =>
    setPending((p) => ({ ...p, [segId]: { ...p[segId], ...patch } }));

  const valueOf = (seg: ComboSegmentInfo) => ({ ...seg, ...pending[seg.id] });

  const applyAndRun = async () => {
    setBusy(true);
    setError(null);
    try {
      for (const [segId, patch] of Object.entries(pending)) {
        await api.patchComboSegment(id, segId, {
          template: patch.template,
          prompt: patch.prompt,
          material: patch.material,
          skip: patch.skip,
        });
      }
      setPending({});
      await api.runCombo(id);
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
              <div className="mb-2 flex flex-wrap items-center gap-2">
                <span className="font-mono text-[12px] text-ink-500">{i + 1}</span>
                <span className="font-mono text-[11px] text-ink-500">{seg.id}</span>
                {/* The role and the heading tell you WHETHER to rewrite this
                    segment, where the material below tells you what to correct.
                    A segment reads very differently once you can see it is the
                    piece's only hook. */}
                {seg.role && <Chip tone="role">{seg.role}</Chip>}
                {seg.heading && (
                  <span className="truncate text-[12px] text-ink-300">{seg.heading}</span>
                )}
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
                  // The editor works on a piece already cut in a theme, and the
                  // page does not know which. Offering everything is right here:
                  // narrowing would hide the template that is actually in use.
                  skin="editorial"
                  value={v.template}
                  onChange={(template) => stage(seg.id, { template })}
                />
              )}
              <textarea
                className="mt-2 min-h-[52px] w-full resize-y rounded-md border border-ink-800 bg-ink-900 p-2 text-[13px] text-ink-100 focus:border-brand focus:outline-none"
                value={v.prompt}
                onChange={(e) => stage(seg.id, { prompt: e.target.value })}
              />
              {/* The facts this segment is planned from, editable. This is the
                  field that decides whether the segment is TRUE — the writer
                  builds from it, so a wrong figure here becomes a wrong figure in
                  the finished video, and correcting it is one edit rather than a
                  re-direct. */}
              <textarea
                className="mt-2 min-h-[38px] w-full resize-y rounded-md border border-ink-800/60 bg-ink-900 p-2 font-mono text-[11.5px] leading-snug text-ink-300 placeholder:text-ink-600 focus:border-brand focus:outline-none"
                placeholder="Facts this part is built from — figures, names, thresholds"
                value={v.material ?? ""}
                onChange={(e) => stage(seg.id, { material: e.target.value })}
                aria-label={`Material for segment ${i + 1}`}
              />
              {seg.why && (
                <p className="mt-2 text-[11.5px] leading-snug text-ink-500">
                  Why this look: {seg.why}
                </p>
              )}
              {compromises.get(seg.id)?.map((line, j) => (
                <p key={j} className="mt-2 text-[11.5px] leading-snug text-amber-400/90">
                  Shipped looser than asked: {line}
                </p>
              ))}
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
            whole combo. Make all your changes first.
          </p>
        )}
      </div>
    </div>
  );
}

export function CombosPage() {
  const combos = useFetch(() => api.combos(), []);
  const templates = useFetch(() => api.snippetTemplates(), []);

  // The four decisions.
  const [subject, setSubject] = useState("");
  const [minutes, setMinutes] = useState(5);
  const [skin, setSkin] = useState<Skin>("default");
  const [captions, setCaptions] = useState<"off" | "on">("off");
  const [mode, setMode] = useState<"dark" | "light">("dark");

  // The proposal, once it exists.
  const [title, setTitle] = useState("");
  const [angle, setAngle] = useState("");
  const [pool, setPool] = useState("");
  const [runtime, setRuntime] = useState("");
  const [segments, setSegments] = useState<CreateComboSegment[]>([]);

  const [planOnly, setPlanOnly] = useState(true);
  const [busy, setBusy] = useState(false);
  const [directing, setDirecting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [open, setOpen] = useState<string | null>(null);

  const move = (i: number, dir: -1 | 1) =>
    setSegments((s) => {
      const next = [...s];
      const j = i + dir;
      if (j < 0 || j >= next.length) return s;
      [next[i], next[j]] = [next[j], next[i]];
      return next;
    });

  /**
   * Ask the director for a piece.
   *
   * Fills the review below rather than creating anything, which is the whole
   * point: what comes back is a proposal you read and change, not a commitment.
   * That is also why the endpoint writes nothing.
   */
  const direct = async () => {
    const text = subject.trim();
    if (!text) return;
    setDirecting(true);
    setError(null);
    try {
      const proposal = await api.directCombo({
        subject: text,
        title: title.trim() || undefined,
        minutes,
        skin,
        captions,
        mode,
      });
      setSegments(proposal.segments);
      setAngle(proposal.angle);
      setPool(proposal.pool);
      setRuntime(proposal.runtime ?? "");
      if (!title.trim()) setTitle(proposal.title);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setDirecting(false);
    }
  };

  const canBuild = segments.length >= 2 && segments.every((s) => s.prompt.trim() !== "");

  const build = async () => {
    setBusy(true);
    setError(null);
    try {
      const created = await api.createCombo({
        title: title.trim() || undefined,
        brief: subject.trim() || undefined,
        // The angle has to make the round trip: it is what the critic scores
        // every finished segment against, and without it that pass can only ask
        // whether a segment is good, which returns opinions about prose.
        angle: angle.trim() || undefined,
        segments,
        mode,
        captions,
        skin,
        plan_only: planOnly,
      });
      setSegments([]);
      setAngle("");
      setTitle("");
      setSubject("");
      setOpen(created.id);
      combos.reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="text-lg font-semibold text-ink-100">Direct a combo</h1>
      <p className="mt-1 text-ink-400">
        Say what the video is about and how long it should run. The director decides what the piece
        argues, how it divides into parts, which look carries each part and how long each one gets —
        then writes them, reads the whole thing back, and repairs anything that does not belong.
      </p>

      <section className="mt-5">
        <label className="mb-2 block text-[11px] uppercase tracking-wide text-ink-500" htmlFor="combo-subject">
          What is the video about?
        </label>
        <textarea
          id="combo-subject"
          className="min-h-[76px] w-full resize-y rounded-lg border border-ink-800 bg-ink-950 p-3 text-[14px] text-ink-100 placeholder:text-ink-500 focus:border-brand focus:outline-none"
          placeholder="What Is Artificial Intelligence?"
          value={subject}
          onChange={(e) => setSubject(e.target.value)}
        />

        <div className="mt-4 flex flex-wrap items-center gap-x-6 gap-y-3">
          <label className="flex items-center gap-2 text-[13px] text-ink-400">
            <span className="text-[11px] uppercase tracking-wide text-ink-500">Minutes</span>
            <input
              type="number"
              min={1}
              max={15}
              className="w-16 rounded-md border border-ink-800 bg-ink-950 px-2 py-1.5 text-[13px] text-ink-100 focus:border-brand focus:outline-none"
              value={minutes}
              onChange={(e) => setMinutes(Math.max(1, Number(e.target.value) || 1))}
            />
          </label>
          {/* The theme is not only a look: it decides which templates can be
              cast at all, because a piece is cut in one house style throughout
              and a template from another family reads as a different production
              however good it is. */}
          <Segmented value={skin} onChange={setSkin} options={ComboSkins} />
          <Segmented
            value={captions}
            onChange={setCaptions}
            options={[
              { value: "off", label: "No captions", hint: "The .vtt sidecar is still written" },
              { value: "on", label: "Captions", hint: "Burn the karaoke caption track into the video" },
            ]}
          />
          <Segmented
            value={mode}
            onChange={setMode}
            options={[
              { value: "dark", label: "Dark", hint: "The default editorial look" },
              { value: "light", label: "Light", hint: "A paper-white video, same branding" },
            ]}
          />
        </div>

        <div className="mt-4 flex items-center gap-3">
          <button
            type="button"
            className="rounded bg-sky-600 px-4 py-1.5 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
            onClick={direct}
            disabled={!subject.trim() || directing}
          >
            {directing ? "Directing…" : "Direct it"}
          </button>
          <span className="text-[12px] text-ink-500">
            Proposes the whole piece below. Nothing is created — change anything before you build.
          </span>
        </div>
      </section>

      {error && <div className="mt-4 rounded-lg border border-danger/40 bg-danger/10 p-3 text-danger">{error}</div>}

      {segments.length > 0 && (
        <>
          {/* The argument, first and editable.
              It is the decision everything else serves, and the one most worth
              disagreeing with: a piece can be about the right subject and be
              making the wrong point, and this is the only place that is visible
              before the render. */}
          <section className="mt-6 rounded-xl border border-ink-800 bg-ink-900 p-4">
            <label className="mb-2 block text-[11px] uppercase tracking-wide text-ink-500" htmlFor="combo-title">
              Title
            </label>
            <input
              id="combo-title"
              className="w-full rounded-lg border border-ink-800 bg-ink-950 p-2.5 text-[14px] text-ink-100 focus:border-brand focus:outline-none"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
            <label className="mb-2 mt-4 block text-[11px] uppercase tracking-wide text-ink-500" htmlFor="combo-angle">
              What it argues{" "}
              <span className="normal-case tracking-normal text-ink-500">
                — every segment is written toward this, and judged against it afterwards
              </span>
            </label>
            <textarea
              id="combo-angle"
              className="min-h-[52px] w-full resize-y rounded-lg border border-ink-800 bg-ink-950 p-2.5 text-[13px] text-ink-100 focus:border-brand focus:outline-none"
              value={angle}
              onChange={(e) => setAngle(e.target.value)}
            />
            <div className="mt-3 flex flex-col gap-1 text-[11.5px] text-ink-500">
              {pool && <span>{pool}</span>}
              {runtime && <span>{runtime}</span>}
            </div>
          </section>

          <section className="mt-4">
            <div className="mb-2 flex items-center justify-between">
              <h2 className="text-[11px] uppercase tracking-wide text-ink-500">The parts, in order</h2>
              <span className="text-[12px] text-ink-500">{segments.length} · at least 2</span>
            </div>
            {templates.error && <ErrorNote error={templates.error} onRetry={templates.reload} />}
            <div className="flex flex-col gap-2">
              {segments.map((seg, i) => (
                <ProposedSegment
                  key={i}
                  index={i}
                  segment={seg}
                  templates={templates.data ?? []}
                  skin={skin}
                  onChange={(next) => setSegments((s) => s.map((v, j) => (j === i ? next : v)))}
                  onRemove={() => setSegments((s) => s.filter((_, j) => j !== i))}
                  onMove={(dir) => move(i, dir)}
                  canMoveUp={i > 0}
                  canMoveDown={i < segments.length - 1}
                />
              ))}
            </div>
          </section>

          <section className="mt-5 flex flex-wrap items-center gap-3">
            <button
              type="button"
              className="rounded bg-sky-600 px-4 py-1.5 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
              onClick={build}
              disabled={!canBuild || busy}
            >
              {busy ? "Starting…" : planOnly ? "Write it" : "Make the video"}
            </button>
            {/* Plan-only defaults ON here and OFF for snippets, deliberately:
                writing a combo is one call per segment plus the critic, and
                rendering is minutes, so reading what it wrote before paying for
                the voice and the render is the sane default. */}
            <label className="flex items-center gap-2 text-[13px] text-ink-400">
              <input type="checkbox" checked={planOnly} onChange={(e) => setPlanOnly(e.target.checked)} />
              Write only — stop before the voice and the render
            </label>
          </section>
        </>
      )}

      {open && <ComboEditor id={open} onClose={() => setOpen(null)} />}

      <section className="mt-8">
        <h2 className="mb-2 text-[11px] uppercase tracking-wide text-ink-500">Your combos</h2>
        {combos.error && <ErrorNote error={combos.error} onRetry={combos.reload} />}
        {combos.data?.length === 0 && <p className="text-ink-500">No combos yet.</p>}
        <div className="flex flex-col gap-2">
          {(combos.data ?? []).map((r: ComboSummary) => (
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
                  await api.deleteCombo(r.id);
                  if (open === r.id) setOpen(null);
                  combos.reload();
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

export { COMBO_COURSE };
