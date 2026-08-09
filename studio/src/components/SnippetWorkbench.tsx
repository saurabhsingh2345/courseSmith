import { useMemo, useState } from "react";
import {
  SNIPPETS_COURSE,
  api,
  type SnippetPlan,
  type SnippetSummary,
  type SnippetTemplateInfo,
} from "../api/client";
import { useFetch } from "../lib/useFetch";
import { useRun } from "../state/RunContext";
import { useScreenShortcuts } from "../state/ShortcutContext";
import { ErrorNote } from "./ErrorNote";
import type { Skin } from "../lib/families";

// SnippetWorkbench is the "write a prompt, pick a look, get a clip" screen.
//
// It exists as a component rather than as a page because there are now two
// front doors onto the same machinery: the snippets gallery and the replica
// gallery. They differ in exactly three things — which family of templates they
// offer, the copy at the top, and whether the clip is cut in a fixed skin — and
// in nothing else. Everything below the gallery is the same prompt box, the same
// runtime picker, the same run log, the same list of what you have made.
//
// Two copies of that would be two places to fix the next bug in the runtime
// picker, and the runtime picker is the part of this screen with the most rules
// in it (see `runtimes` below). So the page components are thin and this is
// where the behaviour lives.

// Button styling follows the rest of the studio's pages rather than the base
// Button component: its `secondary` variant resolves to the light-mode surface
// token and reads as a white pill against this shell.
const PRIMARY_BTN =
  "rounded bg-sky-600 px-4 py-1.5 font-medium text-white hover:bg-sky-500 disabled:opacity-40";
const SECONDARY_BTN =
  "rounded border border-ink-700 px-3 py-1.5 text-ink-300 hover:bg-ink-800 disabled:opacity-40";
const QUIET_BTN = "rounded px-2 py-1 text-ink-500 hover:bg-ink-800 hover:text-ink-200";

// Runtimes are targets the plan aims at, not guarantees: the model writes to
// what the topic needs and can run over, so the labels say "about".
/** The shared default when a template does not ask for its own. */
const DEFAULT_TARGET_SEC = 45;

const TARGETS = [
  { sec: 10, label: "~10s", hint: "A hook, one line" },
  { sec: 20, label: "~20s", hint: "One idea" },
  { sec: 30, label: "~30s", hint: "Idea, with a turn" },
  { sec: 45, label: "~45s", hint: "Idea + example" },
  { sec: 60, label: "~60s", hint: "Room to explain" },
  { sec: 90, label: "~90s", hint: "A short with an arc" },
  { sec: 120, label: "~2 min", hint: "A directed short" },
  { sec: 180, label: "~3 min", hint: "The long form" },
];

/**
 * A segmented control.
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

/**
 * Words in a script, counted the way the pipeline counts them.
 *
 * Shown live under the script box because the length of a written script IS the
 * length of the clip — there is no runtime to pick any more — and somebody
 * pasting three paragraphs deserves to find out it is a two-minute video before
 * they render it rather than after.
 */
const wordCount = (s: string) => s.split(/\s+/).filter(Boolean).length;

/** Small monospace chip used for template metadata. */
function Chip({ children }: { children: React.ReactNode }) {
  return (
    <span className="rounded bg-ink-800 px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wide text-ink-400">
      {children}
    </span>
  );
}

function TemplateCard({
  template,
  selected,
  onSelect,
}: {
  template: SnippetTemplateInfo;
  selected: boolean;
  onSelect: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onSelect}
      aria-pressed={selected}
      className={[
        "group relative overflow-hidden rounded-xl border p-4 text-left transition-colors",
        selected
          ? "border-brand bg-brand/5"
          : "border-ink-800 bg-ink-900 hover:border-ink-700 hover:bg-ink-850",
      ].join(" ")}
    >
      {/* A quiet accent wash so the selected card reads as chosen at a glance
          rather than needing the border alone to carry it. */}
      {selected && (
        <span
          aria-hidden
          className="pointer-events-none absolute -right-10 -top-10 size-32 rounded-full bg-brand/20 blur-2xl"
        />
      )}
      {/* A real frame from this template, not a mock-up. It is downscaled from
          the visual-regression baseline, so the picture on the card is by
          construction what the template actually renders. */}
      <img
        src={`/template-previews/${template.name}.png`}
        alt=""
        aria-hidden
        loading="lazy"
        width={480}
        height={270}
        className="relative mb-3 aspect-video w-full rounded-lg border border-ink-800 bg-ink-950 object-cover"
        // A template whose preview has not been regenerated should lose the
        // image, not show a broken-image glyph on the one screen whose job is
        // helping someone choose.
        onError={(e) => {
          e.currentTarget.style.display = "none";
        }}
      />
      <div className="relative flex items-start justify-between gap-3">
        <span className="font-medium text-ink-100">{template.title}</span>
        {template.shows_code && <Chip>runs code</Chip>}
        {/* The release the template arrived in. A fact rather than a "new"
            badge, so it stays true once the next batch lands. */}
        {template.since && <Chip>{template.since}</Chip>}
      </div>
      {/* The name, which is the id every other surface addresses this template
          by — the CLI's --template, the key in snippet.yaml and combo.yaml. The
          card used to show only the title, so somebody who had read `toggle` in
          a combo.yaml had no way to find which of twelve pictures it was, and
          somebody who liked a picture had no way to name it. The search box
          already matched on name; nothing rendered it. */}
      <code className="relative mt-1 block font-mono text-[12px] text-ink-500">
        {template.name}
      </code>
      <p className="relative mt-1.5 text-[13px] leading-snug text-ink-400">
        {template.description}
      </p>
      <p className="relative mt-3 border-l-2 border-ink-700 pl-2 text-[12px] italic text-ink-500">
        “{template.example}”
      </p>
    </button>
  );
}

function SnippetRow({
  snippet,
  onOpen,
  onDelete,
}: {
  snippet: SnippetSummary;
  onOpen: () => void;
  onDelete: () => void;
}) {
  return (
    <div className="flex items-center gap-3 rounded-lg border border-ink-800 bg-ink-900 p-3">
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="truncate font-medium text-ink-100">{snippet.title}</span>
          <Chip>{snippet.template}</Chip>
          {snippet.ready ? (
            <span className="text-[11px] text-success">ready</span>
          ) : (
            <span className="text-[11px] text-ink-500">not rendered</span>
          )}
        </div>
        <div className="truncate text-[12px] text-ink-500">{snippet.prompt}</div>
      </div>
      {snippet.ready && snippet.video_url && (
        <a
          className="text-[13px] text-sky-400 hover:underline"
          href={snippet.video_url}
          download
        >
          Download
        </a>
      )}
      <button type="button" className={SECONDARY_BTN} onClick={onOpen}>
        Open
      </button>
      <button
        type="button"
        className={QUIET_BTN}
        onClick={onDelete}
        aria-label={`Delete ${snippet.title}`}
      >
        ✕
      </button>
    </div>
  );
}

/** The finished clip plus the plan that produced it. */
function SnippetViewer({ id, onClose }: { id: string; onClose: () => void }) {
  const detail = useFetch(() => api.snippet(id), [id]);
  const plan = detail.data?.plan as SnippetPlan | undefined;

  return (
    <div className="mt-6 rounded-xl border border-ink-800 bg-ink-900 p-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="font-semibold text-ink-100">{detail.data?.title ?? id}</h2>
          <p className="mt-0.5 text-[13px] text-ink-500">{detail.data?.prompt}</p>
        </div>
        <button type="button" className={QUIET_BTN} onClick={onClose}>
          Close
        </button>
      </div>

      {detail.error && <ErrorNote error={detail.error} onRetry={detail.reload} />}

      {detail.data?.video_url ? (
        <video
          className="mt-4 w-full rounded-lg border border-ink-800 bg-black"
          src={detail.data.video_url}
          controls
          preload="metadata"
        />
      ) : (
        <p className="mt-4 text-[13px] text-ink-500">
          No video yet — the render stage has not finished. Watch progress in the log panel
          (press <kbd>L</kbd>).
        </p>
      )}

      {plan && (
        <div className="mt-5">
          <h3 className="mb-2 text-[11px] uppercase tracking-wide text-ink-500">
            Beats ({plan.beats.length})
          </h3>
          <ol className="space-y-2">
            {plan.beats.map((beat, i) => (
              <li key={beat.id} className="rounded-lg border border-ink-800 bg-ink-950 p-3">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-[11px] text-ink-500">{i + 1}</span>
                  <span className="text-[13px] font-medium text-ink-200">{beat.heading}</span>
                  {beat.run && <Chip>runs</Chip>}
                </div>
                <p className="mt-1 text-[13px] leading-relaxed text-ink-400">{beat.narration}</p>
                {beat.code && (
                  <pre className="mt-2 overflow-x-auto rounded bg-ink-900 p-2 font-mono text-[12px] text-ink-300">
                    {beat.code}
                  </pre>
                )}
              </li>
            ))}
          </ol>
        </div>
      )}

      <a
        className="mt-4 inline-block text-[13px] text-sky-400 hover:underline"
        href={`/c/${SNIPPETS_COURSE}/l/${encodeURIComponent(id)}`}
      >
        Open the full artifact view →
      </a>
    </div>
  );
}

export interface SnippetWorkbenchProps {
  /**
   * Which family of templates this gallery offers. "" is the snippets catalog;
   * "replica" is the batch cut to the reference videos' grammar. The filter is
   * on the client because the API ships the whole catalog in one
   * category-ordered response — see handleSnippetTemplates for why that
   * ordering is not something a second endpoint should have to reproduce.
   */
  family: string;
  /** The h1. */
  heading: string;
  /** The line under it. */
  blurb: string;
  /**
   * The house style clips from this gallery are cut in. Undefined leaves it to
   * the course config, which is what the snippets page has always done.
   */
  skin?: Skin;
  /** Copy for the gallery's section heading. */
  galleryHeading?: string;
  /** Rendered between the blurb and the gallery. */
  children?: React.ReactNode;
}

export function SnippetWorkbench({
  family,
  heading,
  blurb,
  skin,
  galleryHeading = "How should it look?",
  children,
}: SnippetWorkbenchProps) {
  const templates = useFetch(() => api.snippetTemplates(), []);
  const [filter, setFilter] = useState("");
  const snippets = useFetch(() => api.snippets(), []);
  const { run } = useRun();

  const [prompt, setPrompt] = useState("");
  const [title, setTitle] = useState("");
  // The creator's own script, for the templates that speak one word for word.
  const [narration, setNarration] = useState("");
  const [template, setTemplate] = useState<string | null>(null);
  const [targetSec, setTargetSec] = useState<number | null>(null);
  const [captions, setCaptions] = useState<"off" | "on">("off");
  const [mode, setMode] = useState<"dark" | "light">("dark");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [open, setOpen] = useState<string | null>(null);

  useScreenShortcuts([{ keys: "r", label: "reload" }], (e) => {
    if (e.key === "r") {
      snippets.reload();
    }
  });

  /** Only this gallery's family. A template with no family is a snippet. */
  const offered = useMemo(
    () => (templates.data ?? []).filter((t) => (t.family ?? "") === family),
    [templates.data, family],
  );

  /**
   * The catalog grouped for display, filtered by the search box.
   *
   * The group ORDER comes from the order the API sent the templates in, not
   * from anything decided here: Go owns the category vocabulary and its
   * sequence, and a second copy in the client would drift the first time a
   * group was added. So the groups are accumulated in first-seen order and a
   * Map preserves it.
   */
  const grouped = useMemo(() => {
    const q = filter.trim().toLowerCase();
    const matches = (t: SnippetTemplateInfo) =>
      !q ||
      t.name.includes(q) ||
      t.title.toLowerCase().includes(q) ||
      t.description.toLowerCase().includes(q);

    const byCat = new Map<string, {category: string; title: string; templates: SnippetTemplateInfo[]}>();
    for (const t of offered) {
      if (!matches(t)) continue;
      const key = t.category ?? "";
      if (!byCat.has(key)) {
        byCat.set(key, {category: key, title: t.category_title || key, templates: []});
      }
      byCat.get(key)!.templates.push(t);
    }
    return [...byCat.values()];
  }, [offered, filter]);

  // Default to the first template so the primary action is never blocked on a
  // choice the user has no opinion about yet.
  const chosen = template ?? offered[0]?.name ?? null;
  const chosenTemplate = offered.find((t) => t.name === chosen);

  /**
   * The runtimes this template can actually satisfy.
   *
   * Not every template can be any length. `story` is built from eight or more
   * beats, so a twenty-second story is not a short clip — it is a plan whose
   * own rules contradict each other, and it fails after three correction
   * rounds having spent a chunk of the day's tokens finding out. Offering the
   * option was the bug; the picker now only shows what the template can do.
   */
  const runtimes = TARGETS.filter((t) => t.sec >= (chosenTemplate?.min_target_sec ?? 0));
  // Prefer the template's own default, then the shared one if this template
  // can satisfy it, and only then the shortest it can. Falling straight to the
  // shortest made every template default to the briefest clip it allowed.
  const fallbackRuntime =
    chosenTemplate?.default_target_sec ||
    runtimes.find((t) => t.sec === DEFAULT_TARGET_SEC)?.sec ||
    runtimes[0]?.sec ||
    DEFAULT_TARGET_SEC;
  // An explicit choice wins, but only while it is still valid for the template
  // the user has since picked.
  const effectiveTargetSec =
    targetSec !== null && runtimes.some((t) => t.sec === targetSec) ? targetSec : fallbackRuntime;

  /**
   * Picking a template fills the prompt with that template's own example.
   *
   * Every template is good at a different kind of subject, so an empty box
   * after choosing one is the worst moment to leave someone: the card just
   * told them what it does and the next step is inventing a prompt that suits
   * it. The example is already the answer.
   *
   * It only overwrites text that came from an example, never something typed.
   * That is checked against the whole example set rather than tracked with a
   * dirty flag, so switching between templates keeps swapping the demo — which
   * is what browsing them feels like — while a single typed character makes
   * the box the user's for good.
   */
  const selectTemplate = (name: string) => {
    setTemplate(name);
    const examples = new Set((templates.data ?? []).map((t) => t.example.trim()));
    const untouched = prompt.trim() === "" || examples.has(prompt.trim());
    if (untouched) {
      const next = offered.find((t) => t.name === name)?.example ?? "";
      setPrompt(next);
    }
  };

  // A script is only sent to a template that will actually speak it, so
  // switching away from `spine` with text still in the box cannot silently
  // change what the clip says.
  const script = chosenTemplate?.takes_narration ? narration.trim() : "";

  const create = async (planOnly: boolean) => {
    // A script stands in for the prompt: somebody who has already written the
    // words should not have to also describe them.
    const text = prompt.trim() || script;
    if (!text || !chosen) return;
    setBusy(true);
    setError(null);
    try {
      const created = await api.createSnippet({
        prompt: text,
        narration: script || undefined,
        title: title.trim() || undefined,
        template: chosen,
        // A script sets its own runtime — it is as long as it is — so the
        // picker's value is not sent with one. Sending it would ask the
        // pipeline to fit the creator's words into somebody's guess.
        target_sec: script ? undefined : effectiveTargetSec,
        captions,
        mode,
        skin,
        plan_only: planOnly,
      });
      setPrompt("");
      setTitle("");
      setNarration("");
      setTargetSec(null);
      setOpen(created.id);
      snippets.reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="text-lg font-semibold text-ink-100">{heading}</h1>
      <p className="mt-1 text-ink-400">{blurb}</p>

      {children}

      <section className="mt-5">
        <div className="mb-2.5 flex items-baseline justify-between gap-4">
          <h2 className="text-[11px] uppercase tracking-wide text-ink-500">{galleryHeading}</h2>
          {/* A filter, because the catalog passed twenty and scrolling a wall of
              cards to find one you already know the name of is the slowest way
              to use this page. It matches the title and description as well as
              the name — somebody looking for "does it fit" should land on
              `gauge` without knowing that is what it is called. */}
          <input
            type="search"
            aria-label="Filter templates"
            className="w-56 rounded-md border border-ink-800 bg-ink-950 px-2.5 py-1 text-[13px] text-ink-100 placeholder:text-ink-500 focus:border-brand focus:outline-none"
            placeholder="Filter templates…"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          />
        </div>
        {templates.error && <ErrorNote error={templates.error} onRetry={templates.reload} />}
        {templates.loading && !templates.data && <div className="text-ink-500">Loading…</div>}
        {grouped.length === 0 && !templates.loading && (
          <div className="rounded-lg border border-ink-800 p-4 text-ink-500">
            Nothing matches “{filter}”.
          </div>
        )}
        {grouped.map((g) => (
          <div key={g.category} className="mb-5 last:mb-0">
            {/* The heading is the job, not the mechanism: somebody arriving
                knows what they want to say, not which template says it. */}
            <h3 className="mb-2 border-b border-ink-800 pb-1.5 text-[12px] font-semibold uppercase tracking-wide text-ink-300">
              {g.title}
              <span className="ml-2 font-normal normal-case tracking-normal text-ink-500">
                {g.templates.length}
              </span>
            </h3>
            <div className="grid gap-3 sm:grid-cols-2">
              {g.templates.map((t) => (
                <TemplateCard
                  key={t.name}
                  template={t}
                  selected={t.name === chosen}
                  onSelect={() => selectTemplate(t.name)}
                />
              ))}
            </div>
          </div>
        ))}
      </section>

      <section className="mt-6">
        <label
          className="mb-2 block text-[11px] uppercase tracking-wide text-ink-500"
          htmlFor="snippet-prompt"
        >
          What should it teach?
        </label>
        <textarea
          id="snippet-prompt"
          className="min-h-[84px] w-full resize-y rounded-lg border border-ink-800 bg-ink-950 p-3 text-[14px] text-ink-100 placeholder:text-ink-500 focus:border-brand focus:outline-none"
          placeholder={chosenTemplate?.example ?? "How for loops work in Python"}
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
        />

        {/* The script box, offered only by the templates that speak one word
            for word. Below the prompt rather than instead of it: the prompt is
            still what the clip is FOR, and the planner uses it to decide what
            each line should be looking at. */}
        {chosenTemplate?.takes_narration && (
          <>
            <label
              className="mb-2 mt-4 block text-[11px] uppercase tracking-wide text-ink-500"
              htmlFor="snippet-narration"
            >
              Your narration{" "}
              <span className="normal-case tracking-normal text-ink-500">
                — optional. Written here, it is spoken word for word and only the pictures are
                decided for you.
              </span>
            </label>
            <textarea
              id="snippet-narration"
              className="min-h-[120px] w-full resize-y rounded-lg border border-ink-800 bg-ink-950 p-3 font-mono text-[13px] leading-relaxed text-ink-100 placeholder:text-ink-500 focus:border-brand focus:outline-none"
              placeholder={
                "Paste or write the script. Every word of it gets spoken, in this order.\n\nThe clip's length comes from the script, so the runtime picker is ignored while there is one here."
              }
              value={narration}
              onChange={(e) => setNarration(e.target.value)}
            />
            {script && (
              <div className="mt-1.5 text-[12px] text-ink-500">
                {wordCount(script)} words — about {Math.round((wordCount(script) * 60) / 150)}s at
                the house pace.
              </div>
            )}
          </>
        )}

        <label
          className="mb-2 mt-4 block text-[11px] uppercase tracking-wide text-ink-500"
          htmlFor="snippet-title"
        >
          Title <span className="normal-case tracking-normal text-ink-500">— optional, the model writes one otherwise</span>
        </label>
        <input
          id="snippet-title"
          className="w-full rounded-lg border border-ink-800 bg-ink-950 p-3 text-[14px] text-ink-100 placeholder:text-ink-500 focus:border-brand focus:outline-none"
          placeholder="Leave empty to let the model title it"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />

        <div className="mt-3 flex flex-wrap items-center gap-3">
          <Segmented
            value={effectiveTargetSec}
            onChange={setTargetSec}
            options={
              runtimes.length > 0
                ? runtimes.map((t) => ({ value: t.sec, label: t.label, hint: t.hint }))
                : [
                    {
                      value: fallbackRuntime,
                      label: `~${fallbackRuntime}s`,
                      hint: "This template's own length",
                    },
                  ]
            }
          />
          <Segmented
            value={mode}
            onChange={setMode}
            options={[
              { value: "dark", label: "Dark", hint: "The default editorial look" },
              { value: "light", label: "Light", hint: "A paper-white clip, same branding" },
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

          <button
            type="button"
            className={PRIMARY_BTN}
            onClick={() => void create(false)}
            disabled={busy || (!prompt.trim() && !script) || !chosen || run.running}
            title={run.running ? "A generation run is already in progress" : undefined}
          >
            {busy ? "Starting…" : "Generate clip"}
          </button>
          <button
            type="button"
            className={SECONDARY_BTN}
            onClick={() => void create(true)}
            disabled={busy || (!prompt.trim() && !script) || !chosen || run.running}
            title="Plan the beats and code without synthesizing audio or rendering"
          >
            Plan only
          </button>
          {run.running && (
            <span className="text-[12px] text-ink-500">
              A pipeline is already running — wait for it to finish.
            </span>
          )}
        </div>

        {error && <div className="mt-3 text-[13px] text-error">{error}</div>}
      </section>

      {open && <SnippetViewer id={open} onClose={() => setOpen(null)} />}

      <section className="mt-8">
        <h2 className="mb-2.5 text-[11px] uppercase tracking-wide text-ink-500">
          Your snippets
        </h2>
        {snippets.error && <ErrorNote error={snippets.error} onRetry={snippets.reload} />}
        {snippets.data?.length === 0 && (
          <p className="text-[13px] text-ink-500">Nothing yet — make one above.</p>
        )}
        <div className="space-y-2">
          {(snippets.data ?? []).map((s) => (
            <SnippetRow
              key={s.id}
              snippet={s}
              onOpen={() => setOpen(s.id)}
              onDelete={async () => {
                // A delete can be refused — the server keeps a snippet its own
                // run is mid-stage on, because deleting the directory out from
                // under a running stage is what strands a half-generated one.
                // Swallowing that rejection would leave a button that does
                // nothing and says nothing.
                try {
                  setError(null);
                  await api.deleteSnippet(s.id);
                } catch (err) {
                  setError(err instanceof Error ? err.message : String(err));
                  return;
                }
                if (open === s.id) setOpen(null);
                snippets.reload();
              }}
            />
          ))}
        </div>
      </section>
    </div>
  );
}
