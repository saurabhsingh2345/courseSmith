---
title: "Data without SQL"
outcomes:
  - Explain what a database is, in terms of tables and rows
  - Design the tables for a small app before building it
  - Say what a relationship between two tables actually means
diagrams:
  - id: table-anatomy
    kind: mermaid
    prompt: "A left-to-right flowchart from a node labelled 'A table' to three
      nodes: 'Columns — what every row has', 'Rows — one thing each', 'Types —
      what may go in a column'. One arrow from the table node to each."
  - id: two-tables
    kind: mermaid
    prompt: "An entity relationship style flowchart with two nodes: a node
      labelled 'habits (id, name)' and a node labelled 'ticks (id, habit_id,
      date)'. A single arrow from the ticks node to the habits node labelled
      'habit_id points here'. Only two nodes and one arrow."
---

# Data without SQL

## Why this is the lesson people skip
- Layer four is invisible when it works, which makes it feel optional.
- It is not optional. It is the difference between an app that remembers you
  and a demo that forgets everything when you close the tab.
- Getting the tables wrong is also the most expensive mistake to fix later,
  because everything else is built on their shape.

## A database is a spreadsheet with rules
- A **table** is a sheet. A **row** is one thing. A **column** is a fact every
  one of those things has.
- The rules are what a spreadsheet lacks: a column that must be a date really
  is a date, and a row that must be unique really is.
- That is the whole idea. Everything else is detail.

[DIAGRAM: table-anatomy]

## Designing the tables before you build
- List the nouns in your idea. For the habit tracker: habits, and ticks.
- Give each noun a table. Give each table the facts you need about it.
- Do this on paper in four minutes. It will save you an afternoon.

## One table or two?
- The tempting version: one `habits` table with a column holding all the dates.
- It breaks the moment you want to ask "what did I do on Tuesday" — you have to
  read every row and pull the list apart.
- Two tables: `habits` for the things, `ticks` for each time one was done, with
  each tick pointing at its habit.

[DIAGRAM: two-tables]

## What a relationship is
- The pointing is the relationship. A tick holds the id of its habit.
- That is all "relational" means, and it is why the databases you will meet are
  called relational databases.
- One habit, many ticks. Most relationships you will need are that shape.

## Doing it without writing SQL
- Modern database services give you a table editor: name the table, add
  columns, pick types, done.
- Your builder or agent will write the queries. You are deciding the shape,
  not the syntax.
- You will still see SQL occasionally. It reads roughly like English and you
  can ask what any of it means.

## The rule that saves you later
- Never store the same fact in two places.
- If a habit's name appears in both tables, one of them will eventually be
  wrong, and nothing will tell you which.

## What's next
- Next lesson: accounts, secrets, and the small number of things here that can
  genuinely hurt you.
