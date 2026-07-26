# Execution tracer for coursesmith's code-visualisation stage (workstream C,
# the "Python Tutor" moment). Runs one lesson code block under sys.settrace and
# emits a JSON trace of every step — line, call stack, local variables (with
# structured container/object values), and the cumulative stdout at that point.
#
# Contract: the Go stage prepends a definition of USER_CODE (the block source,
# base64-decoded) before this body, then runs the whole thing via the code
# runner. This body expects USER_CODE to already be a str. The trace is written
# as a single JSON object to the *real* stdout; the traced program's own prints
# are captured separately so they never corrupt the JSON.
#
# Kept deliberately 3.9-compatible (sandbox is 3.12, host here is 3.9).

import base64 as _b64  # noqa: F401  (may be used by the injected header)
import io
import json
import sys
import traceback

# Standalone fallback so the file can be exercised directly in tests:
# `python3 tracer.py <<< 'CODE'` traces stdin when no USER_CODE was injected.
try:
    USER_CODE  # type: ignore[name-defined]
except NameError:
    USER_CODE = sys.stdin.read()

MAX_STEPS = 400      # hard cap so a runaway loop can't produce an unbounded trace
MAX_REPR = 140       # per-value repr truncation
MAX_ITEMS = 24       # elements shown per container
FILENAME = "<lesson>"  # user frames carry this; everything else is library code

_steps = []
_out = io.StringIO()
_truncated = [False]


def _short_repr(v):
    try:
        r = repr(v)
    except Exception:
        return "<unrepr>"
    if len(r) > MAX_REPR:
        r = r[:MAX_REPR] + "…"
    return r


def _render(v, depth=0):
    """Render a value into a JSON-safe {type, repr, [items|entries|fields]}."""
    t = type(v).__name__
    try:
        if isinstance(v, (bool, int, float, type(None), str)):
            return {"type": t, "repr": _short_repr(v)}
        if isinstance(v, (list, tuple, set, frozenset)):
            items = []
            for i, e in enumerate(v):
                if i >= MAX_ITEMS:
                    break
                items.append(_child(e, depth))
            return {"type": t, "repr": _short_repr(v), "items": items}
        if isinstance(v, dict):
            entries = []
            for i, (k, val) in enumerate(v.items()):
                if i >= MAX_ITEMS:
                    break
                entries.append({"key": _short_repr(k), "value": _child(val, depth)})
            return {"type": t, "repr": _short_repr(v), "entries": entries}
        if hasattr(v, "__dict__") and not isinstance(v, type):
            fields = []
            for i, (k, val) in enumerate(vars(v).items()):
                if i >= MAX_ITEMS:
                    break
                if k.startswith("_"):
                    continue
                fields.append({"key": k, "value": _child(val, depth)})
            return {"type": t, "repr": _short_repr(v), "fields": fields}
    except Exception:
        pass
    return {"type": t, "repr": _short_repr(v)}


def _child(v, depth):
    # One level of nesting keeps traces readable; deeper values collapse to repr.
    if depth >= 1:
        return {"type": type(v).__name__, "repr": _short_repr(v)}
    return _render(v, depth + 1)


def _skip_var(name, val):
    if name.startswith("__"):
        return True
    # Hide imported modules; keep user functions/classes (they are real state).
    return type(val).__name__ == "module"


def _record(frame, event):
    if len(_steps) >= MAX_STEPS:
        _truncated[0] = True
        return
    variables = []
    for name, val in frame.f_locals.items():
        if _skip_var(name, val):
            continue
        variables.append({"name": name, "value": _render(val)})

    stack = []
    f = frame
    while f is not None:
        if f.f_code.co_filename == FILENAME:
            stack.append(f.f_code.co_name)
        f = f.f_back
    stack.reverse()

    _steps.append(
        {
            "step": len(_steps),
            "line": frame.f_lineno,
            "event": event,
            "func": frame.f_code.co_name,
            "vars": variables,
            "stack": stack,
            "stdout": _out.getvalue(),
        }
    )


def _local_trace(frame, event, arg):
    if event in ("line", "return", "exception"):
        _record(frame, event)
    return _local_trace


def _global_trace(frame, event, arg):
    # Only descend into user-code frames; library internals are ignored.
    if frame.f_code.co_filename != FILENAME:
        return None
    return _local_trace


def main():
    error = None
    namespace = {"__name__": "__main__"}
    real_stdout = sys.stdout
    try:
        compiled = compile(USER_CODE, FILENAME, "exec")
    except SyntaxError as e:
        json.dump(
            {
                "code": USER_CODE,
                "lines": USER_CODE.split("\n"),
                "steps": [],
                "truncated": False,
                "error": "SyntaxError: {} (line {})".format(e.msg, e.lineno),
                "stdout": "",
            },
            real_stdout,
        )
        return

    sys.stdout = _out
    sys.settrace(_global_trace)
    try:
        exec(compiled, namespace)
    except Exception as e:  # noqa: BLE001 — surface any runtime error to the trace
        error = "".join(traceback.format_exception_only(type(e), e)).strip()
    finally:
        sys.settrace(None)
        sys.stdout = real_stdout

    json.dump(
        {
            "code": USER_CODE,
            "lines": USER_CODE.split("\n"),
            "steps": _steps,
            "truncated": _truncated[0],
            "error": error,
            "stdout": _out.getvalue(),
        },
        real_stdout,
    )


main()
