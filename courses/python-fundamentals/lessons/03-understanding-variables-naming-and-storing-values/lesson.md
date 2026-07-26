---
title: "Understanding variables: naming and storing values"
outcomes:
  - "Define what a variable is"
  - "Create variables with meaningful names"
  - "Store and retrieve values using variables"
diagrams:
  - id: variables-overview
    kind: mermaid
    prompt: "Top-to-bottom flowchart showing the concept of variables. Start with 'Variable' at the top, pointing down to 'Name' and 'Value' side by side, with arrows indicating that a variable has a name and stores a value."
---

# Understanding variables: naming and storing values

## What is a variable?
- A variable is a container for storing data values.
- It has a name that you use to reference the stored value.
- Variables can hold different types of data, like numbers or text.

## Naming variables
- Use descriptive names that reflect the value stored.
- Start with a letter or underscore, followed by letters, numbers, or underscores.
- Avoid using spaces or special characters in variable names.

[DIAGRAM: variables-overview]

## Creating and storing values in variables
- You can create a variable using the assignment operator '='.
- Assign a value to a variable by writing the name, followed by '=', then the value.
- You can store different types of values in variables.

```python
my_variable = 10
print(my_variable)
```

```output
10
```

## Retrieving values from variables
- To use the value stored in a variable, simply reference its name.
- You can use variables in expressions and functions.
- Variables can be updated with new values at any time.

```python
my_variable = 10
my_variable = my_variable + 5
print(my_variable)
```

```output
15
```

## What’s next?
- Now that you know about variables, you can explore data types further.
- Next, we'll learn about lists and how to store multiple values.
- Understanding variables is key to writing more complex programs.
