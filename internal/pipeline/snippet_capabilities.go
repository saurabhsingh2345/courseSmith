package pipeline

// The capabilities template: what a thing is allowed to touch, and who decided.
//
// The catalog can draw a system's shape (`flow`), a tool's tiers (`stack`) and a
// checklist going green (`spec`). None of them can draw a *boundary* — the line
// around a piece of code with the things it cannot reach on the outside of it,
// and the one or two that were deliberately handed in.
//
// This is the picture underneath WebAssembly's sandbox, containers, iOS
// permissions, OAuth scopes, capability-based security generally, and every
// "why can't my function just read a file" question. It is drawn wrongly almost
// everywhere, in one of two ways: as a wall with nothing crossing it, which
// describes something useless, or as a list of features, which loses the fact
// that they are denied until granted.
//
// Three rules earn it its place, and all three are validators.
//
// The sealed state is established first. "Denied by default" is the whole
// claim, and a frame that opens with capabilities already granted has no default
// to contrast against — the viewer sees permissions rather than a sandbox.
//
// **Something must be granted, and something must stay denied.** This is the
// rule the template exists for, and it is one rule rather than two because
// neither half means anything alone. A boundary with everything denied is a
// wall: correct, useless, and not what anyone ships. A boundary with everything
// granted is the host: also correct, also not a sandbox. What makes the picture
// teach is the GAP — this module asked for four things, was handed one, and runs
// anyway.
//
// A capability is granted at most once. Granting twice says the first grant did
// not take, which is the opposite of what a capability is.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "capabilities",
		Category:    CatSystems,
		Since:       SinceV4,
		Family:      FamilyReplica,
		Title:       "Denied until handed in",
		Description: "A boundary around some code, with what it cannot reach outside it and the one or two things deliberately granted. Reach for it for sandboxes, containers, permissions, OAuth scopes.",
		Example:     "Why a WebAssembly module cannot open a file unless you let it",
		PromptFile:  snippetCapabilitiesTemplateName,
		NeedsCode:   false,
		// The sealed boundary, a grant or two, and what is still shut.
		MinTargetSec:     30,
		DefaultTargetSec: 50,
		MaxBeats:         8,
		Owns:             beatFields{Capabilities: true},
		OwnsPlan:         planFields{Capabilities: true},
		Normalize:        normalizeCapabilitiesPlan,
		Validate:         validateCapabilitiesPlan,
		Scenes:           capabilitiesScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(CapabilityShows(), ", "),
				"MinCaps":       minCapabilities,
				"MaxCaps":       maxCapabilities,
				"MaxLabelWords": maxCapabilityLabelWords,
				"MaxNoteWords":  maxCapabilityNoteWords,
			}
		},
	})
}

const snippetCapabilitiesTemplateName = "snippet_capabilities.tmpl"

const (
	// Below three there is not enough outside the boundary for the gap between
	// asked-for and handed-over to be visible. Past six the ring of chips around
	// the subject is a crowd rather than a boundary.
	minCapabilities = 3
	maxCapabilities = 6

	maxCapabilityLabelWords = 2
	maxCapabilityNoteWords  = 16
)

// capabilityShows is the closed vocabulary of what a beat does.
var capabilityShows = map[string]bool{
	// The subject inside its boundary, everything denied. The first beat.
	"sealed": true,
	// One capability is handed in.
	"grant": true,
	// Hold the picture and say what is still shut.
	"read": true,
}

// CapabilityShows returns the beat vocabulary sorted.
func CapabilityShows() []string {
	out := make([]string, 0, len(capabilityShows))
	for k := range capabilityShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CapabilitySpec is the boundary and what is outside it. On the plan because the
// boundary is one object that persists for the whole clip.
type CapabilitySpec struct {
	// Subject is what is inside the boundary — "WASM module", "the container".
	Subject string `json:"subject"`
	// SubjectNote is the line under it — "app.wasm", "1.2 MB". Optional.
	SubjectNote string `json:"subjectNote,omitempty"`
	// Boundary names the rule the line enforces — "zero default access".
	// Optional; it is the caption on the ring rather than a claim of its own.
	Boundary string `json:"boundary,omitempty"`
	// Granter is who hands capabilities in — "the host", "your manifest".
	// Optional but this template is much weaker without it: a capability that
	// arrives from nowhere is a permission, and the whole point is that
	// something outside chose to give it.
	Granter string `json:"granter,omitempty"`
	// Items are the things the subject might reach, drawn around the boundary.
	Items []Capability `json:"items"`
}

// Capability is one thing the subject might be allowed to touch.
type Capability struct {
	// Label is the capability — "files", "network", "the clock", "sockets".
	Label string `json:"label"`
	// Note says what granting or withholding it means. One sentence.
	Note string `json:"note,omitempty"`
	// Role is what this capability is doing: a metricRoles name.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the capability's role, defaulting to neutral.
func (c Capability) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(c.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// ResolvedGranter names who hands capabilities in.
func (s *CapabilitySpec) ResolvedGranter() string {
	if g := strings.TrimSpace(s.Granter); g != "" {
		return g
	}
	return "the host"
}

// CapabilityBeat is one move: which capability this beat grants.
type CapabilityBeat struct {
	// Show is a capabilityShows name.
	Show string `json:"show"`
	// At indexes CapabilitySpec.Items, for a "grant" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to a grant.
func (b CapabilityBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if capabilityShows[s] {
		return s
	}
	return "grant"
}

func normalizeCapabilitiesPlan(p *SnippetPlan) {
	c := p.Capabilities
	if c == nil {
		return
	}
	c.Subject = clampWords(collapseSpaces(c.Subject), maxCapabilityLabelWords+1)
	c.SubjectNote = clampWords(collapseSpaces(c.SubjectNote), maxCapabilityLabelWords+1)
	c.Boundary = clampWords(collapseSpaces(c.Boundary), maxCapabilityLabelWords+2)
	c.Granter = clampWords(collapseSpaces(c.Granter), maxCapabilityLabelWords+1)

	items := make([]Capability, 0, len(c.Items))
	for _, it := range c.Items {
		it.Label = clampWords(collapseSpaces(it.Label), maxCapabilityLabelWords)
		it.Note = clampWords(collapseSpaces(it.Note), maxCapabilityNoteWords)
		it.Role = it.ResolvedRole()
		if it.Label != "" && len(items) < maxCapabilities {
			items = append(items, it)
		}
	}
	c.Items = items

	for i := range p.Beats {
		b := p.Beats[i].Capabilities
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "grant" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(c.Items); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateCapabilitiesPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Capabilities: true}); err != nil {
		return err
	}

	c := p.Capabilities
	if c == nil {
		return fmt.Errorf("the plan has no boundary — this template is a line around some code and the things it cannot reach across it")
	}
	if strings.TrimSpace(c.Subject) == "" {
		return fmt.Errorf("nothing is named inside the boundary. A ring with an empty middle is a diagram of a rule rather than of a thing obeying one")
	}
	if n := len(c.Items); n < minCapabilities || n > maxCapabilities {
		return fmt.Errorf("there are %d capabilities, want %d-%d. Below three there is not enough outside the boundary for the gap between asked-for and handed-over to be visible; past six the ring is a crowd",
			n, minCapabilities, maxCapabilities)
	}

	seen := map[string]bool{}
	for i, it := range c.Items {
		if strings.TrimSpace(it.Label) == "" {
			return fmt.Errorf("capability %d has no label", i)
		}
		key := strings.ToLower(strings.TrimSpace(it.Label))
		if seen[key] {
			return fmt.Errorf("two capabilities are both %q — each one is a different thing the subject might reach", it.Label)
		}
		seen[key] = true
		if r := strings.ToLower(strings.TrimSpace(it.Role)); r != "" && !metricRoles[r] {
			return fmt.Errorf("capability %d has role %q, which is not one of: %s", i, it.Role, strings.Join(MetricRoles(), ", "))
		}
	}

	// The sealed state comes first: "denied by default" needs a default on
	// screen before anything is handed in.
	if p.Beats[0].Capabilities == nil || p.Beats[0].Capabilities.ResolvedShow() != "sealed" {
		return fmt.Errorf("beat %q does not open on the sealed boundary. Denied-by-default is the whole claim, and a frame that opens with capabilities already granted has no default to contrast against — the viewer sees permissions rather than a sandbox",
			p.Beats[0].ID)
	}

	granted := map[int]bool{}
	seals := 0
	for i, b := range p.Beats {
		if b.Capabilities == nil {
			return fmt.Errorf("beat %q has no capabilities direction — every beat seals the boundary, grants one thing, or reads the result", b.ID)
		}
		switch b.Capabilities.ResolvedShow() {
		case "sealed":
			seals++
			if i != 0 {
				return fmt.Errorf("beat %q seals the boundary again part-way through. A capability once granted is not taken back, and re-sealing would say it was", b.ID)
			}
		case "grant":
			at := b.Capabilities.At
			if at < 0 || at >= len(c.Items) {
				return fmt.Errorf("beat %q grants capability %d, which does not exist", b.ID, at)
			}
			if granted[at] {
				return fmt.Errorf("beat %q grants capability %d again. Granting twice says the first grant did not take, which is the opposite of what a capability is", b.ID, at)
			}
			granted[at] = true
		}
	}
	if seals != 1 {
		return fmt.Errorf("there are %d beats sealing the boundary, want exactly 1", seals)
	}

	// The rule the template exists for, and it is one rule rather than two:
	// neither half means anything alone.
	if len(granted) == 0 {
		return fmt.Errorf("nothing is ever granted, so the picture is a wall rather than a sandbox. What makes a boundary worth drawing is that something crosses it deliberately — grant at least one capability, or the clip has described code that cannot do anything")
	}
	if len(granted) >= len(c.Items) {
		return fmt.Errorf("all %d capabilities end up granted, so there is no boundary left and the subject simply has the host's powers. The picture teaches the GAP — asked for several, handed one, runs anyway — so leave at least one denied",
			len(c.Items))
	}
	return nil
}

// capabilitiesScenes lays the clip out as ONE scene. The boundary persists and
// the beats only say what has been handed in by now.
func capabilitiesScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.Capabilities
	if c == nil {
		return nil, fmt.Errorf("the plan has no boundary")
	}

	items := make([]map[string]any, len(c.Items))
	for i, it := range c.Items {
		items[i] = map[string]any{
			"label": it.Label,
			"note":  it.Note,
			"role":  it.ResolvedRole(),
		}
	}

	granted := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Capabilities == nil {
			return nil, fmt.Errorf("beat %q has no capabilities direction", beat.ID)
		}
		show := beat.Capabilities.ResolvedShow()
		if show == "grant" {
			granted[beat.Capabilities.At] = true
		}
		// The granted set as it stands at this beat, so the renderer draws a
		// whole frame from one step.
		open := make([]int, 0, len(granted))
		for at := range granted {
			open = append(open, at)
		}
		sort.Ints(open)

		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"granted": open,
		}
		if show == "grant" {
			step["at"] = beat.Capabilities.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneCapabilities,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":       in.Plan.Title,
			"subject":     c.Subject,
			"subjectNote": c.SubjectNote,
			"boundary":    c.Boundary,
			"granter":     c.ResolvedGranter(),
			"items":       items,
			"steps":       steps,
		}),
	}}, nil
}
