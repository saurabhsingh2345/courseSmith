// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, expect, it, vi } from "vitest";
import { CoursePreview } from "./CoursePreview";
import { api, type LessonDetail } from "../api/client";

const LESSON: LessonDetail = {
  course: "python-basics",
  id: "01-what-is-python",
  title: "What is Python?",
  source: "lesson.md",
  stages: {},
  stage_order: [],
  artifacts: [
    { name: "final.mp4", size: 1000, url: "/artifacts/python-basics/01/final.mp4" },
    {
      name: "diagrams/where-python-runs.svg",
      size: 200,
      url: "/artifacts/python-basics/01/diagrams/where-python-runs.svg",
    },
    // attempts must be excluded from the diagrams grid.
    {
      name: "diagrams/attempts/where-python-runs-1.svg",
      size: 200,
      url: "/artifacts/python-basics/01/diagrams/attempts/where-python-runs-1.svg",
    },
    {
      name: "code_traces/manifest.json",
      size: 50,
      url: "/artifacts/python-basics/01/code_traces/manifest.json",
    },
    {
      name: "code_traces/sha256:abc.json",
      size: 100,
      url: "/artifacts/python-basics/01/code_traces/sha256:abc.json",
    },
  ],
  quiz: {
    title: "Quiz",
    questions: [
      {
        id: "q1",
        type: "recall",
        prompt: "What is Python?",
        options: ["A snake", "A language"],
        answer_index: 1,
        explanation: "It's a programming language.",
      },
    ],
  },
};

function renderPreview() {
  return render(
    <MemoryRouter initialEntries={["/c/python-basics/l/01-what-is-python/preview"]}>
      <Routes>
        <Route path="/c/:slug/l/:id/preview" element={<CoursePreview />} />
      </Routes>
    </MemoryRouter>,
  );
}

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

it("renders the lesson title and default video tab from real artifacts", async () => {
  vi.spyOn(api, "lesson").mockResolvedValue(LESSON);
  const { container } = renderPreview();

  expect(await screen.findByText("What is Python?")).toBeTruthy();
  await waitFor(() => {
    const video = container.querySelector("video");
    expect(video?.getAttribute("src")).toBe("/artifacts/python-basics/01/final.mp4");
  });
});

it("shows only real diagrams (excluding attempts) on the diagrams tab", async () => {
  vi.spyOn(api, "lesson").mockResolvedValue(LESSON);
  const { container } = renderPreview();
  await screen.findByText("What is Python?");

  // Radix Tabs (automatic activation) switches on focus, not a bare click.
  fireEvent.focus(screen.getByRole("tab", { name: /Diagrams/i }));
  await waitFor(() => {
    const imgs = container.querySelectorAll("img");
    expect(imgs.length).toBe(1);
    expect(imgs[0].getAttribute("src")).toContain("diagrams/where-python-runs.svg");
  });
});

it("renders the quiz with the correct answer marked", async () => {
  vi.spyOn(api, "lesson").mockResolvedValue(LESSON);
  renderPreview();
  await screen.findByText("What is Python?");

  fireEvent.focus(screen.getByRole("tab", { name: /Quiz/i }));
  expect(await screen.findByText(/Q1\. What is Python\?/)).toBeTruthy();
  expect(screen.getByText("✓ A language")).toBeTruthy();
});
