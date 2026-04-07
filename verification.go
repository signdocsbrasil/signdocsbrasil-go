package signdocsbrasil

import (
	"context"
	"fmt"
	"net/http"
)

// VerificationService provides access to public evidence verification endpoints.
// These endpoints do not require authentication.
type VerificationService struct {
	http *httpClient
}

func newVerificationService(h *httpClient) *VerificationService {
	return &VerificationService{http: h}
}

// Verify checks whether an evidence record is valid. This is a public endpoint
// and does not require authentication.
func (s *VerificationService) Verify(ctx context.Context, evidenceID string) (*VerificationResponse, error) {
	var result VerificationResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/verify/%s", evidenceID),
		NoAuth: true,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Downloads retrieves presigned URLs for downloading evidence artifacts
// (evidence report and/or signed document). This is a public endpoint
// and does not require authentication.
func (s *VerificationService) Downloads(ctx context.Context, evidenceID string) (*VerificationDownloadsResponse, error) {
	var result VerificationDownloadsResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/verify/%s/downloads", evidenceID),
		NoAuth: true,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
