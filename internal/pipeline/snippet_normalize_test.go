package pipeline

import (
	"reflect"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

// The reflective strip in stripBeatFields pairs a beatFields flag with the
// SnippetBeat field of the same name. Nothing but this test holds that pairing:
// a beat field added under a name beatFields does not use would never be
// stripped, and a template would silently render another template's payload.
func TestBeatFieldsNameSnippetBeatFields(t *testing.T) {
	beat := reflect.TypeFor[SnippetBeat]()
	owns := reflect.TypeFor[beatFields]()
	for i := range owns.NumField() {
		name := owns.Field(i).Name
		if _, ok := beat.FieldByName(name); !ok {
			t.Errorf("beatFields.%s names no field on SnippetBeat — stripBeatFields cannot clear it", name)
		}
	}
	// And the other way: every optional payload on a beat must be claimable, or
	// no template can own it and every template will strip it.
	for i := range beat.NumField() {
		f := beat.Field(i)
		switch f.Name {
		case "ID", "Heading", "Narration":
			continue // the fields every beat has
		}
		if _, ok := owns.FieldByName(f.Name); !ok {
			t.Errorf("SnippetBeat.%s is not named by beatFields — no template can declare it", f.Name)
		}
	}
}

// Every template has to say what it reads, or normalization will strip the very
// payload it renders.
func TestEveryTemplateDeclaresWhatItOwns(t *testing.T) {
	for _, name := range SnippetTemplateNames() {
		tpl := SnippetTemplates[name]
		if tpl.Owns == (beatFields{}) && tpl.OwnsPlan == (planFields{}) {
			t.Errorf("template %q declares no Owns — every beat payload it plans will be stripped", name)
		}
	}
}

// Owns and the template's own rejectForeignBeatFields call are two statements of
// the same fact, and they used to be able to disagree. A plan carrying every
// payload at once, normalized, must come out the other side without the
// validator complaining that a field belongs to someone else: whatever the
// template owns survived, and whatever it does not was stripped.
func TestNormalizingStripsWhatEveryTemplateDoesNotOwn(t *testing.T) {
	for _, name := range SnippetTemplateNames() {
		tpl := SnippetTemplates[name]
		if tpl.Validate == nil {
			continue
		}
		p := &SnippetPlan{Template: name, Title: "Everything at once"}
		for i := range 4 {
			p.Beats = append(p.Beats, SnippetBeat{
				ID:        "beat",
				Heading:   "A heading here",
				Narration: strings.Repeat("word ", 30),
				Code:      "print('hi')\n",
				Run:       i == 2,
				Sketch:    []SketchItem{{Label: "A box", Icon: "server"}},
				Nodes:     []FlowNode{{ID: "n", Label: "A node", Kind: "service"}},
				Focus:     []string{"n"},
				Art:       &ArtBeat{Figure: "server"},
				Cast:      &CastBeat{Pose: "point"},
				Shot:      &ShotBeat{BeatID: "beat", Staging: "hero", Camera: "hold"},
				Data:      &DataBeat{Caption: "a caption"},
				Work:      &WorkspaceBeat{File: "main.py"},
				Quiz:      &QuizBeat{Show: "think"},
			})
		}
		normalizeSnippetPlan(p)
		if err := tpl.Validate(p); err != nil && strings.Contains(err.Error(), "does not use") {
			t.Errorf("template %q: %v — Owns and its rejectForeignBeatFields call disagree", name, err)
		}
	}
}

// A model that answers a whiteboard request with flow nodes has understood the
// clip and mislabelled the field.
func TestNormalizeMigratesNodesOntoTheBoard(t *testing.T) {
	p := &SnippetPlan{
		Template: "whiteboard",
		Title:    "How a request travels",
		Beats: []SnippetBeat{
			{ID: "one", Heading: "The client", Narration: strings.Repeat("word ", 30), Nodes: []FlowNode{
				{ID: "browser", Label: "Browser", Kind: "client"},
			}},
			{ID: "two", Heading: "The server", Narration: strings.Repeat("word ", 30), Nodes: []FlowNode{
				{ID: "api", Label: "API server", Kind: "service", From: []string{"browser"}},
			}},
		},
	}
	normalizeSnippetPlan(p)
	if len(p.Beats[0].Sketch) != 1 || p.Beats[0].Sketch[0].Label != "Browser" {
		t.Fatalf("the node did not become a board item: %+v", p.Beats[0])
	}
	// The icon comes from the node's kind, and the link from the node it was fed
	// by — by label, which is how the board addresses its boxes.
	if got := p.Beats[0].Sketch[0].Icon; got != "monitor" {
		t.Errorf("client node drew %q, want monitor", got)
	}
	if got := p.Beats[1].Sketch[0].LinkFrom; got != "Browser" {
		t.Errorf("edge became link_from %q, want Browser", got)
	}
	if p.Beats[0].Nodes != nil {
		t.Error("the nodes were left on the plan as well as migrated")
	}
}

func TestNormalizeWhiteboardPlan(t *testing.T) {
	p := whiteboardPlan()
	p.Beats[1].Sketch[0].Label = "  The   browser  " // stray whitespace
	p.Beats[2].Sketch[0].Label = "a label with far too many words"
	p.Beats[2].Sketch[0].Shape = "rectangle"       // not in the vocabulary
	p.Beats[2].Sketch[0].Icon = "a-drawing-of-fog" // nor this
	p.Beats[3].Sketch[0].LinkFrom = "Nothing here" // points at no box
	p.Beats[3].Sketch = append(p.Beats[3].Sketch, SketchItem{Label: "The browser"})

	normalizeSnippetPlan(p)

	if got := p.Beats[1].Sketch[0].Label; got != "The browser" {
		t.Errorf("label %q was not collapsed", got)
	}
	if got := p.Beats[2].Sketch[0].Label; got != "a label with far" {
		t.Errorf("label %q was not cut to %d words", got, maxSketchLabelWords)
	}
	if got := p.Beats[2].Sketch[0].Shape; got != "box" {
		t.Errorf("shape %q was not degraded to a box", got)
	}
	if got := p.Beats[2].Sketch[0].Icon; got != "spark" {
		t.Errorf("icon %q was not degraded to the neutral figure", got)
	}
	if got := p.Beats[3].Sketch[0].LinkFrom; got != "" {
		t.Errorf("link to a box that is not on the board survived as %q", got)
	}
	if n := len(p.Beats[3].Sketch); n != 2 {
		t.Fatalf("beat 3 has %d items, want the duplicate dropped", n)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("a normalized board should be valid, got %v", err)
	}
}

func TestNormalizeQuizPlan(t *testing.T) {
	p := &SnippetPlan{
		Template: "quiz",
		Title:    "What does len() count?",
		Quiz: &QuizSpec{
			Question: "What does len() return for [[1, 2], [3, 4, 5]]?",
			// "5" twice: the duplicate goes, and the answer travels with the
			// option it was pointing at rather than staying on index 2.
			Options: []string{"2", "5", "5", "7"},
			Answer:  3,
			Why:     []string{"counts the top level", "counts every number", "the same again", "adds the counts"},
		},
		Beats: []SnippetBeat{
			{ID: "a", Heading: "A quick check", Narration: strings.Repeat("word ", 25), Quiz: &QuizBeat{Show: "ask"}},
			// A second ask, and an explanation before the answer is out.
			{ID: "b", Heading: "Think", Narration: strings.Repeat("word ", 25), Quiz: &QuizBeat{Show: "ask"}},
			{ID: "c", Heading: "Still thinking", Narration: strings.Repeat("word ", 25), Quiz: &QuizBeat{Show: "explain", Option: 1}},
			{ID: "d", Heading: "The answer", Narration: strings.Repeat("word ", 25), Quiz: &QuizBeat{Show: "reveal"}},
			// An option that does not exist, once the duplicate is gone.
			{ID: "e", Heading: "Why five", Narration: strings.Repeat("word ", 25), Quiz: &QuizBeat{Show: "explain", Option: 9}},
		},
	}
	normalizeSnippetPlan(p)

	if got := p.Quiz.Options; len(got) != 3 {
		t.Fatalf("options %v, want the duplicate dropped", got)
	}
	if got := p.Quiz.Options[p.Quiz.Answer]; got != "7" {
		t.Errorf("the answer moved to %q; it should still be the option it named", got)
	}
	if len(p.Quiz.Why) != len(p.Quiz.Options) {
		t.Errorf("%d explanations for %d options", len(p.Quiz.Why), len(p.Quiz.Options))
	}
	if got := p.Beats[1].Quiz.Show; got != "think" {
		t.Errorf("the second ask became %q, want think", got)
	}
	if got := p.Beats[2].Quiz.Show; got != "think" {
		t.Errorf("an explanation before the reveal became %q, want think", got)
	}
	if got := p.Beats[4].Quiz.Option; got < 0 || got >= len(p.Quiz.Options) {
		t.Errorf("explain beat still points at option %d of %d", got, len(p.Quiz.Options))
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("a normalized quiz should be valid, got %v", err)
	}
}

// Ids are how a flow diagram refers to itself, and a model that writes "API
// Gateway" in one place and "api-gateway" in another has drawn the right
// diagram with the wrong strings in it.
func TestNormalizeFlowPlanResolvesReferences(t *testing.T) {
	p := &SnippetPlan{
		Template: "flow",
		Title:    "A rate-limited gateway",
		Beats: []SnippetBeat{
			{ID: "one", Heading: "The client", Narration: strings.Repeat("word ", 25), Nodes: []FlowNode{
				{ID: "Client App", Label: "Client", Kind: "client"},
			}},
			{ID: "two", Heading: "The gateway", Narration: strings.Repeat("word ", 25), Nodes: []FlowNode{
				{ID: "gateway", Label: "Gateway", Kind: "service", From: []string{"Client App"}},
				{ID: "limiter", Label: "Rate limiter", Kind: "cache", From: []string{"Client App"}},
				{ID: "ghost", Label: "Orphan", Kind: "service", From: []string{"nothing-declared"}},
			},
				Focus: []string{"Client App", "not-a-node"}},
		},
	}
	normalizeSnippetPlan(p)

	if got := p.Beats[0].Nodes[0].ID; got != "client-app" {
		t.Errorf("id %q was not slugged", got)
	}
	if got := p.Beats[1].Nodes[0].From; len(got) != 1 || got[0] != "client-app" {
		t.Errorf("edge %v was not re-pointed at the slugged id", got)
	}
	if got := p.Beats[1].Nodes[2].From; len(got) != 0 {
		t.Errorf("edge from an undeclared node survived as %v", got)
	}
	if got := p.Beats[1].Focus; len(got) != 1 || got[0] != "client-app" {
		t.Errorf("focus %v, want only the node that exists", got)
	}
}

// The last line of defence: a plan the rules rejected still ships if the
// template can actually lay it out.
func TestSalvageShipsAPlanThatRenders(t *testing.T) {
	cfg := config.Defaults()
	spec := SnippetSpec{Prompt: "why caching helps", Template: "whiteboard"}

	// Three boxes where the template asks for four. Editorial, not structural:
	// the board draws, so the creator gets a clip.
	short := whiteboardPlan()
	short.Beats[3].Sketch = nil
	if err := short.Validate(); err == nil {
		t.Fatal("expected the three-item board to be rejected by validation")
	}
	if err := dryRunSnippetScenes(spec, cfg, short); err != nil {
		t.Fatalf("a three-item board should still lay out, got %v", err)
	}

	// Nothing on the board at all is structural: there is no clip to ship.
	empty := whiteboardPlan()
	for i := range empty.Beats {
		empty.Beats[i].Sketch = nil
	}
	if err := dryRunSnippetScenes(spec, cfg, empty); err == nil {
		t.Fatal("a board with nothing drawn on it should not lay out")
	}
}

func TestNormalizeSnippetBeatsNamesEveryBeat(t *testing.T) {
	p := &SnippetPlan{
		Template: "whiteboard",
		Beats: []SnippetBeat{
			{Narration: "the first beat says something about caching"},
			{ID: "the-first-beat", Heading: "Say", Narration: "and so does the second"},
			{ID: "the-first-beat", Heading: "Again", Narration: "and the third"},
			{ID: "silent", Heading: "Nothing", Narration: "   "},
		},
	}
	normalizeSnippetPlan(p)

	if n := len(p.Beats); n != 3 {
		t.Fatalf("%d beats, want the silent one dropped", n)
	}
	seen := map[string]bool{}
	for _, b := range p.Beats {
		if b.ID == "" || seen[b.ID] {
			t.Errorf("beat id %q is empty or repeated", b.ID)
		}
		if b.Heading == "" {
			t.Errorf("beat %q has no heading", b.ID)
		}
		seen[b.ID] = true
	}
	if p.Title == "" {
		t.Error("a plan with beats should end up with a title")
	}
}
