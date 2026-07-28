package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The failure this heals, exactly as it was reported:
//
//	stage plan: rendering "system" section of prompts/snippet_whiteboard.tmpl:
//	executing "system" at <.Shapes>: map has no entry for key "Shapes"
//
// Prompts are read from disk and their data map is compiled into the binary, so
// the two drift whenever a prompt is edited ahead of a rebuild — or, the way it
// usually happens, whenever a studio process outlives one. The creator had
// already written their request and paid for the planning call; losing the run
// over a phrase in a prompt is the wrong trade in every case.
func TestRenderHealsAKeyTheCallerDidNotSupply(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "drifted.tmpl", `{{define "system"}}Shapes: {{.Shapes}}. Icons: {{.Icons}}.{{end}}`+
		`{{define "user"}}Make one about {{.Prompt}}.{{end}}`)

	// A caller that knows about Prompt and nothing else — the stale binary.
	system, user, healed, err := renderPromptFileHealed(dir, "drifted.tmpl", map[string]any{"Prompt": "backpressure"})
	if err != nil {
		t.Fatalf("a drifted prompt should still render, got %v", err)
	}
	if !strings.Contains(user, "backpressure") {
		t.Errorf("the creator's own words were lost: %q", user)
	}
	// .Shapes is defined by some template in this build, so the real vocabulary
	// is what the model is told — not a blank.
	if !strings.Contains(system, "sticky") {
		t.Errorf("system prompt did not pick up the real shape list: %q", system)
	}
	if len(healed) != 2 {
		t.Fatalf("want a note per invented key, got %v", healed)
	}
	for _, note := range healed {
		if !strings.HasPrefix(note, ".Shapes") && !strings.HasPrefix(note, ".Icons") {
			t.Errorf("unexpected note %q", note)
		}
	}
}

// A key nothing in the build has ever heard of still cannot fail the run; it
// renders empty, and says so loudly enough to be fixed.
func TestRenderHealsAKeyNothingDefines(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "invented.tmpl", `{{define "system"}}Aim for {{.MoonPhase}} beats.{{end}}`+
		`{{define "user"}}{{.Prompt}}{{end}}`)

	system, _, healed, err := renderPromptFileHealed(dir, "invented.tmpl", map[string]any{"Prompt": "indexes"})
	if err != nil {
		t.Fatalf("want a rendered prompt, got %v", err)
	}
	if !strings.Contains(system, "Aim for  beats") {
		t.Errorf("want the unknown key rendered empty, got %q", system)
	}
	if len(healed) != 1 || !strings.Contains(healed[0], "drifted") {
		t.Fatalf("want a drift warning naming the key, got %v", healed)
	}
}

// Struct data heals too — several stages pass one rather than a map, and they
// are no less able to fall behind their prompt.
func TestRenderHealsStructData(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "struct.tmpl", `{{define "system"}}{{.ID}} needs {{.MinBeats}} beats{{end}}`+
		`{{define "user"}}{{.Prompt}}{{end}}`)

	system, _, healed, err := renderPromptFileHealed(dir, "struct.tmpl", visualQAPromptData{ID: "fig-1", Prompt: "a diagram"})
	if err != nil {
		t.Fatalf("want a rendered prompt, got %v", err)
	}
	if !strings.Contains(system, "fig-1") {
		t.Errorf("the struct's own fields were lost: %q", system)
	}
	if len(healed) != 1 {
		t.Fatalf("want one healed key, got %v", healed)
	}
}

// Healing must not paper over a template that is actually broken: the original
// error is what the caller should see, not a second confusing one.
func TestRenderStillFailsOnARealTemplateBug(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "nouser.tmpl", `{{define "system"}}fine{{end}}`)

	if _, _, _, err := renderPromptFileHealed(dir, "nouser.tmpl", map[string]any{}); err == nil ||
		!strings.Contains(err.Error(), `must contain {{define "user"}}`) {
		t.Fatalf("want the missing-section error, got %v", err)
	}
}

// The dot inside a range is an element, not the data map, so its fields are not
// keys anyone could supply and must not be invented as such.
func TestPromptKeysIgnoreRangeBodies(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "loop.tmpl", `{{define "system"}}Cover these: {{range .Items}}{{.Label}} {{end}}and stop.{{end}}`+
		`{{define "user"}}{{.Prompt}}{{end}}`)

	_, _, healed, err := renderPromptFileHealed(dir, "loop.tmpl", map[string]any{"Prompt": "x"})
	if err != nil {
		t.Fatalf("want a rendered prompt, got %v", err)
	}
	if len(healed) != 1 || !strings.HasPrefix(healed[0], ".Items") {
		t.Fatalf("want only .Items healed, got %v", healed)
	}
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
