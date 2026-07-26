// @vitest-environment jsdom
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { TemplatesPage } from "./TemplatesPage";
import { api, type ArchetypeCatalog } from "../api/client";

const CATALOG: ArchetypeCatalog = {
  archetypes: [
    {
      name: "concept-first",
      description: "Deep conceptual understanding.",
      default_animation: "smooth",
      pace_hint: 135,
      prompt_hint: "One idea at a time.",
    },
  ],
  animation_styles: ["minimal", "smooth", "playful"],
  palettes: [{ name: "cool", colors: { primary: "#5b4b8a", accent: "#0aa3a3", background: "#f7fbfb" } }],
};

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

it("renders the archetype catalog", async () => {
  vi.spyOn(api, "archetypes").mockResolvedValue(CATALOG);
  render(<TemplatesPage />);

  expect(await screen.findByText("concept-first")).toBeTruthy();
  expect(screen.getByText("One idea at a time.")).toBeTruthy();
  expect(screen.getByText("cool")).toBeTruthy();
});
