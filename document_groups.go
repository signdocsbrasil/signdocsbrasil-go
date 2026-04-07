package signdocsbrasil

import (
	"context"
	"fmt"
	"net/http"
)

// DocumentGroupsService provides access to document group operations.
type DocumentGroupsService struct {
	http *httpClient
}

func newDocumentGroupsService(h *httpClient) *DocumentGroupsService {
	return &DocumentGroupsService{http: h}
}

// CombinedStamp requests a combined stamp document for a document group.
// This merges all signed documents in the group into a single stamped output.
func (s *DocumentGroupsService) CombinedStamp(ctx context.Context, documentGroupID string) (*CombinedStampResponse, error) {
	var result CombinedStampResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1/document-groups/%s/combined-stamp", documentGroupID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
