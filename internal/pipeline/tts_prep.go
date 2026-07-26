package pipeline

// TTS text preprocessing: rewrites narration into the form Kokoro should
// actually speak. Code tokens become their spoken names ("__init__" →
// "dunder init"), numbers and operators become words, over-long sentences
// are split at clause boundaries, and *emphasis* spans become micro-pauses.
//
// The audio stage persists the transformed text as tts_script.json so the
// align stage can compute WER against what was actually sent to the TTS —
// not the written narration.

import (
	"fmt"
	"maps"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// TTSScriptFileName is the spoken-text artifact the audio stage writes to
// the lesson's generated dir. It records exactly what was sent to Kokoro.
const TTSScriptFileName = "tts_script.json"

// TTSFixesFileName holds pronunciation fixes the align stage auto-applies
// when the WER gate catches a misread word with a known dictionary entry.
// The audio stage consumes it, so writing it re-triggers synthesis.
const TTSFixesFileName = "tts_fixes.json"

// maxSentenceWords is the sentence-length guard: TTS prosody degrades on
// long sentences, so anything above this is split at a clause boundary.
const maxSentenceWords = 30

// defaultSpeechDict maps written code tokens to their spoken form for
// Python courses. course.yaml style.pronunciations extends/overrides it.
var defaultSpeechDict = map[string]string{
	// Language & ecosystem names
	"PyPI":       "pie pee eye",
	"pypi.org":   "pie pee eye dot org",
	"python.org": "python dot org",
	"CPython":    "see python",
	"NumPy":      "num pie",
	"SciPy":      "sigh pie",
	"PyPy":       "pie pie",
	"pytest":     "pie test",
	"pip":        "pip",
	"venv":       "v env",
	"virtualenv": "virtual env",
	"pathlib":    "path lib",
	"stdin":      "standard in",
	"stdout":     "standard out",
	"stderr":     "standard error",
	"utf-8":      "you tee eff eight",
	"REPL":       "repple",
	"IDE":        "eye dee ee",
	"CLI":        "see ell eye",
	"JSON":       "jason",
	"YAML":       "yammle",
	"SQL":        "sequel",
	"regex":       "rej ex",
	"kwargs":      "keyword args",
	"elif":        "ell if",
	"init":        "in it",
	"python3":     "python three",
	"python2":     "python two",
	"pip3":        "pip three",
	"NoneType":    "none type",
	"ValueError":  "value error",
	"TypeError":   "type error",
	"IndexError":  "index error",
	"KeyError":    "key error",
	"SyntaxError": "syntax error",

	// Contractions Kokoro expands on its own; spelling them out keeps the
	// spoken text honest for the WER check.
	"wanna": "want to",
	"gonna": "going to",
	"gotta": "got to",

	// Compound written forms
	"f-string":  "eff string",
	"f-strings": "eff strings",
	".py":       "dot pie",
	"PEP 8":     "pep eight",
	"__pycache__": "dunder pie cache",

	// Dunders that don't follow the generic pattern cleanly
	"__init__.py": "dunder init dot pie",
}

// symbolSpeech maps operators and punctuation-as-code to spoken phrases.
// Applied longest-first so "<=" wins over "<".
var symbolSpeech = map[string]string{
	"==": "equals equals",
	"!=": "not equal to",
	"<=": "less than or equal to",
	">=": "greater than or equal to",
	"->": "arrow",
	"+=": "plus equals",
	"-=": "minus equals",
	"//": "double slash",
	"[]": "square brackets",
	"{}": "curly braces",
	"()": "parentheses",
}

// dunderRe matches remaining dunder names: __init__ → "dunder init".
var dunderRe = regexp.MustCompile(`__([A-Za-z][A-Za-z0-9]*(?:_[A-Za-z0-9]+)*)__`)

// powerRe matches integer exponentiation: 2**8 → "2 to the power of 8".
var powerRe = regexp.MustCompile(`(\d+)\s*\*\*\s*(\d+)`)

// percentRe matches "50%" → "50 percent".
var percentRe = regexp.MustCompile(`(\d)%`)

// decimalRe matches decimal literals / version numbers like 3.11.
var decimalRe = regexp.MustCompile(`\b(\d+)\.(\d+)\b`)

// integerRe matches remaining standalone integers.
var integerRe = regexp.MustCompile(`\b\d+\b`)

// emphasisRe matches *emphasis* spans emitted by the script generator.
var emphasisRe = regexp.MustCompile(`\*{1,2}([^*\n]+?)\*{1,2}`)

// SpeechDict returns the effective pronunciation dictionary: the built-in
// Python dictionary overlaid with course/lesson pronunciations.
func SpeechDict(pronunciations map[string]string) map[string]string {
	dict := make(map[string]string, len(defaultSpeechDict)+len(pronunciations))
	maps.Copy(dict, defaultSpeechDict)
	maps.Copy(dict, pronunciations)
	return dict
}

// PrepForSpeech rewrites one narration paragraph into speakable text.
// Transformation order matters: emphasis → dictionary → dunders → operators
// → numbers → sentence-length guard → whitespace cleanup.
func PrepForSpeech(text string, dict map[string]string) string {
	text = applyEmphasis(text)
	text = applyDict(text, dict)
	text = dunderRe.ReplaceAllStringFunc(text, func(m string) string {
		inner := strings.Trim(m, "_")
		return "dunder " + strings.ReplaceAll(inner, "_", " ")
	})
	text = powerRe.ReplaceAllString(text, "$1 to the power of $2")
	text = percentRe.ReplaceAllString(text, "$1 percent")
	text = applySymbols(text)
	text = decimalRe.ReplaceAllStringFunc(text, func(m string) string {
		parts := decimalRe.FindStringSubmatch(m)
		return numberWords(parts[1]) + " point " + fractionWords(parts[2])
	})
	text = integerRe.ReplaceAllStringFunc(text, numberWords)
	text = guardSentenceLength(text, maxSentenceWords)
	return cleanupSpaces(text)
}

// applyEmphasis converts *emphasized* spans into micro-pauses: Kokoro has no
// SSML emphasis control, but a comma either side of the span makes the
// narrator lean on it.
func applyEmphasis(text string) string {
	return emphasisRe.ReplaceAllStringFunc(text, func(m string) string {
		inner := strings.Trim(m, "*")
		return ", " + strings.TrimSpace(inner) + ","
	})
}

// applyDict replaces dictionary tokens longest-first, matching only at
// token boundaries (previous/next rune not a letter or digit), so "pip"
// never rewrites the middle of "pipeline". Matching is case-insensitive.
func applyDict(text string, dict map[string]string) string {
	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) > len(keys[j])
		}
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		text = replaceToken(text, k, dict[k])
	}
	return text
}

// replaceToken replaces case-insensitive occurrences of from with to when
// the occurrence is bounded by non-word runes on both sides.
func replaceToken(text, from, to string) string {
	lower := strings.ToLower(text)
	fromLower := strings.ToLower(from)
	var b strings.Builder
	i := 0
	for {
		j := strings.Index(lower[i:], fromLower)
		if j < 0 {
			b.WriteString(text[i:])
			return b.String()
		}
		start := i + j
		end := start + len(from)
		if wordBoundary(text, start, end) {
			b.WriteString(text[i:start])
			b.WriteString(to)
		} else {
			b.WriteString(text[i:end])
		}
		i = end
	}
}

// wordBoundary reports whether text[start:end] is delimited by non-word
// runes (or the string bounds). A leading/trailing letter or digit inside
// the match itself only needs guarding when the match starts/ends with one.
func wordBoundary(text string, start, end int) bool {
	if start > 0 {
		prev, _ := lastRune(text[:start])
		first, _ := firstRune(text[start:end])
		if isWordRune(prev) && isWordRune(first) {
			return false
		}
	}
	if end < len(text) {
		next, _ := firstRune(text[end:])
		last, _ := lastRune(text[start:end])
		if isWordRune(next) && isWordRune(last) {
			return false
		}
	}
	return true
}

// isWordRune treats underscores as word runes so dictionary keys never fire
// inside identifiers like __init__ (the dunder rule owns those).
func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func firstRune(s string) (rune, bool) {
	for _, r := range s {
		return r, true
	}
	return 0, false
}

func lastRune(s string) (rune, bool) {
	var out rune
	found := false
	for _, r := range s {
		out = r
		found = true
	}
	return out, found
}

// applySymbols replaces operator tokens longest-first with padded spaces so
// "x!=y" reads "x not equal to y".
func applySymbols(text string) string {
	keys := make([]string, 0, len(symbolSpeech))
	for k := range symbolSpeech {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) > len(keys[j])
		}
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		if !strings.Contains(text, k) {
			continue
		}
		text = strings.ReplaceAll(text, k, " "+symbolSpeech[k]+" ")
	}
	return text
}

var onesWords = []string{
	"zero", "one", "two", "three", "four", "five", "six", "seven", "eight",
	"nine", "ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen",
	"sixteen", "seventeen", "eighteen", "nineteen",
}

var tensWords = []string{
	"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy",
	"eighty", "ninety",
}

// numberWords spells out a non-negative integer literal. Numbers too large
// to say naturally (≥ a trillion) are read digit by digit.
func numberWords(digits string) string {
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		return "zero"
	}
	if len(digits) > 12 {
		return digitByDigit(digits)
	}
	var n int64
	if _, err := fmt.Sscanf(digits, "%d", &n); err != nil {
		return digitByDigit(digits)
	}
	return intWords(n)
}

func intWords(n int64) string {
	switch {
	case n < 20:
		return onesWords[n]
	case n < 100:
		if n%10 == 0 {
			return tensWords[n/10]
		}
		return tensWords[n/10] + "-" + onesWords[n%10]
	case n < 1000:
		s := onesWords[n/100] + " hundred"
		if rem := n % 100; rem != 0 {
			s += " " + intWords(rem)
		}
		return s
	}
	for _, scale := range []struct {
		value int64
		name  string
	}{
		{1_000_000_000_000, "trillion"},
		{1_000_000_000, "billion"},
		{1_000_000, "million"},
		{1_000, "thousand"},
	} {
		if n >= scale.value {
			s := intWords(n/scale.value) + " " + scale.name
			if rem := n % scale.value; rem != 0 {
				s += " " + intWords(rem)
			}
			return s
		}
	}
	return digitByDigit(fmt.Sprint(n))
}

// fractionWords reads the digits after a decimal point: "11" → "eleven",
// "05" → "zero five" (leading zeros are read out to keep versions exact).
func fractionWords(digits string) string {
	if strings.HasPrefix(digits, "0") {
		return digitByDigit(digits)
	}
	return numberWords(digits)
}

func digitByDigit(digits string) string {
	words := make([]string, 0, len(digits))
	for _, d := range digits {
		if d < '0' || d > '9' {
			continue
		}
		words = append(words, onesWords[d-'0'])
	}
	return strings.Join(words, " ")
}

// sentenceEndRe finds sentence terminators followed by whitespace.
var sentenceEndRe = regexp.MustCompile(`([.!?])\s+`)

// guardSentenceLength splits sentences longer than maxWords at the clause
// boundary (comma, semicolon, colon, or dash) closest to the midpoint.
func guardSentenceLength(text string, maxWords int) string {
	sentences := splitSentences(text)
	out := make([]string, 0, len(sentences))
	for _, s := range sentences {
		out = append(out, splitLongSentence(s, maxWords)...)
	}
	return strings.Join(out, " ")
}

// splitSentences cuts text into sentences, keeping terminators attached.
func splitSentences(text string) []string {
	var sentences []string
	last := 0
	for _, loc := range sentenceEndRe.FindAllStringIndex(text, -1) {
		sentences = append(sentences, strings.TrimSpace(text[last:loc[1]]))
		last = loc[1]
	}
	if rest := strings.TrimSpace(text[last:]); rest != "" {
		sentences = append(sentences, rest)
	}
	return sentences
}

var clauseBoundaryRe = regexp.MustCompile(`([,;:]|\s—|\s–|\s-)\s+`)

// splitLongSentence recursively splits one sentence at the clause boundary
// nearest its midpoint until every piece fits maxWords (or no boundary is
// left to split at).
func splitLongSentence(s string, maxWords int) []string {
	if len(strings.Fields(s)) <= maxWords {
		return []string{s}
	}
	locs := clauseBoundaryRe.FindAllStringIndex(s, -1)
	if len(locs) == 0 {
		return []string{s}
	}
	mid := len(s) / 2
	best := locs[0]
	for _, loc := range locs[1:] {
		if abs(loc[0]-mid) < abs(best[0]-mid) {
			best = loc
		}
	}
	left := strings.TrimRight(strings.TrimSpace(s[:best[0]]), ",;:—–- ")
	right := strings.TrimSpace(s[best[1]:])
	if left == "" || right == "" {
		return []string{s}
	}
	if !strings.ContainsAny(left[len(left)-1:], ".!?") {
		left += "."
	}
	right = capitalizeFirst(right)
	out := splitLongSentence(left, maxWords)
	return append(out, splitLongSentence(right, maxWords)...)
}

func capitalizeFirst(s string) string {
	for i, r := range s {
		return string(unicode.ToUpper(r)) + s[i+len(string(r)):]
	}
	return s
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

var multiSpaceRe = regexp.MustCompile(`[ \t]{2,}`)
var spaceBeforePunctRe = regexp.MustCompile(` +([,.;:!?])`)
var doubledCommaRe = regexp.MustCompile(`,\s*,+`)
var leadingCommaRe = regexp.MustCompile(`(^|[.!?] ),\s*`)

// cleanupSpaces collapses whitespace artifacts left by the substitutions.
func cleanupSpaces(text string) string {
	text = multiSpaceRe.ReplaceAllString(text, " ")
	text = doubledCommaRe.ReplaceAllString(text, ",")
	text = spaceBeforePunctRe.ReplaceAllString(text, "$1")
	text = strings.ReplaceAll(text, ",.", ".")
	text = leadingCommaRe.ReplaceAllString(text, "$1")
	return strings.TrimSpace(text)
}

// SplitParagraphs cuts narration into paragraphs on blank lines. The audio
// stage inserts the configured paragraph pause between them.
func SplitParagraphs(text string) []string {
	var out []string
	for _, p := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n\n") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{strings.TrimSpace(text)}
	}
	return out
}

// StripEmphasis removes *emphasis* markers for text-facing artifacts
// (transcripts, pages) without inserting pauses.
func StripEmphasis(text string) string {
	return emphasisRe.ReplaceAllString(text, "$1")
}

// SpokenSection is the per-section spoken text sent to the TTS.
type SpokenSection struct {
	ID string `json:"id"`
	// Paragraphs are the prepped narration paragraphs, in order; the audio
	// stage inserts the paragraph pause between them.
	Paragraphs []string `json:"paragraphs"`
}

// SpokenText joins the section's paragraphs for WER reference use.
func (s SpokenSection) SpokenText() string {
	return strings.Join(s.Paragraphs, "\n\n")
}

// TTSScript is the persisted tts_script.json artifact.
type TTSScript struct {
	Sections []SpokenSection `json:"sections"`
}
