import type { ShortcutHint } from "../state/ShortcutContext";

const GLOBAL: ShortcutHint[] = [
  { keys: "?", label: "toggle this help" },
  { keys: "L", label: "toggle logs" },
  { keys: "Esc", label: "close overlays" },
];

function Row({ hint }: { hint: ShortcutHint }) {
  return (
    <li className="flex items-center justify-between gap-4">
      <span className="text-ink-300">{hint.label}</span>
      <kbd>{hint.keys}</kbd>
    </li>
  );
}

export function ShortcutOverlay({
  hints,
  onClose,
}: {
  hints: ShortcutHint[];
  onClose: () => void;
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md rounded-lg border border-ink-700 bg-ink-850 p-5 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-ink-100">Keyboard shortcuts</h2>
          <button onClick={onClose} className="text-ink-500 hover:text-ink-200">
            ✕
          </button>
        </div>
        {hints.length > 0 && (
          <div className="mb-4">
            <div className="mb-1 text-[11px] uppercase tracking-wide text-ink-500">This screen</div>
            <ul className="space-y-1">
              {hints.map((h) => (
                <Row key={h.keys + h.label} hint={h} />
              ))}
            </ul>
          </div>
        )}
        <div className="mb-1 text-[11px] uppercase tracking-wide text-ink-500">Global</div>
        <ul className="space-y-1">
          {GLOBAL.map((h) => (
            <Row key={h.keys + h.label} hint={h} />
          ))}
        </ul>
      </div>
    </div>
  );
}
