package pipeline

import (
	"os"
	"regexp"
	"sort"
	"testing"
)

// The figure vocabulary lives in two places by necessity: Go validates and
// normalizes the names a model may emit, and the renderer draws them. Neither
// can import the other.
//
// A name Go allows but the renderer does not draw leaves half the frame empty —
// and unlike a missing icon, which at least degrades to a visible dot, an
// illustration shot with no figure just looks like a typography slide that was
// meant to have artwork. It is invisible until someone watches the clip.
const figureMirrorPath = "../../renderer/src/components/artwork.tsx"

// The FIGURES block is located first and entries parsed inside it, rather than
// matching entries across the whole file. artwork.tsx also declares types whose
// fields have the same shape as a map entry (`palette: FigurePalette;`), and a
// file-wide regex happily reports those as figures the renderer supports.
var (
	tsFiguresBlockRe = regexp.MustCompile(`(?s)export const FIGURES[^{]*\{(.*?)\n\};`)
	tsFigureEntryRe  = regexp.MustCompile(`(?m)^\s{2}([a-z][a-z0-9]*):\s+[A-Z]`)
)

func TestFigureVocabularyInSync(t *testing.T) {
	src, err := os.ReadFile(figureMirrorPath)
	if err != nil {
		t.Fatalf("reading %s: %v", figureMirrorPath, err)
	}
	block := tsFiguresBlockRe.FindSubmatch(src)
	if block == nil {
		t.Fatalf("no FIGURES map found in %s — has its shape changed?", figureMirrorPath)
	}
	matches := tsFigureEntryRe.FindAllSubmatch(block[1], -1)
	if len(matches) == 0 {
		t.Fatalf("no figure entries parsed from %s", figureMirrorPath)
	}
	renderer := make(map[string]bool, len(matches))
	for _, m := range matches {
		renderer[string(m[1])] = true
	}

	var missingInRenderer, extraInRenderer []string
	for name := range artFigureVocab {
		if !renderer[name] {
			missingInRenderer = append(missingInRenderer, name)
		}
	}
	for name := range renderer {
		if !artFigureVocab[name] {
			extraInRenderer = append(extraInRenderer, name)
		}
	}
	sort.Strings(missingInRenderer)
	sort.Strings(extraInRenderer)

	if len(missingInRenderer) > 0 {
		t.Errorf("artFigureVocab allows %v, which %s does not draw — those shots would have an empty half-frame",
			missingInRenderer, figureMirrorPath)
	}
	if len(extraInRenderer) > 0 {
		t.Errorf("%s draws %v, which artFigureVocab rejects — the model can never ask for them",
			figureMirrorPath, extraInRenderer)
	}
}

// "spark" is the fallback every unknown name normalizes to, so it has to exist
// on both sides no matter what else changes.
func TestFigureVocabularyHasFallback(t *testing.T) {
	if !artFigureVocab["spark"] {
		t.Error("the neutral fallback figure \"spark\" is missing from artFigureVocab")
	}
	if got := normalizeArtFigure("definitely-not-a-figure"); got != "spark" {
		t.Errorf("normalizeArtFigure fallback = %q, want spark", got)
	}
	if got := normalizeArtFigure("  Rocket "); got != "rocket" {
		t.Errorf("normalizeArtFigure(%q) = %q, want rocket", "  Rocket ", got)
	}
}
