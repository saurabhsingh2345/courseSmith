/**
 * Compute the minimal QuizOverrides payload by diffing an edited quiz
 * against the generated (canonical) quiz. Only fields that differ from
 * the generated question are included in each override entry.
 */

export interface QuizQuestion {
  id: string;
  type?: string;
  prompt: string;
  options: string[];
  answer_index: number;
  explanation: string;
  review?: boolean;
}

export interface QuestionOverride {
  id: string;
  drop?: boolean;
  prompt?: string;
  options?: string[];
  answer_index?: number;
  explanation?: string;
}

export interface QuizOverridesPayload {
  questions: QuestionOverride[];
}

/** State of a question in the editor: edited fields plus a drop flag. */
export interface EditedQuestion {
  id: string;
  drop: boolean;
  prompt: string;
  options: string[];
  answer_index: number;
  explanation: string;
}

function sameOptions(a: string[], b: string[]): boolean {
  return a.length === b.length && a.every((v, i) => v === b[i]);
}

/**
 * Diff a single edited question against its generated original.
 * Returns null when nothing differs (no override needed).
 */
export function diffQuestion(
  generated: QuizQuestion,
  edited: EditedQuestion,
): QuestionOverride | null {
  const override: QuestionOverride = { id: generated.id };
  let changed = false;

  if (edited.drop) {
    override.drop = true;
    changed = true;
  }
  if (edited.prompt !== generated.prompt) {
    override.prompt = edited.prompt;
    changed = true;
  }
  if (!sameOptions(edited.options, generated.options)) {
    override.options = [...edited.options];
    changed = true;
  }
  if (edited.answer_index !== generated.answer_index) {
    override.answer_index = edited.answer_index;
    changed = true;
  }
  if (edited.explanation !== generated.explanation) {
    override.explanation = edited.explanation;
    changed = true;
  }
  return changed ? override : null;
}

/**
 * Build the minimal overrides payload for the full quiz.
 * Edited questions with no generated counterpart are ignored (the backend
 * only merges overrides onto generated questions by id).
 */
export function computeOverrides(
  generated: QuizQuestion[],
  edited: EditedQuestion[],
): QuizOverridesPayload {
  const byId = new Map(generated.map((q) => [q.id, q]));
  const questions: QuestionOverride[] = [];
  for (const e of edited) {
    const gen = byId.get(e.id);
    if (!gen) continue;
    const d = diffQuestion(gen, e);
    if (d) questions.push(d);
  }
  return { questions };
}

/** Convert a merged/generated question into the editor's editable state. */
export function toEdited(q: QuizQuestion, drop = false): EditedQuestion {
  return {
    id: q.id,
    drop,
    prompt: q.prompt,
    options: [...q.options],
    answer_index: q.answer_index,
    explanation: q.explanation,
  };
}

/** Which fields of an override entry actually carry a change (for badges). */
export function overriddenFields(o: QuestionOverride): string[] {
  const fields: string[] = [];
  if (o.drop) fields.push("drop");
  if (o.prompt !== undefined) fields.push("prompt");
  if (o.options !== undefined) fields.push("options");
  if (o.answer_index !== undefined) fields.push("answer");
  if (o.explanation !== undefined) fields.push("explanation");
  return fields;
}
