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

// AddSession adds a signer session to an envelope. An X-Idempotency-Key header
// is set automatically; pass WithIdempotencyKey for a specific one.
//
// Use a distinct key per signer. The API scopes its idempotency cache by key
// and resolved path, and every signer on an envelope shares that path, so one
// key across the loop returns signer 1's response — and signer 1's
// ClientSecret — for signer 2.
func (s *EnvelopesService) AddSession(ctx context.Context, envelopeID string, req *AddEnvelopeSessionRequest, opts ...CreateOption) (*EnvelopeSession, error) {
	o := &createOptions{}
	for _, opt := range opts {
		opt(o)
	}

	var result EnvelopeSession
	err := s.http.requestWithIdempotency(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/envelopes/%s/sessions", envelopeID),
		Body:   req,
	}, &result, o.idempotencyKey)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Cancel cancels an entire envelope.
//
// It transitions every non-terminal session and its transaction to CANCELLED
// and marks the envelope CANCELLED, killing the pending signing links.
// Signatures already collected are preserved and reported as
// PreservedSignedCount.
//
// Prefer this over cancelling each session individually: it is one call, it
// records the cancellation as a single auditable terminal event, and it is the
// only way to move the envelope's own status. Cancelling the member sessions
// one by one leaves the envelope itself ACTIVE.
//
// Idempotent: re-cancelling returns CancelledCount 0 and AlreadyCancelled true.
//
// reason is recorded in the audit trail; pass "" to let the API default it to
// "envelope_cancelled".
func (s *EnvelopesService) Cancel(ctx context.Context, envelopeID string, reason string) (*CancelEnvelopeResponse, error) {
	body := map[string]string{}
	if reason != "" {
		body["reason"] = reason
	}
	var result CancelEnvelopeResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/envelopes/%s/cancel", envelopeID),
		Body:   body,
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
