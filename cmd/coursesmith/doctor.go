package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/pipeline"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check ffmpeg, the Kokoro TTS server, API keys, and prompt templates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(cmd)
		},
	}
}

type doctorCheck struct {
	name string
	// run returns a human-readable detail on success.
	run    func() (string, error)
	remedy string
	// optional checks degrade quality when missing but don't fail doctor.
	optional bool
}

func runDoctor(cmd *cobra.Command) error {
	env := newEnv(cmd)
	checks := []doctorCheck{
		{
			name:   "ffmpeg",
			run:    env.CheckFFmpeg,
			remedy: "install ffmpeg: brew install ffmpeg (macOS) / apt install ffmpeg (Debian)",
		},
		{
			name:   "kokoro tts server",
			run:    func() (string, error) { return checkKokoro(kokoroBaseURL()) },
			remedy: "docker run -p 8880:8880 ghcr.io/remsky/kokoro-fastapi-cpu:latest  (or set KOKORO_URL)",
		},
		{
			name:   llm.EnvGroqKey,
			run:    func() (string, error) { return checkEnvKey(llm.EnvGroqKey) },
			remedy: "create a free key at https://console.groq.com/keys, then: export GROQ_API_KEY=<key>",
		},
		{
			name:   llm.EnvOpenAIKey,
			run:    func() (string, error) { return checkEnvKey(llm.EnvOpenAIKey) },
			remedy: "create a key at https://platform.openai.com/api-keys, then: export OPENAI_API_KEY=<key>",
		},
		{
			name:   "prompt templates",
			run:    checkPrompts,
			remedy: "run coursesmith from the project root (the directory containing prompts/)",
		},
		{
			name: "whisperX aligner",
			run: func() (string, error) {
				if env.Aligner == nil {
					return "", fmt.Errorf("not installed — captions/scenes will use segment-level estimates")
				}
				return "installed (word-level sync enabled)", nil
			},
			remedy:   "cd tools/align && uv sync   (or set COURSESMITH_ALIGN)",
			optional: true,
		},
		{
			name: "docker sandbox",
			run: func() (string, error) {
				runner, note := pipeline.ResolveCodeRunner(cmd.Context())
				if note != "" {
					return "", fmt.Errorf("%s", note)
				}
				return fmt.Sprintf("image %s ready (%s)", pipeline.SandboxImage, runner.Name()), nil
			},
			remedy:   "install docker, then: docker build -t " + pipeline.SandboxImage + " sandbox/",
			optional: true,
		},
		{
			name: "remotion renderer",
			run: func() (string, error) {
				if env.Renderer == nil {
					return "", fmt.Errorf("node/npx not found — videos fall back to ffmpeg slides")
				}
				if _, err := os.Stat(filepath.Join(pipeline.DefaultRendererDir, "node_modules", "remotion")); err != nil {
					return "", fmt.Errorf("renderer dependencies missing")
				}
				return "ready (renderer/ + node_modules)", nil
			},
			remedy:   "install Node 18+, then: cd " + pipeline.DefaultRendererDir + " && npm install",
			optional: true,
		},
		{
			name: "vhs demo recorder",
			run: func() (string, error) {
				tapes, note := pipeline.ResolveTapeRunner(cmd.Context())
				if tapes == nil {
					return "", fmt.Errorf("no vhs available — [DEMO] markers cannot be recorded")
				}
				if note != "" {
					return "", fmt.Errorf("%s", note)
				}
				return fmt.Sprintf("ready (%s)", tapes.Name()), nil
			},
			remedy:   "docker build -t " + pipeline.SandboxImage + " sandbox/   (vhs is inside the image)",
			optional: true,
		},
	}

	out := cmd.OutOrStdout()
	failed := 0
	for _, c := range checks {
		detail, err := c.run()
		if err != nil {
			if c.optional {
				fmt.Fprintf(out, "  ○ %-22s %v (optional)\n", c.name, err)
				fmt.Fprintf(out, "    for best quality: %s\n", c.remedy)
				continue
			}
			failed++
			fmt.Fprintf(out, "  ✗ %-22s %v\n", c.name, err)
			fmt.Fprintf(out, "    fix: %s\n", c.remedy)
			continue
		}
		fmt.Fprintf(out, "  ✓ %-22s %s\n", c.name, detail)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d checks failed", failed, len(checks))
	}
	fmt.Fprintln(out, "\nAll checks passed — ready to run the full pipeline.")
	return nil
}

// checkKokoro treats any HTTP response from the server as "reachable".
func checkKokoro(baseURL string) (string, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(baseURL + "/audio/voices")
	if err != nil {
		return "", fmt.Errorf("not reachable at %s", baseURL)
	}
	defer resp.Body.Close()
	return fmt.Sprintf("reachable at %s (HTTP %d)", baseURL, resp.StatusCode), nil
}

func checkEnvKey(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("not set")
	}
	return fmt.Sprintf("set (%d chars)", len(v)), nil
}

func checkPrompts() (string, error) {
	required := []string{
		"script.tmpl", "review_rubric.tmpl", "diagram_svg.tmpl", "quiz.tmpl",
		"review_claims.tmpl", "review_accuracy.tmpl", "review_pedagogy.tmpl", "review_tone.tmpl",
		"quiz_distractors.tmpl", "quiz_difficulty.tmpl", "mistakes.tmpl", "exercises.tmpl",
		"concepts.tmpl", "terminology.tmpl", "bridge.tmpl",
	}
	for _, name := range required {
		path := filepath.Join(promptsDirName, name)
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("%s is missing", path)
		}
	}
	return fmt.Sprintf("%d templates found in %s/", len(required), promptsDirName), nil
}

func kokoroBaseURL() string {
	// TTS_URL is the engine-neutral name (any OpenAI-compatible
	// /v1/audio/speech server works, e.g. chatterbox-tts-api for cloned
	// voices); KOKORO_URL stays as the historical alias.
	if v := os.Getenv("TTS_URL"); v != "" {
		return v
	}
	if v := os.Getenv("KOKORO_URL"); v != "" {
		return v
	}
	return pipeline.DefaultTTSBaseURL
}

// newEnv builds the pipeline environment shared by run and doctor.
func newEnv(cmd *cobra.Command) *pipeline.Env {
	runner, _ := pipeline.ResolveCodeRunner(cmd.Context())
	tapes, _ := pipeline.ResolveTapeRunner(cmd.Context())
	env := &pipeline.Env{
		Router:     llm.NewRouter(llm.DefaultStateDir),
		PromptsDir: promptsDirName,
		Out:        cmd.OutOrStdout(),
		TTSBaseURL: kokoroBaseURL(),
		Aligner:    resolveAligner(),
		CodeRunner: runner,
		TapeRunner: tapes,
		Renderer:   resolveRenderer(),
		// Lazy: Chromium launches on first diagram QA; failures fall back
		// to source-text review inside the stage.
		Screenshotter: &pipeline.RodScreenshotter{},
	}
	if key := os.Getenv(llm.EnvGroqKey); key != "" {
		env.Transcriber = llm.NewGroqTranscriber(key, "", nil)
	}
	return env
}

// alignToolDir holds the whisperX wrapper mini-project.
const alignToolDir = "tools/align"

// resolveRenderer returns the Remotion renderer when node and the renderer
// project are present; nil falls back to the ffmpeg slide assembly.
func resolveRenderer() pipeline.VideoRenderer {
	if _, err := exec.LookPath("npx"); err != nil {
		return nil
	}
	if _, err := os.Stat(filepath.Join(pipeline.DefaultRendererDir, "src", "index.ts")); err != nil {
		return nil
	}
	return &pipeline.RemotionRenderer{}
}

// resolveAligner finds a whisperX alignment command: COURSESMITH_ALIGN
// overrides; otherwise the tools/align virtualenv is used when present.
// nil means "not installed" and the align stage falls back to Groq segments.
func resolveAligner() pipeline.Aligner {
	if cmd := os.Getenv("COURSESMITH_ALIGN"); cmd != "" {
		return &pipeline.SubprocessAligner{Cmd: strings.Fields(cmd)}
	}
	venvPython := filepath.Join(alignToolDir, ".venv", "bin", "python")
	if _, err := os.Stat(venvPython); err == nil {
		return &pipeline.SubprocessAligner{Cmd: []string{venvPython, filepath.Join(alignToolDir, "align.py")}}
	}
	return nil
}
