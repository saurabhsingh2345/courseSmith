#!/usr/bin/env python3
"""Scaffold one Week 3 part in the sticker deck kit.

    python3 newpart.py w3-03 "Slash commands: a routine in one word"

Creates program/part-w3-03-slash-commands/ with a slides.html wired to the kit's
conventions, an empty vo/script.json, and a generate_vo.py that reads the key
from where it already lives instead of copying a secret into the repo root.

The deck is deliberately small. Week 3 averages 91% footage; a part that needs
more than about six slides is a part where we should be filming instead.
"""

from __future__ import annotations

import json
import os
import re
import shutil
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
REPO = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
PROGRAM = os.path.join(REPO, "program")
KIT = os.path.join(PROGRAM, "part-01-the-missing-manual")
ENV = os.path.join(REPO, "tools", "nocode", "nc", "No-Code", ".env")


STOP = {"a", "an", "the", "in", "on", "of", "to", "and", "or", "for",
        "your", "one", "with", "it", "is", "at", "by", "that", "what"}


def slug(s: str) -> str:
    words = [w for w in re.sub(r"[^a-z0-9]+", " ", s.lower()).split() if w not in STOP]
    return "-".join(words[:3])


SLIDES = """<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>__ID__ __TITLE__</title>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
  <link href="https://fonts.googleapis.com/css2?family=Archivo+Black&family=IBM+Plex+Mono:wght@400;600;700&display=swap" rel="stylesheet" />
  <link rel="stylesheet" href="../assets/deck.css" />
</head>
<body>
  <div id="stage">
    <div id="gate"><span>Click to start &middot; Adam + SFX</span></div>
    <div class="veil" id="veil"></div>

    <!-- One hero mark per slide. Draw a subject only when the line needs it.
         Never stamp the hardware lineup on a slide that is not about hardware. -->

    <section class="slide on" data-vo="vo/01.mp3" data-enter="rise">
      <p class="kicker">__KICKER__</p>
      <h1 class="display xl">__TITLE__</h1>
      <div class="grow"></div>
    </section>

    <section class="slide" data-vo="vo/02.mp3" data-enter="whoosh">
      <p class="kicker">replace me</p>
      <h2 class="display lg">second beat</h2>
    </section>

  </div>
  <div class="hud"><span id="pos">1 / 1</span><span>__ID__</span></div>
  <script src="../assets/deck.js"></script>
</body>
</html>
"""

VO_PY = '''#!/usr/bin/env python3
"""Speak this part with Adam. Reads the key from the rig's existing .env."""
import json, os, sys, urllib.request, time
from pathlib import Path

ROOT = Path(__file__).resolve().parent
ENV = Path("__ENV__")
SPEC = ROOT / "vo" / "script.json"


def env():
    d = {}
    for line in ENV.read_text().splitlines():
        if "=" in line and not line.strip().startswith("#"):
            k, v = line.split("=", 1)
            d[k.strip()] = v.strip().strip('"').strip("'")
    return d


def main():
    e = env()
    key, voice = e["ELEVENLABS_API_KEY"], e["ELEVENLABS_VOICE_ID"]
    spec = json.loads(SPEC.read_text())
    for s in spec["slides"]:
        dest = ROOT / "vo" / s["file"]
        if dest.exists() and dest.stat().st_size > 2000:
            # The cache is keyed on the TEXT, not the filename: editing a line
            # and reusing the name used to ship the old take.
            stamp = ROOT / "vo" / (s["file"] + ".txt")
            if stamp.exists() and stamp.read_text() == s["text"]:
                print("cached", s["file"]); continue
        body = json.dumps({
            "text": s["text"],
            "model_id": spec.get("model_id", "eleven_multilingual_v2"),
            "voice_settings": {"stability": 0.42, "similarity_boost": 0.8,
                               "style": 0.15, "use_speaker_boost": True},
        }).encode()
        req = urllib.request.Request(
            f"https://api.elevenlabs.io/v1/text-to-speech/{voice}",
            data=body, method="POST",
            headers={"xi-api-key": key, "Content-Type": "application/json",
                     "Accept": "audio/mpeg"})
        for attempt in range(5):
            try:
                with urllib.request.urlopen(req, timeout=120) as r:
                    dest.write_bytes(r.read())
                break
            except Exception as exc:
                if attempt == 4:
                    raise
                print(f"  retry {attempt+1} after {exc}")
                time.sleep(2 ** attempt)
        (ROOT / "vo" / (s["file"] + ".txt")).write_text(s["text"])
        print("wrote", s["file"], dest.stat().st_size // 1024, "KB")


if __name__ == "__main__":
    main()
'''


def main() -> None:
    if len(sys.argv) < 3:
        sys.exit(__doc__)
    pid, title = sys.argv[1], sys.argv[2]
    name = f"part-{pid}-{slug(title)}"
    dst = os.path.join(PROGRAM, name)
    if os.path.exists(dst):
        sys.exit(f"{name} already exists")
    os.makedirs(os.path.join(dst, "vo"))
    os.makedirs(os.path.join(dst, "images"))

    kicker = {"w3": "WEEK 3"}.get(pid.split("-")[0], "WEEK 3")
    html = (SLIDES.replace("__ID__", pid)
                  .replace("__TITLE__", title)
                  .replace("__KICKER__", kicker))
    open(os.path.join(dst, "slides.html"), "w").write(html)
    open(os.path.join(dst, "generate_vo.py"), "w").write(VO_PY.replace("__ENV__", ENV))
    os.chmod(os.path.join(dst, "generate_vo.py"), 0o755)

    json.dump({"program": "AI Coder — no code", "section": f"{pid} {title}",
               "model_id": "eleven_multilingual_v2", "slides": [
                   {"id": "01", "file": "01.mp3", "text": "First line."},
                   {"id": "02", "file": "02.mp3", "text": "Second line."}]},
              open(os.path.join(dst, "vo", "script.json"), "w"), indent=2)

    for f in ("sfx-whoosh.mp3", "sfx-hit.mp3", "sfx-stamp.mp3",
              "sfx-tick.mp3", "sfx-rise.mp3"):
        src = os.path.join(KIT, "vo", f)
        if os.path.exists(src):
            shutil.copy2(src, os.path.join(dst, "vo", f))

    open(os.path.join(dst, "shot-list.txt"), "w").write(
        f"{pid.upper()} — {title}\n\n"
        "SLIDE ANIMATION   vo/script.json slides\n"
        "SCREEN RECORDING  from tools/nocode/w3/cuts/" + pid + ".mp4\n"
        "  After the cut to footage, do not go back to slides.\n")
    print("created", dst)


if __name__ == "__main__":
    main()
