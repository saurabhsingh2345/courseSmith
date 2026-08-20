package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// DefaultRendererDir is the Remotion project at the repo root.
const DefaultRendererDir = "renderer"

// VideoRenderer turns a scene graph into final.mp4. Implemented by
// RemotionRenderer; nil on Env falls back to the legacy ffmpeg assembly.
type VideoRenderer interface {
	Name() string
	Render(ctx context.Context, l *project.Lesson, graph *SceneGraph, outPath string) error
}

// RemotionRenderer shells out to `npx remotion render` in the renderer/
// project.
type RemotionRenderer struct {
	// Dir is the Remotion project directory ("" uses DefaultRendererDir).
	Dir string
	// Concurrency is passed to remotion render (0 lets Remotion decide).
	Concurrency int
	// FrameTimeoutMs bounds how long one frame may take (0 uses
	// DefaultFrameTimeoutMs).
	FrameTimeoutMs int
}

// DefaultFrameTimeoutMs is the per-frame render budget.
//
// Remotion's own default is 30s, which is ample for these scenes — a frame
// measures ~100ms — but not for a machine that is busy. A single frame missing
// that deadline aborts the whole render, so a loaded laptop turned a finished
// clip into a failed stage. The budget is raised rather than the scenes made
// cheaper, because the scenes are not the problem.
const DefaultFrameTimeoutMs = 180_000

func (r *RemotionRenderer) Name() string { return "remotion" }

func (r *RemotionRenderer) dir() string {
	if r.Dir != "" {
		return r.Dir
	}
	return DefaultRendererDir
}

// assetPaths lists every lesson-relative file the scene graph references.
func (g *SceneGraph) assetPaths() []string {
	paths := []string{g.AudioFile}
	if g.SFXFile != "" {
		// Missing this is silent in the worst way: the render succeeds, the
		// picture is right, and the typing track simply is not there.
		paths = append(paths, g.SFXFile)
	}
	for _, s := range g.Scenes {
		if src, ok := s.Props["src"].(string); ok && src != "" {
			paths = append(paths, src)
		}
	}
	return paths
}

// StageAssets copies the scene graph's assets from the lesson's generated
// dir into the renderer's public dir under subdir, and returns a copy of the
// graph with AssetBase set. Remotion can only load files via staticFile(),
// which resolves inside public/.
func (r *RemotionRenderer) StageAssets(l *project.Lesson, graph *SceneGraph, subdir string) (*SceneGraph, error) {
	jobDir := filepath.Join(r.dir(), "public", filepath.FromSlash(subdir))
	if err := os.RemoveAll(jobDir); err != nil {
		return nil, fmt.Errorf("clearing %s: %w", jobDir, err)
	}
	for _, rel := range graph.assetPaths() {
		src := filepath.Join(l.GeneratedDir(), filepath.FromSlash(rel))
		dst := filepath.Join(jobDir, filepath.FromSlash(rel))
		if err := copyFile(src, dst); err != nil {
			return nil, fmt.Errorf("staging render asset: %w", err)
		}
	}
	staged := *graph
	staged.AssetBase = subdir
	return &staged, nil
}

// checkInstalled verifies node + the renderer project are usable.
func (r *RemotionRenderer) checkInstalled() error {
	if _, err := exec.LookPath("npx"); err != nil {
		return fmt.Errorf("node/npx not found — install Node 18+ (https://nodejs.org) to render videos with Remotion")
	}
	if _, err := os.Stat(filepath.Join(r.dir(), "src", "index.ts")); err != nil {
		return fmt.Errorf("renderer project not found at %s — run coursesmith from the project root", r.dir())
	}
	if _, err := os.Stat(filepath.Join(r.dir(), "node_modules", "remotion")); err != nil {
		return fmt.Errorf("renderer dependencies missing — run: cd %s && npm install", r.dir())
	}
	return nil
}

func (r *RemotionRenderer) Render(ctx context.Context, l *project.Lesson, graph *SceneGraph, outPath string) error {
	if err := r.checkInstalled(); err != nil {
		return err
	}
	staged, err := r.StageAssets(l, graph, "jobs/"+l.ID)
	if err != nil {
		return err
	}
	propsData, err := json.Marshal(staged)
	if err != nil {
		return fmt.Errorf("encoding render props: %w", err)
	}
	propsFile := filepath.Join(r.dir(), "public", "jobs", l.ID, "props.json")
	if err := writeFileAtomic(propsFile, propsData); err != nil {
		return err
	}
	absProps, err := filepath.Abs(propsFile)
	if err != nil {
		return fmt.Errorf("resolving %s: %w", propsFile, err)
	}
	absOut, err := filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("resolving %s: %w", outPath, err)
	}

	timeout := r.FrameTimeoutMs
	if timeout <= 0 {
		timeout = DefaultFrameTimeoutMs
	}
	args := []string{
		"remotion", "render",
		"src/index.ts", "LessonVideo", absOut,
		"--props=" + absProps,
		"--overwrite",
		"--timeout=" + strconv.Itoa(timeout),
		// Lossless intermediate frames, and a quality floor on the encode.
		//
		// Both defaults are wrong for this content and the reason is the same.
		// Remotion hands each frame to the encoder as JPEG (quality 80) and lets
		// x264 pick its own rate; that is a sensible default for footage, where
		// grain and motion hide everything. These frames are the opposite of
		// footage: enormous flat fields of one bone colour with hairline-thin
		// serif edges on top. JPEG spends its bits on the flat area and rings
		// around the edges, and x264 at an unpinned rate does the same, so type
		// that is standing perfectly still acquires a shimmer that moves frame to
		// frame. It is most visible on exactly the frames meant to be calm — a
		// held title card — which is where a viewer is most likely to notice.
		//
		// So: jpeg at quality 100 rather than the default 80, and a pinned crf.
		//
		// NOT lossless png, which is what this used to say and which does not
		// scale. Remotion writes every frame to disk as an image before encoding,
		// so the intermediate cost is per-frame times the frame count — fine for a
		// 5,600-frame clip and fatal for a 46,000-frame lesson, where png ran the
		// disk out of space mid-render. Quality 100 jpeg is visually
		// indistinguishable here and roughly a tenth of the size, which keeps the
		// fix and drops the failure mode.
		"--image-format=jpeg",
		"--jpeg-quality=100",
		"--crf=16",
	}
	if r.Concurrency > 0 {
		args = append(args, "--concurrency="+strconv.Itoa(r.Concurrency))
	}
	cmd := exec.CommandContext(ctx, "npx", args...)
	cmd.Dir = r.dir()
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("remotion render failed: %w\n%s", err, tailLines(output.String(), 20))
	}
	return nil
}

// runRenderStage renders generated/final.mp4 from the scene graph via
// Remotion. Without node/Remotion available it falls back to the legacy
// ffmpeg slides/recording assembly with a warning, so the pipeline still
// completes end to end.
func runRenderStage(ctx context.Context, e *Env, course *project.Course, l *project.Lesson, cfg config.Config) error {
	if e.Renderer == nil {
		fmt.Fprintf(e.out(), "  ⚠ render    Remotion unavailable (install Node 18+ and run `cd %s && npm install`) — falling back to ffmpeg slide assembly\n", DefaultRendererDir)
		return runVideoStage(ctx, e, course, l, cfg)
	}
	graph, err := LoadSceneGraph(l)
	if err != nil {
		return err
	}
	final := filepath.Join(l.GeneratedDir(), FinalVideoName)
	fmt.Fprintf(e.out(), "  → render    %s: %d scenes, %.1fs at 1920x1080/30fps (%s)...\n",
		FinalVideoName, len(graph.Scenes), float64(graph.DurationMs)/1000, e.Renderer.Name())
	if err := e.Renderer.Render(ctx, l, graph, final); err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "    %s written\n", FinalVideoName)

	// Split the finished video into per-chapter chunks so each section is a
	// downloadable clip of its own. Chunking is a convenience product of the
	// render, not a gate on it — a failure warns rather than failing the stage.
	if n, err := splitVideoByChapters(ctx, e, l, final); err != nil {
		fmt.Fprintf(e.out(), "  ⚠ render    section chunks failed: %v\n", err)
	} else if n > 0 {
		fmt.Fprintf(e.out(), "    %s/ written (%d section clips)\n", SectionsDirName, n)
	}
	return nil
}

// SectionsDirName holds the per-chapter clips under generated/.
const SectionsDirName = "sections"

// splitVideoByChapters cuts final.mp4 at the chapter boundaries into
// generated/sections/NN-slug.mp4 and returns how many clips it wrote.
//
// Cuts are re-encoded, not stream-copied: chapter starts land wherever the
// narration puts them, and -c copy can only cut on keyframes — up to a couple
// of seconds of the previous section would leak into each clip. crf 20 veryfast
// matches the master's quality closely and keeps the split to a few seconds
// per chunk.
func splitVideoByChapters(ctx context.Context, e *Env, l *project.Lesson, final string) (int, error) {
	raw, err := os.ReadFile(filepath.Join(l.GeneratedDir(), ChaptersJSONFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // no chapters, nothing to split
		}
		return 0, err
	}
	var chapters []Chapter
	if err := json.Unmarshal(raw, &chapters); err != nil {
		return 0, fmt.Errorf("parsing %s: %w", ChaptersJSONFileName, err)
	}
	if len(chapters) < 2 {
		return 0, nil // a single chapter IS the video; a clip would duplicate it
	}

	dir := filepath.Join(l.GeneratedDir(), SectionsDirName)
	// Rebuild from scratch so renamed sections don't leave stale clips behind.
	if err := os.RemoveAll(dir); err != nil {
		return 0, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, err
	}
	for i, ch := range chapters {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		out := filepath.Join(dir, fmt.Sprintf("%02d-%s.mp4", i+1, slugify(ch.Title)))
		args := []string{
			"-ss", fmt.Sprintf("%.3f", float64(ch.StartMs)/1000),
		}
		// The last chapter runs to the end of the file (the video may outlast
		// the chapter list by the tail padding).
		if i+1 < len(chapters) {
			args = append(args, "-to", fmt.Sprintf("%.3f", float64(ch.EndMs)/1000))
		}
		args = append(args,
			"-i", final,
			"-c:v", "libx264", "-preset", "veryfast", "-crf", "20",
			"-c:a", "aac", "-movflags", "+faststart",
			out,
		)
		if err := e.runFFmpeg(ctx, args...); err != nil {
			return 0, fmt.Errorf("chapter %d (%s): %w", i+1, ch.Title, err)
		}
	}
	return len(chapters), nil
}
