import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

export interface ShortcutHint {
  keys: string;
  label: string;
}

interface ShortcutContextValue {
  hints: ShortcutHint[];
  setHints: (hints: ShortcutHint[]) => void;
  overlayOpen: boolean;
  setOverlayOpen: (open: boolean) => void;
}

const ShortcutContext = createContext<ShortcutContextValue | null>(null);

export function ShortcutProvider({ children }: { children: ReactNode }) {
  const [hints, setHints] = useState<ShortcutHint[]>([]);
  const [overlayOpen, setOverlayOpen] = useState(false);
  const value = useMemo(
    () => ({ hints, setHints, overlayOpen, setOverlayOpen }),
    [hints, overlayOpen],
  );
  return <ShortcutContext.Provider value={value}>{children}</ShortcutContext.Provider>;
}

export function useShortcutContext(): ShortcutContextValue {
  const ctx = useContext(ShortcutContext);
  if (!ctx) throw new Error("useShortcutContext outside provider");
  return ctx;
}

export function isTypingTarget(el: EventTarget | null): boolean {
  if (!(el instanceof HTMLElement)) return false;
  return (
    el.tagName === "INPUT" ||
    el.tagName === "TEXTAREA" ||
    el.tagName === "SELECT" ||
    el.isContentEditable
  );
}

/**
 * Register this screen's footer hints and its key handlers.
 * Handlers are skipped while typing in a form field (except Escape).
 */
export function useScreenShortcuts(
  hints: ShortcutHint[],
  onKey?: (e: KeyboardEvent) => void,
): void {
  const { setHints } = useShortcutContext();

  useEffect(() => {
    setHints(hints);
    return () => setHints([]);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(hints)]);

  useEffect(() => {
    if (!onKey) return;
    const handler = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (isTypingTarget(e.target) && e.key !== "Escape") return;
      onKey(e);
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onKey]);
}
