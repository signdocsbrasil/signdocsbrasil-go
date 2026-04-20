package signdocsbrasil

import (
	"net/http"
	"testing"
	"time"
)

func makeResp(status int, headers map[string]string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{StatusCode: status, Header: h}
}

func TestResponseMetadataFromResponse_RateLimit(t *testing.T) {
	resp := makeResp(200, map[string]string{
		"RateLimit-Limit":     "1000",
		"RateLimit-Remaining": "942",
		"RateLimit-Reset":     "45",
	})

	m := ResponseMetadataFromResponse(resp, "get", "/v1/transactions")

	if m.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", m.StatusCode)
	}
	if m.Method != "GET" {
		t.Errorf("expected method 'GET', got %q", m.Method)
	}
	if m.Path != "/v1/transactions" {
		t.Errorf("expected path '/v1/transactions', got %q", m.Path)
	}
	if m.RateLimitLimit == nil || *m.RateLimitLimit != 1000 {
		t.Errorf("expected RateLimitLimit=1000, got %v", m.RateLimitLimit)
	}
	if m.RateLimitRemaining == nil || *m.RateLimitRemaining != 942 {
		t.Errorf("expected RateLimitRemaining=942, got %v", m.RateLimitRemaining)
	}
	if m.RateLimitReset == nil || *m.RateLimitReset != 45 {
		t.Errorf("expected RateLimitReset=45, got %v", m.RateLimitReset)
	}
}

func TestResponseMetadataFromResponse_Deprecation_UnixTimestamp(t *testing.T) {
	// RFC 8594 §2: `@<unix-seconds>` form.
	resp := makeResp(200, map[string]string{
		"Deprecation": "@1711929600", // 2024-04-01 00:00:00 UTC
		"Sunset":      "@1743465600", // 2025-04-01 00:00:00 UTC
	})

	m := ResponseMetadataFromResponse(resp, "GET", "/v1/x")

	if m.Deprecation == nil {
		t.Fatal("expected Deprecation to be parsed")
	}
	if m.Deprecation.Unix() != 1711929600 {
		t.Errorf("expected deprecation unix=1711929600, got %d", m.Deprecation.Unix())
	}
	if m.Sunset == nil {
		t.Fatal("expected Sunset to be parsed")
	}
	if m.Sunset.Unix() != 1743465600 {
		t.Errorf("expected sunset unix=1743465600, got %d", m.Sunset.Unix())
	}
	if !m.IsDeprecated() {
		t.Error("expected IsDeprecated() to return true")
	}
}

func TestResponseMetadataFromResponse_Deprecation_IMFFixdate(t *testing.T) {
	resp := makeResp(200, map[string]string{
		"Deprecation": "Sun, 01 Sep 2026 00:00:00 GMT",
	})

	m := ResponseMetadataFromResponse(resp, "GET", "/v1/x")

	if m.Deprecation == nil {
		t.Fatal("expected Deprecation to be parsed from IMF-fixdate")
	}
	want := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	if !m.Deprecation.Equal(want) {
		t.Errorf("expected %v, got %v", want, m.Deprecation)
	}
}

func TestResponseMetadataFromResponse_Deprecation_Unparseable(t *testing.T) {
	resp := makeResp(200, map[string]string{
		"Deprecation": "not-a-valid-date",
	})

	m := ResponseMetadataFromResponse(resp, "GET", "/v1/x")

	if m.Deprecation != nil {
		t.Errorf("expected nil Deprecation for unparseable input, got %v", m.Deprecation)
	}
	if m.IsDeprecated() {
		t.Error("IsDeprecated() should be false when parse failed")
	}
}

func TestResponseMetadataFromResponse_RequestID_Primary(t *testing.T) {
	resp := makeResp(200, map[string]string{
		"X-Request-Id": "req_abc123",
	})

	m := ResponseMetadataFromResponse(resp, "GET", "/v1/x")

	if m.RequestID == nil || *m.RequestID != "req_abc123" {
		t.Errorf("expected RequestID 'req_abc123', got %v", m.RequestID)
	}
}

func TestResponseMetadataFromResponse_RequestID_Fallback(t *testing.T) {
	// Only the fallback header is set — should still pick it up.
	resp := makeResp(200, map[string]string{
		"X-SignDocs-Request-Id": "sdb_xyz789",
	})

	m := ResponseMetadataFromResponse(resp, "GET", "/v1/x")

	if m.RequestID == nil || *m.RequestID != "sdb_xyz789" {
		t.Errorf("expected RequestID 'sdb_xyz789' (fallback), got %v", m.RequestID)
	}
}

func TestResponseMetadataFromResponse_RequestID_PrimaryWins(t *testing.T) {
	resp := makeResp(200, map[string]string{
		"X-Request-Id":          "primary",
		"X-SignDocs-Request-Id": "fallback",
	})

	m := ResponseMetadataFromResponse(resp, "GET", "/v1/x")

	if m.RequestID == nil || *m.RequestID != "primary" {
		t.Errorf("expected primary to win, got %v", m.RequestID)
	}
}

func TestResponseMetadataFromResponse_RequestID_Missing(t *testing.T) {
	resp := makeResp(200, map[string]string{})

	m := ResponseMetadataFromResponse(resp, "GET", "/v1/x")

	if m.RequestID != nil {
		t.Errorf("expected nil RequestID, got %q", *m.RequestID)
	}
}

func TestResponseMetadataFromResponse_NoHeaders(t *testing.T) {
	resp := makeResp(204, map[string]string{})

	m := ResponseMetadataFromResponse(resp, "delete", "/v1/webhooks/abc")

	if m.StatusCode != 204 {
		t.Errorf("expected 204, got %d", m.StatusCode)
	}
	if m.Method != "DELETE" {
		t.Errorf("expected uppercased method 'DELETE', got %q", m.Method)
	}
	if m.RateLimitLimit != nil || m.RateLimitRemaining != nil || m.RateLimitReset != nil {
		t.Error("expected all rate-limit fields nil on empty headers")
	}
	if m.Deprecation != nil || m.Sunset != nil || m.RequestID != nil {
		t.Error("expected deprecation/sunset/requestID nil on empty headers")
	}
}

func TestResponseMetadataFromResponse_NonNumericRateLimit(t *testing.T) {
	resp := makeResp(200, map[string]string{
		"RateLimit-Limit": "not-a-number",
	})

	m := ResponseMetadataFromResponse(resp, "GET", "/v1/x")

	if m.RateLimitLimit != nil {
		t.Errorf("expected nil for non-numeric rate limit, got %d", *m.RateLimitLimit)
	}
}

func TestResponseMetadataFromResponse_NilResponse(t *testing.T) {
	m := ResponseMetadataFromResponse(nil, "GET", "/v1/x")

	if m == nil {
		t.Fatal("expected non-nil ResponseMetadata even for nil response")
	}
	if m.StatusCode != 0 {
		t.Errorf("expected StatusCode 0, got %d", m.StatusCode)
	}
	if m.Method != "GET" {
		t.Errorf("expected method preserved, got %q", m.Method)
	}
}
