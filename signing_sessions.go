package signdocsbrasil

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// SigningSessionsService provides access to signing session operations.
type SigningSessionsService struct {
	http *httpClient
}

func newSigningSessionsService(h *httpClient) *SigningSessionsService {
	return &SigningSessionsService{http: h}
}

// Create creates a new signing session. An X-Idempotency-Key header is automatically
// included. Use WithIdempotencyKey to provide a specific key.
func (s *SigningSessionsService) Create(ctx context.Context, req *CreateSigningSessionRequest, opts ...CreateOption) (*SigningSession, error) {
	o := &createOptions{}
	for _, opt := range opts {
		opt(o)
	}

	var result SigningSession
	err := s.http.requestWithIdempotency(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   "/v1/signing-sessions",
		Body:   req,
	}, &result, o.idempotencyKey)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetStatus returns the current status of a signing session.
func (s *SigningSessionsService) GetStatus(ctx context.Context, sessionID string) (*SigningSessionStatus, error) {
	var result SigningSessionStatus
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/signing-sessions/%s/status", sessionID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Cancel cancels an active signing session.
func (s *SigningSessionsService) Cancel(ctx context.Context, sessionID string) (*CancelSigningSessionResponse, error) {
	var result CancelSigningSessionResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/signing-sessions/%s/cancel", sessionID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns a paginated list of signing sessions matching the filter parameters.
func (s *SigningSessionsService) List(ctx context.Context, params *SigningSessionListParams) (*SigningSessionListResponse, error) {
	query := make(map[string]string)
	if params != nil {
		if params.Status != "" {
			query["status"] = params.Status
		}
		if params.Limit > 0 {
			query["limit"] = strconv.Itoa(params.Limit)
		}
		if params.Cursor != "" {
			query["cursor"] = params.Cursor
		}
	}

	var result SigningSessionListResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   "/v1/signing-sessions",
		Query:  query,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns the full bootstrap data for a signing session.
// Used by the embedded signing widget to initialize the UI.
func (s *SigningSessionsService) Get(ctx context.Context, sessionID string) (*SigningSessionBootstrap, error) {
	var result SigningSessionBootstrap
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/signing-sessions/%s", sessionID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Advance advances a signing session through its steps.
// Supports actions: accept, verify_otp, resend_otp, start_liveness,
// complete_liveness, prepare_signing, complete_signing.
func (s *SigningSessionsService) Advance(ctx context.Context, sessionID string, req *AdvanceSessionRequest) (*AdvanceSessionResponse, error) {
	var result AdvanceSessionResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/signing-sessions/%s/advance", sessionID),
		Body:   req,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ResendOTP resends the OTP challenge for a signing session.
func (s *SigningSessionsService) ResendOTP(ctx context.Context, sessionID string) (*AdvanceSessionResponse, error) {
	var result AdvanceSessionResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/signing-sessions/%s/resend-otp", sessionID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// WaitForCompletion polls until the session reaches a terminal state.
// Returns the final status or an error if the timeout is exceeded.
func (s *SigningSessionsService) WaitForCompletion(ctx context.Context, sessionID string, opts ...WaitOption) (*SigningSessionStatus, error) {
	wo := &waitOptions{
		pollInterval: 3 * time.Second,
		timeout:      5 * time.Minute,
	}
	for _, opt := range opts {
		opt(wo)
	}

	deadline := time.Now().Add(wo.timeout)
	for {
		status, err := s.GetStatus(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		if status.Status != "ACTIVE" {
			return status, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("signing session %s did not complete within %v", sessionID, wo.timeout)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wo.pollInterval):
		}
	}
}

// WaitOption configures WaitForCompletion behavior.
type WaitOption func(*waitOptions)

type waitOptions struct {
	pollInterval time.Duration
	timeout      time.Duration
}

// WithPollInterval sets the polling interval for WaitForCompletion.
func WithPollInterval(d time.Duration) WaitOption {
	return func(o *waitOptions) {
		o.pollInterval = d
	}
}

// WithWaitTimeout sets the maximum wait time for WaitForCompletion.
func WithWaitTimeout(d time.Duration) WaitOption {
	return func(o *waitOptions) {
		o.timeout = d
	}
}
