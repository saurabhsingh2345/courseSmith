# QA harness for the delivered lectures

Built 2026-09-02 for the full review of all 61 masters in `videos/nocode/vids/`.
Findings are written up at the artifact linked from the memory note
`coursesmith-lecture-review.md`.

## The tools

| script | what it answers |
|---|---|
| `holds.py <files>` | How long does the picture sit unchanged? Decodes at 2fps on a 64x36 luma and times the runs where nothing changes. **This is the one that found the real problem** — 25 of 35 w3 lectures are >70% frozen. |
| `flash.py <files>` | Where is light-theme footage cut into a dark lecture? Flags sustained frames over luma 190. Also prints each file's baseline brightness, which is how the nocode01/02 light-theme problem got quantified (215/233 vs 11-24). |
| `sheet.py <video> <prefix> <n>` | Contact sheets: n timestamped frames tiled 2-up at 768x432. Readable enough to judge a slide; use a crop for terminal text. |
| `triage.py <prefix> <files>` | One row per lecture, six frames across. For sweeping many lectures at once — catches dialogs, overlays, wasted framing. Not enough resolution to read a terminal. |
| `defects.sh` | silencedetect + astats over every file: dead air, peak, RMS. |

## Traps found while building these

- **Do not use `&` inside a backgrounded Bash tool call.** Double-backgrounding kills the child; both the first sheet run and the first defect scan died silently partway. `setsid` also does not exist on macOS.
- **`flash.py`'s first threshold (`luma > median+45 & std<40`) was ~90% false positives** — it fired on any slide with a bright headline. Only the absolute test (`luma > 190`, held >=1s) is trustworthy.
- **An "overlay in the top band" detector does not work on these decks** — the headline lives in that band. The w3-01 fullscreen banner and volume HUD had to be found by eye.
- **A single sampled frame is not evidence of a blank slide.** nocode11 @8:46 and nocode08 @2:49 both looked broken in a contact sheet and were mid-transition; sample every 2-4s around a suspect before calling it.
- 4K decode is slow: the w3 `holds.py` pass takes ~25 min for 35 files. Run it detached and let it land.
