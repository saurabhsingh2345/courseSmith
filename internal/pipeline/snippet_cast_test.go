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
				Cast: &CastBeat{Pose: "idle", Expression: "sad", Caption: "It sits for days and gets a thumbs up nobody means."}},
			{ID: "the-why", Heading: "The size is the problem", Narration: strings.Repeat("why ", 22),
				Cast: &CastBeat{Pose: "shrug", Expression: "thinking", Prop: "stack", Caption: "Reviewers lose the thread by the second file."}},
			{ID: "the-fix", Heading: "Ship it in slices", Narration: strings.Repeat("fix ", 22),
				Cast: &CastBeat{Pose: "point", Expression: "neutral", Prop: "chart", Caption: "Four small reviews beat one big one."}},
			{ID: "the-payoff", Heading: "Faster, and actually reviewed", Narration: strings.Repeat("payoff ", 22),
				Cast: &CastBeat{Pose: "confident", Expression: "happy", Caption: "Real comments instead of a rubber stamp."}},
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
		p.Beats[2].Cast.Pose = "shrug"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "previous beat already used") {
			t.Fatalf("want repeated-pose error, got %v", err)
		}
	})
	// Reusing a pose later in the clip is a callback, not a freeze.
	t.Run("non-consecutive repeat is fine", func(t *testing.T) {
		p := castPlan()
		p.Beats[3].Cast.Pose = "idle"
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
	wantPrev := []string{"idle", "idle", "shrug", "point"}
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

// Same guard as the figures: a pose Go allows and the renderer does not have
// would fall back to `idle`, so the character just stands there through the
// beat that was meant to be its punchline.
const castMirrorPath = "../../renderer/src/components/cast.tsx"

var (
	tsPosesBlockRe = regexp.MustCompile(`(?s)export const POSES[^{]*\{(.*?)\n\};`)
	tsPoseEntryRe  = regexp.MustCompile(`(?m)^\s{2}([a-zA-Z][a-zA-Z0-9]*):\s*\{`)
	// Expressions used to be a union type. They are a map from our name to an
	// Open Peeps face now, because the name and the drawing are different
	// things — `thinking` is a face called `Suspicious` — and only a map can
	// say so.
	tsFacesBlockRe = regexp.MustCompile(`(?s)export const FACES\s*=\s*\{(.*?)\n\}`)
	tsFaceEntryRe  = regexp.MustCompile(`(?m)^\s{2}([a-zA-Z][a-zA-Z0-9]*):\s*'([A-Za-z]+)'`)
	// Every pose names an Open Peeps bust. A typo in one is not a drift between
	// Go and the renderer, so neither vocabulary test would see it — it is a
	// crash at render time, which is worse.
	tsPoseBodyRe = regexp.MustCompile(`body:\s*'([A-Za-z]+)'`)
)

// The parts react-peeps actually ships, read off the installed package rather
// than copied here — a hand-kept list of somebody else's asset names is a list
// that goes stale the first time they publish.
func peepPartNames(t *testing.T, rel string, decl string) map[string]bool {
	t.Helper()
	path := filepath.Join("../../renderer/node_modules/react-peeps/lib/peeps", rel)
	src, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("react-peeps is not installed (%v); run npm install in renderer/", err)
	}
	block := regexp.MustCompile(`(?s)exports\.` + decl + `\s*=\s*\{(.*?)\n\};`).FindSubmatch(src)
	if block == nil {
		t.Fatalf("no %s map found in %s — has react-peeps changed shape?", decl, path)
	}
	names := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^\s+([A-Za-z][A-Za-z0-9]*):`).FindAllSubmatch(block[1], -1) {
		names[string(m[1])] = true
	}
	if len(names) == 0 {
		t.Fatalf("no %s entries parsed from %s", decl, path)
	}
	return names
}

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
		t.Errorf("castPoseVocab allows %v, which cast.tsx does not have — those beats would stand still", missing)
	}
	if len(extra) > 0 {
		t.Errorf("cast.tsx has %v, which castPoseVocab rejects — the model can never ask for them", extra)
	}
}

func TestCastExpressionVocabularyInSync(t *testing.T) {
	src, err := os.ReadFile(castMirrorPath)
	if err != nil {
		t.Fatalf("reading %s: %v", castMirrorPath, err)
	}
	block := tsFacesBlockRe.FindSubmatch(src)
	if block == nil {
		t.Fatalf("no FACES map found in %s — has its shape changed?", castMirrorPath)
	}
	matches := tsFaceEntryRe.FindAllSubmatch(block[1], -1)
	if len(matches) == 0 {
		t.Fatalf("no face entries parsed from %s", castMirrorPath)
	}
	renderer := map[string]bool{}
	for _, m := range matches {
		renderer[string(m[1])] = true
	}
	for name := range castExpressionVocab {
		if !renderer[name] {
			t.Errorf("castExpressionVocab allows %q, which the FACES map does not — that beat would fall back to neutral", name)
		}
	}
	for name := range renderer {
		if !castExpressionVocab[name] {
			t.Errorf("the FACES map has %q, which castExpressionVocab rejects — the model can never ask for it", name)
		}
	}
}

// The vocabularies can agree with each other perfectly and still name a drawing
// that does not exist. Both sides of that mapping live outside Go — our name on
// one side, react-peeps' asset name on the other — and a typo in the second one
// is not drift, it is a crash on the frame that pose first appears.
func TestCastPartsExistInOpenPeeps(t *testing.T) {
	src, err := os.ReadFile(castMirrorPath)
	if err != nil {
		t.Fatalf("reading %s: %v", castMirrorPath, err)
	}

	busts := peepPartNames(t, "pose/bust/z_options.js", "BustPose")
	posesBlock := tsPosesBlockRe.FindSubmatch(src)
	if posesBlock == nil {
		t.Fatalf("no POSES map found in %s", castMirrorPath)
	}
	bodies := tsPoseBodyRe.FindAllSubmatch(posesBlock[1], -1)
	if len(bodies) != len(castPoseVocab) {
		t.Errorf("parsed %d pose bodies from %s but the vocabulary has %d entries",
			len(bodies), castMirrorPath, len(castPoseVocab))
	}
	for _, m := range bodies {
		if name := string(m[1]); !busts[name] {
			t.Errorf("a pose is drawn from the Open Peeps bust %q, which react-peeps does not ship", name)
		}
	}

	faces := peepPartNames(t, "face/z_options.js", "Face")
	facesBlock := tsFacesBlockRe.FindSubmatch(src)
	if facesBlock == nil {
		t.Fatalf("no FACES map found in %s", castMirrorPath)
	}
	for _, m := range tsFaceEntryRe.FindAllSubmatch(facesBlock[1], -1) {
		if name := string(m[2]); !faces[name] {
			t.Errorf("expression %q is drawn from the Open Peeps face %q, which react-peeps does not ship",
				m[1], name)
		}
	}

	// The blink is a face swap, not an eyelid layer — so it is one more name
	// that has to exist and is referenced nowhere the tests above look.
	if blink := regexp.MustCompile(`BLINK_FACE: FaceType = '([A-Za-z]+)'`).FindSubmatch(src); blink == nil {
		t.Error("no BLINK_FACE found in cast.tsx")
	} else if name := string(blink[1]); !faces[name] {
		t.Errorf("the blink uses the Open Peeps face %q, which react-peeps does not ship", name)
	}

	// Each presenter names a hairstyle, and a style react-peeps does not ship is
	// a crash the moment that presenter's seed comes up — which is one clip in
	// eight, so it would survive a casual look at the output.
	hairs := peepPartNames(t, "hair/z_options.js", "Hair")
	castBlock := regexp.MustCompile(`(?s)export const CASTS[^\[]*\[(.*?)\n\]`).FindSubmatch(src)
	if castBlock == nil {
		t.Fatalf("no CASTS list found in %s", castMirrorPath)
	}
	styles := regexp.MustCompile(`style:\s*'([A-Za-z]+)'`).FindAllSubmatch(castBlock[1], -1)
	if len(styles) < 2 {
		t.Fatalf("parsed %d presenters from CASTS; the cast should have several", len(styles))
	}
	for _, m := range styles {
		if name := string(m[1]); !hairs[name] {
			t.Errorf("a presenter uses the hairstyle %q, which react-peeps does not ship", name)
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

// Busts that must never back a pose, whatever they are called.
//
// Twice now a drawing has been chosen off its filename and turned out to
// contain something no course should ship: `Geek` has a poop emoji on its
// laptop lid, and `Killer` — which reads as hand-to-chin at thumbnail size —
// is a character holding a knife. Both were caught by eye, once each, after
// shipping. A list is not a substitute for looking, but it does stop a
// rejected drawing coming back the next time somebody scans the set for a
// pose that sounds right.
var rejectedBusts = map[string]string{
	"Killer": "the character is holding a knife",
	"Geek":   "there is a poop emoji on the laptop lid",
}

func TestNoPoseUsesARejectedBust(t *testing.T) {
	src, err := os.ReadFile(castMirrorPath)
	if err != nil {
		t.Fatalf("reading %s: %v", castMirrorPath, err)
	}
	block := tsPosesBlockRe.FindSubmatch(src)
	if block == nil {
		t.Fatalf("no POSES map found in %s", castMirrorPath)
	}
	for _, m := range tsPoseBodyRe.FindAllSubmatch(block[1], -1) {
		if why, bad := rejectedBusts[string(m[1])]; bad {
			t.Errorf("a pose is drawn from the Open Peeps bust %q — %s", m[1], why)
		}
	}
}

// A retired pose name has to keep meaning something. Falling through to the
// `idle` fallback would silently drop the gesture from every plan that already
// asked for it.
func TestRetiredPoseNamesRedirect(t *testing.T) {
	if got := normalizeCastPose("think"); got != "point" {
		t.Errorf("normalizeCastPose(\"think\") = %q, want point — the raised finger replaced it", got)
	}
	if castPoseVocab["think"] {
		t.Error("think is still offered as a pose; its only drawing has a knife in it")
	}
	// An alias must not point at a pose that does not exist.
	for from, to := range castPoseAliases {
		if !castPoseVocab[to] {
			t.Errorf("alias %q -> %q, which is not a pose", from, to)
		}
	}
}
