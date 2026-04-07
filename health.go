package signdocsbrasil

import (
	"context"
	"net/http"
)

// HealthService provides access to the health check endpoints.
// These endpoints do not require authentication.
type HealthService struct {
	http *httpClient
}

func newHealthService(h *httpClient) *HealthService {
	return &HealthService{http: h}
}

// Check returns the current health status of the API.
func (s *HealthService) Check(ctx context.Context) (*HealthCheckResponse, error) {
	var result HealthCheckResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   "/health",
		NoAuth: true,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// History returns recent health check history.
func (s *HealthService) History(ctx context.Context) (*HealthHistoryResponse, error) {
	var result HealthHistoryResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   "/health/history",
		NoAuth: true,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
