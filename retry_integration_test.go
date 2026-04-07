package signdocsbrasil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

var successTxJSON = `{
	"tenantId": "abc123",
	"transactionId": "tx-001",
	"status": "CREATED",
	"purpose": "DOCUMENT_SIGNATURE",
	"policy": {"profile": "CLICK_ONLY"},
	"signer": {"name": "Test", "userExternalId": "u1"},
	"steps": [],
	"expiresAt": "2024-12-31T00:00:00Z",
	"createdAt": "2024-01-01T00:00:00Z",
	"updatedAt": "2024-01-01T00:00:00Z"
}`

func retryServer(errorCount int, errorStatus int, errorHeaders map[string]string) (*httptest.Server, *int32) {
	var apiCallCount int32
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(tokenResponseJSON))
			return
		}

		n := int(atomic.AddInt32(&apiCallCount, 1))
		if n <= errorCount {
			for k, v := range errorHeaders {
				w.Header().Set(k, v)
			}
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(errorStatus)
			body := map[string]any{
				"type":   "https://api.signdocs.com.br/errors/test",
				"title":  http.StatusText(errorStatus),
				"status": errorStatus,
				"detail": "test error",
			}
			json.NewEncoder(w).Encode(body)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(successTxJSON))
	})), &apiCallCount
}

func retryClient(t *testing.T, serverURL string, maxRetries int) *Client {
	t.Helper()
	client, err := NewClient("test-client",
		WithClientSecret("test-secret"),
		WithBaseURL(serverURL),
		WithMaxRetries(maxRetries),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client
}

func TestRetryIntegration_503Then200(t *testing.T) {
	server, count := retryServer(1, 503, nil)
	defer server.Close()

	client := retryClient(t, server.URL, 3)
	tx, err := client.Transactions.Get(context.Background(), "tx-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.TransactionID != "tx-001" {
		t.Errorf("TransactionID = %q, want tx-001", tx.TransactionID)
	}
	if got := atomic.LoadInt32(count); got != 2 {
		t.Errorf("API call count = %d, want 2", got)
	}
}

func TestRetryIntegration_429RetryAfterThen200(t *testing.T) {
	server, count := retryServer(1, 429, map[string]string{"Retry-After": "1"})
	defer server.Close()

	client := retryClient(t, server.URL, 3)
	tx, err := client.Transactions.Get(context.Background(), "tx-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.TransactionID != "tx-001" {
		t.Errorf("TransactionID = %q, want tx-001", tx.TransactionID)
	}
	if got := atomic.LoadInt32(count); got != 2 {
		t.Errorf("API call count = %d, want 2", got)
	}
}

func TestRetryIntegration_503x3Then200(t *testing.T) {
	server, count := retryServer(3, 503, nil)
	defer server.Close()

	client := retryClient(t, server.URL, 3)
	tx, err := client.Transactions.Get(context.Background(), "tx-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.TransactionID != "tx-001" {
		t.Errorf("TransactionID = %q, want tx-001", tx.TransactionID)
	}
	if got := atomic.LoadInt32(count); got != 4 {
		t.Errorf("API call count = %d, want 4", got)
	}
}

func TestRetryIntegration_503x4_ExhaustsRetries(t *testing.T) {
	server, count := retryServer(4, 503, nil)
	defer server.Close()

	client := retryClient(t, server.URL, 3)
	_, err := client.Transactions.Get(context.Background(), "tx-001")

	var svcErr *ServiceUnavailableError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected ServiceUnavailableError, got %T: %v", err, err)
	}
	if got := atomic.LoadInt32(count); got != 4 {
		t.Errorf("API call count = %d, want 4", got)
	}
}

func TestRetryIntegration_NonRetryable400_NoRetry(t *testing.T) {
	server, count := retryServer(1, 400, nil)
	defer server.Close()

	client := retryClient(t, server.URL, 3)
	_, err := client.Transactions.Get(context.Background(), "tx-001")

	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Fatalf("expected BadRequestError, got %T: %v", err, err)
	}
	// Should NOT retry 400 - only 1 API call
	if got := atomic.LoadInt32(count); got != 1 {
		t.Errorf("API call count = %d, want 1 (no retry for 400)", got)
	}
}
