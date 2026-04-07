package signdocsbrasil

import (
	"context"
	"fmt"
	"net/http"
)

// UsersService provides access to user enrollment operations.
type UsersService struct {
	http *httpClient
}

func newUsersService(h *httpClient) *UsersService {
	return &UsersService{http: h}
}

// Enroll enrolls a user's biometric reference image for future biometric match steps.
func (s *UsersService) Enroll(ctx context.Context, userExternalID string, req *EnrollUserRequest) (*EnrollUserResponse, error) {
	var result EnrollUserResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPut,
		Path:   fmt.Sprintf("/v1/users/%s/enrollment", userExternalID),
		Body:   req,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
