package signdocsbrasil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestWithRetry_ImmediateSuccess(t *testing.T) {
	calls := 0
	resp, err := withRetry(context.Background(), 3, func() (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	calls := 0
	resp, err := withRetry(context.Background(), 3, func() (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 400, Body: http.NoBody}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_Retries429ThenSucceeds(t *testing.T) {
	calls := 0
	resp, err := withRetry(context.Background(), 3, func() (*http.Response, error) {
		calls++
		if calls == 1 {
			header := http.Header{}
			header.Set("Retry-After", "0")
			return &http.Response{StatusCode: 429, Header: header, Body: http.NoBody}, nil
		}
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestWithRetry_Retries500(t *testing.T) {
	calls := 0
	resp, err := withRetry(context.Background(), 3, func() (*http.Response, error) {
		calls++
		if calls == 1 {
			return &http.Response{StatusCode: 500, Header: http.Header{}, Body: http.NoBody}, nil
		}
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestWithRetry_StopsAfterMaxRetries(t *testing.T) {
	calls := 0
	resp, err := withRetry(context.Background(), 2, func() (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 500, Header: http.Header{}, Body: http.NoBody}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
	if calls != 3 { // initial + 2 retries
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := int32(0)

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := withRetry(ctx, 10, func() (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{StatusCode: 503, Header: http.Header{}, Body: http.NoBody}, nil
	})
	if err == nil {
		t.Fatal("expected error from context cancellation")
	}
}

func TestCalculateDelay_WithRetryAfter(t *testing.T) {
	delay := calculateDelay(0, "3")
	if delay != 3*time.Second {
		t.Errorf("expected 3s, got %v", delay)
	}
}

func TestCalculateDelay_ExponentialBackoff(t *testing.T) {
	delay := calculateDelay(0, "")
	if delay < 1*time.Second || delay > 2*time.Second {
		t.Errorf("expected delay between 1s and 2s for attempt 0, got %v", delay)
	}

	delay = calculateDelay(1, "")
	if delay < 2*time.Second || delay > 3*time.Second {
		t.Errorf("expected delay between 2s and 3s for attempt 1, got %v", delay)
	}
}

func TestCalculateDelay_MaxCap(t *testing.T) {
	delay := calculateDelay(10, "") // 2^10 = 1024s, should be capped at 30s
	if delay > maxDelay {
		t.Errorf("expected delay <= %v, got %v", maxDelay, delay)
	}
}

func TestRetryableStatusCodes(t *testing.T) {
	retryable := []int{429, 500, 503}
	nonRetryable := []int{200, 201, 204, 400, 401, 403, 404, 409, 422}

	for _, code := range retryable {
		if !retryableStatusCodes[code] {
			t.Errorf("expected %d to be retryable", code)
		}
	}
	for _, code := range nonRetryable {
		if retryableStatusCodes[code] {
			t.Errorf("expected %d to NOT be retryable", code)
		}
	}
}

// Integration test using a real httptest server
func TestWithRetry_Integration(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := server.Client()
	resp, err := withRetry(context.Background(), 3, func() (*http.Response, error) {
		return client.Get(server.URL)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
