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

// GetEnrollment reads whether a user is enrolled and, crucially, until when.
//
// Use it to sweep your user base and re-enrol before Expired flips. Nothing
// warns you on its own beyond the ENROLLMENT.EXPIRING webhook, and once the
// grace window closes this returns a not-found error rather than reporting an
// expired enrolment.
func (s *UsersService) GetEnrollment(ctx context.Context, userExternalID string) (*EnrollmentStatusResponse, error) {
	var result EnrollmentStatusResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/users/%s/enrollment", userExternalID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteEnrollment erases a user's biometric enrolment (LGPD art. 18).
//
// Destroys every stored version of the reference image, not just the current
// one, and removes the record. Irreversible.
func (s *UsersService) DeleteEnrollment(ctx context.Context, userExternalID string) (*DeleteEnrollmentResponse, error) {
	var result DeleteEnrollmentResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodDelete,
		Path:   fmt.Sprintf("/v1/users/%s/enrollment", userExternalID),
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// EnrollBatch enrols up to 25 users in one request.
//
// The documented cap is 25 rows, but the binding limit is the request body —
// roughly 6MB, and base64 inflates each photo by a third. Keep photos under
// ~175KB (640x640 is ample) to use all 25 slots.
//
// Set DryRun to inspect the photos without storing anything.
func (s *UsersService) EnrollBatch(ctx context.Context, req *EnrollUsersBatchRequest) (*EnrollUsersBatchResponse, error) {
	var result EnrollUsersBatchResponse
	err := s.http.request(ctx, requestOptions{
		Method: http.MethodPost,
		Path:   "/v1/users/enrollments",
		Body:   req,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
