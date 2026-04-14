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
			w.Write([]byte(`{
				"evidenceId": "ev_1",
				"downloads": {
					"originalDocument": null,
					"evidencePack": {"url": "https://example.com/pack.p7m", "filename": "pack.p7m"},
					"finalPdf": null,
					"signedSignature": {"url": "https://example.com/signature.p7s", "filename": "signature.p7s"}
				}
			}`))
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newVerificationService(hc)
	resp, err := svc.Downloads(context.Background(), "ev_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Downloads.EvidencePack == nil || resp.Downloads.EvidencePack.Filename != "pack.p7m" {
		t.Errorf("expected evidencePack pack.p7m, got %+v", resp.Downloads.EvidencePack)
	}
	if resp.Downloads.SignedSignature == nil || resp.Downloads.SignedSignature.Filename != "signature.p7s" {
		t.Errorf("expected signedSignature signature.p7s, got %+v", resp.Downloads.SignedSignature)
	}
	if resp.Downloads.OriginalDocument != nil {
		t.Errorf("expected originalDocument nil, got %+v", resp.Downloads.OriginalDocument)
	}
}

func TestVerification_VerifyEnvelopeNoAuth(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/verify/envelope/env_1" && r.Method == "GET" {
			if r.Header.Get("Authorization") != "" {
				t.Error("expected no Authorization for envelope verification")
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"envelopeId": "env_1",
				"status": "COMPLETED",
				"signingMode": "SEQUENTIAL",
				"totalSigners": 2,
				"completedSessions": 2,
				"documentHash": "sha256:abc",
				"tenantName": "Acme",
				"tenantCnpj": "12345678000100",
				"signers": [
					{
						"signerIndex": 1,
						"displayName": "João Silva",
						"cpfCnpj": "12345678901",
						"status": "COMPLETED",
						"evidenceId": "ev_a",
						"completedAt": "2026-04-13T18:00:00Z"
					},
					{
						"signerIndex": 2,
						"displayName": "Maria Souza",
						"status": "COMPLETED",
						"evidenceId": "ev_b",
						"completedAt": "2026-04-13T18:30:00Z"
					}
				],
				"downloads": {
					"consolidatedSignature": {"url": "https://example.com/envelope.p7s", "filename": "signature.p7s"}
				},
				"createdAt": "2026-04-13T17:00:00Z",
				"completedAt": "2026-04-13T18:30:00Z"
			}`))
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newVerificationService(hc)
	resp, err := svc.VerifyEnvelope(context.Background(), "env_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.EnvelopeID != "env_1" || resp.SigningMode != "SEQUENTIAL" {
		t.Errorf("unexpected envelope: %+v", resp)
	}
	if len(resp.Signers) != 2 || resp.Signers[0].DisplayName != "João Silva" {
		t.Errorf("unexpected signers: %+v", resp.Signers)
	}
	if resp.Downloads == nil || resp.Downloads.ConsolidatedSignature == nil ||
		resp.Downloads.ConsolidatedSignature.Filename != "signature.p7s" {
		t.Errorf("unexpected downloads: %+v", resp.Downloads)
	}
	if resp.Downloads.CombinedSignedPDF != nil {
		t.Errorf("expected combinedSignedPdf nil, got %+v", resp.Downloads.CombinedSignedPDF)
	}
}
