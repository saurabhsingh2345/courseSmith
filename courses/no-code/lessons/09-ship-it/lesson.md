---
title: "Ship it: hosting, domains, going live"
outcomes:
  - Put a project somewhere other people can open it
  - Explain what a deploy actually does
  - Keep a history you can go back to when something breaks
diagrams:
  - id: what-a-deploy-is
    kind: mermaid
    prompt: "A left-to-right chain of four nodes with arrows: 'Your files',
      'Uploaded', 'Built into something a browser can run', 'Live at an
      address'. Straight chain, one arrow between each pair."
  - id: domain-and-host
    kind: mermaid
    prompt: "A left-to-right flowchart with a node labelled 'yourname.com (the
      address)' pointing with an arrow labelled 'points at' to a node labelled
      'the host (the computer that answers)'. Only two nodes and one arrow."
---

# Ship it

## Until you do this, it does not exist
- Everything so far has run on your machine, for you.
- Shipping is the step that turns it into something you can send someone.
- It is also, reliably, the step people put off longest and find easiest once
  they have done it once.

## What a deploy actually does
- Your files get copied to a computer that stays on.
- That computer builds them into the form a browser can run.
- The result gets an address. Anyone with the address reaches it.
- That is the whole ceremony. It takes under a minute.

[DIAGRAM: what-a-deploy-is]

## History first, deploying second
- Before shipping anything, put the project in version control.
- It is a record of every version, so "it worked on Tuesday" becomes something
  you can act on rather than a feeling.
- It is also how the host knows what to deploy, and how anyone else ever helps
  you.

[CAPTURE: tool=gh, fixture=habit-tracker; show the project's repository and its recent history from the command line]

## Watch a real deploy
- What follows is a genuine recording of a project going live. The address at
  the end is a real one.
- Most of a deploy is a progress bar. The video condenses that waiting so it
  fits, and states the real elapsed time in the corner — because "it deployed
  instantly" would teach you the wrong expectation.

[CAPTURE: tool=vercel; deploy the habit tracker to production and show the live URL when it finishes]

## The domain
- The address you get for free is functional and forgettable.
- A domain is a readable name that points at the same place, rented yearly.
- Buy it from anywhere, point it at your host, wait — the pointing takes
  minutes to hours and there is nothing to do but wait.

[DIAGRAM: domain-and-host]

## Every change is another deploy
- You do not deploy once. You deploy every time you change anything.
- Good hosts do it automatically whenever your version control updates.
- They also keep the previous versions, so undoing a bad deploy is one click.
  Find that button before you need it.

## What to check the first time
- Open it on your phone, not just your laptop.
- Open it signed out, in a private window — this is where "works for me"
  usually dies.
- Ask one person to use it without instructions and watch where they hesitate.

## What's next
- Next lesson: automations — connecting your thing to everything else you
  already use.
