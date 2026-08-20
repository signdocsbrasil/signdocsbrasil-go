package signdocsbrasil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Minting a link is not a metered create. Sending an idempotency key would let a
// retry replay a URL that has already been consumed, instead of issuing the
// fresh one the caller asked for.
func TestSigningSessionsLinkPostsWithoutIdempotencyKey(t *testing.T) {
	var gotPath, gotMethod, gotIdemKey string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"t","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		gotPath, gotMethod = r.URL.Path, r.Method
		gotIdemKey = r.Header.Get("X-Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sessionId":"ss_1","transactionId":"tx_1","url":"https://sign.signdocs.com.br/s/ss_1?cs=abc","expiresAt":"2026-08-27T12:00:00.000Z","expiresIn":3600}`))
	}))
	defer srv.Close()

	client := newIntegrationClient(t, srv.URL)

	resp, err := client.SigningSessions.Link(context.Background(), "ss_1")
	if err != nil {
		t.Fatalf("link: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if want := "/v1/signing-sessions/ss_1/link"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if gotIdemKey != "" {
		t.Errorf("X-Idempotency-Key = %q, want empty", gotIdemKey)
	}
	if want := "https://sign.signdocs.com.br/s/ss_1?cs=abc"; resp.URL != want {
		t.Errorf("URL = %q, want %q", resp.URL, want)
	}
	if resp.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", resp.ExpiresIn)
	}
	if resp.TransactionID != "tx_1" {
		t.Errorf("TransactionID = %q, want tx_1", resp.TransactionID)
	}
}

// A completed session cannot be linked: a link to it would authenticate nothing.
// The 409 must surface as a ConflictError, not as a link the caller can't use.
func TestSigningSessionsLinkReturnsConflictWhenNotActive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"t","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"type":"about:blank","title":"Conflict","status":409,"detail":"Session cannot be linked in status: COMPLETED"}`))
	}))
	defer srv.Close()

	client := newIntegrationClient(t, srv.URL)

	_, err := client.SigningSessions.Link(context.Background(), "ss_done")
	if err == nil {
		t.Fatal("link: expected an error for a COMPLETED session, got nil")
	}
	if _, ok := err.(*ConflictError); !ok {
		t.Errorf("error type = %T, want *ConflictError", err)
	}
}
