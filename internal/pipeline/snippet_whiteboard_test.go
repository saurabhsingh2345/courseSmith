package pipeline

import (
	"strings"
	"testing"
)

// whiteboardPlan is a well-formed five-item board: an intro beat with nothing
// drawn, three beats that build a chain, and a closing beat that adds nothing.
func whiteboardPlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "whiteboard",
		Title:    "How HTTP caching works",
		Subtitle: "Three places a response can be waiting",
		Beats: []SnippetBeat{
			{ID: "the-question", Heading: "The question", Narration: strings.Repeat("question ", 22)},
			{ID: "the-browser", Heading: "The browser", Narration: strings.Repeat("browser ", 22), Sketch: []SketchItem{
				{Label: "Browser", Icon: "monitor"},
				{Label: "Local cache", Icon: "box", LinkFrom: "Browser"},
			}},
			{ID: "the-edge", Heading: "The edge", Narration: strings.Repeat("edge ", 22), Sketch: []SketchItem{
				{Label: "CDN edge", Icon: "globe", LinkFrom: "Local cache"},
			}},
			{ID: "the-origin", Heading: "The origin", Narration: strings.Repeat("origin ", 22), Sketch: []SketchItem{
				{Label: "Origin server", Icon: "server", LinkFrom: "CDN edge"},
				{Label: "Database", Icon: "database", LinkFrom: "Origin server"},
			}},
			{ID: "the-payoff", Heading: "Why it matters", Narration: strings.Repeat("payoff ", 22)},
		},
	}
}

func TestValidateWhiteboardPlan(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := whiteboardPlan().Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	t.Run("too few items", func(t *testing.T) {
		p := whiteboardPlan()
		p.Beats[3].Sketch = nil
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "board has 3 items") {
			t.Fatalf("want item-count error, got %v", err)
		}
	})
	t.Run("duplicate label", func(t *testing.T) {
		p := whiteboardPlan()
		p.Beats[2].Sketch[0].Label = "Browser"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "appears twice") {
			t.Fatalf("want duplicate-label error, got %v", err)
		}
	})
	t.Run("label too long", func(t *testing.T) {
		p := whiteboardPlan()
		p.Beats[2].Sketch[0].Label = "a label with far too many words in it"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "labels are at most") {
			t.Fatalf("want label-length error, got %v", err)
		}
	})
	// An arrow has to start somewhere. A link naming an item that has not been
	// drawn yet would silently vanish, so it is rejected instead.
	t.Run("forward link", func(t *testing.T) {
		p := whiteboardPlan()
		p.Beats[1].Sketch[0].LinkFrom = "Database"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "not on the board yet") {
			t.Fatalf("want forward-link error, got %v", err)
		}
	})
	t.Run("rejects code fields", func(t *testing.T) {
		p := whiteboardPlan()
		p.Beats[1].Code = "print('nope')"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "does not use") {
			t.Fatalf("want wrong-template-field error, got %v", err)
		}
	})
}

func TestWhiteboardScenes(t *testing.T) {
	plan := whiteboardPlan()
	scenes, err := whiteboardScenes(sceneInput(t, plan, 9000))
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != 2 {
		t.Fatalf("got %d scenes, want a title card plus a board", len(scenes))
	}
	title, boardScene := scenes[0], scenes[1]
	if title.Type != SceneTitle || boardScene.Type != SceneWhiteboard {
		t.Fatalf("scene types = %s, %s; want title, whiteboard", title.Type, boardScene.Type)
	}
	if got := title.EndMs - title.StartMs; got > maxTitleCardMs {
		t.Errorf("title card runs %dms, want at most %d", got, maxTitleCardMs)
	}
	// The board takes over the rest of the intro rather than leaving a gap, and
	// runs to the end so the finished picture is still up for the closing beat.
	if boardScene.StartMs != title.EndMs {
		t.Errorf("board starts at %d but the title ends at %d", boardScene.StartMs, title.EndMs)
	}
	if boardScene.EndMs != 45000 {
		t.Errorf("board ends at %d, want the end of the clip (45000)", boardScene.EndMs)
	}

	items, ok := boardScene.Props["items"].([]map[string]any)
	if !ok {
		t.Fatalf("board items have the wrong shape: %#v", boardScene.Props["items"])
	}
	if len(items) != 5 {
		t.Fatalf("got %d items, want 5", len(items))
	}
	// Items must be ordered and timed to the beat that introduces them.
	last := -1
	for i, item := range items {
		at, _ := item["atMs"].(int)
		if at <= last {
			t.Errorf("item %d is timed at %d, not after the previous item (%d)", i, at, last)
		}
		last = at
	}
	// The first item has nothing to link from; every later one in this plan does.
	if _, has := items[0]["from"]; has {
		t.Error("the first item should have no incoming arrow")
	}
	for i := 1; i < len(items); i++ {
		from, has := items[i]["from"].(int)
		if !has {
			t.Errorf("item %d has no link, but its plan entry named one", i)
			continue
		}
		if from >= i {
			t.Errorf("item %d links from %d — an arrow must come from an earlier item", i, from)
		}
	}
	// Links resolve by label, so the chain has to match what the plan described.
	if from, _ := items[2]["from"].(int); from != 1 {
		t.Errorf("CDN edge links from item %d, want 1 (Local cache)", from)
	}
}

// A board whose items all land in the first beat still gets a title card only
// if there is room for one.
func TestWhiteboardScenesWithoutIntroBeat(t *testing.T) {
	plan := whiteboardPlan()
	plan.Beats = plan.Beats[1:] // drop the no-sketch opener
	scenes, err := whiteboardScenes(sceneInput(t, plan, 9000))
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneWhiteboard {
		t.Fatalf("want a single board scene, got %d: %+v", len(scenes), scenes)
	}
}

func TestWhiteboardScenesRejectsEmptyBoard(t *testing.T) {
	plan := whiteboardPlan()
	for i := range plan.Beats {
		plan.Beats[i].Sketch = nil
	}
	if _, err := whiteboardScenes(sceneInput(t, plan, 9000)); err == nil {
		t.Fatal("want an error for a board with nothing on it")
	}
}

func TestSketchKey(t *testing.T) {
	if sketchKey("The Browser!") != sketchKey("the browser") {
		t.Error("sketchKey should ignore case and punctuation so links match loosely")
	}
	if sketchKey("CDN edge") == sketchKey("CDN cache") {
		t.Error("sketchKey collapsed two distinct labels")
	}
}
