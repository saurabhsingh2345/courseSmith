---
title: "Accounts, secrets, and what will bite you"
outcomes:
  - Tell the difference between who someone is and what they may do
  - Keep a secret out of the code without understanding cryptography
  - Name the three mistakes that account for most beginner breaches
diagrams:
  - id: authn-authz
    kind: mermaid
    prompt: "A left-to-right flowchart with two nodes in sequence: a node
      labelled 'Who are you? (signing in)' points to a node labelled 'What may
      you do? (permissions)'. A third node labelled 'Most breaches happen
      here' points at the second node with a dashed edge."
  - id: where-secrets-live
    kind: mermaid
    prompt: "A two-column comparison flowchart. Left column headed 'Safe' has
      nodes 'On the server', 'In an environment variable', 'In a secrets
      manager'. Right column headed 'Not safe' has nodes 'In the code',
      'In the browser', 'In a screenshot'. Columns are not connected."
---

# Accounts, secrets, and what will bite you

## Two different questions
- **Signing in** answers "who are you". It is the part everybody thinks about.
- **Permissions** answer "what may you do". It is the part that gets skipped.
- An app can know exactly who you are and still let you read somebody else's
  data. That is not a login problem.

[DIAGRAM: authn-authz]

## Do not build sign-in yourself
- Every builder and database service has this ready-made, tested by more people
  than will ever use your app.
- Building your own is a genuine security decision, made by accident, by
  somebody not intending to make it.
- Ask for the built-in one by name. It is one sentence in your brief.

## What a secret is
- A key that lets your app act as itself: connect to the database, send email,
  charge a card.
- Anyone holding it can do those things as you, and the bill is yours.
- They are not passwords for people. They are passwords for programs.

## Where a secret may live
- On the server, where only your running app can read it.
- In an environment variable — a setting your host holds and hands to the app.
- Never in the code, because the code goes into version control and version
  control remembers forever.

[DIAGRAM: where-secrets-live]

## The screenshot problem
- The most common way a key leaks is not a hack. It is a screenshot in a chat,
  a video, or a support ticket.
- Assume anything on your screen while recording is public.
- If a key ever appears anywhere it should not, replace it. Do not reason about
  whether anyone saw it.

## The three that account for most of it
- **A secret in the code.** Then the repository goes public and it is over.
- **No permission check.** Signed in as anyone, able to read everything.
- **Trusting the browser.** Anything your app checks in the browser can be
  switched off by the person using it. Checks that matter happen on the server.

## The line where you stop
- Payments, health records, anything about children, anything you would have to
  report if you lost it.
- This is not "be careful". It is "this needs somebody whose job it is".
- Knowing where the line is *is* the professional skill here. Lesson twelve
  goes further into it.

## What's next
- Next lesson: shipping. Putting it somewhere with an address, where other
  people can open it.
