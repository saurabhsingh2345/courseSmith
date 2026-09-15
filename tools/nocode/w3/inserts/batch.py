#!/usr/bin/env python3
"""Run one lecture's whole animation pass, or a list of them.

    python3 batch.py w3-36 w3-37 ...
    python3 batch.py --all              # every lecture with a plan
    python3 batch.py --check w3-37      # validate plans only, encode nothing

Per lecture: keep the untouched master, re-measure the dead spans and the sync
leads of the current source, check the plan against both, render the cutaways,
lay them on, and print the holds figure before and after. Nothing is delivered -
`lay.py --deliver` is a separate, deliberate step.

A failure stops that lecture and moves on, because the useful thing after a long
batch is knowing which ones still need work, not losing the ones that succeeded.
"""

from __future__ import annotations

import argparse
import glob
import json
import os
import shutil
import subprocess
import sys

sys.stdout.reconfigure(line_buffering=True)

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import source as SRC  # noqa: E402

PLANS = os.path.join(HERE, "plans")
ORIG = os.path.join(HERE, "orig")
STAGED = os.path.join(HERE, "staged")
VIDS = os.path.abspath(os.path.join(HERE, "..", "..", "..", "..",
                                    "videos", "nocode", "vids"))
HOLDS = os.path.abspath(os.path.join(HERE, "..", "..", "qa", "holds.py"))
PY = sys.executable


def run(args: list[str], **kw) -> subprocess.CompletedProcess:
    return subprocess.run(args, cwd=HERE, **kw)


def holds(path: str) -> str:
    out = subprocess.run([PY, HOLDS, path], capture_output=True, text=True).stdout
    return out.strip().split("\t", 1)[1] if "\t" in out else out.strip()


def one(lec: str, check_only: bool) -> bool:
    print(f"\n=== {lec} " + "=" * (60 - len(lec)))
    plan = os.path.join(PLANS, f"{lec}.json")
    if not os.path.exists(plan):
        print("  no plan")
        return False

    # Nothing is copied here. `lay.py --deliver` moves the master aside when a
    # cut is actually delivered; until then vids/ is the untouched source and
    # duplicating 11 GB of it up front is what filled the disk.
    keep = os.path.join(ORIG, f"{lec}_adam.mp4")
    master = keep if os.path.exists(keep) else os.path.join(VIDS, f"{lec}_adam.mp4")

    # A second pass re-measures a master that has not changed, which is two
    # full 4K decodes for a map that is already on disk - and worse, a map that
    # might come out slightly different from the one the plan was authored
    # against. Re-measure only when the map is missing or older than the file.
    m = SRC.map_path(lec)
    if not os.path.exists(m) or os.path.getmtime(m) < os.path.getmtime(master):
        SRC.remap(lec)
    else:
        print(f"  map: {os.path.relpath(m, HERE)} (measured already, reused)")
    r = run([PY, "render.py", lec, "--dry"], capture_output=True, text=True)
    print(r.stdout.rstrip() or r.stderr.rstrip())
    if r.returncode:
        return False
    if check_only:
        return True

    for f in glob.glob(os.path.join(HERE, "out", f"{lec}_*")):
        os.remove(f)
    r = run([PY, "render.py", lec], capture_output=True, text=True)
    if r.returncode:
        print("  render failed:", (r.stderr or r.stdout).strip()[-2000:])
        return False
    r = run([PY, "lay.py", lec], capture_output=True, text=True)
    if r.returncode:
        print("  lay failed:", (r.stderr or r.stdout).strip()[-2000:])
        return False
    print("  " + r.stdout.strip().splitlines()[-1])

    # holds() on the master and on the cut is two more 4K decodes for a number
    # verify.py prints anyway, next to the sharp one that is the actual result.
    if os.environ.get("BATCH_HOLDS"):
        print(f"  before: {holds(master)}")
        print(f"  after:  {holds(os.path.join(STAGED, f'{lec}_adam.mp4'))}")

    # The strips are ~150 MB of 4K per lecture and are reproducible from the
    # plan, so they do not need to survive the run.
    for f in glob.glob(os.path.join(HERE, "out", f"{lec}_*")):
        os.remove(f)
    return True


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("lectures", nargs="*")
    ap.add_argument("--all", action="store_true")
    ap.add_argument("--check", action="store_true")
    args = ap.parse_args()

    lecs = args.lectures
    if args.all:
        lecs = sorted(os.path.basename(p)[:-5]
                      for p in glob.glob(os.path.join(PLANS, "*.json")))
    ok, bad = [], []
    for lec in lecs:
        (ok if one(lec, args.check) else bad).append(lec)
    print(f"\ndone: {len(ok)} ok" + (f", FAILED: {', '.join(bad)}" if bad else ""))
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
