package pipeline

// The roster template: who is in the room, as cards rather than as caveats.
//
// A course's first minute has to answer "is this for me", and the catalog only
// had one shape for it. `prereq` answers it as a floor — a list of capabilities,
// each tagged with where it came from and whether it can be skipped — which is
// the right instrument for lesson six of nine, where the viewer already knows
// what the course is and needs to know whether they are behind. It is the wrong
// instrument for lesson one, and the difference is not styling. A floor
// addresses one viewer and tells them what they are missing; a roster addresses
// several and tells each of them that they are the one it was built for. The
// first is a checklist and it reads as an entry exam. The second is a row of
// cards, and every card is somebody recognising themselves.
//
// So this template has no notion of a prerequisite, and refuses to grow one.
// Each card is a KIND OF PERSON, and it carries exactly two lines: what they
// arrive with, and what changes for them. Both are required, and the pairing is
// the whole template — a card with only the first is a demographic, a card with
// only the second is a promise nobody is standing in front of.
//
// Three decisions earn the shape.
//
// The row is up from the first frame and nothing is hidden. Every other card
// template in the catalog reveals on the beat, because a voice is walking the
// row and a card that has already been read steals the line that was about to
// introduce it. This one is built to also work with no voice at all — a
// title-card sequence a viewer reads — so an unlit card is dimmed in its
// CONTENTS and never in its surface. On paper, opacity on a card fades it into
// the page and the row loses the count it exists to show. See the showroom skin.
//
// Two or three cards, never four. The cards carry two labelled lines each, so
// the fourth column takes the type below the size at which a line of ten words
// is read at a glance rather than read at all — and this frame's whole job is
// being understood before the viewer has decided to keep watching.
//
// A card cannot be a rival of another. `versus` is a shape for products, and it
// is exactly wrong here: the two people on this frame are not alternatives being
// weighed, they are two audiences being told the same thing. The role vocabulary
// is available for tinting and the validator refuses "rival" outright, because
// the one thing this frame must not say is that one of these viewers is the
// wrong kind.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "roster",
		Category: CatPresenting,
		Since:    SinceV11,
		// Showroom rather than core: it is cut on paper, with the seated cards
		// and the contents-not-surface dimming that only work on a light ground.
		Family:      FamilyShowroom,
		Title:       "Who this is for",
		Description: "Two or three cards, one per kind of viewer, each saying what that person arrives with and what changes for them. Reach for it in a course's first minute, when the question is whether this was built for me.",
		Example:     "Who this course is for: people who have never written a line of code, and people who write it every day",
		PromptFile:  snippetRosterTemplateName,
		NeedsCode:   false,
		// The row, a beat per card, and the closing frame: four beats at two
		// cards, five at three. Twenty seconds funds four.
		MinTargetSec:     20,
		DefaultTargetSec: 34,
		// Five beats of sixty words is as much narration as the shape holds; past
		// it a card is held on screen for half a minute with nothing moving.
		MaxTargetSec: 120,
		MaxBeats:     5,
		// A beat is a SHOT — one card raised with its two lines read — not a step
		// in an argument. Twenty-four words is eight seconds on one card, which is
		// as long as a card with eighteen words on it can hold a frame.
		IdealWordsPerBeat: 24,
		Owns:              beatFields{Roster: true},
		OwnsPlan:          planFields{Roster: true},
		Normalize:         normalizeRosterPlan,
		Validate:          validateRosterPlan,
		Scenes:            rosterScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":          strings.Join(RosterShows(), ", "),
				"Icons":          strings.Join(PointIconNames(), ", "),
				"Roles":          strings.Join(RosterRoles(), ", "),
				"MinPeople":      minRosterPeople,
				"MaxPeople":      maxRosterPeople,
				"MaxWhoWords":    maxRosterWhoWords,
				"MaxBringsWords": maxRosterBringsWords,
				"MaxGetsWords":   maxRosterGetsWords,
				"MaxCloserWords": maxRosterCloserWords,
			}
		},
	})
}

const snippetRosterTemplateName = "snippet_roster.tmpl"

const (
	// One card is a title card with a subtitle. Four columns of two labelled
	// lines each drops the type below reading size — see the file header.
	minRosterPeople = 2
	maxRosterPeople = 3

	// The card's heading, set in display type: "New to code", "Already ship
	// code", "Coming from design".
	maxRosterWhoWords = 4
	// What they arrive with. One clause.
	maxRosterBringsWords = 10
	// What changes for them. The line the card exists for, so it gets the most
	// room of the three.
	maxRosterGetsWords = 12
	// The line under the finished row.
	maxRosterCloserWords = 14
)

// rosterShows is the closed vocabulary of what a beat does.
var rosterShows = map[string]bool{
	// The whole row up, every card even. The opener.
	"row": true,
	// One card raised, its two lines at full ink, the others muted.
	"person": true,
	// Every card level again and the closer landing. The closer.
	"all": true,
}

// RosterShows returns the beat vocabulary sorted.
func RosterShows() []string {
	out := make([]string, 0, len(rosterShows))
	for k := range rosterShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// rosterRoles is the tint vocabulary: the metric roles minus "rival".
//
// Refused rather than merely unused. "rival" is what a card wears when it is the
// alternative you are being talked out of, and the one claim this frame must
// never make is that one of the people on it came to the wrong course.
var rosterRoles = map[string]bool{
	"neutral":  true,
	"quantity": true,
	"limit":    true,
}

// RosterRoles returns the tint vocabulary sorted.
func RosterRoles() []string {
	out := make([]string, 0, len(rosterRoles))
	for k := range rosterRoles {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// RosterSpec is the row of people. On the plan rather than on a beat because
// every card is on screen for the whole clip — the beats only raise one of them.
type RosterSpec struct {
	// People are the cards, left to right.
	People []RosterPerson `json:"people"`
	// Closer is the line under the finished row — the thing that is true of
	// everyone on it. Optional, and worth having: it is what the last beat has
	// to land on other than a card lighting up again.
	Closer string `json:"closer,omitempty"`
}

// RosterPerson is one card: a kind of viewer, what they bring, what they get.
type RosterPerson struct {
	// Who they are, as they would describe themselves.
	Who string `json:"who"`
	// Brings is what they arrive with. Required.
	Brings string `json:"brings"`
	// Gets is what changes for them by the end. Required, and it is the line
	// the card exists for.
	Gets string `json:"gets"`
	// Icon is a PointIconNames name, drawn in the card's tile.
	Icon string `json:"icon,omitempty"`
	// Role tints the card: a rosterRoles name.
	Role string `json:"role,omitempty"`
}

// ResolvedIcon returns the card's glyph, never empty — a card with an empty
// tile is a hole in the row rather than a plain card.
func (p RosterPerson) ResolvedIcon() string {
	if icon := normalizePointIconName(p.Icon); icon != "" {
		return icon
	}
	return "users"
}

// ResolvedRole returns the card's tint, defaulting to neutral.
func (p RosterPerson) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(p.Role))
	if rosterRoles[r] {
		return r
	}
	return "neutral"
}

// RosterBeat is one move.
type RosterBeat struct {
	Show string `json:"show"`
	At   int    `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to raising a
// card — which is what most beats of this template do.
func (b RosterBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if rosterShows[s] {
		return s
	}
	return "person"
}

func normalizeRosterPlan(p *SnippetPlan) {
	r := p.Roster
	if r == nil {
		return
	}
	r.Closer = clampWords(collapseSpaces(r.Closer), maxRosterCloserWords)

	people := make([]RosterPerson, 0, len(r.People))
	for _, person := range r.People {
		person.Who = clampWords(collapseSpaces(person.Who), maxRosterWhoWords)
		person.Brings = clampWords(collapseSpaces(person.Brings), maxRosterBringsWords)
		person.Gets = clampWords(collapseSpaces(person.Gets), maxRosterGetsWords)
		person.Icon = person.ResolvedIcon()
		person.Role = person.ResolvedRole()
		if person.Who != "" && len(people) < maxRosterPeople {
			people = append(people, person)
		}
	}
	r.People = people

	for i := range p.Beats {
		b := p.Beats[i].Roster
		if b == nil {
			continue
		}
		if !rosterShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			switch {
			case i == 0:
				b.Show = "row"
			case i == len(p.Beats)-1:
				b.Show = "all"
			default:
				b.Show = "person"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "person" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(r.People); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateRosterPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Roster: true}); err != nil {
		return err
	}

	r := p.Roster
	if r == nil {
		return fmt.Errorf("the plan has no roster — this template is a row of people, each with what they bring and what they get")
	}
	if n := len(r.People); n < minRosterPeople || n > maxRosterPeople {
		return fmt.Errorf("there are %d people, want %d-%d. One card is a title card with a subtitle, and a fourth column takes two labelled lines below reading size",
			n, minRosterPeople, maxRosterPeople)
	}

	seen := map[string]bool{}
	for i, person := range r.People {
		if strings.TrimSpace(person.Who) == "" {
			return fmt.Errorf("card %d has nobody on it", i)
		}
		if strings.TrimSpace(person.Brings) == "" {
			return fmt.Errorf("card %q does not say what they arrive with. Without it the card is a demographic", person.Who)
		}
		if strings.TrimSpace(person.Gets) == "" {
			return fmt.Errorf("card %q does not say what changes for them. That line is the reason the card is on the frame", person.Who)
		}
		if strings.EqualFold(strings.TrimSpace(person.Brings), strings.TrimSpace(person.Gets)) {
			return fmt.Errorf("card %q brings and gets the same thing — one of the two lines is not saying anything", person.Who)
		}
		key := strings.ToLower(strings.TrimSpace(person.Who))
		if seen[key] {
			return fmt.Errorf("two cards are both %q — the row is kinds of viewer, so a repeat is one card padding the count", person.Who)
		}
		seen[key] = true
		// Checked here as well as normalized, because normalization silently
		// rewrites "rival" to neutral and the rewrite hides the mistake worth
		// naming: a plan that cast one of these people as the alternative.
		if strings.EqualFold(strings.TrimSpace(person.Role), "rival") {
			return fmt.Errorf("card %q is cast as a rival. Everyone on this frame is being told they belong, so nobody on it is the alternative", person.Who)
		}
	}

	covered := map[int]bool{}
	counts := map[string]int{}
	last := -1
	for i, b := range p.Beats {
		if b.Roster == nil {
			return fmt.Errorf("beat %q has no roster direction — every beat raises the row, one card, or brings them all back", b.ID)
		}
		show := b.Roster.ResolvedShow()
		counts[show]++
		if i == 0 && show != "row" {
			return fmt.Errorf("the clip opens on %q. Raise the whole row first — the viewer has to see how many kinds of person are on the frame before one of them is read out", show)
		}
		if show == "all" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q levels the row again but the clip carries on afterwards. That frame is the close", b.ID)
		}
		if show != "person" {
			continue
		}
		if b.Roster.At < 0 || b.Roster.At >= len(r.People) {
			return fmt.Errorf("beat %q raises card %d, which does not exist", b.ID, b.Roster.At)
		}
		if covered[b.Roster.At] {
			return fmt.Errorf("beat %q raises card %d again; each card gets one beat", b.ID, b.Roster.At)
		}
		if b.Roster.At < last {
			return fmt.Errorf("beat %q goes back to card %d after card %d. The row is read left to right, and a light that jumps around teaches the viewer the order means nothing",
				b.ID, b.Roster.At, last)
		}
		last = b.Roster.At
		covered[b.Roster.At] = true
	}
	if counts["row"] != 1 {
		return fmt.Errorf("there are %d opening beats; the row goes up once", counts["row"])
	}
	if len(covered) != len(r.People) {
		return fmt.Errorf("%d of the %d cards are never spoken about. A card nobody accounts for is somebody on the frame being ignored",
			len(r.People)-len(covered), len(r.People))
	}
	if counts["all"] != 1 {
		return fmt.Errorf("there are %d closing beats; the row levels once, at the end, and that is the frame the viewer screenshots", counts["all"])
	}
	return nil
}

// rosterScenes lays the clip out as ONE scene: the row is on screen throughout
// and the beats only move which card is raised.
func rosterScenes(in SnippetSceneInput) ([]Scene, error) {
	r := in.Plan.Roster
	if r == nil {
		return nil, fmt.Errorf("the plan has no roster")
	}

	people := make([]map[string]any, len(r.People))
	for i, person := range r.People {
		people[i] = map[string]any{
			"who":    person.Who,
			"brings": person.Brings,
			"gets":   person.Gets,
			"icon":   person.ResolvedIcon(),
			"role":   person.ResolvedRole(),
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Roster == nil {
			return nil, fmt.Errorf("beat %q has no roster direction", beat.ID)
		}
		show := beat.Roster.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "person" {
			step["at"] = beat.Roster.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneRoster,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"people": people,
			"closer": r.Closer,
			"steps":  steps,
		}),
	}}, nil
}
