---
title: "Loops: Repeating things with for and while"
outcomes:
  - "Understand what loops are"
  - "Use for loops to iterate over lists"
  - "Use while loops to repeat actions conditionally"
diagrams:
  - id: loop-types
    kind: mermaid
    prompt: "Top-to-bottom flowchart showing two types of loops: 'For Loop' and 'While Loop'. 'For Loop' points to 'Iterate over a collection', while 'While Loop' points to 'Repeat while a condition is true'."
---

# Loops: Repeating things with for and while

## What are loops?
- Loops help repeat actions in code
- They reduce the need for repetitive code
- Two main types: for loops and while loops

## Using for loops
- For loops iterate over a sequence like a list
- They execute a block of code for each item
- Great for tasks that require repetition over collections

```python
for number in range(5):
    print(number)
```

```output
0
1
2
3
4
```

## Understanding while loops
- While loops repeat as long as a condition is true
- They are useful for unknown iteration counts
- Be careful to avoid infinite loops

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

## When to use each loop
- For loops are best for fixed iterations
- While loops are best for conditional repetitions
- Choosing the right loop can simplify your code

[DIAGRAM: loop-types]

## Practice with loops
- Try creating a for loop that prints items in a list
- Create a while loop that counts down from 10
- Experiment with different conditions and collections

[DEMO: In your terminal, create a list of fruits and use a for loop to print each fruit.]

## What’s next?
- Next, we'll explore how to combine loops with conditionals
- Learn how to break and continue in loops
- Build more complex programs using loops and conditions
