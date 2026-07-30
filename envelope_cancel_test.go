package signdocsbrasil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// An envelope must be cancelled through its own endpoint. Cancelling the member
// sessions one by one is not equivalent: it leaves the envelope itself ACTIVE.
func TestEnvelopesCancelPostsToTheEnvelopeEndpoint(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"t","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		gotPath, gotMethod = r.URL.Path, r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"envelopeId":"env_1","status":"CANCELLED","cancelledCount":2,"preservedSignedCount":1}`))
	}))
	defer srv.Close()

	client := newIntegrationClient(t, srv.URL)

	resp, err := client.Envelopes.Cancel(context.Background(), "env_1", "owner_cancelled")
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if want := "/v1/envelopes/env_1/cancel"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if gotBody["reason"] != "owner_cancelled" {
		t.Errorf("reason = %q", gotBody["reason"])
	}
	if resp.CancelledCount != 2 {
		t.Errorf("CancelledCount = %d, want 2", resp.CancelledCount)
	}
	// Signatures already collected are never invalidated by cancelling.
	if resp.PreservedSignedCount != 1 {
		t.Errorf("PreservedSignedCount = %d, want 1", resp.PreservedSignedCount)
	}
}

func TestEnvelopesCancelOmitsReasonWhenEmpty(t *testing.T) {
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"t","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"envelopeId":"env_1","status":"CANCELLED","alreadyCancelled":true}`))
	}))
	defer srv.Close()

	client := newIntegrationClient(t, srv.URL)
	resp, err := client.Envelopes.Cancel(context.Background(), "env_1", "")
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if _, present := gotBody["reason"]; present {
		t.Errorf("reason should be omitted, got %v", gotBody)
	}
	// Re-cancelling is a safe no-op, not an error.
	if !resp.AlreadyCancelled {
		t.Error("AlreadyCancelled = false, want true")
	}
}
