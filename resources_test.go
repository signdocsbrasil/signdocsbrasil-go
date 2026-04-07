package signdocsbrasil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupResourceTest(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *httpClient) {
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

	return server, hc
}

func TestDocuments_Upload(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/document" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(DocumentUploadResponse{TransactionID: "tx_1", DocumentHash: "sha256-abc", Status: "DOCUMENT_UPLOADED", UploadedAt: "2024-11-15T00:00:00Z"})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newDocumentsService(hc)
	resp, err := svc.Upload(context.Background(), "tx_1", &UploadDocumentRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TransactionID != "tx_1" {
		t.Errorf("expected tx_1, got %s", resp.TransactionID)
	}
}

func TestDocuments_Presign(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/document/presign" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PresignResponse{UploadURL: "https://s3.example.com"})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newDocumentsService(hc)
	resp, err := svc.Presign(context.Background(), "tx_1", &PresignRequest{
		ContentType: "application/pdf",
		Filename:    "contract.pdf",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UploadURL == "" {
		t.Error("expected upload URL")
	}
}

func TestDocuments_Download(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/download" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(DownloadResponse{
				TransactionID: "tx_1",
				OriginalURL:   "https://s3.example.com/download",
				ExpiresIn:     3600,
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newDocumentsService(hc)
	resp, err := svc.Download(context.Background(), "tx_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.OriginalURL == "" {
		t.Error("expected original URL")
	}
}

func TestHealth_CheckNoAuth(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" && r.Method == "GET" {
			if r.Header.Get("Authorization") != "" {
				t.Error("expected no Authorization header for health check")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(HealthCheckResponse{Status: "healthy"})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newHealthService(hc)
	resp, err := svc.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "healthy" {
		t.Errorf("expected healthy, got %s", resp.Status)
	}
}

func TestHealth_HistoryNoAuth(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health/history" {
			if r.Header.Get("Authorization") != "" {
				t.Error("expected no Authorization header for health history")
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"entries":[]}`))
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newHealthService(hc)
	_, err := svc.History(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWebhooks_Register(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/webhooks" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(RegisterWebhookResponse{
				WebhookID: "wh_1",
				Secret:    "whsec_123",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newWebhooksService(hc)
	resp, err := svc.Register(context.Background(), &RegisterWebhookRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.WebhookID != "wh_1" {
		t.Errorf("expected wh_1, got %s", resp.WebhookID)
	}
}

func TestWebhooks_Delete204(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/webhooks/wh_1" && r.Method == "DELETE" {
			w.WriteHeader(204)
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newWebhooksService(hc)
	err := svc.Delete(context.Background(), "wh_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWebhooks_List(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/webhooks" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Webhook{{WebhookID: "wh_1"}, {WebhookID: "wh_2"}})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newWebhooksService(hc)
	result, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 webhooks, got %d", len(result))
	}
}

func TestUsers_EnrollPUT(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/users/ext_1/enrollment" && r.Method == "PUT" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(EnrollUserResponse{
				UserExternalID:    "ext_1",
				EnrollmentHash:    "sha256-enroll",
				EnrollmentVersion: 1,
				EnrollmentSource:  "BANK_PROVIDED",
				CPF:               "12345678901",
				FaceConfidence:    0.98,
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newUsersService(hc)
	resp, err := svc.Enroll(context.Background(), "ext_1", &EnrollUserRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UserExternalID != "ext_1" {
		t.Errorf("expected ext_1, got %s", resp.UserExternalID)
	}
}

func TestVerification_VerifyNoAuth(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/verify/ev_1" && r.Method == "GET" {
			if r.Header.Get("Authorization") != "" {
				t.Error("expected no Authorization for verification")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(VerificationResponse{
				EvidenceID: "ev_1",
				Status:     "COMPLETED",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newVerificationService(hc)
	resp, err := svc.Verify(context.Background(), "ev_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "COMPLETED" {
		t.Errorf("expected COMPLETED, got %s", resp.Status)
	}
}

func TestVerification_DownloadsNoAuth(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/verify/ev_1/downloads" && r.Method == "GET" {
			if r.Header.Get("Authorization") != "" {
				t.Error("expected no Authorization for verification downloads")
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"files":[]}`))
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newVerificationService(hc)
	_, err := svc.Downloads(context.Background(), "ev_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
