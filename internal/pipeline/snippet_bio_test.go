package pipeline

import (
	"strings"
	"testing"
)

// The registration guard catches a template with no bio and panics at init, so
// the whole package failing to load is that direction's test. These are the
// checks it structurally cannot make.

// A bio for a template that no longer exists sits in the table looking
// maintained, and the entry it should have replaced is the one that is missing.
func TestNoOrphanBios(t *testing.T) {
	if orphans := OrphanBios(); len(orphans) > 0 {
		t.Errorf("templateBios has entries for templates that do not exist: %s\n"+
			"A renamed template leaves its old bio behind, and the new name then registers with none",
			strings.Join(orphans, ", "))
	}
}

// The arc rules are unsatisfiable unless the catalog can actually staff them.
//
// The piece must open on a hook and close on a payoff, and both are checked
// against the cast. If no offered template declared it could hook, every cast
// would fail validation for a reason no prompt could fix — so this asserts the
// catalog contains a way to obey its own rules, per family pool as well as
// overall, since a pool is what a director actually chooses from.
func TestEveryPoolCanStaffTheArc(t *testing.T) {
	for _, skin := range ComboSkins() {
		pool := ComboPool(skin)
		if len(pool) == 0 {
			t.Errorf("skin %q offers no templates at all", skin)
			continue
		}
		for _, role := range []string{RoleHook, RoleDevelop, RolePayoff} {
			if got := TemplatesForRole(role, pool); len(got) == 0 {
				t.Errorf("skin %q has %d templates and none can carry the %q role — every cast in this theme would fail the arc check",
					skin, len(pool), role)
			}
		}
	}
}

// A bio written as a description rather than a requirement is the failure the
// table's header is about, and length is the only mechanical proxy for it: "some
// data" is short because it is vague, and vagueness is what makes a director
// pick the wrong look.
func TestBiosAreSpecific(t *testing.T) {
	for _, tpl := range SnippetTemplateList() {
		if n := len(strings.Fields(tpl.Bio.Needs)); n < 6 {
			t.Errorf("%s: Needs is %d words (%q) — too vague to check a pick against. Name the actual material",
				tpl.Name, n, tpl.Bio.Needs)
		}
	}
}

// A template that can only develop cannot open or close a piece, which is fine.
// A template that can do NOTHING is a registration mistake the init guard would
// have caught; this catches the subtler one, where a Figures template is offered
// as a hook — the opening segment is cast before any facts are confirmed, and a
// chart is the worst place to discover the numbers do not exist.
func TestFigureTemplatesDeclareTheirRisk(t *testing.T) {
	for _, tpl := range SnippetTemplateList() {
		if tpl.Bio.Figures && strings.TrimSpace(tpl.Bio.Avoid) == "" {
			t.Errorf("%s needs real figures but names no subject to avoid — the trap is exactly what a director has to be told", tpl.Name)
		}
	}
}
