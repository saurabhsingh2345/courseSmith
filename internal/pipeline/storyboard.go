package pipeline

// Storyboard stage: the fix for dead screens. Sections without a declared
// diagram/demo used to render as a bare heading card for their whole
// duration; this stage plans per-section visual beats — keyword points with
// icons, each timed to the narration word it belongs to — so every section
// has a visual that tracks what is being said.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// StoryboardFileName is the storyboard stage output in generated/.
const StoryboardFileName = "storyboard.json"

// maxPointsPerSection bounds a section's beat count (readability, not model
// whim).
const maxPointsPerSection = 5

// pointIconVocab is the closed icon vocabulary the model picks from; the
// renderer maps these names to bundled Lucide icons. "dot" is the neutral
// fallback for anything the model invents.
var pointIconVocab = map[string]bool{
	"idea": true, "code": true, "rocket": true, "book": true, "check": true,
	"alert": true, "list": true, "box": true, "arrow": true, "terminal": true,
	"database": true, "globe": true, "clock": true, "star": true, "search": true,
	"play": true, "layers": true, "link": true, "zap": true, "shield": true,
	"target": true, "gear": true, "heart": true, "flag": true, "brain": true,
	"sparkles": true, "puzzle": true, "wrench": true, "folder": true,
	"message": true, "repeat": true, "bug": true, "file": true, "keyboard": true,
	"monitor": true, "dot": true,
	// Systems and infrastructure. Added for the whiteboard template, whose
	// boards are usually architecture: without these the model reached for
	// "server" and "cloud", got normalized to "dot", and a box that should have
	// read as an origin server showed a bare placeholder circle.
	"server": true, "cloud": true, "lock": true, "users": true, "chart": true,
	"shuffle": true, "filter": true, "network": true,
}

// normalizePointIconName returns the icon name if it is in the vocabulary, and
// "" if it is not — so a caller can fall back to something better than "dot"
// when it has a better default of its own (a canvas card takes its kind's icon).
func normalizePointIconName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if pointIconVocab[n] {
		return n
	}
	return ""
}

// PointIconNames returns the vocabulary sorted, for prompts and docs.
func PointIconNames() []string {
	out := make([]string, 0, len(pointIconVocab))
	for k := range pointIconVocab {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// StoryPoint is one visual beat: a short keyword phrase that appears on the
// exact narration word it belongs to.
type StoryPoint struct {
	// Text is the on-screen phrase (2-6 words, not a sentence).
	Text string `json:"text"`
	// Icon is a pointIconVocab name.
	Icon string `json:"icon"`
	// AtWord is the 0-based index of the word in the section's narration
	// where this point should appear — the same convention as script cues.
	AtWord int `json:"at_word"`
}

// StoryboardSection carries the beats for one script section.
type StoryboardSection struct {
	ID     string       `json:"id"`
	Points []StoryPoint `json:"points"`
}

// Storyboard is the persisted storyboard.json.
type Storyboard struct {
	Sections []StoryboardSection `json:"sections"`
}

// storyboardReply is the model's raw shape before validation.
type storyboardReply struct {
	Sections []StoryboardSection `json:"sections"`
}

// runStoryboardStage plans the visual beats for every section from the
// reviewed script. Requires an LLM; without one the stage is skipped and the
// scenegraph keeps its heading-card fallback.
func runStoryboardStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	if e.Router == nil {
		fmt.Fprintf(e.out(), "  ⚠ storyboard no LLM configured — skipping (sections keep heading cards)\n")
		return nil
	}
	script, err := loadScript(l)
	if err != nil {
		return err
	}

	fmt.Fprintf(e.out(), "  → storyboard planning visual beats for %d sections (%s)...\n",
		len(script.Sections), cfg.Pipeline.LLMContent)

	type sectionCtx struct {
		ID        string
		Narration string
		WordCount int
	}
	sections := make([]sectionCtx, len(script.Sections))
	for i, sec := range script.Sections {
		sections[i] = sectionCtx{ID: sec.ID, Narration: sec.Narration, WordCount: len(strings.Fields(sec.Narration))}
	}

	system, user, err := e.renderPrompt(storyboardTemplateName, map[string]any{
		"Title":     script.Title,
		"Audience":  cfg.Style.Audience,
		"Tone":      cfg.Style.Tone,
		"MaxPoints": maxPointsPerSection,
		"Icons":     strings.Join(PointIconNames(), ", "),
		"Sections":  sections,
	})
	if err != nil {
		return err
	}

	var reply storyboardReply
	if err := e.completeJSON(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.3, 4096, &reply, nil); err != nil {
		return fmt.Errorf("storyboard generation: %w", err)
	}

	board := normalizeStoryboard(&reply, script)
	if err := writeJSON(filepath.Join(l.GeneratedDir(), StoryboardFileName), board); err != nil {
		return err
	}
	total := 0
	for _, s := range board.Sections {
		total += len(s.Points)
	}
	fmt.Fprintf(e.out(), "    %s written (%d beats across %d sections)\n", StoryboardFileName, total, len(board.Sections))
	return nil
}

// normalizeStoryboard validates the model reply against the script: unknown
// sections are dropped, icons fall back to "dot", at_word is clamped into the
// section, points are capped and ordered by at_word.
func normalizeStoryboard(reply *storyboardReply, script *Script) *Storyboard {
	wordCount := map[string]int{}
	order := map[string]int{}
	for i, sec := range script.Sections {
		wordCount[sec.ID] = len(strings.Fields(sec.Narration))
		order[sec.ID] = i
	}

	out := &Storyboard{}
	for _, sec := range reply.Sections {
		n, ok := wordCount[sec.ID]
		if !ok {
			continue
		}
		var points []StoryPoint
		for _, p := range sec.Points {
			text := strings.TrimSpace(p.Text)
			if text == "" {
				continue
			}
			if len(strings.Fields(text)) > 8 {
				text = strings.Join(strings.Fields(text)[:8], " ")
			}
			icon := strings.ToLower(strings.TrimSpace(p.Icon))
			if !pointIconVocab[icon] {
				icon = "dot"
			}
			at := p.AtWord
			if at < 0 {
				at = 0
			}
			if n > 0 && at >= n {
				at = n - 1
			}
			points = append(points, StoryPoint{Text: text, Icon: icon, AtWord: at})
			if len(points) >= maxPointsPerSection {
				break
			}
		}
		sort.SliceStable(points, func(i, j int) bool { return points[i].AtWord < points[j].AtWord })
		out.Sections = append(out.Sections, StoryboardSection{ID: sec.ID, Points: points})
	}
	sort.SliceStable(out.Sections, func(i, j int) bool { return order[out.Sections[i].ID] < order[out.Sections[j].ID] })
	return out
}

// loadStoryboard reads storyboard.json; missing is nil (optional stage).
func loadStoryboard(l *project.Lesson) (*Storyboard, error) {
	path := filepath.Join(l.GeneratedDir(), StoryboardFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var sb Storyboard
	if err := json.Unmarshal(data, &sb); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it and re-run the storyboard stage): %w", path, err)
	}
	return &sb, nil
}
