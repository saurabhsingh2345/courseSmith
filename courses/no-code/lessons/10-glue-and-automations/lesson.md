---
title: "Glue: connecting it to everything else"
outcomes:
  - Describe an automation as a trigger, some steps, and a result
  - Choose between an automation tool and asking a builder for the feature
  - Avoid the two failure modes that make automations untrustworthy
diagrams:
  - id: trigger-steps-result
    kind: mermaid
    prompt: "A left-to-right chain of three nodes with arrows: 'Trigger —
      something happened', 'Steps — do these in order', 'Result — something
      changed somewhere else'. Straight chain, one arrow between each pair."
  - id: build-or-connect
    kind: mermaid
    prompt: "A decision flowchart from a node 'Does it belong in your app?'
      branching to two nodes: 'Yes — your users see it, so build it in' and
      'No — it is your own back office, so connect it'. Two branches only."
---

# Glue

## The unglamorous half of software
- Most working software is not one clever thing. It is several ordinary things
  passing messages.
- A form fills a spreadsheet. The spreadsheet pings a channel. The channel
  reminds a person.
- None of that is impressive and all of it is what people actually pay for.

## Every automation has the same shape
- A **trigger**: something happened. A form was submitted. A row appeared.
- **Steps**: do these things, in order, possibly with a condition.
- A **result**: something changed somewhere else.
- Once you see this shape, every automation tool looks like the same tool with
  a different logo.

[DIAGRAM: trigger-steps-result]

## Build it in, or connect it?
- If the person using your app sees it happen, it belongs in the app.
- If it is your own back office — a notification to you, a record in your
  spreadsheet — connect it and move on.
- The mistake is building a whole feature to do something an automation would
  have done in ten minutes. The opposite mistake is rarer and cheaper.

[DIAGRAM: build-or-connect]

## What these tools are worth
- They connect to things you would otherwise have to read documentation for.
- They retry when something fails, and tell you when it kept failing.
- That second one is most of the value. Anybody can call something once.

## Failure mode one: the silent automation
- It stops working and nothing tells you, because nothing was watching.
- The symptom appears weeks later as "we stopped getting those emails".
- Give every automation a way to complain. A message to yourself when it fails
  is enough.

## Failure mode two: the automation nobody remembers
- Six months on, something happens every Tuesday and nobody knows what does it.
- Name them for what they do, not for what they connect.
- Keep one list of every automation you have running. One page, boring, current.

## Where the cost hides
- These tools charge per run, and a trigger that fires more than you expected
  is the usual surprise.
- Check what your trigger's real frequency is before you connect it to
  something that costs money per fire.

## What's next
- Next lesson: when it breaks. Reading an error without being afraid of it.
