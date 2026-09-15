#!/usr/bin/env python3
"""Shared staging for the nocode01 screen takes.

Everything is shot on the external 16:9 panel (display id 2, avfoundation
index 2, 3840x2160 px / 1920x1080 pt, origin -1920,-85). Capturing a true 16:9
panel is what keeps the finished frame free of side flanks: no letterboxing,
no crop, a clean 2x downscale to 1080p.
"""
import os, subprocess, sys, time, threading
import Quartz as Q

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.dirname(HERE))
import drive

def _detect_capture_index():
    """avfoundation screen indices follow the display arrangement, so probe for
    the 16:9 panel every run rather than trusting a number from last time.

    Two guards, both learned the hard way on 2026-09-02: an orphaned recorder
    ffmpeg from an abandoned take still holds the capture device, and then every
    probe here blocks FOREVER — a shoot sat on an empty screen for half an hour
    before anyone worked out why. So each probe is given a deadline, and
    CAP_INDEX in the environment skips probing altogether.
    """
    import tempfile
    if os.environ.get("CAP_INDEX"):
        print(f"[stage] capture index {os.environ['CAP_INDEX']} (from CAP_INDEX)")
        return os.environ["CAP_INDEX"]
    for i in ("1", "2", "3"):
        out = os.path.join(tempfile.gettempdir(), f"_capprobe{i}.mp4")
        try:
            subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-f", "avfoundation",
                            "-capture_cursor", "0", "-framerate", "30", "-i", i, "-t", "1",
                            "-c:v", "h264_videotoolbox", "-b:v", "4M",
                            "-pix_fmt", "yuv420p", out], capture_output=True, timeout=20)
        except subprocess.TimeoutExpired:
            sys.stderr.write(f"[stage] probe {i} timed out — a stale ffmpeg is holding "
                             "the capture device; pkill it\n")
            continue
        r = subprocess.run(["ffprobe", "-v", "error", "-select_streams", "v",
                            "-show_entries", "stream=width,height", "-of", "csv=p=0", out],
                           capture_output=True, text=True).stdout.strip()
        if not r:
            continue
        w, h = (int(v) for v in r.split(",")[:2])
        if abs(w / h - 16 / 9) < 0.01:
            print(f"[stage] capture index {i} -> {w}x{h} (16:9)")
            return i
    raise SystemExit("no 16:9 capture device found")


CAP_INDEX = _detect_capture_index()
OX, OY = 0.0, 0.0                # the shoot panel is made primary by arrange.py
DW, DH = 1920.0, 1080.0
TAKES = os.path.join(HERE, "takes")
os.makedirs(TAKES, exist_ok=True)


def P(x, y):
    """display-2 relative point -> global point"""
    return (OX + x, OY + y)


def osa(script):
    r = subprocess.run(["osascript", "-e", script], capture_output=True, text=True)
    return r.stdout.strip()


def frontmost():
    return osa('tell application "System Events" to get name of first process '
               'whose frontmost is true')


def clear_mods():
    for kc in (55, 56, 58, 59, 60, 61, 62):
        e = Q.CGEventCreateKeyboardEvent(None, kc, False)
        Q.CGEventSetFlags(e, 0)
        Q.CGEventPost(Q.kCGHIDEventTap, e)
        time.sleep(0.01)


def move(x, y, dur=0.9):
    drive.move_to(*P(x, y), dur)


def click(x, y, dur=0.9):
    drive.move_to(*P(x, y), dur)
    time.sleep(0.15)
    clear_mods()
    drive.click()


def key(name, mods=()):
    drive.key(name, mods)
    clear_mods()


def scroll(dy, dx=0, steps=10):
    drive.scroll(dy, dx, steps)


def type_text(s, cps=18):
    drive.type_text(s, cps)


def wait_front(app, timeout=25):
    t0 = time.time()
    while time.time() - t0 < timeout:
        if frontmost() == app:
            return True
        osa(f'tell application "{app}" to activate')
        time.sleep(0.8)
    return False


def close_finder_windows():
    osa('tell application "Finder" to close every window')


def snap(path, disp="1"):
    """Still of the shoot panel, for eyes-on verification before rolling."""
    subprocess.run(["screencapture", "-x", "-D", disp, path], capture_output=True)
    return path


def move_window_to_panel(app):
    """Park the app's front window on display 2, then take it fullscreen there."""
    osa(f'tell application "{app}" to activate')
    time.sleep(1.0)
    osa(f'''tell application "System Events" to tell process "{app}"
        set position of front window to {{{int(OX) + 40}, {int(OY) + 60}}}
        set size of front window to {{{int(DW) - 80}, {int(DH) - 120}}}
    end tell''')
    time.sleep(1.2)
    osa(f'tell application "{app}" to activate')
    time.sleep(0.6)
    key('f', ('cmd', 'ctrl'))
    time.sleep(3.5)
    clear_mods()


class Rec:
    def __init__(self, name, guard=()):
        self.name = name
        self.guard = guard
        self.path = os.path.join(TAKES, f"{name}.mp4")
        self.p = None
        self.stop_flag = False
        self.killed = False

    def start(self, settle=2.0):
        self.p = subprocess.Popen([
            "ffmpeg", "-y", "-loglevel", "error",
            "-f", "avfoundation", "-capture_cursor", "1",
            "-framerate", "30", "-i", CAP_INDEX,
            "-c:v", "h264_videotoolbox", "-b:v", "40M",
            "-g", "15", "-keyint_min", "15",
            "-pix_fmt", "yuv420p", self.path,
        ], stdin=subprocess.PIPE, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        time.sleep(settle)
        if self.guard:
            threading.Thread(target=self._watch, daemon=True).start()

    def _watch(self):
        while not self.stop_flag and self.p and self.p.poll() is None:
            f = frontmost()
            if f and f not in self.guard:
                self.killed = True
                sys.stderr.write(f"\n!! GUARD {self.name}: frontmost {f!r} not in {self.guard}\n")
                self.p.kill()
                return
            time.sleep(0.4)

    def stop(self):
        self.stop_flag = True
        if self.p and self.p.poll() is None:
            try:
                self.p.stdin.write(b"q"); self.p.stdin.flush()
            except Exception:
                pass
            try:
                self.p.wait(timeout=12)
            except Exception:
                self.p.kill()
        time.sleep(0.6)
        sz = os.path.getsize(self.path) if os.path.exists(self.path) else 0
        ok = sz > 200_000 and not self.killed
        print(f"[{self.name}] {'ok' if ok else 'FAILED'} {sz//1024}KB"
              f"{' guard-killed' if self.killed else ''}", flush=True)
        return ok


def desktop_quiet(on=True):
    if on:
        subprocess.run(["defaults", "write", "com.apple.finder", "CreateDesktop", "false"])
        subprocess.run(["defaults", "write", "com.apple.dock", "autohide", "-bool", "true"])
    else:
        subprocess.run(["defaults", "delete", "com.apple.finder", "CreateDesktop"],
                       capture_output=True)
        subprocess.run(["defaults", "write", "com.apple.dock", "autohide", "-bool", "false"])
    subprocess.run(["killall", "Finder"], capture_output=True)
    subprocess.run(["killall", "Dock"], capture_output=True)
    time.sleep(2.5)
    close_finder_windows()


def drag(x1, y1, x2, y2, **kw):
    """Panel-relative drag. See drive.drag for why a full path is posted."""
    drive.drag(*P(x1, y1), *P(x2, y2), **kw)


def frontmost_fast():
    """NSWorkspace instead of osascript. A long take polls this thousands of
    times and osascript costs ~40ms a call."""
    try:
        from AppKit import NSWorkspace
        app = NSWorkspace.sharedWorkspace().frontmostApplication()
        return app.localizedName() if app else ""
    except Exception:
        return frontmost()


class SegRec:
    """A recorder for LONG unattended takes (a 20-minute agent build).

    The plain Rec guard KILLS the take the moment the frontmost app is not the
    one being filmed. Over twenty minutes that is nearly certain to fire — he
    is at the machine — and it throws away the whole build.

    So this one pauses instead: it records only while an allowed app is
    frontmost, stops cleanly when it is not, and starts a new segment when it
    comes back. His other windows are still never filmed, but a glance at Slack
    costs a cut rather than the take. Segments are numbered and assembled in
    order afterwards.
    """

    def __init__(self, name, allow=("Code",), poll=0.35):
        self.name = name
        self.allow = tuple(allow)
        self.poll = poll
        self.segs = []
        self.p = None
        self._stop = False
        self._t = None
        self.pauses = 0

    def _seg_path(self):
        """Never return a path that already exists.

        Relaunching a take with the same name silently overwrote a good 197s
        segment on 2026-08-28 — a fresh SegRec starts its counter at s01 again.
        Footage is the one thing in this pipeline that cannot be regenerated, so
        the counter now skips past anything already on disk.
        """
        n = len(self.segs) + 1
        while True:
            p = os.path.join(TAKES, f"{self.name}_s{n:02d}.mp4")
            if not os.path.exists(p):
                return p
            n += 1

    def _spawn(self):
        p = self._seg_path()
        self.p = subprocess.Popen([
            "ffmpeg", "-y", "-loglevel", "error",
            "-f", "avfoundation", "-capture_cursor", "1",
            "-framerate", "30", "-i", CAP_INDEX,
            "-c:v", "h264_videotoolbox", "-b:v", "40M",
            "-g", "15", "-keyint_min", "15",
            "-pix_fmt", "yuv420p", p,
        ], stdin=subprocess.PIPE, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        self.segs.append(p)

    def _kill(self):
        if self.p and self.p.poll() is None:
            try:
                self.p.stdin.write(b"q"); self.p.stdin.flush()
                self.p.wait(timeout=10)
            except Exception:
                self.p.kill()
        self.p = None

    def _watch(self):
        while not self._stop:
            ok = frontmost_fast() in self.allow
            if ok and self.p is None:
                self._spawn()
            elif not ok and self.p is not None:
                self._kill()
                self.pauses += 1
                sys.stderr.write(f"[{self.name}] paused (frontmost "
                                 f"{frontmost_fast()!r}) after {len(self.segs)} seg\n")
            time.sleep(self.poll)
        self._kill()

    def start(self, settle=2.0):
        self._t = threading.Thread(target=self._watch, daemon=True)
        self._t.start()
        time.sleep(settle)

    def stop(self):
        self._stop = True
        if self._t:
            self._t.join(timeout=20)
        kept = []
        for p in self.segs:
            sz = os.path.getsize(p) if os.path.exists(p) else 0
            if sz > 300_000:
                kept.append(p)
            elif os.path.exists(p):
                os.remove(p)
        total = 0.0
        for p in kept:
            r = subprocess.run(["ffprobe", "-v", "error", "-show_entries",
                                "format=duration", "-of", "csv=p=0", p],
                               capture_output=True, text=True).stdout.strip()
            try:
                total += float(r)
            except ValueError:
                pass
        print(f"[{self.name}] {len(kept)} segment(s), {total:.1f}s total, "
              f"{self.pauses} pause(s)", flush=True)
        for p in kept:
            print("   ", p)
        return kept


def front_window_title(app):
    return osa('tell application "System Events" to tell process "%s" '
               'to get name of front window' % app)


def wait_front_window(app, must_contain, timeout=20):
    """Raise `app` AND verify its FRONT WINDOW is the one we mean.

    `wait_front` only checks which APPLICATION is frontmost, which is not enough
    when he has his own window open in the same app. On 2026-08-29 a take typed
    `git clone ...` into his courseSmith window because VS Code was frontmost but
    the wrong window was; nothing was damaged only because the active tab held a
    binary file that ignored the keystrokes. A text editor would have taken them.

    Raises rather than typing into the wrong window.
    """
    wait_front(app, timeout=timeout)
    raise_script = (
        'tell application "System Events" to tell process "%s"\n'
        '  repeat with w in windows\n'
        '    if name of w contains "%s" then\n'
        '      perform action "AXRaise" of w\n'
        '      exit repeat\n'
        '    end if\n'
        '  end repeat\n'
        'end tell' % (app, must_contain))
    for _ in range(int(timeout / 0.5)):
        t = front_window_title(app) or ""
        if must_contain.lower() in t.lower():
            return t
        osa(raise_script)
        time.sleep(0.5)
    raise SystemExit(
        "REFUSING to drive %s: front window is %r, which does not contain %r. "
        "Nothing was typed." % (app, front_window_title(app), must_contain))
