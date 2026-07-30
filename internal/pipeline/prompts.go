package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"text/template"
	"text/template/parse"

	"github.com/enfec/coursesmith/internal/config"
)

// Template file names under the prompts directory.
const (
	scriptTemplateName          = "script.tmpl"
	reviewTemplateName          = "review_rubric.tmpl"
	reviewPlanTemplateName      = "review_plan.tmpl"
	reviewClaimsTemplateName    = "review_claims.tmpl"
	reviewAccuracyTemplateName  = "review_accuracy.tmpl"
	reviewPedagogyTemplateName  = "review_pedagogy.tmpl"
	reviewToneTemplateName      = "review_tone.tmpl"
	diagramTemplateName         = "diagram_svg.tmpl"
	d3DiagramTemplateName       = "diagram_d3.tmpl"
	mermaidDiagramTemplateName  = "diagram_mermaid.tmpl"
	excalidrawTemplateName      = "diagram_excalidraw.tmpl"
	diagramVisualQATemplateName = "diagram_visual_qa.tmpl"
	quizTemplateName            = "quiz.tmpl"
	demoTapeTemplateName        = "demo_tape.tmpl"
	conceptsTemplateName        = "concepts.tmpl"
	terminologyTemplateName     = "terminology.tmpl"
	bridgeTemplateName          = "bridge.tmpl"
	quizDistractorsTemplateName = "quiz_distractors.tmpl"
	quizDifficultyTemplateName  = "quiz_difficulty.tmpl"
	mistakesTemplateName        = "mistakes.tmpl"
	exercisesTemplateName       = "exercises.tmpl"
	captionEmphasisTemplateName = "caption_emphasis.tmpl"
	storyboardTemplateName      = "storyboard.tmpl"
	d2LangTemplateName          = "diagram_d2lang.tmpl"
)

// renderPrompt resolves a template through the Env's prompt search path —
// the course's own prompts/ dir (archetype/course overrides) first, the
// project prompts/ dir second — and renders it.
//
// A key the prompt asks for and the caller did not supply is healed rather than
// fatal; the substitution is reported on the run's output so the drift is
// visible without it costing the user their generation.
func (e *Env) renderPrompt(file string, data any) (system, user string, err error) {
	dir := e.PromptsDir
	if e.CoursePromptsDir != "" {
		if _, statErr := os.Stat(filepath.Join(e.CoursePromptsDir, file)); statErr == nil {
			dir = e.CoursePromptsDir
		}
	}
	system, user, healed, err := renderPromptFileHealed(dir, file, data)
	for _, h := range healed {
		fmt.Fprintf(e.out(), "    ! %s asks for %s\n", file, h)
	}
	return system, user, err
}

// renderPromptFile loads a prompt template file and renders its "system"
// and "user" sections with data. Templates live on disk (prompts/*.tmpl)
// so they can be tuned without recompiling; each must contain
// {{define "system"}}...{{end}} and {{define "user"}}...{{end}}.
func renderPromptFile(promptsDir, file string, data any) (system, user string, err error) {
	system, user, _, err = renderPromptFileHealed(promptsDir, file, data)
	return system, user, err
}

// renderPromptFileHealed is renderPromptFile plus the list of keys it had to
// invent, one human-readable line each.
//
// Prompts are read from disk and the data map is compiled into the binary, so
// the two drift independently: a prompt edited to say {{.Shapes}} against a
// binary built before Shapes existed — or, far more common in practice, a
// studio process left running across a rebuild — used to fail the generation
// outright, after the creator had already asked for a clip and paid for the
// planning tokens. It is never worth failing a run over a phrase in a prompt.
//
// So the strict render is tried first, unchanged, and only a failure opens the
// healing path: every key the template actually references is collected from the
// parse tree, the missing ones are filled (from another template's data where
// some template in this build defines that key, otherwise empty), and the render
// is retried. If it still fails, the original error is what the caller sees —
// healing must never turn a genuine template bug into a confusing second error.
func renderPromptFileHealed(promptsDir, file string, data any) (system, user string, healed []string, err error) {
	path := filepath.Join(promptsDir, file)
	if _, statErr := os.Stat(path); statErr != nil {
		return "", "", nil, fmt.Errorf(
			"prompt template %s not found — run coursesmith from the project root (the directory containing %s/): %w",
			path, promptsDir, statErr,
		)
	}
	tmpl, err := template.New(file).Option("missingkey=error").ParseFiles(path)
	if err != nil {
		return "", "", nil, fmt.Errorf("parsing prompt template %s: %w", path, err)
	}
	system, user, err = renderPromptSections(tmpl, path, data)
	if err == nil {
		return system, user, nil, nil
	}
	filled, notes := healPromptData(tmpl, data)
	if len(notes) == 0 {
		return "", "", nil, err
	}
	system2, user2, err2 := renderPromptSections(tmpl, path, filled)
	if err2 != nil {
		return "", "", nil, err
	}
	return system2, user2, notes, nil
}

func renderPromptSections(tmpl *template.Template, path string, data any) (system, user string, err error) {
	system, err = renderSection(tmpl, path, "system", data)
	if err != nil {
		return "", "", err
	}
	user, err = renderSection(tmpl, path, "user", data)
	if err != nil {
		return "", "", err
	}
	return system, user, nil
}

func renderSection(tmpl *template.Template, path, section string, data any) (string, error) {
	if tmpl.Lookup(section) == nil {
		return "", fmt.Errorf(`prompt template %s must contain {{define %q}}...{{end}}`, path, section)
	}
	var sb strings.Builder
	if err := tmpl.ExecuteTemplate(&sb, section, data); err != nil {
		return "", fmt.Errorf("rendering %q section of %s: %w", section, path, err)
	}
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", fmt.Errorf("%q section of %s rendered empty", section, path)
	}
	return out, nil
}

// healPromptData returns a data map carrying every key the template references,
// plus a note for each key it had to invent. It returns no notes — and the
// render is left to fail — when the data is a shape it cannot extend, or when
// nothing was missing in the first place (in which case the failure is a real
// template bug, not drift).
func healPromptData(tmpl *template.Template, data any) (map[string]any, []string) {
	filled := promptDataMap(data)
	if filled == nil {
		return nil, nil
	}
	refs := promptKeysReferenced(tmpl)
	var notes []string
	for _, key := range refs.keys() {
		if _, ok := filled[key]; ok {
			continue
		}
		if v, ok := promptDataFallbacks()[key]; ok {
			filled[key] = v
			notes = append(notes, fmt.Sprintf(".%s, which this stage does not supply — filled in from the shared prompt data", key))
			continue
		}
		// What a blank has to be depends on what the template does with it: a
		// {{range}} over an empty string is an error, not an empty section.
		if refs.ranged[key] {
			filled[key] = []any{}
		} else {
			filled[key] = ""
		}
		notes = append(notes, fmt.Sprintf(".%s, which nothing in this build defines — rendered empty (the prompt file and the binary have drifted; rebuild, or restart a long-running studio)", key))
	}
	return filled, notes
}

// promptDataMap copies the caller's data into a map the healer can add to.
// Structs are flattened by field name, which is exactly how the template was
// addressing them; anything else (a map with non-string keys, a scalar) has no
// safe extension and returns nil.
func promptDataMap(data any) map[string]any {
	switch v := data.(type) {
	case nil:
		return map[string]any{}
	case map[string]any:
		out := make(map[string]any, len(v)+4)
		for k, val := range v {
			out[k] = val
		}
		return out
	}
	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return map[string]any{}
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	rt := rv.Type()
	out := make(map[string]any, rt.NumField()+4)
	for i := range rt.NumField() {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		out[f.Name] = rv.Field(i).Interface()
	}
	return out
}

// promptRefs is what a template asks of its data: every top-level field it
// reads off the dot, and which of those it walks with {{range}}.
type promptRefs struct {
	fields map[string]bool
	ranged map[string]bool
}

func (r promptRefs) keys() []string {
	out := make([]string, 0, len(r.fields))
	for k := range r.fields {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// promptKeysReferenced reads those references off the parse tree.
//
// Fields inside {{range}} and {{with}} bodies are deliberately not collected:
// the dot is rebound there, so those names belong to an element rather than to
// the data map, and inventing them at the top level would say nothing true.
func promptKeysReferenced(tmpl *template.Template) promptRefs {
	refs := promptRefs{fields: map[string]bool{}, ranged: map[string]bool{}}
	for _, t := range tmpl.Templates() {
		if t.Tree == nil {
			continue
		}
		collectPromptFields(t.Tree.Root, refs.fields)
	}
	for _, t := range tmpl.Templates() {
		if t.Tree == nil {
			continue
		}
		collectRangedFields(t.Tree.Root, refs.ranged)
	}
	return refs
}

func collectPromptFields(n parse.Node, out map[string]bool) {
	switch v := n.(type) {
	case nil:
		return
	case *parse.ListNode:
		if v == nil {
			return
		}
		for _, c := range v.Nodes {
			collectPromptFields(c, out)
		}
	case *parse.ActionNode:
		collectPromptFields(v.Pipe, out)
	case *parse.PipeNode:
		if v == nil {
			return
		}
		for _, c := range v.Cmds {
			collectPromptFields(c, out)
		}
	case *parse.CommandNode:
		for _, a := range v.Args {
			collectPromptFields(a, out)
		}
	case *parse.FieldNode:
		if len(v.Ident) > 0 {
			out[v.Ident[0]] = true
		}
	case *parse.ChainNode:
		collectPromptFields(v.Node, out)
	case *parse.IfNode:
		collectPromptFields(v.Pipe, out)
		collectPromptFields(v.List, out)
		collectPromptFields(v.ElseList, out)
	case *parse.RangeNode:
		collectPromptFields(v.Pipe, out) // the body's dot is an element, not the data
	case *parse.WithNode:
		collectPromptFields(v.Pipe, out)
	case *parse.TemplateNode:
		collectPromptFields(v.Pipe, out)
	}
}

// collectRangedFields notes the fields a template iterates, which is the one
// distinction that changes what an invented blank has to be.
func collectRangedFields(n parse.Node, out map[string]bool) {
	switch v := n.(type) {
	case nil:
		return
	case *parse.ListNode:
		if v == nil {
			return
		}
		for _, c := range v.Nodes {
			collectRangedFields(c, out)
		}
	case *parse.IfNode:
		collectRangedFields(v.List, out)
		collectRangedFields(v.ElseList, out)
	case *parse.WithNode:
		collectRangedFields(v.List, out)
		collectRangedFields(v.ElseList, out)
	case *parse.RangeNode:
		collectPromptFields(v.Pipe, out)
		collectRangedFields(v.List, out)
		collectRangedFields(v.ElseList, out)
	}
}

// promptDataFallbacks is every key some template in this build knows how to
// supply: the shared snippet data plus each template's own vocabularies and
// bounds. A prompt asking for {{.Shapes}} from a stage that does not pass it
// still gets the real shape list rather than a blank.
//
// Prompt and Title are excluded on purpose — they are the creator's own words,
// and a blank is more honest than a default.
var promptDataFallbacks = sync.OnceValue(func() map[string]any {
	cfg := config.Defaults()
	out := sharedPromptData(SnippetSpec{}, cfg)
	for _, name := range SnippetTemplateNames() {
		tpl := SnippetTemplates[name]
		if tpl.PromptData == nil {
			continue
		}
		for k, v := range tpl.PromptData(SnippetSpec{Template: name}, cfg) {
			if _, dup := out[k]; !dup {
				out[k] = v
			}
		}
	}
	delete(out, "Prompt")
	delete(out, "Title")
	return out
})
