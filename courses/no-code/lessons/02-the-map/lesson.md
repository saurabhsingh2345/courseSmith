---
title: "The map: what each tool is for"
outcomes:
  - Place any new tool you meet into one of five layers
  - Explain what a database, a host, and a domain each do
  - Pick the right kind of tool for a job instead of the most-advertised one
diagrams:
  - id: the-stack
    kind: mermaid
    prompt: "A top-to-bottom flowchart of five stacked layers, each a single
      node, connected downward in order: 'Where you describe it (chat, prompt)',
      'Where it gets built (app builders, AI editors)', 'Where it looks right
      (design tools)', 'Where the data lives (database)', 'Where it runs
      (hosting and a domain)'. Straight vertical chain, one arrow between each
      pair."
  - id: choosing
    kind: mermaid
    prompt: "A decision flowchart starting from a node 'What do you have?'.
      It branches to 'Just an idea' which points to 'Start in an app builder';
      to 'A design' which points to 'Start in an AI editor'; and to 'A working
      app' which points to 'Go straight to hosting'. Three branches, no
      cross-links."
---

# The map

## Why the names feel like noise
- Every week there is a new tool with a one-word name and a bold claim.
- They are not all doing different things. Most sit in one of five layers.
- Learn the layers once and every new name becomes "oh, that's another one of
  those" instead of something else to keep up with.

## Layer one: where you describe it
- Plain chat assistants. You type what you want in ordinary language.
- Good for thinking, naming, planning, and writing the brief you will paste
  into everything else.
- They do not build or host anything. They are where the idea gets sharp.

## Layer two: where it gets built
- Two flavours, and the difference matters more than the brand names.
- **App builders** run in a browser: you describe an app, it appears, you keep
  talking to it. Nothing is installed. This is where you should start.
- **AI editors** run on your computer, on the real files. More power, more
  ceremony. Lesson five is where we move here, and it explains why you'd want to.

## Layer three: where it looks right
- Design tools decide what the thing looks like before it is built.
- You can skip this layer and let the builder choose. It will look generic —
  competent and forgettable.
- Worth an hour when the thing is going in front of people whose opinion counts.

## Layer four: where the data lives
- A database is a set of tables. A table is a spreadsheet with rules.
- Everything your app remembers between visits lives here: accounts, orders,
  posts, settings.
- This is the layer beginners skip and then rebuild. Lesson seven is entirely
  about it.

## Layer five: where it runs
- Hosting is a computer, somewhere else, that stays on so people can reach
  your app.
- A domain is the readable address that points at it.
- Until you do this, what you have built exists only for you.

[DIAGRAM: the-stack]

## Reading the map backwards
- The layers are also the order things go wrong in, bottom-up: hosting breaks
  loudest, data breaks worst, design breaks quietly.
- When something is broken, ask which layer it lives in before you ask what to
  do about it.

[DIAGRAM: choosing]

## Where to start, for you specifically
- If you have an idea and nothing else: layer two, browser flavour.
- If you have a design already: layer two, editor flavour.
- If you have something working and nobody can see it: layer five.
- Everything else is optimisation.

## What's next
- Next lesson: one sentence in, a working application out. The whole promise
  of this course, demonstrated.
