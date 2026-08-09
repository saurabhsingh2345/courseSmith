package pipeline

import (
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

// The caster no longer invents anything. It receives the outline's parts and
// returns one look each, and every check here is about a pick that would have
// produced a segment nobody could fill.

func castFixture() *LookCast {
	return &LookCast{Picks: []LookPick{
		{Heading: "The assumption", Template: "myth"},
		{Heading: "The three numbers", Template: "rundown"},
		{Heading: "What fits in 24GB", Template: "gauge"},
		{Heading: "The call", Template: "verdict"},
	}}
}

func TestCastAccepted(t *testing.T) {
	if err := validateLookCast(castFixture(), outlineFixture(), SkinDefault); err != nil {
		t.Fatalf("a well-formed cast was rejected: %v", err)
	}
}

// The commonest miscast in this catalog's history, and the reason the Figures
// flag exists: `gauge` over a part with no numbers in it has no ceiling and no
// candidates, so it burns a planning call and three correction rounds before
// shipping as a salvaged clip with an empty chart.
func TestCastRefusesADataLookOverAPartWithNoFigures(t *testing.T) {
	o := outlineFixture()
	o.Parts[2].Material = "the tradeoff between capacity and speed, in general terms"
	err := validateLookCast(castFixture(), o, SkinDefault)
	if err == nil {
		t.Fatal("gauge was accepted over a part containing no figures")
	}
	if !strings.Contains(err.Error(), "contains none") {
		t.Errorf("the error does not say why the pick fails: %v", err)
	}
	// And it must quote the template's own warning, which is the part that tells
	// the caster what to reach for instead.
	if !strings.Contains(err.Error(), "no threshold in it") {
		t.Errorf("the error does not carry the template's Avoid line: %v", err)
	}
}

// A template can only be cast where it can do the job its part was given. Before
// the bios, the caster satisfied the arc by labelling whatever it had picked — an
// `anatomy` declared a hook still opens by taking a URL apart.
func TestCastRefusesATemplateThatCannotCarryTheRole(t *testing.T) {
	c := castFixture()
	c.Picks[0].Template = "anatomy" // develop-only
	err := validateLookCast(c, outlineFixture(), SkinDefault)
	if err == nil {
		t.Fatal("a develop-only template was accepted as the opener")
	}
	if !strings.Contains(err.Error(), "cannot carry that job") {
		t.Errorf("unexpected error: %v", err)
	}
	// The error has to name what CAN do the job, or the caster picks another
	// develop-only look and is rejected again.
	if !strings.Contains(err.Error(), "myth") {
		t.Errorf("the error does not offer templates that can hook: %v", err)
	}
}

// The whole piece is cut in one house style. A look from another family is not a
// worse look, it is a good look from a different production.
func TestCastRefusesATemplateOutsideTheThemesPool(t *testing.T) {
	c := castFixture()
	c.Picks[1].Template = "bitfield" // foundations, so editorial-only
	err := validateLookCast(c, outlineFixture(), SkinDefault)
	if err == nil {
		t.Fatal("a foundations template was accepted in the default theme")
	}
	if !strings.Contains(err.Error(), "not offered in the default theme") {
		t.Errorf("the error does not name the theme that rejected it: %v", err)
	}
	// And the same pick is legal in the theme that batch was drawn for.
	if err := validateLookCast(c, outlineFixture(), SkinEditorial); err != nil {
		t.Fatalf("bitfield was rejected in the editorial theme it belongs to: %v", err)
	}
}

// A reply that drops or reorders an entry silently pairs every later look with
// the wrong part, producing a valid combo.yaml that describes a different video.
func TestCastCatchesAMisalignedReply(t *testing.T) {
	c := castFixture()
	c.Picks[1], c.Picks[2] = c.Picks[2], c.Picks[1]
	err := validateLookCast(c, outlineFixture(), SkinDefault)
	if err == nil {
		t.Fatal("a reply whose looks were paired with the wrong parts was accepted")
	}
	if !strings.Contains(err.Error(), "same order") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCastRequiresOneLookPerPart(t *testing.T) {
	c := castFixture()
	c.Picks = c.Picks[:3]
	err := validateLookCast(c, outlineFixture(), SkinDefault)
	if err == nil {
		t.Fatal("three looks for four parts were accepted")
	}
	if !strings.Contains(err.Error(), "do not merge, drop or add parts") {
		t.Errorf("the error does not say the parts are already decided: %v", err)
	}
}

// Rhythm is enforced rather than requested: a prompt can ask for variety and a
// model will still produce a run of identical looks when the subject leans that
// way.
func TestCastRejectsBackToBackRepeats(t *testing.T) {
	c := castFixture()
	c.Picks[2].Template = "rundown"
	err := validateLookCast(c, outlineFixture(), SkinDefault)
	if err == nil {
		t.Fatal("two identical templates in a row were accepted")
	}
	if !strings.Contains(err.Error(), "one long segment") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCastCapsRepetition(t *testing.T) {
	// Seven parts alternating two looks, so nothing is back to back and only the
	// repetition cap can fire. `illustration` is the one that repeats four times:
	// it is core, it needs no data, and it is the only look that can carry all
	// three roles — which is what lets it sit in the opener, the middle and the
	// closer of the same fixture.
	o := &ComboOutline{Title: "t", Angle: "a"}
	c := &LookCast{}
	for i := range 7 {
		role := RoleDevelop
		switch i {
		case 0:
			role = RoleHook
		case 6:
			role = RolePayoff
		}
		heading := "part " + string(rune('a'+i))
		o.Parts = append(o.Parts, ComboPart{Heading: heading, Establishes: heading, Material: "a familiar thing that maps onto it", Role: role})
		tpl := "analogy"
		if i%2 == 0 {
			tpl = "illustration"
		}
		c.Picks = append(c.Picks, LookPick{Heading: heading, Template: tpl})
	}
	err := validateLookCast(c, o, SkinDefault)
	if err == nil {
		t.Fatal("a template used four times was accepted")
	}
	if !strings.Contains(err.Error(), "nobody finishes") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCastRejectsInventedTemplate(t *testing.T) {
	c := castFixture()
	c.Picks[1].Template = "explainer"
	err := validateLookCast(c, outlineFixture(), SkinDefault)
	if err == nil {
		t.Fatal("an invented template was accepted")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCastNormalizeRepairs(t *testing.T) {
	c := castFixture()
	c.Picks[0].Template = "  MYTH  "
	c.Picks = append(c.Picks, LookPick{Template: "  ", Heading: "nothing"})
	normalizeLookCast(c)

	if c.Picks[0].Template != "myth" {
		t.Errorf("a mis-cased template stayed %q", c.Picks[0].Template)
	}
	for _, p := range c.Picks {
		if p.Template == "" {
			t.Error("a pick naming no template survived normalize")
		}
	}
}

// The catalog handed to the caster is built from the live registry, so a
// template added today is castable today. A hand-maintained list in the prompt
// would silently stop offering the newest looks.
func TestComboCatalogCarriesBiosAndRespectsThePool(t *testing.T) {
	cat := ComboCatalogForPrompt(SkinDefault)
	pool := ComboPool(SkinDefault)
	if len(pool) == 0 {
		t.Fatal("the default theme offers no templates")
	}
	for _, tpl := range pool {
		if !strings.Contains(cat, "\n  "+tpl.Name+" — ") {
			t.Errorf("template %q is in the pool but missing from the caster's catalog", tpl.Name)
		}
		// The bio is the whole reason this catalog differs from the gallery's. A
		// template shown without its requirement reads as the easiest one to
		// satisfy, which makes it the default pick.
		if !strings.Contains(cat, tpl.Bio.Needs) {
			t.Errorf("template %q is offered without its NEEDS line", tpl.Name)
		}
	}
	// A shelved template that cannot be rendered but can be picked produces a
	// video with a voiceover over an empty screen, after paying for the plan, the
	// review rounds and the audio.
	for _, name := range SnippetTemplateNames() {
		if SnippetTemplates[name].Shelved && strings.Contains(cat, "\n  "+name+" — ") {
			t.Errorf("shelved template %q is being offered to the caster", name)
		}
	}
	// And nothing from another family leaks into this theme.
	for _, name := range SnippetTemplateNames() {
		tpl := SnippetTemplates[name]
		if tpl.Family != FamilyCore && strings.Contains(cat, "\n  "+name+" — ") {
			t.Errorf("%q is family %q but appears in the default theme's catalog", name, tpl.Family)
		}
	}
}

func TestCastPromptRendersWithWhatCastLooksSupplies(t *testing.T) {
	cfg := config.Defaults()
	o := outlineFixture()
	parts := make([]map[string]any, 0, len(o.Parts))
	for i, p := range o.Parts {
		parts = append(parts, map[string]any{
			"N": i + 1, "Heading": p.Heading, "Establishes": p.Establishes,
			"Material": p.Material, "Role": p.Role,
			"HasFigures": digitRe.MatchString(p.Material),
		})
	}
	system, user, healed, err := renderPromptFileHealed(repoPromptsDir, comboCastTemplateName, map[string]any{
		"Title": o.Title, "Angle": o.Angle, "Parts": parts,
		"Catalog": ComboCatalogForPrompt(SkinDefault), "Skin": SkinDefault,
		"PoolNote": ComboPoolDescribe(SkinDefault), "MaxSame": maxSameTemplate,
		"Audience": cfg.Style.Audience, "Tone": cfg.Style.Tone,
		"Figures":   strings.Join(ComboFigureTemplates(SkinDefault), ", "),
		"HookLooks": strings.Join(ComboRoleTemplates(SkinDefault, RoleHook), ", "),
		"PayLooks":  strings.Join(ComboRoleTemplates(SkinDefault, RolePayoff), ", "),
	})
	if err != nil {
		t.Fatalf("rendering %s: %v", comboCastTemplateName, err)
	}
	if len(healed) > 0 {
		t.Errorf("the cast prompt references keys nothing supplies: %v", healed)
	}
	// The figures warning has to name the templates it applies to, and that list
	// is derived from the bios rather than written into the prompt — which is the
	// whole point of the Figures flag.
	if !strings.Contains(system, "gauge") {
		t.Error("the figure-hungry templates were not named in the prompt")
	}
	// A part with no numbers must be marked as such per part. Whether a text
	// contains a figure is exactly the judgement a model gets wrong: a part about
	// memory sounds numeric whether or not any figure is in it.
	o.Parts[0].Material = "belief: buy the biggest GPU; truth: bandwidth sets speed"
	if !strings.Contains(user, "has figures") && !strings.Contains(user, "NO FIGURES") {
		t.Error("parts are not marked with whether they contain figures")
	}
}

// A cast has to survive becoming a combo: the same rules the spec enforces.
func TestCastProducesAValidSpec(t *testing.T) {
	o := outlineFixture()
	c := castFixture()
	spec := &ComboSpec{Title: o.Title, Brief: "a brief", Angle: o.Angle}
	for i, p := range c.Picks {
		spec.Segments = append(spec.Segments, ComboSegment{
			Template: p.Template, Prompt: o.Parts[i].Establishes,
			Heading: o.Parts[i].Heading, Role: o.Parts[i].Role, Material: o.Parts[i].Material,
		})
	}
	spec.EnsureSegmentIDs()
	if err := spec.Validate(); err != nil {
		t.Fatalf("a cast combo does not validate as a spec: %v", err)
	}
	seen := map[string]bool{}
	for _, s := range spec.Segments {
		if s.ID == "" || seen[s.ID] {
			t.Errorf("segment id %q is empty or duplicated", s.ID)
		}
		seen[s.ID] = true
	}
}
