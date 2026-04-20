package signdocsbrasil

import (
	"crypto/ecdsa"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

const (
	DefaultBaseURL    = "https://api.signdocs.com.br"
	DefaultTimeout    = 30 * time.Second
	DefaultMaxRetries = 5
)

// DefaultScopes are requested when no custom scopes are provided.
var DefaultScopes = []string{
	"transactions:read",
	"transactions:write",
	"steps:write",
	"evidence:read",
	"webhooks:write",
}

// Config holds the resolved configuration for a Client.
type Config struct {
	ClientID     string
	ClientSecret string
	PrivateKey   *ecdsa.PrivateKey
	Kid          string
	BaseURL      string
	Timeout      time.Duration
	MaxRetries   int
	Scopes       []string
	HTTPClient   *http.Client
	Logger       *slog.Logger

	// TokenCache persists OAuth tokens across requests. If nil, the
	// SDK installs an in-process NewInMemoryTokenCache. Inject a
	// shared-store implementation (Redis, Memcache, etc.) to share
	// tokens across stateless workers. See WithTokenCache.
	TokenCache TokenCache

	// OnResponse, when non-nil, is invoked after every completed
	// HTTP response with response-level observability metadata
	// (rate-limit counters, RFC 8594 deprecation headers, request
	// ID). Panics inside the callback are recovered and logged; they
	// never propagate to the request path. See WithOnResponse.
	OnResponse func(*ResponseMetadata)
}

// Option is a functional option for configuring the Client.
type Option func(*Config)

// WithClientSecret configures client_secret authentication.
func WithClientSecret(secret string) Option {
	return func(c *Config) {
		c.ClientSecret = secret
	}
}

// WithPrivateKey configures private_key_jwt authentication using ES256.
func WithPrivateKey(key *ecdsa.PrivateKey, kid string) Option {
	return func(c *Config) {
		c.PrivateKey = key
		c.Kid = kid
	}
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithTimeout overrides the default HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.Timeout = d
	}
}

// WithMaxRetries overrides the default maximum retry count.
func WithMaxRetries(n int) Option {
	return func(c *Config) {
		c.MaxRetries = n
	}
}

// WithScopes overrides the default OAuth2 scopes.
func WithScopes(scopes ...string) Option {
	return func(c *Config) {
		c.Scopes = scopes
	}
}

// WithHTTPClient provides a custom *http.Client for all API requests.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = client
	}
}

// WithLogger configures structured logging for HTTP requests.
// Only method, path, status code, and duration are logged.
// Authorization headers, request bodies, and tokens are never logged.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithTokenCache injects a custom TokenCache, replacing the default
// in-process InMemoryTokenCache. Stateless hosts (serverless, CLI
// tools) should use this to share OAuth tokens across workers and
// avoid a token fetch on every cold start.
//
// If c is nil, the default in-memory cache is retained.
func WithTokenCache(c TokenCache) Option {
	return func(cfg *Config) {
		if c != nil {
			cfg.TokenCache = c
		}
	}
}

// WithOnResponse registers fn as a response observer. fn is invoked
// after every completed HTTP response (including error responses) with
// the parsed ResponseMetadata. The SDK recovers from panics in fn and
// logs them via the configured Logger (or the default stdlib log if no
// Logger is set) — a misbehaving callback never takes down the request
// path.
//
// Typical uses: push rate-limit counters into metrics, surface
// Deprecation/Sunset warnings, correlate request IDs to local traces.
//
// If fn is nil, any previously installed observer is cleared.
func WithOnResponse(fn func(*ResponseMetadata)) Option {
	return func(cfg *Config) {
		cfg.OnResponse = fn
	}
}

func resolveConfig(clientID string, opts []Option) (*Config, error) {
	if clientID == "" {
		return nil, errors.New("signdocsbrasil: clientId is required")
	}

	cfg := &Config{
		ClientID:   clientID,
		BaseURL:    DefaultBaseURL,
		Timeout:    DefaultTimeout,
		MaxRetries: DefaultMaxRetries,
		Scopes:     DefaultScopes,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.ClientSecret == "" && cfg.PrivateKey == nil {
		return nil, errors.New("signdocsbrasil: either WithClientSecret or WithPrivateKey option is required")
	}
	if cfg.PrivateKey != nil && cfg.Kid == "" {
		return nil, errors.New("signdocsbrasil: kid is required when using WithPrivateKey")
	}

	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: cfg.Timeout}
	}

	return cfg, nil
}
