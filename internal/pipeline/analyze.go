package pipeline

// Course-level intelligence (coursesmith analyze): extracts the concepts
// each lesson introduces and uses, builds the course-wide concept DAG,
// flags dependency violations (a concept used before it is taught is a
// pipeline ERROR), checks terminology consistency, and scores the
// narrative bridge between consecutive lessons.
//
// Outputs land in courses/<slug>/generated/: concepts.json, concepts.svg,
// analysis.json. Per-lesson extractions are cached in each lesson's
// generated/concepts.json keyed by a source hash, so re-running analyze
// only re-extracts changed lessons.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// Analysis artifact names.
const (
	ConceptsFileName    = "concepts.json" // per-lesson cache AND course-level graph
	ConceptsSVGFileName = "concepts.svg"
	AnalysisFileName    = "analysis.json"
)

// CourseGeneratedDirName is the course-level output dir (sibling of lessons/).
const CourseGeneratedDirName = "generated"

// ConceptRef is one occurrence of a concept in a lesson.
type ConceptRef struct {
	Name    string `json:"name"`    // canonical lowercase concept name
	Term    string `json:"term"`    // the exact term the lesson used
	Section string `json:"section"` // section id / heading slug
	Quote   string `json:"quote"`   // short verbatim fragment
}

// LessonConcepts is the cached per-lesson extraction.
type LessonConcepts struct {
	LessonID    string       `json:"lesson_id"`
	SourceHash  string       `json:"source_hash"`
	Introduced  []ConceptRef `json:"introduced"`
	Used        []ConceptRef `json:"used"`
	ExtractedAt time.Time    `json:"extracted_at"`
}

// conceptExtraction is the LLM response shape.
type conceptExtraction struct {
	Introduced []ConceptRef `json:"introduced"`
	Used       []ConceptRef `json:"used"`
}

func (c *conceptExtraction) Validate() error {
	check := func(kind string, refs []ConceptRef) error {
		for i, r := range refs {
			if strings.TrimSpace(r.Name) == "" {
				return fmt.Errorf("%s[%d] has an empty name", kind, i)
			}
		}
		return nil
	}
	if err := check("introduced", c.Introduced); err != nil {
		return err
	}
	return check("used", c.Used)
}

// ConceptNode is one node of the course-wide concept graph.
type ConceptNode struct {
	Name         string   `json:"name"`
	IntroducedIn string   `json:"introduced_in"` // lesson id, "" if never taught
	RequiredBy   []string `json:"required_by"`   // lesson ids using the concept
}

// DependencyViolation is a concept used before (or without) being taught.
type DependencyViolation struct {
	Concept      string `json:"concept"`
	UsedIn       string `json:"used_in"`
	Section      string `json:"section"`
	Quote        string `json:"quote"`
	IntroducedIn string `json:"introduced_in,omitempty"` // later lesson, or "" if never
}

func (v DependencyViolation) String() string {
	where := "never introduced anywhere in the course"
	if v.IntroducedIn != "" {
		where = "only introduced later, in " + v.IntroducedIn
	}
	return fmt.Sprintf("%s (%s/%s: %q) — %s", v.Concept, v.UsedIn, v.Section, v.Quote, where)
}

// TerminologyIssue is one drift finding.
type TerminologyIssue struct {
	Concept   string   `json:"concept"`
	Variants  []string `json:"variants"`
	Canonical string   `json:"canonical"`
	Reason    string   `json:"reason"`
}

type terminologyReview struct {
	Issues []TerminologyIssue `json:"issues"`
}

// BridgeScore is the narrative check between two consecutive lessons.
type BridgeScore struct {
	From       string  `json:"from"`
	To         string  `json:"to"`
	Score      float64 `json:"score"`
	Suggestion string  `json:"suggestion"`
}

type bridgeReview struct {
	Score      float64 `json:"score"`
	Suggestion string  `json:"suggestion"`
}

func (b *bridgeReview) Validate() error {
	if b.Score < 1 || b.Score > 10 {
		return fmt.Errorf("score %v is outside 1-10", b.Score)
	}
	return nil
}

// weakBridgeThreshold marks bridges needing a rewrite suggestion surfaced.
const weakBridgeThreshold = 7

// AnalysisReport is the persisted analysis.json.
type AnalysisReport struct {
	Course      string                `json:"course"`
	Lessons     []string              `json:"lessons"`
	Concepts    []ConceptNode         `json:"concepts"`
	Violations  []DependencyViolation `json:"violations"`
	Terminology []TerminologyIssue    `json:"terminology"`
	Bridges     []BridgeScore         `json:"bridges"`
	AnalyzedAt  time.Time             `json:"analyzed_at"`
}

// conceptsPromptData feeds prompts/concepts.tmpl.
type conceptsPromptData struct {
	LessonID  string
	Title     string
	Outline   string
	Narration string
}

// extractLessonConcepts returns the lesson's concept extraction, using the
// cached generated/concepts.json when the sources are unchanged.
func extractLessonConcepts(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config) (*LessonConcepts, error) {
	narration := ""
	if script, err := loadScript(l); err == nil {
		var parts []string
		for _, sec := range script.Sections {
			parts = append(parts, fmt.Sprintf("[%s]\n%s", sec.ID, sec.Narration))
		}
		narration = strings.Join(parts, "\n\n")
	}
	sourceHash := project.HashBytes([]byte(l.Body + "\x00" + narration))

	cachePath := filepath.Join(l.GeneratedDir(), ConceptsFileName)
	if data, err := os.ReadFile(cachePath); err == nil {
		var cached LessonConcepts
		if json.Unmarshal(data, &cached) == nil && cached.SourceHash == sourceHash {
			return &cached, nil
		}
	}

	data := conceptsPromptData{
		LessonID:  l.ID,
		Title:     l.FrontMatter.Title,
		Outline:   l.Body,
		Narration: narration,
	}
	system, user, err := e.renderPrompt(conceptsTemplateName, data)
	if err != nil {
		return nil, err
	}
	var extraction conceptExtraction
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0, 4096, &extraction, extraction.Validate)
	if err != nil {
		return nil, fmt.Errorf("extracting concepts from %s: %w", l.ID, err)
	}

	out := &LessonConcepts{
		LessonID:    l.ID,
		SourceHash:  sourceHash,
		Introduced:  normalizeConceptRefs(extraction.Introduced),
		Used:        normalizeConceptRefs(extraction.Used),
		ExtractedAt: time.Now().UTC(),
	}
	if err := writeJSON(cachePath, out); err != nil {
		return nil, err
	}
	return out, nil
}

func normalizeConceptRefs(refs []ConceptRef) []ConceptRef {
	out := make([]ConceptRef, 0, len(refs))
	for _, r := range refs {
		r.Name = strings.ToLower(strings.TrimSpace(r.Name))
		if r.Name != "" {
			out = append(out, r)
		}
	}
	return out
}

// BuildConceptGraph merges per-lesson extractions (in lesson order) into
// the course graph and finds dependency violations.
func BuildConceptGraph(lessonOrder []string, perLesson map[string]*LessonConcepts) ([]ConceptNode, []DependencyViolation) {
	ordinal := make(map[string]int, len(lessonOrder))
	for i, id := range lessonOrder {
		ordinal[id] = i
	}

	introducedIn := map[string]string{} // concept → first introducing lesson
	for _, id := range lessonOrder {
		lc := perLesson[id]
		if lc == nil {
			continue
		}
		for _, ref := range lc.Introduced {
			if _, seen := introducedIn[ref.Name]; !seen {
				introducedIn[ref.Name] = id
			}
		}
	}

	requiredBy := map[string][]string{}
	var violations []DependencyViolation
	for _, id := range lessonOrder {
		lc := perLesson[id]
		if lc == nil {
			continue
		}
		seenHere := map[string]bool{}
		for _, ref := range lc.Used {
			if !seenHere[ref.Name] {
				requiredBy[ref.Name] = append(requiredBy[ref.Name], id)
				seenHere[ref.Name] = true
			}
			intro, taught := introducedIn[ref.Name]
			if !taught || ordinal[intro] > ordinal[id] {
				v := DependencyViolation{
					Concept: ref.Name,
					UsedIn:  id,
					Section: ref.Section,
					Quote:   ref.Quote,
				}
				if taught {
					v.IntroducedIn = intro
				}
				violations = append(violations, v)
			}
		}
	}

	names := make([]string, 0, len(introducedIn)+len(requiredBy))
	seen := map[string]bool{}
	for name := range introducedIn {
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	for name := range requiredBy {
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	sort.Strings(names)

	nodes := make([]ConceptNode, 0, len(names))
	for _, name := range names {
		nodes = append(nodes, ConceptNode{
			Name:         name,
			IntroducedIn: introducedIn[name],
			RequiredBy:   requiredBy[name],
		})
	}
	return nodes, violations
}

// terminologyPromptData feeds prompts/terminology.tmpl.
type terminologyPromptData struct {
	TermsDump string
}

// checkTerminology asks the review model for term drift across lessons.
func checkTerminology(ctx context.Context, e *Env, cfg config.Config, lessonOrder []string, perLesson map[string]*LessonConcepts) ([]TerminologyIssue, error) {
	terms := map[string]map[string]map[string]bool{} // concept → lesson → set of terms
	record := func(lessonID string, refs []ConceptRef) {
		for _, r := range refs {
			if terms[r.Name] == nil {
				terms[r.Name] = map[string]map[string]bool{}
			}
			if terms[r.Name][lessonID] == nil {
				terms[r.Name][lessonID] = map[string]bool{}
			}
			if t := strings.TrimSpace(r.Term); t != "" {
				terms[r.Name][lessonID][t] = true
			}
		}
	}
	for _, id := range lessonOrder {
		if lc := perLesson[id]; lc != nil {
			record(id, lc.Introduced)
			record(id, lc.Used)
		}
	}

	var names []string
	for name := range terms {
		names = append(names, name)
	}
	sort.Strings(names)
	var dump strings.Builder
	for _, name := range names {
		fmt.Fprintf(&dump, "%s:\n", name)
		for _, id := range lessonOrder {
			if set := terms[name][id]; len(set) > 0 {
				var ts []string
				for t := range set {
					ts = append(ts, t)
				}
				sort.Strings(ts)
				fmt.Fprintf(&dump, "  %s: %s\n", id, strings.Join(ts, ", "))
			}
		}
	}
	if dump.Len() == 0 {
		return nil, nil
	}

	system, user, err := e.renderPrompt(terminologyTemplateName, terminologyPromptData{TermsDump: dump.String()})
	if err != nil {
		return nil, err
	}
	var review terminologyReview
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskReview, system, user, 0, 2048, &review, nil)
	if err != nil {
		return nil, fmt.Errorf("terminology check: %w", err)
	}
	return review.Issues, nil
}

// bridgePromptData feeds prompts/bridge.tmpl.
type bridgePromptData struct {
	PrevID      string
	PrevTitle   string
	PrevClosing string
	NextID      string
	NextTitle   string
	NextOpening string
}

// scoreBridges scores the narrative connection between every consecutive
// lesson pair with a script.
func scoreBridges(ctx context.Context, e *Env, cfg config.Config, lessons []*project.Lesson) ([]BridgeScore, error) {
	closing := func(l *project.Lesson) string {
		if script, err := loadScript(l); err == nil && len(script.Sections) > 0 {
			return script.Sections[len(script.Sections)-1].Narration
		}
		return tailLines(strings.TrimSpace(l.Body), 6)
	}
	opening := func(l *project.Lesson) string {
		if script, err := loadScript(l); err == nil && len(script.Sections) > 0 {
			return script.Sections[0].Narration
		}
		lines := strings.SplitN(strings.TrimSpace(l.Body), "\n", 7)
		return strings.Join(lines[:min(6, len(lines))], "\n")
	}

	var bridges []BridgeScore
	for i := 1; i < len(lessons); i++ {
		prev, next := lessons[i-1], lessons[i]
		data := bridgePromptData{
			PrevID: prev.ID, PrevTitle: prev.FrontMatter.Title, PrevClosing: closing(prev),
			NextID: next.ID, NextTitle: next.FrontMatter.Title, NextOpening: opening(next),
		}
		system, user, err := e.renderPrompt(bridgeTemplateName, data)
		if err != nil {
			return nil, err
		}
		var review bridgeReview
		err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskReview, system, user, 0, 1024, &review, review.Validate)
		if err != nil {
			return nil, fmt.Errorf("bridge %s → %s: %w", prev.ID, next.ID, err)
		}
		bridges = append(bridges, BridgeScore{
			From: prev.ID, To: next.ID, Score: review.Score, Suggestion: review.Suggestion,
		})
	}
	return bridges, nil
}

// AnalyzeCourse runs the full course-level analysis and writes the
// artifacts to courses/<slug>/generated/. Dependency violations make the
// returned error non-nil AFTER all artifacts are written, so the graph is
// always available for inspection.
func AnalyzeCourse(ctx context.Context, e *Env, course *project.Course) (*AnalysisReport, error) {
	cfg := config.Resolve(course.Config, config.Config{}, config.Config{})
	lessons, err := course.Lessons()
	if err != nil {
		return nil, err
	}

	report := &AnalysisReport{Course: course.Slug, AnalyzedAt: time.Now().UTC()}
	perLesson := map[string]*LessonConcepts{}
	for _, l := range lessons {
		report.Lessons = append(report.Lessons, l.ID)
		fmt.Fprintf(e.out(), "  → concepts  %s...\n", l.ID)
		lc, err := extractLessonConcepts(ctx, e, l, cfg)
		if err != nil {
			return nil, err
		}
		perLesson[l.ID] = lc
		fmt.Fprintf(e.out(), "    %d introduced, %d used\n", len(lc.Introduced), len(lc.Used))
	}

	report.Concepts, report.Violations = BuildConceptGraph(report.Lessons, perLesson)

	fmt.Fprintf(e.out(), "  → terminology consistency (%s)...\n", cfg.Pipeline.LLMReview)
	report.Terminology, err = checkTerminology(ctx, e, cfg, report.Lessons, perLesson)
	if err != nil {
		return nil, err
	}

	if len(lessons) > 1 {
		fmt.Fprintf(e.out(), "  → narrative bridges (%d pair(s))...\n", len(lessons)-1)
		report.Bridges, err = scoreBridges(ctx, e, cfg, lessons)
		if err != nil {
			return nil, err
		}
	}

	outDir := filepath.Join(course.Dir, CourseGeneratedDirName)
	if err := writeJSON(filepath.Join(outDir, ConceptsFileName), map[string]any{
		"concepts":   report.Concepts,
		"violations": report.Violations,
	}); err != nil {
		return nil, err
	}
	svg := conceptsSVG(course, report, perLesson)
	if err := writeFileAtomic(filepath.Join(outDir, ConceptsSVGFileName), []byte(svg)); err != nil {
		return nil, err
	}
	if err := writeJSON(filepath.Join(outDir, AnalysisFileName), report); err != nil {
		return nil, err
	}

	printAnalysis(e, report)
	if len(report.Violations) > 0 {
		return report, fmt.Errorf("%d concept dependency violation(s) — see above and %s",
			len(report.Violations), filepath.Join(outDir, AnalysisFileName))
	}
	return report, nil
}

func printAnalysis(e *Env, report *AnalysisReport) {
	fmt.Fprintf(e.out(), "\nconcept graph: %d concept(s) across %d lesson(s)\n", len(report.Concepts), len(report.Lessons))
	if len(report.Violations) == 0 {
		fmt.Fprintf(e.out(), "  ✓ zero dependency violations — nothing is used before it is taught\n")
	}
	for _, v := range report.Violations {
		fmt.Fprintf(e.out(), "  ✗ VIOLATION: %s\n", v)
	}
	if len(report.Terminology) == 0 {
		fmt.Fprintf(e.out(), "  ✓ terminology is consistent\n")
	}
	for _, issue := range report.Terminology {
		fmt.Fprintf(e.out(), "  ⚠ terminology drift: %s (%s) — use %q (%s)\n",
			issue.Concept, strings.Join(issue.Variants, " / "), issue.Canonical, issue.Reason)
	}
	for _, b := range report.Bridges {
		mark := "✓"
		if b.Score < weakBridgeThreshold {
			mark = "⚠"
		}
		fmt.Fprintf(e.out(), "  %s bridge %s → %s: %.1f/10", mark, b.From, b.To, b.Score)
		if b.Score < weakBridgeThreshold && b.Suggestion != "" {
			fmt.Fprintf(e.out(), " — %s", b.Suggestion)
		}
		fmt.Fprintln(e.out())
	}
}
