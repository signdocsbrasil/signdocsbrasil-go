package signdocsbrasil

import (
	"context"
	"fmt"
	"net/http"
)

// EvidenceService provides access to transaction evidence retrieval.
type EvidenceService struct {
	http *httpClient
}

func newEvidenceService(h *httpClient) *EvidenceService {
	return &EvidenceService{http: h}
}

// Get retrieves the audit evidence for a completed transaction.
func (s *EvidenceService) Get(ctx context.Context, transactionID string) (*Evidence, error) {
	var result Evidence
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/transactions/%s/evidence", transactionID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
