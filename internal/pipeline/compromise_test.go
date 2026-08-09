package pipeline

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// The correction loop joins every round's complaint into one error, and the same
// rule reappears each round. This is the exact shape observed in the field, from
// the rundown segment of the no-code combo.
const observedLoopError = "content response invalid after 3 correction round(s): " +
	`these beats are under the 10-word minimum: "five" (8 words); expand them or fold them into a neighbour; ` +
	`these beats are under the 10-word minimum: "five" (8 words); expand them or fold them into a neighbour; ` +
	"narration totals 61 words but a 55s clip needs 119-246 (aim for 159) — rewrite with fuller sentences"

func TestCompromiseLinesSplitsAndDeduplicates(t *testing.T) {
	got := compromiseLines(fmt.Errorf("%s", observedLoopError))
	if len(got) == 0 {
		t.Fatal("no compromises were recorded from a failing loop")
	}
	// The duplicated round must collapse: stored verbatim it is three copies of
	// one complaint, which is what made the blob unreadable.
	joined := strings.Join(got, "\n")
	if n := strings.Count(joined, `"five" (8 words)`); n != 1 {
		t.Errorf("the repeated rule appears %d times, want 1:\n%s", n, joined)
	}
	// The distinct complaints must both survive.
	if !strings.Contains(joined, "10-word minimum") {
		t.Error("the per-beat complaint was lost")
	}
	if !strings.Contains(joined, "narration totals 61 words") {
		t.Error("the total-words complaint was lost")
	}
	// The loop's own framing is about the machinery, not the clip.
	if strings.Contains(joined, "correction round(s)") {
		t.Errorf("the loop's framing leaked into the record:\n%s", joined)
	}
	if strings.Contains(joined, "content response invalid") {
		t.Errorf("the transport error leaked into the record:\n%s", joined)
	}
}

func TestCompromiseLinesOnNoError(t *testing.T) {
	if got := compromiseLines(nil); got != nil {
		t.Errorf("compromiseLines(nil) = %v", got)
	}
}

// The record has to survive the round trip to disk, or it is a log line with
// extra steps. Both plan paths writeJSON the plan, so this is what makes the
// record durable — and the gate marshals the plan through JSON too.
func TestCompromisesSurviveTheArtifact(t *testing.T) {
	plan := &SnippetPlan{
		Template:    "rundown",
		Title:       "Four mindsets",
		Beats:       []SnippetBeat{{ID: "a", Heading: "One", Narration: "words here"}},
		Compromises: []string{`beats under the 10-word minimum: "five" (8 words)`},
	}
	raw, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "compromises") {
		t.Fatal("compromises are not serialised, so nothing is recorded on disk")
	}
	var back SnippetPlan
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if len(back.Compromises) != 1 || back.Compromises[0] != plan.Compromises[0] {
		t.Errorf("compromises did not survive the round trip: %v", back.Compromises)
	}
}

// A plan that satisfied its rules must carry no record at all — a marker on
// everything is a marker on nothing.
func TestACleanPlanRecordsNoCompromise(t *testing.T) {
	plan := &SnippetPlan{Template: "myth", Title: "A clean plan"}
	raw, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "compromises") {
		t.Errorf("a clean plan serialised a compromises key: %s", raw)
	}
}
