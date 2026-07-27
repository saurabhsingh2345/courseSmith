import { useState } from "react";
import { api } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { ErrorNote } from "../components/ErrorNote";

type Tab = "diagrams" | "questions";
const DIAGRAM_KINDS = ["svg", "d3", "mermaid", "excalidraw"];
const QUESTION_TYPES = ["recall", "application", "debugging", "prediction"];

/** Saved reusable assets: diagram specs and quiz questions, persisted server-
 *  side under the studio state dir. Create/delete round-trips through the API. */
export function LibraryPage() {
  const [tab, setTab] = useState<Tab>("diagrams");
  return (
    <div className="mx-auto max-w-3xl p-6">
      <h1 className="mb-1 text-lg font-semibold text-ink-100">Library</h1>
      <p className="mb-4 text-[13px] text-ink-500">Reusable diagrams and questions.</p>

      <div className="mb-4 flex gap-2">
        <TabChip label="Diagrams" active={tab === "diagrams"} onClick={() => setTab("diagrams")} />
        <TabChip label="Questions" active={tab === "questions"} onClick={() => setTab("questions")} />
      </div>

      {tab === "diagrams" ? <Diagrams /> : <Questions />}
    </div>
  );
}

function Diagrams() {
  const { data, error, reload } = useFetch(() => api.libraryDiagrams(), []);
  const [name, setName] = useState("");
  const [kind, setKind] = useState("mermaid");
  const [source, setSource] = useState("");
  const [err, setErr] = useState<string | null>(null);

  const add = async () => {
    if (!name.trim()) return;
    setErr(null);
    try {
      await api.createLibraryDiagram({ name: name.trim(), kind, source });
      setName("");
      setSource("");
      reload();
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  };
  const remove = async (id: string) => {
    try {
      await api.deleteLibraryDiagram(id);
      reload();
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  };

  return (
    <div>
      <div className="mb-5 rounded-lg border border-ink-800 bg-ink-900 p-4">
        <div className="mb-2 flex gap-2">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Name"
            className="flex-1 rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100 placeholder:text-ink-500"
          />
          <select
            value={kind}
            onChange={(e) => setKind(e.target.value)}
            className="rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100"
          >
            {DIAGRAM_KINDS.map((k) => (
              <option key={k} value={k}>
                {k}
              </option>
            ))}
          </select>
        </div>
        <textarea
          value={source}
          onChange={(e) => setSource(e.target.value)}
          rows={3}
          placeholder="Diagram source (e.g. Mermaid syntax or a prompt)"
          className="mb-2 w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 font-mono text-[12px] text-ink-200 placeholder:text-ink-500"
        />
        {err && <ErrorNote error={err} />}
        <button
          onClick={() => void add()}
          disabled={!name.trim()}
          className="rounded bg-sky-600 px-3 py-1 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
        >
          Save diagram
        </button>
      </div>

      {error && <ErrorNote error={error} onRetry={reload} />}
      <div className="space-y-2">
        {(data ?? []).map((d) => (
          <div
            key={d.id}
            className="flex items-center justify-between rounded-lg border border-ink-800 bg-ink-900 px-3 py-2"
          >
            <div className="min-w-0">
              <div className="truncate text-ink-100">
                {d.name} <span className="text-[11px] text-ink-500">{d.kind}</span>
              </div>
              {d.source && (
                <div className="truncate font-mono text-[11px] text-ink-500">{d.source}</div>
              )}
            </div>
            <button
              onClick={() => void remove(d.id)}
              className="shrink-0 rounded border border-red-500/30 px-2 py-0.5 text-[12px] text-red-300 hover:bg-red-500/10"
            >
              Delete
            </button>
          </div>
        ))}
        {data && data.length === 0 && <div className="text-ink-500">No saved diagrams.</div>}
      </div>
    </div>
  );
}

function Questions() {
  const { data, error, reload } = useFetch(() => api.libraryQuestions(), []);
  const [prompt, setPrompt] = useState("");
  const [type, setType] = useState("recall");
  const [options, setOptions] = useState("");
  const [answerIndex, setAnswerIndex] = useState(0);
  const [explanation, setExplanation] = useState("");
  const [err, setErr] = useState<string | null>(null);

  const add = async () => {
    if (!prompt.trim()) return;
    setErr(null);
    try {
      await api.createLibraryQuestion({
        prompt: prompt.trim(),
        type,
        options: options.split(",").map((o) => o.trim()).filter(Boolean),
        answer_index: answerIndex,
        explanation: explanation.trim(),
      });
      setPrompt("");
      setOptions("");
      setExplanation("");
      setAnswerIndex(0);
      reload();
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  };
  const remove = async (id: string) => {
    try {
      await api.deleteLibraryQuestion(id);
      reload();
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  };

  return (
    <div>
      <div className="mb-5 space-y-2 rounded-lg border border-ink-800 bg-ink-900 p-4">
        <div className="flex gap-2">
          <input
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="Question prompt"
            className="flex-1 rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100 placeholder:text-ink-500"
          />
          <select
            value={type}
            onChange={(e) => setType(e.target.value)}
            className="rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100"
          >
            {QUESTION_TYPES.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
        </div>
        <input
          value={options}
          onChange={(e) => setOptions(e.target.value)}
          placeholder="Options, comma-separated"
          className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-100 placeholder:text-ink-500"
        />
        <div className="flex items-center gap-2">
          <label className="text-[12px] text-ink-400">
            Answer index
            <input
              type="number"
              min={0}
              value={answerIndex}
              onChange={(e) => setAnswerIndex(Number(e.target.value))}
              className="ml-2 w-16 rounded border border-ink-800 bg-ink-950 px-2 py-1 text-ink-100"
            />
          </label>
        </div>
        <input
          value={explanation}
          onChange={(e) => setExplanation(e.target.value)}
          placeholder="Explanation"
          className="w-full rounded border border-ink-800 bg-ink-950 px-2 py-1.5 text-ink-200 placeholder:text-ink-500"
        />
        {err && <ErrorNote error={err} />}
        <button
          onClick={() => void add()}
          disabled={!prompt.trim()}
          className="rounded bg-sky-600 px-3 py-1 font-medium text-white hover:bg-sky-500 disabled:opacity-40"
        >
          Save question
        </button>
      </div>

      {error && <ErrorNote error={error} onRetry={reload} />}
      <div className="space-y-2">
        {(data ?? []).map((q) => (
          <div
            key={q.id}
            className="flex items-start justify-between gap-3 rounded-lg border border-ink-800 bg-ink-900 px-3 py-2"
          >
            <div className="min-w-0">
              <div className="text-ink-100">
                <span className="mr-2 rounded bg-ink-800 px-1.5 py-0.5 text-[10px] uppercase text-ink-400">
                  {q.type}
                </span>
                {q.prompt}
              </div>
              {q.options.length > 0 && (
                <div className="mt-0.5 text-[11px] text-ink-500">
                  {q.options.length} options · answer #{q.answer_index}
                </div>
              )}
            </div>
            <button
              onClick={() => void remove(q.id)}
              className="shrink-0 rounded border border-red-500/30 px-2 py-0.5 text-[12px] text-red-300 hover:bg-red-500/10"
            >
              Delete
            </button>
          </div>
        ))}
        {data && data.length === 0 && <div className="text-ink-500">No saved questions.</div>}
      </div>
    </div>
  );
}

function TabChip({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={
        "rounded-full border px-3 py-1 text-[13px] " +
        (active
          ? "border-sky-500/50 bg-sky-500/10 text-sky-200"
          : "border-ink-700 bg-ink-800 text-ink-300 hover:bg-ink-700")
      }
    >
      {label}
    </button>
  );
}
