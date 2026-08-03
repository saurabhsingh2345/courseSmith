// The habits a person is tracking, and which days each was ticked off.
export const habits = [
  { id: 'water', name: 'Drink water', done: ['2026-07-29', '2026-07-30', '2026-07-31'] },
  { id: 'walk', name: 'Walk outside', done: ['2026-07-30', '2026-07-31'] },
  { id: 'read', name: 'Read ten pages', done: ['2026-07-31'] },
];

// streak counts back from today until it finds a day that was missed.
export function streak(habit, today) {
  let days = 0;
  const day = new Date(today);
  while (habit.done.includes(day.toISOString().slice(0, 10))) {
    days += 1;
    day.setDate(day.getDate() - 1);
  }
  return days;
}
