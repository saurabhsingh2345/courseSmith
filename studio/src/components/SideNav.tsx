// The studio's primary navigation.
//
// It replaced a row of nine text links in the header, which had two problems.
// It could not grow — a tenth surface had nowhere to go, and the row already
// wrapped under the run bar at laptop widths. And it read as nine equal things,
// when the studio is really two groups: the surfaces you *make* something on
// and the ones you *inspect* it from.
//
// Collapsed, it is an icon rail: same targets, same order, labels in tooltips.
// The collapse is persisted because it is a workspace preference, not a mode —
// somebody who works on a 13" screen wants it collapsed every morning.

import { Link, useLocation } from "react-router-dom";
import {
  BarChart3,
  Camera,
  BookOpen,
  ChevronLeft,
  FlaskConical,
  GraduationCap,
  Layers,
  Library,
  PlayCircle,
  Film,
  Receipt,
  Scissors,
  Sparkles,
  Wand2,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "./base/Tooltip";
import { cn } from "../lib/cn";
import { studio, useStudioStore } from "../store/studioStore";

interface NavEntry {
  to: string;
  label: string;
  icon: LucideIcon;
  /** Match nested routes too — /courses stays active on /c/python-basics. */
  match?: (path: string) => boolean;
}

interface NavGroup {
  heading: string;
  entries: NavEntry[];
}

export const NAV_GROUPS: NavGroup[] = [
  {
    heading: "Create",
    entries: [
      { to: "/", label: "Compose", icon: Sparkles },
      { to: "/snippets", label: "Snippets", icon: Scissors },
      { to: "/replica", label: "Replica", icon: Wand2 },
      { to: "/foundations", label: "Foundations", icon: GraduationCap },
      { to: "/reels", label: "Reels", icon: Film },
      { to: "/nocode", label: "No-code", icon: Camera },
      {
        to: "/courses",
        label: "Courses",
        icon: BookOpen,
        match: (p) => p.startsWith("/courses") || p.startsWith("/c/"),
      },
      { to: "/generation", label: "Generation", icon: PlayCircle },
    ],
  },
  {
    heading: "Configure",
    entries: [
      { to: "/templates", label: "Templates", icon: Layers },
      { to: "/adaptive", label: "Adaptive", icon: FlaskConical },
      { to: "/library", label: "Library", icon: Library },
    ],
  },
  {
    heading: "Inspect",
    entries: [
      { to: "/ledger", label: "Ledger", icon: Receipt },
      { to: "/showcase", label: "Showcase", icon: BarChart3 },
    ],
  },
];

function NavItem({ entry, collapsed }: { entry: NavEntry; collapsed: boolean }) {
  const { icon: Icon, label, to, match } = entry;
  const { pathname } = useLocation();

  // Deliberately not NavLink's own `isActive`. A course lives at /c/:slug, not
  // under /courses, so NavLink can never match the two — and a nav that
  // de-highlights Courses while you are reading a course tells you that you
  // have left the section you are plainly still in. `match` is that mapping,
  // and it has to be applied here or it is decoration.
  const isActive = match ? match(pathname) : pathname === to;

  const link = (
    <Link
      to={to}
      onClick={() => studio.setMobileNav(false)}
      aria-current={isActive ? "page" : undefined}
      aria-label={collapsed ? label : undefined}
      className={cn(
        // 44px min target: this is the one control that has to work on a phone.
        "group relative flex min-h-11 items-center gap-3 rounded-[var(--radius-md)] px-3 text-sm",
        "transition-colors duration-[var(--motion-fast)]",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2 focus-visible:ring-offset-bg",
        collapsed && "justify-center px-0",
        isActive ? "bg-ink-800/60 font-medium text-fg" : "text-muted hover:bg-ink-800/40 hover:text-fg",
      )}
    >
      {/* The active rail. Rendered always and scaled, so switching pages
          animates the bar rather than mounting a new element. */}
      <span
        aria-hidden
        className={cn(
          "absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-r bg-brand",
          "origin-center transition-transform duration-[var(--motion-fast)]",
          isActive ? "scale-y-100" : "scale-y-0",
        )}
      />
      <Icon className="size-4 shrink-0" aria-hidden />
      {!collapsed && <span className="truncate">{label}</span>}
    </Link>
  );

  if (!collapsed) return link;
  return (
    <Tooltip>
      <TooltipTrigger asChild>{link}</TooltipTrigger>
      <TooltipContent side="right">{label}</TooltipContent>
    </Tooltip>
  );
}

/**
 * The nav body. Rendered twice by the layout — once as the persistent desktop
 * rail and once inside the mobile drawer — so it takes `collapsed` rather than
 * reading it, since the drawer is never collapsed however the rail is set.
 */
export function SideNav({ collapsed = false }: { collapsed?: boolean }) {
  return (
    <TooltipProvider delayDuration={200}>
      <nav
        aria-label="Studio sections"
        className="flex h-full flex-col gap-6 overflow-y-auto px-3 py-4"
      >
        {NAV_GROUPS.map((group) => (
          <div key={group.heading} className="flex flex-col gap-1">
            {collapsed ? (
              <div aria-hidden className="mx-auto mb-1 h-px w-6 bg-border" />
            ) : (
              <h2 className="px-3 pb-1 text-[10px] font-semibold uppercase tracking-widest text-muted">
                {group.heading}
              </h2>
            )}
            {group.entries.map((entry) => (
              <NavItem key={entry.to} entry={entry} collapsed={collapsed} />
            ))}
          </div>
        ))}
      </nav>
    </TooltipProvider>
  );
}

/** The collapse control. Desktop only — the drawer closes by other means. */
export function NavCollapseButton() {
  const collapsed = useStudioStore((s) => s.navCollapsed);
  return (
    <button
      type="button"
      onClick={studio.toggleNav}
      aria-label={collapsed ? "Expand navigation" : "Collapse navigation"}
      aria-pressed={collapsed}
      className={cn(
        "flex min-h-11 w-full items-center gap-2 border-t border-border px-4 text-xs text-muted",
        "transition-colors duration-[var(--motion-fast)] hover:text-fg",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-brand",
        collapsed && "justify-center px-0",
      )}
    >
      <ChevronLeft
        className={cn(
          "size-4 shrink-0 transition-transform duration-[var(--motion-normal)]",
          collapsed && "rotate-180",
        )}
        aria-hidden
      />
      {!collapsed && <span>Collapse</span>}
    </button>
  );
}
