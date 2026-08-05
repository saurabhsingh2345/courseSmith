package pipeline

// Categories: how somebody finds the template they want.
//
// The catalog passed twenty and a flat alphabetical list stopped working. That
// is not a cosmetic problem — a creator opening the gallery is not thinking
// "I want the gauge template", they are thinking "I need to show whether this
// fits", and an A-to-Z run from `analogy` to `workspace` gives them no way to
// get from the second thought to the first.
//
// So the grouping is by the JOB, not by the mechanism. It would have been
// easier to sort by what is on screen — charts here, editors there — and it
// would have been useless, because the thing a creator knows when they arrive
// is what they are trying to say. `gauge` and `metric` both draw numbers and
// sit together; `trace` and `flow` both draw boxes and edges and sit together
// for a completely different reason, which is that both answer "how does this
// work".
//
// Every template must declare one, and registerSnippetTemplate panics on a
// missing or unknown category. A catalog that could grow an uncategorised entry
// would grow one, and it would land in whatever bucket the UI used for
// leftovers — which is where templates go to be never used again.

import "fmt"

// The category names. Kept as constants so a template's declaration is checked
// at compile time rather than spelled as a string in twenty-nine files.
const (
	// CatNumbers is for clips whose subject is a quantity.
	CatNumbers = "numbers"
	// CatConcepts is for clips that install or correct an idea.
	CatConcepts = "concepts"
	// CatSystems is for clips that show how something works or is built.
	CatSystems = "systems"
	// CatCode is for clips whose subject is on a screen — an editor, a canvas,
	// a page being assembled.
	CatCode = "code"
	// CatDecisions is for clips that help somebody choose.
	CatDecisions = "decisions"
	// CatPresenting is for clips whose shape is a performance or a pace.
	CatPresenting = "presenting"
)

// SnippetCategory is one group in the gallery.
type SnippetCategory struct {
	// Name is the id used in the API and on the CLI.
	Name string
	// Title is the heading shown above the group.
	Title string
	// Blurb is the one line under the heading. It answers "is what I want in
	// here?" — so it is written as the question a creator arrives with, not as
	// a description of the templates underneath.
	Blurb string
}

// snippetCategories is the catalog's groups, in the order they are shown.
//
// The order is deliberate and is not alphabetical: it runs from the most
// concrete thing a creator might want to say to the most stylistic. Numbers
// first because a quantity is the most common reason a short clip exists at
// all, and presenting last because those are choices about delivery rather than
// about content.
var snippetCategories = []SnippetCategory{
	{CatNumbers, "Numbers & scale", "How big, how much, how long — and whether it fits."},
	{CatConcepts, "Ideas & mental models", "Explaining something, or replacing what someone already believes."},
	{CatSystems, "Systems & process", "How it works, how it is built, what happens in what order."},
	{CatCode, "Code & screens", "Anything whose subject is on a screen — an editor, a canvas, a page."},
	{CatDecisions, "Choices & verdicts", "Weighing options and telling somebody what to do."},
	{CatPresenting, "Presenting & pacing", "The shape of the delivery rather than the content."},
}

// SnippetCategories returns the groups in display order.
func SnippetCategories() []SnippetCategory {
	out := make([]SnippetCategory, len(snippetCategories))
	copy(out, snippetCategories)
	return out
}

// snippetCategoryByName indexes the groups for validation.
var snippetCategoryByName = func() map[string]SnippetCategory {
	m := make(map[string]SnippetCategory, len(snippetCategories))
	for _, c := range snippetCategories {
		m[c.Name] = c
	}
	return m
}()

// LookupSnippetCategory returns a category by name.
func LookupSnippetCategory(name string) (SnippetCategory, bool) {
	c, ok := snippetCategoryByName[name]
	return c, ok
}

// Release tags. A template's Since is the catalog release it arrived in, which
// is a fact rather than a status — unlike a "new" flag, it does not rot the
// moment the next batch lands.
const (
	// SinceCore is the empty tag carried by the templates that predate
	// versioning. They are the catalog's original set.
	SinceCore = ""
	// SinceV1 is the batch built to match the reference explainer clips.
	SinceV1 = "v1"
	// SinceV2 is the batch written to carry a whole course rather than to
	// answer one question: the break between sections, the loop that comes back
	// round, and the ladder of magnitudes a chart cannot draw.
	SinceV2 = "v2"
	// SinceV3 is the capture batch: templates whose frame is a real recording
	// rather than anything drawn. They arrived with the no-code surface, where
	// every segment has to stand on evidence.
	SinceV3 = "v3"
	// SinceV4 is the replica batch, built frame-by-frame against four reference
	// videos rather than from a description of them. See FamilyReplica.
	SinceV4 = "v4"
	// SinceV5 is the course-scaffolding batch: the five templates that structure
	// a course rather than answer a question inside one.
	//
	// The catalog reached forty-four templates that could each explain
	// something, and still had no way to say what a lesson is FOR, what it
	// assumes, what the last one established, where people go wrong at this
	// exact step, or whether the viewer can now do the thing. Every one of the
	// forty-four is an answer; none of them is the scaffolding a course hangs
	// answers on, which is why a run of them reads as a playlist rather than a
	// course.
	//
	// What makes these five a batch rather than five more looks: each one's
	// validator checks a relationship to the REST of the course — an outcome
	// that must be observable, a prerequisite that must resolve to something
	// actually taught, a recap claim that must trace to a prior lesson. They are
	// the first templates that cannot be validated by looking at one clip alone.
	SinceV5 = "v5"
)

// Template families. A family is which surface offers a template, and it exists
// because the catalog now has two audiences rather than one.
//
// This is not a second category axis. A category says what a template is *for*
// ("systems & process"), and the replica batch spans several of them. A family
// says which front door it appears behind, so the snippets gallery keeps
// offering the catalog it has always offered while a second page can offer a set
// with its own house rules.
const (
	// FamilyCore is the empty family: every template that predates the split,
	// offered on the snippets page. Empty rather than "core" so no registration
	// site has to be touched and no persisted plan changes shape.
	FamilyCore = ""
	// FamilyReplica is the batch cut to the reference videos' visual grammar —
	// a threshold bar racing a requirement line, a wrong number struck through
	// and replaced. They are offered on their own page because they assume the
	// broadcast skin's air and chrome, and dropped into a default-skin course
	// they would read as a different production.
	FamilyReplica = "replica"
)

// SnippetCategoryGroup is one category with the templates in it.
type SnippetCategoryGroup struct {
	SnippetCategory
	Templates []*SnippetTemplate
}

// SnippetTemplatesByCategory returns the catalog grouped for display: the
// groups in their declared order, and the templates inside each sorted by name.
//
// Empty groups are dropped rather than shown as headings with nothing under
// them, so a category can be declared here before anything uses it.
func SnippetTemplatesByCategory() []SnippetCategoryGroup {
	byCat := map[string][]*SnippetTemplate{}
	for _, t := range SnippetTemplateList() {
		byCat[t.Category] = append(byCat[t.Category], t)
	}
	out := make([]SnippetCategoryGroup, 0, len(snippetCategories))
	for _, c := range snippetCategories {
		if len(byCat[c.Name]) == 0 {
			continue
		}
		// SnippetTemplateList is already name-sorted, so each bucket is too.
		out = append(out, SnippetCategoryGroup{SnippetCategory: c, Templates: byCat[c.Name]})
	}
	return out
}

// checkTemplateCategory is the guard registerSnippetTemplate runs. It panics
// rather than returning an error because a miscategorised template is a
// programming mistake caught at init, and the alternative — degrading to a
// leftovers bucket — is how a catalog quietly accumulates entries nobody finds.
func checkTemplateCategory(t *SnippetTemplate) {
	if t.Category == "" {
		panic(fmt.Sprintf("snippet template %q has no category — pick one of %v so it can be found in the gallery",
			t.Name, snippetCategoryNames()))
	}
	if _, ok := snippetCategoryByName[t.Category]; !ok {
		panic(fmt.Sprintf("snippet template %q has unknown category %q (want one of %v)",
			t.Name, t.Category, snippetCategoryNames()))
	}
}

func snippetCategoryNames() []string {
	out := make([]string, 0, len(snippetCategories))
	for _, c := range snippetCategories {
		out = append(out, c.Name)
	}
	return out
}
