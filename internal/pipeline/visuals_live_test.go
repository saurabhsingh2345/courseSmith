package pipeline

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// TestGenerateD3DiagramLive exercises the real generation path (diagram_d3.tmpl
// → content model → D3Spec.Validate) against the configured LLM. Skipped unless
// COURSESMITH_LIVE=1 and the provider key is present, so normal `go test`
// stays offline and deterministic.
func TestGenerateD3DiagramLive(t *testing.T) {
	if os.Getenv("COURSESMITH_LIVE") != "1" {
		t.Skip("set COURSESMITH_LIVE=1 to run live LLM generation")
	}
	if os.Getenv("GROQ_API_KEY") == "" && os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("no LLM API key in the environment")
	}

	e := &Env{Router: llm.NewRouter(""), PromptsDir: "../../prompts", Out: os.Stderr}
	l := &project.Lesson{FrontMatter: project.FrontMatter{Title: "What is a Python class?"}}
	spec := project.DiagramSpec{ID: "class-tree", Kind: project.DiagramKindD3,
		Prompt: "A small class hierarchy: Animal as the base class, with Dog and Cat subclasses, and Dog has a Puppy subclass."}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	d3, err := generateD3Diagram(ctx, e, l, config.Defaults(), spec, "")
	if err != nil {
		t.Fatalf("live generation failed: %v", err)
	}
	if err := d3.Validate(); err != nil {
		t.Fatalf("generated spec is invalid: %v", err)
	}
	t.Logf("layout=%s nodes=%d edges=%d", d3.Layout, len(d3.Nodes), len(d3.Edges))
	for _, n := range d3.Nodes {
		t.Logf("  node %s = %q (group %d)", n.ID, n.Label, n.Group)
	}
}
