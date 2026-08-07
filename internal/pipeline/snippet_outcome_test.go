package pipeline

import (
	"strings"
	"testing"
)

const outcNarration = "By the end of this you can look at a memory dump and say what you see."

func outcomePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "outcome",
		Title:    "What memory will stop hiding",
		Outcome: &OutcomeSpec{
			Lesson: "Memory & Storage",
			Abilities: []OutcomeAbility{
				{Skill: "Read a hex dump without guessing", Payoff: "crash logs stop looking like random letters"},
				{Skill: "Trace a byte from RAM to disk", Payoff: "you can say where your data actually sits"},
				{Skill: "Size a cache for a workload", Payoff: "you stop buying memory you will never use"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "promise", Heading: "Three things", Narration: outcNarration, Outcome: &OutcomeBeat{Show: "promise"}},
			{ID: "hex-dumps", Heading: "Reading a dump", Narration: outcNarration, Outcome: &OutcomeBeat{Show: "ability", At: 0}},
			{ID: "the-trace", Heading: "Following a byte", Narration: outcNarration, Outcome: &OutcomeBeat{Show: "ability", At: 1}},
			{ID: "sizing", Heading: "Sizing a cache", Narration: outcNarration, Outcome: &OutcomeBeat{Show: "ability", At: 2}},
			{ID: "contract", Heading: "The whole deal", Narration: outcNarration, Outcome: &OutcomeBeat{Show: "contract"}},
		},
	}
	p.targetWords = 5 * 40
	return p
}

func TestOutcomePlanAccepted(t *testing.T) {
	if err := validateOutcomePlan(outcomePlan()); err != nil {
		t.Fatalf("a well-formed outcome plan was rejected: %v", err)
	}
}

func TestOutcomeRejectsASinglePromise(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Abilities = p.Outcome.Abilities[:1]
	err := validateOutcomePlan(p)
	if err == nil {
		t.Fatal("a one-ability lesson opener was accepted, and one promise is a title")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Fatalf("the error does not explain why: %v", err)
	}
}

func TestOutcomeRejectsTooManyPromises(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Abilities = append(p.Outcome.Abilities,
		OutcomeAbility{Skill: "Draw the storage stack", Payoff: "you can sketch it for a colleague"},
		OutcomeAbility{Skill: "Decode a permission byte", Payoff: "file errors stop being mysterious"},
	)
	if err := validateOutcomePlan(p); err == nil {
		t.Fatal("a five-ability lesson was accepted, and a lesson promising five will deliver two")
	}
}

// The rule this template exists for, and the error that teaches rather than
// sending the model looking for a synonym.
func TestOutcomeRejectsABeliefVerbWithItsReason(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Abilities[0].Skill = "Understand how RAM works"
	err := validateOutcomePlan(p)
	if err == nil {
		t.Fatal("an ability nobody could be watched doing was accepted")
	}
	if !strings.Contains(err.Error(), "understand") {
		t.Fatalf("the error does not quote the refused verb: %v", err)
	}
	if !strings.Contains(err.Error(), "nobody has ever watched someone understand") {
		t.Fatalf("the error gives no reason, so the model will just reach for a synonym: %v", err)
	}
}

func TestOutcomeRejectsAVerbOutsideTheVocabulary(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Abilities[1].Skill = "Enjoy a hex dump"
	err := validateOutcomePlan(p)
	if err == nil {
		t.Fatal("an ability opening on an unlisted verb was accepted")
	}
	if !strings.Contains(err.Error(), "enjoy") {
		t.Fatalf("the error does not quote the verb: %v", err)
	}
	if !strings.Contains(err.Error(), "trace") {
		t.Fatalf("the error does not offer the vocabulary: %v", err)
	}
}

func TestOutcomeRejectsAnAbilityWithNoSkill(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Abilities[0].Skill = "   "
	if err := validateOutcomePlan(p); err == nil {
		t.Fatal("an ability naming no skill was accepted")
	}
}

func TestOutcomeRejectsAMissingPayoff(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Abilities[2].Payoff = ""
	err := validateOutcomePlan(p)
	if err == nil {
		t.Fatal("an ability with no job-relevance was accepted, and that is a syllabus line rather than a skill")
	}
	if !strings.Contains(err.Error(), "Size a cache for a workload") {
		t.Fatalf("the error does not quote the ability it means: %v", err)
	}
}

func TestOutcomeRejectsARepeatedPromise(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Abilities[1].Skill = p.Outcome.Abilities[0].Skill
	err := validateOutcomePlan(p)
	if err == nil {
		t.Fatal("the same promise made twice was accepted, and the count chip then claims one thing too many")
	}
	if !strings.Contains(err.Error(), "repeats an earlier one") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOutcomeRequiresOpeningOnThePromise(t *testing.T) {
	p := outcomePlan()
	p.Beats[0].Outcome = &OutcomeBeat{Show: "ability", At: 0}
	p.Beats[1].Outcome = &OutcomeBeat{Show: "promise"}
	if err := validateOutcomePlan(p); err == nil {
		t.Fatal("a card stamping onto a ledger nobody has been promised was accepted")
	}
}

func TestOutcomeRejectsPromisingTwice(t *testing.T) {
	p := outcomePlan()
	p.Beats[2].Outcome = &OutcomeBeat{Show: "promise"}
	err := validateOutcomePlan(p)
	if err == nil {
		t.Fatal("a clip that makes the promise again mid-fill was accepted")
	}
	if !strings.Contains(err.Error(), "resets a ledger") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOutcomeRejectsLandingAnAbilityTwice(t *testing.T) {
	p := outcomePlan()
	p.Beats[3].Outcome = &OutcomeBeat{Show: "ability", At: 0}
	err := validateOutcomePlan(p)
	if err == nil {
		t.Fatal("a card stamped twice was accepted, and it breaks the count on screen")
	}
	if !strings.Contains(err.Error(), "breaks the count") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOutcomeRejectsAnAbilityOffTheLedger(t *testing.T) {
	p := outcomePlan()
	p.Beats[1].Outcome = &OutcomeBeat{Show: "ability", At: 9}
	if err := validateOutcomePlan(p); err == nil {
		t.Fatal("a beat landing an ability that does not exist was accepted")
	}
}

func TestOutcomeRejectsTheContractBeforeTheEnd(t *testing.T) {
	p := outcomePlan()
	p.Beats[2].Outcome = &OutcomeBeat{Show: "contract"}
	err := validateOutcomePlan(p)
	if err == nil {
		t.Fatal("a full ledger shown mid-clip was accepted")
	}
	if !strings.Contains(err.Error(), "already claimed to be complete") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOutcomeRequiresClosingOnTheContract(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Abilities = append(p.Outcome.Abilities,
		OutcomeAbility{Skill: "Decode a permission byte", Payoff: "file errors stop being mysterious"})
	p.Beats[4].Outcome = &OutcomeBeat{Show: "ability", At: 3}
	err := validateOutcomePlan(p)
	if err == nil {
		t.Fatal("a clip that never lights the whole ledger was accepted")
	}
	if !strings.Contains(err.Error(), "close on the contract") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOutcomeRejectsAnAbilityNeverLanded(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Abilities = append(p.Outcome.Abilities,
		OutcomeAbility{Skill: "Decode a permission byte", Payoff: "file errors stop being mysterious"})
	err := validateOutcomePlan(p)
	if err == nil {
		t.Fatal("a promise on the ledger that no beat says out loud was accepted")
	}
	if !strings.Contains(err.Error(), "never landed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// "You will be able to trace..." and "Trace..." are the same outcome, and only
// one of them opens with a verb the vocabulary can see.
func TestOutcomeNormalizeStripsTheLeadIn(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Abilities[0].Skill = "You will be able to read a hex dump"
	p.Outcome.Abilities[1].Skill = "You'll trace a byte to disk"
	normalizeOutcomePlan(p)
	if got := p.Outcome.Abilities[0].Verb(); got != "read" {
		t.Fatalf("the lead-in survived: verb is %q", got)
	}
	if got := p.Outcome.Abilities[1].Verb(); got != "trace" {
		t.Fatalf("the contracted lead-in survived: verb is %q", got)
	}
	if err := validateOutcomePlan(p); err != nil {
		t.Fatalf("a plan whose skills only wore a lead-in was rejected after normalize: %v", err)
	}
}

func TestOutcomeNormalizeClampsAndCapsTheLedger(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Lesson = "the lesson about memory and storage and caching"
	p.Outcome.Abilities[0].Skill = "Read a hex dump of a running process without any guesswork"
	p.Outcome.Abilities[0].Payoff = "crash logs and core dumps stop looking like a wall of random letters"
	p.Outcome.Abilities = append(p.Outcome.Abilities,
		OutcomeAbility{Skill: "Draw the storage stack", Payoff: "you can sketch it"},
		OutcomeAbility{Skill: "Decode a permission byte", Payoff: "file errors stop being mysterious"})
	normalizeOutcomePlan(p)
	if n := len(strings.Fields(p.Outcome.Lesson)); n != maxOutcomeLessonWords {
		t.Fatalf("the lesson chip survived at %d words", n)
	}
	if n := len(strings.Fields(p.Outcome.Abilities[0].Skill)); n != maxOutcomeSkillWords {
		t.Fatalf("a skill survived at %d words", n)
	}
	if n := len(strings.Fields(p.Outcome.Abilities[0].Payoff)); n != maxOutcomePayoffWords {
		t.Fatalf("a payoff survived at %d words", n)
	}
	if n := len(p.Outcome.Abilities); n != maxOutcomeAbilities {
		t.Fatalf("want %d abilities after normalize, got %d", maxOutcomeAbilities, n)
	}
}

func TestOutcomeNormalizeClampsAnAbilityOffTheLedger(t *testing.T) {
	p := outcomePlan()
	p.Beats[1].Outcome.At = 99
	p.Beats[4].Outcome.At = 2
	normalizeOutcomePlan(p)
	if at := p.Beats[1].Outcome.At; at != len(p.Outcome.Abilities)-1 {
		t.Fatalf("want the ability clamped to the last card, got %d", at)
	}
	// The contract lights everything, so an index on it is noise.
	if at := p.Beats[4].Outcome.At; at != 0 {
		t.Fatalf("the contract beat kept its index %d", at)
	}
}

func TestOutcomeShowDefaultsToAbility(t *testing.T) {
	b := OutcomeBeat{Show: "boast"}
	if got := b.ResolvedShow(); got != "ability" {
		t.Fatalf("an unknown show resolved to %q, want ability", got)
	}
	b = OutcomeBeat{Show: " CONTRACT "}
	if got := b.ResolvedShow(); got != "contract" {
		t.Fatalf("a shouted contract resolved to %q", got)
	}
}

// The ledger accumulates and the contract lights everything, which is the frame
// that gets screenshotted.
func TestOutcomeScenesFillTheLedger(t *testing.T) {
	p := outcomePlan()
	scenes, err := outcomeScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props
	abilities, _ := props["abilities"].([]map[string]any)
	if len(abilities) != 3 {
		t.Fatalf("want 3 ability cards, got %d", len(abilities))
	}
	if abilities[0]["payoff"] != "crash logs stop looking like random letters" {
		t.Fatalf("the first card lost its payoff: %v", abilities[0])
	}

	steps, _ := props["steps"].([]map[string]any)
	first := steps[0]
	lit, _ := first["lit"].([]int)
	if first["show"] != "promise" || len(lit) != 0 {
		t.Fatalf("the promise beat already lights cards: %v", first)
	}
	last := steps[len(steps)-1]
	on, _ := last["lit"].([]int)
	if last["show"] != "contract" || len(on) != len(p.Outcome.Abilities) {
		t.Fatalf("the contract does not light every card: %v", last)
	}
	for i, at := range on {
		if at != i {
			t.Fatalf("the lit set is not sorted or complete: %v", on)
		}
	}
}

// The opener is the lesson's name over the count of what it promises, so a
// plan without a name opens on a number attached to nothing.
func TestOutcomeRequiresTheLessonName(t *testing.T) {
	p := outcomePlan()
	p.Outcome.Lesson = ""
	err := validateOutcomePlan(p)
	if err == nil {
		t.Fatal("an outcome with no lesson name was accepted")
	}
	if !strings.Contains(err.Error(), "name the lesson") {
		t.Fatalf("the error does not say what is missing: %v", err)
	}
}

// The belief-verb check matches whole verbs. A prefix test would read
// "understand" inside any longer allowed verb starting the same way and reject
// it with a reason that is not true.
func TestOutcomeBeliefCheckDoesNotMatchOnAPrefix(t *testing.T) {
	for _, verb := range OutcomeVerbs() {
		for belief := range outcomeBeliefVerbs {
			if verb != belief && strings.HasPrefix(verb, belief) {
				t.Errorf("allowed verb %q starts with refused verb %q; a prefix match would reject it as a belief", verb, belief)
			}
		}
	}
}
