package pipeline

// The mockup template: a screen assembling itself, block by block.
//
// Every no-code builder — Bubble, Webflow, Framer, Softr, Lovable, v0 — makes
// the same kind of thing, and it is a *screen*. A lesson about any of them has
// to put that screen on camera, and until now the catalog could not: the board
// and the diagram draw relationships between boxes, `anatomy` takes apart a line
// of text, `illustration` sets a headline beside a figure. None of them can
// render a page.
//
// The alternative was a screen recording of the tool, and it is the wrong trade
// twice over. Those tools redraw their interface every quarter, so the recording
// ages in weeks; and a recording of somebody dragging a component shows the
// dragging, which is the part nobody needs taught. A synthesized wireframe shows
// the *result* of each decision, stays on the design system, and is legible at
// 1080p in a way a screen capture of a 1440p editor never is.
//
// The layers panel down the right is not decoration either. It is the one piece
// of chrome every builder in that list shares, and it does the job the note
// under a timeline does: it says where you are in something with a shape, so a
// viewer three blocks in still knows how much screen is left to build.
//
// The blocks are declared in page order and walked forward, because that is the
// order a page is read and the order a builder stacks it. A clip that jumps back
// up the page is describing a layout, not building one, and the whiteboard draws
// layouts better than a device frame does.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:             "mockup",
		Category:         CatCode,
		Title:            "Screen mockup",
		Description:      "A page assembling itself inside a device frame, with the layer list filling in beside it.",
		Example:          "Building a signup page in Webflow without writing any code",
		PromptFile:       snippetMockupTemplateName,
		NeedsCode:        false,
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		Owns:             beatFields{Mockup: true},
		OwnsPlan:         planFields{Mockup: true},
		Normalize:        normalizeMockupPlan,
		Validate:         validateMockupPlan,
		Scenes:           mockupScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Kinds":         strings.Join(MockupBlockKinds(), ", "),
				"Devices":       strings.Join(MockupDevices(), ", "),
				"MinBlocks":     minMockupBlocks,
				"MaxBlocks":     maxMockupBlocks,
				"MaxLabelWords": maxMockupLabelWords,
				"MaxTextWords":  maxMockupTextWords,
				"MaxNoteWords":  maxMockupNoteWords,
			}
		},
	})
}

const snippetMockupTemplateName = "snippet_mockup.tmpl"

// Page capacity. Five blocks is what a 520px device viewport holds before the
// wireframe stops being a page and becomes a stack of grey bars; three is the
// floor because a screen with two things on it is a component, and the
// illustration template draws a single component better than a browser frame
// wrapped around one.
const (
	minMockupBlocks = 3
	maxMockupBlocks = 5

	maxMockupLabelWords = 4
	maxMockupTextWords  = 5
	maxMockupNoteWords  = 16
)

// mockupBlockKinds is the closed vocabulary of things a page is made of. Each
// one is a distinct wireframe drawing in the renderer, which is what stops this
// degenerating into rectangles with captions on them.
//
// The set is deliberately small. A vocabulary with thirty entries would be one
// the model picks badly from and one whose drawings nobody maintains; these nine
// cover what a landing page, a form page and a list page are actually built out
// of, which is what the courses this template was written for are about.
var mockupBlockKinds = map[string]bool{
	"header": true,
	"hero":   true,
	"text":   true,
	"image":  true,
	"grid":   true,
	"button": true,
	"input":  true,
	"list":   true,
	"footer": true,
}

// MockupBlockKinds returns the vocabulary sorted, for prompts and docs.
func MockupBlockKinds() []string {
	out := make([]string, 0, len(mockupBlockKinds))
	for k := range mockupBlockKinds {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// mockupDevices is what the page is drawn inside.
var mockupDevices = map[string]bool{"browser": true, "phone": true}

// MockupDevices returns the device vocabulary sorted.
func MockupDevices() []string {
	out := make([]string, 0, len(mockupDevices))
	for k := range mockupDevices {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// MockupSpec is the screen being built. On the plan rather than per-beat for the
// reason the canvas's automation is: the page is the subject of the clip and the
// beats only walk down it.
type MockupSpec struct {
	// Device is what the page is drawn inside: a mockupDevices name.
	Device string `json:"device,omitempty"`
	// Screen names the page — it sits in the browser's address pill, or under
	// the phone's status bar.
	Screen string `json:"screen,omitempty"`
	// Blocks are the page's sections, in the order they appear DOWN the page.
	Blocks []MockupBlock `json:"blocks"`
}

// MockupBlock is one section of the page.
type MockupBlock struct {
	// Kind is a mockupBlockKinds name; anything else degrades to "text".
	Kind string `json:"kind"`
	// Label is what this section is, for the layer list beside the frame.
	Label string `json:"label"`
	// Text is rendered inside the block itself — a hero's headline, a button's
	// caption, an input's placeholder. Optional; the wireframe draws grey rules
	// where there is none, which is what a wireframe is supposed to look like.
	Text string `json:"text,omitempty"`
	// Note expands on the block, and is shown only while it is current.
	Note string `json:"note,omitempty"`
}

// ResolvedKind returns the block's kind, defaulting the unknown to a text block.
func (b MockupBlock) ResolvedKind() string {
	k := strings.ToLower(strings.TrimSpace(b.Kind))
	if mockupBlockKinds[k] {
		return k
	}
	return "text"
}

// MockupBeat says which block this beat is adding, or that it is showing the
// finished screen.
type MockupBeat struct {
	// At indexes MockupSpec.Blocks.
	At int `json:"at"`
	// Whole marks the closing beat that shows the assembled page with nothing
	// singled out. A flag rather than an out-of-range index for the same reason
	// the timeline's is: `at` omitted decodes to 0, which is a real block.
	Whole bool `json:"whole,omitempty"`
}

func normalizeMockupPlan(p *SnippetPlan) {
	m := p.Mockup
	if m == nil {
		return
	}
	device := strings.ToLower(strings.TrimSpace(m.Device))
	if !mockupDevices[device] {
		device = "browser"
	}
	m.Device = device
	m.Screen = clampWords(collapseSpaces(m.Screen), maxMockupLabelWords)
	for i := range m.Blocks {
		b := &m.Blocks[i]
		b.Kind = b.ResolvedKind()
		b.Label = clampWords(collapseSpaces(b.Label), maxMockupLabelWords)
		b.Text = clampWords(collapseSpaces(b.Text), maxMockupTextWords)
		b.Note = clampWords(collapseSpaces(b.Note), maxMockupNoteWords)
		if b.Label == "" {
			// The layer list is the one place a block must be nameable. Its kind
			// is a worse name than the model would have written and a better one
			// than a blank row.
			b.Label = strings.ToUpper(b.Kind[:1]) + b.Kind[1:]
		}
	}
	for i := range p.Beats {
		if b := p.Beats[i].Mockup; b != nil && b.At < 0 {
			b.At, b.Whole = 0, true
		}
	}
}

func validateMockupPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Mockup: true}); err != nil {
		return err
	}

	m := p.Mockup
	if m == nil {
		return fmt.Errorf("the plan has no mockup — this template is one screen, assembled block by block")
	}
	if n := len(m.Blocks); n < minMockupBlocks || n > maxMockupBlocks {
		return fmt.Errorf("the page has %d blocks, want %d-%d — a screen with two things on it is a component, and the illustration template draws one of those better than a browser frame wrapped around it",
			n, minMockupBlocks, maxMockupBlocks)
	}
	// A page has one top and one bottom. Two of either is not a layout the
	// device frame can draw honestly, and it is usually a model listing
	// sections rather than stacking them.
	counts := map[string]int{}
	for _, b := range m.Blocks {
		counts[b.ResolvedKind()]++
	}
	for _, once := range []string{"header", "footer"} {
		if counts[once] > 1 {
			return fmt.Errorf("the page has %d %s blocks. A page has one — if these are different sections, say what they are instead", counts[once], once)
		}
	}
	seen := map[string]bool{}
	for i, b := range m.Blocks {
		if strings.TrimSpace(b.Label) == "" {
			return fmt.Errorf("block %d has no label — the layer list beside the frame needs a name for it", i)
		}
		key := strings.ToLower(strings.TrimSpace(b.Label))
		if seen[key] {
			return fmt.Errorf("block %d repeats the label %q — each layer is a different part of the page", i, b.Label)
		}
		seen[key] = true
	}

	visited := map[int]bool{}
	last := -1
	sawWhole := false
	for _, beat := range p.Beats {
		if beat.Mockup == nil {
			return fmt.Errorf("beat %q has no mockup direction — every beat is adding a block or showing the finished screen", beat.ID)
		}
		if beat.Mockup.Whole {
			sawWhole = true
			continue
		}
		if beat.Mockup.At < 0 || beat.Mockup.At >= len(m.Blocks) {
			return fmt.Errorf("beat %q adds block %d, which does not exist", beat.ID, beat.Mockup.At)
		}
		// A page is read and built top to bottom.
		if beat.Mockup.At < last {
			return fmt.Errorf("beat %q goes back up to block %d after %d. The page is built downward — a clip that jumps back up is describing a layout rather than building one, and the whiteboard draws layouts better than a device frame does",
				beat.ID, beat.Mockup.At, last)
		}
		if visited[beat.Mockup.At] {
			return fmt.Errorf("beat %q adds block %d again; each block gets one beat", beat.ID, beat.Mockup.At)
		}
		visited[beat.Mockup.At] = true
		last = beat.Mockup.At
	}
	if len(visited) != len(m.Blocks) {
		return fmt.Errorf("%d of the %d blocks are never narrated — a block nobody explains is a grey rectangle",
			len(m.Blocks)-len(visited), len(m.Blocks))
	}
	if !sawWhole {
		return fmt.Errorf("no beat shows the finished screen. Close with a beat carrying \"whole\": true — the assembled page is what the viewer came to see")
	}
	if lastBeat := p.Beats[len(p.Beats)-1].Mockup; lastBeat != nil && !lastBeat.Whole {
		return fmt.Errorf("the clip ends mid-build; end on the finished screen instead (\"whole\": true)")
	}
	return nil
}

// mockupScenes lays the clip out as ONE scene: the device frame is on screen
// throughout and the beats only move which block has landed.
func mockupScenes(in SnippetSceneInput) ([]Scene, error) {
	m := in.Plan.Mockup
	if m == nil {
		return nil, fmt.Errorf("the plan has no mockup")
	}

	blocks := make([]map[string]any, len(m.Blocks))
	for i, b := range m.Blocks {
		blocks[i] = map[string]any{
			"kind":  b.ResolvedKind(),
			"label": b.Label,
			"text":  b.Text,
			"note":  b.Note,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Mockup == nil {
			return nil, fmt.Errorf("beat %q has no mockup direction", beat.ID)
		}
		step := map[string]any{"startMs": startMs, "endMs": endMs}
		if beat.Mockup.Whole {
			step["whole"] = true
		} else {
			step["at"] = beat.Mockup.At
		}
		steps = append(steps, step)
	}

	device := strings.ToLower(strings.TrimSpace(m.Device))
	if !mockupDevices[device] {
		device = "browser"
	}
	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneMockup,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":  in.Plan.Title,
			"device": device,
			"screen": m.Screen,
			"blocks": blocks,
			"steps":  steps,
		},
	}}, nil
}
