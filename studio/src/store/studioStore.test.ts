// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getStudioState, studio } from "./studioStore";
import { api, type Course } from "../api/client";

const COURSES = [
  { slug: "a", name: "A" },
  { slug: "b", name: "B" },
] as unknown as Course[];

beforeEach(() => {
  studio.__reset();
  localStorage.clear();
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("courses", () => {
  it("loads once and serves the cache afterwards", async () => {
    const spy = vi.spyOn(api, "courses").mockResolvedValue(COURSES);

    await studio.loadCourses();
    await studio.loadCourses();

    expect(spy).toHaveBeenCalledTimes(1);
    expect(getStudioState().courses).toEqual(COURSES);
    expect(getStudioState().coursesStatus).toBe("ready");
  });

  it("shares one request between concurrent callers", async () => {
    // The sidebar and the page it navigated to mount together. Two GETs is a
    // race over which response lands last, and the loser wins the render.
    const spy = vi.spyOn(api, "courses").mockResolvedValue(COURSES);

    await Promise.all([studio.loadCourses(), studio.loadCourses(), studio.loadCourses()]);

    expect(spy).toHaveBeenCalledTimes(1);
  });

  it("refetches after invalidation, but not otherwise", async () => {
    const spy = vi.spyOn(api, "courses").mockResolvedValue(COURSES);

    await studio.loadCourses();
    studio.invalidateCourses();
    await studio.loadCourses();

    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("records the failure and stays loadable", async () => {
    vi.spyOn(api, "courses").mockRejectedValueOnce(new Error("offline"));
    await studio.loadCourses();
    expect(getStudioState().coursesStatus).toBe("error");
    expect(getStudioState().coursesError).toBe("offline");

    // An error is not a cache: the next call must actually retry, or a blip at
    // boot leaves the sidebar empty until a reload.
    vi.spyOn(api, "courses").mockResolvedValue(COURSES);
    await studio.loadCourses();
    expect(getStudioState().coursesStatus).toBe("ready");
  });
});

describe("ui state", () => {
  it("persists the nav collapse but not the mobile drawer", () => {
    studio.toggleNav();
    expect(getStudioState().navCollapsed).toBe(true);
    expect(localStorage.getItem("cs-nav-collapsed")).toBe("1");

    studio.setMobileNav(true);
    expect(getStudioState().mobileNavOpen).toBe(true);
    expect(localStorage.getItem("cs-mobile-nav")).toBeNull();
  });

  it("toggles the theme, applies it to the document, and remembers it", () => {
    const start = getStudioState().theme;
    studio.toggleTheme();
    const next = getStudioState().theme;

    expect(next).not.toBe(start);
    expect(document.documentElement.dataset.theme).toBe(next);
    expect(localStorage.getItem("cs-theme")).toBe(next);
    // The ramp has to move with it, or only the shell changes colour.
    expect(document.documentElement.style.getPropertyValue("--ink-950")).toBeTruthy();
  });

  it("survives a localStorage that throws", () => {
    const setItem = vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("QuotaExceededError");
    });
    expect(() => studio.toggleNav()).not.toThrow();
    expect(getStudioState().navCollapsed).toBe(true);
    setItem.mockRestore();
  });
});
