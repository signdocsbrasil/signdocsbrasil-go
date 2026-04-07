// Package signdocsbrasil provides a Go client for the SignDocs Brasil API.
//
// The client supports both client_secret and private_key_jwt (ES256) authentication,
// automatic token caching and refresh, retry with exponential backoff, pagination,
// webhook signature verification, and idempotent transaction creation.
//
// # Quick Start
//
//	client, err := signdocsbrasil.NewClient("your-client-id",
//	    signdocsbrasil.WithClientSecret("your-client-secret"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	tx, err := client.Transactions.Create(ctx, &signdocsbrasil.CreateTransactionRequest{
//	    Purpose: signdocsbrasil.TransactionPurposeDocumentSignature,
//	    Policy:  signdocsbrasil.Policy{Profile: signdocsbrasil.PolicyProfileClickOnly},
//	    Signer:  signdocsbrasil.Signer{Name: "Jane Doe", UserExternalID: "user-123"},
//	})
package signdocsbrasil

// Client is the top-level entry point for the SignDocs Brasil API.
// It exposes service objects for each API resource.
type Client struct {
	// Health provides access to the /health endpoints.
	Health *HealthService

	// Transactions provides access to the /v1/transactions endpoints.
	Transactions *TransactionsService

	// Documents provides access to document upload, presign, confirm, and download.
	Documents *DocumentsService

	// Steps provides access to verification step operations.
	Steps *StepsService

	// Signing provides access to digital signing prepare and complete.
	Signing *SigningService

	// Evidence provides access to evidence retrieval.
	Evidence *EvidenceService

	// Verification provides access to public verification endpoints.
	Verification *VerificationService

	// Users provides access to user enrollment.
	Users *UsersService

	// Webhooks provides access to webhook management.
	Webhooks *WebhooksService

	// DocumentGroups provides access to document group operations.
	DocumentGroups *DocumentGroupsService

	// SigningSessions provides access to signing session operations.
	SigningSessions *SigningSessionsService

	// Envelopes provides access to multi-signer envelope operations.
	Envelopes *EnvelopesService
}

// NewClient creates a new SignDocs Brasil API client.
//
// The clientID is required. Authentication must be configured with either
// WithClientSecret or WithPrivateKey. Additional options can customize
// the base URL, timeout, retry count, scopes, and HTTP client.
//
// Example with client secret:
//
//	client, err := signdocsbrasil.NewClient("client-id",
//	    signdocsbrasil.WithClientSecret("client-secret"),
//	)
//
// Example with private key (ES256):
//
//	key, _ := signdocsbrasil.ParseES256PrivateKeyFromPEM(pemBytes)
//	client, err := signdocsbrasil.NewClient("client-id",
//	    signdocsbrasil.WithPrivateKey(key, "my-kid"),
//	)
func NewClient(clientID string, opts ...Option) (*Client, error) {
	cfg, err := resolveConfig(clientID, opts)
	if err != nil {
		return nil, err
	}

	auth := newAuthHandler(cfg)
	http := newHTTPClient(cfg, auth)

	return &Client{
		Health:         newHealthService(http),
		Transactions:   newTransactionsService(http),
		Documents:      newDocumentsService(http),
		Steps:          newStepsService(http),
		Signing:        newSigningService(http),
		Evidence:       newEvidenceService(http),
		Verification:   newVerificationService(http),
		Users:          newUsersService(http),
		Webhooks:       newWebhooksService(http),
		DocumentGroups:  newDocumentGroupsService(http),
		SigningSessions: newSigningSessionsService(http),
		Envelopes:       newEnvelopesService(http),
	}, nil
}
