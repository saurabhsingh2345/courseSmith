package pipeline

import (
	"context"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

func d2TestTheme() SceneTheme {
	return deriveVideoTheme(config.Colors{Primary: "#2563eb", Accent: "#f5b841", Background: "#f8fafc"}, config.Fonts{}, "test", "")
}

func TestCompileD2ProducesSVG(t *testing.T) {
	svg, err := compileD2(context.Background(), "user -> app: request\napp -> db: query\ndb.shape: cylinder", d2TestTheme())
	if err != nil {
		t.Fatalf("compileD2: %v", err)
	}
	s := string(svg)
	if !strings.Contains(s, "<svg") {
		t.Fatalf("no <svg> in output (%d bytes)", len(svg))
	}
	for _, label := range []string{"user", "app", "db", "request", "query"} {
		if !strings.Contains(s, label) {
			t.Errorf("compiled SVG missing label %q", label)
		}
	}
}

func TestCompileD2RejectsBadSource(t *testing.T) {
	if _, err := compileD2(context.Background(), "a -> {unclosed", d2TestTheme()); err == nil {
		t.Fatal("invalid d2 source must fail to compile")
	}
}

func TestCompileD2Deterministic(t *testing.T) {
	src := "a -> b\nb -> c"
	one, err := compileD2(context.Background(), src, d2TestTheme())
	if err != nil {
		t.Fatal(err)
	}
	two, err := compileD2(context.Background(), src, d2TestTheme())
	if err != nil {
		t.Fatal(err)
	}
	if string(one) != string(two) {
		t.Fatal("d2 compilation must be deterministic (renderer flickers otherwise)")
	}
}
