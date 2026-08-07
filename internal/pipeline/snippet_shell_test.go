package pipeline

import (
	"strings"
	"testing"
)

const shNarration = "Watch what the command actually prints back, because the flag in the middle is doing all the work."

func shellPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "shell",
		Title:    "What the long listing tells you",
		Shell: &ShellSpec{
			Host: "ubuntu",
			Entries: []ShellEntry{
				{
					Cmd: "ls -la",
					Output: []string{
						"total 24",
						"drwxr-xr-x 5 kai kai 4096 Feb 3 09:12 .",
						"-rw-r--r-- 1 kai kai  812 Feb 3 09:07 notes.txt",
					},
					Note: "the -l gives one line per file",
				},
				{
					Cmd:    "du -sh Documents",
					Output: []string{"1.2G    Documents"},
					Note:   "-h prints sizes people can read",
				},
			},
		},
		Beats: []SnippetBeat{
			{ID: "blank", Heading: "An empty prompt", Narration: shNarration, Shell: &ShellBeat{Show: "prompt"}},
			{ID: "type-ls", Heading: "Typing the command", Narration: shNarration, Shell: &ShellBeat{Show: "type", At: 0}},
			{ID: "read-ls", Heading: "Reading the columns", Narration: shNarration, Shell: &ShellBeat{Show: "output", At: 0}},
			{ID: "type-du", Heading: "Asking for a size", Narration: shNarration, Shell: &ShellBeat{Show: "type", At: 1}},
			{ID: "read-du", Heading: "One human number", Narration: shNarration, Shell: &ShellBeat{Show: "output", At: 1}},
			{ID: "session", Heading: "The whole session", Narration: shNarration, Shell: &ShellBeat{Show: "recap"}},
		},
	}
	p.targetWords = 6 * 40
	return p
}

func TestShellPlanAccepted(t *testing.T) {
	if err := validateShellPlan(shellPlan()); err != nil {
		t.Fatalf("a well-formed shell plan was rejected: %v", err)
	}
}

func TestShellRejectsAnEmptySession(t *testing.T) {
	p := shellPlan()
	p.Shell.Entries = nil
	if err := validateShellPlan(p); err == nil {
		t.Fatal("a terminal with no commands in it was accepted")
	}
}

func TestShellRejectsTooManyCommands(t *testing.T) {
	p := shellPlan()
	for i := 0; i < 3; i++ {
		p.Shell.Entries = append(p.Shell.Entries, ShellEntry{Cmd: "pwd"})
	}
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("a five-command session was accepted, and past four no command gets enough of the voice")
	}
	if !strings.Contains(err.Error(), "5 command(s)") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestShellRejectsAnEntryWithNoCommand(t *testing.T) {
	p := shellPlan()
	p.Shell.Entries[1].Cmd = "   "
	if err := validateShellPlan(p); err == nil {
		t.Fatal("an entry with nothing typed into it was accepted")
	}
}

// A session of narrated comments is a slide wearing a terminal's chrome.
func TestShellRejectsACommentAsACommand(t *testing.T) {
	p := shellPlan()
	p.Shell.Entries[0].Cmd = "# list everything in the folder"
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("a comment posing as a command was accepted")
	}
	if !strings.Contains(err.Error(), "comment") {
		t.Fatalf("the error does not say what a hash makes it: %v", err)
	}
	if !strings.Contains(err.Error(), "# list everything in the folder") {
		t.Fatalf("the error does not quote the line: %v", err)
	}
}

func TestShellRejectsAWallOfOutput(t *testing.T) {
	p := shellPlan()
	p.Shell.Entries[0].Output = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("nine lines of output were accepted, and they scroll the command out of view")
	}
	if !strings.Contains(err.Error(), "out of view") {
		t.Fatalf("the error does not say what it costs: %v", err)
	}
}

func TestShellRequiresOpeningOnTheEmptyPrompt(t *testing.T) {
	p := shellPlan()
	p.Beats[0].Shell = &ShellBeat{Show: "type", At: 0}
	if err := validateShellPlan(p); err == nil {
		t.Fatal("a session that starts mid-scroll was accepted")
	}
}

func TestShellRequiresClosingOnTheRecap(t *testing.T) {
	p := shellPlan()
	p.Beats[5].Shell = &ShellBeat{Show: "output", At: 1}
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("a clip that ends mid-command was accepted")
	}
	if !strings.Contains(err.Error(), "recap") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShellRejectsASecondEmptyPrompt(t *testing.T) {
	p := shellPlan()
	p.Beats[3].Shell = &ShellBeat{Show: "prompt"}
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("a blank screen part-way through was accepted, and a terminal is append-only")
	}
	if !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShellRejectsARecapBeforeTheEnd(t *testing.T) {
	p := shellPlan()
	p.Beats[3].Shell = &ShellBeat{Show: "recap"}
	if err := validateShellPlan(p); err == nil {
		t.Fatal("a recap before every command has run was accepted")
	}
}

func TestShellRejectsTypingACommandTwice(t *testing.T) {
	p := shellPlan()
	p.Beats[3].Shell = &ShellBeat{Show: "type", At: 0}
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("a command typed twice was accepted")
	}
	if !strings.Contains(err.Error(), "typed once") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// A terminal happens in order: the commands are typed exactly as listed.
func TestShellRejectsTypingOutOfOrder(t *testing.T) {
	p := shellPlan()
	p.Beats[1].Shell = &ShellBeat{Show: "type", At: 1}
	p.Beats[2].Shell = &ShellBeat{Show: "output", At: 1}
	p.Beats[3].Shell = &ShellBeat{Show: "type", At: 0}
	p.Beats[4].Shell = &ShellBeat{Show: "output", At: 0}
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("a session that runs the second command first was accepted")
	}
	if !strings.Contains(err.Error(), "entry 0 is next") {
		t.Fatalf("the error does not name the command that should have come first: %v", err)
	}
}

func TestShellRejectsATypeOffTheList(t *testing.T) {
	p := shellPlan()
	p.Beats[1].Shell = &ShellBeat{Show: "type", At: 9}
	if err := validateShellPlan(p); err == nil {
		t.Fatal("typing an entry that does not exist was accepted")
	}
}

// Output appearing before anyone typed is a terminal running backwards.
func TestShellRejectsOutputBeforeTheCommandIsTyped(t *testing.T) {
	p := shellPlan()
	p.Beats[1].Shell = &ShellBeat{Show: "output", At: 0}
	p.Beats[2].Shell = &ShellBeat{Show: "type", At: 0}
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("output landing before its command was typed was accepted")
	}
	if !strings.Contains(err.Error(), "running backwards") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// A terminal appends: it cannot show an earlier command's output after a later
// command has been typed.
func TestShellRejectsOutputForACommandThatIsNoLongerAtThePrompt(t *testing.T) {
	p := shellPlan()
	p.Beats[2].Shell = &ShellBeat{Show: "type", At: 1}
	p.Beats[3].Shell = &ShellBeat{Show: "output", At: 0}
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("an earlier command printing after a later one was typed was accepted")
	}
	if !strings.Contains(err.Error(), "at the prompt") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShellRejectsPrintingTheSameOutputTwice(t *testing.T) {
	p := shellPlan()
	p.Beats[3].Shell = &ShellBeat{Show: "output", At: 0}
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("an entry printing twice was accepted")
	}
	if !strings.Contains(err.Error(), "prints once") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShellRejectsACommandNeverTyped(t *testing.T) {
	p := shellPlan()
	p.Beats = []SnippetBeat{p.Beats[0], p.Beats[1], p.Beats[2], p.Beats[5]}
	p.targetWords = 4 * 40
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("a command nobody typed was accepted, and it is dead weight in the plan")
	}
	if !strings.Contains(err.Error(), "never typed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShellNormalizeClampsTheSession(t *testing.T) {
	p := shellPlan()
	p.Shell.Host = "my very own laptop"
	p.Shell.Entries[0].Cmd = "ls -la --color=always --time-style=full-iso /var/log/journal/remote/archive"
	p.Shell.Entries[0].Output = []string{
		strings.Repeat("x", 90), "2", "3", "4", "5", "6", "7", "8", "9", "10",
	}
	p.Shell.Entries[0].Note = "the dash l flag is the one that turns this into one line for every file"
	for i := 0; i < 4; i++ {
		p.Shell.Entries = append(p.Shell.Entries, ShellEntry{Cmd: "pwd"})
	}
	normalizeShellPlan(p)

	if n := len(strings.Fields(p.Shell.Host)); n != maxShellHostWords {
		t.Fatalf("the host label survived at %d words", n)
	}
	if n := len([]rune(p.Shell.Entries[0].Cmd)); n > maxShellCmdChars {
		t.Fatalf("the command survived at %d characters", n)
	}
	if n := len(p.Shell.Entries[0].Output); n != maxShellOutputLines {
		t.Fatalf("want %d output lines after normalize, got %d", maxShellOutputLines, n)
	}
	if n := len([]rune(p.Shell.Entries[0].Output[0])); n > maxShellOutputChars {
		t.Fatalf("an output line survived at %d characters", n)
	}
	if n := len(strings.Fields(p.Shell.Entries[0].Note)); n != maxShellNoteWords {
		t.Fatalf("the note survived at %d words", n)
	}
	if n := len(p.Shell.Entries); n != maxShellEntries {
		t.Fatalf("want %d entries after normalize, got %d", maxShellEntries, n)
	}
}

func TestShellNormalizeClampsBeatTargets(t *testing.T) {
	p := shellPlan()
	p.Beats[1].Shell.At = 99
	p.Beats[4].Shell.At = 99
	p.Beats[5].Shell.At = 1
	normalizeShellPlan(p)
	if at := p.Beats[1].Shell.At; at != len(p.Shell.Entries)-1 {
		t.Fatalf("want the type clamped to the last entry, got %d", at)
	}
	if at := p.Beats[4].Shell.At; at != len(p.Shell.Entries)-1 {
		t.Fatalf("want the output clamped to the last entry, got %d", at)
	}
	// The recap shows the whole session, so an index on it is noise.
	if at := p.Beats[5].Shell.At; at != 0 {
		t.Fatalf("the recap beat kept its index %d", at)
	}
}

func TestShellShowDefaultsToType(t *testing.T) {
	b := ShellBeat{Show: "scroll"}
	if got := b.ResolvedShow(); got != "type" {
		t.Fatalf("an unknown show resolved to %q, want type", got)
	}
	b = ShellBeat{Show: " RECAP "}
	if got := b.ResolvedShow(); got != "recap" {
		t.Fatalf("a shouted recap resolved to %q", got)
	}
}

func TestShellHostDefaultsToUbuntu(t *testing.T) {
	s := &ShellSpec{Host: "  "}
	if got := s.ResolvedHost(); got != "ubuntu" {
		t.Fatalf("an empty host resolved to %q, want ubuntu", got)
	}
	s = &ShellSpec{Host: "web-01"}
	if got := s.ResolvedHost(); got != "web-01" {
		t.Fatalf("a named host resolved to %q", got)
	}
}

// The session is append-only, so every frame is drawn from its own step: the
// typed and shown sets arrive accumulated, and the recap shows everything.
func TestShellScenesAccumulateTheSession(t *testing.T) {
	p := shellPlan()
	scenes, err := shellScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props
	if props["host"] != "ubuntu" {
		t.Fatalf("the scene lost its host: %v", props["host"])
	}
	entries, _ := props["entries"].([]map[string]any)
	if len(entries) != 2 {
		t.Fatalf("want 2 entries in the scene, got %d", len(entries))
	}
	if entries[1]["note"] != "-h prints sizes people can read" {
		t.Fatalf("the second entry lost its note: %v", entries[1])
	}

	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != len(p.Beats) {
		t.Fatalf("want %d steps, got %d", len(p.Beats), len(steps))
	}
	first := steps[0]
	typed, _ := first["typed"].([]int)
	shown, _ := first["shown"].([]int)
	if first["show"] != "prompt" || len(typed) != 0 || len(shown) != 0 {
		t.Fatalf("the blank prompt already has a session on it: %v", first)
	}

	last := steps[len(steps)-1]
	typedLast, _ := last["typed"].([]int)
	shownLast, _ := last["shown"].([]int)
	if last["show"] != "recap" {
		t.Fatalf("the last step is %v, want the recap", last["show"])
	}
	if len(typedLast) != 2 || typedLast[0] != 0 || typedLast[1] != 1 {
		t.Fatalf("the recap does not show every command typed: %v", typedLast)
	}
	if len(shownLast) != 2 || shownLast[0] != 0 || shownLast[1] != 1 {
		t.Fatalf("the recap does not show every command's output: %v", shownLast)
	}
}

// Output that no beat lands still reaches the screen at the recap, so it would
// arrive in the closing frame with nothing said about it.
func TestShellRequiresOutputToBeLanded(t *testing.T) {
	p := shellPlan()
	kept := p.Beats[:0]
	var dropped int
	for _, b := range p.Beats {
		if b.Shell != nil && b.Shell.ResolvedShow() == "output" && b.Shell.At == 1 {
			dropped++
			continue
		}
		kept = append(kept, b)
	}
	p.Beats = kept
	if dropped == 0 {
		t.Skip("the fixture has no output beat for entry 1")
	}
	p.targetWords = len(p.Beats) * 40
	err := validateShellPlan(p)
	if err == nil {
		t.Fatal("a command whose output no beat lands was accepted")
	}
	if !strings.Contains(err.Error(), "no beat ever lands") {
		t.Fatalf("the error does not name the unexplained output: %v", err)
	}
}

// A command that prints nothing needs no output beat — plenty of real commands
// succeed silently, and demanding a beat for a blank result is asking the model
// to narrate nothing.
func TestShellAllowsASilentCommand(t *testing.T) {
	p := shellPlan()
	kept := p.Beats[:0]
	for _, b := range p.Beats {
		if b.Shell != nil && b.Shell.ResolvedShow() == "output" && b.Shell.At == 1 {
			continue
		}
		kept = append(kept, b)
	}
	p.Beats = kept
	p.Shell.Entries[1].Output = nil
	p.targetWords = len(p.Beats) * 40
	if err := validateShellPlan(p); err != nil {
		t.Fatalf("a silent command was rejected: %v", err)
	}
}
