// @vitest-environment jsdom
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { AdaptiveOverlay } from "./AdaptiveOverlay";
import type { Quiz } from "../api/client";

const QUIZ: Quiz = {
  title: "Q",
  questions: [
    { id: "q1", type: "recall", prompt: "a", options: ["x", "y"], answer_index: 0, explanation: "" },
    { id: "q2", type: "recall", prompt: "b", options: ["x", "y"], answer_index: 1, explanation: "" },
    { id: "q3", type: "application", prompt: "c", options: ["x", "y"], answer_index: 0, explanation: "" },
  ],
};

// BKT mock: p_known scales with the number of correct responses in the seed, so
// different dimensions land at different mastery.
function installFetch() {
  const fetchMock = vi.fn(async (_url: string, init?: RequestInit) => {
    const body = JSON.parse((init?.body as string) ?? "{}") as {
      responses: { correct: boolean }[];
    };
    const n = body.responses.filter((r) => r.correct).length;
    const pk = Math.min(0.99, 0.2 + n * 0.15);
    return {
      ok: true,
      status: 200,
      json: async () => ({
        p_known: pk,
        p_next_correct: pk,
        mastered: pk >= 0.95,
        difficulty_hint: pk > 0.8 ? "harder" : "medium",
        recommendation: "Keep going.",
      }),
    } as Response;
  });
  vi.stubGlobal("fetch", fetchMock);
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

it("renders a mastery card per question type present in the quiz", async () => {
  installFetch();
  render(<AdaptiveOverlay quiz={QUIZ} />);

  // Only recall + application appear (not debugging/prediction).
  expect(await screen.findByText("Recall")).toBeTruthy();
  expect(screen.getByText("Application")).toBeTruthy();
  expect(screen.queryByText("Debugging")).toBeNull();

  // Recall seed [T,T,T,F,T] → 4 correct → pk 0.80 → 80%.
  await waitFor(() => expect(screen.getByText("80%")).toBeTruthy());
});

it("shows a friendly note when the tutor service is unreachable", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => {
      throw new TypeError("Failed to fetch");
    }),
  );
  render(<AdaptiveOverlay quiz={QUIZ} />);
  expect(await screen.findByText(/Tutor service not running/i)).toBeTruthy();
});

it("prompts to run the quiz stage when there are no questions", () => {
  installFetch();
  render(<AdaptiveOverlay quiz={{ title: "Q", questions: [] }} />);
  expect(screen.getByText(/No quiz questions to model/i)).toBeTruthy();
});
