import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { api, type RunRequest, type RunStatus } from "../api/client";
import { SSEClient, type SSEStatus, type StudioEvent } from "../lib/sse";

export interface LogLine {
  seq: number;
  stage?: string;
  line: string;
}

/** Live per-stage status overlay from SSE, keyed "course|lesson|stage". */
export type LiveStageStatus = "running" | "done" | "failed" | "skipped";

interface RunContextValue {
  run: RunStatus;
  connection: SSEStatus;
  logs: LogLine[];
  liveStages: Record<string, LiveStageStatus>;
  /** Bumped when a run finishes/fails so views can refetch project state. */
  refreshTick: number;
  lastError: string | null;
  startRun: (req: RunRequest) => Promise<void>;
  cancelRun: () => Promise<void>;
  subscribe: (handler: (e: StudioEvent) => void) => () => void;
}

const RunContext = createContext<RunContextValue | null>(null);

const LOG_LIMIT = 500;

export function stageKey(course: string, lesson: string, stage: string): string {
  return `${course}|${lesson}|${stage}`;
}

/** The backend reports lesson ids sometimes short ("02"), sometimes full. */
export function lessonMatches(a: string | undefined, b: string): boolean {
  if (!a) return false;
  return a === b || a.startsWith(`${b}-`) || b.startsWith(`${a}-`);
}

export function RunProvider({ children }: { children: ReactNode }) {
  const [run, setRun] = useState<RunStatus>({ running: false });
  const [connection, setConnection] = useState<SSEStatus>("connecting");
  const [logs, setLogs] = useState<LogLine[]>([]);
  const [liveStages, setLiveStages] = useState<Record<string, LiveStageStatus>>({});
  const [refreshTick, setRefreshTick] = useState(0);
  const [lastError, setLastError] = useState<string | null>(null);
  const subscribers = useRef(new Set<(e: StudioEvent) => void>());

  useEffect(() => {
    let alive = true;
    api
      .runStatus()
      .then((s) => alive && setRun(s))
      .catch(() => {});

    const client = new SSEClient("/api/events");
    client.onStatus((s) => alive && setConnection(s));
    client.on("*", (e) => {
      if (!alive) return;
      for (const fn of subscribers.current) fn(e);
      switch (e.type) {
        case "run-started":
          setRun({
            running: true,
            run_id: e.run_id,
            course: e.course,
            lesson: e.lesson,
            stage: e.stage,
          });
          setLiveStages({});
          setLastError(null);
          setLogs([]);
          break;
        case "stage-started":
          setRun((r) => ({ ...r, running: true, stage: e.stage }));
          if (e.course && e.lesson && e.stage) {
            const k = stageKey(e.course, e.lesson, e.stage);
            setLiveStages((m) => ({ ...m, [k]: "running" }));
          }
          break;
        case "stage-finished":
        case "stage-skipped":
        case "stage-failed":
          if (e.course && e.lesson && e.stage) {
            const k = stageKey(e.course, e.lesson, e.stage);
            const status: LiveStageStatus =
              e.type === "stage-failed"
                ? "failed"
                : e.type === "stage-skipped"
                  ? "skipped"
                  : "done";
            setLiveStages((m) => ({ ...m, [k]: status }));
          }
          if (e.type === "stage-failed" && e.error) setLastError(e.error);
          break;
        case "log":
          if (e.line !== undefined) {
            setLogs((prev) => {
              const next = [...prev, { seq: e.seq, stage: e.stage, line: e.line ?? "" }];
              return next.length > LOG_LIMIT ? next.slice(next.length - LOG_LIMIT) : next;
            });
          }
          break;
        case "run-finished":
        case "run-failed":
          setRun({ running: false });
          if (e.type === "run-failed" && e.error) setLastError(e.error);
          setRefreshTick((t) => t + 1);
          break;
        case "quota":
          setRefreshTick((t) => t + 1);
          break;
      }
    });
    client.start();
    return () => {
      alive = false;
      client.close();
    };
  }, []);

  const startRun = useCallback(async (req: RunRequest) => {
    setLastError(null);
    try {
      const { run_id } = await api.startRun(req);
      setRun({
        running: true,
        run_id,
        course: req.course,
        lesson: req.lesson,
        stage: req.stage,
      });
    } catch (err) {
      setLastError(err instanceof Error ? err.message : String(err));
      throw err;
    }
  }, []);

  const cancelRun = useCallback(async () => {
    try {
      await api.cancelRun();
    } catch (err) {
      setLastError(err instanceof Error ? err.message : String(err));
    }
  }, []);

  const subscribe = useCallback((handler: (e: StudioEvent) => void) => {
    subscribers.current.add(handler);
    return () => {
      subscribers.current.delete(handler);
    };
  }, []);

  const value = useMemo(
    () => ({
      run,
      connection,
      logs,
      liveStages,
      refreshTick,
      lastError,
      startRun,
      cancelRun,
      subscribe,
    }),
    [run, connection, logs, liveStages, refreshTick, lastError, startRun, cancelRun, subscribe],
  );

  return <RunContext.Provider value={value}>{children}</RunContext.Provider>;
}

export function useRun(): RunContextValue {
  const ctx = useContext(RunContext);
  if (!ctx) throw new Error("useRun outside RunProvider");
  return ctx;
}
