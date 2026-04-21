package signdocsbrasil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// WebhooksService provides access to webhook registration and management.
type WebhooksService struct {
	http *httpClient
}

func newWebhooksService(h *httpClient) *WebhooksService {
	return &WebhooksService{http: h}
}

// Register creates a new webhook endpoint. Returns 201 Created with the webhook
// details including the signing secret.
func (s *WebhooksService) Register(ctx context.Context, req *RegisterWebhookRequest) (*RegisterWebhookResponse, error) {
	var result RegisterWebhookResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   "/v1/webhooks",
		Body:   req,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns all registered webhooks.
//
// The API responds with {"webhooks":[...],"count":N}; this method unwraps
// the envelope and returns the inner slice. A raw-RawMessage decode first
// lets us fall back to a bare-array shape for legacy test fixtures.
func (s *WebhooksService) List(ctx context.Context) ([]Webhook, error) {
	var raw json.RawMessage
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/webhooks",
	}, &raw)
	if err != nil {
		return nil, err
	}

	// Try envelope shape first.
	var envelope struct {
		Webhooks []Webhook `json:"webhooks"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && envelope.Webhooks != nil {
		return envelope.Webhooks, nil
	}

	// Fall back to bare array.
	var bare []Webhook
	if err := json.Unmarshal(raw, &bare); err != nil {
		return nil, fmt.Errorf("signdocsbrasil: decode webhook list: %w", err)
	}
	return bare, nil
}

// Delete removes a webhook endpoint. Returns 204 No Content on success.
func (s *WebhooksService) Delete(ctx context.Context, webhookID string) error {
	return s.http.request(ctx, requestOptions{
		Method: http.MethodDelete,
		Path:   fmt.Sprintf("/v1/webhooks/%s", webhookID),
	}, nil)
}

// Test sends a test delivery to a webhook endpoint.
func (s *WebhooksService) Test(ctx context.Context, webhookID string) (*WebhookTestResponse, error) {
	var result WebhookTestResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/webhooks/%s/test", webhookID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
