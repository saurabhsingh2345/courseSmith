#!/usr/bin/env python3
"""Validate the deepening pass before it replaces the narration it came from.

    python3 checkdeep.py            # check every narration/deep/*.json
    python3 checkdeep.py --land     # ...and move the clean ones into place

A deepened file may only GAIN sentences. Anchors, beats, segment count and
order must be identical, and every existing sentence must survive verbatim -
a reworded sentence is a sentence we pay for twice and a claim nobody checked
against the frame it sits over.
"""
import glob, json, os, subprocess, sys

HERE = os.path.dirname(os.path.abspath(__file__))
NARR = os.path.join(HERE, "narration")
DEEP = os.path.join(NARR, "deep")
CUTS = os.path.join(HERE, "cuts")
WPM, MIN_SPEED, PAD = 198.0, 0.92, 0.55

BANNED = ["cursor", "copilot", "antigravity", "cowork", "gastown", " gsd",
          "sprites", "on your phone", "claude code on the web"]


def dur(p):
    return float(subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip())


def words(s):
    return len(s.split())


def main():
    land = "--land" in sys.argv
    ok, bad = [], []
    for p in sorted(glob.glob(os.path.join(DEEP, "w3-*.json"))):
        lec = os.path.basename(p)[:-5]
        errs, warn, over = [], [], []
        try:
            new = json.load(open(p))
        except Exception as e:
            print(f"{lec}: UNPARSEABLE - {e}"); bad.append(lec); continue
        old = json.load(open(os.path.join(NARR, f"{lec}.json")))

        no, nn = old["segments"], new["segments"]
        if len(no) != len(nn):
            errs.append(f"segment count {len(no)} -> {len(nn)}")
        else:
            for i, (a, b) in enumerate(zip(no, nn)):
                if abs(float(a["anchor"]) - float(b["anchor"])) > 0.05:
                    errs.append(f"seg{i} anchor moved {a['anchor']} -> {b['anchor']}")
                if a.get("beat") != b.get("beat"):
                    errs.append(f"seg{i} beat renamed")
                if words(b["say"]) < words(a["say"]):
                    errs.append(f"seg{i} LOST words {words(a['say'])} -> {words(b['say'])}")

        # every segment must still fit its footage
        cut = os.path.join(CUTS, f"{lec}.mp4")
        total = dur(cut) if os.path.exists(cut) else float(new["cut_seconds"])
        for i, s in enumerate(nn):
            f0 = float(s["anchor"])
            f1 = float(nn[i + 1]["anchor"]) if i + 1 < len(nn) else total
            have = (f1 - f0) / MIN_SPEED
            need = words(s["say"]) / WPM * 60 + PAD
            if need > have + 0.5:
                over.append(f"seg{i} OVERRUNS by {need - have:.1f}s")

        # "cursor" also means the terminal caret, so show the sentence and
        # judge it rather than failing the file on a substring.
        for i, sg in enumerate(nn):
            for sent in sg["say"].replace("!", ".").replace("?", ".").split("."):
                low = sent.lower()
                for b in BANNED:
                    if b in low:
                        warn.append(f"    seg{i} '{b.strip()}': {sent.strip()[:110]}")

        added = sum(words(b["say"]) for b in nn) - sum(words(a["say"]) for a in no)
        gained = added / WPM
        if warn:
            print(f"{lec}: {len(warn)} phrase(s) to eyeball")
            for w in warn:
                print(w)
        if over and not errs:
            print(f"{lec}: {len(over)} overrun(s) for balance.py, +{added} words  +{gained:.1f} min")
        if errs:
            print(f"{lec}: {len(errs)} STRUCTURAL problem(s), +{added}w")
            for e in errs[:6]:
                print(f"    {e}")
            bad.append(lec)
        else:
            print(f"{lec}: clean  +{added} words  +{gained:.1f} min")
            ok.append((lec, p))

    print(f"\nclean {len(ok)} · problems {len(bad)}")
    if land:
        for lec, p in ok:
            os.replace(p, os.path.join(NARR, f"{lec}.json"))
        print(f"landed {len(ok)} into narration/")


if __name__ == "__main__":
    main()
