package pipeline

import (
	"strings"
	"testing"
)

const myNarration = "Everybody reaches for Redis as a cache first, and it is genuinely very good at that."

func mythPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "myth",
		Title:    "What everyone gets wrong about Redis",
		Myth: &MythSpec{
			Claim: "Redis is just a cache",
			Truth: "Redis is a data structure server",
			Why:   "Because the first thing anyone uses it for is caching",
			Evidence: []string{
				"Sorted sets give you a leaderboard",
				"Streams give you a durable log",
				"Lua scripts run atomically",
			},
		},
		Beats: []SnippetBeat{
			{ID: "claim", Heading: "What everyone says", Narration: myNarration, Myth: &MythBeat{Show: "claim"}},
			{ID: "strike", Heading: "Not quite", Narration: myNarration, Myth: &MythBeat{Show: "strike"}},
			{ID: "sets", Heading: "Sorted sets", Narration: myNarration, Myth: &MythBeat{Show: "evidence", At: 0}},
			{ID: "streams", Heading: "Streams", Narration: myNarration, Myth: &MythBeat{Show: "evidence", At: 1}},
			{ID: "lua", Heading: "Atomic scripts", Narration: myNarration, Myth: &MythBeat{Show: "evidence", At: 2}},
			{ID: "why", Heading: "Why we thought so", Narration: myNarration, Myth: &MythBeat{Show: "why"}},
		},
	}
	p.targetWords = 6 * 40
	return p
}

func TestMythPlanAccepted(t *testing.T) {
	if err := validateMythPlan(mythPlan()); err != nil {
		t.Fatalf("a well-formed myth was rejected: %v", err)
	}
}

// The rule this template exists for. A viewer left with a hole where their
// model used to be is worse off than one with a wrong model.
func TestMythRejectsMereNegation(t *testing.T) {
	for _, truth := range []string{
		"Redis is not just a cache",
		"Redis isn't just a cache",
		"Not just a cache",
		"No, Redis is more than that",
		"Wrong",
	} {
		p := mythPlan()
		p.Myth.Truth = truth
		if err := validateMythPlan(p); err == nil {
			t.Errorf("the correction %q was accepted, but it only denies the claim", truth)
		}
	}
}

// A real statement survives, including one that happens to contain a negation
// somewhere — the guard is for corrections that are *nothing but* a denial.
func TestMythAcceptsRealCorrections(t *testing.T) {
	for _, truth := range []string{
		"Redis is a data structure server",
		"Redis is a database that keeps everything in memory",
		"Sorted sets make it a ranking engine",
	} {
		p := mythPlan()
		p.Myth.Truth = truth
		if err := validateMythPlan(p); err != nil {
			t.Errorf("the correction %q was rejected: %v", truth, err)
		}
	}
}

func TestMythOpensOnTheClaim(t *testing.T) {
	p := mythPlan()
	p.Beats[0].Myth = &MythBeat{Show: "why"}
	p.Beats[5].Myth = &MythBeat{Show: "claim"}
	err := validateMythPlan(p)
	if err == nil {
		t.Fatal("a clip that did not open on the belief was accepted")
	}
	if !strings.Contains(err.Error(), "corrects nobody") {
		t.Errorf("the error does not say why: %v", err)
	}
}

// Nothing can be evidence for a claim the clip has not made yet.
func TestMythEvidenceComesAfterTheStrike(t *testing.T) {
	p := mythPlan()
	p.Beats[1].Myth = &MythBeat{Show: "evidence", At: 0}
	p.Beats[2].Myth = &MythBeat{Show: "strike"}
	err := validateMythPlan(p)
	if err == nil {
		t.Fatal("evidence before the strike was accepted")
	}
	if !strings.Contains(err.Error(), "has not made yet") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMythStrikesExactlyOnce(t *testing.T) {
	p := mythPlan()
	p.Beats[3].Myth = &MythBeat{Show: "strike"}
	err := validateMythPlan(p)
	if err == nil {
		t.Fatal("two strike beats were accepted")
	}
	if !strings.Contains(err.Error(), "exactly once") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMythRequiresEvidence(t *testing.T) {
	p := mythPlan()
	p.Myth.Evidence = nil
	err := validateMythPlan(p)
	if err == nil {
		t.Fatal("a correction with nothing behind it was accepted")
	}
	if !strings.Contains(err.Error(), "different assertion") {
		t.Errorf("unexpected error: %v", err)
	}
}

// The concession keeps the clip from being smug, and a smug correction is one
// the viewer argues with rather than accepts.
func TestMythWhyMustBeNarratedIfWritten(t *testing.T) {
	p := mythPlan()
	// Leave the shape otherwise legal and make only the concession unspoken:
	// give the `why` beat a fourth piece of evidence to carry instead. Every
	// other rule — one claim, one strike, no repeats, full coverage — still
	// passes, so the failure can only be the one under test.
	p.Myth.Evidence = append(p.Myth.Evidence, "Pub/sub fans messages out")
	p.Beats[5].Myth = &MythBeat{Show: "evidence", At: 3}

	err := validateMythPlan(p)
	if err == nil {
		t.Fatal("a written-but-unspoken concession was accepted")
	}
	if !strings.Contains(err.Error(), "smug") {
		t.Errorf("the error does not say why it matters: %v", err)
	}
}

func TestMythNormalizeRepairs(t *testing.T) {
	p := mythPlan()
	p.Beats[2].Myth.Show = "nonsense"
	p.Beats[3].Myth.At = 99
	p.Myth.Claim = strings.Repeat("word ", 30)
	normalizeMythPlan(p)

	if p.Beats[2].Myth.Show != "evidence" {
		t.Errorf("an unknown show became %q, want evidence", p.Beats[2].Myth.Show)
	}
	if p.Beats[3].Myth.At != len(p.Myth.Evidence)-1 {
		t.Errorf("an out-of-range beat points at %d", p.Beats[3].Myth.At)
	}
	if n := len(strings.Fields(p.Myth.Claim)); n > maxMythClaimWords {
		t.Errorf("the claim is %d words after normalize, want <= %d", n, maxMythClaimWords)
	}
}

func TestMythScenesShape(t *testing.T) {
	p := mythPlan()
	scenes, err := mythScenes(sceneInput(t, p, 7000))
	if err != nil {
		t.Fatalf("mythScenes: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneMyth {
		t.Fatalf("want one myth scene, got %d of %q", len(scenes), scenes[0].Type)
	}
	// The claim and the truth are both scene-level props: the claim stays on
	// screen struck through for the rest of the clip, so it cannot live on the
	// step that happened to state it.
	if scenes[0].Props["claim"] != p.Myth.Claim || scenes[0].Props["truth"] != p.Myth.Truth {
		t.Error("the claim and truth are not both on the scene")
	}
	steps := scenes[0].Props["steps"].([]map[string]any)
	if steps[1]["show"] != "strike" {
		t.Errorf("step 1 = %v, want the strike", steps[1]["show"])
	}
}
