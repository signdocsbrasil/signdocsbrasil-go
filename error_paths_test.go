package signdocsbrasil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestErrorPath_TransactionCreate400(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(ProblemDetail{
				Type:   "https://api.signdocs.com.br/errors/bad-request",
				Title:  "Bad Request",
				Status: 400,
				Detail: "Invalid policy profile: UNKNOWN_PROFILE",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	_, err := svc.Create(context.Background(), &CreateTransactionRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Fatalf("expected *BadRequestError, got %T: %v", err, err)
	}
	if badReq.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", badReq.StatusCode)
	}
	if badReq.ProblemDetail.Detail != "Invalid policy profile: UNKNOWN_PROFILE" {
		t.Errorf("expected detail about invalid policy, got %q", badReq.ProblemDetail.Detail)
	}
}

func TestErrorPath_TransactionGet404(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/nonexistent" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(404)
			json.NewEncoder(w).Encode(ProblemDetail{
				Type:   "https://api.signdocs.com.br/errors/not-found",
				Title:  "Not Found",
				Status: 404,
				Detail: "Transaction tx-nonexistent not found",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	_, err := svc.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
	}
	if notFound.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", notFound.StatusCode)
	}
	if notFound.ProblemDetail.Title != "Not Found" {
		t.Errorf("expected title 'Not Found', got %q", notFound.ProblemDetail.Title)
	}
}

func TestErrorPath_TransactionCreate409(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(409)
			json.NewEncoder(w).Encode(ProblemDetail{
				Type:   "https://api.signdocs.com.br/errors/conflict",
				Title:  "Conflict",
				Status: 409,
				Detail: "Transaction tx-uuid-001 is already finalized",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	_, err := svc.Create(context.Background(), &CreateTransactionRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected *ConflictError, got %T: %v", err, err)
	}
	if conflict.StatusCode != 409 {
		t.Errorf("expected status 409, got %d", conflict.StatusCode)
	}
	if conflict.ProblemDetail.Detail != "Transaction tx-uuid-001 is already finalized" {
		t.Errorf("unexpected detail: %q", conflict.ProblemDetail.Detail)
	}
}

func TestErrorPath_DocumentUpload422(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/document" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(422)
			json.NewEncoder(w).Encode(ProblemDetail{
				Type:   "https://api.signdocs.com.br/errors/unprocessable-entity",
				Title:  "Unprocessable Entity",
				Status: 422,
				Detail: "CPF must be exactly 11 digits",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newDocumentsService(hc)
	_, err := svc.Upload(context.Background(), "tx_1", &UploadDocumentRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var unprocessable *UnprocessableEntityError
	if !errors.As(err, &unprocessable) {
		t.Fatalf("expected *UnprocessableEntityError, got %T: %v", err, err)
	}
	if unprocessable.StatusCode != 422 {
		t.Errorf("expected status 422, got %d", unprocessable.StatusCode)
	}
	if unprocessable.ProblemDetail.Detail != "CPF must be exactly 11 digits" {
		t.Errorf("unexpected detail: %q", unprocessable.ProblemDetail.Detail)
	}
}

func TestErrorPath_DocumentConfirm400(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/document/confirm" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(ProblemDetail{
				Type:   "https://api.signdocs.com.br/errors/bad-request",
				Title:  "Bad Request",
				Status: 400,
				Detail: "Missing sha256Hash field",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newDocumentsService(hc)
	_, err := svc.Confirm(context.Background(), "tx_1", &ConfirmDocumentRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Fatalf("expected *BadRequestError, got %T: %v", err, err)
	}
	if badReq.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", badReq.StatusCode)
	}
}

func TestErrorPath_WebhookRegister400(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/webhooks" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(ProblemDetail{
				Type:   "https://api.signdocs.com.br/errors/bad-request",
				Title:  "Bad Request",
				Status: 400,
				Detail: "URL must be HTTPS",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newWebhooksService(hc)
	_, err := svc.Register(context.Background(), &RegisterWebhookRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Fatalf("expected *BadRequestError, got %T: %v", err, err)
	}
	if badReq.ProblemDetail.Detail != "URL must be HTTPS" {
		t.Errorf("unexpected detail: %q", badReq.ProblemDetail.Detail)
	}
}

func TestErrorPath_TransactionFinalize409(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/finalize" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(409)
			json.NewEncoder(w).Encode(ProblemDetail{
				Type:   "https://api.signdocs.com.br/errors/conflict",
				Title:  "Conflict",
				Status: 409,
				Detail: "Transaction tx_1 is already finalized",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	_, err := svc.Finalize(context.Background(), "tx_1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected *ConflictError, got %T: %v", err, err)
	}
	if conflict.StatusCode != 409 {
		t.Errorf("expected status 409, got %d", conflict.StatusCode)
	}
}

func TestErrorPath_TransactionCancel400(t *testing.T) {
	server, svc := setupTransactionTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1" && r.Method == "DELETE" {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(ProblemDetail{
				Type:   "https://api.signdocs.com.br/errors/bad-request",
				Title:  "Bad Request",
				Status: 400,
				Detail: "Transaction cannot be cancelled in current state",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	_, err := svc.Cancel(context.Background(), "tx_1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Fatalf("expected *BadRequestError, got %T: %v", err, err)
	}
	if badReq.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", badReq.StatusCode)
	}
}
