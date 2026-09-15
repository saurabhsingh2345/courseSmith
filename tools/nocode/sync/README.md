# Two Macs, one program

Machine A cuts Day 2 (L11-L14), machine B cuts Day 3. This directory keeps the
two in step. Only the no-code program is covered — none of the Claude-tutorial,
remaster or social tracks are in the kit, and the films do not need Go, Kokoro,
Docker or a Groq key.

## First time on the new Mac

On this machine:

    tools/nocode/sync/pack.sh            # ~/Desktop/nocode-kit-<stamp>.tar.gz, 2.8 MB

AirDrop it over. On the new Mac, from anywhere:

    tar -xzf nocode-kit-<stamp>.tar.gz
    nocode-kit/sync/unpack.sh nocode-kit

It clones the repo, checks out `v10-atelier-and-the-101`, drops the untracked
source in, installs `renderer/node_modules` and the python deps, writes the
memory, and finishes with a checklist. Re-run the checks any time with
`tools/nocode/sync/unpack.sh --check`.

Then open Claude in the repo and say **"let's finish that Udemy program"** — it
reads `coursesmith-nocode-RESUME` and picks up where this machine left off.

## Every day after

Whoever finished a lecture packs the brain and sends it:

    tools/nocode/sync/pack.sh --memory-only     # ~800 KB: memory + docs + markers
    # other machine:
    tools/nocode/sync/unpack.sh <kit>

`unpack.sh` merges rather than overwrites, but `MEMORY.md` is one file both
machines edit. If it differs you get `MEMORY.md.incoming` beside it and a
warning — paste the new index lines across by hand. Everything else (one fact
per file) merges cleanly.

Finished films are 40-140 MB each and are not in the kit. Send those separately
when a lecture is done; `videos/` is gitignored on purpose.

## What is in the kit, and why

| | why it cannot be regenerated |
|---|---|
| `memory/` (42 files) | the whole doctrine: pacing, gap discipline, house rules, what shipped |
| `docs/nocode-course/` | his transcripts, `ref-L01..L34`, `PLAN.md`, `curriculum.md` |
| `videos/nocode/markers/` | every deviation from his lectures, per film |
| `nc/No-Code/.env` | ElevenLabs key **and the Adam voice id** |
| `renderer/src/nocode/` | `kit.tsx` and the `l03..l10` decks |
| `tools/nocode/scripts/` | `adam_deck.py`, `adam_lay.py`, `stage.py`, `drive.py`, the `shoot_*.py` rig |
| `~/arena/` | the game the lectures build on screen, carried lecture to lecture |

Left behind deliberately: 1.3 GB of takes, 449 MB of finished films,
`_archive/`, `node_modules`, `.voice` caches, and the 94 MB extracted No-Code
pack (`pack.sh --with-pack` adds the zip if you are doing template design).

## The three things that break a second machine

1. **Screen geometry.** Footage from a different display will not intercut with
   Day 1-2. This machine is a Mac15,7, built-in Retina 3456x2234, and shoots
   are cropped against that. If the other Mac's built-in panel differs, set it
   to the same scaled resolution before the first shoot and re-probe the crop
   constants — do not assume Day 1's numbers.
2. **Permissions.** Screen Recording *and* Accessibility for whatever terminal
   runs the shoot. Without Accessibility the pyobjc CGEvents silently do
   nothing and you get footage of a still screen.
3. **The capture index.** `Capture screen 0` is not always the built-in one.
   Probe it (`unpack.sh --check` prints the list) every time you shoot on a
   machine whose display setup changed.

## Keeping the work from colliding

One machine owns a lecture end to end — script, shoot, voice, assemble. Never
two machines on the same lecture: the marker sheet and the deck file are both
per-lecture and there is no merge story for footage. Branch per lecture group
if you want the source in git (`nocode/L11-L14`), and keep `docs/nocode-course`
and the `.env` out of the public repo — the transcripts are his and the key is
a key.
