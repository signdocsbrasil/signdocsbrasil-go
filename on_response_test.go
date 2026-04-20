package signdocsbrasil

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// onResponseSetup mirrors testSetup but lets the caller install an
// onResponse observer and custom HTTP handler.
func onResponseSetup(t *testing.T, observer func(*ResponseMetadata), apiHandler http.HandlerFunc) (*httptest.Server, *httpClient) {
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
		OnResponse:   observer,
	}

	auth := newAuthHandler(cfg)
	hc := newHTTPClient(cfg, auth)

	return server, hc
}

func TestOnResponse_FiresOnSuccess(t *testing.T) {
	var captured *ResponseMetadata
	var mu sync.Mutex

	observer := func(m *ResponseMetadata) {
		mu.Lock()
		defer mu.Unlock()
		captured = m
	}

	server, hc := onResponseSetup(t, observer, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("RateLimit-Limit", "500")
		w.Header().Set("RateLimit-Remaining", "499")
		w.Header().Set("X-Request-Id", "req_test")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	})
	defer server.Close()

	var result map[string]any
	if err := hc.request(context.Background(), requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/transactions",
	}, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if captured == nil {
		t.Fatal("onResponse was not invoked")
	}
	if captured.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", captured.StatusCode)
	}
	if captured.Method != "GET" {
		t.Errorf("expected method 'GET', got %q", captured.Method)
	}
	if captured.Path != "/v1/transactions" {
		t.Errorf("expected path '/v1/transactions', got %q", captured.Path)
	}
	if captured.RateLimitLimit == nil || *captured.RateLimitLimit != 500 {
		t.Errorf("expected RateLimitLimit=500, got %v", captured.RateLimitLimit)
	}
	if captured.RequestID == nil || *captured.RequestID != "req_test" {
		t.Errorf("expected RequestID 'req_test', got %v", captured.RequestID)
	}
}

func TestOnResponse_FiresOnError(t *testing.T) {
	var calls int32
	observer := func(m *ResponseMetadata) {
		atomic.AddInt32(&calls, 1)
		if m.StatusCode != 400 {
			t.Errorf("expected status 400 in observer, got %d", m.StatusCode)
		}
	}

	server, hc := onResponseSetup(t, observer, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(ProblemDetail{Type: "about:blank", Title: "Bad Request", Status: 400})
	})
	defer server.Close()

	var result map[string]any
	err := hc.request(context.Background(), requestOptions{Method: http.MethodPost, Path: "/v1/x"}, &result)
	if err == nil {
		t.Fatal("expected error from 400 response")
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected observer to fire once for an error response, got %d calls", got)
	}
}

func TestOnResponse_CallbackPanicDoesNotBreakRequest(t *testing.T) {
	// Redirect the stdlib log output to capture the panic message.
	var logBuf bytes.Buffer
	oldOutput := log.Writer()
	log.SetOutput(&logBuf)
	defer log.SetOutput(oldOutput)

	observer := func(m *ResponseMetadata) {
		panic("boom — user callback is misbehaving")
	}

	server, hc := onResponseSetup(t, observer, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	})
	defer server.Close()

	var result map[string]any
	err := hc.request(context.Background(), requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/x",
	}, &result)
	if err != nil {
		t.Fatalf("panic in observer should not surface as request error, got: %v", err)
	}

	// The observer panic must have been logged through our fallback.
	if !strings.Contains(logBuf.String(), "onResponse callback panicked") {
		t.Errorf("expected panic log entry, got:\n%s", logBuf.String())
	}
}

func TestOnResponse_NotInvokedWhenNil(t *testing.T) {
	// Should not crash — just a smoke test that nil OnResponse is a no-op.
	server, hc := onResponseSetup(t, nil, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	})
	defer server.Close()

	var result map[string]any
	if err := hc.request(context.Background(), requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/x",
	}, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
