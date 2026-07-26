import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type ArtifactFile } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { ErrorNote } from "../components/ErrorNote";
import { DownloadButton } from "../components/DownloadButton";
import { formatBytes } from "../lib/format";

type Category = "video" | "audio" | "diagram" | "trace" | "data" | "captions" | "other";

const CATEGORY_LABEL: Record<Category, string> = {
  video: "Video",
  audio: "Audio",
  diagram: "Diagrams",
  trace: "Code traces",
  captions: "Captions",
  data: "Data",
  other: "Other",
};

const ORDER: Category[] = ["video", "audio", "diagram", "trace", "captions", "data", "other"];

function categorize(a: ArtifactFile): Category {
  const name = a.name.toLowerCase();
  if (/\.(mp4|mov|webm)$/.test(name)) return "video";
  if (/\.(wav|mp3|m4a|aac|ogg)$/.test(name)) return "audio";
  if (name.endsWith(".svg") || name.startsWith("diagrams/")) return "diagram";
  if (name.startsWith("code_traces/")) return "trace";
  if (/\.(vtt|srt)$/.test(name)) return "captions";
  if (name.endsWith(".json")) return "data";
  return "other";
}

/** Browse and download the generated artifacts for one lesson. */
export function ResultsGalleryPage() {
  const { slug = "", id = "" } = useParams();
  const { data, loading, error, reload } = useFetch(() => api.lesson(slug, id), [slug, id]);
  const [active, setActive] = useState<Category | "all">("all");

  const artifacts = data?.artifacts ?? [];
  const grouped = new Map<Category, ArtifactFile[]>();
  for (const a of artifacts) {
    const cat = categorize(a);
    (grouped.get(cat) ?? grouped.set(cat, []).get(cat)!).push(a);
  }
  const presentCats = ORDER.filter((c) => grouped.has(c));

  return (
    <div className="mx-auto max-w-4xl p-6">
      <div className="mb-1 text-[12px] text-ink-500">
        <Link className="hover:underline" to={`/c/${encodeURIComponent(slug)}`}>
          {slug}
        </Link>{" "}
        / <span className="text-ink-400">{data?.title ?? id}</span>
      </div>
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-lg font-semibold text-ink-100">Results</h1>
        <Link
          to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(id)}`}
          className="rounded border border-ink-700 bg-ink-800 px-3 py-1 text-[13px] text-ink-200 hover:bg-ink-700"
        >
          Lesson workbench →
        </Link>
      </div>

      {error && <ErrorNote error={error} onRetry={reload} />}
      {loading && !data && <div className="text-ink-500">Loading…</div>}

      {data && artifacts.length === 0 && (
        <div className="rounded-lg border border-ink-800 bg-ink-900 p-6 text-ink-500">
          No artifacts yet. Run the pipeline from the{" "}
          <Link className="text-sky-300 hover:underline" to="/generation">
            Generation
          </Link>{" "}
          page.
        </div>
      )}

      {artifacts.length > 0 && (
        <>
          <div className="mb-4 flex flex-wrap gap-2">
            <TabChip label={`All (${artifacts.length})`} active={active === "all"} onClick={() => setActive("all")} />
            {presentCats.map((c) => (
              <TabChip
                key={c}
                label={`${CATEGORY_LABEL[c]} (${grouped.get(c)!.length})`}
                active={active === c}
                onClick={() => setActive(c)}
              />
            ))}
          </div>

          <div className="space-y-6">
            {presentCats
              .filter((c) => active === "all" || active === c)
              .map((c) => (
                <section key={c}>
                  <h2 className="mb-2 text-[11px] uppercase tracking-wide text-ink-500">
                    {CATEGORY_LABEL[c]}
                  </h2>
                  <div className="grid gap-3 sm:grid-cols-2">
                    {grouped.get(c)!.map((a) => (
                      <ArtifactCard key={a.name} artifact={a} category={c} />
                    ))}
                  </div>
                </section>
              ))}
          </div>
        </>
      )}
    </div>
  );
}

function TabChip({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={
        "rounded-full border px-3 py-1 text-[12px] " +
        (active
          ? "border-sky-500/50 bg-sky-500/10 text-sky-200"
          : "border-ink-700 bg-ink-800 text-ink-300 hover:bg-ink-700")
      }
    >
      {label}
    </button>
  );
}

function ArtifactCard({ artifact, category }: { artifact: ArtifactFile; category: Category }) {
  const base = artifact.name.split("/").pop() ?? artifact.name;
  return (
    <div className="overflow-hidden rounded-lg border border-ink-800 bg-ink-900">
      <Preview artifact={artifact} category={category} />
      <div className="p-3">
        <div className="truncate font-mono text-[12px] text-ink-200" title={artifact.name}>
          {artifact.name}
        </div>
        <div className="mt-0.5 text-[11px] text-ink-500">{formatBytes(artifact.size)}</div>
        <div className="mt-2">
          <DownloadButton url={artifact.url} filename={base} size={artifact.size} />
        </div>
      </div>
    </div>
  );
}

function Preview({ artifact, category }: { artifact: ArtifactFile; category: Category }) {
  if (category === "video") {
    return <video src={artifact.url} controls className="max-h-56 w-full bg-black" preload="metadata" />;
  }
  if (category === "audio") {
    return (
      <div className="flex items-center bg-ink-950 p-3">
        <audio src={artifact.url} controls className="w-full" preload="none" />
      </div>
    );
  }
  if (category === "diagram" && artifact.name.toLowerCase().endsWith(".svg")) {
    return (
      <div className="flex max-h-56 items-center justify-center bg-white p-2">
        <img src={artifact.url} alt={artifact.name} className="max-h-52 max-w-full" />
      </div>
    );
  }
  return null;
}
