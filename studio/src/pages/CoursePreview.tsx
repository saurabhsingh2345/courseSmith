import { lazy, Suspense, useMemo, type ReactNode } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type ArtifactFile } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { ErrorNote } from "../components/ErrorNote";
import { AdaptiveOverlay } from "../components/AdaptiveOverlay";
import { Tabs, TabsContent, TabsList, TabsTrigger, Badge } from "../components/base";

// CoursePreview is the stakeholder-facing "everything in one place" view for a
// single lesson (workstream H): video, diagrams, live code-execution, quiz, and
// adaptive-learning estimates as tabs over the lesson's REAL generated artifacts
// (api.lesson → LessonDetail). Nothing here is mocked — the code tab embeds the
// same renderer component that produces the final video, the diagrams are the
// exact SVGs the pipeline drew, and the adaptive tab calls the tutor service.

// @remotion/player is heavy; only load it when the Code tab is actually opened.
const CodeTracePlayer = lazy(() => import("../components/CodeTracePlayer"));

function isDiagram(a: ArtifactFile): boolean {
  return (
    a.name.startsWith("diagrams/") &&
    a.name.endsWith(".svg") &&
    !a.name.includes("/attempts/")
  );
}

function isCodeTrace(a: ArtifactFile): boolean {
  return (
    a.name.startsWith("code_traces/") &&
    a.name.endsWith(".json") &&
    !a.name.endsWith("manifest.json")
  );
}

function baseName(name: string): string {
  const slash = name.lastIndexOf("/");
  return slash === -1 ? name : name.slice(slash + 1);
}

export function CoursePreview() {
  const { slug = "", id = "" } = useParams();
  const { data, loading, error, reload } = useFetch(() => api.lesson(slug, id), [slug, id]);

  const video = useMemo(
    () => data?.artifacts.find((a) => a.name === "final.mp4"),
    [data],
  );
  const diagrams = useMemo(() => data?.artifacts.filter(isDiagram) ?? [], [data]);
  const traces = useMemo(() => data?.artifacts.filter(isCodeTrace) ?? [], [data]);
  const quiz = data?.quiz;

  return (
    <div className="mx-auto max-w-5xl p-6">
      <div className="mb-4 flex items-center gap-2 text-ink-500">
        <Link to="/" className="hover:text-ink-200">
          Courses
        </Link>
        <span>/</span>
        <Link to={`/c/${encodeURIComponent(slug)}`} className="hover:text-ink-200">
          {slug}
        </Link>
        <span>/</span>
        <Link
          to={`/c/${encodeURIComponent(slug)}/l/${encodeURIComponent(id)}`}
          className="hover:text-ink-200"
        >
          {data?.id ?? id}
        </Link>
        <span>/</span>
        <span className="text-ink-200">preview</span>
      </div>

      {error && <ErrorNote error={error} onRetry={reload} />}
      {loading && !data && <div className="text-ink-500">Loading…</div>}

      {data && (
        <>
          <h1 className="mb-4 text-lg font-semibold text-ink-100">{data.title}</h1>

          <Tabs defaultValue="video">
            <TabsList>
              <TabsTrigger value="video">Video</TabsTrigger>
              <TabsTrigger value="diagrams">
                Diagrams
                {diagrams.length > 0 && <Count>{diagrams.length}</Count>}
              </TabsTrigger>
              <TabsTrigger value="code">
                Code
                {traces.length > 0 && <Count>{traces.length}</Count>}
              </TabsTrigger>
              <TabsTrigger value="quiz">
                Quiz
                {quiz?.questions?.length ? <Count>{quiz.questions.length}</Count> : null}
              </TabsTrigger>
              <TabsTrigger value="adaptive">Adaptive</TabsTrigger>
            </TabsList>

            {/* Video */}
            <TabsContent value="video">
              <Section title="Lesson video">
                {video ? (
                  <video
                    src={video.url}
                    controls
                    className="w-full max-w-3xl rounded-lg border border-ink-800"
                  />
                ) : (
                  <Empty>Video not generated yet — run the video stage.</Empty>
                )}
              </Section>
            </TabsContent>

            {/* Diagrams */}
            <TabsContent value="diagrams">
              <Section title="Diagrams">
                {diagrams.length > 0 ? (
                  <div className="grid grid-cols-[repeat(auto-fit,minmax(320px,1fr))] gap-4">
                    {diagrams.map((d) => (
                      <figure
                        key={d.name}
                        className="rounded-lg border border-ink-800 bg-ink-900 p-3"
                      >
                        <img
                          src={d.url}
                          alt={baseName(d.name)}
                          className="w-full rounded bg-white"
                        />
                        <figcaption className="mt-2 font-mono text-[11px] text-ink-500">
                          {baseName(d.name)}
                        </figcaption>
                      </figure>
                    ))}
                  </div>
                ) : (
                  <Empty>No diagrams yet — run the visuals stage.</Empty>
                )}
              </Section>
            </TabsContent>

            {/* Code execution */}
            <TabsContent value="code">
              <Section title="Code execution">
                {traces.length > 0 ? (
                  <div className="space-y-6">
                    {traces.map((t) => (
                      <div key={t.name}>
                        <div className="mb-1 font-mono text-[11px] text-ink-500">
                          {baseName(t.name)}
                        </div>
                        <Suspense
                          fallback={<div className="text-[12px] text-ink-500">Loading player…</div>}
                        >
                          <CodeTracePlayer url={t.url} title={data.title} />
                        </Suspense>
                      </div>
                    ))}
                  </div>
                ) : (
                  <Empty>No code traces yet — run the trace stage.</Empty>
                )}
              </Section>
            </TabsContent>

            {/* Quiz */}
            <TabsContent value="quiz">
              <Section title="Quiz preview">
                {quiz?.questions?.length ? (
                  <div className="space-y-3">
                    {quiz.questions.map((q, idx) => (
                      <div key={q.id} className="rounded-lg border border-ink-800 bg-ink-900 p-3">
                        <div className="mb-1 flex items-center gap-2">
                          <span className="text-ink-200">
                            Q{idx + 1}. {q.prompt}
                          </span>
                          <Badge variant="secondary">{q.type}</Badge>
                          {q.review && <Badge variant="outline">review</Badge>}
                        </div>
                        <ul className="mt-2 space-y-1">
                          {q.options.map((opt, oi) => (
                            <li
                              key={oi}
                              className={
                                oi === q.answer_index
                                  ? "text-success"
                                  : "text-ink-400"
                              }
                            >
                              {oi === q.answer_index ? "✓ " : "· "}
                              {opt}
                            </li>
                          ))}
                        </ul>
                        {q.explanation && (
                          <p className="mt-2 text-[12px] text-ink-500">{q.explanation}</p>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <Empty>No quiz yet — run the quiz stage.</Empty>
                )}
              </Section>
            </TabsContent>

            {/* Adaptive */}
            <TabsContent value="adaptive">
              <Section title="Adaptive learning">
                <AdaptiveOverlay quiz={quiz} />
              </Section>
            </TabsContent>
          </Tabs>
        </>
      )}
    </div>
  );
}

function Count({ children }: { children: ReactNode }) {
  return <span className="ml-1.5 text-[11px] text-ink-500">{children}</span>;
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="mt-4">
      <h2 className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-ink-500">
        {title}
      </h2>
      {children}
    </section>
  );
}

function Empty({ children }: { children: ReactNode }) {
  return (
    <div className="rounded-lg border border-dashed border-ink-800 p-6 text-center text-ink-500">
      {children}
    </div>
  );
}
