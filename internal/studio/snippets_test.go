package studio

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/pipeline"
	"github.com/enfec/coursesmith/internal/project"
)

// waitForRun blocks until the server's single run slot is free again. Snippet
// creation returns as soon as the pipeline is queued, so assertions about what
// it produced have to wait for it.
func waitForRun(t *testing.T, server *Server) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if !server.runs.Status().Running {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("the run did not finish within 10s")
}

// planningRouter returns a fixed valid vscode snippet plan for any request.
// The narration is deliberately full-length: the plan stage rejects a draft
// that comes in under the word budget for the requested runtime, so a terse
// fixture would fail the same gate a terse model does.
type planningRouter struct{}

func (planningRouter) Complete(_ context.Context, _ config.Pipeline, _ llm.TaskType, _ llm.Request) (*llm.Response, error) {
	return &llm.Response{Content: `{
		"title": "For Loops in Python",
		"subtitle": "Repeat work without repeating yourself",
		"beats": [
			{"id":"the-idea","heading":"The idea",
			 "narration":"A for loop repeats a block of code once for every item in a sequence, which means you write the step a single time and let Python do the repeating for you. That is the whole idea, and it is why loops show up in almost every program you will ever read."},
			{"id":"write-it","heading":"Writing the loop",
			 "narration":"Here we ask range for the numbers five down to one, and print each one as it comes. The third argument is the step, and making it negative is what walks the numbers backwards. After the loop finishes, one last print announces the liftoff.",
			 "code":"for i in range(5, 0, -1):\n    print(i)\nprint('Liftoff!')"},
			{"id":"run-it","heading":"Running it","run":true,
			 "narration":"Running the file walks the loop from five down to one, printing as it goes, and then falls through to the final line. The countdown and the liftoff appear in exactly the order the code describes, with nothing hidden in between."}
		]
	}`}, nil
}

// snippetFixture wires the snippet prompt template and a planning router so
// POST /api/snippets can actually plan.
func snippetFixture(t *testing.T) (*Server, string) {
	t.Helper()
	server, root := fixture(t)
	tmpl := `{{define "system"}}plan{{end}}{{define "user"}}{{.Prompt}}{{.Title}}{{.TargetSec}}` +
		`{{.TargetWords}}{{.MinWords}}{{.MaxWords}}{{.MinWordsPerBeat}}{{.MaxWordsPerBeat}}` +
		`{{.Tone}}{{.Audience}}{{.Language}}{{.CodeLanguage}}{{.PaceWPM}}{{.MinBeats}}{{.MaxBeats}}{{end}}`
	if err := os.WriteFile(filepath.Join(root, "prompts", "snippet_vscode.tmpl"), []byte(tmpl), 0o644); err != nil {
		t.Fatal(err)
	}
	server.env.Router = planningRouter{}
	return server, root
}

func TestSnippetTemplatesEndpoint(t *testing.T) {
	server, _ := fixture(t)
	var out []SnippetTemplateInfo
	if rec := doJSON(t, server.Handler(), "GET", "/api/snippet-templates", "", &out); rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if len(out) == 0 {
		t.Fatal("the gallery is empty")
	}
	var vscode *SnippetTemplateInfo
	for i := range out {
		if out[i].Name == "vscode" {
			vscode = &out[i]
		}
	}
	if vscode == nil {
		t.Fatal("no vscode template in the catalog")
	}
	if vscode.Title == "" || vscode.Description == "" || vscode.Example == "" {
		t.Errorf("gallery copy is incomplete: %+v", vscode)
	}
	if !vscode.ShowsCode {
		t.Error("the vscode template should be marked as running code")
	}
}

func TestSnippetCreateRejectsBadRequests(t *testing.T) {
	server, _ := snippetFixture(t)
	h := server.Handler()
	cases := map[string]string{
		"no prompt":        `{"template":"vscode"}`,
		"no template":      `{"prompt":"loops"}`,
		"unknown template": `{"prompt":"loops","template":"nope"}`,
		"absurd length":    `{"prompt":"loops","template":"vscode","target_sec":9000}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if rec := doJSON(t, h, "POST", "/api/snippets", body, nil); rec.Code != 400 {
				t.Fatalf("status %d, want 400: %s", rec.Code, rec.Body)
			}
		})
	}
}

// The end-to-end shape of the surface: create → the plan lands on disk → the
// list and detail endpoints describe it → delete removes it.
func TestSnippetCreatePlanListDelete(t *testing.T) {
	server, root := snippetFixture(t)
	h := server.Handler()

	var created CreateSnippetResponse
	rec := doJSON(t, h, "POST", "/api/snippets",
		`{"prompt":"how for loops work in python","template":"vscode","plan_only":true}`, &created)
	if rec.Code != 201 {
		t.Fatalf("create status %d: %s", rec.Code, rec.Body)
	}
	if created.ID == "" || created.Template != "vscode" {
		t.Fatalf("unexpected create response: %+v", created)
	}
	if created.Ready {
		t.Error("a plan-only snippet must not report a finished video")
	}

	// The snippet lives in the state dir, not in courses/.
	dir := filepath.Join(root, ".coursesmith", "snippets", "lessons", created.ID)
	if _, err := os.Stat(filepath.Join(dir, pipeline.SnippetFileName)); err != nil {
		t.Fatalf("snippet request not stored: %v", err)
	}
	waitForRun(t, server)

	planPath := filepath.Join(dir, "generated", pipeline.SnippetPlanFileName)
	raw, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("plan stage did not write %s: %v", pipeline.SnippetPlanFileName, err)
	}
	if !strings.Contains(string(raw), "For Loops in Python") {
		t.Errorf("plan does not carry the model's title:\n%s", raw)
	}

	var list []SnippetSummary
	if rec := doJSON(t, h, "GET", "/api/snippets", "", &list); rec.Code != 200 {
		t.Fatalf("list status %d: %s", rec.Code, rec.Body)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("list = %+v, want the one snippet", list)
	}
	if list[0].Title != "For Loops in Python" {
		t.Errorf("list title = %q, want the planned title", list[0].Title)
	}

	var detail SnippetDetail
	if rec := doJSON(t, h, "GET", "/api/snippets/"+created.ID, "", &detail); rec.Code != 200 {
		t.Fatalf("detail status %d: %s", rec.Code, rec.Body)
	}
	if detail.TargetSec == 0 {
		t.Error("detail is missing the target runtime")
	}
	var plan struct {
		Beats []struct {
			ID   string `json:"id"`
			Run  bool   `json:"run"`
			Code string `json:"code"`
		} `json:"beats"`
	}
	if err := json.Unmarshal(detail.Plan, &plan); err != nil {
		t.Fatalf("detail plan is not decodable: %v", err)
	}
	if len(plan.Beats) != 3 {
		t.Fatalf("got %d beats, want 3", len(plan.Beats))
	}
	if !plan.Beats[2].Run {
		t.Error("the last beat should be the one that runs the file")
	}

	// A snippet resolves as a lesson of the synthetic snippets course, which is
	// what makes the artifact and stage-status routes work for it.
	var lesson LessonDetail
	if rec := doJSON(t, h, "GET", "/api/lessons/snippets/"+created.ID, "", &lesson); rec.Code != 200 {
		t.Fatalf("lesson detail status %d: %s", rec.Code, rec.Body)
	}
	if lesson.Stages[project.StagePlan] != "done" {
		t.Errorf("plan stage status = %q, want done", lesson.Stages[project.StagePlan])
	}

	if rec := doJSON(t, h, "DELETE", "/api/snippets/"+created.ID, "", nil); rec.Code != 204 {
		t.Fatalf("delete status %d: %s", rec.Code, rec.Body)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("delete left the snippet directory behind")
	}
}

func TestSnippetDetailUnknownID(t *testing.T) {
	server, _ := snippetFixture(t)
	if rec := doJSON(t, server.Handler(), "GET", "/api/snippets/nope", "", nil); rec.Code != 404 {
		t.Fatalf("status %d, want 404", rec.Code)
	}
}
