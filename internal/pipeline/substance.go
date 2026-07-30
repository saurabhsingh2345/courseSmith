package pipeline

// The substance stage: establish the facts before anything chooses a picture.
//
// This inverts the order the short-form path used to run in, and the inversion is
// the whole point. It used to be:
//
//	prompt → pick a template → fill that template's schema
//
// A template's schema is a demand. `gauge` demands a numeric ceiling and things
// measured against it; `costing` demands line items that add up. Handed a subject
// with no figures in it, a writer asked to fill that schema does not fail — it
// invents. That is not a hypothetical: a reel cast from a brief about no-code
// tools produced a gauge captioned "MAX CONCEPTS IN A SESSION — 10 concepts",
// bars at 2/5/10/12, rendered as authoritative data. The number does not exist.
// The same run produced a developer named Emily whose app got "thousands of
// downloads".
//
// So the facts come first, and they come with provenance:
//
//	brief → establish the facts → pick looks those facts can fill → fill them
//
// The stage's output (substance.json) is what casting reads to choose templates
// and what each segment's writer draws on. A template whose material would have
// to be invented is now a template the caster can see it should not pick.
//
// == Why provenance and not just facts ==
//
// A fact sheet the model wrote from memory is the same fabrication engine with an
// extra step. What makes this load-bearing is that every fact says where it came
// from, and only facts with a real origin may be rendered:
//
//	given     — stated in the brief. The creator's own words; the strongest kind.
//	sourced   — from a web search, with the URL that backs it.
//	derived   — arithmetic on given or sourced facts, with the working shown.
//	unverified — the model believes it. NOT renderable.
//
// `unverified` exists rather than being dropped because a model asked to omit
// what it cannot support will instead quietly relabel it. Giving the doubt a
// place to live is what makes the label honest, and the gate then refuses that
// label rather than trusting a silence.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// SubstanceFileName is the fact sheet, in the lesson's generated dir.
const SubstanceFileName = "substance.json"

const substanceTemplateName = "substance.tmpl"

// Provenance is where a fact came from. See the package comment for why the
// vocabulary is closed and why `unverified` is one of the options.
type Provenance string

const (
	// ProvGiven is stated in the brief.
	ProvGiven Provenance = "given"
	// ProvSourced came from a web search and carries the URL that backs it.
	ProvSourced Provenance = "sourced"
	// ProvDerived is arithmetic on other facts, with the working in Working.
	ProvDerived Provenance = "derived"
	// ProvUnverified is the model's own belief, and may not be rendered.
	ProvUnverified Provenance = "unverified"
)

// Renderable reports whether a template may put this fact on screen.
func (p Provenance) Renderable() bool {
	return p == ProvGiven || p == ProvSourced || p == ProvDerived
}

// Fact is one thing the piece may state as true.
type Fact struct {
	// Claim is the fact in one sentence, with its units and its subject named.
	Claim string `json:"claim"`
	// Provenance is where it came from. Anything but the four known values is
	// normalized to unverified, which is the safe direction to fail.
	Provenance Provenance `json:"provenance"`
	// Source is the URL backing a sourced fact. Required for `sourced` and
	// meaningless otherwise.
	Source string `json:"source,omitempty"`
	// Working is the arithmetic behind a derived fact ("70B params x 2 bytes =
	// 140GB"). Required for `derived`: a derivation nobody can check is a belief
	// with a confident tone.
	Working string `json:"working,omitempty"`
}

// Substance is the fact sheet for one piece.
type Substance struct {
	// Subject is what the piece is actually about, in one line.
	Subject string `json:"subject"`
	// Facts are the renderable and non-renderable claims, in no order.
	Facts []Fact `json:"facts"`
	// Misconceptions are beliefs the audience genuinely holds, for `myth` to
	// correct. Kept apart from Facts because a misconception is a fact ABOUT the
	// audience, and a myth segment that draws from the general pool ends up
	// correcting a strawman nobody believed.
	Misconceptions []string `json:"misconceptions,omitempty"`
	// Examples are concrete worked cases — real names, real numbers.
	Examples []string `json:"examples,omitempty"`
	// Gaps are what could not be established. Written down rather than discarded
	// because a gap is the caster's best signal that a data-hungry template is
	// the wrong choice for that part.
	Gaps []string `json:"gaps,omitempty"`
	// Sources are every URL the search leaned on, from the provider's own
	// citations rather than from the model's prose — so the list cannot itself be
	// fabricated.
	Sources []llm.Citation `json:"sources,omitempty"`
	// Grounded records whether a search actually ran. False means every fact here
	// is `given`, `derived` or `unverified`, which changes how much the gate
	// downstream can trust it.
	Grounded bool `json:"grounded"`
}

// Renderable returns the facts a template may put on screen.
func (s Substance) Renderable() []Fact {
	out := make([]Fact, 0, len(s.Facts))
	for _, f := range s.Facts {
		if f.Provenance.Renderable() {
			out = append(out, f)
		}
	}
	return out
}

// Validate enforces what provenance is for. A label without its backing is the
// failure this whole stage exists to prevent, so it is rejected rather than
// downgraded — downgrading would teach the model that `sourced` is free.
func (s *Substance) Validate() error {
	if strings.TrimSpace(s.Subject) == "" {
		return fmt.Errorf("substance has no subject — say in one line what the piece is about")
	}
	if len(s.Facts) == 0 {
		return fmt.Errorf("substance has no facts. Even a brief with no figures in it states things; list what it states")
	}
	renderable := 0
	for i, f := range s.Facts {
		if strings.TrimSpace(f.Claim) == "" {
			return fmt.Errorf("fact %d has no claim", i+1)
		}
		switch f.Provenance {
		case ProvGiven, ProvUnverified:
		case ProvSourced:
			if strings.TrimSpace(f.Source) == "" {
				return fmt.Errorf("fact %d (%q) is marked sourced but names no URL. A sourced fact without its source is an unverified one: either give the URL you read it in, or mark it unverified",
					i+1, truncateForLog(f.Claim, 60))
			}
		case ProvDerived:
			if strings.TrimSpace(f.Working) == "" {
				return fmt.Errorf("fact %d (%q) is marked derived but shows no working. Show the arithmetic, or mark it unverified",
					i+1, truncateForLog(f.Claim, 60))
			}
		default:
			return fmt.Errorf("fact %d (%q) has provenance %q, which is not one of given, sourced, derived, unverified",
				i+1, truncateForLog(f.Claim, 60), f.Provenance)
		}
		if f.Provenance.Renderable() {
			renderable++
		}
	}
	if renderable == 0 {
		return fmt.Errorf("every fact is unverified, so the piece has nothing it may actually state. Work from what the brief says")
	}
	return nil
}

// normalizeSubstance repairs the model's mechanical mistakes before Validate
// judges it: casing on a provenance label, whitespace, a source that is not a
// URL. An unrecognised label becomes `unverified` rather than an error, because
// the safe direction to fail is "cannot be rendered".
func normalizeSubstance(s *Substance) {
	s.Subject = collapseSpaces(s.Subject)
	for i := range s.Facts {
		f := &s.Facts[i]
		f.Claim = collapseSpaces(f.Claim)
		f.Working = collapseSpaces(f.Working)
		f.Source = strings.TrimSpace(f.Source)
		f.Provenance = Provenance(strings.ToLower(strings.TrimSpace(string(f.Provenance))))
		// A "source" that is not a URL is prose the model wrote to satisfy the
		// field, which is worse than an empty one: it looks like a citation.
		if f.Provenance == ProvSourced && !strings.HasPrefix(f.Source, "http") {
			f.Source = ""
		}
	}
	s.Misconceptions = collapseEach(s.Misconceptions)
	s.Examples = collapseEach(s.Examples)
	s.Gaps = collapseEach(s.Gaps)
}

func collapseEach(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s = collapseSpaces(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// LoadSubstance reads the fact sheet a lesson's substance stage wrote.
//
// A missing file is (nil, nil), not an error: a snippet created before this
// stage existed, or one resumed from a later stage, has none, and planning
// without a fact sheet is how the whole pipeline worked until now. Callers pass
// the nil straight through.
// A corrupt file is an ERROR, not an absence — deliberately unlike readJSONFile,
// which treats the two the same. Silently planning ungrounded because the fact
// sheet failed to parse is precisely the quiet degradation this stage exists to
// remove: the plan would come out looking exactly like a grounded one.
func LoadSubstance(l *project.Lesson) (*Substance, error) {
	path := filepath.Join(l.GeneratedDir(), SubstanceFileName)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var sub Substance
	if err := json.Unmarshal(data, &sub); err != nil {
		return nil, fmt.Errorf("parsing %s: %w — delete it and re-run the substance stage", path, err)
	}
	return &sub, nil
}

// substanceLines renders the renderable facts one per line, each carrying its
// backing, for a prompt to drop in.
//
// Only the renderable ones. An `unverified` fact shown to a writer is a fact
// that writer will use — the label means something to this code and nothing to a
// model reading a list — so the filter happens here rather than being requested
// in the prompt. Nil substance gives nil, and the prompt's {{if}} handles it.
func substanceLines(s *Substance) []string {
	if s == nil {
		return nil
	}
	facts := s.Renderable()
	out := make([]string, 0, len(facts))
	for _, f := range facts {
		switch {
		case f.Source != "":
			out = append(out, fmt.Sprintf("%s [%s]", f.Claim, f.Source))
		case f.Working != "":
			out = append(out, fmt.Sprintf("%s [%s]", f.Claim, f.Working))
		default:
			out = append(out, f.Claim)
		}
	}
	return out
}

// substanceGaps renders what could not be established, so a writer knows which
// specifics it must not reach for. Merged with the examples' absence: a gap is
// the one part of the sheet whose value is entirely negative.
func substanceGaps(s *Substance) []string {
	if s == nil {
		return nil
	}
	return s.Gaps
}

// runSubstanceStage establishes the facts for a snippet or a reel.
func runSubstanceStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	brief, err := substanceBrief(l)
	if err != nil {
		return err
	}
	if e.Router == nil {
		return fmt.Errorf("the substance stage needs an LLM — set OPENAI_API_KEY and retry")
	}

	sub, err := e.establishSubstance(ctx, brief, cfg)
	if err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "    %d facts (%d renderable), %d sources%s\n",
		len(sub.Facts), len(sub.Renderable()), len(sub.Sources), groundedNote(sub.Grounded))
	for _, g := range sub.Gaps {
		fmt.Fprintf(e.out(), "    gap: %s\n", truncateForLog(g, 80))
	}
	return writeJSON(filepath.Join(l.GeneratedDir(), SubstanceFileName), sub)
}

func groundedNote(grounded bool) string {
	if grounded {
		return ""
	}
	return " (ungrounded — llm_search is off, so nothing was looked up)"
}

// substanceBrief is the whole input to establish facts from: a reel's brief, or
// a snippet's prompt.
func substanceBrief(l *project.Lesson) (string, error) {
	if IsReel(l) {
		spec, err := LoadReelSpec(l.Dir)
		if err != nil {
			return "", err
		}
		brief := strings.TrimSpace(spec.Brief)
		if brief == "" {
			// A hand-authored reel need not have one; the segment prompts are
			// then all there is to go on, and they are better than nothing.
			var parts []string
			for _, seg := range spec.Active() {
				parts = append(parts, seg.Prompt)
			}
			brief = strings.Join(parts, "; ")
		}
		if spec.Title != "" {
			brief = spec.Title + ". " + brief
		}
		return brief, nil
	}
	spec, err := LoadSnippetSpec(l.Dir)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(spec.Prompt), nil
}

// establishSubstance is the two calls that produce the fact sheet.
//
// TWO calls, not one, and the split is forced rather than chosen. The
// search-capable models live on the chat-completions search path, which does not
// take the JSON response format the rest of this pipeline is built on — and a
// grounded call is exactly the one where a parse failure is most expensive,
// because it has already paid for the search. So:
//
//	1. search: prose about the subject, with the provider's own citations.
//	2. structure: an ordinary JSON call that turns that prose into the fact sheet.
//
// The second call is told to treat the first's prose as its source material,
// which is why it can label things `sourced` at all. With grounding off, step 1
// is skipped and step 2 works from the brief alone — every fact then comes back
// `given` or `unverified`, which is honest rather than degraded.
func (e *Env) establishSubstance(ctx context.Context, brief string, cfg config.Config) (*Substance, error) {
	var (
		findings  string
		citations []llm.Citation
		grounded  bool
	)
	if llm.SearchEnabled(cfg.Pipeline) {
		var err error
		findings, citations, err = e.searchForSubstance(ctx, brief, cfg)
		if err != nil {
			// Grounding is best-effort at the stage level even though it is
			// strict at the provider level. A search that fails — no key, a
			// model that is not search-capable, a rate limit — should not stop a
			// creator making a video; it should produce a fact sheet that says
			// plainly it was not grounded, and let the gate downstream be
			// correspondingly less trusting.
			fmt.Fprintf(e.out(), "    ! grounding failed, continuing from the brief alone: %s\n", errSummary(err))
		} else {
			grounded = true
		}
	}

	system, user, err := e.renderPrompt(substanceTemplateName, map[string]any{
		"Brief":    brief,
		"Findings": findings,
		"Audience": cfg.Style.Audience,
	})
	if err != nil {
		return nil, err
	}

	var sub Substance
	err = e.completeJSONLenientRounds(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.2, 4096, substanceRepairRounds, &sub, func() error {
		normalizeSubstance(&sub)
		return sub.Validate()
	})
	if err != nil {
		return nil, err
	}
	sub.Grounded = grounded
	// The citations come from the provider, not from the prose, so the source
	// list cannot be invented even if the fact labels are wrong.
	sub.Sources = citations
	dropUncitedSources(&sub, citations, grounded)
	return &sub, nil
}

// substanceRepairRounds is how many times the model is told what it got wrong.
// Three, matching the other structured stages: the rules here are mechanical
// (a label needs its backing) rather than a stack of interacting budgets.
const substanceRepairRounds = 3

// dropUncitedSources downgrades any `sourced` fact whose URL was not among the
// citations the provider actually returned.
//
// This is the check the whole provenance scheme rests on. Asking a model to
// label its own facts and then believing the labels is not verification — a
// model that wants to mark something `sourced` can write any plausible URL, and
// a plausible URL is indistinguishable from a real one by eye. The provider's
// citation list is the one part of this that the model did not author, so it is
// the only thing worth checking against.
func dropUncitedSources(s *Substance, citations []llm.Citation, grounded bool) {
	cited := make(map[string]bool, len(citations))
	for _, c := range citations {
		cited[c.URL] = true
	}
	for i := range s.Facts {
		f := &s.Facts[i]
		if f.Provenance != ProvSourced {
			continue
		}
		// With grounding off nothing was cited, so every `sourced` label is
		// wrong by construction.
		if !grounded || !cited[f.Source] {
			f.Provenance = ProvUnverified
			f.Source = ""
		}
	}
}

// searchForSubstance runs the grounded call and returns its prose and sources.
func (e *Env) searchForSubstance(ctx context.Context, brief string, cfg config.Config) (string, []llm.Citation, error) {
	system, user, err := e.renderPrompt(substanceSearchTemplateName, map[string]any{
		"Brief":    brief,
		"Audience": cfg.Style.Audience,
	})
	if err != nil {
		return "", nil, err
	}
	// Straight to the router rather than through a completeJSON helper: this is
	// the one call in the pipeline that wants prose, and routing it here still
	// puts it through cache → retry → rate limit like everything else.
	resp, err := e.Router.Complete(ctx, cfg.Pipeline, llm.TaskSearch, llm.Request{
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: system},
			{Role: llm.RoleUser, Content: user},
		},
		MaxTokens: 3000,
		WebSearch: true,
	})
	if err != nil {
		return "", nil, err
	}
	return resp.Content, resp.Citations, nil
}

const substanceSearchTemplateName = "substance_search.tmpl"
