# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.5.0] - 2026-04-27

### Added

- `EnvelopeID string` on `VerificationResponse` (`json:"envelopeId,omitempty"`) — populated when the verified evidence belongs to a multi-signer envelope. Use it with `client.Verification.VerifyEnvelope(ctx, envelopeID)` for cross-signer drill-down.
- Three new `WebhookEventType` constants:
  - `WebhookEventEnvelopeCreated` (`ENVELOPE.CREATED`)
  - `WebhookEventEnvelopeAllSigned` (`ENVELOPE.ALL_SIGNED`)
  - `WebhookEventEnvelopeExpired` (`ENVELOPE.EXPIRED`)

### Changed

- `User-Agent` bumped to `signdocs-brasil-go/1.5.0`.

## [1.4.1] - 2026-04-27

### Fixed

- `WebhookTestResponse` shape now matches the API. Was `{deliveryId, status, statusCode}`, now `{webhookId, testDelivery: {httpStatus, success, error?, timestamp}}` per `WebhookTestResponse` in `openapi.yaml`. The previous typed wrapper unmarshalled all-empty fields against the live HML API. Introduces a new `WebhookTestDelivery` struct for the nested object.

### Changed

- `User-Agent` bumped to `signdocs-brasil-go/1.4.1`.

## [1.4.0] - 2026-04-23

### Added

- `Owner` struct — optional requester identity (`Email`, `Name`) on `CreateSigningSessionRequest` and `CreateEnvelopeRequest`. When provided, SignDocs automatically emails each signer an invitation with their signing URL (when `Signer.Email` differs from `Owner.Email`, case-insensitive) and emails the owner a completion notification per signer completion (plus a final "all signed" message for envelopes). Leave `Owner` nil to keep the traditional behavior.
- `InviteSent bool` field on `SigningSession` and `EnvelopeSession` response structs. Populated by the API when an invitation email was dispatched.

### Changed

- `User-Agent` bumped to `signdocs-brasil-go/1.4.0`.

## [1.3.0] - 2026-04-20

### Fixed

- `WebhooksService.List` now correctly returns `[]Webhook`. Previously `json.Unmarshal` of `{"webhooks":[...],"count":N}` into `[]Webhook` failed with "cannot unmarshal object into Go value of type []Webhook". The method now decodes via an envelope shape with a bare-array fallback for test fixtures.

### Added

- `TokenCache` interface — pluggable OAuth token cache. Inject via the `WithTokenCache` functional option to share tokens across stateless workers (serverless, CLI). Default `NewInMemoryTokenCache()` preserves pre-1.3 single-process behavior.
- `CachedToken` struct and `NewInMemoryTokenCache()` constructor (thread-safe via `sync.Mutex`).
- `DeriveCacheKey(clientID, baseURL, scopes)` exported helper for custom cache implementations. Returns `signdocs.oauth.<32-hex>` SHA-256 derivative of canonical material (sorted scopes, trimmed trailing slash). Keys never leak the raw client ID.
- `ResponseMetadata` struct — captures `RateLimit-*`, `Deprecation`, `Sunset`, and `X-Request-Id` / `X-SignDocs-Request-Id` headers from every API response. `IsDeprecated()` helper. RFC 8594 parser accepts both `@<unix-seconds>` and IMF-fixdate forms.
- `WithOnResponse(fn func(*ResponseMetadata))` functional option — registers a response observer. Fires after every HTTP response (including errors). Panics in the callback are recovered and logged; they never reach the request path.
- `IsNT65Event(WebhookEventType) bool` exported predicate for identifying NT65 consignado events.

### Changed

- `authHandler` now reads and writes tokens through the configured `TokenCache`. Refresh is still serialized via `sync.Mutex` so a cold cache + bursty concurrency results in a single upstream token fetch.
- `authHandler.invalidate()` now deletes the cache entry instead of clearing internal fields.
- SDK now officially aligned with OpenAPI spec `WebhookEventType` enum at 17 events. Go was already ahead on `STEP.PURPOSE_DISCLOSURE_SENT` and `TRANSACTION.DEADLINE_APPROACHING` prior to 1.3.0; as of spec v1.1.0 these are part of the canonical set.
- User-Agent bumped to `1.3.0`.

## [1.2.0] - 2026-04-14

### Added

- `VerificationService.VerifyEnvelope(ctx, envelopeID)` — public method for the new `GET /v1/verify/envelope/{envelopeId}` endpoint. Returns envelope status, signers list (each with `EvidenceID` for drill-down via `Verify()`), and consolidated download URLs.
- `EnvelopeVerificationResponse`, `EnvelopeVerificationSigner`, and `EnvelopeVerificationDownloads` types. For non-PDF envelopes signed with digital certificates, `Downloads.ConsolidatedSignature` exposes a single PKCS#7 / CMS detached `.p7s` containing every signer's `SignerInfo`. For PDF envelopes, `Downloads.CombinedSignedPDF` exposes the merged PDF.
- `VerificationSigner.CPFCNPJ` and `VerificationResponse.TenantCNPJ` fields (previously returned by the API but not typed by the SDK).
- `VerificationDownloads.OriginalDocument` and `SignedSignature` fields (previously undocumented), matching the real shape the API returns.

### Changed

- `VerificationDownloads.SignedSignature` is now `nil` when the evidence belongs to a multi-signer envelope (the API omits the field). For standalone signing sessions (single-signer non-PDF with digital certificate) the field is still populated. To retrieve the consolidated `.p7s` for an envelope, use `VerificationService.VerifyEnvelope()` instead.

### Removed

- `VerificationDownloads.SignedPDF` — the field was typed by the SDK but never actually returned by the API. No real-world consumer could have depended on it.

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
