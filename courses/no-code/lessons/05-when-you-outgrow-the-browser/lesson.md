---
title: "When you outgrow the browser"
outcomes:
  - Recognise the three signs a project has outgrown an app builder
  - Understand what an AI coding agent is actually doing to your files
  - Watch an agent read a real project and change it
diagrams:
  - id: builder-vs-agent
    kind: mermaid
    prompt: "A two-column comparison flowchart. Left column headed 'App
      builder' has nodes 'Runs in a browser', 'Owns the whole project',
      'One conversation'. Right column headed 'AI editor' has nodes 'Runs on
      your machine', 'Edits real files you keep', 'Reads the whole project
      before answering'. The columns are not connected."
  - id: what-the-agent-does
    kind: mermaid
    prompt: "A left-to-right chain of four nodes with arrows: 'Reads your
      request', 'Searches the project for what is relevant', 'Edits the files
      it decided to change', 'Runs it to check'. Straight chain, one arrow
      between each pair."
---

# When you outgrow the browser

## Three signs it is time
- You are describing the same context every session because the builder keeps
  losing it.
- You want something the builder will not do — a specific library, an unusual
  integration, a file laid out your way.
- Somebody else needs to work on it with you.
- None of these are about difficulty. They are about ownership.

## What actually changes
- The project stops being inside somebody's product and becomes files on your
  computer, which you keep.
- You can put those files somewhere safe, hand them to a developer, or move to
  a different tool entirely. None of that is possible from inside a builder.
- The cost is ceremony: things to install, a terminal, a few new words.

[DIAGRAM: builder-vs-agent]

## What an AI coding agent is doing
- It is not autocomplete and it is not a chat window that happens to know code.
- Given a request, it searches your project for what is relevant, decides which
  files to change, makes the edits, and can run the thing to see if it worked.
- That last step is the one that matters. It checks its own work.

[DIAGRAM: what-the-agent-does]

## Watch it happen on a real project
- What follows is a genuine recording. Nothing is staged, and nothing is cut.
- The agent's thinking time is condensed so it fits — the corner of the frame
  says how long it really took and how long you are watching. Nothing is hidden
  by the speed-up; you see every step, just not every second of waiting.
- The project is the habit tracker from lesson three, exported to files.
- The request is a small feature that touches more than one place — which is
  exactly where a builder starts struggling and an agent starts helping.

[CAPTURE: tool=claude, fixture=habit-tracker; ask the agent to add a weekly summary to the habit tracker, then show the files it changed]

## Reading what just happened
- It looked at the project before it wrote anything. That is why it put the new
  code in the right place instead of a new file of its own.
- It changed more than one file, because a feature usually is more than one file.
- It told you what it did. Read that part. It is the only record of what
  changed if you do not check the files yourself.

## The habits that keep this safe
- Keep your project in version control from day one — lesson nine sets this up,
  and it is what makes "undo everything since Tuesday" possible.
- Read the summary of what changed, even if you do not read the code.
- Ask it to explain anything you do not recognise. It is not a stupid question
  and the explanation is free.

## What's next
- Next lesson: design. Making the thing look like somebody chose how it looks.
