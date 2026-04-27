package signdocsbrasil

import (
	"encoding/json"
	"os"
	"testing"
)

// fixtureResponseBody loads a fixture JSON file and returns the "response.body"
// portion as raw JSON bytes. Fixtures follow the format:
//
//	{ "response": { "status": N, "body": { ... } } }
func fixtureResponseBody(t *testing.T, name string) []byte {
	t.Helper()

	path := "fixtures/" + name
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}

	var wrapper struct {
		Response struct {
			Body json.RawMessage `json:"body"`
		} `json:"response"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("failed to parse fixture wrapper %s: %v", name, err)
	}

	return wrapper.Response.Body
}

func TestModel_TransactionCreate(t *testing.T) {
	body := fixtureResponseBody(t, "transactions-create.json")

	var tx Transaction
	if err := json.Unmarshal(body, &tx); err != nil {
		t.Fatalf("failed to unmarshal Transaction: %v", err)
	}

	if tx.TenantID != "abc123" {
		t.Errorf("expected tenantId 'abc123', got %q", tx.TenantID)
	}
	if tx.TransactionID != "tx-uuid-001" {
		t.Errorf("expected transactionId 'tx-uuid-001', got %q", tx.TransactionID)
	}
	if tx.Status != TransactionStatusCreated {
		t.Errorf("expected status CREATED, got %q", tx.Status)
	}
	if tx.Purpose != TransactionPurposeDocumentSignature {
		t.Errorf("expected purpose DOCUMENT_SIGNATURE, got %q", tx.Purpose)
	}

	// Nested Policy
	if tx.Policy.Profile != PolicyProfileClickOnly {
		t.Errorf("expected policy profile CLICK_ONLY, got %q", tx.Policy.Profile)
	}

	// Nested Signer
	if tx.Signer.Name != "João Silva" {
		t.Errorf("expected signer name 'João Silva', got %q", tx.Signer.Name)
	}
	if tx.Signer.Email != "joao@example.com" {
		t.Errorf("expected signer email 'joao@example.com', got %q", tx.Signer.Email)
	}
	if tx.Signer.UserExternalID != "user-ext-001" {
		t.Errorf("expected signer userExternalId 'user-ext-001', got %q", tx.Signer.UserExternalID)
	}
	if tx.Signer.CPF != "12345678901" {
		t.Errorf("expected signer cpf '12345678901', got %q", tx.Signer.CPF)
	}

	// Steps
	if len(tx.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(tx.Steps))
	}
	step := tx.Steps[0]
	if step.StepID != "step-uuid-001" {
		t.Errorf("expected stepId 'step-uuid-001', got %q", step.StepID)
	}
	if step.Type != StepTypeClickAccept {
		t.Errorf("expected step type CLICK_ACCEPT, got %q", step.Type)
	}
	if step.Status != StepStatusPending {
		t.Errorf("expected step status PENDING, got %q", step.Status)
	}
	if step.Order != 1 {
		t.Errorf("expected step order 1, got %d", step.Order)
	}
	if step.MaxAttempts != 3 {
		t.Errorf("expected maxAttempts 3, got %d", step.MaxAttempts)
	}

	// Metadata
	if tx.Metadata == nil {
		t.Fatal("expected metadata to be non-nil")
	}
	if tx.Metadata["contractId"] != "CTR-2024-001" {
		t.Errorf("expected metadata contractId 'CTR-2024-001', got %q", tx.Metadata["contractId"])
	}

	// Timestamps
	if tx.CreatedAt != "2024-11-15T00:00:00.000Z" {
		t.Errorf("expected createdAt '2024-11-15T00:00:00.000Z', got %q", tx.CreatedAt)
	}
	if tx.ExpiresAt != "2024-11-16T00:00:00.000Z" {
		t.Errorf("expected expiresAt '2024-11-16T00:00:00.000Z', got %q", tx.ExpiresAt)
	}
}

func TestModel_TransactionGet(t *testing.T) {
	body := fixtureResponseBody(t, "transactions-get.json")

	var tx Transaction
	if err := json.Unmarshal(body, &tx); err != nil {
		t.Fatalf("failed to unmarshal Transaction: %v", err)
	}

	if tx.Status != TransactionStatusInProgress {
		t.Errorf("expected status IN_PROGRESS, got %q", tx.Status)
	}
	if tx.Policy.Profile != PolicyProfileClickPlusOTP {
		t.Errorf("expected policy profile CLICK_PLUS_OTP, got %q", tx.Policy.Profile)
	}

	// 2 steps
	if len(tx.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(tx.Steps))
	}

	// First step: completed with Click result
	first := tx.Steps[0]
	if first.Status != StepStatusCompleted {
		t.Errorf("expected first step status COMPLETED, got %q", first.Status)
	}
	if first.Result == nil {
		t.Fatal("expected first step result to be non-nil")
	}
	if first.Result.Click == nil {
		t.Fatal("expected first step Click result to be non-nil")
	}
	if !first.Result.Click.Accepted {
		t.Error("expected first step Click.Accepted to be true")
	}
	if first.Result.Click.TextVersion != "v1.0" {
		t.Errorf("expected textVersion 'v1.0', got %q", first.Result.Click.TextVersion)
	}
	if first.CompletedAt != "2024-11-15T00:01:00.000Z" {
		t.Errorf("expected completedAt for first step, got %q", first.CompletedAt)
	}

	// Second step: pending OTP
	second := tx.Steps[1]
	if second.Type != StepTypeOTPChallenge {
		t.Errorf("expected second step type OTP_CHALLENGE, got %q", second.Type)
	}
	if second.Status != StepStatusPending {
		t.Errorf("expected second step status PENDING, got %q", second.Status)
	}
}

func TestModel_TransactionList(t *testing.T) {
	body := fixtureResponseBody(t, "transactions-list.json")

	var resp TransactionListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal TransactionListResponse: %v", err)
	}

	// Pagination fields
	if resp.Count != 2 {
		t.Errorf("expected count 2, got %d", resp.Count)
	}
	if resp.NextToken == "" {
		t.Error("expected nextToken to be non-empty")
	}

	// 2 transactions
	if len(resp.Transactions) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(resp.Transactions))
	}

	tx1 := resp.Transactions[0]
	if tx1.TransactionID != "tx-uuid-002" {
		t.Errorf("expected first tx ID 'tx-uuid-002', got %q", tx1.TransactionID)
	}
	if tx1.Status != TransactionStatusCompleted {
		t.Errorf("expected first tx status COMPLETED, got %q", tx1.Status)
	}
	if tx1.Signer.Name != "Maria Santos" {
		t.Errorf("expected first tx signer 'Maria Santos', got %q", tx1.Signer.Name)
	}

	tx2 := resp.Transactions[1]
	if tx2.TransactionID != "tx-uuid-003" {
		t.Errorf("expected second tx ID 'tx-uuid-003', got %q", tx2.TransactionID)
	}
	if tx2.Policy.Profile != PolicyProfileBiometric {
		t.Errorf("expected second tx policy BIOMETRIC, got %q", tx2.Policy.Profile)
	}
}

func TestModel_Evidence(t *testing.T) {
	body := fixtureResponseBody(t, "evidence-get.json")

	var ev Evidence
	if err := json.Unmarshal(body, &ev); err != nil {
		t.Fatalf("failed to unmarshal Evidence: %v", err)
	}

	if ev.TenantID != "abc123" {
		t.Errorf("expected tenantId 'abc123', got %q", ev.TenantID)
	}
	if ev.TransactionID != "tx-uuid-001" {
		t.Errorf("expected transactionId 'tx-uuid-001', got %q", ev.TransactionID)
	}
	if ev.EvidenceID != "ev-uuid-001" {
		t.Errorf("expected evidenceId 'ev-uuid-001', got %q", ev.EvidenceID)
	}
	if ev.Status != "COMPLETED" {
		t.Errorf("expected status 'COMPLETED', got %q", ev.Status)
	}

	// Nested signer
	if ev.Signer.Name != "João Silva" {
		t.Errorf("expected signer name 'João Silva', got %q", ev.Signer.Name)
	}
	if ev.Signer.CPF != "12345678901" {
		t.Errorf("expected signer CPF '12345678901', got %q", ev.Signer.CPF)
	}
	if ev.Signer.UserExternalID != "user-ext-001" {
		t.Errorf("expected signer userExternalId 'user-ext-001', got %q", ev.Signer.UserExternalID)
	}

	// Steps
	if len(ev.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(ev.Steps))
	}
	step := ev.Steps[0]
	if step.Type != "CLICK_ACCEPT" {
		t.Errorf("expected step type 'CLICK_ACCEPT', got %q", step.Type)
	}
	if step.Status != "COMPLETED" {
		t.Errorf("expected step status 'COMPLETED', got %q", step.Status)
	}
	if step.CompletedAt != "2024-11-15T00:01:00.000Z" {
		t.Errorf("expected step completedAt, got %q", step.CompletedAt)
	}

	// Document
	if ev.Document == nil {
		t.Fatal("expected document to be non-nil")
	}
	if ev.Document.Hash != "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2" {
		t.Errorf("unexpected document hash: %q", ev.Document.Hash)
	}
	if ev.Document.Filename != "contract.pdf" {
		t.Errorf("expected document filename 'contract.pdf', got %q", ev.Document.Filename)
	}

	// Timestamps
	if ev.CreatedAt != "2024-11-15T00:00:00.000Z" {
		t.Errorf("unexpected createdAt: %q", ev.CreatedAt)
	}
	if ev.CompletedAt != "2024-11-15T00:01:00.000Z" {
		t.Errorf("unexpected completedAt: %q", ev.CompletedAt)
	}
}

func TestModel_WebhookTestResponse_RoundTrip(t *testing.T) {
	// Mirrors the live API shape per openapi.yaml WebhookTestResponse:
	// { "webhookId": "...", "testDelivery": { "httpStatus", "success", "error?", "timestamp" } }.
	const raw = `{
		"webhookId": "wh_abc123",
		"testDelivery": {
			"httpStatus": 200,
			"success": true,
			"timestamp": "2026-04-27T01:23:28.323Z"
		}
	}`

	var resp WebhookTestResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal WebhookTestResponse: %v", err)
	}

	if resp.WebhookID != "wh_abc123" {
		t.Errorf("expected webhookId 'wh_abc123', got %q", resp.WebhookID)
	}
	if resp.TestDelivery.HTTPStatus != 200 {
		t.Errorf("expected httpStatus 200, got %d", resp.TestDelivery.HTTPStatus)
	}
	if !resp.TestDelivery.Success {
		t.Error("expected success true")
	}
	if resp.TestDelivery.Timestamp != "2026-04-27T01:23:28.323Z" {
		t.Errorf("unexpected timestamp: %q", resp.TestDelivery.Timestamp)
	}
	if resp.TestDelivery.Error != "" {
		t.Errorf("expected empty error on success, got %q", resp.TestDelivery.Error)
	}

	// Round-trip should preserve the canonical shape (omitempty drops the empty error).
	out, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal WebhookTestResponse: %v", err)
	}
	const want = `{"webhookId":"wh_abc123","testDelivery":{"httpStatus":200,"success":true,"timestamp":"2026-04-27T01:23:28.323Z"}}`
	if string(out) != want {
		t.Errorf("round-trip mismatch:\n got  %s\n want %s", string(out), want)
	}

	// Failure case: error field is populated and serialized.
	const rawErr = `{
		"webhookId": "wh_xyz",
		"testDelivery": {
			"httpStatus": 502,
			"success": false,
			"error": "connection refused",
			"timestamp": "2026-04-27T01:24:00.000Z"
		}
	}`
	var failResp WebhookTestResponse
	if err := json.Unmarshal([]byte(rawErr), &failResp); err != nil {
		t.Fatalf("unmarshal failure WebhookTestResponse: %v", err)
	}
	if failResp.TestDelivery.Success {
		t.Error("expected success false")
	}
	if failResp.TestDelivery.Error != "connection refused" {
		t.Errorf("expected error 'connection refused', got %q", failResp.TestDelivery.Error)
	}
	if failResp.TestDelivery.HTTPStatus != 502 {
		t.Errorf("expected httpStatus 502, got %d", failResp.TestDelivery.HTTPStatus)
	}
}

func TestModel_ProblemDetail(t *testing.T) {
	path := "fixtures/error-400.json"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var wrapper struct {
		Response struct {
			Body ProblemDetail `json:"body"`
		} `json:"response"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("failed to parse fixture: %v", err)
	}

	pd := wrapper.Response.Body

	if pd.Type != "https://api.signdocs.com.br/errors/bad-request" {
		t.Errorf("expected type 'https://api.signdocs.com.br/errors/bad-request', got %q", pd.Type)
	}
	if pd.Title != "Bad Request" {
		t.Errorf("expected title 'Bad Request', got %q", pd.Title)
	}
	if pd.Status != 400 {
		t.Errorf("expected status 400, got %d", pd.Status)
	}
	if pd.Detail != "Invalid policy profile: UNKNOWN_PROFILE" {
		t.Errorf("expected detail about invalid policy, got %q", pd.Detail)
	}
	if pd.Instance != "/v1/transactions" {
		t.Errorf("expected instance '/v1/transactions', got %q", pd.Instance)
	}
}
