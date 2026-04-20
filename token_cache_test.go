package signdocsbrasil

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestInMemoryTokenCache_HitAndMiss(t *testing.T) {
	c := NewInMemoryTokenCache()

	if _, ok := c.Get("missing"); ok {
		t.Fatal("expected miss on empty cache")
	}

	tok := &CachedToken{AccessToken: "abc", ExpiresAt: time.Now().Add(time.Hour)}
	c.Set("k1", tok)

	got, ok := c.Get("k1")
	if !ok {
		t.Fatal("expected hit after Set")
	}
	if got.AccessToken != "abc" {
		t.Errorf("expected 'abc', got %q", got.AccessToken)
	}
}

func TestInMemoryTokenCache_ExpiredEviction(t *testing.T) {
	c := NewInMemoryTokenCache()

	// Already-expired token should be evicted on read.
	c.Set("k1", &CachedToken{AccessToken: "old", ExpiresAt: time.Now().Add(-time.Second)})

	if _, ok := c.Get("k1"); ok {
		t.Fatal("expected expired token to be a miss")
	}
}

func TestInMemoryTokenCache_Delete(t *testing.T) {
	c := NewInMemoryTokenCache()

	c.Set("k1", &CachedToken{AccessToken: "abc", ExpiresAt: time.Now().Add(time.Hour)})
	c.Delete("k1")

	if _, ok := c.Get("k1"); ok {
		t.Fatal("expected miss after Delete")
	}

	// Deleting a missing key is a no-op.
	c.Delete("missing")
}

func TestInMemoryTokenCache_ConcurrentSafe(t *testing.T) {
	c := NewInMemoryTokenCache()

	const (
		goroutines = 50
		iterations = 500
	)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				key := fmt.Sprintf("k%d", i%10)
				c.Set(key, &CachedToken{
					AccessToken: fmt.Sprintf("tok-%d-%d", gid, i),
					ExpiresAt:   time.Now().Add(time.Hour),
				})
				_, _ = c.Get(key)
				if i%5 == 0 {
					c.Delete(key)
				}
			}
		}(g)
	}

	wg.Wait()

	// Nothing to assert beyond "no panic / no data race". Race
	// detector (go test -race) is the real verification here.
}

func TestCachedToken_IsExpired(t *testing.T) {
	now := time.Now()
	future := &CachedToken{AccessToken: "a", ExpiresAt: now.Add(time.Hour)}
	past := &CachedToken{AccessToken: "b", ExpiresAt: now.Add(-time.Hour)}

	if future.IsExpired(now, 0) {
		t.Error("future token reported as expired")
	}
	if !past.IsExpired(now, 0) {
		t.Error("past token not reported as expired")
	}
	// With skew: a token expiring in 10s is "expired" against a 30s skew.
	near := &CachedToken{AccessToken: "c", ExpiresAt: now.Add(10 * time.Second)}
	if !near.IsExpired(now, 30*time.Second) {
		t.Error("near-expiry token not expired against 30s skew")
	}

	// Nil receiver must not panic and must report expired.
	var nilToken *CachedToken
	if !nilToken.IsExpired(now, 0) {
		t.Error("nil token should report expired")
	}
}

func TestDeriveCacheKey_Deterministic(t *testing.T) {
	k1 := DeriveCacheKey("client-1", "https://api.signdocs.com.br", []string{"transactions:read", "webhooks:write"})
	k2 := DeriveCacheKey("client-1", "https://api.signdocs.com.br", []string{"transactions:read", "webhooks:write"})

	if k1 != k2 {
		t.Errorf("expected deterministic keys, got %q vs %q", k1, k2)
	}
}

func TestDeriveCacheKey_ScopeOrderInvariant(t *testing.T) {
	k1 := DeriveCacheKey("client-1", "https://api.x", []string{"a:read", "b:write", "c:read"})
	k2 := DeriveCacheKey("client-1", "https://api.x", []string{"c:read", "a:read", "b:write"})

	if k1 != k2 {
		t.Errorf("scope order leaked into key: %q vs %q", k1, k2)
	}
}

func TestDeriveCacheKey_TrailingSlashInvariant(t *testing.T) {
	k1 := DeriveCacheKey("c", "https://api.x", []string{"s"})
	k2 := DeriveCacheKey("c", "https://api.x/", []string{"s"})
	k3 := DeriveCacheKey("c", "https://api.x///", []string{"s"})

	if k1 != k2 || k2 != k3 {
		t.Errorf("trailing slash leaked into key: %q / %q / %q", k1, k2, k3)
	}
}

func TestDeriveCacheKey_PrefixAndLength(t *testing.T) {
	k := DeriveCacheKey("client-1", "https://api.x", []string{"s"})

	const prefix = "signdocs.oauth."
	if !strings.HasPrefix(k, prefix) {
		t.Errorf("missing prefix, got %q", k)
	}
	hex := strings.TrimPrefix(k, prefix)
	if len(hex) != 32 {
		t.Errorf("expected 32-char hex suffix, got %d chars (%q)", len(hex), hex)
	}
	for _, r := range hex {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			t.Errorf("non-hex character %q in key %q", r, k)
			break
		}
	}
}

func TestDeriveCacheKey_DoesNotLeakCredentials(t *testing.T) {
	// The clientID must not appear anywhere in the derived key.
	clientID := "SUPER-SECRET-CLIENT-ID"
	k := DeriveCacheKey(clientID, "https://api.x", []string{"s"})

	if strings.Contains(k, clientID) {
		t.Errorf("clientID leaked into cache key: %q", k)
	}
	if strings.Contains(strings.ToLower(k), strings.ToLower(clientID)) {
		t.Errorf("clientID leaked into cache key (case-insensitive): %q", k)
	}
}

func TestDeriveCacheKey_DifferentInputsDifferentKeys(t *testing.T) {
	base := DeriveCacheKey("c", "https://api.x", []string{"s"})

	cases := []struct {
		name string
		key  string
	}{
		{"different clientID", DeriveCacheKey("c2", "https://api.x", []string{"s"})},
		{"different baseURL", DeriveCacheKey("c", "https://api.y", []string{"s"})},
		{"different scopes", DeriveCacheKey("c", "https://api.x", []string{"s2"})},
		{"extra scope", DeriveCacheKey("c", "https://api.x", []string{"s", "s2"})},
	}
	for _, tc := range cases {
		if tc.key == base {
			t.Errorf("%s: expected different key from base %q, got same", tc.name, base)
		}
	}
}
