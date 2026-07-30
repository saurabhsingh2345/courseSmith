package pipeline

import (
	"strings"
	"testing"
)

// Shelving has to hold on two sides at once, and they pull in opposite
// directions: nothing may *offer* a shelved template, and everything must still
// *render* one. A change that only got the first half right would look correct
// in the gallery and break every reel already on disk that names `story`.

// shelvedTemplate returns the name of some shelved template, or skips. Written
// this way rather than hard-coding `story` so these tests keep testing the
// mechanism after the current two are un-shelved.
func shelvedTemplate(t *testing.T) string {
	t.Helper()
	for _, name := range SnippetTemplateNames() {
		if SnippetTemplates[name].Shelved {
			return name
		}
	}
	t.Skip("no shelved templates in the catalog")
	return ""
}

// The caster reads its catalog off the live registry, which is what makes a new
// template castable the moment it registers — and is exactly why shelving has to
// reach it. A look left in the caster's catalog gets picked.
func TestShelvedTemplatesAreNotCastable(t *testing.T) {
	name := shelvedTemplate(t)
	catalog := SnippetCatalogForPrompt()
	// Matched with the surrounding spaces of the catalog's own "  %-14s %s"
	// column, so a name that happens to be a substring of another template's
	// prose ("cast" inside "screencast") does not read as a false positive.
	if strings.Contains(catalog, "  "+name+" ") {
		t.Errorf("shelved template %q is in the catalog the caster picks from", name)
	}
	// Positive control. The match above depends on the catalog's column format,
	// so without this a formatting change would turn this test green forever
	// instead of failing.
	if !strings.Contains(catalog, "  myth ") {
		t.Error("`myth` is not matchable in the catalog, so the absence check above proves nothing")
	}
}

func TestShelvedTemplatesAreNotOnOffer(t *testing.T) {
	name := shelvedTemplate(t)
	for _, tpl := range SnippetTemplateList() {
		if tpl.Name == name {
			t.Errorf("shelved template %q is on offer via SnippetTemplateList", name)
		}
	}
	// Still registered, though — the whole point.
	if _, ok := SnippetTemplates[name]; !ok {
		t.Errorf("shelved template %q was removed from the registry, not shelved", name)
	}
}

// A reel authored before the template was shelved must still plan and render.
// This is the half that a delete-the-template change would have broken, and it
// is not hypothetical: three reels under .coursesmith/ name `story`.
func TestShelvedTemplateStillValidatesInAReel(t *testing.T) {
	name := shelvedTemplate(t)
	spec := ReelSpec{
		ID:    "already-on-disk",
		Title: "A reel cast before the shelf",
		Segments: []ReelSegment{
			{ID: "a", Template: name, Prompt: "the part that was cast with a character in it"},
			{ID: "b", Template: "myth", Prompt: "the belief this corrects"},
			{ID: "c", Template: "verdict", Prompt: "what to do about it"},
		},
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("a reel naming the shelved template %q no longer validates: %v", name, err)
	}
}

// And the error a typo produces should still name it, because naming it
// explicitly is still a working thing to do.
func TestShelvedTemplateIsStillNamedInTheHint(t *testing.T) {
	name := shelvedTemplate(t)
	spec := ReelSpec{
		ID: "typo",
		Segments: []ReelSegment{
			{ID: "a", Template: name + "x", Prompt: "misspelled"},
			{ID: "b", Template: "myth", Prompt: "the belief this corrects"},
			{ID: "c", Template: "verdict", Prompt: "what to do about it"},
		},
	}
	err := spec.Validate()
	if err == nil {
		t.Fatal("an unknown template was accepted")
	}
	if !strings.Contains(err.Error(), name) {
		t.Errorf("the hint omits the shelved template %q, so somebody who meant to name it cannot find it: %v", name, err)
	}
}
