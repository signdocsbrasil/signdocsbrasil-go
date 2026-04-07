package signdocsbrasil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"testing"
	"time"
)

func TestResolveConfig_Defaults(t *testing.T) {
	cfg, err := resolveConfig("client-1", []Option{WithClientSecret("secret")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("expected BaseURL %s, got %s", DefaultBaseURL, cfg.BaseURL)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Errorf("expected Timeout %v, got %v", DefaultTimeout, cfg.Timeout)
	}
	if cfg.MaxRetries != DefaultMaxRetries {
		t.Errorf("expected MaxRetries %d, got %d", DefaultMaxRetries, cfg.MaxRetries)
	}
	if len(cfg.Scopes) != len(DefaultScopes) {
		t.Errorf("expected %d scopes, got %d", len(DefaultScopes), len(cfg.Scopes))
	}
}

func TestResolveConfig_CustomValues(t *testing.T) {
	cfg, err := resolveConfig("client-1", []Option{
		WithClientSecret("secret"),
		WithBaseURL("https://custom.api.com"),
		WithTimeout(5 * time.Second),
		WithMaxRetries(2),
		WithScopes("custom:scope"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != "https://custom.api.com" {
		t.Errorf("expected custom BaseURL, got %s", cfg.BaseURL)
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", cfg.Timeout)
	}
	if cfg.MaxRetries != 2 {
		t.Errorf("expected 2 retries, got %d", cfg.MaxRetries)
	}
	if len(cfg.Scopes) != 1 || cfg.Scopes[0] != "custom:scope" {
		t.Errorf("expected custom scopes, got %v", cfg.Scopes)
	}
}

func TestResolveConfig_MissingClientID(t *testing.T) {
	_, err := resolveConfig("", []Option{WithClientSecret("secret")})
	if err == nil {
		t.Fatal("expected error for empty clientID")
	}
}

func TestResolveConfig_NoAuth(t *testing.T) {
	_, err := resolveConfig("client-1", nil)
	if err == nil {
		t.Fatal("expected error when no auth is provided")
	}
}

func TestResolveConfig_WithPrivateKey(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	cfg, err := resolveConfig("client-1", []Option{WithPrivateKey(key, "kid-1")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PrivateKey != key {
		t.Error("expected private key to be set")
	}
	if cfg.Kid != "kid-1" {
		t.Errorf("expected kid 'kid-1', got %s", cfg.Kid)
	}
}

func TestResolveConfig_PrivateKeyWithoutKid(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	_, err := resolveConfig("client-1", []Option{WithPrivateKey(key, "")})
	if err == nil {
		t.Fatal("expected error when kid is empty")
	}
}

func TestResolveConfig_WithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 10 * time.Second}
	cfg, err := resolveConfig("client-1", []Option{
		WithClientSecret("secret"),
		WithHTTPClient(custom),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPClient != custom {
		t.Error("expected custom HTTP client to be set")
	}
}
