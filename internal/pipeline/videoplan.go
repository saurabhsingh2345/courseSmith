package pipeline

// video-plan.yaml: the human edit layer over the generated scene graph.
// The pipeline decides the default scenes; this file (next to lesson.md,
// like review-notes.yaml and quiz overrides) lets an editor retarget any
// scene — swap its template variant, patch its props, or drop it entirely
// (the previous scene extends over its span, keeping audio sync intact).
// Editing the file re-stales the scenegraph stage.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/enfec/coursesmith/internal/project"
)

// VideoPlanFileName is the per-lesson editable plan, in the lesson dir.
const VideoPlanFileName = "video-plan.yaml"

// VideoPlanEdit is one edit, addressed by generated-scene index (the index
// printed by `coursesmith status` / visible in lesson-video.json).
type VideoPlanEdit struct {
	Scene int `yaml:"scene"`
	// Template swaps the scene's template variant (e.g. points: rows|grid).
	Template string `yaml:"template,omitempty"`
	// Props are merged over the scene's generated props.
	Props map[string]any `yaml:"props,omitempty"`
	// Skip drops the scene; the previous scene extends across its span.
	Skip bool `yaml:"skip,omitempty"`
}

// VideoPlan is the parsed video-plan.yaml.
type VideoPlan struct {
	Edits []VideoPlanEdit `yaml:"edits"`
}

// loadVideoPlan reads <lesson>/video-plan.yaml; missing is nil.
func loadVideoPlan(l *project.Lesson) (*VideoPlan, error) {
	path := filepath.Join(l.Dir, VideoPlanFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var plan VideoPlan
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&plan); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &plan, nil
}

// applyVideoPlan applies the edits to the generated scenes. Out-of-range
// scene indices are an error (the plan is stale against the current graph);
// skipping the first scene is allowed only when a scene follows it.
func applyVideoPlan(graph *SceneGraph, plan *VideoPlan) error {
	if plan == nil {
		return nil
	}
	skip := map[int]bool{}
	for _, edit := range plan.Edits {
		if edit.Scene < 0 || edit.Scene >= len(graph.Scenes) {
			return fmt.Errorf("%s: scene %d does not exist (graph has %d scenes 0-%d) — regenerate or fix the plan",
				VideoPlanFileName, edit.Scene, len(graph.Scenes), len(graph.Scenes)-1)
		}
		sc := &graph.Scenes[edit.Scene]
		if edit.Template != "" {
			sc.Props["template"] = edit.Template
		}
		for k, v := range edit.Props {
			sc.Props[k] = v
		}
		if edit.Skip {
			skip[edit.Scene] = true
		}
	}
	if len(skip) == 0 {
		return nil
	}
	if len(skip) >= len(graph.Scenes) {
		return fmt.Errorf("%s: cannot skip every scene", VideoPlanFileName)
	}
	// Drop skipped scenes; each removed span is absorbed by the previous
	// surviving scene (or the next one when the first scene is skipped).
	var kept []Scene
	for i, sc := range graph.Scenes {
		if !skip[i] {
			kept = append(kept, sc)
			continue
		}
		if len(kept) > 0 {
			kept[len(kept)-1].EndMs = sc.EndMs
		} else {
			// Skipped head: the next surviving scene starts early.
			for j := i + 1; j < len(graph.Scenes); j++ {
				if !skip[j] {
					graph.Scenes[j].StartMs = sc.StartMs
					break
				}
			}
		}
	}
	graph.Scenes = kept
	return nil
}
