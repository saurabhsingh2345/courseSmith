package studio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/pipeline"
)

// The studio's template gallery shows a preview thumbnail per template, served
// as a static asset named after the template. Nothing at runtime checks that
// the file is there — a missing one is a broken image on the one screen whose
// whole job is helping someone choose.
//
// So the catalog and the asset directory are asserted to match. Adding a
// template without running `node test/template_previews.mjs` fails here rather
// than in front of a user.
const previewDir = "../../studio/public/template-previews"

func TestEveryTemplateHasAPreview(t *testing.T) {
	entries, err := os.ReadDir(previewDir)
	if err != nil {
		t.Fatalf("reading %s: %v", previewDir, err)
	}
	have := map[string]bool{}
	for _, e := range entries {
		if name := strings.TrimSuffix(e.Name(), ".png"); name != e.Name() {
			have[name] = true
		}
	}

	for _, name := range pipeline.SnippetTemplateNames() {
		if !have[name] {
			t.Errorf("template %q has no preview at %s/%s.png — regenerate with `node test/template_previews.mjs`",
				name, previewDir, name)
		}
	}
	for name := range have {
		if _, ok := pipeline.SnippetTemplates[name]; !ok {
			t.Errorf("%s/%s.png has no template — a preview for something the catalog dropped",
				previewDir, name)
		}
	}
}

// A zero-byte or truncated PNG renders as a broken image exactly like a missing
// one, and an ffmpeg run that half-failed leaves precisely that.
func TestPreviewsAreRealImages(t *testing.T) {
	for _, name := range pipeline.SnippetTemplateNames() {
		path := filepath.Join(previewDir, name+".png")
		data, err := os.ReadFile(path)
		if err != nil {
			continue // the missing-file case is reported by the test above
		}
		if len(data) < 1024 {
			t.Errorf("%s is %d bytes — that is not a rendered preview", path, len(data))
		}
		if len(data) < 8 || string(data[1:4]) != "PNG" {
			t.Errorf("%s does not start with a PNG signature", path)
		}
	}
}
