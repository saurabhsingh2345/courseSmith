package pipeline

// The handshake template: both sides of the wire.
//
// Protocols are taught as lists of message names — SYN, SYN-ACK, ACK — and the
// list is why nobody remembers them. A list has one column, and a protocol has
// two: the whole content of a handshake is WHO is speaking and what the other
// side is allowed to assume once they have heard it. Written down the middle of
// a page, "SYN-ACK" is a word to memorise. Drawn as an arrow coming BACK, it is
// obviously an answer, and the viewer stops memorising and starts reading.
//
// So this template is a sequence diagram and nothing else: two named columns,
// two lifelines, and arrows that fire across one at a time. Each arrow carries
// what it says on the wire and, underneath, what it ACCOMPLISHES — because the
// label is the protocol's vocabulary and the meaning is the lesson, and a clip
// that shows only the label has drawn the list again with better typography.
//
// The validators keep the picture a dialogue. Messages are covered in order
// with no skips, because a sequence diagram whose arrows arrive out of order is
// not a sequence diagram. And a plan whose arrows all travel the same way is
// rejected: a one-directional exchange is a broadcast, and if the subject
// really is one side talking then this is the wrong template rather than a
// handshake with a missing reply. That rejection is written as a redirection,
// not a scolding — the model has usually understood the protocol and picked the
// wrong picture for it.
//
// The clip closes on the channel being open, because that is the state the
// whole exchange existed to reach, and it is the one frame that says what a
// handshake is FOR.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "handshake",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "Both sides of the wire",
		Description: "Two named columns, lifelines down the frame, and the messages of an exchange firing across one at a time with what each one accomplishes. Reach for it when the subject is a protocol or a negotiation — a handshake, an auth flow, a request and its answer.",
		Example:     "The TCP three-way handshake",
		PromptFile:  snippetHandshakeTemplateName,
		NeedsCode:   false,
		// The empty wire, three arrows, the open channel: five states, and
		// under thirty-five seconds an arrow's meaning cannot be read before
		// the next one fires.
		MinTargetSec: 35,
		// Three or four messages plus the two framing shots is six beats, which
		// is what fifty seconds funds.
		DefaultTargetSec: 50,
		// Six messages plus the wire and the open channel is eight. Past that
		// the exchange is a session, not a handshake.
		MaxBeats: 8,
		// A beat here is a SHOT — one arrow crossing — not a step in an
		// argument, so a beat holds the frame for about nine seconds.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Handshake: true},
		OwnsPlan:          planFields{Handshake: true},
		Normalize:         normalizeHandshakePlan,
		Validate:          validateHandshakePlan,
		Scenes:            handshakeScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(HandshakeShows(), ", "),
				"Dirs":          strings.Join(HandshakeDirs(), ", "),
				"MinMsgs":       minHandshakeMsgs,
				"MaxMsgs":       maxHandshakeMsgs,
				"MaxSideWords":  maxHandshakeSideWords,
				"MaxLabelWords": maxHandshakeLabelWords,
				"MaxMeansWords": maxHandshakeMeansWords,
			}
		},
	})
}

const snippetHandshakeTemplateName = "snippet_handshake.tmpl"

const (
	// One arrow is not an exchange; it is a message. Two is the smallest thing
	// with a reply in it, which is the smallest thing this picture explains.
	minHandshakeMsgs = 2
	// Six arrows stack down the middle of the frame at a readable size with
	// their meanings under them; a seventh row starts colliding with the
	// column headers.
	maxHandshakeMsgs = 6

	// A column header is a role — "your browser", "the server". Three words
	// names a role; four is describing one.
	maxHandshakeSideWords = 3
	// What goes on the wire, riding the arrow: "SYN, sequence 0".
	maxHandshakeLabelWords = 6
	// What the arrow accomplishes, set under it. A caption, not a paragraph.
	maxHandshakeMeansWords = 10
)

// handshakeDirs is the closed vocabulary of which way an arrow travels.
var handshakeDirs = map[string]bool{
	// Left column to right column: the initiator speaking.
	"ltr": true,
	// Right column back to left: the answer.
	"rtl": true,
}

// HandshakeDirs returns the direction vocabulary sorted.
func HandshakeDirs() []string {
	out := make([]string, 0, len(handshakeDirs))
	for k := range handshakeDirs {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// handshakeShows is the closed vocabulary of what a beat does.
var handshakeShows = map[string]bool{
	// Two columns and an empty wire between them. The opener.
	"wire": true,
	// Arrow At fires across, its label riding it and its meaning under it.
	"msg": true,
	// The established channel: the wire fills and the summary lands. The
	// closer.
	"open": true,
}

// HandshakeShows returns the beat vocabulary sorted.
func HandshakeShows() []string {
	out := make([]string, 0, len(handshakeShows))
	for k := range handshakeShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// HandshakeMsg is one arrow across the wire.
type HandshakeMsg struct {
	// Dir is which way it travels, from handshakeDirs.
	Dir string `json:"dir,omitempty"`
	// Label is what goes on the wire — "SYN, sequence 0".
	Label string `json:"label"`
	// Means is what the message accomplishes — "asks to start talking".
	Means string `json:"means,omitempty"`
}

// ResolvedDir returns the arrow's direction, defaulting the unknown to
// left-to-right. The initiator speaks first in every exchange this template
// draws, so an unstated direction is far more often the outbound one.
func (m HandshakeMsg) ResolvedDir() string {
	d := strings.ToLower(strings.TrimSpace(m.Dir))
	if handshakeDirs[d] {
		return d
	}
	return "ltr"
}

// HandshakeSpec is the two parties and everything they say to each other. On
// the plan because the columns stand for the whole clip.
type HandshakeSpec struct {
	// Left names the initiator's column — "your browser".
	Left string `json:"left"`
	// Right names the responder's column — "the server".
	Right string `json:"right"`
	// Msgs are the arrows, in the order they fire.
	Msgs []HandshakeMsg `json:"msgs"`
}

// HandshakeBeat is one shot: which state of the wire this beat shows.
type HandshakeBeat struct {
	// Show is a handshakeShows name.
	Show string `json:"show"`
	// At is the message this beat fires, for "msg".
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a message —
// the workhorse state most beats of this template are in.
func (b HandshakeBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if handshakeShows[s] {
		return s
	}
	return "msg"
}

func normalizeHandshakePlan(p *SnippetPlan) {
	h := p.Handshake
	if h == nil {
		return
	}
	h.Left = clampWords(collapseSpaces(h.Left), maxHandshakeSideWords)
	h.Right = clampWords(collapseSpaces(h.Right), maxHandshakeSideWords)
	if len(h.Msgs) > maxHandshakeMsgs {
		h.Msgs = h.Msgs[:maxHandshakeMsgs]
	}
	for i := range h.Msgs {
		h.Msgs[i].Dir = h.Msgs[i].ResolvedDir()
		h.Msgs[i].Label = clampWords(collapseSpaces(h.Msgs[i].Label), maxHandshakeLabelWords)
		h.Msgs[i].Means = clampWords(collapseSpaces(h.Msgs[i].Means), maxHandshakeMeansWords)
	}

	last := len(h.Msgs) - 1
	for i := range p.Beats {
		b := p.Beats[i].Handshake
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.At < 0 {
			b.At = 0
		}
		if last >= 0 && b.At > last {
			b.At = last
		}
	}
}

func validateHandshakePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Handshake: true}); err != nil {
		return err
	}

	h := p.Handshake
	if h == nil {
		return fmt.Errorf("the plan has no exchange — this template is two columns and the arrows between them, so without messages there is an empty wire and nothing on it")
	}
	if strings.TrimSpace(h.Left) == "" || strings.TrimSpace(h.Right) == "" {
		return fmt.Errorf("the two sides are not both named (left %q, right %q). The columns are the picture: an arrow means something only when the viewer knows who is at each end of it",
			h.Left, h.Right)
	}
	if n := len(h.Msgs); n < minHandshakeMsgs || n > maxHandshakeMsgs {
		return fmt.Errorf("the exchange has %d messages, want %d-%d. One arrow is a message rather than a handshake, and past %d the rows stack into the column headers",
			n, minHandshakeMsgs, maxHandshakeMsgs, maxHandshakeMsgs)
	}
	sameDir := true
	for i, m := range h.Msgs {
		if strings.TrimSpace(m.Label) == "" {
			return fmt.Errorf("message %d has nothing on it. Every arrow carries what actually goes over the wire — \"SYN, sequence 0\" — because that is the protocol's own vocabulary and the viewer will meet it again", i)
		}
		if strings.TrimSpace(m.Means) == "" {
			return fmt.Errorf("message %d (%q) never says what it accomplishes. The label is the vocabulary and the meaning is the lesson — without it this clip is a list of names in a nicer font",
				i, m.Label)
		}
		if m.ResolvedDir() != h.Msgs[0].ResolvedDir() {
			sameDir = false
		}
	}
	// A handshake has both sides speaking, by definition. Worded as a
	// redirection because a plan that gets here has usually understood its
	// subject and chosen the wrong picture for it.
	if sameDir {
		return fmt.Errorf("every message travels %q, so only %q ever speaks. A handshake is a negotiation and the whole reason it is drawn as two columns is that the answer comes BACK — set the replies to the other direction. If the subject really is one side sending in one direction, this is a sequence rather than a handshake, and `flow` or `trace` draws it better",
			h.Msgs[0].ResolvedDir(), map[string]string{"ltr": h.Left, "rtl": h.Right}[h.Msgs[0].ResolvedDir()])
	}

	if p.Beats[0].Handshake == nil || p.Beats[0].Handshake.ResolvedShow() != "wire" {
		return fmt.Errorf("beat %q does not open on the empty wire. The columns have to be named and standing before anything crosses them, or the first arrow is a line between two words the viewer has not read yet — open with {\"show\": \"wire\"}",
			p.Beats[0].ID)
	}
	if lastBeat := p.Beats[len(p.Beats)-1]; lastBeat.Handshake == nil || lastBeat.Handshake.ResolvedShow() != "open" {
		return fmt.Errorf("the clip does not close on the open channel. That state is what the whole exchange existed to reach and it is the one frame that says what a handshake is FOR — end with {\"show\": \"open\"}")
	}

	want := 0
	for i, b := range p.Beats {
		hb := b.Handshake
		if hb == nil {
			return fmt.Errorf("beat %q has no handshake direction — every beat shows one state of the wire", b.ID)
		}
		switch hb.ResolvedShow() {
		case "wire":
			if i != 0 {
				return fmt.Errorf("beat %q clears the wire part-way through. Delivered arrows stay on screen — the stack down the middle IS the transcript — so a second {\"show\": \"wire\"} throws away the record the clip is building", b.ID)
			}
		case "msg":
			if hb.At > want {
				return fmt.Errorf("beat %q fires message %d (%s) while message %d (%s) has not been sent. The arrows are a sequence; skipping one leaves a reply on screen answering something the viewer never saw",
					b.ID, hb.At, handshakeMsgName(h, hb.At), want, handshakeMsgName(h, want))
			}
			if hb.At < want {
				return fmt.Errorf("beat %q fires message %d (%s) again, after it has already crossed. Each arrow crosses once — if it needs more time, give the beat more narration rather than a second flight",
					b.ID, hb.At, handshakeMsgName(h, hb.At))
			}
			want++
		case "open":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q opens the channel before the end. \"open\" is the closer — once the wire is lit there is nothing left for an arrow to establish", b.ID)
			}
		}
	}
	if want != len(h.Msgs) {
		return fmt.Errorf("the clip fires %d of the %d messages, so %s is written into the exchange and never sent. Either give it a beat or take it out of the plan",
			want, len(h.Msgs), handshakeMsgName(h, want))
	}
	return nil
}

// handshakeMsgName quotes a message for an error. Index-safe because a
// validator may be handed a plan that never went through normalize, and a
// rejection that panics is worse than the plan it was rejecting.
func handshakeMsgName(h *HandshakeSpec, i int) string {
	if i < 0 || i >= len(h.Msgs) {
		return "no such message"
	}
	return fmt.Sprintf("%q", h.Msgs[i].Label)
}

// handshakeScenes lays the clip out as ONE scene. Which arrows have landed at
// each moment is resolved here, so the component draws a transcript it is
// handed rather than replaying the protocol in TypeScript.
func handshakeScenes(in SnippetSceneInput) ([]Scene, error) {
	h := in.Plan.Handshake
	if h == nil {
		return nil, fmt.Errorf("the plan has no exchange")
	}

	msgs := make([]map[string]any, 0, len(h.Msgs))
	for _, m := range h.Msgs {
		msgs = append(msgs, map[string]any{
			"dir":   m.ResolvedDir(),
			"label": m.Label,
			"means": m.Means,
		})
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	sent := make([]int, 0, len(h.Msgs))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Handshake == nil {
			return nil, fmt.Errorf("beat %q has no handshake direction", beat.ID)
		}
		show := beat.Handshake.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "msg" {
			step["at"] = beat.Handshake.At
			sent = append(sent, beat.Handshake.At)
		}
		delivered := append([]int(nil), sent...)
		sort.Ints(delivered)
		step["delivered"] = delivered
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneHandshake,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title": in.Plan.Title,
			"left":  h.Left,
			"right": h.Right,
			"msgs":  msgs,
			"steps": steps,
		}),
	}}, nil
}
