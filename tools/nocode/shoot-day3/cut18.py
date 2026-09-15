#!/usr/bin/env python3
"""Cut the L18 takes into exact slide windows, masking the one piece of PII.

Same contract as tools/nocode/shoot06/cut.py — each clip's picture length is
retimed so it equals its slide's window (narration + gap), which means swapping
a graphic for footage costs no timing anywhere else in the film.

One addition: VS Code puts the signed-in GitHub account's AVATAR — his actual
photograph — in the bottom-left of the activity bar, and Copilot cannot be used
signed out, so it is in every frame of every L18 take. It is masked here rather
than blurred: the activity bar is a flat #181818, so a filled box in exactly
that colour reads as an empty slot instead of a smudge. Measured position at
1:1 points was x 14-35, y 974-997, which is x 28-70, y 1948-1994 in the
3840x2160 source; the box below carries margin on all four sides.
"""
import os, subprocess, sys

R = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith"
T = f"{R}/tools/nocode/scripts/takes"

# the avatar, in SOURCE pixels (3840x2160)
MASK = "drawbox=x=18:y=1934:w=68:h=76:color=0x181818:t=fill"


def dur(p):
    """Duration, or -1 if ffprobe cannot read the file.

    A span that runs past the end of its take produces an output ffprobe reports
    as 'N/A', which used to raise and abort the whole cut run partway through —
    leaving a half-cut lecture and a render that failed for no obvious reason.
    Now it returns -1 so the caller can flag exactly which clip is wrong."""
    out = subprocess.run(
        ["ffprobe", "-v", "error", "-show_entries", "format=duration",
         "-of", "csv=p=0", p], capture_output=True, text=True).stdout.strip()
    try:
        return float(out)
    except ValueError:
        return -1.0


def cut(lec, idx, take, ss, to, win, crop=None):
    """crop is (w,h,x,y) in source pixels, applied BEFORE the scale, for beats
    that need a zoom onto the chat panel rather than the whole screen.

    The avatar mask is applied ONLY to editor takes. The browser takes have no
    avatar to hide, and painting the activity-bar's #181818 into a Chrome frame
    left a faintly lighter square in the bottom-left corner of the board —
    (24,24,24) against a (15,15,15) background, subtle but real.
    """
    src = f"{T}/{take}.mp4"
    out = f"{R}/renderer/public/nocode{lec}/shots/{idx:02d}.mp4"
    os.makedirs(os.path.dirname(out), exist_ok=True)
    # Rate is computed against the window PLUS a pad, not the window itself.
    # With rate = win/(to-ss) the output is exactly `win` long in theory and one
    # or two frames short in practice, once ffmpeg rounds to whole frames — and
    # a clip a frame short of its slide is a black flash at the cut. Overshoot
    # by 0.3s instead and let the Sequence trim it; 0.3s over ~15s is a 2% speed
    # change, which is invisible.
    PAD = 0.30
    rate = (win + PAD) / (to - ss)
    # The mask covers ONE thing: the signed-in GitHub avatar (his photograph) in
    # VS Code's activity bar, bottom-left, which Copilot and the Claude Code
    # extension both require a login for. Only the L18 and L19 editor takes have
    # it. Antigravity puts no avatar there — painting #181818 into its activity
    # bar would just cover its own settings icons — and the browser takes have
    # nothing to hide either.
    # Every VS Code take carries the signed-in GitHub avatar — his photograph —
    # in the activity bar, bottom left. Antigravity does not (nothing is there to
    # hide), and the browser takes obviously do not.
    editor_take = ("checks" not in take
                   and any(take.startswith(p) for p in
                           ("L18_", "L19_", "L30_", "L31_", "L32_", "L33_", "L34_")))
    vf = [f"setpts=PTS*{rate:.6f}"] + ([MASK] if editor_take else [])
    if crop:
        w, h, x, y = crop
        vf.append(f"crop={w}:{h}:{x}:{y}")
    vf += ["scale=1920:1080:flags=lanczos", "fps=30"]
    # ask for a touch MORE than the window. setpts + -t rounding can land a
    # clip a frame or two short, and a short clip means black at the end of that
    # slide; the Sequence trims the surplus for free.
    cmd = ["ffmpeg", "-y", "-loglevel", "error", "-ss", str(ss), "-to", str(to),
           "-i", src, "-an", "-vf", ",".join(vf), "-t", f"{win + PAD:.3f}",
           "-c:v", "libx264", "-preset", "medium", "-crf", "18",
           "-pix_fmt", "yuv420p", "-g", "30", out]
    r = subprocess.run(cmd, capture_output=True, text=True)
    if r.returncode:
        print(f"FAIL {lec}/{idx}: {r.stderr[-400:]}")
        return False
    d = dur(out)
    if d < 0:
        print(f"  BAD nocode{lec}/shots/{idx:02d}.mp4 — unreadable. "
              f"{take} {ss:.1f}-{to:.1f}s probably runs past the end of the take")
        return False
    print(f"  nocode{lec}/shots/{idx:02d}.mp4  {d:5.2f}s (want {win:.2f})  "
          f"{take} {ss:.1f}-{to:.1f} @ {rate:.2f}x"
          f"{'  CROP' if crop else ''}{'' if editor_take else '  no-mask'}")
    return True


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "jobs":
        import json
        for j in json.load(open(sys.argv[2])):
            cut(*j[:6], crop=tuple(j[6]) if len(j) > 6 and j[6] else None)
    else:
        print(__doc__)
        print("usage: cut18.py jobs <jobs.json>   # [[lec,idx,take,ss,to,win,crop?],...]")
