package pipeline

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func castPlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "cast",
		Title:    "Why code review is fast",
		Subtitle: "The bottleneck is the batch, not the reading",
		Beats: []SnippetBeat{
			{ID: "the-pain", Heading: "Nobody reads a five hundred line diff", Narration: strings.Repeat("pain ", 22),
				Cast: &CastBeat{Pose: "defeated", Expression: "concerned", Caption: "It sits for days and gets a thumbs up nobody means."}},
			{ID: "the-why", Heading: "The size is the problem", Narration: strings.Repeat("why ", 22),
				Cast: &CastBeat{Pose: "think", Expression: "thinking", Prop: "stack", Caption: "Reviewers lose the thread by the second file."}},
			{ID: "the-fix", Heading: "Ship it in slices", Narration: strings.Repeat("fix ", 22),
				Cast: &CastBeat{Pose: "point", Expression: "neutral", Prop: "chart", Caption: "Four small reviews beat one big one."}},
			{ID: "the-payoff", Heading: "Faster, and actually reviewed", Narration: strings.Repeat("payoff ", 22),
				Cast: &CastBeat{Pose: "celebrate", Expression: "happy", Caption: "Real comments instead of a rubber stamp."}},
		},
	}
}

func TestValidateCastPlan(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := castPlan().Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	t.Run("beat without cast", func(t *testing.T) {
		p := castPlan()
		p.Beats[1].Cast = nil
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "no cast direction") {
			t.Fatalf("want missing-cast error, got %v", err)
		}
	})
	// The failure this template exists to avoid: a character who holds still
	// across two beats is a photo with the text changing beside it.
	t.Run("consecutive identical poses", func(t *testing.T) {
		p := castPlan()
		p.Beats[2].Cast.Pose = "think"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "previous beat already used") {
			t.Fatalf("want repeated-pose error, got %v", err)
		}
	})
	// Reusing a pose later in the clip is a callback, not a freeze.
	t.Run("non-consecutive repeat is fine", func(t *testing.T) {
		p := castPlan()
		p.Beats[3].Cast.Pose = "defeated"
		if err := p.Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	// An unknown pose normalizes to idle, which would silently collide with a
	// neighbouring idle — so the rule has to be checked on normalized names.
	t.Run("invented pose normalizes and still collides", func(t *testing.T) {
		p := castPlan()
		p.Beats[1].Cast.Pose = "idle"
		p.Beats[2].Cast.Pose = "moonwalk"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "previous beat already used") {
			t.Fatalf("want repeated-pose error after normalization, got %v", err)
		}
	})
	t.Run("heading too long", func(t *testing.T) {
		p := castPlan()
		p.Beats[0].Heading = "this heading runs on and on and on for far too many words indeed"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "headlines are") {
			t.Fatalf("want headline-length error, got %v", err)
		}
	})
	t.Run("foreign beat field", func(t *testing.T) {
		p := castPlan()
		p.Beats[0].Sketch = []SketchItem{{Label: "Browser", Icon: "monitor"}}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "does not use") {
			t.Fatalf("want foreign-field error, got %v", err)
		}
	})
}

func TestCastScenes(t *testing.T) {
	plan := castPlan()
	scenes, err := castScenes(sceneInput(t, plan, 6000))
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != len(plan.Beats) {
		t.Fatalf("got %d scenes, want one per beat (%d)", len(scenes), len(plan.Beats))
	}
	// Each shot has to know where the character is coming from, or the pose
	// change is a jump on the first frame instead of a move.
	wantPrev := []string{"idle", "defeated", "think", "point"}
	for i, s := range scenes {
		if s.Type != SceneCast {
			t.Errorf("scene %d is %q, want %q", i, s.Type, SceneCast)
		}
		if got := s.Props["prevPose"]; got != wantPrev[i] {
			t.Errorf("scene %d prevPose = %v, want %q", i, got, wantPrev[i])
		}
		if got, want := s.Props["flip"], i%2 == 1; got != want {
			t.Errorf("scene %d flip = %v, want %v", i, got, want)
		}
	}
	// A beat with no prop must not carry an empty one — the renderer reserves
	// space for a prop whenever the key is present.
	if _, set := scenes[0].Props["prop"]; set {
		t.Error("a beat with no prop should not set the prop key")
	}
	if got := scenes[1].Props["prop"]; got != "stack" {
		t.Errorf("prop = %v, want stack", got)
	}
}

func TestCastNormalization(t *testing.T) {
	if got := normalizeCastPose("  Point "); got != "point" {
		t.Errorf("normalizeCastPose = %q, want point", got)
	}
	if got := normalizeCastPose("breakdance"); got != "idle" {
		t.Errorf("normalizeCastPose fallback = %q, want idle", got)
	}
	if got := normalizeCastExpression("ECSTATIC"); got != "neutral" {
		t.Errorf("normalizeCastExpression fallback = %q, want neutral", got)
	}
	if got := normalizeCastExpression("Happy"); got != "happy" {
		t.Errorf("normalizeCastExpression = %q, want happy", got)
	}
}

// Same guard as the figures: a pose Go allows and the rig does not have would
// fall back to `idle`, so the character just stands there through the beat that
// was meant to be its punchline.
const castMirrorPath = "../../renderer/src/components/cast.tsx"

var (
	tsPosesBlockRe = regexp.MustCompile(`(?s)export const POSES[^{]*\{(.*?)\n\};`)
	tsPoseEntryRe  = regexp.MustCompile(`(?m)^\s{2}([a-zA-Z][a-zA-Z0-9]*):\s*\{`)
	tsExpressionRe = regexp.MustCompile(`export type Expression\s*=\s*([^;]+);`)
)

func TestCastPoseVocabularyInSync(t *testing.T) {
	src, err := os.ReadFile(castMirrorPath)
	if err != nil {
		t.Fatalf("reading %s: %v", castMirrorPath, err)
	}
	block := tsPosesBlockRe.FindSubmatch(src)
	if block == nil {
		t.Fatalf("no POSES map found in %s — has its shape changed?", castMirrorPath)
	}
	matches := tsPoseEntryRe.FindAllSubmatch(block[1], -1)
	if len(matches) == 0 {
		t.Fatalf("no pose entries parsed from %s", castMirrorPath)
	}
	rig := make(map[string]bool, len(matches))
	for _, m := range matches {
		rig[string(m[1])] = true
	}
	var missing, extra []string
	for name := range castPoseVocab {
		if !rig[name] {
			missing = append(missing, name)
		}
	}
	for name := range rig {
		if !castPoseVocab[name] {
			extra = append(extra, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 {
		t.Errorf("castPoseVocab allows %v, which the rig does not have — those beats would stand still", missing)
	}
	if len(extra) > 0 {
		t.Errorf("the rig has %v, which castPoseVocab rejects — the model can never ask for them", extra)
	}
}

func TestCastExpressionVocabularyInSync(t *testing.T) {
	src, err := os.ReadFile(castMirrorPath)
	if err != nil {
		t.Fatalf("reading %s: %v", castMirrorPath, err)
	}
	m := tsExpressionRe.FindSubmatch(src)
	if m == nil {
		t.Fatalf("no Expression union found in %s", castMirrorPath)
	}
	rig := map[string]bool{}
	for _, part := range strings.Split(string(m[1]), "|") {
		name := strings.Trim(strings.TrimSpace(part), "'\"")
		if name != "" {
			rig[name] = true
		}
	}
	for name := range castExpressionVocab {
		if !rig[name] {
			t.Errorf("castExpressionVocab allows %q, which the rig's Expression union does not", name)
		}
	}
	for name := range rig {
		if !castExpressionVocab[name] {
			t.Errorf("the rig's Expression union has %q, which castExpressionVocab rejects", name)
		}
	}
}

// The prompt's example has to satisfy the rules the same prompt states —
// especially the no-consecutive-poses rule, which is exactly the kind of thing
// an example gets wrong and the model then copies faithfully.
func TestCastPromptExampleIsValid(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("../../prompts", snippetCastTemplateName))
	if err != nil {
		t.Fatalf("reading prompt: %v", err)
	}
	at := bytes.Index(src, []byte(`{"title":`))
	if at < 0 {
		t.Fatalf("no example reply found in %s", snippetCastTemplateName)
	}
	line := src[at:]
	if end := bytes.IndexByte(line, '\n'); end >= 0 {
		line = line[:end]
	}
	var plan SnippetPlan
	if err := json.Unmarshal(line, &plan); err != nil {
		t.Fatalf("the example reply in %s is not valid JSON: %v", snippetCastTemplateName, err)
	}
	plan.Template = "cast"
	for i := range plan.Beats {
		plan.Beats[i].Narration = strings.Repeat("narration ", 22)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("the example reply in %s does not satisfy the rules that same prompt states: %v",
			snippetCastTemplateName, err)
	}
}
