package pipeline

import (
	"os"
	"regexp"
	"sort"
	"testing"
)

// The icon vocabulary lives in two places by necessity: Go validates and
// normalizes the names a model may emit, and the renderer maps them to Lucide
// components. Neither can import the other.
//
// A name Go allows but the renderer does not map renders as a placeholder dot —
// which is exactly the bug that made an "origin server" box show a bare circle,
// and it is invisible until someone watches the finished video. This test is the
// cheap guard: the two sets must be identical.
const iconMirrorPath = "../../renderer/src/components/icons.ts"

var tsIconEntryRe = regexp.MustCompile(`(?m)^\s{2}([a-z][a-z0-9]*):\s+[A-Z]`)

func TestIconVocabularyInSync(t *testing.T) {
	src, err := os.ReadFile(iconMirrorPath)
	if err != nil {
		t.Fatalf("reading %s: %v", iconMirrorPath, err)
	}
	matches := tsIconEntryRe.FindAllSubmatch(src, -1)
	if len(matches) == 0 {
		t.Fatalf("no icon entries parsed from %s — has the map's shape changed?", iconMirrorPath)
	}
	renderer := make(map[string]bool, len(matches))
	for _, m := range matches {
		renderer[string(m[1])] = true
	}

	var missingInRenderer, extraInRenderer []string
	for name := range pointIconVocab {
		if !renderer[name] {
			missingInRenderer = append(missingInRenderer, name)
		}
	}
	for name := range renderer {
		if !pointIconVocab[name] {
			extraInRenderer = append(extraInRenderer, name)
		}
	}
	sort.Strings(missingInRenderer)
	sort.Strings(extraInRenderer)

	if len(missingInRenderer) > 0 {
		t.Errorf("pointIconVocab allows %v, which %s does not map — those boxes would render a placeholder dot",
			missingInRenderer, iconMirrorPath)
	}
	if len(extraInRenderer) > 0 {
		t.Errorf("%s maps %v, which pointIconVocab rejects — the model can never ask for them",
			iconMirrorPath, extraInRenderer)
	}
}

// "dot" is the fallback every unknown name normalizes to, so it has to exist on
// both sides no matter what else changes.
//
// The whiteboard used to normalize against this vocabulary and no longer does —
// its items are figures now, so those assertions live beside that template. What
// still reads the icon set is the `points` scene, where a small chip beside a
// phrase is the right object and a 200-unit figure would be absurd.
func TestIconVocabularyHasFallback(t *testing.T) {
	if !pointIconVocab["dot"] {
		t.Error("the neutral fallback icon \"dot\" is missing from pointIconVocab")
	}
}
