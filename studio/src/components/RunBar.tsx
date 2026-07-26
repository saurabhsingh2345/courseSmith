import { useRun } from "../state/RunContext";
import type { SSEStatus } from "../lib/sse";
import { Spinner } from "./Spinner";

const connLabel: Record<SSEStatus, string> = {
  connecting: "connecting",
  open: "live",
  reconnecting: "reconnecting",
  closed: "offline",
};

const connDot: Record<SSEStatus, string> = {
  connecting: "bg-amber-400",
  open: "bg-emerald-400",
  reconnecting: "bg-amber-400 animate-pulse-fast",
  closed: "bg-red-400",
};

export function RunBar() {
  const { run, connection, cancelRun, lastError } = useRun();
  return (
    <div className="flex items-center gap-3 text-[12px]">
      {lastError && (
        <span className="max-w-[280px] truncate text-red-400" title={lastError}>
          {lastError}
        </span>
      )}
      {run.running ? (
        <span className="flex items-center gap-2 text-sky-300">
          <Spinner />
          <span className="text-ink-300">
            {run.course}/{run.lesson}
            {run.stage ? ` · ${run.stage}` : ""}
          </span>
          <button
            onClick={() => void cancelRun()}
            className="rounded border border-red-500/40 bg-red-500/10 px-2 py-0.5 text-red-300 hover:bg-red-500/20"
          >
            Cancel
          </button>
        </span>
      ) : (
        <span className="text-ink-500">idle</span>
      )}
      <span className="flex items-center gap-1 text-ink-500" title={`events: ${connection}`}>
        <span className={`h-1.5 w-1.5 rounded-full ${connDot[connection]}`} />
        {connLabel[connection]}
      </span>
    </div>
  );
}
