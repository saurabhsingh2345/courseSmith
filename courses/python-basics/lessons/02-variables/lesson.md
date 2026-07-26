---
title: "Variables: how programs remember"
outcomes:
  - Store a value in a variable and use it later
  - Choose clear, legal variable names
  - Update a variable and predict what prints
diagrams:
  - id: labeled-box
    kind: mermaid
    prompt: "A left-to-right flowchart showing a variable as a labelled box.
      A node labelled 'name = \"Ada\"' points with an arrow to a node
      labelled 'Box \"name\"\\nholds: \"Ada\"'. Below, a third node holds the
      caption 'the label stays, the contents can change' connected to the box
      node. Two nodes, clear left-to-right arrow, plus the caption node."
  - id: reassignment
    kind: mermaid
    prompt: "A left-to-right flowchart showing variable reassignment as a
      before/after. A node labelled 'score = 0\\nBox \"score\": 0' points with
      an arrow labelled 'assign again' to a node labelled
      'score = 10\\nBox \"score\": 10'. A third caption node reads 'assigning
      again replaces what was inside' and connects to the after node. The old
      value is simply gone in the after node."
callouts:
  - section: giving-a-value-a-name
    shape: circle
    x: 0.38
    y: 0.45
    label: no quotes on the left!
    at: "name of the box"
  - section: changing-what-is-inside
    shape: arrow
    x: 0.6
    y: 0.5
    label: old value is gone
    at: "replaces what was inside"
---

# Variables: how programs remember

## You already used one
- Last lesson's challenge: make print() say your own name. If you wrote
  your name straight inside print(), you had to repeat it every time.
- Last lesson's final script had a line like: name = "Ada" — that line is
  a variable. This lesson is about what that line really does.
- A variable is how a program remembers a value so it can use it again later.

## Giving a value a name
- Writing name = "Ada" stores the value "Ada" under the name name.
- The = sign means "store this", not "equals" like in math.
- The name of the box goes on the left, the value goes on the right.
- After that, writing name anywhere means "whatever is in that box".

[DIAGRAM: labeled-box]

```python
name = "Ada"
print(name)
print(name)
```

```output
Ada
Ada
```

## Names without quotes, text with quotes
- print("name") shows the text name; print(name) shows what the variable holds.
- Quotes mean "these exact characters"; no quotes means "look inside the box".

```python
name = "Ada"
print("name")
print(name)
```

```output
name
Ada
```

## Picking good names
- Legal names use letters, digits, and underscores, and can't start with a digit.
- Python is case-sensitive: age and Age are two different boxes.
- Choose names that say what's inside: birth_year beats by.
- Multi-word names use underscores: favorite_color, not favoritecolor.

## Changing what is inside
- Assigning again replaces the old value — the box keeps its label, the
  contents are swapped.
- The old value is gone; Python only remembers the latest assignment.

[DIAGRAM: reassignment]

```python
score = 0
print(score)
```

```output
0
```

Then assign again — the box keeps its label, the contents are swapped:

```python
score = 0
print(score)
score = 10
print(score)
```

```output
0
10
```

[DEMO: open the Python REPL with python3, create a variable mood = "curious", print it, reassign mood = "confident", print it again, and exit]

## Variables can hold numbers too
- Everything from last lesson's math works inside variables.
- You can use a variable's current value to compute a new one.

```python
apples = 3
apples = apples + 2
print(apples)
```

```output
5
```

## Variables working together
- Real programs mix several variables in one line.
- print() can take several things separated by commas — it prints them
  with spaces in between.

```python
name = "Ada"
age = 12
print(name, "is", age)
print(name, "turns", age + 1, "next year")
```

```output
Ada is 12
Ada turns 13 next year
```

[DEMO: create a script about_me.py with a heredoc containing name and age variables and a print that combines them, then run it with python3 about_me.py]

## What's next
- Next lesson: text itself — what strings can do, from joining to
  changing case.
- Challenge before then: make a script with two variables, your name and
  your favorite number, and print one sentence using both.
