package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Cache is a disk cache of LLM responses. Entries are keyed by a hash of
// (provider, full request), so re-running a pipeline stage with unchanged
// inputs costs zero API calls.
type Cache struct {
	dir string
}

// NewCache returns a cache rooted at dir (e.g. ".coursesmith/cache").
// The directory is created on first Put.
func NewCache(dir string) *Cache {
	return &Cache{dir: dir}
}

// cacheEntry is the on-disk format. The request is stored alongside the
// response so cache files are auditable and debuggable.
type cacheEntry struct {
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	Request   Request   `json:"request"`
	Response  Response  `json:"response"`
}

// Key returns the cache key for a request sent to a named provider:
// hex(sha256(provider + canonical request JSON)).
func (c *Cache) Key(provider string, req Request) string {
	// Request marshals with a fixed field order, so this is deterministic.
	data, err := json.Marshal(struct {
		Provider string  `json:"provider"`
		Request  Request `json:"request"`
	}{provider, req})
	if err != nil {
		// Request contains only plain values; Marshal cannot fail on it.
		panic(fmt.Sprintf("marshaling cache key: %v", err))
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (c *Cache) path(key string) string {
	return filepath.Join(c.dir, key+".json")
}

// Get returns the cached response for key, or (nil, false) on any miss —
// including unreadable or corrupt entries, which are treated as absent.
func (c *Cache) Get(key string) (*Response, bool) {
	data, err := os.ReadFile(c.path(key))
	if err != nil {
		return nil, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}
	resp := entry.Response
	return &resp, true
}

// Put stores a response under key, atomically (temp file + rename).
func (c *Cache) Put(key, provider string, req Request, resp *Response) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return fmt.Errorf("creating cache dir %s: %w", c.dir, err)
	}
	stored := *resp
	stored.FromCache = false
	entry := cacheEntry{
		Provider:  provider,
		CreatedAt: time.Now().UTC(),
		Request:   req,
		Response:  stored,
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding cache entry: %w", err)
	}
	tmp, err := os.CreateTemp(c.dir, key+".tmp-*")
	if err != nil {
		return fmt.Errorf("creating cache temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing cache entry: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing cache temp file: %w", err)
	}
	if err := os.Rename(tmpName, c.path(key)); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("storing cache entry: %w", err)
	}
	return nil
}

// cachedProvider serves repeated requests from the cache without touching
// the wrapped provider (and therefore without consuming rate-limit tokens).
type cachedProvider struct {
	inner Provider
	cache *Cache
}

func withCache(inner Provider, cache *Cache) Provider {
	return &cachedProvider{inner: inner, cache: cache}
}

func (p *cachedProvider) Name() string { return p.inner.Name() }

func (p *cachedProvider) Complete(ctx context.Context, req Request) (*Response, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	key := p.cache.Key(p.inner.Name(), req)
	if resp, ok := p.cache.Get(key); ok {
		resp.FromCache = true
		return resp, nil
	}
	resp, err := p.inner.Complete(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := p.cache.Put(key, p.inner.Name(), req, resp); err != nil {
		// The response is valid; losing the cache write costs a future API
		// call, not correctness. Surface it without failing the request.
		fmt.Fprintf(os.Stderr, "warning: could not cache %s response: %v\n", p.inner.Name(), err)
	}
	return resp, nil
}
