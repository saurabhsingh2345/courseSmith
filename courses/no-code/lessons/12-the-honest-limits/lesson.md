---
title: "The honest limits"
outcomes:
  - Say what these tools genuinely cannot do yet
  - Estimate what a running project costs, including the parts nobody mentions
  - Judge how stuck you would be if a tool disappeared tomorrow
diagrams:
  - id: the-cost-lines
    kind: mermaid
    prompt: "A hub-and-spoke flowchart with a central node 'What it costs to
      run' connected to five nodes: 'The builder subscription', 'The AI, per
      use', 'Hosting', 'The database', 'The domain'. Five spokes only."
  - id: lock-in
    kind: mermaid
    prompt: "A left-to-right flowchart from a node 'If this tool vanished'
      to three nodes: 'Do I have the files?', 'Do I have the data?', 'Could
      somebody else run it?'. One arrow to each of the three."
---

# The honest limits

## Why this lesson is last and not first
- Put it first and it reads as a warning nobody has context for.
- Put it last, after you have built something that works, and every line of it
  means something.
- This is the lesson that separates somebody who can use these tools from
  somebody who believes the advertising.

## What they are genuinely bad at
- **Anything with a lot of rules that interact.** Tax, payroll, scheduling with
  exceptions. It will produce something that looks right and is wrong in the
  fifth case.
- **Performance at scale.** It writes code that works for a hundred rows. Ten
  million rows is a different craft.
- **Anything that has to be exactly right every time.** It is a very good
  approximator, and approximation is the wrong tool for money.

## What they are bad at in a way you will not notice
- Producing code that is fine today and unmaintainable in six months.
- Silently choosing an approach that closes off the thing you will want next.
- These do not announce themselves. They show up as "why is everything so hard
  now" a quarter later.

## What it actually costs
- The builder's subscription, monthly.
- The AI itself, per use — this one grows with your usage and is the line
  people get wrong.
- Hosting, cheap until it is not.
- The database, usually free until a real number of rows.
- The domain, yearly and trivial.
- Add them up before you promise anybody anything. Then double the AI line.

[DIAGRAM: the-cost-lines]

## The three questions about lock-in
- **Do I have the files?** If the answer is only "they are in the tool", you do
  not have a project, you have a subscription.
- **Do I have the data?** Export it once, now, and see what you get.
- **Could somebody else run this?** If a developer could pick it up, you are
  fine. If it only exists inside one product, you are not.

[DIAGRAM: lock-in]

## Where you should stop and hire somebody
- Payments and anything that touches a card.
- Health, financial, or children's data.
- Anything where being wrong costs somebody else money or safety.
- Anything you would have to disclose if you lost it.
- Hiring here is not admitting defeat. It is the same judgement as not doing
  your own electrics.

## What is genuinely true
- You can build and ship a real, useful thing without writing code. You just
  did.
- The ceiling is much higher than it was and it is not absent.
- The people who do best with these tools are the ones who know exactly where
  that ceiling is — which, having got this far, you now do.

## What's next
- Last lesson: the capstone. The whole flow, start to finish, in one go.
