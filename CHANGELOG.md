# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-03-27

### Added

- Envelopes service (`client.Envelopes`): Create, Get, AddSession, CombinedStamp — multi-signer workflows with parallel or sequential signing
- New types: CreateEnvelopeRequest, Envelope, AddEnvelopeSessionRequest, EnvelopeSession, EnvelopeSessionSummary, EnvelopeDetail, EnvelopeCombinedStampResponse

### Fixed

- Removed duplicate `ActionMetadata` type declaration in signing_sessions_types.go (was already defined in models.go)
- Renamed `WithTimeout` to `WithWaitTimeout` in WaitForCompletion options to avoid conflict with client config `WithTimeout`

## [1.0.0] - 2026-03-02

### Added

- Full API coverage: Transactions, Documents, Steps, Signing, Evidence, Verification, Users, Webhooks, DocumentGroups, Health
- OAuth2 `client_credentials` authentication with client secret
- Private Key JWT (ES256) authentication with `client_assertion`
- Automatic token caching with 30-second refresh buffer
- Thread-safe token refresh via `sync.Mutex`
- Auto-pagination via `ListAutoPaginate()` with generic `PageIterator[T]`
- Custom HTTP client injection via `WithHTTPClient(*http.Client)`
- Per-request timeout via `context.Context` deadlines
- Exponential backoff retry with jitter (429, 500, 503)
- Retry-After header support
- Idempotency keys (auto-generated UUID) on POST requests
- Typed errors for all HTTP error codes (RFC 7807 Problem Details)
- Helper functions: `IsNotFound()`, `IsRateLimit()`, `IsConflict()`
- Webhook signature verification (HMAC-SHA256, constant-time comparison)
- Configurable base URL, timeout, max retries, and scopes
- Functional options pattern for client configuration
- Zero external dependencies (Go standard library only)
- Go 1.21+ support
