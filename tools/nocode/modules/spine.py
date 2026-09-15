#!/usr/bin/env python3
"""The course spine: the 98 delivered lectures, in upload order, with the
block each one belongs to and a SHORT original title for the chapter cards.

Two things this file exists to fix.

**Order.** `MANIFEST.md` warns that filename numbers are production numbers with
gaps in them, and Week 3's upload order is nothing like its filename order
(w3-36..38 sit between w3-13 and w3-15). The order here IS the upload order.

**Titles.** Weeks 1-2 carry the reference course's lecture titles in both the
manifest and the file metadata - `QA.md` says he should rewrite them. A chapter
card needs a short title anyway (a row is ~42 characters before it wraps), so
these are written fresh from the films and are ours.

`block` is what a module may never straddle. For Weeks 1-2 that is the teaching
Day, because the course is taught in days. Week 3 has no days in the manifest -
it runs as one continuous build - so its blocks are the thematic runs the films
actually fall into, chosen so they also pack near 25 minutes.
"""
from __future__ import annotations

# (file stem, week, block key, block title, short original title)
SPINE = [
    # ---- Week 1 ---------------------------------------------------------
    ("nocode01", 1, "w1d1", "Day 1", "Welcome: a 3D game in an hour"),
    ("nocode02", 1, "w1d1", "Day 1", "The agent builds a shooter"),
    ("nocode03", 1, "w1d1", "Day 1", "The missing manual for agentic coding"),
    ("nocode04", 1, "w1d1", "Day 1", "Your instructor, and the three-week map"),
    ("nocode06", 1, "w1d1", "Day 1", "Vibe coding, agents, and coding agents"),
    ("nocode07", 1, "w1d1", "Day 1", "The eight stages of AI coding"),
    ("nocode08", 1, "w1d1", "Day 1", "Wrapping up day one"),

    ("nocode09", 1, "w1d2", "Day 2", "How an LLM actually works"),
    ("nocode10", 1, "w1d2", "Day 2", "Tools, loops, and what an agent is"),
    ("nocode11", 1, "w1d2", "Day 2", "Context engineering and the system prompt"),
    ("nocode12", 1, "w1d2", "Day 2", "Writing agents.md"),
    ("nocode13", 1, "w1d2", "Day 2", "From YOLO mode to Ralph loops"),
    ("nocode14", 1, "w1d2", "Day 2", "Comparing the models on the numbers"),

    ("nocode15", 1, "w1d3", "Day 3", "Four IDEs, one build, hands on keys"),
    ("nocode16", 1, "w1d3", "Day 3", "agents.md and the settings that matter"),
    ("nocode17", 1, "w1d3", "Day 3", "A Kanban app in Cursor, YOLO mode"),
    ("nocode18", 1, "w1d3", "Day 3", "The same app in Copilot and VS Code"),
    ("nocode19", 1, "w1d3", "Day 3", "Codex takes it zero-shot"),
    ("nocode20", 1, "w1d3", "Day 3", "Antigravity, and Gemini 3 Pro"),
    ("nocode21", 1, "w1d3", "Day 3", "The verdict on the four"),

    ("nocode22", 1, "w1d4", "Day 4", "Picking the model for the job"),
    ("nocode23", 1, "w1d4", "Day 4", "Five principles: be the boss"),
    ("nocode24", 1, "w1d4", "Day 4", "Responsible YOLO: setting up OpenRouter"),
    ("nocode25", 1, "w1d4", "Day 4", "A Next.js site with Codex in Cursor"),
    ("nocode26", 1, "w1d4", "Day 4", "A digital twin, wired to OpenRouter"),
    ("nocode27", 1, "w1d4", "Day 4", "Code review across two models"),

    ("nocode28", 1, "w1d5", "Day 5", "Karpathy's rules, and a real MVP"),
    ("nocode29", 1, "w1d5", "Day 5", "Web apps 101: front, back, and Docker"),
    ("nocode30", 1, "w1d5", "Day 5", "Standing up a full-stack project"),
    ("nocode31", 1, "w1d5", "Day 5", "Planning and scaffolding with Copilot"),
    ("nocode32", 1, "w1d5", "Day 5", "The Kanban app: Docker and FastAPI"),
    ("nocode33", 1, "w1d5", "Day 5", "Debugging drag and drop"),
    ("nocode34", 1, "w1d5", "Day 5", "The app is done: week one wrap"),

    # ---- Week 2 ---------------------------------------------------------
    ("nocode35", 2, "w2d1", "Day 1", "From vibe coding to vibe engineering"),
    ("nocode36", 2, "w2d1", "Day 1", "Installing Claude Code"),
    ("nocode37", 2, "w2d1", "Day 1", "First run: init, context, a test"),
    ("nocode38", 2, "w2d1", "Day 1", "Fixing a hallucination, then refactoring"),
    ("nocode39", 2, "w2d1", "Day 1", "OpenCode, and the free models"),
    ("nocode40", 2, "w2d1", "Day 1", "AMP, OpenRouter, and Ollama"),

    ("nocode41", 2, "w2d2", "Day 2", "Commands, shortcuts, configuration"),
    ("nocode42", 2, "w2d2", "Day 2", "Sessions, checkpoints, and git"),
    ("nocode43", 2, "w2d2", "Day 2", "Rewind: undoing what the agent did"),
    ("nocode44", 2, "w2d2", "Day 2", "Bypassing permissions on purpose"),
    ("nocode45", 2, "w2d2", "Day 2", "Ralph loops in Claude Code"),

    ("nocode46", 2, "w2d3", "Day 3", "MCP, skills, plugins: the big three"),
    ("nocode47", 2, "w2d3", "Day 3", "What an MCP server really is"),
    ("nocode48", 2, "w2d3", "Day 3", "Adding MCP servers: Context7, Polygon"),
    ("nocode49", 2, "w2d3", "Day 3", "Skills: the simpler way to add ability"),
    ("nocode50", 2, "w2d3", "Day 3", "Skill marketplaces, and agent browser"),
    ("nocode51", 2, "w2d3", "Day 3", "Plugins: the marketplace and the rules"),
    ("nocode52", 2, "w2d3", "Day 3", "MCP vs skills vs plugins: choosing"),

    ("nocode53", 2, "w2d4", "Day 4", "The workflow: Claude Code, Jira, MCP"),
    ("nocode54", 2, "w2d4", "Day 4", "Wiring Claude Code to Jira"),
    ("nocode55", 2, "w2d4", "Day 4", "The GitHub MCP server"),
    ("nocode56", 2, "w2d4", "Day 4", "From a Jira issue to a pull request"),
    ("nocode57", 2, "w2d4", "Day 4", "A whole Next.js app from one ticket"),
    ("nocode58", 2, "w2d4", "Day 4", "A disciplined way to debug"),

    ("nocode59", 2, "w2d5", "Day 5", "A SaaS build: skills and claude.md"),
    ("nocode60", 2, "w2d5", "Day 5", "Writing claude.md, building a skill"),
    ("nocode61", 2, "w2d5", "Day 5", "Filing the tickets for V1"),
    ("nocode62", 2, "w2d5", "Day 5", "Building the features"),
    ("nocode63", 2, "w2d5", "Day 5", "Testing the legal doc generator"),
    ("nocode64", 2, "w2d5", "Day 5", "Final merge, full demo, week two"),

    # ---- Week 3 ---------------------------------------------------------
    # Blocks here are thematic runs, not days. w3-36..38 really do belong
    # between w3-13 and w3-15 - that is the upload order in MANIFEST.md.
    ("w3-01", 3, "w3a", "Configuring the agent", "What changes when you stop typing"),
    ("w3-02", 3, "w3a", "Configuring the agent", "The brief: Control Tower"),
    ("w3-03", 3, "w3a", "Configuring the agent", "Slash commands: a routine in one word"),
    ("w3-04", 3, "w3a", "Configuring the agent", "Sub-agents: specialists inside an agent"),
    ("w3-05", 3, "w3a", "Configuring the agent", "Hooks: rules that fire by themselves"),

    ("w3-06", 3, "w3b", "Packaging and sandboxing", "Plugins: packaging your whole setup"),
    ("w3-07", 3, "w3b", "Packaging and sandboxing", "Why sandboxing is the unlock"),
    ("w3-08", 3, "w3b", "Packaging and sandboxing", "YOLO without the fear: the container"),

    ("w3-12", 3, "w3c", "Scale, and the SDK", "A big codebase: the seven rules"),
    ("w3-13", 3, "w3c", "Scale, and the SDK", "Driving Claude from code: the SDK"),

    ("w3-36", 3, "w3d", "No screen at all", "No screen at all: the print flag"),
    ("w3-37", 3, "w3d", "No screen at all", "Agents in a pipeline"),
    ("w3-38", 3, "w3d", "No screen at all", "A stack of documents, not a codebase"),

    ("w3-15", 3, "w3e", "Agent teams", "Sub-agents vs agent teams"),
    ("w3-16", 3, "w3e", "Agent teams", "Clean house, and pick the roster"),
    ("w3-17", 3, "w3e", "Agent teams", "Launch: seven agents, one task list"),

    ("w3-18", 3, "w3f", "What the team did", "What the swarm actually did"),
    ("w3-19", 3, "w3f", "What the team did", "Are these tests worth anything?"),
    ("w3-20", 3, "w3f", "What the team did", "First run: Control Tower is alive"),
    ("w3-23", 3, "w3f", "What the team did", "Roll your own orchestrator"),
    ("w3-24", 3, "w3f", "What the team did", "The verdict on running a team"),

    ("w3-27", 3, "w3g", "Nobody in the chair", "A second brief, nobody in the chair"),
    ("w3-28", 3, "w3g", "Nobody in the chair", "Turning it into a real pipeline"),

    ("w3-29", 3, "w3h", "The unattended run", "The run: no terminal to watch"),

    ("w3-30", 3, "w3i", "What a program built", "What a program built"),
    ("w3-31", 3, "w3i", "What a program built", "First run, and what is wrong"),
    ("w3-32", 3, "w3i", "What a program built", "The fix pass"),

    ("w3-33", 3, "w3j", "Off your laptop", "The repo, and its CI"),
    ("w3-34", 3, "w3j", "Off your laptop", "An issue is a brief with an address"),
    ("w3-35", 3, "w3j", "Off your laptop", "The pull request"),
    ("w3-39", 3, "w3j", "Off your laptop", "Diagnosing from the evidence"),

    ("w3-40", 3, "w3k", "Two products, and a wrap", "The history the pipeline lost"),
    ("w3-41", 3, "w3k", "Two products, and a wrap", "Two products, honestly compared"),
    ("w3-25", 3, "w3k", "Two products, and a wrap", "Into a container, with none of you"),
    ("w3-26", 3, "w3k", "Two products, and a wrap", "Wrap: from no-code to director"),
]

WEEK_TITLE = {
    1: "Vibe Coding for fun and profit",
    2: "Vibe Engineering as a professional",
    3: "Vibe Engineering as an expert",
}

def check() -> None:
    stems = [s[0] for s in SPINE]
    assert len(stems) == len(set(stems)) == 98, f"{len(stems)} rows, {len(set(stems))} unique"
    long = [(s[0], s[4], len(s[4])) for s in SPINE if len(s[4]) > 42]
    if long:
        print("titles over 42 chars (a row wraps):")
        for a, b, n in long:
            print(f"  {a:10} {n:3}  {b}")
    else:
        print("98 rows, all unique, every title <= 42 chars")

if __name__ == "__main__":
    check()
