# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.10.0] - 2026-08-20

### Added

- **`SigningSessions.Link(ctx, sessionID)`** — `POST /v1/signing-sessions/{sessionId}/link`. The endpoint has
  been in the API and documented in the OpenAPI spec all along, but no SDK in any
  language exposed it, so there was no supported way to recover a signing link
  once the create response was gone.
  - A signing link is single-use: after the signer finishes — or the embed token
    is otherwise consumed — reopening the same URL returns
    `401 Embed token has been consumed`. This mints a new one **without creating
    another transaction and without consuming quota**.
  - Works for standalone and envelope sessions alike.
  - The session must be `ACTIVE`. A completed or cancelled one returns 409: a
    link to a finished session would authenticate nothing. Reach the signed
    document through the envelope's combined stamp or the transaction download
    instead.
  - `expiresAt` is inherited from the original session and is **not** extended.
  - Sends no idempotency key, deliberately. A retry must mint a fresh URL, not
    replay one that has already been consumed.
  - **Authorises the tenant, not the end user.** The API cannot tell which of
    your users is entitled to a given link, so an application whose users share
    one tenant has to establish that itself before calling — otherwise this is a
    way for one user to obtain another's signing credential.
- `MintSigningLinkResponse` model (`SessionID`, `TransactionID`, `URL`, `ExpiresAt`, `ExpiresIn`).

### Fixed

- **The `User-Agent` reported a version nobody was running.** v1.9.0 shipped
  announcing itself as `signdocs-brasil-go/1.8.0`. The constant now moves with
  the release. Unlike the other SDKs there is no in-repo manifest to compare it
  against — the module version lives in the git tag — so this one is guarded by
  the release checklist rather than by a test.

## [1.9.0] - 2026-08-20

### Fixed

- **`addSession`/`verifyDocument` sent no idempotency key** while the client
  retries 429/500/503, so a 500 on an add-session became a second signer, a
  second quota charge and a second invitation, and a retried `verifyDocument`
  paid the metered verification quota twice for an identical result. Pass a
  distinct key per signer: the API scopes its cache by key and resolved path,
  and all signers on an envelope share that path.

## [1.8.0] - 2026-07-30

### Added

- **Envelope cancellation** — `POST /v1/envelopes/{envelopeId}/cancel` has existed since envelopes shipped and is what the Telegram bot calls, but no SDK exposed it. Consumers were left cancelling each member session by hand, which is not the same operation: it leaves the envelope's own status ACTIVE (verified against HML — an envelope whose sessions are every one CANCELLED still reports ACTIVE), costs a call per signer, and records N separate cancellations instead of one auditable terminal event.
  - Transitions every non-terminal session and its transaction to CANCELLED, then marks the envelope CANCELLED.
  - Signatures already collected are preserved and reported as `preservedSignedCount` — cancelling stops the pending signers, it never invalidates evidence already gathered.
  - Idempotent: re-cancelling returns `cancelledCount` 0 and `alreadyCancelled` true.
  - Optional `reason` is recorded in the audit trail; the API defaults it to `envelope_cancelled`.
  - Shipped in lockstep with signdocs-brasil-php 1.9.0.

### Changed

- `User-Agent` bumped to `signdocsbrasil-go/1.8.0`.

## [1.7.0] - 2026-07-29

### Added

- **`signatureUrl` and `documentFormat` on the download response.** `GET /v1/transactions/{id}/download` has always returned these for non-PDF transactions, but the model parsed only `originalUrl` / `signedUrl` and silently dropped them — so there was no way to reach a detached CAdES signature through the SDK at all. Verified against HML: the API returns six fields where the model exposed four.
  - `documentFormat` is `'pdf'` or `'generic'`, derived by the API from the uploaded bytes (not the filename).
  - `signatureUrl` is the presigned URL for the detached `.p7s`, returned **instead of** `signedUrl` when `documentFormat` is `'generic'` — a non-PDF cannot carry an embedded signature.
  - Caveat worth knowing when consuming it: the API presigns that S3 key without checking that the object exists, so a non-PDF signed under a click/OTP policy still comes back with a `signatureUrl` — one that 404s on GET, because only the digital-certificate step writes a `.p7s`. Branch on the signing policy, not on the field being set.
  - Shipped in lockstep with signdocs-brasil-php 1.8.0.

## [1.6.1] - 2026-06-25

### Changed

- API-documentation link in README now points to https://docs.signdocs.com.br (was a dead relative path).

## [1.6.0] - 2026-06-19

### Added

- `VerificationService.VerifyDocument(ctx, *VerifyDocumentRequest)` — POSTs a base64-encoded PDF to `/v1/verify/document` and reports detected signatures. Unlike the other `VerificationService` methods, this endpoint is authenticated: it requires a Bearer token with the `verification:write` scope and is restricted to production credentials at runtime. New `VerifyDocumentRequest` (`Content`, `Filename`), `VerifyDocumentResponse` (`Signed`, `SignatureCount`, `Signatures`, `CheckedAt`), and `DetectedSignature` (`Method`, `Type`, `SubFilter`, `Filter`, `Confidence`) structs. `Type` is one of `pades`, `pkcs7`, `legacy`, or `digital_certificate`.
- `OtpChannelSelectable bool` on `Signer` (`json:"otpChannelSelectable,omitempty"`) — lets a signer choose their OTP delivery channel instead of being locked to `OtpChannel`.
- `OtpChannelSelectable bool` and `AvailableOtpChannels []OtpChannel` on `BootstrapSigner` (`json:"otpChannelSelectable,omitempty"` / `json:"availableOtpChannels,omitempty"`) — surface the selectable channels in the bootstrap response.
- `OtpChannel OtpChannel` on `AdvanceSessionRequest` (`json:"otpChannel,omitempty"`) — select the channel when verifying/advancing an OTP step.
- New `ResendOtpRequest` struct (`Channel OtpChannel`, `json:"channel,omitempty"`) for choosing the resend delivery channel.

### Changed

- `SigningSessionsService.ResendOTP` now accepts a `*ResendOtpRequest` argument (pass `nil` to resend over the session's default channel) and POSTs it as the request body. This is a signature change.
- `User-Agent` bumped to `signdocs-brasil-go/1.6.0`.

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
