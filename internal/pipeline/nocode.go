package pipeline

// No-code pieces: the fourth surface, and the one where nothing is drawn.
//
//	courses   a document first, a video second — chapters, quizzes, a site
//	snippets  one prompt, one template, one clip
//	reels     several templates cut onto one timeline
//	nocode    several segments, every one of them backed by something real
//
// == The rule that earns it ==
//
// A reel about no-code tools is not this. A reel will happily draw a person
// gesturing at a chart of numbers nobody measured, because a reel's contract is
// "each segment names material it can be filled with" and prose counts as
// material. That is right for a reel and wrong here.
//
// **Every segment of a no-code piece must be backed by evidence that exists on
// disk**: a real capture — a browser driven against the product, a terminal
// session, an operator in a native app — or a fact with provenance from the
// substance sheet. A segment that can name neither is refused before a single
// planning call is spent.
//
// That single rule is the whole surface. It is why the template catalog here is
// a subset (§noCodeTemplates), why the capture stage runs first, and why the
// pieces this produces are worth more than the same words over stock animation:
// a viewer who has never seen Lovable's prompt box will not believe a rectangle
// we drew that says "Lovable", and will believe a recording of it.
//
// == What is deliberately excluded ==
//
// The character templates — `cast`, `story`, `illustration` — are the ones that
// put a drawn figure on screen and let it stand in for the thing being
// discussed. They are good templates and they are exactly wrong for a surface
// whose entire claim is "this really happened". Excluding them is not a
// judgement about the artwork; it is that a face is the fastest way to fill a
// frame without evidence, and this surface exists to refuse that.

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
	"gopkg.in/yaml.v3"
)

const (
	// NoCodeFileName is the spec in a no-code piece's directory.
	NoCodeFileName = "nocode.yaml"
	// NoCodeCourseSlug is the slug of the synthetic course pieces live in. It
	// must match the `slug:` in noCodeCourseYAML — it is what artifact URLs are
	// built from, so a piece's finished video is served at
	// /artifacts/nocode/<id>/final.mp4 like every other lesson's.
	NoCodeCourseSlug = "nocode"
)

// NoCodeRoot is where no-code pieces live, relative to the project root.
var NoCodeRoot = filepath.Join(".coursesmith", NoCodeCourseSlug)

// EvidenceKind is what backs a segment.
type EvidenceKind string

const (
	// EvidenceCapture is a real recording: a browser take, a terminal session,
	// an operator in a native app. The strongest kind, and the reason this
	// surface exists.
	EvidenceCapture EvidenceKind = "capture"
	// EvidenceFact is a claim from the substance sheet carrying its own
	// provenance — sourced, given or derived. Weaker than a recording and still
	// checkable, which is the bar.
	EvidenceFact EvidenceKind = "fact"
)

// NoCodeSegment is one part of the piece.
type NoCodeSegment struct {
	// ID is stable across edits, so an edit stays addressed to the same segment
	// when a neighbour moves.
	ID string `yaml:"id"`
	// Template names a registered template, restricted to the no-code catalog.
	Template string `yaml:"template"`
	// Prompt is what this segment covers, in the creator's words.
	Prompt string `yaml:"prompt"`
	// Evidence is what backs it — the field that makes this surface what it is.
	Evidence NoCodeEvidence `yaml:"evidence"`
	// Skip drops a segment from the cut without deleting the prompt that
	// produced it. The commonest edit after a first watch is "lose that bit"
	// and the second is "put it back".
	Skip bool `yaml:"skip,omitempty"`
}

// NoCodeCapture is the recording a segment stands on — both what to shoot and,
// once shot, what the segment narrates.
//
// It declares the recording rather than referencing one, and that is
// deliberate. A course says what to record with a `[CAPTURE: tool=…]` marker in
// its lesson.md; a piece has no lesson.md until after planning, so an id
// pointing at something declared elsewhere would point at nothing. Saying it
// here keeps the segment and its evidence in one place, which is also where
// somebody editing the piece expects to find it.
type NoCodeCapture struct {
	// Tool is the allowlist key: a terminal tool, a web site, or a desktop app.
	Tool string `yaml:"tool"`
	// Of is what to record, in the creator's words. It becomes the capture's
	// description — for a terminal tool it is what the generated tape is asked
	// to show.
	Of string `yaml:"of,omitempty"`
	// Take names a file under the course's takes/ for the web and desktop
	// kinds, which cannot be driven from a description.
	Take string `yaml:"take,omitempty"`
	// Fixture seeds a terminal capture's scratch directory.
	Fixture string `yaml:"fixture,omitempty"`
}

// NoCodeEvidence is what a segment is allowed to stand on.
type NoCodeEvidence struct {
	Kind EvidenceKind `yaml:"kind"`
	// Capture is the recording, for the capture kind. Required for it.
	Capture *NoCodeCapture `yaml:"capture,omitempty"`
	// Facts are the claims this segment stands on, for the fact kind. Written
	// out rather than referenced by index, because a fact sheet is regenerated
	// and an index into it is a reference that silently retargets.
	Facts []string `yaml:"facts,omitempty"`
}

// NoCodeSpec is a course's `nocode.yaml`.
type NoCodeSpec struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title,omitempty"`
	// Brief is the whole piece in the creator's own words — the input casting
	// reads, kept even when a piece was authored by hand so it can be re-cast
	// later without the intent having been thrown away.
	Brief    string          `yaml:"brief,omitempty"`
	Segments []NoCodeSegment `yaml:"segments"`
	Config   config.Config   `yaml:",inline"`

	CreatedAt time.Time `yaml:"created_at,omitempty"`
}

// noCodeExcluded are the templates that put a drawn figure on screen.
//
// Not a judgement about the artwork — these are good templates. A face is
// simply the fastest way to fill a frame with no evidence behind it, and this
// surface exists to refuse that.
var noCodeExcluded = map[string]string{
	"cast":         "it puts a drawn presenter on screen; show the tool instead",
	"story":        "it stages drawn characters and props; show what really happened",
	"illustration": "it is a figure beside a headline, which is decoration standing in for evidence",
}

// NoCodeTemplateNames lists the templates a no-code segment may use, in
// catalog order.
func NoCodeTemplateNames() []string {
	var out []string
	for _, name := range SnippetTemplateNames() {
		if _, banned := noCodeExcluded[name]; banned {
			continue
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// CaptureToolNames lists everything recordable, across all three kinds.
func CaptureToolNames() []string {
	out := append([]string{}, captureToolKeys()...)
	out = append(out, captureSiteKeys()...)
	out = append(out, captureAppKeys()...)
	return out
}

// IsNoCode reports whether a lesson directory is a no-code piece.
func IsNoCode(l *project.Lesson) bool {
	_, err := os.Stat(filepath.Join(l.Dir, NoCodeFileName))
	return err == nil
}

// LoadNoCodeSpec reads nocode.yaml from a piece's directory.
func LoadNoCodeSpec(dir string) (*NoCodeSpec, error) {
	raw, err := os.ReadFile(filepath.Join(dir, NoCodeFileName))
	if err != nil {
		return nil, err
	}
	var spec NoCodeSpec
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", NoCodeFileName, err)
	}
	if spec.ID == "" {
		spec.ID = filepath.Base(dir)
	}
	spec.assignIDs()
	return &spec, nil
}

// assignIDs fills in missing segment ids once, from position. Renumbering on
// every edit would make every id a lie the moment a segment moved.
func (s *NoCodeSpec) assignIDs() {
	used := map[string]bool{}
	for i := range s.Segments {
		if s.Segments[i].ID != "" {
			used[s.Segments[i].ID] = true
		}
	}
	for i := range s.Segments {
		if s.Segments[i].ID != "" {
			continue
		}
		base := s.Segments[i].Template
		if base == "" {
			base = "segment"
		}
		id := fmt.Sprintf("%s-%d", base, i+1)
		for n := 2; used[id]; n++ {
			id = fmt.Sprintf("%s-%d-%d", base, i+1, n)
		}
		used[id] = true
		s.Segments[i].ID = id
	}
}

// Live returns the segments that are actually in the cut.
func (s *NoCodeSpec) Live() []NoCodeSegment {
	out := make([]NoCodeSegment, 0, len(s.Segments))
	for _, seg := range s.Segments {
		if !seg.Skip {
			out = append(out, seg)
		}
	}
	return out
}

// Validate enforces the rule that earns this surface.
//
// It runs before any planning call, because the whole point is to move "this
// segment has nothing real behind it" from late and expensive to immediate and
// free. A reel discovers a hollow segment after the caster has gone and a
// template's validator refuses the material; here it is a parse error.
func (s *NoCodeSpec) Validate() error {
	live := s.Live()
	if len(live) == 0 {
		return fmt.Errorf("a no-code piece has no segments in the cut — every one is skipped, or there are none")
	}
	allowed := NoCodeTemplateNames()
	for i, seg := range live {
		n := i + 1
		label := seg.ID
		if label == "" {
			label = fmt.Sprintf("segment %d", n)
		}
		if seg.Template == "" {
			return fmt.Errorf("%s names no template. Available: %s", label, strings.Join(allowed, ", "))
		}
		if why, banned := noCodeExcluded[seg.Template]; banned {
			return fmt.Errorf("%s uses %q, which a no-code piece may not: %s. This surface exists so that nothing on screen is drawn in place of evidence",
				label, seg.Template, why)
		}
		if _, ok := SnippetTemplates[seg.Template]; !ok {
			return fmt.Errorf("%s uses unknown template %q. Available: %s", label, seg.Template, strings.Join(allowed, ", "))
		}
		if strings.TrimSpace(seg.Prompt) == "" {
			return fmt.Errorf("%s has no prompt — say what it covers", label)
		}
		if err := seg.Evidence.validate(label); err != nil {
			return err
		}
	}
	return nil
}

// validate checks a segment stands on something that exists.
func (e NoCodeEvidence) validate(label string) error {
	switch e.Kind {
	case EvidenceCapture:
		if e.Capture == nil || strings.TrimSpace(e.Capture.Tool) == "" {
			return fmt.Errorf("%s is backed by a capture but names no tool. Write `evidence: {kind: capture, capture: {tool: claude, of: what to record}}`", label)
		}
		c := e.Capture
		_, isTool := captureTools[c.Tool]
		_, isSite := captureSites[c.Tool]
		_, isApp := captureApps[c.Tool]
		switch {
		case isTool:
			if strings.TrimSpace(c.Of) == "" {
				return fmt.Errorf("%s records %s but does not say what to record. Set `of:`", label, c.Tool)
			}
		case isSite, isApp:
			if strings.TrimSpace(c.Take) == "" {
				return fmt.Errorf("%s records %s, which needs `take:` — nobody can invent selectors for somebody else's page, or the beats for a native app", label, c.Tool)
			}
		default:
			return fmt.Errorf("%s records %q, which is not recordable. Terminal tools: %s. Web: %s. Desktop: %s",
				label, c.Tool, strings.Join(captureToolKeys(), ", "), strings.Join(captureSiteKeys(), ", "), strings.Join(captureAppKeys(), ", "))
		}
	case EvidenceFact:
		if len(e.Facts) == 0 {
			return fmt.Errorf("%s is backed by facts but lists none. Write the claims out, so a regenerated fact sheet cannot silently change what this segment stands on", label)
		}
		for _, f := range e.Facts {
			if strings.TrimSpace(f) == "" {
				return fmt.Errorf("%s lists an empty fact", label)
			}
		}
	case "":
		return fmt.Errorf("%s names no evidence. Every segment of a no-code piece stands on a real capture or a fact with provenance — that rule is the whole surface. Write `evidence: {kind: capture, capture: capture-1}` or `evidence: {kind: fact, facts: [...]}`", label)
	default:
		return fmt.Errorf("%s has evidence kind %q, which is not one of %s, %s", label, e.Kind, EvidenceCapture, EvidenceFact)
	}
	return nil
}

// CaptureIDs lists the captures this piece stands on, in order.
//
// A capture's id is its segment's id: one segment, one recording. Deriving it
// rather than asking for it removes the whole class of mistake where a segment
// references a capture that was renamed or never declared.
func (s *NoCodeSpec) CaptureIDs() []string {
	var out []string
	for _, seg := range s.Live() {
		if seg.Evidence.Kind == EvidenceCapture && seg.Evidence.Capture != nil {
			out = append(out, seg.ID)
		}
	}
	return out
}

// CaptureSpecs is what the capture stage records for this piece.
//
// A piece has no lesson.md when the capture stage runs — it is written by the
// plan stage, afterwards — so the markers a course path scans for do not exist
// here. These come from the segments instead, which is where the evidence was
// declared anyway.
func (s *NoCodeSpec) CaptureSpecs() []DemoSpec {
	var out []DemoSpec
	for _, seg := range s.Live() {
		if seg.Evidence.Kind != EvidenceCapture || seg.Evidence.Capture == nil {
			continue
		}
		c := seg.Evidence.Capture
		spec := DemoSpec{
			ID:          seg.ID,
			Description: c.Of,
			Tool:        c.Tool,
			Take:        c.Take,
			Fixture:     c.Fixture,
		}
		switch {
		case captureTools[c.Tool].Binary != "":
			spec.Kind = CaptureKindTool
		case captureSites[c.Tool].Origin != "":
			spec.Kind = CaptureKindWeb
		case captureApps[c.Tool].Bundle != "":
			spec.Kind = CaptureKindDesktop
		}
		out = append(out, spec)
	}
	return out
}

// NoCodeStageOrder is the pipeline for a no-code piece.
//
// It is the snippet path with **capture first**, and that order is the surface's
// whole argument made mechanical: the recordings have to exist before anything
// decides what to say about them. Substance follows, so the fact sheet can carry
// what the captures actually showed; then planning, which is the first point at
// which anything is written.
var NoCodeStageOrder = func() []string {
	out := []string{project.StageCapture}
	return append(out, project.SnippetStageOrder...)
}()

// noCodeCourseYAML is the synthetic course no-code pieces live in.
const noCodeCourseYAML = `name: "No-code pieces"
slug: nocode
description: >
  Short pieces where every segment is backed by a real recording or a fact with
  provenance. Nothing on screen is drawn in place of evidence.

style:
  voice: af_heart
  tone: plain-spoken and practical, never breathless about AI
  pace_wpm: 175
  voice_speed: 0.9
  audience: >
    people who are not developers — founders, product managers, designers —
    who have an idea and no way to build it
  language: en

branding:
  colors:
    primary: "#5b5bd6"
    accent: "#f5c26b"
    background: "#0b0d12"
  diagram_style: clean, flat, generous whitespace, few words per node

pipeline:
  llm_content: openai/gpt-4o-mini
  video_only: true
`

// EnsureNoCodeCourse creates (or loads) the synthetic course pieces live in.
func EnsureNoCodeCourse(root string) (*project.Course, error) {
	dir := filepath.Join(root, NoCodeRoot)
	manifest := filepath.Join(dir, project.CourseFileName)
	if _, err := os.Stat(manifest); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Join(dir, "lessons"), 0o755); err != nil {
			return nil, fmt.Errorf("creating the no-code course: %w", err)
		}
		// takes/ is made empty rather than on demand. A web or desktop segment
		// names a file here, and the recorder joins `take:` onto this exact
		// path — so a directory that does not exist until somebody guesses its
		// name is a required field with an unfindable answer.
		if err := os.MkdirAll(filepath.Join(dir, "takes"), 0o755); err != nil {
			return nil, fmt.Errorf("creating the no-code course: %w", err)
		}
		if err := writeFileAtomic(manifest, []byte(noCodeCourseYAML)); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("checking %s: %w", manifest, err)
	}
	return project.LoadCourse(dir)
}

// NoCodeStages is the stage list for one piece: the shared order, minus verify
// when nothing in it shows code.
func NoCodeStages(l *project.Lesson) ([]string, error) {
	spec, err := LoadNoCodeSpec(l.Dir)
	if err != nil {
		return nil, err
	}
	needsCode := false
	for _, seg := range spec.Live() {
		if tpl, ok := SnippetTemplates[seg.Template]; ok && tpl.NeedsCode {
			needsCode = true
		}
	}
	stages := slices.Clone(NoCodeStageOrder)
	if !needsCode {
		stages = slices.DeleteFunc(stages, func(s string) bool { return s == project.StageVerify })
	}
	return stages, nil
}

// CaptureDisplayNameFor is how a recordable is named to a reader, across all
// three registries. Exported so the studio does not carry its own copy of the
// vocabulary — a second copy is how the gallery ends up offering a tool the
// engine no longer records.
func CaptureDisplayNameFor(key string) string {
	if t, ok := captureTools[key]; ok {
		return t.Display
	}
	if s, ok := captureSites[key]; ok {
		return s.Display
	}
	if a, ok := captureApps[key]; ok {
		return a.Display
	}
	return key
}

// Recordable is one thing a capture segment may record.
type Recordable struct {
	Key     string `json:"key"`
	Display string `json:"display"`
	// Kind is terminal | web | desktop. It decides what the segment must
	// supply: a terminal tool is driven from a description, the other two need
	// a checked-in take.
	Kind string `json:"kind"`
	// NeedsTake is true when a description cannot drive it — nobody can invent
	// selectors for somebody else's page, or the beats for a native app.
	NeedsTake bool `json:"needs_take"`
}

// Recordables lists everything that can be captured, terminal first.
func Recordables() []Recordable {
	var out []Recordable
	for _, k := range captureToolKeys() {
		out = append(out, Recordable{Key: k, Display: captureTools[k].Display, Kind: "terminal"})
	}
	for _, k := range captureSiteKeys() {
		out = append(out, Recordable{Key: k, Display: captureSites[k].Display, Kind: "web", NeedsTake: true})
	}
	for _, k := range captureAppKeys() {
		out = append(out, Recordable{Key: k, Display: captureApps[k].Display, Kind: "desktop", NeedsTake: true})
	}
	return out
}

// AvailableTake is one take file a piece can name.
type AvailableTake struct {
	// Name is what a segment writes in `take:` — the file's base name with no
	// extension, because that is what the recorder joins onto takes/.
	Name string `json:"name"`
	// Tool is the site or app key the take drives, read out of the file. A take
	// only fits a segment that records the same thing, so the page filters by it
	// rather than offering every take for every tool.
	Tool string `json:"tool"`
	Kind string `json:"kind"`
	// Steps is how many beats or steps it works through — enough to tell two
	// takes for the same tool apart without opening either.
	Steps int `json:"steps"`
}

// NoCodeTakesDir is where a piece's takes live: the synthetic course's own
// takes/, since that is the directory the recorder joins `take:` onto.
func NoCodeTakesDir(root string) string {
	return filepath.Join(root, NoCodeRoot, "takes")
}

// AvailableTakes lists the takes a no-code piece can actually name.
//
// Served rather than guessed, because `take:` is the one required field on this
// surface that is neither prose nor a choice from a list — it is a filename, in
// a directory nobody would find on their own, spelled without its extension. A
// text box for that is a spelling test with no answer key, and the failure lands
// minutes later inside a run.
func AvailableTakes(root string) []AvailableTake {
	dir := NoCodeTakesDir(root)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []AvailableTake
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".yaml")
		path := filepath.Join(dir, e.Name())
		// Which kind it is comes from the file, not the filename: a take names
		// the thing it drives, and trusting a naming convention here is how a
		// `cursor-*.yaml` that actually drives Figma gets offered for Cursor.
		if t, err := LoadWebTake(path); err == nil {
			out = append(out, AvailableTake{Name: name, Tool: t.Site, Kind: "web", Steps: len(t.Steps)})
			continue
		}
		if t, err := LoadDesktopTake(path); err == nil {
			out = append(out, AvailableTake{Name: name, Tool: t.App, Kind: "desktop", Steps: len(t.Beats)})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
