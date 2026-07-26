package pipeline

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/enfec/coursesmith/internal/config"
)

// These tests drive the real headless-Chromium compile path (embedded
// Mermaid.js / Rough.js). rod provisions a Chromium on first use, so they are
// gated behind COURSESMITH_LIVE=1 to keep normal `go test` offline and fast.

func liveScreenshotter(t *testing.T) *RodScreenshotter {
	t.Helper()
	if os.Getenv("COURSESMITH_LIVE") != "1" {
		t.Skip("set COURSESMITH_LIVE=1 to run headless-browser compile tests")
	}
	return &RodScreenshotter{}
}

func TestRenderMermaidLive(t *testing.T) {
	r := liveScreenshotter(t)
	defer r.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	theme := deriveVideoTheme(config.Colors{Primary: "#2563eb", Accent: "#f5b841", Background: "#f8fafc"}, config.Fonts{}, "live")
	svg, err := r.RenderMermaid(ctx, "flowchart TD\n  A[Start] --> B{Empty?}\n  B -->|yes| C[Grow]\n  B -->|no| D[Append]", theme)
	if err != nil {
		t.Fatalf("RenderMermaid failed: %v", err)
	}
	if err := validateCompiledSVG(svg); err != nil {
		t.Fatalf("mermaid produced invalid SVG: %v", err)
	}
	if !strings.Contains(string(svg), "<svg") {
		t.Fatalf("output is not an SVG: %.80q", svg)
	}
	// The theme must actually reach the output: branded node fill, and no
	// max-width cap left on the ROOT element (the renderer scales to fill
	// the stage; label-wrapping max-widths inside the stylesheet are fine).
	if !strings.Contains(string(svg), theme.Surface) {
		t.Errorf("themed surface colour %s not found in compiled SVG", theme.Surface)
	}
	rootTag := string(svg[:strings.IndexByte(string(svg), '>')+1])
	if strings.Contains(rootTag, "max-width") {
		t.Errorf("compiled SVG root still carries a max-width cap: %s", rootTag)
	}

	// Invalid syntax must surface as an error (drives the model's repair loop).
	if _, err := r.RenderMermaid(ctx, "flowchart TD\n  A --> --> ??? {{{", theme); err == nil {
		t.Fatalf("expected an error for invalid mermaid syntax")
	}
}

func TestRenderExcalidrawLive(t *testing.T) {
	r := liveScreenshotter(t)
	defer r.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	sceneJSON, _ := json.Marshal(validScene())
	theme := deriveVideoTheme(config.Colors{Primary: "#2563eb", Accent: "#f5b841", Background: "#f8fafc"}, config.Fonts{}, "live")
	svg, err := r.RenderExcalidraw(ctx, sceneJSON, theme)
	if err != nil {
		t.Fatalf("RenderExcalidraw failed: %v", err)
	}
	if err := validateCompiledSVG(svg); err != nil {
		t.Fatalf("excalidraw produced invalid SVG: %v", err)
	}
	// One <g> group per element lets the renderer stagger the reveal.
	if n := strings.Count(string(svg), "<g"); n < len(validScene().Elements) {
		t.Fatalf("expected >= %d <g> groups, got %d", len(validScene().Elements), n)
	}
}
