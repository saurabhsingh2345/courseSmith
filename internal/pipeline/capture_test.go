package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractCaptureMarkers(t *testing.T) {
	body := "## Ship it\n\n[CAPTURE: tool=claude, fixture=habit-tracker; ask the agent to add streaks]\n\n" +
		"text\n\n[CAPTURE: tool=vercel; deploy to production]\n"
	specs, err := extractCaptureMarkers(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 2 {
		t.Fatalf("specs = %+v", specs)
	}
	if specs[0].ID != "capture-1" || specs[0].Tool != "claude" || specs[0].Fixture != "habit-tracker" {
		t.Errorf("spec 0 = %+v", specs[0])
	}
	if specs[0].Description != "ask the agent to add streaks" {
		t.Errorf("description = %q", specs[0].Description)
	}
	if specs[0].Kind != CaptureKindTool {
		t.Errorf("kind = %q", specs[0].Kind)
	}
	if specs[1].ID != "capture-2" || specs[1].Tool != "vercel" || specs[1].Fixture != "" {
		t.Errorf("spec 1 = %+v", specs[1])
	}
}

// Capture ids must not share a counter with demo ids. If they did, adding a
// capture would renumber every demo after it, re-staling clips that did not
// change and re-shooting them — and a re-shoot of a tool capture costs a real
// API call against a real credential.
func TestCaptureAndDemoIDsAreNumberedIndependently(t *testing.T) {
	body := "[DEMO: one]\n[CAPTURE: tool=gh; two]\n[DEMO: three]\n"
	demos := extractDemoMarkers(body)
	captures, err := extractCaptureMarkers(body)
	if err != nil {
		t.Fatal(err)
	}
	if demos[0].ID != "demo-1" || demos[1].ID != "demo-2" {
		t.Errorf("demo ids = %s, %s", demos[0].ID, demos[1].ID)
	}
	if captures[0].ID != "capture-1" {
		t.Errorf("capture id = %s", captures[0].ID)
	}
	if demos[0].Kind != CaptureKindPython {
		t.Errorf("a [DEMO] marker must stay the python kind, got %q", demos[0].Kind)
	}
}

func TestCaptureMarkerRejectsUnknownTool(t *testing.T) {
	_, err := extractCaptureMarkers("[CAPTURE: tool=photoshop; do a thing]\n")
	if err == nil {
		t.Fatal("an off-allowlist tool was accepted")
	}
	if !strings.Contains(err.Error(), "claude") {
		t.Errorf("the error should list what is available: %v", err)
	}
}

func TestCaptureMarkerNeedsATool(t *testing.T) {
	if _, err := extractCaptureMarkers("[CAPTURE: fixture=x; do a thing]\n"); err == nil {
		t.Fatal("a capture with no tool was accepted")
	}
}

func TestCaptureToolKeysAreDeterministic(t *testing.T) {
	first := strings.Join(captureToolKeys(), ",")
	for range 5 {
		if got := strings.Join(captureToolKeys(), ","); got != first {
			t.Fatalf("order changed: %s vs %s", got, first)
		}
	}
	if first != "claude,gh,git,npm,supabase,vercel" {
		t.Errorf("keys = %s", first)
	}
}

const goodToolTape = `Type "claude -p 'add streaks' --allowedTools Write,Edit"
Enter
# MARK sent
Wait
# MARK done
Sleep 3s
`

// A `claude -p` with no tools granted, which is the shape that passed every
// check and recorded nothing: the agent cannot write, so it prints a paragraph
// about the approval it did not get and exits. Used by the cases below that are
// testing something else, so they keep failing for their own reason.
const ungrantedClaude = `claude -p go --allowedTools Write`

func TestLintToolTapeBody(t *testing.T) {
	claude := captureTools["claude"]
	for _, tt := range []struct {
		name    string
		body    string
		wantErr string
	}{
		{"accepts a real session", goodToolTape, ""},
		{
			"rejects a tape that never runs the tool",
			"Type \"ls\"\nEnter\n# MARK listed\n",
			"never runs claude",
		},
		{
			"rejects echo",
			"Type \"" + ungrantedClaude + "\"\nEnter\nType \"echo done\"\nEnter\n# MARK m\n",
			"echo is forbidden",
		},
		{
			"rejects a tape with no marks",
			"Type \"" + ungrantedClaude + "\"\nEnter\nSleep 2s\n",
			"no `# MARK",
		},
		{
			"rejects engine-owned directives",
			"Set FontSize 12\nType \"" + ungrantedClaude + "\"\nEnter\n# MARK m\n",
			"engine-owned",
		},
		{
			"rejects rm",
			"Type \"" + ungrantedClaude + "\"\nEnter\nType \"rm -rf build\"\nEnter\n# MARK m\n",
			"deleting files",
		},
		{
			"rejects git push",
			"Type \"" + ungrantedClaude + "\"\nEnter\nType \"git push origin main\"\nEnter\n# MARK m\n",
			"irreversible",
		},
		{
			"rejects curl",
			"Type \"" + ungrantedClaude + "\"\nEnter\nType \"curl https://example.com\"\nEnter\n# MARK m\n",
			"except through the tool itself",
		},
		{
			"rejects command substitution",
			"Type \"claude -p $(cat secret)\"\nEnter\n# MARK m\n",
			"command substitution",
		},
		{
			"rejects an upper-case mark name",
			"Type \"" + ungrantedClaude + "\"\nEnter\n# MARK Deploy-Green\n",
			"lowercase",
		},
		{
			// The failure a real run produced: the model writes the shell
			// session it is imagining rather than a tape. Caught by name, so
			// the correction round is told what it did instead of being told
			// the tool was never run.
			"rejects a raw shell command",
			"ls\nSleep 1s\nclaude -p \"add a summary\"\nWait\n# MARK done\n",
			"is not a VHS directive",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lintToolTapeBody(tt.body, claude)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(got, "claude") {
					t.Errorf("body came back mangled:\n%s", got)
				}
				return
			}
			if err == nil {
				t.Fatalf("accepted, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

// The prose told the model "never a bare claude"; it wrote
// `claude 'add a summary'`, which is not bare, is still interactive, and still
// stalls the take on the trust prompt. A rule satisfiable in letter while
// broken in spirit belongs in the validator.
func TestClaudeMustBeInvokedWithPrintFlag(t *testing.T) {
	claude := captureTools["claude"]
	_, err := lintToolTapeBody("Type \"claude 'add a weekly summary'\"\nEnter\n# MARK done\n", claude)
	if err == nil {
		t.Fatal("an interactive claude invocation was accepted")
	}
	if !strings.Contains(err.Error(), "-p") {
		t.Errorf("the error should name the flag that was missing: %v", err)
	}
	if _, err := lintToolTapeBody("Type \"claude -p 'add a weekly summary' --allowedTools Write,Edit\"\nEnter\n# MARK done\n", claude); err != nil {
		t.Fatalf("a correct -p invocation was rejected: %v", err)
	}
	if _, err := lintToolTapeBody("Type \"claude --print 'go' --allowed-tools Write\"\nEnter\n# MARK done\n", claude); err != nil {
		t.Fatalf("--print with the hyphenated flag spelling was rejected: %v", err)
	}
}

// The second half of the same lesson, learned the same way.
//
// `claude -p "<request>"` satisfied every rule this validator had. It also
// recorded nothing: in print mode the Write tool needs an approval no one is
// there to give, so the agent prints a paragraph about the permission it did
// not get, exits 0, and leaves the directory empty. The clip that produced was
// four minutes of an agent describing a page it never built — on the one
// surface whose entire claim is that what you see really happened.
func TestClaudeMustBeInvokedWithItsToolsGranted(t *testing.T) {
	claude := captureTools["claude"]
	_, err := lintToolTapeBody("Type \"claude -p 'build me a landing page'\"\nEnter\n# MARK done\n", claude)
	if err == nil {
		t.Fatal("a claude invocation that cannot write a single file was accepted")
	}
	if !strings.Contains(err.Error(), "--allowedTools") {
		t.Errorf("the error should name the flag that was missing: %v", err)
	}
	// Both spellings the CLI accepts.
	for _, flag := range []string{"--allowedTools Write", "--allowed-tools Write,Edit,Bash"} {
		body := "Type \"claude -p 'go' " + flag + "\"\nEnter\n# MARK done\n"
		if _, err := lintToolTapeBody(body, claude); err != nil {
			t.Errorf("%s was rejected: %v", flag, err)
		}
	}
}

// The binary check is word-bounded because a substring test fails in exactly
// the direction that matters: `gh` is inside "high", so a tape that only talks
// about the tool would pass the one check standing between us and a capture
// that never ran it.
func TestToolMentionIsWordBounded(t *testing.T) {
	gh := captureTools["gh"]
	_, err := lintToolTapeBody("Type \"echoes from a high place\"\nEnter\n# MARK m\n", gh)
	if err == nil {
		t.Fatal("\"high\" was accepted as running gh")
	}
	if _, err := lintToolTapeBody("Type \"gh pr list\"\nEnter\n# MARK m\n", gh); err != nil {
		t.Fatalf("a real gh invocation was rejected: %v", err)
	}
}

func TestPrepareCaptureWorkdirSeedsFromFixture(t *testing.T) {
	courseDir := t.TempDir()
	fixture := filepath.Join(courseDir, "fixtures", "habit-tracker")
	if err := os.MkdirAll(filepath.Join(fixture, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture, "src", "app.js"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	dir, cleanup, err := prepareCaptureWorkdir(courseDir, "habit-tracker")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	got, err := os.ReadFile(filepath.Join(dir, "src", "app.js"))
	if err != nil || string(got) != "hi\n" {
		t.Errorf("fixture not seeded: %v / %q", err, got)
	}
	// The scratch dir is not the course dir — that is the containment.
	if dir == courseDir || strings.HasPrefix(dir, courseDir) {
		t.Errorf("scratch dir %s sits inside the course tree", dir)
	}
	// The tape's home is above the recorded directory, so it cannot appear in
	// its own recording. Nothing but the fixture is visible to the shell.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "src" {
		t.Errorf("the recorded directory holds more than the fixture: %+v", entries)
	}
	if _, err := os.Stat(filepath.Join(dir, captureTapeRelPath)); err != nil {
		t.Errorf("no scratch root above the recorded directory: %v", err)
	}
}

// A capture that names a fixture and does not get one records an empty
// directory, which is a clip of nothing. Better to fail than to ship it.
func TestPrepareCaptureWorkdirFailsOnAMissingFixture(t *testing.T) {
	if _, _, err := prepareCaptureWorkdir(t.TempDir(), "nope"); err == nil {
		t.Fatal("a missing fixture was accepted")
	}
}

func TestPrepareCaptureWorkdirWithNoFixtureIsEmpty(t *testing.T) {
	dir, cleanup, err := prepareCaptureWorkdir(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("scratch dir is not empty: %+v", entries)
	}
}
