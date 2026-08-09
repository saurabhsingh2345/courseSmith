package pipeline

// The pool: which templates a theme is allowed to cast from.
//
// A combo cuts between several looks, and the failure that gets reported as
// "that clip did not belong" is usually not a bad clip. It is a good clip from
// another production. The replica batch is drawn for a near-black stage with
// standing chrome; the foundations batch is composed against the editorial
// skin's hard left axis, where the headline holds one edge and the picture is
// deliberately unbalanced. Each is right in its own house and each reads as an
// intrusion in the other's — so a piece that mixes them freely is a piece that
// changes production partway through, however well every individual segment was
// planned.
//
// This is why the family exists at all (snippet_categories.go), and until now it
// only gated which studio PAGE offered a template. A caster handed the whole
// registry ignored it entirely, which is how a bit-row diagram ended up in the
// middle of a broadcast explainer.
//
// So the theme decides the pool. That is a narrowing, and it is the right
// trade: the alternative is asking a model to hold a visual-consistency rule it
// has no way to see, and a rule a model cannot see is a rule that is obeyed
// about two thirds of the time. Two thirds is exactly the rate at which a piece
// contains one segment that does not belong.
//
// The core family is in every pool. That is not a hedge — those forty-one
// templates predate the split and are drawn to be skin-neutral, so they inherit
// whichever house style they are rendered in rather than fighting it.

import (
	"fmt"
	"sort"
	"strings"
)

// comboPools maps a skin onto the families it may cast from.
//
// Written as data rather than a switch because the interesting property is that
// it is TOTAL: every skin the renderer can produce appears here, so a skin added
// to videoskin.go and forgotten here fails a test rather than silently casting
// from nothing. See TestEveryPoolCanStaffTheArc and TestEverySkinHasAPool.
var comboPools = map[string][]string{
	// The catalog's own look. Core only: the default skin is what the replica and
	// foundations batches were explicitly drawn NOT to assume, so admitting them
	// here is the mixing this file exists to prevent.
	SkinDefault: {FamilyCore},
	// Flat charcoal, one accent, no furniture. Core only for the same reason, and
	// one more: the replica batch's whole vocabulary is standing chrome, which is
	// the thing minimal removes.
	SkinMinimal: {FamilyCore},
	// The near-black explainer stage the replica batch was cut against, frame by
	// frame, from the reference videos. This is the pool those twelve templates
	// were built for and the only one they read correctly in.
	SkinBroadcast: {FamilyCore, FamilyReplica},
	// The hard left axis the foundations batch composes against. Thirty templates
	// for teaching computing itself — bit rows, call stacks, packet journeys —
	// which is also why a CS subject should be directed in this theme rather than
	// in default: the pool is where its pictures live.
	SkinEditorial: {FamilyCore, FamilyFoundations},
}

// ComboSkins returns the themes a combo can be cut in, in display order.
//
// Ordered rather than ranged over the map, because this backs a picker and a
// picker whose options move between page loads is one nobody builds a habit
// with.
func ComboSkins() []string {
	return []string{SkinDefault, SkinBroadcast, SkinMinimal, SkinEditorial}
}

// ComboPool returns the templates a director may cast from in this theme,
// sorted by name.
//
// Shelved templates are excluded, because SnippetTemplateList excludes them and
// this is an offer rather than an enumeration — the same distinction that file
// draws. A shelved template named by hand in a combo.yaml still plans and still
// renders; it is only kept out of what the machine chooses from.
func ComboPool(skin string) []*SnippetTemplate {
	families := comboPools[normalizeSkin(skin)]
	allowed := make(map[string]bool, len(families))
	for _, f := range families {
		allowed[f] = true
	}
	out := make([]*SnippetTemplate, 0, len(SnippetTemplates))
	for _, t := range SnippetTemplateList() {
		if allowed[t.Family] {
			out = append(out, t)
		}
	}
	return out
}

// InComboPool reports whether a template may be cast in this theme.
//
// The check the cast validator runs. Separate from ComboPool so a single name
// can be tested without building the whole list, and because the ERROR wants to
// say which theme rejected it — a director told only "not in the catalog" will
// pick another template from the same family and be rejected again.
func InComboPool(skin, name string) bool {
	tpl, ok := SnippetTemplates[name]
	if !ok || tpl.Shelved {
		return false
	}
	for _, f := range comboPools[normalizeSkin(skin)] {
		if tpl.Family == f {
			return true
		}
	}
	return false
}

// ComboPoolDescribe is the one line a person or a log reads to understand why
// the choice was narrowed.
func ComboPoolDescribe(skin string) string {
	skin = normalizeSkin(skin)
	pool := ComboPool(skin)
	names := make([]string, 0, 3)
	for _, f := range comboPools[skin] {
		switch f {
		case FamilyCore:
			names = append(names, "the core catalog")
		case FamilyReplica:
			names = append(names, "the replica batch")
		case FamilyFoundations:
			names = append(names, "the foundations batch")
		}
	}
	return fmt.Sprintf("%s theme — casting from %s (%d templates)", skin, strings.Join(names, " and "), len(pool))
}

// ComboCatalogForPrompt renders the pool as the director sees it: grouped by
// category, and carrying each template's bio rather than its gallery copy.
//
// This is the difference between the old caster and this one. The gallery copy
// answers "is this the look I want?", which is a person's question; the bio
// answers "can I actually fill this?", which is the director's. Both are shown —
// the description says what lands on screen and the requirement says what it
// costs — but the requirement is the line that decides, so it is the line that
// is impossible to skim past.
//
// Built from the live registry, so a template added to the catalog is castable
// the moment it registers and a bio written for it is what the director reads.
func ComboCatalogForPrompt(skin string) string {
	pool := ComboPool(skin)
	inPool := make(map[string]bool, len(pool))
	for _, t := range pool {
		inPool[t.Name] = true
	}

	var sb strings.Builder
	for _, g := range SnippetTemplatesByCategory() {
		var shown []*SnippetTemplate
		for _, t := range g.Templates {
			if inPool[t.Name] {
				shown = append(shown, t)
			}
		}
		if len(shown) == 0 {
			continue
		}
		fmt.Fprintf(&sb, "\n%s — %s\n", g.Title, g.Blurb)
		for _, t := range shown {
			fmt.Fprintf(&sb, "\n  %s — %s. %s\n", t.Name, t.Title, t.Description)
			fmt.Fprintf(&sb, "    NEEDS: %s\n", t.Bio.Needs)
			if t.Bio.Avoid != "" {
				fmt.Fprintf(&sb, "    NOT FOR: %s\n", t.Bio.Avoid)
			}
			roles := append([]string(nil), t.Bio.Roles...)
			sort.Strings(roles)
			fmt.Fprintf(&sb, "    CAN BE: %s\n", strings.Join(roles, ", "))
		}
	}
	return sb.String()
}

// ComboFigureTemplates lists the pool's templates that cannot be planned
// without real numbers, for the prompt's do-not-cast-over-a-gap warning.
//
// Derived from the bios rather than written into the prompt, which is the whole
// point of the Figures flag: the prompt used to name four templates by hand and
// was silent about every one added since.
func ComboFigureTemplates(skin string) []string {
	var out []string
	for _, t := range ComboPool(skin) {
		if t.Bio.Figures {
			out = append(out, t.Name)
		}
	}
	return out
}

// ComboRoleTemplates lists the pool's templates that can carry an arc role.
//
// Handed to the director for the two roles that are structurally scarce. The
// arc check rejects a piece that does not open on a hook or close on a payoff,
// and a director that cannot see which looks can do those jobs satisfies the
// rule by mislabelling — an `anatomy` declared a hook still opens by taking a
// URL apart, and the check passes on a clip that puts nothing at stake.
func ComboRoleTemplates(skin, role string) []string {
	var out []string
	for _, t := range TemplatesForRole(role, ComboPool(skin)) {
		out = append(out, t.Name)
	}
	return out
}
