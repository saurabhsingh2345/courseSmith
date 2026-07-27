package pipeline

// The VS Code walkthrough template.
//
// The clip everyone actually wants when they say "show me the code": an
// editor opens, a file is picked from the tree, code types itself in line by
// line, and then the integrated terminal slides up and runs it.
//
// The output in that terminal is not written by the model. The plan's code
// goes through the ordinary verify stage — really executed, in the sandbox —
// and the scene shows what the interpreter really printed. A clip that claims
// an output the code does not produce cannot be built.

import (
	"fmt"
	"strings"

	"github.com/enfec/coursesmith/internal/project"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "vscode",
		Title:       "VS Code walkthrough",
		Description: "An editor opens, code types itself in, and the terminal runs it for real.",
		Example:     "How for loops work in Python, with a countdown example",
		PromptFile:  snippetVSCodeTemplateName,
		NeedsCode:   true,
		Validate:    validateVSCodePlan,
		Scenes:      vscodeScenes,
	})
}

const snippetVSCodeTemplateName = "snippet_vscode.tmpl"

// runOpensAtFraction places the terminal's appearance inside the beat that
// runs the code: far enough in that the narrator has said "let's run it",
// early enough that the output has time to breathe.
const runOpensAtFraction = 0.28

// Title-card bounds.
//
// The opening beat's narration sets up the idea, and it can run fifteen
// seconds. Holding a title card for all of it is the dead-screen failure this
// engine already learned once: the card says everything it has to say in about
// four seconds. So the card is capped, and the editor takes over early —
// opening the window, browsing the tree, picking the file — which is motion
// that belongs to the intro rather than filler invented to cover it.
//
// Below the floor there is not enough room for the card to land at all, and it
// is folded into the walkthrough instead.
const (
	minTitleCardMs = 2200
	maxTitleCardMs = 5200
)

// minTypingLeadMs is how long the editor gets to open itself before the first
// character is typed, when there was no intro beat to cover it.
const minTypingLeadMs = 1300

func validateVSCodePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	seenCode := false
	for _, b := range p.Beats {
		if b.Code != "" {
			seenCode = true
		}
		if b.Run && !seenCode {
			return fmt.Errorf("beat %q runs the file before any beat has written code", b.ID)
		}
	}
	if !seenCode {
		return fmt.Errorf("a vscode snippet needs at least one beat with code")
	}
	if !hasRunBeat(p) {
		return fmt.Errorf("a vscode snippet must end by running the code — set run: true on a beat")
	}
	return rejectForeignBeatFields(p, beatFields{Code: true, Run: true})
}

func hasRunBeat(p *SnippetPlan) bool {
	for _, b := range p.Beats {
		if b.Run {
			return true
		}
	}
	return false
}

// vscodeScenes lays the clip out as an opening title card (the beats before
// any code exists) followed by one continuous editor scene that runs to the
// end. Keeping the editor mounted for the whole back half is what makes it
// feel like a session rather than a slideshow: the buffer persists, the
// terminal opens under it, and nothing re-mounts mid-thought.
func vscodeScenes(in SnippetSceneInput) ([]Scene, error) {
	firstCode := -1
	for i, b := range in.Plan.Beats {
		if b.Code != "" {
			firstCode = i
			break
		}
	}
	if firstCode < 0 {
		return nil, fmt.Errorf("no beat carries code")
	}

	_, clipStart, _ := in.Beat(0)
	_, firstCodeStart, _ := in.Beat(firstCode)

	// The title card takes the front of the intro, capped; the editor takes
	// the rest of it, so the run-up to the first keystroke is the window
	// opening rather than a held card.
	var scenes []Scene
	walkStart := clipStart
	if firstCodeStart-clipStart >= minTitleCardMs {
		titleEnd := min(firstCodeStart, clipStart+maxTitleCardMs)
		scenes = append(scenes, Scene{
			Type:    SceneTitle,
			StartMs: clipStart,
			EndMs:   titleEnd,
			Props: map[string]any{
				"heading":  in.Plan.Title,
				"subtitle": in.Plan.Subtitle,
				"intro":    true,
				"outcomes": leadHeadings(in, firstCode),
			},
		})
		walkStart = titleEnd
	}

	language := in.Spec.ResolvedCodeLanguage()
	file := snippetFileName(in.Plan.Title, language)
	command := runCommand(language, file)

	// Steps carry the buffer forward: a beat that only runs the file inherits
	// the code already on screen, so the model never has to repeat it.
	var steps []map[string]any
	lastCode := ""
	for i := firstCode; i < len(in.Plan.Beats); i++ {
		b, startMs, endMs := in.Beat(i)
		code := b.Code
		if code == "" {
			code = lastCode
		}
		if code == "" {
			continue
		}
		newBuffer := code != lastCode
		lastCode = code

		// A beat that neither changes the buffer nor runs anything is the
		// narrator talking over what is already on screen — no new step.
		if !newBuffer && !b.Run {
			continue
		}
		step := map[string]any{
			"code": code,
			"atMs": startMs,
		}
		if b.Run {
			out, ok := in.VerifiedOutput[project.HashBytes([]byte(code))]
			if !ok {
				return nil, fmt.Errorf("beat %q runs code that the verify stage did not execute — re-run the verify stage", b.ID)
			}
			step["run"] = true
			step["command"] = command
			step["output"] = out
			step["runAtMs"] = startMs + int(float64(endMs-startMs)*runOpensAtFraction)
		}
		steps = append(steps, step)
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("no editor steps were produced")
	}

	_, _, walkEnd := in.Beat(len(in.Plan.Beats) - 1)
	scenes = append(scenes, Scene{
		Type:    SceneWalkthrough,
		StartMs: walkStart,
		EndMs:   walkEnd,
		Props: map[string]any{
			"title":    in.Plan.Title,
			"language": language,
			"file":     file,
			"project":  workspaceName(in.Plan.Title),
			"files":    []string{file},
			"steps":    steps,
			// The snippet path plays the full choreography: the window scales
			// up, the file is picked out of the tree, the tab opens, and only
			// then does the cursor start typing. The lesson path leaves this
			// off — mid-lesson the editor should already be there.
			"intro": true,
			// When the first character lands. Everything before it is the
			// opening gesture, paced to fill exactly this gap however long the
			// intro beat turned out to be.
			"typeAtMs": max(firstCodeStart, walkStart+minTypingLeadMs),
		},
	})
	return scenes, nil
}

// workspaceName is the folder name shown in the editor chrome. It is
// deliberately not the title: repeating the clip's title in the title bar, the
// tab, and the scene header reads as three labels for one thing.
func workspaceName(title string) string {
	parts := strings.Split(slugify(title), "-")
	if len(parts) > 2 {
		parts = parts[:2]
	}
	name := strings.Join(parts, "-")
	if name == "" {
		return "workspace"
	}
	return name
}

// leadHeadings returns the headings of the beats before the code starts, used
// as the title card's bullet list. One heading is not a list — the card shows
// nothing rather than a lonely bullet.
func leadHeadings(in SnippetSceneInput, firstCode int) []string {
	if firstCode < 2 {
		return nil
	}
	out := make([]string, 0, firstCode)
	for i := 0; i < firstCode; i++ {
		out = append(out, in.Plan.Beats[i].Heading)
	}
	return out
}

// runCommand is what the terminal shows itself typing.
func runCommand(language, file string) string {
	switch strings.ToLower(language) {
	case "python":
		return "python3 " + file
	case "javascript":
		return "node " + file
	case "typescript":
		return "npx tsx " + file
	case "go":
		return "go run " + file
	case "ruby":
		return "ruby " + file
	case "bash":
		return "bash " + file
	default:
		return "run " + file
	}
}
