#!/usr/bin/env python3
"""What would rendering cost, in characters, against what is left.

    python3 quota.py                 # every lecture with a cut
    python3 quota.py w3-17 w3-08

The voice is a metered API with a monthly allowance, and a deepening pass that
rewrites twenty lectures can spend a month's worth in one run. Nothing here
calls the API: it reads the sidecar text files beside every clip already voiced,
works out which sentences are genuinely new, and prices only those.

The sidecars are also why the cache should be content-addressed rather than
keyed on `<lecture>_<segment>_<index>`. Inserting one sentence into segment two
renumbers every clip after it, and every one of those is then re-spoken at full
price even though the words did not change. `voicecache.py` fixes that; this
reports the difference.
"""

from __future__ import annotations

import glob
import json
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import adam_lay as A  # noqa: E402

CUTS = os.path.join(HERE, "cuts")
NARR = os.path.join(HERE, "narration")
WORK = os.path.join(HERE, "_assemble")


class Paid:
    """Membership test for "this exact sentence has already been bought".

    The content store files clips under a hash of their text, so it cannot be
    enumerated back into sentences - it has to be asked. Sidecars are read too,
    because clips voiced before the store existed only live beside their mp3.
    Anything `adam_lay.speak()` would find, this finds.
    """

    def __init__(self) -> None:
        self.texts: set[str] = set()
        for side in glob.glob(os.path.join(WORK, "*", "raw", "*.mp3.txt")):
            mp3 = side[:-4]
            if os.path.exists(mp3) and os.path.getsize(mp3) > 1500:
                try:
                    self.texts.add(open(side, encoding="utf-8").read())
                except OSError:
                    pass

    def __contains__(self, text: str) -> bool:
        if text in self.texts:
            return True
        keep = A._store_path(text)
        return os.path.exists(keep) and os.path.getsize(keep) > 1500

    def __len__(self) -> int:
        return len(self.texts) + len(glob.glob(os.path.join(A.STORE, "*.mp3")))


def spoken() -> "Paid":
    return Paid()


def main() -> None:
    want = [a for a in sys.argv[1:] if a.startswith("w3-")]
    if not want:
        want = sorted({f[:-4] for f in os.listdir(CUTS)
                       if f.endswith(".mp4") and ".orig" not in f})
    have = spoken()
    print(f"{len(have)} sentence(s) already voiced\n")

    grand_new = grand_naive = 0
    for lec in want:
        p = os.path.join(NARR, f"{lec}.json")
        if not os.path.exists(p):
            continue
        d = json.load(open(p))
        new = naive = 0
        for sg in d["segments"]:
            for s in A.sentences(sg["say"]):
                naive += len(s)
                if s not in have:
                    new += len(s)
        grand_new += new
        grand_naive += naive
        flag = "" if new == 0 else f"   <- {new:6d} new"
        print(f"{lec}  {naive:7d} chars of narration{flag}")

    print(f"\ntotal narration      {grand_naive:8d} chars")
    print(f"already paid for     {grand_naive - grand_new:8d}")
    print(f"WOULD SPEND          {grand_new:8d}")

    if "--live" in sys.argv:
        import json as _j
        import urllib.request
        req = urllib.request.Request(
            "https://api.elevenlabs.io/v1/user/subscription",
            headers={"xi-api-key": A.KEY})
        sub = _j.load(urllib.request.urlopen(req, timeout=30))
        left = sub["character_limit"] - sub["character_count"]
        print(f"\nallowance left       {left:8d}")
        print(f"after this render    {left - grand_new:8d}"
              + ("" if left - grand_new >= 0 else "   <- NOT ENOUGH"))


if __name__ == "__main__":
    main()
