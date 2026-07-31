package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const goodTake = `
site: lovable
viewport: {width: 1440, height: 900}
steps:
  - do: goto
    path: /
  - do: wait
    selector: "[data-testid=prompt]"
    timeout: 30s
  - do: shot
    mark: landing
    focus: "[data-testid=prompt]"
  - do: type
    selector: "[data-testid=prompt]"
    text: a habit tracker with streaks
  - do: shot
    mark: prompt-typed
`

func takeFrom(t *testing.T, yaml string) *WebTake {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "take.yaml")
	if err := os.WriteFile(p, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	take, err := LoadWebTake(p)
	if err != nil {
		t.Fatalf("loading take: %v", err)
	}
	return take
}

func TestLoadWebTake(t *testing.T) {
	take := takeFrom(t, goodTake)
	if take.Site != "lovable" {
		t.Errorf("site = %q", take.Site)
	}
	if take.Viewport.Width != 1440 || take.Viewport.Height != 900 {
		t.Errorf("viewport = %+v", take.Viewport)
	}
	if len(take.Steps) != 5 {
		t.Fatalf("steps = %d", len(take.Steps))
	}
	if countShots(take) != 2 {
		t.Errorf("shots = %d, want 2", countShots(take))
	}
	if len(take.Steps[2].Focus) != 1 || take.Steps[2].Focus[0] != "[data-testid=prompt]" {
		t.Errorf("focus = %v", take.Steps[2].Focus)
	}
}

func TestWebTakeValidation(t *testing.T) {
	for _, tt := range []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{"accepts a real take", goodTake, ""},
		{
			"rejects an unknown site",
			"site: photoshop\nsteps:\n  - do: shot\n    mark: a\n",
			"not recordable",
		},
		{
			"rejects an unknown step",
			"site: lovable\nsteps:\n  - do: teleport\n",
			"is not a step",
		},
		{
			"rejects a take with no shots",
			"site: lovable\nsteps:\n  - do: goto\n    path: /\n",
			"records nothing",
		},
		{
			"rejects a shot with no mark",
			"site: lovable\nsteps:\n  - do: shot\n",
			"needs a mark",
		},
		{
			"rejects a duplicate mark",
			"site: lovable\nsteps:\n  - do: shot\n    mark: a\n  - do: shot\n    mark: a\n",
			"used twice",
		},
		{
			"rejects an upper-case mark",
			"site: lovable\nsteps:\n  - do: shot\n    mark: Landing\n",
			"lowercase",
		},
		{
			"rejects a click with no selector",
			"site: lovable\nsteps:\n  - do: click\n  - do: shot\n    mark: a\n",
			"needs a selector",
		},
		{
			"rejects type with no text",
			"site: lovable\nsteps:\n  - do: type\n    selector: x\n  - do: shot\n    mark: a\n",
			"needs text",
		},
		{
			"rejects a bad timeout",
			"site: lovable\nsteps:\n  - do: wait\n    selector: x\n    timeout: soon\n  - do: shot\n    mark: a\n",
			"not a duration",
		},
		{
			"rejects an empty take",
			"site: lovable\n",
			"no steps",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, "take.yaml")
			if err := os.WriteFile(p, []byte(tt.yaml), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := LoadWebTake(p)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("accepted, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

// A selector may be one string or a list tried in order. Fallbacks are what let
// a take outlive a redesign — the precise selector first, a looser one behind
// it — and the scalar form has to keep working, because most steps need one and
// a list of one is noise.
func TestSelectorAcceptsScalarOrList(t *testing.T) {
	take := takeFrom(t, `
site: lovable
steps:
  - do: wait
    selector: "textarea"
  - do: click
    selector:
      - "[data-testid=submit]"
      - "button[type=submit]"
      - "button"
  - do: shot
    mark: a
`)
	if got := take.Steps[0].Selector; len(got) != 1 || got[0] != "textarea" {
		t.Errorf("scalar selector = %v", got)
	}
	got := take.Steps[1].Selector
	if len(got) != 3 || got[0] != "[data-testid=submit]" || got[2] != "button" {
		t.Errorf("list selector = %v", got)
	}
	// The error a broken take produces has to name every alternative, so the
	// person repairing it knows what was already ruled out.
	if s := got.String(); !strings.Contains(s, "[data-testid=submit]") || !strings.Contains(s, "button") {
		t.Errorf("selector renders as %s", s)
	}
}

// The web half of the provenance gate. footage.json records the origin as
// evidence that the frames really are of the product they claim, so a take that
// navigates off that origin would produce a clip whose stated provenance is
// false — which is the exact failure this whole track exists to prevent.
func TestWebTakeCannotLeaveItsOrigin(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "take.yaml")
	yaml := "site: lovable\nsteps:\n  - do: goto\n    path: https://example.com/pricing\n  - do: shot\n    mark: a\n"
	if err := os.WriteFile(p, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadWebTake(p)
	if err == nil {
		t.Fatal("a take navigating off its origin was accepted")
	}
	if !strings.Contains(err.Error(), "leaves https://lovable.dev") {
		t.Errorf("error = %v", err)
	}

	// A relative path is always fine, and an absolute URL on the same origin is
	// fine — people write both and neither is a provenance problem.
	for _, path := range []string{"/projects", "https://lovable.dev/projects"} {
		yaml := "site: lovable\nsteps:\n  - do: goto\n    path: " + path + "\n  - do: shot\n    mark: a\n"
		if err := os.WriteFile(p, []byte(yaml), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadWebTake(p); err != nil {
			t.Errorf("path %q was rejected: %v", path, err)
		}
	}
}

// A web mark and a tape mark must be the same vocabulary. Two spellings of the
// same idea is how a downstream consumer ends up handling one and silently
// dropping the other.
func TestWebMarksUseTheSameVocabularyAsTapeMarks(t *testing.T) {
	for _, name := range []string{"landing", "app-built", "step-2"} {
		if !markNameRe.MatchString(name) {
			t.Errorf("web rejects %q", name)
		}
		if !markCommentRe.MatchString("# MARK " + name) {
			t.Errorf("a tape rejects %q", name)
		}
	}
	for _, name := range []string{"Landing", "app_built", "-leading"} {
		if markNameRe.MatchString(name) {
			t.Errorf("web accepts %q", name)
		}
		if markCommentRe.MatchString("# MARK " + name) {
			t.Errorf("a tape accepts %q", name)
		}
	}
}

// Every take checked into the repo must load and validate.
//
// A take is the one artifact in this system a person writes by hand, against
// somebody else's markup, and it is not exercised until a capture runs — which
// needs a browser, a login, and several minutes. A typo in a step name would
// otherwise sit there until the day somebody tried to shoot the lesson.
func TestShippedTakesAreValid(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "courses", "*", "takes", "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Skip("no takes checked in yet")
	}
	for _, path := range matches {
		t.Run(filepath.Base(path), func(t *testing.T) {
			// A take file is one kind or the other and says so in its first
			// key. Trying web first and falling back would report the web
			// parser's complaint about a desktop take, which is a confusing
			// way to be told about a typo.
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(raw), "\napp:") || strings.HasPrefix(string(raw), "app:") {
				take, err := LoadDesktopTake(path)
				if err != nil {
					t.Fatalf("%s does not validate: %v", path, err)
				}
				if _, ok := captureApps[take.App]; !ok {
					t.Errorf("app %q is no longer recordable", take.App)
				}
				return
			}
			take, err := LoadWebTake(path)
			if err != nil {
				t.Fatalf("%s does not validate: %v", path, err)
			}
			// A take nobody can re-shoot from is a take that rots silently, so
			// the site it names has to still be one we know how to drive.
			if _, ok := captureSites[take.Site]; !ok {
				t.Errorf("site %q is no longer recordable", take.Site)
			}
		})
	}
}

// A key may not appear in both registries: the marker's `tool=` looks in both,
// so a collision would make which one you get depend on lookup order.
func TestToolAndSiteKeysDoNotCollide(t *testing.T) {
	for key := range captureSites {
		if _, clash := captureTools[key]; clash {
			t.Errorf("%q is both a terminal tool and a web site", key)
		}
	}
}

func TestCaptureMarkerRoutesToTheRightKind(t *testing.T) {
	specs, err := extractCaptureMarkers(
		"[CAPTURE: tool=claude; agent stuff]\n[CAPTURE: tool=lovable, take=first-build; one sentence in]\n")
	if err != nil {
		t.Fatal(err)
	}
	if specs[0].Kind != CaptureKindTool {
		t.Errorf("claude routed to %q", specs[0].Kind)
	}
	if specs[1].Kind != CaptureKindWeb || specs[1].Take != "first-build" {
		t.Errorf("lovable spec = %+v", specs[1])
	}
}

// The two kinds take different attributes, and using one kind's attribute on
// the other is a mistake worth naming rather than ignoring — a `fixture=` on a
// web capture means somebody expected a working directory that does not exist.
func TestCaptureMarkerRejectsMismatchedAttributes(t *testing.T) {
	if _, err := extractCaptureMarkers("[CAPTURE: tool=lovable; no take named]\n"); err == nil {
		t.Error("a web capture with no take was accepted")
	}
	_, err := extractCaptureMarkers("[CAPTURE: tool=claude, take=x; a tape is generated]\n")
	if err == nil || !strings.Contains(err.Error(), "no take file") {
		t.Errorf("a terminal capture with a take= was accepted or misreported: %v", err)
	}
	_, err = extractCaptureMarkers("[CAPTURE: tool=lovable, take=x, fixture=y; no workdir]\n")
	if err == nil || !strings.Contains(err.Error(), "no working directory") {
		t.Errorf("a web capture with a fixture= was accepted or misreported: %v", err)
	}
}
