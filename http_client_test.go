package signdocsbrasil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testSetup(t *testing.T, apiHandler http.HandlerFunc) (*httptest.Server, *httpClient) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-token",
			"expires_in":   3600,
		})
	})
	mux.HandleFunc("/", apiHandler)

	server := httptest.NewServer(mux)

	cfg := &Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		BaseURL:      server.URL,
		MaxRetries:   0,
		Scopes:       []string{"transactions:read"},
		HTTPClient:   server.Client(),
	}

	auth := newAuthHandler(cfg)
	hc := newHTTPClient(cfg, auth)

	return server, hc
}

func TestHTTPClient_AuthorizationHeader(t *testing.T) {
	server, hc := testSetup(t, func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("expected Bearer token, got %s", authHeader)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	})
	defer server.Close()

	var result map[string]any
	err := hc.request(context.Background(), requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/test",
	}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClient_UserAgent(t *testing.T) {
	server, hc := testSetup(t, func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if !strings.Contains(ua, "signdocs-brasil-go/") {
			t.Errorf("expected User-Agent with signdocs-brasil-go, got %s", ua)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	})
	defer server.Close()

	var result map[string]any
	hc.request(context.Background(), requestOptions{Method: http.MethodGet, Path: "/v1/test"}, &result)
}

func TestHTTPClient_NoAuth(t *testing.T) {
	server, hc := testSetup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("expected no Authorization header for noAuth request")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})
	defer server.Close()

	var result map[string]any
	hc.request(context.Background(), requestOptions{
		Method: http.MethodGet,
		Path:   "/health",
		NoAuth: true,
	}, &result)
}

func TestHTTPClient_JSONBody(t *testing.T) {
	server, hc := testSetup(t, func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"tx_1"}`))
	})
	defer server.Close()

	var result map[string]any
	hc.request(context.Background(), requestOptions{
		Method: http.MethodPost,
		Path:   "/v1/transactions",
		Body:   map[string]string{"name": "test"},
	}, &result)
}

func TestHTTPClient_204NoContent(t *testing.T) {
	server, hc := testSetup(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	defer server.Close()

	err := hc.request(context.Background(), requestOptions{
		Method: http.MethodDelete,
		Path:   "/v1/webhooks/123",
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClient_400Error(t *testing.T) {
	server, hc := testSetup(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(ProblemDetail{Type: "about:blank", Title: "Bad Request", Status: 400})
	})
	defer server.Close()

	var result map[string]any
	err := hc.request(context.Background(), requestOptions{Method: http.MethodPost, Path: "/v1/test"}, &result)
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*BadRequestError); !ok {
		t.Errorf("expected BadRequestError, got %T", err)
	}
}

func TestHTTPClient_IdempotencyKey(t *testing.T) {
	server, hc := testSetup(t, func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Idempotency-Key")
		if key != "my-key" {
			t.Errorf("expected my-key, got %s", key)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"tx_1"}`))
	})
	defer server.Close()

	var result map[string]any
	hc.requestWithIdempotency(context.Background(), requestOptions{
		Method: http.MethodPost,
		Path:   "/v1/transactions",
		Body:   map[string]string{"name": "test"},
	}, &result, "my-key")
}

func TestHTTPClient_AutoIdempotencyKey(t *testing.T) {
	server, hc := testSetup(t, func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Idempotency-Key")
		if key == "" {
			t.Error("expected auto-generated idempotency key")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"tx_1"}`))
	})
	defer server.Close()

	var result map[string]any
	hc.requestWithIdempotency(context.Background(), requestOptions{
		Method: http.MethodPost,
		Path:   "/v1/transactions",
	}, &result, "")
}
