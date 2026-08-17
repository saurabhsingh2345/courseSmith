package pipeline

import (
	"context"
	"strings"
	"testing"
)

const cardNarration = "Every card is up from the first frame, and the light moves across the row while the arrangement holds still."

func cardsPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "cards",
		Title:    "Three assistants, one coding job",
		Cards: &CardsSpec{
			Relation: "versus",
			Items: []Card{
				{Title: "Claude", Note: "Holds a long argument about a codebase without losing the thread", Role: "quantity", Brand: "claude", Site: "anthropic.com", Icon: "brain"},
				{Title: "Gemini", Note: "Reads a whole repository at once and answers across all of it", Role: "rival", Brand: "googlegemini", Site: "gemini.google.com", Icon: "sparkles"},
				{Title: "ChatGPT", Note: "Fastest to a working answer, and the one everybody already has open", Role: "neutral", Site: "openai.com", Icon: "message"},
			},
			Closer: "Claude for long refactors, Gemini for whole repos, ChatGPT for a quick answer",
		},
		Beats: []SnippetBeat{
			{ID: "row", Heading: "Three assistants", Narration: cardNarration, Cards: &CardsBeat{Show: "row"}},
			{ID: "claude", Heading: "The long thread", Narration: cardNarration, Cards: &CardsBeat{Show: "card", At: 0}},
			{ID: "gemini", Heading: "The whole repo", Narration: cardNarration, Cards: &CardsBeat{Show: "card", At: 1}},
			{ID: "chatgpt", Heading: "The quick answer", Narration: cardNarration, Cards: &CardsBeat{Show: "card", At: 2}},
			{ID: "pick", Heading: "Which one", Narration: cardNarration, Cards: &CardsBeat{Show: "all"}},
		},
	}
	// Five beats at this template's own per-beat budget, so the shared beat-shape
	// check scores the plan against the range the prompt would have quoted.
	p.targetWords = 5 * 26
	return p
}

func TestCardsPlanAccepted(t *testing.T) {
	if err := validateCardsPlan(cardsPlan()); err != nil {
		t.Fatalf("a well-formed row was rejected: %v", err)
	}
}

// The rule the template turns on: a card that is a logo and a name is a sticker.
func TestCardsRejectsACardWithNoLine(t *testing.T) {
	p := cardsPlan()
	p.Cards.Items[1].Note = ""
	err := validateCardsPlan(p)
	if err == nil {
		t.Fatal("a card with no line under it was accepted")
	}
	if !strings.Contains(err.Error(), "Gemini") {
		t.Fatalf("the error does not name the card that is missing its line: %v", err)
	}
}

// A contest that ends having named the contenders is a scoreboard with the score
// unread. Borrowed from versus, and it fails here the same way.
func TestCardsRequiresACloserOnAContest(t *testing.T) {
	p := cardsPlan()
	p.Cards.Closer = ""
	if err := validateCardsPlan(p); err == nil {
		t.Fatal("a versus row that closes on nothing was accepted")
	}

	// The same row with nothing in the gaps makes no claim to settle, so the
	// closer is genuinely optional there.
	p.Cards.Relation = "none"
	if err := validateCardsPlan(p); err != nil {
		t.Fatalf("a plain row with no closer was rejected: %v", err)
	}
}

func TestCardsRejectsACloserThatIsJustOneName(t *testing.T) {
	p := cardsPlan()
	p.Cards.Closer = "Claude."
	err := validateCardsPlan(p)
	if err == nil {
		t.Fatal("a closer that is only one card's name was accepted")
	}
	if !strings.Contains(err.Error(), "Claude") {
		t.Fatalf("the error does not quote the closer back: %v", err)
	}
}

func TestCardsRejectsTwoCardsForTheSameThing(t *testing.T) {
	p := cardsPlan()
	p.Cards.Items[2].Title = "claude"
	if err := validateCardsPlan(p); err == nil {
		t.Fatal("a row showing one thing twice was accepted")
	}
}

// The row is on screen from the first frame — that is what this format has over
// a list — so a card nobody speaks about sits lit and unaccounted for.
func TestCardsRequiresEveryCardToGetABeat(t *testing.T) {
	p := cardsPlan()
	p.Beats = append(p.Beats[:3], p.Beats[4])
	err := validateCardsPlan(p)
	if err == nil {
		t.Fatal("a row with a card nobody spoke about was accepted")
	}
	if !strings.Contains(err.Error(), "ChatGPT") {
		t.Fatalf("the error does not name the card that was skipped: %v", err)
	}
}

// A plan with more card beats than cards. Found by the studio rather than by
// this file: the ordering rule below names "the next card due", and once every
// card has been lit there is no next one — so reading it panicked the whole
// server mid-run rather than failing the plan.
func TestCardsRejectsMoreCardBeatsThanCards(t *testing.T) {
	p := cardsPlan()
	p.Cards.Items = p.Cards.Items[:2]
	p.Cards.Closer = "Claude for long refactors, Gemini when the context is a whole repo"
	// Three card beats over a two-card row: the third has nothing left to light.
	p.Beats[3].Cards = &CardsBeat{Show: "card", At: 1}
	err := validateCardsPlan(p)
	if err == nil {
		t.Fatal("a clip with more card beats than cards was accepted")
	}
	if !strings.Contains(err.Error(), "all 2 cards") {
		t.Fatalf("the error does not say the row was already finished: %v", err)
	}
}

func TestCardsRequiresTheCardsInOrder(t *testing.T) {
	p := cardsPlan()
	p.Beats[1].Cards.At, p.Beats[2].Cards.At = 1, 0
	if err := validateCardsPlan(p); err == nil {
		t.Fatal("a row read out of order was accepted")
	}
}

func TestCardsRequiresTheRowFirstAndAllLast(t *testing.T) {
	p := cardsPlan()
	p.Beats[0].Cards = &CardsBeat{Show: "card", At: 0}
	if err := validateCardsPlan(p); err == nil {
		t.Fatal("a clip that opens on one card rather than the row was accepted")
	}

	p = cardsPlan()
	p.Beats[len(p.Beats)-1].Cards = &CardsBeat{Show: "card", At: 2}
	if err := validateCardsPlan(p); err == nil {
		t.Fatal("a clip that ends lit on one card was accepted")
	}
}

// The beat budget and the card count are the same arithmetic, and the error has
// to say so — "too many cards" without the sum leaves the model choosing between
// cutting cards and cutting beats.
func TestCardsRejectsMoreCardsThanTheRuntimeFunds(t *testing.T) {
	p := cardsPlan()
	// A twenty-second clip: four beats at most, so two cards, so this row of
	// three cannot be covered however it is written. The beats are cut to four
	// as well, so the plan is inside the shared beat range and this is the only
	// rule left to catch it.
	p.targetWords = 50
	p.Beats = append(p.Beats[:3], p.Beats[4])
	err := validateCardsPlan(p)
	if err == nil {
		t.Fatal("a row with more cards than the runtime can light was accepted")
	}
	if !strings.Contains(err.Error(), "3 cards") {
		t.Fatalf("the error does not count the cards: %v", err)
	}
}

func TestCardsRejectsAForeignBeatField(t *testing.T) {
	p := cardsPlan()
	p.Beats[1].Versus = &VersusBeat{Show: "row"}
	if err := validateCardsPlan(p); err == nil {
		t.Fatal("a versus payload on a cards beat was accepted")
	}
}

// Normalization repairs what only has one sensible repair, so the correction
// rounds are spent on what only the model can fix.
func TestCardsNormalizeFillsTheDefaults(t *testing.T) {
	p := cardsPlan()
	p.Cards.Relation = "side by side"
	p.Cards.Items[0].Icon = "not-an-icon"
	p.Cards.Items[0].Role = "invented"
	p.Cards.Items[1].Brand = "Google Gemini"
	p.Cards.Items[1].Site = "https://www.gemini.google.com/app"
	p.Beats[1].Cards.At = 99
	normalizeCardsPlan(p)

	if got := p.Cards.Relation; got != "none" {
		t.Errorf("an invented relation normalized to %q, want none — a drawn connector is a claim the narration never made", got)
	}
	if got := p.Cards.Items[0].Icon; got != "box" {
		t.Errorf("an unknown glyph normalized to %q, want box", got)
	}
	if got := p.Cards.Items[0].Role; got != "neutral" {
		t.Errorf("an unknown role normalized to %q, want neutral", got)
	}
	if got := p.Cards.Items[1].Brand; got != "googlegemini" {
		t.Errorf("the brand slug normalized to %q, want googlegemini", got)
	}
	if got := p.Cards.Items[1].Site; got != "gemini.google.com" {
		t.Errorf("the site normalized to %q, want gemini.google.com", got)
	}
	if got := p.Beats[1].Cards.At; got != 2 {
		t.Errorf("a beat pointing past the row normalized to %d, want the last card", got)
	}
}

func TestCardsScenesCarryTheRowAndTheSteps(t *testing.T) {
	p := cardsPlan()
	scenes, err := dryRunScenes(t, p)
	if err != nil {
		t.Fatalf("laying the clip out: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("got %d scenes, want 1 — the row persists for the whole clip", len(scenes))
	}
	props := scenes[0].Props
	if props["relation"] != "versus" {
		t.Errorf("the scene does not carry the relation: %v", props["relation"])
	}
	items, _ := props["items"].([]map[string]any)
	if len(items) != 3 {
		t.Fatalf("the scene carries %d cards, want 3", len(items))
	}
	// Absent art is an absent key rather than an empty string: the component
	// draws whatever is present, so "" would be a blank tile where the fallback
	// glyph belongs.
	if _, set := items[2]["mark"]; set {
		t.Error("a card with no fetched mark still carries a mark key")
	}
	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != len(p.Beats) {
		t.Fatalf("got %d steps for %d beats", len(steps), len(p.Beats))
	}
	if lit, _ := steps[0]["lit"].([]int); len(lit) != 3 {
		t.Errorf("the opening step lights %v, want every card — the row is up from the first frame", steps[0]["lit"])
	}
	if lit, _ := steps[2]["lit"].([]int); len(lit) != 1 || lit[0] != 1 {
		t.Errorf("a card step lights %v, want only card 1", steps[2]["lit"])
	}
	if lit, _ := steps[len(steps)-1]["lit"].([]int); len(lit) != 3 {
		t.Errorf("the closing step lights %v, want every card", steps[len(steps)-1]["lit"])
	}
}

// dryRunScenes lays a plan out against estimated timings, which is enough to
// answer "will the render stage accept this?" for a template whose scene builder
// is a pure function of the plan.
func dryRunScenes(t *testing.T, p *SnippetPlan) ([]Scene, error) {
	t.Helper()
	spans := make([]SectionSpan, len(p.Beats))
	ends := make([]int, len(p.Beats))
	for i, b := range p.Beats {
		spans[i] = SectionSpan{ID: b.ID, StartMs: i * 5000, EndMs: (i + 1) * 5000}
		ends[i] = (i + 1) * 5000
	}
	return cardsScenes(SnippetSceneInput{Plan: p, Spans: spans, BeatEndMs: ends, DurationMs: len(p.Beats) * 5000})
}

// == The art ==

// Geometry only, and only from the grid the renderer draws on. A mark from a
// different viewBox would land somewhere off the card with nothing to say so.
func TestSVGPathDataTakesGeometryFromTheRightGrid(t *testing.T) {
	ok := []byte(`<svg fill="#D97757" role="img" viewBox="0 0 24 24"><title>Claude</title><path d="M1 2h3z"/></svg>`)
	if got := svgPathData(ok); got != "M1 2h3z" {
		t.Errorf("extracted %q, want the path data alone", got)
	}
	two := []byte(`<svg viewBox="0 0 24 24"><path d="M1 2h3z"/><path d="M4 5h6z"/></svg>`)
	if got := svgPathData(two); got != "M1 2h3z M4 5h6z" {
		t.Errorf("extracted %q from a two-path mark", got)
	}
	wrong := []byte(`<svg viewBox="0 0 512 512"><path d="M1 2h3z"/></svg>`)
	if got := svgPathData(wrong); got != "" {
		t.Errorf("a mark on a 512 grid was accepted as %q", got)
	}
	if got := svgPathData([]byte(`<html>not found</html>`)); got != "" {
		t.Errorf("an error page was read as a mark: %q", got)
	}
}

func TestCardsHostAndSlugNormalize(t *testing.T) {
	for in, want := range map[string]string{
		"https://www.anthropic.com/claude": "anthropic.com",
		"ANTHROPIC.COM":                    "anthropic.com",
		"gemini.google.com":                "gemini.google.com",
		"not a domain":                     "",
		"localhost":                        "",
	} {
		if got := cardsHost(in); got != want {
			t.Errorf("cardsHost(%q) = %q, want %q", in, got, want)
		}
	}
	for in, want := range map[string]string{
		"Google Gemini":      "googlegemini",
		"visual-studio-code": "visualstudiocode",
		"claude":             "claude",
	} {
		if got := cardsSlug(in); got != want {
			t.Errorf("cardsSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

// The one field where a document a model wrote decides what gets requested.
func TestCardImageURLTakesOnlyHTTPS(t *testing.T) {
	for _, raw := range []string{"http://example.com/a.png", "file:///etc/passwd", "javascript:alert(1)", "not a url at all", ""} {
		if got := (Card{ImageURL: raw}).ResolvedImageURL(); got != "" {
			t.Errorf("ResolvedImageURL(%q) = %q, want it refused", raw, got)
		}
	}
	if got := (Card{ImageURL: "https://example.com/a.png"}).ResolvedImageURL(); got != "https://example.com/a.png" {
		t.Errorf("a plain https image was refused: %q", got)
	}
}

// The whole fallback chain, in one test, because the chain IS the design: a
// template that reaches the open web has to degrade rather than fail.
func TestResolveCardArtFallsAllTheWayDown(t *testing.T) {
	served := map[string]struct {
		body []byte
		mime string
	}{
		"https://cdn.simpleicons.org/claude":                          {[]byte(`<svg viewBox="0 0 24 24"><path d="M1 2h3z"/></svg>`), "image/svg+xml"},
		"https://www.google.com/s2/favicons?sz=128&domain=openai.com": {[]byte("\x89PNG\r\n"), "image/png"},
	}
	restore := cardArtFetch
	cardArtFetch = func(_ context.Context, u string) ([]byte, string, error) {
		if r, ok := served[u]; ok {
			return r.body, r.mime, nil
		}
		return nil, "", context.DeadlineExceeded
	}
	defer func() { cardArtFetch = restore }()

	spec := &CardsSpec{Items: []Card{
		// Has a brand mark: the good case.
		{Title: "Claude", Brand: "claude", Site: "anthropic.com", Icon: "brain"},
		// Its slug 404s — OpenAI's marks are in no icon set — so the favicon.
		{Title: "ChatGPT", Brand: "openai", Site: "openai.com", Icon: "message"},
		// Nothing to fetch at all: the drawn glyph, which is always there.
		{Title: "A concept", Icon: "idea"},
	}}
	resolveCardArt(context.Background(), &Env{}, spec)

	if got := spec.Items[0].Mark; got != "M1 2h3z" {
		t.Errorf("the brand mark resolved to %q", got)
	}
	if got := spec.Items[0].MarkFrom; got != "simpleicons:claude" {
		t.Errorf("the provenance is %q, want the service that served it", got)
	}
	if !strings.HasPrefix(spec.Items[1].Image, "data:image/png;base64,") {
		t.Errorf("the favicon fallback resolved to %q", spec.Items[1].Image)
	}
	if spec.Items[1].Mark != "" {
		t.Error("a card whose slug 404d still carries a vector mark")
	}
	if spec.Items[2].Mark != "" || spec.Items[2].Image != "" || spec.Items[2].MarkFrom != "" {
		t.Error("a card with nothing to fetch came back with art")
	}
	if got := spec.Items[2].ResolvedIcon(); got != "idea" {
		t.Errorf("the floor glyph is %q, want the one the card named", got)
	}
}

// Art already on the plan is a re-plan of something that was fetched once, and
// re-fetching it would spend the network to arrive back where it started.
func TestResolveCardArtLeavesResolvedCardsAlone(t *testing.T) {
	restore := cardArtFetch
	cardArtFetch = func(context.Context, string) ([]byte, string, error) {
		t.Fatal("a card that already had its mark was fetched again")
		return nil, "", nil
	}
	defer func() { cardArtFetch = restore }()

	spec := &CardsSpec{Items: []Card{{Title: "Claude", Brand: "claude", Mark: "M1 2h3z", MarkFrom: "simpleicons:claude"}}}
	resolveCardArt(context.Background(), &Env{}, spec)
}
