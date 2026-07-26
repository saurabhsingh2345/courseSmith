import { useCallback, useMemo, useState } from "react";
import {
  createTutorClient,
  TUTOR_URL,
  type BKTEstimate,
  type BKTParams,
  type FSRSRating,
  type FSRSResult,
} from "../api/tutor";

/**
 * React wrapper around the coursesmith-tutor client. Each call updates shared
 * `loading`/`error` state and returns the result (or null on failure), so
 * components can render adaptive insights without their own try/catch. The
 * service is optional: when it is down, calls resolve to null and `error`
 * carries a friendly message rather than throwing.
 */
export function useTutorAPI(baseURL: string = TUTOR_URL) {
  const client = useMemo(() => createTutorClient(baseURL), [baseURL]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const run = useCallback(async <T,>(fn: () => Promise<T>): Promise<T | null> => {
    setLoading(true);
    setError(null);
    try {
      return await fn();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  const estimateMastery = useCallback(
    (corrects: boolean[], params?: BKTParams): Promise<BKTEstimate | null> =>
      run(() => client.estimateBKT(corrects, params)),
    [client, run],
  );

  const scheduleReview = useCallback(
    (rating: FSRSRating, stability?: number, difficulty?: number): Promise<FSRSResult | null> =>
      run(() => client.scheduleFSRS(rating, stability, difficulty)),
    [client, run],
  );

  const checkHealth = useCallback(
    (): Promise<{ status: string } | null> => run(() => client.health()),
    [client, run],
  );

  return { estimateMastery, scheduleReview, checkHealth, loading, error };
}
