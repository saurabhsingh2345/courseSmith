package pipeline

import (
	"strings"
	"testing"
)

// Every template must be findable. A catalog that can grow an uncategorised
// entry does, and it lands in whatever bucket the UI keeps for leftovers.
func TestEveryTemplateHasAKnownCategory(t *testing.T) {
	for _, tpl := range SnippetTemplateList() {
		if tpl.Category == "" {
			t.Errorf("template %q has no category", tpl.Name)
			continue
		}
		if _, ok := LookupSnippetCategory(tpl.Category); !ok {
			t.Errorf("template %q has unknown category %q", tpl.Name, tpl.Category)
		}
	}
}

// registerSnippetTemplate is the gate: a template with no category cannot make
// it into the catalog at init, so the check above can never start failing
// silently in a build somebody shipped.
func TestRegisterRejectsUncategorisedTemplate(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("registering a template with no category did not panic")
		}
		if msg, _ := r.(string); !strings.Contains(msg, "no category") {
			t.Errorf("panic did not explain the problem: %v", r)
		}
	}()
	registerSnippetTemplate(&SnippetTemplate{Name: "zz-uncategorised-probe"})
}

func TestRegisterRejectsUnknownCategory(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("registering a template with an invented category did not panic")
		}
	}()
	registerSnippetTemplate(&SnippetTemplate{Name: "zz-bad-category-probe", Category: "vibes"})
}

// The grouping is what both the CLI and the gallery render, so it has to
// account for the whole catalog exactly once.
func TestGroupingCoversEveryTemplateOnce(t *testing.T) {
	seen := map[string]int{}
	for _, g := range SnippetTemplatesByCategory() {
		if len(g.Templates) == 0 {
			t.Errorf("category %q is shown with nothing in it", g.Name)
		}
		if g.Title == "" || g.Blurb == "" {
			t.Errorf("category %q has no heading or no blurb — the gallery needs both", g.Name)
		}
		for _, tpl := range g.Templates {
			seen[tpl.Name]++
			if tpl.Category != g.Name {
				t.Errorf("template %q is grouped under %q", tpl.Name, g.Name)
			}
		}
	}
	for _, tpl := range SnippetTemplateList() {
		if seen[tpl.Name] != 1 {
			t.Errorf("template %q appears %d times in the grouping, want exactly 1", tpl.Name, seen[tpl.Name])
		}
	}
}

// Templates are sorted inside each group, so the gallery order is stable
// between renders and between the CLI and the studio.
func TestGroupsAreSortedByName(t *testing.T) {
	for _, g := range SnippetTemplatesByCategory() {
		for i := 1; i < len(g.Templates); i++ {
			if g.Templates[i-1].Name > g.Templates[i].Name {
				t.Errorf("category %q is out of order: %q before %q",
					g.Name, g.Templates[i-1].Name, g.Templates[i].Name)
			}
		}
	}
}

// No category should swallow the catalog. A group holding half of everything is
// a group nobody can scan, which is the problem the taxonomy exists to solve.
func TestNoCategoryIsOversized(t *testing.T) {
	total := len(SnippetTemplateNames())
	for _, g := range SnippetTemplatesByCategory() {
		if len(g.Templates)*3 > total {
			t.Errorf("category %q holds %d of %d templates — that is a wall, not a group; split it",
				g.Name, len(g.Templates), total)
		}
	}
}

// Since is a release tag, not a free-text field: a typo would put a template in
// a chip nobody recognises.
func TestSinceIsAKnownRelease(t *testing.T) {
	for _, tpl := range SnippetTemplateList() {
		switch tpl.Since {
		case SinceCore, SinceV1, SinceV2:
		default:
			t.Errorf("template %q has release tag %q, want one of %q, %q, %q",
				tpl.Name, tpl.Since, SinceCore, SinceV1, SinceV2)
		}
	}
}

// The v1 batch is the ten built against the reference clips. Asserted by name
// so a template cannot quietly join or leave it.
func TestV1BatchIsTheTenReferenceTemplates(t *testing.T) {
	want := map[string]bool{
		"metric": true, "gauge": true, "verdict": true, "decision": true, "myth": true,
		"analogy": true, "rundown": true, "trace": true, "costing": true, "constellation": true,
	}
	got := map[string]bool{}
	for _, tpl := range SnippetTemplateList() {
		if tpl.Since == SinceV1 {
			got[tpl.Name] = true
		}
	}
	for name := range want {
		if !got[name] {
			t.Errorf("%q is not tagged %s", name, SinceV1)
		}
	}
	for name := range got {
		if !want[name] {
			t.Errorf("%q is tagged %s but is not part of that batch", name, SinceV1)
		}
	}
}

// The v2 batch is the three written to carry a whole course rather than to
// answer one question. Asserted by name for the same reason v1 is.
func TestV2BatchIsTheThreeCourseTemplates(t *testing.T) {
	want := map[string]bool{"chapter": true, "cycle": true, "scale": true}
	got := map[string]bool{}
	for _, tpl := range SnippetTemplateList() {
		if tpl.Since == SinceV2 {
			got[tpl.Name] = true
		}
	}
	for name := range want {
		if !got[name] {
			t.Errorf("%q is not tagged %s", name, SinceV2)
		}
	}
	for name := range got {
		if !want[name] {
			t.Errorf("%q is tagged %s but is not part of that batch", name, SinceV2)
		}
	}
}
