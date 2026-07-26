import { useMemo } from "react";
import { useRun, lessonMatches, type LiveStageStatus } from "../state/RunContext";
import type { RunRequest } from "../api/client";

export interface GenerationView {
  run: ReturnType<typeof useRun>["run"];
  connection: ReturnType<typeof useRun>["connection"];
  logs: ReturnType<typeof useRun>["logs"];
  liveStages: Record<string, LiveStageStatus>;
  lastError: string | null;
  /** 0..100 across the given stage order for the currently running lesson. */
  percent: number;
  /** The stage currently reported running, if any. */
  activeStage: string | undefined;
  start: (req: RunRequest) => Promise<void>;
  cancel: () => Promise<void>;
}

/**
 * Page-level wrapper over RunContext that derives a progress percentage from the
 * live per-stage overlay. The pipeline runs one lesson at a time, so progress is
 * "stages done or skipped / total stages" for the running course+lesson.
 */
export function useGeneration(stageOrder: string[] | undefined): GenerationView {
  const { run, connection, logs, liveStages, lastError, startRun, cancelRun } = useRun();

  const percent = useMemo(() => {
    if (!stageOrder || stageOrder.length === 0 || !run.course || !run.lesson) return 0;
    let done = 0;
    for (const stage of stageOrder) {
      const status = liveStatusFor(liveStages, run.course, run.lesson, stage);
      if (status === "done" || status === "skipped") done++;
    }
    return Math.round((done / stageOrder.length) * 100);
  }, [stageOrder, liveStages, run.course, run.lesson]);

  return {
    run,
    connection,
    logs,
    liveStages,
    lastError,
    percent,
    activeStage: run.stage,
    start: startRun,
    cancel: cancelRun,
  };
}

// liveStatusFor tolerates the backend reporting lesson ids short ("02") or full
// ("02-variables"), mirroring LessonPage's matcher.
function liveStatusFor(
  live: Record<string, LiveStageStatus>,
  course: string,
  lesson: string,
  stage: string,
): LiveStageStatus | undefined {
  for (const [key, value] of Object.entries(live)) {
    const [c, l, s] = key.split("|");
    if (c === course && s === stage && lessonMatches(l, lesson)) return value;
  }
  return undefined;
}
