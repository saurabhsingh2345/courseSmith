package pipeline

import (
	"strings"
	"testing"
)

const hsNarration = "Nothing is assumed until the other side has answered and both ends agree on where they are."

func handshakePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "handshake",
		Title:    "Three messages before a single byte",
		Handshake: &HandshakeSpec{
			Left:  "your browser",
			Right: "the server",
			Msgs: []HandshakeMsg{
				{Dir: "ltr", Label: "SYN, sequence 0", Means: "asks to start talking and picks a number"},
				{Dir: "rtl", Label: "SYN-ACK, sequence 0", Means: "agrees, and picks a number of its own"},
				{Dir: "ltr", Label: "ACK, sequence 1", Means: "confirms it heard the answer come back"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "the-wire", Heading: "Two ends", Narration: hsNarration, Handshake: &HandshakeBeat{Show: "wire"}},
			{ID: "syn", Heading: "The knock", Narration: hsNarration, Handshake: &HandshakeBeat{Show: "msg", At: 0}},
			{ID: "syn-ack", Heading: "The answer", Narration: hsNarration, Handshake: &HandshakeBeat{Show: "msg", At: 1}},
			{ID: "ack", Heading: "The confirmation", Narration: hsNarration, Handshake: &HandshakeBeat{Show: "msg", At: 2}},
			{ID: "open", Heading: "Open", Narration: hsNarration, Handshake: &HandshakeBeat{Show: "open"}},
		},
	}
	// Against this template's own ideal of 28 words per beat, not the shared
	// 40: at 40 the shared bounds demand more beats than the fixture has, and
	// it would be rejected for its length before any rule under test ran.
	p.targetWords = 5 * 28
	return p
}

func TestHandshakePlanAccepted(t *testing.T) {
	if err := validateHandshakePlan(handshakePlan()); err != nil {
		t.Fatalf("a well-formed handshake plan was rejected: %v", err)
	}
}

// A handshake has both sides speaking by definition, and the rejection is
// written as a redirection because the model has usually understood the
// subject and picked the wrong picture for it.
func TestHandshakeRejectsAnExchangeWithOnlyOneSideSpeaking(t *testing.T) {
	p := handshakePlan()
	p.Handshake.Msgs[1].Dir = "ltr"
	err := validateHandshakePlan(p)
	if err == nil {
		t.Fatal("an exchange where only one side ever speaks was accepted")
	}
	if !strings.Contains(err.Error(), "your browser") {
		t.Fatalf("the error does not name the side doing all the talking: %v", err)
	}
	if !strings.Contains(err.Error(), "flow") {
		t.Fatalf("the error does not offer a template that fits instead: %v", err)
	}
}

func TestHandshakeRejectsSkippingAMessage(t *testing.T) {
	p := handshakePlan()
	p.Beats[2].Handshake = &HandshakeBeat{Show: "msg", At: 2}
	p.Beats[3].Handshake = &HandshakeBeat{Show: "msg", At: 1}
	err := validateHandshakePlan(p)
	if err == nil {
		t.Fatal("a reply that crossed before the message it answers was accepted")
	}
	if !strings.Contains(err.Error(), "ACK, sequence 1") || !strings.Contains(err.Error(), "SYN-ACK, sequence 0") {
		t.Fatalf("the error does not quote both messages: %v", err)
	}
}

func TestHandshakeRejectsFiringTheSameMessageTwice(t *testing.T) {
	p := handshakePlan()
	p.Beats[3].Handshake = &HandshakeBeat{Show: "msg", At: 1}
	err := validateHandshakePlan(p)
	if err == nil {
		t.Fatal("an arrow that crossed twice was accepted")
	}
	if !strings.Contains(err.Error(), "again") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandshakeRejectsAMessageThatIsNeverSent(t *testing.T) {
	p := handshakePlan()
	p.Handshake.Msgs = append(p.Handshake.Msgs, HandshakeMsg{Dir: "rtl", Label: "FIN", Means: "asks to close the connection down"})
	err := validateHandshakePlan(p)
	if err == nil {
		t.Fatal("a message written into the exchange and never sent was accepted")
	}
	if !strings.Contains(err.Error(), "FIN") {
		t.Fatalf("the error does not name the message left on the floor: %v", err)
	}
}

func TestHandshakeRequiresOpeningOnTheWire(t *testing.T) {
	p := handshakePlan()
	p.Beats[0].Handshake = &HandshakeBeat{Show: "msg", At: 0}
	p.Beats[1].Handshake = &HandshakeBeat{Show: "wire"}
	if err := validateHandshakePlan(p); err == nil {
		t.Fatal("an arrow crossing before the columns were named was accepted")
	}
}

func TestHandshakeRequiresClosingOnTheOpenChannel(t *testing.T) {
	p := handshakePlan()
	p.Beats[4].Handshake = &HandshakeBeat{Show: "msg", At: 2}
	p.Beats[3].Handshake = &HandshakeBeat{Show: "open"}
	if err := validateHandshakePlan(p); err == nil {
		t.Fatal("a clip that never reached the open channel was accepted")
	}
}

func TestHandshakeRejectsAMessageWithNoMeaning(t *testing.T) {
	p := handshakePlan()
	p.Handshake.Msgs[1].Means = ""
	err := validateHandshakePlan(p)
	if err == nil {
		t.Fatal("an arrow that carries a name and no meaning was accepted")
	}
	if !strings.Contains(err.Error(), "SYN-ACK") {
		t.Fatalf("the error does not name the silent message: %v", err)
	}
}

func TestHandshakeRejectsASingleMessage(t *testing.T) {
	p := handshakePlan()
	p.Handshake.Msgs = p.Handshake.Msgs[:1]
	p.Beats = append(p.Beats[:2], p.Beats[4])
	p.targetWords = 3 * 28
	err := validateHandshakePlan(p)
	if err == nil {
		t.Fatal("a one-arrow exchange was accepted, and one arrow is a message rather than a handshake")
	}
	if !strings.Contains(err.Error(), "1 messages") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestHandshakeRejectsAnUnnamedSide(t *testing.T) {
	p := handshakePlan()
	p.Handshake.Right = "  "
	if err := validateHandshakePlan(p); err == nil {
		t.Fatal("a wire with only one end named was accepted")
	}
}

func TestHandshakeNormalizeTrimsTheColumnsAndTheArrows(t *testing.T) {
	p := handshakePlan()
	p.Handshake.Left = "your   own  personal web browser"
	p.Handshake.Msgs[0].Label = "SYN with an initial sequence number of exactly zero"
	p.Handshake.Msgs[0].Means = "asks to start talking and picks a number that both ends will count from"
	p.Handshake.Msgs[2].Dir = "sideways"
	p.Beats[3].Handshake.At = 42
	normalizeHandshakePlan(p)
	if got := p.Handshake.Left; got != "your own personal" {
		t.Fatalf("the left column normalized to %q, want three collapsed words", got)
	}
	if got := len(strings.Fields(p.Handshake.Msgs[0].Label)); got != maxHandshakeLabelWords {
		t.Fatalf("the label is %d words, want it clamped to %d", got, maxHandshakeLabelWords)
	}
	if got := len(strings.Fields(p.Handshake.Msgs[0].Means)); got != maxHandshakeMeansWords {
		t.Fatalf("the meaning is %d words, want it clamped to %d", got, maxHandshakeMeansWords)
	}
	if got := p.Handshake.Msgs[2].Dir; got != "ltr" {
		t.Fatalf("an unknown direction normalized to %q, want ltr", got)
	}
	if got := p.Beats[3].Handshake.At; got != 2 {
		t.Fatalf("an index off the end normalized to %d, want the last message 2", got)
	}
	if err := validateHandshakePlan(p); err != nil {
		t.Fatalf("a wordy-but-sound plan was rejected after normalize: %v", err)
	}
}

func TestHandshakeNormalizeCapsTheExchange(t *testing.T) {
	p := handshakePlan()
	for i := 0; i < 5; i++ {
		p.Handshake.Msgs = append(p.Handshake.Msgs, HandshakeMsg{Dir: "rtl", Label: "more", Means: "says something else again"})
	}
	normalizeHandshakePlan(p)
	if got := len(p.Handshake.Msgs); got != maxHandshakeMsgs {
		t.Fatalf("the exchange normalized to %d messages, want it capped at %d", got, maxHandshakeMsgs)
	}
}

func TestHandshakeShowDefaultsToMsg(t *testing.T) {
	b := HandshakeBeat{Show: "shout"}
	if got := b.ResolvedShow(); got != "msg" {
		t.Fatalf("an unknown show resolved to %q, want msg", got)
	}
}

func TestHandshakeDirDefaultsToLeftToRight(t *testing.T) {
	m := HandshakeMsg{Dir: "diagonal"}
	if got := m.ResolvedDir(); got != "ltr" {
		t.Fatalf("an unknown direction resolved to %q, want ltr", got)
	}
}

// The transcript down the middle of the frame is handed to the component, not
// replayed by it.
func TestHandshakeScenesAccumulateTheTranscript(t *testing.T) {
	p := handshakePlan()
	scenes, err := handshakeScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("want one scene spanning the clip, got %d", len(scenes))
	}
	props := scenes[0].Props

	if props["left"] != "your browser" || props["right"] != "the server" {
		t.Fatalf("the columns are wrong: %v / %v", props["left"], props["right"])
	}
	msgs, _ := props["msgs"].([]map[string]any)
	if len(msgs) != 3 {
		t.Fatalf("want 3 messages, got %d", len(msgs))
	}
	if msgs[1]["dir"] != "rtl" {
		t.Fatalf("the reply travels %v, want rtl", msgs[1]["dir"])
	}

	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != 5 {
		t.Fatalf("want 5 steps, got %d", len(steps))
	}
	if steps[0]["show"] != "wire" {
		t.Fatalf("first step shows %v, want wire", steps[0]["show"])
	}
	if first, _ := steps[0]["delivered"].([]int); len(first) != 0 {
		t.Fatalf("the opener has delivered %v, want nothing", first)
	}
	last := steps[len(steps)-1]
	if last["show"] != "open" {
		t.Fatalf("last step shows %v, want open", last["show"])
	}
	delivered, _ := last["delivered"].([]int)
	if len(delivered) != 3 || delivered[0] != 0 || delivered[2] != 2 {
		t.Fatalf("the closer has delivered %v, want every message 0 through 2", delivered)
	}
}
