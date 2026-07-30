package studio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enfec/coursesmith/internal/pipeline"
)

// The gallery groups by first-seen category and keeps no copy of the
// vocabulary, so the order templates arrive in IS the order the headings
// render in. Sending the flat name-sorted list looked fine and silently threw
// away the sequence Go declares.
func TestSnippetTemplatesArriveInCategoryOrder(t *testing.T) {
	s := &Server{}
	rec := httptest.NewRecorder()
	s.handleSnippetTemplates(rec, httptest.NewRequest(http.MethodGet, "/api/snippet-templates", nil))

	var got []SnippetTemplateInfo
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// SnippetTemplateList, not SnippetTemplateNames: the gallery offers a
	// choice, so it serves the templates on offer. Shelved ones stay registered
	// and stay renderable, and must not appear here.
	if len(got) != len(pipeline.SnippetTemplateList()) {
		t.Fatalf("got %d templates, want every template on offer (%d)", len(got), len(pipeline.SnippetTemplateList()))
	}
	for _, tpl := range got {
		if pipeline.SnippetTemplates[tpl.Name].Shelved {
			t.Errorf("shelved template %q is being offered in the gallery", tpl.Name)
		}
	}

	// The categories, in the order they first appear in the payload.
	var seen []string
	inList := map[string]bool{}
	for _, tpl := range got {
		if !inList[tpl.Category] {
			inList[tpl.Category] = true
			seen = append(seen, tpl.Category)
		}
	}
	want := []string{}
	for _, g := range pipeline.SnippetTemplatesByCategory() {
		want = append(want, g.Name)
	}
	if len(seen) != len(want) {
		t.Fatalf("payload has %d groups, want %d", len(seen), len(want))
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Errorf("group %d is %q, want %q — the declared order is not reaching the gallery", i, seen[i], want[i])
		}
	}

	// And a category's templates must arrive contiguously, or grouping by
	// first-seen would split one heading into two.
	for i := 1; i < len(got); i++ {
		if got[i].Category == got[i-1].Category {
			continue
		}
		for j := i + 1; j < len(got); j++ {
			if got[j].Category == got[i-1].Category {
				t.Fatalf("category %q is not contiguous — it resumes at index %d", got[i-1].Category, j)
			}
		}
	}

	// The display title ships with the value so no client keeps its own map.
	for _, tpl := range got {
		if tpl.CategoryTitle == "" {
			t.Errorf("template %q has no category title", tpl.Name)
		}
	}
}
