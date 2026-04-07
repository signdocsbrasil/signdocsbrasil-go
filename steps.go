package signdocsbrasil

import (
	"context"
	"fmt"
	"net/http"
)

// StepsService provides access to verification step operations within a transaction.
type StepsService struct {
	http *httpClient
}

func newStepsService(h *httpClient) *StepsService {
	return &StepsService{http: h}
}

// List returns all steps for a transaction.
func (s *StepsService) List(ctx context.Context, transactionID string) (*StepListResponse, error) {
	var result StepListResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/transactions/%s/steps", transactionID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Start initiates a specific step in the transaction. The request body is optional;
// pass nil if no parameters are needed.
func (s *StepsService) Start(ctx context.Context, transactionID, stepID string, req *StartStepRequest) (*StartStepResponse, error) {
	var body any
	if req != nil {
		body = req
	} else {
		body = map[string]any{}
	}

	var result StartStepResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/transactions/%s/steps/%s/start", transactionID, stepID),
		Body:   body,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Complete completes a specific step. The request body depends on the step type.
// Pass the appropriate request type (CompleteClickRequest, CompleteOTPRequest, etc.)
// or nil/empty map if no body is needed.
func (s *StepsService) Complete(ctx context.Context, transactionID, stepID string, req any) (*StepCompleteResponse, error) {
	var body any
	if req != nil {
		body = req
	} else {
		body = map[string]any{}
	}

	var result StepCompleteResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/transactions/%s/steps/%s/complete", transactionID, stepID),
		Body:   body,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
