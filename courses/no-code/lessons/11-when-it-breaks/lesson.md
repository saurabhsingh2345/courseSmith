---
title: "When it breaks"
outcomes:
  - Read an error message for the two things that matter
  - Give an AI enough to actually fix it, first time
  - Work out which layer a problem lives in before trying to fix it
diagrams:
  - id: anatomy-of-an-error
    kind: mermaid
    prompt: "A left-to-right flowchart from a node labelled 'An error message'
      to two nodes: 'What went wrong (the message)' and 'Where it went wrong
      (the file and line)'. One arrow to each. A third node labelled
      'Everything else is noise' points at the error node with a dashed edge."
  - id: which-layer
    kind: mermaid
    prompt: "A decision flowchart from a node 'Nothing works at all?' branching
      to 'Yes — hosting' and to a node 'Wrong information showing?' which
      branches to 'Yes — data' and to 'No — the logic'. A chain of two
      decisions."
---

# When it breaks

## It breaking is not a sign you did it wrong
- Everything breaks. Professionals spend most of their time on this.
- The difference is not that theirs breaks less. It is that they are not
  frightened by the message.
- The message is not shouting at you. It is the most cooperative thing in
  software: it is telling you exactly what it could not do.

## An error says two things
- **What** went wrong: one sentence, usually near the top.
- **Where** it went wrong: a file and a line number.
- Everything between those two is the path it took to get there, and you can
  ignore it until you need it.

[DIAGRAM: anatomy-of-an-error]

## Read the first line and the last line
- The first line is usually the actual problem.
- The last line is usually your code, as opposed to somebody's library.
- If those two disagree about what happened, the answer is in between and
  that is the moment to ask for help.

## Which layer is it?
- Nothing loads at all: hosting.
- It loads but shows the wrong information: data.
- It loads, the data is right, it behaves wrong: the logic.
- Guessing the layer first stops you fixing a data problem in the design.

[DIAGRAM: which-layer]

## Giving an AI enough to fix it
- Paste the **whole** error, not your summary of it.
- Say what you expected and what happened instead.
- Say what you changed most recently. This is the single most useful sentence
  and almost nobody offers it.
- Do not tell it what the fix is. You are the expert on the symptom.

## Watch a real one get fixed
- What follows is a genuine recording. The bug is real, the failure is real,
  and the fix is whatever actually happened — including any wrong turn.
- The waiting is condensed and the frame says by how much. Nothing is cut:
  every step is there, including the ones that went nowhere.
- Watch how much of the work is *finding* the problem rather than fixing it.

[CAPTURE: tool=claude, fixture=habit-tracker; show the streak calculation giving the wrong answer, then have the agent find the cause and fix it]

## When to stop and get help
- You have tried the same fix three times.
- The error mentions money, accounts, or other people's data.
- You do not understand the fix that worked, and it touches something that
  matters.
- Stopping is a decision, not a failure. Knowing when is the skill.

## What's next
- Next lesson: the honest limits. What this genuinely cannot do, and what it
  costs.
