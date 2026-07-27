package pipeline

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// storyPlan is a well-formed eight-shot piece: enough stagings, enough camera
// movement, enough of the presenter, and no cut repeating both axes.
func storyPlan() *SnippetPlan {
	shots := []struct {
		id, heading, staging, camera, pose, prop, propB string
	}{
		{"the-slow-query", "Your query reads every row", "duo", "push", "reading", "stack", ""},
		{"why-scanning-fails", "Scanning is fine until it is not", "object", "hold", "", "clock", ""},
		{"the-turn", "An index is a sorted copy", "hero", "push", "point", "", ""},
		{"the-mechanism", "Sorted means you can skip", "pair", "pan", "", "chart", "network"},
		{"the-cost", "Every write pays for it", "duo", "drift", "think", "gears", ""},
		{"the-tradeoff", "Reads get cheap, writes get dear", "object", "rise", "", "shield", ""},
		{"the-rule", "Index what you filter on", "hero", "hold", "confident", "", ""},
		{"the-landing", "One line, a thousand times faster", "empty", "pull", "", "", ""},
	}
	beats := make([]SnippetBeat, 0, len(shots))
	for _, s := range shots {
		beats = append(beats, SnippetBeat{
			ID:        s.id,
			Heading:   s.heading,
			Narration: strings.Repeat("narration ", 22),
			Shot: &ShotBeat{
				BeatID: s.id, Staging: s.staging, Camera: s.camera,
				Pose: s.pose, Expression: "neutral", Prop: s.prop, PropB: s.propB,
				Caption: "A supporting line that sits under the heading.",
			},
		})
	}
	return &SnippetPlan{Template: "story", Title: "How an index finds your row", Beats: beats}
}

func TestValidateStoryPlan(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := storyPlan().Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	t.Run("too few beats", func(t *testing.T) {
		p := storyPlan()
		p.Beats = p.Beats[:5]
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "want 8-14") {
			t.Fatalf("want beat-count error, got %v", err)
		}
	})
	t.Run("beat without a shot", func(t *testing.T) {
		p := storyPlan()
		p.Beats[3].Shot = nil
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "has no shot") {
			t.Fatalf("want missing-shot error, got %v", err)
		}
	})
	// The cut is the unit of this template. Repeating both axes across one is
	// literally the same picture twice.
	t.Run("cut repeating staging and camera", func(t *testing.T) {
		p := storyPlan()
		p.Beats[1].Shot.Staging = "duo"
		p.Beats[1].Shot.Camera = "push"
		p.Beats[1].Shot.Pose = "shrug"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "repeats the previous shot") {
			t.Fatalf("want repeated-shot error, got %v", err)
		}
	})
	// Repeating one axis is not just allowed, it is how a scene gets built.
	t.Run("holding a staging across a cut is fine", func(t *testing.T) {
		p := storyPlan()
		p.Beats[1].Shot.Staging = "duo"
		p.Beats[1].Shot.Camera = "rise"
		p.Beats[1].Shot.Pose = "shrug"
		if err := p.Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	t.Run("presenter barely on stage", func(t *testing.T) {
		p := storyPlan()
		for _, i := range []int{0, 2, 4} {
			p.Beats[i].Shot.Staging = "object"
			p.Beats[i].Shot.Prop = "rocket"
			p.Beats[i].Shot.Pose = ""
		}
		p.Beats[1].Shot.Staging = "pair"
		p.Beats[1].Shot.Prop = "clock"
		p.Beats[1].Shot.PropB = "gears"
		p.Beats[3].Shot.Staging = "empty"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "presenter on stage") {
			t.Fatalf("want presenter-floor error, got %v", err)
		}
	})
	t.Run("camera never moves", func(t *testing.T) {
		p := storyPlan()
		for i := range p.Beats {
			p.Beats[i].Shot.Camera = "hold"
		}
		// Keep the cuts legal on the staging axis so the moving-camera rule is
		// what fails rather than the repeat rule.
		stagings := []string{"duo", "object", "hero", "pair", "duo", "object", "hero", "empty"}
		for i := range p.Beats {
			p.Beats[i].Shot.Staging = stagings[i]
		}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "move the camera") {
			t.Fatalf("want camera-movement error, got %v", err)
		}
	})
	t.Run("too few stagings", func(t *testing.T) {
		p := storyPlan()
		cams := []string{"push", "hold", "push", "hold", "push", "hold", "push", "hold"}
		for i := range p.Beats {
			if i%2 == 0 {
				p.Beats[i].Shot.Staging = "hero"
				p.Beats[i].Shot.Pose = "point"
				p.Beats[i].Shot.Prop = ""
			} else {
				p.Beats[i].Shot.Staging = "object"
				p.Beats[i].Shot.Prop = "clock"
				p.Beats[i].Shot.Pose = ""
			}
			p.Beats[i].Shot.Camera = cams[i]
		}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "only 2 stagings") {
			t.Fatalf("want staging-variety error, got %v", err)
		}
	})
	t.Run("staging without its prop", func(t *testing.T) {
		p := storyPlan()
		p.Beats[1].Shot.Prop = ""
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "names no prop") {
			t.Fatalf("want missing-prop error, got %v", err)
		}
	})
	// A comparison of a thing with itself is not a comparison.
	t.Run("pair of identical props", func(t *testing.T) {
		p := storyPlan()
		p.Beats[3].Shot.PropB = "chart"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "with itself") {
			t.Fatalf("want identical-pair error, got %v", err)
		}
	})
	t.Run("character staging without a pose", func(t *testing.T) {
		p := storyPlan()
		p.Beats[2].Shot.Pose = ""
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "no pose") {
			t.Fatalf("want missing-pose error, got %v", err)
		}
	})
	t.Run("foreign beat field", func(t *testing.T) {
		p := storyPlan()
		p.Beats[0].Sketch = []SketchItem{{Label: "Browser", Icon: "monitor"}}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "does not use") {
			t.Fatalf("want foreign-field error, got %v", err)
		}
	})
}

func TestStoryScenes(t *testing.T) {
	plan := storyPlan()
	scenes, err := storyScenes(sceneInput(t, plan, 8000))
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != len(plan.Beats) {
		t.Fatalf("got %d scenes, want one per beat (%d)", len(scenes), len(plan.Beats))
	}
	for i, s := range scenes {
		if s.Type != SceneStory {
			t.Errorf("scene %d is %q, want %q", i, s.Type, SceneStory)
		}
		// The renderer eases the camera over the shot's own length, so it has
		// to be told what that is — a scene component only knows its frame.
		if got := s.Props["durationMs"]; got != 8000 {
			t.Errorf("scene %d durationMs = %v, want 8000", i, got)
		}
	}
	// A shot with no presenter must not advance the pose the next one eases
	// from, or the character teleports across the intervening object shots.
	if got := scenes[2].Props["prevPose"]; got != "reading" {
		t.Errorf("hero shot after an object shot has prevPose %v, want the last posed shot's (reading)", got)
	}
	if _, set := scenes[7].Props["prop"]; set {
		t.Error("an empty staging should not carry a prop")
	}
	if got := scenes[3].Props["propB"]; got != "network" {
		t.Errorf("pair propB = %v, want network", got)
	}
}

// A story is eight beats minimum; the shared 45-second default cannot fund
// that, so the template declares its own.
func TestStoryRuntimeDefault(t *testing.T) {
	story := SnippetSpec{Template: "story"}
	if got := story.ResolvedTargetSec(); got != 90 {
		t.Errorf("story default runtime = %ds, want 90", got)
	}
	// An explicit request still wins.
	if got := (SnippetSpec{Template: "story", TargetSec: 120}).ResolvedTargetSec(); got != 120 {
		t.Errorf("explicit runtime = %d, want 120", got)
	}
	// Templates without their own default are unaffected.
	if got := (SnippetSpec{Template: "illustration"}).ResolvedTargetSec(); got != defaultSnippetTargetSec {
		t.Errorf("illustration default = %d, want %d", got, defaultSnippetTargetSec)
	}
	// And the runtime must actually be able to pay for the beat floor: the
	// word budget's own minimum has to clear eight beats of the per-beat floor.
	minWords := 90 * 175 / 60 * 75 / 100
	if minWords < minStoryBeats*minWordsPerBeat {
		t.Errorf("a 90s story funds only %d words but %d beats need at least %d",
			minWords, minStoryBeats, minStoryBeats*minWordsPerBeat)
	}
}

func TestStoryNormalization(t *testing.T) {
	if got := normalizeStaging("  Duo "); got != "duo" {
		t.Errorf("normalizeStaging = %q, want duo", got)
	}
	if got := normalizeStaging("montage"); got != "empty" {
		t.Errorf("normalizeStaging fallback = %q, want empty", got)
	}
	if got := normalizeCamera("PUSH"); got != "push" {
		t.Errorf("normalizeCamera = %q, want push", got)
	}
	if got := normalizeCamera("dolly-zoom"); got != "hold" {
		t.Errorf("normalizeCamera fallback = %q, want hold", got)
	}
}

// Same guard as the figures and the poses: a camera move Go allows and the
// renderer does not implement silently becomes a locked-off shot, and a shot
// list full of those is the slideshow this template exists to avoid.
const cameraMirrorPath = "../../renderer/src/components/camera.ts"

var tsCameraCaseRe = regexp.MustCompile(`(?m)^\s*case '([a-z]+)':`)

func TestCameraVocabularyInSync(t *testing.T) {
	src, err := os.ReadFile(cameraMirrorPath)
	if err != nil {
		t.Fatalf("reading %s: %v", cameraMirrorPath, err)
	}
	matches := tsCameraCaseRe.FindAllSubmatch(src, -1)
	if len(matches) == 0 {
		t.Fatalf("no camera cases parsed from %s", cameraMirrorPath)
	}
	impl := map[string]bool{}
	for _, m := range matches {
		impl[string(m[1])] = true
	}
	// "hold" is the default arm of the switch rather than a case of its own.
	impl["hold"] = true

	var missing []string
	for name := range cameraVocab {
		if !impl[name] {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("cameraVocab allows %v, which %s does not implement — those shots would be locked off",
			missing, cameraMirrorPath)
	}
	for name := range impl {
		if !cameraVocab[name] {
			t.Errorf("%s implements %q, which cameraVocab rejects — the director can never ask for it",
				cameraMirrorPath, name)
		}
	}
}

// The staging vocabulary has the same failure mode: an unknown staging falls
// back to `empty`, so a shot the director meant to carry the presenter would
// render as type on a bare stage.
func TestStagingVocabularyInSync(t *testing.T) {
	src, err := os.ReadFile("../../renderer/src/components/StoryScene.tsx")
	if err != nil {
		t.Fatalf("reading StoryScene: %v", err)
	}
	matches := tsCameraCaseRe.FindAllSubmatch(src, -1)
	impl := map[string]bool{}
	for _, m := range matches {
		impl[string(m[1])] = true
	}
	for name := range stagingVocab {
		if !impl[name] {
			t.Errorf("stagingVocab allows %q, which StoryScene does not stage", name)
		}
	}
}

// Both prompts ship an example, and both examples have to satisfy the rules
// their own prompt states — the director's example especially, since the
// no-repeated-cut rule is exactly the sort of thing an example gets wrong and
// the model then copies faithfully.
func TestStoryShotsPromptExampleObeysTheCutRule(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("../../prompts", snippetStoryShotsTemplateName))
	if err != nil {
		t.Fatalf("reading prompt: %v", err)
	}
	shots := extractExampleShots(t, string(src))
	if len(shots) < 2 {
		t.Fatalf("the director prompt's example shows %d shots — too few to demonstrate a cut", len(shots))
	}
	for i := 1; i < len(shots); i++ {
		if shots[i].Staging == shots[i-1].Staging && shots[i].Camera == shots[i-1].Camera {
			t.Errorf("the example in %s repeats staging %q and camera %q across a cut — the same prompt forbids it",
				snippetStoryShotsTemplateName, shots[i].Staging, shots[i].Camera)
		}
	}
	for _, s := range shots {
		if !stagingVocab[s.Staging] {
			t.Errorf("the example uses staging %q, which is not in the vocabulary the prompt lists", s.Staging)
		}
		if !cameraVocab[s.Camera] {
			t.Errorf("the example uses camera %q, which is not in the vocabulary the prompt lists", s.Camera)
		}
	}
}

var exampleShotRe = regexp.MustCompile(`"staging":\s*"([a-z]+)",\s*"camera":\s*"([a-z]+)"`)

func extractExampleShots(t *testing.T, src string) []ShotBeat {
	t.Helper()
	var out []ShotBeat
	for _, m := range exampleShotRe.FindAllStringSubmatch(src, -1) {
		out = append(out, ShotBeat{Staging: m[1], Camera: m[2]})
	}
	return out
}

// The writer prompt must not leak staging vocabulary: its whole job is that it
// does not stage, and a writer that starts emitting camera moves produces beats
// the director then contradicts.
func TestStoryScriptPromptStaysOutOfStaging(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("../../prompts", snippetStoryScriptTemplateName))
	if err != nil {
		t.Fatalf("reading prompt: %v", err)
	}
	text := string(src)
	for _, word := range []string{"staging", "\"camera\"", "\"pose\"", "\"prop\""} {
		if strings.Contains(strings.ToLower(text), strings.ToLower(word)) {
			t.Errorf("%s mentions %s — the writer call must not stage anything",
				snippetStoryScriptTemplateName, word)
		}
	}
}
