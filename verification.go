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
// (original document, evidence pack, final PDF, and — for standalone
// signing sessions only — the detached .p7s signature). This is a public
// endpoint and does not require authentication.
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

// VerifyEnvelope retrieves public verification data for a multi-signer
// envelope, including the list of signers (each with an EvidenceID for
// drill-down via Verify) and consolidated download URLs. For non-PDF
// envelopes signed with digital certificates, the consolidated .p7s
// containing every signer's SignerInfo is exposed via
// Downloads.ConsolidatedSignature. This is a public endpoint and does
// not require authentication.
func (s *VerificationService) VerifyEnvelope(ctx context.Context, envelopeID string) (*EnvelopeVerificationResponse, error) {
	var result EnvelopeVerificationResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/verify/envelope/%s", envelopeID),
		NoAuth: true,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
