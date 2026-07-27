package pipeline

// The cast template.
//
// A character explains something: they stand beside the headline, hold a pose
// that carries an attitude to it, and change pose between beats.
//
// It is the illustration template's sibling — same one-beat-one-shot shape,
// same kinetic headline — and the difference is the point. An object can show
// what a thing *is*; only a person can show how to *feel* about it. "This is
// the problem" is a slumped figure, "I'm not sure" is a shrug, "here it is" is
// a point. That register is what an explainer opens and closes on, and neither
// the diagram nor the board nor a floating rocket can do it.
//
// As everywhere else in the catalog, the model does not draw. It picks a pose
// and an expression from closed vocabularies and the rig (renderer's cast.tsx)
// does forward kinematics. That is what makes a pose *change* possible at all:
// the poses interpolate, so the character moves from thinking to pointing
// instead of cutting between two drawings of themselves.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "cast",
		Title:       "Explainer character",
		Description: "A person who reacts as they explain — poses, expressions, and a headline beside them.",
		Example:     "Why code review makes teams faster, not slower",
		PromptFile:  snippetCastTemplateName,
		NeedsCode:   false,
		Validate:    validateCastPlan,
		Scenes:      castScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Poses":            strings.Join(CastPoseNames(), ", "),
				"Expressions":      strings.Join(CastExpressionNames(), ", "),
				"Props":            strings.Join(ArtFigureNames(), ", "),
				"MinHeadlineWords": minHeadlineWords,
				"MaxHeadlineWords": maxHeadlineWords,
				"MaxCaptionWords":  maxCaptionWords,
			}
		},
	})
}

const snippetCastTemplateName = "snippet_cast.tmpl"

// castPoseVocab mirrors POSES in renderer/src/components/cast.tsx; a drift test
// keeps the two identical. A pose Go allows and the rig does not have would
// silently fall back to `idle`, so the character would simply stand there
// through the beat that was supposed to be its punchline.
var castPoseVocab = map[string]bool{
	"idle":      true,
	"point":     true,
	"think":     true,
	"wave":      true,
	"celebrate": true,
	"shrug":     true,
	"confident": true,
	"defeated":  true,
	"walk":      true,
}

// castExpressionVocab mirrors the Expression union in cast.tsx.
var castExpressionVocab = map[string]bool{
	"neutral":   true,
	"happy":     true,
	"thinking":  true,
	"surprised": true,
	"concerned": true,
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
