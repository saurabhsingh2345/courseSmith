import { formatBytes } from "../lib/format";

/**
 * Anchor styled as a button that downloads an artifact. Uses the `download`
 * attribute so the browser saves rather than navigates; the studio serves
 * artifacts with Content-Disposition, and same-origin `download` handles the
 * rest. Optional `size` renders a byte badge.
 */
export function DownloadButton({
  url,
  filename,
  label = "Download",
  size,
}: {
  url: string;
  filename?: string;
  label?: string;
  size?: number;
}) {
  return (
    <a
      href={url}
      download={filename ?? true}
      className="inline-flex items-center gap-1.5 rounded border border-ink-700 bg-ink-800 px-2.5 py-1 text-[12px] text-ink-200 hover:bg-ink-700"
    >
      {label}
      {typeof size === "number" && (
        <span className="text-[10px] text-ink-500">{formatBytes(size)}</span>
      )}
    </a>
  );
}
