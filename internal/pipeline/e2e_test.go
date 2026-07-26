package pipeline

// End-to-end pipeline run on the real python-basics/01 lesson.
//
// LLM responses are canned (no API keys needed); everything else is real:
// docker-sandboxed code verification, VHS terminal demos really executing
// python3 in docker, Kokoro TTS (local server), whisperX word alignment,
// silence compression, WebVTT captions, scene graph, and a full Remotion
// render to final.mp4.
//
// Gated because it needs docker + the sandbox image + a Kokoro server +
// node_modules and takes minutes:
//
//	COURSESMITH_E2E=1 go test ./internal/pipeline/ -run TestEndToEndLesson01 -v -timeout 45m

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/enfec/coursesmith/internal/project"
)

const e2eScriptJSON = `{
  "title": "What is Python?",
  "sections": [
    {"id": "a-language-for-talking-to-computers",
     "narration": "Computers are powerful, but they cannot guess what you mean. A programming language is a precise way of writing instructions, and Python is one designed to read almost like English. Here is the big picture.",
     "duration_est_sec": 12,
     "cues": [{"type": "diagram", "ref": "python-translator", "at_word": 33}]},
    {"id": "why-beginners-start-with-python",
     "narration": "Beginners love Python because one readable line can do real work, and the community is enormous. Almost any question you will ever have already has an answer online.",
     "duration_est_sec": 10, "cues": []},
    {"id": "where-python-is-used-in-the-real-world",
     "narration": "Python runs websites, crunches data, automates boring tasks, and powers most of modern artificial intelligence. Take a look at where it shows up.",
     "duration_est_sec": 9,
     "cues": [{"type": "diagram", "ref": "where-python-runs", "at_word": 20}]},
    {"id": "installing-python",
     "narration": "Head to python dot org and download the installer for your system. On Windows, check the box that says add Python to PATH before you continue. Then open a terminal and type python3 dash dash version to check it worked.",
     "duration_est_sec": 14, "cues": []},
    {"id": "your-first-line-of-python",
     "narration": "Time for the traditional first program. The print function takes whatever you put inside the quotes and shows it on screen. Watch it happen for real.",
     "duration_est_sec": 10,
     "cues": [{"type": "demo", "ref": "open the Python REPL with python3, print a hello message, then print your own name, and exit", "at_word": 24}]},
    {"id": "python-does-math-too",
     "narration": "Numbers do not need quotes. Give Python some arithmetic and it answers immediately, like a very obedient calculator.",
     "duration_est_sec": 8, "cues": []},
    {"id": "putting-lines-together",
     "narration": "A program is just lines executed top to bottom. Save a few lines in a file, run the file, and Python performs each one in order. Let us try that with a real script.",
     "duration_est_sec": 12,
     "cues": [{"type": "demo", "ref": "create a two-line script hello.py with a heredoc and run it with python3 hello.py", "at_word": 33}]},
    {"id": "what-s-next",
     "narration": "Next lesson we meet variables, the way programs remember things. Before then, make print say your own name. See you there.",
     "duration_est_sec": 8, "cues": []}
  ]
}`

func e2eSVG(title string) string {
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 420">
  <style>svg { --primary: #306998; --accent: #ffd43b; --bg: #ffffff; }
    .box { fill: var(--bg); stroke: var(--primary); stroke-width: 3; }
    .label { font-family: sans-serif; font-size: 24px; fill: #1f2937; text-anchor: middle; }
    .arrow { stroke: var(--accent); stroke-width: 5; }</style>
  <g id="background"><rect width="800" height="420" fill="var(--bg)"/>
    <text class="label" x="400" y="50" font-weight="bold">` + title + `</text></g>
  <g id="left"><rect class="box" x="60" y="150" width="180" height="110" rx="14"/>
    <text class="label" x="150" y="212">your code</text></g>
  <g id="middle"><rect class="box" x="310" y="150" width="180" height="110" rx="14"/>
    <text class="label" x="400" y="212">Python</text></g>
  <g id="right"><rect class="box" x="560" y="150" width="180" height="110" rx="14"/>
    <text class="label" x="650" y="212">result</text></g>
  <g id="arrows"><line class="arrow" x1="245" y1="205" x2="305" y2="205"/>
    <line class="arrow" x1="495" y1="205" x2="555" y2="205"/></g>
</svg>`
}

const e2eQuizJSON = `{
  "title": "Check your understanding",
  "questions": [
    {"id": "q1", "type": "recall", "prompt": "What is Python?",
     "options": ["A snake-care manual", "A programming language", "A web browser", "An operating system"],
     "answer_index": 1, "explanation": "Python is a programming language — a precise way to give computers instructions."},
    {"id": "q2", "type": "prediction", "prompt": "What does this print?\n` + "```python\\nprint(7 * 6)\\n```" + `",
     "options": ["76", "42", "7 * 6", "an error"],
     "answer_index": 1, "explanation": "Numbers without quotes are calculated: 7 times 6 is 42."},
    {"id": "q3", "type": "application", "prompt": "On Windows, which installer option should you check?",
     "options": ["Add Python to PATH", "Install for all snakes", "Disable pip", "Skip the license"],
     "answer_index": 0, "explanation": "Add Python to PATH makes the python command work in your terminal."},
    {"id": "q4", "type": "debugging", "prompt": "A learner writes print(Hello) without quotes and gets an error. Why?",
     "options": ["print needs a number", "Text must be inside quotes", "Hello is a keyword", "print is misspelled"],
     "answer_index": 1, "explanation": "Without quotes Python looks for a variable named Hello, which does not exist."}
  ]
}`

// Real beginner mistakes: the sandbox executes these and captures the
// authentic tracebacks.
const e2eMistakesJSON = `{
  "mistakes": [
    {"title": "Forgetting the quotes", "explanation": "print(Hello) reads naturally but Hello is not defined.",
     "broken_code": "print(Hello)", "fix": "Wrap text in quotes.", "fixed_code": "print(\"Hello\")"},
    {"title": "Mismatched quotes", "explanation": "Opening with one quote style and closing with another.",
     "broken_code": "print(\"Hello')", "fix": "Use the same quote character on both ends.", "fixed_code": "print(\"Hello\")"},
    {"title": "Capitalizing print", "explanation": "Python names are case-sensitive.",
     "broken_code": "Print(\"Hello\")", "fix": "Python's function is lowercase print.", "fixed_code": "print(\"Hello\")"}
  ]
}`

// Real exercises: solutions are verified against the hidden pytest tests in
// the sandbox; starters must fail them.
const e2eExercisesJSON = `{
  "exercises": [
    {"slug": "say-hello", "title": "Say hello",
     "description": "Write a function greet() that returns the text Hello, world!",
     "starter_code": "def greet():\n    # TODO: return the greeting\n    ...",
     "solution_code": "def greet():\n    return \"Hello, world!\"",
     "test_code": "from exercise import greet\n\ndef test_greet():\n    assert greet() == \"Hello, world!\"\n\ndef test_returns_string():\n    assert isinstance(greet(), str)"},
    {"slug": "double-it", "title": "Double it",
     "description": "Write a function double(n) that returns n times two.",
     "starter_code": "def double(n):\n    # TODO: return n times two\n    ...",
     "solution_code": "def double(n):\n    return n * 2",
     "test_code": "from exercise import double\n\ndef test_double():\n    assert double(2) == 4\n\ndef test_zero():\n    assert double(0) == 0"}
  ]
}`

const e2eReplTape = `Type "python3"
Enter
Sleep 1.5s
Type "print('Hello, world!')"
Enter
Sleep 2s
Type "print('My name is Ada')"
Enter
Sleep 2s
Type "exit()"
Enter
Sleep 2s`

const e2eScriptTape = `Type "cat > hello.py << 'EOF'"
Enter
Type "name = 'Ada'"
Enter
Type "print('Hello,', name)"
Enter
Type "EOF"
Enter
Sleep 1s
Type "python3 hello.py"
Enter
Sleep 3s`

func TestEndToEndLesson01(t *testing.T) {
	if os.Getenv("COURSESMITH_E2E") == "" {
		t.Skip("set COURSESMITH_E2E=1 to run the full end-to-end pipeline")
	}
	requireFFmpeg(t)
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	// Copy the real course into a temp dir so the repo stays clean.
	work := t.TempDir()
	courseDir := filepath.Join(work, "python-basics")
	lessonDir := filepath.Join(courseDir, "lessons", "01-what-is-python")
	if err := os.MkdirAll(lessonDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for src, dst := range map[string]string{
		filepath.Join(repoRoot, "courses/python-basics/course.yaml"):                            filepath.Join(courseDir, "course.yaml"),
		filepath.Join(repoRoot, "courses/python-basics/lessons/01-what-is-python", "lesson.md"): filepath.Join(lessonDir, "lesson.md"),
	} {
		if err := copyFile(src, dst); err != nil {
			t.Fatal(err)
		}
	}
	course, err := project.LoadCourse(courseDir)
	if err != nil {
		t.Fatal(err)
	}
	lesson, err := course.FindLesson("01")
	if err != nil {
		t.Fatal(err)
	}

	// Canned LLM responses in stage order; everything else is real —
	// mistakes tracebacks and exercise verification really run in docker.
	fake := &fakeRouter{
		content: []string{
			e2eScriptJSON,                // script stage
			e2eSVG("Python, translated"), // visuals: python-translator
			e2eSVG("Where Python runs"),  // visuals: where-python-runs
			e2eQuizJSON,                  // quiz stage
			e2eMistakesJSON,              // mistakes stage
			e2eExercisesJSON,             // exercises stage
			e2eReplTape,                  // demos: demo-1
			e2eScriptTape,                // demos: demo-2
		},
		review: []string{
			// script three-pass review; the checkable claim really executes
			// in the docker sandbox.
			`{"claims":[
				{"claim":"7 times 6 is 42","section":"python-does-math-too","checkable":true,"code":"assert 7 * 6 == 42"},
				{"claim":"print shows text on screen","section":"your-first-line-of-python","checkable":false}]}`,
			accuracyJSON(9, "Accurate, warm, and well paced."),
			pedagogyJSON(9, "Concepts land in order with worked examples."),
			toneJSON(9, "Matches the intended warmth."),
			visualVerdictJSON(true),                  // diagram 1 visual QA
			visualVerdictJSON(true),                  // diagram 2 visual QA
			reviewJSON(9, "Grounded in the lesson."), // quiz rubric
			distractorsJSON(8),                       // quiz distractor scoring
			difficultyJSON(7),                        // quiz difficulty simulation
		},
	}

	ctx := context.Background()
	codeRunner, note := ResolveCodeRunner(ctx)
	if note != "" {
		t.Fatalf("docker sandbox required for E2E: %s", note)
	}
	tapeRunner, tapeNote := ResolveTapeRunner(ctx)
	if tapeRunner == nil || tapeNote != "" {
		t.Fatalf("docker sandbox (vhs) required for E2E")
	}

	kokoroURL := os.Getenv("KOKORO_URL")
	if kokoroURL == "" {
		kokoroURL = DefaultTTSBaseURL
	}
	if resp, err := http.Get(strings.TrimSuffix(kokoroURL, "/") + "/audio/voices"); err != nil {
		t.Fatalf("Kokoro server not reachable at %s — start it first", kokoroURL)
	} else {
		resp.Body.Close()
	}

	alignerPython := filepath.Join(repoRoot, "tools/align/.venv/bin/python")
	var aligner Aligner
	if _, err := os.Stat(alignerPython); err == nil {
		aligner = &SubprocessAligner{Cmd: []string{alignerPython, filepath.Join(repoRoot, "tools/align/align.py")}}
	} else {
		t.Log("whisperX not installed — align stage will need a transcriber; failing")
		t.Fatal("run `cd tools/align && uv sync` first")
	}

	env := &Env{
		Router:        fake,
		PromptsDir:    filepath.Join(repoRoot, "prompts"),
		Out:           io.MultiWriter(os.Stdout, &strings.Builder{}),
		TTSBaseURL:    kokoroURL,
		Aligner:       aligner,
		CodeRunner:    codeRunner,
		TapeRunner:    tapeRunner,
		Renderer:      &RemotionRenderer{Dir: filepath.Join(repoRoot, "renderer")},
		Screenshotter: &fakeScreenshotter{}, // vision verdicts are canned anyway
		SiteDir:       filepath.Join(work, "site"),
	}

	start := time.Now()
	if err := env.RunLesson(ctx, course, lesson, RunOptions{}); err != nil {
		t.Fatal(err)
	}
	t.Logf("full pipeline completed in %s", time.Since(start).Round(time.Second))

	// The deliverables of the definition of done.
	for _, artifact := range []string{
		VerificationFileName, ScriptFileName, QuizFileName,
		VoiceoverFileName, AlignmentFileName, CaptionsFileName,
		SceneGraphFileName, FinalVideoName,
		filepath.Join(DemosDirName, "demo-1.mp4"),
		filepath.Join(DemosDirName, "demo-2.mp4"),
		filepath.Join(DiagramsDirName, "python-translator.svg"),
	} {
		if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), artifact)); err != nil {
			t.Errorf("missing artifact: %v", err)
		}
	}
	graph, err := LoadSceneGraph(lesson)
	if err != nil {
		t.Fatal(err)
	}
	types := map[string]int{}
	for _, s := range graph.Scenes {
		types[s.Type]++
	}
	t.Logf("scene graph: %d scenes (%v), %d caption words, %.1fs",
		len(graph.Scenes), types, len(graph.Captions), float64(graph.DurationMs)/1000)
	if types[SceneTerminal] < 2 || types[SceneDiagram] < 2 || types[SceneCode] < 1 {
		t.Errorf("scene mix incomplete: %v", types)
	}
	callouts := 0
	for _, s := range graph.Scenes {
		callouts += len(s.Callouts)
	}
	if callouts < 2 {
		t.Errorf("callouts = %d, want the 2 declared in front-matter", callouts)
	}

	// Preserve the outputs for inspection.
	keep := os.Getenv("COURSESMITH_E2E_KEEP")
	if keep != "" {
		if err := os.CopyFS(keep, os.DirFS(lesson.GeneratedDir())); err != nil {
			t.Logf("could not preserve outputs: %v", err)
		} else {
			t.Logf("outputs preserved in %s", keep)
		}
	}
	if fmt.Sprint(len(fake.content), len(fake.review)) != "0 0" {
		t.Errorf("unconsumed canned responses: %d content, %d review", len(fake.content), len(fake.review))
	}
}
