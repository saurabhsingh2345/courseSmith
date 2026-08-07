package pipeline

import (
	"strings"
	"testing"
)

const bpNarration = "The address goes out on one bus while the data it asked for comes back on another."

func blueprintPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "blueprint",
		Title:    "How the CPU talks to memory",
		Blueprint: &BlueprintSpec{
			Blocks: []BlueprintBlock{
				{ID: "cpu", Label: "CPU", Role: "unit"},
				{ID: "ram", Label: "Main memory", Role: "store"},
				{ID: "disk", Label: "Disk", Role: "io"},
			},
			Wires: []BlueprintWire{
				{From: "cpu", To: "ram", Label: "address bus"},
				{From: "ram", To: "cpu", Label: "data bus"},
				{From: "cpu", To: "disk", Label: "control lines"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "board", Heading: "The whole board", Narration: bpNarration, Blueprint: &BlueprintBeat{Show: "board"}},
			{ID: "cpu", Heading: "The unit in charge", Narration: bpNarration, Blueprint: &BlueprintBeat{Show: "block", At: 0}},
			{ID: "asking", Heading: "Asking for an address", Narration: bpNarration, Blueprint: &BlueprintBeat{Show: "path", At: 0}},
			{ID: "answering", Heading: "The bytes come back", Narration: bpNarration, Blueprint: &BlueprintBeat{Show: "path", At: 1}},
			{ID: "the-slow-road", Heading: "Out to the disk", Narration: bpNarration, Blueprint: &BlueprintBeat{Show: "path", At: 2}},
			{ID: "whole", Heading: "The whole machine", Narration: bpNarration, Blueprint: &BlueprintBeat{Show: "whole"}},
		},
	}
	// A beat of this template is a SHOT — one block forward or one pulse — so
	// the fixture's budget is sized against its 28-word ideal rather than the
	// shared forty, which would demand more beats than the picture has moves.
	p.targetWords = 6 * 28
	return p
}

func TestBlueprintPlanAccepted(t *testing.T) {
	if err := validateBlueprintPlan(blueprintPlan()); err != nil {
		t.Fatalf("a well-formed blueprint plan was rejected: %v", err)
	}
}

func TestBlueprintRejectsTooFewBlocks(t *testing.T) {
	p := blueprintPlan()
	p.Blueprint.Blocks = p.Blueprint.Blocks[:2]
	err := validateBlueprintPlan(p)
	if err == nil {
		t.Fatal("a two-block board was accepted, and two boxes with a line between them is a sentence")
	}
	if !strings.Contains(err.Error(), "2 block(s)") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestBlueprintRejectsTooFewWires(t *testing.T) {
	p := blueprintPlan()
	p.Blueprint.Wires = p.Blueprint.Wires[:1]
	if err := validateBlueprintPlan(p); err == nil {
		t.Fatal("a single-wire board was accepted, and one connection has no which-path question in it")
	}
}

// The wiring rule the template exists for: an endpoint that resolves to nothing
// is a connection to a component that is not on the board.
func TestBlueprintRejectsAWireToAnUnknownBlock(t *testing.T) {
	p := blueprintPlan()
	p.Blueprint.Wires[0].To = "gpu"
	err := validateBlueprintPlan(p)
	if err == nil {
		t.Fatal("a wire ending on a block that is not drawn was accepted")
	}
	if !strings.Contains(err.Error(), "gpu") {
		t.Fatalf("the error does not quote the bad id: %v", err)
	}
	if !strings.Contains(err.Error(), "cpu") || !strings.Contains(err.Error(), "ram") {
		t.Fatalf("the error does not list the ids that do exist: %v", err)
	}
}

func TestBlueprintRejectsAWireFromAnUnknownBlock(t *testing.T) {
	p := blueprintPlan()
	p.Blueprint.Wires[2].From = "cpu-core"
	err := validateBlueprintPlan(p)
	if err == nil {
		t.Fatal("a wire starting on a block that is not drawn was accepted")
	}
	if !strings.Contains(err.Error(), "cpu-core") {
		t.Fatalf("the error does not quote the bad id: %v", err)
	}
}

func TestBlueprintRejectsAWireToItself(t *testing.T) {
	p := blueprintPlan()
	p.Blueprint.Wires[0].To = "cpu"
	err := validateBlueprintPlan(p)
	if err == nil {
		t.Fatal("a wire from a block to itself was accepted")
	}
	if !strings.Contains(err.Error(), "itself") {
		t.Fatalf("the error does not say what is wrong with it: %v", err)
	}
}

func TestBlueprintRejectsTwoBlocksSharingAnID(t *testing.T) {
	p := blueprintPlan()
	p.Blueprint.Blocks[1].ID = "cpu"
	err := validateBlueprintPlan(p)
	if err == nil {
		t.Fatal("two blocks with the same id were accepted, and a wire pointing at it could mean either")
	}
	if !strings.Contains(err.Error(), "share the id") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBlueprintRejectsAnUnlabelledBlock(t *testing.T) {
	p := blueprintPlan()
	p.Blueprint.Blocks[1].Label = "   "
	if err := validateBlueprintPlan(p); err == nil {
		t.Fatal("an unlabelled block was accepted, and an unlabelled box never gets explained")
	}
}

func TestBlueprintRejectsAnInventedRole(t *testing.T) {
	p := blueprintPlan()
	p.Blueprint.Blocks[1].Role = "gadget"
	err := validateBlueprintPlan(p)
	if err == nil {
		t.Fatal("a block role outside the vocabulary was accepted, and an invented kind has no look")
	}
	if !strings.Contains(err.Error(), "gadget") {
		t.Fatalf("the error does not quote the invented role: %v", err)
	}
}

func TestBlueprintRequiresOpeningOnTheBoard(t *testing.T) {
	p := blueprintPlan()
	p.Beats[0].Blueprint = &BlueprintBeat{Show: "block", At: 0}
	if err := validateBlueprintPlan(p); err == nil {
		t.Fatal("a clip that focuses a block before the board is up was accepted")
	}
}

func TestBlueprintRejectsASecondBoardBeat(t *testing.T) {
	p := blueprintPlan()
	p.Beats[1].Blueprint = &BlueprintBeat{Show: "board"}
	err := validateBlueprintPlan(p)
	if err == nil {
		t.Fatal("a clip that returns to the establishing shot was accepted")
	}
	if !strings.Contains(err.Error(), "re-opens") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBlueprintRejectsAFocusOffTheBoard(t *testing.T) {
	p := blueprintPlan()
	p.Beats[1].Blueprint = &BlueprintBeat{Show: "block", At: 9}
	if err := validateBlueprintPlan(p); err == nil {
		t.Fatal("a focus on a block that does not exist was accepted")
	}
}

func TestBlueprintRejectsAPathOnAWireThatDoesNotExist(t *testing.T) {
	p := blueprintPlan()
	p.Beats[2].Blueprint = &BlueprintBeat{Show: "path", At: 9}
	if err := validateBlueprintPlan(p); err == nil {
		t.Fatal("a pulse along a wire that is not drawn was accepted")
	}
}

// The wires ARE the content: repeating a lit one while another sits dark leaves
// a drawn connection unexplained, which is the textbook diagram this fixes.
func TestBlueprintRejectsRepeatingAWireWhileAnotherIsDark(t *testing.T) {
	p := blueprintPlan()
	p.Beats[4].Blueprint = &BlueprintBeat{Show: "path", At: 0}
	err := validateBlueprintPlan(p)
	if err == nil {
		t.Fatal("a second pulse on wire 0 was accepted while wire 2 had never lit")
	}
	if !strings.Contains(err.Error(), "never lit") {
		t.Fatalf("the error does not say what went unexplained: %v", err)
	}
}

// Once every wire has had its turn a repeat is fine — the rule is about
// coverage, not about repetition for its own sake.
func TestBlueprintAllowsARepeatOnceEveryWireHasLit(t *testing.T) {
	p := blueprintPlan()
	// Inserted BEFORE the closer rather than over it. The last beat has to stay
	// the whole board — overwriting it tests the closer rule instead of the
	// coverage rule, and would pass for the wrong reason if the closer rule
	// were ever dropped.
	repeat := SnippetBeat{ID: "again", Heading: "Back down the bus", Narration: bpNarration,
		Blueprint: &BlueprintBeat{Show: "path", At: 0}}
	p.Beats = append(p.Beats[:5], append([]SnippetBeat{repeat}, p.Beats[5:]...)...)
	p.targetWords = len(p.Beats) * 28
	if err := validateBlueprintPlan(p); err != nil {
		t.Fatalf("a repeat after full coverage was rejected: %v", err)
	}
}

// The closer is a rule, not a convention: a clip that lights one edge at a time
// and stops has never shown the machine working as a machine.
func TestBlueprintRequiresTheWholeBoardAtTheEnd(t *testing.T) {
	p := blueprintPlan()
	p.Beats[5].Blueprint = &BlueprintBeat{Show: "path", At: 0}
	err := validateBlueprintPlan(p)
	if err == nil {
		t.Fatal("a clip that never lights the whole board was accepted")
	}
	if !strings.Contains(err.Error(), "whole") {
		t.Fatalf("the error does not name the missing closer: %v", err)
	}
}

// And it may only be the last shot: once everything is on, a later beat is a
// step backwards.
func TestBlueprintRejectsTheWholeBoardPartWayThrough(t *testing.T) {
	p := blueprintPlan()
	p.Beats[2].Blueprint = &BlueprintBeat{Show: "whole"}
	err := validateBlueprintPlan(p)
	if err == nil {
		t.Fatal("lighting the whole board mid-clip was accepted")
	}
	if !strings.Contains(err.Error(), "part-way through") {
		t.Fatalf("the error does not say why an early whole is wrong: %v", err)
	}
}

// The closer lights every wire, including any the beats never pulsed. Without
// it the final frame is indistinguishable from the beat before it.
func TestBlueprintClosingShotLightsEveryWire(t *testing.T) {
	p := blueprintPlan()
	scenes, err := blueprintScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	lit, _ := steps[len(steps)-1]["lit"].([]int)
	if len(lit) != len(p.Blueprint.Wires) {
		t.Fatalf("the closing frame lights %d of %d wires", len(lit), len(p.Blueprint.Wires))
	}
}

func TestBlueprintNormalizeSlugifiesIDs(t *testing.T) {
	p := blueprintPlan()
	p.Blueprint.Blocks[0].ID = "  CPU Core "
	p.Blueprint.Wires[0].From = "CPU Core"
	p.Blueprint.Wires[1].To = "CPU  Core"
	p.Blueprint.Wires[2].From = "cpu core"
	normalizeBlueprintPlan(p)
	if got := p.Blueprint.Blocks[0].ID; got != "cpu-core" {
		t.Fatalf("the block id normalized to %q, want cpu-core", got)
	}
	if err := validateBlueprintPlan(p); err != nil {
		t.Fatalf("a plan whose ids only differed in shouting was rejected after normalize: %v", err)
	}
}

func TestBlueprintNormalizeClampsLabels(t *testing.T) {
	p := blueprintPlan()
	p.Blueprint.Blocks[1].Label = "the main memory of this whole machine"
	p.Blueprint.Wires[0].Label = "the bus that carries the address"
	normalizeBlueprintPlan(p)
	if n := len(strings.Fields(p.Blueprint.Blocks[1].Label)); n != maxBlueprintLabelWords {
		t.Fatalf("the block label survived at %d words", n)
	}
	if n := len(strings.Fields(p.Blueprint.Wires[0].Label)); n != maxBlueprintWireLabelWords {
		t.Fatalf("the wire label survived at %d words", n)
	}
}

func TestBlueprintNormalizeCapsTheBoard(t *testing.T) {
	p := blueprintPlan()
	for i := 0; i < 6; i++ {
		p.Blueprint.Blocks = append(p.Blueprint.Blocks, BlueprintBlock{ID: "extra", Label: "Extra"})
		p.Blueprint.Wires = append(p.Blueprint.Wires, BlueprintWire{From: "cpu", To: "ram"})
	}
	normalizeBlueprintPlan(p)
	if n := len(p.Blueprint.Blocks); n != maxBlueprintBlocks {
		t.Fatalf("want %d blocks after normalize, got %d", maxBlueprintBlocks, n)
	}
	if n := len(p.Blueprint.Wires); n != maxBlueprintWires {
		t.Fatalf("want %d wires after normalize, got %d", maxBlueprintWires, n)
	}
}

func TestBlueprintNormalizeClampsBeatTargets(t *testing.T) {
	p := blueprintPlan()
	p.Beats[1].Blueprint.At = 99
	p.Beats[2].Blueprint.At = 99
	p.Beats[5].Blueprint.At = 4
	normalizeBlueprintPlan(p)
	if got := p.Beats[1].Blueprint.At; got != len(p.Blueprint.Blocks)-1 {
		t.Fatalf("a block focus off the end clamps to the last block, got %d", got)
	}
	if got := p.Beats[2].Blueprint.At; got != len(p.Blueprint.Wires)-1 {
		t.Fatalf("a path off the end clamps to the last wire, got %d", got)
	}
	// A beat that does not act on an index carries none: "whole" lights
	// everything, so an index on it is noise the renderer would have to ignore.
	if got := p.Beats[5].Blueprint.At; got != 0 {
		t.Fatalf("a non-acting beat kept its index %d", got)
	}
}

func TestBlueprintShowDefaultsToPath(t *testing.T) {
	b := BlueprintBeat{Show: "sparkle"}
	if got := b.ResolvedShow(); got != "path" {
		t.Fatalf("an unknown show resolved to %q, want path", got)
	}
	b = BlueprintBeat{Show: " BOARD "}
	if got := b.ResolvedShow(); got != "board" {
		t.Fatalf("a shouted board resolved to %q", got)
	}
}

func TestBlueprintRoleDefaultsToUnit(t *testing.T) {
	b := BlueprintBlock{Role: "gadget"}
	if got := b.ResolvedRole(); got != "unit" {
		t.Fatalf("an unknown role resolved to %q, want unit", got)
	}
}

// The renderer never matches strings to find where a pulse starts: the wire
// endpoints arrive as block indices, and the lit set arrives accumulated.
func TestBlueprintScenesResolveEndpointsAndAccumulateLitWires(t *testing.T) {
	p := blueprintPlan()
	scenes, err := blueprintScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	wires, _ := props["wires"].([]map[string]any)
	if len(wires) != 3 {
		t.Fatalf("want 3 wires in the scene, got %d", len(wires))
	}
	if wires[0]["from"] != 0 || wires[0]["to"] != 1 {
		t.Fatalf("the first wire's endpoints are not block indices: %v", wires[0])
	}
	if wires[2]["from"] != 0 || wires[2]["to"] != 2 {
		t.Fatalf("the last wire's endpoints are wrong: %v", wires[2])
	}

	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != len(p.Beats) {
		t.Fatalf("want %d steps, got %d", len(p.Beats), len(steps))
	}
	first := steps[0]
	lit, _ := first["lit"].([]int)
	if first["show"] != "board" || len(lit) != 0 {
		t.Fatalf("the opening board already has lit wires: %v", first)
	}
	if _, ok := first["at"]; ok {
		t.Fatalf("the board beat carries an index it does not act on: %v", first)
	}
	last := steps[len(steps)-1]
	litLast, _ := last["lit"].([]int)
	if last["show"] != "whole" || len(litLast) != 3 {
		t.Fatalf("the closer does not carry every pulsed wire: %v", last)
	}
	for i, at := range litLast {
		if at != i {
			t.Fatalf("the lit set is not sorted or complete: %v", litLast)
		}
	}
}
