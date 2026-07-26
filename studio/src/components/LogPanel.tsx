import { useEffect, useRef } from "react";
import { useRun } from "../state/RunContext";

export function LogPanel({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { logs } = useRun();
  const endRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (open) endRef.current?.scrollIntoView({ block: "end" });
  }, [logs, open]);

  if (!open) return null;

  return (
    <div className="flex h-56 flex-col border-t border-ink-800 bg-ink-950">
      <div className="flex items-center justify-between border-b border-ink-800 px-3 py-1 text-[11px] text-ink-400">
        <span>
          logs <span className="text-ink-600">({logs.length})</span>
        </span>
        <button onClick={onClose} className="hover:text-ink-200">
          close ✕
        </button>
      </div>
      <div className="min-h-0 flex-1 overflow-auto px-3 py-1 font-mono text-[11px] leading-relaxed">
        {logs.length === 0 ? (
          <div className="text-ink-600">no output yet</div>
        ) : (
          logs.map((l) => (
            <div key={l.seq} className="whitespace-pre-wrap break-words text-ink-300">
              {l.stage && <span className="text-ink-600">[{l.stage}] </span>}
              {l.line}
            </div>
          ))
        )}
        <div ref={endRef} />
      </div>
    </div>
  );
}
