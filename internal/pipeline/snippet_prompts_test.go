package pipeline

import (
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

// Every snippet prompt must render with the data its own template supplies.
//
// Prompts are files read from disk; the data map is compiled into the binary.
// The two are a pair, and nothing until now checked that they agreed —
// prompts_test.go covers the lesson prompts and stopped there. Templates render
// with missingkey=error, so a prompt referencing a key no template provides is
// a hard failure at generation time, after the user has already asked for a
// clip.
//
// That is not hypothetical: adding {{.SuggestBeats}} to six prompts and the key
// to sharedPromptData was one commit, and the only thing standing between a
// half-applied version of it and a broken planner was nobody having written
// this test.
func TestEverySnippetPromptRenders(t *testing.T) {
	cfg := config.Defaults()
	cfg.Style.PaceWPM = 175

	for _, name := range SnippetTemplateNames() {
		t.Run(name, func(t *testing.T) {
			tpl := SnippetTemplates[name]
			spec := SnippetSpec{
				Prompt:   "Why every engineer should understand backpressure",
				Template: name,
			}
			data := sharedPromptData(spec, cfg)
			if tpl.PromptData != nil {
				for k, v := range tpl.PromptData(spec, cfg) {
					data[k] = v
				}
			}
			system, user, err := renderPromptFile(repoPromptsDir, tpl.PromptFile, data)
			if err != nil {
				t.Fatalf("rendering %s: %v", tpl.PromptFile, err)
			}
			if strings.TrimSpace(system) == "" {
				t.Errorf("%s produced an empty system prompt", tpl.PromptFile)
			}
			if !strings.Contains(user, spec.Prompt) {
				t.Errorf("%s did not carry the creator's prompt into the user message", tpl.PromptFile)
			}
			// missingkey=error turns an unknown key into an error, but a key
			// that exists and is empty renders as "" and is silently useless.
			if strings.Contains(system, "<no value>") || strings.Contains(user, "<no value>") {
				t.Errorf("%s rendered a <no value> placeholder", tpl.PromptFile)
			}
		})
	}
}

// The story template's second call builds its own data map rather than going
// through PromptData, so the loop above cannot reach it. It is the prompt most
// likely to drift, being the only one whose inputs are assembled by hand.
func TestStoryShotsPromptRenders(t *testing.T) {
	cfg := config.Defaults()
	cfg.Style.PaceWPM = 175
	spec := SnippetSpec{Prompt: "How an index finds your row", Template: "story"}

	data := sharedPromptData(spec, cfg)
	data["StoryTitle"] = "How an index finds your row"
	data["Logline"] = "An index is a sorted copy of one column."
	data["Beats"] = "1. id=the-slow-query  heading=\"Your query reads every row\"\n   ...\n"
	data["BeatCount"] = 8
	data["Stagings"] = strings.Join(StoryStagingNames(), ", ")
	data["Cameras"] = strings.Join(StoryCameraNames(), ", ")
	data["Poses"] = strings.Join(CastPoseNames(), ", ")
	data["Expressions"] = strings.Join(CastExpressionNames(), ", ")
	data["Props"] = strings.Join(ArtFigureNames(), ", ")
	data["MinCharacterShots"] = minCharacterShots(8)

	system, user, err := renderPromptFile(repoPromptsDir, snippetStoryShotsTemplateName, data)
	if err != nil {
		t.Fatalf("rendering %s: %v", snippetStoryShotsTemplateName, err)
	}
	if !strings.Contains(system, "hero") || !strings.Contains(system, "push") {
		t.Error("the director prompt does not list its own vocabularies")
	}
	if !strings.Contains(user, "the-slow-query") {
		t.Error("the director prompt did not carry the script into the user message")
	}
}
