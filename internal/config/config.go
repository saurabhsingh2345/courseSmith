// Package config defines the merged configuration model for coursesmith.
//
// Configuration is layered, later layers winning field-by-field:
//
//	defaults < course.yaml < lesson front-matter < CLI flags
//
// Every layer is expressed as a Config value; zero-valued fields in an
// overlay mean "inherit from the layer below".
package config

import "maps"

// Config is the fully merged configuration for one pipeline run of one lesson.
type Config struct {
	Style    Style    `yaml:"style"`
	Branding Branding `yaml:"branding"`
	Pipeline Pipeline `yaml:"pipeline"`
	Audio    Audio    `yaml:"audio"`
}

// Style controls the authorial voice of generated content.
type Style struct {
	// Voice is the TTS voice id, e.g. "af_heart". Kokoro also accepts
	// weighted blends: "af_bella(2)+af_sky(1)" mixes voice embeddings.
	Voice string `yaml:"voice"`
	// VoiceSpeed is the TTS speaking-rate multiplier (0.5–2.0; 0 = 1.0).
	// Composes with the auto-pace correction the align stage computes.
	VoiceSpeed float64 `yaml:"voice_speed"`
	Tone       string  `yaml:"tone"`     // e.g. "friendly, conversational"
	PaceWPM    int     `yaml:"pace_wpm"` // narration pace, words per minute
	Audience   string  `yaml:"audience"` // e.g. "absolute beginners"
	Language   string  `yaml:"language"` // BCP-47-ish, e.g. "en"
	// Archetype selects a course preset (project-based, concept-first,
	// practical-skills, story-driven, reference) that supplies motion,
	// palette, and prompt-hint defaults. See internal/pipeline/archetypes.go.
	Archetype string `yaml:"archetype,omitempty"`
	// AnimationStyle is the motion philosophy: minimal | smooth | playful.
	AnimationStyle string `yaml:"animation_style,omitempty"`
	// ColorPalette is a named palette: corporate | warm | cool | colorblind.
	ColorPalette string `yaml:"color_palette,omitempty"`
	// Pronunciations extends the built-in speech dictionary: written form →
	// spoken form, applied to narration before TTS (e.g. "NumPy": "num pie").
	Pronunciations map[string]string `yaml:"pronunciations"`
	// Captions controls the on-screen karaoke caption track: "" or "off"
	// keeps it out of the video (the default — the captions.vtt sidecar is
	// still generated for players and uploads), "on" burns it in.
	// omitempty so the field's introduction doesn't change config
	// fingerprints recorded before it existed.
	Captions string `yaml:"captions,omitempty"`
	// Mode is the video's light/dark polarity: "" or "dark" is the default
	// editorial look, "light" derives a paper-white counterpart from the same
	// branding colours. It changes only the derived tokens — every scene reads
	// those, so no scene knows which mode it is rendering in.
	// omitempty so the field's introduction doesn't change config
	// fingerprints recorded before it existed.
	Mode string `yaml:"mode,omitempty"`
	// Skin is the house style the video is cut in: "" or "default" is the look
	// the catalog has always had, "broadcast" is the near-black explainer look
	// (standing chrome, large uppercase headlines, one small precise diagram in
	// a lot of air), "minimal" is the flat single-accent look where the diagram
	// is the whole frame. It is an axis independent of Mode — every skin
	// derives in both polarities. See internal/pipeline/videoskin.go.
	// omitempty so the field's introduction doesn't change config
	// fingerprints recorded before it existed.
	Skin string `yaml:"skin,omitempty"`
	// Watermark is the standing corner mark a chrome-carrying skin sets on
	// every frame. Empty falls back to the course name.
	// omitempty so the field's introduction doesn't change config
	// fingerprints recorded before it existed.
	Watermark string `yaml:"watermark,omitempty"`
}

// Audio controls voiceover post-production in the audio stage.
type Audio struct {
	// SectionPauseMs is the silence inserted between sections (default 700).
	SectionPauseMs int `yaml:"section_pause_ms"`
	// ParagraphPauseMs is the silence inserted between narration paragraphs
	// within a section (default 350).
	ParagraphPauseMs int `yaml:"paragraph_pause_ms"`
	// SentencePauseMs is the minimum silence at every sentence end (default
	// 400). Unlike the two above it is not synthesized in — a paragraph is
	// still read as one continuous take, so the voice keeps its intonation
	// across full stops, and the align stage widens the gaps afterwards to
	// this floor. It is a floor, not an addition: a sentence end that already
	// breathes for longer is left alone, so the rhythm stays the narrator's
	// rather than becoming metronomic.
	//
	// Capped by the align stage's long-gap compression (1500ms), which would
	// otherwise squash anything larger back down. Set -1 to turn it off (0
	// means "inherit", as it does for every other field here).
	//
	// omitempty so the field's introduction doesn't change config
	// fingerprints recorded before it existed.
	SentencePauseMs int `yaml:"sentence_pause_ms,omitempty"`
	// CrossfadeMs is the fade length at every audio join (default 50).
	CrossfadeMs int `yaml:"crossfade_ms"`
	// TargetLUFS is the integrated loudness target for the two-pass
	// normalization (default -16, the podcast/YouTube voice standard).
	TargetLUFS float64 `yaml:"target_lufs"`
	// MusicBed mixes courses/<slug>/assets/music/*.mp3 under the voice,
	// ducked via sidechain compression. Off by default.
	MusicBed bool `yaml:"music_bed"`
	// MusicDuckDB is how far the bed sits below the voice while the narrator
	// speaks (default -18).
	MusicDuckDB float64 `yaml:"music_duck_db"`
}

// Branding controls the look of generated diagrams and pages.
type Branding struct {
	Colors       Colors `yaml:"colors"`
	Fonts        Fonts  `yaml:"fonts"`
	DiagramStyle string `yaml:"diagram_style"` // e.g. "clean, flat, rounded"
}

// Fonts overrides the video type stack. Names must be Google Fonts families
// bundled by the renderer (see renderer/src/theme/fonts.ts). Empty fields
// inherit the defaults (Space Grotesk / Inter / JetBrains Mono).
type Fonts struct {
	Display string `yaml:"display" json:"display,omitempty"`
	Body    string `yaml:"body" json:"body,omitempty"`
	Mono    string `yaml:"mono" json:"mono,omitempty"`
}

// Colors are CSS color values injected into SVG <style> blocks.
type Colors struct {
	Primary    string `yaml:"primary" json:"primary"`
	Accent     string `yaml:"accent" json:"accent"`
	Background string `yaml:"background" json:"background"`
}

// Pipeline selects models and thresholds for the generation stages.
// Model references use "provider/model" form, e.g. "openai/gpt-4o-mini".
type Pipeline struct {
	LLMContent string `yaml:"llm_content"`
	LLMReview  string `yaml:"llm_review"`
	// LLMVision judges rendered diagram screenshots. Vision spatial reasoning
	// needs a stronger model than text review — a weak judge reports overlaps
	// that are not there on clean, layout-engine-produced diagrams. Empty
	// falls back to LLMReview.
	LLMVision string `yaml:"llm_vision"`
	// LLMSearch grounds the substance stage in real sources. It must name a
	// search-capable OpenAI model — an ordinary one answers from memory and
	// returns no citations, which the provider treats as an error rather than
	// letting an ungrounded answer pass for a grounded one.
	//
	// Named separately from LLMContent because searching is a different capability
	// with a different (and much shorter) list of models that offer it, and
	// because it is the one model reference somebody may want to turn OFF: empty
	// disables grounding and the substance stage falls back to what the brief
	// states, rather than failing the run.
	LLMSearch       string  `yaml:"llm_search"`
	ReviewThreshold float64 `yaml:"review_threshold"`
	CaptionsModel   string  `yaml:"captions_model"`
	// VideoOnly skips the companion-material stages (quiz, quiz-strategy,
	// mistakes, exercises, hugo): the course is just its videos. Merge is
	// true-wins — a course that opts in stays video-only for all its lessons.
	// omitempty keeps it out of the config fingerprint (StageInputs zeroes it
	// before hashing): stage hashes recorded before the field existed must not
	// go stale just because the key started being marshaled.
	VideoOnly bool `yaml:"video_only,omitempty"`
}

// Defaults returns the base configuration layer.
func Defaults() Config {
	return Config{
		Style: Style{
			Voice: "af_heart",
			// af_heart at its natural rate reads a touch faster than is
			// comfortable to learn from. 0.9 is the house rate; the align
			// stage scales the pace target by it, so slowing down here does
			// not read as being 10% under pace.
			VoiceSpeed: 0.9,
			Tone:       "friendly, conversational teacher",
			PaceWPM:    150,
			Audience:   "absolute beginners with no programming experience",
			Language:   "en",
		},
		Branding: Branding{
			Colors: Colors{
				Primary:    "#2563eb",
				Accent:     "#f59e0b",
				Background: "#ffffff",
			},
			DiagramStyle: "clean, flat, rounded corners, generous whitespace",
		},
		Pipeline: Pipeline{
			LLMContent: "openai/gpt-4o-mini",
			LLMReview:  "openai/gpt-4o-mini",
			LLMVision:  "openai/gpt-4o",
			// gpt-5-search-api rather than the gpt-4o-*-search-preview pair,
			// which are deprecated.
			LLMSearch:       "openai/gpt-5-search-api",
			ReviewThreshold: 8,
			CaptionsModel:   "whisper-large-v3",
		},
		Audio: Audio{
			SectionPauseMs:   700,
			ParagraphPauseMs: 350,
			SentencePauseMs:  400,
			CrossfadeMs:      50,
			TargetLUFS:       -16,
			MusicBed:         false,
			MusicDuckDB:      -18,
		},
	}
}

// Merge overlays over onto base, field by field. Zero-valued fields in over
// inherit from base. Neither argument is mutated.
func Merge(base, over Config) Config {
	out := base
	if over.Style.Voice != "" {
		out.Style.Voice = over.Style.Voice
	}
	if over.Style.VoiceSpeed != 0 {
		out.Style.VoiceSpeed = over.Style.VoiceSpeed
	}
	if over.Style.Tone != "" {
		out.Style.Tone = over.Style.Tone
	}
	if over.Style.PaceWPM != 0 {
		out.Style.PaceWPM = over.Style.PaceWPM
	}
	if over.Style.Audience != "" {
		out.Style.Audience = over.Style.Audience
	}
	if over.Style.Language != "" {
		out.Style.Language = over.Style.Language
	}
	if over.Style.Archetype != "" {
		out.Style.Archetype = over.Style.Archetype
	}
	if over.Style.AnimationStyle != "" {
		out.Style.AnimationStyle = over.Style.AnimationStyle
	}
	if over.Style.ColorPalette != "" {
		out.Style.ColorPalette = over.Style.ColorPalette
	}
	if over.Style.Captions != "" {
		out.Style.Captions = over.Style.Captions
	}
	if over.Style.Mode != "" {
		out.Style.Mode = over.Style.Mode
	}
	if over.Style.Skin != "" {
		out.Style.Skin = over.Style.Skin
	}
	if over.Style.Watermark != "" {
		out.Style.Watermark = over.Style.Watermark
	}
	if len(over.Style.Pronunciations) > 0 {
		merged := make(map[string]string, len(base.Style.Pronunciations)+len(over.Style.Pronunciations))
		maps.Copy(merged, base.Style.Pronunciations)
		maps.Copy(merged, over.Style.Pronunciations)
		out.Style.Pronunciations = merged
	}
	if over.Branding.Colors.Primary != "" {
		out.Branding.Colors.Primary = over.Branding.Colors.Primary
	}
	if over.Branding.Colors.Accent != "" {
		out.Branding.Colors.Accent = over.Branding.Colors.Accent
	}
	if over.Branding.Colors.Background != "" {
		out.Branding.Colors.Background = over.Branding.Colors.Background
	}
	if over.Branding.Fonts.Display != "" {
		out.Branding.Fonts.Display = over.Branding.Fonts.Display
	}
	if over.Branding.Fonts.Body != "" {
		out.Branding.Fonts.Body = over.Branding.Fonts.Body
	}
	if over.Branding.Fonts.Mono != "" {
		out.Branding.Fonts.Mono = over.Branding.Fonts.Mono
	}
	if over.Branding.DiagramStyle != "" {
		out.Branding.DiagramStyle = over.Branding.DiagramStyle
	}
	if over.Pipeline.LLMContent != "" {
		out.Pipeline.LLMContent = over.Pipeline.LLMContent
	}
	if over.Pipeline.LLMReview != "" {
		out.Pipeline.LLMReview = over.Pipeline.LLMReview
	}
	if over.Pipeline.LLMVision != "" {
		out.Pipeline.LLMVision = over.Pipeline.LLMVision
	}
	if over.Pipeline.LLMSearch != "" {
		out.Pipeline.LLMSearch = over.Pipeline.LLMSearch
	}
	if over.Pipeline.ReviewThreshold != 0 {
		out.Pipeline.ReviewThreshold = over.Pipeline.ReviewThreshold
	}
	if over.Pipeline.VideoOnly {
		out.Pipeline.VideoOnly = true
	}
	if over.Pipeline.CaptionsModel != "" {
		out.Pipeline.CaptionsModel = over.Pipeline.CaptionsModel
	}
	if over.Audio.SectionPauseMs != 0 {
		out.Audio.SectionPauseMs = over.Audio.SectionPauseMs
	}
	if over.Audio.ParagraphPauseMs != 0 {
		out.Audio.ParagraphPauseMs = over.Audio.ParagraphPauseMs
	}
	if over.Audio.SentencePauseMs != 0 {
		out.Audio.SentencePauseMs = over.Audio.SentencePauseMs
	}
	if over.Audio.CrossfadeMs != 0 {
		out.Audio.CrossfadeMs = over.Audio.CrossfadeMs
	}
	if over.Audio.TargetLUFS != 0 {
		out.Audio.TargetLUFS = over.Audio.TargetLUFS
	}
	if over.Audio.MusicBed {
		out.Audio.MusicBed = true
	}
	if over.Audio.MusicDuckDB != 0 {
		out.Audio.MusicDuckDB = over.Audio.MusicDuckDB
	}
	return out
}

// Resolve merges the standard layer stack into a final Config.
// Any layer may be the zero Config.
func Resolve(course, lesson, flags Config) Config {
	out := Merge(Defaults(), course)
	out = Merge(out, lesson)
	out = Merge(out, flags)
	return out
}
