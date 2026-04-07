package signdocsbrasil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTransactionTest(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *TransactionsService) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
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
	}
	auth := newAuthHandler(cfg)
	hc := newHTTPClient(cfg, auth)
	svc := newTransactionsService(hc)

	return server, svc
}

func TestTransactions_Create(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions" && r.Method == "POST" {
			if r.Header.Get("X-Idempotency-Key") == "" {
				t.Error("expected X-Idempotency-Key header")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Transaction{TransactionID: "tx_1", Status: TransactionStatusCreated})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	tx, err := svc.Create(context.Background(), &CreateTransactionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.TransactionID != "tx_1" {
		t.Errorf("expected tx_1, got %s", tx.TransactionID)
	}
}

func TestTransactions_CreateWithIdempotencyKey(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions" && r.Method == "POST" {
			key := r.Header.Get("X-Idempotency-Key")
			if key != "custom-key" {
				t.Errorf("expected custom-key, got %s", key)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Transaction{TransactionID: "tx_2"})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	_, err := svc.Create(context.Background(), &CreateTransactionRequest{}, WithIdempotencyKey("custom-key"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransactions_List(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions" && r.Method == "GET" {
			if r.URL.Query().Get("status") != "COMPLETED" {
				t.Errorf("expected status=COMPLETED, got %s", r.URL.Query().Get("status"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(TransactionListResponse{
				Transactions: []Transaction{{TransactionID: "tx_1"}},
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	resp, err := svc.List(context.Background(), &TransactionListParams{Status: TransactionStatusCompleted})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(resp.Transactions))
	}
}

func TestTransactions_Get(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/transactions/tx_1") && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Transaction{TransactionID: "tx_1", Status: TransactionStatusInProgress})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	tx, err := svc.Get(context.Background(), "tx_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.TransactionID != "tx_1" {
		t.Errorf("expected tx_1, got %s", tx.TransactionID)
	}
}

func TestTransactions_Cancel(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/transactions/tx_1") && r.Method == "DELETE" {
			w.Header().Set("Content-Type", "application/json")
			// Cancel returns 200 with body (not 204)
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(CancelTransactionResponse{
				TransactionID: "tx_1",
				Status:        string(TransactionStatusCancelled),
				CancelledAt:   "2024-11-15T00:00:00Z",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	tx, err := svc.Cancel(context.Background(), "tx_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Status != string(TransactionStatusCancelled) {
		t.Errorf("expected cancelled, got %s", tx.Status)
	}
}

func TestTransactions_Finalize(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/finalize") && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(FinalizeResponse{
				TransactionID: "tx_1",
				Status:        string(TransactionStatusCompleted),
				EvidenceID:    "ev-001",
				EvidenceHash:  "sha256-evidence",
				CompletedAt:   "2024-11-15T00:01:00Z",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	tx, err := svc.Finalize(context.Background(), "tx_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Status != string(TransactionStatusCompleted) {
		t.Errorf("expected completed, got %s", tx.Status)
	}
}

func TestTransactions_ListAutoPaginate(t *testing.T) {
	calls := 0
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions" && r.Method == "GET" {
			calls++
			w.Header().Set("Content-Type", "application/json")
			if calls == 1 {
				json.NewEncoder(w).Encode(TransactionListResponse{
					Transactions: []Transaction{{TransactionID: "tx_1"}, {TransactionID: "tx_2"}},
					NextToken:    "page2",
				})
			} else {
				json.NewEncoder(w).Encode(TransactionListResponse{
					Transactions: []Transaction{{TransactionID: "tx_3"}},
				})
			}
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	iter := svc.ListAutoPaginate(nil)
	var items []Transaction
	for iter.Next(context.Background()) {
		items = append(items, iter.Value())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("expected 3 transactions, got %d", len(items))
	}
}
