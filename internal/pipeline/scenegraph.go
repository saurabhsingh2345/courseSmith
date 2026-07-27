package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// SceneGraphFileName is the scenegraph stage output: the input props for the
// Remotion renderer.
const SceneGraphFileName = "lesson-video.json"

// Scene types.
const (
	SceneTitle    = "title"
	SceneCode     = "code"
	SceneDiagram  = "diagram"
	SceneTerminal = "terminal"
	ScenePoints   = "points"
	// SceneWalkthrough is the VS Code walkthrough: a synthesized editor whose
	// buffer evolves through timed steps (one per outline code block).
	SceneWalkthrough = "walkthrough"
	// SceneWhiteboard is one continuous board that accumulates hand-drawn
	// boxes, arrows and labels as the narrator speaks.
	SceneWhiteboard = "whiteboard"
	// SceneFlow is a layered systems diagram with traffic moving along its
	// edges and a focus that follows the narration.
	SceneFlow = "flow"
)

// maxFileNameWords caps how many slug words reach the editor tab and file
// tree. Whole section slugs produced names like
// "names_without_quotes_text_with_quotes.py" — far wider than the 264px
// sidebar, and real projects don't name files in sentences either.
const maxFileNameWords = 3

// sectionFileName derives the editor-tab file name for a section's code:
// the first few slug words, snake_cased, so it reads like a plausible module
// and still fits the sidebar.
func sectionFileName(sectionID string) string {
	words := make([]string, 0, maxFileNameWords)
	for w := range strings.SplitSeq(sectionID, "-") {
		if w == "" || isFileNameStopWord(w) {
			continue
		}
		words = append(words, w)
		if len(words) == maxFileNameWords {
			break
		}
	}
	if len(words) == 0 {
		// Every word was a stop word (or the slug was empty): fall back to the
		// raw slug so distinct sections still get distinct files.
		return strings.ReplaceAll(sectionID, "-", "_") + ".py"
	}
	return strings.Join(words, "_") + ".py"
}

// isFileNameStopWord drops the articles and prepositions that pad section
// headings but carry no meaning in a file name.
func isFileNameStopWord(w string) bool {
	switch w {
	case "a", "an", "the", "of", "in", "on", "to", "for", "with", "without", "and", "or", "is", "are", "it", "its", "your", "you":
		return true
	}
	return false
}

// defaultCalloutDurMs is how long a callout stays on screen when the lesson
// doesn't specify.
const defaultCalloutDurMs = 4000

// videoTailMs pads the video past the last spoken word.
const videoTailMs = 800

// SceneGraph is the complete render input for one lesson video
// (generated/lesson-video.json).
type SceneGraph struct {
	Theme SceneTheme `json:"theme"`
	// Motion is the animation language for this lesson's video — baseline
	// DefaultMotion(), later overridden per course by the archetype system.
	// The renderer reads these tokens instead of hardcoding timings.
	Motion Motion `json:"motion"`
	// AssetBase is where relative asset paths resolve from inside the
	// renderer's public dir; the render/preview steps set it when staging
	// assets. Empty in the persisted scene graph.
	AssetBase  string        `json:"assetBase,omitempty"`
	AudioFile  string        `json:"audioFile"`
	DurationMs int           `json:"durationMs"`
	Scenes     []Scene       `json:"scenes"`
	Captions   []AlignedWord `json:"captions"`
	// CaptionEmphasis holds global caption-word indices the emphasis pass
	// marked as keywords (accent colour in the caption track). Optional.
	CaptionEmphasis []int `json:"captionEmphasis,omitempty"`
}

// SceneTheme carries course branding into the renderer. Beyond the three
// course colours it includes the derived video tokens (dark gradient,
// surfaces, type stack) — see videotheme.go. New fields are omitempty so
// scene graphs generated before the design system still parse and render.
type SceneTheme struct {
	Primary    string `json:"primary"`
	Accent     string `json:"accent"`
	Background string `json:"background"`
	CourseName string `json:"courseName"`

	// Derived design tokens (Go-owned; renderer falls back when absent).
	Mode          string  `json:"mode,omitempty"`          // "dark" (default) | "light"
	BgTop         string  `json:"bgTop,omitempty"`         // scene gradient start
	BgBottom      string  `json:"bgBottom,omitempty"`      // scene gradient end
	Surface       string  `json:"surface,omitempty"`       // card fill
	SurfaceBorder string  `json:"surfaceBorder,omitempty"` // card hairline
	Text          string  `json:"text,omitempty"`          // main text on bg
	TextMuted     string  `json:"textMuted,omitempty"`     // secondary text
	FontDisplay   string  `json:"fontDisplay,omitempty"`   // headings
	FontBody      string  `json:"fontBody,omitempty"`      // body/captions
	FontMono      string  `json:"fontMono,omitempty"`      // code
	Grain         float64 `json:"grain,omitempty"`         // film-grain opacity 0..1
}

// Scene is one visual span of the lesson video.
type Scene struct {
	Type    string         `json:"type"` // title|code|diagram|terminal
	StartMs int            `json:"startMs"`
	EndMs   int            `json:"endMs"`
	Props   map[string]any `json:"props"`
	// Callouts are annotation overlays active during this scene.
	Callouts []Callout `json:"callouts,omitempty"`
}

// Callout is one arrow/circle/label overlay with resolved timing.
type Callout struct {
	AtMs  int     `json:"atMs"`
	DurMs int     `json:"durMs"`
	Shape string  `json:"shape"` // arrow|circle
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Label string  `json:"label"`
}

var headingRe = regexp.MustCompile(`(?m)^##\s+(.+)$`)

// sectionsFromOutline splits a lesson body into per-heading chunks keyed by
// the heading's slug — the same slugs the script generator uses as section
// ids, which is what ties outline content (code blocks) to script sections.
func sectionsFromOutline(body string) map[string]string {
	locs := headingRe.FindAllStringSubmatchIndex(body, -1)
	out := make(map[string]string, len(locs))
	for i, loc := range locs {
		heading := body[loc[2]:loc[3]]
		end := len(body)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		out[sectionKey(heading)] = body[loc[1]:end]
	}
	return out
}

// sectionKey normalizes a heading (or a script section id) to something both
// sides agree on.
//
// Section ids are written by the LLM from the heading text, and it does not
// slugify punctuation the way slugifyHeading does: "## What's next" comes back
// as "whats-next" while slugifyHeading yields "what-s-next". Keying on the
// slug alone meant such a section matched no outline chunk at all — it lost
// its code blocks *and* its title, silently, which is how "Whats Next" ended
// up as a bare heading card. Dropping every non-alphanumeric makes the two
// agree without depending on either side's punctuation rules.
func sectionKey(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// headingsFromOutline maps each section slug back to the heading exactly as
// the author wrote it. Scene titles used to be rebuilt from the slug by
// humanizeSlug, which upper-cases every word and cannot recover punctuation —
// "## Giving a value a name" came back as "Giving A Value A Name" and
// "## What's next" as "Whats Next". The real text is right here; keep it.
func headingsFromOutline(body string) map[string]string {
	matches := headingRe.FindAllStringSubmatch(body, -1)
	out := make(map[string]string, len(matches))
	for _, m := range matches {
		heading := strings.TrimSpace(m[1])
		out[sectionKey(heading)] = heading
	}
	return out
}


// cueTimestamp resolves a cue's at_word (section-relative index) to the
// timestamp of that word, clamped into the section's span.
func cueTimestamp(a *Alignment, span SectionSpan, atWord int) int {
	if span.WordEnd <= span.WordStart {
		return span.StartMs
	}
	idx := span.WordStart + atWord
	if idx >= span.WordEnd {
		idx = span.WordEnd - 1
	}
	if idx < span.WordStart {
		idx = span.WordStart
	}
	return a.Words[idx].StartMs
}

// findPhrase locates a phrase inside a section's aligned words and returns
// the start time of its first word, or -1 when absent.
func findPhrase(a *Alignment, span SectionSpan, phrase string) int {
	want := normalizeWords(strings.Fields(phrase))
	if len(want) == 0 {
		return -1
	}
	words := a.Words[span.WordStart:span.WordEnd]
	for i := 0; i+len(want) <= len(words); i++ {
		match := true
		for j, w := range want {
			if normalizeToken(words[i+j].Word) != w {
				match = false
				break
			}
		}
		if match {
			return words[i].StartMs
		}
	}
	return -1
}

// buildSceneGraph assembles the renderer input from the script, alignment,
// demo manifest, verified code, and lesson metadata. Pure — fully unit
// testable.
func buildSceneGraph(
	course *project.Course,
	l *project.Lesson,
	cfg config.Config,
	script *Script,
	alignment *Alignment,
	demos *DemoManifest,
	verification *VerificationReport,
	traces map[string]*CodeTrace,
	storyboard *Storyboard,
	audioDurMs int,
) (*SceneGraph, error) {
	// Archetype (workstream F) supplies the motion philosophy and, when set, a
	// palette that overrides the course's branding colours.
	arch, err := ResolveArchetype(cfg.Style)
	if err != nil {
		return nil, err
	}
	colors := cfg.Branding.Colors
	if arch.HasPalette {
		colors = arch.Palette
	}

	graph := &SceneGraph{
		Theme:     deriveVideoTheme(colors, cfg.Branding.Fonts, course.Name),
		Motion:    arch.Motion, // DefaultMotion() + the archetype's philosophy
		AudioFile: VoiceoverFileName,
	}
	// On-screen captions are opt-in (style.captions: on). Off, the words stay
	// out of the scene graph entirely — the renderer draws nothing and scenes
	// keep the frame. captions.vtt is still produced for players/uploads.
	if cfg.Style.Captions == "on" {
		graph.Captions = alignment.CaptionWords()
	}

	if len(alignment.Sections) != len(script.Sections) {
		return nil, fmt.Errorf("alignment has %d sections but the script has %d — re-run the align stage", len(alignment.Sections), len(script.Sections))
	}

	outline := sectionsFromOutline(l.Body)
	headings := headingsFromOutline(l.Body)
	// sectionTitle prefers the author's own heading; humanizeSlug is only the
	// fallback for sections the script invented without an outline heading.
	sectionTitle := func(id string) string {
		if h := headings[sectionKey(id)]; h != "" {
			return h
		}
		return humanizeSlug(id)
	}
	diagramKinds := diagramKindByID(l)
	boardByID := map[string][]StoryPoint{}
	if storyboard != nil {
		for _, s := range storyboard.Sections {
			boardByID[s.ID] = s.Points
		}
	}
	// The walkthrough's file tree lists every code-bearing section, so the
	// editor sidebar looks like a real project.
	var walkthroughFiles []string
	for _, sec := range script.Sections {
		if len(extractCodeBlocks(outline[sectionKey(sec.ID)], sec.ID)) > 0 {
			walkthroughFiles = append(walkthroughFiles, sectionFileName(sec.ID))
		}
	}
	demosByDesc := make(map[string]DemoEntry, len(demos.Demos))
	for _, d := range demos.Demos {
		demosByDesc[normalizeToken(d.Description)] = d
	}
	verifiedOutput := map[string]string{}
	if verification != nil {
		for _, b := range verification.Blocks {
			verifiedOutput[project.HashBytes([]byte(b.Code))] = b.Stdout
		}
	}

	for si, sec := range script.Sections {
		span := alignment.Sections[si]
		sectionEnd := span.EndMs
		if si+1 < len(alignment.Sections) {
			sectionEnd = alignment.Sections[si+1].StartMs
		} else {
			sectionEnd += videoTailMs
		}

		// The base visual for the section: the lesson title card for the
		// first section, a code scene when the outline section carries a
		// Python block, a heading card otherwise.
		base := Scene{Type: SceneTitle, StartMs: span.StartMs, Props: map[string]any{
			"heading":  sectionTitle(sec.ID),
			"subtitle": script.Title,
		}}
		if si == 0 {
			base.Props = map[string]any{
				"heading":  script.Title,
				"subtitle": course.Name,
				"outcomes": l.FrontMatter.Outcomes,
				"intro":    true,
			}
		} else if points := boardByID[sec.ID]; len(points) > 0 {
			// The storyboard turns what would be a bare heading card into
			// keyword beats that pop in on the exact spoken word.
			items := make([]map[string]any, len(points))
			for i, p := range points {
				items[i] = map[string]any{
					"text": p.Text,
					"icon": p.Icon,
					"atMs": cueTimestamp(alignment, span, p.AtWord),
				}
			}
			base = Scene{Type: ScenePoints, StartMs: span.StartMs, Props: map[string]any{
				"title": sectionTitle(sec.ID),
				"items": items,
			}}
		}
		if blocks := extractCodeBlocks(outline[sectionKey(sec.ID)], sec.ID); len(blocks) > 1 && si > 0 {
			// Multi-block sections become a VS Code walkthrough: each block is
			// one step of the same editor buffer, spread across the section.
			// (Previously only the first block was shown — the rest were lost.)
			stepDur := (sectionEnd - span.StartMs) / len(blocks)
			steps := make([]map[string]any, len(blocks))
			for i, b := range blocks {
				steps[i] = map[string]any{
					"code":   b.Code,
					"atMs":   span.StartMs + i*stepDur,
					"output": verifiedOutput[project.HashBytes([]byte(b.Code))],
				}
			}
			base = Scene{Type: SceneWalkthrough, StartMs: span.StartMs, Props: map[string]any{
				"title":    sectionTitle(sec.ID),
				"language": "python",
				"file":     sectionFileName(sec.ID),
				"project":  course.Slug,
				"files":    walkthroughFiles,
				"steps":    steps,
			}}
		} else if len(blocks) > 0 && si > 0 {
			b := blocks[0] // one code scene per section; the first block leads
			hash := project.HashBytes([]byte(b.Code))
			props := map[string]any{
				"title":    sectionTitle(sec.ID),
				"code":     b.Code,
				"language": "python",
				"output":   verifiedOutput[hash],
			}
			// Attach the execution trace when one exists and has real steps, so
			// the renderer shows variable state stepping through instead of a
			// plain typing scene. Static/expression-only blocks (no steps) stay
			// as ordinary code scenes.
			if tr := traces[hash]; tr != nil && len(tr.Steps) > 0 {
				props["trace"] = tr
			}
			base = Scene{Type: SceneCode, StartMs: span.StartMs, Props: props}
		}

		// Cues split the section into consecutive scenes.
		scenes := []Scene{base}
		for _, cue := range sec.Cues {
			at := cueTimestamp(alignment, span, cue.AtWord)
			var next *Scene
			switch cue.Type {
			case CueDiagram:
				kind := diagramKinds[cue.Ref]
				if kind == "" {
					kind = project.DiagramKindSVG
				}
				next = &Scene{Type: SceneDiagram, StartMs: at, Props: map[string]any{
					"src":     diagramSceneSrc(cue.Ref, kind),
					"kind":    kind,
					"title":   sectionTitle(sec.ID),
					"caption": cue.Ref,
				}}
			case CueDemo:
				demo, ok := demosByDesc[normalizeToken(cue.Ref)]
				if !ok {
					// The script's demo description drifted from the marker;
					// fall back to any single demo, else skip the cue.
					if len(demos.Demos) == 1 {
						demo = demos.Demos[0]
					} else {
						continue
					}
				}
				next = &Scene{Type: SceneTerminal, StartMs: at, Props: map[string]any{
					"src":        demo.Path,
					"title":      demo.Description,
					"durationMs": demo.DurationMs,
				}}
			default: // pause cues don't change the visual
				continue
			}
			if next.StartMs <= scenes[len(scenes)-1].StartMs {
				// Degenerate timing (estimated alignment): replace the base
				// rather than emitting a zero-length scene.
				scenes[len(scenes)-1] = *next
			} else {
				scenes = append(scenes, *next)
			}
		}
		for i := range scenes {
			if i+1 < len(scenes) {
				scenes[i].EndMs = scenes[i+1].StartMs
			} else {
				scenes[i].EndMs = sectionEnd
			}
		}

		// Attach callouts declared for this section to the scene covering
		// their timestamp.
		for _, spec := range l.FrontMatter.Callouts {
			if spec.Section != sec.ID {
				continue
			}
			at := findPhrase(alignment, span, spec.At)
			if at < 0 {
				at = span.StartMs // phrase not found: show at section start
			}
			dur := spec.DurMs
			if dur <= 0 {
				dur = defaultCalloutDurMs
			}
			callout := Callout{AtMs: at, DurMs: dur, Shape: spec.Shape, X: spec.X, Y: spec.Y, Label: spec.Label}
			for i := range scenes {
				if at >= scenes[i].StartMs && (at < scenes[i].EndMs || i == len(scenes)-1) {
					scenes[i].Callouts = append(scenes[i].Callouts, callout)
					break
				}
			}
		}
		graph.Scenes = append(graph.Scenes, scenes...)
	}

	// Stamp each scene with its archetype-selected template variant (the
	// renderer falls back to its default for "" or unknown names).
	for i := range graph.Scenes {
		if _, set := graph.Scenes[i].Props["template"]; !set {
			if tpl := arch.TemplateFor(graph.Scenes[i].Type); tpl != "" {
				graph.Scenes[i].Props["template"] = tpl
			}
		}
	}

	last := 0
	if n := len(graph.Scenes); n > 0 {
		last = graph.Scenes[n-1].EndMs
	}
	graph.DurationMs = max(audioDurMs, last)
	if n := len(graph.Scenes); n > 0 {
		graph.Scenes[n-1].EndMs = graph.DurationMs
	}
	return graph, nil
}

// LoadSceneGraph reads the lesson's generated lesson-video.json.
func LoadSceneGraph(l *project.Lesson) (*SceneGraph, error) {
	path := filepath.Join(l.GeneratedDir(), SceneGraphFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no %s yet — the scenegraph stage must run first", SceneGraphFileName)
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var g SceneGraph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it and re-run the scenegraph stage): %w", path, err)
	}
	return &g, nil
}

// runScenegraphStage builds generated/lesson-video.json from the script,
// word alignment, demo manifest, and verified code outputs.
func runScenegraphStage(ctx context.Context, e *Env, course *project.Course, l *project.Lesson, cfg config.Config) error {
	// A snippet's scenes come from its template, not from the lesson-shaped
	// script/storyboard/diagram machinery. Everything after this branch —
	// captions, the video-plan edit layer, the written artifact — is shared.
	if IsSnippet(l) {
		return runSnippetScenegraph(ctx, e, course, l, cfg)
	}
	script, err := loadScript(l)
	if err != nil {
		return err
	}
	alignment, err := loadAlignment(l)
	if err != nil {
		return err
	}
	demos, err := loadDemoManifest(l)
	if err != nil {
		return err
	}
	verification, err := loadVerification(l)
	if err != nil {
		return err
	}
	traces, err := loadTraces(l)
	if err != nil {
		return err
	}
	audioDur, err := wavDuration(filepath.Join(l.GeneratedDir(), VoiceoverFileName))
	if err != nil {
		return fmt.Errorf("no usable %s — the audio stage must run first: %w", VoiceoverFileName, err)
	}

	storyboard, err := loadStoryboard(l)
	if err != nil {
		return err
	}

	fmt.Fprintf(e.out(), "  → scenegraph building %s...\n", SceneGraphFileName)
	graph, err := buildSceneGraph(course, l, cfg, script, alignment, demos, verification, traces, storyboard, int(audioDur.Milliseconds()))
	if err != nil {
		return err
	}
	return finishSceneGraph(e, l, graph)
}

// finishSceneGraph applies the shared post-build layers to any scene graph —
// caption emphasis, the human video-plan edits — and writes it out. Lessons
// and snippets differ in how their scenes are built and in nothing after.
func finishSceneGraph(e *Env, l *project.Lesson, graph *SceneGraph) error {
	// Emphasis indices address caption words; without captions there is
	// nothing for them to point at.
	if len(graph.Captions) > 0 {
		if emphasis, err := loadCaptionEmphasis(l); err != nil {
			return err
		} else if emphasis != nil {
			graph.CaptionEmphasis = emphasis.Indices
		}
	}
	// Human edit layer: apply video-plan.yaml (template swaps, prop patches,
	// skipped scenes) over the generated scenes.
	plan, err := loadVideoPlan(l)
	if err != nil {
		return err
	}
	if plan != nil {
		if err := applyVideoPlan(graph, plan); err != nil {
			return err
		}
		fmt.Fprintf(e.out(), "    applied %d edit(s) from %s\n", len(plan.Edits), VideoPlanFileName)
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), SceneGraphFileName), graph); err != nil {
		return err
	}
	types := map[string]int{}
	for _, s := range graph.Scenes {
		types[s.Type]++
	}
	fmt.Fprintf(e.out(), "    %d scenes (%d title, %d points, %d code, %d walkthrough, %d whiteboard, %d flow, %d diagram, %d terminal), %d captions, %.1fs\n",
		len(graph.Scenes), types[SceneTitle], types[ScenePoints], types[SceneCode], types[SceneWalkthrough],
		types[SceneWhiteboard], types[SceneFlow], types[SceneDiagram], types[SceneTerminal],
		len(graph.Captions), float64(graph.DurationMs)/1000)
	return nil
}
