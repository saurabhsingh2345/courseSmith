package pipeline

import (
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

// The arc lives on the OUTLINE now, which is a move rather than a rewrite.
//
// The rules are the same three they always were — open on something at stake,
// close on something resolved, never go backwards — and they used to be checked
// after the looks were chosen. That was one stage too late: by then the roles
// are the only thing left that can move, so a shapeless piece gets repaired by
// relabelling segments rather than by fixing the argument. Checked here, the
// only available repair is rewriting the parts, which is the actual defect.

func outlineFixture() *ComboOutline {
	return &ComboOutline{
		Title: "What decides whether a model runs",
		Angle: "The card matters less than three numbers nobody quotes",
		Parts: []ComboPart{
			{Heading: "The assumption", Establishes: "that buying the biggest card is not the decision they think", Material: "belief: buy the biggest GPU; truth: bandwidth sets speed", Role: RoleHook},
			{Heading: "The three numbers", Establishes: "they can name capacity, bandwidth and compute", Material: "capacity GB; bandwidth TB/s; compute TFLOPs", Role: RoleDevelop},
			{Heading: "What fits in 24GB", Establishes: "which model sizes their own card can hold", Material: "24GB ceiling; 7B 14GB, 13B 26GB, 13B-4bit 8GB", Role: RoleDevelop},
			{Heading: "The call", Establishes: "what to buy and where the advice breaks", Material: "holds: under 24GB; breaks: compliance, past a few TB", Role: RolePayoff},
		},
	}
}

func TestOutlineArcAccepted(t *testing.T) {
	if err := validateOutline(outlineFixture(), 4); err != nil {
		t.Fatalf("a well-shaped outline was rejected: %v", err)
	}
}

// Opening on a definition or a ruling gives nobody a reason to watch part two.
func TestOutlineMustOpenOnAHook(t *testing.T) {
	o := outlineFixture()
	o.Parts[0].Role = RoleDevelop
	err := validateOutlineArc(o)
	if err == nil {
		t.Fatal("a piece opening on a develop part was accepted")
	}
	if !strings.Contains(err.Error(), "at stake") {
		t.Errorf("the error does not say what an opener is for: %v", err)
	}
}

// Ending mid-explanation leaves the viewer with notes rather than an answer.
func TestOutlineMustCloseOnAPayoff(t *testing.T) {
	o := outlineFixture()
	o.Parts[len(o.Parts)-1].Role = RoleDevelop
	err := validateOutlineArc(o)
	if err == nil {
		t.Fatal("a piece ending on a develop part was accepted")
	}
	if !strings.Contains(err.Error(), "resolve") {
		t.Errorf("the error does not say what a closer is for: %v", err)
	}
}

// The failure shapelessness actually looks like: hook, develop, hook, develop
// reads as four unrelated clips, because every return to a hook restarts the
// piece before the first question was paid off.
func TestOutlineArcMustNotGoBackwards(t *testing.T) {
	o := outlineFixture()
	o.Parts[2].Role = RoleHook
	err := validateOutlineArc(o)
	if err == nil {
		t.Fatal("an arc that returned to a hook halfway through was accepted")
	}
	if !strings.Contains(err.Error(), "restarts the video") {
		t.Errorf("the error does not explain why going back is fatal: %v", err)
	}
}

// A hook and a payoff with no middle promises something and rules on it without
// ever explaining it.
func TestOutlineArcNeedsAMiddle(t *testing.T) {
	o := &ComboOutline{Title: "t", Angle: "a", Parts: []ComboPart{
		{Heading: "the belief", Establishes: "x", Material: "belief: x; truth: y", Role: RoleHook},
		{Heading: "what to do", Establishes: "y", Material: "holds: a; breaks: b", Role: RolePayoff},
	}}
	if err := validateOutlineArc(o); err == nil || !strings.Contains(err.Error(), "no middle") {
		t.Fatalf("a piece with no develop part was accepted: %v", err)
	}
}

func TestOutlineRejectsAnInventedRole(t *testing.T) {
	o := outlineFixture()
	o.Parts[1].Role = "climax"
	err := validateOutline(o, 4)
	if err == nil {
		t.Fatal("an invented role was accepted")
	}
	if !strings.Contains(err.Error(), "hook, develop, payoff") {
		t.Errorf("the error does not list the vocabulary: %v", err)
	}
}

// An outline with no angle is a list of true statements about a subject, which
// is the shape a topic naturally falls into and the shape nobody finishes. It is
// also what the critic later scores every segment against, so a piece without
// one cannot be criticised either.
func TestOutlineRequiresAnAngle(t *testing.T) {
	o := outlineFixture()
	o.Angle = "   "
	err := validateOutline(o, 4)
	if err == nil {
		t.Fatal("an outline with no angle was accepted")
	}
	if !strings.Contains(err.Error(), "ARGUES") {
		t.Errorf("the error does not say what an angle is: %v", err)
	}
}

// A part with nothing concrete in it is a part the caster cannot place and the
// writer cannot fill — the failure moved one stage earlier, where it is cheap.
func TestOutlineRequiresMaterial(t *testing.T) {
	o := outlineFixture()
	o.Parts[2].Material = ""
	err := validateOutline(o, 4)
	if err == nil {
		t.Fatal("a part with no material was accepted")
	}
	if !strings.Contains(err.Error(), "merged into one that is") {
		t.Errorf("the error does not tell the model what to do: %v", err)
	}
}

// The defect a viewer recognises instantly as "we already covered that", and the
// one the whole outline stage exists to catch. Compared on the claim rather than
// the heading: two parts can be headed differently and land on the same
// increment, which is how a piece repeats itself while looking varied on paper.
func TestOutlineRejectsTwoPartsThatEstablishTheSameThing(t *testing.T) {
	o := outlineFixture()
	o.Parts[2].Heading = "A different heading entirely"
	o.Parts[2].Establishes = "They can name the capacity, bandwidth, and compute."
	err := validateOutline(o, 4)
	if err == nil {
		t.Fatal("two parts establishing the same thing were accepted")
	}
	if !strings.Contains(err.Error(), "establish the same thing") {
		t.Errorf("unexpected error: %v", err)
	}
}

// Casing and whitespace are the model's mechanical mistakes, repaired before the
// shape is judged.
func TestOutlineNormalizesTheRole(t *testing.T) {
	o := &ComboOutline{Parts: []ComboPart{
		{Heading: "h", Establishes: "e", Material: "m", Role: "  HOOK "},
	}}
	normalizeOutline(o)
	if o.Parts[0].Role != RoleHook {
		t.Errorf("role = %q, want %q", o.Parts[0].Role, RoleHook)
	}
}

// The outliner is given the facts AND the gaps, and the gaps are the
// load-bearing half: a part whose numbers were looked for and not found must not
// be written at all, because the segment would have to invent them.
func TestOutlinePromptCarriesFactsAndGaps(t *testing.T) {
	cfg := config.Defaults()
	sub := substanceFixture()
	sub.Gaps = []string{"No current figure for professional no-code adoption"}
	sub.Misconceptions = []string{"You need a CS degree to build software"}

	system, _, healed, err := renderPromptFileHealed(repoPromptsDir, comboOutlineTemplateName, comboOutlinePromptData(cfg, sub))
	if err != nil {
		t.Fatalf("rendering %s: %v", comboOutlineTemplateName, err)
	}
	if len(healed) > 0 {
		t.Errorf("the outline prompt references keys nothing supplies: %v", healed)
	}
	if !strings.Contains(system, "https://webflow.com/pricing") {
		t.Error("the outliner was not given the facts, so its material can only be a guess")
	}
	if !strings.Contains(system, "professional no-code adoption") {
		t.Error("the outliner was not given the gaps")
	}
	if !strings.Contains(system, "CS degree") {
		t.Error("the audience's beliefs never reached the outliner, which is where the strongest hooks are")
	}
	for _, r := range []string{RoleHook, RoleDevelop, RolePayoff} {
		if !strings.Contains(system, r) {
			t.Errorf("the prompt never mentions the %q role it will be validated on", r)
		}
	}
	// The outliner must not be shown the catalog. Seeing it is what made the old
	// caster write toward templates instead of toward an argument.
	if strings.Contains(system, "gauge") || strings.Contains(system, "constellation") {
		t.Error("the outline prompt names templates — it must not see the catalog, or it writes toward looks rather than toward the argument")
	}
}

// An ungrounded outline renders without those blocks and behaves as it did
// before the fact sheet existed.
func TestOutlinePromptWithoutASheet(t *testing.T) {
	system, _, healed, err := renderPromptFileHealed(repoPromptsDir, comboOutlineTemplateName, comboOutlinePromptData(config.Defaults(), nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(healed) > 0 {
		t.Errorf("unsupplied keys: %v", healed)
	}
	if strings.Contains(system, "THE FACTS THIS PIECE HAS") {
		t.Error("the facts block rendered with nothing in it")
	}
	if strings.Contains(system, "WHAT WAS LOOKED FOR AND NOT FOUND") {
		t.Error("the gaps block rendered with nothing in it")
	}
	if !strings.Contains(system, "hook → develop → payoff") {
		t.Error("the arc instruction is missing from an ungrounded outline")
	}
}

// comboOutlinePromptData mirrors what OutlineCombo supplies, so a key added there and
// forgotten in the prompt (or the reverse) shows up as a healed key here.
func comboOutlinePromptData(cfg config.Config, sub *Substance) map[string]any {
	return map[string]any{
		"Subject": "a subject", "Title": "",
		"WantParts": 5, "MinParts": minComboSegments, "MaxParts": maxComboSegments,
		"PerPartSec": 0, "TotalSec": 0,
		"Audience": cfg.Style.Audience, "Tone": cfg.Style.Tone,
		"MaxPartWords": maxPartWords, "MaxMaterial": maxOutlineMaterialWords,
		"Roles":          strings.Join([]string{RoleHook, RoleDevelop, RolePayoff}, ", "),
		"Facts":          substanceLines(sub),
		"Gaps":           substanceGaps(sub),
		"Misconceptions": substanceMisconceptions(sub),
	}
}
