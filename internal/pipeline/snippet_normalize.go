package pipeline

// Normalizing a snippet plan: turning what the model said into what the
// template requires, wherever that conversion is mechanical.
//
// A validator and a normalizer answer two different questions. "Is this label
// five words when the box holds four?" is a validation question only if the
// answer changes what the model should write next; if the fix is *cut it to
// four*, there is nothing to tell the model, and a correction round spent
// saying so is a round not spent on the clip. The same goes for a shape named
// "rectangle" instead of "box", an id the model forgot, a link pointing at a
// box drawn two beats later, an explanation list one entry short of its
// options — every one of these used to cost a full generation round trip, and
// every one of them has exactly one sensible repair.
//
// So the correction loop normalizes first and validates second, and the rules
// that survive into validation are the ones only the model can fix: too little
// narration, a question with no answer, a board with three boxes on it. That
// division is what keeps the prompts honest too — a rule worth stating in a
// prompt is a rule worth a round, and anything else belongs here.
//
// Normalization never invents content. It cuts, renames, re-points and drops;
// it does not write narration, and it does not make up a fourth option for a
// quiz that offered three. What it cannot repair, validation still rejects.

import (
	"fmt"
	"reflect"
	"strings"
)

// normalizeSnippetPlan repairs everything repairable about a decoded plan, in
// place. It is safe to call more than once and on a plan built by hand.
func normalizeSnippetPlan(p *SnippetPlan) {
	if p == nil {
		return
	}
	p.Title = collapseSpaces(p.Title)
	p.Subtitle = collapseSpaces(p.Subtitle)
	normalizeSnippetBeats(p)

	tpl := SnippetTemplates[p.Template]
	if tpl != nil {
		migrateBeatFields(p, tpl.Owns)
		stripBeatFields(p, tpl.Owns)
		stripPlanFields(p, tpl.OwnsPlan)
		if tpl.Normalize != nil {
			tpl.Normalize(p)
		}
	}
	if p.Title == "" && len(p.Beats) > 0 {
		p.Title = p.Beats[0].Heading
	}
}

// normalizeSnippetBeats trims the beats, drops the ones with nothing to say,
// and gives every survivor the id and heading the pipeline downstream assumes
// it has — the aligner reports timings per id, so a missing or duplicated one
// is not a style problem.
func normalizeSnippetBeats(p *SnippetPlan) {
	out := make([]SnippetBeat, 0, len(p.Beats))
	seen := map[string]bool{}
	for i, b := range p.Beats {
		b.Narration = collapseSpaces(b.Narration)
		b.Heading = collapseSpaces(b.Heading)
		if b.Narration == "" {
			continue // a beat the voice never speaks has no span to time against
		}
		if b.Heading == "" {
			b.Heading = headingFromNarration(b.Narration)
		}
		base := slugify(b.ID)
		if base == "" {
			base = slugify(b.Heading)
		}
		if base == "" {
			base = fmt.Sprintf("beat-%d", i+1)
		}
		b.ID = base
		for n := 2; seen[b.ID]; n++ {
			b.ID = fmt.Sprintf("%s-%d", base, n)
		}
		seen[b.ID] = true
		out = append(out, b)
	}
	p.Beats = out
}

// planFields names the plan-level payloads a template consumes, the way
// beatFields names the per-beat ones.
type planFields struct {
	Chart     bool
	Project   bool
	Quiz      bool
	Compare   bool
	Anatomy   bool
	Timeline  bool
	Canvas    bool
	Loop      bool
	Mockup    bool
	Stack     bool
	Spec      bool
	Showcase  bool
	Breakdown bool
}

// migrateBeatFields moves a payload the model put under the wrong name onto the
// field this template actually reads.
//
// The confusion it repairs is a real one and not the model being careless: a
// board item and a flow node are the same idea in two vocabularies, and a model
// asked for a whiteboard that answers with `nodes` has understood the clip and
// mislabelled the field. Rejecting that costs a round to be told a synonym.
// Anything with no equivalent under this template is left for stripBeatFields.
func migrateBeatFields(p *SnippetPlan, owns beatFields) {
	// Node ids are referenced by other nodes' `from`, so a board built out of
	// them needs the id→label map before it can re-point the links.
	labelOf := map[string]string{}
	if owns.Sketch {
		for _, b := range p.Beats {
			for _, n := range b.Nodes {
				labelOf[strings.TrimSpace(n.ID)] = strings.TrimSpace(n.Label)
			}
		}
	}
	for i := range p.Beats {
		b := &p.Beats[i]
		switch {
		case owns.Sketch && len(b.Sketch) == 0 && len(b.Nodes) > 0:
			for _, n := range b.Nodes {
				item := SketchItem{Label: n.Label, Icon: flowNodeKinds[n.ResolvedKind()]}
				if len(n.From) > 0 {
					item.LinkFrom = labelOf[strings.TrimSpace(n.From[0])]
				}
				b.Sketch = append(b.Sketch, item)
			}
			b.Nodes = nil
		case owns.Nodes && len(b.Nodes) == 0 && len(b.Sketch) > 0:
			for _, item := range b.Sketch {
				node := FlowNode{ID: slugify(item.Label), Label: item.Label}
				if item.LinkFrom != "" {
					node.From = []string{slugify(item.LinkFrom)}
				}
				b.Nodes = append(b.Nodes, node)
			}
			b.Sketch = nil
		case owns.Art && b.Art == nil && b.Cast != nil:
			b.Art = &ArtBeat{Figure: b.Cast.Prop, Caption: b.Cast.Caption}
			b.Cast = nil
		case owns.Cast && b.Cast == nil && b.Art != nil:
			b.Cast = &CastBeat{Prop: b.Art.Figure, Caption: b.Art.Caption}
			b.Art = nil
		}
	}
}

// stripBeatFields clears what this template does not read.
//
// Left in place it is dead weight in snippet-plan.json and a claim the clip
// never honours; rejected, it is a correction round spent on something the
// renderer was going to ignore anyway.
//
// It works by name rather than by a switch, because a switch is a list that
// gets forgotten: beatFields already names each optional payload exactly as
// SnippetBeat does, so adding a field to both — which is what adding a template
// requires anyway — is enough for it to be stripped here too. The pairing is
// held by TestBeatFieldsNameSnippetBeatFields rather than by remembering.
func stripBeatFields(p *SnippetPlan, owns beatFields) {
	ot, ov := reflect.TypeOf(owns), reflect.ValueOf(owns)
	for i := range p.Beats {
		bv := reflect.ValueOf(&p.Beats[i]).Elem()
		for f := range ot.NumField() {
			if ov.Field(f).Bool() {
				continue
			}
			if fv := bv.FieldByName(ot.Field(f).Name); fv.IsValid() && fv.CanSet() {
				fv.SetZero()
			}
		}
	}
}

func stripPlanFields(p *SnippetPlan, owns planFields) {
	if !owns.Chart {
		p.Chart = nil
	}
	if !owns.Project {
		p.Project = nil
	}
	if !owns.Quiz {
		p.Quiz = nil
	}
	if !owns.Compare {
		p.Compare = nil
	}
	if !owns.Anatomy {
		p.Anatomy = nil
	}
	if !owns.Timeline {
		p.Timeline = nil
	}
	if !owns.Canvas {
		p.Canvas = nil
	}
	if !owns.Loop {
		p.Loop = nil
	}
	if !owns.Mockup {
		p.Mockup = nil
	}
	if !owns.Stack {
		p.Stack = nil
	}
	if !owns.Spec {
		p.Spec = nil
	}
	if !owns.Showcase {
		p.Showcase = nil
	}
	if !owns.Breakdown {
		p.Breakdown = nil
	}
}

// clampWords cuts a phrase to at most n words, keeping the first ones — the
// front of a label is the part that identifies it.
func clampWords(s string, n int) string {
	f := strings.Fields(s)
	if len(f) <= n {
		return strings.Join(f, " ")
	}
	return strings.Join(f[:n], " ")
}

// collapseSpaces trims a string and squeezes its internal whitespace, so a
// heading that arrived with a newline in it still counts its own words.
func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// headingFromNarration makes a label out of a beat that arrived without one.
func headingFromNarration(narration string) string {
	words := strings.Fields(narration)
	if len(words) > 4 {
		words = words[:4]
	}
	s := strings.TrimRight(strings.Join(words, " "), ".,;:!?—-")
	if s == "" {
		return "This beat"
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
