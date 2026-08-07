package pipeline

// The encode template: from a character to bytes.
//
// "Everything is bytes" is the sentence every foundations course opens with,
// and almost nobody who repeats it can say what happens between the letter on
// the keyboard and the number in memory. There are two steps, not one, and
// collapsing them is the source of every mojibake bug the viewer will ever
// file: first the character is assigned a NUMBER by Unicode — a codepoint, a
// fact about the character that no file format can change — and only then is
// that number written down as bytes by an ENCODING, which is a choice. 'A' is
// U+0041 is 0x41, and the three-step chain is boring, which is exactly why the
// clip should use an emoji: U+1F600 is one character, one codepoint and FOUR
// bytes, and the moment the fourth box lands is the moment "everything is
// bytes" stops being a slogan.
//
// So the picture is three stations left to right, joined by arrows: a glyph
// card, a codepoint, a row of byte boxes. The byte boxes are drawn with their
// UTF-8 marker bits tinted apart from their payload bits, because that is the
// answer to the question the picture provokes — how does anything know where
// one character ends — and it costs nothing to show.
//
// The validator does not check byte counts by rule. It encodes the codepoint
// with Go's own UTF-8 encoder and demands the plan's bytes match it exactly.
// That one check subsumes every rule about lead bytes and continuation bytes
// and can never disagree with a real machine, which a hand-written rule
// eventually would. A model asked for the bytes of an emoji will produce a
// plausible-looking four-byte sequence with one digit wrong, and a diagram of
// an encoding that is wrong is worse than no diagram: it teaches the viewer
// that the picture is decoration rather than the actual bytes on disk. The
// rejection quotes both sequences in hex, so the fix is mechanical.
//
// The glyph is checked against the codepoint for the same reason: a card
// showing 'é' over U+0065 is a clip whose first arrow is a lie.

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "encode",
		Category:    CatConcepts,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "From 'A' to bytes",
		Description: "One character walked across three stations — glyph, codepoint, bytes — with the UTF-8 marker bits tinted apart from the payload. Reach for it when the subject is text as data: what a codepoint is, why an emoji is four bytes, where mojibake comes from.",
		Example:     "How the emoji becomes four bytes of UTF-8",
		PromptFile:  snippetEncodeTemplateName,
		NeedsCode:   false,
		// Four stations, four states. Under thirty seconds the byte boxes land
		// before the codepoint they came from has been read.
		MinTargetSec:     30,
		DefaultTargetSec: 45,
		// Glyph + codepoint + one beat per byte + the note. Four bytes is the most
		// UTF-8 ever uses, so seven is exactly what the longest legal sequence
		// needs and not one beat more.
		MaxBeats: 7,
		// A beat is a SHOT — one station lighting up — not a step in an argument.
		// Twenty-eight words is about nine seconds, which is as long as a single
		// station holds anybody.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Encode: true},
		OwnsPlan:          planFields{Encode: true},
		Normalize:         normalizeEncodePlan,
		Validate:          validateEncodePlan,
		Scenes:            encodeScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":        strings.Join(MetricRoles(), ", "),
				"Shows":        strings.Join(EncodeShows(), ", "),
				"MaxBytes":     maxEncodeBytes,
				"MaxGlyphRune": maxEncodeGlyphRunes,
				"MaxNoteWords": maxEncodeNoteWords,
			}
		},
	})
}

const snippetEncodeTemplateName = "snippet_encode.tmpl"

const (
	// UTF-8 never uses more than four bytes for a codepoint. This is not a taste
	// judgement, it is the encoding.
	maxEncodeBytes = 4
	// One rune is the character. Two is allowed only because a great many emoji
	// are written with a trailing variation selector, and rejecting the string a
	// creator actually pasted would be pedantry — the SECOND rune is decoration,
	// and the first is the one the clip is about.
	maxEncodeGlyphRunes = 2
	// The note is one line under the finished row. Twelve words is an
	// observation; more is a paragraph the closer has no room for.
	maxEncodeNoteWords = 12
)

// encodeShows is the closed vocabulary of what a beat does.
var encodeShows = map[string]bool{
	// The character alone, set very large. The opener.
	"glyph": true,
	// The U+ number slides out of the glyph: Unicode's fact about the character.
	"codepoint": true,
	// The byte boxes land, one at a time, with their UTF-8 prefix bits marked.
	"bytes": true,
	// The note line under the finished row. The closer.
	"note": true,
}

// EncodeShows returns the beat vocabulary sorted.
func EncodeShows() []string {
	out := make([]string, 0, len(encodeShows))
	for k := range encodeShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// EncodeSpec is one character and its journey into memory. On the plan because
// a single character persists across the whole clip; the beats only move along
// the chain.
type EncodeSpec struct {
	// Glyph is the character itself — "A", "é", "😀".
	Glyph string `json:"glyph"`
	// Codepoint is Unicode's number for it, in the usual form — "U+0041".
	// Checked against the glyph, not trusted.
	Codepoint string `json:"codepoint"`
	// Bytes are the UTF-8 encoding, each written "0x41". Recomputed in Go.
	Bytes []string `json:"bytes"`
	// Note is the one line the clip closes on — the thing worth noticing.
	Note string `json:"note"`
}

// EncodeBeat is one shot: which station of the chain this beat lights.
type EncodeBeat struct {
	// Show is an encodeShows name.
	Show string `json:"show"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to the bytes —
// the station that carries most of this template's beats.
func (b EncodeBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if encodeShows[s] {
		return s
	}
	return "bytes"
}

// parseEncodeCodepoint reads "U+1F600", "0x1F600" or "1F600" as a rune. The
// decoration is a phrasing habit, not a wrong answer, so it is stripped rather
// than argued with.
func parseEncodeCodepoint(s string) (rune, error) {
	t := strings.ToUpper(strings.TrimSpace(s))
	t = strings.ReplaceAll(t, " ", "")
	t = strings.TrimPrefix(t, "U+")
	t = strings.TrimPrefix(t, "0X")
	if t == "" {
		return 0, fmt.Errorf("empty")
	}
	v, err := strconv.ParseUint(t, 16, 32)
	if err != nil {
		return 0, err
	}
	return rune(v), nil
}

// parseEncodeByte reads "0x9F", "9f" or "0X9F" as a byte value.
func parseEncodeByte(s string) (byte, error) {
	t := strings.ToUpper(strings.TrimSpace(s))
	t = strings.TrimPrefix(t, "0X")
	if t == "" {
		return 0, fmt.Errorf("empty")
	}
	v, err := strconv.ParseUint(t, 16, 8)
	if err != nil {
		return 0, err
	}
	return byte(v), nil
}

// formatEncodeBytes renders a byte sequence the way the picture writes it, so
// the two sequences in a rejection can be compared at a glance.
func formatEncodeBytes(bs []byte) string {
	out := make([]string, len(bs))
	for i, b := range bs {
		out[i] = fmt.Sprintf("0x%02X", b)
	}
	return strings.Join(out, " ")
}

func normalizeEncodePlan(p *SnippetPlan) {
	e := p.Encode
	if e == nil {
		return
	}
	e.Glyph = strings.TrimSpace(e.Glyph)
	e.Note = clampWords(collapseSpaces(e.Note), maxEncodeNoteWords)
	// Canonical spellings, so the card and the boxes read the same whatever the
	// model typed. A value that does not parse is left exactly as written, for
	// the validator to quote back.
	if cp, err := parseEncodeCodepoint(e.Codepoint); err == nil {
		e.Codepoint = fmt.Sprintf("U+%04X", cp)
	}
	bytes := make([]string, 0, len(e.Bytes))
	for _, raw := range e.Bytes {
		if len(bytes) >= maxEncodeBytes {
			break
		}
		if b, err := parseEncodeByte(raw); err == nil {
			bytes = append(bytes, fmt.Sprintf("0x%02X", b))
			continue
		}
		bytes = append(bytes, strings.TrimSpace(raw))
	}
	e.Bytes = bytes

	for i := range p.Beats {
		if b := p.Beats[i].Encode; b != nil {
			b.Show = b.ResolvedShow()
		}
	}
}

func validateEncodePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Encode: true}); err != nil {
		return err
	}

	e := p.Encode
	if e == nil {
		return fmt.Errorf("the plan has no character — this template walks ONE character from glyph to codepoint to bytes, so the character is the clip")
	}
	if strings.TrimSpace(e.Glyph) == "" {
		return fmt.Errorf("the plan has no glyph. The first station is the character set very large, and an empty card is an arrow pointing out of nothing")
	}
	if n := utf8.RuneCountInString(e.Glyph); n > maxEncodeGlyphRunes {
		return fmt.Errorf("the glyph %q is %d characters. This clip encodes ONE character — pick the single letter, symbol or emoji whose bytes you want to show",
			e.Glyph, n)
	}
	cp, err := parseEncodeCodepoint(e.Codepoint)
	if err != nil {
		return fmt.Errorf("the codepoint %q does not read as a Unicode codepoint. Write it the way Unicode writes it, hexadecimal after a U+ — \"U+0041\", \"U+1F600\"", e.Codepoint)
	}
	if cp > 0x10FFFF {
		return fmt.Errorf("the codepoint %s is past U+10FFFF, which is the largest codepoint Unicode has. There is nothing there to encode", e.Codepoint)
	}
	if cp >= 0xD800 && cp <= 0xDFFF {
		return fmt.Errorf("the codepoint %s is a surrogate half. Surrogates are a UTF-16 bookkeeping device, not characters, and UTF-8 has no encoding for one — pick the real codepoint of the character you mean", e.Codepoint)
	}
	// The glyph and the codepoint have to be the same character, or the clip's
	// first arrow is a lie.
	if first, _ := utf8.DecodeRuneInString(e.Glyph); first != cp {
		return fmt.Errorf("the glyph %q is U+%04X but the plan labels it %s. The first arrow of the picture says the card and the number are the same character, so they have to be",
			e.Glyph, first, e.Codepoint)
	}

	if len(e.Bytes) == 0 {
		return fmt.Errorf("the plan lists no bytes. The byte boxes are the payoff of the whole chain — without them the clip stops at a number and never reaches memory")
	}
	if len(e.Bytes) > maxEncodeBytes {
		return fmt.Errorf("the plan lists %d bytes, and UTF-8 never uses more than %d for one codepoint. That is the encoding, not a layout limit",
			len(e.Bytes), maxEncodeBytes)
	}
	got := make([]byte, 0, len(e.Bytes))
	for i, raw := range e.Bytes {
		b, err := parseEncodeByte(raw)
		if err != nil {
			return fmt.Errorf("byte %d is %q, which does not read as a byte. Write each one as two hexadecimal digits after 0x — \"0x41\", \"0x9F\"", i, raw)
		}
		got = append(got, b)
	}

	// THE ENCODING, done by Go's own encoder rather than by a rule about lead
	// bytes. One check, and it cannot disagree with a real machine.
	want := []byte(string(cp))
	if string(got) != string(want) {
		return fmt.Errorf("the plan says %s encodes to %s, but %s in UTF-8 is %s. Every box on screen is drawn from the codepoint, marker bits and all, so wrong bytes would visibly disagree with the number they came from",
			e.Codepoint, formatEncodeBytes(got), e.Codepoint, formatEncodeBytes(want))
	}

	if strings.TrimSpace(e.Note) == "" {
		return fmt.Errorf("the plan has no note. The clip closes on one line about what the row of boxes means — \"one character, four bytes, and every one of them needed\" — and a closer with nothing to say is a held frame")
	}

	// The shape: the chain is walked in the only order it happens in.
	if p.Beats[0].Encode == nil || p.Beats[0].Encode.ResolvedShow() != "glyph" {
		return fmt.Errorf("beat %q does not open on the glyph. The clip is a journey and the character is where it starts — open with {\"show\": \"glyph\"}",
			p.Beats[0].ID)
	}
	last := p.Beats[len(p.Beats)-1]
	if last.Encode == nil || last.Encode.ResolvedShow() != "note" {
		return fmt.Errorf("beat %q does not close on the note. The last frame is the whole chain lit with one line under it, and that line is the thing the viewer keeps",
			last.ID)
	}

	count := map[string]int{}
	firstAt := map[string]int{}
	for i, b := range p.Beats {
		if b.Encode == nil {
			return fmt.Errorf("beat %q has no encode direction — every beat lights one station of the chain", b.ID)
		}
		show := b.Encode.ResolvedShow()
		if _, seen := firstAt[show]; !seen {
			firstAt[show] = i
		}
		count[show]++
		switch show {
		case "glyph":
			if i != 0 {
				return fmt.Errorf("beat %q goes back to the bare glyph part-way through. The character is the opener; returning to it walks the journey backwards", b.ID)
			}
		case "note":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q lands the note before the bytes are down. The note is the closer — it comments on a row that has to already be there", b.ID)
			}
		case "codepoint":
			if count["codepoint"] > 1 {
				return fmt.Errorf("beat %q slides the codepoint out again. The number appears once and then stays on screen — a second reveal re-teaches a station the viewer has already passed", b.ID)
			}
		}
	}
	for _, need := range []string{"codepoint", "bytes"} {
		if _, ok := firstAt[need]; !ok {
			return fmt.Errorf("no beat shows %q. The chain is glyph, then codepoint, then bytes: skip a station and the clip claims a character becomes bytes by magic", need)
		}
	}
	if firstAt["codepoint"] > firstAt["bytes"] {
		return fmt.Errorf("the bytes land before the codepoint appears. The bytes are an encoding OF the number, so a byte box arriving first is a box with nothing behind it — put the codepoint beat before the bytes beat")
	}
	if n := count["bytes"]; n > len(e.Bytes) {
		return fmt.Errorf("there are %d \"bytes\" beats but only %d byte(s) to land. A byte beat with no box left to place holds a finished row for nine seconds — use one beat per byte at most",
			n, len(e.Bytes))
	}
	return nil
}

// encodeScenes lays the clip out as ONE scene. Each byte arrives with its bits
// already split into the UTF-8 marker prefix and the payload, so the component
// never decodes anything: it only decides which station is lit.
func encodeScenes(in SnippetSceneInput) ([]Scene, error) {
	e := in.Plan.Encode
	if e == nil {
		return nil, fmt.Errorf("the plan has no character")
	}

	boxes := make([]map[string]any, 0, len(e.Bytes))
	for _, raw := range e.Bytes {
		b, err := parseEncodeByte(raw)
		if err != nil {
			return nil, fmt.Errorf("byte %q does not parse", raw)
		}
		bits := fmt.Sprintf("%08b", b)
		// How many leading bits are structure rather than payload, read off the
		// byte itself rather than from the sequence length — the honest source,
		// and the one that stays right if the sequence ever changes.
		marker := 1
		switch {
		case b&0x80 == 0x00:
			marker = 1
		case b&0xC0 == 0x80:
			marker = 2
		case b&0xE0 == 0xC0:
			marker = 3
		case b&0xF0 == 0xE0:
			marker = 4
		case b&0xF8 == 0xF0:
			marker = 5
		}
		boxes = append(boxes, map[string]any{
			"hex":     fmt.Sprintf("0x%02X", b),
			"bits":    bits,
			"marker":  bits[:marker],
			"payload": bits[marker:],
			"lead":    b&0xC0 != 0x80,
		})
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	// How many byte boxes have landed by each beat, accumulated in Go so the
	// renderer never replays the beat list to find out. The boxes are dealt out
	// evenly across however many byte beats the plan used.
	byteBeats := 0
	for i := range in.Plan.Beats {
		if b := in.Plan.Beats[i].Encode; b != nil && b.ResolvedShow() == "bytes" {
			byteBeats++
		}
	}
	seenByteBeats := 0
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Encode == nil {
			return nil, fmt.Errorf("beat %q has no encode direction", beat.ID)
		}
		show := beat.Encode.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "bytes" {
			seenByteBeats++
			step["at"] = seenByteBeats - 1
		}
		landed := 0
		if byteBeats > 0 {
			landed = len(boxes) * seenByteBeats / byteBeats
		}
		if seenByteBeats == byteBeats && byteBeats > 0 {
			landed = len(boxes)
		}
		step["landed"] = landed
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneEncode,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":     in.Plan.Title,
			"glyph":     e.Glyph,
			"codepoint": e.Codepoint,
			"note":      e.Note,
			"bytes":     boxes,
			"steps":     steps,
		}),
	}}, nil
}
