package pipeline

// Narration has to survive being SPOKEN, and that is a different standard from
// reading well.
//
// == Why this exists ==
//
// The catalog's prompts were written against a small model whose failure was
// vagueness, so they push relentlessly for specificity. Put a competent model
// behind them and the failure inverts: it writes something factually excellent
// and completely unspeakable —
//
//	"ANN indexes trade perfect recall for speed; depending on tuning and data,
//	 recall and latency vary—test on your dataset."
//
// Every clause is true. Nobody has ever said it out loud. It goes to a TTS voice
// that reads a semicolon as a full stop and an em-dash as a stumble, and the
// viewer hears a paragraph where they needed a sentence.
//
// == Why these three rules and not the seven that were tried ==
//
// Guidance in a prompt with nothing checking it is read as decoration — this
// repo learned that once already with the cast arc, where the structural advice
// had always been there and was ignored until validateCastArc made it
// checkable. So the rules are enforced. But a rule that fires on good work costs
// a correction round and real money, so each was measured against 476 sentences
// of shipped narration before being allowed in:
//
//	rule                    fires on shipped   fires on the new output
//	semicolon                     0.6%                 11.1%
//	em/en dash                    1.1%                 14.8%
//	over 30 words                 0.0%                  0.0%
//	hedge word                    1.7%                 10.0%
//	4+ clause breaks              2.9%                  3.7%
//	3+ item list                  5.7%                 12.5%
//	over 20 words AND 3+ breaks   2.1%                  3.7%
//
// Only the first three separate the two corpora, by eighteen-, thirteen- and
// sixfold respectively.
// The length rules do not discriminate at all: shipped narration runs a median
// of ELEVEN words per sentence and never exceeds thirty, so a ceiling high
// enough to be safe never fires, and one low enough to fire is a coin toss.
// They are left to spokenVoiceAdvice, where being ignored costs nothing.
//
// The ~1-2% of shipped sentences these three would have caught are genuinely
// rewritable as speech. That is the standard, not collateral damage.

import (
	"fmt"
	"regexp"
	"strings"
)

// hedges are the words a writer reaches for instead of finding out.
//
// Admitted on the same measured basis as the punctuation rules: they appear in
// 1.7% of shipped narration and 10.0% of the first drafts this catalog now
// produces — a sixfold separation. A three-or-more-item list was tested at the
// same time and REJECTED: 5.7% against 12.5% is barely a doubling, and a rule
// that fires on one good sentence in eighteen is not worth a correction round.
//
// == The tension this creates, deliberately ==
//
// Banning the hedge pushes toward false precision, and the plan reviewer scores
// fabrication with a fatal floor. That is the point. Together they leave exactly
// one way out: say the thing you actually know. "Recall depends on tuning" and
// an invented "recall is 94%" are both closed off, and what is left is either a
// real figure or an honest direction — "you trade a little accuracy for a lot of
// speed" — which is what the viewer needed in the first place.
var hedges = regexp.MustCompile(`(?i)\b(typically|roughly|generally|usually|often|depending on|in many cases|may vary|can vary|tends? to|approximately|somewhat|relatively)\b`)

// validateSpokenVoice rejects narration that reads as written prose.
//
// Checked with the template's editorial rules rather than in validateShape,
// because it is a standard the model is held to while it still has correction
// rounds — not a fact about the file. A plan salvaged deliberately loose must
// still load and render.
func validateSpokenVoice(p *SnippetPlan) error {
	for _, b := range p.Beats {
		if i := strings.IndexRune(b.Narration, ';'); i >= 0 {
			return fmt.Errorf("beat %q joins clauses with a semicolon: %q. This is narration — it will be SPOKEN, and a voice reads a semicolon as a full stop, so the sentence lands as two halves with no join. Write two sentences, or one shorter one",
				b.ID, excerptAround(b.Narration, i))
		}
		if i := strings.IndexAny(b.Narration, "—–"); i >= 0 {
			return fmt.Errorf("beat %q uses a dash to bolt a clause on: %q. A dash is a punctuation mark for the eye; spoken, it is a stumble. Say it as its own sentence, or fold it in with a plain word like \"and\" or \"because\"",
				b.ID, excerptAround(b.Narration, i))
		}
		if m := hedges.FindStringIndex(b.Narration); m != nil {
			return fmt.Errorf("beat %q hedges with %q: %q. A hedge is what gets written when the work of finding out was not done, and it leaves the viewer with nothing they can act on. Do NOT swap it for a number you have not got — that is a fabrication and it is scored separately. Either state the figure you actually have, or say plainly which way it goes and why: \"you give up a little accuracy for a lot of speed\"",
				b.ID, b.Narration[m[0]:m[1]], excerptAround(b.Narration, m[0]))
		}
	}
	return nil
}

// excerptAround quotes the offending stretch rather than the whole beat, so the
// correction round is pointed at the clause instead of re-litigating the beat.
func excerptAround(s string, i int) string {
	const pad = 45
	lo, hi := max(0, i-pad), min(len(s), i+pad)
	out := s[lo:hi]
	if lo > 0 {
		out = "..." + out
	}
	if hi < len(s) {
		out += "..."
	}
	return out
}

// spokenVoiceAdvice is appended to every template's prompt centrally, for the
// reason the budget arithmetic is: it is shared guidance rather than a property
// of any one look, and forty-four copies means the forty-fifth is wrong.
//
// It carries what could NOT be made into a rule — the failures that are real but
// not mechanically separable from good writing. Being advice, ignoring it costs
// nothing; the three things above that are worth a correction round are enforced.
func spokenVoiceAdvice() string {
	return "\n\nHOW IT HAS TO SOUND. This narration is spoken aloud by a synthetic voice over a moving picture. It is not read off a page, and nobody can rewind a sentence they lost halfway through.\n" +
		"- One idea per sentence. Shipped narration in this catalog runs about ELEVEN words a sentence; if yours is running to twenty, you are writing an article.\n" +
		"- No semicolons and no dashes. Both are rejected. They are marks for the eye — a voice reads them as a stop or a stumble.\n" +
		"- Say the thing before you name it. \"It keeps a shortcut so it does not have to compare against everything\" earns the word HNSW; leading with the acronym spends the viewer's attention on a term they cannot use yet.\n" +
		"- Do not hedge. \"typically\", \"roughly\", \"often\", \"depending on\" and their relatives are rejected. They tell the viewer nothing they can act on. Do not replace one with an invented number either — fabrication is scored separately and fatally. Give the figure you have, or say which way it goes and why.\n" +
		"- Do not stack a list of three technical nouns in one breath. Pick the one that carries the point and drop the others.\n" +
		"- Read your last beat back as speech before you answer. If you would not say it to somebody sitting next to you, rewrite it."
}
