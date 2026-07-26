---
title: "If and else: making decisions in code"
outcomes:
  - "Understand the concept of conditional statements"
  - "Write if and else statements in Python"
  - "Use comparison operators to make decisions"
diagrams:
  - id: if-else-flowchart
    kind: mermaid
    prompt: "Top-to-bottom flowchart showing the flow of an if-else statement. Start with 'Condition Check'. If 'True', flow to 'Execute Code Block A'. If 'False', flow to 'Execute Code Block B'. End after executing either block."
---

# If and else: making decisions in code

## Understanding conditional statements
- Conditional statements allow us to make decisions in code.
- They evaluate conditions and execute code based on the result.
- The main types are 'if' and 'else' statements.

## The structure of if statements
- An if statement checks a condition.
- If the condition is true, it runs the associated code block.
- If the condition is false, it skips the code block.

```python
age = 18
if age >= 18:
    print("You are an adult.")
```

```output
You are an adult.
```

## Adding else statements
- An else statement provides an alternative action.
- It runs when the if condition is false.
- Combining if and else allows for two possible outcomes.

```python
age = 16
if age >= 18:
    print("You are an adult.")
else:
    print("You are a minor.")
```

```output
You are a minor.
```

## Using comparison operators
- Comparison operators help define conditions.
- Common operators include ==, !=, >, <, >=, and <=.
- They compare values to determine true or false.

```python
temperature = 30
if temperature > 25:
    print("It's a hot day.")
else:
    print("It's a cool day.")
```

```output
It's a hot day.
```

## Flowchart for if and else statements
- Flowcharts visually represent the logic of if-else statements.
- They show the flow from condition checks to actions.
- Understanding flowcharts helps in planning code.

[DIAGRAM: if-else-flowchart]

## What's next?
- Now that you know about if and else, you can explore elif.
- Elif allows for multiple conditions to be checked.
- Next, we'll learn about loops to repeat actions based on conditions.
