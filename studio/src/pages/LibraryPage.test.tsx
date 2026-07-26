// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { LibraryPage } from "./LibraryPage";
import { api, type LibraryDiagram } from "../api/client";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

it("shows the empty state then a created diagram", async () => {
  const listSpy = vi.spyOn(api, "libraryDiagrams").mockResolvedValue([]);
  vi.spyOn(api, "libraryQuestions").mockResolvedValue([]);
  const created: LibraryDiagram = {
    id: "1",
    name: "Flow",
    kind: "mermaid",
    source: "graph TD; A-->B",
    created_at: "2026-01-01T00:00:00Z",
  };
  const createSpy = vi.spyOn(api, "createLibraryDiagram").mockResolvedValue(created);

  render(<LibraryPage />);
  expect(await screen.findByText("No saved diagrams.")).toBeTruthy();

  // After a create, the list reloads — return the new item on the next fetch.
  listSpy.mockResolvedValue([created]);
  fireEvent.change(screen.getByPlaceholderText("Name"), { target: { value: "Flow" } });
  fireEvent.click(screen.getByText("Save diagram"));

  await waitFor(() => expect(createSpy).toHaveBeenCalled());
  expect(await screen.findByText(/Flow/)).toBeTruthy();
});
