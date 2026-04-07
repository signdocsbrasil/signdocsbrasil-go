package signdocsbrasil

import (
	"context"
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
func (s *WebhooksService) List(ctx context.Context) ([]Webhook, error) {
	var result []Webhook
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/webhooks",
	}, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
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
