package studio

// Token/cost ledger: every LLM response is cached on disk with its usage,
// so the ledger is derived by scanning the cache — no separate accounting
// that can drift from reality.

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/enfec/coursesmith/internal/llm"
)

// modelPricing is USD per 1M tokens (prompt, completion). Models not
// listed cost zero (Groq free tier, local models).
var modelPricing = map[string][2]float64{
	"gpt-4o-mini":  {0.15, 0.60},
	"gpt-4o":       {2.50, 10.00},
	"gpt-4.1":      {2.00, 8.00},
	"gpt-4.1-mini": {0.40, 1.60},
}

// priceFor matches a model id against the pricing table by prefix, so
// dated ids ("gpt-4o-mini-2024-07-18") price correctly.
func priceFor(model string) (perPrompt, perCompletion float64) {
	best := ""
	for prefix := range modelPricing {
		if len(prefix) > len(best) && len(model) >= len(prefix) && model[:len(prefix)] == prefix {
			best = prefix
		}
	}
	if best == "" {
		return 0, 0
	}
	p := modelPricing[best]
	return p[0] / 1e6, p[1] / 1e6
}

// LedgerRow aggregates usage for one (day, provider, model).
type LedgerRow struct {
	Day              string  `json:"day"` // YYYY-MM-DD (UTC)
	Provider         string  `json:"provider"`
	Model            string  `json:"model"`
	Calls            int     `json:"calls"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	CostUSD          float64 `json:"cost_usd"`
}

// QuotaStatus is one provider's live rate-limit state.
type QuotaStatus struct {
	Provider  string `json:"provider"`
	PerMinute int    `json:"per_minute"`
	PerDay    int    `json:"per_day"`
	DayUsed   int    `json:"day_used"`
}

// Ledger is the GET /api/ledger response.
type Ledger struct {
	Rows         []LedgerRow   `json:"rows"`
	TotalCostUSD float64       `json:"total_cost_usd"`
	TotalCalls   int           `json:"total_calls"`
	Quotas       []QuotaStatus `json:"quotas"`
}

// cacheFileEntry mirrors the llm cache's on-disk format (read-only here).
type cacheFileEntry struct {
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	Response  struct {
		Model string `json:"model"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	} `json:"response"`
}

// BuildLedger scans the LLM cache dir and the rate-limit state files.
func BuildLedger(stateDir string) (*Ledger, error) {
	ledger := &Ledger{}
	byKey := map[string]*LedgerRow{}

	entries, err := os.ReadDir(filepath.Join(stateDir, "cache"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(stateDir, "cache", entry.Name()))
		if err != nil {
			continue
		}
		var ce cacheFileEntry
		if json.Unmarshal(data, &ce) != nil {
			continue
		}
		day := ce.CreatedAt.UTC().Format("2006-01-02")
		key := day + "|" + ce.Provider + "|" + ce.Response.Model
		row, ok := byKey[key]
		if !ok {
			row = &LedgerRow{Day: day, Provider: ce.Provider, Model: ce.Response.Model}
			byKey[key] = row
		}
		row.Calls++
		row.PromptTokens += ce.Response.Usage.PromptTokens
		row.CompletionTokens += ce.Response.Usage.CompletionTokens
		pp, pc := priceFor(ce.Response.Model)
		row.CostUSD += float64(ce.Response.Usage.PromptTokens)*pp + float64(ce.Response.Usage.CompletionTokens)*pc
	}

	for _, row := range byKey {
		ledger.Rows = append(ledger.Rows, *row)
		ledger.TotalCostUSD += row.CostUSD
		ledger.TotalCalls += row.Calls
	}
	sort.Slice(ledger.Rows, func(i, j int) bool {
		if ledger.Rows[i].Day != ledger.Rows[j].Day {
			return ledger.Rows[i].Day < ledger.Rows[j].Day
		}
		return ledger.Rows[i].Model < ledger.Rows[j].Model
	})

	// Live quota state from the rate limiter's persisted buckets.
	for provider, cfg := range llm.DefaultLimits {
		q := QuotaStatus{Provider: provider, PerMinute: cfg.PerMinute, PerDay: cfg.PerDay}
		data, err := os.ReadFile(filepath.Join(stateDir, "ratelimit", provider+".json"))
		if err == nil {
			var state struct {
				DayUsed  int       `json:"day_used"`
				DayStart time.Time `json:"day_start"`
			}
			if json.Unmarshal(data, &state) == nil && time.Since(state.DayStart) < 24*time.Hour {
				q.DayUsed = state.DayUsed
			}
		}
		ledger.Quotas = append(ledger.Quotas, q)
	}
	sort.Slice(ledger.Quotas, func(i, j int) bool { return ledger.Quotas[i].Provider < ledger.Quotas[j].Provider })
	return ledger, nil
}

func (s *Server) handleLedger(w http.ResponseWriter, r *http.Request) {
	ledger, err := BuildLedger(s.stateDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, ledger)
}
