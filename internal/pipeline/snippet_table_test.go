package pipeline

import (
	"strings"
	"testing"
)

const tbNarration = "Six numbers on the page, and the one that decides it is sitting fifth in the same weight as the rest."

func tablePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "table",
		Title:    "The line the spec sheet buries",
		Table: &TableSpec{
			Source: "RTX 5090, from the product page",
			Rows: []TableRow{
				{Label: "CUDA Cores", Value: "21,760"},
				{Label: "Tensor Cores", Value: "680"},
				{Label: "Boost Clock", Value: "2.52 GHz"},
				{Label: "Memory Bandwidth", Value: "1,792 GB/s"},
				{Label: "Memory Capacity", Value: "32 GB"},
				{Label: "TDP", Value: "575 W"},
			},
			At:   4,
			Note: "Every other number is irrelevant if the model does not fit",
			Role: "limit",
		},
		Beats: []SnippetBeat{
			{ID: "sheet", Heading: "The spec sheet", Narration: tbNarration, Table: &TableBeat{Show: "sheet"}},
			{ID: "row", Heading: "The one that counts", Narration: tbNarration, Table: &TableBeat{Show: "focus"}},
			{ID: "decides", Heading: "What it decides", Narration: tbNarration, Table: &TableBeat{Show: "read"}},
		},
	}
	p.targetWords = 3 * 40
	return p
}

func TestTablePlanAccepted(t *testing.T) {
	if err := validateTablePlan(tablePlan()); err != nil {
		t.Fatalf("a well-formed table plan was rejected: %v", err)
	}
}

// The rule the template exists for. A number at the top of a sheet is the
// headline and one at the bottom is the summary; neither is buried, and a clip
// claiming otherwise makes a rhetorical point its own picture contradicts.
func TestTableRejectsARowThatIsNotActuallyBuried(t *testing.T) {
	for _, at := range []int{0, 5} {
		p := tablePlan()
		p.Table.At = at
		err := validateTablePlan(p)
		if err == nil {
			t.Fatalf("a row at index %d was accepted as buried", at)
		}
		if !strings.Contains(err.Error(), "metric template") {
			t.Fatalf("the error does not name the alternative: %v", err)
		}
	}
}

// The burial only reads if the viewer has seen the row looking ordinary.
func TestTableRequiresTheWholeSheetFirst(t *testing.T) {
	p := tablePlan()
	p.Beats[0].Table = &TableBeat{Show: "focus"}
	p.Beats[1].Table = &TableBeat{Show: "sheet"}
	err := validateTablePlan(p)
	if err == nil {
		t.Fatal("a clip that focuses before showing the sheet was accepted")
	}
	if !strings.Contains(err.Error(), "highlight rather than a burial") {
		t.Fatalf("the error does not name the difference: %v", err)
	}
}

// Once the weighting is gone it does not come back.
func TestTableRejectsRestoringTheSheet(t *testing.T) {
	p := tablePlan()
	p.Beats[2].Table = &TableBeat{Show: "sheet"}
	if err := validateTablePlan(p); err == nil {
		t.Fatal("a clip that restored the even weighting was accepted")
	}
}

func TestTableRequiresExactlyOneFocus(t *testing.T) {
	p := tablePlan()
	p.Beats[2].Table = &TableBeat{Show: "focus"}
	if err := validateTablePlan(p); err == nil {
		t.Fatal("a clip focusing twice was accepted")
	}
}

// With three rows and the ends excluded there is exactly one legal position, so
// the burial stops being a choice.
func TestTableRejectsASheetTooShortToBuryIn(t *testing.T) {
	p := tablePlan()
	p.Table.Rows = p.Table.Rows[:3]
	p.Table.At = 1
	if err := validateTablePlan(p); err == nil {
		t.Fatal("a three-row sheet was accepted")
	}
}

// Lighting a row is the gesture; saying what it decides is the content.
func TestTableRequiresANote(t *testing.T) {
	p := tablePlan()
	p.Table.Note = ""
	err := validateTablePlan(p)
	if err == nil {
		t.Fatal("a buried row with nothing said about it was accepted")
	}
	if !strings.Contains(err.Error(), "pointed at a number and stopped") {
		t.Fatalf("the error does not say what is missing: %v", err)
	}
}

func TestTableRejectsARowWithNoValue(t *testing.T) {
	p := tablePlan()
	p.Table.Rows[2].Value = ""
	if err := validateTablePlan(p); err == nil {
		t.Fatal("a sheet line with an empty right-hand column was accepted")
	}
}

func TestTableRejectsADuplicateSpec(t *testing.T) {
	p := tablePlan()
	p.Table.Rows[3].Label = "CUDA Cores"
	if err := validateTablePlan(p); err == nil {
		t.Fatal("a sheet listing the same spec twice was accepted")
	}
}

// Values are cut by characters: "1,792 GB/s" clamped to a word count is not a
// spec sheet value.
func TestTableNormalizeClampsValuesByCharacters(t *testing.T) {
	p := tablePlan()
	p.Table.Rows[0].Value = strings.Repeat("9", maxTableValueChars+10)
	normalizeTablePlan(p)
	if got := len(p.Table.Rows[0].Value); got > maxTableValueChars {
		t.Fatalf("value is %d chars, want at most %d", got, maxTableValueChars)
	}
}

// An index off the end of the sheet is clamped rather than rejected: which row
// the model meant is not recoverable, but a sheet with a focus on its last row is
// at least a drawable frame, and the validator then explains why it is wrong.
func TestTableNormalizeClampsTheIndexOntoTheSheet(t *testing.T) {
	p := tablePlan()
	p.Table.At = 99
	normalizeTablePlan(p)
	if p.Table.At != len(p.Table.Rows)-1 {
		t.Fatalf("at is %d after normalize, want the last row %d", p.Table.At, len(p.Table.Rows)-1)
	}
}

// The buried row's role defaults to the limit: the number a sheet buries is
// almost always the one that stops you.
func TestTableRoleDefaultsToLimit(t *testing.T) {
	s := &TableSpec{}
	if s.ResolvedRole() != "limit" {
		t.Fatalf("default role is %q, want limit", s.ResolvedRole())
	}
}

func TestTableScenesCarryTheSheetAndTheIndex(t *testing.T) {
	p := tablePlan()
	scenes, err := tableScenes(sceneInput(t, p, 12000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	rows, _ := scenes[0].Props["rows"].([]map[string]any)
	if len(rows) != 6 {
		t.Fatalf("want 6 rows in the scene, got %d", len(rows))
	}
	// The order is the subject, so it has to survive into the scene unchanged.
	if rows[4]["label"] != "Memory Capacity" {
		t.Fatalf("row 4 is %v, want Memory Capacity — the sheet's order is the point", rows[4]["label"])
	}
	if scenes[0].Props["at"] != 4 {
		t.Fatalf("the buried index reached the scene as %v, want 4", scenes[0].Props["at"])
	}
}
