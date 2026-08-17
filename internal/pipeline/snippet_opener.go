package pipeline

// The opener template: the title IS the picture.
//
// Every intro card in this catalog is built the same way — a kicker, a headline
// at 64pt, a rule under it, and a lot of empty stage around the group. It is a
// perfectly good slide and it is the reason a course's first frame looks like its
// fortieth: the headline is set at the size a *label* is set at, so the frame
// reads as a diagram whose diagram has not loaded yet.
//
// This one sets the title at the size a book sets a title page: enormous, filling
// the frame edge to edge, in a display SERIF, at an opacity low enough that it
// stops being type and becomes the ground. Then the things you actually read —
// what this is, who it is from — sit on top of it, small and solid.
//
// Two decisions carry it, and both are the opposite of what a normal headline
// wants.
//
// THE BIG TYPE IS NOT MEANT TO BE READ. It is at 12% ink. A viewer takes it in as
// texture and shape, the way they take in a watermark on paper, and the words they
// actually read are the small solid ones. That is why the title may be long here
// when every other template caps its headline at seven words — length is what
// fills the frame. A four-word title in this treatment leaves two thirds of the
// page empty and the effect collapses.
//
// AND IT IS A SERIF. The catalog has three families and all three are sans, which
// means every title it has ever set is the same voice at a different size. A high
// contrast serif at 300pt is a different register entirely — it is the one thing
// on this frame doing the work that a designed intro does, and it costs one font.
//
// What it is NOT: a syllabus. There is no list of what the course contains, no
// outcomes, no modules. Those are `syllabus` and `objective`, and an opener that
// grew a bulleted list would be those templates with a bigger headline. This frame
// answers two questions — what is this, and whose is it — and then gets out of the
// way.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "opener",
		Category: CatPresenting,
		Since:    SinceV9,
		Family:   FamilyShowroom,
		Title:    "The title page",
		Description: "The subject set enormous in a display serif as the ground of the frame, with the promise and the byline solid on top of it. " +
			"Reach for it to open a piece — it is a title page, not a contents page.",
		Example:    "Your first prompt in Claude Code",
		PromptFile: snippetOpenerTemplateName,
		NeedsCode:  false,
		// Three beats is the shape: the ground, the promise, the mark. There is no
		// long version of a title page — past about twenty-five seconds an opener
		// is stalling, and the templates that legitimately take longer are the ones
		// that list something.
		MinTargetSec:     12,
		DefaultTargetSec: 20,
		MaxBeats:         4,
		// Deliberately low. A beat here is a held frame with one line landing on
		// it, and the temptation the prompt has to fight is writing a paragraph
		// over a title card.
		IdealWordsPerBeat: 20,
		Owns:              beatFields{Opener: true},
		OwnsPlan:          planFields{Opener: true},
		Normalize:         normalizeOpenerPlan,
		Validate:          validateOpenerPlan,
		Scenes:            openerScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":           strings.Join(OpenerShows(), ", "),
				"MinGroundWords":  minOpenerGroundWords,
				"MaxGroundWords":  maxOpenerGroundWords,
				"MaxPromiseWords": maxOpenerPromiseWords,
				"MaxBylineWords":  maxOpenerBylineWords,
				"MaxKickerWords":  maxOpenerKickerWords,
			}
		},
	})
}

const snippetOpenerTemplateName = "snippet_opener.tmpl"

const (
	// The big type. A floor rather than only a ceiling, and the floor is the
	// unusual part: this treatment needs LENGTH to work, because the words are
	// what fill the frame. Three words at 300pt is a logo, not a title page.
	minOpenerGroundWords = 4
	maxOpenerGroundWords = 9

	// The line that is actually read: what the viewer will be able to do.
	maxOpenerPromiseWords = 16
	// The mark at the foot — a name, a course, a series.
	maxOpenerBylineWords = 5
	// The eyebrow above the promise.
	maxOpenerKickerWords = 4
)

// openerShows is the closed vocabulary of what a beat does.
var openerShows = map[string]bool{
	// The ground alone: the big serif title, nothing on top of it.
	"ground": true,
	// The kicker and the promise land.
	"promise": true,
	// The byline plate lands, and the frame is complete.
	"mark": true,
}

// OpenerShows returns the beat vocabulary sorted.
func OpenerShows() []string {
	out := make([]string, 0, len(openerShows))
	for k := range openerShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// OpenerSpec is the title page.
type OpenerSpec struct {
	// Ground is the big type — the subject, in the viewer's words. Long on
	// purpose: see minOpenerGroundWords.
	Ground string `json:"ground"`
	// Kicker is the eyebrow over the promise. Optional.
	Kicker string `json:"kicker,omitempty"`
	// Promise is the one line a viewer actually reads: what they will be able to
	// do. Required — a title page with no promise is a name on a wall.
	Promise string `json:"promise"`
	// Byline is the mark at the foot: who this is from.
	Byline string `json:"byline,omitempty"`
}

// OpenerBeat is one shot of the title page.
type OpenerBeat struct {
	// Show is an openerShows name.
	Show string `json:"show"`
}

// ResolvedShow defaults the unknown to the promise landing, which is the beat
// this template is really about.
func (b OpenerBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if openerShows[s] {
		return s
	}
	return "promise"
}

func normalizeOpenerPlan(p *SnippetPlan) {
	o := p.Opener
	if o == nil {
		return
	}
	o.Ground = clampWords(collapseSpaces(o.Ground), maxOpenerGroundWords)
	o.Kicker = clampWords(collapseSpaces(o.Kicker), maxOpenerKickerWords)
	o.Promise = clampWords(collapseSpaces(o.Promise), maxOpenerPromiseWords)
	o.Byline = clampWords(collapseSpaces(o.Byline), maxOpenerBylineWords)

	for i := range p.Beats {
		if b := p.Beats[i].Opener; b != nil {
			b.Show = b.ResolvedShow()
		}
	}
}

func validateOpenerPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Opener: true}); err != nil {
		return err
	}

	o := p.Opener
	if o == nil {
		return fmt.Errorf("the plan has no title page — this template is one frame with the subject set enormous behind it, so the opener is the clip")
	}
	words := len(strings.Fields(o.Ground))
	if words < minOpenerGroundWords {
		return fmt.Errorf("the big type is %d word(s): %q. This treatment is filled BY the words — at the size they are set, %d of them cover a third of the frame and the rest is empty page, which reads as a logo somebody forgot to finish rather than as a title. Say the subject the long way round, in %d-%d words",
			words, o.Ground, words, minOpenerGroundWords, maxOpenerGroundWords)
	}
	if strings.TrimSpace(o.Promise) == "" {
		return fmt.Errorf("the title page makes no promise. The big type is at twelve percent ink — it is texture, and a viewer does not read it — so the one line they DO read has to say what they will be able to do. Without it the frame is a name on a wall")
	}
	// The promise repeating the title is the failure this template invites: the
	// big words are right there, and restating them feels like reinforcement.
	if openerFold(o.Promise) == openerFold(o.Ground) {
		return fmt.Errorf("the promise is the same words as the big type. The whole composition is that one thing is texture and the other is read — printing the title twice leaves nothing for the viewer to take away. Say what they will be able to DO")
	}

	if p.Beats[0].Opener == nil || p.Beats[0].Opener.ResolvedShow() != "ground" {
		return fmt.Errorf("beat %q does not open on the ground alone. The big type has to be there before anything sits on it, or the small lines land on an empty page and the frame assembles backwards — open with {\"show\": \"ground\"}", p.Beats[0].ID)
	}

	seen := map[string]int{}
	for _, b := range p.Beats {
		d := b.Opener
		if d == nil {
			return fmt.Errorf("beat %q has no opener direction — every beat raises the ground, lands the promise, or lands the mark", b.ID)
		}
		seen[d.ResolvedShow()]++
	}
	if seen["ground"] > 1 {
		return fmt.Errorf("two beats raise the ground. It is one held frame; raising it twice is a cut to the same picture")
	}
	if seen["promise"] == 0 {
		return fmt.Errorf("no beat lands the promise, so the clip is a title nobody reads and a name at the foot. Add a {\"show\": \"promise\"} beat")
	}
	if seen["promise"] > 1 {
		return fmt.Errorf("two beats land the promise. There is one line to read here — a second one is the paragraph this template exists to refuse")
	}
	if strings.TrimSpace(o.Byline) == "" && seen["mark"] > 0 {
		return fmt.Errorf("a beat lands the mark but there is no byline for it to land. Give the opener a byline, or drop the beat")
	}
	return nil
}

// openerFold reduces a phrase to its letters, so "Your First Prompt" and
// "your first prompt." are the same answer.
func openerFold(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// openerScenes lays the clip out as ONE scene. The ground never moves; the steps
// only say which of the two solid lines are up.
func openerScenes(in SnippetSceneInput) ([]Scene, error) {
	o := in.Plan.Opener
	if o == nil {
		return nil, fmt.Errorf("the plan has no title page")
	}

	// Latched, like the duel's bars: once a line is up it stays up, so the last
	// frame of the clip is the finished title page rather than whatever the final
	// beat happened to be about.
	promise, mark := false, false
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Opener == nil {
			return nil, fmt.Errorf("beat %q has no opener direction", beat.ID)
		}
		show := beat.Opener.ResolvedShow()
		switch show {
		case "promise":
			promise = true
		case "mark":
			mark = true
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"promise": promise,
			"mark":    mark,
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneOpener,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"ground":  o.Ground,
			"kicker":  o.Kicker,
			"promise": o.Promise,
			"byline":  o.Byline,
			"steps":   steps,
		},
	}}, nil
}
