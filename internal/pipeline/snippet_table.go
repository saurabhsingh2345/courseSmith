package pipeline

// The table template: a spec sheet, and the row everybody skips.
//
// The boundary with `anatomy` is what the subject is made of. Anatomy holds one
// STRING still — a signature, a URL, a command line — and lights literal
// substrings of it; its parts are spans of text and its subject is capped at a
// single line. A spec sheet is not a string. It is labelled rows, and the move
// this template makes cannot be expressed as a substring of anything.
//
// The move is burial. A manufacturer's table is a rhetorical object: the
// impressive numbers go at the top, and the one that decides whether the thing
// is any use for your purpose sits fifth, in the same weight as the rest, where
// nobody reads it. The clip reproduces the sheet honestly, then takes the
// weighting away — everything recedes except the row that mattered all along.
// The reference clips do this exactly once and it is the most quoted frame in
// them.
//
// Three rules earn it its place, and all three are validators.
//
// The whole sheet is shown evenly weighted first. The burial only reads if the
// viewer has seen the row sitting there looking ordinary — a sheet that opens
// with one row already lit has drawn a highlight, not a burial.
//
// **The buried row cannot be first or last.** This is the rule the template
// exists for. A number at the top of a spec sheet is the headline, and one at the
// bottom is the summary; neither is buried, and a clip claiming otherwise is
// making a rhetorical point its own picture contradicts. If the number really is
// at the top, it is not being hidden and `metric` states it better.
//
// There have to be enough rows to bury something in. Three rows is a list with a
// highlight; past seven the type is small and the sheet stops looking like a
// sheet and starts looking like a table nobody would print.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:             "table",
		Category:         CatSystems,
		Since:            SinceV4,
		Family:           FamilyReplica,
		Title:            "The row they bury",
		Description:      "A spec sheet shown straight, then stripped back to the one row that actually decides it. Reach for it when the number that matters is sitting in plain sight where nobody reads it.",
		Example:          "The line on a GPU spec sheet that decides whether a model fits",
		PromptFile:       snippetTableTemplateName,
		NeedsCode:        false,
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		MaxBeats:         7,
		Owns:             beatFields{Table: true},
		OwnsPlan:         planFields{Table: true},
		Normalize:        normalizeTablePlan,
		Validate:         validateTablePlan,
		Scenes:           tableScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(TableShows(), ", "),
				"MinRows":       minTableRows,
				"MaxRows":       maxTableRows,
				"MaxLabelWords": maxTableLabelWords,
				"MaxValueChars": maxTableValueChars,
				"MaxNoteWords":  maxTableNoteWords,
			}
		},
	})
}

const snippetTableTemplateName = "snippet_table.tmpl"

const (
	// Four is the floor: with three rows, one of which cannot be first or last,
	// there is exactly one legal position and the burial is not a choice. Past
	// seven the sheet stops looking like a sheet.
	minTableRows = 4
	maxTableRows = 7

	maxTableLabelWords = 4
	maxTableValueChars = 18
	maxTableNoteWords  = 18
)

// tableShows is the closed vocabulary of what a beat does.
var tableShows = map[string]bool{
	// The whole sheet, every row weighted the same. The first beat, always.
	"sheet": true,
	// Everything recedes except the row that mattered.
	"focus": true,
	// Hold the stripped-back sheet and say what the row decides.
	"read": true,
}

// TableShows returns the beat vocabulary sorted.
func TableShows() []string {
	out := make([]string, 0, len(tableShows))
	for k := range tableShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TableSpec is the sheet and the row that matters. On the plan because the sheet
// is one object every beat looks at.
type TableSpec struct {
	// Source names the sheet — "RTX 5090, from the product page". Optional, and
	// worth having: a spec sheet with an attribution is a document, and a
	// document is what the clip is arguing with.
	Source string `json:"source,omitempty"`
	// Rows are the sheet's lines, in the order they appear on it. That order is
	// the subject: it is where the burial happens.
	Rows []TableRow `json:"rows"`
	// At indexes Rows — the row that actually decides it.
	At int `json:"at"`
	// Note is what that row decides. One sentence, and the payload of the clip.
	Note string `json:"note"`
	// Role is what the buried row is doing: a metricRoles name.
	Role string `json:"role,omitempty"`
}

// TableRow is one line of the sheet.
type TableRow struct {
	// Label is the spec's name — "Memory Capacity", "Boost Clock".
	Label string `json:"label"`
	// Value is what the sheet says, exactly as printed — "24 GB", "1,792 GB/s".
	Value string `json:"value"`
}

// ResolvedRole returns the buried row's role, defaulting to the limit — the
// number a sheet buries is almost always the one that stops you.
func (s *TableSpec) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(s.Role))
	if metricRoles[r] {
		return r
	}
	return "limit"
}

// TableBeat is one move.
type TableBeat struct {
	// Show is a tableShows name.
	Show string `json:"show"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to the sheet.
func (b TableBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if tableShows[s] {
		return s
	}
	return "sheet"
}

func normalizeTablePlan(p *SnippetPlan) {
	t := p.Table
	if t == nil {
		return
	}
	t.Source = clampWords(collapseSpaces(t.Source), maxTableLabelWords+3)
	t.Note = clampWords(collapseSpaces(t.Note), maxTableNoteWords)
	t.Role = t.ResolvedRole()

	rows := make([]TableRow, 0, len(t.Rows))
	for _, r := range t.Rows {
		r.Label = clampWords(collapseSpaces(r.Label), maxTableLabelWords)
		// Values are cut by characters: "1,792 GB/s" clamped to a word count is
		// not a spec sheet value.
		r.Value = clampChars(collapseSpaces(r.Value), maxTableValueChars)
		if r.Label != "" && r.Value != "" && len(rows) < maxTableRows {
			rows = append(rows, r)
		}
	}
	t.Rows = rows

	if t.At < 0 {
		t.At = 0
	}
	if n := len(t.Rows); n > 0 && t.At >= n {
		t.At = n - 1
	}

	for i := range p.Beats {
		b := p.Beats[i].Table
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
	}
}

func validateTablePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Table: true}); err != nil {
		return err
	}

	t := p.Table
	if t == nil {
		return fmt.Errorf("the plan has no sheet — this template is a spec table and the row in it that actually decides things")
	}
	if n := len(t.Rows); n < minTableRows || n > maxTableRows {
		return fmt.Errorf("the sheet has %d rows, want %d-%d. With three rows, one of which cannot be first or last, there is exactly one legal place to bury something and the burial stops being a choice; past seven it stops looking like a sheet anybody would print",
			n, minTableRows, maxTableRows)
	}

	seen := map[string]bool{}
	for i, r := range t.Rows {
		if strings.TrimSpace(r.Label) == "" {
			return fmt.Errorf("row %d has no label", i)
		}
		if strings.TrimSpace(r.Value) == "" {
			return fmt.Errorf("row %d (%q) has no value — a spec sheet line with nothing in the right-hand column is not a spec sheet line", i, r.Label)
		}
		key := strings.ToLower(strings.TrimSpace(r.Label))
		if seen[key] {
			return fmt.Errorf("two rows are both %q — a sheet does not list the same spec twice", r.Label)
		}
		seen[key] = true
	}

	if t.At < 0 || t.At >= len(t.Rows) {
		return fmt.Errorf("the buried row is %d, which is not on a sheet of %d rows", t.At, len(t.Rows))
	}
	// The rule the template exists for.
	if t.At == 0 || t.At == len(t.Rows)-1 {
		where := "top"
		if t.At != 0 {
			where = "bottom"
		}
		return fmt.Errorf("the row that matters (%q) is at the %s of the sheet, so it is not buried — a number at the top is the headline and one at the bottom is the summary. This template's whole claim is that the deciding number sits in the middle where nobody reads it; if it really is at the %s, it is not being hidden and the metric template states it better",
			t.Rows[t.At].Label, where, where)
	}
	if strings.TrimSpace(t.Note) == "" {
		return fmt.Errorf("the buried row (%q) has no note. Lighting a row is the gesture; saying what it decides is the content, and without it the clip has pointed at a number and stopped",
			t.Rows[t.At].Label)
	}
	if r := strings.ToLower(strings.TrimSpace(t.Role)); r != "" && !metricRoles[r] {
		return fmt.Errorf("the buried row has role %q, which is not one of: %s", t.Role, strings.Join(MetricRoles(), ", "))
	}

	// The sheet is shown straight before it is stripped back.
	if p.Beats[0].Table == nil || p.Beats[0].Table.ResolvedShow() != "sheet" {
		return fmt.Errorf("beat %q does not open on the whole sheet. The burial only reads if the viewer has seen that row sitting there looking ordinary — a sheet that opens with one row already lit has drawn a highlight rather than a burial",
			p.Beats[0].ID)
	}

	counts := map[string]int{}
	order := make([]string, 0, len(p.Beats))
	for _, b := range p.Beats {
		if b.Table == nil {
			return fmt.Errorf("beat %q has no table direction — every beat shows the sheet, focuses the row, or reads what it decides", b.ID)
		}
		show := b.Table.ResolvedShow()
		counts[show]++
		order = append(order, show)
	}
	if counts["sheet"] != 1 {
		return fmt.Errorf("there are %d beats showing the whole sheet, want exactly 1. Once the weighting is gone it does not come back — putting the sheet back would undo the only move the clip makes", counts["sheet"])
	}
	if counts["focus"] != 1 {
		return fmt.Errorf("there are %d beats focusing the row, want exactly 1", counts["focus"])
	}
	if indexOfShow(order, "sheet") > indexOfShow(order, "focus") {
		return fmt.Errorf("the sheet is shown after the row is focused. The order is the argument: an ordinary-looking sheet, and then the row that was never ordinary")
	}
	return nil
}

// tableScenes lays the clip out as ONE scene. The sheet persists and the beats
// only change the weighting on it.
func tableScenes(in SnippetSceneInput) ([]Scene, error) {
	t := in.Plan.Table
	if t == nil {
		return nil, fmt.Errorf("the plan has no sheet")
	}

	rows := make([]map[string]any, len(t.Rows))
	for i, r := range t.Rows {
		rows[i] = map[string]any{"label": r.Label, "value": r.Value}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Table == nil {
			return nil, fmt.Errorf("beat %q has no table direction", beat.ID)
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Table.ResolvedShow(),
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneTable,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"source": t.Source,
			"rows":   rows,
			"at":     t.At,
			"note":   t.Note,
			"role":   t.ResolvedRole(),
			"steps":  steps,
		}),
	}}, nil
}
