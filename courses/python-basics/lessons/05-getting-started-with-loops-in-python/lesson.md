---
title: "Getting started with loops in Python"
outcomes:
  - "Understand what loops are"
  - "Write simple loops to repeat actions"
  - "Use loops to process lists of items"
diagrams:
  - id: loop-structure
    kind: mermaid
    prompt: "Top-to-bottom flowchart showing the structure of a loop. Start with 'Start Loop', flow down to 'Condition Check', then branch to 'Action' if true, and back to 'Condition Check'. If false, flow down to 'End Loop'."
  - id: for-loop-example
    kind: mermaid
    prompt: "Top-to-bottom flowchart illustrating a for loop example. Start with 'Define List', flow down to 'For each item in List', then to 'Print item', and finally to 'End Loop'."
---

# Getting started with loops in Python

## What are loops?
- Loops allow us to repeat actions in our code.
- Think of loops like a chef following a recipe multiple times.
- They help automate repetitive tasks, saving time and effort.

## The structure of a loop
- A loop starts with a condition that needs to be checked.
- If the condition is true, the loop executes an action.
- Once the action is complete, it checks the condition again.

[DIAGRAM: loop-structure]

## Using a for loop
- A for loop iterates over a sequence, like a list.
- It allows us to perform an action for each item in that sequence.
- For example, we can print each item in a list.

```python
items = ['apple', 'banana', 'cherry']
for item in items:
    print(item)
```

```output
apple
banana
cherry
```

[DIAGRAM: for-loop-example]

## Using a while loop
- A while loop continues as long as a condition is true.
- It's useful when the number of iterations isn't known beforehand.
- Be careful to avoid infinite loops by ensuring the condition eventually becomes false.

```python
count = 0
while count < 5:
    print(count)
    count += 1
```

```output
0
1
2
3
4
```

## Practical example: counting items
- Let's use a loop to count items in a list.
- This helps us understand how many items we have.
- We can combine loops with lists for powerful results.

```python
items = ['dog', 'cat', 'mouse']
count = 0
for item in items:
    count += 1
print(count)
```

```output
3
```

## What’s next?
- Now that you know about loops, you can explore nested loops.
- Nested loops allow you to loop within a loop.
- This opens up new possibilities for complex tasks.
