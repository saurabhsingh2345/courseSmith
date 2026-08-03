package pipeline

// Tool captures: recording a real CLI session with a real credential.
//
// == Why this cannot reuse the Python path ==
//
// `DockerTapeRunner` runs the sandbox with `--network none`. That is exactly
// right for executing a stranger's Python and fatal for every tool worth
// recording: `claude`, `vercel`, `gh` and `supabase` are network clients
// holding credentials, and with the network off they do not fail interestingly,
// they fail at the login check in the first second.
//
// So a tool capture runs on the host, with the network up and the operator's
// real credentials. There is no version of this that is as isolated as the
// Python path, and pretending otherwise would be worse than saying it: an
// LLM-written tape body is about to drive authenticated CLIs on a real machine.
//
// Three things stand there instead, and they are aimed at the actual threat.
// The model writing these tapes is our own content model working from a
// description we wrote — the risk is an *accident* (a stray `rm`, a real
// `git push`, money spent) rather than an attack, and accidents are what a
// denylist is genuinely good at:
//
//  1. **A scratch working directory.** The tape runs in a throwaway dir, never
//     in the course tree. This is a containment control and also a correctness
//     one — an agent session recorded inside `generated/` would see and edit the
//     course's own files, which is not the demo anybody wanted.
//  2. **An engine-owned allowlist.** The tape must run the tool the lesson
//     declared, and the set of declarable tools is in this file, not in a
//     prompt.
//  3. **A denylist of irreversible verbs.** Destructive, exfiltrating, and
//     publishing commands are refused outright — `rm`, `sudo`, `curl`,
//     `git push`, `npm publish`. A recording is never worth any of them.
//
// What this does not defend against is a deliberately hostile model, and the
// containment is a scratch dir rather than a sandbox. That limit is real and is
// written down in docs/no-code-track.md rather than papered over.

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// captureTool is one recordable CLI. The set is closed and lives here rather
// than in a prompt, because "which binaries may a generated script run on this
// machine" is not a content decision.
type captureTool struct {
	// Key is what a lesson writes in `tool=`.
	Key string
	// Binary is what must appear in the tape and be present on PATH.
	Binary string
	// Display is how the tool is named to a reader.
	Display string
	// VersionArgs asks the binary for its own version. Observed at capture
	// time so footage.json records what really ran, not what we assumed.
	VersionArgs []string
	// Invocation tells the tape writer how to drive this tool inside a
	// recording, and it is engine-owned for the same reason the allowlist is:
	// "how do I run this binary without it stopping to ask a question" is not
	// something a content model should be guessing per take.
	//
	// It exists because of a real failure. Claude Code's interactive UI asks
	// whether you trust the folder it was opened in, and a capture always runs
	// in a fresh scratch directory, so the recording stalled at a prompt no
	// keystroke in the tape was going to answer. Empty means the obvious
	// invocation is fine.
	Invocation string
	// Invoke is the shape a working invocation must take, checked against the
	// tape rather than requested in the prose.
	//
	// It is here because the prose was not enough, twice. Told "never type a
	// bare `claude`", the model wrote `claude 'add a summary'` — not bare, still
	// interactive, still stalls on the trust prompt. A rule the model can
	// satisfy in letter while breaking in spirit is a rule that belongs in the
	// validator; the recording is the only place the mistake shows up, and by
	// then a real API call has been spent.
	//
	// Nil means any invocation naming the binary is fine.
	Invoke *regexp.Regexp
	// InvokeWhy explains the requirement to the model that just missed it.
	InvokeWhy string
}

// captureTools is the allowlist. Adding one is a deliberate act: it grants a
// generated tape permission to run that binary, authenticated, on the host.
var captureTools = map[string]captureTool{
	"claude": {Key: "claude", Binary: "claude", Display: "Claude Code", VersionArgs: []string{"--version"},
		// Both halves are required, and the second one is here because the
		// first was not enough. `claude -p "<request>"` satisfies every rule
		// this validator had and still records nothing: in print mode the Write
		// tool needs an approval nobody is there to give, so the agent prints a
		// paragraph explaining that it *would* have written the file and exits
		// 0, leaving the directory empty. The recording that produced was four
		// minutes of prose about a page that was never built — evidence of
		// nothing, on the one surface whose whole claim is that what you see
		// really happened.
		Invoke:    regexp.MustCompile(`\bclaude\s+(-p|--print)\b.*--allowed-?tools\b`),
		InvokeWhy: "claude must be run with -p AND with the tools it needs granted, as `claude -p \"<the request>\" --allowedTools Write,Edit,Bash`. Without -p — including `claude \"<the request>\"` with the request as a plain argument — it opens its interactive interface, which asks whether you trust the folder before doing anything, and a capture always starts in a fresh directory. Without --allowedTools it cannot write a single file: it prints a paragraph saying the write needs approval, exits, and the recording shows an agent describing work it never did.",
		Invocation: `Run it as ` + "`claude -p \"<the request>\" --allowedTools Write,Edit,Bash`" + `, with BOTH parts, always.

The ` + "`--allowedTools`" + ` flag goes AFTER the request, never before it. It takes a list, so ` + "`claude -p --allowedTools Write \"<the request>\"`" + ` swallows the request as another tool name and the command fails outright.

Without it the agent cannot write anything — it prints a paragraph explaining that the write needs approval, and the recording shows nothing happening. The scratch directory is a throwaway made fresh for this recording, which is why granting the tools there is safe and is the only way the shot exists.

` + "`claude \"<the request>\"`" + ` — the request as a plain argument, no -p — is WRONG and is the mistake to avoid. So is a bare ` + "`claude`" + `. Both open the interactive interface and the recording will sit on the trust prompt until the take times out.

It takes real time, so put the one ` + "`Wait`" + ` immediately after it. Afterwards, show what it produced — ` + "`cat`" + ` the file it wrote, or ` + "`ls`" + ` the directory. A viewer believes the file, not the summary.`},
	"vercel": {Key: "vercel", Binary: "vercel", Display: "Vercel CLI", VersionArgs: []string{"--version"},
		Invocation: `Pass ` + "`--yes`" + ` so it never stops to ask a setup question a recording cannot answer.
Deploying takes real time, so put the one ` + "`Wait`" + ` immediately after the deploy command.`},
	"supabase": {Key: "supabase", Binary: "supabase", Display: "Supabase CLI", VersionArgs: []string{"--version"},
		Invocation: `Prefer read-only commands. Anything that would create or destroy a real hosted project is out of scope for a recording.`},
	"gh": {Key: "gh", Binary: "gh", Display: "GitHub CLI", VersionArgs: []string{"--version"},
		Invocation: `Prefer read-only subcommands (` + "`gh repo view`, `gh pr list`, `gh run list`" + `). They need no confirmation and change nothing.`},
	"git": {Key: "git", Binary: "git", Display: "git", VersionArgs: []string{"--version"}},
	"npm": {Key: "npm", Binary: "npm", Display: "npm", VersionArgs: []string{"--version"}},
}

// captureToolKeys lists the allowlist in a stable order, for error messages.
func captureToolKeys() []string {
	keys := make([]string, 0, len(captureTools))
	for k := range captureTools {
		keys = append(keys, k)
	}
	// Small and fixed; a simple insertion keeps the message deterministic.
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

// forbiddenCommand is a verb no recording justifies, with the reason attached
// so the regenerating model is told why rather than just refused.
type forbiddenCommand struct {
	Pattern *regexp.Regexp
	Why     string
}

var forbiddenCommands = []forbiddenCommand{
	{regexp.MustCompile(`\brm\b`), "deleting files is never part of a demo — start from a fixture instead"},
	{regexp.MustCompile(`\b(sudo|chmod|chown|dd|mkfs)\b`), "a recording never needs elevated or destructive system commands"},
	{regexp.MustCompile(`\b(curl|wget|scp|ssh|nc)\b`), "the demo may not reach the network except through the tool itself"},
	{regexp.MustCompile(`\bgit\s+push\b`), "pushing is irreversible and visible to other people; record the commit, not the push"},
	{regexp.MustCompile(`\bnpm\s+publish\b`), "publishing is irreversible; record the build, not the release"},
	{regexp.MustCompile(`\b(vercel|supabase)\s+(remove|rm|delete|destroy)\b`), "deleting a real deployment or project is irreversible"},
	{regexp.MustCompile(`>\s*/`), "writing outside the scratch directory is not allowed"},
	{regexp.MustCompile(`\$\(|` + "`"), "command substitution hides what is really being run"},
}

// captureMarkerRe finds `[CAPTURE: tool=claude; description]` markers, with an
// optional `fixture=` attribute:
//
//	[CAPTURE: tool=claude, fixture=habit-tracker; ask the agent to add streaks]
//
// The attributes sit before a `;` so the description stays free prose — a
// lesson author should not have to escape anything to describe a demo.
var captureMarkerRe = regexp.MustCompile(`\[CAPTURE:\s*([^;\]]+);\s*([^\]]+)\]`)

// parseCaptureAttrs reads the `key=value, key=value` half of a CAPTURE marker.
func parseCaptureAttrs(s string) (map[string]string, error) {
	attrs := map[string]string{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("%q is not key=value", part)
		}
		attrs[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return attrs, nil
}

// extractCaptureMarkers finds every [CAPTURE: ...] marker in outline order and
// assigns stable ids capture-1, capture-2, ...
//
// Numbering is independent of the [DEMO:] sequence on purpose: sharing a
// counter would mean adding a capture renumbers every demo after it, which
// re-stales clips that did not change and re-shoots them. Ids are what make a
// clip's cache entry stable, so they must not move when a neighbour appears.
func extractCaptureMarkers(body string) ([]DemoSpec, error) {
	matches := captureMarkerRe.FindAllStringSubmatch(body, -1)
	specs := make([]DemoSpec, 0, len(matches))
	for i, m := range matches {
		attrs, err := parseCaptureAttrs(m[1])
		if err != nil {
			return nil, fmt.Errorf("capture-%d: %w", i+1, err)
		}
		toolKey := attrs["tool"]
		if toolKey == "" {
			return nil, fmt.Errorf("capture-%d names no tool — write [CAPTURE: tool=claude; what to record]", i+1)
		}
		spec := DemoSpec{
			ID:          fmt.Sprintf("capture-%d", i+1),
			Description: strings.TrimSpace(m[2]),
			Tool:        toolKey,
			Fixture:     attrs["fixture"],
			Take:        attrs["take"],
		}
		switch {
		case captureTools[toolKey].Binary != "":
			spec.Kind = CaptureKindTool
			if spec.Take != "" {
				return nil, fmt.Errorf("capture-%d: %s is a terminal tool, so it has no take file — the tape is generated. Drop `take=`", i+1, toolKey)
			}
		case captureApps[toolKey].Bundle != "":
			spec.Kind = CaptureKindDesktop
			if spec.Take == "" {
				return nil, fmt.Errorf("capture-%d: a %s capture needs `take=<name>`, naming a file in the course's takes/ directory. A native app has no selectors to drive, so the take is a list of beats an operator works through",
					i+1, captureApps[toolKey].Display)
			}
			if spec.Fixture != "" {
				return nil, fmt.Errorf("capture-%d: a desktop capture has no working directory, so `fixture=` means nothing here", i+1)
			}
		case captureSites[toolKey].Origin != "":
			spec.Kind = CaptureKindWeb
			if spec.Take == "" {
				return nil, fmt.Errorf("capture-%d: a %s capture needs `take=<name>`, naming a file in the course's takes/ directory. Nobody can invent selectors for somebody else's page, so a web take is written once and kept",
					i+1, captureSites[toolKey].Display)
			}
			if spec.Fixture != "" {
				return nil, fmt.Errorf("capture-%d: a web capture has no working directory, so `fixture=` means nothing here", i+1)
			}
		default:
			return nil, fmt.Errorf("capture-%d: %q is not recordable. Terminal tools: %s. Web products: %s. Desktop apps: %s",
				i+1, toolKey, strings.Join(captureToolKeys(), ", "), strings.Join(captureSiteKeys(), ", "), strings.Join(captureAppKeys(), ", "))
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

// tapeDirectiveRe matches any line that is a VHS directive rather than
// something the model invented.
//
// The tool lint needs this and the python lint never did, because of how the
// two fail. A python tape that goes wrong still looks like a tape; a tool tape
// goes wrong by the model writing the *shell session it is imagining* —
//
//	claude -p "add a weekly summary"
//	Wait
//
// — which is a plausible-looking file, is not a tape, and slips past a lint
// that only inspects `Type` lines. The failure then surfaces as "the tape never
// runs claude", which is true and tells the model nothing about what it did
// wrong. Naming the actual mistake is the difference between a correction round
// that converges and one that repeats itself.
var tapeDirectiveRe = regexp.MustCompile(`^(Type|Enter|Sleep|Wait|Backspace|Delete|Tab|Space|Up|Down|Left|Right|PageUp|PageDown|Escape|Hide|Show|Screenshot|Copy|Paste|Env|Ctrl\+|Alt\+|Shift\+)`)

// lintToolTapeBody enforces the real-execution contract on a tool tape.
//
// It shares two rules with the Python lint and they matter more here. `echo` is
// still banned, and typing a program's output is still banned — a model that
// types Vercel's success banner has forged a deploy, which is the exact failure
// this whole track exists to make impossible.
func lintToolTapeBody(content string, tool captureTool) (string, error) {
	body := stripFences(content)
	sawBinary := false
	sawMark := false
	for i, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			if markCommentRe.MatchString(trimmed) {
				sawMark = true
				continue
			}
			if strings.HasPrefix(strings.ToUpper(trimmed), "# MARK") {
				return "", fmt.Errorf("line %d: a mark name must be lowercase letters, digits and dashes — %q is not", i+1, trimmed)
			}
			continue
		}
		first, _, _ := strings.Cut(trimmed, " ")
		switch first {
		case "Output", "Set", "Source", "Require":
			return "", fmt.Errorf("line %d: %s directives are engine-owned — write only the tape body", i+1, first)
		}
		if !tapeDirectiveRe.MatchString(trimmed) {
			return "", fmt.Errorf("line %d: %q is not a VHS directive. A tape does not contain shell commands — every command is typed, so `%s` has to be written as `Type %q` followed by `Enter`",
				i+1, truncateForLog(trimmed, 40), tool.Binary, trimmed)
		}
		if !strings.HasPrefix(trimmed, "Type") {
			continue
		}
		typed := strings.ToLower(trimmed)
		if strings.Contains(typed, "echo ") || strings.Contains(typed, "echo\"") || strings.Contains(typed, "echo'") {
			return "", fmt.Errorf("line %d: echo is forbidden — every character on screen must come from really running %s", i+1, tool.Binary)
		}
		for _, f := range forbiddenCommands {
			if f.Pattern.MatchString(typed) {
				return "", fmt.Errorf("line %d: %s", i+1, f.Why)
			}
		}
		// Word-bounded, because a substring test is wrong in both directions
		// that matter: `gh` is inside "high" and `git` is inside "digit", so a
		// tape that merely mentions the tool in prose would pass the one check
		// standing between us and a capture that never runs the tool.
		if tool.Invoke != nil {
			if tool.Invoke.MatchString(typed) {
				sawBinary = true
			}
		} else if binaryMentionRe(tool.Binary).MatchString(typed) {
			sawBinary = true
		}
	}
	if strings.TrimSpace(body) == "" {
		return "", fmt.Errorf("tape body is empty")
	}
	if !sawBinary {
		if tool.InvokeWhy != "" {
			return "", fmt.Errorf("the tape never runs %s in a form that will work: %s", tool.Binary, tool.InvokeWhy)
		}
		return "", fmt.Errorf("the tape never runs %s — a capture of %s must actually run it", tool.Binary, tool.Display)
	}
	if !sawMark {
		return "", fmt.Errorf("the tape carries no `# MARK <name>` comments — without at least one, nothing downstream can reference a moment in this recording")
	}
	return strings.TrimSpace(body) + "\n", nil
}

// resolveToolTapeRunner picks the recorder for a tool capture.
//
// It is deliberately *only* HostTapeRunner. The docker runner is not offered
// and not falling back to — it runs `--network none`, so a tool capture inside
// it would not fail loudly, it would record an authentic-looking clip of a CLI
// failing to reach its API, which is the worst possible outcome for a stage
// whose entire purpose is that the recording is real.
func resolveToolTapeRunner(ctx context.Context, tool captureTool) (TapeRunner, error) {
	if _, err := exec.LookPath("vhs"); err != nil {
		return nil, fmt.Errorf("recording %s needs vhs on the host (the docker sandbox runs with the network off, which a %s capture cannot use) — install it: brew install vhs",
			tool.Display, tool.Display)
	}
	if _, err := exec.LookPath(tool.Binary); err != nil {
		return nil, fmt.Errorf("%s is not on PATH, so there is nothing to record", tool.Binary)
	}
	probe, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := execCommandOutput(probe, tool.Binary, tool.VersionArgs...); err != nil {
		return nil, fmt.Errorf("%s is installed but would not run (%v) — a capture cannot record a tool that is not working", tool.Binary, err)
	}
	return HostTapeRunner{}, nil
}

// CaptureReadiness reports whether tool captures can run at all on this
// machine, and which allowlisted tools are missing from PATH.
//
// The two answers are separate because they fail differently: no host vhs means
// no capture of anything, while a missing binary only blocks the captures that
// name it — a course recording Claude Code does not care that `supabase` is
// absent.
func CaptureReadiness(ctx context.Context) (bool, []string) {
	if _, err := exec.LookPath("vhs"); err != nil {
		return false, nil
	}
	var missing []string
	for _, key := range captureToolKeys() {
		if _, err := exec.LookPath(captureTools[key].Binary); err != nil {
			missing = append(missing, key)
		}
	}
	return true, missing
}

// execCommandOutput runs a command and returns its combined output.
func execCommandOutput(ctx context.Context, name string, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	return string(out), err
}

// binaryMentionCache memoises the per-binary word-boundary matchers.
var binaryMentionCache sync.Map

// binaryMentionRe matches the binary as a whole word.
func binaryMentionRe(binary string) *regexp.Regexp {
	if re, ok := binaryMentionCache.Load(binary); ok {
		return re.(*regexp.Regexp)
	}
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(strings.ToLower(binary)) + `\b`)
	binaryMentionCache.Store(binary, re)
	return re
}

// observeToolVersion asks the binary what it is. A failure is not fatal: a
// missing version makes the clip's freshness harder to judge, which is worth a
// blank field and not worth failing a capture that otherwise recorded fine.
func observeToolVersion(ctx context.Context, tool captureTool) string {
	probe, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := execCommandOutput(probe, tool.Binary, tool.VersionArgs...)
	if err != nil {
		return ""
	}
	line, _, _ := strings.Cut(strings.TrimSpace(out), "\n")
	return strings.TrimSpace(line)
}

// captureTapeRelPath is where a tool tape sits relative to the directory the
// recording runs in: one level up, deliberately.
//
// VHS gives the recorded shell the same working directory the tape is invoked
// from, so a tape written *into* that directory appears in the recording — a
// `git status` demo listing `capture-1.tape` as an untracked file, on screen,
// in the finished video. The scratch root holds the tape and its `work/` child
// is what gets recorded, so the shell only ever sees the fixture.
const captureTapeRelPath = ".."

// prepareCaptureWorkdir makes the throwaway directory a tool tape records in,
// and seeds it from the course's fixtures when the marker named one. It returns
// the directory the recording runs in and the cleanup for the whole scratch
// tree, which is one level above it (see captureTapeRelPath).
func prepareCaptureWorkdir(courseDir, fixture string) (string, func(), error) {
	root, err := os.MkdirTemp("", "coursesmith-capture-*")
	if err != nil {
		return "", func() {}, fmt.Errorf("creating capture scratch dir: %w", err)
	}
	cleanup := func() { os.RemoveAll(root) }
	work := filepath.Join(root, "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("creating capture scratch dir: %w", err)
	}
	if fixture == "" {
		return work, cleanup, nil
	}
	src := filepath.Join(courseDir, "fixtures", fixture)
	info, err := os.Stat(src)
	if err != nil || !info.IsDir() {
		cleanup()
		return "", func() {}, fmt.Errorf("fixture %q not found at %s — a capture that names a fixture needs it to exist, or the recording starts in an empty directory and shows nothing", fixture, src)
	}
	if err := copyTree(src, work); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("seeding capture scratch dir from %s: %w", src, err)
	}
	return work, cleanup, nil
}

// copyTree copies a fixture directory's contents into dst.
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}
