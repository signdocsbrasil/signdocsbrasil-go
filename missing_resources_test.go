package signdocsbrasil

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSteps_List(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/steps" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(StepListResponse{
				Steps: []StepDetail{
					{
						StepID:      "step_1",
						Type:        string(StepTypeClickAccept),
						Status:      string(StepStatusPending),
						Order:       1,
						MaxAttempts: 3,
					},
					{
						StepID:      "step_2",
						Type:        string(StepTypeOTPChallenge),
						Status:      string(StepStatusPending),
						Order:       2,
						MaxAttempts: 3,
					},
				},
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newStepsService(hc)
	steps, err := svc.List(context.Background(), "tx_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps.Steps))
	}
	if steps.Steps[0].StepID != "step_1" {
		t.Errorf("expected step_1, got %s", steps.Steps[0].StepID)
	}
	if steps.Steps[0].Type != string(StepTypeClickAccept) {
		t.Errorf("expected CLICK_ACCEPT, got %s", steps.Steps[0].Type)
	}
	if steps.Steps[1].StepID != "step_2" {
		t.Errorf("expected step_2, got %s", steps.Steps[1].StepID)
	}
}

func TestSteps_Start(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/steps/step_1/start" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(StartStepResponse{
				StepID:  "step_1",
				Type:    "CLICK_ACCEPT",
				Status:  "STARTED",
				Message: "Step started successfully",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newStepsService(hc)
	resp, err := svc.Start(context.Background(), "tx_1", "step_1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StepID != "step_1" {
		t.Errorf("expected step_1, got %s", resp.StepID)
	}
	if resp.Status != "STARTED" {
		t.Errorf("expected STARTED, got %s", resp.Status)
	}
	if resp.Message != "Step started successfully" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
}

func TestSteps_Complete(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/steps/step_1/complete" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(StepCompleteResponse{
				StepID:   "step_1",
				Type:     string(StepTypeClickAccept),
				Status:   string(StepStatusCompleted),
				Attempts: 1,
				Result: &StepResult{
					Click: &ClickResult{Accepted: true, TextVersion: "v1.0"},
				},
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newStepsService(hc)
	resp, err := svc.Complete(context.Background(), "tx_1", "step_1", &CompleteClickRequest{Accepted: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StepID != "step_1" {
		t.Errorf("expected step_1, got %s", resp.StepID)
	}
	if resp.Status != string(StepStatusCompleted) {
		t.Errorf("expected COMPLETED, got %s", resp.Status)
	}
	if resp.Result == nil || resp.Result.Click == nil {
		t.Fatal("expected Click result to be non-nil")
	}
	if !resp.Result.Click.Accepted {
		t.Error("expected Click.Accepted to be true")
	}
}

func TestSigning_Prepare(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/signing/prepare" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PrepareSigningResponse{
				SignatureRequestID: "sigreq-uuid-001",
				HashToSign:         "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				HashAlgorithm:      "SHA-256",
				SignatureAlgorithm: "RSASSA-PKCS1-v1_5",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newSigningService(hc)
	resp, err := svc.Prepare(context.Background(), "tx_1", &PrepareSigningRequest{
		CertificateChainPEMs: []string{"-----BEGIN CERTIFICATE-----\nMIIB...leaf...\n-----END CERTIFICATE-----"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SignatureRequestID != "sigreq-uuid-001" {
		t.Errorf("expected sigreq-uuid-001, got %s", resp.SignatureRequestID)
	}
	if resp.HashToSign == "" {
		t.Error("expected hashToSign to be non-empty")
	}
	if resp.HashAlgorithm != "SHA-256" {
		t.Errorf("expected SHA-256, got %s", resp.HashAlgorithm)
	}
	if resp.SignatureAlgorithm != "RSASSA-PKCS1-v1_5" {
		t.Errorf("expected RSASSA-PKCS1-v1_5, got %s", resp.SignatureAlgorithm)
	}
}

func TestSigning_Complete(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/signing/complete" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CompleteSigningResponse{
				StepID: "step-uuid-sign-001",
				Status: "COMPLETED",
				Result: CompleteSigningResult{
					DigitalSignature: CompleteSigningDigitalSignatureResult{
						CertificateSubject: "CN=JOAO SILVA:12345678901, OU=AC EXAMPLE, O=ICP-Brasil",
						CertificateSerial:  "1234567890ABCDEF",
						CertificateIssuer:  "CN=AC EXAMPLE v5, O=ICP-Brasil",
						Algorithm:          "SHA256withRSA",
						SignedAt:           "2024-11-15T12:05:00.000Z",
						SignedPDFHash:      "b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3",
						SignatureFieldName: "SignDocs_1",
					},
				},
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newSigningService(hc)
	resp, err := svc.Complete(context.Background(), "tx_1", &CompleteSigningRequest{
		SignatureRequestID: "sigreq-uuid-001",
		RawSignatureBase64: "MEUCIA...base64signature...",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StepID != "step-uuid-sign-001" {
		t.Errorf("expected step-uuid-sign-001, got %s", resp.StepID)
	}
	if resp.Status != "COMPLETED" {
		t.Errorf("expected COMPLETED, got %s", resp.Status)
	}
	if resp.Result.DigitalSignature.CertificateSubject != "CN=JOAO SILVA:12345678901, OU=AC EXAMPLE, O=ICP-Brasil" {
		t.Errorf("unexpected certificate subject: %s", resp.Result.DigitalSignature.CertificateSubject)
	}
	if resp.Result.DigitalSignature.Algorithm != "SHA256withRSA" {
		t.Errorf("expected SHA256withRSA, got %s", resp.Result.DigitalSignature.Algorithm)
	}
}

func TestEvidence_Get(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transactions/tx_1/evidence" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Evidence{
				TenantID:      "abc123",
				TransactionID: "tx_1",
				EvidenceID:    "ev_1",
				Status:        "COMPLETED",
				Signer: EvidenceSigner{
					Name:           "João Silva",
					CPF:            "12345678901",
					UserExternalID: "user-ext-001",
				},
				Steps: []EvidenceStep{
					{
						Type:        "CLICK_ACCEPT",
						Status:      "COMPLETED",
						CompletedAt: "2024-11-15T00:01:00.000Z",
					},
				},
				Document: &EvidenceDocument{
					Hash:     "a1b2c3d4",
					Filename: "contract.pdf",
				},
				CreatedAt:   "2024-11-15T00:00:00.000Z",
				CompletedAt: "2024-11-15T00:01:00.000Z",
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newEvidenceService(hc)
	ev, err := svc.Get(context.Background(), "tx_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.EvidenceID != "ev_1" {
		t.Errorf("expected ev_1, got %s", ev.EvidenceID)
	}
	if ev.Status != "COMPLETED" {
		t.Errorf("expected COMPLETED, got %s", ev.Status)
	}
	if ev.Signer.Name != "João Silva" {
		t.Errorf("expected signer João Silva, got %s", ev.Signer.Name)
	}
	if len(ev.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(ev.Steps))
	}
	if ev.Document == nil {
		t.Fatal("expected document to be non-nil")
	}
	if ev.Document.Filename != "contract.pdf" {
		t.Errorf("expected contract.pdf, got %s", ev.Document.Filename)
	}
}

func TestDocumentGroups_CombinedStamp(t *testing.T) {
	server, hc := setupResourceTest(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/document-groups/grp_1/combined-stamp" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CombinedStampResponse{
				GroupID:     "grp_1",
				SignerCount: 2,
				DownloadURL: "https://s3.example.com/stamped.pdf",
				ExpiresIn:   3600,
			})
			return
		}
		w.WriteHeader(404)
	})
	defer server.Close()

	svc := newDocumentGroupsService(hc)
	resp, err := svc.CombinedStamp(context.Background(), "grp_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GroupID != "grp_1" {
		t.Errorf("expected grp_1, got %s", resp.GroupID)
	}
	if resp.DownloadURL != "https://s3.example.com/stamped.pdf" {
		t.Errorf("unexpected download URL: %s", resp.DownloadURL)
	}
	if resp.ExpiresIn != 3600 {
		t.Errorf("expected expiresIn 3600, got %d", resp.ExpiresIn)
	}
}
