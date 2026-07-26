// Client for the coursesmith-tutor adaptive-learning service (workstream D).
// Unlike api/client.ts (same-origin /api on the studio backend), the tutor runs
// as a separate service — default http://localhost:8765, overridable with
// VITE_TUTOR_URL — so this speaks absolute URLs and relies on the service's
// CORS headers. Shapes mirror internal/adaptive.

export interface BKTParams {
  p_init: number;
  p_learn: number;
  p_slip: number;
  p_guess: number;
}

export interface BKTEstimate {
  p_known: number;
  p_next_correct: number;
  mastered: boolean;
  difficulty_hint: "easier" | "medium" | "harder";
  recommendation: string;
}

export type FSRSRating = "again" | "hard" | "good" | "easy";

export interface FSRSResult {
  stability: number;
  difficulty: number;
  interval_days: number;
  note: string;
}

export interface IRTObservation {
  question_id: string;
  correct: boolean;
}

export interface IRTItem {
  question_id: string;
  difficulty: number;
  discrimination: number;
}

export interface IRTResult {
  calibrated: IRTItem[];
  note: string;
}

export class TutorError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = "TutorError";
    this.status = status;
  }
}

export const TUTOR_URL: string =
  import.meta.env.VITE_TUTOR_URL ?? "http://localhost:8765";

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(url, {
      ...init,
      headers: {
        ...(init?.body !== undefined ? { "Content-Type": "application/json" } : {}),
        ...init?.headers,
      },
    });
  } catch (e) {
    // Network error / service down / CORS — give a friendly, actionable message.
    throw new TutorError(
      0,
      `tutor service unreachable — is coursesmith-tutor running? (${e instanceof Error ? e.message : String(e)})`,
    );
  }
  if (!res.ok) {
    let message = `HTTP ${res.status}`;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      /* non-JSON error body */
    }
    throw new TutorError(res.status, message);
  }
  return (await res.json()) as T;
}

/** A thin, typed client bound to a tutor base URL. */
export function createTutorClient(baseURL: string = TUTOR_URL) {
  return {
    baseURL,
    health: () => request<{ status: string; service?: string }>(`${baseURL}/health`),

    /** Posterior mastery from a sequence of correct/incorrect answers. */
    estimateBKT: (corrects: boolean[], params?: BKTParams) =>
      request<BKTEstimate>(`${baseURL}/bkt/estimate`, {
        method: "POST",
        body: JSON.stringify({
          params,
          responses: corrects.map((correct) => ({ correct })),
        }),
      }),

    /** Next review interval for a rated review. */
    scheduleFSRS: (rating: FSRSRating, stability = 0, difficulty = 0) =>
      request<FSRSResult>(`${baseURL}/fsrs/schedule`, {
        method: "POST",
        body: JSON.stringify({ rating, stability, difficulty }),
      }),

    /** Per-item difficulty/discrimination (currently neutral-prior stubs). */
    calibrateIRT: (responses: IRTObservation[]) =>
      request<IRTResult>(`${baseURL}/irt/calibrate`, {
        method: "POST",
        body: JSON.stringify({ responses }),
      }),
  };
}

export type TutorClient = ReturnType<typeof createTutorClient>;
