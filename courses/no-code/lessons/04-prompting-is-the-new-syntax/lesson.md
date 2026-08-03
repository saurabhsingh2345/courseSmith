---
title: "Prompting is the new syntax"
outcomes:
  - Write a request that produces the same result twice
  - Give context deliberately instead of hoping it was remembered
  - Recognise the four ways a prompt usually fails
diagrams:
  - id: prompt-failures
    kind: mermaid
    prompt: "A hub-and-spoke flowchart with a central node 'Why it built the
      wrong thing' connected to four surrounding nodes: 'You were ambiguous',
      'You assumed it remembered', 'You asked for too much at once', 'You
      described the fix, not the problem'. Four spokes, no other links."
  - id: context-window
    kind: mermaid
    prompt: "A left-to-right flowchart showing three nodes feeding into one:
      'What you just typed', 'What is on screen', and 'What you said earlier'
      all point to a node labelled 'What it actually knows'. A fourth node
      labelled 'What you assumed it knows' points at the same node with a
      dashed edge."
---

# Prompting is the new syntax

## Why this is a real skill
- The old skill was remembering exact syntax; a misplaced bracket broke everything.
- The new skill is describing precisely; a vague sentence builds the wrong thing.
- Both are unforgiving. Only one of them can be practised in ordinary English.

## It is not a magic phrase
- There is no secret wording that unlocks better output.
- Prompt templates you find online mostly work because they are *specific*,
  not because of the words they use.
- What actually moves the result: concrete nouns, stated constraints, and one
  request at a time.

## The four ways it goes wrong
- **Ambiguity**: "make the list better" — better how? Sorted? Shorter? Prettier?
- **Assumed memory**: it does not necessarily still know what you said twenty
  messages ago, and it will not tell you it forgot.
- **Too much at once**: five changes in one message, three get done, and you
  cannot tell which two were dropped.
- **Prescribing the fix**: "add a filter to the query" assumes your diagnosis.
  Say the symptom and let it find the cause.

[DIAGRAM: prompt-failures]

## Say the symptom, not the cure
- Bad: "wrap the date in a formatter."
- Good: "the date shows as 2026-07-31T08:00:00Z and should read like
  31 July." — now it can fix the actual cause, which may not be formatting.
- You are not the expert on the code. You are the expert on what is wrong.

## Context is something you give, not something it has
- Paste the error. Paste the relevant part. Say which screen you are on.
- "It's broken" starts a guessing game you pay for by the round.

[DIAGRAM: context-window]

## The loop, deliberately
- One change. Look at the result. Say the next thing.
- It feels slower than a big request and it is faster, because you never lose
  track of which change caused what.
- When something works, say so — "keep that" — before asking for the next thing.

## A brief you can reuse
- Keep a short document describing your project: what it is, who uses it, what
  the data looks like, what the rules are.
- Paste it at the start of a new session. This is the single highest-value
  habit in the course.
- It is also what you would hand a new colleague, which is the point.

## What's next
- Next lesson: leaving the browser. When your project outgrows a chat box,
  and what an AI editor gives you that a builder cannot.
