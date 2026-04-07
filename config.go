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
