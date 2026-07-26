package llm

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/enfec/coursesmith/internal/config"
)

// TaskType names a pipeline job that needs an LLM; config maps each to a
// "provider/model" reference.
type TaskType string

const (
	// TaskContent generates lesson artifacts (scripts, quizzes, diagrams).
	TaskContent TaskType = "content"
	// TaskReview critiques generated artifacts against the rubric.
	TaskReview TaskType = "review"
	// TaskVision inspects rendered images (diagram screenshots). It needs a
	// vision-capable model with strong spatial perception; a weak judge
	// hallucinates overlaps on clean, layout-engine-produced diagrams.
	TaskVision TaskType = "vision"
)

// DefaultStateDir is where the router keeps its cache and rate-limit state,
// relative to the project root.
const DefaultStateDir = ".coursesmith"

// ModelFor returns the configured "provider/model" reference for a task.
func ModelFor(p config.Pipeline, task TaskType) (string, error) {
	switch task {
	case TaskContent:
		if p.LLMContent == "" {
			return "", fmt.Errorf("pipeline.llm_content is not configured")
		}
		return p.LLMContent, nil
	case TaskReview:
		if p.LLMReview == "" {
			return "", fmt.Errorf("pipeline.llm_review is not configured")
		}
		return p.LLMReview, nil
	case TaskVision:
		// Fall back to the review model when no dedicated vision model is set,
		// so older configs keep working.
		if p.LLMVision != "" {
			return p.LLMVision, nil
		}
		if p.LLMReview == "" {
			return "", fmt.Errorf("pipeline.llm_vision (and fallback llm_review) is not configured")
		}
		return p.LLMReview, nil
	default:
		return "", fmt.Errorf("unknown LLM task type %q", task)
	}
}

// ParseModelRef splits "groq/llama-3.3-70b-versatile" into provider and
// model. Models may themselves contain slashes; only the first one splits.
func ParseModelRef(ref string) (provider, model string, err error) {
	provider, model, ok := strings.Cut(ref, "/")
	if !ok || provider == "" || model == "" {
		return "", "", fmt.Errorf("model reference %q must be provider/model, e.g. %q", ref, "groq/llama-3.3-70b-versatile")
	}
	return provider, model, nil
}

// Router resolves task types to providers and completes requests through the
// full stack: cache → retry → rate limit → HTTP. Providers are built lazily,
// so an API key is only required for providers that are actually routed to.
type Router struct {
	mu        sync.Mutex
	stateDir  string
	cache     *Cache
	limits    map[string]LimitConfig
	providers map[string]Provider

	// newProvider builds the base (HTTP) provider for a name; injectable
	// for tests. Defaults to env-key construction.
	newProvider func(name string) (Provider, error)
}

// NewRouter creates a router keeping cache and rate-limit state under
// stateDir ("" uses DefaultStateDir).
func NewRouter(stateDir string) *Router {
	if stateDir == "" {
		stateDir = DefaultStateDir
	}
	return &Router{
		stateDir:    stateDir,
		cache:       NewCache(filepath.Join(stateDir, "cache")),
		limits:      DefaultLimits,
		providers:   map[string]Provider{},
		newProvider: baseProviderFromEnv,
	}
}

func baseProviderFromEnv(name string) (Provider, error) {
	switch name {
	case "groq":
		return GroqFromEnv()
	case "openai":
		return OpenAIFromEnv()
	default:
		return nil, fmt.Errorf("unknown LLM provider %q (supported: groq, openai)", name)
	}
}

// Complete resolves the task's configured model and completes the request.
// req.Model is filled in from config; any value already set is overridden.
func (r *Router) Complete(ctx context.Context, pcfg config.Pipeline, task TaskType, req Request) (*Response, error) {
	ref, err := ModelFor(pcfg, task)
	if err != nil {
		return nil, err
	}
	providerName, model, err := ParseModelRef(ref)
	if err != nil {
		return nil, fmt.Errorf("resolving %s model: %w", task, err)
	}
	p, err := r.provider(providerName)
	if err != nil {
		return nil, err
	}
	req.Model = model
	return p.Complete(ctx, req)
}

// provider returns the memoized full stack for a provider name, building it
// on first use.
func (r *Router) provider(name string) (Provider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.providers[name]; ok {
		return p, nil
	}
	base, err := r.newProvider(name)
	if err != nil {
		return nil, err
	}
	limits, ok := r.limits[name]
	if !ok {
		// Unknown provider: throttle conservatively rather than not at all.
		limits = LimitConfig{PerMinute: 20, PerDay: 500}
	}
	limiter := NewLimiter(name, limits, filepath.Join(r.stateDir, "ratelimit", name+".json"))
	stack := withCache(withRetry(base, limiter), r.cache)
	r.providers[name] = stack
	return stack, nil
}
