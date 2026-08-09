import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Camera,
  CheckCircle2,
  Circle,
  Download,
  Eye,
  FileText,
  Loader2,
  Plus,
  Sparkles,
  Trash2,
  Video,
  XCircle,
} from "lucide-react";
import { api } from "../api/client";
import { lessonMatches, useRun, type LiveStageStatus } from "../state/RunContext";
import type {
  AvailableTake,
  NoCodeCatalogEntry,
  NoCodeDetail,
  NoCodeSummary,
  NoCodeTakes,
  Recordable,
} from "../api/types";

// The no-code surface.
//
// Every other builder page in the studio asks two questions — what do you want
// to say, and what should it look like. This one asks a third, and puts it
// where it cannot be skipped: **what is this standing on?**
//
// That is the whole surface, so the page is built around it rather than around
// the template gallery. A segment's evidence block is as prominent as its
// prompt, an unbacked segment cannot be submitted, and a piece that names
// recordings it has not made yet says so on its card instead of failing later.
//
// == Why the page is shaped the way it is ==
//
// Making a piece takes minutes and costs real tokens, so three things have to be
// true before somebody presses the button, and each one is a piece of the layout:
//
//  1. **Every field says whether it is needed and who reads it.** A form where
//     half the boxes are optional and one of them silently steers an LLM is a
//     form people fill in wrong once and then distrust. Hence the two markers —
//     `Required` and the eye — and a legend that names them.
//  2. **There is one button, it says what it does, and it is the only blue
//     thing on the page.** "Create piece" used to write a spec and stop, with
//     the actual generate hidden in a detail panel you could only reach by first
//     creating and then clicking the thing you just made. Now creating and
//     generating are one action, and saving without generating is the quiet
//     secondary.
//  3. **The run is visible and the file has an address.** Generating used to
//     fire and return nothing: no progress, no error, no player, and the video
//     URL the API advertised pointed at a route that did not exist. So the piece
//     view shows the real stage list ticking over, and afterwards the player,
//     a download, and the path on disk.
//
// The template picker is a grouped select rather than the snippets gallery for
// the same reason the combos page uses one: you are choosing for the third
// segment of five, and the choice has to sit inline beside its prompt.

/** The synthetic course pieces live in; run events are tagged with this slug. */
const NOCODE_COURSE = "nocode";

type EvidenceKind = "capture" | "fact";

interface DraftSegment {
  template: string;
  prompt: string;
  kind: EvidenceKind;
  tool: string;
  of: string;
  take: string;
  facts: string;
}

const emptySegment = (template: string): DraftSegment => ({
  template,
  prompt: "",
  kind: template === "footage" ? "capture" : "fact",
  tool: "claude",
  of: "",
  take: "",
  facts: "",
});

const secs = (ms?: number) => (ms ? `${Math.round(ms / 1000)}s` : "");

/**
 * What each stage of a run is doing, in words a founder can read.
 *
 * The engine's stage names are the engine's business — `scenegraph` means
 * nothing to somebody waiting on a video, and a progress list nobody can read is
 * the same as a spinner.
 */
const STAGE_LABEL: Record<string, string> = {
  capture: "Recording the tools",
  substance: "Checking the facts",
  plan: "Writing the segments",
  verify: "Running the code for real",
  audio: "Recording the voiceover",
  align: "Timing the words",
  captions: "Captions",
  chapters: "Chapters",
  scenegraph: "Laying out the scenes",
  render: "Rendering the video",
};

/**
 * Who actually shoots each kind of recording.
 *
 * The question the page could not answer, and the one that decides whether
 * pressing Generate finishes or stops: two of these three kinds run themselves,
 * and the third stops and waits for a person the studio cannot reach. A run
 * with a desktop capture in it refuses outright — the engine wants somebody at
 * the keyboard and a browser tab is not a keyboard — so that has to be said
 * while the tool is being chosen, not discovered when the run dies.
 */
const WHO_RECORDS: Record<string, { automatic: boolean; note: string }> = {
  terminal: {
    automatic: true,
    note: "Recorded for you. The tool really runs in a scratch directory and the session is filmed — nobody has to be at the keyboard.",
  },
  web: {
    automatic: true,
    note: "Recorded for you, in a real browser signed in as you. Do `coursesmith footage login <tool>` once first, or the recorder ends up filming a login page.",
  },
  desktop: {
    automatic: false,
    note: "You record this one — it films your screen while you work through the take's beats, so it cannot run from the studio. In a terminal: `coursesmith footage shoot nocode/<piece>`.",
  },
};

/** A worked example, so the first piece is a change rather than a blank page. */
const EXAMPLE = {
  title: "Ship a landing page without writing code",
  brief:
    "What a founder with no engineering help can actually get live in an afternoon: which tool does which part, what it costs, and what it looks like when it really runs.",
  segments: [
    {
      ...emptySegment("footage"),
      prompt: "Watch an AI agent build the page from a plain-English description.",
      kind: "capture" as EvidenceKind,
      tool: "claude",
      of: "ask the agent to scaffold a landing page with a signup form, then run it locally",
    },
    {
      ...emptySegment("costing"),
      prompt: "What this actually costs to run for a month.",
      kind: "fact" as EvidenceKind,
      facts: "Hosting on a hobby plan is free\nA custom domain is about $12 a year\nThe AI agent is $20 a month",
    },
    {
      ...emptySegment("verdict"),
      prompt: "Who this approach is right for, and who should not use it.",
      kind: "fact" as EvidenceKind,
      facts:
        "Right for a landing page, a waitlist, an internal tool\nWrong for anything holding payment details or medical records\nYou own the code either way — it is a normal repo",
    },
  ],
};

/** `Required`, and its quieter twin. The single most-asked question this page had. */
function Need({ required }: { required?: boolean }) {
  return required ? (
    <span className="rounded bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide text-amber-300">
      required
    </span>
  ) : (
    <span className="rounded bg-ink-800 px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-ink-500">
      optional
    </span>
  );
}

/**
 * "The model reads this."
 *
 * Worth its own marker because the answer is not guessable. The title and the
 * brief are fed to the fact-finder; a segment's prompt and its facts are what
 * the writer plans from; `of` is what the recording is asked to show. A take
 * filename and a fixture are files — no model ever sees them, and knowing which
 * is which is the difference between writing a prompt and writing a label.
 */
function ModelReads({ what }: { what: string }) {
  return (
    <span
      className="flex items-center gap-1 text-[10px] uppercase tracking-wide text-sky-400/80"
      title={what}
    >
      <Eye size={11} /> the model reads this
    </span>
  );
}

/** One labelled field, so every input on the page answers the same questions. */
function Field({
  label,
  required,
  modelReads,
  hint,
  children,
}: {
  label: string;
  required?: boolean;
  modelReads?: string;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="mt-4 first:mt-0">
      <div className="mb-1.5 flex flex-wrap items-center gap-2">
        <span className="text-[11px] font-medium uppercase tracking-wide text-ink-400">
          {label}
        </span>
        <Need required={required} />
        {modelReads && <ModelReads what={modelReads} />}
      </div>
      {children}
      {hint && <p className="mt-1.5 text-[12px] leading-snug text-ink-500">{hint}</p>}
    </div>
  );
}

function Step({ n, title, sub }: { n: number; title: string; sub?: string }) {
  return (
    <div className="mb-3 flex items-baseline gap-2.5">
      <span className="grid h-5 w-5 shrink-0 place-items-center rounded-full bg-ink-800 text-[11px] font-medium text-ink-300">
        {n}
      </span>
      <h2 className="text-[13px] font-semibold text-ink-100">{title}</h2>
      {sub && <span className="text-[12px] text-ink-500">{sub}</span>}
    </div>
  );
}

const inputClass =
  "w-full rounded-md border border-ink-800 bg-ink-950 px-3 py-2 text-[13px] text-ink-100 placeholder:text-ink-600 focus:border-brand focus:outline-none";

export function NoCodePage() {
  const { run, liveStages, logs, cancelRun, refreshTick, lastError } = useRun();

  const [pieces, setPieces] = useState<NoCodeSummary[]>([]);
  const [catalog, setCatalog] = useState<NoCodeCatalogEntry[]>([]);
  const [recordables, setRecordables] = useState<Recordable[]>([]);
  const [takes, setTakes] = useState<NoCodeTakes>({ dir: "", takes: [] });
  const [detail, setDetail] = useState<NoCodeDetail | null>(null);
  const [view, setView] = useState<"new" | "piece">("new");
  const [title, setTitle] = useState("");
  const [brief, setBrief] = useState("");
  const [segments, setSegments] = useState<DraftSegment[]>([emptySegment("footage")]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const refresh = useCallback(
    () => api.noCodePieces().then(setPieces).catch(() => {}),
    [],
  );

  useEffect(() => {
    refresh();
    api.noCodeCatalog().then(setCatalog).catch(() => {});
    api.noCodeRecordables().then(setRecordables).catch(() => {});
    api.noCodeTakes().then(setTakes).catch(() => {});
  }, [refresh]);

  // A finished run is the moment the answer to "where did it go" changes, so
  // that is when the piece is re-read rather than on a timer.
  useEffect(() => {
    if (refreshTick === 0) return;
    refresh();
    if (detail) api.noCodePiece(detail.id).then(setDetail).catch(() => {});
    // detail is deliberately not a dependency: this fires on run completion.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshTick, refresh]);

  const open = async (id: string) => {
    setView("piece");
    setError("");
    try {
      setDetail(await api.noCodePiece(id));
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  const byKind = useMemo(() => {
    const groups: Record<string, Recordable[]> = {};
    for (const r of recordables) (groups[r.kind] ??= []).push(r);
    return groups;
  }, [recordables]);

  const byCategory = useMemo(() => {
    const groups = new Map<string, NoCodeCatalogEntry[]>();
    for (const c of catalog) {
      if (!groups.has(c.category)) groups.set(c.category, []);
      groups.get(c.category)!.push(c);
    }
    return [...groups.entries()];
  }, [catalog]);

  const templateOf = (name: string) => catalog.find((c) => c.name === name);
  const needsCapture = (template: string) => templateOf(template)?.needs_capture ?? false;

  // A recordable that cannot be driven from a description needs a checked-in
  // take. Surfaced here so the choice is made while picking the tool, not
  // discovered when the browser opens.
  const recordableOf = (tool: string) => recordables.find((r) => r.key === tool);
  const takeRequired = (tool: string) => recordableOf(tool)?.needs_take ?? false;

  const update = (i: number, patch: Partial<DraftSegment>) =>
    setSegments((prev) => prev.map((s, n) => (n === i ? { ...s, ...patch } : s)));

  /**
   * Everything standing between this draft and a video, named.
   *
   * The engine refuses a hollow segment with a good message, but only after the
   * POST — so a disabled button with no explanation was the page's way of saying
   * "no" without saying why. These are the same rules, checked early, listed.
   */
  const problems = useMemo(() => {
    const out: string[] = [];
    if (!title.trim()) out.push("The piece needs a title.");
    segments.forEach((s, i) => {
      const n = i + 1;
      if (!s.prompt.trim()) out.push(`Segment ${n}: say what it covers.`);
      if (s.kind === "capture") {
        const rec = recordableOf(s.tool);
        if (!rec) out.push(`Segment ${n}: pick a tool to record.`);
        else if (rec.needs_take && !s.take.trim())
          out.push(
            `Segment ${n}: recording ${rec.display} needs a take file — nobody can invent selectors for somebody else's page.`,
          );
        else if (!rec.needs_take && !s.of.trim())
          out.push(`Segment ${n}: say what the recording should show.`);
      } else if (!s.facts.trim()) {
        out.push(`Segment ${n}: it is backed by facts, so write at least one.`);
      }
    });
    return out;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [title, segments, recordables]);

  const loadExample = () => {
    setTitle(EXAMPLE.title);
    setBrief(EXAMPLE.brief);
    setSegments(EXAMPLE.segments.map((s) => ({ ...s })));
    setError("");
  };

  /**
   * Create the piece and, unless told otherwise, generate it.
   *
   * One action, because "create" on its own produced a thing that looked
   * finished and was not. Saving without generating stays available — a piece
   * whose web takes are not written yet is a real and normal state — but it is
   * the secondary.
   */
  const submit = async (alsoGenerate: boolean) => {
    setBusy(true);
    setError("");
    try {
      const created = await api.createNoCodePiece({
        title,
        brief,
        segments: segments.map((s) => ({
          template: s.template,
          prompt: s.prompt,
          kind: s.kind,
          ...(s.kind === "capture"
            ? { tool: s.tool, of: s.of, take: s.take }
            : { facts: s.facts.split("\n").map((f) => f.trim()).filter(Boolean) }),
        })),
      });
      if (alsoGenerate) await api.runNoCodePiece(created.id);
      setTitle("");
      setBrief("");
      setSegments([emptySegment("footage")]);
      await refresh();
      setView("piece");
      setDetail(await api.noCodePiece(created.id));
    } catch (e) {
      // The engine's validation messages are written to be read by whoever is
      // editing, so they are shown as-is rather than replaced with "invalid".
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  const generate = async (id: string) => {
    setBusy(true);
    setError("");
    try {
      await api.runNoCodePiece(id);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  // --- live run state, narrowed to this piece ---

  const runningThis =
    run.running && run.course === NOCODE_COURSE && lessonMatches(run.lesson, detail?.id ?? " ");

  /**
   * Per-stage status for the open piece.
   *
   * Matched out of the global map rather than keyed directly, because the
   * backend reports lesson ids sometimes short and sometimes full — and after a
   * run finishes `run.lesson` is cleared, so a key built from it would take the
   * finished list off the screen at the exact moment it became the answer.
   */
  const stageStatus = useMemo(() => {
    const out: Record<string, LiveStageStatus> = {};
    if (!detail) return out;
    for (const [k, v] of Object.entries(liveStages)) {
      const [course, lesson, stage] = k.split("|");
      if (course === NOCODE_COURSE && lessonMatches(lesson, detail.id)) out[stage] = v;
    }
    return out;
  }, [liveStages, detail]);

  const anyStageSeen = Object.keys(stageStatus).length > 0;

  /**
   * The stage this piece's last run died on, if it did.
   *
   * Without this the page had no idea a run had failed. `lastError` lives in the
   * run context and was rendered only by the top bar — a 280px truncated chip in
   * the corner — so a run that started and died in half a second was, from here,
   * indistinguishable from a button that did nothing at all. That is the single
   * likeliest reading of "the generate button does not work", and it was not
   * about the button.
   */
  const failedStage = useMemo(
    () => Object.entries(stageStatus).find(([, v]) => v === "failed")?.[0],
    [stageStatus],
  );

  return (
    <div className="mx-auto max-w-6xl px-6 py-8">
      <header className="mb-6">
        <h1 className="text-2xl font-semibold text-ink-100">No-code pieces</h1>
        <p className="mt-2 max-w-2xl text-sm leading-relaxed text-ink-400">
          Every segment stands on something real — a recording of the tool actually
          running, or facts you can point at. Nothing here is drawn in place of
          evidence, which is what separates a piece from a combo about the same
          subject.
        </p>
        {/* What actually happens, before anybody spends ten minutes finding out. */}
        <ol className="mt-4 flex flex-wrap items-center gap-x-2 gap-y-1 text-[12px] text-ink-500">
          <li className="rounded-md border border-ink-800 px-2 py-1">
            <b className="text-ink-300">1</b> You describe the piece
          </li>
          <span>→</span>
          <li className="rounded-md border border-ink-800 px-2 py-1">
            <b className="text-ink-300">2</b> The tools get recorded for real
          </li>
          <span>→</span>
          <li className="rounded-md border border-ink-800 px-2 py-1">
            <b className="text-ink-300">3</b> It is written, voiced and rendered
          </li>
          <span>→</span>
          <li className="rounded-md border border-ink-800 px-2 py-1">
            <b className="text-ink-300">4</b> An MP4 you can play here
          </li>
        </ol>
        {/* The legend for the two markers every field carries. */}
        <div className="mt-3 flex flex-wrap items-center gap-3 text-[11px] text-ink-500">
          <span className="flex items-center gap-1.5">
            <Need required /> must be filled in
          </span>
          <span className="flex items-center gap-1.5">
            <Need /> can be left blank
          </span>
          <span className="flex items-center gap-1.5">
            <ModelReads what="fed to the model" /> — it steers what gets written
          </span>
        </div>
      </header>

      <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_19rem]">
        <section>
          {view === "new" ? (
            <>
              <div className="mb-4 flex items-center justify-between gap-3">
                <h2 className="text-[15px] font-semibold text-ink-100">A new piece</h2>
                <button
                  className="flex items-center gap-1.5 rounded-md border border-ink-700 px-2.5 py-1.5 text-[12px] text-ink-300 hover:text-ink-100"
                  onClick={loadExample}
                >
                  <Sparkles size={13} /> Fill in an example
                </button>
              </div>

              <div className="rounded-xl border border-ink-800 bg-ink-900/40 p-5">
                <Step n={1} title="What is the piece?" />
                <Field
                  label="Title"
                  required
                  modelReads="It opens the brief the fact-finder searches from, and titles the finished video."
                >
                  <input
                    className={inputClass}
                    placeholder="Ship a landing page without writing code"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                  />
                </Field>
                <Field
                  label="What it is about"
                  modelReads="The whole input the fact-finding stage searches from. Leave it blank and the segment prompts are all it has to go on."
                  hint="Two or three sentences in your own words. Worth writing — it is what stops the facts being generic."
                >
                  <textarea
                    className={`${inputClass} resize-y`}
                    rows={3}
                    placeholder="What a founder with no engineering help can get live in an afternoon…"
                    value={brief}
                    onChange={(e) => setBrief(e.target.value)}
                  />
                </Field>
              </div>

              <div className="mt-6">
                <Step
                  n={2}
                  title="The segments, in order"
                  sub={`${segments.length} · each one becomes a part of the video`}
                />
              </div>

              <ol className="space-y-4">
                {segments.map((seg, i) => {
                  const tpl = templateOf(seg.template);
                  const rec = recordableOf(seg.tool);
                  return (
                    <li
                      key={i}
                      className="rounded-xl border border-ink-800 bg-ink-900/40 p-5"
                    >
                      <div className="flex items-center gap-3">
                        <span className="grid h-6 w-6 shrink-0 place-items-center rounded-full bg-ink-800 text-xs text-ink-300">
                          {i + 1}
                        </span>
                        <select
                          className="min-w-0 flex-1 rounded-md border border-ink-700 bg-ink-950 px-2 py-1.5 text-sm text-ink-200 focus:border-brand focus:outline-none"
                          value={seg.template}
                          aria-label={`Look for segment ${i + 1}`}
                          onChange={(e) => {
                            const t = e.target.value;
                            update(i, {
                              template: t,
                              // A template whose frame IS a recording can only be
                              // backed by one, so the choice is made for you rather
                              // than offered and then refused.
                              kind: needsCapture(t) ? "capture" : seg.kind,
                            });
                          }}
                        >
                          {byCategory.map(([category, items]) => (
                            <optgroup key={category} label={category}>
                              {items.map((c) => (
                                <option key={c.name} value={c.name}>
                                  {c.title} ({c.name})
                                </option>
                              ))}
                            </optgroup>
                          ))}
                        </select>
                        {segments.length > 1 && (
                          <button
                            className="text-ink-500 hover:text-ink-300"
                            onClick={() =>
                              setSegments((p) => p.filter((_, n) => n !== i))
                            }
                            aria-label={`Remove segment ${i + 1}`}
                          >
                            <Trash2 size={16} />
                          </button>
                        )}
                      </div>
                      {/* What the chosen look actually does. A name alone makes
                          the picker a lottery for anybody who has not seen all
                          thirty of them. */}
                      {tpl?.description && (
                        <p className="mt-2 pl-9 text-[12px] leading-snug text-ink-500">
                          {tpl.description}
                        </p>
                      )}

                      <Field
                        label="What this segment covers"
                        required
                        modelReads="The writer plans this segment from it."
                      >
                        <textarea
                          className={`${inputClass} resize-y`}
                          rows={2}
                          placeholder="Watch an AI agent build the page from a plain-English description."
                          value={seg.prompt}
                          onChange={(e) => update(i, { prompt: e.target.value })}
                        />
                      </Field>

                      {/* The evidence block, as prominent as the prompt. */}
                      <div className="mt-4 rounded-md border border-dashed border-ink-700 p-3">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="text-[11px] font-medium uppercase tracking-wide text-ink-400">
                            Backed by
                          </span>
                          <Need required />
                          <div className="flex gap-1">
                            {(["capture", "fact"] as EvidenceKind[]).map((k) => (
                              <button
                                key={k}
                                disabled={needsCapture(seg.template) && k === "fact"}
                                onClick={() => update(i, { kind: k })}
                                className={`rounded px-2 py-0.5 text-xs ${
                                  seg.kind === k
                                    ? "bg-brand/20 text-brand"
                                    : "text-ink-400 hover:text-ink-200 disabled:opacity-40"
                                }`}
                              >
                                {k === "capture" ? "a recording" : "facts"}
                              </button>
                            ))}
                          </div>
                        </div>
                        <p className="mt-1.5 text-[12px] leading-snug text-ink-500">
                          {needsCapture(seg.template)
                            ? "This look is a recording or it is nothing, so the choice is made for you."
                            : seg.kind === "capture"
                              ? "We drive the real tool and film it. Slower, and the strongest thing a segment can stand on."
                              : "Claims you can point at. Written out here so a regenerated fact sheet cannot change what this segment stands on."}
                        </p>

                        {seg.kind === "capture" ? (
                          <>
                            <Field label="Which tool" required>
                              <select
                                className={inputClass}
                                value={seg.tool}
                                aria-label={`Tool for segment ${i + 1}`}
                                onChange={(e) => update(i, { tool: e.target.value })}
                              >
                                {Object.entries(byKind).map(([kind, list]) => (
                                  <optgroup key={kind} label={kind}>
                                    {list.map((r) => (
                                      <option key={r.key} value={r.key}>
                                        {r.display}
                                      </option>
                                    ))}
                                  </optgroup>
                                ))}
                              </select>
                            </Field>
                            {/* Who shoots it. The one thing that decides whether
                                pressing Generate finishes or stops dead. */}
                            {rec && <WhoRecords kind={rec.kind} />}
                            {takeRequired(seg.tool) ? (
                              <TakeField
                                index={i}
                                value={seg.take}
                                tool={seg.tool}
                                display={rec?.display ?? seg.tool}
                                kind={rec?.kind ?? "web"}
                                takes={takes}
                                onChange={(take) => update(i, { take })}
                              />
                            ) : (
                              <Field
                                label="What the recording should show"
                                required
                                modelReads="It is what the recording script is written from — the session really runs, so say what a beginner would actually do."
                              >
                                <input
                                  className={inputClass}
                                  placeholder="ask the agent to add a weekly summary, then run it"
                                  value={seg.of}
                                  onChange={(e) => update(i, { of: e.target.value })}
                                />
                              </Field>
                            )}
                          </>
                        ) : (
                          <Field
                            label="The facts"
                            required
                            modelReads="These become the segment's material — the writer builds from them, so a wrong figure here is wrong in the finished video."
                            hint="One claim per line."
                          >
                            <textarea
                              className={`${inputClass} resize-y font-mono text-[12px] leading-snug`}
                              rows={3}
                              placeholder={
                                "Hosting on a hobby plan is free\nA custom domain is about $12 a year"
                              }
                              value={seg.facts}
                              onChange={(e) => update(i, { facts: e.target.value })}
                            />
                          </Field>
                        )}
                      </div>
                    </li>
                  );
                })}
              </ol>

              <button
                className="mt-4 flex items-center gap-1.5 rounded-md border border-ink-700 px-3 py-1.5 text-sm text-ink-300 hover:text-ink-100"
                onClick={() => setSegments((p) => [...p, emptySegment("verdict")])}
              >
                <Plus size={15} /> Add a segment
              </button>

              <div className="mt-8">
                <Step n={3} title="Make the video" />
                {/* Why the button is off, in the page's own words rather than
                    the engine's — the engine's version arrives after the POST. */}
                {problems.length > 0 && (
                  <ul className="mb-3 space-y-1 rounded-lg border border-ink-800 bg-ink-900/40 p-3 text-[12.5px] text-ink-400">
                    {problems.map((p, n) => (
                      <li key={n} className="flex gap-2">
                        <Circle size={13} className="mt-0.5 shrink-0 text-ink-600" />
                        {p}
                      </li>
                    ))}
                  </ul>
                )}
                <div className="flex flex-wrap items-center gap-3">
                  <button
                    className="flex items-center gap-2 rounded-md bg-sky-600 px-4 py-2 text-sm font-medium text-white hover:bg-sky-500 disabled:opacity-40"
                    disabled={busy || problems.length > 0}
                    onClick={() => submit(true)}
                  >
                    {busy && <Loader2 size={15} className="animate-spin" />}
                    <Video size={15} /> Generate the video
                  </button>
                  <button
                    className="rounded-md border border-ink-700 px-3 py-2 text-sm text-ink-300 hover:text-ink-100 disabled:opacity-40"
                    disabled={busy || problems.length > 0}
                    onClick={() => submit(false)}
                  >
                    Save it, do not generate yet
                  </button>
                </div>
                <p className="mt-2 text-[12px] text-ink-500">
                  Generating records the tools, establishes the facts, writes and
                  voices the narration, and renders an MP4. It takes minutes and
                  spends tokens. Progress shows up right here.
                </p>
              </div>
            </>
          ) : (
            detail && (
              <PieceView
                detail={detail}
                recordables={recordables}
                takes={takes}
                busy={busy}
                running={runningThis}
                stageStatus={stageStatus}
                anyStageSeen={anyStageSeen}
                failedStage={failedStage}
                lastError={lastError}
                logs={logs}
                onGenerate={() => generate(detail.id)}
                onCancel={() => void cancelRun()}
                onNew={() => {
                  setView("new");
                  setDetail(null);
                  setError("");
                }}
              />
            )
          )}

          {error && (
            <pre className="mt-4 whitespace-pre-wrap rounded-md border border-red-900/60 bg-red-950/30 p-3 text-sm text-red-300">
              {error}
            </pre>
          )}
        </section>

        <aside>
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-xs uppercase tracking-wide text-ink-500">Pieces</h2>
            <button
              className="flex items-center gap-1 rounded border border-ink-700 px-2 py-0.5 text-[12px] text-ink-300 hover:text-ink-100"
              onClick={() => {
                setView("new");
                setDetail(null);
              }}
            >
              <Plus size={12} /> New
            </button>
          </div>
          {pieces.length === 0 && <p className="text-sm text-ink-500">Nothing yet.</p>}
          <ul className="space-y-2">
            {pieces.map((p) => (
              <li key={p.id}>
                <button
                  className={`w-full rounded-lg border bg-ink-900/40 p-3 text-left hover:border-ink-700 ${
                    detail?.id === p.id && view === "piece"
                      ? "border-brand/60"
                      : "border-ink-800"
                  }`}
                  onClick={() => void open(p.id)}
                >
                  <div className="truncate text-sm font-medium text-ink-100">
                    {p.title || p.id}
                  </div>
                  <div className="mt-1.5 flex flex-wrap items-center gap-2 text-xs text-ink-400">
                    <span>{p.segments} segments</span>
                    {p.captures > 0 && (
                      <span
                        className={`flex items-center gap-1 ${
                          p.recorded < p.captures ? "text-amber-400" : "text-ink-400"
                        }`}
                      >
                        <Camera size={12} />
                        {p.recorded}/{p.captures} recorded
                      </span>
                    )}
                    {p.ready ? (
                      <span className="flex items-center gap-1 text-emerald-400">
                        <Video size={12} /> ready
                      </span>
                    ) : (
                      <span className="text-ink-500">not generated</span>
                    )}
                  </div>
                </button>
              </li>
            ))}
          </ul>
        </aside>
      </div>
    </div>
  );
}

/**
 * What to do about a run that died, for the failures that are not about the
 * piece at all.
 *
 * A 401 from the model provider is the clearest case: the piece is fine, the
 * evidence is fine, nothing here can be edited to fix it — and the raw error
 * says "You do not have access to the organization tied to the API key", which
 * reads like a problem with courseSmith rather than with a key in the shell that
 * started the server. Kept to the handful worth naming; anything else shows its
 * own message, which the engine writes to be read.
 */
function diagnose(err: string): string | null {
  const e = err.toLowerCase();
  if (e.includes("401") || e.includes("api key") || e.includes("incorrect api key")) {
    return "This is the model provider rejecting the key, not a problem with the piece. Check the key in the shell that started `coursesmith serve` — the server reads it at launch, so a key exported afterwards will not have reached it. Restart the server after fixing it.";
  }
  if (e.includes("429") || e.includes("quota") || e.includes("rate limit")) {
    return "The provider is rate-limiting or out of quota. Wait and run it again, or point pipeline.llm_content at another provider.";
  }
  if (e.includes("reading take") || e.includes("no such file")) {
    return "A take file the run needed is not on disk. The path in the message is exactly where it was looked for.";
  }
  if (e.includes("somebody at the keyboard")) {
    return "This capture records your screen, so it cannot run from the studio. Shoot it from a terminal first, then generate again.";
  }
  if (e.includes("kokoro") || e.includes("tts")) {
    return "The voice server is not reachable. Start it, or set KOKORO_URL to where it is running.";
  }
  return null;
}

/** Who shoots this kind of recording, said where the tool is chosen. */
function WhoRecords({ kind }: { kind: string }) {
  const who = WHO_RECORDS[kind];
  if (!who) return null;
  return (
    <p
      className={`mt-2 text-[12px] leading-snug ${
        who.automatic ? "text-ink-500" : "text-amber-400"
      }`}
    >
      <b className="font-medium">{who.automatic ? "We record this." : "You record this."}</b>{" "}
      {who.note}
    </p>
  );
}

/**
 * The take field: a pick from the files that exist, not a spelling test.
 *
 * The name is the file's base name with no extension, because that is what the
 * recorder joins onto `takes/` — a placeholder reading `something.yaml` sends
 * people to `takes/something.yaml.yaml`, which fails minutes into a run with an
 * error about a file they are sure they wrote. A datalist rather than a plain
 * select because naming a take you have not written yet is legitimate: the piece
 * is saved, the take gets written, the run comes later. So an unknown name is a
 * warning with the full path in it, never a block.
 */
function TakeField({
  index,
  value,
  tool,
  display,
  kind,
  takes,
  onChange,
}: {
  index: number;
  value: string;
  tool: string;
  display: string;
  kind: string;
  takes: NoCodeTakes;
  onChange: (v: string) => void;
}) {
  const listID = `takes-${tool}`;
  const forTool: AvailableTake[] = takes.takes.filter((t) => t.tool === tool);
  const known = forTool.some((t) => t.name === value.trim());
  const looksLikeAFilename = /\.(ya?ml)$/i.test(value.trim());
  return (
    <Field
      label="Take file"
      required
      hint={`${display} is ${kind === "desktop" ? "a native app, so a take is the list of beats an operator works through" : "somebody else's website, and nobody can invent selectors for a page they do not own"} — so it is driven by a take you write once and keep, in ${takes.dir || "the course's takes/"}. Name it without the .yaml.`}
    >
      <input
        className={inputClass}
        list={forTool.length > 0 ? listID : undefined}
        placeholder={forTool[0]?.name ?? "first-build"}
        value={value}
        aria-label={`Take file for segment ${index + 1}`}
        onChange={(e) => onChange(e.target.value)}
      />
      {forTool.length > 0 && (
        <datalist id={listID}>
          {forTool.map((t) => (
            <option key={t.name} value={t.name}>
              {t.steps} steps
            </option>
          ))}
        </datalist>
      )}
      {forTool.length === 0 ? (
        <p className="mt-1.5 text-[12px] leading-snug text-amber-400">
          No take for {display} exists yet. You can name one now and write it before
          you generate.
        </p>
      ) : (
        <p className="mt-1.5 text-[12px] text-ink-500">
          Written already: {forTool.map((t) => t.name).join(", ")}
        </p>
      )}
      {value.trim() !== "" && !known && (
        <p className="mt-1 text-[12px] leading-snug text-amber-400">
          {looksLikeAFilename
            ? `Drop the extension — the engine adds it, so this would look for ${value.trim()}.yaml.`
            : `Nothing at ${takes.dir}/${value.trim()}.yaml yet. The piece saves fine; the run will stop here until it exists.`}
        </p>
      )}
    </Field>
  );
}

/** The stage list of a run, ticking over. */
function StageRow({ stage, status }: { stage: string; status?: LiveStageStatus }) {
  const label = STAGE_LABEL[stage] ?? stage;
  const icon =
    status === "running" ? (
      <Loader2 size={13} className="animate-spin text-sky-400" />
    ) : status === "done" ? (
      <CheckCircle2 size={13} className="text-emerald-400" />
    ) : status === "failed" ? (
      <XCircle size={13} className="text-red-400" />
    ) : (
      <Circle size={13} className="text-ink-700" />
    );
  return (
    <li className="flex items-center gap-2 text-[12.5px]">
      {icon}
      <span
        className={
          status === "running"
            ? "text-ink-100"
            : status === "done"
              ? "text-ink-400"
              : status === "failed"
                ? "text-red-300"
                : status === "skipped"
                  ? "text-ink-600 line-through"
                  : "text-ink-600"
        }
      >
        {label}
      </span>
    </li>
  );
}

/**
 * One piece: what it stands on, the run, and — the part that was missing — the
 * finished file with an address.
 */
function PieceView({
  detail,
  recordables,
  takes,
  busy,
  running,
  stageStatus,
  anyStageSeen,
  failedStage,
  lastError,
  logs,
  onGenerate,
  onCancel,
  onNew,
}: {
  detail: NoCodeDetail;
  recordables: Recordable[];
  takes: NoCodeTakes;
  busy: boolean;
  running: boolean;
  stageStatus: Record<string, LiveStageStatus>;
  anyStageSeen: boolean;
  failedStage?: string;
  lastError: string | null;
  logs: { seq: number; stage?: string; line: string }[];
  onGenerate: () => void;
  onCancel: () => void;
  onNew: () => void;
}) {
  const stages = detail.stages ?? [];

  // Which of the outstanding recordings the run can make on its own, and which
  // it will stop and wait for a person on. "2 still to shoot" was true and
  // useless — the number nobody could act on was how many of them were theirs.
  const kindOf = (tool?: string) => recordables.find((r) => r.key === tool)?.kind ?? "";
  const outstanding = detail.segment_list.filter(
    (s) => !s.skip && s.evidence.kind === "capture" && !s.evidence.recorded,
  );
  const byHand = outstanding.filter((s) => WHO_RECORDS[kindOf(s.evidence.tool)]?.automatic === false);
  const automatic = outstanding.length - byHand.length;

  // A segment naming a take that has not been written is the commonest way a
  // run of this surface dies, and it dies at the capture stage — minutes in,
  // after the pipeline has already started. Checking it here costs nothing.
  const missingTakes = outstanding.filter(
    (s) => s.evidence.take && !takes.takes.some((t) => t.name === s.evidence.take),
  );

  return (
    <div>
      <div className="mb-4 flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="truncate text-[16px] font-semibold text-ink-100">
            {detail.title || detail.id}
          </h2>
          <p className="mt-0.5 text-[12px] text-ink-500">
            {detail.segments} segments
            {detail.captures > 0 && ` · ${detail.recorded}/${detail.captures} recorded`}
          </p>
        </div>
        <button
          className="shrink-0 rounded border border-ink-700 px-2.5 py-1 text-[12px] text-ink-300 hover:text-ink-100"
          onClick={onNew}
        >
          New piece
        </button>
      </div>

      {/* The finished thing, first — it is what somebody opening a piece came
          for, and it used to be last and pointed at a dead URL. */}
      {detail.ready && detail.video_url && (
        <div className="mb-5 rounded-xl border border-ink-800 bg-ink-900/40 p-4">
          <video className="w-full rounded-lg" controls src={detail.video_url} />
          <div className="mt-3 flex flex-wrap items-center gap-3 text-[12px]">
            <a
              className="flex items-center gap-1.5 rounded-md border border-ink-700 px-2.5 py-1.5 text-ink-200 hover:text-ink-100"
              href={detail.video_url}
              download
            >
              <Download size={13} /> Download the MP4
            </a>
            {detail.video_path && (
              <span className="text-ink-500">
                on disk:{" "}
                <code className="rounded bg-ink-950 px-1.5 py-0.5 font-mono text-[11px] text-ink-300">
                  {detail.video_path}
                </code>
              </span>
            )}
          </div>
        </div>
      )}

      <div className="rounded-xl border border-ink-800 bg-ink-900/40 p-4">
        <div className="flex flex-wrap items-center gap-3">
          <button
            className="flex items-center gap-2 rounded-md bg-sky-600 px-4 py-2 text-sm font-medium text-white hover:bg-sky-500 disabled:opacity-40"
            disabled={busy || running}
            onClick={onGenerate}
          >
            {(busy || running) && <Loader2 size={15} className="animate-spin" />}
            <Video size={15} />
            {running ? "Generating…" : detail.ready ? "Generate it again" : "Generate the video"}
          </button>
          {running && (
            <button
              className="rounded-md border border-red-500/40 bg-red-500/10 px-3 py-2 text-[13px] text-red-300 hover:bg-red-500/20"
              onClick={onCancel}
            >
              Cancel
            </button>
          )}
          {!running && automatic > 0 && (
            <span className="text-[12px] text-ink-500">
              {automatic} recording{automatic > 1 ? "s" : ""} still to shoot — generating
              does {automatic > 1 ? "those" : "that"} for you first.
            </span>
          )}
        </div>

        {missingTakes.length > 0 && (
          <div className="mt-3 rounded-md border border-amber-500/30 bg-amber-500/5 p-3 text-[12px] leading-relaxed text-amber-300">
            <b className="font-medium">
              {missingTakes.length === 1 ? "A take file is" : `${missingTakes.length} take files are`}{" "}
              missing.
            </b>{" "}
            The run will stop at the capture stage until{" "}
            {missingTakes.length === 1 ? "it exists" : "they exist"}:
            <ul className="mt-1.5 space-y-1">
              {missingTakes.map((s) => (
                <li key={s.id}>
                  <code className="rounded bg-ink-950 px-1.5 py-0.5 font-mono text-[11px] text-ink-300">
                    {takes.dir}/{s.evidence.take}.yaml
                  </code>{" "}
                  <span className="text-amber-300/70">
                    — drives {s.evidence.display} for segment “{s.prompt}”
                  </span>
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* The one thing a run cannot do by itself, with the command that can. */}
        {byHand.length > 0 && (
          <div className="mt-3 rounded-md border border-amber-500/30 bg-amber-500/5 p-3 text-[12px] leading-relaxed text-amber-300">
            <b className="font-medium">
              {byHand.length} recording{byHand.length > 1 ? "s need" : " needs"} you at the
              keyboard.
            </b>{" "}
            {byHand.map((s) => s.evidence.display).join(", ")}{" "}
            {byHand.length > 1 ? "are native apps" : "is a native app"} — the recorder films
            your screen while you work through the take, so it cannot run from a browser
            tab. Generating from here will stop at that segment. Shoot{" "}
            {byHand.length > 1 ? "them" : "it"} first with:
            <code className="mt-1.5 block rounded bg-ink-950 px-2 py-1.5 font-mono text-[11px] text-ink-300">
              coursesmith footage shoot {NOCODE_COURSE}/{detail.id}
            </code>
          </div>
        )}

        {/* A failed run, where the run was. */}
        {!running && failedStage && lastError && (
          <div className="mt-3 rounded-md border border-red-900/60 bg-red-950/30 p-3">
            <p className="text-[12.5px] font-medium text-red-300">
              The last run stopped at “{STAGE_LABEL[failedStage] ?? failedStage}”.
            </p>
            <pre className="mt-1.5 whitespace-pre-wrap font-mono text-[11px] leading-snug text-red-300/90">
              {lastError}
            </pre>
            {diagnose(lastError) && (
              <p className="mt-2 text-[12px] leading-relaxed text-red-200/90">
                {diagnose(lastError)}
              </p>
            )}
          </div>
        )}

        {(running || anyStageSeen) && stages.length > 0 && (
          <>
            <ul className="mt-4 space-y-1.5">
              {stages.map((s) => (
                <StageRow key={s} stage={s} status={stageStatus[s]} />
              ))}
            </ul>
            {running && logs.length > 0 && (
              <pre className="mt-3 max-h-32 overflow-auto whitespace-pre-wrap rounded-md bg-ink-950 p-2.5 font-mono text-[11px] leading-snug text-ink-500">
                {logs.slice(-8).map((l) => l.line).join("\n")}
              </pre>
            )}
          </>
        )}
        {!running && !detail.ready && !anyStageSeen && (
          <p className="mt-3 text-[12px] text-ink-500">
            Nothing has been generated yet. The finished MP4 will land at{" "}
            <code className="rounded bg-ink-950 px-1.5 py-0.5 font-mono text-[11px] text-ink-400">
              {detail.video_path}
            </code>{" "}
            and play right here.
          </p>
        )}
      </div>

      <h3 className="mb-2 mt-6 text-xs uppercase tracking-wide text-ink-500">
        What each segment stands on
      </h3>
      <ul className="space-y-2">
        {detail.segment_list.map((s, i) => (
          <li key={s.id} className="rounded-lg border border-ink-800 bg-ink-900/40 p-3">
            <div className="flex items-center gap-2 text-[12px]">
              <span className="font-mono text-ink-600">{i + 1}</span>
              <span className="text-ink-300">{s.template_title}</span>
              <span className="font-mono text-[11px] text-ink-600">{s.template}</span>
              {s.skip && (
                <span className="rounded bg-amber-500/15 px-1.5 py-0.5 text-[10px] uppercase text-amber-300">
                  not in the cut
                </span>
              )}
            </div>
            {s.prompt && (
              <p className="mt-1 text-[12.5px] leading-snug text-ink-400">{s.prompt}</p>
            )}
            <div className="mt-1.5 flex items-center gap-1.5 text-[12px]">
              {s.evidence.kind === "capture" ? (
                <>
                  <Camera size={12} className="shrink-0 text-ink-500" />
                  <span
                    className={s.evidence.recorded ? "text-emerald-400" : "text-amber-400"}
                  >
                    {s.evidence.display}
                    {s.evidence.recorded
                      ? ` · ${secs(s.evidence.duration_ms)} recorded`
                      : " · not recorded yet"}
                  </span>
                </>
              ) : (
                <>
                  <FileText size={12} className="shrink-0 text-ink-500" />
                  <span className="text-ink-500">
                    {s.evidence.facts?.length ?? 0} fact
                    {(s.evidence.facts?.length ?? 0) === 1 ? "" : "s"}
                  </span>
                </>
              )}
            </div>
            {s.evidence.marks && s.evidence.marks.length > 0 && (
              <div className="mt-1 text-[11px] text-ink-600">
                moments: {s.evidence.marks.join(", ")}
              </div>
            )}
          </li>
        ))}
      </ul>

      <p className="mt-4 text-[12px] leading-snug text-ink-500">
        To change a segment, edit{" "}
        <code className="rounded bg-ink-950 px-1.5 py-0.5 font-mono text-[11px] text-ink-400">
          .coursesmith/nocode/lessons/{detail.id}/nocode.yaml
        </code>{" "}
        and generate again.
      </p>
    </div>
  );
}
