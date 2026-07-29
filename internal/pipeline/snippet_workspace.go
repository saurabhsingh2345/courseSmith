package pipeline

// The workspace template: an editor shot like a screen recording.
//
// It is the `vscode` template's bigger sibling, and the difference is the
// *presentation*, not the code. `vscode` is a window floating on the brand
// stage — the right frame for eight lines that read large and sit still. This
// one fills the screen, has no stage behind it, and moves: it zooms into the
// function being written, pans to the tree when a file opens, pulls back out
// for the wide shot, drops to the terminal for the payoff. It is imitating a
// capture of somebody working.
//
// That is the whole selection rule. Reach for this one when the clip should
// feel like a screen recording, when the code is longer than a window holds,
// or when there is more than one file. Several files are *supported*, not
// required — one file shot with a moving camera is a perfectly good clip, and
// insisting on a second file produced exactly the thing nobody wants: an
// invented module that exists to satisfy a validator.
//
// The model never writes a coordinate. It names a subject (`focus`) and the
// renderer owns every number, the same bargain the story template's staging
// makes and for the same reason.
//
// **The program is executed while it is being planned.** A project spanning
// files cannot go through the verify stage — verify runs each fenced block on
// its own, where `import greet` is a ModuleNotFoundError — so the plan stage
// runs the whole file set through the sandbox (CodeRunner.RunProject) and
// feeds the real stdout back into the clip. A program that does not run is a
// *correction round*, not a shipped video: the model is handed its own
// traceback and asked again. That is what makes this template able to attempt
// something complicated — the loop that catches it is execution, not taste.

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "workspace",
		Category:    CatCode,
		Title:       "Screen recording",
		Description: "A full-screen editor moving like a screencast across several files. Reach for it when the subject is a project rather than a snippet.",
		Example:     "Write a function that finds duplicates in a list, then run it",
		PromptFile:  snippetWorkspaceTemplateName,
		NeedsCode:   true,
		// A camera move is only a camera move if it has time to be one — a
		// zoom that has to complete in a second is a cut. Below this the
		// template's whole reason for existing does not fit.
		MinTargetSec:     30,
		DefaultTargetSec: 60,
		Owns:             beatFields{Work: true},
		OwnsPlan:         planFields{Project: true},
		Normalize:        normalizeWorkspacePlan,
		Validate:         validateWorkspacePlan,
		Scenes:           workspaceScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Focuses":  strings.Join(WorkspaceFocusNames(), ", "),
				"MinFiles": minProjectFiles,
				"MaxFiles": maxProjectFiles,
				"MaxLines": maxProjectFileLines,
			}
		},
	})
}

const snippetWorkspaceTemplateName = "snippet_workspace.tmpl"

// Project bounds.
//
// One file is allowed: this template is chosen for how it *looks*, and a
// single file shot with a moving camera is a legitimate clip. Requiring two
// only ever produced a second file invented to satisfy the rule.
//
// Four is the ceiling because the tab bar is read at a glance and a fifth tab
// is a tab nobody looks at.
//
// The line cap is generous because the camera can move down a file — the code
// area scrolls to keep the active line in frame, so the constraint is reading
// time rather than screen space. Past about forty lines a clip is scrolling
// faster than anyone reads.
const (
	minProjectFiles     = 1
	maxProjectFiles     = 4
	maxProjectFileLines = 40
)

// workspaceFocusVocab is where the camera can look. It mirrors the switch in
// renderer/src/components/WorkspaceScene.tsx; a drift test keeps them equal.
//
// Subjects, not coordinates. "The imports at the top of this file" is a thing
// a writer can mean; y=214 is not, and a model handed pixel values produces
// camera moves that frame the gap between two panels.
var workspaceFocusVocab = map[string]string{
	// The whole editor, no zoom — the establishing shot, and where a beat that
	// is really about the narration should sit.
	"wide": "the whole editor",
	// In on the code being written, following the line the typing is on.
	"code": "the code under the cursor",
	// The file tree, for the beat that says which file we are opening.
	"tree": "the file tree",
	// The tab bar, for the beat about moving between files.
	"tabs": "the open tabs",
	// The integrated terminal. Implied by a beat that runs, so it rarely has
	// to be asked for.
	"terminal": "the terminal",
}

func WorkspaceFocusNames() []string {
	out := make([]string, 0, len(workspaceFocusVocab))
	for k := range workspaceFocusVocab {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func normalizeWorkspaceFocus(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))
	if _, ok := workspaceFocusVocab[n]; ok {
		return n
	}
	return "wide"
}

// workspaceScenes lays the clip out as ONE scene spanning the whole clip.
//
// One scene, not one per beat, because the camera has to be continuous. A
// scene per beat remounts the editor every few seconds, which is a cut — and a
// cut is the opposite of the screen recording this template is imitating. The
// beats become timed steps inside it, the same shape the data template uses
// for the same reason.
func workspaceScenes(in SnippetSceneInput) ([]Scene, error) {
	proj := in.Plan.Project
	if proj == nil {
		return nil, fmt.Errorf("the plan has no project")
	}

	files := make([]map[string]any, 0, len(proj.Files))
	for _, f := range proj.Files {
		files = append(files, map[string]any{"path": f.Path, "code": f.Code})
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Work == nil {
			return nil, fmt.Errorf("beat %q has no work direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"file":    beat.Work.File,
			"through": beat.Work.Through,
			"focus":   normalizeWorkspaceFocus(beat.Work.Focus),
			"run":     beat.Work.Run,
		}
		if beat.Work.Caption != "" {
			step["caption"] = beat.Work.Caption
		}
		steps = append(steps, step)
	}

	command := proj.Command
	if command == "" {
		command = "python3 " + proj.Entry
	}

	_, clipStart, _ := in.Beat(0)
	_, _, lastEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneWorkspace,
		StartMs: clipStart,
		EndMs:   lastEnd,
		Props: map[string]any{
			"title":   in.Plan.Title,
			"project": projectName(proj.Entry),
			"files":   files,
			"entry":   proj.Entry,
			"command": command,
			// What the interpreter really printed, captured when the plan was
			// executed. See the file header.
			"output": proj.Output,
			"steps":  steps,
		},
	}}, nil
}

// projectName is the folder the tree is labelled with. Derived rather than
// asked for: one more field the model can get wrong, for a string nobody reads
// closely.
func projectName(entry string) string {
	base := strings.TrimSuffix(path.Base(entry), path.Ext(entry))
	if base == "" || base == "main" || base == "run" {
		return "project"
	}
	return base
}

// normalizeWorkspacePlan points every beat at a file that exists.
//
// A beat naming "main.py" when the project calls it "app.py" is the model
// losing track of its own file list halfway down a long reply, not a
// disagreement about the walkthrough — and the renderer cannot open a file that
// is not there. Falling back to the entry file keeps the camera somewhere
// sensible; the focus and the caption the beat carries still apply.
func normalizeWorkspacePlan(p *SnippetPlan) {
	if p.Project == nil {
		return
	}
	paths := map[string]string{}
	for i := range p.Project.Files {
		f := &p.Project.Files[i]
		f.Path = strings.TrimSpace(f.Path)
		paths[strings.ToLower(f.Path)] = f.Path
		paths[strings.ToLower(path.Base(f.Path))] = f.Path
	}
	p.Project.Entry = strings.TrimSpace(p.Project.Entry)
	if resolved, ok := paths[strings.ToLower(p.Project.Entry)]; ok {
		p.Project.Entry = resolved
	}
	for i := range p.Beats {
		w := p.Beats[i].Work
		if w == nil {
			continue
		}
		w.Caption = collapseSpaces(w.Caption)
		w.Focus = normalizeWorkspaceFocus(w.Focus)
		if w.Through < 0 {
			w.Through = 0
		}
		if resolved, ok := paths[strings.ToLower(strings.TrimSpace(w.File))]; ok {
			w.File = resolved
		} else {
			w.File = p.Project.Entry
		}
	}
}

func validateWorkspacePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Work: true}); err != nil {
		return err
	}
	proj := p.Project
	if proj == nil {
		return fmt.Errorf("the plan has no project — this template is a program that spans files")
	}
	if n := len(proj.Files); n < minProjectFiles || n > maxProjectFiles {
		return fmt.Errorf("the project has %d files, want %d-%d", n, minProjectFiles, maxProjectFiles)
	}

	seen := map[string]bool{}
	for _, f := range proj.Files {
		p := strings.TrimSpace(f.Path)
		if p == "" {
			return fmt.Errorf("a project file has no path")
		}
		// The paths are written to a real directory before the program runs,
		// so this is a containment check and not a style rule.
		if path.IsAbs(p) || strings.Contains(p, "..") {
			return fmt.Errorf("file path %q must be a plain relative path inside the project", p)
		}
		if !strings.HasSuffix(p, ".py") {
			return fmt.Errorf("file %q must be a .py file", p)
		}
		if seen[p] {
			return fmt.Errorf("file %q appears twice", p)
		}
		seen[p] = true
		if strings.TrimSpace(f.Code) == "" {
			return fmt.Errorf("file %q is empty", p)
		}
		if n := len(strings.Split(strings.TrimRight(f.Code, "\n"), "\n")); n > maxProjectFileLines {
			return fmt.Errorf("file %q is %d lines; at most %d — a clip cannot read more than that",
				p, n, maxProjectFileLines)
		}
	}
	if !seen[strings.TrimSpace(proj.Entry)] {
		return fmt.Errorf("entry %q is not one of the project's files (%s)",
			proj.Entry, strings.Join(sortedFilePaths(seen), ", "))
	}

	ran := false
	for _, b := range p.Beats {
		if b.Work == nil {
			return fmt.Errorf("beat %q has no work direction — every beat in this template is on screen somewhere", b.ID)
		}
		if !seen[strings.TrimSpace(b.Work.File)] {
			return fmt.Errorf("beat %q is in file %q, which the project does not have", b.ID, b.Work.File)
		}
		if b.Work.Through < 0 {
			return fmt.Errorf("beat %q has a negative `through`", b.ID)
		}
		if n := len(strings.Fields(b.Work.Caption)); n > maxCaptionWords {
			return fmt.Errorf("beat %q has a %d-word caption; at most %d", b.ID, n, maxCaptionWords)
		}
		if b.Work.Run {
			ran = true
		}
	}
	// A project you never run is a project nobody can believe. Running it is
	// also the only moment the real output appears.
	if !ran {
		return fmt.Errorf("no beat runs the project — mark the beat where the terminal opens with \"run\": true")
	}
	if p.Beats[0].Work.Run {
		return fmt.Errorf("the first beat runs the project before any of it is on screen — run it after the code has been written")
	}
	return nil
}

func sortedFilePaths(seen map[string]bool) []string {
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// runPlannedProject executes a plan's file set and records what it printed.
//
// Called from inside the planner's correction loop, so a failure is a retry
// with the traceback attached rather than a broken clip. Returns nil for plans
// that have no project, which is every template but this one.
func (e *Env) runPlannedProject(ctx context.Context, p *SnippetPlan) error {
	if p.Project == nil {
		return nil
	}
	if e.CodeRunner == nil {
		// No sandbox configured (a unit test, a dry run). The template's own
		// validation has already checked the shape; refusing here would make
		// the planner untestable without docker.
		return nil
	}
	res, err := e.CodeRunner.RunProject(ctx, p.Project.FileMap(), p.Project.Entry, codeTimeout)
	if err != nil {
		// Infrastructure, not the model's fault — do not spend a correction
		// round on it.
		return nil
	}
	switch {
	case res.TimedOut:
		return fmt.Errorf("the project did not finish within %s — remove the long-running work; this clip runs it for real",
			codeTimeout)
	case res.ExitCode != 0:
		return fmt.Errorf("running %s failed (exit %d) and this clip shows the real terminal, so the program has to work. Fix the code:\n%s",
			p.Project.Entry, res.ExitCode, tailLines(res.Stderr, 8))
	case strings.TrimSpace(res.Stdout) == "":
		return fmt.Errorf("running %s printed nothing — the payoff of this clip is the terminal, so the program must print something worth watching",
			p.Project.Entry)
	}
	p.Project.Output = strings.TrimRight(res.Stdout, "\n")
	return nil
}
