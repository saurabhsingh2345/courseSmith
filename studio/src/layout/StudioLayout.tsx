// The studio shell: sidebar, header, scrolling content, log panel, hint bar.
//
// Pulled out of App.tsx, which had grown into the layout, the keyboard handler
// and the route table at once. The routes are the part that changes weekly; the
// shell is the part that has to stay still.
//
// One rail, two presentations. Above `sm` it is a persistent column that can
// collapse to icons; below, it is an off-canvas drawer over a scrim, because a
// 380px viewport has no column to spare. Both render the same `SideNav`, so
// there is no second copy of the nav to keep in step.

import { useEffect, type ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";
import { Menu, X } from "lucide-react";
import { RunBar } from "../components/RunBar";
import { LogPanel } from "../components/LogPanel";
import { ShortcutOverlay } from "../components/ShortcutOverlay";
import { NavCollapseButton, SideNav } from "../components/SideNav";
import { ThemeToggle } from "../components/ThemeToggle";
import { Button } from "../components/base/Button";
import { cn } from "../lib/cn";
import { isTypingTarget, useShortcutContext } from "../state/ShortcutContext";
import { studio, useStudioStore } from "../store/studioStore";

function Wordmark() {
  return (
    <Link
      to="/"
      className={cn(
        "flex items-center gap-1.5 rounded-[var(--radius-md)] font-semibold text-fg",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2 focus-visible:ring-offset-bg",
      )}
    >
      coursesmith <span className="font-normal text-muted">studio</span>
    </Link>
  );
}

export function StudioLayout({
  children,
  logOpen,
  onCloseLog,
}: {
  children: ReactNode;
  logOpen: boolean;
  onCloseLog: () => void;
}) {
  const { hints, overlayOpen, setOverlayOpen } = useShortcutContext();
  const collapsed = useStudioStore((s) => s.navCollapsed);
  const mobileOpen = useStudioStore((s) => s.mobileNavOpen);
  const { pathname } = useLocation();

  // A drawer that survives navigation covers the page you just asked for.
  useEffect(() => {
    studio.setMobileNav(false);
  }, [pathname]);

  const footerHints = [...hints, { keys: "L", label: "logs" }, { keys: "?", label: "shortcuts" }];

  return (
    <div className="flex h-full flex-col bg-bg text-fg">
      <div className="flex min-h-0 flex-1">
        {/* Desktop rail. */}
        <aside
          className={cn(
            "hidden shrink-0 flex-col border-r border-border bg-surface sm:flex",
            // Written as a bracketed property because an arbitrary easing value
            // in the shorthand form is ambiguous to Tailwind and warns at build.
            "transition-[width] duration-[var(--motion-normal)] [transition-timing-function:var(--ease-subtle)]",
            collapsed ? "w-16" : "w-56",
          )}
        >
          <div
            className={cn(
              "flex min-h-14 items-center border-b border-border px-4",
              collapsed && "justify-center px-0",
            )}
          >
            {collapsed ? (
              <Link
                to="/"
                aria-label="coursesmith studio"
                className="rounded-[var(--radius-md)] font-semibold text-brand focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand"
              >
                cs
              </Link>
            ) : (
              <Wordmark />
            )}
          </div>
          <div className="min-h-0 flex-1">
            <SideNav collapsed={collapsed} />
          </div>
          <NavCollapseButton />
        </aside>

        {/* Mobile drawer. Mounted only while open so its links stay out of the
            tab order behind the scrim. */}
        {mobileOpen && (
          <div className="fixed inset-0 z-50 sm:hidden">
            {/* The scrim. aria-hidden and not a button: closing is already
                reachable from the X and from Escape, and a second control with
                the same name is one more thing between a screen-reader user and
                the nav they opened. */}
            <div
              aria-hidden
              onClick={() => studio.setMobileNav(false)}
              className="absolute inset-0 bg-black/60 animate-in fade-in-0"
            />
            <div className="absolute inset-y-0 left-0 flex w-64 flex-col border-r border-border bg-surface animate-in slide-in-from-left duration-[var(--motion-normal)]">
              <div className="flex min-h-14 items-center justify-between border-b border-border px-4">
                <Wordmark />
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-11"
                  aria-label="Close navigation"
                  onClick={() => studio.setMobileNav(false)}
                >
                  <X className="size-4" aria-hidden />
                </Button>
              </div>
              <div className="min-h-0 flex-1">
                <SideNav />
              </div>
            </div>
          </div>
        )}

        <div className="flex min-w-0 flex-1 flex-col">
          <header className="flex min-h-14 items-center gap-3 border-b border-border bg-surface px-3 sm:px-4">
            <Button
              variant="ghost"
              size="icon"
              className="size-11 sm:hidden"
              aria-label="Open navigation"
              aria-expanded={mobileOpen}
              onClick={() => studio.setMobileNav(true)}
            >
              <Menu className="size-4" aria-hidden />
            </Button>
            <div className="sm:hidden">
              <Wordmark />
            </div>
            <div className="ml-auto flex items-center gap-2">
              <RunBar />
              <ThemeToggle />
            </div>
          </header>

          <main id="main" className="min-h-0 flex-1 overflow-auto">
            {children}
          </main>
        </div>
      </div>

      <LogPanel open={logOpen} onClose={onCloseLog} />

      {/* Hints are a desktop affordance — they name keys a phone has none of. */}
      <footer className="hidden flex-wrap items-center gap-3 border-t border-border bg-surface px-4 py-1.5 text-[11px] text-muted sm:flex">
        {footerHints.map((h) => (
          <span key={h.keys + h.label} className="flex items-center gap-1">
            <kbd>{h.keys}</kbd> {h.label}
          </span>
        ))}
      </footer>

      {overlayOpen && <ShortcutOverlay hints={hints} onClose={() => setOverlayOpen(false)} />}
    </div>
  );
}

/** The shell's keyboard shortcuts, kept beside the shell they drive. */
export function useShellShortcuts({
  overlayOpen,
  setOverlayOpen,
  toggleLog,
}: {
  overlayOpen: boolean;
  setOverlayOpen: (v: boolean) => void;
  toggleLog: () => void;
}) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (e.key === "Escape") {
        if (overlayOpen) setOverlayOpen(false);
        studio.setMobileNav(false);
        return;
      }
      if (isTypingTarget(e.target)) return;
      if (e.key === "?") {
        e.preventDefault();
        setOverlayOpen(!overlayOpen);
      } else if (e.key === "l" || e.key === "L") {
        toggleLog();
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [overlayOpen, setOverlayOpen, toggleLog]);
}
