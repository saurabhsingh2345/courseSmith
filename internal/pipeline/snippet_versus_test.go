package pipeline

import (
	"strings"
	"testing"
)

const duelNarration = "The row lands on both panels at once, and the side that wins this dimension takes the tint."

func versusPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "versus",
		Title:    "Reliability against raw speed",
		Versus: &VersusSpec{
			Left:  "TCP",
			Right: "UDP",
			Rows: []VersusRow{
				{Dim: "delivery", LeftVal: "guaranteed, in order", RightVal: "best effort only", Edge: "left"},
				{Dim: "setup cost", LeftVal: "a three way handshake", RightVal: "none at all", Edge: "right"},
				{Dim: "overhead", LeftVal: "twenty bytes per packet", RightVal: "eight bytes per packet", Edge: "right"},
				{Dim: "typical use", LeftVal: "web pages and email", RightVal: "live video and games", Edge: "even"},
			},
			Verdict: "Use TCP when losing data hurts more than waiting does",
		},
		Beats: []SnippetBeat{
			{ID: "face-off", Heading: "The contenders", Narration: duelNarration, Versus: &VersusBeat{Show: "face"}},
			{ID: "delivery", Heading: "Does it arrive", Narration: duelNarration, Versus: &VersusBeat{Show: "row", At: 0}},
			{ID: "setup", Heading: "What it costs", Narration: duelNarration, Versus: &VersusBeat{Show: "row", At: 1}},
			{ID: "overhead", Heading: "Bytes on the wire", Narration: duelNarration, Versus: &VersusBeat{Show: "row", At: 2}},
			{ID: "uses", Heading: "Where each lives", Narration: duelNarration, Versus: &VersusBeat{Show: "row", At: 3}},
			{ID: "the-call", Heading: "Which one", Narration: duelNarration, Versus: &VersusBeat{Show: "verdict"}},
		},
	}
	p.targetWords = 6 * 40
	return p
}

func TestVersusPlanAccepted(t *testing.T) {
	if err := validateVersusPlan(versusPlan()); err != nil {
		t.Fatalf("a well-formed head-to-head was rejected: %v", err)
	}
}

// A verdict that reduces to one contender's name is a preference rather than
// advice, and it is the commonest way this template stops being useful.
func TestVersusRejectsAVerdictThatIsJustOneSide(t *testing.T) {
	p := versusPlan()
	p.Versus.Verdict = "TCP."
	err := validateVersusPlan(p)
	if err == nil {
		t.Fatal("a verdict that is only the name of one side was accepted")
	}
	if !strings.Contains(err.Error(), "TCP") {
		t.Fatalf("the error does not quote the verdict back: %v", err)
	}
}

// The clean sweep is counted in Go: when every row goes the same way, the
// verdict has to be long enough to name the case the loser still wins.
func TestVersusRequiresALongerVerdictOnACleanSweep(t *testing.T) {
	p := versusPlan()
	for i := range p.Versus.Rows {
		p.Versus.Rows[i].Edge = "left"
	}
	p.Versus.Verdict = "Always reach for TCP"
	err := validateVersusPlan(p)
	if err == nil {
		t.Fatal("a clean sweep signed off with four words was accepted")
	}
	if !strings.Contains(err.Error(), "UDP") {
		t.Fatalf("the error does not name the side that lost every row: %v", err)
	}
	if !strings.Contains(err.Error(), "4 rows") {
		t.Fatalf("the error does not count the rows that swept: %v", err)
	}
}

func TestVersusAcceptsACleanSweepWithAnExplainedVerdict(t *testing.T) {
	p := versusPlan()
	for i := range p.Versus.Rows {
		p.Versus.Rows[i].Edge = "left"
	}
	p.Versus.Verdict = "Reach for UDP only when a dropped packet beats a late one"
	if err := validateVersusPlan(p); err != nil {
		t.Fatalf("a sweep with a verdict that explains itself was rejected: %v", err)
	}
}

func TestVersusRejectsBothSidesNamedTheSame(t *testing.T) {
	p := versusPlan()
	p.Versus.Right = "tcp"
	if err := validateVersusPlan(p); err == nil {
		t.Fatal("a comparison of one thing with itself was accepted")
	}
}

func TestVersusRejectsARowCountOutOfRange(t *testing.T) {
	p := versusPlan()
	p.Versus.Rows = p.Versus.Rows[:2]
	if err := validateVersusPlan(p); err == nil {
		t.Fatal("a two-row comparison was accepted, and two rows is an assertion with an example")
	}
}

func TestVersusRejectsARowWithAnEmptySide(t *testing.T) {
	p := versusPlan()
	p.Versus.Rows[2].RightVal = ""
	err := validateVersusPlan(p)
	if err == nil {
		t.Fatal("a row with only one cell filled was accepted")
	}
	if !strings.Contains(err.Error(), "overhead") {
		t.Fatalf("the error does not name the row: %v", err)
	}
}

func TestVersusRequiresOpeningOnTheFace(t *testing.T) {
	p := versusPlan()
	p.Beats[0].Versus = &VersusBeat{Show: "row", At: 0}
	if err := validateVersusPlan(p); err == nil {
		t.Fatal("a row landing before the contenders were named was accepted")
	}
}

func TestVersusRequiresClosingOnTheVerdict(t *testing.T) {
	p := versusPlan()
	p.Beats[5].Versus = &VersusBeat{Show: "face"}
	err := validateVersusPlan(p)
	if err == nil {
		t.Fatal("a comparison that never delivers its verdict was accepted")
	}
	if !strings.Contains(err.Error(), "verdict") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVersusRejectsRowsLandingOutOfOrder(t *testing.T) {
	p := versusPlan()
	p.Beats[2].Versus = &VersusBeat{Show: "row", At: 2}
	err := validateVersusPlan(p)
	if err == nil {
		t.Fatal("rows landing out of order were accepted")
	}
	if !strings.Contains(err.Error(), "overhead") || !strings.Contains(err.Error(), "setup cost") {
		t.Fatalf("the error does not quote the row landed and the one that was due: %v", err)
	}
}

func TestVersusRejectsLeavingARowUnlanded(t *testing.T) {
	p := versusPlan()
	p.Versus.Rows = append(p.Versus.Rows, VersusRow{Dim: "congestion", LeftVal: "backs off politely", RightVal: "keeps sending", Edge: "left"})
	err := validateVersusPlan(p)
	if err == nil {
		t.Fatal("a row nobody narrated was accepted")
	}
	if !strings.Contains(err.Error(), "congestion") {
		t.Fatalf("the error does not name the row left on screen: %v", err)
	}
}

// Over-long phrases and an unknown edge are phrasing, not wrong answers, so
// they are repaired rather than argued with.
func TestVersusNormalizeRepairsLabelsAndEdges(t *testing.T) {
	p := versusPlan()
	p.Versus.Left = "the transmission control protocol"
	p.Versus.Rows[0].Edge = "  LEFT "
	p.Versus.Rows[1].Edge = "sideways"
	p.Versus.Rows[2].LeftVal = "twenty whole bytes of header on every single packet"
	p.Versus.Verdict = "use it whenever the data really matters more than the latency does for this particular job"
	p.Beats[1].Versus.At = 77
	normalizeVersusPlan(p)

	if n := len(strings.Fields(p.Versus.Left)); n != maxVersusSideWords {
		t.Fatalf("the left contender normalized to %d words, want %d", n, maxVersusSideWords)
	}
	if p.Versus.Rows[0].Edge != "left" {
		t.Fatalf("a shouted edge normalized to %q, want left", p.Versus.Rows[0].Edge)
	}
	if p.Versus.Rows[1].Edge != "even" {
		t.Fatalf("an unknown edge normalized to %q, want even", p.Versus.Rows[1].Edge)
	}
	if n := len(strings.Fields(p.Versus.Rows[2].LeftVal)); n != maxVersusValWords {
		t.Fatalf("a cell normalized to %d words, want %d", n, maxVersusValWords)
	}
	if n := len(strings.Fields(p.Versus.Verdict)); n != maxVersusVerdictWords {
		t.Fatalf("the verdict normalized to %d words, want %d", n, maxVersusVerdictWords)
	}
	if p.Beats[1].Versus.At != 3 {
		t.Fatalf("an out-of-range row index normalized to %d, want the last row", p.Beats[1].Versus.At)
	}
}

func TestVersusEdgeDefaultsToEven(t *testing.T) {
	r := VersusRow{Edge: "both"}
	if got := r.ResolvedEdge(); got != "even" {
		t.Fatalf("an unknown edge resolved to %q, want even", got)
	}
}

func TestVersusShowDefaultsToRow(t *testing.T) {
	b := VersusBeat{Show: "duel"}
	if got := b.ResolvedShow(); got != "row" {
		t.Fatalf("an unknown show resolved to %q, want row", got)
	}
}

// The component paints a result it was handed: the tally and the sweep flag are
// counted in Go, and each step carries the rows that are on screen.
func TestVersusScenesTallyTheRows(t *testing.T) {
	p := versusPlan()
	scenes, err := versusScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	if props["leftWins"] != 1 || props["rightWins"] != 2 || props["evens"] != 1 {
		t.Fatalf("the tally is wrong: %v / %v / %v", props["leftWins"], props["rightWins"], props["evens"])
	}
	if props["sweep"] != false {
		t.Fatalf("a mixed comparison was flagged as a clean sweep")
	}

	steps, _ := props["steps"].([]map[string]any)
	first := steps[0]
	if first["show"] != "face" {
		t.Fatalf("the first step shows %v, want face", first["show"])
	}
	if up, _ := first["landed"].([]int); len(up) != 0 {
		t.Fatalf("no row has landed on the opener, but the step reports %v", up)
	}

	last := steps[len(steps)-1]
	if last["show"] != "verdict" {
		t.Fatalf("the last step shows %v, want verdict", last["show"])
	}
	up, _ := last["landed"].([]int)
	if len(up) != 4 || up[0] != 0 || up[3] != 3 {
		t.Fatalf("the verdict lands over %v, want every row sorted", up)
	}
}
