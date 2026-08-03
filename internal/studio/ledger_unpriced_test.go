package studio

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

// writeCacheEntry writes one llm-cache-shaped file, the way the ledger reads them.
func writeCacheEntry(t *testing.T, dir, name, provider, model string, prompt, completion int) {
	t.Helper()
	entry := map[string]any{
		"provider":   provider,
		"created_at": time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC),
		"response": map[string]any{
			"model": model,
			"usage": map[string]any{"prompt_tokens": prompt, "completion_tokens": completion},
		},
	}
	b, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
}

// A model absent from the pricing table must be reported as unpriced, not as
// free. Conflating the two hid real money: adding gpt-5-search-api for the
// substance stage's grounding put a model in the pipeline the table had never
// heard of, so every grounded run recorded its search as costing exactly $0.00 —
// a wrong number that looked authoritative.
func TestLedgerFlagsUnpricedModels(t *testing.T) {
	state := t.TempDir()
	cache := filepath.Join(state, "cache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCacheEntry(t, cache, "a", "openai", "gpt-4o-mini-2024-07-18", 100000, 10000)
	writeCacheEntry(t, cache, "b", "openai", "gpt-5-search-api-2025-10-14", 32000, 2000)

	l, err := BuildLedger(state)
	if err != nil {
		t.Fatal(err)
	}

	var priced, unpriced *LedgerRow
	for i := range l.Rows {
		if l.Rows[i].Priced {
			priced = &l.Rows[i]
		} else {
			unpriced = &l.Rows[i]
		}
	}
	if priced == nil {
		t.Fatal("the known model was not marked priced")
	}
	if unpriced == nil {
		t.Fatal("the unknown model was marked priced — its spend is silently missing")
	}
	if unpriced.Model != "gpt-5-search-api-2025-10-14" {
		t.Errorf("wrong row flagged: %s", unpriced.Model)
	}

	// The known model still prices exactly as before: 100k prompt at $0.15/1M
	// plus 10k completion at $0.60/1M.
	want := 100000*0.15/1e6 + 10000*0.60/1e6
	if diff := priced.CostUSD - want; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("priced row = %v, want %v", priced.CostUSD, want)
	}
	// And the total is only the priced part, declared as such.
	if diff := l.TotalCostUSD - want; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("total = %v, want only the priced spend %v", l.TotalCostUSD, want)
	}
	if !slices.Contains(l.UnpricedModels, "gpt-5-search-api-2025-10-14") {
		t.Errorf("UnpricedModels = %v, does not name the unpriced model", l.UnpricedModels)
	}
	// The size of the gap, not only its existence: a caller deciding whether it
	// matters needs to know it is 34k tokens rather than a handful.
	if l.UnpricedTokens != 34000 {
		t.Errorf("UnpricedTokens = %d, want 34000", l.UnpricedTokens)
	}
}

// A free provider's zero is a fact, not a gap. Flagging Groq would fire the
// warning on every ledger that has ever used it, and a warning that cries wolf is
// one nobody reads by the third time — which would leave the real gap (a paid
// model nobody priced) exactly as hidden as it was before.
func TestFreeProvidersAreNotFlagged(t *testing.T) {
	state := t.TempDir()
	cache := filepath.Join(state, "cache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	// A Groq model nobody has ever priced, on the free tier.
	writeCacheEntry(t, cache, "a", "groq", "llama-3.3-70b-versatile", 50000, 17000)

	l, err := BuildLedger(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.UnpricedModels) != 0 {
		t.Errorf("a free provider was flagged as unpriced: %v", l.UnpricedModels)
	}
	if l.UnpricedTokens != 0 {
		t.Errorf("UnpricedTokens = %d for a free provider", l.UnpricedTokens)
	}
	if len(l.Rows) != 1 || !l.Rows[0].Priced {
		t.Error("a free provider's row is not marked priced, so the UI would show it as unknown")
	}
	if l.TotalCostUSD != 0 {
		t.Errorf("a free provider cost %v", l.TotalCostUSD)
	}
}

// A ledger with nothing unpriced must say nothing, or the warning stops carrying
// information.
func TestLedgerSilentWhenEverythingIsPriced(t *testing.T) {
	state := t.TempDir()
	cache := filepath.Join(state, "cache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCacheEntry(t, cache, "a", "openai", "gpt-4o-mini-2024-07-18", 1000, 100)

	l, err := BuildLedger(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.UnpricedModels) != 0 || l.UnpricedTokens != 0 {
		t.Errorf("a fully priced ledger reported a gap: %v / %d tokens", l.UnpricedModels, l.UnpricedTokens)
	}
	for _, r := range l.Rows {
		if !r.Priced {
			t.Errorf("row %s is not marked priced", r.Model)
		}
	}
}

// One unknown call in a bucket makes the bucket's total a floor. Marking the row
// priced because most of its calls were would be the same lie at finer grain.
func TestOneUnknownCallUnpricesItsRow(t *testing.T) {
	state := t.TempDir()
	cache := filepath.Join(state, "cache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	// Same day, same provider, same model string → one row, two calls.
	writeCacheEntry(t, cache, "a", "openai", "mystery-model", 10, 1)
	writeCacheEntry(t, cache, "b", "openai", "mystery-model", 10, 1)

	l, err := BuildLedger(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(l.Rows))
	}
	if l.Rows[0].Priced {
		t.Error("a row of entirely unknown calls was marked priced")
	}
	if l.Rows[0].Calls != 2 {
		t.Errorf("calls = %d, want 2", l.Rows[0].Calls)
	}
}

func TestPriceForReportsWhetherItKnows(t *testing.T) {
	if _, _, known := priceFor("gpt-4o-mini-2024-07-18"); !known {
		t.Error("a dated known model was not recognised")
	}
	if _, _, known := priceFor("gpt-5-search-api-2025-10-14"); known {
		t.Error("an unknown model reported a price")
	}
	// Longest-prefix wins, so gpt-4o-mini must not be priced as gpt-4o.
	pp, _, _ := priceFor("gpt-4o-mini")
	pp4o, _, _ := priceFor("gpt-4o")
	if pp >= pp4o {
		t.Errorf("gpt-4o-mini priced at %v, not below gpt-4o at %v — prefix matching is wrong", pp, pp4o)
	}
}
