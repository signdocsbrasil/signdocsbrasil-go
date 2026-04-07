package signdocsbrasil

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	maxTotalDuration = 60 * time.Second
	maxDelay         = 30 * time.Second
)

var retryableStatusCodes = map[int]bool{
	429: true,
	500: true,
	503: true,
}

// withRetry executes fn with exponential backoff retry logic for retryable
// status codes (429, 500, 503). It respects the Retry-After header and
// enforces a 60-second total duration budget. The context is checked between
// retries for cancellation.
func withRetry(ctx context.Context, maxRetries int, fn func() (*http.Response, error)) (*http.Response, error) {
	startTime := time.Now()

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if time.Since(startTime) > maxTotalDuration {
			return nil, &TimeoutError{Message: "request exceeded maximum retry duration of 60s"}
		}

		resp, err := fn()
		if err != nil {
			return nil, err
		}

		if !retryableStatusCodes[resp.StatusCode] {
			return resp, nil
		}

		// Last attempt - return the response as-is for the caller to handle the error
		if attempt == maxRetries {
			return resp, nil
		}

		delay := calculateDelay(attempt, resp.Header.Get("Retry-After"))

		select {
		case <-ctx.Done():
			resp.Body.Close()
			return nil, ctx.Err()
		case <-time.After(delay):
			resp.Body.Close()
		}
	}

	return nil, &TimeoutError{Message: "max retries exceeded"}
}

func calculateDelay(attempt int, retryAfterHeader string) time.Duration {
	if retryAfterHeader != "" {
		if seconds, err := strconv.Atoi(retryAfterHeader); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}

	// Exponential backoff: 2^attempt seconds + random jitter up to 1 second
	baseDelay := math.Pow(2, float64(attempt)) * float64(time.Second)
	jitter := rand.Float64() * float64(time.Second) //nolint:gosec // jitter does not need cryptographic randomness
	delay := time.Duration(baseDelay + jitter)

	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}
