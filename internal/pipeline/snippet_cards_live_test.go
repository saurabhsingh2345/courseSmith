package pipeline

import (
	"context"
	"os"
	"strings"
	"testing"
)

// The card marks come from a service nobody here controls, and every other test
// around them asserts the parsing against bytes this repository wrote. That
// cannot notice the CDN changing what it serves, dropping a brand, or refusing
// a client — all three of which have already happened once each in the course of
// building this: OpenAI's marks were removed from the set on a trademark
// request, and the service answers 403 to a request with no user agent.
//
// So this one goes out to the real thing. It is not part of the default run
// because a test that fails when a laptop is offline is a test that gets
// disabled; set COURSESMITH_NET=1 to include it.
//
// It costs about a second.
func TestCardArtFetchesRealMarks(t *testing.T) {
	if os.Getenv("COURSESMITH_NET") == "" {
		t.Skip("set COURSESMITH_NET=1 to fetch real brand marks")
	}
	spec := &CardsSpec{Items: []Card{
		{Title: "Claude", Brand: "claude", Site: "anthropic.com", Icon: "brain"},
		{Title: "Gemini", Brand: "googlegemini", Site: "gemini.google.com", Icon: "sparkles"},
		// The slug that is gone. It has to fall through to the favicon rather
		// than to nothing, because this is the case the fallback chain exists for
		// and it is not hypothetical.
		{Title: "ChatGPT", Brand: "openai", Site: "openai.com", Icon: "message"},
		// Nothing to look up at all.
		{Title: "Some idea", Icon: "idea"},
	}}
	resolveCardArt(context.Background(), &Env{Out: os.Stdout}, spec)

	for i := range 2 {
		it := spec.Items[i]
		if it.Mark == "" {
			t.Errorf("%s came back with no brand mark (from %q)", it.Title, it.MarkFrom)
			continue
		}
		// Path data, not a document: if the service starts serving something
		// else, this is where it shows.
		if !strings.HasPrefix(strings.TrimSpace(it.Mark), "M") && !strings.HasPrefix(strings.TrimSpace(it.Mark), "m") {
			t.Errorf("%s's mark does not start with a moveto: %.40q", it.Title, it.Mark)
		}
		if !strings.HasPrefix(it.MarkFrom, "simpleicons:") {
			t.Errorf("%s's provenance is %q", it.Title, it.MarkFrom)
		}
	}
	if got := spec.Items[2]; !strings.HasPrefix(got.Image, "data:image/") {
		t.Errorf("the dropped brand did not fall through to a favicon: mark=%q image=%.30q from=%q", got.Mark, got.Image, got.MarkFrom)
	}
	if got := spec.Items[3]; got.Mark != "" || got.Image != "" {
		t.Errorf("a card with nothing to look up came back with art from %q", got.MarkFrom)
	}
}
