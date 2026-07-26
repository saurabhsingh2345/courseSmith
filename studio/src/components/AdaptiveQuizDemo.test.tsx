// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { AdaptiveQuizDemo } from "./AdaptiveQuizDemo";

// Route the mock by endpoint. BKT reflects the number of correct answers so the
// bar visibly moves; FSRS returns a fixed interval.
function installFetch() {
  const fetchMock = vi.fn(async (url: string, init?: RequestInit) => {
    if (String(url).includes("/bkt/estimate")) {
      const body = JSON.parse((init?.body as string) ?? "{}") as {
        responses: { correct: boolean }[];
      };
      const nCorrect = body.responses.filter((r) => r.correct).length;
      const pk = Math.min(0.99, 0.2 + nCorrect * 0.15);
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
    }
    return {
      ok: true,
      status: 200,
      json: async () => ({ stability: 3, difficulty: 5, interval_days: 3, note: "x" }),
    } as Response;
  });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

beforeEach(() => installFetch());
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

it("renders a mastery estimate and raises it when a correct answer is added", async () => {
  render(<AdaptiveQuizDemo />);

  // Initial sequence [true,true,false,true] → 3 correct → pk 0.65 → 65%.
  const bar = await screen.findByText(/mastered \d+% of this concept/);
  expect(bar.textContent).toContain("65%");

  fireEvent.click(screen.getByRole("button", { name: /Answered correctly/i }));

  // 4 correct → pk 0.80 → 80%.
  await waitFor(() =>
    expect(screen.getByText(/mastered \d+% of this concept/).textContent).toContain("80%"),
  );
});

it("shows an FSRS next-review interval after rating a review", async () => {
  render(<AdaptiveQuizDemo />);
  await screen.findByText(/mastered \d+%/);

  fireEvent.click(screen.getByRole("button", { name: "Good" }));

  const line = await screen.findByText(/Next review in/);
  expect(line.textContent).toContain("3 days");
});
