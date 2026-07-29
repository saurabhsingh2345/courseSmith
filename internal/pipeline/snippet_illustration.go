package pipeline

// The illustration template.
//
// Kinetic typography: a short phrase lands on screen word by word, one word of
// it carries a marker stroke, and a flat-vector figure assembles beside it. The
// side alternates beat to beat, and the clip cuts between shots.
//
// This is the one template in the catalog that does *not* accumulate. The
// whiteboard and the flow diagram are both about a picture being built and
// staying built; this one is about a phrase landing, and a phrase cannot land
// on a stage still holding the last four. Making a beat a shot is the whole
// design, and it is why this template suits the openings, the framings and the
// payoffs that the other two are bad at — the parts of an explainer that are
// rhetoric rather than architecture.
//
// As with the whiteboard, the model does not draw. It picks a figure from a
// closed vocabulary and the renderer assembles it from parts (artwork.tsx).
// Bundling a CC0 illustration set was the obvious alternative and it loses the
// thing that matters: a downloaded SVG is one flat blob, so its parts cannot
// move, and a figure that assembles and then freezes is a slide.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "illustration",
		Category:    CatConcepts,
		Title:       "Kinetic type",
		Description: "Big animated type beside flat-vector artwork. Reach for it for a hook, a single idea, or a payoff line that needs no diagram.",
		Example:     "Why every engineer should understand backpressure",
		PromptFile:  snippetIllustrationTemplateName,
		NeedsCode:   false,
		Owns:        beatFields{Art: true},
		Normalize:   normalizeIllustrationPlan,
		Validate:    validateIllustrationPlan,
		Scenes:      illustrationScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Figures": strings.Join(ArtFigureNames(), ", "),
			}
		},
	})
}

const snippetIllustrationTemplateName = "snippet_illustration.tmpl"

// Headline bounds. One word is a caption on a slide, not a line worth setting
// at 100px; past eight it wraps to three lines and stops being a headline.
const (
	minHeadlineWords = 2
	maxHeadlineWords = 8
)

// maxCaptionWords keeps the supporting line to a sentence. Past this it
// competes with the narration, and the viewer reads instead of listening.
const maxCaptionWords = 18

// artFigureVocab is the closed set of figures the renderer can draw.
//
// Go owns this list and normalizes anything outside it to "spark"; a drift test
// (artwork_test.go) keeps it identical to FIGURES in artwork.tsx, because a
// figure Go allows and the renderer does not have would render as an empty
// half-frame — invisible until someone watches the clip.
var artFigureVocab = map[string]bool{
	// Software, infrastructure and the machines it runs on.
	"server":    true,
	"database":  true,
	"terminal":  true,
	"code":      true,
	"branch":    true,
	"bug":       true,
	"package":   true,
	"api":       true,
	"pipeline":  true,
	"cache":     true,
	"queue":     true,
	"container": true,
	"firewall":  true,
	"cpu":       true,
	"memory":    true,
	"disk":      true,
	"laptop":    true,
	"phone":     true,
	"monitor":   true,
	"cloud":     true,
	"network":   true,
	"lock":      true,

	// The web, and the things people do on it.
	"browser":   true,
	"cursor":    true,
	"search":    true,
	"form":      true,
	"cart":      true,
	"tag":       true,
	"wallet":    true,
	"receipt":   true,
	"star":      true,
	"heart":     true,
	"bell":      true,
	"envelope":  true,
	"chat":      true,
	"megaphone": true,
	"share":     true,
	"video":     true,

	// Measuring, testing and looking closely.
	"atom":       true,
	"flask":      true,
	"microscope": true,
	"telescope":  true,
	"dna":        true,
	"magnet":     true,
	"battery":    true,
	"prism":      true,
	"balance":    true,
	"compass":    true,
	"ruler":      true,
	"calculator": true,
	"satellite":  true,
	"gauge":      true,
	"stack":      true,
	"clock":      true,
	"hourglass":  true,

	// The world, weather and growing things.
	"globe":    true,
	"tree":     true,
	"leaf":     true,
	"mountain": true,
	"sun":      true,
	"moon":     true,
	"drop":     true,
	"fire":     true,
	"wind":     true,
	"seed":     true,
	"wave":     true,
	"recycle":  true,

	// Desks, days and the things people aim at.
	"book":      true,
	"notebook":  true,
	"pencil":    true,
	"folder":    true,
	"calendar":  true,
	"briefcase": true,
	"coffee":    true,
	"target":    true,
	"trophy":    true,
	"flag":      true,
	"key":       true,
	"door":      true,
	"handshake": true,
	"brain":     true,
	"eye":       true,
	"checklist": true,
	"lightbulb": true,
	"gears":     true,

	// Teaching, and the shapes a lesson is made of.
	"question":    true,
	"chalkboard":  true,
	"insight":     true,
	"timer":       true,
	"certificate": true,
	"answer":      true,
	"library":     true,
	"highlighter": true,
	"signpost":    true,
	"foundation":  true,
	"progress":    true,
	"discussion":  true,
	"study":       true,
	"steps":       true,
	"graduate":    true,
	"bookmark":    true,

	// Shapes for ideas that are not objects.
	"arrow":  true,
	"loop":   true,
	"funnel": true,
	"filter": true,
	"switch": true,
	"slider": true,
	"puzzle": true,
	"ladder": true,
	"bridge": true,
	"maze":   true,
	"orbit":  true,
	"growth": true,
	"rocket": true,
	"shield": true,
	"chart":  true,
	"spark":  true,
}

// ArtFigureNames returns the figure vocabulary, sorted — for the prompt and
// for anything that wants to show a creator what is available.
func ArtFigureNames() []string {
	out := make([]string, 0, len(artFigureVocab))
	for name := range artFigureVocab {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// normalizeArtFigure maps an invented figure name onto the neutral fallback.
func normalizeArtFigure(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if artFigureVocab[n] {
		return n
	}
	return "spark"
}

// illustrationScenes lays the clip out as one shot per beat.
//
// There is no opening title card. The first beat's headline *is* the title
// card — it is already big type on an empty stage — and cutting to a separate
// card first would say the same thing twice before the clip has said anything.
func illustrationScenes(in SnippetSceneInput) ([]Scene, error) {
	scenes := make([]Scene, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Art == nil {
			return nil, fmt.Errorf("beat %q has no art", beat.ID)
		}
		scenes = append(scenes, Scene{
			Type:    SceneIllustration,
			StartMs: startMs,
			EndMs:   endMs,
			Props: map[string]any{
				"headline": beat.Heading,
				"emphasis": beat.Art.Emphasis,
				"caption":  beat.Art.Caption,
				"figure":   normalizeArtFigure(beat.Art.Figure),
				// The figure changes sides every shot. Holding it on one side
				// for a whole clip makes six cuts read as one slide with the
				// words swapped out.
				"flip": i%2 == 1,
			},
		})
	}
	return scenes, nil
}

// normalizeIllustrationPlan settles the parts of a shot the renderer has an
// opinion about anyway.
//
// The emphasis is the interesting one. It is a marker stroke drawn under part
// of the headline, so a phrase that is not in the headline has nothing to draw
// under — and the mismatch is almost always the model quoting the idea rather
// than the words ("speed" under a headline about "faster"). Clearing it costs
// the shot its accent; rejecting it costs the clip a round to be told a fact
// about string matching.
func normalizeIllustrationPlan(p *SnippetPlan) {
	for i := range p.Beats {
		b := &p.Beats[i]
		if len(strings.Fields(b.Heading)) < minHeadlineWords {
			b.Heading = headingFromNarration(b.Narration)
		}
		b.Heading = clampWords(b.Heading, maxHeadlineWords)
		if b.Art == nil {
			continue
		}
		b.Art.Figure = normalizeArtFigure(b.Art.Figure)
		b.Art.Caption = collapseSpaces(b.Art.Caption)
		b.Art.Emphasis = collapseSpaces(b.Art.Emphasis)
		if b.Art.Emphasis != "" && !containsPhrase(b.Heading, b.Art.Emphasis) {
			b.Art.Emphasis = ""
		}
	}
}

func validateIllustrationPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Art: true}); err != nil {
		return err
	}
	seenFigure := map[string]int{}
	for _, b := range p.Beats {
		if b.Art == nil {
			return fmt.Errorf("beat %q has no art — every beat in this template is a shot", b.ID)
		}
		n := len(strings.Fields(b.Heading))
		if n < minHeadlineWords || n > maxHeadlineWords {
			return fmt.Errorf("beat %q has a %d-word heading; headlines are %d-%d words",
				b.ID, n, minHeadlineWords, maxHeadlineWords)
		}
		// The emphasis is a marker stroke drawn under part of the headline, so
		// a phrase that is not in the headline has nothing to underline and
		// the shot silently loses its accent.
		if e := strings.TrimSpace(b.Art.Emphasis); e != "" && !containsPhrase(b.Heading, e) {
			return fmt.Errorf("beat %q emphasises %q, which does not appear in its heading %q",
				b.ID, e, b.Heading)
		}
		if n := len(strings.Fields(b.Art.Caption)); n > maxCaptionWords {
			return fmt.Errorf("beat %q has a %d-word caption; at most %d", b.ID, n, maxCaptionWords)
		}
		seenFigure[normalizeArtFigure(b.Art.Figure)]++
	}
	// A run of shots that all show the same drawing is a static image with
	// changing text, which is the failure this template exists to avoid. One
	// repeat is a callback; two is a rut.
	for figure, n := range seenFigure {
		if n > 2 {
			return fmt.Errorf("the %q figure is used in %d beats; use a different figure for each idea (at most 2 beats may share one)", figure, n)
		}
	}
	return nil
}

// containsPhrase reports whether `phrase` occurs in `text`, comparing on
// letters and digits only.
//
// The comparison has to be loose because the two strings come from different
// places in the same JSON object and models are inconsistent about echoing
// punctuation and case: a heading of "Ship it — twice." and an emphasis of
// "Twice" are the same word, and rejecting that pair buys a correction round
// that teaches the model nothing.
func containsPhrase(text, phrase string) bool {
	return strings.Contains(phraseKey(text), phraseKey(phrase))
}

// phraseKey reduces a string to lower-case alphanumerics and single spaces.
func phraseKey(s string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(s) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			if space && b.Len() > 0 {
				b.WriteByte(' ')
			}
			space = false
			b.WriteRune(r)
		default:
			space = true
		}
	}
	return b.String()
}
