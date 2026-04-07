package signdocsbrasil

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestPagination_EmptyFirstPage(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(TransactionListResponse{
				Transactions: []Transaction{},
				Count:        0,
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	resp, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Transactions) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(resp.Transactions))
	}
	if resp.Count != 0 {
		t.Errorf("expected count 0, got %d", resp.Count)
	}
	if resp.NextToken != "" {
		t.Errorf("expected empty nextToken, got %q", resp.NextToken)
	}
}

func TestPagination_SinglePageNoNextToken(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(TransactionListResponse{
				Transactions: []Transaction{
					{TransactionID: "tx_1", Status: TransactionStatusCompleted},
					{TransactionID: "tx_2", Status: TransactionStatusCompleted},
				},
				Count: 2,
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	resp, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Transactions) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(resp.Transactions))
	}
	if resp.NextToken != "" {
		t.Errorf("expected empty nextToken, got %q", resp.NextToken)
	}
	if resp.Transactions[0].TransactionID != "tx_1" {
		t.Errorf("expected tx_1, got %s", resp.Transactions[0].TransactionID)
	}
	if resp.Transactions[1].TransactionID != "tx_2" {
		t.Errorf("expected tx_2, got %s", resp.Transactions[1].TransactionID)
	}
}

func TestPagination_AutoPaginateSinglePage(t *testing.T) {
	calls := 0
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions" && r.Method == "GET" {
			calls++
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(TransactionListResponse{
				Transactions: []Transaction{
					{TransactionID: "tx_1"},
					{TransactionID: "tx_2"},
				},
				Count: 2,
				// No NextToken — single page
			})
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

	if len(items) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(items))
	}
	if calls != 1 {
		t.Errorf("expected 1 API call for single page, got %d", calls)
	}
	if items[0].TransactionID != "tx_1" {
		t.Errorf("expected tx_1, got %s", items[0].TransactionID)
	}
	if items[1].TransactionID != "tx_2" {
		t.Errorf("expected tx_2, got %s", items[1].TransactionID)
	}
}
