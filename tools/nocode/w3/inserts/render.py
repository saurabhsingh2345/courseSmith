#!/usr/bin/env python3
"""Render one lecture's animated cutaways from its plan.

    python3 render.py w3-20            # every insert in the plan
    python3 render.py w3-20 --only 2   # just insert #2, after editing it

A plan is `plans/w3-NN.json`:

    {"lecture": "w3-20",
     "inserts": [
       {"at": 12.5, "secs": [8, 8, 8], "slides": [ ...Slide... ]}
     ]}

`at` is a time in the DELIVERED master and `secs` are the per-slide lengths, so
an insert's length is `sum(secs)` - there is no separate `dur` field to fall out
of step with it. The whole approach rests on that number matching the hole
`deadzones.py` found, which `check()` enforces here rather than at overlay time,
because a wrong length is far cheaper to catch before a 4K render than after.

Rendered at 1920x1080 with `--scale=2`: Remotion supersamples the frame instead
of upscaling it, which is the difference between crisp Archivo Black on a 4K
master and a soft one.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
PLANS = os.path.join(HERE, "plans")
sys.path.insert(0, HERE)
import source as SRC  # noqa: E402
OUT = os.path.join(HERE, "out")
RENDERER = os.path.abspath(os.path.join(HERE, "..", "..", "..", "..", "renderer"))
FPS = 30


def frames(secs: list[float]) -> int:
    """Must agree with `insertFrames` in w3insert.tsx, frame for frame."""
    return sum(round(s * FPS) for s in secs)


def dead_spans(lec: str) -> list[tuple[float, float]]:
    """Dead spans of the CURRENT source - retimed if resync.py has run."""
    return SRC.spans(lec)


ROLES = ("open", "beat", "close")
ENFORCE_LEAD = 8.0   # a lead smaller than this is jitter, not a defect


def check(plan: dict) -> list[str]:
    """Refuse a plan before it costs a 4K render.

    Three things are enforced, and each of them caught a real mistake:

    1. **Windows stay inside measured dead spans.** The first cut of w3-03
       covered 89.0-96.5, where the next request is being typed on camera,
       because it was planned against holds.py's coarse detector.
    2. **Every anchor the footage got ahead of is covered.** `align.py` lists the
       sentences that land on a picture which finished printing seconds earlier.
       A cutaway ending just before such an anchor turns its own cut-back into
       the reveal, so the picture arrives with the words. This is the sync fix -
       retiming the video was tried and moved every other anchor with it.
    3. **The lecture has a tutorial shape.** First cutaway `open` (what this
       lecture is and what you will do), last `close` (what you now know),
       everything between `beat`. Without that the cards read as interruptions
       rather than as a lesson with a spine.
    """
    bad = []
    lec = plan["lecture"]
    spans = SRC.spans(lec)
    anchors = SRC.anchors(lec)
    if not spans:
        bad.append(f"no map for {lec} - run: python3 source.py {lec} --remap")

    ins = plan["inserts"]
    roles = [i.get("role") for i in ins]
    if any(r not in ROLES for r in roles):
        bad.append(f"every insert needs a role from {ROLES}; got {roles}")
    elif len(roles) == 1:
        if roles[0] != "open":
            bad.append("a single-insert plan's one insert must be the open")
    elif roles[0] != "open" or roles[-1] != "close" or "open" in roles[1:] \
            or "close" in roles[:-1]:
        bad.append(f"the shape must be open, beat.., close; got {roles}")

    last_end = -1.0
    ends = []
    for i, item in enumerate(ins):
        secs = item["secs"]
        if len(secs) != len(item["slides"]):
            bad.append(f"#{i}: {len(secs)} secs for {len(item['slides'])} slides")
        if min(secs, default=0) < 2.0:
            bad.append(f"#{i}: a slide under 2s is a flash, not a beat")
        at = float(item["at"])
        end = at + frames(secs) / FPS
        ends.append(end)
        if at < last_end:
            bad.append(f"#{i}: starts at {at:.1f}s, inside the previous insert")
        last_end = end

        # Consecutive spans that touch count as one covered stretch, because a
        # span boundary can be nothing more than a cursor moving.
        reach = None
        for a, b in spans:
            if reach is None:
                if a <= at < b:
                    reach = b
            elif a <= reach + 0.75:
                reach = max(reach, b)
        if reach is None:
            bad.append(f"#{i}: starts at {at:.1f}s, which is not in a dead span")
        elif end > reach + 0.01:
            bad.append(f"#{i}: covers to {end:.1f}s but the dead picture ends "
                       f"at {reach:.1f}s - {end - reach:.1f}s of live footage "
                       f"would be hidden")

    # Enforced above ENFORCE_LEAD only. A four-second lead is not what makes a
    # lecture feel out of step, and it often cannot be fixed anyway: there may
    # be no dead frame in the three seconds before that anchor to end a card on.
    # Smaller ones are printed by source.py and left to judgement.
    waived = set(plan.get("sync_waived") or [])
    for anchor, lead in SRC.leads(lec):
        if lead < ENFORCE_LEAD or anchor in waived:
            continue
        if not any(anchor - 3.0 <= e <= anchor + 1.5 for e in ends):
            bad.append(f"the footage reaches anchor {anchor:.1f} {lead:.1f}s early "
                       f"and no cutaway ends just before it - the words will land "
                       f"on a picture that has already finished. Cover it, or waive "
                       f"it in the plan's sync_waived with a reason.")
    return bad


def report(plan: dict) -> None:
    """Where each cut-back lands relative to the sentence it hands over to."""
    anchors = SRC.anchors(plan["lecture"])
    for i, item in enumerate(plan["inserts"]):
        end = float(item["at"]) + frames(item["secs"]) / FPS
        nxt = [a for a in anchors if a >= end - 0.5]
        gap = f"{nxt[0] - end:+.1f}s to anchor {nxt[0]:.1f}" if nxt else "no anchor after"
        print(f"  #{i:02d} {item.get('role','?'):5} {item['at']:7.1f}-{end:7.1f}  "
              f"cut back {gap}")


CHUNK_FRAMES = 120 * FPS   # a strip past ~140s exits Remotion with code 1


def chunks(plan: dict) -> list[list[int]]:
    """Group the inserts into strips of at most CHUNK_FRAMES each.

    One strip per lecture was the right idea and the wrong size: seven lectures
    have more than ~140s of cutaways and Remotion exits 1 rendering them as one
    composition (w3-29 is 578s). Inserts are grouped in plan order, so a chunk
    is a contiguous run and the frame offsets inside it stay cumulative.
    """
    out, cur, at = [], [], 0
    for i, item in enumerate(plan["inserts"]):
        n = frames(item["secs"])
        if cur and at + n > CHUNK_FRAMES:
            out.append(cur)
            cur, at = [], 0
        cur.append(i)
        at += n
    if cur:
        out.append(cur)
    return out


def render_strip(lec: str, plan: dict) -> list[str]:
    """Render every cutaway in the lecture as a few strips.

    Remotion bundles the project on each invocation, so a lecture with eighteen
    inserts paid for eighteen bundles and eighteen Chrome launches. A strip is
    one bundle, and `lay.py` trims the pieces back out by frame offset.

    Offsets are cumulative `round(sec * FPS)` within the strip - exactly what
    `insertFrames` sums in w3insert.tsx - so the arithmetic here and in the
    composition cannot drift. The manifest records which strip each insert is
    in as well as its offset inside it.
    """
    ins = plan["inserts"]
    groups = chunks(plan)
    chunk_of = [0] * len(ins)
    offs = [0] * len(ins)
    files = []
    for c, group in enumerate(groups):
        slides, secs, at = [], [], 0
        for i in group:
            chunk_of[i] = c
            offs[i] = at
            slides += ins[i]["slides"]
            secs += ins[i]["secs"]
            at += frames(ins[i]["secs"])
        props = os.path.join(OUT, f"{lec}_strip_{c:02d}.props.json")
        with open(props, "w") as fh:
            json.dump({"slides": slides, "secs": secs}, fh)
        dst = os.path.join(OUT, f"{lec}_strip_{c:02d}.mp4")
        print(f"  strip {c + 1}/{len(groups)}: inserts {group[0]}-{group[-1]}, "
              f"{at / FPS:.0f}s", flush=True)
        subprocess.run(
            ["npx", "remotion", "render", "src/nocode/index.tsx", "W3Insert", dst,
             f"--props={props}", "--scale=2", "--log=error"],
            cwd=RENDERER, check=True)
        files.append(os.path.basename(dst))
    with open(os.path.join(OUT, f"{lec}_strip.json"), "w") as fh:
        json.dump({"chunks": files, "chunk": chunk_of, "offsets": offs,
                   "lengths": [frames(i["secs"]) for i in ins]}, fh)
    return files


def render(lec: str, ins: dict, i: int) -> str:
    props = os.path.join(OUT, f"{lec}_{i:02d}.props.json")
    with open(props, "w") as fh:
        json.dump({"slides": ins["slides"], "secs": ins["secs"]}, fh)
    dst = os.path.join(OUT, f"{lec}_{i:02d}.mp4")
    subprocess.run(
        ["npx", "remotion", "render", "src/nocode/index.tsx", "W3Insert", dst,
         f"--props={props}", "--scale=2", "--log=error"],
        cwd=RENDERER, check=True)
    return dst


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("lecture")
    ap.add_argument("--only", type=int, default=None,
                    help="render just this insert index")
    ap.add_argument("--dry", action="store_true",
                    help="check the plan and print the cut-backs, render nothing")
    args = ap.parse_args()

    plan = json.load(open(os.path.join(PLANS, f"{args.lecture}.json")))
    bad = check(plan)
    if bad:
        print("plan is not renderable:", *bad, sep="\n  ")
        return 1

    report(plan)
    if args.dry:
        return 0

    os.makedirs(OUT, exist_ok=True)
    if args.only is None:
        n = sum(frames(i["secs"]) for i in plan["inserts"])
        print(f"{args.lecture}  {len(chunks(plan))} strip(s): "
              f"{len(plan['inserts'])} inserts, {n / FPS:.0f}s, {n} frames",
              flush=True)
        render_strip(args.lecture, plan)
        return 0

    for i, ins in enumerate(plan["inserts"]):
        if args.only is not None and i != args.only:
            continue
        n = frames(ins["secs"])
        print(f"{args.lecture} #{i:02d}  at {ins['at']:7.1f}s  "
              f"{n / FPS:5.1f}s  {len(ins['slides'])} slides", flush=True)
        render(args.lecture, ins, i)
    return 0


if __name__ == "__main__":
    sys.exit(main())
