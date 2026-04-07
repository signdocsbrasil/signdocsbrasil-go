package signdocsbrasil

import (
	"fmt"
)

// ProblemDetail represents an RFC 7807 Problem Details response.
type ProblemDetail struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Status   int            `json:"status"`
	Detail   string         `json:"detail,omitempty"`
	Instance string         `json:"instance,omitempty"`
	Extra    map[string]any `json:"-"`
}

// SignDocsBrasilError is the interface implemented by all SDK errors.
type SignDocsBrasilError interface {
	error
	Unwrap() error
}

// ApiError represents an error response from the API with ProblemDetail.
type ApiError struct {
	StatusCode    int
	ProblemDetail ProblemDetail
}

func (e *ApiError) Error() string {
	if e.ProblemDetail.Detail != "" {
		return fmt.Sprintf("signdocsbrasil: API error %d: %s", e.StatusCode, e.ProblemDetail.Detail)
	}
	return fmt.Sprintf("signdocsbrasil: API error %d: %s", e.StatusCode, e.ProblemDetail.Title)
}

func (e *ApiError) Unwrap() error { return nil }

// BadRequestError represents a 400 response.
type BadRequestError struct{ *ApiError }

// UnauthorizedError represents a 401 response.
type UnauthorizedError struct{ *ApiError }

// ForbiddenError represents a 403 response.
type ForbiddenError struct{ *ApiError }

// NotFoundError represents a 404 response.
type NotFoundError struct{ *ApiError }

// ConflictError represents a 409 response.
type ConflictError struct{ *ApiError }

// UnprocessableEntityError represents a 422 response.
type UnprocessableEntityError struct{ *ApiError }

// RateLimitError represents a 429 response.
type RateLimitError struct {
	*ApiError
	RetryAfterSeconds int
}

// InternalServerError represents a 500 response.
type InternalServerError struct{ *ApiError }

// ServiceUnavailableError represents a 503 response.
type ServiceUnavailableError struct{ *ApiError }

// AuthenticationError represents a failure during token acquisition.
type AuthenticationError struct {
	Message string
	Err     error
}

func (e *AuthenticationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("signdocsbrasil: authentication error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("signdocsbrasil: authentication error: %s", e.Message)
}

func (e *AuthenticationError) Unwrap() error { return e.Err }

// ConnectionError represents a network-level failure.
type ConnectionError struct {
	Message string
	Err     error
}

func (e *ConnectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("signdocsbrasil: connection error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("signdocsbrasil: connection error: %s", e.Message)
}

func (e *ConnectionError) Unwrap() error { return e.Err }

// TimeoutError represents a request that exceeded the retry budget.
type TimeoutError struct {
	Message string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("signdocsbrasil: timeout: %s", e.Message)
}

func (e *TimeoutError) Unwrap() error { return nil }

// parseAPIError constructs the appropriate typed error from a status code and
// parsed ProblemDetail. If the response body is not valid ProblemDetail JSON,
// a synthetic one is constructed from the status code.
func parseAPIError(status int, pd ProblemDetail, retryAfter int) error {
	if pd.Type == "" {
		pd.Type = fmt.Sprintf("https://api.signdocs.com.br/errors/%d", status)
	}
	if pd.Title == "" {
		pd.Title = fmt.Sprintf("HTTP %d", status)
	}
	pd.Status = status

	base := &ApiError{StatusCode: status, ProblemDetail: pd}

	switch status {
	case 400:
		return &BadRequestError{base}
	case 401:
		return &UnauthorizedError{base}
	case 403:
		return &ForbiddenError{base}
	case 404:
		return &NotFoundError{base}
	case 409:
		return &ConflictError{base}
	case 422:
		return &UnprocessableEntityError{base}
	case 429:
		return &RateLimitError{ApiError: base, RetryAfterSeconds: retryAfter}
	case 500:
		return &InternalServerError{base}
	case 503:
		return &ServiceUnavailableError{base}
	default:
		return base
	}
}

// IsNotFound reports whether err is a 404 NotFoundError.
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// IsRateLimit reports whether err is a 429 RateLimitError.
func IsRateLimit(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

// IsConflict reports whether err is a 409 ConflictError.
func IsConflict(err error) bool {
	_, ok := err.(*ConflictError)
	return ok
}
