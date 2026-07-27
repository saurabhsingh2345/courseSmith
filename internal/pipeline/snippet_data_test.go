package pipeline

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func dataPlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "data",
		Title:    "Where the time goes",
		Chart: &ChartSpec{
			Kind: "bars",
			Unit: "ms",
			Points: []DataPoint{
				{Label: "Database", Value: 412},
				{Label: "Template render", Value: 96},
				{Label: "Auth check", Value: 41},
				{Label: "Routing", Value: 9},
			},
		},
		Beats: []SnippetBeat{
			{ID: "the-request", Heading: "One request", Narration: strings.Repeat("request ", 22),
				Data: &DataBeat{Caption: "A single request, by where the milliseconds land."}},
			{ID: "the-culprit", Heading: "Where it goes", Narration: strings.Repeat("culprit ", 22),
				Data: &DataBeat{Highlight: []string{"Database"}, Caption: "Four fifths of the wait."}},
			{ID: "the-rest", Heading: "What we tune", Narration: strings.Repeat("rest ", 22),
				Data: &DataBeat{Highlight: []string{"Routing", "Auth check"}, Caption: "Everything people tune is here."}},
		},
	}
}

func TestValidateDataPlan(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := dataPlan().Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	t.Run("no chart", func(t *testing.T) {
		p := dataPlan()
		p.Chart = nil
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "no chart") {
			t.Fatalf("want missing-chart error, got %v", err)
		}
	})
	t.Run("too few points", func(t *testing.T) {
		p := dataPlan()
		p.Chart.Points = p.Chart.Points[:2]
		p.Beats[1].Data.Highlight = []string{"Database"}
		p.Beats[2].Data.Highlight = []string{"Template render"}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "chart has 2 points") {
			t.Fatalf("want point-count error, got %v", err)
		}
	})
	t.Run("negative value", func(t *testing.T) {
		p := dataPlan()
		p.Chart.Points[1].Value = -3
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "negative value") {
			t.Fatalf("want negative-value error, got %v", err)
		}
	})
	t.Run("duplicate label", func(t *testing.T) {
		p := dataPlan()
		p.Chart.Points[2].Label = "Database"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "appears twice") {
			t.Fatalf("want duplicate-label error, got %v", err)
		}
	})
	// A highlight that names nothing is a beat whose emphasis silently does not
	// happen — the failure this whole family of guards exists to prevent.
	t.Run("highlight names a point that does not exist", func(t *testing.T) {
		p := dataPlan()
		p.Beats[1].Data.Highlight = []string{"Databse"}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "not one of the chart's points") {
			t.Fatalf("want unknown-highlight error, got %v", err)
		}
	})
	t.Run("consecutive beats highlight the same set", func(t *testing.T) {
		p := dataPlan()
		p.Beats[2].Data.Highlight = []string{"Database"}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "previous beat did") {
			t.Fatalf("want repeated-highlight error, got %v", err)
		}
	})
	// Order must not matter: the same two points named the other way round is
	// still the same picture.
	t.Run("same set in a different order still repeats", func(t *testing.T) {
		p := dataPlan()
		p.Beats[1].Data.Highlight = []string{"Routing", "Auth check"}
		p.Beats[2].Data.Highlight = []string{"Auth check", "Routing"}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "previous beat did") {
			t.Fatalf("want repeated-highlight error, got %v", err)
		}
	})
	t.Run("nobody points at the chart", func(t *testing.T) {
		p := dataPlan()
		p.Beats[1].Data.Highlight = nil
		p.Beats[2].Data.Highlight = nil
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "highlight anything") {
			t.Fatalf("want highlight-floor error, got %v", err)
		}
	})
	t.Run("foreign beat field", func(t *testing.T) {
		p := dataPlan()
		p.Beats[0].Sketch = []SketchItem{{Label: "Browser", Icon: "monitor"}}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "does not use") {
			t.Fatalf("want foreign-field error, got %v", err)
		}
	})
}

func TestValidateDataMapCountries(t *testing.T) {
	mapPlan := func() *SnippetPlan {
		p := dataPlan()
		p.Chart.Kind = "map"
		p.Chart.Unit = ""
		p.Chart.Points = []DataPoint{
			{Label: "USA", Value: 88},
			{Label: "uk", Value: 61},
			{Label: "India", Value: 35},
			{Label: "Brazil", Value: 26},
		}
		p.Beats[1].Data.Highlight = []string{"USA"}
		p.Beats[2].Data.Highlight = []string{"India", "Brazil"}
		return p
	}
	// Nobody writes "United States of America". The aliases are the difference
	// between the commonest countries in any explainer rendering and not.
	t.Run("aliases and casing resolve", func(t *testing.T) {
		if err := mapPlan().Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	t.Run("unknown country", func(t *testing.T) {
		p := mapPlan()
		p.Chart.Points[0].Label = "Atlantis"
		p.Beats[1].Data.Highlight = []string{"Atlantis"}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "no country called") {
			t.Fatalf("want unknown-country error, got %v", err)
		}
	})
	// "USA" and "United States of America" are the same country, so a chart
	// naming both is a duplicate however differently it is spelled.
	t.Run("two spellings of one country collide", func(t *testing.T) {
		p := mapPlan()
		p.Chart.Points[1].Label = "United States of America"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "appears twice") {
			t.Fatalf("want duplicate-country error, got %v", err)
		}
	})
	// The scene must hand the renderer canonical names, since that is what the
	// atlas is keyed by — an alias reaching the map is a country that vanishes.
	t.Run("scenes canonicalize country names", func(t *testing.T) {
		scenes, err := dataScenes(sceneInput(t, mapPlan(), 6000))
		if err != nil {
			t.Fatal(err)
		}
		points, _ := scenes[0].Props["points"].([]map[string]any)
		if len(points) == 0 || points[0]["label"] != "United States of America" {
			t.Errorf("point label = %v, want the atlas's canonical name", points[0]["label"])
		}
		windows, _ := scenes[0].Props["highlight"].([]map[string]any)
		labels, _ := windows[0]["labels"].([]string)
		if len(labels) == 0 || labels[0] != "United States of America" {
			t.Errorf("highlight label = %v, want canonical", labels)
		}
	})
}

func TestDataScenes(t *testing.T) {
	plan := dataPlan()
	scenes, err := dataScenes(sceneInput(t, plan, 6000))
	if err != nil {
		t.Fatal(err)
	}
	// One chart for the whole clip, not one per beat. That is the format.
	if len(scenes) != 1 {
		t.Fatalf("got %d scenes, want exactly 1 — the chart is held for the clip", len(scenes))
	}
	s := scenes[0]
	if s.Type != SceneData {
		t.Errorf("scene type = %q, want %q", s.Type, SceneData)
	}
	if s.StartMs != 0 || s.EndMs != 3*6000 {
		t.Errorf("scene spans %d-%d, want the whole clip", s.StartMs, s.EndMs)
	}
	windows, _ := s.Props["highlight"].([]map[string]any)
	if len(windows) != 2 {
		t.Fatalf("got %d highlight windows, want one per highlighting beat (2)", len(windows))
	}
	if windows[0]["startMs"] != 6000 {
		t.Errorf("first window starts at %v, want its beat's start (6000)", windows[0]["startMs"])
	}
	captions, _ := s.Props["captions"].([]map[string]any)
	if len(captions) != 3 {
		t.Errorf("got %d captions, want one per beat that has one (3)", len(captions))
	}
}

func TestNormalizeChartKind(t *testing.T) {
	if got := normalizeChartKind("  MAP "); got != "map" {
		t.Errorf("normalizeChartKind = %q, want map", got)
	}
	if got := normalizeChartKind("sankey"); got != "bars" {
		t.Errorf("normalizeChartKind fallback = %q, want bars", got)
	}
}

// The chart kinds Go accepts and the ones DataScene draws must be the same set:
// a kind Go allows and the scene does not draw silently falls through to bars,
// so a clip planned as a treemap renders as a bar chart and the plan's own
// captions stop describing what is on screen.
//
// The kinds used to be a chain of `kind === '...'` comparisons and are a map
// now — thirteen arms of a ternary is not something anybody can read, and the
// map is also the thing this test wants to look at anyway.
var (
	tsChartsBlockRe = regexp.MustCompile(`(?s)const CHARTS: Record[^{]*\{(.*?)\n\};`)
	tsChartEntryRe  = regexp.MustCompile(`(?m)^\s{2}([a-z]+):`)
)

func TestChartKindsInSync(t *testing.T) {
	src, err := os.ReadFile("../../renderer/src/components/DataScene.tsx")
	if err != nil {
		t.Fatalf("reading DataScene: %v", err)
	}
	block := tsChartsBlockRe.FindSubmatch(src)
	if block == nil {
		t.Fatalf("no CHARTS map found in DataScene.tsx — has its shape changed?")
	}
	drawn := map[string]bool{}
	for _, m := range tsChartEntryRe.FindAllSubmatch(block[1], -1) {
		drawn[string(m[1])] = true
	}
	if len(drawn) == 0 {
		t.Fatal("no chart kinds parsed from DataScene.tsx")
	}

	var missing []string
	for kind := range chartKindVocab {
		if !drawn[kind] {
			missing = append(missing, kind)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("chartKindVocab allows %v, which DataScene does not draw — those clips would silently render as bars", missing)
	}
	for kind := range drawn {
		if _, ok := chartKindVocab[kind]; !ok {
			t.Errorf("DataScene draws %q, which chartKindVocab rejects — nobody can ask for it", kind)
		}
	}
}

// Go's country list has to match the atlas the renderer actually draws, or a
// country Go accepts is a country that does not appear.
func TestCountryVocabularyMatchesAtlas(t *testing.T) {
	raw, err := os.ReadFile("../../renderer/node_modules/world-atlas/countries-110m.json")
	if err != nil {
		t.Skipf("world-atlas not installed (%v); run npm install in renderer/", err)
	}
	var topo struct {
		Objects struct {
			Countries struct {
				Geometries []struct {
					Properties struct {
						Name string `json:"name"`
					} `json:"properties"`
				} `json:"geometries"`
			} `json:"countries"`
		} `json:"objects"`
	}
	if err := json.Unmarshal(raw, &topo); err != nil {
		t.Fatalf("parsing atlas: %v", err)
	}
	// geo.ts drops Antarctica, so Go must not accept it: a country the renderer
	// never draws is a plan that validates and then shows nothing. Any *other*
	// mismatch is real drift.
	omitted := map[string]bool{"Antarctica": true}

	atlas := map[string]bool{}
	for _, g := range topo.Objects.Countries.Geometries {
		if g.Properties.Name != "" && !omitted[g.Properties.Name] {
			atlas[g.Properties.Name] = true
		}
	}
	for name := range omitted {
		if mapCountryVocab[name] {
			t.Errorf("mapCountryVocab accepts %q, which geo.ts does not draw", name)
		}
	}
	if len(atlas) == 0 {
		t.Fatal("parsed no country names from the atlas")
	}
	for name := range mapCountryVocab {
		if !atlas[name] {
			t.Errorf("mapCountryVocab has %q, which the atlas does not — regenerate countries.go", name)
		}
	}
	for name := range atlas {
		if !mapCountryVocab[name] {
			t.Errorf("the atlas has %q, which mapCountryVocab rejects — regenerate countries.go", name)
		}
	}
	// Every alias must land on a name the atlas really has, or the alias is a
	// redirect to nowhere.
	for alias, canonical := range mapCountryAliases {
		if !atlas[canonical] {
			t.Errorf("alias %q points at %q, which is not in the atlas", alias, canonical)
		}
	}
}

// The prompt's example must satisfy the rules the prompt states.
func TestDataPromptExampleIsValid(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("../../prompts", snippetDataTemplateName))
	if err != nil {
		t.Fatalf("reading prompt: %v", err)
	}
	at := bytes.Index(src, []byte(`{"title":`))
	if at < 0 {
		t.Fatalf("no example reply found in %s", snippetDataTemplateName)
	}
	line := src[at:]
	if end := bytes.IndexByte(line, '\n'); end >= 0 {
		line = line[:end]
	}
	var plan SnippetPlan
	if err := json.Unmarshal(line, &plan); err != nil {
		t.Fatalf("the example in %s is not valid JSON: %v", snippetDataTemplateName, err)
	}
	plan.Template = "data"
	for i := range plan.Beats {
		plan.Beats[i].Narration = strings.Repeat("narration ", 22)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("the example in %s does not satisfy the rules that same prompt states: %v",
			snippetDataTemplateName, err)
	}
}

// A chart with more than one number per label is where this template can go
// wrong quietly: a short values row still renders a perfectly good bar, it just
// states a total that is not the total. So the shape is checked hard rather
// than padded.
func seriesPlan(kind string, series []string, values [][]float64) *SnippetPlan {
	p := dataPlan()
	p.Chart = &ChartSpec{Kind: kind, Unit: "ms", Series: series}
	labels := []string{"Search", "Checkout", "Home"}
	p.Chart.Points = nil
	for i, v := range values {
		p.Chart.Points = append(p.Chart.Points, DataPoint{Label: labels[i], Values: v})
	}
	for i := range p.Beats {
		if p.Beats[i].Data != nil && len(p.Beats[i].Data.Highlight) > 0 {
			p.Beats[i].Data.Highlight = []string{labels[i%len(labels)]}
		}
	}
	return p
}

func TestValidateChartSeries(t *testing.T) {
	ok := seriesPlan("stackedbars", []string{"Database", "Render"},
		[][]float64{{310, 84}, {120, 210}, {40, 66}})
	if err := ok.Validate(); err != nil {
		t.Fatalf("want valid stacked bars, got %v", err)
	}

	t.Run("point missing a value", func(t *testing.T) {
		p := seriesPlan("stackedbars", []string{"Database", "Render"},
			[][]float64{{310, 84}, {120}, {40, 66}})
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "declares 2 series") {
			t.Fatalf("want series-length error, got %v", err)
		}
	})
	t.Run("series on a kind that takes one number", func(t *testing.T) {
		p := seriesPlan("stackedbars", []string{"Database", "Render"},
			[][]float64{{310, 84}, {120, 210}, {40, 66}})
		p.Chart.Kind = "donut"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "takes one number per point") {
			t.Fatalf("want no-series error, got %v", err)
		}
	})
	t.Run("scatter needs exactly two", func(t *testing.T) {
		p := seriesPlan("scatter", []string{"Team size", "Deploys", "Extra"},
			[][]float64{{6, 22, 1}, {14, 9, 2}, {4, 31, 3}})
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "exactly 2 series") {
			t.Fatalf("want scatter-series error, got %v", err)
		}
	})
	t.Run("duplicate series name", func(t *testing.T) {
		p := seriesPlan("groupedbars", []string{"Before", "before"},
			[][]float64{{310, 84}, {120, 210}, {40, 66}})
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "appears twice") {
			t.Fatalf("want duplicate-series error, got %v", err)
		}
	})
	// The scene has to carry the parts through, or the renderer draws one
	// segment where the plan asked for three.
	t.Run("scenes carry the parts", func(t *testing.T) {
		scenes, err := dataScenes(sceneInput(t, ok, 6000))
		if err != nil {
			t.Fatal(err)
		}
		if got := scenes[0].Props["series"]; got == nil {
			t.Fatal("the scene dropped the series names")
		}
		pts, _ := scenes[0].Props["points"].([]map[string]any)
		if len(pts) == 0 || pts[0]["values"] == nil {
			t.Fatalf("the scene dropped the per-series values: %#v", pts)
		}
		// The total is what a single-value reading of the point should see.
		if got := pts[0]["value"]; got != 394.0 {
			t.Errorf("point value = %v, want the sum of its parts (394)", got)
		}
	})
}

// A funnel drawn from values that rise is a picture of a funnel with numbers
// written on it — worse than no chart, because it looks like it means
// something.
func TestValidateFunnelNarrows(t *testing.T) {
	p := dataPlan()
	p.Chart = &ChartSpec{Kind: "funnel", Unit: "", Points: []DataPoint{
		{Label: "Visited", Value: 10000},
		{Label: "Signed up", Value: 3200},
		{Label: "Paid", Value: 310},
	}}
	for i := range p.Beats {
		if p.Beats[i].Data != nil && len(p.Beats[i].Data.Highlight) > 0 {
			p.Beats[i].Data.Highlight = []string{[]string{"Visited", "Signed up", "Paid"}[i%3]}
		}
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("want valid funnel, got %v", err)
	}
	p.Chart.Points[2].Value = 99999
	if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "a funnel narrows") {
		t.Fatalf("want widening-funnel error, got %v", err)
	}
}

// Several kinds run out of room before the reader does, and a kind-specific
// ceiling is the only place that can be said.
func TestChartKindPointCeilings(t *testing.T) {
	p := dataPlan()
	p.Chart = &ChartSpec{Kind: "waffle", Unit: "%", Points: []DataPoint{
		{Label: "One", Value: 20}, {Label: "Two", Value: 20}, {Label: "Three", Value: 20},
		{Label: "Four", Value: 20}, {Label: "Five", Value: 10}, {Label: "Six", Value: 10},
	}}
	for i := range p.Beats {
		if p.Beats[i].Data != nil && len(p.Beats[i].Data.Highlight) > 0 {
			p.Beats[i].Data.Highlight = []string{[]string{"One", "Two", "Three"}[i%3]}
		}
	}
	if err := p.Validate(); err == nil || !strings.Contains(err.Error(), `for kind "waffle"`) {
		t.Fatalf("want per-kind point-count error, got %v", err)
	}
}

// The near-misses a model actually writes. Without these a plan that says
// "stacked" degrades to bars and silently drops every series it declared.
func TestNormalizeChartKindAliases(t *testing.T) {
	for in, want := range map[string]string{
		"stacked":      "stackedbars",
		"Grouped-Bars": "groupedbars",
		"KPI Cards":    "bars", // spaces are not one of the near-misses
		"kpi-cards":    "kpi",
		"treemap":      "treemap",
	} {
		if got := normalizeChartKind(in); got != want {
			t.Errorf("normalizeChartKind(%q) = %q, want %q", in, got, want)
		}
	}
}
