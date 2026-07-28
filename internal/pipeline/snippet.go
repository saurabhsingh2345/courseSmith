package pipeline

// Snippets: a short, standalone video from one prompt plus one visual
// template.
//
// A course lesson is a document first and a video second — you write
// lesson.md, the pipeline drafts narration from it, reviews it, storyboards
// it, and only then renders. That is the right shape for a 12-minute lesson
// and the wrong shape for the 30-second clip a creator actually wants to drop
// into a landing page: it asks for authoring effort the clip does not deserve.
//
// A snippet inverts it. The prompt IS the input, the template decides what the
// screen looks like, and one LLM call produces the narration and the visual
// spec together. Everything downstream — TTS, word alignment, captions,
// scenegraph, Remotion — is the ordinary video path, reused unchanged, so a
// snippet inherits the whole quality moat (real executed code, word-accurate
// timing, the design system) without a second engine.
//
// On disk a snippet is an ordinary lesson directory inside a synthetic
// single-purpose course, which is what lets the existing stage machinery,
// state tracking, and studio artifact serving work with no special cases:
//
//	.coursesmith/snippets/
//	  course.yaml
//	  lessons/<id>/
//	    snippet.yaml        the request: prompt, template, overrides
//	    lesson.md           synthesized by the plan stage (verify reads it)
//	    generated/          snippet-plan.json, script.json, …, final.mp4

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

const (
	// SnippetFileName is the request file in a snippet's directory.
	SnippetFileName = "snippet.yaml"
	// SnippetPlanFileName is the plan stage output in generated/.
	SnippetPlanFileName = "snippet-plan.json"
	// SnippetsCourseSlug is the slug of the synthetic course snippets live in.
	SnippetsCourseSlug = "snippets"
)

// SnippetsRoot is where snippet courses live, relative to the project root.
var SnippetsRoot = filepath.Join(".coursesmith", "snippets")

// defaultSnippetTargetSec is the runtime a snippet aims for when the request
// does not say. Short enough to hold attention, long enough to teach one idea.
const defaultSnippetTargetSec = 45

// snippetTargetBounds clamp what a caller may ask for. Below the floor there
// is no room for narration to land; above the ceiling it is a lesson, and the
// lesson path (with its review gates) is the right tool.
//
// Ten seconds is about thirty words at a normal pace — two beats of fifteen,
// which is a hook rather than an explanation, and exactly what a landing page
// or a social post wants. It is only expressible because the beat count is
// derived from the budget (beatBounds); against the old fixed three-beat floor
// a ten-second clip was arithmetically impossible, and so, it turned out, was
// a twenty-second one.
const (
	minSnippetTargetSec = 10
	maxSnippetTargetSec = 180
)

// SnippetSpec is snippet.yaml: everything the creator asked for.
type SnippetSpec struct {
	// ID is the snippet's directory name, and its stable handle everywhere.
	ID string `yaml:"id"`
	// Prompt is the creator's request in their own words — the whole input.
	Prompt string `yaml:"prompt"`
	// Template names a registered visual template (see snippet_templates.go).
	Template string `yaml:"template"`
	// Title overrides the title the model would have written ("" = let it).
	Title string `yaml:"title,omitempty"`
	// TargetSec is the runtime to aim for (0 = defaultSnippetTargetSec).
	TargetSec int `yaml:"target_sec,omitempty"`
	// CodeLanguage is the language for code-bearing templates ("" = python).
	CodeLanguage string `yaml:"code_language,omitempty"`
	// Config overrides the course defaults for this snippet alone (voice,
	// palette, captions…), merged in the ordinary layered way.
	Config config.Config `yaml:",inline"`

	CreatedAt time.Time `yaml:"created_at,omitempty"`
}

// ResolvedTargetSec returns the runtime to aim for, defaulted and clamped.
// The default comes from the template when it declares one. A story has a beat
// floor of eight; funding that from the shared 45-second budget would demand
// eight beats of barely more than the ten-word minimum, so the template that
// sets the floor also sets the runtime that can pay for it.
func (s SnippetSpec) ResolvedTargetSec() int {
	t := s.TargetSec
	if t == 0 {
		t = defaultSnippetTargetSec
		if tpl, ok := SnippetTemplates[s.Template]; ok && tpl.DefaultTargetSec > 0 {
			t = tpl.DefaultTargetSec
		}
	}
	return min(max(t, minSnippetTargetSec), maxSnippetTargetSec)
}

// ResolvedCodeLanguage returns the language for code-bearing templates.
func (s SnippetSpec) ResolvedCodeLanguage() string {
	if s.CodeLanguage == "" {
		return "python"
	}
	return s.CodeLanguage
}

// Validate checks a request before anything is written to disk.
func (s SnippetSpec) Validate() error {
	if strings.TrimSpace(s.Prompt) == "" {
		return fmt.Errorf("prompt is required — say what the snippet should teach")
	}
	if s.Template == "" {
		return fmt.Errorf("template is required (templates: %s)", strings.Join(SnippetTemplateNames(), ", "))
	}
	if _, ok := SnippetTemplates[s.Template]; !ok {
		return fmt.Errorf("unknown template %q (templates: %s)", s.Template, strings.Join(SnippetTemplateNames(), ", "))
	}
	if s.TargetSec != 0 && (s.TargetSec < minSnippetTargetSec || s.TargetSec > maxSnippetTargetSec) {
		return fmt.Errorf("target_sec %d is out of range (%d-%d)", s.TargetSec, minSnippetTargetSec, maxSnippetTargetSec)
	}
	// Caught here rather than three correction rounds later: a runtime below a
	// template's floor is not a clip that plans badly, it is one whose rules
	// contradict each other, and the model cannot be told anything that helps.
	// The same failure used to be reachable through the *shared* rules too — a
	// 20-second clip was told to write three beats of forty words against an
	// eighty-nine-word ceiling — which is why the beat count is derived from
	// the budget now rather than fixed (beatBounds).
	if tpl, ok := SnippetTemplates[s.Template]; ok && s.TargetSec != 0 &&
		tpl.MinTargetSec > 0 && s.TargetSec < tpl.MinTargetSec {
		return fmt.Errorf("the %s template needs at least %d seconds (asked for %d) — it is built from %d or more beats and a shorter clip cannot fund them",
			s.Template, tpl.MinTargetSec, s.TargetSec, minStoryBeats)
	}
	return nil
}

// SnippetPlan is snippet-plan.json: the model's complete design for the clip,
// narration and visuals together.
type SnippetPlan struct {
	Template string `json:"template"`
	Title    string `json:"title"`
	// Subtitle is the one-line promise shown under the title on the opening
	// card ("" = no card).
	Subtitle string        `json:"subtitle,omitempty"`
	Beats    []SnippetBeat `json:"beats"`
	// Chart is the data template's dataset. It sits on the plan rather than on
	// a beat because a data clip is one chart read several ways — the beats
	// only move the emphasis around it. Every other template's visual state is
	// per-beat; this is the one case where it genuinely is not.
	Chart *ChartSpec `json:"chart,omitempty"`
	// Project is the workspace template's file set. Like Chart it sits on the
	// plan rather than on a beat, and for a stronger reason: the files are one
	// program, and a program is not a property of a moment in the clip.
	Project *ProjectSpec `json:"project,omitempty"`
	// Timeline is the timeline template's run of milestones. On the plan for the
	// same reason as Quiz: the sequence is the subject of the clip.
	Timeline *TimelineSpec `json:"timeline,omitempty"`
	// Anatomy is the anatomy template's artefact and its labelled parts. On the
	// plan for the same reason as Quiz: the thing being taken apart is the
	// subject of the clip.
	Anatomy *AnatomySpec `json:"anatomy,omitempty"`
	// Compare is the compare template's pair. On the plan for the same reason
	// as Quiz: the two things being weighed are the subject of the clip.
	Compare *CompareSpec `json:"compare,omitempty"`
	// Quiz is the quiz template's question. On the plan for the same reason as
	// Chart: it is the subject of the clip, not a property of one beat in it.
	Quiz *QuizSpec `json:"quiz,omitempty"`
	// Canvas is the canvas template's automation — the whole chain of cards and
	// the payload that runs them. On the plan for the same reason as Timeline:
	// the workflow is the subject of the clip and the beats only walk it.
	Canvas *CanvasSpec `json:"canvas,omitempty"`
	// Loop is the promptloop template's goal — the one thing that does not
	// change across the clip, which is exactly why it is worth pinning.
	Loop *PromptLoopSpec `json:"loop,omitempty"`
	// Mockup is the mockup template's screen. On the plan for the same reason as
	// Canvas: the page is the subject of the clip and the beats walk down it.
	Mockup *MockupSpec `json:"mockup,omitempty"`
	// Stack is the stack template's tiers. On the plan for the same reason as
	// Mockup: the arrangement is the subject and the beats only walk it.
	Stack *StackSpec `json:"stack,omitempty"`

	// targetWords is the narration budget this plan was asked for. Not part of
	// the model's reply — the planner stashes it after decoding so the shared
	// validators can size the beat count against the same budget the prompt
	// quoted. Zero (a plan built by hand in a test) falls back to the fixed
	// range; see beatBounds.
	targetWords int
}

// ProjectSpec is a small multi-file program: what the workspace template
// writes on screen, and what actually gets executed.
type ProjectSpec struct {
	Files []ProjectFile `json:"files"`
	// Entry is the file the terminal runs.
	Entry string `json:"entry"`
	// Command overrides what the terminal shows itself typing. Empty renders
	// as `python3 <entry>`, which is what actually runs.
	Command string `json:"command,omitempty"`
	// Output is what the interpreter really printed. NOT written by the model
	// — the plan stage executes the file set and fills this in, and a program
	// that fails to run is sent back for correction rather than shipped with
	// invented output. It is the single-file template's verify gate, applied
	// to a program that spans files and therefore cannot go through verify.
	Output string `json:"output,omitempty"`
}

// ProjectFile is one file of that program.
type ProjectFile struct {
	Path string `json:"path"`
	Code string `json:"code"`
}

// FileMap is the file set in the shape the runner takes.
func (p *ProjectSpec) FileMap() map[string]string {
	out := make(map[string]string, len(p.Files))
	for _, f := range p.Files {
		out[f.Path] = f.Code
	}
	return out
}

// ChartSpec is a dataset and how to draw it.
type ChartSpec struct {
	// Kind is one of chartKindVocab; anything else degrades to bars.
	Kind string `json:"kind"`
	// Unit is appended to every value shown ("%", "ms", "M"). Optional.
	Unit string `json:"unit,omitempty"`
	// Points are the data. For a map, each label is a country name.
	Points []DataPoint `json:"points"`
	// Series names the dimensions each point carries, for the kinds that need
	// more than one number per label.
	//
	// It means two different things depending on the kind, and that is
	// deliberate rather than sloppy: for stacked and grouped bars these are the
	// *parts* of each label ("Cache", "Database", "Render"), and for a scatter
	// they are the two *axes* ("Team size", "Deploys per week"). Both are "one
	// number per named dimension", which is why one field carries both instead
	// of a second one that is empty three quarters of the time.
	//
	// Empty for every other kind, where a point is a single number.
	Series []string `json:"series,omitempty"`
}

// DataPoint is one labelled number — or, when the chart declares Series, one
// labelled row of them.
type DataPoint struct {
	Label string `json:"label"`
	// Value is the number, for the kinds that take one.
	Value float64 `json:"value,omitempty"`
	// Values is one number per entry in ChartSpec.Series, in the same order.
	Values []float64 `json:"values,omitempty"`
}

// total is the point's magnitude however it was written: a stacked bar's height
// is the sum of its parts, and a plain bar's is its own value.
func (p DataPoint) total() float64 {
	if len(p.Values) == 0 {
		return p.Value
	}
	sum := 0.0
	for _, v := range p.Values {
		sum += v
	}
	return sum
}

// SnippetBeat is one narrated step of a snippet.
//
// Every beat carries narration — that is what makes it a beat, and what the
// aligner times the visuals against. The remaining fields are the union of
// what the templates need; each template uses its own subset and rejects the
// rest in Validate, so the model can never smuggle a whiteboard field into a
// VS Code clip and have it silently ignored.
type SnippetBeat struct {
	// ID is a slug, unique within the plan; it becomes the script section id
	// the aligner reports timings for.
	ID string `json:"id"`
	// Heading is the short on-screen label for this beat (2-5 words).
	Heading string `json:"heading"`
	// Narration is what the voice says during this beat.
	Narration string `json:"narration"`

	// --- vscode template ---
	// Code is the complete buffer contents as of this beat: not a diff, the
	// whole file. The first beat that carries code types itself in; later
	// ones swap the buffer and flash the lines that changed.
	Code string `json:"code,omitempty"`
	// Run executes the file in the integrated terminal during this beat. The
	// output shown is whatever the interpreter really printed (verify stage),
	// never what the model imagined.
	Run bool `json:"run,omitempty"`

	// --- whiteboard template ---
	// Sketch is what this beat adds to the board. The board accumulates, so
	// each item is drawn once and stays for the rest of the clip.
	Sketch []SketchItem `json:"sketch,omitempty"`

	// --- flow template ---
	// Nodes are the boxes this beat adds to the diagram, each naming the nodes
	// that feed it. Like the board, the diagram accumulates.
	Nodes []FlowNode `json:"nodes,omitempty"`
	// Focus lists node ids to light up while this beat is spoken; everything
	// else dims and its traffic stops. This is how one diagram carries several
	// explanations without being redrawn.
	Focus []string `json:"focus,omitempty"`

	// --- illustration template ---
	// Art is this beat's shot: the figure beside the headline, which word of
	// the headline carries the emphasis, and the line under it. Unlike the
	// board and the diagram, nothing here accumulates — one beat is one shot.
	Art *ArtBeat `json:"art,omitempty"`

	// --- cast template ---
	// Cast is this beat's direction for the character: what they do and how
	// they feel about what is being said. Like Art, one beat is one shot.
	Cast *CastBeat `json:"cast,omitempty"`

	// --- story template ---
	// Shot is how this beat is staged and shot. Produced by the director
	// stage, not by the writer — see snippet_story.go.
	Shot *ShotBeat `json:"shot,omitempty"`

	// --- data template ---
	// Data is what this beat points at in the clip's one chart.
	Data *DataBeat `json:"data,omitempty"`

	// --- workspace template ---
	// Work is which file this beat is in, how much of it exists yet, and
	// where the camera is looking.
	Work *WorkspaceBeat `json:"work,omitempty"`

	// --- timeline template ---
	// Timeline says which milestone this beat is standing on, or that it is
	// showing the finished run.
	Timeline *TimelineBeat `json:"timeline,omitempty"`

	// --- anatomy template ---
	// Anatomy says which piece of the artefact this beat is on, or that it is
	// showing the whole thing.
	Anatomy *AnatomyBeat `json:"anatomy,omitempty"`

	// --- compare template ---
	// Compare is what this beat does to the two columns: introduce one,
	// introduce the other, light both, or deliver the verdict.
	Compare *CompareBeat `json:"compare,omitempty"`

	// --- quiz template ---
	// Quiz is what this beat does to the question on screen: pose it, hold
	// while the viewer thinks, reveal the answer, or explain one option.
	Quiz *QuizBeat `json:"quiz,omitempty"`

	// --- canvas template ---
	// Canvas says which card of the automation this beat is standing on, or
	// that it is running the payload down the whole chain.
	Canvas *CanvasBeat `json:"canvas,omitempty"`

	// --- promptloop template ---
	// Loop is one turn of the conversation: who is speaking, what is in the
	// bubble, and — when it is the model's turn — how the attempt came out.
	Loop *PromptLoopBeat `json:"loop,omitempty"`

	// --- mockup template ---
	// Mockup says which block of the page this beat is adding, or that it is
	// showing the finished screen.
	Mockup *MockupBeat `json:"mockup,omitempty"`

	// --- stack template ---
	// Stack says which tier this beat is standing on, or that it is showing the
	// whole stack.
	Stack *StackBeat `json:"stack,omitempty"`
}

// QuizSpec is the clip's one question.
//
// It sits on the plan rather than on a beat for the same reason the data
// template's chart does: the question is the *subject* of the clip, not a
// property of one moment in it. A question per beat would be a quiz nobody has
// time to attempt, which is the opposite of what this template is for — the
// whole value of retrieval practice is the gap between being asked and being
// told, and that gap needs the clip's whole middle.
type QuizSpec struct {
	// Question is what the viewer is asked. One sentence.
	Question string `json:"question"`
	// Options are the answers offered, in the order they appear.
	Options []string `json:"options"`
	// Answer indexes the correct option.
	Answer int `json:"answer"`
	// Why explains each option, index-aligned with Options: why the right one
	// is right, and — much more usefully — why each wrong one is tempting. A
	// distractor nobody would pick teaches nothing; saying what makes it
	// plausible is where the learning actually is.
	Why []string `json:"why,omitempty"`
}

// QuizBeat is one move in the question's life on screen.
type QuizBeat struct {
	// Show is the action: ask | think | reveal | explain.
	Show string `json:"show"`
	// Option is which answer an `explain` beat is talking about.
	Option int `json:"option,omitempty"`
}

// WorkspaceBeat is one moment of a multi-file walkthrough.
//
// Note what it does *not* carry: the code. The single-file template repeats a
// whole buffer per step, which is fine for eight lines and absurd for three
// files — so the plan declares each file once and a beat says how far into it
// the typing has got. The renderer reveals lines 1..Through, which is also
// what makes the reveal continuous rather than a cut between two buffers.
type WorkspaceBeat struct {
	// File is the path shown during this beat; it must be one of the
	// project's files.
	File string `json:"file"`
	// Through is how many lines of that file exist by the end of this beat.
	// 0 means the whole file — the right answer for any beat after it is
	// written, and for a file the clip never types out.
	Through int `json:"through,omitempty"`
	// Focus is where the camera looks: one of workspaceFocusVocab. The model
	// picks the *subject*, never a coordinate — the renderer owns every
	// number, the same bargain the story template's staging makes.
	Focus string `json:"focus,omitempty"`
	// Run opens the terminal and runs the entry file during this beat.
	Run bool `json:"run,omitempty"`
	// Caption is the supporting line under the editor. Optional.
	Caption string `json:"caption,omitempty"`
}

// DataBeat is one beat's reading of the chart: what it points at, and the line
// under it while it does.
type DataBeat struct {
	// Highlight names data points to light up while this beat is spoken;
	// everything else dims. Empty means the whole chart sits neutral, which is
	// the right choice for an opening or closing beat.
	Highlight []string `json:"highlight,omitempty"`
	// Caption is the supporting line shown during this beat. Optional.
	Caption string `json:"caption,omitempty"`
}

// ShotBeat stages one beat of a story: where things stand, how the camera
// moves, and what the character is doing.
//
// The model chooses only from closed vocabularies. Every coordinate belongs to
// the renderer, which is what stops a director putting a character on top of
// its own caption.
type ShotBeat struct {
	// BeatID ties a shot to its written beat. The director is handed the
	// script and answers per-id rather than positionally, so a reordered or
	// short reply is caught instead of being silently mis-applied.
	BeatID string `json:"beat_id"`
	// Staging is the arrangement: hero | duo | object | pair | empty.
	Staging string `json:"staging"`
	// Camera is the move: hold | push | pull | pan | rise | drift.
	Camera string `json:"camera"`
	// Pose and Expression direct the character, on the stagings that have one.
	Pose       string `json:"pose,omitempty"`
	Expression string `json:"expression,omitempty"`
	// Prop, and PropB for a "pair" comparison, name figures from the
	// illustration vocabulary.
	Prop  string `json:"prop,omitempty"`
	PropB string `json:"prop_b,omitempty"`
	// Caption is the supporting line under the headline. Optional.
	Caption string `json:"caption,omitempty"`
}

// CastBeat directs the character for one shot.
//
// The beat's Heading is the headline, as in ArtBeat — the on-screen phrase and
// the beat's label are the same thing rather than two things to keep in step.
type CastBeat struct {
	// Pose is a name from the closed pose vocabulary (see CastPoseNames);
	// anything else degrades to "idle".
	Pose string `json:"pose"`
	// Expression is a name from the closed expression vocabulary; anything
	// else degrades to "neutral".
	Expression string `json:"expression,omitempty"`
	// Prop optionally puts one of the illustration figures above the
	// character — the thing they are talking about.
	Prop string `json:"prop,omitempty"`
	// Caption is the supporting line under the headline. Optional.
	Caption string `json:"caption,omitempty"`
}

// ArtBeat is one kinetic-typography shot: a figure, and the phrasing that
// lands next to it.
//
// The beat's Heading is the headline, so it is not repeated here — the whole
// point of the template is that the on-screen phrase and the beat's label are
// the same thing rather than two things to keep in step.
type ArtBeat struct {
	// Figure is a name from the closed figure vocabulary (see ArtFigureNames);
	// anything else degrades to the neutral "spark".
	Figure string `json:"figure"`
	// Emphasis is the word or short phrase inside the heading that gets the
	// accent and the marker stroke. It must actually occur in the heading.
	Emphasis string `json:"emphasis,omitempty"`
	// Caption is the supporting line under the headline: one sentence that
	// says the thing the headline only gestures at. Optional.
	Caption string `json:"caption,omitempty"`
}

// SketchItem is one thing drawn on the whiteboard: a labelled box with an
// icon, optionally reached by an arrow from an idea already on the board.
type SketchItem struct {
	// Label is the box's caption — a noun phrase of at most a few words.
	Label string `json:"label"`
	// Icon is a name from the closed icon vocabulary (see PointIconNames);
	// anything else degrades to a neutral dot.
	Icon string `json:"icon"`
	// LinkFrom names an earlier item's label. An arrow is drawn from that box
	// to this one, which is how the picture becomes a diagram rather than a
	// grid of unrelated boxes.
	LinkFrom string `json:"link_from,omitempty"`
	// Shape is how this item is drawn — see sketchShapes. Everything used to
	// be a rounded box, which is the right default and the wrong only option:
	// a board where "the internet" and "the retry queue" are the same rectangle
	// is a board that has stopped distinguishing between a thing with edges and
	// a thing without any.
	Shape string `json:"shape,omitempty"`
}

// Validate checks a plan's structure. Template-specific rules live in the
// template's own Validate.
func (p *SnippetPlan) Validate() error {
	if err := p.validateShape(); err != nil {
		return err
	}
	if tpl, ok := SnippetTemplates[p.Template]; ok && tpl.Validate != nil {
		return tpl.Validate(p)
	}
	return nil
}

// validateShape is what has to be true of any plan for the rest of the pipeline
// to work at all: a title, beats, and one unique id per beat for the aligner to
// report timings against.
//
// It is separated from the template's rules because the two are checked at
// different moments and for different reasons. The template's rules are a
// standard the model is held to while it still has rounds left to meet it; this
// is a fact about the file. Re-applying the editorial rules every time
// snippet-plan.json is read would mean a plan that shipped deliberately loose —
// see salvageSnippetPlan — could not be loaded by the stage that renders it,
// and the run would die one stage after the one that decided to continue.
func (p *SnippetPlan) validateShape() error {
	if strings.TrimSpace(p.Title) == "" {
		return fmt.Errorf("title is empty")
	}
	if len(p.Beats) == 0 {
		return fmt.Errorf("plan has no beats")
	}
	seen := map[string]bool{}
	for i, b := range p.Beats {
		if strings.TrimSpace(b.ID) == "" {
			return fmt.Errorf("beat %d has an empty id", i)
		}
		if seen[b.ID] {
			return fmt.Errorf("duplicate beat id %q", b.ID)
		}
		seen[b.ID] = true
		if strings.TrimSpace(b.Narration) == "" {
			return fmt.Errorf("beat %q has empty narration", b.ID)
		}
		if strings.TrimSpace(b.Heading) == "" {
			return fmt.Errorf("beat %q has an empty heading", b.ID)
		}
	}
	return nil
}

// Script converts the plan's narration into the ordinary script.json the
// audio, align, and chapters stages already consume. Duration estimates are
// derived from word count at the configured pace, which is all the audio
// stage uses them for.
func (p *SnippetPlan) Script(paceWPM int) *Script {
	if paceWPM <= 0 {
		paceWPM = 150
	}
	script := &Script{Title: p.Title}
	for _, b := range p.Beats {
		words := len(strings.Fields(b.Narration))
		est := max(1, int(float64(words)/float64(paceWPM)*60+0.5))
		script.Sections = append(script.Sections, Section{
			ID:             b.ID,
			Narration:      b.Narration,
			DurationEstSec: est,
		})
	}
	return script
}

// Markdown renders the plan as an ordinary lesson.md.
//
// Nothing downstream needs to know a snippet was not hand-written: verify
// finds its code blocks here, and the chapters/transcript stages read the same
// headings a lesson would have.
func (p *SnippetPlan) Markdown(spec SnippetSpec) (string, error) {
	fm := project.FrontMatter{
		Title:    p.Title,
		Style:    spec.Config.Style,
		Branding: spec.Config.Branding,
		Pipeline: spec.Config.Pipeline,
	}
	fmData, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("encoding snippet front-matter: %w", err)
	}
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmData)
	sb.WriteString("---\n\n")
	if p.Subtitle != "" {
		sb.WriteString(p.Subtitle + "\n\n")
	}
	lang := spec.ResolvedCodeLanguage()
	// One code block per distinct buffer state, so verify executes exactly
	// the states the video shows — including the intermediate ones.
	var lastCode string
	for _, b := range p.Beats {
		sb.WriteString("## " + b.Heading + "\n\n")
		sb.WriteString(b.Narration + "\n\n")
		if b.Code != "" && b.Code != lastCode {
			sb.WriteString("```" + lang + "\n" + strings.TrimRight(b.Code, "\n") + "\n```\n\n")
			lastCode = b.Code
		}
	}
	return sb.String(), nil
}

// LoadSnippetSpec reads a snippet directory's snippet.yaml.
func LoadSnippetSpec(dir string) (*SnippetSpec, error) {
	path := filepath.Join(dir, SnippetFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var spec SnippetSpec
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)
	if err := dec.Decode(&spec); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if spec.ID == "" {
		spec.ID = filepath.Base(dir)
	}
	return &spec, nil
}

// IsSnippet reports whether a lesson directory is a snippet.
func IsSnippet(l *project.Lesson) bool {
	_, err := os.Stat(filepath.Join(l.Dir, SnippetFileName))
	return err == nil
}

// LoadSnippetPlan reads generated/snippet-plan.json.
func LoadSnippetPlan(l *project.Lesson) (*SnippetPlan, error) {
	path := filepath.Join(l.GeneratedDir(), SnippetPlanFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no %s yet — the plan stage must run first", SnippetPlanFileName)
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var plan SnippetPlan
	// Read back the way it was written: leniently, and judged only on the shape
	// the rest of the pipeline depends on. The plan stage has already applied
	// the template's standards, with the model in the room to answer for them —
	// re-litigating them here can only reject a file that is already the best
	// answer anyone is going to get.
	if err := parseJSONLenient(string(data), &plan, nil); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", SnippetPlanFileName, err)
	}
	normalizeSnippetPlan(&plan)
	if err := plan.validateShape(); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", SnippetPlanFileName, err)
	}
	return &plan, nil
}

// snippetsCourseYAML is the manifest for the synthetic snippets course. It is
// video-only by definition: a snippet has no quiz, no exercises, no web page.
const snippetsCourseYAML = `name: Snippets
slug: snippets
description: Short standalone clips generated from a prompt and a template.
style:
  tone: crisp, confident, direct
  audience: developers and learners watching a short clip
  language: en
  # Measured: Kokoro af_heart delivers full snippet narration at ~174 wpm at
  # speed 1.0, so the target is set where the voice already lives and the
  # first render lands in pace instead of costing an auto-pace round trip.
  pace_wpm: 175
  captions: "on"
pipeline:
  video_only: true
  # Snippet planning runs on gpt-4o-mini rather than the default Groq model.
  # Groq's free tier is 100k tokens per day and one snippet plan is ~14k, so a
  # handful of clips — or one clip whose plan needs correction rounds —
  # exhausts the day and every later plan fails on a 429 that looks like a bug
  # in the planner. gpt-4o-mini plans these as well and is not rationed.
  llm_content: openai/gpt-4o-mini
`

// EnsureSnippetsCourse creates (or opens) the synthetic snippets course under
// root and returns it.
func EnsureSnippetsCourse(root string) (*project.Course, error) {
	dir := filepath.Join(root, SnippetsRoot)
	manifest := filepath.Join(dir, project.CourseFileName)
	if _, err := os.Stat(manifest); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Join(dir, "lessons"), 0o755); err != nil {
			return nil, fmt.Errorf("creating snippets course: %w", err)
		}
		if err := writeFileAtomic(manifest, []byte(snippetsCourseYAML)); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("checking %s: %w", manifest, err)
	}
	return project.LoadCourse(dir)
}

// CreateSnippet writes a new snippet directory and returns its lesson handle.
//
// The stub lesson.md exists only so the directory is a valid lesson from the
// moment it is created — the plan stage overwrites it with the real thing.
func CreateSnippet(root string, spec SnippetSpec) (*project.Course, *project.Lesson, error) {
	if err := spec.Validate(); err != nil {
		return nil, nil, err
	}
	course, err := EnsureSnippetsCourse(root)
	if err != nil {
		return nil, nil, err
	}
	if spec.ID == "" {
		spec.ID, err = uniqueSnippetID(course.Dir, spec)
		if err != nil {
			return nil, nil, err
		}
	}
	if spec.CreatedAt.IsZero() {
		spec.CreatedAt = time.Now().UTC().Truncate(time.Second)
	}

	dir := filepath.Join(course.Dir, "lessons", spec.ID)
	if _, err := os.Stat(dir); err == nil {
		return nil, nil, fmt.Errorf("snippet %q already exists at %s", spec.ID, dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("creating %s: %w", dir, err)
	}

	specData, err := yaml.Marshal(spec)
	if err != nil {
		return nil, nil, fmt.Errorf("encoding %s: %w", SnippetFileName, err)
	}
	if err := writeFileAtomic(filepath.Join(dir, SnippetFileName), specData); err != nil {
		return nil, nil, err
	}

	title := spec.Title
	if title == "" {
		title = snippetStubTitle(spec.Prompt)
	}
	stub := fmt.Sprintf("---\ntitle: %q\n---\n\n## Pending\n\n%s\n",
		title, "This snippet has not been planned yet — run the plan stage.")
	if err := writeFileAtomic(filepath.Join(dir, project.LessonFileName), []byte(stub)); err != nil {
		return nil, nil, err
	}

	l, err := project.LoadLesson(dir)
	if err != nil {
		return nil, nil, err
	}
	return course, l, nil
}

// ListSnippets returns every snippet in the project, newest request first.
func ListSnippets(root string) ([]*project.Lesson, error) {
	dir := filepath.Join(root, SnippetsRoot, "lessons")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}
	var out []*project.Lesson
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		l, err := project.LoadLesson(filepath.Join(dir, e.Name()))
		if err != nil {
			continue // a half-written snippet should not break the list
		}
		if IsSnippet(l) {
			out = append(out, l)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

// FindSnippet resolves a snippet by id.
func FindSnippet(root, id string) (*project.Course, *project.Lesson, error) {
	course, err := EnsureSnippetsCourse(root)
	if err != nil {
		return nil, nil, err
	}
	dir := filepath.Join(course.Dir, "lessons", id)
	l, err := project.LoadLesson(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("no snippet %q (looked in %s)", id, dir)
	}
	return course, l, nil
}

// uniqueSnippetID derives a readable directory name from the prompt, with a
// numeric suffix if that name is taken.
func uniqueSnippetID(courseDir string, spec SnippetSpec) (string, error) {
	base := slugify(snippetStubTitle(spec.Prompt))
	if base == "" {
		base = spec.Template
	}
	if len(base) > 48 {
		base = strings.Trim(base[:48], "-")
	}
	for i := 0; i < 200; i++ {
		id := base
		if i > 0 {
			id = fmt.Sprintf("%s-%d", base, i+1)
		}
		if _, err := os.Stat(filepath.Join(courseDir, "lessons", id)); os.IsNotExist(err) {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not find a free snippet id for %q", base)
}

// snippetStubTitle makes a provisional title from the prompt: the first
// sentence, trimmed to a headline length.
func snippetStubTitle(prompt string) string {
	s := strings.TrimSpace(prompt)
	if i := strings.IndexAny(s, ".!?\n"); i > 0 {
		s = s[:i]
	}
	words := strings.Fields(s)
	if len(words) > 9 {
		words = words[:9]
	}
	s = strings.Join(words, " ")
	if s == "" {
		return "Snippet"
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// runPlanStage is the snippet pipeline's first stage: prompt + template →
// snippet-plan.json, and from it the script.json and lesson.md the rest of the
// pipeline expects.
func runPlanStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	spec, err := LoadSnippetSpec(l.Dir)
	if err != nil {
		return err
	}
	tpl, ok := SnippetTemplates[spec.Template]
	if !ok {
		return fmt.Errorf("unknown template %q (templates: %s)", spec.Template, strings.Join(SnippetTemplateNames(), ", "))
	}
	if e.Router == nil {
		return fmt.Errorf("planning a snippet needs an LLM — set GROQ_API_KEY (or an OpenAI-compatible provider) and retry")
	}

	fmt.Fprintf(e.out(), "  → plan      %s template, ~%ds target (%s)...\n",
		tpl.Name, spec.ResolvedTargetSec(), cfg.Pipeline.LLMContent)

	planner := tpl.Plan
	if planner == nil {
		planner = planSnippetDefault
	}
	plan, err := planner(ctx, e, *spec, cfg)
	if err != nil {
		return err
	}
	plan.Template = spec.Template
	if spec.Title != "" {
		plan.Title = spec.Title
	}

	if err := writeJSON(filepath.Join(l.GeneratedDir(), SnippetPlanFileName), plan); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), ScriptFileName), plan.Script(cfg.Style.PaceWPM)); err != nil {
		return err
	}
	md, err := plan.Markdown(*spec)
	if err != nil {
		return err
	}
	if err := writeFileAtomic(l.SourcePath(), []byte(md)); err != nil {
		return err
	}
	// The caller loaded this lesson before the plan existed, so its in-memory
	// body is still the stub. Later stages in the same run read the struct,
	// not the file — verify would find no code blocks — so reload in place.
	reloaded, err := project.LoadLesson(l.Dir)
	if err != nil {
		return fmt.Errorf("planned snippet does not load (bug): %w", err)
	}
	*l = *reloaded

	words := 0
	for _, b := range plan.Beats {
		words += len(strings.Fields(b.Narration))
	}
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	fmt.Fprintf(e.out(), "    %q — %d beats, %d words (~%ds at %d wpm)\n",
		plan.Title, len(plan.Beats), words, words*60/pace, pace)
	return nil
}

// SnippetStages returns the stage list a snippet actually runs: the snippet
// pipeline, minus the verify stage for templates that never show code.
//
// Exported because the studio drives runs stage by stage (so its SSE stream can
// report each one) and therefore needs the same list the CLI walks.
func SnippetStages(l *project.Lesson) ([]string, error) {
	spec, err := LoadSnippetSpec(l.Dir)
	if err != nil {
		return nil, err
	}
	tpl, ok := SnippetTemplates[spec.Template]
	if !ok {
		return nil, fmt.Errorf("snippet %s: unknown template %q (templates: %s)",
			l.ID, spec.Template, strings.Join(SnippetTemplateNames(), ", "))
	}
	stages := slices.Clone(project.SnippetStageOrder)
	if !tpl.NeedsCode {
		stages = slices.DeleteFunc(stages, func(s string) bool { return s == project.StageVerify })
	}
	return stages, nil
}

// RunSnippet executes the snippet pipeline for one snippet, skipping
// up-to-date stages exactly like RunLesson.
func (e *Env) RunSnippet(ctx context.Context, course *project.Course, l *project.Lesson, opts RunOptions) error {
	if !IsSnippet(l) {
		return fmt.Errorf("%s is not a snippet (no %s)", l.Dir, SnippetFileName)
	}
	spec, err := LoadSnippetSpec(l.Dir)
	if err != nil {
		return err
	}
	stages, err := SnippetStages(l)
	if err != nil {
		return err
	}
	cfg := config.Resolve(course.Config, l.FrontMatter.Overrides(), spec.Config)
	fmt.Fprintf(e.out(), "%s — %s template\n", l.ID, spec.Template)
	return e.runStages(ctx, course, l, cfg, stages, opts)
}
