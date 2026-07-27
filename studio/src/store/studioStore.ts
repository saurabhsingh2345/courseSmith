// The studio's shell state: the things more than one page needs and no single
// page owns — the nav, the theme, and the course list the sidebar and the
// Courses page both read.
//
// It is a plain external store read through `useSyncExternalStore`, not a
// context, and the reason is the nav. Collapsing the sidebar changes one
// boolean; behind a context that re-renders every page under the provider,
// including a Remotion player mid-frame. Selector subscriptions mean a
// component re-renders only when the slice it selected actually changed.
//
// It is deliberately *not* a cache of everything. Per-page data still belongs
// to that page's `useFetch` — a store that mirrors every endpoint has to be
// invalidated from every mutation, and the second one anybody forgets is a page
// showing something the server no longer has. What lives here is state with no
// other home: UI that outlives a route, and the course list, which is read on
// nearly every screen and changes only through the two actions below.

import { useCallback, useSyncExternalStore } from "react";
import { api, type Course } from "../api/client";
import { applyTheme, preferredMode, type ThemeMode } from "../theme/applyTheme";

export type LoadStatus = "idle" | "loading" | "ready" | "error";

export interface StudioState {
  /** Sidebar collapsed to an icon rail. Persisted; desktop only. */
  navCollapsed: boolean;
  /** The mobile drawer. Never persisted — a reload should not open a modal. */
  mobileNavOpen: boolean;
  theme: ThemeMode;
  courses: Course[];
  coursesStatus: LoadStatus;
  coursesError: string | null;
}

const NAV_KEY = "cs-nav-collapsed";
const THEME_KEY = "cs-theme";

const readStoredBool = (key: string): boolean => {
  try {
    return localStorage.getItem(key) === "1";
  } catch {
    // Safari in private mode throws on access, not just on write.
    return false;
  }
};

const store = (() => {
  let state: StudioState = {
    navCollapsed: readStoredBool(NAV_KEY),
    mobileNavOpen: false,
    theme: preferredMode(),
    courses: [],
    coursesStatus: "idle",
    coursesError: null,
  };
  const listeners = new Set<() => void>();

  return {
    get: () => state,
    subscribe(fn: () => void) {
      listeners.add(fn);
      return () => listeners.delete(fn);
    },
    /** Merge a patch and notify. A no-op patch does not notify — every
     *  `set` that changes nothing would otherwise re-run every selector. */
    set(patch: Partial<StudioState>) {
      let changed = false;
      for (const k of Object.keys(patch) as (keyof StudioState)[]) {
        if (!Object.is(state[k], patch[k])) {
          changed = true;
          break;
        }
      }
      if (!changed) return;
      state = { ...state, ...patch };
      listeners.forEach((fn) => fn());
    },
  };
})();

/**
 * Read a slice. The selector's result is compared with `Object.is`, so select
 * primitives or stable references — `s => s.courses` is fine, `s => ({...})`
 * re-renders forever.
 */
export function useStudioStore<T>(selector: (s: StudioState) => T): T {
  const getSnapshot = useCallback(() => selector(store.get()), [selector]);
  return useSyncExternalStore(store.subscribe, getSnapshot, getSnapshot);
}

/** Non-reactive read, for event handlers that need the value once. */
export const getStudioState = (): StudioState => store.get();

// ---------------------------------------------------------------------------
// Actions
// ---------------------------------------------------------------------------

/** The in-flight course request, so concurrent callers share one GET. */
let inFlight: Promise<void> | null = null;

export const studio = {
  toggleNav() {
    const next = !store.get().navCollapsed;
    store.set({ navCollapsed: next });
    try {
      localStorage.setItem(NAV_KEY, next ? "1" : "0");
    } catch {
      // A persisted preference is a nicety; losing it must not break the nav.
    }
  },

  setMobileNav(open: boolean) {
    store.set({ mobileNavOpen: open });
  },

  /** Set the mode, repaint the tokens, and remember the choice. */
  setTheme(mode: ThemeMode) {
    store.set({ theme: mode });
    applyTheme(mode);
    try {
      localStorage.setItem(THEME_KEY, mode);
    } catch {
      // Same as above: the mode still applies to this session.
    }
  },

  toggleTheme() {
    studio.setTheme(store.get().theme === "dark" ? "light" : "dark");
  },

  /**
   * Load the course list once. Concurrent callers share the one request — the
   * sidebar and the Courses page both mount on the same navigation, and two
   * identical GETs is a race over which response lands last.
   */
  async loadCourses(force = false): Promise<void> {
    const { coursesStatus } = store.get();
    if (coursesStatus === "loading") return inFlight ?? undefined;
    if (coursesStatus === "ready" && !force) return;
    store.set({ coursesStatus: "loading", coursesError: null });
    inFlight = (async () => {
      try {
        const courses = await api.courses();
        store.set({ courses, coursesStatus: "ready", coursesError: null });
      } catch (e) {
        store.set({
          coursesStatus: "error",
          coursesError: e instanceof Error ? e.message : String(e),
        });
      } finally {
        inFlight = null;
      }
    })();
    return inFlight;
  },

  /** Drop the cache so the next read refetches. Call after create/delete. */
  invalidateCourses() {
    store.set({ coursesStatus: "idle" });
  },

  /** Test seam. Not used by the app. */
  __reset() {
    inFlight = null;
    store.set({
      navCollapsed: false,
      mobileNavOpen: false,
      courses: [],
      coursesStatus: "idle",
      coursesError: null,
    });
  },
};
