package signdocsbrasil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixture is the top-level structure of a shared JSON fixture file.
type fixture struct {
	Description     string          `json:"description"`
	Input           json.RawMessage `json:"input"`
	ExpectedRequest json.RawMessage `json:"expected_request"`
	Response        struct {
		Status  int               `json:"status"`
		Headers map[string]string `json:"headers"`
		Body    json.RawMessage   `json:"body"`
	} `json:"response"`
	ExpectedError json.RawMessage `json:"expected_error"`
}

func loadFixture(t *testing.T, name string) fixture {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("fixtures", name+".json"))
	if err != nil {
		t.Fatalf("failed to load fixture %s: %v", name, err)
	}
	var f fixture
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("failed to parse fixture %s: %v", name, err)
	}
	return f
}

// tokenResponseJSON is a valid OAuth2 token response for mock servers.
var tokenResponseJSON = `{
	"access_token": "test-integration-token",
	"token_type": "Bearer",
	"expires_in": 900,
	"scope": "transactions:read transactions:write"
}`

// newIntegrationServer creates an httptest.Server that routes requests to
// the appropriate handler based on the request path. tokenHandler handles
// /oauth2/token; apiHandler handles all other paths.
func newIntegrationServer(t *testing.T, f fixture) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(tokenResponseJSON))
			return
		}

		for k, v := range f.Response.Headers {
			w.Header().Set(k, v)
		}
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(f.Response.Status)
		w.Write(f.Response.Body)
	}))
}

func newIntegrationClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	client, err := NewClient("test-client-id",
		WithClientSecret("test-secret"),
		WithBaseURL(serverURL),
		WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client
}

func TestIntegration_TransactionsCreate(t *testing.T) {
	f := loadFixture(t, "transactions-create")
	server := newIntegrationServer(t, f)
	defer server.Close()

	client := newIntegrationClient(t, server.URL)
	ctx := context.Background()

	result, err := client.Transactions.Create(ctx, &CreateTransactionRequest{
		Purpose:  TransactionPurposeDocumentSignature,
		Policy:   Policy{Profile: PolicyProfileClickOnly},
		Signer:   Signer{Name: "João Silva", Email: "joao@example.com", UserExternalID: "user-ext-001", CPF: "12345678901"},
		Metadata: map[string]string{"contractId": "CTR-2024-001"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TenantID != "abc123" {
		t.Errorf("TenantID = %q, want %q", result.TenantID, "abc123")
	}
	if result.TransactionID != "tx-uuid-001" {
		t.Errorf("TransactionID = %q, want %q", result.TransactionID, "tx-uuid-001")
	}
	if result.Status != TransactionStatusCreated {
		t.Errorf("Status = %q, want %q", result.Status, TransactionStatusCreated)
	}
	if result.Policy.Profile != PolicyProfileClickOnly {
		t.Errorf("Policy.Profile = %q, want %q", result.Policy.Profile, PolicyProfileClickOnly)
	}
	if result.Signer.Name != "João Silva" {
		t.Errorf("Signer.Name = %q, want %q", result.Signer.Name, "João Silva")
	}
	if len(result.Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1", len(result.Steps))
	}
	if result.Steps[0].StepID != "step-uuid-001" {
		t.Errorf("Steps[0].StepID = %q, want %q", result.Steps[0].StepID, "step-uuid-001")
	}
	if result.Steps[0].Type != StepTypeClickAccept {
		t.Errorf("Steps[0].Type = %q, want %q", result.Steps[0].Type, StepTypeClickAccept)
	}
	if result.Metadata["contractId"] != "CTR-2024-001" {
		t.Errorf("Metadata[contractId] = %q, want %q", result.Metadata["contractId"], "CTR-2024-001")
	}
}

func TestIntegration_TransactionsList(t *testing.T) {
	f := loadFixture(t, "transactions-list")
	server := newIntegrationServer(t, f)
	defer server.Close()

	client := newIntegrationClient(t, server.URL)
	ctx := context.Background()

	result, err := client.Transactions.List(ctx, &TransactionListParams{Status: TransactionStatusCompleted, Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Transactions) != 2 {
		t.Fatalf("len(Transactions) = %d, want 2", len(result.Transactions))
	}
	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}
	if result.NextToken == "" {
		t.Error("NextToken should not be empty")
	}
	if result.Transactions[0].TransactionID != "tx-uuid-002" {
		t.Errorf("Transactions[0].TransactionID = %q, want %q", result.Transactions[0].TransactionID, "tx-uuid-002")
	}
	if result.Transactions[1].Policy.Profile != PolicyProfileBiometric {
		t.Errorf("Transactions[1].Policy.Profile = %q, want %q", result.Transactions[1].Policy.Profile, PolicyProfileBiometric)
	}
}

func TestIntegration_TransactionsGet(t *testing.T) {
	f := loadFixture(t, "transactions-get")
	server := newIntegrationServer(t, f)
	defer server.Close()

	client := newIntegrationClient(t, server.URL)
	ctx := context.Background()

	result, err := client.Transactions.Get(ctx, "tx-uuid-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != TransactionStatusInProgress {
		t.Errorf("Status = %q, want %q", result.Status, TransactionStatusInProgress)
	}
	if len(result.Steps) != 2 {
		t.Fatalf("len(Steps) = %d, want 2", len(result.Steps))
	}
	if result.Steps[0].Status != StepStatusCompleted {
		t.Errorf("Steps[0].Status = %q, want %q", result.Steps[0].Status, StepStatusCompleted)
	}
	if result.Steps[0].Result == nil || result.Steps[0].Result.Click == nil {
		t.Fatal("Steps[0].Result.Click should not be nil")
	}
	if !result.Steps[0].Result.Click.Accepted {
		t.Error("Steps[0].Result.Click.Accepted should be true")
	}
	if result.Steps[1].Type != StepTypeOTPChallenge {
		t.Errorf("Steps[1].Type = %q, want %q", result.Steps[1].Type, StepTypeOTPChallenge)
	}
}

func TestIntegration_DocumentsUpload(t *testing.T) {
	f := loadFixture(t, "documents-upload")
	server := newIntegrationServer(t, f)
	defer server.Close()

	client := newIntegrationClient(t, server.URL)
	ctx := context.Background()

	result, err := client.Documents.Upload(ctx, "tx-uuid-001", &UploadDocumentRequest{
		Content:  "JVBERi0xLjQKMSAwIG9iago8PAovVHlwZQ==",
		Filename: "contract.pdf",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "DOCUMENT_UPLOADED" {
		t.Errorf("Status = %q, want %q", result.Status, "DOCUMENT_UPLOADED")
	}
	if result.DocumentHash == "" {
		t.Error("DocumentHash should not be empty")
	}
}

func TestIntegration_WebhooksRegister(t *testing.T) {
	f := loadFixture(t, "webhooks-register")
	server := newIntegrationServer(t, f)
	defer server.Close()

	client := newIntegrationClient(t, server.URL)
	ctx := context.Background()

	result, err := client.Webhooks.Register(ctx, &RegisterWebhookRequest{
		URL:    "https://example.com/webhooks/signdocs",
		Events: []WebhookEventType{WebhookEventTransactionCompleted, WebhookEventTransactionFailed},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.WebhookID != "wh-uuid-001" {
		t.Errorf("WebhookID = %q, want %q", result.WebhookID, "wh-uuid-001")
	}
	if result.Secret != "whsec_generated_secret_abc123" {
		t.Errorf("Secret = %q, want %q", result.Secret, "whsec_generated_secret_abc123")
	}
	if result.Status != "ACTIVE" {
		t.Errorf("Status = %q, want %q", result.Status, "ACTIVE")
	}
}

func TestIntegration_HealthCheck_NoAuth(t *testing.T) {
	f := loadFixture(t, "health-check")
	var gotAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.Response.Status)
		w.Write(f.Response.Body)
	}))
	defer server.Close()

	client := newIntegrationClient(t, server.URL)
	ctx := context.Background()

	result, err := client.Health.Check(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "healthy" {
		t.Errorf("Status = %q, want %q", result.Status, "healthy")
	}
	if result.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.0.0")
	}
	if result.Services["dynamodb"].Status != "healthy" {
		t.Errorf("Services[dynamodb].Status = %q, want %q", result.Services["dynamodb"].Status, "healthy")
	}
	if gotAuthHeader != "" {
		t.Errorf("Authorization header should be empty for health check, got %q", gotAuthHeader)
	}
}

func TestIntegration_VerificationVerify_NoAuth(t *testing.T) {
	f := loadFixture(t, "verification-verify")
	var gotAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.Response.Status)
		w.Write(f.Response.Body)
	}))
	defer server.Close()

	client := newIntegrationClient(t, server.URL)
	ctx := context.Background()

	result, err := client.Verification.Verify(ctx, "ev-uuid-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "COMPLETED" {
		t.Errorf("Status = %q, want %q", result.Status, "COMPLETED")
	}
	if result.EvidenceID != "ev-uuid-001" {
		t.Errorf("EvidenceID = %q, want %q", result.EvidenceID, "ev-uuid-001")
	}
	if gotAuthHeader != "" {
		t.Errorf("Authorization header should be empty for verification, got %q", gotAuthHeader)
	}
}

// Error path tests

func TestIntegration_Error400_BadRequest(t *testing.T) {
	f := loadFixture(t, "error-400")
	server := newIntegrationServer(t, f)
	defer server.Close()

	client := newIntegrationClient(t, server.URL)
	ctx := context.Background()

	_, err := client.Transactions.Create(ctx, &CreateTransactionRequest{
		Purpose: TransactionPurposeDocumentSignature,
		Policy:  Policy{Profile: "UNKNOWN"},
		Signer:  Signer{Name: "Test", UserExternalID: "u1"},
	})

	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Fatalf("expected BadRequestError, got %T: %v", err, err)
	}
	if badReq.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", badReq.StatusCode)
	}
	if !strings.Contains(badReq.ProblemDetail.Detail, "Invalid policy profile") {
		t.Errorf("Detail = %q, should contain 'Invalid policy profile'", badReq.ProblemDetail.Detail)
	}
}

func TestIntegration_Error404_NotFound(t *testing.T) {
	f := loadFixture(t, "error-404")
	server := newIntegrationServer(t, f)
	defer server.Close()

	client := newIntegrationClient(t, server.URL)
	ctx := context.Background()

	_, err := client.Transactions.Get(ctx, "tx-nonexistent")

	if !IsNotFound(err) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestIntegration_Error429_RateLimit(t *testing.T) {
	f := loadFixture(t, "error-429")
	server := newIntegrationServer(t, f)
	defer server.Close()

	client := newIntegrationClient(t, server.URL)
	ctx := context.Background()

	_, err := client.Transactions.Create(ctx, &CreateTransactionRequest{
		Purpose: TransactionPurposeDocumentSignature,
		Policy:  Policy{Profile: PolicyProfileClickOnly},
		Signer:  Signer{Name: "Test", UserExternalID: "u1"},
	})

	var rateErr *RateLimitError
	if !errors.As(err, &rateErr) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}
	if rateErr.RetryAfterSeconds != 5 {
		t.Errorf("RetryAfterSeconds = %d, want 5", rateErr.RetryAfterSeconds)
	}
}

func TestIntegration_Error409_Conflict(t *testing.T) {
	f := loadFixture(t, "error-409")
	server := newIntegrationServer(t, f)
	defer server.Close()

	client := newIntegrationClient(t, server.URL)
	ctx := context.Background()

	_, err := client.Transactions.Create(ctx, &CreateTransactionRequest{
		Purpose: TransactionPurposeDocumentSignature,
		Policy:  Policy{Profile: PolicyProfileClickOnly},
		Signer:  Signer{Name: "Test", UserExternalID: "u1"},
	})

	if !IsConflict(err) {
		t.Fatalf("expected ConflictError, got %T: %v", err, err)
	}
}
