import type { components } from "./schema";

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
};

export function artifactUrl(course: string, fullLessonId: string, path: string): string {
  return `/artifacts/${encodeURIComponent(course)}/${encodeURIComponent(fullLessonId)}/${path}`;
}
