package pipeline

// The footage template: a real recording, narrated.
//
// Every other template in the catalog draws something. This one draws nothing —
// it puts a recording of somebody else's product on the stage and talks over
// it. That is the whole idea, and it is why the no-code surface exists: a
// viewer who has never seen Lovable's prompt box will not believe a rectangle
// we drew that says "Lovable", and will believe a recording of it.
//
// == The rule that earns it ==
//
// **A beat may only refer to a mark the recording actually has.** The clip's
// footage.json lists the moments the recorder stamped — `prompt-typed`,
// `app-built` — and a beat naming anything else is describing a moment that
// does not exist. That is the failure this template is most exposed to, because
// the writer is narrating something it cannot see: it has the mark names and
// the durations and nothing else, so an unchecked prompt will happily invent
// "and then the tests pass".
//
// The check is possible precisely because marks are measured rather than
// described. Everything else in the catalog validates shape; this one validates
// against evidence on disk.
//
// == Why it has no artwork of its own ==
//
// The frame is the clip, inside chrome that says where it came from, with the
// capture credit stating the tool, its observed version, and — when the waiting
// was condensed — both durations. There is nothing to lay out and nothing to
// theme, which is the point: the design work here went into making the real
// thing sit properly on the stage rather than into replacing it.

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "footage",
		Category:    CatCode,
		Since:       SinceV3,
		Title:       "A real recording",
		Description: "An actual recording of a tool running, narrated over. Reach for it when the claim is that something really happens — an app building itself, an agent editing files, a deploy going green — and a drawn version would only be an assertion.",
		Example:     "one sentence into Lovable, and the app that comes out",
		PromptFile:  snippetFootageTemplateName,
		NeedsCode:   false,
		// A recording needs room to breathe: an opening, the moments, and a
		// closing read. Below about thirty seconds there is no point cutting to
		// footage at all.
		MinTargetSec:     30,
		DefaultTargetSec: 60,
		MaxBeats:         10,
		Owns:             beatFields{Footage: true},
		OwnsPlan:         planFields{Footage: true},
		Plan:             planFootage,
		NoSalvage:        true,
		PreValidate:      preValidateFootage,
		Normalize:        normalizeFootagePlan,
		Validate:         validateFootagePlan,
		Scenes:           footageScenes,
		PromptData: func(spec SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Marks":   strings.Join(spec.FootageMarks, ", "),
				"Tool":    spec.FootageTool,
				"RealSec": (spec.FootageMs + 500) / 1000,
			}
		},
	})
}

// planFootage refuses to plan until a recording has been resolved.
//
// This template is the only one in the catalog that cannot be filled from a
// prompt alone: its frame IS a clip, and without one there is nothing to put on
// the stage. The failure that guard exists to prevent is silent — planning
// succeeds, the scene renders an empty browser window reading "capture
// unavailable", and the run reports done. A clip that is quietly empty is worse
// than a build that stopped, because nothing downstream can tell.
//
// The resolution itself — capture id → clip path, duration, marks — belongs to
// the `nocode` surface, whose segments name their evidence. Until a caller
// supplies it, this says so rather than producing a hollow frame.
func planFootage(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) (*SnippetPlan, error) {
	if len(spec.FootageMarks) == 0 && spec.FootageTool == "" && spec.FootageMs == 0 {
		return nil, fmt.Errorf(
			"the `footage` template narrates a recording, and this request has none attached.\n" +
				"It is not a template you can pick from a prompt: record something first with a [CAPTURE] marker, then reference that clip from a no-code piece (`evidence: {kind: capture, capture: capture-1}`).\n" +
				"For a drawn frame of a tool's UI, `mockup` or `showcase` is what you want.")
	}
	return planSnippetDefault(ctx, e, spec, cfg)
}

// preValidateFootage puts the recording's own facts on the plan before it is
// judged. Never trusted from the model: these are what it is checked against.
func preValidateFootage(spec SnippetSpec, p *SnippetPlan) {
	p.FootageKnownMarks = spec.FootageMarks
	p.FootageMs = spec.FootageMs
	p.FootageToolName = spec.FootageTool
}

// FootagePlan is what the writer decides about a recording.
type FootagePlan struct {
	// Clip is the capture id this piece narrates.
	Clip string `json:"clip"`
}

// FootageBeat ties a beat to a moment in the recording.
type FootageBeat struct {
	// Mark is the moment this beat talks over. Empty means "wherever we are" —
	// legitimate for an opening or closing beat that is not about a specific
	// instant.
	Mark string `json:"mark,omitempty"`
}

// normalizeFootagePlan repairs the mechanical mistakes before validation sees
// them: casing and spacing on a mark name, which is arithmetic-and-spelling
// rather than a claim, so spending a correction round on it teaches nothing.
func normalizeFootagePlan(p *SnippetPlan) {
	for i := range p.Beats {
		f := p.Beats[i].Footage
		if f == nil {
			continue
		}
		f.Mark = strings.ToLower(strings.TrimSpace(f.Mark))
		f.Mark = strings.ReplaceAll(f.Mark, " ", "-")
		f.Mark = strings.ReplaceAll(f.Mark, "_", "-")
	}
}

// validateFootagePlan is the one validator in the catalog that checks a plan
// against evidence on disk rather than against its own shape.
//
// The writer is narrating something it cannot see — it is handed the mark names
// and the durations and nothing else — so the failure to defend against is a
// beat about a moment the recording does not contain. Marks are measured, which
// is what makes the check possible at all.
func validateFootagePlan(p *SnippetPlan) error {
	if p.Footage == nil || strings.TrimSpace(p.Footage.Clip) == "" {
		return fmt.Errorf("the plan names no clip — a footage piece is about one recording, and it has to say which")
	}
	if err := checkNarratesTheRecordedTool(p); err != nil {
		return err
	}
	known := map[string]bool{}
	for _, m := range p.FootageKnownMarks {
		known[m] = true
	}
	used := map[string]bool{}
	anchored := 0
	for i, b := range p.Beats {
		if b.Footage == nil || b.Footage.Mark == "" {
			continue // an opening or closing read, not about an instant
		}
		mark := b.Footage.Mark
		if len(known) > 0 && !known[mark] {
			return fmt.Errorf("beat %d talks over a moment called %q, which this recording does not contain. It has: %s. Do not describe moments the recording does not have — you cannot see it, and the marks are the only record of what is in it",
				i+1, mark, strings.Join(sortedKeys(known), ", "))
		}
		if used[mark] {
			return fmt.Errorf("beat %d returns to %q, which an earlier beat already covered. A recording runs forwards", i+1, mark)
		}
		used[mark] = true
		anchored++
	}
	if anchored == 0 && len(known) > 0 {
		return fmt.Errorf("no beat is anchored to a moment in the recording, so nothing tells the cut where to be. Anchor at least one beat to one of: %s",
			strings.Join(sortedKeys(known), ", "))
	}
	return nil
}

// checkNarratesTheRecordedTool is the gate the whole surface rests on: the
// narration must be about the tool in the clip.
//
// This is not theoretical. The first real piece recorded a Claude Code session
// and narrated four beats about "the Vercel Agent" — because a *different*
// segment's evidence mentioned Vercel and the fact sheet carried it across. The
// video showed one tool while the voice described another, and every other check
// passed: the beats were the right length, the plan had a clip, the marks were
// fine. A course whose entire claim is "the tool really did that" had produced
// a clip where it demonstrably did not.
//
// Two rules, both cheap:
//
//   - the recorded tool must be named at least once. A piece that never says
//     what you are watching is not narrating the recording.
//   - no other recordable tool may be named more often than it. Mentioning a
//     neighbour is fine — "unlike Cursor, this…" — being *about* one is not.
func checkNarratesTheRecordedTool(p *SnippetPlan) error {
	recorded := strings.TrimSpace(p.FootageToolName)
	if recorded == "" {
		return nil // nothing resolved the tool; the mark check still applies
	}
	var text strings.Builder
	for _, b := range p.Beats {
		text.WriteString(strings.ToLower(b.Narration))
		text.WriteByte(' ')
	}
	body := text.String()

	mine := countBrand(body, recorded)
	if mine == 0 {
		return fmt.Errorf("the recording is of %s and the narration never names it. Say what the viewer is watching — a piece that never names the tool on screen is not narrating the recording",
			recorded)
	}
	for _, other := range allRecordableDisplayNames() {
		if strings.EqualFold(other, recorded) {
			continue
		}
		if brandToken(other) == brandToken(recorded) {
			continue // "Vercel CLI" and a hypothetical "Vercel" are one brand
		}
		if n := countBrand(body, other); n > mine {
			return fmt.Errorf("the recording is of %s, but the narration is about %s (named %d times against %d). Write about what is on screen — this clip shows %s and nothing else",
				recorded, other, n, mine, recorded)
		}
	}
	return nil
}

// brandToken is the word people actually say for a tool.
//
// Display names carry a suffix nobody uses in a sentence — the failing run
// narrated "the Vercel Agent", which contains the brand and never the
// registered display name "Vercel CLI", so matching on the full name found
// nothing and the gate stayed silent on exactly the case it exists for.
func brandToken(display string) string {
	fields := strings.Fields(strings.ToLower(display))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// countBrand counts whole-word mentions of a tool's brand.
func countBrand(body, display string) int {
	tok := brandToken(display)
	if tok == "" {
		return 0
	}
	re, err := regexp.Compile(`\b` + regexp.QuoteMeta(tok) + `\b`)
	if err != nil {
		return 0
	}
	return len(re.FindAllString(body, -1))
}

// allRecordableDisplayNames is every tool a clip could be of, by display name.
func allRecordableDisplayNames() []string {
	var out []string
	for _, k := range captureToolKeys() {
		out = append(out, captureTools[k].Display)
	}
	for _, k := range captureSiteKeys() {
		out = append(out, captureSites[k].Display)
	}
	for _, k := range captureAppKeys() {
		out = append(out, captureApps[k].Display)
	}
	return out
}

// footageScenes puts the recording on the stage for the whole clip.
//
// One scene rather than one per beat: the recording is continuous and cutting
// it into per-beat scenes would restart the video at every sentence. The scene
// spans the piece and the pacing plan — computed in Go, written into the scene
// graph — decides how the clip's own length is fitted into it.
func footageScenes(in SnippetSceneInput) ([]Scene, error) {
	if in.Plan.Footage == nil {
		return nil, fmt.Errorf("footage plan missing")
	}
	// The capture id, from the engine rather than from the plan.
	//
	// Two consumers look the recording up by this id — PlanTerminalPacing, which
	// compresses the dead air, and applyCaptureProvenance, which puts the tool
	// and version on screen — and both did `loadFootageFor(l, id)` against
	// whatever the model happened to write in `footage.clip`. One plan said
	// "Claude Code recording"; no sidecar has that name, so both lookups missed.
	// Provenance silently vanished, and pacing was skipped altogether, which put
	// a 161-second recording into a 61-second slot uncompressed: the clip played
	// its opening seconds and cut away before anything happened in it. The
	// segment's own id is what the capture was recorded under (see
	// NoCodeSpec.CaptureIDs) and it is not the model's to choose.
	clipID := in.Spec.ID
	props := map[string]any{
		"src":        in.Plan.FootageSrc,
		"durationMs": in.Plan.FootageMs,
		"title":      in.Plan.FootageTitle,
		"origin":     in.Plan.FootageOrigin,
		"clipId":     clipID,
		"provClipId": clipID,
	}
	kind := SceneFootage
	if in.Plan.FootageIsTerminal {
		kind = SceneTerminal
	}
	// The span comes from this segment's own beats, never from in.DurationMs.
	//
	// DurationMs is the whole finished piece, which for a standalone snippet is
	// the same thing and for a multi-segment piece is not remotely. Reading it
	// here made the clip claim the entire timeline: on a three-segment no-code
	// piece the recording ran 0→226s while the segments that follow it drew
	// themselves at 61s and 141s, inside the terminal window, over the top of
	// the still-playing footage. Every other template takes its span from its
	// beats (see workspaceScenes, the closest analogue) — this one now does too.
	_, clipStart, _ := in.Beat(0)
	_, _, lastEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    kind,
		StartMs: clipStart,
		EndMs:   lastEnd,
		Props:   props,
	}}, nil
}
