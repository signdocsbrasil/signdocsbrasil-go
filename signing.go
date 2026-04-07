package signdocsbrasil

import (
	"context"
	"fmt"
	"net/http"
)

// SigningService provides access to digital certificate signing operations.
type SigningService struct {
	http *httpClient
}

func newSigningService(h *httpClient) *SigningService {
	return &SigningService{http: h}
}

// Prepare begins a digital signing operation by providing the certificate chain.
// The API returns a hash that must be signed by the client's private key.
func (s *SigningService) Prepare(ctx context.Context, transactionID string, req *PrepareSigningRequest) (*PrepareSigningResponse, error) {
	var result PrepareSigningResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/transactions/%s/signing/prepare", transactionID),
		Body:   req,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Complete finalizes a digital signing operation by providing the raw signature
// produced by the client's private key over the hash from Prepare.
func (s *SigningService) Complete(ctx context.Context, transactionID string, req *CompleteSigningRequest) (*CompleteSigningResponse, error) {
	var result CompleteSigningResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/transactions/%s/signing/complete", transactionID),
		Body:   req,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
