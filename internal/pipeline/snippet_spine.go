package pipeline

// The spine template: the one you reach for when the words come first.
//
// Every other template in the catalog is chosen for what the clip is ABOUT. A
// number gets `gauge`, a choice gets `decision`, an editor gets `vscode`. That
// works for the parts of a course that have a subject, and it leaves the rest
// of a course with nothing — the opening, the hand-off into a chapter, the
// aside that reframes what just happened, the sign-off. Those are most of the
// running time of a real course and their subject is whatever the narrator is
// saying at that moment, which is not a shape any content template has.
//
// So this template's subject is the narration itself. Two ways in:
//
//	prompt    → it writes the narration and stages it, like every other template.
//	narration → YOU write it, word for word, and it only decides the pictures.
//
// The second mode is the point. A creator with a script does not want a model's
// paraphrase of it; they want their sentences, timed to their voice, with
// something worth looking at while each one is spoken. So the words are copied
// through unchanged and checked word for word (validateSpineScript), and the
// runtime is derived FROM the script rather than the script being squeezed into
// a runtime somebody guessed at.
//
// == Shots, not diagrams ==
//
// A beat carries a SHOT: one of twelve arrangements of figures around a line of
// type. `open` is a title that is still moving, `chapter` marks a named part of
// the course, `state` is the workhorse claim, `pair` holds two things against
// each other, `row` lays out three to five, `orbit` hangs parts off a centre,
// `steps` is an order of operations, `recap` ticks off what is already done,
// `aside` is a parenthesis, `focus` is a breath, `quote` is a line landing on
// its own, and `close` ends the clip. The planner picks per beat from what the
// sentence is doing.
//
// Three of those — `chapter`, `recap`, `aside` — are the ones that make this a
// COURSE template rather than a clip template. A clip only ever needs to open,
// explain and close; a course needs to say which part you are in, what you have
// already finished, and when it is stepping sideways for a moment. Those were
// the sentences a spine clip could not stage, and staging them meant cutting
// away to a different template, which is exactly what a backbone is supposed to
// make unnecessary.
//
// The figures come from the vocabulary the illustration template already owns
// (artwork.tsx, ~113 objects) rather than a second set, because the vocabulary
// is the expensive part and a template with its own private one would be a
// second thing to keep in sync with the renderer for no gain.

import (
	"context"
	"fmt"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "spine",
		Category:    CatPresenting,
		Since:       SinceV6,
		Title:       "Narration spine",
		Description: "The backbone. Write the narration — or a prompt — and it stages every line as its own shot: openings, chapter marks, explanations, asides, recaps and sign-offs in one clip.",
		Example:     "Open a chapter on why prompts beat clicking, then hand off to the demo",
		PromptFile:  snippetSpineTemplateName,
		NeedsCode:   false,
		// The one template you can hand a script to. See SnippetSpec.Narration.
		TakesNarration: true,
		// A spine clip is a run of *shots*, and a script long enough to need
		// twelve of them is an ordinary thing for a course opening. Seven is the
		// right ceiling for a clip built from a handful of moments; this one is
		// built from however many sentences the creator wrote.
		MaxBeats: 12,
		// A beat here is a SHOT, and a shot is a thing you cut away from. At the
		// shared forty words a 107-word script came back as three beats — one of
		// them sixty words, which is twenty seconds of a single static
		// arrangement, and the whole clip cut twice. Twenty-four words is about
		// eight seconds, which is the longest a picture with no moving parts
		// holds anybody. It does not shorten the clip; it cuts the same
		// narration into more shots.
		IdealWordsPerBeat: 24,
		Owns:              beatFields{Spine: true},
		Plan:              planSpine,
		// The creator's script has to be on the plan before Validate runs,
		// because Validate runs inside the correction loop — attaching it after
		// the planner returns is attaching it after every judgement was made.
		PreValidate: func(spec SnippetSpec, p *SnippetPlan) { p.spineScript = spec.Narration },
		Normalize:   normalizeSpinePlan,
		Validate:    validateSpinePlan,
		Scenes:      spineScenes,
		PromptData: func(spec SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Figures":        strings.Join(ArtFigureNames(), ", "),
				"Shots":          spineShotHelp(),
				"Script":         strings.TrimSpace(spec.Narration),
				"MaxLabelWords":  maxSpineLabelWords,
				"MaxDetailWords": maxSpineDetailWords,
				"MaxQuoteWords":  maxSpineQuoteWords,
				"MaxOrdinal":     maxSpineOrdinal,
			}
		},
	})
}

const snippetSpineTemplateName = "snippet_spine.tmpl"

// Label and detail bounds. A tile is read at a glance while the narrator keeps
// talking, so its label is a name and not a sentence; the detail line is the
// one place a tile may explain itself, and past ten words it competes with the
// voice for the same half-second.
const (
	maxSpineLabelWords  = 4
	maxSpineDetailWords = 10
	// A `quote` headline is a sentence by definition, so it is the one shot
	// whose type is allowed past the headline ceiling.
	maxSpineQuoteWords = 16
	// maxSpineOrdinal is the highest part number a `chapter` shot may carry.
	// Two digits is what the numeral is designed around; a three-digit part
	// number is not a course, it is a mistake in the plan.
	maxSpineOrdinal = 99
	// maxSpineBeatWords is the longest a single shot may hold the frame.
	//
	// The shared ceiling is sixty, which is right for a beat whose picture is
	// still being drawn under the narration and far too long for one that is
	// already finished. Forty-two words is about fourteen seconds at the house
	// pace — past that the viewer has finished reading, finished looking, and is
	// waiting. IdealWordsPerBeat steers the draft here; this is the edge.
	maxSpineBeatWords = 42
)

// SpineObject is one figure on the stage, with what it is called.
type SpineObject struct {
	// Figure is a name from the shared vocabulary (artFigureVocab). Anything
	// else normalizes to "spark" rather than rendering an empty half-frame.
	Figure string `json:"figure"`
	// Label names it — a noun phrase, not a sentence.
	Label string `json:"label,omitempty"`
	// Detail is the one line a tile may say about itself. Only the shots that
	// draw panels (pair, row, steps) show it.
	Detail string `json:"detail,omitempty"`
}

// SpineBeat is one shot: how this moment is arranged, and what is in it.
type SpineBeat struct {
	// Shot is the layout, from spineShots. An unknown one is repaired to the
	// shot whose object count matches, which is almost always what was meant.
	Shot string `json:"shot"`
	// Note is the small letterspaced line above the headline — a chapter, a
	// section, a cue ("COMING UP", "PART TWO"). Optional and usually absent:
	// on every beat it is chrome, on one beat it is a signpost.
	Note string `json:"note,omitempty"`
	// Emphasis is a phrase copied out of the beat's heading, struck through
	// with the marker. Empty is fine; a phrase that is not in the heading is
	// not, because there would be nothing on screen to draw under.
	Emphasis string `json:"emphasis,omitempty"`
	// Caption is the supporting line under the headline. On `close` it is the
	// next step rather than a supporting point, and it renders as a pill.
	Caption string `json:"caption,omitempty"`
	// Ordinal is the part number a `chapter` shot sets in enormous type behind
	// its title. Only `chapter` has one; it is cleared everywhere else.
	//
	// It cannot be derived. The obvious source is the beat's own index, and
	// that was the first implementation — a chapter marker sitting second in a
	// clip rendered "Part one" under a giant 2, which is a lie the frame tells
	// on its own. Where a clip sits in a course is something only the creator
	// and the planner know.
	Ordinal int `json:"ordinal,omitempty"`
	// Objects are the figures this shot arranges, in reading order. For
	// `orbit` the first is the centre and the rest are its spokes.
	Objects []SpineObject `json:"objects,omitempty"`
}

// spineShot is one layout and what it is for.
//
// The arity is part of the definition rather than a rule applied to it: `pair`
// with three objects is not a loose pair, it is a `row` the planner mislabelled,
// and the repair is to say so rather than to drop the third.
type spineShot struct {
	Name string
	// Min and Max objects the layout draws. For `orbit` this counts the centre.
	Min, Max int
	// Use is the one line the prompt gives the planner. It says when to reach
	// for the shot, not what it looks like — the planner cannot see it, and
	// "when" is the only question it is actually answering.
	Use string
}

// spineShots is the closed vocabulary, in the order a clip usually uses them.
// The renderer mirrors this list (SpineShot in SpineScene.tsx).
var spineShots = []spineShot{
	{"open", 0, 1, "the first beat of a clip: the title, big, with one figure. Only the first beat may be an open."},
	{"chapter", 0, 5, "a section marker — the clip is entering a named part of the course. Set `ordinal` to the part number; it is set enormous behind the title. The objects are what that part covers, drawn as small chips."},
	{"state", 1, 1, "the workhorse — one claim, one figure beside it, one supporting line. Most beats are this."},
	{"pair", 2, 2, "two things held against each other: before and after, theirs and ours, wrong and right."},
	{"row", 3, 5, "three to five things laid out together: what a chapter covers, the tools involved, the parts of a whole."},
	{"orbit", 3, 6, "one idea with things hanging off it. The FIRST object is the centre; the rest orbit it."},
	{"steps", 2, 5, "an order of operations — numbered, connected, one after another in the order they happen."},
	{"recap", 2, 5, "looking back at things already covered, each one ticked off. Use it AFTER the work, never before — the tick means done."},
	{"aside", 0, 1, "a parenthesis: a remark the clip steps sideways for and then steps back from. Set smaller and indented behind a bar. At most one per clip."},
	{"focus", 1, 1, "a breath. One object very large and as few words as the sentence allows. Use it between two dense beats."},
	{"quote", 0, 0, "a line that has to land on its own. No figure at all — an object beside a punchline is what stops it being one."},
	{"close", 0, 1, "the last beat: the sign-off, centred, with the caption as the next step. Only the last beat may be a close."},
}

var spineShotByName = func() map[string]spineShot {
	m := make(map[string]spineShot, len(spineShots))
	for _, s := range spineShots {
		m[s.Name] = s
	}
	return m
}()

// SpineShotNames returns the shot vocabulary in catalog order.
func SpineShotNames() []string {
	out := make([]string, 0, len(spineShots))
	for _, s := range spineShots {
		out = append(out, s.Name)
	}
	return out
}

// spineShotHelp renders the vocabulary for the prompt, one shot per line.
func spineShotHelp() string {
	var b strings.Builder
	for _, s := range spineShots {
		objects := fmt.Sprintf("%d objects", s.Min)
		switch {
		case s.Min != s.Max:
			objects = fmt.Sprintf("%d-%d objects", s.Min, s.Max)
		case s.Min == 0:
			objects = "no objects"
		case s.Min == 1:
			objects = "1 object"
		}
		fmt.Fprintf(&b, "- %s (%s): %s\n", s.Name, objects, s.Use)
	}
	return strings.TrimRight(b.String(), "\n")
}

// planSpine sizes the clip to the creator's script, then plans it normally.
//
// The runtime has to follow the words rather than the other way round. A
// creator who wrote 300 words and left target_sec alone would otherwise get the
// default 45-second budget, which asks for about 112 words — and the model,
// told to keep every word AND hit the budget, is being given two instructions
// that cannot both be obeyed. Every downstream bound (beat count, words per
// beat) is derived from that budget, so setting it correctly here is what makes
// the rest of the machinery agree with the script that actually exists.
func planSpine(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) (*SnippetPlan, error) {
	if script := strings.TrimSpace(spec.Narration); script != "" {
		spec.TargetSec = scriptTargetSec(script, cfg.Style.PaceWPM)
	}
	return planSnippetDefault(ctx, e, spec, cfg)
}

// normalizeSpineShot repairs a shot name.
//
// An unknown name is resolved by object count rather than falling back to a
// fixed default, because the count is the part the planner got right: it laid
// out three objects and called the arrangement something this catalog does not
// have. Mapping that to `state` would throw two of the three away; mapping it
// to `row` keeps the shot it planned.
func normalizeSpineShot(name string, objects int) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if _, ok := spineShotByName[n]; ok {
		return n
	}
	switch objects {
	case 0:
		return "quote"
	case 1:
		return "state"
	case 2:
		return "pair"
	default:
		return "row"
	}
}

func normalizeSpinePlan(p *SnippetPlan) {
	for i := range p.Beats {
		b := &p.Beats[i]
		if b.Spine == nil {
			continue
		}
		s := b.Spine
		s.Shot = normalizeSpineShot(s.Shot, len(s.Objects))
		s.Note = collapseSpaces(s.Note)
		s.Caption = collapseSpaces(s.Caption)
		s.Emphasis = collapseSpaces(s.Emphasis)
		// Only a chapter marker has a part number. A planner that puts one on
		// an ordinary beat meant nothing by it, and the renderer would draw a
		// four-hundred-pixel numeral behind a claim.
		if s.Shot != "chapter" || s.Ordinal < 0 {
			s.Ordinal = 0
		}
		if len(strings.Fields(b.Heading)) < minHeadlineWords {
			b.Heading = headingFromNarration(b.Narration)
		}
		b.Heading = clampWords(b.Heading, spineHeadlineMax(s.Shot))
		// The emphasis is a stroke drawn under part of the heading, so a phrase
		// that is not in the heading has nothing to draw under. Almost always
		// the planner quoting the idea rather than the words; clearing it costs
		// the shot its accent, rejecting it costs a round to be told a fact
		// about string matching.
		if s.Emphasis != "" && !containsPhrase(b.Heading, s.Emphasis) {
			s.Emphasis = ""
		}
		// A shot draws what its layout has room for. Extra objects are dropped
		// rather than rejected — the planner was right about the arrangement
		// and generous about the contents, and there is nothing to be learned
		// from a round spent on that.
		if shot, ok := spineShotByName[s.Shot]; ok && len(s.Objects) > shot.Max {
			s.Objects = s.Objects[:shot.Max]
		}
		for j := range s.Objects {
			o := &s.Objects[j]
			o.Figure = normalizeArtFigure(o.Figure)
			o.Label = clampWords(collapseSpaces(o.Label), maxSpineLabelWords)
			o.Detail = clampWords(collapseSpaces(o.Detail), maxSpineDetailWords)
		}
	}
}

// spineHeadlineMax is the headline ceiling for a shot. A `quote` is a sentence
// landing on its own, which is what the shot is for; every other shot's
// headline is a headline.
func spineHeadlineMax(shot string) int {
	if shot == "quote" {
		return maxSpineQuoteWords
	}
	return maxHeadlineWords
}

func validateSpinePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Spine: true}); err != nil {
		return err
	}
	last := len(p.Beats) - 1
	shots := map[string]int{}
	var run struct {
		shot string
		n    int
	}
	for i, b := range p.Beats {
		if b.Spine == nil {
			return fmt.Errorf("beat %q has no shot — every beat in this template is a shot, with a `spine` object saying which", b.ID)
		}
		s := b.Spine
		shot, ok := spineShotByName[s.Shot]
		if !ok {
			return fmt.Errorf("beat %q has shot %q, which is not one of: %s",
				b.ID, s.Shot, strings.Join(SpineShotNames(), ", "))
		}
		// An open in the middle of a clip is a second title card; a close in
		// the middle ends the clip and then keeps going. Both read as an edit
		// mistake rather than as a choice.
		if s.Shot == "open" && i != 0 {
			return fmt.Errorf("beat %q is an `open` but it is beat %d — only the first beat may open the clip", b.ID, i+1)
		}
		if s.Shot == "close" && i != last {
			return fmt.Errorf("beat %q is a `close` but it is beat %d of %d — only the last beat may close the clip", b.ID, i+1, len(p.Beats))
		}
		if n := len(s.Objects); n < shot.Min || n > shot.Max {
			return fmt.Errorf("beat %q is a %q shot with %d objects; that shot takes %d-%d",
				b.ID, s.Shot, n, shot.Min, shot.Max)
		}
		n := len(strings.Fields(b.Heading))
		if n < minHeadlineWords || n > spineHeadlineMax(s.Shot) {
			return fmt.Errorf("beat %q has a %d-word heading; a %s headline is %d-%d words",
				b.ID, n, s.Shot, minHeadlineWords, spineHeadlineMax(s.Shot))
		}
		if e := strings.TrimSpace(s.Emphasis); e != "" && !containsPhrase(b.Heading, e) {
			return fmt.Errorf("beat %q emphasises %q, which does not appear in its heading %q — copy the phrase across exactly",
				b.ID, e, b.Heading)
		}
		if n := len(strings.Fields(s.Caption)); n > maxCaptionWords {
			return fmt.Errorf("beat %q has a %d-word caption; at most %d", b.ID, n, maxCaptionWords)
		}
		// A chapter marker with no part number is a title card that has
		// forgotten what it is marking — the numeral IS the shot.
		if s.Shot == "chapter" && (s.Ordinal < 1 || s.Ordinal > maxSpineOrdinal) {
			return fmt.Errorf("beat %q is a `chapter` with ordinal %d; set it to the part number this section is, between 1 and %d",
				b.ID, s.Ordinal, maxSpineOrdinal)
		}
		// An aside is an interruption. Two of them is not an interruption, it
		// is the register the clip is now in — and the shot is drawn smaller
		// and quieter than everything around it, so a clip made of them
		// recedes with nothing left to recede from.
		if s.Shot == "aside" && shots["aside"] > 0 {
			return fmt.Errorf("beat %q is the second `aside` in this clip; a clip gets one — the shot is a step sideways, and it only reads as one if the clip steps back", b.ID)
		}
		// Two tiles showing the same drawing is a layout that has stopped
		// distinguishing the things it is laying out — the exact failure a row
		// of five exists to avoid.
		seen := map[string]bool{}
		for _, o := range s.Objects {
			f := normalizeArtFigure(o.Figure)
			if seen[f] {
				return fmt.Errorf("beat %q uses the %q figure twice in one shot; every object in a shot needs its own figure", b.ID, f)
			}
			seen[f] = true
			if w := len(strings.Fields(o.Label)); w > maxSpineLabelWords {
				return fmt.Errorf("beat %q has a %d-word label %q; labels are at most %d words", b.ID, w, o.Label, maxSpineLabelWords)
			}
			if w := len(strings.Fields(o.Detail)); w > maxSpineDetailWords {
				return fmt.Errorf("beat %q has a %d-word detail line; at most %d", b.ID, w, maxSpineDetailWords)
			}
		}
		// A shot the narration sits on for fourteen seconds is a still image
		// with a voiceover, whatever it is arranged as.
		if w := len(strings.Fields(b.Narration)); w > maxSpineBeatWords {
			return fmt.Errorf("beat %q holds one shot for %d words; a shot may carry at most %d — split it into two beats and cut between them",
				b.ID, w, maxSpineBeatWords)
		}
		shots[s.Shot]++
		if s.Shot == run.shot {
			run.n++
		} else {
			run.shot, run.n = s.Shot, 1
		}
		if run.n > 2 {
			return fmt.Errorf("beats up to %q are the third %q shot in a row; vary the shot — a run of one layout is a slideshow with the words swapped out", b.ID, s.Shot)
		}
	}
	// A long clip that never changes its arrangement is the failure this
	// template exists to avoid: the whole reason a beat carries a shot rather
	// than a fixed layout is that a course opening does not look like the aside
	// three sentences later. Short clips are exempt — two beats cannot be
	// expected to show three arrangements.
	if len(p.Beats) >= 5 && len(shots) < 3 {
		return fmt.Errorf("this clip has %d beats but only %d different shots; use at least three of: %s",
			len(p.Beats), len(shots), strings.Join(SpineShotNames(), ", "))
	}
	return validateSpineScript(p)
}

// validateSpineScript enforces the promise of the write-your-own mode: these
// are YOUR words.
//
// Compared as normalized word sequences rather than as strings, because the
// beats are the script cut into pieces and every join loses whitespace and
// gains punctuation. What may not change is which words are said and in what
// order. A model that "improves" a sentence has broken the one guarantee this
// mode exists to make, and a clip that quietly says something the creator did
// not write is worse than a clip that fails to plan.
func validateSpineScript(p *SnippetPlan) error {
	if strings.TrimSpace(p.spineScript) == "" {
		return nil
	}
	want := scriptWords(p.spineScript)
	var got []string
	for _, b := range p.Beats {
		got = append(got, scriptWords(b.Narration)...)
	}
	for i := 0; i < len(want) && i < len(got); i++ {
		if want[i] != got[i] {
			return fmt.Errorf("the narration must be the creator's script word for word, and it diverges at word %d: the script says %q, your beats say %q. Split the script into beats — do not rewrite it, tighten it, or fix its punctuation",
				i+1, scriptContext(want, i), scriptContext(got, i))
		}
	}
	switch {
	case len(got) < len(want):
		return fmt.Errorf("the beats are missing the last %d words of the creator's script, starting at %q. Every word has to be spoken — add beats until the script is used up",
			len(want)-len(got), scriptContext(want, len(got)))
	case len(got) > len(want):
		return fmt.Errorf("the beats add %d words the creator did not write, starting at %q. Say exactly what the script says and nothing else",
			len(got)-len(want), scriptContext(got, len(want)))
	}
	return nil
}

// scriptWords reduces narration to comparable words: lower case, letters and
// digits only, so a curly apostrophe or a full stop is not a difference.
func scriptWords(s string) []string {
	fields := strings.Fields(s)
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if w := phraseKey(f); w != "" {
			out = append(out, w)
		}
	}
	return out
}

// scriptContext quotes a few words around i, so a divergence report points at
// somewhere in the script the creator can actually find.
func scriptContext(words []string, i int) string {
	if i >= len(words) {
		return "(the end)"
	}
	end := min(i+6, len(words))
	return strings.Join(words[i:end], " ")
}

// spineScenes lays the clip out as one shot per beat.
//
// No opening title card: the first beat's `open` shot IS the title card, and it
// is already big type on an empty stage. Cutting to a separate card first would
// say the same thing twice before the clip has said anything.
func spineScenes(in SnippetSceneInput) ([]Scene, error) {
	scenes := make([]Scene, 0, len(in.Plan.Beats))
	total := len(in.Plan.Beats)
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Spine == nil {
			return nil, fmt.Errorf("beat %q has no shot", beat.ID)
		}
		objects := make([]map[string]any, 0, len(beat.Spine.Objects))
		for _, o := range beat.Spine.Objects {
			objects = append(objects, map[string]any{
				"figure": normalizeArtFigure(o.Figure),
				"label":  o.Label,
				"detail": o.Detail,
			})
		}
		scenes = append(scenes, Scene{
			Type:    SceneSpine,
			StartMs: startMs,
			EndMs:   endMs,
			Props: map[string]any{
				"shot":     beat.Spine.Shot,
				"headline": beat.Heading,
				"emphasis": beat.Spine.Emphasis,
				"caption":  beat.Spine.Caption,
				"note":     beat.Spine.Note,
				"ordinal":  beat.Spine.Ordinal,
				"objects":  objects,
				// Where this beat sits in the clip. The rail is drawn from it,
				// and it is the one prop that is about the clip rather than
				// about the beat — which is exactly why the scene cannot
				// compute it for itself.
				"index": i,
				"total": total,
			},
		})
	}
	return scenes, nil
}
