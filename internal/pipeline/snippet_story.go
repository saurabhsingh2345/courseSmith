package pipeline

// The story template: a directed, one-to-two-minute piece.
//
// Every other template in the catalog is one idea rendered one way. This one is
// a *film* — long enough to have an arc, and staged shot by shot: a character
// who acts, objects that come and go, a camera that moves, and a cut that
// changes framing rather than merely swapping content.
//
// == Why the plan is two LLM calls ==
//
// The writer and the director are different jobs, and asking one call to do both
// produces neither. A single call that has to invent narration *and* stage it
// spends its attention on the words and stages everything identically —
// fourteen beats of "character on the left, object on the right", which is a
// slideshow with a person in it.
//
// So: call one writes the whole script and nothing else. Call two is handed the
// finished script — every beat at once, in order — and does only staging. That
// second call can see the arc, which is the entire point: you cannot decide
// that *this* is the beat to push in close unless you can see the beats on
// either side of it. Variety rules (below) are enforceable for the same reason.
//
// == Why the model does not place anything ==
//
// The same trade as the whiteboard and the flow diagram. The director picks a
// staging and a camera move from closed vocabularies; the renderer owns every
// coordinate. A model given x/y positions produces characters standing on their
// own captions. A model given "duo" and "push" produces a shot that is always
// composed, and the vocabulary is small enough that its choices can be checked.

import (
	"context"
	"fmt"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "story",
		Title:       "Directed short",
		Description: "A one-to-two-minute piece: a character, objects, a moving camera, and a shot list with an arc.",
		Example:     "How a database index actually finds your row",
		PromptFile:  snippetStoryScriptTemplateName,
		NeedsCode:   false,
		// Ninety seconds: eight beats is the floor, and at the shared 45s
		// budget eight beats could only be funded by writing every one of them
		// at the ten-word minimum.
		DefaultTargetSec: 90,
		// Eight beats is the floor; at 175 wpm a minute is the least that funds
		// them at a length worth watching.
		MinTargetSec: 60,
		Plan:         planStorySnippet,
		Validate:     validateStoryPlan,
		Scenes:       storyScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"MinStoryBeats": minStoryBeats,
				"MaxStoryBeats": maxStoryBeats,
			}
		},
	})
}

const (
	snippetStoryScriptTemplateName = "story_script.tmpl"
	snippetStoryShotsTemplateName  = "story_shots.tmpl"
)

// Story length. A piece this long needs enough beats to have an arc — under
// eight and it is a snippet that outstays its welcome — and past fourteen the
// director stage starts repeating itself and the clip runs past two minutes
// whatever runtime was asked for.
const (
	minStoryBeats = 8
	maxStoryBeats = 14
)

// stagingVocab is where things stand. The renderer owns the coordinates; this
// is only the choice of arrangement.
var stagingVocab = map[string]bool{
	// The character alone, centre stage, large.
	"hero": true,
	// The character on one side, an object on the other — the workhorse.
	"duo": true,
	// One object alone, large. No character.
	"object": true,
	// Two objects side by side: a comparison, a before and after.
	"pair": true,
	// Type alone on an empty stage. The beat that needs no illustration.
	"empty": true,
}

// cameraVocab is how the shot moves. Every one of these is a deterministic
// transform over the shot's own duration, so a move never depends on the
// wall clock or on what the previous shot did.
var cameraVocab = map[string]bool{
	"hold":  true, // locked off
	"push":  true, // slow zoom in — lands on an idea
	"pull":  true, // zoom out — reveals context
	"pan":   true, // drift sideways
	"rise":  true, // drift up
	"drift": true, // a slow diagonal, for a held thought
}

func StoryStagingNames() []string { return sortedKeys(stagingVocab) }
func StoryCameraNames() []string  { return sortedKeys(cameraVocab) }

func normalizeStaging(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))
	if stagingVocab[n] {
		return n
	}
	return "empty"
}

func normalizeCamera(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))
	if cameraVocab[n] {
		return n
	}
	return "hold"
}

// StoryScript is what the writer call returns: the piece, with no staging.
type StoryScript struct {
	Title    string      `json:"title"`
	Subtitle string      `json:"subtitle,omitempty"`
	Beats    []StoryLine `json:"beats"`
	// Logline is one sentence naming what the piece is about. It is not shown
	// anywhere — it exists so the director call has the writer's own statement
	// of intent rather than having to infer it from fourteen beats.
	Logline string `json:"logline,omitempty"`
}

// StoryLine is one written beat, before it has been staged.
type StoryLine struct {
	ID        string `json:"id"`
	Heading   string `json:"heading"`
	Narration string `json:"narration"`
}

// StoryShots is what the director call returns: one shot per beat, keyed by
// the beat id so a reordered or partial reply is caught rather than silently
// mis-applied.
type StoryShots struct {
	Shots []ShotBeat `json:"shots"`
}

// planStorySnippet runs the writer, then the director.
//
// The two calls are cached independently, which matters more than it sounds:
// iterating on the director prompt (where most of the tuning happens) does not
// re-pay for the script, and a script that survives a director rewrite keeps
// its narration audio unchanged downstream.
func planStorySnippet(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) (*SnippetPlan, error) {
	script, err := planStoryScript(ctx, e, spec, cfg)
	if err != nil {
		return nil, err
	}
	shots, err := planStoryShots(ctx, e, spec, cfg, script)
	if err != nil {
		return nil, err
	}

	byID := make(map[string]ShotBeat, len(shots))
	for _, s := range shots {
		byID[s.BeatID] = s
	}
	plan := &SnippetPlan{
		Template: spec.Template,
		Title:    script.Title,
		Subtitle: script.Subtitle,
		Beats:    make([]SnippetBeat, 0, len(script.Beats)),
	}
	for _, line := range script.Beats {
		shot, ok := byID[line.ID]
		if !ok {
			return nil, fmt.Errorf("the shot list has no shot for beat %q — every written beat must be staged", line.ID)
		}
		s := shot
		plan.Beats = append(plan.Beats, SnippetBeat{
			ID:        line.ID,
			Heading:   line.Heading,
			Narration: line.Narration,
			Shot:      &s,
		})
	}
	if err := plan.Validate(); err != nil {
		return nil, fmt.Errorf("staged story is invalid: %w", err)
	}
	return plan, nil
}

// planStoryScript is call one: the words, and only the words.
func planStoryScript(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) (*StoryScript, error) {
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	target := spec.ResolvedTargetSec()
	wantWords, minWords, maxWords := wordBudget(target, pace)

	data := sharedPromptData(spec, cfg)
	data["MinStoryBeats"] = minStoryBeats
	data["MaxStoryBeats"] = maxStoryBeats
	// A story writes 8-14 beats, not the shared 2-7, so the shared per-beat
	// arithmetic is wrong here by exactly the factor between those ranges —
	// and a per-beat number that does not multiply up to the budget is the
	// failure this whole calculation exists to avoid. Recomputed against the
	// range this template actually enforces.
	storyBeats := min(max(wantWords/idealWordsPerBeat, minStoryBeats), maxStoryBeats)
	data["SuggestBeats"] = storyBeats
	data["WordsPerBeat"] = wantWords / storyBeats

	system, user, err := e.renderPrompt(snippetStoryScriptTemplateName, data)
	if err != nil {
		return nil, err
	}
	var script StoryScript
	// A two-minute script is several times the length of a snippet's, so the
	// token ceiling has to rise with it or the reply is truncated mid-beat and
	// comes back as a JSON parse error that looks like a model failure.
	err = e.completeJSONRounds(ctx, cfg.Pipeline, llm.TaskContent, system, user, nil, 0.6, 12288, snippetPlanRepairRounds, &script, func() error {
		if n := len(script.Beats); n < minStoryBeats || n > maxStoryBeats {
			return fmt.Errorf("script has %d beats, want %d-%d", n, minStoryBeats, maxStoryBeats)
		}
		seen := map[string]bool{}
		total := 0
		for i, b := range script.Beats {
			if strings.TrimSpace(b.ID) == "" {
				return fmt.Errorf("beat %d has an empty id", i)
			}
			if seen[b.ID] {
				return fmt.Errorf("duplicate beat id %q", b.ID)
			}
			seen[b.ID] = true
			if strings.TrimSpace(b.Heading) == "" {
				return fmt.Errorf("beat %q has an empty heading", b.ID)
			}
			n := len(strings.Fields(b.Narration))
			if n < minWordsPerBeat || n > maxWordsPerBeat {
				return fmt.Errorf("beat %q has %d words of narration, want %d-%d", b.ID, n, minWordsPerBeat, maxWordsPerBeat)
			}
			total += n
		}
		if total < minWords || total > maxWords {
			return fmt.Errorf(
				"narration totals %d words but a %ds piece needs %d-%d (aim for %d) — adjust the beats you have rather than adding past %d",
				total, target, minWords, maxWords, wantWords, maxStoryBeats)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("writing story script: %w", err)
	}
	return &script, nil
}

// planStoryShots is call two: the direction.
//
// It is handed the whole script — every beat, in order, with its narration —
// precisely so it can stage for contrast. The variety rules below are only
// answerable by something that can see the neighbours.
func planStoryShots(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config, script *StoryScript) ([]ShotBeat, error) {
	var b strings.Builder
	for i, line := range script.Beats {
		fmt.Fprintf(&b, "%d. id=%s  heading=%q\n   %s\n", i+1, line.ID, line.Heading, line.Narration)
	}

	data := sharedPromptData(spec, cfg)
	data["StoryTitle"] = script.Title
	data["Logline"] = script.Logline
	data["Beats"] = b.String()
	data["BeatCount"] = len(script.Beats)
	data["Stagings"] = strings.Join(StoryStagingNames(), ", ")
	data["Cameras"] = strings.Join(StoryCameraNames(), ", ")
	data["Poses"] = strings.Join(CastPoseNames(), ", ")
	data["Expressions"] = strings.Join(CastExpressionNames(), ", ")
	data["Props"] = strings.Join(ArtFigureNames(), ", ")
	data["MinCharacterShots"] = minCharacterShots(len(script.Beats))

	system, user, err := e.renderPrompt(snippetStoryShotsTemplateName, data)
	if err != nil {
		return nil, err
	}
	ids := map[string]bool{}
	for _, line := range script.Beats {
		ids[line.ID] = true
	}
	var shots StoryShots
	err = e.completeJSONRounds(ctx, cfg.Pipeline, llm.TaskContent, system, user, nil, 0.7, 8192, snippetPlanRepairRounds, &shots, func() error {
		if len(shots.Shots) != len(script.Beats) {
			return fmt.Errorf("got %d shots for %d beats — stage every beat exactly once", len(shots.Shots), len(script.Beats))
		}
		seen := map[string]bool{}
		for _, s := range shots.Shots {
			if !ids[s.BeatID] {
				return fmt.Errorf("shot names beat %q, which is not in the script", s.BeatID)
			}
			if seen[s.BeatID] {
				return fmt.Errorf("beat %q is staged twice", s.BeatID)
			}
			seen[s.BeatID] = true
		}
		return checkShotList(shots.Shots)
	})
	if err != nil {
		return nil, fmt.Errorf("directing story: %w", err)
	}
	return shots.Shots, nil
}

// minCharacterShots is how many shots must actually contain the character.
//
// Without a floor the director reliably stages everything as `object` and
// `empty` — those are easier to justify per-beat — and the result is the
// illustration template with extra steps. A third is enough that the piece has
// a presenter without demanding one in every shot.
func minCharacterShots(beats int) int {
	n := beats / 3
	if n < 2 {
		n = 2
	}
	return n
}

// checkShotList enforces the rules that make a shot list read as directed
// rather than generated. Each is here because its absence has a name:
// a slideshow, a locked-off camera, or a template with a person in it.
func checkShotList(shots []ShotBeat) error {
	var prev *ShotBeat
	moving := 0
	character := 0
	stagings := map[string]int{}
	for i := range shots {
		s := &shots[i]
		staging := normalizeStaging(s.Staging)
		camera := normalizeCamera(s.Camera)
		stagings[staging]++
		if camera != "hold" {
			moving++
		}
		if staging == "hero" || staging == "duo" {
			character++
			if strings.TrimSpace(s.Pose) == "" {
				return fmt.Errorf("shot %q is staged %q but gives the character no pose", s.BeatID, staging)
			}
		}
		switch staging {
		case "duo", "object":
			if strings.TrimSpace(s.Prop) == "" {
				return fmt.Errorf("shot %q is staged %q but names no prop", s.BeatID, staging)
			}
		case "pair":
			if strings.TrimSpace(s.Prop) == "" || strings.TrimSpace(s.PropB) == "" {
				return fmt.Errorf("shot %q is staged \"pair\" but does not name two props (prop and prop_b)", s.BeatID)
			}
			if normalizeArtFigure(s.Prop) == normalizeArtFigure(s.PropB) {
				return fmt.Errorf("shot %q pairs %q with itself — a comparison needs two different objects", s.BeatID, s.Prop)
			}
		}
		// Consecutive shots that match on both axes are the same picture twice.
		// Matching on one is fine and often right: holding a staging across a
		// cut while the camera changes is how a scene is built.
		if prev != nil {
			if normalizeStaging(prev.Staging) == staging && normalizeCamera(prev.Camera) == camera {
				return fmt.Errorf("shot %q repeats the previous shot's staging (%s) and camera (%s) — change at least one across every cut",
					s.BeatID, staging, camera)
			}
		}
		prev = s
	}
	if character < minCharacterShots(len(shots)) {
		return fmt.Errorf("only %d of %d shots put the presenter on stage (staging hero or duo); want at least %d — this template is presented by someone",
			character, len(shots), minCharacterShots(len(shots)))
	}
	if moving*2 < len(shots) {
		return fmt.Errorf("only %d of %d shots move the camera; at least half should be something other than \"hold\"",
			moving, len(shots))
	}
	if len(stagings) < 3 {
		return fmt.Errorf("the shot list uses only %d stagings (%v) across %d shots — a piece this long needs more than that",
			len(stagings), sortedKeys(boolSet(stagings)), len(shots))
	}
	return nil
}

func boolSet(m map[string]int) map[string]bool {
	out := make(map[string]bool, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}

// storyScenes lays the piece out as one scene per shot.
func storyScenes(in SnippetSceneInput) ([]Scene, error) {
	scenes := make([]Scene, 0, len(in.Plan.Beats))
	prevPose := "idle"
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Shot == nil {
			return nil, fmt.Errorf("beat %q has no shot", beat.ID)
		}
		staging := normalizeStaging(beat.Shot.Staging)
		pose := normalizeCastPose(beat.Shot.Pose)
		props := map[string]any{
			"headline":   beat.Heading,
			"caption":    beat.Shot.Caption,
			"staging":    staging,
			"camera":     normalizeCamera(beat.Shot.Camera),
			"pose":       pose,
			"prevPose":   prevPose,
			"expression": normalizeCastExpression(beat.Shot.Expression),
			// The renderer eases the camera over the shot's own length, so it
			// has to be told what that is — a scene component only knows the
			// frame it is on.
			"durationMs": endMs - startMs,
			"index":      i,
		}
		if beat.Shot.Prop != "" {
			props["prop"] = normalizeArtFigure(beat.Shot.Prop)
		}
		if beat.Shot.PropB != "" {
			props["propB"] = normalizeArtFigure(beat.Shot.PropB)
		}
		scenes = append(scenes, Scene{
			Type: SceneStory, StartMs: startMs, EndMs: endMs, Props: props,
		})
		if staging == "hero" || staging == "duo" {
			prevPose = pose
		}
	}
	return scenes, nil
}

func validateStoryPlan(p *SnippetPlan) error {
	if n := len(p.Beats); n < minStoryBeats || n > maxStoryBeats {
		return fmt.Errorf("story has %d beats, want %d-%d", n, minStoryBeats, maxStoryBeats)
	}
	for _, b := range p.Beats {
		n := len(strings.Fields(b.Narration))
		if n < minWordsPerBeat || n > maxWordsPerBeat {
			return fmt.Errorf("beat %q has %d words of narration, want %d-%d",
				b.ID, n, minWordsPerBeat, maxWordsPerBeat)
		}
	}
	if err := rejectForeignBeatFields(p, beatFields{Shot: true}); err != nil {
		return err
	}
	shots := make([]ShotBeat, 0, len(p.Beats))
	for _, b := range p.Beats {
		if b.Shot == nil {
			return fmt.Errorf("beat %q has no shot — every beat in this template is staged", b.ID)
		}
		if n := len(strings.Fields(b.Heading)); n < minHeadlineWords || n > maxHeadlineWords {
			return fmt.Errorf("beat %q has a %d-word heading; headlines are %d-%d words",
				b.ID, n, minHeadlineWords, maxHeadlineWords)
		}
		if n := len(strings.Fields(b.Shot.Caption)); n > maxCaptionWords {
			return fmt.Errorf("beat %q has a %d-word caption; at most %d", b.ID, n, maxCaptionWords)
		}
		s := *b.Shot
		s.BeatID = b.ID
		shots = append(shots, s)
	}
	return checkShotList(shots)
}
