---
title: What is Python?
outcomes:
  - Explain what a programming language does
  - Install Python on your own computer
  - Run your very first line of Python
diagrams:
  - id: python-translator
    kind: mermaid
    prompt: "A left-to-right flowchart showing Python translating a person's
      code into computer actions. Three nodes connected by arrows: a node
      labelled 'You write:\\nprint(\"Hello\")', then a node labelled
      'Python interpreter\\nreads your code line by line\\nand carries it out',
      then a node labelled 'Computer shows:\\nHello'. Two arrows flow left to
      right, one from the first node to the interpreter and one from the
      interpreter to the computer."
  - id: where-python-runs
    kind: mermaid
    prompt: "A hub-and-spoke flowchart with a central node labelled 'Python'
      connected by a plain edge to each of four surrounding nodes showing where
      Python is used: 'Websites', 'Data & Science', 'Automation', and
      'AI & Machine Learning'. The central Python node links to all four; the
      four are not connected to each other."
callouts:
  - section: your-first-line-of-python
    shape: circle
    x: 0.5
    y: 0.42
    label: quotes matter!
    at: "inside the quotes"
  - section: installing-python
    shape: arrow
    x: 0.62
    y: 0.55
    label: don't skip this box
    at: "add python to path"
---

# What is Python?

## A language for talking to computers
- Computers only follow instructions; they can't guess what you mean.
- A programming language is a precise way of writing instructions.
- Python is one of the most popular ones — designed to read almost like English.
- Created by Guido van Rossum, first released in 1991; named after
  Monty Python, not the snake.

[DIAGRAM: python-translator]

## Why beginners start with Python
- Very little ceremony: one readable line can do real work.
- Huge community: almost any question you'll have has an answer online.
- The same language scales from tiny scripts to serious software.

## Where Python is used in the real world
- Websites and apps (Instagram, Spotify backends).
- Data analysis and science (spreadsheets on steroids).
- Automating boring tasks (renaming 1,000 files in a second).
- AI and machine learning — most of it is written in Python.

[DIAGRAM: where-python-runs]

## Installing Python
- Go to python.org/downloads and grab the latest version for your system.
- On Windows: check the "Add Python to PATH" box in the installer.
- On macOS: the installer just works; Linux usually has Python already.
- To check it worked, open a terminal and type: python3 --version

## Your first line of Python
- The traditional first program: make the computer say hello.
- print() takes whatever is inside the quotes and shows it on screen.

```python
print("Hello, world!")
```

```output
Hello, world!
```

[DEMO: open the Python REPL with python3, print a hello message, then print your own name, and exit]

## Python does math too
- You don't need quotes for numbers — Python calculates them.
- The REPL answers back immediately, like a very obedient calculator.

```python
print(7 * 6)
print(10 + 5)
```

```output
42
15
```

## Putting lines together
- A program is just lines executed top to bottom, one after another.
- Save them in a file, run the file, and Python performs each line in order.

```python
name = "Ada"
print("Hello,", name)
print("Welcome to Python!")
```

```output
Hello, Ada
Welcome to Python!
```

[DEMO: create a two-line script hello.py with a heredoc and run it with python3 hello.py]

## What's next
- Next lesson: variables — how programs remember things.
- Challenge before then: make print() say your own name.
