import { habits, streak } from './habits.js';

const TODAY = '2026-07-31';

export function render() {
  return habits
    .map((h) => `${h.name}: ${streak(h, TODAY)} day streak`)
    .join('\n');
}

console.log(render());
