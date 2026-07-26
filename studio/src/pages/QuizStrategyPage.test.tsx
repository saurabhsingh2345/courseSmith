// @vitest-environment jsdom
import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { QuizStrategyPage, interleave, adjacentRepeats } from "./QuizStrategyPage";
import { api, type Question, type QuizWithOverrides } from "../api/client";

function q(id: string, type: Question["type"]): Question {
  return { id, type, prompt: `${id}?`, options: ["a", "b"], answer_index: 0, explanation: "e" };
}

describe("interleave", () => {
  it("drives adjacent same-type pairs to the floor for a spreadable mix", () => {
    // 3 recall + 2 application can interleave to zero adjacent repeats.
    const input = [
      q("a", "recall"),
      q("b", "recall"),
      q("c", "recall"),
      q("d", "application"),
      q("e", "application"),
    ];
    expect(adjacentRepeats(interleave(input))).toBe(0);
    // never drops or duplicates a question
    expect(interleave(input).map((x) => x.id).sort()).toEqual(["a", "b", "c", "d", "e"]);
  });

  it("is deterministic", () => {
    const input = [q("a", "recall"), q("b", "application"), q("c", "debugging")];
    expect(interleave(input).map((x) => x.id)).toEqual(interleave(input).map((x) => x.id));
  });
});

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

it("renders the quiz strategy analysis from a lesson quiz", async () => {
  const quiz: QuizWithOverrides = {
    merged: {
      title: "Q",
      questions: [q("a", "recall"), q("b", "recall"), q("c", "application"), q("d", "debugging")],
    },
  };
  vi.spyOn(api, "quiz").mockResolvedValue(quiz);

  render(
    <MemoryRouter initialEntries={["/c/py/l/01/strategy"]}>
      <Routes>
        <Route path="/c/:slug/l/:id/strategy" element={<QuizStrategyPage />} />
      </Routes>
    </MemoryRouter>,
  );

  // The first prompt appears once loaded, and the type-mix legend lists types.
  expect(await screen.findByText("Quiz strategy")).toBeTruthy();
  expect(await screen.findByText(/recall: 2/)).toBeTruthy();
});
