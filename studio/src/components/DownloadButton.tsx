import { formatBytes } from "../lib/format";

/**
 * Anchor styled as a button that downloads an artifact.
 *
 * `filename` should be the artifact's `download_name` from the API, never its
 * URL's last segment: on disk every lesson's video is `final.mp4`. The server
 * sends the same name as Content-Disposition and browsers prefer that over
 * this attribute, so passing it here is belt and braces — but it is also what
 * makes the intent readable at the call site. Optional `size` renders a byte
 * badge.
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
