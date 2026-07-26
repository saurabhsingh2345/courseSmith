package pipeline

// Voice auditions: render one fixed paragraph in every Kokoro voice that
// matches the course language, side by side in an HTML page, so a human can
// pick the course voice by ear.

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// AuditionsDirName is created under the course directory.
const AuditionsDirName = "auditions"

// auditionParagraph exercises normal prose, a code term, a number, and a
// question so voices can be compared on the sounds a course actually uses.
const auditionParagraph = `Hi there, and welcome to the course. In the next few lessons we'll write
real Python together — small programs first, bigger ideas soon after.
Python 3.12 makes this easier than ever: you'll meet f-strings, the
__init__ method, and the print function. Ready? Let's find out what your
first line of code can do.`

// kokoroLangPrefixes maps a course language to Kokoro voice id prefixes
// (first letter of the voice id encodes language: a/b = US/UK English,
// e = Spanish, f = French, h = Hindi, i = Italian, j = Japanese,
// p = Portuguese, z = Mandarin).
var kokoroLangPrefixes = map[string][]string{
	"en": {"a", "b"},
	"es": {"e"},
	"fr": {"f"},
	"hi": {"h"},
	"it": {"i"},
	"ja": {"j"},
	"pt": {"p"},
	"zh": {"z"},
}

// ListVoices asks the Kokoro server for its installed voice ids.
func (e *Env) ListVoices(ctx context.Context) ([]string, error) {
	url := e.ttsBaseURL() + "/audio/voices"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot reach the Kokoro TTS server at %s — %s\n    (%v)", e.ttsBaseURL(), ttsStartHelp, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading voices response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Kokoro server returned HTTP %d for %s", resp.StatusCode, url)
	}

	// Kokoro-FastAPI variants: {"voices":[{"id":...}]} or {"voices":["id"]}.
	var withObjects struct {
		Voices []struct {
			ID string `json:"id"`
		} `json:"voices"`
	}
	if err := json.Unmarshal(body, &withObjects); err == nil && len(withObjects.Voices) > 0 && withObjects.Voices[0].ID != "" {
		out := make([]string, 0, len(withObjects.Voices))
		for _, v := range withObjects.Voices {
			out = append(out, v.ID)
		}
		return out, nil
	}
	var withStrings struct {
		Voices []string `json:"voices"`
	}
	if err := json.Unmarshal(body, &withStrings); err == nil && len(withStrings.Voices) > 0 {
		return withStrings.Voices, nil
	}
	return nil, fmt.Errorf("could not parse voice list from %s: %s", url, truncate(string(body), 200))
}

// filterVoicesForLanguage keeps voices whose id prefix matches the course
// language; unknown languages keep everything.
func filterVoicesForLanguage(voices []string, language string) []string {
	lang := strings.ToLower(language)
	if i := strings.IndexAny(lang, "-_"); i > 0 {
		lang = lang[:i]
	}
	prefixes, ok := kokoroLangPrefixes[lang]
	if !ok {
		return voices
	}
	var out []string
	for _, v := range voices {
		for _, p := range prefixes {
			if strings.HasPrefix(v, p) {
				out = append(out, v)
				break
			}
		}
	}
	return out
}

// auditionPage is the HTML template for the side-by-side player page.
var auditionPage = template.Must(template.New("auditions").Parse(`<!doctype html>
<meta charset="utf-8">
<title>Voice auditions — {{.Course}}</title>
<style>
  body { font-family: system-ui, sans-serif; margin: 2rem auto; max-width: 48rem; padding: 0 1rem; }
  h1 { font-size: 1.4rem; } p.hint { color: #555; }
  .voice { display: flex; align-items: center; gap: 1rem; padding: .6rem .8rem;
           border: 1px solid #ddd; border-radius: .5rem; margin: .5rem 0; }
  .voice code { min-width: 10rem; font-size: 1rem; }
  .voice.current { border-color: #2563eb; background: #eff6ff; }
  audio { flex: 1; }
</style>
<h1>Voice auditions — {{.Course}}</h1>
<p class="hint">Every Kokoro voice matching language <code>{{.Language}}</code>, reading the same
paragraph. Pick one, then run:<br>
<code>coursesmith audition {{.Slug}} --choose &lt;voice&gt;</code> to write it to course.yaml.</p>
{{range .Voices}}
<div class="voice{{if .Current}} current{{end}}">
  <code>{{.ID}}{{if .Current}} ✓{{end}}</code>
  <audio controls preload="none" src="{{.File}}"></audio>
</div>
{{end}}
`))

type auditionVoice struct {
	ID      string
	File    string
	Current bool
}

// RunAudition renders the audition paragraph in every matching voice into
// courses/<slug>/auditions/ and writes an index.html player page. Existing
// recordings are kept, so re-runs only fill gaps.
func RunAudition(ctx context.Context, e *Env, course *project.Course, cfg config.Config) (string, error) {
	voices, err := e.ListVoices(ctx)
	if err != nil {
		return "", err
	}
	matching := filterVoicesForLanguage(voices, cfg.Style.Language)
	if len(matching) == 0 {
		return "", fmt.Errorf("no Kokoro voices match language %q (server has %d voices)", cfg.Style.Language, len(voices))
	}
	sort.Strings(matching)

	outDir := filepath.Join(course.Dir, AuditionsDirName)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", outDir, err)
	}

	text := PrepForSpeech(strings.Join(strings.Fields(auditionParagraph), " "), SpeechDict(cfg.Style.Pronunciations))
	fmt.Fprintf(e.out(), "auditioning %d voice(s) for language %q...\n", len(matching), cfg.Style.Language)
	page := make([]auditionVoice, 0, len(matching))
	for _, voice := range matching {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		wavPath := filepath.Join(outDir, voice+".wav")
		if _, err := os.Stat(wavPath); err != nil {
			fmt.Fprintf(e.out(), "  → %s\n", voice)
			wav, err := e.ttsSpeak(ctx, voice, text, cfg.Style.VoiceSpeed)
			if err != nil {
				return "", fmt.Errorf("voice %s: %w", voice, err)
			}
			if err := writeFileAtomic(wavPath, wav); err != nil {
				return "", err
			}
		} else {
			fmt.Fprintf(e.out(), "  ✓ %s (already rendered)\n", voice)
		}
		page = append(page, auditionVoice{ID: voice, File: voice + ".wav", Current: voice == cfg.Style.Voice})
	}

	var sb strings.Builder
	err = auditionPage.Execute(&sb, map[string]any{
		"Course":   course.Name,
		"Slug":     course.Slug,
		"Language": cfg.Style.Language,
		"Voices":   page,
	})
	if err != nil {
		return "", fmt.Errorf("rendering audition page: %w", err)
	}
	indexPath := filepath.Join(outDir, "index.html")
	if err := writeFileAtomic(indexPath, []byte(sb.String())); err != nil {
		return "", err
	}
	return indexPath, nil
}

var voiceLineRe = regexp.MustCompile(`(?m)^(\s*voice:\s*)\S+`)

// ChooseVoice writes the picked voice into the course's course.yaml,
// replacing the existing style.voice line in place.
func ChooseVoice(course *project.Course, voice string) error {
	path := filepath.Join(course.Dir, project.CourseFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	if voiceLineRe.Match(data) {
		data = voiceLineRe.ReplaceAll(data, []byte("${1}"+voice))
	} else if styleRe := regexp.MustCompile(`(?m)^style:\s*$`); styleRe.Match(data) {
		data = styleRe.ReplaceAll(data, []byte("style:\n  voice: "+voice))
	} else {
		data = append(data, []byte("\nstyle:\n  voice: "+voice+"\n")...)
	}
	return writeFileAtomic(path, data)
}
