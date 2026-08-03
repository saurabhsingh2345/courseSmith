package studio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/enfec/coursesmith/internal/pipeline"
)

// A finished piece has to be reachable at the URL the API advertises.
//
// It was not: the summary built `/api/lessons/{id}/artifacts/final.mp4`, a route
// this server does not have, so the request fell through to the SPA handler and
// the studio's <video> element was handed 416 bytes of HTML. Nothing errored and
// nothing played — the render was fine and looked lost. Two things had to be
// true and neither was, so both are asserted here: the URL is the artifact route
// every other surface uses, and `nocode` resolves to the synthetic course that
// route serves from.
func TestNoCodeVideoURLIsServable(t *testing.T) {
	server, root := fixture(t)

	_, lesson, err := pipeline.NewNoCodePiece(root, pipeline.NoCodeSpec{
		Title: "Watch it build itself",
		Segments: []pipeline.NoCodeSegment{{
			Template: "verdict",
			Prompt:   "who this is for",
			Evidence: pipeline.NoCodeEvidence{
				Kind:  pipeline.EvidenceFact,
				Facts: []string{"a landing page is a fine place to start"},
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Before the render there is no URL, but there is still an address to point
	// at — "it will land here" is what the page shows while nothing exists.
	var before NoCodeDetail
	if rec := doJSON(t, server.Handler(), "GET", "/api/nocode/"+lesson.ID, "", &before); rec.Code != 200 {
		t.Fatalf("detail before render = %d: %s", rec.Code, rec.Body)
	}
	if before.VideoURL != "" {
		t.Errorf("VideoURL = %q before anything was rendered, want empty", before.VideoURL)
	}
	if before.VideoPath == "" {
		t.Error("VideoPath is empty; the page has nowhere to point")
	}
	if len(before.Stages) == 0 {
		t.Error("Stages is empty; the page cannot show progress against a real list")
	}

	if err := os.MkdirAll(lesson.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	final := filepath.Join(lesson.GeneratedDir(), pipeline.FinalVideoName)
	if err := os.WriteFile(final, []byte("fake mp4 bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	var after NoCodeDetail
	if rec := doJSON(t, server.Handler(), "GET", "/api/nocode/"+lesson.ID, "", &after); rec.Code != 200 {
		t.Fatalf("detail after render = %d: %s", rec.Code, rec.Body)
	}
	want := "/artifacts/" + pipeline.NoCodeCourseSlug + "/" + lesson.ID + "/" + pipeline.FinalVideoName
	if after.VideoURL != want {
		t.Errorf("VideoURL = %q, want %q", after.VideoURL, want)
	}

	// The part that actually broke: fetching it.
	rec := doJSON(t, server.Handler(), "GET", after.VideoURL, "", nil)
	if rec.Code != 200 {
		t.Fatalf("GET %s = %d: %s", after.VideoURL, rec.Code, rec.Body)
	}
	if got := rec.Body.String(); got != "fake mp4 bytes" {
		t.Errorf("served %q, want the video bytes — a fall-through to the SPA is the old bug", got)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "video/mp4" {
		t.Errorf("Content-Type = %q, want video/mp4", ct)
	}
}
