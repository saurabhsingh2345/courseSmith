---
title: "Design: making it look chosen"
outcomes:
  - Say why generated interfaces look the way they do
  - Make four decisions that lift a project out of the default look
  - Hand a design to a builder in a form it can actually use
diagrams:
  - id: four-decisions
    kind: mermaid
    prompt: "A hub-and-spoke flowchart with a central node 'Looking chosen'
      connected to four surrounding nodes: 'One typeface, two sizes', 'One
      accent colour', 'Space, deliberately', 'One idea per screen'. Four
      spokes, nothing else connected."
  - id: design-handoff
    kind: mermaid
    prompt: "A left-to-right flowchart from 'A picture of the screen' and
      'The rules behind it' both pointing to a node 'What the builder can
      use'. Below, a node 'Just a screenshot' points to a node 'It guesses the
      rules' with a dashed edge."
---

# Design

## Why everything generated looks the same
- A builder with no direction picks the safest option at every decision.
- Safe choices are, by definition, the ones everybody else also picked.
- The result is competent and forgettable — which is fine for a prototype and
  fatal for anything asking a stranger for their attention.

## You do not need to be a designer
- You need to make four decisions and then not undermine them.
- Design tools are where those decisions get made visible, so that you and the
  builder are looking at the same thing.
- If you make the four decisions in a sentence and skip the tool, that is
  still better than not making them.

[DIAGRAM: four-decisions]

## Decision one: one typeface, two sizes
- One family for everything. Two sizes: the thing you read and the thing you
  scan.
- Three typefaces is the most reliable way to look amateur.

## Decision two: one accent colour
- Everything is a shade of grey except one colour, which means "this is the
  thing to press".
- If two things are the accent, neither is.

## Decision three: space, on purpose
- Beginners fill the screen. The difference between generated and designed is
  usually the emptiness.
- Doubling the space around the important thing is the cheapest improvement
  available.

## Decision four: one idea per screen
- A screen that does three things teaches the visitor nothing about what
  matters.
- If you cannot say what a screen is for in four words, it is two screens.

## Handing it over
- A screenshot alone makes the builder guess the rules — it will infer a
  colour and miss the reason for it.
- Say the rules alongside the picture: the accent colour, the typeface, the
  spacing unit, what the screen is for.
- The picture shows the destination. The rules are what let it get there twice.

[DIAGRAM: design-handoff]

## The honest limit of this lesson
- Everything here will make an ordinary project look deliberate.
- None of it will make a beautiful thing. That is still a craft, it still takes
  years, and there is no prompt for it.
- Knowing which one you need is the useful part.

## What's next
- Next lesson: data. Where the things your app remembers actually live.
