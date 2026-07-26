// @vitest-environment jsdom
import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useTutorAPI } from "./useTutorAPI";

function mockFetchOnce(body: unknown, ok = true, status = 200) {
  const fetchMock = vi.fn().mockResolvedValue({
    ok,
    status,
    json: async () => body,
  } as Response);
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("useTutorAPI", () => {
  it("maps a correct/incorrect sequence to the BKT wire shape and returns the estimate", async () => {
    const estimate = {
      p_known: 0.72,
      p_next_correct: 0.8,
      mastered: false,
      difficulty_hint: "medium",
      recommendation: "Keep practicing at the current level.",
    };
    const fetchMock = mockFetchOnce(estimate);

    const { result } = renderHook(() => useTutorAPI("http://tutor.test"));

    let value: unknown;
    await act(async () => {
      value = await result.current.estimateMastery([true, false, true]);
    });

    expect(value).toEqual(estimate);
    expect(result.current.error).toBeNull();
    expect(result.current.loading).toBe(false);

    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe("http://tutor.test/bkt/estimate");
    expect(JSON.parse((init as RequestInit).body as string)).toEqual({
      responses: [{ correct: true }, { correct: false }, { correct: true }],
    });
  });

  it("posts the rating for a review schedule", async () => {
    const fsrs = { stability: 2.5, difficulty: 4.9, interval_days: 3, note: "x" };
    const fetchMock = mockFetchOnce(fsrs);

    const { result } = renderHook(() => useTutorAPI("http://tutor.test"));
    let value: unknown;
    await act(async () => {
      value = await result.current.scheduleReview("good");
    });

    expect(value).toEqual(fsrs);
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe("http://tutor.test/fsrs/schedule");
    expect(JSON.parse((init as RequestInit).body as string).rating).toBe("good");
  });

  it("returns null and sets a friendly error when the service is unreachable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("ECONNREFUSED")));

    const { result } = renderHook(() => useTutorAPI("http://tutor.test"));
    let value: unknown = "sentinel";
    await act(async () => {
      value = await result.current.estimateMastery([true]);
    });

    expect(value).toBeNull();
    expect(result.current.error).toContain("unreachable");
  });

  it("surfaces a server error body", async () => {
    mockFetchOnce({ error: "bad request" }, false, 400);
    const { result } = renderHook(() => useTutorAPI("http://tutor.test"));
    await act(async () => {
      await result.current.scheduleReview("easy");
    });
    expect(result.current.error).toBe("bad request");
  });
});
