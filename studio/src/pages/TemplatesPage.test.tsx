// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, expect, it, vi } from "vitest";
import { TemplatesPage } from "./TemplatesPage";
import { api, type ArchetypeCatalog, type Course } from "../api/client";
import { studio } from "../store/studioStore";

const CATALOG: ArchetypeCatalog = {
  archetypes: [
    {
      name: "concept-first",
      description: "Deep conceptual understanding.",
      default_animation: "smooth",
      pace_hint: 135,
      prompt_hint: "One idea at a time.",
    },
    {
      name: "project-led",
      description: "Build something end to end.",
      default_animation: "playful",
      pace_hint: 160,
      prompt_hint: "Ship it, then explain it.",
    },
  ],
  animation_styles: ["minimal", "smooth", "playful"],
  palettes: [{ name: "cool", colors: { primary: "#5b4b8a", accent: "#0aa3a3", background: "#f7fbfb" } }],
};

const COURSES = [{ slug: "python-basics", name: "Python Basics" }] as unknown as Course[];

const renderPage = () =>
  render(
    <MemoryRouter>
      <TemplatesPage />
    </MemoryRouter>,
  );

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  studio.__reset();
});

it("renders the archetype catalog and selects the first by default", async () => {
  vi.spyOn(api, "archetypes").mockResolvedValue(CATALOG);
  vi.spyOn(api, "courses").mockResolvedValue(COURSES);
  renderPage();

  expect(await screen.findByRole("radio", { name: /concept-first/ })).toBeTruthy();
  // The detail panel resolves the first archetype without anything being clicked.
  expect(screen.getByRole("radio", { name: /concept-first/ }).getAttribute("aria-checked")).toBe(
    "true",
  );
  expect(screen.getByText("One idea at a time.")).toBeTruthy();
  expect(screen.getByText("cool")).toBeTruthy();
});

it("switches the detail panel when another archetype is picked", async () => {
  vi.spyOn(api, "archetypes").mockResolvedValue(CATALOG);
  vi.spyOn(api, "courses").mockResolvedValue(COURSES);
  renderPage();

  fireEvent.click(await screen.findByRole("radio", { name: /project-led/ }));

  expect(screen.getByText("Ship it, then explain it.")).toBeTruthy();
  expect(screen.queryByText("One idea at a time.")).toBeNull();
  expect(screen.getByRole("radio", { name: /project-led/ }).getAttribute("aria-checked")).toBe(
    "true",
  );
});

it("applies the selected archetype to a course through the real endpoint", async () => {
  vi.spyOn(api, "archetypes").mockResolvedValue(CATALOG);
  vi.spyOn(api, "courses").mockResolvedValue(COURSES);
  const update = vi
    .spyOn(api, "updateCourse")
    .mockResolvedValue({} as Awaited<ReturnType<typeof api.updateCourse>>);
  renderPage();

  fireEvent.click(await screen.findByRole("radio", { name: /project-led/ }));
  fireEvent.change(await screen.findByLabelText("Course"), {
    target: { value: "python-basics" },
  });
  fireEvent.click(screen.getByRole("button", { name: /^Apply$/ }));

  await waitFor(() =>
    expect(update).toHaveBeenCalledWith("python-basics", { archetype: "project-led" }),
  );
  expect(await screen.findByText(/is now the archetype for Python Basics/)).toBeTruthy();
});

it("says so rather than offering Apply when there are no courses", async () => {
  vi.spyOn(api, "archetypes").mockResolvedValue(CATALOG);
  vi.spyOn(api, "courses").mockResolvedValue([]);
  renderPage();

  expect(await screen.findByText(/No courses yet/)).toBeTruthy();
  expect(screen.queryByRole("button", { name: /^Apply$/ })).toBeNull();
});
