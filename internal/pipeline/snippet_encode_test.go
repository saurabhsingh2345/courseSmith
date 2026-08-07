package pipeline

import (
	"strings"
	"testing"
)

const enNarration = "The character gets a number from Unicode, and only then does the number become bytes."

func encodePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "encode",
		Title:    "One character, four bytes",
		Encode: &EncodeSpec{
			Glyph:     "\U0001F600",
			Codepoint: "U+1F600",
			Bytes:     []string{"0xF0", "0x9F", "0x98", "0x80"},
			Note:      "one character, four bytes, and every one of them is needed",
		},
		Beats: []SnippetBeat{
			{ID: "glyph", Heading: "The character", Narration: enNarration, Encode: &EncodeBeat{Show: "glyph"}},
			{ID: "codepoint", Heading: "Its number", Narration: enNarration, Encode: &EncodeBeat{Show: "codepoint"}},
			{ID: "lead", Heading: "The lead byte", Narration: enNarration, Encode: &EncodeBeat{Show: "bytes"}},
			{ID: "rest", Heading: "The rest", Narration: enNarration, Encode: &EncodeBeat{Show: "bytes"}},
			{ID: "note", Heading: "What it means", Narration: enNarration, Encode: &EncodeBeat{Show: "note"}},
		},
	}
	// The template's ideal is 28 words per beat, so the fixture budget is sized
	// against that — nBeats * 40 would demand more beats than it has.
	p.targetWords = 5 * 28
	return p
}

func TestEncodePlanAccepted(t *testing.T) {
	if err := validateEncodePlan(encodePlan()); err != nil {
		t.Fatalf("a well-formed encode plan was rejected: %v", err)
	}
}

// The family's signature rule: the bytes are produced by Go's own UTF-8
// encoder and a mismatch is rejected with BOTH sequences quoted in hex.
func TestEncodeRejectsBytesThatAreNotTheEncoding(t *testing.T) {
	p := encodePlan()
	p.Encode.Bytes = []string{"0xF0", "0x9F", "0x99", "0x80"}
	err := validateEncodePlan(p)
	if err == nil {
		t.Fatal("a four-byte sequence with one digit wrong was accepted")
	}
	if !strings.Contains(err.Error(), "0xF0 0x9F 0x99 0x80") {
		t.Fatalf("the error does not quote the plan's bytes: %v", err)
	}
	if !strings.Contains(err.Error(), "0xF0 0x9F 0x98 0x80") {
		t.Fatalf("the error does not quote the real bytes: %v", err)
	}
}

// The byte-count rule falls out of the encoding check rather than being stated
// separately, which is why it cannot disagree with a real machine.
func TestEncodeRejectsTooFewBytesForTheCodepoint(t *testing.T) {
	p := encodePlan()
	p.Encode.Bytes = []string{"0xF0", "0x9F"}
	err := validateEncodePlan(p)
	if err == nil {
		t.Fatal("a truncated byte sequence was accepted")
	}
	if !strings.Contains(err.Error(), "0xF0 0x9F 0x98 0x80") {
		t.Fatalf("the error does not quote the real bytes: %v", err)
	}
}

func TestEncodeRejectsAGlyphThatIsNotTheCodepoint(t *testing.T) {
	p := encodePlan()
	p.Encode.Glyph = "A"
	err := validateEncodePlan(p)
	if err == nil {
		t.Fatal("a card showing one character over another character's number was accepted")
	}
	if !strings.Contains(err.Error(), "U+0041") {
		t.Fatalf("the error does not quote the glyph's real codepoint: %v", err)
	}
}

func TestEncodeRejectsASurrogateCodepoint(t *testing.T) {
	p := encodePlan()
	p.Encode.Codepoint = "U+D800"
	p.Encode.Glyph = "A"
	if err := validateEncodePlan(p); err == nil {
		t.Fatal("a surrogate half was accepted, and UTF-8 has no encoding for one")
	}
}

func TestEncodeRejectsACodepointThatDoesNotParse(t *testing.T) {
	p := encodePlan()
	p.Encode.Codepoint = "the letter A"
	if err := validateEncodePlan(p); err == nil {
		t.Fatal("a codepoint written as prose was accepted")
	}
}

func TestEncodeRejectsAGlyphOfSeveralCharacters(t *testing.T) {
	p := encodePlan()
	p.Encode.Glyph = "hello"
	if err := validateEncodePlan(p); err == nil {
		t.Fatal("a whole word was accepted as the glyph")
	}
}

func TestEncodeRequiresOpeningOnTheGlyph(t *testing.T) {
	p := encodePlan()
	p.Beats[0].Encode = &EncodeBeat{Show: "codepoint"}
	p.Beats[1].Encode = &EncodeBeat{Show: "glyph"}
	if err := validateEncodePlan(p); err == nil {
		t.Fatal("a clip that starts at the codepoint was accepted")
	}
}

func TestEncodeRequiresClosingOnTheNote(t *testing.T) {
	p := encodePlan()
	p.Beats[4].Encode = &EncodeBeat{Show: "bytes"}
	if err := validateEncodePlan(p); err == nil {
		t.Fatal("a clip with no closing note was accepted")
	}
}

func TestEncodeRequiresTheCodepointBeforeTheBytes(t *testing.T) {
	p := encodePlan()
	p.Beats[1].Encode = &EncodeBeat{Show: "bytes"}
	p.Beats[2].Encode = &EncodeBeat{Show: "codepoint"}
	err := validateEncodePlan(p)
	if err == nil {
		t.Fatal("byte boxes landing before the number they came from was accepted")
	}
	if !strings.Contains(err.Error(), "codepoint") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEncodeRejectsMoreByteBeatsThanBytes(t *testing.T) {
	p := encodePlan()
	p.Encode.Glyph = "A"
	p.Encode.Codepoint = "U+0041"
	p.Encode.Bytes = []string{"0x41"}
	err := validateEncodePlan(p)
	if err == nil {
		t.Fatal("two byte beats for a one-byte character was accepted")
	}
	if !strings.Contains(err.Error(), "only 1 byte") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Decoration and casing on a codepoint or a byte are phrasing habits, not wrong
// answers, so they are repaired rather than argued.
func TestEncodeNormalizeCanonicalisesTheHex(t *testing.T) {
	p := encodePlan()
	p.Encode.Codepoint = "u+1f600"
	p.Encode.Bytes = []string{"f0", "0x9f", "98", "0X80"}
	normalizeEncodePlan(p)
	if p.Encode.Codepoint != "U+1F600" {
		t.Fatalf("codepoint normalized to %q, want U+1F600", p.Encode.Codepoint)
	}
	if got := strings.Join(p.Encode.Bytes, " "); got != "0xF0 0x9F 0x98 0x80" {
		t.Fatalf("bytes normalized to %q", got)
	}
	if err := validateEncodePlan(p); err != nil {
		t.Fatalf("a decorated-but-correct plan was rejected after normalize: %v", err)
	}
}

func TestEncodeShowDefaultsToBytes(t *testing.T) {
	b := EncodeBeat{Show: "sparkle"}
	if got := b.ResolvedShow(); got != "bytes" {
		t.Fatalf("an unknown show resolved to %q, want bytes", got)
	}
}

// The renderer never decodes anything: each byte arrives with its bits already
// split into the UTF-8 marker prefix and the payload.
func TestEncodeScenesSplitTheMarkerBits(t *testing.T) {
	p := encodePlan()
	scenes, err := encodeScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	boxes, _ := props["bytes"].([]map[string]any)
	if len(boxes) != 4 {
		t.Fatalf("want 4 byte boxes, got %d", len(boxes))
	}
	// The lead byte of a four-byte sequence announces its own length: 11110.
	if boxes[0]["hex"] != "0xF0" || boxes[0]["marker"] != "11110" || boxes[0]["payload"] != "000" {
		t.Fatalf("the lead byte is wrong: %v", boxes[0])
	}
	if boxes[0]["lead"] != true {
		t.Fatalf("the lead byte is not marked as one: %v", boxes[0])
	}
	// Every continuation byte announces itself as one: 10.
	if boxes[1]["marker"] != "10" || boxes[1]["payload"] != "011111" {
		t.Fatalf("the first continuation byte is wrong: %v", boxes[1])
	}
	if boxes[3]["lead"] != false {
		t.Fatalf("a continuation byte is marked as a lead: %v", boxes[3])
	}

	steps, _ := props["steps"].([]map[string]any)
	if steps[0]["show"] != "glyph" || steps[0]["landed"] != 0 {
		t.Fatalf("the opener has already landed bytes: %v", steps[0])
	}
	last := steps[len(steps)-1]
	if last["show"] != "note" || last["landed"] != 4 {
		t.Fatalf("the closer does not hold the whole row: %v", last)
	}
}
