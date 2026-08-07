package pipeline

// Batch-level guards for the foundations family.
//
// The per-template tests check what each template says about its own subject.
// These check the things that are only true of the BATCH, and that no single
// template's test is in a position to notice: that the family is complete, that
// its members agree on the facts a gallery sorts and filters on, and that the
// page which offers them will actually have something to show.
//
// The failure this exists to catch is a quiet one. `Family` is a plain string
// with no compiler behind it, and it is read in exactly one place — the studio
// catalog endpoint — so a template that misspells it, or forgets it, does not
// break: it simply appears on no page at all. The template is registered, the
// CLI can name it, every other test passes, and the only symptom is a gallery
// card nobody can find. A count is the cheapest thing that notices.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// foundationsRoster is the batch, by name. Written out rather than derived from
// the registry, because a test that asks the registry what is in the family and
// then checks the family contains those things cannot fail. This is the list the
// batch was commissioned as; if a template is dropped or renamed, the deletion
// should be a deliberate edit here rather than a silent shrink.
var foundationsRoster = []string{
	// Course scaffolding.
	"syllabus", "outcome", "bridge", "drill", "labcard", "mission",
	// Machines and systems.
	"machine", "blueprint", "relay", "layers", "pipeline",
	// Binary and logic.
	"radix", "carry", "bitfield", "encode", "gates",
	// Memory and storage.
	"ladder", "regions", "lookup",
	// Operating systems.
	"states", "scheduler", "shell",
	// Networking.
	"journey", "handshake",
	// Algorithms.
	"stepper", "growth", "callstack",
	// Git and history.
	"history", "versus", "eras",
}

func TestFoundationsRosterIsRegistered(t *testing.T) {
	for _, name := range foundationsRoster {
		tpl, ok := SnippetTemplates[name]
		if !ok {
			t.Errorf("foundations template %q is not registered", name)
			continue
		}
		if tpl.Family != FamilyFoundations {
			t.Errorf("template %q has family %q, want %q — a template outside the family appears on no gallery page at all",
				name, tpl.Family, FamilyFoundations)
		}
		if tpl.Since != SinceV7 {
			t.Errorf("template %q has Since %q, want %q", name, tpl.Since, SinceV7)
		}
	}
}

// TestFoundationsIsTheWholeFamily is the roster check run the other way: no
// template may join the family without being written down here, so the list
// above stays the description of the batch rather than a stale subset of it.
func TestFoundationsIsTheWholeFamily(t *testing.T) {
	inRoster := map[string]bool{}
	for _, n := range foundationsRoster {
		inRoster[n] = true
	}
	for _, tpl := range SnippetTemplates {
		if tpl.Family == FamilyFoundations && !inRoster[tpl.Name] {
			t.Errorf("template %q declares the foundations family but is not in foundationsRoster — add it there so the batch stays described in one place",
				tpl.Name)
		}
	}
}

// TestFoundationsTemplatesAreOffered guards the gallery itself. A shelved
// template is invisible to the studio, `snippet --list` and the caster alike, so
// a whole family that shipped shelved would be a page with nothing on it.
func TestFoundationsTemplatesAreOffered(t *testing.T) {
	offered := map[string]bool{}
	for _, tpl := range SnippetTemplateList() {
		offered[tpl.Name] = true
	}
	for _, name := range foundationsRoster {
		if _, ok := SnippetTemplates[name]; !ok {
			continue // reported by TestFoundationsRosterIsRegistered
		}
		if !offered[name] {
			t.Errorf("foundations template %q is shelved, so the foundations gallery will not offer it", name)
		}
	}
}

// TestFoundationsTemplatesHaveTheirOwnPrompt checks the one file the registry
// points at but does not own. A missing prompt is not caught until a clip is
// actually planned, which is the most expensive place to discover it.
func TestFoundationsTemplatesHaveTheirOwnPrompt(t *testing.T) {
	for _, name := range foundationsRoster {
		tpl, ok := SnippetTemplates[name]
		if !ok {
			continue
		}
		want := "snippet_" + name + ".tmpl"
		if tpl.PromptFile != want {
			t.Errorf("template %q reads prompt %q, want %q", name, tpl.PromptFile, want)
		}
		if _, err := os.Stat(filepath.Join("..", "..", "prompts", tpl.PromptFile)); err != nil {
			t.Errorf("template %q names prompt %q, which is not on disk", name, tpl.PromptFile)
		}
	}
}

// TestFoundationsTemplatesOwnAPayload catches the copy-paste mistake this batch
// is most exposed to: thirty templates written from one exemplar, where the
// second one to be adapted keeps the first one's Owns line. The symptom is not a
// compile error — it is a template that silently strips its own beats' payloads
// and then fails validation for reasons that make no sense.
func TestFoundationsTemplatesOwnAPayload(t *testing.T) {
	seen := map[beatFields]string{}
	for _, name := range foundationsRoster {
		tpl, ok := SnippetTemplates[name]
		if !ok {
			continue
		}
		if (tpl.Owns == beatFields{}) {
			t.Errorf("template %q owns no beat field, so every beat payload the model writes is stripped before validation", name)
			continue
		}
		if prior, dup := seen[tpl.Owns]; dup {
			t.Errorf("templates %q and %q declare the same Owns — one of them is reading the other's beat payload", prior, name)
		}
		seen[tpl.Owns] = name
	}
}

// TestFoundationsGalleryCopyIsWritten checks the words a creator actually reads.
// The catalog-wide test already requires these to be non-empty; what it cannot
// say is whether they were written or pasted, and a description that does not
// say when to reach for the template is the difference between a catalog and a
// list of names.
func TestFoundationsGalleryCopyIsWritten(t *testing.T) {
	for _, name := range foundationsRoster {
		tpl, ok := SnippetTemplates[name]
		if !ok {
			continue
		}
		if n := len(strings.Fields(tpl.Description)); n < 12 {
			t.Errorf("template %q has a %d-word description; say what the picture is AND when to reach for it", name, n)
		}
		if !strings.Contains(strings.ToLower(tpl.Description), "reach for it") {
			t.Errorf("template %q never says when to reach for it, which is the one thing a gallery card is for", name)
		}
		if n := len(strings.Fields(tpl.Example)); n < 4 {
			t.Errorf("template %q has a %d-word example prompt; it should read like something a creator would really type", name, n)
		}
	}
}
