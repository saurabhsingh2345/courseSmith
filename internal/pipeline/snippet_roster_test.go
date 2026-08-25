package pipeline

import (
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

const rosterNarration = "If you have never written a line of code you are not behind, and an idea is the whole entry fee."

func rosterPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "roster",
		Title:    "Who this is for",
		Roster: &RosterSpec{
			Closer: "Same course, two different first weeks",
			People: []RosterPerson{
				{Who: "New to code", Brings: "an idea and a laptop", Gets: "a working app before you learn a language", Icon: "sprout", Role: "quantity"},
				{Who: "Already write code", Brings: "years of shipping software", Gets: "the same build in an afternoon", Icon: "terminal"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "the-room", Heading: "Two kinds", Narration: rosterNarration, Roster: &RosterBeat{Show: "row"}},
			{ID: "new", Heading: "Never written code", Narration: rosterNarration, Roster: &RosterBeat{Show: "person", At: 0}},
			{ID: "already", Heading: "Already a developer", Narration: rosterNarration, Roster: &RosterBeat{Show: "person", At: 1}},
			{ID: "both", Heading: "Either way", Narration: rosterNarration, Roster: &RosterBeat{Show: "all"}},
		},
	}
	p.targetWords = 4 * 24
	return p
}

func TestRosterPlanAccepted(t *testing.T) {
	p := rosterPlan()
	normalizeRosterPlan(p)
	if err := validateRosterPlan(p); err != nil {
		t.Fatalf("a well-formed roster was rejected: %v", err)
	}
}

// The pairing is the template: a card that says who somebody is and what they
// get, without saying what they arrive with, is a promise with nobody standing
// in front of it.
func TestRosterNeedsBothLines(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		mutate     func(*RosterPerson)
	}{
		{"no brings", "arrive with", func(p *RosterPerson) { p.Brings = "" }},
		{"no gets", "changes for them", func(p *RosterPerson) { p.Gets = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := rosterPlan()
			tc.mutate(&p.Roster.People[0])
			err := validateRosterPlan(p)
			if err == nil {
				t.Fatalf("a card with %s was accepted", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("the error does not say what is missing: %v", err)
			}
		})
	}
}

// Two lines that say the same thing are one line printed twice, and the card
// stops being a before and after.
func TestRosterRejectsARestatement(t *testing.T) {
	p := rosterPlan()
	p.Roster.People[0].Gets = p.Roster.People[0].Brings
	if err := validateRosterPlan(p); err == nil {
		t.Fatal("a card whose two lines are identical was accepted")
	}
}

// The one claim this frame must never make.
func TestRosterRefusesARival(t *testing.T) {
	p := rosterPlan()
	p.Roster.People[1].Role = "rival"
	err := validateRosterPlan(p)
	if err == nil {
		t.Fatal("a viewer cast as the alternative was accepted")
	}
	if !strings.Contains(err.Error(), "rival") {
		t.Errorf("the error does not name the problem: %v", err)
	}
	// And normalization does not let the value through to the frame either, so a
	// plan that skips validation cannot paint one of these people as the loser.
	normalizeRosterPlan(p)
	if got := p.Roster.People[1].ResolvedRole(); got == "rival" {
		t.Errorf("normalize kept the rival role: %q", got)
	}
}

// Every card gets exactly one beat, in reading order.
func TestRosterBeatShape(t *testing.T) {
	for _, tc := range []struct {
		name  string
		beats func(*SnippetPlan)
		want  string
	}{
		{"opens on a card", func(p *SnippetPlan) { p.Beats[0].Roster = &RosterBeat{Show: "person", At: 0} }, "Raise the whole row first"},
		{"card never spoken about", func(p *SnippetPlan) {
			p.Beats = append(p.Beats[:2], p.Beats[3])
		}, "never spoken about"},
		{"card twice", func(p *SnippetPlan) { p.Beats[2].Roster = &RosterBeat{Show: "person", At: 0} }, "again"},
		{"row read backwards", func(p *SnippetPlan) {
			p.Beats[1].Roster = &RosterBeat{Show: "person", At: 1}
			p.Beats[2].Roster = &RosterBeat{Show: "person", At: 0}
		}, "left to right"},
		{"no closing frame", func(p *SnippetPlan) { p.Beats = p.Beats[:3] }, "closing beats"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := rosterPlan()
			tc.beats(p)
			err := validateRosterPlan(p)
			if err == nil {
				t.Fatalf("%s was accepted", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("the error does not explain the shape: %v", err)
			}
		})
	}
}

// One card is a title card with a subtitle, and a fourth column takes the type
// below reading size.
func TestRosterBoundsTheRow(t *testing.T) {
	p := rosterPlan()
	p.Roster.People = p.Roster.People[:1]
	p.Beats = append(p.Beats[:2], p.Beats[3])
	if err := validateRosterPlan(p); err == nil {
		t.Fatal("a row of one was accepted")
	}

	p = rosterPlan()
	p.Roster.People = append(p.Roster.People,
		RosterPerson{Who: "Coming from design", Brings: "a Figma file", Gets: "the file running in a browser"},
		RosterPerson{Who: "Coming from data", Brings: "a spreadsheet", Gets: "the spreadsheet behind a login"})
	normalizeRosterPlan(p)
	if n := len(p.Roster.People); n != maxRosterPeople {
		t.Errorf("normalize kept %d cards, want the row capped at %d", n, maxRosterPeople)
	}
}

// The scene is laid out for real against fabricated timings — the check that
// catches a plan which satisfies every rule and still cannot be drawn.
func TestRosterLaysOut(t *testing.T) {
	p := rosterPlan()
	normalizeRosterPlan(p)
	if err := dryRunSnippetScenes(SnippetSpec{Template: "roster", Prompt: "who this is for", TargetSec: 30}, config.Defaults(), p); err != nil {
		t.Fatalf("a valid roster does not lay out: %v", err)
	}
}
