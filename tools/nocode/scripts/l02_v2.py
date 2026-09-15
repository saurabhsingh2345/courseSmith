#!/usr/bin/env python3
"""Rebuild lecture 2 so the waiting is cut and the words keep their picture.

nocode02 is 6:54 of screen recording and 55% pixel-frozen, with a 44-second
held frame at 2:19. Unlike lecture 1 the material here is good - it is a game
being played and an agent being driven, which moves by itself - and thirteen of
its fifteen segments have all the picture their narration needs once the waits
are ramped out.

Only two do not:

  * **N07** wants 34.6s and the take has 9 seconds of live picture. Most of that
    narration is spoken over nothing: an aside about how the result differs
    every run.
  * **N08** wants 11.6s and the take has **no** live picture at all - it is the
    agent editing files while he says "I'll come back in a second". That is the
    44-second hold.

So those get cards and everything else gets densified footage.

    /usr/bin/python3 l02_v2.py --manifest
    /usr/bin/python3 adam_deck.py .../l02_v2.json renderer/public/nocode02v2/vo
    /usr/bin/python3 l02_v2.py --slides
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
TAKES = os.path.join(ROOT, "videos", "nocode", "takes", "nocode02")
PUB = os.path.join(ROOT, "renderer", "public", "nocode02v2")
DECK = os.path.join(ROOT, "renderer", "src", "nocode")

sys.path.insert(0, HERE)
import densify as D                                            # noqa: E402
from adam_lay import sentences                                 # noqa: E402

# assemble_l02.py's own in/out points, so a shot still shows what its narration
# was written against.
RANGES = {
    "N01": ("L2A_files.mp4",  1.5,  24.0), "N02": ("L2A_files.mp4", 24.0,  50.0),
    "N03": ("L2B_serve.mp4",  2.0,  24.0), "N04": ("L2C_play.mp4",   1.5,  18.0),
    "N05": ("L2C_play.mp4",  26.0,  56.0), "N06": ("L2C_play.mp4",  56.0, 100.0),
    "N07": ("L2D_iter1.mp4",  2.0,  42.0),
    "N09": ("L2E_test1.mp4",  8.0,  40.0), "N10": ("L2F_iter2.mp4",  2.0,  40.0),
    "N11": ("L2G_test2.mp4", 10.0,  38.0), "N12": ("L2H_wrap.mp4",   1.5,  58.0),
    "N13": ("L2F_iter2.mp4", 45.0, 228.0), "N14": ("L2G_test2.mp4", 16.0,  52.0),
    "N15": ("L2C_play.mp4",  72.0, 104.0),
}


# Rewritten on his note: "in the gameplay please don't have commentary on the
# gameplay, it's not matching at all - speak about the process, how good it
# looks and feels, what it lacks, and what we just achieved."
#
# The originals were play-by-play ("whoa", "I was killed", "I'm not much good at
# this", "I won again"), which only works if the cut happens to be on the exact
# moment being described - and after densifying, it never is. These lines are
# true of any run of the game, so the words and the picture cannot drift apart.
REVOICE = {
    "N05": [
        "Stop for a second and look at what is actually on that screen.",
        "Movement, turning, shooting, and an opponent that reacts to you.",
        "None of that was written by hand.",
        "About thirty words of plain English produced a playable first person "
        "shooter.",
        "Ten minutes ago this folder was completely empty.",
    ],
    "N06": [
        "It is entirely possible that when you ran yours, something did not work.",
        "That is normal, and it is fixable - you go back to the agent, tell it "
        "what went wrong, and it repairs its own work.",
        "We are going to do exactly that in a moment.",
        "And it is worth being honest about what this is.",
        "The arena is plain, the lighting is flat, there is no sound, and the "
        "opponent is not remotely clever.",
        "But it runs, it plays, and it came out of a single sentence.",
    ],
    "N09": [
        "It says it has done it, so let us go and look.",
        "And there it is, the opponent has some detail on it now.",
        "It reads as an enemy rather than a shape.",
        "That is the loop you will use for the rest of this program.",
        "Look at what came back, decide what you want changed, and say so in "
        "plain English.",
    ],
    "N11": [
        "It says it has done both of them.",
        "And there is the heads up display, with the difficulty raised behind it.",
        "Two instructions, one round trip.",
        "Neither of them needed me to know where in the code any of that lives.",
    ],
    "N14": [
        "Here is what came back.",
        "Look at the difference against the one we built together.",
        "There is an entry screen, there is lighting, and the arena has real "
        "structure to it.",
        "There is a heads up display carrying health and kills.",
        "Same brief, same tool, one prompt, and no feedback from me at any point.",
        "The only thing that changed is the technique wrapped around the request.",
        "And that technique is what the next three weeks are for.",
    ],
}


def source() -> dict[str, list[str]]:
    by = {it["file"].replace(".mp3", "").split("_")[0]: sentences(it["text"])
          for it in json.load(open(os.path.join(NARR, "l2_manifest.json")))}
    by.update(REVOICE)
    return by


def shot(take: str, **kw) -> dict:
    return {"k": "shot", "take": take, **kw}


SEGS: list[tuple[str, list[int], dict]] = [
    # ------------------------------------------------- it finished, files ---
    ("N01", [0, 1], shot("N01", kicker="it finished", lines=["A whole set", "of *files*"],
                         size=58)),
    ("N01", [2, 3, 4], shot("N01")),
    ("N02", [0, 1, 2], shot("N02", kicker="what it wrote", lines=["Open them *up*"],
                            size=62)),
    ("N02", [3, 4], shot("N02")),

    # ------------------------------------------------------------ run it ---
    ("N03", [0, 1], shot("N03", kicker="time to try it", lines=["*Run* it"], size=76)),
    ("N03", [2], shot("N03")),

    # ------------------------------------------------------------- play ----
    ("N04", [0, 1, 2, 3], shot("N04", kicker="and up it comes",
                               lines=["One arena.", "No *second place*."], size=54)),
    ("N04", [4, 5, 6, 7], shot("N04")),
    ("N05", [0, 1, 2], shot("N05", kicker="what is on the screen",
                            lines=["Nobody *typed* this"], size=60)),
    ("N05", [3, 4], shot("N05")),
    ("N06", [0, 1, 2], shot("N06", kicker="if yours went wrong",
                            lines=["Tell the agent,", "it *fixes* it"], size=52)),
    # What it lacks, said plainly - and a list is the honest way to show it.
    ("N06", [3, 4, 5], {"k": "rows", "kicker": "and being honest about it",
                        "size": 64, "trans": "push",
                        "lines": ["What it", "*is not*"],
                        "rows": [{"t": "A plain arena, flat lighting, no sound"},
                                 {"t": "An opponent that is not remotely clever"},
                                 {"t": "But it runs \u2014 from one sentence", "hot": True}]}),

    # --------------------------------------- N07: mostly spoken over nothing -
    ("N07", [0, 1, 2], shot("N07", kicker="from one prompt", lines=["It just *worked*"],
                            size=68)),
    ("N07", [3, 4, 5, 6, 7, 8], {"k": "rows", "kicker": "and it is different every run",
                                 "size": 62, "trans": "push",
                                 "lines": ["Yours will", "*differ*"],
                                 "rows": [{"t": "Maybe the shooting did not work"},
                                          {"t": "Once the keys did not work at all"},
                                          {"t": "I have run this a lot — never the same twice",
                                           "hot": True}]}),
    ("N07", [9, 10, 11, 12, 13], shot("N07", kicker="so let us improve it",
                                      lines=["Make the opponent", "*look* like one"],
                                      size=52)),

    # ------------------------------ N08: no live picture anywhere in the take -
    ("N08", [0], {"k": "flow", "kicker": "and that is the instruction",
                  "size": 76, "trans": "push", "lines": ["Off it *goes*"],
                  "nodes": [{"t": "Read the files"}, {"t": "Add the detail"},
                            {"t": "Save"}]}),
    ("N08", [1], {"k": "rows", "kicker": "while it edits", "size": 70, "trans": "push",
                  "lines": ["Back in a", "*second*"],
                  "rows": [{"t": "It is editing and changing files"},
                           {"t": "We pick it up when it is ready to test", "hot": True}]}),

    # ------------------------------------------------------- test round 1 ---
    ("N09", [0, 1, 2], shot("N09", kicker="it says it is done",
                            lines=["An *enemy*, now"], size=68)),
    ("N09", [3, 4], {"k": "loops", "kicker": "and this is the whole method",
                     "size": 72, "trans": "push", "lines": ["The *loop*"],
                     "inner": ["Look at what came back", "Decide what to change",
                               "Say it in plain English"],
                     "label": "for the rest of the program"}),

    # ------------------------------------------------------- round 2 --------
    ("N10", [0, 1, 2], shot("N10", kicker="one more edit",
                            lines=["A *HUD*, and", "make it harder"], size=54)),
    ("N10", [3, 4, 5, 6], shot("N10")),
    ("N11", [0, 1], shot("N11", kicker="and again", lines=["A *HUD*, and harder"],
                         size=62)),
    ("N11", [2, 3], shot("N11")),

    # ---------------------------------------------------------- the wrap ----
    ("N12", [0, 1], shot("N12", kicker="so there you have it",
                         lines=["An agent wrote", "*all* of it"], size=54)),
    ("N12", [2, 3, 4], shot("N12")),
    ("N12", [5, 6, 7], shot("N12")),
    ("N12", [8, 9, 10, 11], shot("N12")),

    # --------------------------------------------------------- the teaser ---
    ("N13", [0, 1, 2], shot("N13", kicker="one more thing",
                            lines=["A *technique*", "from later on"], size=54)),
    ("N13", [3, 4, 5, 6], shot("N13")),
    ("N14", [0, 1], shot("N14", kicker="zero shot, one prompt",
                         lines=["No feedback.", "*This* came back."], size=50)),
    ("N14", [2, 3], shot("N14")),
    ("N14", [4, 5, 6], shot("N14")),
    ("N15", [0, 1, 2], shot("N15", kicker="and that is the teaser",
                            lines=["Going *pro*"], size=76)),
    ("N15", [3, 4, 5], shot("N15")),
    ("N15", [6, 7, 8, 9, 10], shot("N15")),
]


def check_narration(src):
    seen = {}
    for blk, idx, _ in SEGS:
        seen.setdefault(blk, []).extend(idx)
    bad = []
    for blk, sents in src.items():
        got = seen.get(blk, [])
        if got != list(range(len(sents))):
            bad.append(f"  {blk}: {len(got)} of {len(sents)}, "
                       f"missing {[i for i in range(len(sents)) if i not in got]}")
    return bad


def write_manifest() -> int:
    src = source()
    bad = check_narration(src)
    if bad:
        print("narration does not round-trip:\n" + "\n".join(bad))
        return 1
    man = [{"file": f"s{n:02d}.mp3", "text": " ".join(src[blk][i] for i in idx)}
           for n, (blk, idx, _) in enumerate(SEGS)]
    p = os.path.join(NARR, "l02_v2.json")
    json.dump(man, open(p, "w"), indent=1, ensure_ascii=False)
    shots = sum(1 for _, _, s in SEGS if s["k"] == "shot")
    print(f"{len(SEGS)} slides ({shots} footage, {len(SEGS)-shots} cards)")
    print(f"  -> {os.path.relpath(p, ROOT)}")
    return 0


def write_slides(min_dead: float) -> int:
    vo = os.path.join(PUB, "vo")
    durs = json.load(open(os.path.join(vo, "durs.json")))
    if len(durs) != len(SEGS):
        print(f"durs.json has {len(durs)}, SEGS has {len(SEGS)}")
        return 1
    os.makedirs(os.path.join(PUB, "clips"), exist_ok=True)

    need = {}
    for n, (_, _, s) in enumerate(SEGS):
        if s["k"] == "shot":
            need[s["take"]] = need.get(s["take"], 0) + durs[n]

    dense, short = {}, {}
    for seg in need:
        take, a, b = RANGES[seg]
        src_take = os.path.join(TAKES, take)
        want = need[seg] + 0.4
        pieces = D.fit(src_take, a, b, want, min_dead)
        if pieces is None:
            loose = D.plan(src_take, min_dead, 4.5, 24.0, 1.0, a, b)
            short[seg] = (want, sum(p["len"] for p in loose))
            continue
        dst = os.path.join(PUB, "clips", f"{seg}.mp4")
        D.render(src_take, dst, pieces, 18)
        dense[seg] = (dst, D.dur(dst))
        ramps = [p for p in pieces if p["kind"] == "ramp"]
        print(f"  {seg}: {b-a:6.1f}s -> {dense[seg][1]:5.1f}s"
              + (f"  ({len(ramps)} waits at {ramps[0]['len']:.1f}s)" if ramps else ""))
    if short:
        print("footage too short for the words over it:")
        for k, (want, have) in short.items():
            print(f"  {k}: needs {want:.1f}s, has at most {have:.1f}s")
        return 1

    slides, cursor = [], {}
    for n, (_, _, s) in enumerate(SEGS):
        if s["k"] != "shot":
            slides.append(s)
            continue
        take = s["take"]
        at = cursor.get(take, 0.0)
        clip = os.path.join(PUB, "clips", f"{take}_{n:02d}.mp4")
        subprocess.run(["ffmpeg", "-nostdin", "-y", "-loglevel", "error",
                        "-ss", f"{at:.3f}", "-i", dense[take][0],
                        "-t", f"{durs[n] + 0.25:.3f}", "-an", "-c:v", "libx264",
                        "-preset", "medium", "-crf", "18", "-pix_fmt", "yuv420p",
                        "-g", "60", clip], check=True)
        cursor[take] = at + durs[n]
        out = {k: v for k, v in s.items() if k != "take"}
        out["src"] = f"nocode02v2/clips/{os.path.basename(clip)}"
        slides.append(out)

    p = os.path.join(DECK, "l02.slides.json")
    json.dump(slides, open(p, "w"), indent=1, ensure_ascii=False)
    print(f"{len(slides)} slides, {sum(1 for s in slides if s['k']=='shot')} footage")
    print(f"  -> {os.path.relpath(p, ROOT)}")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--manifest", action="store_true")
    ap.add_argument("--slides", action="store_true")
    ap.add_argument("--min-dead", type=float, default=6.0)
    a = ap.parse_args()
    if a.manifest:
        return write_manifest()
    if a.slides:
        return write_slides(a.min_dead)
    ap.print_help()
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
