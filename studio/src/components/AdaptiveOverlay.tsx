import { useEffect, useMemo, useState } from "react";
import type { Question, Quiz } from "../api/client";
import { useTutorAPI } from "../hooks/useTutorAPI";
import type { BKTEstimate } from "../api/tutor";
import { Alert, AlertDescription, AlertTitle, Badge, Skeleton } from "./base";

// AdaptiveOverlay renders per-dimension mastery from the real coursesmith-tutor
// BKT engine (workstream D). It groups the lesson's quiz by question *type*
// (recall / application / debugging / prediction) — the cognitive dimensions
// present on every Question — and, for each, asks the tutor service to estimate
// mastery for a *simulated* learner. There is no student persistence yet, so the
// response patterns below are seeded demo profiles, not real telemetry; the math,
// shapes, and service round-trip are real.
//
// The service is optional. When it is unreachable every estimate resolves to
// null and we show a friendly "not running" note instead of throwing.

type QuestionType = Question["type"];

// Canonical order + a seeded response history per dimension. Each array is a
// plausible learner's correct/incorrect streak; the BKT recurrence turns it into
// a mastery estimate, so different dimensions land at visibly different mastery.
const DIMENSIONS: { type: QuestionType; label: string; seed: boolean[] }[] = [
  { type: "recall", label: "Recall", seed: [true, true, true, false, true] },
  { type: "application", label: "Application", seed: [true, false, true, true] },
  { type: "debugging", label: "Debugging", seed: [false, true, false, true, true] },
  { type: "prediction", label: "Prediction", seed: [true, false, true] },
];

function masteryColor(p: number): { bar: string; text: string } {
  if (p > 0.7) return { bar: "bg-success", text: "text-success" };
  if (p > 0.4) return { bar: "bg-warning", text: "text-warning" };
  return { bar: "bg-error", text: "text-error" };
}

function difficultyVariant(hint: BKTEstimate["difficulty_hint"]) {
  return hint === "harder" ? "success" : hint === "easier" ? "warning" : "secondary";
}

type Status = "loading" | "ready" | "down";

export function AdaptiveOverlay({ quiz }: { quiz?: Quiz | null }) {
  const { estimateMastery } = useTutorAPI();

  // Only show dimensions that actually appear in this lesson's quiz, with a count.
  const dimensions = useMemo(() => {
    const counts = new Map<QuestionType, number>();
    for (const q of quiz?.questions ?? []) {
      counts.set(q.type, (counts.get(q.type) ?? 0) + 1);
    }
    return DIMENSIONS.filter((d) => counts.has(d.type)).map((d) => ({
      ...d,
      count: counts.get(d.type) ?? 0,
    }));
  }, [quiz]);

  const dimensionsKey = dimensions.map((d) => d.type).join(",");
  const [estimates, setEstimates] = useState<Record<string, BKTEstimate>>({});
  const [status, setStatus] = useState<Status>("loading");

  useEffect(() => {
    if (dimensions.length === 0) {
      setStatus("ready");
      return;
    }
    let alive = true;
    setStatus("loading");
    (async () => {
      const next: Record<string, BKTEstimate> = {};
      let anyOk = false;
      for (const dim of dimensions) {
        const est = await estimateMastery(dim.seed);
        if (!alive) return;
        if (est) {
          next[dim.type] = est;
          anyOk = true;
        }
      }
      if (!alive) return;
      setEstimates(next);
      setStatus(anyOk ? "ready" : "down");
    })();
    return () => {
      alive = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dimensionsKey, estimateMastery]);

  if (dimensions.length === 0) {
    return (
      <div className="text-sm text-ink-500">
        No quiz questions to model yet — run the quiz stage first.
      </div>
    );
  }

  if (status === "loading") {
    return (
      <div className="grid grid-cols-[repeat(auto-fit,minmax(280px,1fr))] gap-4">
        {dimensions.map((d) => (
          <Skeleton key={d.type} className="h-40 rounded-[var(--radius-md)]" />
        ))}
      </div>
    );
  }

  if (status === "down") {
    return (
      <Alert variant="warning">
        <AlertTitle>Tutor service not running</AlertTitle>
        <AlertDescription>
          Mastery estimates come from <code>coursesmith-tutor</code>. Start it with{" "}
          <code>go run ./cmd/coursesmith-tutor</code> (defaults to <code>:8765</code>), or set{" "}
          <code>VITE_TUTOR_URL</code>, then reopen this tab.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div>
      <p className="mb-3 text-[12px] text-ink-500">
        BKT mastery for a <span className="text-ink-300">simulated learner</span>, per question
        type. Live from <code>coursesmith-tutor</code>.
      </p>
      <div className="grid grid-cols-[repeat(auto-fit,minmax(280px,1fr))] gap-4">
        {dimensions.map((dim) => {
          const est = estimates[dim.type];
          if (!est) return null;
          const pct = Math.round(est.p_known * 100);
          const nextPct = Math.round(est.p_next_correct * 100);
          const color = masteryColor(est.p_known);
          return (
            <div
              key={dim.type}
              className="rounded-[var(--radius-md)] border border-ink-800 bg-ink-900 p-4"
            >
              <div className="mb-2 flex items-center justify-between">
                <h4 className="font-medium text-ink-100">
                  {dim.label}{" "}
                  <span className="text-[11px] font-normal text-ink-500">
                    · {dim.count} question{dim.count === 1 ? "" : "s"}
                  </span>
                </h4>
                {est.mastered && <Badge variant="success">mastered</Badge>}
              </div>

              <div className="mb-1 flex items-center justify-between text-[12px]">
                <span className="text-ink-400">Mastery</span>
                <span className={`font-semibold ${color.text}`}>{pct}%</span>
              </div>
              <div className="mb-3 h-2 w-full overflow-hidden rounded-full bg-ink-800">
                <div className={`h-full ${color.bar}`} style={{ width: `${pct}%` }} />
              </div>

              <div className="mb-2 flex items-center gap-2 text-[11px] text-ink-500">
                <span>P(next correct) {nextPct}%</span>
                <span>·</span>
                <Badge variant={difficultyVariant(est.difficulty_hint)}>{est.difficulty_hint}</Badge>
              </div>

              <p className="text-[12px] text-ink-400">{est.recommendation}</p>
            </div>
          );
        })}
      </div>
    </div>
  );
}
