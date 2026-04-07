package signdocsbrasil

import (
	"context"
	"fmt"
	"net/http"
)

// EnvelopesService provides access to envelope operations.
type EnvelopesService struct {
	http *httpClient
}

func newEnvelopesService(h *httpClient) *EnvelopesService {
	return &EnvelopesService{http: h}
}

// Create creates a new envelope. An X-Idempotency-Key header is automatically
// included. Use WithIdempotencyKey to provide a specific key.
func (s *EnvelopesService) Create(ctx context.Context, req *CreateEnvelopeRequest, opts ...CreateOption) (*Envelope, error) {
	o := &createOptions{}
	for _, opt := range opts {
		opt(o)
	}

	var result Envelope
	err := s.http.requestWithIdempotency(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   "/v1/envelopes",
		Body:   req,
	}, &result, o.idempotencyKey)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves an envelope by ID.
func (s *EnvelopesService) Get(ctx context.Context, envelopeID string) (*EnvelopeDetail, error) {
	var result EnvelopeDetail
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/envelopes/%s", envelopeID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// AddSession adds a signing session to an envelope.
func (s *EnvelopesService) AddSession(ctx context.Context, envelopeID string, req *AddEnvelopeSessionRequest) (*EnvelopeSession, error) {
	var result EnvelopeSession
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/envelopes/%s/sessions", envelopeID),
		Body:   req,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CombinedStamp retrieves the combined signed PDF for a completed envelope.
func (s *EnvelopesService) CombinedStamp(ctx context.Context, envelopeID string) (*EnvelopeCombinedStampResponse, error) {
	var result EnvelopeCombinedStampResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/envelopes/%s/combined-stamp", envelopeID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
