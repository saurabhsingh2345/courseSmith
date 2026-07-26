package config

import (
	"reflect"
	"testing"
)

func TestDefaultsAreComplete(t *testing.T) {
	d := Defaults()
	if d.Style.Voice == "" || d.Style.Tone == "" || d.Style.PaceWPM == 0 ||
		d.Style.Audience == "" || d.Style.Language == "" {
		t.Errorf("Defaults() has zero-valued Style fields: %+v", d.Style)
	}
	if d.Branding.Colors.Primary == "" || d.Branding.Colors.Accent == "" ||
		d.Branding.Colors.Background == "" || d.Branding.DiagramStyle == "" {
		t.Errorf("Defaults() has zero-valued Branding fields: %+v", d.Branding)
	}
	if d.Pipeline.LLMContent == "" || d.Pipeline.LLMReview == "" ||
		d.Pipeline.ReviewThreshold == 0 || d.Pipeline.CaptionsModel == "" {
		t.Errorf("Defaults() has zero-valued Pipeline fields: %+v", d.Pipeline)
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name string
		base Config
		over Config
		want func(t *testing.T, got Config)
	}{
		{
			name: "zero overlay inherits everything",
			base: Defaults(),
			over: Config{},
			want: func(t *testing.T, got Config) {
				if !reflect.DeepEqual(got, Defaults()) {
					t.Errorf("got %+v, want defaults", got)
				}
			},
		},
		{
			name: "set fields win, unset inherit",
			base: Defaults(),
			over: Config{
				Style:    Style{Tone: "dry and witty", PaceWPM: 170},
				Pipeline: Pipeline{ReviewThreshold: 9},
			},
			want: func(t *testing.T, got Config) {
				if got.Style.Tone != "dry and witty" {
					t.Errorf("Tone = %q, want overlay value", got.Style.Tone)
				}
				if got.Style.PaceWPM != 170 {
					t.Errorf("PaceWPM = %d, want 170", got.Style.PaceWPM)
				}
				if got.Pipeline.ReviewThreshold != 9 {
					t.Errorf("ReviewThreshold = %v, want 9", got.Pipeline.ReviewThreshold)
				}
				if got.Style.Voice != Defaults().Style.Voice {
					t.Errorf("Voice = %q, want inherited default", got.Style.Voice)
				}
				if got.Pipeline.LLMContent != Defaults().Pipeline.LLMContent {
					t.Errorf("LLMContent = %q, want inherited default", got.Pipeline.LLMContent)
				}
			},
		},
		{
			name: "nested colors merge field by field",
			base: Defaults(),
			over: Config{Branding: Branding{Colors: Colors{Accent: "#ff0000"}}},
			want: func(t *testing.T, got Config) {
				if got.Branding.Colors.Accent != "#ff0000" {
					t.Errorf("Accent = %q, want #ff0000", got.Branding.Colors.Accent)
				}
				if got.Branding.Colors.Primary != Defaults().Branding.Colors.Primary {
					t.Errorf("Primary = %q, want inherited default", got.Branding.Colors.Primary)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Merge(tt.base, tt.over)
			tt.want(t, got)
		})
	}
}

func TestMergeDoesNotMutateInputs(t *testing.T) {
	base := Defaults()
	over := Config{Style: Style{Voice: "am_adam"}}
	_ = Merge(base, over)
	if base.Style.Voice != Defaults().Style.Voice {
		t.Errorf("Merge mutated base: Voice = %q", base.Style.Voice)
	}
}

func TestResolveLayerPrecedence(t *testing.T) {
	course := Config{Style: Style{Tone: "course tone", Voice: "course voice"}}
	lesson := Config{Style: Style{Tone: "lesson tone"}}
	flags := Config{Pipeline: Pipeline{LLMContent: "openai/gpt-4o-mini"}}

	got := Resolve(course, lesson, flags)

	if got.Style.Tone != "lesson tone" {
		t.Errorf("Tone = %q, want lesson layer to beat course layer", got.Style.Tone)
	}
	if got.Style.Voice != "course voice" {
		t.Errorf("Voice = %q, want course layer to beat defaults", got.Style.Voice)
	}
	if got.Pipeline.LLMContent != "openai/gpt-4o-mini" {
		t.Errorf("LLMContent = %q, want flags layer to win", got.Pipeline.LLMContent)
	}
	if got.Style.PaceWPM != Defaults().Style.PaceWPM {
		t.Errorf("PaceWPM = %d, want default to survive all layers", got.Style.PaceWPM)
	}
}
