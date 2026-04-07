package signdocsbrasil

import (
	"errors"
	"testing"
)

func TestParseAPIError_StatusCodes(t *testing.T) {
	tests := []struct {
		status   int
		errType  string
	}{
		{400, "*signdocsbrasil.BadRequestError"},
		{401, "*signdocsbrasil.UnauthorizedError"},
		{403, "*signdocsbrasil.ForbiddenError"},
		{404, "*signdocsbrasil.NotFoundError"},
		{409, "*signdocsbrasil.ConflictError"},
		{422, "*signdocsbrasil.UnprocessableEntityError"},
		{429, "*signdocsbrasil.RateLimitError"},
		{500, "*signdocsbrasil.InternalServerError"},
		{503, "*signdocsbrasil.ServiceUnavailableError"},
	}

	for _, tt := range tests {
		t.Run(tt.errType, func(t *testing.T) {
			pd := ProblemDetail{
				Type:   "about:blank",
				Title:  "Error",
				Status: tt.status,
			}
			err := parseAPIError(tt.status, pd, 0)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestParseAPIError_UnknownStatus(t *testing.T) {
	pd := ProblemDetail{Type: "about:blank", Title: "Teapot", Status: 418}
	err := parseAPIError(418, pd, 0)

	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected *ApiError")
	}
	if apiErr.StatusCode != 418 {
		t.Errorf("expected 418, got %d", apiErr.StatusCode)
	}
}

func TestParseAPIError_RateLimitRetryAfter(t *testing.T) {
	pd := ProblemDetail{Type: "about:blank", Title: "Rate Limited", Status: 429}
	err := parseAPIError(429, pd, 5)

	var rlErr *RateLimitError
	if !errors.As(err, &rlErr) {
		t.Fatal("expected *RateLimitError")
	}
	if rlErr.RetryAfterSeconds != 5 {
		t.Errorf("expected RetryAfterSeconds 5, got %d", rlErr.RetryAfterSeconds)
	}
}

func TestParseAPIError_FillsEmptyType(t *testing.T) {
	pd := ProblemDetail{Title: "Error"}
	err := parseAPIError(400, pd, 0)

	var brErr *BadRequestError
	if !errors.As(err, &brErr) {
		t.Fatal("expected *BadRequestError")
	}
	if brErr.ProblemDetail.Type == "" {
		t.Error("expected type to be filled in")
	}
}

func TestIsNotFound(t *testing.T) {
	pd := ProblemDetail{Type: "about:blank", Title: "Not Found", Status: 404}
	err := parseAPIError(404, pd, 0)
	if !IsNotFound(err) {
		t.Error("expected IsNotFound to return true")
	}
	if IsNotFound(errors.New("other")) {
		t.Error("expected IsNotFound to return false for non-404")
	}
}

func TestIsRateLimit(t *testing.T) {
	pd := ProblemDetail{Type: "about:blank", Title: "Rate Limited", Status: 429}
	err := parseAPIError(429, pd, 0)
	if !IsRateLimit(err) {
		t.Error("expected IsRateLimit to return true")
	}
}

func TestIsConflict(t *testing.T) {
	pd := ProblemDetail{Type: "about:blank", Title: "Conflict", Status: 409}
	err := parseAPIError(409, pd, 0)
	if !IsConflict(err) {
		t.Error("expected IsConflict to return true")
	}
}

func TestApiError_Error(t *testing.T) {
	err := &ApiError{
		StatusCode:    400,
		ProblemDetail: ProblemDetail{Title: "Bad Request", Detail: "Invalid input"},
	}
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestApiError_ErrorWithoutDetail(t *testing.T) {
	err := &ApiError{
		StatusCode:    500,
		ProblemDetail: ProblemDetail{Title: "Internal Server Error"},
	}
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}
