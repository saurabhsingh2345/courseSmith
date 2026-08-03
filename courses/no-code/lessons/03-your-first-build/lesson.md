---
title: "Your first build: one sentence, one app"
outcomes:
  - Turn an idea into a working app in a browser, in one sitting
  - Write a first prompt that gives the builder enough to work with
  - Change something and see it change, without touching a file
diagrams:
  - id: first-loop
    kind: mermaid
    prompt: "A closed cycle of four nodes with arrows returning to the start:
      'Describe what you want', 'Watch it build', 'Look at what you got',
      'Say what is wrong'. The arrow from the last node returns to the first.
      Label the return arrow 'every time'."
  - id: anatomy-of-a-brief
    kind: mermaid
    prompt: "A left-to-right flowchart from a node labelled 'A first prompt'
      to four nodes stacked vertically: 'Who it is for', 'What they do with
      it', 'What it must remember', 'What it should feel like'. One arrow from
      the first node to each of the four."
---

# Your first build

## What we are making
- A habit tracker: you list the things you mean to do daily, tick them off,
  and it shows you a streak.
- Small enough to finish today, real enough that the parts are all here —
  a screen, data that persists, a rule about time.
- Pick your own idea if you have one. Keep it to one sentence and one screen.

## The first prompt is a brief, not a wish
- "Make me a habit tracker" produces something generic, because it is generic.
- The builder cannot ask a follow-up question you did not invite. Whatever you
  leave out, it decides.
- Four things make the difference between generic and usable.

[DIAGRAM: anatomy-of-a-brief]

## The four things to say
- **Who it is for**: "for one person, on their phone, first thing in the morning."
- **What they do with it**: "add a habit, tick it off, see how many days in a row."
- **What it must remember**: "the habits, and which days each was ticked."
- **What it should feel like**: "calm, one screen, no accounts, no settings page."
- That is four sentences. It is worth the ninety seconds.

## Watch what it does with that
- It writes the screen, the data shape, and the logic in one pass.
- It will show you the app running, not a description of it.
- The first result is usually eighty percent right and confidently wrong about
  one thing. That is normal and it is what the next step is for.

## Then you talk to it
- Do not start over. Say what is wrong, specifically, one thing at a time.
- "The streak resets if I miss a day — it should forgive one missed day a week."
- Specific beats polite. It is not offended and vagueness costs you a round.

[DIAGRAM: first-loop]

## What you just did, in the terms of the map
- Layer two built you a screen and some logic.
- Layer four — the data — is being handled for you, invisibly, for now.
- Layer five has not happened: this exists in your browser and nowhere else.
- Lessons seven and nine are those two layers becoming yours.

## When it goes sideways
- If it produces something unrecognisable, your brief was ambiguous, not wrong.
  Read it back and find the sentence that could mean two things.
- If it keeps breaking the same thing, say so explicitly — "you have changed
  the streak logic three times and it is still wrong; here is exactly what I
  expect."
- If it is thirty percent right, start again with a better brief. Rescuing a
  bad start costs more than a restart.

## What's next
- Next lesson: prompting properly — why the same request gets a different
  answer, and how to stop that being luck.
