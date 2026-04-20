package signdocsbrasil

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ResponseMetadata captures response-level signals consumed for
// observability and lifecycle handling: IETF RateLimit-* headers, RFC
// 8594 Deprecation/Sunset, and the upstream request ID.
//
// Exposed via the WithOnResponse functional option. The SDK does not
// otherwise surface these headers to resource methods, so the callback
// is the single place to plug in observability.
//
// All header-derived fields are pointer types so "absent" is
// distinguishable from "present and zero".
type ResponseMetadata struct {
	// RateLimitLimit mirrors the IETF `RateLimit-Limit` header.
	RateLimitLimit *int
	// RateLimitRemaining mirrors the IETF `RateLimit-Remaining` header.
	RateLimitRemaining *int
	// RateLimitReset is the number of seconds from now until the
	// quota resets, mirroring the IETF `RateLimit-Reset` header.
	RateLimitReset *int
	// Deprecation is the parsed RFC 8594 Deprecation header.
	Deprecation *time.Time
	// Sunset is the parsed RFC 8594 Sunset header.
	Sunset *time.Time
	// RequestID is the upstream `X-Request-Id` (or
	// `X-SignDocs-Request-Id` fallback) value, when present.
	RequestID *string
	// StatusCode is the HTTP status code.
	StatusCode int
	// Method is the uppercased HTTP method.
	Method string
	// Path is the request path (with query string if any).
	Path string
}

// IsDeprecated reports whether the endpoint returned a Deprecation
// header that we could parse.
func (m *ResponseMetadata) IsDeprecated() bool {
	return m != nil && m.Deprecation != nil
}

// ResponseMetadataFromResponse extracts observability metadata from
// resp. method and path are passed through unchanged (aside from
// method being uppercased) because neither is recoverable from the
// response alone.
//
// resp may be nil — the returned ResponseMetadata then has StatusCode
// zero and all header fields nil. This matches the SDK behavior when
// the underlying transport never produced a response (e.g. a network
// error before receiving headers).
func ResponseMetadataFromResponse(resp *http.Response, method, path string) *ResponseMetadata {
	m := &ResponseMetadata{
		Method: strings.ToUpper(method),
		Path:   path,
	}
	if resp == nil {
		return m
	}

	m.StatusCode = resp.StatusCode
	m.RateLimitLimit = intHeader(resp, "RateLimit-Limit")
	m.RateLimitRemaining = intHeader(resp, "RateLimit-Remaining")
	m.RateLimitReset = intHeader(resp, "RateLimit-Reset")
	m.Deprecation = rfc8594Date(resp.Header.Get("Deprecation"))
	m.Sunset = rfc8594Date(resp.Header.Get("Sunset"))
	m.RequestID = firstHeader(resp, "X-Request-Id", "X-SignDocs-Request-Id")

	return m
}

var intHeaderPattern = regexp.MustCompile(`^-?\d+$`)

func intHeader(resp *http.Response, name string) *int {
	raw := resp.Header.Get(name)
	if raw == "" {
		return nil
	}
	if !intHeaderPattern.MatchString(raw) {
		return nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &n
}

func firstHeader(resp *http.Response, names ...string) *string {
	for _, name := range names {
		if v := resp.Header.Get(name); v != "" {
			return &v
		}
	}
	return nil
}

var unixTimestampPattern = regexp.MustCompile(`^@(-?\d+)$`)

// rfc8594Date parses an RFC 8594 Deprecation/Sunset header value.
// Accepts the two forms permitted by the spec:
//
//   - `@<unix-seconds>` (e.g. `@1711929600`)
//   - IMF-fixdate / HTTP-date (e.g. `Sun, 01 Sep 2026 00:00:00 GMT`)
//
// Returns nil for any unparseable or empty input.
func rfc8594Date(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	if m := unixTimestampPattern.FindStringSubmatch(raw); m != nil {
		secs, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return nil
		}
		t := time.Unix(secs, 0).UTC()
		return &t
	}

	// http.ParseTime handles IMF-fixdate (RFC 7231), RFC 850, and
	// asctime — the three HTTP-date formats historically accepted.
	if t, err := http.ParseTime(raw); err == nil {
		return &t
	}
	return nil
}
