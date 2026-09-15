#!/usr/bin/env python3
"""Add cutaways to an existing plan without reformatting it.

The plans are hand-laid out - `secs` inline, a slide's keys grouped two or
three to a line - because they are read far more often than they are written,
and `json.dump` would flatten all thirty-five of them on the first second-pass
edit. So this splices a block of new inserts in as TEXT, at the position its
`at` belongs, and leaves every existing line byte-identical.

    python3 splice.py w3-35 new/w3-35.json

The block file is a JSON list of inserts. Ordering, overlap and dead-span
containment are still `render.py --dry`'s job; this only puts them in sequence.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
PLANS = os.path.join(HERE, "plans")
IND = " " * 4


def items(text: str) -> tuple[int, int, list[tuple[int, int]]]:
    """(start, end, spans) of the top-level objects inside "inserts": [ ... ]."""
    m = re.search(r'"inserts"\s*:\s*\[', text)
    if not m:
        raise SystemExit("no inserts array")
        
    i = m.end()
    depth = 0
    start = None
    spans = []
    while i < len(text):
        c = text[i]
        if c == '"':                      # skip strings, braces live in them
            i += 1
            while i < len(text) and text[i] != '"':
                i += 2 if text[i] == "\\" else 1
        elif c == "{":
            if depth == 0:
                start = i
            depth += 1
        elif c == "}":
            depth -= 1
            if depth == 0:
                spans.append((start, i + 1))
        elif c == "]" and depth == 0:
            return m.end(), i, spans
        i += 1
    raise SystemExit("unterminated inserts array")


def fmt(item: dict) -> str:
    """One insert, in the plans' own layout."""
    out = [IND + "{"]
    out.append(f'{IND}  "at": {float(item["at"]):g},')
    out.append(f'{IND}  "role": {json.dumps(item["role"])},')
    secs = ", ".join(f"{float(s):g}" for s in item["secs"])
    out.append(f'{IND}  "secs": [{secs}],')
    out.append(f'{IND}  "why": {json.dumps(item["why"], ensure_ascii=False)},')
    out.append(f'{IND}  "slides": [')
    for n, s in enumerate(item["slides"]):
        body = json.dumps(s, ensure_ascii=False)
        tail = "" if n == len(item["slides"]) - 1 else ","
        out.append(f'{IND}    {body}{tail}')
    out.append(f'{IND}  ]')
    out.append(IND + "}")
    return "\n".join(out)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("lecture")
    ap.add_argument("block")
    args = ap.parse_args()

    path = os.path.join(PLANS, f"{args.lecture}.json")
    text = open(path).read()
    new = json.load(open(args.block))
    if isinstance(new, dict):
        new = [new]

    for item in sorted(new, key=lambda x: float(x["at"])):
        _, close, spans = items(text)
        ats = [float(json.loads(text[a:b])["at"]) for a, b in spans]
        at = float(item["at"])
        k = next((i for i, x in enumerate(ats) if x > at), len(ats))
        block = fmt(item)
        if k == len(spans):                      # after the last one
            a, b = spans[-1]
            text = text[:b] + ",\n" + block + text[b:]
        else:
            a, _ = spans[k]
            line = text.rfind("\n", 0, a) + 1    # keep that object's indent
            text = text[:line] + block + ",\n" + text[line:]

    json.loads(text)                             # never write a broken plan
    open(path, "w").write(text)
    n = len(items(text)[2])
    print(f"{args.lecture}: +{len(new)} inserts, {n} total")
    return 0


if __name__ == "__main__":
    sys.exit(main())
