/** Pure formatting helpers. No DOM, no side effects. */

/** Milliseconds -> "M:SS" (e.g. 61000 -> "1:01"). Hours roll into minutes. */
export function formatMs(ms: number): string {
  if (!Number.isFinite(ms) || ms < 0) ms = 0;
  const totalSec = Math.floor(ms / 1000);
  const min = Math.floor(totalSec / 60);
  const sec = totalSec % 60;
  return `${min}:${String(sec).padStart(2, "0")}`;
}

/** Seconds -> "M:SS". */
export function formatSec(sec: number): string {
  return formatMs(sec * 1000);
}

/** Bytes -> human string with binary-ish 1024 steps ("1.5 MB"). */
export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) bytes = 0;
  if (bytes < 1024) return `${Math.round(bytes)} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let value = bytes;
  let unit = "B";
  for (const u of units) {
    if (value < 1024) break;
    value /= 1024;
    unit = u;
  }
  const rounded = value >= 100 ? Math.round(value).toString() : value.toFixed(1);
  return `${rounded.replace(/\.0$/, "")} ${unit}`;
}

/** USD amount -> "$1.23", small amounts keep 4 decimals ("$0.0717"). */
export function formatUsd(amount: number): string {
  if (!Number.isFinite(amount)) amount = 0;
  const abs = Math.abs(amount);
  const sign = amount < 0 ? "-" : "";
  if (abs > 0 && abs < 1) return `${sign}$${abs.toFixed(4)}`;
  return `${sign}$${abs.toFixed(2)}`;
}

/** 0..1 ratio -> "12%" (rounded). */
export function formatPct(ratio: number): string {
  if (!Number.isFinite(ratio)) ratio = 0;
  return `${Math.round(ratio * 100)}%`;
}

/** Compact integer with thousands separators. */
export function formatInt(n: number): string {
  if (!Number.isFinite(n)) n = 0;
  return Math.round(n).toLocaleString("en-US");
}
