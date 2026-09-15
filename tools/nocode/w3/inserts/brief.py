#!/usr/bin/env python3
"""Everything needed to author one lecture's cutaways, on one screen.

    python3 brief.py w3-17

For each dead span it prints the words actually being spoken over it, because
that - not the lecture's topic - is what the card has to say. A card written
from the topic is how you get a deck of true-but-unrelated statements; a card
written from the sentence underneath it reads as the lesson continuing.

Also prints, per span, whether an anchor lands inside it. A span containing an
anchor is where the `open`/`beat` card should END: cut back on the word, and the
card's own out-point becomes the reveal.
"""

from __future__ import annotations

import argparse
import json
import os
import sys
import textwrap

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import source as SRC  # noqa: E402

NARR = os.path.abspath(os.path.join(HERE, "..", "narration"))


def words_over(segs: list[dict], a: float, b: float, total: float) -> str:
    """The narration audible between a and b.

    A segment is spoken from its anchor until the next one, so slice each
    overlapping segment proportionally by character count. Crude, and good
    enough to tell which sentences of a ninety-second segment land on a span.
    """
    out = []
    for i, s in enumerate(segs):
        s0 = float(s["anchor"])
        s1 = float(segs[i + 1]["anchor"]) if i + 1 < len(segs) else total
        if s1 <= a or s0 >= b:
            continue
        text = " ".join(s["say"].split())
        if s1 - s0 <= 0:
            out.append(text)
            continue
        f0 = max(0.0, (a - s0) / (s1 - s0))
        f1 = min(1.0, (b - s0) / (s1 - s0))
        c0, c1 = int(len(text) * f0), int(len(text) * f1)
        frag = text[c0:c1].strip()
        if frag:
            lead = "..." if c0 > 0 else ""
            tail = "..." if c1 < len(text) else ""
            out.append(f"{lead}{frag}{tail}")
    return "  ||  ".join(out)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("lecture")
    args = ap.parse_args()
    lec = args.lecture

    segs = json.load(open(os.path.join(NARR, f"{lec}.json")))["segments"]
    spans = SRC.spans(lec)
    anchors = SRC.anchors(lec)
    leads = dict(SRC.leads(lec))

    import subprocess
    total = float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", SRC.source(lec)], capture_output=True, text=True,
        check=True).stdout.strip())

    dead = sum(b - a for a, b in spans)
    print(f"{lec}   runtime {total:.0f}s   dead {dead:.0f}s ({dead / total * 100:.0f}%)"
          f"   {len(spans)} spans   target animated ~{total / 3:.0f}s")
    if leads:
        print("MUST end a cutaway just before: " +
              ", ".join(f"{a:.1f} (leads {l:.0f}s)" for a, l in leads.items()))
    print()
    for a, b in spans:
        hit = [x for x in anchors if a - 1 <= x <= b + 2]
        tag = ""
        if hit:
            tag = "   <- anchor " + ", ".join(f"{h:.1f}" for h in hit)
            if any(h in leads for h in hit):
                tag += "  *** END A CARD JUST BEFORE IT ***"
        print(f"[{a:7.1f} - {b:7.1f}]  {b - a:5.1f}s{tag}")
        said = words_over(segs, a, b, total)
        for line in textwrap.wrap(said, 100) or ["(silence)"]:
            print(f"    {line}")
        print()
    return 0


if __name__ == "__main__":
    sys.exit(main())
