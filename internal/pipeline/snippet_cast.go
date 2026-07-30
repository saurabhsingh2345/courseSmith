package pipeline

// The cast template.
//
// A character explains something: they stand beside the headline, hold a pose
// that carries an attitude to it, and change pose between beats.
//
// It is the illustration template's sibling — same one-beat-one-shot shape,
// same kinetic headline — and the difference is the point. An object can show
// what a thing *is*; only a person can show how to *feel* about it. "This is
// the problem" is a sad face, "I'm not sure" is a shrug, "here it is" is a
// point. That register is what an explainer opens and closes on, and neither
// the diagram nor the board nor a floating rocket can do it.
//
// As everywhere else in the catalog, the model does not draw. It picks a pose
// and an expression from closed vocabularies and the renderer (cast.tsx)
// assembles the character from Open Peeps parts. What the vocabularies can
// offer is therefore what somebody drew, which is a real constraint and the
// reason this file's lists are shorter and stranger than they would be if we
// had invented them — see castPoseVocab.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "cast",
		Category:    CatPresenting,
		Title:       "Explainer character",
		Description: "A person who reacts as they explain, with a headline beside them. Reach for it when a presence on screen carries the idea better than a diagram.",
		Example:     "Why code review makes teams faster, not slower",
		PromptFile:  snippetCastTemplateName,
		NeedsCode:   false,
		Owns:        beatFields{Cast: true},
		Normalize:   normalizeCastPlan,
		Validate:    validateCastPlan,
		Scenes:      castScenes,
		// Shelved 2026-07-30, for the reason set out at length on `story`.
		//
		// It lands harder here, because where `story` merely contains a
		// character this template *is* one: the file header argues that a shrug
		// and a raised finger carry a register no diagram can, and that is
		// right. But the drawn vocabulary that survived castPoseVocab is five
		// poses, two of which are the same arms-down stance read differently,
		// and the register the argument depends on is exactly what got cut.
		//
		// So the promise and the parts are a whole rig apart, and the honest
		// state is off the shelf until the parts catch up.
		Shelved: true,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Poses":       strings.Join(CastPoseNames(), ", "),
				"Expressions": strings.Join(CastExpressionNames(), ", "),
				"Props":       strings.Join(ArtFigureNames(), ", "),
			}
		},
	})
}

const snippetCastTemplateName = "snippet_cast.tmpl"

// castPoseVocab mirrors POSES in renderer/src/components/cast.tsx; a drift test
// keeps the two identical. A pose Go allows and the renderer does not have
// would silently fall back to `idle`, so the character would simply stand there
// through the beat that was supposed to be its punchline.
//
// This list is no longer ours to choose. The character used to be drawn from a
// skeleton, so a pose was whatever eleven joint angles we cared to write down;
// it is Open Peeps artwork now, so a pose exists exactly when somebody drew it.
// `wave`, `celebrate`, `defeated` and `walk` were dropped because no drawing of
// them exists. `think` was dropped because its only drawing has the character
// holding a knife (see castPoseAliases). `explain`, `coffee` and `phone` were
// dropped because their drawings cannot be coloured: Open Peeps fills those garments with the same
// value that paints the hands, so keeping the hands right dresses the character
// in their own skin. Offering a name the artwork cannot satisfy is how a model
// ends up asking for a shot that silently renders as somebody standing still.
var castPoseVocab = map[string]bool{
	"idle":      true,
	"point":     true,
	"shrug":     true,
	"confident": true,
	"reading":   true,
}

// castPoseAliases redirect a retired name to the pose that replaced it.
//
// `think` was backed by the only hand-to-chin drawing in the set, and that
// hand is holding a knife — Open Peeps calls the bust `Killer`, which is the
// clue the filename gave and the thumbnail hid. Dropping the name outright
// would send every plan that already uses it to the `idle` fallback, losing
// the gesture as well as the knife; the raised finger is the honest
// replacement, since the beat `think` was written for is the one where an idea
// lands.
var castPoseAliases = map[string]string{
	"think": "point",
}

// castExpressionVocab mirrors FACES in cast.tsx.
//
// `sad` and `serious` are here because the register the dropped `defeated` pose
// carried had to go somewhere, and a face carries it better than a slump did.
var castExpressionVocab = map[string]bool{
	"neutral":   true,
	"happy":     true,
	"thinking":  true,
	"surprised": true,
	"concerned": true,
	"sad":       true,
	"serious":   true,
	"talking":   true,
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CastPoseNames returns the pose vocabulary, sorted.
func CastPoseNames() []string { return sortedKeys(castPoseVocab) }

// CastExpressionNames returns the expression vocabulary, sorted.
func CastExpressionNames() []string { return sortedKeys(castExpressionVocab) }

func normalizeCastPose(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if castPoseVocab[n] {
		return n
	}
	if to, ok := castPoseAliases[n]; ok {
		return to
	}
	return "idle"
}

func normalizeCastExpression(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if castExpressionVocab[n] {
		return n
	}
	return "neutral"
}

// castScenes lays the clip out as one shot per beat.
//
// Each scene is told the *previous* beat's pose as well as its own, which is
// what lets the renderer interpolate rather than cut. Without it a scene has no
// way to know where the character is coming from, and the pose change would be
// a jump on the first frame of every shot — the exact thing a rig is for
// avoiding.
func castScenes(in SnippetSceneInput) ([]Scene, error) {
	scenes := make([]Scene, 0, len(in.Plan.Beats))
	prevPose := "idle"
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Cast == nil {
			return nil, fmt.Errorf("beat %q has no cast direction", beat.ID)
		}
		pose := normalizeCastPose(beat.Cast.Pose)
		props := map[string]any{
			"headline":   beat.Heading,
			"caption":    beat.Cast.Caption,
			"pose":       pose,
			"prevPose":   prevPose,
			"expression": normalizeCastExpression(beat.Cast.Expression),
			"flip":       i%2 == 1,
		}
		if beat.Cast.Prop != "" {
			props["prop"] = normalizeArtFigure(beat.Cast.Prop)
		}
		scenes = append(scenes, Scene{
			Type: SceneCast, StartMs: startMs, EndMs: endMs, Props: props,
		})
		prevPose = pose
	}
	return scenes, nil
}

// normalizeCastPlan writes the vocabularies back onto the plan.
//
// The renderer already degrades an unknown pose to `idle`, silently. That
// silence is the problem the write-back fixes: with the resolved value on the
// plan, the validator's "this beat holds the pose the last one used" sees what
// the viewer will see, rather than two different invented names that both end
// up as a character standing still for two beats.
func normalizeCastPlan(p *SnippetPlan) {
	for i := range p.Beats {
		b := &p.Beats[i]
		if len(strings.Fields(b.Heading)) < minHeadlineWords {
			b.Heading = headingFromNarration(b.Narration)
		}
		b.Heading = clampWords(b.Heading, maxHeadlineWords)
		if b.Cast == nil {
			continue
		}
		b.Cast.Pose = normalizeCastPose(b.Cast.Pose)
		b.Cast.Expression = normalizeCastExpression(b.Cast.Expression)
		b.Cast.Caption = collapseSpaces(b.Cast.Caption)
		if b.Cast.Prop != "" {
			b.Cast.Prop = normalizeArtFigure(b.Cast.Prop)
		}
	}
}

func validateCastPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Cast: true}); err != nil {
		return err
	}
	poses := map[string]int{}
	var prev string
	for _, b := range p.Beats {
		if b.Cast == nil {
			return fmt.Errorf("beat %q has no cast direction — every beat in this template is a shot", b.ID)
		}
		n := len(strings.Fields(b.Heading))
		if n < minHeadlineWords || n > maxHeadlineWords {
			return fmt.Errorf("beat %q has a %d-word heading; headlines are %d-%d words",
				b.ID, n, minHeadlineWords, maxHeadlineWords)
		}
		if n := len(strings.Fields(b.Cast.Caption)); n > maxCaptionWords {
			return fmt.Errorf("beat %q has a %d-word caption; at most %d", b.ID, n, maxCaptionWords)
		}
		pose := normalizeCastPose(b.Cast.Pose)
		// Consecutive identical poses are the failure this template has to
		// avoid: the character freezes for two beats and the clip becomes a
		// still image of a person with the text changing beside them. The
		// whole reason for a rig is that the pose is what moves.
		if pose == prev {
			return fmt.Errorf("beat %q holds the %q pose that the previous beat already used — give the character something to do on every beat",
				b.ID, pose)
		}
		prev = pose
		poses[pose]++
	}
	return nil
}
