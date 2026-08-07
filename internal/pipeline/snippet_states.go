package pipeline

// The states template: nodes, labelled transitions, and one token that moves.
//
// A state machine is the shape a foundations course keeps drawing without ever
// naming: the five states of a process, the TCP handshake and teardown, a
// connection pool entry, a promise, a lock. Beginners are handed the circles
// and the arrows and take away a picture of a place they could be in ANY of at
// once, because a static graph makes no claim about where you are. The claim
// is the whole point: you are in exactly one state, and you only leave it by
// an event that has a name.
//
// So the picture has a token, and the token is why this template exists rather
// than a generic node diagram. The graph is drawn dim; a bright dot sits on
// one node; a beat fires one arc and the dot SLIDES along it while the event
// caption shows. The clip is a walk, not a map, and the closer leaves the path
// lit so the route the token took is the thing the viewer screenshots.
//
// Which makes the token the thing the validator has to protect. A model
// writing transitions from memory will happily fire "waiting to ready" while
// the token is standing on "running" — the arc exists, the sentence is true in
// general, and the diagram is a lie in particular, because the dot has to jump
// a gap with no arrow across it. So the validator WALKS the token: it starts
// where the opening beat puts it, every fired arc must START where the token
// currently is, and a fire that does not is rejected with both the arc and the
// token's real position quoted. A state machine whose token teleports is a
// diagram telling a lie, and it is a lie in exactly the place a beginner is
// least equipped to catch. The endpoints are resolved against the node ids for
// the same reason: an arrow to a state that is not on screen has nowhere to go.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "states",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The state machine",
		Description: "A handful of states, the labelled events between them, and one bright token walking the graph beat by beat. Reach for it when the subject is a thing that is always in exactly one condition and changes only on a named event — process states, TCP, a connection's life.",
		Example:     "The five states of a process, and what moves it between them",
		PromptFile:  snippetStatesTemplateName,
		NeedsCode:   false,
		// The graph, three or four transitions, a dwell, the route: under
		// thirty-five seconds the token arrives somewhere new before the viewer
		// has read the event that moved it.
		MinTargetSec:     35,
		DefaultTargetSec: 55,
		// Opener + up to six moves and dwells + closer. Past nine the walk is
		// longer than a viewer can hold the route in their head.
		MaxBeats: 9,
		// A beat here is a SHOT — one slide of one dot — not a step in an
		// argument. Twenty-eight words is about nine seconds, which is as long
		// as one transition holds anybody.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{States: true},
		OwnsPlan:          planFields{States: true},
		Normalize:         normalizeStatesPlan,
		Validate:          validateStatesPlan,
		Scenes:            statesScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(StatesShows(), ", "),
				"MinNodes":      minStatesNodes,
				"MaxNodes":      maxStatesNodes,
				"MinArcs":       minStatesArcs,
				"MaxArcs":       maxStatesArcs,
				"MaxLabelWords": maxStatesLabelWords,
				"MaxEventWords": maxStatesEventWords,
			}
		},
	})
}

const snippetStatesTemplateName = "snippet_states.tmpl"

const (
	// Two states and one arrow is a toggle, not a machine — the idea only
	// starts to bite when there is somewhere else the token could have gone.
	minStatesNodes = 3
	// Past six the pills no longer fit two rows at a size their labels can be
	// read at, and the arcs start crossing each other more than they connect.
	maxStatesNodes = 6

	// One arc is a straight line, not a graph.
	minStatesArcs = 2
	// Past eight the curves overlap enough that following one with your eye
	// stops being possible, which is the only thing the picture is for.
	maxStatesArcs = 8

	// A state name sits inside a pill: "ready", "waiting for io".
	maxStatesLabelWords = 3
	// The event rides on the arc as a small caption. Six words is an event;
	// more is a sentence draped over a curve.
	maxStatesEventWords = 6
)

// statesShows is the closed vocabulary of what a beat does.
var statesShows = map[string]bool{
	// The whole graph, dim, with the token placed on node At. The opener.
	"machine": true,
	// Arc At fires: it lights, the token slides along it, the event shows.
	"fire": true,
	// A dwell on node At — what being in this state actually means.
	"state": true,
	// The route the token took, left lit. The closer.
	"walk": true,
}

// StatesShows returns the beat vocabulary sorted.
func StatesShows() []string {
	out := make([]string, 0, len(statesShows))
	for k := range statesShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// StatesSpec is the graph. On the plan because the same machine is on screen
// for the whole clip; only the token moves.
type StatesSpec struct {
	// Nodes are the states, drawn in this order.
	Nodes []StatesNode `json:"nodes"`
	// Arcs are the labelled transitions between them.
	Arcs []StatesArc `json:"arcs"`
}

// StatesNode is one state.
type StatesNode struct {
	// ID is a slug the arcs refer to — "ready".
	ID string `json:"id"`
	// Label is what the pill reads — "ready".
	Label string `json:"label"`
}

// StatesArc is one labelled transition.
type StatesArc struct {
	// From is the id of the state this arc leaves.
	From string `json:"from"`
	// To is the id of the state it arrives at.
	To string `json:"to"`
	// On is the event that fires it — "the scheduler picks it".
	On string `json:"on"`
}

// StatesBeat is one shot. At means a NODE index for "machine" and "state", and
// an ARC index for "fire" — each show has exactly one thing it can point at.
type StatesBeat struct {
	// Show is a statesShows name.
	Show string `json:"show"`
	// At indexes Nodes for "machine"/"state", and Arcs for "fire".
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a fire —
// the workhorse state most beats of this template are in.
func (b StatesBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if statesShows[s] {
		return s
	}
	return "fire"
}

// statesNodeIndex maps node ids to their position, so the validator and the
// scene builder resolve endpoints exactly once and the same way.
func statesNodeIndex(nodes []StatesNode) map[string]int {
	byID := make(map[string]int, len(nodes))
	for i, n := range nodes {
		if _, seen := byID[n.ID]; !seen {
			byID[n.ID] = i
		}
	}
	return byID
}

func normalizeStatesPlan(p *SnippetPlan) {
	s := p.States
	if s == nil {
		return
	}
	nodes := make([]StatesNode, 0, len(s.Nodes))
	for _, n := range s.Nodes {
		n.ID = slugify(n.ID)
		n.Label = clampWords(collapseSpaces(n.Label), maxStatesLabelWords)
		if n.ID == "" {
			n.ID = slugify(n.Label)
		}
		if len(nodes) < maxStatesNodes {
			nodes = append(nodes, n)
		}
	}
	s.Nodes = nodes

	arcs := make([]StatesArc, 0, len(s.Arcs))
	for _, a := range s.Arcs {
		a.From = slugify(a.From)
		a.To = slugify(a.To)
		a.On = clampWords(collapseSpaces(a.On), maxStatesEventWords)
		if len(arcs) < maxStatesArcs {
			arcs = append(arcs, a)
		}
	}
	s.Arcs = arcs

	for i := range p.Beats {
		b := p.Beats[i].States
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		limit := len(s.Nodes)
		if b.Show == "fire" {
			limit = len(s.Arcs)
		} else if b.Show == "walk" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if limit > 0 && b.At >= limit {
			b.At = limit - 1
		}
	}
}

func validateStatesPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{States: true}); err != nil {
		return err
	}

	s := p.States
	if s == nil {
		return fmt.Errorf("the plan has no state machine — this template is a graph with a token walking it, so the graph is the clip")
	}
	if n := len(s.Nodes); n < minStatesNodes || n > maxStatesNodes {
		return fmt.Errorf("the machine has %d state(s), want %d-%d. Two states and one arrow is a toggle rather than a machine — the idea only bites when there is somewhere else the token could have gone — and past %d the pills stop fitting two rows at a readable size",
			n, minStatesNodes, maxStatesNodes, maxStatesNodes)
	}
	seenID := map[string]bool{}
	for i, n := range s.Nodes {
		if strings.TrimSpace(n.ID) == "" {
			return fmt.Errorf("state %d has no id. The arcs refer to states by id, so a state without one cannot be an endpoint of anything", i)
		}
		if strings.TrimSpace(n.Label) == "" {
			return fmt.Errorf("state %d (%q) has no label — an unnamed pill is a circle the token visits for no stated reason", i, n.ID)
		}
		if seenID[n.ID] {
			return fmt.Errorf("state %d repeats the id %q. Ids are how an arc names its endpoints, so two states sharing one means every arrow to %q points at both of them",
				i, n.ID, n.ID)
		}
		seenID[n.ID] = true
	}
	if n := len(s.Arcs); n < minStatesArcs || n > maxStatesArcs {
		return fmt.Errorf("the machine has %d transition(s), want %d-%d. One arc is a straight line rather than a graph, and past %d the curves overlap enough that following one with your eye — the only thing the picture is for — stops being possible",
			n, minStatesArcs, maxStatesArcs, maxStatesArcs)
	}
	byID := statesNodeIndex(s.Nodes)
	seenArc := map[string]bool{}
	for i, a := range s.Arcs {
		if _, ok := byID[a.From]; !ok {
			return fmt.Errorf("transition %d leaves %q, which is not a state in this machine. The states are: %s. An arrow out of somewhere that is not on screen has nowhere to start",
				i, a.From, statesIDList(s.Nodes))
		}
		if _, ok := byID[a.To]; !ok {
			return fmt.Errorf("transition %d arrives at %q, which is not a state in this machine. The states are: %s. An arrow to somewhere that is not on screen has nowhere to go",
				i, a.To, statesIDList(s.Nodes))
		}
		if strings.TrimSpace(a.On) == "" {
			return fmt.Errorf("transition %d (%s to %s) has no event. The label IS the lesson: an unlabelled arrow says the token moves for reasons the viewer is not told", i, a.From, a.To)
		}
		key := a.From + "\x00" + a.To + "\x00" + strings.ToLower(a.On)
		if seenArc[key] {
			return fmt.Errorf("transition %d repeats %s to %s on %q. The same arc drawn twice lands on the same curve, so the second one is invisible and only makes the fire beats ambiguous",
				i, a.From, a.To, a.On)
		}
		seenArc[key] = true
	}

	// The shape. The graph is seen whole, with the token placed, before it moves.
	if p.Beats[0].States == nil || p.Beats[0].States.ResolvedShow() != "machine" {
		return fmt.Errorf("beat %q does not open on the whole machine. A dot sliding along a curve nobody has been shown is a moving dot — the first beat is {\"show\": \"machine\", \"at\": N} with N the state the token starts in",
			p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.States == nil || last.States.ResolvedShow() != "walk" {
		return fmt.Errorf("the clip does not close on the route. The final frame is the path the token actually took, left lit — end with {\"show\": \"walk\"}")
	}

	// THE TOKEN WALK. The dot is in exactly one state at a time, and it only
	// leaves by an arc that starts where it is standing.
	token := p.Beats[0].States.At
	if token < 0 || token >= len(s.Nodes) {
		return fmt.Errorf("beat %q starts the token in state %d, which does not exist — the machine holds states 0-%d", p.Beats[0].ID, token, len(s.Nodes)-1)
	}
	for i, b := range p.Beats {
		if b.States == nil {
			return fmt.Errorf("beat %q has no states direction — every beat shows one state of the graph", b.ID)
		}
		switch b.States.ResolvedShow() {
		case "machine":
			if i != 0 {
				return fmt.Errorf("beat %q re-places the token part-way through. The machine beat is the opener; setting the token down again mid-walk erases the route the clip has been building", b.ID)
			}
		case "walk":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q lights the route before the walk is over. \"walk\" is the closer — a route drawn with transitions still to come is the wrong route", b.ID)
			}
		case "state":
			if b.States.At < 0 || b.States.At >= len(s.Nodes) {
				return fmt.Errorf("beat %q dwells on state %d, which does not exist — the machine holds states 0-%d", b.ID, b.States.At, len(s.Nodes)-1)
			}
		case "fire":
			at := b.States.At
			if at < 0 || at >= len(s.Arcs) {
				return fmt.Errorf("beat %q fires transition %d, which does not exist — the machine holds transitions 0-%d", b.ID, at, len(s.Arcs)-1)
			}
			arc := s.Arcs[at]
			if arc.From != s.Nodes[token].ID {
				return fmt.Errorf("beat %q fires the transition %s to %s on %q, but the token is standing on %q. An arc has to START where the token IS: firing this one makes the dot jump a gap with no arrow across it, and a state machine whose token teleports is a diagram telling a lie. Either fire a transition out of %q first, or fire one that leaves %q",
					b.ID, arc.From, arc.To, arc.On, s.Nodes[token].ID, s.Nodes[token].ID, s.Nodes[token].ID)
			}
			token = byID[arc.To]
		}
	}
	return nil
}

// statesIDList renders the machine's state ids for a rejection message.
func statesIDList(nodes []StatesNode) string {
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ID
	}
	return strings.Join(ids, ", ")
}

// statesScenes lays the clip out as ONE scene. The endpoints are resolved to
// indices here and the token's position is walked here, so the component
// never looks up an id or works out where the dot should be.
func statesScenes(in SnippetSceneInput) ([]Scene, error) {
	s := in.Plan.States
	if s == nil {
		return nil, fmt.Errorf("the plan has no state machine")
	}
	byID := statesNodeIndex(s.Nodes)

	nodes := make([]map[string]any, len(s.Nodes))
	for i, n := range s.Nodes {
		nodes[i] = map[string]any{
			"id":    n.ID,
			"label": n.Label,
		}
	}
	arcs := make([]map[string]any, len(s.Arcs))
	for i, a := range s.Arcs {
		from, okFrom := byID[a.From]
		to, okTo := byID[a.To]
		if !okFrom || !okTo {
			return nil, fmt.Errorf("transition %d joins %q to %q, and one of them is not a state", i, a.From, a.To)
		}
		arcs[i] = map[string]any{
			"from": from,
			"to":   to,
			"on":   a.On,
		}
	}

	token := 0
	if b := in.Plan.Beats[0].States; b != nil && b.At >= 0 && b.At < len(s.Nodes) {
		token = b.At
	}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	fired := map[int]bool{}
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.States == nil {
			return nil, fmt.Errorf("beat %q has no states direction", beat.ID)
		}
		show := beat.States.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		// The token's node BEFORE this beat acts, so a fire beat can animate
		// the slide from where it was to where it lands.
		step["from"] = token
		switch show {
		case "state":
			step["at"] = beat.States.At
		case "fire":
			at := beat.States.At
			if at < 0 || at >= len(s.Arcs) {
				return nil, fmt.Errorf("beat %q fires transition %d, which does not exist", beat.ID, at)
			}
			fired[at] = true
			step["at"] = at
			token = byID[s.Arcs[at].To]
		}
		step["token"] = token
		lit := make([]int, 0, len(fired))
		for at := range fired {
			lit = append(lit, at)
		}
		sort.Ints(lit)
		step["lit"] = lit
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneStates,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title": in.Plan.Title,
			"nodes": nodes,
			"arcs":  arcs,
			"steps": steps,
		}),
	}}, nil
}
