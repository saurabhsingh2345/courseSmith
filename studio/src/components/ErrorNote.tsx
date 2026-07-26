export function ErrorNote({ error, onRetry }: { error: string; onRetry?: () => void }) {
  return (
    <div className="mb-4 flex items-center justify-between gap-3 rounded border border-red-500/40 bg-red-500/10 px-3 py-2 text-red-300">
      <span className="min-w-0 break-words">{error}</span>
      {onRetry && (
        <button
          onClick={onRetry}
          className="shrink-0 rounded border border-red-500/40 px-2 py-0.5 text-red-200 hover:bg-red-500/20"
        >
          Retry
        </button>
      )}
    </div>
  );
}
