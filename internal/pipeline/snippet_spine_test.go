package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// spinePlan is a well-formed five-shot clip that uses four different shots —
// the shape a real course opening has, and the smallest one that exercises the
// variety rule.
func spinePlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "spine",
		Title:    "Why prompts beat clicking",
		Subtitle: "The habit that makes everything after it faster",
		Beats: []SnippetBeat{
			{ID: "open", Heading: "You are still clicking", Narration: strings.Repeat("open ", 22),
				Spine: &SpineBeat{Shot: "open", Note: "PART ONE", Emphasis: "clicking",
					Caption: "An hour of it produces exactly one result.",
					Objects: []SpineObject{{Figure: "cursor", Label: "Clicking"}}}},
			{ID: "the-turn", Heading: "A prompt is reusable", Narration: strings.Repeat("turn ", 22),
				Spine: &SpineBeat{Shot: "pair", Emphasis: "reusable",
					Objects: []SpineObject{
						{Figure: "cursor", Label: "A click", Detail: "Gone the moment it finishes."},
						{Figure: "chat", Label: "A prompt", Detail: "Runs again tomorrow."},
					}}},
			{ID: "what-it-buys", Heading: "Three things you get back", Narration: strings.Repeat("buys ", 22),
				Spine: &SpineBeat{Shot: "row",
					Objects: []SpineObject{
						{Figure: "clock", Label: "Time"},
						{Figure: "recycle", Label: "Repeatability"},
						{Figure: "share", Label: "Handover"},
					}}},
			{ID: "the-line", Heading: "The work you can hand over is the only work that scales",
				Narration: strings.Repeat("line ", 22),
				Spine:     &SpineBeat{Shot: "quote", Emphasis: "hand over"}},
			{ID: "close", Heading: "Write your first one today", Narration: strings.Repeat("close ", 22),
				Spine: &SpineBeat{Shot: "close", Emphasis: "today",
					Caption: "Open the next lesson and write one",
					Objects: []SpineObject{{Figure: "rocket"}}}},
		},
	}
}

func TestValidateSpinePlan(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := spinePlan().Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	t.Run("beat without a shot", func(t *testing.T) {
		p := spinePlan()
		p.Beats[2].Spine = nil
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "has no shot") {
			t.Fatalf("want missing-shot error, got %v", err)
		}
	})
	t.Run("unknown shot", func(t *testing.T) {
		p := spinePlan()
		p.Beats[1].Spine.Shot = "montage"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "not one of") {
			t.Fatalf("want unknown-shot error, got %v", err)
		}
	})
	// The two structural rules. An `open` in the middle is a second title card
	// and a `close` in the middle ends the clip and then keeps going; both read
	// as an edit mistake rather than as a choice, which is why they are errors
	// and not repairs.
	t.Run("open away from the front", func(t *testing.T) {
		p := spinePlan()
		p.Beats[2].Spine.Shot = "open"
		p.Beats[2].Spine.Objects = p.Beats[2].Spine.Objects[:1]
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "only the first beat may open") {
			t.Fatalf("want open-position error, got %v", err)
		}
	})
	t.Run("close away from the end", func(t *testing.T) {
		p := spinePlan()
		p.Beats[1].Spine.Shot = "close"
		p.Beats[1].Spine.Objects = p.Beats[1].Spine.Objects[:1]
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "only the last beat may close") {
			t.Fatalf("want close-position error, got %v", err)
		}
	})
	t.Run("wrong object count for the shot", func(t *testing.T) {
		p := spinePlan()
		p.Beats[1].Spine.Objects = p.Beats[1].Spine.Objects[:1] // a pair with one
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "that shot takes") {
			t.Fatalf("want arity error, got %v", err)
		}
	})
	t.Run("one figure used twice in a shot", func(t *testing.T) {
		p := spinePlan()
		p.Beats[2].Spine.Objects[1].Figure = "clock"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "twice in one shot") {
			t.Fatalf("want duplicate-figure error, got %v", err)
		}
	})
	t.Run("emphasis that is not in the heading", func(t *testing.T) {
		p := spinePlan()
		p.Beats[1].Spine.Emphasis = "recyclable"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "does not appear in its heading") {
			t.Fatalf("want emphasis error, got %v", err)
		}
	})
	// A quote's heading IS the sentence, so it is the one shot allowed past the
	// headline ceiling — and every other shot must still be held to it.
	t.Run("a long heading is a quote's right and nobody else's", func(t *testing.T) {
		p := spinePlan()
		long := "The work you can hand over is the only work that scales"
		p.Beats[2].Heading = long
		p.Beats[2].Spine.Emphasis = ""
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "headline is") {
			t.Fatalf("want headline-length error on a row shot, got %v", err)
		}
	})
	t.Run("three of the same shot in a row", func(t *testing.T) {
		p := spinePlan()
		for _, i := range []int{1, 2, 3} {
			p.Beats[i].Heading = "One claim per shot"
			p.Beats[i].Spine = &SpineBeat{Shot: "state", Objects: []SpineObject{{Figure: "code"}}}
		}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "in a row") {
			t.Fatalf("want repetition error, got %v", err)
		}
	})
	t.Run("a long clip with only two arrangements", func(t *testing.T) {
		p := spinePlan()
		// Alternating, so the three-in-a-row rule cannot be what fires.
		for i := range p.Beats {
			p.Beats[i].Heading = "One claim per shot"
			if i%2 == 0 {
				p.Beats[i].Spine = &SpineBeat{Shot: "state", Objects: []SpineObject{{Figure: "code"}}}
			} else {
				p.Beats[i].Spine = &SpineBeat{Shot: "focus", Objects: []SpineObject{{Figure: "lock"}}}
			}
		}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "different shots") {
			t.Fatalf("want variety error, got %v", err)
		}
	})
	// A beat here is a shot, and a shot is a thing you cut away from. The shared
	// sixty-word ceiling let a 107-word script come back as three beats, one of
	// them twenty seconds long on a single static arrangement.
	t.Run("one shot held too long", func(t *testing.T) {
		p := spinePlan()
		p.Beats[1].Narration = strings.Repeat("held ", maxSpineBeatWords+5)
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "holds one shot for") {
			t.Fatalf("want a shot-length error, got %v", err)
		}
	})
	t.Run("a label that is a sentence", func(t *testing.T) {
		p := spinePlan()
		p.Beats[2].Spine.Objects[0].Label = "The time you get back on every single later run"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "labels are at most") {
			t.Fatalf("want label-length error, got %v", err)
		}
	})
	// The ordinal is the whole `chapter` shot: it is the four-hundred-pixel
	// numeral set behind the title, and without one the layout is an ordinary
	// heading that has reserved a third of the frame for nothing.
	t.Run("a chapter with no part number", func(t *testing.T) {
		p := spinePlan()
		p.Beats[2].Heading = "The instruction you keep"
		p.Beats[2].Spine = &SpineBeat{Shot: "chapter",
			Objects: []SpineObject{{Figure: "prompt", Label: "Writing one"}}}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "set it to the part number") {
			t.Fatalf("want a missing-ordinal error, got %v", err)
		}
	})
	t.Run("a chapter numbered past the design", func(t *testing.T) {
		p := spinePlan()
		p.Beats[2].Heading = "The instruction you keep"
		p.Beats[2].Spine = &SpineBeat{Shot: "chapter", Ordinal: maxSpineOrdinal + 1,
			Objects: []SpineObject{{Figure: "prompt", Label: "Writing one"}}}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "set it to the part number") {
			t.Fatalf("want an ordinal-range error, got %v", err)
		}
	})
	// An aside is drawn smaller and quieter than the shots around it, so a
	// second one has nothing left to be quieter than.
	t.Run("two asides in one clip", func(t *testing.T) {
		p := spinePlan()
		for _, i := range []int{1, 3} {
			p.Beats[i].Heading = "You already do this elsewhere"
			p.Beats[i].Spine = &SpineBeat{Shot: "aside",
				Objects: []SpineObject{{Figure: "spreadsheet"}}}
		}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "second `aside`") {
			t.Fatalf("want a second-aside error, got %v", err)
		}
	})
}

// The three shots that make this a course template rather than a clip template.
// They were added together and they are the reason a spine clip no longer has
// to cut away to another template to say which part you are in, what is already
// behind you, or that it is stepping sideways for a sentence.
func TestSpineStagesTheCourseShots(t *testing.T) {
	p := spinePlan()
	p.Beats[1].Heading = "The instruction you keep"
	p.Beats[1].Spine = &SpineBeat{Shot: "chapter", Ordinal: 2, Note: "CHAPTER",
		Objects: []SpineObject{
			{Figure: "prompt", Label: "Writing one"},
			{Figure: "guardrail", Label: "Keeping it honest"},
		}}
	p.Beats[2].Heading = "You already do this in a spreadsheet"
	p.Beats[2].Spine = &SpineBeat{Shot: "aside",
		Objects: []SpineObject{{Figure: "spreadsheet"}}}
	p.Beats[3].Heading = "That is the whole habit"
	p.Beats[3].Spine = &SpineBeat{Shot: "recap",
		Objects: []SpineObject{
			{Figure: "blueprint", Label: "Planned"},
			{Figure: "blocks", Label: "Assembled"},
			{Figure: "deploy", Label: "Shipped"},
		}}
	if err := p.Validate(); err != nil {
		t.Fatalf("want the course shots to validate, got %v", err)
	}
	in := SnippetSceneInput{Plan: p}
	for i := range p.Beats {
		in.Spans = append(in.Spans, SectionSpan{StartMs: i * 4000})
		in.BeatEndMs = append(in.BeatEndMs, (i+1)*4000)
	}
	scenes, err := spineScenes(in)
	if err != nil {
		t.Fatalf("spineScenes: %v", err)
	}
	// The ordinal has to reach the renderer, because it is the one prop the
	// scene cannot work out for itself — the beat index is not the part number.
	if got := scenes[1].Props["ordinal"]; got != 2 {
		t.Errorf("chapter beat sent ordinal %v, want 2", got)
	}
	if got := scenes[3].Props["ordinal"]; got != 0 {
		t.Errorf("recap beat sent ordinal %v, want it cleared", got)
	}
}

// An ordinal on anything but a chapter is meaningless, and drawn it is a
// four-hundred-pixel numeral sitting behind a claim. Cleared rather than
// rejected: the planner meant nothing by it, and there is nothing to learn from
// a correction round spent on that.
func TestSpineClearsOrdinalOffNonChapterShots(t *testing.T) {
	p := spinePlan()
	p.Beats[1].Spine.Ordinal = 7
	normalizeSpinePlan(p)
	if p.Beats[1].Spine.Ordinal != 0 {
		t.Errorf("a %q beat kept ordinal %d", p.Beats[1].Spine.Shot, p.Beats[1].Spine.Ordinal)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("want the plan still valid after the repair, got %v", err)
	}
}

// A beat carrying another template's payload is the failure the ownership
// declaration exists for: it plans, it validates, and it renders a frame with
// nothing in it.
func TestSpineRejectsForeignBeatFields(t *testing.T) {
	p := spinePlan()
	p.Beats[1].Art = &ArtBeat{Figure: "gears"}
	if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "art") {
		t.Fatalf("want foreign-field error, got %v", err)
	}
}

func TestNormalizeSpinePlan(t *testing.T) {
	// An unknown shot is repaired by object count rather than to a fixed
	// default: the count is the part the planner got right.
	t.Run("unknown shot resolves by object count", func(t *testing.T) {
		for _, tc := range []struct {
			objects int
			want    string
		}{{0, "quote"}, {1, "state"}, {2, "pair"}, {4, "row"}} {
			if got := normalizeSpineShot("montage", tc.objects); got != tc.want {
				t.Errorf("a %d-object montage became %q, want %q", tc.objects, got, tc.want)
			}
		}
	})
	t.Run("a figure outside the vocabulary becomes a spark", func(t *testing.T) {
		p := spinePlan()
		p.Beats[2].Spine.Objects[0].Figure = "flux-capacitor"
		normalizeSpinePlan(p)
		if got := p.Beats[2].Spine.Objects[0].Figure; got != "spark" {
			t.Errorf("figure = %q, want spark", got)
		}
	})
	// Clearing rather than rejecting: the mismatch is almost always the model
	// quoting the idea rather than the words, and a round spent teaching it
	// about string matching buys nothing.
	t.Run("an emphasis not in the heading is dropped", func(t *testing.T) {
		p := spinePlan()
		p.Beats[1].Spine.Emphasis = "recyclable"
		normalizeSpinePlan(p)
		if p.Beats[1].Spine.Emphasis != "" {
			t.Errorf("emphasis = %q, want it cleared", p.Beats[1].Spine.Emphasis)
		}
		if err := p.Validate(); err != nil {
			t.Errorf("a normalized plan should validate, got %v", err)
		}
	})
	t.Run("objects past what the layout draws are dropped", func(t *testing.T) {
		p := spinePlan()
		p.Beats[1].Spine.Objects = append(p.Beats[1].Spine.Objects, SpineObject{Figure: "rocket"})
		normalizeSpinePlan(p)
		if n := len(p.Beats[1].Spine.Objects); n != 2 {
			t.Errorf("a pair kept %d objects, want 2", n)
		}
	})
}

// The write-your-own-narration mode's whole promise, and the only rule in the
// catalog that is about the creator's words rather than about the clip.
func TestSpineSpeaksTheCreatorsScriptVerbatim(t *testing.T) {
	const script = "You are still clicking. A prompt is an instruction you keep. " +
		"Write it once and it runs again tomorrow, unchanged."

	withScript := func(beats ...string) *SnippetPlan {
		p := spinePlan()
		p.spineScript = script
		for i, n := range beats {
			p.Beats[i].Narration = n
		}
		for i := len(beats); i < len(p.Beats); i++ {
			p.Beats[i].Narration = ""
		}
		return p
	}

	t.Run("split at sentence boundaries", func(t *testing.T) {
		p := withScript(
			"You are still clicking.",
			"A prompt is an instruction you keep.",
			"Write it once and it runs again tomorrow, unchanged.",
		)
		if err := validateSpineScript(p); err != nil {
			t.Fatalf("want the split accepted, got %v", err)
		}
	})
	// Punctuation and case are not the creator's words in any sense that
	// matters — the beats are the script cut into pieces and every join loses
	// whitespace and gains a full stop.
	t.Run("punctuation and case are not divergences", func(t *testing.T) {
		p := withScript(
			"You are still clicking",
			"a prompt is an instruction you keep —",
			"write it once, and it runs again tomorrow unchanged!",
		)
		if err := validateSpineScript(p); err != nil {
			t.Fatalf("want punctuation ignored, got %v", err)
		}
	})
	t.Run("a reworded line is refused", func(t *testing.T) {
		p := withScript(
			"You are still clicking.",
			"A prompt is an instruction you save.", // "keep" -> "save"
			"Write it once and it runs again tomorrow, unchanged.",
		)
		err := validateSpineScript(p)
		if err == nil || !strings.Contains(err.Error(), "diverges at word") {
			t.Fatalf("want a divergence error, got %v", err)
		}
		// The report has to point somewhere the creator can find in their own
		// file, which means quoting their words and not just an index.
		if !strings.Contains(err.Error(), "keep") || !strings.Contains(err.Error(), "save") {
			t.Errorf("the divergence should quote both sides: %v", err)
		}
	})
	t.Run("a dropped tail is refused", func(t *testing.T) {
		p := withScript(
			"You are still clicking.",
			"A prompt is an instruction you keep.",
		)
		if err := validateSpineScript(p); err == nil || !strings.Contains(err.Error(), "missing the last") {
			t.Fatalf("want a truncation error, got %v", err)
		}
	})
	t.Run("an added line is refused", func(t *testing.T) {
		p := withScript(
			"You are still clicking.",
			"A prompt is an instruction you keep.",
			"Write it once and it runs again tomorrow, unchanged.",
			"And that is the whole trick.",
		)
		if err := validateSpineScript(p); err == nil || !strings.Contains(err.Error(), "did not write") {
			t.Fatalf("want an addition error, got %v", err)
		}
	})
	// Without a script the mode is off and the ordinary word budget applies —
	// the check must not fire on the templates' normal path.
	t.Run("no script, no rule", func(t *testing.T) {
		if err := validateSpineScript(spinePlan()); err != nil {
			t.Fatalf("want no check without a script, got %v", err)
		}
	})
}

// A script sets its own runtime. Left to the default, a 300-word script would
// be planned against a 45-second budget — about 112 words — and the model would
// be given two instructions it cannot both obey.
func TestSpineRuntimeFollowsTheScript(t *testing.T) {
	words := func(n int) string { return strings.TrimSpace(strings.Repeat("word ", n)) }
	for _, tc := range []struct {
		name    string
		script  string
		pace    int
		wantSec int
	}{
		{"150 words at the house pace", words(150), 150, 60},
		{"a slower voice needs longer", words(150), 120, 75},
		{"pace unset falls back to 150", words(75), 0, 30},
		{"a one-line script cannot go below the floor", words(4), 150, minSnippetTargetSec},
		{"a script longer than a snippet is capped", words(2000), 150, maxSnippetTargetSec},
	} {
		if got := scriptTargetSec(tc.script, tc.pace); got != tc.wantSec {
			t.Errorf("%s: got %ds, want %ds", tc.name, got, tc.wantSec)
		}
	}

	// And the routing: the derived runtime applies to the templates that speak
	// a script and to no others. A creator who supplies a script and picks
	// `gauge` asked for a gauge, and a gauge writes its own narration to the
	// runtime that was requested.
	script := SnippetSpec{Template: "spine", Narration: words(300), TargetSec: 20}
	if got := script.ScriptTargetSec(150); got != 120 {
		t.Errorf("a spine with a 300-word script planned to %ds, want 120", got)
	}
	ignored := SnippetSpec{Template: "illustration", Narration: words(300), TargetSec: 20}
	if got := ignored.ScriptTargetSec(150); got != 20 {
		t.Errorf("a template that does not take a script planned to %ds, want the requested 20", got)
	}
	none := SnippetSpec{Template: "spine", TargetSec: 30}
	if got := none.ScriptTargetSec(150); got != 30 {
		t.Errorf("a spine with no script planned to %ds, want the requested 30", got)
	}
}

// The knob that produces the extra beats: the same word budget cut into more
// shots, which is what "cut more often" means to everything downstream.
func TestSpineCutsMoreOftenThanTheSharedDefault(t *testing.T) {
	const words = 107 // the script that came back as three beats
	_, _, shared, sharedPerBeat := beatBounds(words, templateBeatCeiling("illustration"), templateIdealWords("illustration"))
	_, spineMax, spineSuggest, spinePerBeat := beatBounds(words, templateBeatCeiling("spine"), templateIdealWords("spine"))
	if spineSuggest <= shared {
		t.Errorf("spine suggests %d beats for %d words, the shared default suggests %d — it should ask for more",
			spineSuggest, words, shared)
	}
	if spinePerBeat >= sharedPerBeat {
		t.Errorf("spine calibrates %d words a beat against the shared %d — it should ask for shorter",
			spinePerBeat, sharedPerBeat)
	}
	// And the number the model is told has to be one its own validator accepts,
	// or every plan arrives needing a correction round.
	if spinePerBeat > maxSpineBeatWords {
		t.Errorf("the prompt calibrates %d words a beat but the validator caps a shot at %d",
			spinePerBeat, maxSpineBeatWords)
	}
	if spineMax < spineSuggest {
		t.Errorf("spine may use %d beats but is told to write %d", spineMax, spineSuggest)
	}
}

func TestSpineScenes(t *testing.T) {
	p := spinePlan()
	in := SnippetSceneInput{Plan: p}
	for i := range p.Beats {
		in.Spans = append(in.Spans, SectionSpan{StartMs: i * 4000})
		in.BeatEndMs = append(in.BeatEndMs, (i+1)*4000)
	}
	scenes, err := spineScenes(in)
	if err != nil {
		t.Fatalf("spineScenes: %v", err)
	}
	if len(scenes) != len(p.Beats) {
		t.Fatalf("got %d scenes, want one per beat (%d)", len(scenes), len(p.Beats))
	}
	for i, s := range scenes {
		if s.Type != SceneSpine {
			t.Errorf("scene %d is %q, want %q", i, s.Type, SceneSpine)
		}
		// The rail is drawn from these two, and they are the only props that
		// describe the CLIP rather than the beat — so the scene cannot work
		// them out for itself and a missing one is an invisible regression.
		if s.Props["index"] != i {
			t.Errorf("scene %d has index %v", i, s.Props["index"])
		}
		if s.Props["total"] != len(p.Beats) {
			t.Errorf("scene %d has total %v, want %d", i, s.Props["total"], len(p.Beats))
		}
	}
	objects, ok := scenes[2].Props["objects"].([]map[string]any)
	if !ok || len(objects) != 3 {
		t.Fatalf("the row shot's objects did not survive: %#v", scenes[2].Props["objects"])
	}
	if objects[0]["figure"] != "clock" || objects[0]["label"] != "Time" {
		t.Errorf("object 0 = %#v", objects[0])
	}
}

// The shot vocabulary lives in Go (which validates it) and in the renderer
// (which draws it), and neither can import the other. A shot Go allows and the
// renderer has no case for renders as an empty frame — invisible until somebody
// watches the clip, which is the same failure the figure mirror test exists for.
const spineShotMirrorPath = "../../renderer/src/components/SpineScene.tsx"

var tsSpineShotRe = regexp.MustCompile(`(?s)export type SpineShot =(.*?);`)

func TestSpineShotVocabularyInSync(t *testing.T) {
	src, err := os.ReadFile(spineShotMirrorPath)
	if err != nil {
		t.Fatalf("reading %s: %v", spineShotMirrorPath, err)
	}
	block := tsSpineShotRe.FindSubmatch(src)
	if block == nil {
		t.Fatalf("no SpineShot union found in %s — has its shape changed?", spineShotMirrorPath)
	}
	renderer := map[string]bool{}
	for _, m := range regexp.MustCompile(`'([a-z]+)'`).FindAllStringSubmatch(string(block[1]), -1) {
		renderer[m[1]] = true
	}
	if len(renderer) == 0 {
		t.Fatalf("no shots parsed out of the SpineShot union")
	}
	for _, name := range SpineShotNames() {
		if !renderer[name] {
			t.Errorf("Go allows the %q shot and %s cannot draw it", name, spineShotMirrorPath)
		}
		delete(renderer, name)
	}
	for name := range renderer {
		t.Errorf("%s draws a %q shot that Go will never emit", spineShotMirrorPath, name)
	}
}

// The example reply in the prompt is the one piece of the instructions a model
// copies most literally, so an example that would fail this template's own
// validator teaches it to produce exactly that failure.
func TestSpinePromptExampleIsAValidPlan(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("../../prompts", snippetSpineTemplateName))
	if err != nil {
		t.Fatalf("reading the prompt: %v", err)
	}
	start := strings.Index(string(src), `{"title":`)
	if start < 0 {
		t.Fatalf("no example reply found in %s", snippetSpineTemplateName)
	}
	end := strings.Index(string(src[start:]), "\n")
	if end < 0 {
		t.Fatalf("the example reply in %s is not on one line", snippetSpineTemplateName)
	}
	var plan SnippetPlan
	if err := json.Unmarshal(src[start:start+end], &plan); err != nil {
		t.Fatalf("the example reply is not valid JSON: %v", err)
	}
	plan.Template = "spine"
	// The narration fields are elided in the example, so the shared word-count
	// rules cannot apply; what is being checked is the SHAPE the model copies.
	for i := range plan.Beats {
		plan.Beats[i].Narration = strings.Repeat("word ", 22)
	}
	normalizeSnippetPlan(&plan)
	if err := plan.Validate(); err != nil {
		t.Errorf("the example reply in %s would be rejected by this template: %v",
			snippetSpineTemplateName, err)
	}
	// Every figure named in the example has to exist, or the prompt is
	// advertising a drawing the renderer does not have.
	for _, b := range plan.Beats {
		for _, o := range b.Spine.Objects {
			if !artFigureVocab[o.Figure] {
				t.Errorf("the example names the figure %q, which is not in the vocabulary", o.Figure)
			}
		}
	}
}
