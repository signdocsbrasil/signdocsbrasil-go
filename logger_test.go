package signdocsbrasil

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// captureHandler is a custom slog.Handler that records all log entries for test assertions.
type captureHandler struct {
	records []slog.Record
	mu      sync.Mutex
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r)
	return nil
}

func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func (h *captureHandler) getRecords() []slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	dst := make([]slog.Record, len(h.records))
	copy(dst, h.records)
	return dst
}

// testSetupWithLogger mirrors testSetup but injects a logger into the httpClient.
func testSetupWithLogger(t *testing.T, handler http.HandlerFunc, logger *slog.Logger) (*httptest.Server, *httpClient) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-token",
			"expires_in":   3600,
		})
	})
	mux.HandleFunc("/", handler)

	server := httptest.NewServer(mux)

	cfg := &Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		BaseURL:      server.URL,
		MaxRetries:   0,
		Scopes:       []string{"transactions:read"},
		HTTPClient:   server.Client(),
		Logger:       logger,
	}

	auth := newAuthHandler(cfg)
	hc := newHTTPClient(cfg, auth)

	return server, hc
}

func TestLogger_SuccessfulRequest(t *testing.T) {
	ch := &captureHandler{}
	logger := slog.New(ch)

	server, hc := testSetupWithLogger(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}, logger)
	defer server.Close()

	var result map[string]any
	err := hc.request(context.Background(), requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/transactions",
	}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := ch.getRecords()
	// Filter for our request log (skip the /oauth2/token request)
	var found *slog.Record
	for i := range records {
		r := &records[i]
		if r.Message == "request completed" {
			found = r
			break
		}
	}
	if found == nil {
		t.Fatal("expected 'request completed' log record")
	}
	if found.Level != slog.LevelInfo {
		t.Errorf("expected Info level, got %v", found.Level)
	}

	attrs := recordAttrs(found)
	if attrs["method"] != "GET" {
		t.Errorf("expected method GET, got %v", attrs["method"])
	}
	if attrs["path"] != "/v1/transactions" {
		t.Errorf("expected path /v1/transactions, got %v", attrs["path"])
	}
	if attrs["status"] != int64(200) {
		t.Errorf("expected status 200, got %v", attrs["status"])
	}
	if _, ok := attrs["duration_ms"]; !ok {
		t.Error("expected duration_ms attribute")
	}
}

func TestLogger_ErrorResponse(t *testing.T) {
	ch := &captureHandler{}
	logger := slog.New(ch)

	server, hc := testSetupWithLogger(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(ProblemDetail{Type: "about:blank", Title: "Bad Request", Status: 400})
	}, logger)
	defer server.Close()

	var result map[string]any
	_ = hc.request(context.Background(), requestOptions{
		Method: http.MethodPost,
		Path:   "/v1/transactions",
		Body:   map[string]string{"name": "test"},
	}, &result)

	records := ch.getRecords()
	var found *slog.Record
	for i := range records {
		r := &records[i]
		if r.Message == "request failed" {
			found = r
			break
		}
	}
	if found == nil {
		t.Fatal("expected 'request failed' log record")
	}
	if found.Level != slog.LevelWarn {
		t.Errorf("expected Warn level, got %v", found.Level)
	}

	attrs := recordAttrs(found)
	if attrs["method"] != "POST" {
		t.Errorf("expected method POST, got %v", attrs["method"])
	}
	if attrs["path"] != "/v1/transactions" {
		t.Errorf("expected path /v1/transactions, got %v", attrs["path"])
	}
	if attrs["status"] != int64(400) {
		t.Errorf("expected status 400, got %v", attrs["status"])
	}
}

func TestLogger_NilLoggerNoPanic(t *testing.T) {
	server, hc := testSetupWithLogger(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}, nil)
	defer server.Close()

	var result map[string]any
	err := hc.request(context.Background(), requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/test",
	}, &result)
	if err != nil {
		t.Fatalf("unexpected error with nil logger: %v", err)
	}
}

func TestLogger_NoAuthorizationHeaderInLogs(t *testing.T) {
	ch := &captureHandler{}
	logger := slog.New(ch)

	server, hc := testSetupWithLogger(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}, logger)
	defer server.Close()

	var result map[string]any
	err := hc.request(context.Background(), requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/test",
	}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := ch.getRecords()
	for _, r := range records {
		r.Attrs(func(a slog.Attr) bool {
			key := a.Key
			if key == "authorization" || key == "Authorization" || key == "token" || key == "bearer" {
				t.Errorf("sensitive attribute %q found in log record", key)
			}
			return true
		})
	}
}

func TestLogger_ServerError(t *testing.T) {
	ch := &captureHandler{}
	logger := slog.New(ch)

	server, hc := testSetupWithLogger(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"title":"Internal Server Error","status":500}`))
	}, logger)
	defer server.Close()

	var result map[string]any
	_ = hc.request(context.Background(), requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/health",
	}, &result)

	records := ch.getRecords()
	var found *slog.Record
	for i := range records {
		r := &records[i]
		if r.Message == "request failed" {
			found = r
			break
		}
	}
	if found == nil {
		t.Fatal("expected 'request failed' log record for 500 response")
	}
	if found.Level != slog.LevelWarn {
		t.Errorf("expected Warn level for 500, got %v", found.Level)
	}

	attrs := recordAttrs(found)
	if attrs["status"] != int64(500) {
		t.Errorf("expected status 500, got %v", attrs["status"])
	}
}

func TestWithLogger_Option(t *testing.T) {
	ch := &captureHandler{}
	logger := slog.New(ch)

	cfg, err := resolveConfig("client-1", []Option{
		WithClientSecret("secret"),
		WithLogger(logger),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Logger != logger {
		t.Error("expected logger to be set on config")
	}
}

// recordAttrs extracts all attributes from a slog.Record into a map for easy assertion.
func recordAttrs(r *slog.Record) map[string]any {
	attrs := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	return attrs
}
