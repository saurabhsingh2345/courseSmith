package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// Every prompt's worked example must have `beats` where SnippetPlan reads it —
// at the TOP level, beside the template's own object, not nested inside it.
//
// This is not a style rule. A model copies the shape of the example it is given
// far more faithfully than it follows the prose above it, so an example with
// beats in the wrong place produces a reply that decodes to a plan with no
// beats at all — and the failure surfaces three correction rounds later as
// "plan has no beats", which names the symptom and not the cause. Two prompts
// shipped that way and neither was caught by anything: the templates' own tests
// build plans in Go, and the render tests use fixtures.
var exampleStartRe = regexp.MustCompile(`(?m)^\{"title"`)

// findExample returns the worked example, which may span several lines — so it
// is matched by balancing braces rather than by taking one line.
func findExample(raw []byte) []byte {
	loc := exampleStartRe.FindIndex(raw)
	if loc == nil {
		return nil
	}
	depth, inStr, esc := 0, false, false
	for i := loc[0]; i < len(raw); i++ {
		c := raw[i]
		switch {
		case esc:
			esc = false
		case c == '\\' && inStr:
			esc = true
		case c == '"':
			inStr = !inStr
		case inStr:
			// brace inside a string is not structure
		case c == '{':
			depth++
		case c == '}':
			depth--
			if depth == 0 {
				return raw[loc[0] : i+1]
			}
		}
	}
	return nil
}

func TestPromptExamplesPutBeatsAtTheTopLevel(t *testing.T) {
	for _, tpl := range SnippetTemplateList() {
		if tpl.PromptFile == "" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join("..", "..", "prompts", tpl.PromptFile))
		if err != nil {
			t.Errorf("%s: %v", tpl.Name, err)
			continue
		}
		m := findExample(raw)
		if m == nil {
			t.Errorf("%s: no worked example — the model has nothing to copy the shape from", tpl.Name)
			continue
		}
		var doc map[string]any
		if err := json.Unmarshal(m, &doc); err != nil {
			t.Errorf("%s: the worked example is not valid JSON: %v", tpl.Name, err)
			continue
		}
		beats, ok := doc["beats"].([]any)
		if !ok || len(beats) == 0 {
			t.Errorf("%s: the example has no top-level \"beats\". A model copies the example's shape, so this ships plans with no beats at all", tpl.Name)
		}
		// And it must not ALSO be nested, which is how the mistake looked.
		if own, ok := doc[tpl.Name].(map[string]any); ok {
			if _, nested := own["beats"]; nested {
				t.Errorf("%s: the example nests \"beats\" inside %q as well as at the top level", tpl.Name, tpl.Name)
			}
		}
	}
}
