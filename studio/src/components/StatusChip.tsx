export type ChipStatus =
  | "done"
  | "stale"
  | "pending"
  | "running"
  | "failed"
  | "skipped";

const styles: Record<ChipStatus, string> = {
  done: "bg-emerald-500/15 text-emerald-400 border-emerald-500/30",
  stale: "bg-amber-500/15 text-amber-400 border-amber-500/30",
  pending: "bg-ink-700/40 text-ink-400 border-ink-600",
  running: "bg-sky-500/20 text-sky-300 border-sky-500/40 animate-pulse-fast",
  failed: "bg-red-500/15 text-red-400 border-red-500/40",
  skipped: "bg-ink-700/40 text-ink-300 border-ink-600",
};

export const chipDot: Record<ChipStatus, string> = {
  done: "bg-emerald-400",
  stale: "bg-amber-400",
  pending: "bg-ink-500",
  running: "bg-sky-400 animate-pulse-fast",
  failed: "bg-red-400",
  skipped: "bg-ink-400",
};

export function StatusChip({
  status,
  label,
  title,
  onClick,
}: {
  status: ChipStatus;
  label?: string;
  title?: string;
  onClick?: () => void;
}) {
  const Tag = onClick ? "button" : "span";
  return (
    <Tag
      title={title}
      onClick={onClick}
      className={`inline-flex items-center gap-1 rounded border px-1.5 py-0.5 text-[11px] font-medium leading-none ${styles[status]} ${
        onClick ? "cursor-pointer hover:brightness-125" : ""
      }`}
    >
      <span className={`h-1.5 w-1.5 rounded-full ${chipDot[status]}`} />
      {label ?? status}
    </Tag>
  );
}
