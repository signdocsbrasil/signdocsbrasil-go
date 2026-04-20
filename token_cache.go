package signdocsbrasil

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"
)

// CachedToken is the value stored in a TokenCache. It is safe to share
// across goroutines as long as the underlying cache is concurrency-safe.
type CachedToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

// IsExpired reports whether the token is past (or within skew of) its
// expiry. Callers typically pass a small positive skew (e.g. 30s) so
// that a token on the verge of expiring is refreshed preemptively; the
// default InMemoryTokenCache uses skew=0 when evicting on read and
// relies on the AuthHandler's own refresh buffer for preemption.
func (c *CachedToken) IsExpired(now time.Time, skew time.Duration) bool {
	if c == nil {
		return true
	}
	return !now.Before(c.ExpiresAt.Add(-skew))
}

// TokenCache is the pluggable store for OAuth2 access tokens. The SDK
// writes to the cache once per token fetch and reads on every outbound
// request. Implementations MUST be safe for concurrent use.
//
// The default implementation (InMemoryTokenCache) lives for the process
// lifetime and is suitable for long-running daemons. Stateless hosts
// (serverless, short-lived CLI invocations) should inject a shared-store
// implementation to avoid a token fetch on every cold start.
//
// Implementations SHOULD treat the key as opaque; the SDK derives it
// deterministically from credentials + baseURL + scopes via
// DeriveCacheKey. Implementations SHOULD return (nil, false) on any
// backend error rather than panicking or propagating the error — a
// cache miss simply forces a fresh token fetch.
type TokenCache interface {
	// Get returns the cached token for key. The bool is false when the
	// entry is missing or expired; implementations MAY evict expired
	// entries on read.
	Get(key string) (*CachedToken, bool)

	// Set stores token under key. Implementations SHOULD honor the
	// token's ExpiresAt as a TTL upper bound.
	Set(key string, token *CachedToken)

	// Delete removes the entry for key. Deleting a missing key is a
	// no-op.
	Delete(key string)
}

// InMemoryTokenCache is the default, process-local TokenCache. It is
// safe for concurrent use via an internal sync.Mutex.
type InMemoryTokenCache struct {
	mu    sync.Mutex
	store map[string]*CachedToken
}

// NewInMemoryTokenCache returns a new, empty InMemoryTokenCache. The
// returned value satisfies TokenCache.
func NewInMemoryTokenCache() TokenCache {
	return &InMemoryTokenCache{
		store: make(map[string]*CachedToken),
	}
}

// Get returns the token for key, evicting the entry if it has already
// expired at wall-clock time.
func (c *InMemoryTokenCache) Get(key string) (*CachedToken, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.store[key]
	if !ok || entry == nil {
		return nil, false
	}
	if entry.IsExpired(time.Now(), 0) {
		delete(c.store, key)
		return nil, false
	}
	return entry, true
}

// Set stores token under key.
func (c *InMemoryTokenCache) Set(key string, token *CachedToken) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = token
}

// Delete removes the entry for key.
func (c *InMemoryTokenCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

// DeriveCacheKey produces a deterministic cache key from the clientID,
// baseURL, and scopes. Scopes are sorted so that callers that pass the
// same scopes in different orders hit the same entry. The baseURL is
// right-trimmed of trailing slashes so that "https://api.x" and
// "https://api.x/" share a cache slot.
//
// The output is "signdocs.oauth." + the first 32 hex chars of
// SHA-256(material). Truncation is deliberate: a leaked cache key
// cannot be reversed to recover the clientID, but is still wide enough
// (128 bits) to avoid collision in practice.
//
// This helper is exported so that consumers implementing their own
// TokenCache can derive the same key the SDK uses and share tokens
// between processes.
func DeriveCacheKey(clientID, baseURL string, scopes []string) string {
	canonicalScopes := make([]string, len(scopes))
	copy(canonicalScopes, scopes)
	sort.Strings(canonicalScopes)

	material := clientID + "|" + strings.TrimRight(baseURL, "/") + "|" + strings.Join(canonicalScopes, " ")
	sum := sha256.Sum256([]byte(material))
	return "signdocs.oauth." + hex.EncodeToString(sum[:])[:32]
}
