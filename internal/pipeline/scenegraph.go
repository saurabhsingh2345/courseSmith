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
	SceneFootage  = "footage"
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
	// SceneIllustration is one kinetic-typography shot: a headline set word by
	// word beside a flat-vector figure. Unlike the board and the diagram it
	// does not accumulate — a clip is a run of these, one per beat.
	SceneIllustration = "illustration"
	// SceneSpine is one staged shot of a narration: a headline, an arrangement
	// of figures chosen from nine layouts, and a rail down the left edge saying
	// where in the clip this beat sits. Like the illustration it does not
	// accumulate — but unlike it, consecutive shots may be arranged completely
	// differently, which is what lets one clip open, explain, turn and close.
	SceneSpine = "spine"
	// SceneCast is one shot of a character explaining something: a posed,
	// breathing person beside a kinetic headline.
	SceneCast = "cast"
	// SceneStory is one shot of a directed piece: a staged arrangement of
	// character and objects, framed by a moving camera.
	SceneStory = "story"
	// SceneWorkspace is a full-frame editor across several files, shot like a
	// screen recording: tabs, a camera that moves between the tree, the code
	// and the terminal, and the real output of running the project.
	SceneWorkspace = "workspace"
	// SceneData is one persistent chart or world map whose highlight follows
	// the narration.
	SceneData = "data"
	// SceneQuiz is a question posed, held, and answered — the one template
	// built around the viewer doing something rather than watching.
	SceneQuiz = "quiz"
	// SceneCompare is two subjects in one frame, introduced in turn and then
	// judged — the only template that shows more than one thing at a time.
	SceneCompare = "compare"
	// SceneAnatomy is one artefact held still while callouts reach to each of
	// its labelled parts in turn.
	SceneAnatomy = "anatomy"
	// SceneTimeline is a spine of milestones filling in as the narration walks
	// it — for anything whose subject is order.
	SceneTimeline = "timeline"
	// SceneCanvas is an automation wired across a builder's canvas: app cards
	// on a dotted grid, and a payload that runs the chain end to end.
	SceneCanvas = "canvas"
	// ScenePromptLoop is a conversation with something that builds: prompts
	// stacking down one column, and what came back beside them.
	ScenePromptLoop = "promptloop"
	// SceneMockup is a page assembling itself inside a device frame, with the
	// layer list filling in beside it.
	SceneMockup = "mockup"
	// SceneStack is a set of tiers stacked vertically — which tool does what,
	// walked from the top of the stack down.
	SceneStack = "stack"
	// SceneSpec is a checklist written a line at a time and then checked all at
	// once — acceptance criteria, and the moment they go green.
	SceneSpec = "spec"
	// SceneShowcase is one tool's card — what it is, what it costs, what it is
	// good and bad at — ending on a hand-off plate a demo recording cuts onto.
	SceneShowcase = "showcase"
	// SceneBreakdown is a path of phases where the current one opens into its
	// own detail — the description, and the items inside it.
	SceneBreakdown = "breakdown"
	// SceneMetric is one figure at a time, set large enough to be the whole
	// frame, with its unit, what it counts, and what that means.
	SceneMetric = "metric"
	// SceneGauge is a bar filling toward a marked ceiling — what clears it,
	// what runs past, and by how much.
	SceneGauge = "gauge"
	// SceneVerdict is a ruling: the ground it holds on, the asterisk that
	// qualifies it, and the call alone on the closing frame.
	SceneVerdict = "verdict"
	// SceneDecision is one question on an axis split into tiers, each band
	// carrying the answer for landing in it.
	SceneDecision = "decision"
	// SceneMyth is a widely-held belief struck through in place and replaced by
	// what is actually the case.
	SceneMyth = "myth"
	// SceneRundown is a numbered row that promises how many things there are
	// and then lights them one at a time.
	SceneRundown = "rundown"
	// SceneAnalogy is a familiar picture in one column and what each of its
	// parts really is in the other, walked pair by pair.
	SceneAnalogy = "analogy"
	// SceneTrace is a system caught in the act: actors issuing work into a
	// queue that drains against one shared value.
	SceneTrace = "trace"
	// SceneCosting is a bill built line by line, with a running total that
	// moves as each cost lands.
	SceneCosting = "costing"
	// SceneConstellation is one idea in the middle with its properties
	// radiating out, lit one spoke at a time.
	SceneConstellation = "constellation"
	// SceneChapter is the break between two stretches of teaching: a huge
	// ordinal, the section starting now, and the path it sits on.
	SceneChapter = "chapter"
	// SceneCycle is a closed ring of stages with a light running round it and
	// a hub that says what is different next lap.
	SceneCycle = "cycle"
	// SceneScale is a ladder of nested worlds with the camera pulling back an
	// order of magnitude at a time.
	SceneScale = "scale"
	// SceneOccupancy is a grid of identical units, all of it visible at once,
	// with bands of it claimed one at a time.
	SceneOccupancy = "occupancy"
	// SceneRanking is an ordered board that re-sorts as entries land on it, the
	// rows sliding to their new places rather than being redrawn.
	SceneRanking = "ranking"
	// SceneJournal is an append-only file growing at the bottom, then replayed
	// from the top with a cursor walking down it.
	SceneJournal = "journal"
	// SceneMultiplex is a pool of identical sources where several go ready at
	// once and a single worker takes them in one pass.
	SceneMultiplex = "multiplex"
	// SceneFork is two processes over one memory, where a write splits a single
	// page onto the writer's side and leaves the rest shared.
	SceneFork = "fork"
	// SceneCapabilities is a boundary around a subject with the things it cannot
	// reach outside it, and the one or two deliberately handed in.
	SceneCapabilities = "capabilities"
	// SceneBudget is a fixed pot with claims taken out of it one at a time,
	// closing on what is left.
	SceneBudget = "budget"
	// SceneLatency is a logarithmic time axis with its decades named and
	// operations placed along it.
	SceneLatency = "latency"
	// SceneMultiply is a per-unit figure, a row of glyphs for the count, and the
	// product that comes of the two.
	SceneMultiply = "multiply"
	// SceneRatio is two measurements on a shared scale with the proportion
	// between them named in words.
	SceneRatio = "ratio"
	// SceneTable is a spec sheet shown evenly weighted and then stripped back to
	// the one row that decides things.
	SceneTable = "table"
	// SceneToggle is a question answered by a switch in the first beat, with the
	// qualifiers accumulating under it.
	SceneToggle = "toggle"
	// The v5 course-scaffolding scenes. Each draws the shape of a lesson's
	// relationship to the rest of the course rather than a subject inside it.
	SceneObjective  = "objective"
	ScenePrereq     = "prereq"
	SceneRecap      = "recap"
	ScenePitfall    = "pitfall"
	SceneCheckpoint = "checkpoint"
	// The v7 foundations scenes: the pictures a computer-science course keeps
	// asking for. Course scaffolding first, then the machines, the data, the
	// runtime, the network, the algorithms, and the history.
	SceneSyllabus  = "syllabus"
	SceneOutcome   = "outcome"
	SceneBridge    = "bridge"
	SceneDrill     = "drill"
	SceneLabCard   = "labcard"
	SceneMission   = "mission"
	SceneMachine   = "machine"
	SceneBlueprint = "blueprint"
	SceneRelay     = "relay"
	SceneLayers    = "layers"
	ScenePipeline  = "pipeline"
	SceneRadix     = "radix"
	SceneCarry     = "carry"
	SceneBitfield  = "bitfield"
	SceneEncode    = "encode"
	SceneGates     = "gates"
	SceneLadder    = "ladder"
	SceneRegions   = "regions"
	SceneLookup    = "lookup"
	SceneStates    = "states"
	SceneScheduler = "scheduler"
	SceneShell     = "shell"
	SceneJourney   = "journey"
	SceneHandshake = "handshake"
	SceneStepper   = "stepper"
	SceneGrowth    = "growth"
	SceneCallStack = "callstack"
	SceneHistory   = "history"
	SceneVersus    = "versus"
	SceneEras      = "eras"
	// SceneCards is a row of named things wearing their own fetched marks, with
	// vs, an arrow, or nothing at all in the gaps between them.
	SceneCards = "cards"
	// SceneDuel is two named things on two cards with one measured bar each.
	SceneDuel = "duel"
	// SceneSpotlight is one card on the left and its claims stacked on the right.
	SceneSpotlight = "spotlight"
	// SceneOpener is the title page: the subject set enormous in a serif as the
	// ground of the frame, with the promise and the byline solid on top.
	SceneOpener = "opener"
	// SceneChangePlan is the rail of files a change touches, one open beside it.
	SceneChangePlan = "changeplan"
	// ScenePatch is one hunk of a diff at a size it can be read at.
	ScenePatch = "patch"
	// SceneApproval is a permission prompt and what each answer hands over.
	SceneApproval = "approval"
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
	AssetBase string `json:"assetBase,omitempty"`
	AudioFile string `json:"audioFile"`
	// SFXFile is the generated keystroke track, played as a second audio layer
	// under the voice. Only present when the video actually types something —
	// see keysound.go. Empty for every template that shows no editor.
	SFXFile    string        `json:"sfxFile,omitempty"`
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
	Mode string `json:"mode,omitempty"` // "dark" (default) | "light"
	// Skin is the house style: "default" (unchanged, and the default),
	// "broadcast" or "minimal". It is an axis independent of Mode — every skin
	// derives in both polarities. See videoskin.go.
	Skin string `json:"skin,omitempty"`
	// Air is how far a skin pulls content in from the stage edges, as a
	// fraction of the drawing box (0 = fill the stage, the default).
	Air           float64 `json:"air,omitempty"`
	BgTop         string  `json:"bgTop,omitempty"`         // scene gradient start
	BgBottom      string  `json:"bgBottom,omitempty"`      // scene gradient end
	Surface       string  `json:"surface,omitempty"`       // card fill
	SurfaceBorder string  `json:"surfaceBorder,omitempty"` // card hairline
	Text          string  `json:"text,omitempty"`          // main text on bg
	TextMuted     string  `json:"textMuted,omitempty"`     // secondary text
	// Semantic accents: the three roles a precise diagram colours by. Unlike
	// Accent these are *not* branding — a bar that overruns its ceiling is red
	// whatever the course is branded with, because the colour is saying what
	// the picture means. Derived for every skin. See videoskin.go.
	AccentQuantity string `json:"accentQuantity,omitempty"` // the measured number
	AccentLimit    string `json:"accentLimit,omitempty"`    // the ceiling it hits
	AccentRival    string `json:"accentRival,omitempty"`    // the alternative weighed
	// Watermark is the standing mark a chrome-carrying skin sets in the corner
	// of every frame. Empty leaves the corner clean.
	Watermark string `json:"watermark,omitempty"`
	// Mass is the body fill of drawn artwork, and Ink is the shading laid over
	// a mass to give it a lit and an unlit face. They are a pair and they flip
	// together: on the dark stage a mass is near-white, on paper it is a
	// mid-tone, and in both cases Ink is darker than Mass so the same shading
	// code works either way. A figure painted in a literal colour instead of
	// these is a figure that vanishes in one of the two modes.
	Mass string `json:"mass,omitempty"`
	Ink  string `json:"ink,omitempty"`
	// Elevation: how an object is seated on this background. Shadow is the colour
	// a cast shadow is drawn in and ShadowStrength how opaque it is; Rim is the
	// hairline highlight along a lit object's top edge. They exist because
	// "lift this off the surface" is two opposite effects — on the near-black
	// stage the rim does the seating and the shadow is nearly invisible, on paper
	// it is the other way round — and a scene that hardcodes either one is a
	// scene that looks pasted on in the other polarity. See deriveElevation.
	Shadow         string  `json:"shadow,omitempty"`
	ShadowStrength float64 `json:"shadowStrength,omitempty"`
	Rim            string  `json:"rim,omitempty"`
	// AccentText is the accent adjusted to be legible as text on this mode's
	// background. Accent itself stays the brand colour and is what fills and
	// strokes use; only type takes this one. See readableOn.
	AccentText  string  `json:"accentText,omitempty"`
	FontDisplay string  `json:"fontDisplay,omitempty"` // headings
	FontBody    string  `json:"fontBody,omitempty"`    // body/captions
	FontMono    string  `json:"fontMono,omitempty"`    // code
	FontSerif   string  `json:"fontSerif,omitempty"`   // intro cards
	Grain       float64 `json:"grain,omitempty"`       // film-grain opacity 0..1
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
		Theme: deriveVideoThemeSkinned(colors, cfg.Branding.Fonts, course.Name,
			cfg.Style.Mode, cfg.Style.Skin, cfg.Style.Watermark),
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
				// A web capture is stills behind browser chrome rather than a
				// clip in a terminal window, and its address bar shows the
				// origin the driver recorded — not a caption. A frame whose
				// sidecar is missing or empty is skipped rather than drawn as
				// an empty browser: the narration will be talking over a hole
				// either way, and an empty chrome asserts something happened.
				if demo.Kind == CaptureKindWeb {
					f, ok := loadFootageFor(l, demo.ID)
					if !ok || (len(f.Frames) == 0 && demo.DurationMs == 0) {
						continue
					}
					p := map[string]any{"origin": f.Origin, "title": f.Origin, "provClipId": demo.ID}
					if demo.DurationMs > 0 {
						// A recorded clip. Its marks were measured against our
						// own clock rather than modelled from a script, so they
						// are never approximate and pacing can always use them.
						p["src"] = demo.Path
						p["durationMs"] = demo.DurationMs
						p["clipId"] = demo.ID
					} else {
						p["frames"] = f.Frames
					}
					next = &Scene{Type: SceneFootage, StartMs: at, Props: p}
					break
				}

				// The window title is the demo's description for a python
				// demo, and the tool's own name for a tool capture. The
				// credibility of a capture comes from looking like the thing
				// on the viewer's second monitor — a terminal running Claude
				// Code has "Claude Code" in its title bar, not a sentence
				// describing what we are about to do with it.
				title := demo.Description
				if demo.Kind == CaptureKindTool {
					if tool, ok := captureTools[demo.Tool]; ok {
						title = tool.Display
					}
				}
				if demo.Kind == CaptureKindDesktop {
					if app, ok := captureApps[demo.Tool]; ok {
						title = app.Display
					}
				}
				next = &Scene{Type: SceneTerminal, StartMs: at, Props: map[string]any{
					"src":        demo.Path,
					"title":      title,
					"durationMs": demo.DurationMs,
					"provClipId": demo.ID,
					// "segments" is added once every scene has an end — see
					// applyTerminalPacing.
					"clipId": demo.ID,
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
	applyTerminalPacing(graph, l)
	return graph, nil
}

// applyTerminalPacing decides how each recording is played, once every scene
// knows how long it is on screen for.
//
// It runs here rather than where the scene is built because a slot's length is
// only known after the following scene's start is — and the whole decision is
// "does this recording fit in that". The plan is written into the scene graph
// rather than computed in the renderer so it is auditable with everything else:
// `lesson-video.json` records what was decided, and a Go test can check the
// arithmetic. The renderer only plays what it is told.
func applyTerminalPacing(graph *SceneGraph, l *project.Lesson) {
	for i := range graph.Scenes {
		s := &graph.Scenes[i]
		// Both kinds of recording need this. A web clip of an app being built
		// runs long for exactly the same reason an agent session does — real
		// work takes real time — and its dead air is the same dead air.
		if s.Type != SceneTerminal && s.Type != SceneFootage {
			continue
		}
		id, _ := s.Props["clipId"].(string)
		clipMs, _ := s.Props["durationMs"].(int)
		delete(s.Props, "clipId")
		slotMs := s.EndMs - s.StartMs
		if id == "" || clipMs <= 0 || slotMs <= 0 || clipMs <= slotMs {
			continue
		}
		// Only exact marks are cut points. This is where the approximate flag
		// finally earns its place: a clip whose timing could not be attributed
		// gets a uniform speed-up rather than a confident cut in the wrong spot.
		var marks []FootageMark
		if f, ok := loadFootageFor(l, id); ok && f.Exact() {
			marks = f.Marks
		}
		s.Props["segments"] = PlanTerminalPacing(clipMs, slotMs, marks)
	}
	applyCaptureProvenance(graph, l)
}

// applyCaptureProvenance gives every capture scene the credit it states on
// screen: what tool it is of, what version, and — when the clip was compressed
// to fit — both durations.
//
// It runs after pacing because the "shown in" figure is the slot, which is only
// settled once every scene has an end. Nothing here is generated: the tool name
// comes from the engine's own registry and the version and durations were
// measured at capture time.
func applyCaptureProvenance(graph *SceneGraph, l *project.Lesson) {
	for i := range graph.Scenes {
		s := &graph.Scenes[i]
		if s.Type != SceneTerminal && s.Type != SceneFootage {
			continue
		}
		id, _ := s.Props["provClipId"].(string)
		delete(s.Props, "provClipId")
		if id == "" {
			continue
		}
		f, ok := loadFootageFor(l, id)
		if !ok {
			continue
		}
		// A python demo is our own code running, not somebody else's product,
		// so it makes no provenance claim and needs no chip.
		display := ""
		switch f.Kind {
		case CaptureKindTool:
			display = captureTools[f.Tool].Display
		case CaptureKindWeb:
			display = captureSites[f.Tool].Display
		case CaptureKindDesktop:
			display = captureApps[f.Tool].Display
		default:
			continue
		}
		clipMs, _ := s.Props["durationMs"].(int)
		p := captureCreditFor(f, display, clipMs, s.EndMs-s.StartMs)
		if p.Tool == "" && p.Version == "" && p.ShownMs == 0 {
			continue
		}
		s.Props["provenance"] = p
	}
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
	// A no-code piece is segments on one timeline, which is what the combo
	// assembler already does — the surfaces differ in what a segment must
	// stand on, not in how the cut is made.
	if IsCombo(l) || IsNoCode(l) {
		return runComboScenegraph(ctx, e, course, l, cfg)
	}
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

	// The keystroke track. Generated after the video plan, so an edit that
	// retimed a scene retimes its typing sound with it, and from the scene
	// graph's own schedule, so the clicks and the characters are the same
	// numbers rather than two implementations that agree until one is changed.
	if times, newlines := collectKeystrokes(graph); len(times) > 0 {
		path := filepath.Join(l.GeneratedDir(), KeySoundFileName)
		n, err := WriteKeystrokeTrack(path, times, newlines, graph.DurationMs)
		if err != nil {
			// Texture, not content: a clip with no typing sound is a clip, and
			// failing the build over it would be the wrong trade.
			fmt.Fprintf(e.out(), "  ⚠ scenegraph could not write %s: %v\n", KeySoundFileName, err)
		} else if n > 0 {
			graph.SFXFile = KeySoundFileName
			fmt.Fprintf(e.out(), "    %s: %d keystrokes\n", KeySoundFileName, n)
		}
	}

	if err := writeJSON(filepath.Join(l.GeneratedDir(), SceneGraphFileName), graph); err != nil {
		return err
	}
	types := map[string]int{}
	for _, s := range graph.Scenes {
		types[s.Type]++
	}
	fmt.Fprintf(e.out(), "    %d scenes (%d title, %d points, %d code, %d walkthrough, %d whiteboard, %d flow, %d illustration, %d cast, %d story, %d data, %d diagram, %d terminal), %d captions, %.1fs\n",
		len(graph.Scenes), types[SceneTitle], types[ScenePoints], types[SceneCode], types[SceneWalkthrough],
		types[SceneWhiteboard], types[SceneFlow], types[SceneIllustration], types[SceneCast], types[SceneStory], types[SceneData], types[SceneDiagram], types[SceneTerminal],
		len(graph.Captions), float64(graph.DurationMs)/1000)
	return nil
}
