package pipeline

import (
	"strings"
	"testing"
)

// A reel's chapter list used to read "Myth 1  Everyone Says": the template name,
// the caster's ordinal, and a double space where the "--" in the section id was.
// Nothing was wrong with the heading — it was known when the section was built
// and dropped one field short of the chapters stage, which then recovered a title
// by matching slugs against lesson.md and, finding no match for a composed id,
// humanized the id itself.
func TestReelChaptersUseTheBeatHeading(t *testing.T) {
	plan := &ReelPlan{
		Title: "Welcome to the era of no-code",
		Segments: []ReelPlanSegment{{
			ID:       "myth-1",
			Template: "myth",
			Plan: &SnippetPlan{Beats: []SnippetBeat{
				{ID: "everyone-says", Heading: "What everyone says", Narration: "Many people believe you need years of experience."},
				{ID: "not-quite", Heading: "Not quite", Narration: "You can build software with very little of it."},
			}},
		}},
	}
	script := plan.Script(174)
	titles := sectionTitles("", script)

	for _, sec := range script.Sections {
		got := titles[sec.ID]
		if strings.Contains(got, "  ") {
			t.Errorf("chapter %q has a double space, so the id is still being humanized", got)
		}
		if strings.Contains(strings.ToLower(got), "myth") {
			t.Errorf("chapter %q leaks the template name", got)
		}
		if strings.ContainsAny(got, "0123456789") {
			t.Errorf("chapter %q leaks the cast ordinal", got)
		}
	}
	if titles["myth-1--everyone-says"] != "What everyone says" {
		t.Errorf("first chapter is %q, want %q", titles["myth-1--everyone-says"], "What everyone says")
	}
	if titles["myth-1--not-quite"] != "Not quite" {
		t.Errorf("second chapter is %q, want %q", titles["myth-1--not-quite"], "Not quite")
	}
}

func TestSnippetChaptersUseTheBeatHeading(t *testing.T) {
	plan := &SnippetPlan{
		Title: "What a 70B model costs",
		Beats: []SnippetBeat{
			// A heading whose slug does not round-trip: the ampersand and the colon
			// vanish, so matching against lesson.md would have degraded this one.
			{ID: "cards-hosting", Heading: "Cards & hosting: the real bill", Narration: "The sticker price is not the price you pay."},
		},
	}
	titles := sectionTitles("", plan.Script(174))
	if got := titles["cards-hosting"]; got != "Cards & hosting: the real bill" {
		t.Errorf("chapter is %q, want the heading verbatim", got)
	}
}

// A hand-written lesson has no Section.Title, and must keep working exactly as it
// did — its headings live in lesson.md and slug matching is the only route.
func TestHandWrittenLessonStillMatchesItsHeadings(t *testing.T) {
	body := "## What is a variable?\n\nSome prose.\n\n## Naming variables\n\nMore prose.\n"
	script := &Script{Sections: []Section{
		{ID: "what-is-a-variable"},
		{ID: "naming-variables"},
	}}
	titles := sectionTitles(body, script)
	if got := titles["what-is-a-variable"]; got != "What is a variable?" {
		t.Errorf("got %q, want the lesson.md heading", got)
	}
	if got := titles["naming-variables"]; got != "Naming variables" {
		t.Errorf("got %q, want the lesson.md heading", got)
	}
}

// And a section with neither a title nor a matching heading still gets something
// readable rather than a raw slug.
func TestUnmatchedSectionFallsBackToTheHumanizedID(t *testing.T) {
	script := &Script{Sections: []Section{{ID: "some-orphan-section"}}}
	if got := sectionTitles("", script)["some-orphan-section"]; got != "Some Orphan Section" {
		t.Errorf("got %q, want a humanized id", got)
	}
}

// A title carrying stray whitespace must not reintroduce the very artefact this
// fixes.
func TestSectionTitleWhitespaceIsCollapsed(t *testing.T) {
	script := &Script{Sections: []Section{{ID: "x", Title: "Two  spaces   here"}}}
	if got := sectionTitles("", script)["x"]; got != "Two spaces here" {
		t.Errorf("got %q, want collapsed whitespace", got)
	}
}
