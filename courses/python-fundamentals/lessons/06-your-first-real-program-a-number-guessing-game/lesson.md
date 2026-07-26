---
title: "Your first real program: a number guessing game"
outcomes:
  - "Combine print, variables, if-else, and loops"
  - "Build a complete number guessing game"
  - "Understand the flow of a simple program"
diagrams:
  - id: number-guessing-game-flow
    kind: mermaid
    prompt: "Top-to-bottom flowchart showing the flow of the number guessing game. Start with 'Pick a secret number', then 'Take the next guess', then 'Check the guess' as a decision: 'correct' leads to 'Print congratulations' and 'End game', while 'wrong' leads to 'Print try again' which loops back up to 'Take the next guess'."
---

# Your first real program: a number guessing game

## What we're building
- A number guessing game — the classic first program
- The program picks a secret number, the player tries to find it
- It uses everything from this course: print, variables, if-else, and loops

## Picking the secret number
- The random module can pick a number for us
- random.seed() makes the choice repeatable — perfect for testing
- random.randint(1, 10) gives a whole number from 1 to 10

```python
import random
random.seed(7)
secret_number = random.randint(1, 10)
print('The secret number is chosen!')
```

```output
The secret number is chosen!
```

## Checking a guess with if-else
- Compare the guess to the secret number with ==
- Print a win message when they match
- Print a hint when they don't

```python
secret_number = 6
user_guess = 4
if user_guess == secret_number:
    print('Congratulations! You guessed it!')
else:
    print('Not quite — try again!')
```

```output
Not quite — try again!
```

## The game loop: guessing until it's right
- A program can't type, so we give it a list of guesses to play with
- The for loop tries each guess in order, just like a player would
- if-else decides what to print for every guess

```python
import random
random.seed(7)
secret_number = random.randint(1, 10)
guesses = [2, 9, secret_number]
for user_guess in guesses:
    if user_guess == secret_number:
        print('Congratulations!', user_guess, 'is right!')
    else:
        print(user_guess, 'is too far off — try again!')
```

```output
2 is too far off — try again!
9 is too far off — try again!
Congratulations! 6 is right!
```

[DIAGRAM: number-guessing-game-flow]

## What your program just did
- Picked a secret number and stored it in a variable
- Looped over the guesses one at a time
- Made a decision about each guess and printed the result
- That is a real program — the same shape as software you use every day

## Where to go from here
- Change the range to 1-100 and add more guesses
- Count the attempts with a variable that grows by 1 each loop
- When you run Python on your own computer, swap the guess list for input() to make it truly interactive
