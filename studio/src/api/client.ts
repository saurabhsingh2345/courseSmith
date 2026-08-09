import type { components } from "./schema";
import type {
  NoCodeSummary,
  NoCodeDetail,
  NoCodeCatalogEntry,
  NoCodeTakes,
  Recordable,
  CreateNoCodeRequest,
} from "./types";

export type Course = components["schemas"]["Course"];
export type CourseDetail = components["schemas"]["CourseDetail"];
export type LessonSummary = components["schemas"]["LessonSummary"];
export type LessonDetail = components["schemas"]["LessonDetail"];
export type ArtifactFile = components["schemas"]["Artifact"];
export type Script = components["schemas"]["Script"];
export type Quiz = components["schemas"]["Quiz"];
export type Question = components["schemas"]["Question"];
export type QuizOverrides = components["schemas"]["QuizOverrides"];
export type QuizWithOverrides = components["schemas"]["QuizWithOverrides"];
export type Chapter = components["schemas"]["Chapter"];
export type RunRequest = components["schemas"]["RunRequest"];
export type RunStatus = components["schemas"]["RunStatus"];
export type FeedbackRequest = components["schemas"]["FeedbackRequest"];
export type RegenerateRequest = components["schemas"]["RegenerateRequest"];
export type Ledger = components["schemas"]["Ledger"];
export type QuotaStatus = components["schemas"]["QuotaStatus"];
export type StageStatus = "done" | "stale" | "pending";

// Phase 7 editing/catalog/library shapes. These endpoints post-date the
// generated schema.d.ts, so their types live here alongside the calls.
export interface Archetype {
  name: string;
  description: string;
  default_animation: string;
  pace_hint: number;
  prompt_hint: string;
}
export interface PaletteColors {
  name: string;
  colors: { primary: string; accent: string; background: string };
}
export interface ArchetypeCatalog {
  archetypes: Archetype[];
  animation_styles: string[];
  palettes: PaletteColors[];
}
export interface CourseMeta {
  archetype: string;
  animation_style: string;
  color_palette: string;
  colors: { primary: string; accent: string; background: string };
}
// CourseDetail plus the editable style selection (backend adds `meta`; it
// post-dates the generated schema so we intersect it in here).
export type CourseDetailFull = CourseDetail & { meta: CourseMeta };
export interface UpdateCourseRequest {
  name?: string;
  description?: string;
  archetype?: string;
  animation_style?: string;
  color_palette?: string;
  colors?: { primary?: string; accent?: string; background?: string };
}
export interface LibraryDiagram {
  id: string;
  name: string;
  kind: string;
  source: string;
  created_at: string;
}
export interface LibraryQuestion {
  id: string;
  prompt: string;
  type: string;
  options: string[];
  answer_index: number;
  explanation: string;
  created_at: string;
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...init,
    headers: {
      ...(init?.body !== undefined ? { "Content-Type": "application/json" } : {}),
      ...init?.headers,
    },
  });
  if (!res.ok) {
    let message = `HTTP ${res.status}`;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      /* non-JSON error body */
    }
    throw new ApiError(res.status, message);
  }
  if (res.status === 204) return undefined as T;
  const text = await res.text();
  return (text === "" ? undefined : JSON.parse(text)) as T;
}

/** An unfiled lesson: drafted from a prompt, not yet in any course. */
export type DraftMeta = {
  id: string;
  prompt: string;
  title: string;
  summary: string;
  created_at: string;
  course?: string;
};

export type DraftDetail = DraftMeta & { source: string };

/** One card in the snippet template gallery. */
export type SnippetTemplateInfo = components["schemas"]["SnippetTemplateInfo"];
export type ComboSummary = components["schemas"]["ComboSummary"];
export type ComboDetail = components["schemas"]["ComboDetail"];
export type ComboSegmentInfo = components["schemas"]["ComboSegmentInfo"];

/** One requested segment when creating a combo. */
export interface CreateComboSegment {
  template: string;
  /** What this segment establishes — the increment the viewer gains, not a
   *  topic. That difference is what stops two segments arriving at the same
   *  content from different directions. */
  prompt: string;
  /** The director's reasoning, carried through so the page can show WHY the
   *  piece is shaped this way rather than only what it chose. All optional: a
   *  hand-built combo supplies none of them. */
  heading?: string;
  role?: string;
  why?: string;
  /** The concrete facts this template gets filled with. The director returns it
   *  with each proposed segment and it must be POSTed back: a segment created
   *  without it is planned from `prompt` alone, and its writer invents the
   *  specifics rather than using the ones already chosen. */
  material?: string;
  target_sec?: number;
}

/** The whole combo request: the subject, the argument, and the ordered segments. */
export interface CreateComboRequest {
  title?: string;
  brief?: string;
  /** What the piece argues. Must be POSTed back with a directed proposal — it
   *  is what the critic scores every finished segment against, and a combo
   *  created without one gets a critic that can only ask whether a segment is
   *  good, which returns opinions about prose. */
  angle?: string;
  segments: CreateComboSegment[];
  voice?: string;
  captions?: string;
  mode?: string;
  skin?: string;
  plan_only?: boolean;
}

/** The four choices a creator makes. Everything else the director decides. */
export interface DirectComboRequest {
  subject: string;
  title?: string;
  /** How long the piece should run. 0 reads a length out of the subject if one
   *  is stated there, and otherwise takes the default. */
  minutes?: number;
  /** The theme, which also decides which templates may be cast. */
  skin?: string;
  captions?: string;
  mode?: string;
}

/** The proposed piece. Nothing has been written yet. */
export interface DirectComboResponse {
  title: string;
  angle: string;
  /** Which catalog the theme narrowed the casting to, in words. */
  pool: string;
  /** How the runtime was spread over segments, and whether the ask could be met. */
  runtime?: string;
  segments: CreateComboSegment[];
}

/** A segment edit. Every field is optional: omitting one leaves it alone,
 *  which is what almost every edit means. */
export interface PatchComboSegmentRequest {
  template?: string;
  prompt?: string;
  material?: string;
  target_sec?: number;
  skip?: boolean;
}
export type SnippetSummary = components["schemas"]["SnippetSummary"];
export type SnippetDetail = components["schemas"]["SnippetDetail"];
export type CreateSnippetRequest = components["schemas"]["CreateSnippetRequest"];
export type CreateSnippetResponse = components["schemas"]["CreateSnippetResponse"];

/** One narrated step of a snippet, as the plan stage wrote it. */
export interface SnippetBeat {
  id: string;
  heading: string;
  narration: string;
  code?: string;
  run?: boolean;
}
export interface SnippetPlan {
  template: string;
  title: string;
  subtitle?: string;
  beats: SnippetBeat[];
}

export const api = {
  courses: () => request<Course[]>("/api/courses"),

  drafts: () => request<DraftMeta[]>("/api/drafts"),
  createDraft: (req: { prompt: string; course?: string }) =>
    request<DraftDetail>("/api/drafts", {
      method: "POST",
      body: JSON.stringify(req),
    }),
  draft: (id: string) => request<DraftDetail>(`/api/drafts/${encodeURIComponent(id)}`),
  updateDraft: (id: string, source: string) =>
    request<DraftDetail>(`/api/drafts/${encodeURIComponent(id)}`, {
      method: "PUT",
      body: JSON.stringify({ source }),
    }),
  deleteDraft: (id: string) =>
    request<void>(`/api/drafts/${encodeURIComponent(id)}`, { method: "DELETE" }),
  /** File a draft into a course — pass an existing slug or a new course name. */
  assignDraft: (id: string, req: { course?: string; new_course?: string }) =>
    request<LessonDetail>(`/api/drafts/${encodeURIComponent(id)}/assign`, {
      method: "POST",
      body: JSON.stringify(req),
    }),

  createCourse: (req: { name: string; description?: string }) =>
    request<Course>("/api/courses", {
      method: "POST",
      body: JSON.stringify(req),
    }),
  course: (slug: string) =>
    request<CourseDetailFull>(`/api/courses/${encodeURIComponent(slug)}`),
  updateCourse: (slug: string, req: UpdateCourseRequest) =>
    request<CourseDetailFull>(`/api/courses/${encodeURIComponent(slug)}`, {
      method: "PUT",
      body: JSON.stringify(req),
    }),
  deleteCourse: (slug: string) =>
    request<void>(`/api/courses/${encodeURIComponent(slug)}`, { method: "DELETE" }),
  lesson: (course: string, id: string) =>
    request<LessonDetail>(
      `/api/lessons/${encodeURIComponent(course)}/${encodeURIComponent(id)}`,
    ),
  createLesson: (course: string, req: { title: string }) =>
    request<LessonDetail>(`/api/lessons/${encodeURIComponent(course)}`, {
      method: "POST",
      body: JSON.stringify(req),
    }),
  updateLesson: (course: string, id: string, req: { source: string }) =>
    request<LessonDetail>(
      `/api/lessons/${encodeURIComponent(course)}/${encodeURIComponent(id)}`,
      { method: "PUT", body: JSON.stringify(req) },
    ),
  deleteLesson: (course: string, id: string) =>
    request<void>(
      `/api/lessons/${encodeURIComponent(course)}/${encodeURIComponent(id)}`,
      { method: "DELETE" },
    ),
  archetypes: () => request<ArchetypeCatalog>("/api/archetypes"),
  libraryDiagrams: () => request<LibraryDiagram[]>("/api/library/diagrams"),
  createLibraryDiagram: (d: Omit<LibraryDiagram, "id" | "created_at">) =>
    request<LibraryDiagram>("/api/library/diagrams", {
      method: "POST",
      body: JSON.stringify(d),
    }),
  deleteLibraryDiagram: (id: string) =>
    request<void>(`/api/library/diagrams/${encodeURIComponent(id)}`, { method: "DELETE" }),
  libraryQuestions: () => request<LibraryQuestion[]>("/api/library/questions"),
  createLibraryQuestion: (q: Omit<LibraryQuestion, "id" | "created_at">) =>
    request<LibraryQuestion>("/api/library/questions", {
      method: "POST",
      body: JSON.stringify(q),
    }),
  deleteLibraryQuestion: (id: string) =>
    request<void>(`/api/library/questions/${encodeURIComponent(id)}`, { method: "DELETE" }),
  runStatus: () => request<RunStatus>("/api/run"),
  startRun: (req: RunRequest) =>
    request<{ run_id: string }>("/api/run", {
      method: "POST",
      body: JSON.stringify(req),
    }),
  cancelRun: () => request<void>("/api/run", { method: "DELETE" }),
  feedback: (req: FeedbackRequest) =>
    request<unknown>("/api/feedback", {
      method: "POST",
      body: JSON.stringify(req),
    }),
  regenerate: (req: RegenerateRequest) =>
    request<{ run_id?: string; stages?: string[] }>("/api/regenerate", {
      method: "POST",
      body: JSON.stringify(req),
    }),
  quiz: (course: string, id: string) =>
    request<QuizWithOverrides>(
      `/api/quiz/${encodeURIComponent(course)}/${encodeURIComponent(id)}`,
    ),
  putQuizOverrides: (course: string, id: string, payload: QuizOverrides) =>
    request<QuizWithOverrides>(
      `/api/quiz-overrides/${encodeURIComponent(course)}/${encodeURIComponent(id)}`,
      { method: "PUT", body: JSON.stringify(payload) },
    ),
  ledger: () => request<Ledger>("/api/ledger"),

  snippetTemplates: () => request<SnippetTemplateInfo[]>("/api/snippet-templates"),
  combos: () => request<ComboSummary[]>("/api/combos"),
  directCombo: (req: DirectComboRequest) =>
    request<DirectComboResponse>("/api/combos/direct", {
      method: "POST",
      body: JSON.stringify(req),
    }),
  combo: (id: string) => request<ComboDetail>(`/api/combos/${encodeURIComponent(id)}`),
  createCombo: (req: CreateComboRequest) =>
    request<ComboSummary & { run_id?: string }>("/api/combos", {
      method: "POST",
      body: JSON.stringify(req),
    }),
  runCombo: (id: string) =>
    request<{ run_id: string }>(`/api/combos/${encodeURIComponent(id)}/run`, { method: "POST" }),
  patchComboSegment: (id: string, segment: string, req: PatchComboSegmentRequest) =>
    request<ComboSegmentInfo[]>(
      `/api/combos/${encodeURIComponent(id)}/segments/${encodeURIComponent(segment)}`,
      { method: "PATCH", body: JSON.stringify(req) },
    ),
  deleteCombo: (id: string) =>
    request<void>(`/api/combos/${encodeURIComponent(id)}`, { method: "DELETE" }),
  // No-code pieces. Every segment stands on a recording or on stated facts,
  // so the catalog is the no-code subset and the recordables list is served
  // rather than hard-coded — the page must not carry its own copy of either.
  noCodePieces: () => request<NoCodeSummary[]>("/api/nocode"),
  noCodePiece: (id: string) => request<NoCodeDetail>(`/api/nocode/${encodeURIComponent(id)}`),
  noCodeCatalog: () => request<NoCodeCatalogEntry[]>("/api/nocode/catalog"),
  noCodeRecordables: () => request<Recordable[]>("/api/nocode/recordables"),
  noCodeTakes: () => request<NoCodeTakes>("/api/nocode/takes"),
  createNoCodePiece: (req: CreateNoCodeRequest) =>
    request<NoCodeSummary>("/api/nocode", { method: "POST", body: JSON.stringify(req) }),
  runNoCodePiece: (id: string) =>
    request<{ run_id: string }>(`/api/nocode/${encodeURIComponent(id)}/run`, { method: "POST" }),

  snippets: () => request<SnippetSummary[]>("/api/snippets"),
  snippet: (id: string) => request<SnippetDetail>(`/api/snippets/${encodeURIComponent(id)}`),
  createSnippet: (req: CreateSnippetRequest) =>
    request<CreateSnippetResponse>("/api/snippets", {
      method: "POST",
      body: JSON.stringify(req),
    }),
  deleteSnippet: (id: string) =>
    request<void>(`/api/snippets/${encodeURIComponent(id)}`, { method: "DELETE" }),
};

/** The synthetic course snippets live in; their lesson routes use this slug. */
export const SNIPPETS_COURSE = "snippets";

export function artifactUrl(course: string, fullLessonId: string, path: string): string {
  return `/artifacts/${encodeURIComponent(course)}/${encodeURIComponent(fullLessonId)}/${path}`;
}
