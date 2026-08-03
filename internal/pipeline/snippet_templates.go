package pipeline

// The snippet template catalog.
//
// A template is the answer to "what does this clip look like?" — it owns the
// prompt that plans the clip, the rules that plan must satisfy, and the
// mapping from a planned-and-timed clip onto renderer scenes. Adding a
// template means adding one file here and one Remotion component; nothing in
// the pipeline needs to know the catalog grew.
//
// The split of responsibility is deliberate: the *shared* code owns timing,
// theming, captions, and the scene-graph envelope, so no template can get
// those wrong or drift from the design system. A template only decides what
// fills the frame.

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// SnippetTemplate is one entry in the catalog.
type SnippetTemplate struct {
	// Name is the id used on the CLI and in snippet.yaml.
	Name string
	// Title and Description are the gallery copy shown in the studio.
	Title       string
	Description string
	// Category is which group this template appears under, from the closed
	// vocabulary in snippet_categories.go. Required — registration panics
	// without it, because a catalog that can grow an uncategorised entry
	// eventually does, and it lands somewhere nobody looks.
	Category string
	// Since is the catalog release this template arrived in ("" for the
	// original set, "v1" for the reference-look batch). A fact rather than a
	// status, so it does not go stale when the next batch lands.
	Since string
	// Family is which surface offers this template ("" = the snippets gallery).
	// See snippet_categories.go — it is a routing fact, not a second category:
	// the registry stays single so the CLI, the reel caster and every plan on
	// disk keep resolving a template by name regardless of which page offered it.
	Family string
	// Example is a prompt that shows this template at its best; the studio
	// offers it as a starting point.
	Example string
	// Shelved keeps this template out of everything that *offers* a choice —
	// the studio gallery, `snippet --list`, and the caster's catalog — while
	// leaving it fully usable when named explicitly in a snippet.yaml or a
	// reel.yaml.
	//
	// The distinction matters because the two failure modes are not symmetric.
	// Deleting a template breaks every piece already on disk that names it;
	// leaving a template that does not meet the bar in the gallery means it
	// keeps being picked, and a caster handed a catalog will use everything in
	// it. Shelving is the only state that stops new pieces choosing a look
	// without invalidating the old ones.
	//
	// A shelved template is a judgement about the *output*, so the reason
	// belongs on the field at the registration site, not here.
	Shelved bool
	// PromptFile is the prompt template rendered to plan a clip.
	PromptFile string
	// NeedsCode makes the verify stage part of this template's pipeline, so
	// any code shown is code that really ran.
	NeedsCode bool
	// DefaultTargetSec is the runtime this template aims for when the request
	// does not say (0 = defaultSnippetTargetSec). A template with a beat floor
	// well above the shared one needs its own default, or the standard 45s
	// budget cannot fund the beats its own validator demands.
	DefaultTargetSec int
	// MaxBeats raises this template's beat ceiling above the shared
	// maxSnippetBeats (0 = the shared one).
	//
	// Seven is the right cap for a clip built from a handful of *moments*, which
	// almost every template is. It is the wrong cap for one whose subject is a
	// list — a tool with four facts to state, a build with five phases each
	// worth opening — because there the beat count is a property of the content
	// rather than of how long a viewer's attention lasts on one picture.
	MaxBeats int
	// MinTargetSec is the shortest runtime this template can actually satisfy
	// (0 = the shared floor). A default is a suggestion and callers override it;
	// this is the arithmetic. `story` needs eight beats, and eight beats of the
	// ten-word minimum cannot be written inside a twenty-second word budget — so
	// asking for one is not a plan that comes out badly, it is a plan that
	// cannot come out at all, and the correction loop burns three rounds and a
	// token budget discovering that.
	MinTargetSec int

	// Plan produces the clip's design. Nil uses planSnippetDefault, which
	// renders PromptFile and decodes a SnippetPlan — enough for every
	// template whose plan fits the standard shape.
	Plan func(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) (*SnippetPlan, error)
	// Owns and OwnsPlan declare which of SnippetBeat's and SnippetPlan's
	// optional payloads this template reads. Everything else is migrated onto
	// the field it belongs in, or dropped, before the plan is validated.
	Owns     beatFields
	OwnsPlan planFields
	// NoSalvage refuses the near-miss path when the correction rounds run out.
	//
	// Most templates ship their closest draft rather than nothing, because a
	// slightly loose clip still renders and still teaches. A template whose
	// rules are about truth rather than shape cannot take that trade: its
	// closest draft is a clip that says something the evidence does not support.
	NoSalvage bool
	// PreValidate copies what the caller resolved onto the plan, before it is
	// judged. A validator can only check against what is on the plan when
	// Validate runs, and Validate runs inside the correction loop.
	PreValidate func(spec SnippetSpec, p *SnippetPlan)
	// Normalize repairs this template's own mechanical mistakes — a label a
	// word too long, a vocabulary term the model invented, a link pointing at
	// nothing — before Validate sees the plan. See snippet_normalize.go for
	// where the line between the two sits.
	Normalize func(p *SnippetPlan)
	// Validate enforces this template's own rules on a plan, and rejects
	// beat fields the template does not own.
	Validate func(p *SnippetPlan) error
	// Scenes maps the planned, timed clip onto renderer scenes.
	Scenes func(in SnippetSceneInput) ([]Scene, error)
	// PromptData contributes extra top-level fields to the prompt beyond the
	// shared set — a template's own vocabularies and bounds. Keys collide at
	// the template author's peril; the shared keys are listed in
	// sharedPromptData.
	PromptData func(spec SnippetSpec, cfg config.Config) map[string]any
}

// SnippetTemplates is the catalog, keyed by name. Templates register
// themselves from their own files' init().
var SnippetTemplates = map[string]*SnippetTemplate{}

// registerSnippetTemplate adds a template to the catalog.
func registerSnippetTemplate(t *SnippetTemplate) {
	if _, dup := SnippetTemplates[t.Name]; dup {
		panic("duplicate snippet template " + t.Name)
	}
	checkTemplateCategory(t)
	SnippetTemplates[t.Name] = t
}

// SnippetTemplateNames returns every registered name, sorted — shelved
// templates included.
//
// This is the *enumeration*, not the offer. It backs the "templates: ..." hint
// on a validation error, where naming a shelved template is correct because
// naming one explicitly still works, and promptDataFallbacks, which needs every
// template's vocabularies whether or not the gallery shows it. Use
// SnippetTemplateList for anything a creator or a model chooses from.
func SnippetTemplateNames() []string {
	out := make([]string, 0, len(SnippetTemplates))
	for name := range SnippetTemplates {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// SnippetTemplateList returns the templates on offer, sorted by name, for the
// studio gallery, `coursesmith snippet --list` and the caster's catalog.
//
// Shelved templates are filtered here rather than at each of the three call
// sites, so shelving one cannot land in two of them and be missed in the third
// — which is the whole failure this single choke point exists to prevent.
func SnippetTemplateList() []*SnippetTemplate {
	out := make([]*SnippetTemplate, 0, len(SnippetTemplates))
	for _, name := range SnippetTemplateNames() {
		if t := SnippetTemplates[name]; !t.Shelved {
			out = append(out, t)
		}
	}
	return out
}

// SnippetSceneInput is everything a template needs to lay its clip out on the
// timeline. Timing is already resolved: spans are real, measured word
// timings from the aligner, not estimates.
type SnippetSceneInput struct {
	Spec SnippetSpec
	Plan *SnippetPlan
	Cfg  config.Config
	// Course is the synthetic snippets course (branding, name).
	Course *project.Course
	// Spans are the aligned section spans, one per beat, in plan order.
	Spans []SectionSpan
	// BeatEndMs is the ms at which each beat's visual should give way to the
	// next — the next beat's start, or the padded end of the audio.
	BeatEndMs []int
	// Verification maps a code block's hash to what running it really
	// printed; empty for templates that do not show code.
	VerifiedOutput map[string]string
	// DurationMs is the finished clip's length.
	DurationMs int
}

// Beat returns beat i with its resolved timing.
func (in SnippetSceneInput) Beat(i int) (b SnippetBeat, startMs, endMs int) {
	return in.Plan.Beats[i], in.Spans[i].StartMs, in.BeatEndMs[i]
}

// sharedPromptData is the field set every template's prompt can rely on. A
// template adds its own with SnippetTemplate.PromptData.
//
// It is a map rather than a struct so templates can extend it; prompts still
// render with missingkey=error, so a typo in a template is a build failure
// rather than a silently empty instruction.
func sharedPromptData(spec SnippetSpec, cfg config.Config) map[string]any {
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	target := spec.ResolvedTargetSec()
	wantWords, minWords, maxWords := wordBudget(target, pace)
	minBeats, maxBeats, suggest, perBeat := beatBounds(wantWords, templateBeatCeiling(spec.Template))
	return map[string]any{
		"Prompt":          spec.Prompt,
		"Title":           spec.Title,
		"TargetSec":       target,
		"TargetWords":     wantWords,
		"MinWords":        minWords,
		"MaxWords":        maxWords,
		"MinWordsPerBeat": minWordsPerBeat,
		"MaxWordsPerBeat": maxWordsPerBeat,
		"Tone":            cfg.Style.Tone,
		"Audience":        cfg.Style.Audience,
		"Language":        cfg.Style.Language,
		"CodeLanguage":    spec.ResolvedCodeLanguage(),
		"PaceWPM":         pace,
		"MinBeats":        minBeats,
		"MaxBeats":        maxBeats,
		"SuggestBeats":    suggest,
		// The words one beat affords at *this* runtime. Every prompt calibrates
		// against it, so the number the model is told to write and the number
		// it is scored against are the same number.
		"WordsPerBeat": perBeat,
		// Headline and caption bounds are shared rather than per-template:
		// four prompts reference them and only three templates were supplying
		// them, so `story` rendered {{.MinHeadlineWords}} against a map that
		// had never heard of it. Bounds that several templates use are shared
		// data by definition — a copy per template is three chances to forget
		// the fourth.
		"MinHeadlineWords": minHeadlineWords,
		"MaxHeadlineWords": maxHeadlineWords,
		"MaxCaptionWords":  maxCaptionWords,
		// A reel segment's planning context, empty for a standalone snippet.
		//
		// The facts reach a template's writer through Prompt, which enrichment
		// has already rewritten to carry them (snippet_enrich.go) — that is the
		// path that works without editing twenty-seven prompt files. These keys
		// are here so a template prompt that wants the piece's brief, or the
		// ground already covered, can say {{.Brief}} or {{.Priors}} and get it
		// rather than getting an empty string from the healing path.
		"Brief":    spec.Brief,
		"Material": spec.Material,
		"Priors":   spec.Priors,
		// The established facts and the known gaps. Shared rather than owned by
		// the enrich prompt alone, for the reason the headline bounds above are:
		// more than one prompt wants them, and a key that only one supplier
		// defines renders empty through the healing path in every other — which
		// looks like a drift warning rather than the missing data it is.
		"Facts": substanceLines(spec.Substance),
		"Gaps":  substanceGaps(spec.Substance),
		// Shared as well as owned by `myth`, so any template that wants to open on
		// a belief the viewer recognises can — and so a prompt referencing it does
		// not fall through the healing path and render empty with a drift warning.
		"Misconceptions": substanceMisconceptions(spec.Substance),
	}
}

// beatBounds sizes the beat count against the narration budget, and says how
// many words each beat then affords.
//
// A fixed 3-7 range was right for the runtimes this started with and is a
// *contradiction* below about thirty seconds. Every one of these prompts also
// calibrates a beat at roughly forty words, so a three-beat floor is a
// hundred-and-twenty-word floor — more than a 20-second clip's entire ceiling
// of eighty-nine. The model obeys the concrete per-beat number, blows the
// total, and the correction rounds cannot rescue it because nothing it could
// write satisfies both rules at once. Observed in the wild as a plan walking
// 128 → 114 → 96 → 93 words across three rounds and failing: it was converging
// on a floor the instructions themselves imposed.
//
// So the count comes from the budget. `wordsPerBeat` is the same arithmetic
// run back the other way, and it is what the prompts calibrate against now —
// which means the number the model is told to hit and the number it is scored
// against are the same number at every runtime, instead of only above 45s.
// The `ceiling` is the most beats this template can hold — maxSnippetBeats for
// almost everything, and higher for the few whose structure is a *list* rather
// than a handful of moments. See SnippetTemplate.MaxBeats.
func beatBounds(targetWords, ceiling int) (minBeats, maxBeats, suggest, wordsPerBeat int) {
	if ceiling <= 0 {
		ceiling = maxSnippetBeats
	}
	if targetWords <= 0 {
		// A plan built by hand (a test, a fixture) has no budget to size
		// against; fall back to the range that was fixed before this existed.
		return floorSnippetBeats, ceiling, idealWordsPerBeat, idealWordsPerBeat
	}
	suggest = max((targetWords+idealWordsPerBeat/2)/idealWordsPerBeat, floorSnippetBeats)
	// A beat can only hold so much, so the *words* set a floor on the beat count
	// as surely as the ideal sets a target. Without this the two halves of the
	// instructions contradict each other at long runtimes exactly as they used to
	// at short ones: a 180-second clip was told to write seven beats of 75 words
	// against a 60-word per-beat maximum, which no plan can satisfy. That is the
	// same failure documented above, at the other end of the range, and it went
	// unnoticed because nothing asked for three minutes until now.
	suggest = max(suggest, (targetWords+maxWordsPerBeat-1)/maxWordsPerBeat)
	suggest = min(suggest, ceiling)
	minBeats = min(max(suggest-1, floorSnippetBeats), ceiling)
	maxBeats = min(max(suggest+2, minBeats), ceiling)
	// A list-shaped template must be able to reach the ceiling it declared.
	//
	// suggest+2 is the right width for a clip made of *moments* — it keeps a
	// 45-second piece from becoming eight captions. It silently overrode the
	// ceiling for the templates whose beat count is a property of their content,
	// which is the one case MaxBeats exists to express. Measured: constellation
	// defaults to 55s, so suggest+2 allowed 6 beats while its shape — centre,
	// spokes, whole — needs 7 for five spokes; rundown allowed 6 while five cards
	// plus an opener and a summary needs 7. Every run of those templates therefore
	// failed its own validator, burned three correction rounds, salvaged the
	// closest draft, and shipped it. "1 of the 5 cards are never covered" was the
	// arithmetic saying so.
	//
	// So the ceiling is reachable when the budget can fund it at a substantial
	// beat — half the ideal, not the bare ten-word minimum, because funding eight
	// beats at ten words each is how a clip becomes the slideshow the width was
	// protecting against. max() rather than assignment: this only ever widens the
	// range, so no runtime that worked before gets a narrower one.
	if fundable := targetWords / leanWordsPerBeat; fundable > maxBeats {
		maxBeats = min(fundable, ceiling)
	}
	// And when even the ceiling cannot fund the budget, advise the most a beat
	// may actually carry rather than the arithmetic answer. The clip will run
	// short of its target; being told to write something that is rejected on
	// arrival would leave it with no clip at all.
	return minBeats, maxBeats, suggest, min(targetWords/suggest, maxWordsPerBeat)
}

// templateBeatCeiling is the most beats the named template may use.
func templateBeatCeiling(name string) int {
	if tpl, ok := SnippetTemplates[name]; ok && tpl.MaxBeats > 0 {
		return tpl.MaxBeats
	}
	return maxSnippetBeats
}

// Beat-count bounds.
//
// The floor is two rather than three: two beats is one cut, which is the least
// that still makes this a film rather than a held shot, and it is what lets a
// ten- or twenty-second clip exist at all. Seven is where a clip made of
// *moments* stops landing — a template whose subject is a list of things is a
// different case, and raises its own ceiling with MaxBeats.
const (
	floorSnippetBeats = 2
	maxSnippetBeats   = 7
	// idealWordsPerBeat is how much narration one visual comfortably holds —
	// the divisor that turns a word budget into a beat count.
	idealWordsPerBeat = 40
	// leanWordsPerBeat is the least a beat can carry and still be a thought
	// rather than a label. It is what decides how far a list-shaped template may
	// stretch toward its declared ceiling: dividing the budget by the ten-word
	// hard minimum would licence eight beats of ten words, which is a slideshow.
	leanWordsPerBeat = idealWordsPerBeat / 2
)

// Per-beat narration bounds. Under ten words a beat is a caption, not a
// thought; over sixty it outlasts any single visual.
// beatRole is the job a beat does, derived from where it sits.
//
// Position rather than a field the model declares, and that is a deliberate
// choice about where to spend a rule. A `role` on SnippetBeat would need adding
// to twenty-seven prompts, would be got wrong by some of them, and would tell us
// nothing position does not: every template in this catalog opens by naming its
// subject, develops through the middle, and closes on the whole. `constellation`
// is centre → spokes → whole, `myth` is claim → evidence → why, `rundown` is
// promise → items → all. The shape is the catalog's, not the model's to invent.
type beatRole int

const (
	// roleOpen names the subject. Short by convention — "No-code is visual
	// programming", "Four mindsets to adopt" — because its job is to put one
	// thing on screen, not to explain it.
	roleOpen beatRole = iota
	// roleDevelop carries the actual teaching, and is where the words belong.
	roleDevelop
	// roleLand closes on the whole. Short again: a summary that runs as long as
	// the thing it summarises has not summarised anything.
	roleLand
)

// roleOf returns the job of beat i of n.
//
// A two-beat clip is open then land, with no middle: at ten to twenty seconds
// there is one cut and no room to develop anything, which is what makes those
// runtimes a hook rather than an explanation.
func roleOf(i, n int) beatRole {
	switch {
	case i == 0:
		return roleOpen
	case i == n-1 && n > 1:
		return roleLand
	default:
		return roleDevelop
	}
}

// minWordsFor is the floor for a beat in this role.
//
// The flat ten-word floor was the last thing failing every run of the
// list-shaped templates, and it was failing them on the beats the templates are
// *designed* to keep short. Measured on the no-code reel after the beat-ceiling
// fix, every remaining violation was an opener or a closer one or two words shy:
// constellation's "No-code is visual programming" at eight, rundown's "Four
// mindsets to adopt" at nine. The rule and the format were arguing, the format
// was right, and the correction rounds spent settling it are most of why 71% of
// this pipeline's token spend went on re-planning.
func minWordsFor(role beatRole) int {
	if role == roleDevelop {
		return minWordsPerBeat
	}
	return minWordsOpenLand
}

const (
	minWordsPerBeat = 10
	// minWordsOpenLand is the floor for an opener or a closer. Six words is still
	// a sentence — "No-code means building without code" — and below it a beat
	// really is a label with a voice track.
	minWordsOpenLand = 6
	maxWordsPerBeat  = 60
)

// snippetPlanRepairRounds is how many correction attempts a plan gets. Replies
// are cached, so the cost lands once per distinct prompt.
const snippetPlanRepairRounds = 3

// wordBudget is how much narration a clip of the requested length needs.
//
// Models systematically under-write to a seconds target — they have no clock —
// so the budget is enforced, not suggested: a plan outside the band is
// rejected and regenerated with the miss quoted back.
//
// The band is asymmetric because the two failures are not symmetric. Coming in
// short breaks the clip: the visuals are timed to the voice, so half the
// narration is half a video. Running long only makes it longer than asked —
// every beat still lands, the viewer just gets more of them.
//
// The ceiling is deliberately loose. A tighter one (135%) was tried and did not
// hold: on a topic the model judged to need ~185 words it produced 184-185 in
// three consecutive correction rounds, ignoring the stated target entirely.
// Models write to what the content seems to need, not to a word count, and no
// amount of restating the budget changed that. Rejecting those plans bought a
// failed generation rather than a shorter clip.
//
// So the target is treated as a target: it steers the draft, the floor is
// enforced because undershooting is fatal, and the runtime a creator picks is
// documented as approximate. The finished duration is always reported.
func wordBudget(targetSec, paceWPM int) (target, minWords, maxWords int) {
	target = targetSec * paceWPM / 60
	return target, target * 75 / 100, target * 155 / 100
}

// narrationWords counts the words the voice will actually speak.
func narrationWords(p *SnippetPlan) int {
	n := 0
	for _, b := range p.Beats {
		n += len(strings.Fields(b.Narration))
	}
	return n
}

// planSnippetDefault renders the template's prompt and decodes the reply.
func planSnippetDefault(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) (*SnippetPlan, error) {
	tpl := SnippetTemplates[spec.Template]
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	target := spec.ResolvedTargetSec()
	wantWords, minWords, maxWords := wordBudget(target, pace)
	// Hoisted above the render because the appended arithmetic below quotes these
	// bounds, and they must be the same ones the validator scores against — two
	// calls to beatBounds either side of the prompt is how the number the model is
	// told and the number it is judged by drift apart.
	minBeats, maxBeats, suggest, perBeat := beatBounds(wantWords, templateBeatCeiling(spec.Template))
	data := sharedPromptData(spec, cfg)
	if tpl.PromptData != nil {
		for k, v := range tpl.PromptData(spec, cfg) {
			data[k] = v
		}
	}
	system, user, err := e.renderPrompt(tpl.PromptFile, data)
	if err != nil {
		return nil, err
	}
	// The budget arithmetic, spelled out. Appended centrally for the same reason
	// the critique below is: it is shared guidance, not a property of any one
	// template's look, and twenty-seven copies means the twenty-eighth is wrong.
	//
	// Every prompt already quotes a per-beat word count. What none of them could
	// say is what happens when the model takes the latitude in the beat range:
	// told "3-7 beats, about 36 words each", it writes seven beats of nine words
	// and fails the ten-word floor on all of them. Both numbers were followed —
	// the count from one rule and the length from nowhere — and the arithmetic
	// connecting them was never stated. Observed on every list-shaped template:
	// constellation and rundown failed this and nothing else after the beat
	// ceiling was fixed.
	user += beatVariationAdvice(wantWords, minBeats, maxBeats) + budgetTotalsAdvice(minWords, maxWords)

	// A review critique is appended to the rendered user message rather than
	// rendered into the prompt file. Blunt, and chosen deliberately: the
	// alternative is the same block copied into twenty-seven templates, where the
	// twenty-eighth to be added silently ignores its reviewer. The prompts stay
	// the description of a template's own rules, and regeneration stays one thing
	// in one place.
	if c := strings.TrimSpace(spec.Critique); c != "" {
		user += "\n\nA reviewer scored your previous plan below the quality bar. Produce a new plan that fixes every point of this critique. Keep what was already good; change what it names.\n\n" + c
	}
	var plan SnippetPlan
	// The closest attempt seen so far, kept for the salvage below: a plan that
	// decoded and normalized is a clip, even when it never satisfied every rule.
	var closest *SnippetPlan
	// A plan has more independent numeric rules than anything else the pipeline
	// asks for — beat count, per-beat words, total words, and whatever the
	// template adds on top. One correction round is not enough to land them all.
	// The bounds themselves are computed above, before the prompt quotes them.
	err = e.completeJSONLenientRounds(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.5, 6144, snippetPlanRepairRounds, &plan, func() error {
		plan.Template = spec.Template // so Validate dispatches to this template
		// The budget the prompt quoted, so the shared validators score the plan
		// against the same beat range the model was asked for.
		plan.targetWords = wantWords
		// Repair what is mechanically repairable before judging the reply, so
		// the correction rounds are spent on what only the model can fix.
		normalizeSnippetPlan(&plan)
		// Facts the caller resolved, injected before the plan is judged.
		//
		// Validation happens inside this loop, so anything a validator checks
		// against must be on the plan *here* — setting it after the planner
		// returns is setting it after every judgement has already been made.
		// That is not hypothetical: the footage template's whole rule is
		// "a beat may only name a moment the recording has", and with the marks
		// attached afterwards the check silently passed on an empty set.
		if tpl.PreValidate != nil {
			tpl.PreValidate(spec, &plan)
		}
		snapshot := plan
		closest = &snapshot
		if err := plan.Validate(); err != nil {
			return err
		}
		// A multi-file plan is executed here, in the correction loop, rather
		// than downstream in verify — verify runs each fenced block on its
		// own, where `import greet` is a ModuleNotFoundError. Running it here
		// means a program that does not work comes back to the model as its
		// own traceback and gets another attempt, which is the whole reason
		// this template can be pointed at something complicated.
		if err := e.runPlannedProject(ctx, &plan); err != nil {
			return err
		}
		// The advice has to point the way the plan actually needs to move. One
		// message served both directions and it said "rewrite with fuller
		// sentences" — which is right when the plan is short and is pushing the
		// wrong way when it is long. A 20-second clip that came back at 128
		// words against an 89-word ceiling was being told, three rounds
		// running, to write more.
		if n := narrationWords(&plan); n < minWords {
			return fmt.Errorf(
				"narration totals %d words but a %ds clip needs %d-%d (aim for %d) — rewrite with fuller sentences, about %d words per beat; do not add beats past %d",
				n, target, minWords, maxWords, wantWords, perBeat, maxBeats)
		} else if n > maxWords {
			return fmt.Errorf(
				"narration totals %d words but a %ds clip needs %d-%d (aim for %d) — cut it back to about %d words per beat across %d beat(s); say less, do not drop below %d beats",
				n, target, minWords, maxWords, wantWords, perBeat, suggest, minBeats)
		}
		return nil
	})
	if err != nil {
		// Salvage is the right trade for a template whose rules are about
		// shape: a clip that is a little loose still renders and still teaches.
		// It is the wrong trade for one whose rules are about *truth*. The
		// closest draft of a footage piece is a clip whose narration does not
		// match the recording — shipping it, with a warning, is precisely the
		// failure this surface exists to refuse, and a warning on stdout is not
		// a defence against a video that says the wrong tool did the work.
		if tpl.NoSalvage {
			return nil, fmt.Errorf("planning %s: %w\nThis template does not ship a near miss: its rules are about whether the clip tells the truth, not about how it is shaped", spec.Template, err)
		}
		if salvaged := salvageSnippetPlan(ctx, e, spec, cfg, closest); salvaged != nil {
			fmt.Fprintf(e.out(), "    ! the plan never satisfied every rule (%v)\n", err)
			fmt.Fprintf(e.out(), "      shipping the closest one — it renders, so the clip is real; expect it to be looser than asked\n")
			// Record it on the plan, not only on stdout. A warning nobody scrolled
			// back to is how three segments of a finished reel shipped under their
			// word floor with no way to tell them from the ones that passed.
			salvaged.Compromises = compromiseLines(err)
			return salvaged, nil
		}
		return nil, fmt.Errorf("planning %s snippet: %w", spec.Template, err)
	}
	return &plan, nil
}

// beatVariationAdvice is the arithmetic and the pacing, stated to every template.
//
// Appended by the shared planner rather than written into twenty-seven prompt
// files, for the reason the critique is: this is shared guidance, not a property
// of any one template's look, and twenty-seven copies means the twenty-eighth is
// wrong.
//
// Two things it has to say, and neither was being said anywhere.
//
// The ARITHMETIC, because every prompt quoted a per-beat word count and none of
// them could say what happens when the model uses the latitude in the beat range:
// told "3-7 beats, about 36 words each" it wrote seven beats of nine words and
// failed the floor on all of them. Both numbers were obeyed and the sum was never
// mentioned.
//
// The VARIATION, because uniform beats are the commonest thing wrong with these
// plans and nothing had ever asked for anything else. Every beat the same length,
// the same shape and the same energy is a list read aloud — the review gate scores
// it 6/10 on non-redundancy run after run, and it is why the narration reads
// metronomic even when every individual sentence is fine.
func beatVariationAdvice(wantWords, minBeats, maxBeats int) string {
	return fmt.Sprintf(
		"\n\nARITHMETIC, before you answer. You are writing %d-%d beats. That number and the word count are linked: at %d beats each is about %d words, at %d beats each is about %d words. Work out which you are writing and size the beats to match.\n\n"+
			"THE BEATS MUST NOT ALL BE THE SAME LENGTH. This is the difference between a clip and a list read aloud, and it is the commonest thing wrong with these plans: every beat the same size, the same shape and the same energy, so nothing lands anywhere.\n"+
			"- The FIRST beat names the subject and can be short — %d words is enough, and one clean sentence beats three hedged ones.\n"+
			"- The MIDDLE beats do the teaching and should be the longest, comfortably past %d words. This is where the explanation lives; a short middle beat has stated something instead of explaining it.\n"+
			"- The LAST beat closes and can be short again, %d words. A summary that runs as long as the thing it summarises has not summarised anything.\n"+
			"No middle beat may be under %d words. If you cannot write one up to %d, you have a beat too many: merge it into its neighbour rather than padding it out.",
		minBeats, maxBeats,
		minBeats, wantWords/max(minBeats, 1), maxBeats, wantWords/max(maxBeats, 1),
		minWordsOpenLand, minWordsPerBeat+10, minWordsOpenLand,
		minWordsPerBeat, minWordsPerBeat,
	)
}

// budgetTotalsAdvice states the one number the validators actually score the
// whole plan against.
func budgetTotalsAdvice(minWords, maxWords int) string {
	return fmt.Sprintf("\nThe narration across every beat must total %d-%d words.", minWords, maxWords)
}

// compromiseLines splits a correction-loop error into one line per rule broken.
//
// The loop accumulates every round's complaint into one error joined by "; ", and
// the same rule usually appears in several rounds — the observed shape is "these
// beats are under the 10-word minimum: X; these beats are under the 10-word
// minimum: Y; narration totals 61 words but...". Stored as one blob it is
// unreadable and unqueryable; split and de-duplicated it answers the question
// somebody actually has, which is "what is wrong with this clip".
func compromiseLines(err error) []string {
	if err == nil {
		return nil
	}
	var out []string
	seen := map[string]bool{}
	for _, part := range strings.Split(err.Error(), ";") {
		part = collapseSpaces(part)
		// The loop's own framing ("content response invalid after 3 correction
		// round(s):") is about the machinery rather than the clip.
		if i := strings.Index(part, "correction round(s):"); i >= 0 {
			part = collapseSpaces(part[i+len("correction round(s):"):])
		}
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	return out
}

// salvageSnippetPlan is the last thing between a creator and an empty hand.
//
// The correction rounds enforce rules of two kinds. Some are structural — a
// quiz with no answer, a board with nothing on it — and a plan that breaks one
// cannot be rendered at all. The rest are editorial: the clip runs short of its
// word budget, the board has three boxes where the template asks for four, one
// figure carries three shots instead of two. Failing the whole generation on
// the second kind is a bad trade. The creator asked for a clip about something;
// a slightly loose clip is an answer to that and an error message is not, and
// they can always re-run for a tighter one.
//
// So the test applied here is the only one that cannot be argued with: lay the
// plan out on a fabricated timeline and see whether the template can actually
// produce scenes from it. If it can, the clip exists and it ships with a
// warning. If it cannot, the original failure stands and is reported honestly.
func salvageSnippetPlan(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config, p *SnippetPlan) *SnippetPlan {
	if p == nil || len(p.Beats) == 0 || strings.TrimSpace(p.Title) == "" {
		return nil
	}
	normalizeSnippetPlan(p)
	// A workspace clip shows what its program really printed. If the program
	// still does not run its output stays empty, which the renderer draws as a
	// terminal that was never opened — wrong-looking, but not a lie.
	if p.Project != nil {
		_ = e.runPlannedProject(ctx, p)
	}
	if err := dryRunSnippetScenes(spec, cfg, p); err != nil {
		return nil
	}
	return p
}

// dryRunSnippetScenes asks a template to lay the plan out against estimated
// timings, and reports whether it could. Every scene builder in the catalog is
// a pure function of the plan and its spans, so this answers "will the render
// stage accept this?" without a voice track, an alignment, or a frame.
func dryRunSnippetScenes(spec SnippetSpec, cfg config.Config, p *SnippetPlan) error {
	tpl, ok := SnippetTemplates[p.Template]
	if !ok || tpl.Scenes == nil {
		return fmt.Errorf("unknown template %q", p.Template)
	}
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	spans := make([]SectionSpan, len(p.Beats))
	ends := make([]int, len(p.Beats))
	at := 0
	for i, b := range p.Beats {
		dur := max(1500, len(strings.Fields(b.Narration))*60_000/pace)
		spans[i] = SectionSpan{ID: b.ID, StartMs: at, EndMs: at + dur}
		ends[i] = at + dur
		at += dur
	}
	_, err := tpl.Scenes(SnippetSceneInput{
		Spec:       spec,
		Plan:       p,
		Cfg:        cfg,
		Course:     &project.Course{Name: "Snippets"},
		Spans:      spans,
		BeatEndMs:  ends,
		DurationMs: at,
	})
	return err
}

// checkBeatShape is the shared structural rule every template applies: how
// many beats, and how much narration each one carries.
// checkBeatShape reports *every* offending beat, and says what to do about it.
//
// Reporting one beat per round was a whack-a-mole: a plan with four over-long
// beats spent its three correction rounds fixing three of them and failed on
// the fourth, having been told about each in turn. This is the same lesson the
// snippet planner already learned for its numeric rules — a single-error retry
// makes models oscillate — applied in the validator that was written before it.
//
// The advice matters as much as the count. An over-long beat is a model with
// more to say than one beat holds, and told only "want 10-60" it *trims*,
// fighting the content it just decided was necessary. Told to split, it keeps
// the writing and satisfies the rule, which is usually the honest fix: a beat
// is one visual, so sixty words is not a style guide, it is how long any single
// image stays interesting.
func checkBeatShape(p *SnippetPlan) error {
	minBeats, maxBeats, _, _ := beatBounds(p.targetWords, templateBeatCeiling(p.Template))
	if n := len(p.Beats); n < minBeats || n > maxBeats {
		return fmt.Errorf("plan has %d beats, want %d-%d", n, minBeats, maxBeats)
	}

	var long, short []string
	for i, b := range p.Beats {
		floor := minWordsFor(roleOf(i, len(p.Beats)))
		switch n := len(strings.Fields(b.Narration)); {
		case n > maxWordsPerBeat:
			long = append(long, fmt.Sprintf("%q (%d words)", b.ID, n))
		case n < floor:
			short = append(short, fmt.Sprintf("%q (%d words, floor %d)", b.ID, n, floor))
		}
	}
	if len(long) == 0 && len(short) == 0 {
		return nil
	}

	var msg []string
	if len(long) > 0 {
		m := fmt.Sprintf("these beats are over the %d-word maximum: %s",
			maxWordsPerBeat, strings.Join(long, ", "))
		// Against this template's own ceiling, not the shared one: telling a
		// twelve-beat template it may not split because it has seven beats is
		// advice that contradicts the range it was just given.
		if room := maxBeats - len(p.Beats); room > 0 {
			m += fmt.Sprintf("; you have %d beats and may use up to %d, so SPLIT each long beat into two rather than cutting it — keep the narration, give it another beat to live in",
				len(p.Beats), maxBeats)
		} else {
			m += fmt.Sprintf("; you are already at the %d-beat maximum, so these have to be tightened", maxBeats)
		}
		msg = append(msg, m)
	}
	if len(short) > 0 {
		msg = append(msg, fmt.Sprintf("these beats are under their floor: %s; the first and last beat may be as short as %d words, every beat between them needs %d. Expand them or fold them into a neighbour",
			strings.Join(short, ", "), minWordsOpenLand, minWordsPerBeat))
	}
	return fmt.Errorf("%s", strings.Join(msg, ". "))
}

// beatFields names the optional SnippetBeat fields a template consumes.
//
// SnippetBeat is the union of what every template needs, so a field one
// template owns is meaningless to the others. Declaring ownership once per
// template — rather than each template hand-checking the others' fields, which
// is quadratic and rots as the catalog grows — means a model that puts a
// whiteboard sketch on a flow diagram gets a loud error instead of silence.
type beatFields struct {
	Footage       bool
	Code          bool
	Run           bool
	Sketch        bool
	Nodes         bool
	Focus         bool
	Art           bool
	Cast          bool
	Shot          bool
	Data          bool
	Work          bool
	Quiz          bool
	Compare       bool
	Anatomy       bool
	Timeline      bool
	Canvas        bool
	Loop          bool
	Mockup        bool
	Stack         bool
	Spec          bool
	Showcase      bool
	Breakdown     bool
	Metric        bool
	Gauge         bool
	Verdict       bool
	Decision      bool
	Myth          bool
	Rundown       bool
	Analogy       bool
	Trace         bool
	Costing       bool
	Constellation bool
	Chapter       bool
	Cycle         bool
	Scale         bool
	Occupancy     bool
	Ranking       bool
	Journal       bool
	Multiplex     bool
	Fork          bool
	Capabilities  bool
	Budget        bool
	Latency       bool
	Multiply      bool
}

// rejectForeignBeatFields fails when a beat sets a field its template does not
// own. Adding a field to SnippetBeat means adding one case here.
func rejectForeignBeatFields(p *SnippetPlan, owned beatFields) error {
	for _, b := range p.Beats {
		var set string
		switch {
		case !owned.Code && b.Code != "":
			set = "code"
		case !owned.Run && b.Run:
			set = "run"
		case !owned.Quiz && b.Quiz != nil:
			set = "quiz"
		case !owned.Compare && b.Compare != nil:
			set = "compare"
		case !owned.Anatomy && b.Anatomy != nil:
			set = "anatomy"
		case !owned.Timeline && b.Timeline != nil:
			set = "timeline"
		case !owned.Canvas && b.Canvas != nil:
			set = "canvas"
		case !owned.Loop && b.Loop != nil:
			set = "loop"
		case !owned.Mockup && b.Mockup != nil:
			set = "mockup"
		case !owned.Stack && b.Stack != nil:
			set = "stack"
		case !owned.Spec && b.Spec != nil:
			set = "spec"
		case !owned.Showcase && b.Showcase != nil:
			set = "showcase"
		case !owned.Breakdown && b.Breakdown != nil:
			set = "breakdown"
		case !owned.Metric && b.Metric != nil:
			set = "metric"
		case !owned.Gauge && b.Gauge != nil:
			set = "gauge"
		case !owned.Verdict && b.Verdict != nil:
			set = "verdict"
		case !owned.Decision && b.Decision != nil:
			set = "decision"
		case !owned.Myth && b.Myth != nil:
			set = "myth"
		case !owned.Rundown && b.Rundown != nil:
			set = "rundown"
		case !owned.Analogy && b.Analogy != nil:
			set = "analogy"
		case !owned.Trace && b.Trace != nil:
			set = "trace"
		case !owned.Costing && b.Costing != nil:
			set = "costing"
		case !owned.Constellation && b.Constellation != nil:
			set = "constellation"
		case !owned.Chapter && b.Chapter != nil:
			set = "chapter"
		case !owned.Cycle && b.Cycle != nil:
			set = "cycle"
		case !owned.Scale && b.Scale != nil:
			set = "scale"
		case !owned.Occupancy && b.Occupancy != nil:
			set = "occupancy"
		case !owned.Ranking && b.Ranking != nil:
			set = "ranking"
		case !owned.Journal && b.Journal != nil:
			set = "journal"
		case !owned.Multiplex && b.Multiplex != nil:
			set = "multiplex"
		case !owned.Fork && b.Fork != nil:
			set = "fork"
		case !owned.Capabilities && b.Capabilities != nil:
			set = "capabilities"
		case !owned.Budget && b.Budget != nil:
			set = "budget"
		case !owned.Latency && b.Latency != nil:
			set = "latency"
		case !owned.Multiply && b.Multiply != nil:
			set = "multiply"
		case !owned.Sketch && len(b.Sketch) > 0:
			set = "sketch"
		case !owned.Nodes && len(b.Nodes) > 0:
			set = "nodes"
		case !owned.Focus && len(b.Focus) > 0:
			set = "focus"
		case !owned.Art && b.Art != nil:
			set = "art"
		case !owned.Cast && b.Cast != nil:
			set = "cast"
		case !owned.Shot && b.Shot != nil:
			set = "shot"
		case !owned.Data && b.Data != nil:
			set = "data"
		case !owned.Work && b.Work != nil:
			set = "work"
		default:
			continue
		}
		return fmt.Errorf("beat %q sets %s, which the %s template does not use", b.ID, set, p.Template)
	}
	return nil
}

// buildSnippetSceneGraph assembles the renderer input for a snippet: the
// shared envelope (theme, motion, captions, duration) plus whatever scenes the
// template lays out.
func buildSnippetSceneGraph(
	course *project.Course,
	l *project.Lesson,
	cfg config.Config,
	spec SnippetSpec,
	plan *SnippetPlan,
	alignment *Alignment,
	verification *VerificationReport,
	audioDurMs int,
) (*SceneGraph, error) {
	tpl, ok := SnippetTemplates[spec.Template]
	if !ok {
		return nil, fmt.Errorf("unknown template %q", spec.Template)
	}
	arch, err := ResolveArchetype(cfg.Style)
	if err != nil {
		return nil, err
	}

	graph := &SceneGraph{
		Theme:     videoThemeForConfig(cfg, course.Name),
		Motion:    arch.Motion,
		AudioFile: VoiceoverFileName,
	}
	if cfg.Style.Captions == "on" {
		graph.Captions = alignment.CaptionWords()
	}

	spans := alignment.CaptionSections()
	if len(spans) != len(plan.Beats) {
		return nil, fmt.Errorf("alignment has %d sections but the plan has %d beats — re-run the align stage", len(spans), len(plan.Beats))
	}

	// A beat's visual holds until the next beat starts; the last one runs to
	// the padded end of the audio so the clip does not cut on the final word.
	ends := make([]int, len(spans))
	for i := range spans {
		if i+1 < len(spans) {
			ends[i] = spans[i+1].StartMs
		} else {
			ends[i] = spans[i].EndMs + videoTailMs
		}
	}

	verified := map[string]string{}
	if verification != nil {
		for _, b := range verification.Blocks {
			verified[project.HashBytes([]byte(b.Code))] = b.Stdout
		}
	}

	scenes, err := tpl.Scenes(SnippetSceneInput{
		Spec:           spec,
		Plan:           plan,
		Cfg:            cfg,
		Course:         course,
		Spans:          spans,
		BeatEndMs:      ends,
		VerifiedOutput: verified,
		DurationMs:     max(audioDurMs, ends[len(ends)-1]),
	})
	if err != nil {
		return nil, fmt.Errorf("template %s: %w", tpl.Name, err)
	}
	if len(scenes) == 0 {
		return nil, fmt.Errorf("template %s produced no scenes", tpl.Name)
	}

	graph.Scenes = scenes
	graph.DurationMs = max(audioDurMs, scenes[len(scenes)-1].EndMs)
	graph.Scenes[len(graph.Scenes)-1].EndMs = graph.DurationMs
	for i := range graph.Scenes {
		if graph.Scenes[i].Props == nil {
			graph.Scenes[i].Props = map[string]any{}
		}
		if _, set := graph.Scenes[i].Props["template"]; !set {
			if v := arch.TemplateFor(graph.Scenes[i].Type); v != "" {
				graph.Scenes[i].Props["template"] = v
			}
		}
	}
	return graph, nil
}

// snippetFileName is the filename shown in a code-bearing snippet's editor
// chrome: derived from the title so it reads like a real file in a real
// project rather than "main.py" every time.
func snippetFileName(title, language string) string {
	ext := map[string]string{
		"python": "py", "javascript": "js", "typescript": "ts", "go": "go",
		"rust": "rs", "java": "java", "ruby": "rb", "sql": "sql", "bash": "sh",
	}[strings.ToLower(language)]
	if ext == "" {
		ext = "txt"
	}
	// Short enough to sit in the editor's file tree without ellipsis: the
	// first couple of words of the title, not the whole thing.
	words := strings.Split(slugify(title), "-")
	if len(words) > 2 {
		words = words[:2]
	}
	base := strings.Join(words, "_")
	if len(base) > 16 {
		base = strings.Trim(base[:16], "_")
	}
	if base == "" {
		base = "main"
	}
	return base + "." + ext
}
