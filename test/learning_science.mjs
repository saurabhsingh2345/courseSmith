// Learning-science invariants for a generated quiz (workstream I, working).
//
// Validates a lesson's quiz_sequence.json against the cognitive-science rules
// the quiz-strategy stage (workstream G) is supposed to enforce:
//   1. the sequence is a permutation of the quiz's questions (no drops/dupes)
//   2. interleaving: adjacent same-type questions stay under a threshold
//   3. difficulty targets sum to the question count and skew toward medium
//
// Usage: node test/learning_science.mjs <lesson-generated-dir>
// Exit 0 = all invariants hold; exit 1 = a violation (printed).

import { readFileSync } from "node:fs";
import { join } from "node:path";

const MAX_ADJACENT_SAME_TYPE_RATIO = 0.34; // at most ~1/3 of transitions repeat

const dir = process.argv[2];
if (!dir) {
  console.error("usage: node test/learning_science.mjs <lesson-generated-dir>");
  process.exit(2);
}

const readJSON = (name) => JSON.parse(readFileSync(join(dir, name), "utf8"));

let quiz;
let sequence;
try {
  quiz = readJSON("quiz.json");
  sequence = readJSON("quiz_sequence.json");
} catch (err) {
  console.error(`could not read quiz artifacts in ${dir}: ${err.message}`);
  process.exit(2);
}

const failures = [];
const typeOf = new Map(quiz.questions.map((q) => [q.id, q.type]));
const order = sequence.order ?? [];

// 1. Permutation check.
const quizIds = new Set(quiz.questions.map((q) => q.id));
const orderIds = new Set(order);
if (order.length !== quiz.questions.length || orderIds.size !== quizIds.size) {
  failures.push(`sequence has ${order.length} ids for ${quiz.questions.length} questions`);
}
for (const id of order) {
  if (!quizIds.has(id)) failures.push(`sequence references unknown question id ${id}`);
}

// 2. Interleaving check.
let adjacent = 0;
for (let i = 1; i < order.length; i++) {
  if (typeOf.get(order[i]) === typeOf.get(order[i - 1])) adjacent++;
}
const transitions = Math.max(1, order.length - 1);
if (adjacent / transitions > MAX_ADJACENT_SAME_TYPE_RATIO) {
  failures.push(`interleaving weak: ${adjacent}/${transitions} adjacent same-type (max ${MAX_ADJACENT_SAME_TYPE_RATIO})`);
}

// 3. Difficulty distribution check.
const targets = sequence.difficulty_targets ?? {};
const sum = (targets.easy ?? 0) + (targets.medium ?? 0) + (targets.hard ?? 0);
if (sum !== quiz.questions.length) {
  failures.push(`difficulty targets sum to ${sum}, want ${quiz.questions.length}`);
}
if ((targets.medium ?? 0) < (targets.easy ?? 0) || (targets.medium ?? 0) < (targets.hard ?? 0)) {
  failures.push(`difficulty targets should skew medium, got ${JSON.stringify(targets)}`);
}

if (failures.length > 0) {
  console.error(`FAIL ${dir}`);
  for (const f of failures) console.error(`  - ${f}`);
  process.exit(1);
}
console.log(`PASS ${dir}: ${order.length} questions, ${adjacent} adjacent same-type, targets ${JSON.stringify(targets)}`);
