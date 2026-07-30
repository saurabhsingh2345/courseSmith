package pipeline

// Chapters stage: turns the aligned section boundaries into distribution
// artifacts — chapters.json (machine-readable), chapters.txt (YouTube
// description format), and transcript.md (clean punctuated transcript with
// section headings and timestamps, later a page on the course site).

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// Chapters stage outputs, under the lesson's generated dir.
const (
	ChaptersJSONFileName = "chapters.json"
	ChaptersTxtFileName  = "chapters.txt"
	TranscriptFileName   = "transcript.md"
)

// Chapter is one entry of chapters.json.
type Chapter struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	StartMs int    `json:"start_ms"`
	EndMs   int    `json:"end_ms"`
}

// headingRe is shared with the scenegraph stage (scenegraph.go).
var slugCleanRe = regexp.MustCompile(`[^a-z0-9]+`)

// slugify mirrors the script prompt's rule: lowercase, hyphen-separated.
func slugify(s string) string {
	return strings.Trim(slugCleanRe.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

// sectionTitles maps section ids to human titles by re-slugifying the
// lesson outline's ## headings; sections without a matching heading get a
// humanized version of their id.
func sectionTitles(body string, script *Script) map[string]string {
	bySlug := map[string]string{}
	for _, m := range headingRe.FindAllStringSubmatch(body, -1) {
		title := strings.TrimSpace(m[1])
		bySlug[slugify(title)] = title
	}
	titles := make(map[string]string, len(script.Sections))
	for _, sec := range script.Sections {
		switch {
		// What the section says it is called, when it knows. Assembled scripts
		// (reels, snippets) set this; matching slugs cannot work for them because
		// their ids are composed rather than derived from a heading.
		case strings.TrimSpace(sec.Title) != "":
			titles[sec.ID] = collapseSpaces(sec.Title)
		case bySlug[sec.ID] != "":
			titles[sec.ID] = bySlug[sec.ID]
		default:
			titles[sec.ID] = humanizeSlug(sec.ID)
		}
	}
	return titles
}

// youtubeTimestamp formats milliseconds as M:SS (or H:MM:SS past an hour),
// the format YouTube chapter lists require.
func youtubeTimestamp(ms int) string {
	if ms < 0 {
		ms = 0
	}
	total := ms / 1000
	h, m, s := total/3600, total%3600/60, total%60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// buildChapters derives chapter boundaries from the aligned section spans:
// each chapter runs from its section's first word to the next section's
// first word (the last to the end of audio), and the first always starts at
// zero as YouTube requires.
func buildChapters(alignment *Alignment, titles map[string]string) []Chapter {
	endOfAudio := 0
	if n := len(alignment.Words); n > 0 {
		endOfAudio = alignment.Words[n-1].EndMs
	}
	chapters := make([]Chapter, 0, len(alignment.Sections))
	for si, span := range alignment.Sections {
		start := span.StartMs
		if si == 0 {
			start = 0
		}
		end := endOfAudio
		if si+1 < len(alignment.Sections) {
			end = alignment.Sections[si+1].StartMs
		}
		if end < start {
			end = start
		}
		chapters = append(chapters, Chapter{
			ID:      span.ID,
			Title:   titles[span.ID],
			StartMs: start,
			EndMs:   end,
		})
	}
	return chapters
}

// chaptersTxt renders the YouTube-description chapter list.
func chaptersTxt(chapters []Chapter) string {
	var b strings.Builder
	for _, c := range chapters {
		fmt.Fprintf(&b, "%s %s\n", youtubeTimestamp(c.StartMs), c.Title)
	}
	return b.String()
}

// transcriptMD renders the clean punctuated transcript: the written
// narration (emphasis markers stripped) under timestamped section headings.
func transcriptMD(script *Script, chapters []Chapter) string {
	starts := make(map[string]int, len(chapters))
	for _, c := range chapters {
		starts[c.ID] = c.StartMs
	}
	titles := make(map[string]string, len(chapters))
	for _, c := range chapters {
		titles[c.ID] = c.Title
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", script.Title)
	for _, sec := range script.Sections {
		fmt.Fprintf(&b, "\n## %s [%s]\n\n", titles[sec.ID], youtubeTimestamp(starts[sec.ID]))
		for _, para := range SplitParagraphs(sec.Narration) {
			fmt.Fprintf(&b, "%s\n\n", StripEmphasis(para))
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// runChaptersStage is the "chapters" stage: script.json + alignment.json →
// chapters.json, chapters.txt, transcript.md.
func runChaptersStage(_ context.Context, e *Env, _ *project.Course, l *project.Lesson, _ config.Config) error {
	script, err := loadScript(l)
	if err != nil {
		return err
	}
	alignment, err := loadAlignment(l)
	if err != nil {
		return err
	}

	titles := sectionTitles(l.Body, script)
	chapters := buildChapters(alignment, titles)

	if err := writeJSON(filepath.Join(l.GeneratedDir(), ChaptersJSONFileName), chapters); err != nil {
		return err
	}
	if err := writeFileAtomic(filepath.Join(l.GeneratedDir(), ChaptersTxtFileName), []byte(chaptersTxt(chapters))); err != nil {
		return err
	}
	if err := writeFileAtomic(filepath.Join(l.GeneratedDir(), TranscriptFileName), []byte(transcriptMD(script, chapters))); err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "  → chapters  %d chapter(s) → %s, %s, %s\n",
		len(chapters), ChaptersJSONFileName, ChaptersTxtFileName, TranscriptFileName)
	return nil
}
