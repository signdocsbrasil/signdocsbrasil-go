package signdocsbrasil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAuthHandler_UsesInjectedCache(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: "tok_injected", ExpiresIn: 3600})
	}))
	defer server.Close()

	cache := NewInMemoryTokenCache()

	cfg := &Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		BaseURL:      server.URL,
		Scopes:       []string{"transactions:read"},
		HTTPClient:   server.Client(),
		TokenCache:   cache,
	}

	auth := newAuthHandler(cfg)

	tok1, err := auth.getAccessToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok1 != "tok_injected" {
		t.Errorf("expected 'tok_injected', got %q", tok1)
	}

	// Second call should be served from the cache.
	tok2, err := auth.getAccessToken()
	if err != nil {
		t.Fatalf("unexpected error on cached call: %v", err)
	}
	if tok2 != "tok_injected" {
		t.Errorf("expected cached token, got %q", tok2)
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected exactly 1 token fetch, got %d", got)
	}

	// The injected cache should hold the derived key.
	expectedKey := DeriveCacheKey("test-client", server.URL, []string{"transactions:read"})
	cached, ok := cache.Get(expectedKey)
	if !ok {
		t.Fatalf("expected injected cache to hold key %q", expectedKey)
	}
	if cached.AccessToken != "tok_injected" {
		t.Errorf("expected cached token 'tok_injected', got %q", cached.AccessToken)
	}
}

func TestAuthHandler_InvalidateForcesRefresh(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		tok := "tok_1"
		if n > 1 {
			tok = "tok_2"
		}
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: tok, ExpiresIn: 3600})
	}))
	defer server.Close()

	cfg := &Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		BaseURL:      server.URL,
		Scopes:       []string{"transactions:read"},
		HTTPClient:   server.Client(),
	}
	auth := newAuthHandler(cfg)

	tok1, _ := auth.getAccessToken()
	if tok1 != "tok_1" {
		t.Errorf("expected tok_1, got %q", tok1)
	}

	auth.invalidate()

	tok2, _ := auth.getAccessToken()
	if tok2 != "tok_2" {
		t.Errorf("expected tok_2 after invalidate, got %q", tok2)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 fetches after invalidate, got %d", got)
	}
}

func TestAuthHandler_SharedCacheAcrossHandlers(t *testing.T) {
	// Two authHandler instances pointed at the same cache and
	// credentials should share a single token — the second handler
	// must not trigger a fetch.
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: "shared_tok", ExpiresIn: 3600})
	}))
	defer server.Close()

	cache := NewInMemoryTokenCache()
	baseCfg := func() *Config {
		return &Config{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			BaseURL:      server.URL,
			Scopes:       []string{"transactions:read"},
			HTTPClient:   server.Client(),
			TokenCache:   cache,
		}
	}

	auth1 := newAuthHandler(baseCfg())
	auth2 := newAuthHandler(baseCfg())

	if _, err := auth1.getAccessToken(); err != nil {
		t.Fatal(err)
	}
	tok, err := auth2.getAccessToken()
	if err != nil {
		t.Fatal(err)
	}
	if tok != "shared_tok" {
		t.Errorf("expected shared token, got %q", tok)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 fetch across both handlers, got %d", got)
	}
}

func TestAuthHandler_ConcurrentFetchesCoalesceToOne(t *testing.T) {
	// A burst of concurrent getAccessToken calls on a cold cache must
	// result in at most one token fetch — the fetch mutex serializes
	// fetches and the recheck-under-lock catches winners.
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		// Small delay to widen the race window.
		time.Sleep(20 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: "tok_race", ExpiresIn: 3600})
	}))
	defer server.Close()

	cfg := &Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		BaseURL:      server.URL,
		Scopes:       []string{"transactions:read"},
		HTTPClient:   server.Client(),
	}
	auth := newAuthHandler(cfg)

	const goroutines = 30
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if _, err := auth.getAccessToken(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 coalesced fetch, got %d", got)
	}
}
