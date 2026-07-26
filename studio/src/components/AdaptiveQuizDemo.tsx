import { useEffect, useState } from "react";
import { useTutorAPI } from "../hooks/useTutorAPI";
import type { BKTEstimate, FSRSRating, FSRSResult } from "../api/tutor";
import { Badge } from "./base/Badge";
import { Button } from "./base/Button";
import { Alert, AlertDescription, AlertTitle } from "./base/Alert";

const CONCEPT = "Python lists";
const RATINGS: { rating: FSRSRating; label: string }[] = [
  { rating: "again", label: "Again" },
  { rating: "hard", label: "Hard" },
  { rating: "good", label: "Good" },
  { rating: "easy", label: "Easy" },
];

function hintBadge(hint: BKTEstimate["difficulty_hint"]) {
  const variant = hint === "harder" ? "success" : hint === "easier" ? "warning" : "secondary";
  const label = hint === "harder" ? "Advance — harder" : hint === "easier" ? "Ease off — easier" : "Hold — medium";
  return <Badge variant={variant}>{label}</Badge>;
}

/** Mastery meter: a labelled bar that fills to p_known. */
function MasteryBar({ pct }: { pct: number }) {
  return (
    <div className="flex flex-col gap-1">
      <div className="bg-ink-800 h-3 w-full overflow-hidden rounded-full">
        <div
          className="h-full rounded-full transition-[width] duration-500"
          style={{ width: `${Math.round(pct * 100)}%`, background: "var(--color-brand)" }}
        />
      </div>
      <div className="text-ink-100 text-sm font-semibold">
        You&rsquo;ve mastered {Math.round(pct * 100)}% of this concept
      </div>
    </div>
  );
}

/**
 * AdaptiveQuizDemo (workstream D) drives the live coursesmith-tutor service:
 * simulate quiz answers for a concept and watch BKT mastery + the difficulty
 * recommendation update, then rate a review and see the FSRS next-review date.
 * The service is optional — if it is down, the panel explains how to start it.
 */
export function AdaptiveQuizDemo() {
  const { estimateMastery, scheduleReview, loading, error } = useTutorAPI();
  const [responses, setResponses] = useState<boolean[]>([true, true, false, true]);
  const [estimate, setEstimate] = useState<BKTEstimate | null>(null);
  const [schedule, setSchedule] = useState<FSRSResult | null>(null);

  // Re-estimate whenever the simulated answer sequence changes.
  useEffect(() => {
    let alive = true;
    estimateMastery(responses).then((e) => {
      if (alive && e) setEstimate(e);
    });
    return () => {
      alive = false;
    };
  }, [responses, estimateMastery]);

  const answer = (correct: boolean) => setResponses((r) => [...r, correct]);
  const reset = () => {
    setResponses([]);
    setSchedule(null);
  };

  const rate = async (rating: FSRSRating) => {
    const s = await scheduleReview(rating);
    if (s) setSchedule(s);
  };

  const nextReviewDate =
    schedule &&
    new Date(Date.now() + schedule.interval_days * 86_400_000).toLocaleDateString(undefined, {
      weekday: "short",
      month: "short",
      day: "numeric",
    });

  return (
    <div className="border-ink-700 bg-ink-900/40 flex flex-col gap-5 rounded-lg border p-5">
      {error ? (
        <Alert variant="info">
          <AlertTitle>Tutor service offline</AlertTitle>
          <AlertDescription>
            Showing the last estimate. Start it with{" "}
            <code className="font-mono text-[11px]">go run ./cmd/coursesmith-tutor</code> to go live.
          </AlertDescription>
        </Alert>
      ) : null}

      {/* BKT mastery */}
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <h3 className="text-ink-100 text-sm font-semibold">
            Mastery of <span className="text-brand">{CONCEPT}</span>{" "}
            <span className="text-ink-500 font-normal">(BKT)</span>
          </h3>
          <span className="text-ink-500 font-mono text-[10px]">
            {responses.length} response{responses.length === 1 ? "" : "s"}
            {loading ? " · …" : ""}
          </span>
        </div>

        {estimate ? (
          <>
            <MasteryBar pct={estimate.p_known} />
            <div className="flex flex-wrap items-center gap-3">
              {hintBadge(estimate.difficulty_hint)}
              {estimate.mastered ? <Badge variant="success">Mastered</Badge> : null}
              <span className="text-ink-500 text-xs">
                Next answer predicted correct: {Math.round(estimate.p_next_correct * 100)}%
              </span>
            </div>
            <p className="text-ink-300 text-xs">{estimate.recommendation}</p>
          </>
        ) : (
          <p className="text-ink-500 text-xs">Answer a question to estimate mastery.</p>
        )}

        <div className="flex flex-wrap gap-2">
          <Button size="sm" variant="primary" onClick={() => answer(true)}>
            Answered correctly
          </Button>
          <Button size="sm" variant="secondary" onClick={() => answer(false)}>
            Answered incorrectly
          </Button>
          <Button size="sm" variant="ghost" onClick={reset}>
            Reset
          </Button>
        </div>
      </div>

      {/* FSRS review scheduling */}
      <div className="border-ink-800 flex flex-col gap-3 border-t pt-4">
        <h3 className="text-ink-100 text-sm font-semibold">
          Spaced review <span className="text-ink-500 font-normal">(FSRS)</span>
        </h3>
        <p className="text-ink-500 text-xs">Rate how the review felt to schedule the next one:</p>
        <div className="flex flex-wrap gap-2">
          {RATINGS.map(({ rating, label }) => (
            <Button key={rating} size="sm" variant="outline" onClick={() => rate(rating)}>
              {label}
            </Button>
          ))}
        </div>
        {schedule ? (
          <div className="text-ink-300 text-sm">
            Next review in{" "}
            <span className="text-ink-100 font-semibold">
              {schedule.interval_days} day{schedule.interval_days === 1 ? "" : "s"}
            </span>{" "}
            <span className="text-ink-500">— {nextReviewDate}</span>
          </div>
        ) : null}
      </div>
    </div>
  );
}
