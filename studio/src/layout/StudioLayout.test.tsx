// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { StudioLayout } from "./StudioLayout";
import { NAV_GROUPS } from "../components/SideNav";
import { RunProvider } from "../state/RunContext";
import { ShortcutProvider } from "../state/ShortcutContext";
import { api } from "../api/client";
import { studio } from "../store/studioStore";

const renderShell = (path = "/") =>
  render(
    <MemoryRouter initialEntries={[path]}>
      <ShortcutProvider>
        <RunProvider>
          <StudioLayout logOpen={false} onCloseLog={() => {}}>
            <div>page body</div>
          </StudioLayout>
        </RunProvider>
      </ShortcutProvider>
    </MemoryRouter>,
  );

beforeEach(() => {
  studio.__reset();
  localStorage.clear();
  // RunBar polls on mount; keep the shell tests off the network.
  vi.spyOn(api, "runStatus").mockResolvedValue({
    running: false,
  } as Awaited<ReturnType<typeof api.runStatus>>);
});

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

it("renders every nav destination exactly once", () => {
  renderShell();
  const nav = screen.getAllByRole("navigation", { name: "Studio sections" })[0];
  const entries = NAV_GROUPS.flatMap((g) => g.entries);

  for (const e of entries) {
    expect(within(nav).getByRole("link", { name: e.label }), e.label).toBeTruthy();
  }
  expect(within(nav).getAllByRole("link")).toHaveLength(entries.length);
});

it("marks the destination for the current route as current", () => {
  renderShell("/library");
  const nav = screen.getAllByRole("navigation", { name: "Studio sections" })[0];
  expect(within(nav).getByRole("link", { name: "Library" }).getAttribute("aria-current")).toBe(
    "page",
  );
});

it("keeps Courses current on a nested course route", () => {
  // /c/:slug is a course page; a nav that de-highlights there tells the reader
  // they have left the section they are plainly still in.
  renderShell("/c/python-basics/l/01");
  const nav = screen.getAllByRole("navigation", { name: "Studio sections" })[0];
  expect(within(nav).getByRole("link", { name: "Courses" }).getAttribute("aria-current")).toBe(
    "page",
  );
});

it("collapses and expands the rail, and says which it will do", () => {
  renderShell();
  const collapse = screen.getByRole("button", { name: "Collapse navigation" });
  fireEvent.click(collapse);

  expect(screen.getByRole("button", { name: "Expand navigation" })).toBeTruthy();
  // Collapsed is a rail of icons, so the labels leave the accessible name of
  // the link and become its aria-label — the destinations must still be there.
  const nav = screen.getAllByRole("navigation", { name: "Studio sections" })[0];
  expect(within(nav).getByRole("link", { name: "Ledger" })).toBeTruthy();
});

it("opens and closes the mobile drawer", () => {
  renderShell();
  expect(screen.getAllByRole("navigation", { name: "Studio sections" })).toHaveLength(1);

  fireEvent.click(screen.getByRole("button", { name: "Open navigation" }));
  expect(screen.getAllByRole("navigation", { name: "Studio sections" })).toHaveLength(2);

  fireEvent.click(screen.getByRole("button", { name: "Close navigation" }));
  expect(screen.getAllByRole("navigation", { name: "Studio sections" })).toHaveLength(1);
});

it("closes the drawer on navigation", () => {
  // A drawer that survives the tap covers the page it just opened.
  renderShell();
  fireEvent.click(screen.getByRole("button", { name: "Open navigation" }));
  const drawerNav = screen.getAllByRole("navigation", { name: "Studio sections" })[1];

  fireEvent.click(within(drawerNav).getByRole("link", { name: "Adaptive" }));

  expect(screen.getAllByRole("navigation", { name: "Studio sections" })).toHaveLength(1);
});

it("toggles the theme from the header", () => {
  renderShell();
  const toggle = screen.getByRole("button", { name: /Switch to (light|dark) mode/ });
  const before = document.documentElement.dataset.theme;

  fireEvent.click(toggle);

  expect(document.documentElement.dataset.theme).not.toBe(before);
});

it("renders the page body it is given", () => {
  renderShell();
  expect(screen.getByText("page body")).toBeTruthy();
});
