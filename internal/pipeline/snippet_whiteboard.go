package pipeline

// The whiteboard template.
//
// A hand-drawn board that fills in as the narrator talks: a box is sketched,
// its icon and label appear, an arrow reaches across to the next idea, and by
// the end the whole picture is on screen at once. The board never wipes — that
// accumulation is the point, and it is what separates a whiteboard explainer
// from a slideshow of drawings.
//
// The model does not draw. Letting an LLM author freehand SVG produced
// overlapping text and inconsistent line weight often enough that the visuals
// stage needed a vision-QA gate and a repair loop to cope with it. Here the
// model instead picks *what* goes on the board — a short label and an icon from
// a closed vocabulary, plus which earlier idea it follows from — and the
// renderer draws it, laid out on a grid it owns. That trade gives up nothing
// that reads on screen and gains a composition that cannot come out crooked.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "whiteboard",
		Title:       "Whiteboard sketch",
		Description: "A hand-drawn board that fills in as you talk — boxes, icons and arrows, sketched live.",
		Example:     "Why HTTP caching matters, from browser to CDN to origin",
		PromptFile:  snippetWhiteboardTemplateName,
		NeedsCode:   false,
		Owns:        beatFields{Sketch: true},
		Normalize:   normalizeWhiteboardPlan,
		Validate:    validateWhiteboardPlan,
		Scenes:      whiteboardScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				// The figure vocabulary, not the icon one. A board item is a
				// large object in the middle of a box, which is what the figures
				// were drawn for and what a single-stroke glyph never filled;
				// and where an icon is a static shape, a figure keeps running —
				// the terminal's line types, the database's query sweeps down —
				// so a board that has stopped being drawn is still alive.
				"Icons":         strings.Join(ArtFigureNames(), ", "),
				"Shapes":        strings.Join(SketchShapeNames(), ", "),
				"MinItems":      minSketchItems,
				"MaxItems":      maxSketchItems,
				"MaxLabelWords": maxSketchLabelWords,
			}
		},
	})
}

const snippetWhiteboardTemplateName = "snippet_whiteboard.tmpl"

// Board capacity. Below four items there is no picture to build — three boxes
// leaves the back half of the clip narrating a board that stopped changing.
// Above eight it is a wall of boxes and nothing is legible at 1080p. The
// renderer's layout grid is sized against both ends.
const (
	minSketchItems = 4
	maxSketchItems = 8
)

// maxSketchLabelWords keeps a label a label. Anything longer stops fitting the
// box and starts competing with the narration for the viewer's reading.
const maxSketchLabelWords = 4

// whiteboardScenes lays the clip out as an optional opening title card
// followed by one board that runs to the end.
//
// Items are timed to the beat that introduces them and spread across it, so
// what appears on the board tracks what is being said without ever needing the
// board to reset.
func whiteboardScenes(in SnippetSceneInput) ([]Scene, error) {
	_, clipStart, _ := in.Beat(0)

	var scenes []Scene
	boardStart := clipStart
	// The board wants the same treatment the editor gets: a short card, then
	// the surface it draws on, rather than a card held for a whole beat.
	firstSketch := -1
	for i, b := range in.Plan.Beats {
		if len(b.Sketch) > 0 {
			firstSketch = i
			break
		}
	}
	if firstSketch < 0 {
		return nil, fmt.Errorf("no beat puts anything on the board")
	}
	_, firstSketchStart, _ := in.Beat(firstSketch)
	if firstSketchStart-clipStart >= minTitleCardMs {
		titleEnd := min(firstSketchStart, clipStart+maxTitleCardMs)
		scenes = append(scenes, Scene{
			Type:    SceneTitle,
			StartMs: clipStart,
			EndMs:   titleEnd,
			Props: map[string]any{
				"heading":  in.Plan.Title,
				"subtitle": in.Plan.Subtitle,
				"intro":    true,
			},
		})
		boardStart = titleEnd
	}

	// Item index by label, so a link can name an earlier idea instead of
	// counting positions the model would have to track across beats.
	indexOf := map[string]int{}
	items := make([]map[string]any, 0, maxSketchItems)
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if len(beat.Sketch) == 0 {
			continue
		}
		// Spread this beat's items across the first two thirds of its span, so
		// the last one still has time to be read before the beat ends.
		window := (endMs - startMs) * 2 / 3
		step := window / len(beat.Sketch)
		for j, item := range beat.Sketch {
			props := map[string]any{
				"label": item.Label,
				"icon":  normalizeSketchIcon(item.Icon),
				"shape": normalizeSketchShape(item.Shape),
				"atMs":  startMs + j*step,
			}
			if item.LinkFrom != "" {
				if from, ok := indexOf[sketchKey(item.LinkFrom)]; ok {
					props["from"] = from
				}
			}
			indexOf[sketchKey(item.Label)] = len(items)
			items = append(items, props)
		}
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no board items were produced")
	}

	_, _, boardEnd := in.Beat(len(in.Plan.Beats) - 1)
	scenes = append(scenes, Scene{
		Type:    SceneWhiteboard,
		StartMs: boardStart,
		EndMs:   boardEnd,
		Props: map[string]any{
			"title": in.Plan.Title,
			"items": items,
		},
	})
	return scenes, nil
}

// normalizeWhiteboardPlan makes the board drawable without asking the model
// anything.
//
// Every repair here was a correction round at some point, and none of them
// taught the model anything it could not have been told by the renderer's own
// defaults: a five-word label is cut to four because the box holds four, a
// shape called "rectangle" is a box, a link to a box that gets drawn two beats
// later is a link the layout cannot draw and the sentence does not need. The
// one thing it will not do is invent a fifth box for a four-box template — that
// is content, and content is the model's job.
func normalizeWhiteboardPlan(p *SnippetPlan) {
	seen := map[string]bool{}
	total := 0
	for i := range p.Beats {
		items := make([]SketchItem, 0, len(p.Beats[i].Sketch))
		for _, item := range p.Beats[i].Sketch {
			item.Label = clampWords(collapseSpaces(item.Label), maxSketchLabelWords)
			if item.Label == "" {
				continue // a box with no caption says nothing
			}
			key := sketchKey(item.Label)
			if seen[key] || total >= maxSketchItems {
				continue // each box is drawn once and stays; the rest is overflow
			}
			item.Icon = normalizeSketchIcon(item.Icon)
			item.Shape = normalizeSketchShape(item.Shape)
			// An arrow can only start from a box already on the board. A link to
			// something drawn later has nowhere to begin, and dropping it costs
			// the picture one arrow rather than the clip a round trip.
			if item.LinkFrom != "" && !seen[sketchKey(item.LinkFrom)] {
				item.LinkFrom = ""
			}
			seen[key] = true
			total++
			items = append(items, item)
		}
		p.Beats[i].Sketch = items
	}
}

func validateWhiteboardPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Sketch: true}); err != nil {
		return err
	}
	total := 0
	seen := map[string]bool{}
	for _, b := range p.Beats {
		for _, item := range b.Sketch {
			label := strings.TrimSpace(item.Label)
			if label == "" {
				return fmt.Errorf("beat %q has a board item with no label", b.ID)
			}
			if n := len(strings.Fields(label)); n > maxSketchLabelWords {
				return fmt.Errorf("board item %q is %d words; labels are at most %d", label, n, maxSketchLabelWords)
			}
			key := sketchKey(label)
			if seen[key] {
				return fmt.Errorf("board item %q appears twice — each item is drawn once and stays", label)
			}
			// A link must point at something already on the board, or there is
			// nothing for the arrow to start from.
			if item.LinkFrom != "" && !seen[sketchKey(item.LinkFrom)] {
				return fmt.Errorf("board item %q links from %q, which is not on the board yet", label, item.LinkFrom)
			}
			seen[key] = true
			total++
		}
	}
	if total < minSketchItems || total > maxSketchItems {
		return fmt.Errorf("the board has %d items, want %d-%d", total, minSketchItems, maxSketchItems)
	}
	return nil
}

// sketchKey normalizes a label for link matching, so "The Browser" and
// "browser" name the same box.
func sketchKey(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// normalizeSketchIcon maps an invented figure name onto the neutral fallback,
// which is what the renderer's closed vocabulary expects.
func normalizeSketchIcon(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if artFigureVocab[n] {
		return n
	}
	return "spark"
}

// sketchShapes is how a board item can be drawn. Four, not more: every one has
// to mean something a person at a board would actually mean by it, and a
// vocabulary with two shapes nobody can tell apart is worse than one shape.
//
//	box     a component — something with edges. The default and the common case.
//	circle  an actor or a moment: a user, a request, the point where it fails.
//	cloud   something deliberately vague — the internet, "everything else".
//	sticky  an aside: a caveat, a note, the thing you add after the diagram.
var sketchShapes = map[string]bool{
	"box":    true,
	"circle": true,
	"cloud":  true,
	"sticky": true,
}

// SketchShapeNames returns the shape vocabulary, sorted, for the prompt.
func SketchShapeNames() []string {
	out := make([]string, 0, len(sketchShapes))
	for s := range sketchShapes {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// normalizeSketchShape falls back to the box, which is the shape that is never
// wrong — an item drawn as a component when it is really an aside still reads.
func normalizeSketchShape(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if sketchShapes[n] {
		return n
	}
	return "box"
}
