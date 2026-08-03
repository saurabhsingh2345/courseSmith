package pipeline

import (
	"strings"
	"testing"
)

const cpNarration = "It asked for four things, the host handed over one, and it runs anyway."

func capabilitiesPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "capabilities",
		Title:    "It cannot open a file unless you let it",
		Capabilities: &CapabilitySpec{
			Subject:     "WASM module",
			SubjectNote: "app.wasm",
			Boundary:    "zero default access",
			Granter:     "the host",
			Items: []Capability{
				{Label: "files", Note: "Handed in as one directory", Role: "quantity"},
				{Label: "network", Note: "So it cannot phone home", Role: "limit"},
				{Label: "the clock", Note: "Denied, so timing attacks get harder", Role: "limit"},
				{Label: "random", Note: "Still shut unless passed a source", Role: "neutral"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "sealed", Heading: "Nothing by default", Narration: cpNarration, Capabilities: &CapabilityBeat{Show: "sealed"}},
			{ID: "one-dir", Heading: "One directory", Narration: cpNarration, Capabilities: &CapabilityBeat{Show: "grant", At: 0}},
			{ID: "shut", Heading: "What stays shut", Narration: cpNarration, Capabilities: &CapabilityBeat{Show: "read"}},
		},
	}
	p.targetWords = 3 * 40
	return p
}

func TestCapabilitiesPlanAccepted(t *testing.T) {
	if err := validateCapabilitiesPlan(capabilitiesPlan()); err != nil {
		t.Fatalf("a well-formed capabilities plan was rejected: %v", err)
	}
}

// The rule the template exists for, first half. A boundary with everything
// denied is a wall: correct, useless, and not what anyone ships.
func TestCapabilitiesRejectsAWallWithNothingGranted(t *testing.T) {
	p := capabilitiesPlan()
	p.Beats[1].Capabilities = &CapabilityBeat{Show: "read"}
	err := validateCapabilitiesPlan(p)
	if err == nil {
		t.Fatal("a boundary with nothing ever granted was accepted")
	}
	if !strings.Contains(err.Error(), "wall") {
		t.Fatalf("the error does not name what was drawn instead: %v", err)
	}
}

// The rule's second half. A boundary with everything granted is just the host.
func TestCapabilitiesRejectsGrantingEverything(t *testing.T) {
	p := capabilitiesPlan()
	p.Capabilities.Items = p.Capabilities.Items[:3]
	p.Beats = []SnippetBeat{
		{ID: "sealed", Heading: "Nothing by default", Narration: cpNarration, Capabilities: &CapabilityBeat{Show: "sealed"}},
		{ID: "g0", Heading: "Files", Narration: cpNarration, Capabilities: &CapabilityBeat{Show: "grant", At: 0}},
		{ID: "g1", Heading: "Network", Narration: cpNarration, Capabilities: &CapabilityBeat{Show: "grant", At: 1}},
		{ID: "g2", Heading: "The clock", Narration: cpNarration, Capabilities: &CapabilityBeat{Show: "grant", At: 2}},
	}
	p.targetWords = 4 * 40
	err := validateCapabilitiesPlan(p)
	if err == nil {
		t.Fatal("a boundary with everything granted was accepted")
	}
	if !strings.Contains(err.Error(), "GAP") {
		t.Fatalf("the error does not say what the picture teaches: %v", err)
	}
}

func TestCapabilitiesRequiresTheSealedBoundaryFirst(t *testing.T) {
	p := capabilitiesPlan()
	p.Beats[0].Capabilities = &CapabilityBeat{Show: "grant", At: 0}
	p.Beats[1].Capabilities = &CapabilityBeat{Show: "sealed"}
	err := validateCapabilitiesPlan(p)
	if err == nil {
		t.Fatal("a clip that grants before sealing was accepted")
	}
	if !strings.Contains(err.Error(), "permissions rather than a sandbox") {
		t.Fatalf("the error does not explain the difference: %v", err)
	}
}

// A capability once given is not taken back, so re-sealing would say it was.
func TestCapabilitiesRejectsResealing(t *testing.T) {
	p := capabilitiesPlan()
	p.Beats[2].Capabilities = &CapabilityBeat{Show: "sealed"}
	if err := validateCapabilitiesPlan(p); err == nil {
		t.Fatal("a clip that re-sealed the boundary was accepted")
	}
}

// Granting twice says the first grant did not take, which is the opposite of
// what a capability is.
func TestCapabilitiesRejectsGrantingTheSameThingTwice(t *testing.T) {
	p := capabilitiesPlan()
	p.Beats[2].Capabilities = &CapabilityBeat{Show: "grant", At: 0}
	if err := validateCapabilitiesPlan(p); err == nil {
		t.Fatal("a capability granted twice was accepted")
	}
}

func TestCapabilitiesRequiresASubject(t *testing.T) {
	p := capabilitiesPlan()
	p.Capabilities.Subject = ""
	if err := validateCapabilitiesPlan(p); err == nil {
		t.Fatal("a boundary with an empty middle was accepted")
	}
}

func TestCapabilitiesRejectsTooFewToShowTheGap(t *testing.T) {
	p := capabilitiesPlan()
	p.Capabilities.Items = p.Capabilities.Items[:2]
	if err := validateCapabilitiesPlan(p); err == nil {
		t.Fatal("a two-capability boundary was accepted")
	}
}

func TestCapabilitiesRejectsADuplicateLabel(t *testing.T) {
	p := capabilitiesPlan()
	p.Capabilities.Items[2].Label = "files"
	if err := validateCapabilitiesPlan(p); err == nil {
		t.Fatal("two capabilities with the same label were accepted")
	}
}

func TestCapabilitiesGranterHasADefault(t *testing.T) {
	c := &CapabilitySpec{}
	if c.ResolvedGranter() == "" {
		t.Fatal("a boundary that names no granter renders with a blank line")
	}
	c.Granter = "your manifest"
	if c.ResolvedGranter() != "your manifest" {
		t.Fatalf("a stated granter was ignored: %q", c.ResolvedGranter())
	}
}

// Each step carries the granted set as it stands, so the renderer draws a whole
// frame from one step rather than replaying the beats.
func TestCapabilitiesScenesAccumulateGrants(t *testing.T) {
	p := capabilitiesPlan()
	scenes, err := capabilitiesScenes(sceneInput(t, p, 12000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if got, _ := steps[0]["granted"].([]int); len(got) != 0 {
		t.Fatalf("the sealed beat already has grants: %v", got)
	}
	if got, _ := steps[1]["granted"].([]int); len(got) != 1 || got[0] != 0 {
		t.Fatalf("the grant beat carries %v, want [0]", got)
	}
	// The closing beat still carries it: a grant does not lapse when the
	// narrator moves on.
	if got, _ := steps[2]["granted"].([]int); len(got) != 1 {
		t.Fatalf("the closing beat lost the grant: %v", got)
	}
}
