#!/usr/bin/env python3
"""Rebuild lecture 1 so the picture never stops and the words match the frame.

The shipped nocode01 is 10:12 and **81% pixel-frozen**, with a single 140-second
held frame at 6:22. The cause is in `assemble_l01.py`: each segment is retimed
by ONE factor to hit its narration length, so segment J plays 138.5s of source
at 1.13x - and 90 seconds of that source is an agent thinking with the screen
not moving at all. Worse, thirteen sentences of J's narration are a spoken aside
about which projects are worth building, which has no picture anywhere in the
take. The lecture was 10 minutes built out of about 4 minutes of material.

This rebuild:

  * **Densifies the footage** (`densify.py`): live picture plays, every wait is
    ramped to a beat. 486s of build take becomes 99s that is 72% live.
  * **Covers the surplus narration with animated cards** instead of a held
    frame. 144 seconds of this lecture's words have no footage to sit under -
    that is not a cutting problem, it is a missing picture, and a card that
    teaches the point is the honest answer.
  * **Re-voices everything with Adam**, including the opening deck, which was
    still Kokoro - the lecture changed voice a third of the way through.
  * Runs on `kit.tsx` like lecture 3, so module 1 looks like one course.

Two passes, because a shot has to be exactly as long as the words over it and
that length is only known once the line is spoken:

    /usr/bin/python3 l01_v2.py --manifest     # 1. write the voice manifest
    /usr/bin/python3 adam_deck.py .../l01_v2.json renderer/public/nocode01v2/vo
    /usr/bin/python3 l01_v2.py --slides       # 2. cut footage to fit, emit slides
"""
from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
NARR = os.path.join(ROOT, "tools", "nocode", "narration")
TAKES = os.path.join(ROOT, "videos", "nocode", "takes", "nocode01")
PUB = os.path.join(ROOT, "renderer", "public", "nocode01v2")
DECK = os.path.join(ROOT, "renderer", "src", "nocode")

sys.path.insert(0, HERE)
import densify as D                                            # noqa: E402
from adam_lay import sentences                                 # noqa: E402

# Which slice of which take each footage segment comes from. These are
# `assemble_l01.py`'s own in/out points, kept so a shot still shows what its
# narration was written against.
RANGES = {
    "B":  ("BC_browser.mp4",  1.2,  27.0),
    "C":  ("BC_browser.mp4", 27.0,  70.0),
    "E":  ("E_drag.mp4",      4.0,  12.0),
    "G":  ("GH_cursor.mp4",   6.0,  37.4),
    # No H1. GH_cursor 37.4-54.9 is him browsing his own home directory and the
    # company name is legible on screen from 39.0s to 46.5s of that take. It is
    # cut rather than blurred, on his instruction - a blurred smear over the
    # thing you are being told to look at teaches nothing anyway.
    # H2 starts where the new folder is being named, so the home folder and its
    # contents are never on screen.
    "H2": ("H2_folder.mp4",  17.5,  40.5),
    "J":  ("JKLM_build.mp4",  1.5, 140.0),
    "K":  ("JKLM_build.mp4", 140.0, 174.0),
    "L":  ("JKLM_build.mp4", 174.0, 194.0),
    "M":  ("JKLM_build.mp4", 194.0, 486.5),
}


# Lines replaced because the picture under them changed. The originals walked
# through his own home directory, which is exactly the footage that had to go.
REVOICE = {
    "H_folder": [
        "So Cursor is based on VS Code, and you may be very familiar with this "
        "already if you use VS Code.",
        "But open a project really means: pick a folder on your computer.",
        "Everything to do with this project is going to live inside it, so it is "
        "worth making a fresh one rather than reusing something.",
        "So I press New Folder, and I am going to call this one Instant.",
        "Create.",
        "And there it is, an empty folder with nothing in it at all.",
        "Then Open.",
        "And Cursor opens the project called Instant.",
        "You can tell that it worked, because it says Instant in block capitals "
        "right there.",
    ],
}


def source() -> dict[str, list[str]]:
    by = {}
    for f in ("slides_manifest.json", "screen_manifest.json"):
        for it in json.load(open(os.path.join(NARR, f))):
            by[it["file"].replace(".mp3", "")] = sentences(it["text"])
    by.update(REVOICE)
    return by


def shot(take: str, **kw) -> dict:
    """A slide backed by footage. `take` names a densified segment; the clip is
    cut in pass two, sequentially, so consecutive shots from one take continue
    through it rather than restarting."""
    return {"k": "shot", "take": take, **kw}


# (block, [sentence indices], slide)
SEGS: list[tuple[str, list[int], dict]] = [
    # ============================================================ opening ====
    ("01", [0], {"k": "head", "kicker": "welcome", "size": 92, "trans": "fade",
                 "lines": ["The definitive", "*AI Coder* program"]}),
    ("01", [1], {"k": "head", "kicker": "and what it is called", "size": 128,
                 "trans": "rise", "lines": ["The *Missing*", "Manual"],
                 "stamp": "three weeks"}),
    ("02", [0], {"k": "timeline", "kicker": "where this ends up", "size": 66,
                 "trans": "push", "lines": ["The *journey*"],
                 "items": [{"when": "week one", "t": "The tools, and the ideas under them"},
                           {"when": "week two", "t": "Building for real"},
                           {"when": "week three", "t": "An expert at agentic engineering",
                            "hot": True}]}),
    ("03", [0, 1, 2], {"k": "head", "kicker": "but first, a word of caution",
                       "size": 104, "trans": "rise", "accent": "#e53935",
                       "lines": ["*Buckle* up"], "stamp": "a wild ride"}),
    ("03", [3, 4, 5], {"k": "curve", "kicker": "what is in store", "size": 76,
                       "trans": "push", "lines": ["A *roller coaster*"],
                       "rising": True, "accel": True,
                       "yLabel": "what you can build", "xLabel": "three weeks"}),
    ("04", [0, 1], {"k": "rows", "kicker": "how these programs usually start",
                    "size": 66, "trans": "push", "lines": ["Standard *fare*"], "numbered": True,
                    "rows": [{"t": "Begin with the objectives"},
                             {"t": "Introduce myself"},
                             {"t": "The curriculum"}, {"t": "The logistics"}]}),
    ("04", [2], {"k": "head", "kicker": "and here is the exception",
                 "size": 118, "trans": "rise", "accent": "#e53935",
                 "lines": ["Except,", "*I don't*"]}),
    ("05", [0], {"k": "head", "kicker": "if you have been here before", "size": 92,
                 "trans": "fade", "lines": ["Instant", "*gratification*"]}),
    ("05", [1, 2, 3], {"k": "rows", "kicker": "so here is the plan", "size": 78,
                       "trans": "push", "lines": ["Build *first*"],
                       "rows": [{"t": "Roll up sleeves"},
                                {"t": "Build something together", "hot": True},
                                {"t": "The rest of it can wait"}]}),
    ("06", [0], {"k": "tokens", "kicker": "over the next three weeks", "size": 72,
                 "trans": "push", "lines": ["Several tools", "that *write code*"],
                 "chips": ["Cursor", "Claude Code", "Copilot", "Codex",
                           "Antigravity", "\u2026and more"]}),
    ("07", [0, 1, 2], {"k": "rows", "kicker": "but today, just one", "size": 110,
                       "trans": "push", "lines": ["*Cursor*"],
                       "rows": [{"t": "Super popular"}, {"t": "Very easy to use"},
                                {"t": "Free trial if you are new", "hot": True}]}),
    ("08", [0, 1], {"k": "rows", "kicker": "and if you have used it already",
                    "size": 70, "trans": "push", "lines": ["No trial *left*?"],
                    "rows": [{"t": "You do not need one"},
                             {"t": "Watch what I do"},
                             {"t": "Or follow along in another product", "hot": True}]}),
    ("09", [0, 1], {"k": "head", "kicker": "so let us do something simple",
                    "size": 96, "trans": "rise",
                    "lines": ["Generate code", "*right away*"],
                    "stamp": "let's go do it"}),

    # ============================================== B - open the browser ====
    ("B_browser", [0, 1, 2], shot("B", kicker="step one", lines=["Open a *browser*"],
                                  size=64)),
    ("B_browser", [3, 4, 5, 6], shot("B")),
    ("B_browser", [7, 8, 9, 10], shot("B")),

    # ==================================================== C - download it ====
    ("C_download", [0, 1], shot("C", kicker="step two", lines=["*Download* it"],
                                size=64)),
    ("C_download", [2, 3, 4], shot("C")),
    ("C_download", [5, 6, 7], shot("C")),

    # ======================================= D - the same thing on Windows ===
    # No Windows footage was ever shot; the original insert was a text card too,
    # it simply held still for ten seconds at a time.
    ("D_windows", [0, 1], {"k": "rows", "kicker": "and for the pc people",
                           "size": 86, "trans": "push", "lines": ["On *Windows*"],
                           "rows": [{"t": "Go to cursor.com"},
                                    {"t": "“Download for Windows” is the default button",
                                     "hot": True}]}),
    ("D_windows", [2, 3, 4], {"k": "flow", "kicker": "once it has downloaded",
                              "size": 78, "trans": "push", "lines": ["Then the *usual*"],
                              "nodes": [{"t": "Open the file"}, {"t": "Yes, run it"},
                                        {"t": "Installed"}]}),
    ("D_windows", [5, 6], {"k": "head", "kicker": "and that is the pc done",
                           "size": 92, "trans": "rise",
                           "lines": ["Same app,", "*either way*"]}),

    # ================================================= E - drag to install ===
    ("E_drag", [0], shot("E", kicker="on a mac", lines=["Drag to", "*Applications*"],
                         size=62)),
    ("E_drag", [1], {"k": "head", "kicker": "and that is the install done",
                     "size": 104, "trans": "rise", "lines": ["Now let's", "*start*"]}),

    # ======================================================== F - sign in ====
    ("F_signin", [0, 1, 2], {"k": "rows", "kicker": "the first time you open it",
                             "size": 80, "trans": "push", "lines": ["Sign in, or", "*sign up*"],
                             "rows": [{"t": "You will be prompted"},
                                      {"t": "Sign up if you have no account", "hot": True}]}),
    ("F_signin", [3, 4, 5], {"k": "meter", "kicker": "what it costs you today",
                             "size": 86, "trans": "push", "lines": ["*Free* trial"],
                             "pct": 100, "fill": "free for the first few weeks",
                             "rest": "which is all we need"}),
    ("F_signin", [6, 7, 8], {"k": "rows", "kicker": "if a setup screen appears",
                             "size": 76, "trans": "push", "lines": ["Just press *through*"],
                             "rows": [{"t": "You can accept the defaults"},
                                      {"t": "Nothing here matters yet", "hot": True}]}),
    ("F_signin", [9, 10, 11, 12], {"k": "head", "kicker": "and then you are in",
                                   "size": 100, "trans": "rise",
                                   "lines": ["Signed in,", "*ready*"]}),

    # ================================================== G - the welcome ====
    ("G_welcome", [0, 1], shot("G", kicker="launching it", lines=["The *Cursor* screen"],
                               size=62)),
    ("G_welcome", [2, 3, 4], shot("G")),
    ("G_welcome", [5, 6], shot("G")),
    ("G_welcome", [7, 8], {"k": "rows", "kicker": "if your screen looks different",
                           "size": 72, "trans": "push", "lines": ["Not *identical*?"],
                           "rows": [{"t": "It may not say Pro — that is fine"},
                                    {"t": "A free trial account looks slightly different",
                                     "hot": True}]}),

    # ============================================= H - open a project =======
    # Cards for the idea, footage only from the point the folder is named.
    ("H_folder", [0], {"k": "rows", "kicker": "and you may know this already",
                       "size": 72, "trans": "push", "lines": ["Built on", "*VS Code*"],
                       "rows": [{"t": "If you use VS Code, this is familiar"}]}),
    ("H_folder", [1, 2], {"k": "rows", "kicker": "what \u201copen a project\u201d means",
                          "size": 68, "trans": "push", "lines": ["Pick a *folder*"],
                          "rows": [{"t": "Everything for this project lives in it"},
                                   {"t": "Make a fresh one rather than reuse something",
                                    "hot": True}]}),
    ("H_folder", [3, 4], shot("H2", kicker="so, a new one",
                              lines=["Call it *Instant*"], size=62)),
    ("H_folder", [5, 6], shot("H2")),
    ("H_folder", [7, 8], shot("H2", kicker="and there is the project",
                              lines=["*Instant*, open"], size=62)),

    # ================================== I - the same thing on Windows ======
    ("I_winfolder", [0, 1, 2], {"k": "rows", "kicker": "now the same on a pc",
                                "size": 82, "trans": "push", "lines": ["On the *PC*"],
                                "rows": [{"t": "Here is Cursor on Windows"},
                                         {"t": "If it says Sign In, click it first"}]}),
    ("I_winfolder", [3, 4, 5, 6], {"k": "flow", "kicker": "otherwise",
                                   "size": 76, "trans": "push",
                                   "lines": ["Open *Project*"],
                                   "nodes": [{"t": "Open Project"},
                                             {"t": "The PC file browser"},
                                             {"t": "Your Home folder"}]}),
    ("I_winfolder", [7, 8, 9, 10], {"k": "rows", "kicker": "same idea as the mac",
                                    "size": 70, "trans": "push",
                                    "lines": ["Find, or *make* it"],
                                    "rows": [{"t": "Go to your Projects folder"},
                                             {"t": "Or create one"},
                                             {"t": "Then a folder for this project",
                                              "hot": True}]}),
    ("I_winfolder", [11, 12, 13, 14], {"k": "rows", "kicker": "name it the same thing",
                                       "size": 84, "trans": "push",
                                       "lines": ["Call it *Instant*"],
                                       "rows": [{"t": "New folder"}, {"t": "Type Instant"},
                                                {"t": "Create"}, {"t": "Open", "hot": True}]}),
    ("I_winfolder", [15, 16, 17, 18], {"k": "head", "kicker": "and both machines meet here",
                                       "size": 88, "trans": "rise",
                                       "lines": ["Same project,", "*either way*"]}),

    # ============================================== J - orient, and a detour =
    ("J_panes", [0, 1], shot("J", kicker="orienting you", lines=["*Three* panes"],
                             size=72)),
    ("J_panes", [2, 3, 4, 5], shot("J")),
    ("J_panes", [6, 7], shot("J")),
    ("J_panes", [8, 9, 10], {"k": "rows", "kicker": "if you cannot see all three",
                             "size": 66, "trans": "push", "lines": ["View → *Appearance*"],
                             "rows": [{"t": "Toggle things until the three panes appear"},
                                      {"t": "You will figure it out — otherwise, message me",
                                       "hot": True}]}),
    # The thirteen-sentence aside. There is no footage for any of it anywhere in
    # the take - this is the stretch the delivered lecture held one frame for.
    ("J_panes", [11], {"k": "head", "kicker": "so here is what we will do", "size": 84,
                       "trans": "fade", "lines": ["Put the agent", "*to work*"]}),
    ("J_panes", [12], {"k": "rows", "kicker": "something i feel strongly about",
                       "size": 62, "trans": "push", "lines": ["Projects with", "*real impact*"],
                       "rows": [{"t": "On every one of our programs"},
                                {"t": "When we build, it should matter", "hot": True}]}),
    ("J_panes", [13, 14, 15], {"k": "rows", "kicker": "what that means in practice",
                               "size": 66, "trans": "push", "lines": ["*Business* first"],
                               "rows": [{"t": "Things that are actually valuable"},
                                        {"t": "Something you could monetise"},
                                        {"t": "Or a springboard for what is next", "hot": True}]}),
    ("J_panes", [16, 17, 18], {"k": "myth", "kicker": "and then there is today",
                               "size": 92, "trans": "push", "lines": ["*One* exception"],
                               "wrong": "Every project here has business impact",
                               "right": "Except this one",
                               "note": "the only exception on the whole program"}),
    ("J_panes", [19, 20, 21], {"k": "head", "kicker": "because of how this should start",
                               "size": 104, "trans": "rise",
                               "lines": ["Today we", "have *fun*"],
                               "stamp": "what better way to begin"}),
    ("J_panes", [22], {"k": "rows", "kicker": "and if you are all business",
                       "size": 70, "trans": "push", "lines": ["Do not *worry*"],
                       "rows": [{"t": "It is all business after this point", "hot": True}]}),
    ("J_panes", [23, 24, 25], shot("J", kicker="back to it", lines=["Talk to the *agent*"],
                                   size=62)),
    ("J_panes", [26, 27, 28], {"k": "rows", "kicker": "and if these words are new",
                               "size": 66, "trans": "push", "lines": ["All of it, *tomorrow*"],
                               "rows": [{"t": "Agents, how they work, what they do"},
                                        {"t": "Today we are just having fun", "hot": True}]}),

    # =================================================== K - pick a model ====
    ("K_model", [0, 1], shot("K", kicker="the box on the right",
                             lines=["A message to", "your *agent*"], size=58)),
    ("K_model", [2, 3], shot("K")),
    ("K_model", [4, 5, 6], {"k": "rows", "kicker": "and which one you can pick",
                            "size": 70, "trans": "push", "lines": ["Depends on", "your *plan*"],
                            "rows": [{"t": "You may not see every model"},
                                     {"t": "Any current one will do this job", "hot": True}]}),

    # ======================================================= L - the prompt ==
    ("L_prompt", [0, 1], shot("L", kicker="and here is the ask",
                              lines=["The *request*"], size=72)),
    ("L_prompt", [2, 3], shot("L")),

    # ======================================================== M - the build ==
    ("M_build", [0, 1, 2, 3], shot("M", kicker="and off it goes",
                                   lines=["It *starts* working"], size=62)),
    ("M_build", [4, 5, 6], shot("M")),
    ("M_build", [7, 8, 9, 10], {"k": "rows", "kicker": "it may stop and ask you",
                                "size": 64, "trans": "push", "lines": ["It might ask", "*permission*"],
                                "rows": [{"t": "If you are happy with it, say yes"},
                                         {"t": "Mine does not — it depends on your settings"},
                                         {"t": "The default will just go and do it", "hot": True}]}),
    ("M_build", [11, 12, 13, 14], {"k": "rows", "kicker": "and this matters all program",
                                   "size": 62, "trans": "push",
                                   "lines": ["Yours will", "*differ*"],
                                   "rows": [{"t": "I am on one model, you may be on another"},
                                            {"t": "Every one of us gets a different run"},
                                            {"t": "What happens next is not fully predictable",
                                             "hot": True}]}),
    ("M_build", [15], shot("M", kicker="but it is away", lines=["Writing *index.html*"],
                           size=62)),
]


def check_narration(src: dict) -> list[str]:
    seen: dict[str, list[int]] = {}
    for blk, idx, _ in SEGS:
        seen.setdefault(blk, []).extend(idx)
    bad = []
    for blk, sents in src.items():
        got = seen.get(blk, [])
        if got != list(range(len(sents))):
            missing = [i for i in range(len(sents)) if i not in got]
            bad.append(f"  {blk}: {len(got)} of {len(sents)}"
                       + (f", missing {missing}" if missing else "")
                       + ("" if got == sorted(got) else ", OUT OF ORDER"))
    return bad


def write_manifest() -> int:
    src = source()
    bad = check_narration(src)
    if bad:
        print("narration does not round-trip:\n" + "\n".join(bad))
        return 1
    man = [{"file": f"s{n:02d}.mp3",
            "text": " ".join(src[blk][i] for i in idx)}
           for n, (blk, idx, _) in enumerate(SEGS)]
    p = os.path.join(NARR, "l01_v2.json")
    json.dump(man, open(p, "w"), indent=1, ensure_ascii=False)
    shots = sum(1 for _, _, s in SEGS if s["k"] == "shot")
    print(f"{len(SEGS)} slides ({shots} footage, {len(SEGS)-shots} cards), "
          f"{sum(len(m['text'].split()) for m in man)} words")
    print(f"  -> {os.path.relpath(p, ROOT)}")
    return 0


def cut(src_clip: str, at: float, length: float, dst: str) -> None:
    subprocess.run(["ffmpeg", "-nostdin", "-y", "-loglevel", "error",
                    "-ss", f"{at:.3f}", "-i", src_clip, "-t", f"{length:.3f}",
                    "-an", "-c:v", "libx264", "-preset", "medium", "-crf", "18",
                    "-pix_fmt", "yuv420p", "-g", "60", dst], check=True)


def write_slides(hold: float, min_dead: float) -> int:
    vo = os.path.join(PUB, "vo")
    durs = json.load(open(os.path.join(vo, "durs.json")))
    if len(durs) != len(SEGS):
        print(f"durs.json has {len(durs)} entries, SEGS has {len(SEGS)} - "
              "re-run --manifest and adam_deck.py")
        return 1
    os.makedirs(os.path.join(PUB, "clips"), exist_ok=True)

    # Exactly as much picture as there are words over it. Every live frame is
    # kept and the waits absorb the difference, so a shot can neither run out
    # of footage nor sit on a frozen frame waiting for the line to end.
    need: dict[str, float] = {}
    for n, (_, _, s) in enumerate(SEGS):
        if s["k"] == "shot":
            need[s["take"]] = need.get(s["take"], 0) + durs[n]

    dense: dict[str, tuple[str, float]] = {}
    short = {}
    for seg in need:
        take, a, b = RANGES[seg]
        src_take = os.path.join(TAKES, take)
        want = need[seg] + 0.4                      # a little tail per segment
        pieces = D.fit(src_take, a, b, want, min_dead)
        if pieces is None:
            loose = D.plan(src_take, min_dead, 4.5, 24.0, 1.0, a, b)
            short[seg] = (want, sum(p["len"] for p in loose))
            continue
        dst = os.path.join(PUB, "clips", f"{seg}.mp4")
        D.render(src_take, dst, pieces, 18)
        got = D.dur(dst)
        dense[seg] = (dst, got)
        ramps = [p for p in pieces if p["kind"] == "ramp"]
        print(f"  {seg}: {b-a:6.1f}s source -> {got:5.1f}s  "
              f"({len(ramps)} waits at {ramps[0]['len']:.1f}s each)"
              if ramps else f"  {seg}: {b-a:6.1f}s -> {got:5.1f}s (no waits)")
    if short:
        # Not a cutting problem: the range genuinely has less picture than the
        # words need, so those sentences belong on a card.
        print("footage too short for the words over it, even with every wait stretched:")
        for k, (want, have) in short.items():
            print(f"  {k}: needs {want:.1f}s, has at most {have:.1f}s "
                  f"(short {want-have:.1f}s) - move sentences to a card")
        return 1

    slides, cursor = [], {}
    for n, (_, _, s) in enumerate(SEGS):
        if s["k"] != "shot":
            slides.append(s)
            continue
        take = s["take"]
        at = cursor.get(take, 0.0)
        clip = os.path.join(PUB, "clips", f"{take}_{n:02d}.mp4")
        cut(dense[take][0], at, durs[n] + 0.25, clip)
        cursor[take] = at + durs[n]
        out = {k: v for k, v in s.items() if k != "take"}
        out["src"] = f"nocode01v2/clips/{os.path.basename(clip)}"
        slides.append(out)

    p = os.path.join(DECK, "l01.slides.json")
    json.dump(slides, open(p, "w"), indent=1, ensure_ascii=False)
    print(f"{len(slides)} slides, {sum(1 for s in slides if s['k']=='shot')} footage")
    for k in sorted(need):
        print(f"  {k}: used {need[k]:5.1f}s of {dense[k][1]:5.1f}s densified")
    print(f"  -> {os.path.relpath(p, ROOT)}")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--manifest", action="store_true")
    ap.add_argument("--slides", action="store_true")
    ap.add_argument("--hold", type=float, default=1.5)
    ap.add_argument("--min-dead", type=float, default=6.0)
    a = ap.parse_args()
    if a.manifest:
        return write_manifest()
    if a.slides:
        return write_slides(a.hold, a.min_dead)
    ap.print_help()
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
