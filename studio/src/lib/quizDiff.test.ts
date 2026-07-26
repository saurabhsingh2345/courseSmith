import { describe, expect, it } from "vitest";
import {
  computeOverrides,
  diffQuestion,
  overriddenFields,
  toEdited,
  type QuizQuestion,
} from "./quizDiff";

const q1: QuizQuestion = {
  id: "q1",
  type: "recall",
  prompt: "What is a variable?",
  options: ["A named box", "An error", "A file", "A loop"],
  answer_index: 0,
  explanation: "A variable names a value.",
};

const q2: QuizQuestion = {
  id: "q2",
  type: "application",
  prompt: "What prints the variable color?",
  options: ["print(color)", "print('color')", "print[color]", "echo color"],
  answer_index: 0,
  explanation: "No quotes around variable names.",
};

describe("diffQuestion", () => {
  it("returns null when nothing changed", () => {
    expect(diffQuestion(q1, toEdited(q1))).toBeNull();
  });

  it("includes only the changed field", () => {
    const edited = toEdited(q1);
    edited.prompt = "What is a variable, really?";
    expect(diffQuestion(q1, edited)).toEqual({
      id: "q1",
      prompt: "What is a variable, really?",
    });
  });

  it("detects option edits by value and order", () => {
    const reordered = toEdited(q1);
    reordered.options = ["An error", "A named box", "A file", "A loop"];
    expect(diffQuestion(q1, reordered)).toEqual({
      id: "q1",
      options: ["An error", "A named box", "A file", "A loop"],
    });

    const addedOption = toEdited(q1);
    addedOption.options = [...q1.options, "A number"];
    expect(diffQuestion(q1, addedOption)?.options).toHaveLength(5);
  });

  it("detects answer_index change to 0-adjacent values", () => {
    const edited = toEdited(q1);
    edited.answer_index = 2;
    expect(diffQuestion(q1, edited)).toEqual({ id: "q1", answer_index: 2 });
  });

  it("emits drop-only override for an untouched dropped question", () => {
    const edited = toEdited(q1, true);
    expect(diffQuestion(q1, edited)).toEqual({ id: "q1", drop: true });
  });

  it("combines multiple changed fields", () => {
    const edited = toEdited(q2, true);
    edited.explanation = "Quotes make it a string literal.";
    edited.answer_index = 1;
    expect(diffQuestion(q2, edited)).toEqual({
      id: "q2",
      drop: true,
      answer_index: 1,
      explanation: "Quotes make it a string literal.",
    });
  });
});

describe("computeOverrides", () => {
  it("returns an empty payload when nothing changed", () => {
    const payload = computeOverrides([q1, q2], [toEdited(q1), toEdited(q2)]);
    expect(payload).toEqual({ questions: [] });
  });

  it("only includes questions that changed", () => {
    const e2 = toEdited(q2);
    e2.prompt = "Which line prints the value of color?";
    const payload = computeOverrides([q1, q2], [toEdited(q1), e2]);
    expect(payload.questions).toHaveLength(1);
    expect(payload.questions[0]).toEqual({
      id: "q2",
      prompt: "Which line prints the value of color?",
    });
  });

  it("ignores edited questions with no generated counterpart", () => {
    const ghost = toEdited({ ...q1, id: "q99" });
    ghost.prompt = "edited";
    expect(computeOverrides([q1], [ghost])).toEqual({ questions: [] });
  });

  it("resetting an edit back to generated removes the override", () => {
    const edited = toEdited(q1);
    edited.prompt = "changed";
    edited.prompt = q1.prompt; // reset
    expect(computeOverrides([q1], [edited])).toEqual({ questions: [] });
  });

  it("handles empty inputs", () => {
    expect(computeOverrides([], [])).toEqual({ questions: [] });
    expect(computeOverrides([q1], [])).toEqual({ questions: [] });
  });
});

describe("overriddenFields", () => {
  it("labels each carried field", () => {
    expect(
      overriddenFields({ id: "q1", drop: true, answer_index: 1, options: ["a"] }),
    ).toEqual(["drop", "options", "answer"]);
    expect(overriddenFields({ id: "q1" })).toEqual([]);
  });
});

describe("toEdited", () => {
  it("deep-copies options so edits do not mutate the source", () => {
    const edited = toEdited(q1);
    edited.options[0] = "mutated";
    expect(q1.options[0]).toBe("A named box");
  });
});
