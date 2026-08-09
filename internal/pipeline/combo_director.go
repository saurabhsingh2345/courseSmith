package pipeline

// The director: one prompt in, a whole piece out.
//
// This is the front door the combo surface is built around. A creator says what
// the video is about, how long it should run, what theme it is cut in and
// whether it carries captions — and everything between that and a finished file
// is decided here.
//
// It is four decisions in a fixed order, and the order is the design:
//
//	1. SUBSTANCE  what is actually known about this subject, and what is not.
//	2. OUTLINE    what the piece argues, divided into parts. No looks in sight.
//	3. CAST       which template holds each part. No latitude to change the parts.
//	4. WRITE      each part planned through its template's own writer (the plan
//	              stage), then read back as a whole piece and repaired (the
//	              critic).
//
// Each stage can only see what the one before it settled. That is what stops the
// failure this surface was rebuilt to fix: a stage that can revisit an earlier
// decision will, and it will revisit it in favour of whatever is easiest to
// satisfy. Casting given the power to change the outline does not choose better
// looks — it chooses easier parts.
//
// == Why the runtime is decided first ==
//
// How many parts a piece has is a function of how long it runs, not a
// preference. Five parts cannot carry ten minutes without each running two,
// which is past what most templates hold; twelve parts cannot fit in ninety
// seconds without each being a caption. So the budget is computed before the
// outliner is asked for anything, and it is handed the part count rather than
// being asked to pick one.

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
)

// ComboRequest is everything a creator decides.
//
// Deliberately small. The whole argument for this surface is that a person
// should have to make four choices and a machine should make the other forty —
// so anything that could be inferred is inferred, and what is left is what
// genuinely differs between two people asking for the same video.
type ComboRequest struct {
	// Subject is what the piece is about, in the creator's words. The only
	// required field.
	Subject string
	// Title is the finished piece's title. Empty lets the director write one.
	Title string
	// Minutes is how long the piece should run. Zero reads a length out of the
	// subject if one is stated there, and otherwise takes the default.
	Minutes int
	// Skin is the house style, which also decides which templates may be cast
	// (see combo_pool.go). Empty is SkinDefault.
	Skin string
	// Captions burns the caption track into the video: "on" or "off".
	Captions string
	// Mode is light or dark. Empty is the course default.
	Mode string
	// Voice and Model override the combos course's defaults.
	Voice string
	Model string
}

// DirectResult is what the director produced, and the reasoning behind it.
//
// The outline is returned alongside the spec rather than only being written into
// it, because the outline is the decision worth disagreeing with and it is
// cheapest to disagree with here — before a planning call has been spent on any
// part. The CLI prints it; the studio shows it.
type DirectResult struct {
	Spec    *ComboSpec
	Outline *ComboOutline
	Budget  RuntimeBudget
	// Picks is the look chosen for each part, with the caster's reasoning.
	Picks []LookPick
}

// DirectCombo turns a request into a combo ready to run.
//
// Writes nothing to disk. Creating the piece is the caller's step, so a director
// run that produces a structure somebody dislikes costs two LLM calls rather
// than a directory they then have to delete.
func DirectCombo(ctx context.Context, e *Env, req ComboRequest, cfg config.Config) (*DirectResult, error) {
	subject := strings.TrimSpace(req.Subject)
	if subject == "" {
		return nil, fmt.Errorf("directing a combo needs a subject — say what the video is about")
	}
	if e.Router == nil {
		return nil, fmt.Errorf("directing a combo needs an LLM — set GROQ_API_KEY (or an OpenAI-compatible provider) and retry")
	}
	skin := normalizeSkin(req.Skin)

	budget := comboBudget(req, subject)
	fmt.Fprintf(e.out(), "  → direct    %s\n", ComboPoolDescribe(skin))
	if d := budget.Describe(); d != "" {
		fmt.Fprintf(e.out(), "              %s\n", d)
	}

	// The facts, before anything is argued from them.
	//
	// Best-effort, and that is deliberate: a failure here should produce a less
	// well-informed piece, never no piece. It is also not a doubled cost —
	// establishSubstance is a pure function of (subject, cfg) and the substance
	// stage calls it again with identical inputs, so the second call is a cache
	// hit and the facts are paid for once.
	var sub *Substance
	if s, err := e.establishSubstance(ctx, subject, cfg); err == nil {
		sub = s
		fmt.Fprintf(e.out(), "              %d facts established, %d gaps\n", len(s.Renderable()), len(s.Gaps))
	} else {
		fmt.Fprintf(e.out(), "              ! directing without a fact sheet: %s\n", errSummary(err))
	}

	outline, err := OutlineCombo(ctx, e, subject, req.Title, budget, sub, cfg)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(e.out(), "  → outline   %d parts\n", len(outline.Parts))

	picks, err := CastLooks(ctx, e, outline, skin, cfg)
	if err != nil {
		return nil, err
	}

	spec := &ComboSpec{
		Title: outline.Title,
		// The subject is kept verbatim as the brief. Every segment's writer is
		// handed it, so it is what keeps twelve independently-planned parts
		// sounding like one piece rather than twelve answers to twelve questions.
		Brief:     subject,
		Angle:     outline.Angle,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	spec.Config.Style.Skin = skin
	if req.Captions != "" {
		spec.Config.Style.Captions = req.Captions
	}
	if req.Mode != "" {
		spec.Config.Style.Mode = req.Mode
	}
	if req.Voice != "" {
		spec.Config.Style.Voice = req.Voice
	}
	if req.Model != "" {
		spec.Config.Pipeline.LLMContent = req.Model
	}

	for i, part := range outline.Parts {
		seg := ComboSegment{
			Template: picks[i].Template,
			// What this part covers is the outline's claim, not a restatement of
			// the subject. The writer is asked for the increment rather than for
			// "something about X", which is what stops two segments arriving at
			// the same content from different directions.
			Prompt:   part.Establishes,
			Heading:  part.Heading,
			Role:     part.Role,
			Material: part.Material,
			Why:      picks[i].Why,
		}
		if budget.PerSegmentSec > 0 {
			seg.TargetSec = segmentTargetFor(picks[i].Template, budget.PerSegmentSec)
		}
		spec.Segments = append(spec.Segments, seg)
	}
	spec.EnsureSegmentIDs()
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	return &DirectResult{Spec: spec, Outline: outline, Budget: budget, Picks: picks}, nil
}

// comboBudget resolves how long the piece runs and over how many parts.
//
// An explicit minutes figure wins over anything written in the subject, because
// a creator who typed a number into a box has answered the question the parser
// was guessing at. Falling back to reading the subject keeps the CLI's one-string
// path working — "a 6 minute intro to recursion" has always meant six minutes,
// and it should not stop meaning that because a field exists elsewhere.
func comboBudget(req ComboRequest, subject string) RuntimeBudget {
	if req.Minutes > 0 {
		return BudgetRuntime(req.Minutes*60, defaultComboSegments)
	}
	if sec, ok := ParseRequestedRuntime(subject); ok {
		return BudgetRuntime(sec, defaultComboSegments)
	}
	// No length anywhere. Segments then take their templates' own defaults, which
	// is the right answer for a request that said nothing about how long it should
	// be — and the outliner is still told how many parts to aim for.
	return RuntimeBudget{Segments: defaultComboSegments}
}
