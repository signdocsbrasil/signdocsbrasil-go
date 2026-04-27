package signdocsbrasil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// sdkVersion is reported in the User-Agent header.
const sdkVersion = "1.5.0"

// httpClient is the internal HTTP client used by all service methods.
type httpClient struct {
	baseURL    string
	maxRetries int
	auth       *authHandler
	client     *http.Client
	logger     *slog.Logger
	onResponse func(*ResponseMetadata)
}

type requestOptions struct {
	Method  string
	Path    string
	Body    any
	Query   map[string]string
	Headers map[string]string
	NoAuth  bool
}

func newHTTPClient(cfg *Config, auth *authHandler) *httpClient {
	return &httpClient{
		baseURL:    cfg.BaseURL,
		maxRetries: cfg.MaxRetries,
		auth:       auth,
		client:     cfg.HTTPClient,
		logger:     cfg.Logger,
		onResponse: cfg.OnResponse,
	}
}

// fireOnResponse invokes the configured observer, if any, wrapped in
// a panic-recovery boundary so that a panic inside user code cannot
// bubble into the request path. A recovered panic is logged via the
// configured slog.Logger, or the stdlib log package as a fallback.
func (h *httpClient) fireOnResponse(meta *ResponseMetadata) {
	if h.onResponse == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			if h.logger != nil {
				h.logger.Warn("onResponse callback panicked",
					"panic", fmt.Sprint(r),
					"method", meta.Method,
					"path", meta.Path,
					"status", meta.StatusCode,
				)
			} else {
				log.Printf("signdocsbrasil: onResponse callback panicked: %v (method=%s path=%s status=%d)",
					r, meta.Method, meta.Path, meta.StatusCode)
			}
		}
	}()
	h.onResponse(meta)
}

// request performs an HTTP request with authentication, retry, and error handling.
// The result is JSON-decoded into the value pointed to by result. If result is nil,
// the response body is discarded (useful for 204 responses).
func (h *httpClient) request(ctx context.Context, opts requestOptions, result any) error {
	resp, err := withRetry(ctx, h.maxRetries, func() (*http.Response, error) {
		reqURL, err := h.buildURL(opts.Path, opts.Query)
		if err != nil {
			return nil, &ConnectionError{Message: "failed to build URL", Err: err}
		}

		var bodyReader io.Reader
		if opts.Body != nil {
			bodyBytes, err := json.Marshal(opts.Body)
			if err != nil {
				return nil, fmt.Errorf("signdocsbrasil: failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, opts.Method, reqURL, bodyReader)
		if err != nil {
			return nil, &ConnectionError{Message: "failed to create request", Err: err}
		}

		req.Header.Set("User-Agent", "signdocs-brasil-go/"+sdkVersion)

		if !opts.NoAuth {
			token, err := h.auth.getAccessToken()
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", "Bearer "+token)
		}

		if opts.Body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}

		start := time.Now()
		resp, err := h.client.Do(req)
		if h.logger != nil {
			dur := time.Since(start)
			if resp != nil {
				if resp.StatusCode >= 400 {
					h.logger.Warn("request failed", "method", opts.Method, "path", opts.Path, "status", resp.StatusCode, "duration_ms", dur.Milliseconds())
				} else {
					h.logger.Info("request completed", "method", opts.Method, "path", opts.Path, "status", resp.StatusCode, "duration_ms", dur.Milliseconds())
				}
			} else {
				h.logger.Warn("request error", "method", opts.Method, "path", opts.Path, "error", err, "duration_ms", dur.Milliseconds())
			}
		}
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, &ConnectionError{Message: fmt.Sprintf("failed to connect to %s", reqURL), Err: err}
		}
		return resp, nil
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Fire the onResponse observer with the final response (after any
	// retry loop has converged). Retried responses are intentionally
	// not surfaced — consumers want the outcome, not the intermediate
	// 503s the retry layer already handled.
	h.fireOnResponse(ResponseMetadataFromResponse(resp, opts.Method, opts.Path))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ConnectionError{Message: "failed to read response body", Err: err}
	}

	// Handle 204 No Content
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		retryAfter := 0
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			retryAfter, _ = strconv.Atoi(ra)
		}

		var pd ProblemDetail
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "application/problem+json") {
			_ = json.Unmarshal(body, &pd)
		} else {
			pd.Detail = string(body)
		}

		return parseAPIError(resp.StatusCode, pd, retryAfter)
	}

	// Decode successful response
	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("signdocsbrasil: failed to decode response: %w", err)
		}
	}

	return nil
}

// requestWithIdempotency performs a request with an X-Idempotency-Key header.
// If idempotencyKey is empty, a new UUID is generated.
func (h *httpClient) requestWithIdempotency(ctx context.Context, opts requestOptions, result any, idempotencyKey string) error {
	if idempotencyKey == "" {
		var err error
		idempotencyKey, err = generateUUID()
		if err != nil {
			return fmt.Errorf("signdocsbrasil: failed to generate idempotency key: %w", err)
		}
	}

	if opts.Headers == nil {
		opts.Headers = make(map[string]string)
	}
	opts.Headers["X-Idempotency-Key"] = idempotencyKey

	return h.request(ctx, opts, result)
}

func (h *httpClient) buildURL(path string, query map[string]string) (string, error) {
	u, err := url.Parse(h.baseURL)
	if err != nil {
		return "", err
	}
	u.Path = path

	if len(query) > 0 {
		q := u.Query()
		for k, v := range query {
			if v != "" {
				q.Set(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}
