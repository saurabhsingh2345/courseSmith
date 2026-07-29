package pipeline

// The canvas template: an automation wired together on a builder's canvas, and
// then run.
//
// The thirteenth template, and the first one written for a course that is not
// about code. Every no-code tool a lesson could be about — n8n, Make, Zapier,
// Airtable automations, Power Automate — is the same picture: app cards on a
// dotted grid, wired left to right, with a record travelling the wire. Owning
// that picture means a whole curriculum can be taught without recording a
// single tool, which matters more here than anywhere else in the catalog: the
// tools change their UI every quarter and a screen recording ages in weeks.
//
// The flow template draws boxes and edges too, and the difference is the reason
// both exist. A flow diagram is a *topology* — it must fork or join, its traffic
// is ambient, and its claim is "these things are connected like this". A canvas
// is a *procedure*: it is a chain by construction, it runs once, and its claim
// is "this fires, then this happens, and here is one real record going through
// it end to end". Rendering a chain through the flow template gives you columns
// of one node each with most of the frame empty; rendering a topology here
// flattens it into a queue it never was.
//
// So the shape rules are the procedure's rules. Something has to start it, which
// is why the first node is a trigger and there is only one. The clip only moves
// forward, because an automation read backwards is not an automation. And the
// last beat runs the payload, because a workflow nobody watched fire is a
// diagram of a workflow.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "canvas",
		Category:    CatCode,
		Title:       "Automation canvas",
		Description: "App cards wired on a builder's canvas with a real record running the chain. Reach for it for no-code automations and integrations.",
		Example:     "How a form submission ends up in a spreadsheet and a Slack message",
		PromptFile:  snippetCanvasTemplateName,
		NeedsCode:   false,
		// One beat per step plus the run is four at the floor, and a step that
		// gets less than a few seconds is a card nobody read.
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		Owns:             beatFields{Canvas: true},
		OwnsPlan:         planFields{Canvas: true},
		Normalize:        normalizeCanvasPlan,
		Validate:         validateCanvasPlan,
		Scenes:           canvasScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Kinds":           strings.Join(CanvasNodeKinds(), ", "),
				"Icons":           strings.Join(PointIconNames(), ", "),
				"MinNodes":        minCanvasNodes,
				"MaxNodes":        maxCanvasNodes,
				"MaxAppWords":     maxCanvasAppWords,
				"MaxTitleWords":   maxCanvasTitleWords,
				"MaxNoteWords":    maxCanvasNoteWords,
				"MaxPayloadWords": maxCanvasPayloadWords,
			}
		},
	})
}

const snippetCanvasTemplateName = "snippet_canvas.tmpl"

// Canvas capacity. Five cards across the 1700px stage leaves 300px each, which
// is what a five-word title needs at two lines and 26px; a sixth turns the row
// into a strip of labels nobody can read at a glance. Three is the floor
// because two cards is a trigger and one action, which is a sentence rather
// than a workflow.
const (
	minCanvasNodes = 3
	maxCanvasNodes = 5

	maxCanvasAppWords     = 2
	maxCanvasTitleWords   = 5
	maxCanvasNoteWords    = 16
	maxCanvasPayloadWords = 3
)

// canvasNodeKinds is the closed vocabulary of card types, mapped to the icon
// each one falls back to when the model does not name a better one.
//
// Kind is not decoration: it drives the chip colour and the badge, and the
// validator reads it to enforce that exactly one card starts the automation.
// The set is deliberately the one every no-code builder converges on, whatever
// it calls them — something happens, something is decided, something is done.
var canvasNodeKinds = map[string]string{
	"trigger": "zap",
	"action":  "gear",
	"filter":  "filter",
	"branch":  "shuffle",
	"output":  "check",
}

// CanvasNodeKinds returns the vocabulary sorted, for prompts and docs.
func CanvasNodeKinds() []string {
	out := make([]string, 0, len(canvasNodeKinds))
	for k := range canvasNodeKinds {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CanvasSpec is the whole automation. On the plan rather than per-beat for the
// reason the timeline's milestones are: the workflow is the subject of the clip
// and the beats only walk it.
type CanvasSpec struct {
	// Payload is what rides the wire when the automation runs — the label on
	// the token, in the terms the viewer would use ("New signup", "One row").
	Payload string `json:"payload"`
	// Nodes are the steps, in the order they fire.
	Nodes []CanvasNode `json:"nodes"`
}

// CanvasNode is one card on the canvas.
//
// It has no id, and that is a choice rather than an omission. A flow node needs
// one because its edges are declared by name and can point anywhere; a canvas is
// a chain, so a card's only relationship is "after the one before it" and
// position says that already. Every id in this template would be a string the
// model has to keep consistent across three places for no gain — which is
// exactly the failure normalizeFlowPlan exists to repair.
type CanvasNode struct {
	// App is the tool this step happens in — "Gmail", "Airtable", "Slack". It
	// sits in the card's chip header, small, above the title.
	App string `json:"app,omitempty"`
	// Title is what the step does.
	Title string `json:"title"`
	// Kind is a canvasNodeKinds name; anything else degrades to "action".
	Kind string `json:"kind"`
	// Icon is a PointIconNames name drawn in the chip; empty takes the kind's.
	Icon string `json:"icon,omitempty"`
	// Note expands on the step, and is shown only while this card is current.
	Note string `json:"note,omitempty"`
}

// ResolvedKind returns the node's kind, defaulting the unknown to an action.
func (n CanvasNode) ResolvedKind() string {
	k := strings.ToLower(strings.TrimSpace(n.Kind))
	if _, ok := canvasNodeKinds[k]; ok {
		return k
	}
	return "action"
}

// ResolvedIcon returns the icon drawn in the card's chip.
func (n CanvasNode) ResolvedIcon() string {
	if icon := normalizePointIconName(n.Icon); icon != "" {
		return icon
	}
	return canvasNodeKinds[n.ResolvedKind()]
}

// CanvasBeat says which card this beat is standing on, or that it is running
// the automation.
type CanvasBeat struct {
	// At indexes CanvasSpec.Nodes.
	At int `json:"at"`
	// Run marks the closing beat that sends the payload down the whole chain.
	// A separate flag rather than an out-of-range index for the same reason the
	// timeline's Whole is one: `at` omitted decodes to 0, which is a real card.
	Run bool `json:"run,omitempty"`
}

func normalizeCanvasPlan(p *SnippetPlan) {
	c := p.Canvas
	if c == nil {
		return
	}
	c.Payload = clampWords(collapseSpaces(c.Payload), maxCanvasPayloadWords)
	if c.Payload == "" {
		// The token has to say something — it is the whole reason the run beat
		// reads as a record moving rather than as a dot sliding.
		c.Payload = "New record"
	}
	for i := range c.Nodes {
		n := &c.Nodes[i]
		n.App = clampWords(collapseSpaces(n.App), maxCanvasAppWords)
		n.Title = clampWords(collapseSpaces(n.Title), maxCanvasTitleWords)
		n.Note = clampWords(collapseSpaces(n.Note), maxCanvasNoteWords)
		n.Kind = n.ResolvedKind()
		n.Icon = n.ResolvedIcon()
	}
	// The first card starts the automation. A model that describes a trigger
	// and files it as an action has understood the workflow and mislabelled the
	// card, which is a repair; a model that puts a *second* trigger in the
	// middle has said something about the workflow, and that is validation's.
	if len(c.Nodes) > 0 && c.Nodes[0].Kind != "trigger" {
		triggers := 0
		for _, n := range c.Nodes {
			if n.Kind == "trigger" {
				triggers++
			}
		}
		if triggers == 0 {
			c.Nodes[0].Kind = "trigger"
			if c.Nodes[0].Icon == canvasNodeKinds["action"] {
				c.Nodes[0].Icon = canvasNodeKinds["trigger"]
			}
		}
	}
	for i := range p.Beats {
		if b := p.Beats[i].Canvas; b != nil && b.At < 0 {
			b.At = 0
		}
	}
}

func validateCanvasPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Canvas: true}); err != nil {
		return err
	}

	c := p.Canvas
	if c == nil {
		return fmt.Errorf("the plan has no canvas — this template is one automation, wired up and then run")
	}
	if n := len(c.Nodes); n < minCanvasNodes || n > maxCanvasNodes {
		return fmt.Errorf("the automation has %d steps, want %d-%d — two cards is a trigger and one action, which is a sentence rather than a workflow",
			n, minCanvasNodes, maxCanvasNodes)
	}
	seen := map[string]bool{}
	for i, n := range c.Nodes {
		if strings.TrimSpace(n.Title) == "" {
			return fmt.Errorf("step %d has no title — say what happens in that card", i)
		}
		key := strings.ToLower(strings.TrimSpace(n.Title))
		if seen[key] {
			return fmt.Errorf("step %d repeats the title %q — each card is a different step", i, n.Title)
		}
		seen[key] = true
	}
	// Something has to start it. This is the one claim an automation makes that
	// a picture of boxes does not, and getting it wrong teaches the viewer that
	// workflows begin wherever you feel like.
	if c.Nodes[0].ResolvedKind() != "trigger" {
		return fmt.Errorf("the first step is a %s — an automation begins with a trigger, the thing that happens and sets everything else off",
			c.Nodes[0].ResolvedKind())
	}
	for i, n := range c.Nodes[1:] {
		if n.ResolvedKind() == "trigger" {
			return fmt.Errorf("step %d is a second trigger. One thing starts an automation; if two separate events start two separate chains, that is two clips",
				i+1)
		}
	}

	visited := map[int]bool{}
	last := -1
	sawRun := false
	for _, b := range p.Beats {
		if b.Canvas == nil {
			return fmt.Errorf("beat %q has no canvas direction — every beat is standing on a card or running the automation", b.ID)
		}
		if b.Canvas.Run {
			sawRun = true
			continue
		}
		if b.Canvas.At < 0 || b.Canvas.At >= len(c.Nodes) {
			return fmt.Errorf("beat %q stands on step %d, which does not exist", b.ID, b.Canvas.At)
		}
		// An automation read backwards is not an automation.
		if b.Canvas.At < last {
			return fmt.Errorf("beat %q goes back to step %d after %d. A canvas only moves forward — if the story genuinely revisits an earlier step, it is a diagram and the flow template will tell it properly",
				b.ID, b.Canvas.At, last)
		}
		if visited[b.Canvas.At] {
			return fmt.Errorf("beat %q stands on step %d again; each card gets one beat", b.ID, b.Canvas.At)
		}
		visited[b.Canvas.At] = true
		last = b.Canvas.At
	}
	if len(visited) != len(c.Nodes) {
		return fmt.Errorf("%d of the %d steps are never narrated — a card nobody explains is a box with a logo on it",
			len(c.Nodes)-len(visited), len(c.Nodes))
	}
	// The payoff. A workflow nobody watched fire is a diagram of a workflow.
	if !sawRun {
		return fmt.Errorf("no beat runs the automation. Close with a beat carrying \"run\": true — watching one real record travel the whole chain is what the viewer came for")
	}
	if lastBeat := p.Beats[len(p.Beats)-1].Canvas; lastBeat != nil && !lastBeat.Run {
		return fmt.Errorf("the clip ends standing on one card; end by running the automation instead (\"run\": true)")
	}
	return nil
}

// canvasScenes lays the clip out as ONE scene: the canvas is on screen for the
// whole clip and the beats only move which card is current.
func canvasScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.Canvas
	if c == nil {
		return nil, fmt.Errorf("the plan has no canvas")
	}

	nodes := make([]map[string]any, len(c.Nodes))
	for i, n := range c.Nodes {
		nodes[i] = map[string]any{
			"app":   n.App,
			"title": n.Title,
			"kind":  n.ResolvedKind(),
			"icon":  n.ResolvedIcon(),
			"note":  n.Note,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Canvas == nil {
			return nil, fmt.Errorf("beat %q has no canvas direction", beat.ID)
		}
		step := map[string]any{"startMs": startMs, "endMs": endMs}
		if beat.Canvas.Run {
			step["run"] = true
		} else {
			step["at"] = beat.Canvas.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneCanvas,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":   in.Plan.Title,
			"payload": c.Payload,
			"nodes":   nodes,
			"steps":   steps,
		},
	}}, nil
}
