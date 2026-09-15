#!/usr/bin/env python3
"""Re-cut lecture 3's deck so the picture moves with the voice.

The shipped nocode03 is 22 slides over 389 seconds - one slide per narration
paragraph, averaging 17.5s each. Every reveal on those slides fired inside the
first 1.65 seconds (`d = 0.55 + i*0.22`) and the Ken Burns finished at 8s, so
the back half of every slide was pixel-identical. `deadzones.py` finds a dead
span every twenty seconds through the whole lecture, and that is what the team
watched and called "the audio is talking about something else".

This re-segments the SAME narration - every sentence kept verbatim, in order -
into ~50 short slides of one or two sentences each, and picks a template per
slide instead of putting nine of them through `rows`. Two consequences:

  * Nothing is re-written, so every sentence is already in the content-addressed
    voice store and the rebuild costs zero characters of the monthly allowance.
  * The manifest and the deck are emitted from ONE list here, so the voice and
    the picture cannot drift apart - with fifty slides, hand-syncing two files
    is a defect waiting to happen.

Writes:
    tools/nocode/narration/l3_v2.json        manifest for adam_deck.py
    renderer/src/nocode/l03.slides.json      the slides l03.tsx renders

    /usr/bin/python3 l03_v2.py
"""
from __future__ import annotations

import json
import os

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
NARR = os.path.join(ROOT, "tools", "nocode", "narration")
DECK = os.path.join(ROOT, "renderer", "src", "nocode")


def sentences(text: str) -> list[str]:
    """Exactly adam_lay.sentences, so a cue index here is a clip index there."""
    out, cur = [], ""
    for ch in text:
        cur += ch
        if ch in ".!?":
            out.append(cur.strip())
            cur = ""
    if cur.strip():
        out.append(cur.strip())
    return [s for s in out if s]


def source() -> dict[str, list[str]]:
    by = {}
    for f in ("l3_manifest.json", "l3_extra.json"):
        for it in json.load(open(os.path.join(NARR, f))):
            by[it["file"].replace(".mp3", "")] = sentences(it["text"])
    return by


# (block, [sentence indices], slide)
#
# The sentence indices are the contract: they say which words are spoken over
# this slide, and `adam_deck.py` turns them into the cues that land each row.
# Keep them in reading order across the whole list - the lecture is the
# concatenation of these, and nothing may be dropped or reordered.
SEGS: list[tuple[str, list[int], dict]] = [
    # ---------------------------------------------------------- the post ----
    ("01", [0], {"k": "head", "kicker": "the post that started this",
                 "lines": ["It starts with", "one *post*"], "size": 96, "trans": "fade"}),
    ("01", [1], {"k": "head", "kicker": "if you keep one thing from today",
                 "lines": ["Keep *this*"], "size": 132,
                 "stamp": "andrej karpathy", "trans": "rise"}),
    ("02", [0, 1], {"k": "rows", "kicker": "who is saying it", "lines": ["Worth *listening* to"],
                    "size": 62, "trans": "push",
                    "rows": [{"t": "A founding member of OpenAI"},
                             {"t": "Tesla, OpenAI, Stanford"}]}),
    ("02", [2, 3], {"k": "rows", "kicker": "and one term you already know",
                    "lines": ["He named", "*vibe coding*"], "size": 62, "trans": "push",
                    "rows": [{"t": "He coined the term “vibe coding”", "hot": True},
                             {"t": "So when he says he feels behind — listen"}]}),
    ("03", [0, 1], {"k": "quote", "kicker": "what he wrote", "who": "Andrej Karpathy",
                    "text": "I’ve never felt this much behind as a programmer",
                    "trans": "fade"}),
    ("03", [2], {"k": "head", "kicker": "his word for it, not ours",
                 "lines": ["Dramatically", "*refactored*"], "size": 86,
                 "stamp": "the human parts are getting sparser", "trans": "rise"}),
    ("03", [3], {"k": "meter", "kicker": "what he thinks is on the table",
                 "lines": ["*Ten times*"], "size": 96, "pct": 100,
                 "fill": "if you could string it all together",
                 "rest": "everything from the last year", "trans": "push"}),
    ("03", [4], {"k": "head", "kicker": "and if you do not", "lines": ["A *skill* issue"],
                 "size": 110, "accent": "#e53935", "trans": "rise"}),

    # the list, read out loud - a chip per spoken word is the whole point
    ("04", [0], {"k": "head", "kicker": "so he lists it",
                 "lines": ["What you", "would have to", "*string together*"], "size": 74,
                 "trans": "fade"}),
    ("04", [1, 2, 3, 4, 5, 6, 7, 8], {
        "k": "tokens", "kicker": "he reads it out, and it keeps going",
        "lines": ["The *list*"], "size": 66, "trans": "push",
        "chips": ["Agents", "Sub-agents", "Prompts", "Context", "Memory",
                  "Modes", "Permissions", "Tools"]}),
    ("04", [9, 10, 11, 12, 13, 14, 15], {
        "k": "tokens", "kicker": "…and it is still going",
        "lines": ["Still *going*"], "size": 66, "trans": "none",
        "chips": ["Plugins", "Skills", "Hooks", "MCP", "Slash commands",
                  "Workflows", "Editor integrations"]}),
    # Not a myth being corrected - a list of three properties, which is what
    # the sentence actually is. `myth` struck its line out four seconds before
    # the words justified it and then held half an empty frame.
    ("04", [16], {"k": "rows", "kicker": "and then the hard part, on top of all of it",
                  "lines": ["And a", "*mental model*"], "size": 72, "trans": "push",
                  "rows": [{"t": "Fundamentally stochastic"},
                           {"t": "Fallible"},
                           {"t": "Changing under your feet", "hot": True}],
                  "stamp": "for their strengths AND their failure modes"}),
    ("05", [0, 1], {"k": "rows", "kicker": "here is the good news",
                    "lines": ["Not a list of", "things you *should* know"], "size": 58,
                    "trans": "push",
                    "rows": [{"t": "This should reassure you, not frighten you"},
                             {"t": "Nobody expects you to know these already"}]}),
    ("05", [2, 3, 4], {"k": "head", "kicker": "it is something much better",
                       "lines": ["It is the", "*syllabus*"], "size": 104,
                       "stamp": "every single item, one at a time", "trans": "rise"}),

    # ----------------------------------------------- the vocabulary pass ----
    ("E1", [0, 1], {"k": "head", "kicker": "before we go anywhere near a tool",
                    "lines": ["Plain", "*English*"], "size": 118, "trans": "fade"}),
    ("E1", [2, 3], {"k": "rows", "kicker": "how to listen to the next two minutes",
                    "lines": ["Do *not* memorise"], "size": 70, "trans": "push",
                    "rows": [{"t": "You do not need to remember any of it"},
                             {"t": "It just needs to stop being frightening", "hot": True}]}),
    ("E2", [0, 1], {"k": "loops", "kicker": "the one definition that matters",
                    "lines": ["*Agents*"], "size": 104, "trans": "push",
                    "inner": ["Think", "Pick a tool", "Run it", "Read the result"],
                    "label": "until the goal is met",
                    "caption": "that is the whole definition"}),
    ("E2", [2, 3], {"k": "tree", "kicker": "and when one is not enough",
                    "lines": ["*Sub-agents*"], "size": 86, "trans": "push",
                    "leaves": [{"t": "Read the docs"}, {"t": "Write the test"},
                               {"t": "Fix the build"}],
                    "caption": "one instruction becomes a hundred actions"}),
    ("E3", [0, 1], {"k": "stack", "kicker": "everything it can see, right now",
                    "lines": ["*Context*"], "size": 104, "trans": "push",
                    "layers": [{"t": "Your instruction", "h": 1},
                               {"t": "The files it has read", "h": 2.4},
                               {"t": "What it said a minute ago", "h": 1.6},
                               {"t": "What it just ran", "h": 1.3}],
                    "upto": 4}),
    ("E3", [2], {"k": "rows", "kicker": "the part you keep on purpose",
                 "lines": ["*Memory*"], "size": 104, "trans": "push",
                 "rows": [{"t": "What survives between sessions"},
                          {"t": "So it does not start from nothing each morning"}]}),
    ("E3", [3], {"k": "head", "kicker": "remember this one line",
                 "lines": ["A *context* problem", "in disguise"], "size": 72,
                 "stamp": "almost every failure you will hit", "trans": "rise"}),
    ("E4", [0, 1], {"k": "rows", "kicker": "what you allow it to do",
                    "lines": ["*Tools*"], "size": 110, "trans": "push",
                    "rows": [{"t": "Read a file"}, {"t": "Write a file"},
                             {"t": "Run a command"}, {"t": "Search the web"}]}),
    ("E4", [2, 3], {"k": "myth", "kicker": "and this is the whole difference",
                    "lines": ["Talk, or *work*"], "size": 86, "trans": "push",
                    "wrong": "Without tools, a model can only talk",
                    "right": "With tools, it can work",
                    "note": "everything else is detail"}),
    ("E4", [4], {"k": "nest", "kicker": "one plug, many sockets", "lines": ["*MCP*"],
                 "size": 118, "trans": "push",
                 "outer": "One agreed way to plug tools in",
                 "inner": "Your agent",
                 "ring": ["Issue tracker", "Database", "Design files"],
                 "caption": "no custom glue for each one"}),
    ("E5", [0], {"k": "rows", "kicker": "so you stop repeating yourself",
                 "lines": ["*Skills*"], "size": 110, "trans": "push",
                 "rows": [{"t": "Short written instructions"},
                          {"t": "How you want a particular job done", "hot": True}]}),
    ("E5", [1], {"k": "head", "kicker": "and bundled up for everyone else",
                 "lines": ["*Plugins*"], "size": 132,
                 "stamp": "so a team can share them", "trans": "rise"}),
    ("E5", [2, 3, 4, 5], {"k": "flow", "kicker": "the opposite direction",
                          "lines": ["*Hooks*"], "size": 110, "trans": "push",
                          "nodes": [{"t": "When this happens"},
                                    {"t": "automatically do that"},
                                    {"t": "Run the tests", "sub": "after every edit"},
                                    {"t": "Review the code", "sub": "before every commit"}],
                          "caption": "you say it once, it happens every time"}),
    ("E6", [0, 1, 2], {"k": "ladder", "kicker": "how much rope you hand over",
                       "lines": ["Modes and", "*permissions*"], "size": 62, "trans": "push",
                       "items": ["Ask before anything", "Ask for the risky things",
                                 "Run free inside a sandbox", "Do whatever it likes"],
                       "pick": 3}),
    ("E6", [3, 4], {"k": "rows", "kicker": "and there is no correct answer",
                    "lines": ["Neither is", "*right or wrong*"], "size": 62, "trans": "push",
                    "rows": [{"t": "Choosing well for the job in front of you"},
                             {"t": "That is one of the real skills here", "hot": True}]}),
    ("E7", [0, 1, 2, 3, 4], {"k": "cols", "kicker": "same tools, very different results",
                             "lines": ["*Workflows*"], "size": 104, "trans": "push",
                             "cols": [{"head": "Plan", "items": ["Plan it first", "Then build"]},
                                      {"head": "Verify", "items": ["Build", "Review", "Test"]},
                                      {"head": "Loop", "items": ["Set it running",
                                                                 "Let it grade itself"]}]}),

    # ------------------------------------------------------- the manual ----
    ("06", [0], {"k": "head", "kicker": "and then the line this program is named for",
                 "lines": ["One more", "*line*"], "size": 96, "trans": "fade"}),
    ("06", [1], {"k": "quote", "kicker": "andrej karpathy", "who": "Andrej Karpathy",
                 "text": "Some powerful alien tool was handed around, "
                         "except it comes with no manual", "trans": "fade"}),
    ("06", [2], {"k": "head", "kicker": "his advice, and ours",
                 "lines": ["Roll up your", "*sleeves*"], "size": 92,
                 "stamp": "so you do not fall behind", "trans": "rise"}),
    ("07", [0, 1], {"k": "rows", "kicker": "so that is the gap", "lines": ["The *gap*"],
                    "size": 110, "trans": "push",
                    "rows": [{"t": "Extraordinarily powerful tools"},
                             {"t": "Handed out with no instructions", "hot": True}]}),
    ("07", [2, 3], {"k": "head", "kicker": "and that is the whole idea",
                    "lines": ["This program", "is the *manual*"], "size": 88, "trans": "rise"}),
    ("07", [4], {"k": "rows", "kicker": "what that means by the end",
                 "lines": ["Read it *again*"], "size": 84, "trans": "push",
                 "rows": [{"t": "Know what every one of those words means"},
                          {"t": "Where each works well, and where it does not"},
                          {"t": "And how they fit together", "hot": True}]}),

    # ---------------------------------------------------- who it is for ----
    ("08", [0, 1], {"k": "head", "kicker": "who is this actually for",
                    "lines": ["*Two* people"], "size": 132,
                    "stamp": "you are probably one of them", "trans": "fade"}),
    ("09", [0, 1, 2, 3], {"k": "rows", "kicker": "the first",
                          "lines": ["The aspiring", "*engineer*"], "size": 66, "trans": "push",
                          "rows": [{"t": "Maybe you are junior"},
                                   {"t": "Maybe you have never written a line"},
                                   {"t": "Maybe a bit, but no expert"}]}),
    ("09", [4, 5], {"k": "rows", "kicker": "and here is what changes",
                    "lines": ["You will *own* it"], "size": 80, "trans": "push",
                    "rows": [{"t": "Complete products, at real scale"},
                             {"t": "With coding agents doing the typing"},
                             {"t": "You own the code, even if you cannot write it yet",
                              "hot": True}]}),
    ("10", [0, 1, 2], {"k": "rows", "kicker": "the second",
                       "lines": ["The senior", "*engineer*"], "size": 66, "trans": "push",
                       "rows": [{"t": "Years of experience, new to agentic coding"},
                                {"t": "Or you read that list and felt what he felt"},
                                {"t": "Too much, too fast, no way to fit it together"}]}),
    ("10", [3], {"k": "rows", "kicker": "two things, and the second matters",
                 "lines": ["Faster — and", "still *enjoyable*"], "size": 62, "trans": "push",
                 "rows": [{"t": "Accelerate what you already do"},
                          {"t": "And keep enjoying it", "hot": True}]}),
    ("11", [0, 1], {"k": "head", "kicker": "and honestly, not only those two",
                    "lines": ["Too advanced.", "Too *basic*."], "size": 86, "trans": "fade"}),
    ("11", [2, 3, 4], {"k": "rows", "kicker": "so use it at your own speed",
                       "lines": ["Two *speeds*"], "size": 96, "trans": "push",
                       "rows": [{"t": "Obvious? Two times speed, keep moving"},
                                {"t": "Over your head? Take the gist, keep going"},
                                {"t": "It comes back around", "hot": True}]}),
    ("12", [0, 1], {"k": "rows", "kicker": "what you walk away with",
                    "lines": ["Any *scale*"], "size": 110, "trans": "push",
                    "rows": [{"t": "Build"}, {"t": "Debug"}, {"t": "Troubleshoot"},
                             {"t": "Extend"}]}),
    ("12", [2, 3], {"k": "rows", "kicker": "and it really does mean any",
                    "lines": ["Game, or", "*codebase*"], "size": 76, "trans": "push",
                    "rows": [{"t": "A small browser game"},
                             {"t": "A team codebase across several repositories"},
                             {"t": "With agents working alongside you", "hot": True}]}),
    ("13", [0, 1], {"k": "head", "kicker": "and there is a market reality here",
                    "lines": ["They *expect*", "this now"], "size": 88, "trans": "fade"}),
    ("13", [2, 3], {"k": "rows", "kicker": "straight off the job descriptions",
                    "lines": ["On the *spec*"], "size": 92, "trans": "push",
                    "rows": [{"t": "What MCP is"}, {"t": "Skills and hooks"},
                             {"t": "Long autonomous loops"},
                             {"t": "Ship quickly, without shipping slop", "hot": True}]}),

    # --------------------------------------------------------- not hype ----
    ("14", [0, 1], {"k": "head", "kicker": "one last thing, and it matters most",
                    "lines": ["This is", "*not* hype"], "size": 110, "accent": "#e53935",
                    "trans": "rise"}),
    ("14", [2, 3], {"k": "myth", "kicker": "you will see both, on camera",
                    "lines": ["Both *sides*"], "size": 96, "trans": "push",
                    "wrong": "Genuinely limited, with real drawbacks",
                    "right": "Genuinely powerful, with real benefits",
                    "note": "nothing edited out"}),
    ("15", [0, 1], {"k": "rows", "kicker": "because the skill is not belief",
                    "lines": ["Know *where*"], "size": 104, "trans": "push",
                    "rows": [{"t": "Where the strengths genuinely are"},
                             {"t": "Where the weaknesses genuinely are"},
                             {"t": "A way of working that leans on one, manages the other",
                              "hot": True}]}),
    ("15", [2, 3], {"k": "head", "kicker": "that is the whole job",
                    "lines": ["That is", "the *skill*"], "size": 118,
                    "stamp": "that is what we are here to build", "trans": "rise"}),
]


# What each template actually takes, mirrored from the `Slide` union in
# renderer/src/nocode/kit.tsx. l03.tsx has to cast the generated JSON to
# `Slide[]` (TypeScript infers `k: string` from a .json import, which matches no
# member of the union), and that cast means a wrong field here would typecheck
# and then render as an empty panel. So the shape is checked on this side.
#   kind: (required fields, optional fields, {field: element-shape})
COMMON_OPT = {"kicker", "size", "trans", "caption", "lines"}
SPEC: dict[str, tuple[set, set, dict]] = {
    "head":   ({"lines"}, {"accent", "stamp", "stampColor"}, {}),
    "rows":   ({"lines", "rows"}, {"numbered", "stamp"}, {"rows": {"t"}}),
    "ladder": ({"lines", "items"}, {"pick", "picks", "upto", "tone"}, {}),
    "cols":   ({"lines", "cols"}, set(), {"cols": {"head", "items"}}),
    "quote":  ({"text", "who"}, set(), {}),
    "tokens": ({"chips"}, {"hot"}, {}),
    "flow":   ({"nodes"}, {"loop"}, {"nodes": {"t"}}),
    "nest":   ({"outer", "inner"}, {"ring"}, {}),
    "tree":   ({"leaves"}, set(), {"leaves": {"t"}}),
    "meter":  ({"pct", "fill", "rest"}, {"tone"}, {}),
    "myth":   ({"wrong", "right"}, {"note"}, {}),
    "stack":  ({"layers"}, {"upto", "hot", "foot"}, {"layers": {"t", "h"}}),
    "loops":  ({"inner"}, {"outer", "passes", "innerRpm", "label"}, {}),
}


def check(slides: list[dict]) -> list[str]:
    bad = []
    for n, s in enumerate(slides):
        k = s.get("k")
        if k not in SPEC:
            bad.append(f"  slide {n:2d}: unknown kind {k!r}")
            continue
        req, opt, shapes = SPEC[k]
        keys = set(s) - {"k"}
        for miss in sorted(req - keys):
            bad.append(f"  slide {n:2d} ({k}): missing {miss!r}")
        for extra in sorted(keys - req - opt - COMMON_OPT):
            bad.append(f"  slide {n:2d} ({k}): unknown field {extra!r}")
        for field, want in shapes.items():
            for j, it in enumerate(s.get(field, [])):
                if not isinstance(it, dict):
                    bad.append(f"  slide {n:2d} ({k}): {field}[{j}] is "
                               f"{type(it).__name__}, want an object with {sorted(want)}")
                elif want - set(it):
                    bad.append(f"  slide {n:2d} ({k}): {field}[{j}] missing "
                               f"{sorted(want - set(it))}")
    return bad


def main() -> int:
    src = source()

    # Every sentence must be used exactly once, in order. With fifty slides a
    # dropped or duplicated sentence is silent - the voice would simply skip a
    # line and nothing would fail - so it is checked rather than trusted.
    seen: dict[str, list[int]] = {}
    for blk, idx, _ in SEGS:
        seen.setdefault(blk, []).extend(idx)
    bad = []
    for blk, sents in src.items():
        got = seen.get(blk, [])
        if got != list(range(len(sents))):
            missing = [i for i in range(len(sents)) if i not in got]
            dupes = [i for i in set(got) if got.count(i) > 1]
            bad.append(f"  {blk}: have {len(got)} of {len(sents)}"
                       + (f", missing {missing}" if missing else "")
                       + (f", duplicated {dupes}" if dupes else "")
                       + ("" if got == sorted(got) else ", OUT OF ORDER"))
    if bad:
        print("narration does not round-trip:\n" + "\n".join(bad))
        return 1

    manifest, slides = [], []
    for n, (blk, idx, slide) in enumerate(SEGS):
        text = " ".join(src[blk][i] for i in idx)
        manifest.append({"file": f"s{n:02d}.mp3", "text": text})
        slides.append(slide)

    wrong = check(slides)
    if wrong:
        print("slides do not match the templates in kit.tsx:\n" + "\n".join(wrong))
        return 1

    mp = os.path.join(NARR, "l3_v2.json")
    sp = os.path.join(DECK, "l03.slides.json")
    json.dump(manifest, open(mp, "w"), indent=1, ensure_ascii=False)
    json.dump(slides, open(sp, "w"), indent=1, ensure_ascii=False)

    words = sum(len(m["text"].split()) for m in manifest)
    kinds: dict[str, int] = {}
    for s in slides:
        kinds[s["k"]] = kinds.get(s["k"], 0) + 1
    print(f"{len(SEGS)} slides (was 22), {words} words, "
          f"{sum(len(i) for _, i, _ in SEGS)} sentences")
    print("  templates: " + ", ".join(f"{k} x{v}" for k, v in
                                      sorted(kinds.items(), key=lambda x: -x[1])))
    print(f"  -> {os.path.relpath(mp, ROOT)}")
    print(f"  -> {os.path.relpath(sp, ROOT)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
