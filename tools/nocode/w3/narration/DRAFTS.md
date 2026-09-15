# Drafts for the top-up lectures

Prose written before the cuts exist, for the beats whose content is known
because we wrote the instruction — the brief, the rewrite, the shape of the
pipeline. **The result beats are deliberately missing.** Nothing in this file
goes into a narration file until the frames have been looked at, and anything
here that the footage contradicts loses.

---

## w3-27 — A second brief, and nobody in the chair

**empty.** Second empty folder of the week, and this one is going to be filled
differently. Everything up to now has had a person in the loop: you type, it
works, it answers, you read it. Today the person leaves the room after the
first two instructions and comes back to a finished repository. So the setup
matters more than usual, because there is nobody to catch a mistake at minute
nine.

**git.** Version control before anything else, and for a sharper reason than
tidiness. In an hour a program is going to write into this folder with no
approval prompts and nobody watching. `git init` here is what makes that a
reversible decision rather than a gamble, and it costs one command.

**brief.** Forty-three lines. Read the shape of it rather than the content:
what the screens are, what the data is, and a short block of rules at the
bottom that are mostly about what *not* to do — no dependencies, no build step,
no timers inside the simulator, no personal data. A brief that only says what
you want produces something that works and that you would not want to maintain.
Half of a good brief is constraints.

**read.** [ANCHOR FROM FRAMES — this is the brief we wrote, so it is safe to
describe; check the scroll position as you go.]

Four screens, and read how they are specified, because it is not the way people
usually write a brief. Not "a watchlist". Thirty instruments; each row a symbol,
a value, a signed percentage against the open, and a sparkline of the last two
minutes; rows re-sorting by the size of the move and animating rather than
jumping. Not "a chart". Candles, one every five seconds, an hour of scrollback,
with rule thresholds drawn across it. Not "alerts". Newest first, each one
saying which rule fired, on what, at what value, how long ago — and clicking one
moves the chart to the moment it fired.

Every one of those sentences is a decision somebody would otherwise make for you
at three in the afternoon without telling you. And then the replay section,
which is the only genuinely hard requirement in the document: the scrubber moves
the whole console back through the last ten minutes, and all four panels follow
it.

Then the rules at the bottom, and this is the half most briefs are missing. Node
twenty-four, no dependencies for the running app, no framework, no build step.
Port eight thousand and eighty-one. The simulator owns no timer — the caller
ticks it. Same seed, same data. Tests exercise real behaviour rather than mocks.
No emojis, no personal data. Split long files. And the one I would put in every
brief I ever write again: do not add a defensive branch that no case reaches.

Read those two halves against each other. The first half is what it does. The
second half is what it is allowed to be. Left to itself an agent will satisfy
the first and invent the second, and you will get a working console with a
framework in it, a build step, four dependencies and a test suite made of mocks.
None of that is disobedience. It is what happens when nobody says otherwise.

**claude.** One agent, pinned to sonnet, in a folder with one file in it.

**plan.** And the instruction is not build this. It is: divide this into areas
that different people could own at the same time without treading on each
other, and for each one name the files it owns, what it must expose to the
others, and what it must not touch. That is the difference between a brief and
a plan, and it is the only piece of work in this lecture that has to be done
with a person in the room — because everything downstream inherits it. Put the
simulator first, because everything else consumes it.

**planread.** [ANCHOR AND TRIM FROM FRAMES — the text below is written from
plan.md and CLAUDE.md as they were actually produced, so it is grounded, but the
order has to follow what is on screen as it scrolls.]

Six areas, and the first line of the document is the reasoning rather than the
list: the simulator is first because every other area consumes its output. Then
the sentence that makes the rest of the week possible — the other five can all
start in parallel once they are building against the *documented* contracts
rather than waiting for finished code from somebody else.

Look at what each area gets. It owns a named list of files. It exposes an
interface written out in full — for the simulator, an actual function signature
and the exact shape of a candle: symbol, time, open, high, low, close, change
percent, one entry per instrument, in a fixed order. And then a section headed
"must not touch", which is the half people leave out. The simulator must not
know what a rule is. It must not decide how much history to keep, because that
belongs to the feed and not to the generator.

And there is one paragraph near the top that is worth more than the six areas
under it: areas talk to each other only through the interfaces listed, nobody
reaches into another area's files, and if a contract has to change, that is a
cross-area edit — update this file in the same change and say so in the commit.
That is not documentation. That is the rule that decides whether six agents
working at once produce a product or a merge conflict.

It also wrote the memory file, and read the first line of it, because it is a
sentence I did not ask for: these come directly from the brief and are not
optional style preferences — code that violates them is wrong, not just
unpolished. Everything under it is a constraint I wrote once in a document,
turned into something that will be read at the start of every session by every
agent that ever works here.

---

## w3-28 — Turning the orchestrator into a real pipeline

**tool.** This is the orchestrator we wrote on Friday, copied into a repository
it has never seen. Eighty-two lines: a list of tasks, one query per task, all
fired at once, a report at the end. And the honest verdict we recorded on it
was that it is good at cheap isolated read-then-write jobs and that it cannot
edit an existing file or run a test. Both of those are about to matter.

**grow.** So read the instruction, because it is a specification and not a
wish. Three phases, each runnable on its own. Plan: one agent reads the plan
and writes a task list — an id, an area, the files it owns, the tools it is
allowed, what it depends on, and a brief written for somebody who cannot ask a
question. Run: execute that list respecting the dependencies, at most three at
once, each in its own session with only its own tools, appending a line to a log
every time something starts, finishes or fails. Review: one agent with
read-only tools inspects what was built against the plan and writes down what is
present, what is missing and what is wrong.

Three things in there are doing all the work. The **dependency list**, because
without it the server starts before the simulator exists and builds against a
guess. The **per-task tool grant**, because that is the only thing standing
between an unattended run and a bad afternoon — there is no permission prompt in
this design, on purpose, since there is nobody to answer one. And **the log**,
because when you come back to a finished repository the log is the only account
of what happened that you did not have to take on trust.

**read.** [ANCHOR FROM FRAMES — written from the file it actually produced.]

Three hundred and sixty six lines, from eighty two. And read down it, because
almost none of the growth is orchestration.

At the top, two constants: the model, named once, and a maximum of three agents
at a time. Then the function that runs one agent, and the comment above it is
the most important thing in the file: permissions are bypassed deliberately,
because this pipeline has no human to answer a prompt — so each task's safety
comes from its tool list being scoped to what that task needs, not from a
dialog somebody would have dismissed anyway. That is the trade, written down,
where the next person will find it.

Then a validator. It checks that every task has the fields it needs, that every
dependency names a task that exists, and that the graph has no cycle in it —
and if any of that fails it refuses to start rather than running half of a
build. Nobody asked for that. It is what happens when you say the tool has to
be debuggable at two in the morning.

Then the scheduler, and it is thirty lines. A task whose dependencies are all
done becomes eligible. A task whose dependency *failed* is not run at all — it
is marked skipped, with the reason written to the log, because building the
server against a simulator that did not get built produces something worse than
nothing. And there is a check for the case where nothing is running and nothing
can start, which throws instead of spinning forever.

And then a table at the end: id, area, status, time, note. That is the whole
interface. No dashboard. One table and a log file.

**dep.** One dependency, and it is not part of the application. The memory file
this repository wrote for itself says zero npm dependencies for the running app,
and that rule is about to be bent in exactly the same deliberate way it was bent
on Friday: the orchestrator is a development tool, not part of the thing that
ships, and the console still starts with no network and nothing beyond Node.

**dep.** One dependency, and it is not part of the application. [FROM FRAMES]

---

## w3-29 — The run

**plan phase.** [FROM FRAMES — how long it took, what it decided]

**tasks.** [FROM FRAMES — the graph it chose, and whether the ordering is right]

**run.** The thing to say over this, while the log scrolls, is what is NOT on
screen. There is no dashboard. No panes to cycle. No approval prompts. Three
agents are writing into this repository at once and the entire interface to that
is a log file and a shell that has not returned yet.

**build.** [ANCHOR FROM FRAMES. These numbers are off the run log, not
estimated — check them against what is on screen before recording.]

Three start at once: the simulator, the watchlist and chart, and the rules and
alerts panels. The two client areas can start immediately because they are
building against the contract in the plan rather than against code — which is
the thing the planning lecture was for, and this is where it pays.

The simulator finishes first, at a minute and a half, and the moment it does the
rules engine starts, because the rules engine declared it as a dependency and
the scheduler was holding it. Panels at two and a half minutes. Chart at three
and a half. Rules engine at five. Then the server — which needed both the
simulator and the rules engine — three and a half minutes. Then the app shell,
which needed everything, six and a half. And the cross-cutting verification pass
at the end, just over a minute.

Seven tasks, all done, none failed, none skipped. Seventeen minutes and
fifty-three seconds of wall clock.

And here is the honest arithmetic, because it is more interesting than the
headline. Add up the seven task times and you get just under twenty-four
minutes. So running three at a time saved about six minutes out of twenty-four,
not two thirds of it. The reason is the chain down the middle: the simulator has
to finish before the rules engine can start, the server needs both of those, and
the shell needs the server. That chain is thirteen minutes long on its own, and
no amount of concurrency shortens it.

Which is worth knowing before you go and build one of these. The speed-up from
running agents in parallel is bounded by the longest chain of things that
genuinely have to happen in order — and that chain is decided in the planning
lecture, not in the orchestrator. If you want a faster build, do not raise the
concurrency limit. Go back and look at why the server needed to wait.

**table.** [FROM FRAMES — the table it prints, and the log]

---

## w3-30 — What a program built

**tree.** The first thing to do with a finished unattended run is not to read
what the agent said about it. It is to read the repository. Git status, the file
list, the test output — those are facts. Everything else is a summary written by
the thing you are checking on, and this week has been one long argument for
knowing the difference.

**size.** [FROM FRAMES — the actual line counts]

**tests.** [FROM FRAMES — how many pass, how many fail. Whatever it is, say it
plainly. A failing suite here is not a problem for the lecture, it is the
lecture: the review phase is about to be asked what is wrong, and having
something genuinely wrong is what makes the answer worth reading.]

**review.** And this is the third phase, and it is the one people leave out. An
agent with read-only tools, that had no part in building any of this, reads what
is on disk against the plan and writes down what is present, what is missing and
what is wrong. Read-only matters: it cannot quietly fix something and then
report that it was fine. It can only tell you.

That is a different job from the tests. The tests answer "does the code do what
the tests say". This answers "does the repository contain what the plan asked
for" — and those two questions have completely different failure modes. A whole
area can be missing with a green suite, because nobody wrote a test for the
thing that is not there.

[The review it produced is 342 lines and found three real things. Verify each
one on camera before saying it — the point of this lecture is not to trust a
review either.]

- The chart is supposed to hold the **last hour**, scrollable. The only
  retention anywhere — a ring buffer in the server, mirrored on the client — is
  capped at **ten minutes**. Nothing in the suite noticed, because nothing in
  the suite tests retention.
- `alerts-panel.js`'s `update` function has quietly **grown a second parameter**
  that the plan does not document. A contract drifted, in a repository whose
  whole premise is that contracts do not drift.
- And the one worth stopping on: **`package.json` carries a dependency**, which
  is a direct violation of the memory file's zero-dependency rule. It is the
  Agent SDK — the orchestrator's own dependency. The reviewer is right, and it
  is right for the reason we already knew on Friday: an exception that is not
  written down is indistinguishable from a mistake. The fix is not to remove it.
  It is to record it, in the file, the way we did in the other repository.

Notice the shape of that third one. An agent with read-only tools, asked to
check the repository against its own stated rules, found the place where we
knowingly broke one and did not say so. That is exactly what you want this phase
for, and it is the reason the reviewer must not be the builder.

And the first one is worth more than it looks, because the bug is not in the
code. Go back to the brief: the chart holds the last hour, scrollable — and the
replay scrubber rewinds through the last ten minutes. Two numbers, in two
paragraphs, that nobody reconciled. The plan carried both of them forward, the
server owner implemented a ten-minute ring buffer because ten minutes is what
their section said, and the chart owner wrote a component that will happily
render an hour it is never given. Every one of those decisions was locally
correct. The defect was written on the first morning, by me, in English.

That is the failure mode of this whole way of working, and it does not look like
bad code. Six agents will each do the locally correct thing with what you gave
them, faster than you can check, and the contradiction only becomes visible when
something reads the whole repository at once. Which is what this phase is.

---

## w3-31 — First run, and what is wrong with it

**start.** One command. And the only question that has mattered since Monday:
does it run.

**alive.** [Verified against frames of the running console, 2026-09-01. Check
again before recording.]

Four panels, dark, exactly the shape the brief asked for. Down the left, the
watchlist: about seventeen instruments visible, each one a symbol, a live value,
a signed change against its open in red or green, and a sparkline that is
drawing itself as we watch. Look at the order — KLN at the top on minus three
and a half, ALP just under it on plus three and a half. It is sorted by the size
of the move, not the direction, and it re-sorts itself every few seconds. Nobody
demonstrated that; it is in the brief and it is on screen.

In the middle, candles, one every five simulated seconds. On the right, the rule
editor: a symbol box, a dropdown that already knows the four kinds of rule the
engine can parse, a threshold, and a button. Under the chart, a scrubber, with
the word LIVE next to it. Along the bottom, the alerts panel, empty, because
there are no rules yet and therefore nothing has fired.

And clicking a row does what it should: the chart follows the selection.

**honest.** [Verified defects. Each one is visible in the frame, so point at it
rather than describing it.]

- **The chart does not fill its panel.** The candles are jammed into the left
  edge and about eighty percent of that panel is empty. It is drawing at a fixed
  candle width and letting the rest be blank, so it will look correct in an hour
  and looks broken now. Nothing in the brief said "fit the width", and nothing
  in the test suite could have caught it.
- **The alerts panel has no empty state.** A large black band with nothing in it
  reads as broken rather than as quiet. One line of text would fix it. Nobody
  wrote the line because nobody was asked to.
- **The watchlist shows seventeen of thirty** with no visible indication that
  there are more below.

And the one you cannot see, which the review found and the screen cannot: the
chart is supposed to scroll back an hour and only ten minutes of history exists
anywhere in the system.

Four defects. None of them is a crash, none of them would fail a test, and every
one of them is the kind of thing a person notices in four seconds and a suite
never notices at all. That is the division of labour now.

**honest.** [FROM FRAMES — the defects, named. Do not soften them and do not
manufacture them. If it looks good, say it looks good and name what is missing
against the brief instead: the brief asked for a replay scrubber, rows that
animate to their new position, and alerts that move the chart to the moment
they fired. Check each one on camera.]

The framing for this beat, whatever the outcome: a program built this in
eighteen minutes with nobody watching, and the interesting question is not
whether that is impressive. It is what it got wrong, because that is the part
you would have to find and fix in a real project, and finding it is now your
whole job.

---

## w3-32 — The fix pass

**hand.** The review is a list. Lists are what agents are good at. So the last
step of the pipeline is not a phase in the orchestrator at all — it is a person
reading the review, deciding it is worth acting on, and handing it back.

And notice the shape of the instruction: work through it highest impact first,
actually edit the files rather than describing the fix, run the tests after each
one, and when you stop, write down what you changed and what you deliberately
left. That last clause is the one that stops a fix pass turning into a rewrite.

**what it actually did** [verified in FIXED.md and on disk, 2026-09-01 — this is
the best beat in the lecture and it was not planned]

It took the review's three findings highest-impact first, and the first one it
reached was the dependency. The rule in the memory file says zero dependencies.
So it removed the dependency. It deleted the devDependencies block, deleted
node_modules, deleted the lock file, verified that the console still starts and
the suite still passes, and wrote down that it had done so.

Which is correct. It is also the thing that makes `tools/orchestrator.mjs` — the
program that built this entire repository forty minutes ago — no longer able to
run.

Sit with that, because it is not a mistake by anybody. The rule said zero
dependencies. The exception was real, and I knew about it, and I said so out
loud in the last lecture, and I did not write it in the file. So an agent
reading only what is written found a violation, and fixed it exactly as
instructed, and the cost of my not writing one sentence down is that the tool is
gone.

That is the whole of this week in one incident. Not a model doing something
strange. A model doing precisely what the written standard said, at a moment
when the written standard was incomplete — and doing it faster than anybody
could have caught it.

The second finding is the one I would frame, though. The chart was supposed to
hold an hour and the buffer held ten minutes, and rather than change one number
it went and read the brief again, worked out that the ten-minute figure belonged
to *replay* and the hour belonged to the *chart*, split the single constant into
two named ones, updated the retention, updated the test to match the corrected
contract — and edited plan.md to record that the contract had moved. Which is
the rule the plan itself set on the first morning: if a contract turns out to be
wrong, fix it here, in the same change.

**verify.** [FROM FRAMES — the suite, the commit, the second look]

---

## w3-33 to w3-35 — the loop that runs on somebody else's computer

**Framing for w3-33.** Everything this week has happened on one laptop. That is
fine for learning and useless as a way of working, because the only record of
what an agent did is a folder you happen to have. The three things that fix
that are not agent features at all: a remote you can push to, a test suite that
somebody else's computer runs, and a review step that happens before code lands.

**w3-33, written from what actually happened. Anchor when the cut exists.**

*local.* One commit, on one laptop, in a folder that exists on one disk. That is
where the second product has been for the whole of the last hour, and it is fine
for learning and useless as a way of working - because the only record of what
seven agents did is a directory you happen to have.

*create.* One command. Create the repository, use this directory as its source,
add the remote, push. Private, and inside an organisation rather than under a
personal account - which is a habit worth having for its own sake, but here it
is also because almost every command from now on will print the owner on screen.

*ci.* Now the part that is not about agents at all. The instruction is: run on
push to main and on every pull request, one job, check out the code, set up Node
twenty-four, run the suite. And two negative clauses, which are the ones that
make it short: the app has no dependencies, so do not add an install step it
does not need, and do not add a cache for a lockfile that may not exist. Left to
itself an agent will write you the workflow it has seen most often, which has an
npm ci in it and a cache keyed on a lock file, and both of those will fail on
this repository for the same reason - the thing they are for does not exist.

*read.* Twenty lines, three steps, and a comment above each one saying what it
is for. Checkout, so later steps have code. Node twenty-four, matching the
engines field in the package file. And npm test, with the reason for the absent
install step written down next to it. That last comment is the one that matters
in six months, when somebody adds a dependency and wonders why nothing installs
it.

*push.* And now something genuinely different happens, and it is worth being
precise about what. Up to this point every check in this course has run on this
machine, which means every check has been run by somebody who wanted it to pass.
This one runs on a computer that has never seen this project, from a clean
checkout, with nothing cached and nothing left over.

*green.* Fifteen seconds, and green. And the useful thing about that is not the
colour - it is that the suite now runs whether or not anybody remembers to run
it. Everything else this week has depended on a person choosing to check. This
does not.

**Framing for w3-34.** An issue is a brief with an address. That is the whole
idea. Every brief we have written this week was typed into a terminal and lost
when the session ended. An issue is the same words, kept, numbered, and
readable by somebody who was not in the room — which means it can be handed to
an agent, or a colleague, or to you in three weeks.

**w3-34, from what happened.** The instruction to the first agent is not "find a
bug". It is: read the review and the fix log, pick the single most worthwhile
thing still not done, small enough to fix in one change and real enough to
matter, and file it. Then stop. Do not fix anything.

That last sentence is doing something specific. An agent that finds a problem
will fix it, because fixing is what it is for, and then the issue is a
description of work already done and the whole exercise is theatre. Separating
the finding from the fixing is not bureaucracy - it is what makes the issue an
artefact somebody else could act on.

And what it found is a good one: the alerts array in the feed grows without
limit, while the candle history next to it is trimmed on every tick to a fixed
window. One of those two was thought about and the other was not. That is
exactly the kind of defect that survives a green suite and a careful review of
each area separately, because neither half is wrong - the inconsistency between
them is.

Read the body of the issue rather than the title. What is wrong, where, with the
actual lines quoted. How to see it. And what done looks like. A stranger could
act on that, which is the only test an issue has to pass.

*hand.* And now the second agent gets the address, not the work. Read issue one.
Create a branch named for it. Write a test that fails because of the bug and
would pass once it is fixed - and run it, and watch it fail, before touching the
code. Then fix it, run the suite again, commit with a message that explains the
change rather than restating the diff, push, and open a pull request that closes
the issue. And if the fix turns out to be bigger than the issue described, say
so in the pull request rather than quietly widening it.

That is not an agent instruction. That is how you would want a colleague to work,
written down. The only reason it has to be written down is that this colleague
has never worked anywhere.

**Framing for w3-35.** The pull request is where the week's argument lands. We
have spent five days saying that reading is the job now. This is the interface
that was designed for exactly that, twenty years before any of this existed: a
diff, a description, and a check that does not care how good the description
was.

*diff.* And this is the beat to slow down on, because it is the one people skip.
Eighty lines. Read the test first, then the change, in that order - because the
test tells you what the author believed the bug was, and if that belief is wrong
the fix is wrong however good it looks. Then ask the two questions that catch
most of it: is anything here not mentioned in the issue, and is anything in the
issue not here.

*ci.* Then the part that does not care how good the explanation was. It is worth
saying plainly why this matters more now than it did before agents: the volume
of plausible code you are asked to review has gone up by an order of magnitude,
and your capacity to read it has not moved at all. A check that runs itself is
the only part of this that scales with the thing that changed.

---

## w3-36 to w3-38 — the agent with no screen

**w3-36.** *tui.* Every session this week has had a face on it. A banner, a
spinner, a box you type into, a status line telling you what it is thinking
about. That interface is genuinely good, and it has quietly taught you something
untrue: that using an agent means sitting in front of one.

*p.* Same binary, same login, same tools. Minus p, a prompt, and the screen taken
away. It reads the repository, answers in one line, and exits. And notice what
did not happen: no banner, no session, nothing to close, and nothing to approve -
because there is nobody to approve it, and the tool list decided instead.

*exit.* And it has an exit code, like any other command. That single fact is the
whole lecture. An exit code is what lets this appear in a shell script, a
Makefile, a git hook, a scheduled job - anywhere a program can go. Everything
else here is detail.

*json.* Then ask for structure instead of prose. Output format json, and what
comes back is an object: the result, how many turns it took, how long it took,
what it cost. Pipe that through jq and you have a program's answer rather than a
person's. The cost field is the one to keep - it is the only place in this whole
week where the bill is machine-readable.

*tools.* And the safety model, which is not a permission prompt because there is
nobody to prompt. Ask it to delete the test directory, with read and glob and
nothing else, and it cannot. Not "declines to" - has no tool that does it. That
is the same idea as Tuesday's container, one level up: the boundary is what is
in the list, not what the agent decides.

**w3-37.** *pipe.* It reads standard input, which means it goes in a pipe, which
means everything you already know about shell composition applies unchanged. A
diff on the left, a reviewer on the right. And the instruction is written to make
flattery difficult: only defects that are real and specific, one per line, with
the file and what goes wrong; if there are none, reply with the single word NONE;
do not praise anything. Without those clauses you get three paragraphs about how
well-structured it is.

*chain.* Two of them in a row is a pipeline. The first names the three riskiest
files for a newcomer, path only, no commentary - into a file. The second reads
that file and says what a newcomer would get wrong in each. Neither knows the
other exists. The connective tissue is a text file and a pipe, which is the
oldest idea in this entire course.

*point.* And that is the thing to take from this lecture. Not that the CLI has a
minus p flag. That an agent with an exit code and a standard input is a unix
program, and can be composed with every other unix program you have, by you,
today, without anybody shipping an integration.

**w3-38.** *docs.* And none of this was ever about code. Twelve PDFs in a folder.
Receipts - vendors, dates, line items, totals - the sort of thing that arrives
attached to an email and gets typed into a spreadsheet by somebody on a Tuesday.

*one.* One document, one line of structured output, and the schema in the prompt:
vendor, date, currency as a three-letter code inferred from the symbol, total as
a number, items as an integer count. No prose, no code fence. Being that specific
is the difference between output you can parse and output you have to read.

*all.* Then the same thing twelve times, from a shell loop. Not a batch feature,
not a queue, not an integration - a for loop, appending one line of JSON per file.
This is the beat that replaces a whole desktop application, and it is four lines
of shell.

*table.* Twelve PDFs in, one table out, with a total on the bottom. And be honest
about what is and is not proven here: nothing has checked those numbers. What has
been demonstrated is the shape - documents in, structured records out, at whatever
scale the folder happens to be - and the checking is the same job it always was.

*point.* One more thing worth saying about this, because it is the lecture that
generalises furthest. Nothing in that loop knew it was reading receipts. Change
the schema and it reads invoices, or CVs, or survey responses, or the six hundred
markdown files in a documentation site. The agent is not the product here. The
loop is.

**Extra material for w3-36 to w3-38, if the footage carries it.**

*On what -p actually is.* It is not a lite mode and it is not a different
product. It is the same binary, the same login, the same tool set and the same
memory files - the session is just not interactive, so there is nowhere to put a
question. Everything that follows from that is a consequence of one fact: there
is no human in the loop, so anything that would have been asked has to have been
decided in advance, in the flags.

*On the answer it gave.* Three test files, twenty-nine test cases, and it named
the command it ran to find out. That last part was in the instruction on purpose.
An agent that answers a factual question about a repository should be made to say
how it knows, because the difference between counting and remembering is
invisible in the answer and total in the reliability.

*On the exit code.* Zero. Which sounds like nothing and is the reason this whole
lecture exists. A program with an exit code composes: it can be `&&`-ed, it can
gate a deploy, it can fail a build, it can go in a git hook. Everything else in
this course has been a thing you talk to. This is a thing you can put in a
sentence with other programs.

*On the json.* The result, the number of turns, the wall-clock milliseconds and
the cost. Two of those are useful immediately: the turns tell you whether your
prompt was one job or an argument, and the cost is the only place all week where
the bill is a number a program can read. If you are going to run this over a
thousand documents, that field is how you find out what a thousand costs before
you run nine hundred and ninety-nine of them.

*On the tool grant.* Asked to delete a directory, with read and glob and nothing
else, it cannot - and the distinction that matters is that it is not declining.
There is no tool in its hands that removes a file. That is Tuesday's container
argument one level up: the boundary is the list, not the judgement. And unlike
the container, this one costs nothing and takes one flag.

*On the pipe, for w3-37.* Standard input is the oldest interface in computing and
it is why this composes with everything you already have. `git diff` on the left.
A reviewer on the right. No integration, no plugin, no vendor. And the review
instruction is written to make flattery difficult, because the default output of
"review this" is three paragraphs about how well-structured it is: only defects
that are real and specific, one per line, with the file; if there are none, the
single word NONE; do not praise anything.

*On chaining.* The first names the three riskiest files for a newcomer, path only,
no commentary. The second reads that list and says what a newcomer would get
wrong in each. Neither knows the other exists; the connective tissue is a text
file.

And look at what actually landed in that file, because it is the honest half of
this lecture. It says path only, no commentary - and it came back wrapped in a
code fence. Three correct paths with three backticks above and below them. Which
is fine here, because the next stage is another language model and it does not
care. It would not be fine if the next stage were `xargs`, and that is the whole
difficulty with putting these in a pipeline: the output is *almost* structured.
If you are building this for real, the line after the agent is a `grep` or a
`sed` that throws away everything that is not the shape you asked for - which is
exactly what we do two lectures from now with the receipts. That is not a lesser version of a multi-agent framework. For work shaped
like this it is the same thing with nothing to install and nothing to learn.

*On the batch, for w3-38.* Twelve documents, one shell loop, one line of JSON per
file. It is worth counting what is not here: no queue, no worker pool, no retry,
no state. If one of the twelve fails you get eleven lines and a gap, and you run
that one again. For a folder of twelve that is exactly the right amount of
machinery, and knowing that is worth more than knowing how to build the other
kind.

*On what the table actually shows.* Two things are wrong with it and both are
worth stopping on, because they are the two failure modes of this entire
technique.

The first: six of the twelve have a currency and six are blank. Same instruction,
same model, same kind of document - and half of them decided the currency was not
stated and returned null rather than guessing. That is arguably the right
behaviour and it is certainly the honest one, and it means the column is
unusable. If you need a field to be present you cannot ask nicely for it in a
prompt; you check for it afterwards and re-run the ones that are missing.

The second is the total, and it is not the model's fault at all. Two thousand
eight hundred and fifty-six point eight nine, followed by eleven more digits.
That is floating point doing what floating point does, in the jq at the end of my
own pipeline. It is exactly the sort of thing that ships, because the interesting
part - the extraction - worked. The lesson is not about money types. It is that
the moment the agent's part of a pipeline succeeds, your attention moves on, and
everything downstream of it stops being read.

*On honesty about the output.* Nothing has checked those numbers. A total at the
bottom of a table is very convincing and it is only as good as twelve extractions
nobody read. What has been demonstrated is the shape - documents in, records out,
at whatever scale the folder happens to be. The checking is the same job it
always was, and it is now the only job left.

---

## w3-39 to w3-41 — when it is wrong, when it went wrong, and what each is for

**w3-39.** *state.* Start from evidence, not from a feeling. The history, then
the suite. And the reason to run the suite yourself before asking anybody about
it is that you want to know the answer before you hear an explanation of it -
otherwise the explanation is doing the work your eyes should be doing.

*ask.* Then hand over the evidence, not the conclusion. Read the failing test.
Read the code it exercises. Say which of the two is wrong and why. Show me the
evidence before proposing a change. That last clause is the entire lecture: an
agent asked to fix something will fix something, and the thing it fixes will be
whatever makes the symptom go away. Asked to diagnose first, it has to commit to
a claim you can check.

[If the suite is green when this rolls - and it may well be - the beat becomes
the same technique pointed at the review instead: take the strongest claim in
REVIEW.md, check it against the code, and say whether the review was right. Say
on camera that the suite is green. Do not manufacture a failure.]

*fix.* And only then change something. The smallest change that fixes what you
just described, and nothing else, with the suite before and after both on
screen. If the fix does not work, say so rather than trying a second thing on top
of the first - which is how a twenty-minute debugging session becomes an
afternoon and a diff nobody can read.

**w3-40.** *when.* Here is the question nobody asks an agent, and it is the one
that most often ends the argument: not what is wrong with this, but when did it
stop being right. Those have completely different answers and only one of them
tells you why.

*command.* And the first move is not git at all. It is: write me a one-line
command that exits zero when this behaviour is right and non-zero when it is
wrong. Everything after that is mechanical. Without it, bisecting is a person
squinting at each revision and deciding, which is slow and inconsistent and
exactly the thing a machine should be doing.

*bisect.* Then the search. Git has had this for twenty years and almost nobody
uses it, because writing the test command was always the hard part - and that is
the part that just got cheap. Ten commits is three or four checkouts. A thousand
is ten.

*point.* And notice what this gives you that a fix does not. A fix makes the
symptom go away. A first bad commit tells you what somebody was trying to do when
they caused it, which is usually the difference between patching a behaviour and
understanding a design.

**w3-41.** *two.* Two whole products, in two directories, built two different
ways. Control Tower, by an interactive agent team with a person in the room.
Signal Desk, by a program, with nobody watching. Same model, same week, same
person writing the brief.

*numbers.* Commits, files, lines, counted the same way on both sides - and that
is a deliberate small thing. Two comparisons measured differently is worse than
no numbers at all, and it is the easiest way to accidentally lie in a table.

*verdict.* And then the instruction, which is written to make a winner difficult:
read both repositories, compare them on what each delivered against its own plan,
where each went wrong, what each cost in wall-clock time, and which you would
choose for what kind of work. Ground every claim in a file you read. Where the
evidence is missing, say it is missing. And do not decide which is better in
general - decide what each is for.

That last clause is there because "which is better" has no answer and every
comparison video pretends it does. The team could ask, argue, and route around a
surprise; the program could not, and does not need to be awake. Those are not
two grades of the same thing.

---

*merge.* Squash, delete the branch, and the issue closes itself because the pull
request said it would. And that is the loop. An issue, a branch, a test, a
change, a review, a check, a merge. Every step of it existed before any of this,
and none of it had to be invented for agents - which is the point. You do not
need a new process. You need the one you already had, actually used.


---

## w3-40, from what actually happened

*log.* Seven agents, an hour of work, nineteen hundred lines - and how many
commits. One. Then two, once the CI workflow landed, and three after the pull
request merged. All of the building is in the first one.

*cost.* And that is not a tidiness complaint. Ask what it costs, in commands
that do not work. `git log` on a file tells you nothing, because every file
arrived at the same instant. `git blame` attributes nineteen hundred lines to one
commit made by me, which is false about all of it. `git bisect` has nothing to
search. `git show` on the commit that introduced a behaviour shows you the whole
product. Every one of those is a question you will want to ask in three weeks and
cannot.

*fix.* And the fix is in the runner, not in the briefs, which is the interesting
part. You could tell each task to commit its own work - and then a task that
fails halfway commits half of itself, and a task that touches a file another task
owns commits somebody else's change with it. The runner already knows when a task
started, when it finished, and whether it succeeded. It is the only thing that
does. So the commit belongs there: after a task reports done, stage what changed
and commit it in that task's name.

*exception.* And then the sentence that should have been there this morning. The
memory file says zero dependencies. The orchestrator needs one. I knew that, I
said it out loud in the lecture where we installed it, and I did not write it
down - so an agent reading only what was written deleted the dependency, exactly
as instructed, and left the program that built this repository unable to run.

Read what it wrote, because it is better than what I would have written. It names
the file. It says why that file is different - a development tool invoked by a
person, not code the running app loads. It states the boundary: nothing under
server or public, and nothing `npm start` reaches, may import from it. And then
the clause I did not ask for and would not have thought of: this does not reopen
the door in general, and a new development dependency needs the same
justification this one has, not "it was convenient."

That is what a written standard looks like when it is finished. Not a rule - a
rule and its exception and the reason the exception does not generalise. And the
whole cost of not having it was one deleted package and one tool that stopped
working, on a day when somebody happened to be watching.


---

## w3-39, from what actually happened

*state.* Start from evidence rather than from a feeling. The history first, then
the suite - and run the suite yourself before you ask anybody about it, because
you want to know the answer before you hear an explanation of it. Read the two
in that order every time. If you read the explanation first you will find
yourself confirming it, and you will not notice that is what you are doing.

*green.* And it is green. Twenty-nine tests, all passing, which is not what this
lecture was going to be about. So say so, and take the other branch: if there is
no failure to diagnose, take the strongest claim in the review and check whether
it is still true.

That is worth doing on any project, not just this one. A review is a snapshot.
The moment somebody acts on it, half of it is stale - and there is nothing in the
document that tells you which half.

*ask.* Read the instruction, because the shape of it is the transferable part.
Take the claim. Check it against what is actually on disk right now. Say whether
the review was right. Show me the evidence before you conclude anything. And do
not change anything yet.

That last clause is the one that keeps this honest. An agent asked to check
something and fix it will fix it, and then report that everything is fine - which
is true, and useless, because you never find out what state you were actually in.

*result.* And what it did is the right answer: it went to the file, confirmed the
dependency really is gone from the package file and off the disk, and then
annotated the review itself with the evidence and the date. So the review is no
longer a snapshot of this morning. It is a document that says what was true then
and what is true now, and which of its own findings have been dealt with.

Which is the small habit worth stealing from this whole lecture. When you act on
a document, write the outcome back into it. Reviews, issues, plans, memory files
- the ones that stay useful are the ones somebody amends, and the ones that
become fiction are the ones nobody ever touched again.

---

## w3-41, from what actually happened

*two.* Two products, in two directories, built two different ways. Control Tower,
by an interactive agent team with a person in the room. Signal Desk, by a program
we wrote, with nobody watching. Same model. Same week. Same person writing the
brief.

*numbers.* Commits, files, lines - counted the same way on both sides. That is a
deliberate small thing: two things measured differently is worse than no numbers
at all, and it is the easiest way to accidentally lie in a table.

*verdict.* And then the instruction, written to make a winner difficult. Read
both repositories - the plans, the run logs, the review, the test output, the
history. Compare them on four things: what each delivered against its own plan,
where each went wrong, what each cost in wall-clock time, and which you would
reach for, for what kind of work. Ground every claim in a file you read. Where
the evidence is missing, say it is missing. And do not decide which is better in
general.

That last clause is there because "which is better" has no answer, and every
comparison you have ever watched pretends it does. The team could ask, argue,
change its mind and route around a surprise. The program could do none of that
and does not need anyone to be awake. Those are not two grades of the same thing,
and a verdict that ranks them is a verdict that has stopped paying attention.

**And what it produced is better than the lecture I had planned.** Fourteen
thousand words, every claim carrying a file path and a line number, and it
refuses the overall winner exactly as instructed - and then finds two things
about our own week that we did not know.

*On time.* The program's build has a machine-generated record: seventeen minutes
and fifty-two seconds of unattended execution for six areas, and fifty-one
minutes from first commit to last including planning, review and the fix pass.
The team's has no equivalent, because nothing logged it - only three commits
spanning six hours and fourteen minutes. And it says plainly that those two
numbers are not comparable: one is machine-measured agent execution, the other is
wall clock across a supervised session with a person thinking in it. Then it says
what neither repository records at all, which is what either of them cost in
money. That is a gap in our own work, found by reading it.

*On what went wrong.* Two findings about Control Tower that nobody in that
six-hour session caught. The finished work is still sitting on an unmerged
branch. And the Agent SDK dependency went into the wrong field of the package
file. Both are the sort of thing a review pass catches in thirty seconds, and
there was no review pass, because a person was in the room and a person in the
room feels like one.

*The conclusion,* which is not the one I expected and is better than the one I
would have written: a person being present did not substitute for a review. It
substituted for some of one, and missed things the program's automated review
step is built to catch. So the recommendation is not to pick a side. It is to
take the one piece of process the unattended pipeline has - a read-only agent
that checks the repository against the plan when the work stops - and bolt it
onto the interactive way of working too.

Which is the whole week, arriving as a conclusion rather than as advice. The
thing that made the unattended build trustworthy was not the agents. It was that
somebody wrote down what done looks like, and then made something check.


---

## w3-32, second half (capture E2) - the console actually used

*using.* And now the part w3-31 could not do, because there was nothing worth
touching yet. The fixes are in, the suite is green, and the console has been
running for a few minutes - so this time we use it rather than look at it.

*sort.* Watch the list re-order itself. Biggest mover at the top, direction
ignored, re-sorting every few seconds. That is one sentence of the brief, four
hours ago, and nobody demonstrated it or tested it.

*pick.* Click a row and the chart follows. Which is shared state that two
separately built panels both agreed about - the watchlist and the chart were
written at the same time by different agents who never saw each other's code, and
the only thing making them agree is a container id and a function signature
written in a plan.

*rule.* Now write a rule, in the sentence the engine understands. A symbol, a
kind, a threshold. And this is the seam I was least confident about all day,
because it is the only place in the product where two areas had to agree on
something neither of them owns: the panel has to emit sentences the parser can
read, and they were built in parallel by agents who never spoke. ALP crosses
above four hundred.

*second.* And a second one of a different kind, to check the dropdown is not
decoration. KLN crosses below seventy.

*wait.* Then leave it alone. Which is the honest way to test an alerting system
and the one nobody films, because for a minute or two nothing happens and nothing
is supposed to.

*alert.* And there it is. ALP crosses above four hundred, fired at four hundred
and eight point nine one. That is the whole product working end to end, in one
line: a simulator that ticks, a feed that retains, a rules engine that parses a
sentence a person typed into a box, an evaluator that noticed a crossing, and a
panel that rendered it - five of the six areas, built in parallel by five agents
in eighteen minutes, meeting for the first time.

[Click the alert and say honestly whether the chart moves to the moment it fired.
That is the requirement; if it does not, that is the finding.]

*replay.* Then the scrubber, which is the hardest single requirement in the
brief, because moving it has to move all four panels together. [Narrate what
actually happens. Do not describe the requirement as though it were the result.]

*theme.* And the light one, because the brief asked for both and a light theme
that is an inverted dark theme is a thing you can see in a second.
