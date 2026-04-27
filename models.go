package signdocsbrasil

// TransactionStatus represents the lifecycle state of a transaction.
type TransactionStatus string

const (
	TransactionStatusCreated          TransactionStatus = "CREATED"
	TransactionStatusDocumentUploaded TransactionStatus = "DOCUMENT_UPLOADED"
	TransactionStatusInProgress       TransactionStatus = "IN_PROGRESS"
	TransactionStatusCompleted        TransactionStatus = "COMPLETED"
	TransactionStatusCancelled        TransactionStatus = "CANCELLED"
	TransactionStatusExpired          TransactionStatus = "EXPIRED"
	TransactionStatusFailed           TransactionStatus = "FAILED"
)

// StepType represents the type of verification step.
type StepType string

const (
	StepTypeClickAccept        StepType = "CLICK_ACCEPT"
	StepTypeOTPChallenge       StepType = "OTP_CHALLENGE"
	StepTypeOTPVerify          StepType = "OTP_VERIFY"
	StepTypeBiometricLive      StepType = "BIOMETRIC_LIVENESS"
	StepTypeBiometricMatch     StepType = "BIOMETRIC_MATCH"
	StepTypeDigitalSignA1      StepType = "DIGITAL_SIGN_A1"
	StepTypeSerproIdentity     StepType = "SERPRO_IDENTITY_CHECK"
	StepTypeDocumentPhotoMatch StepType = "DOCUMENT_PHOTO_MATCH"
	StepTypePurposeDisclosure  StepType = "PURPOSE_DISCLOSURE"
)

// StepStatus represents the lifecycle state of a step.
type StepStatus string

const (
	StepStatusPending   StepStatus = "PENDING"
	StepStatusStarted   StepStatus = "STARTED"
	StepStatusCompleted StepStatus = "COMPLETED"
	StepStatusFailed    StepStatus = "FAILED"
)

// PolicyProfile represents a predefined verification policy.
type PolicyProfile string

const (
	PolicyProfileClickOnly              PolicyProfile = "CLICK_ONLY"
	PolicyProfileClickPlusOTP           PolicyProfile = "CLICK_PLUS_OTP"
	PolicyProfileBiometric              PolicyProfile = "BIOMETRIC"
	PolicyProfileBiometricPlusOTP       PolicyProfile = "BIOMETRIC_PLUS_OTP"
	PolicyProfileDigitalCertificate     PolicyProfile = "DIGITAL_CERTIFICATE"
	PolicyProfileBiometricSerpro        PolicyProfile = "BIOMETRIC_SERPRO"
	PolicyProfileBiometricDocFallback   PolicyProfile = "BIOMETRIC_DOCUMENT_FALLBACK"
	PolicyProfileCustom                 PolicyProfile = "CUSTOM"
)

// TransactionPurpose represents the purpose of a transaction.
type TransactionPurpose string

const (
	TransactionPurposeDocumentSignature    TransactionPurpose = "DOCUMENT_SIGNATURE"
	TransactionPurposeActionAuthentication TransactionPurpose = "ACTION_AUTHENTICATION"
)

// CaptureMode represents how biometric capture is performed.
type CaptureMode string

const (
	CaptureModeBankApp    CaptureMode = "BANK_APP"
	CaptureModeHostedPage CaptureMode = "HOSTED_PAGE"
)

// WebhookEventType represents types of webhook events. The canonical
// set of events is defined by the OpenAPI spec `WebhookEventType` enum
// at `openapi/openapi.yaml`; the SDK stays in lockstep. Events tagged
// NT65 are only emitted for tenants with `nt65ComplianceEnabled` (INSS
// consignado flow) — use IsNT65Event to check.
type WebhookEventType string

const (
	// Transaction events
	WebhookEventTransactionCreated             WebhookEventType = "TRANSACTION.CREATED"
	WebhookEventTransactionCompleted           WebhookEventType = "TRANSACTION.COMPLETED"
	WebhookEventTransactionCancelled           WebhookEventType = "TRANSACTION.CANCELLED"
	WebhookEventTransactionFailed              WebhookEventType = "TRANSACTION.FAILED"
	WebhookEventTransactionExpired             WebhookEventType = "TRANSACTION.EXPIRED"
	WebhookEventTransactionFallback            WebhookEventType = "TRANSACTION.FALLBACK"
	WebhookEventTransactionDeadlineApproaching WebhookEventType = "TRANSACTION.DEADLINE_APPROACHING"

	// Step events
	WebhookEventStepStarted               WebhookEventType = "STEP.STARTED"
	WebhookEventStepCompleted             WebhookEventType = "STEP.COMPLETED"
	WebhookEventStepFailed                WebhookEventType = "STEP.FAILED"
	WebhookEventStepPurposeDisclosureSent WebhookEventType = "STEP.PURPOSE_DISCLOSURE_SENT"

	// Tenant-level events
	WebhookEventQuotaWarning   WebhookEventType = "QUOTA.WARNING"
	WebhookEventAPIDeprecation WebhookEventType = "API.DEPRECATION_NOTICE"

	// Signing session events (added in 1.3.0 — new in OpenAPI R0.1)
	WebhookEventSigningSessionCreated   WebhookEventType = "SIGNING_SESSION.CREATED"
	WebhookEventSigningSessionCompleted WebhookEventType = "SIGNING_SESSION.COMPLETED"
	WebhookEventSigningSessionCancelled WebhookEventType = "SIGNING_SESSION.CANCELLED"
	WebhookEventSigningSessionExpired   WebhookEventType = "SIGNING_SESSION.EXPIRED"

	// Deprecated aliases. The original 1.2.x constants used truncated
	// names for the two NT65 events; keep them pointing at the same
	// underlying strings so existing consumer code continues to
	// compile. Prefer the full names above for new code.
	//
	// Deprecated: Use WebhookEventStepPurposeDisclosureSent.
	WebhookEventStepPurposeDisclosure = WebhookEventStepPurposeDisclosureSent
	// Deprecated: Use WebhookEventTransactionDeadlineApproaching.
	WebhookEventTransactionDeadline = WebhookEventTransactionDeadlineApproaching
)

// IsNT65Event reports whether the event is part of the NT65 INSS
// consignado flow and is only emitted for tenants with
// `nt65ComplianceEnabled`. See docs/18-nt65-consignado.md.
func IsNT65Event(e WebhookEventType) bool {
	switch e {
	case WebhookEventTransactionDeadlineApproaching,
		WebhookEventStepPurposeDisclosureSent:
		return true
	default:
		return false
	}
}

// OtpChannel represents the channel for OTP delivery.
type OtpChannel string

const (
	OtpChannelEmail OtpChannel = "email"
	OtpChannelSMS   OtpChannel = "sms"
)

// GeolocationSource represents the source of geolocation data.
type GeolocationSource string

const (
	GeolocationSourceGPS  GeolocationSource = "GPS"
	GeolocationSourceIP   GeolocationSource = "IP"
	GeolocationSourceWIFI GeolocationSource = "WIFI"
	GeolocationSourceCELL GeolocationSource = "CELL"
)

// Geolocation contains geographic coordinates captured during a step.
type Geolocation struct {
	Latitude  float64           `json:"latitude"`
	Longitude float64           `json:"longitude"`
	Accuracy  *float64          `json:"accuracy,omitempty"`
	Source    GeolocationSource `json:"source,omitempty"`
}

// Policy defines the verification policy for a transaction.
type Policy struct {
	Profile     PolicyProfile `json:"profile"`
	CustomSteps []StepType    `json:"customSteps,omitempty"`
}

// Signer identifies the person performing the transaction.
type Signer struct {
	Name           string     `json:"name"`
	Email          string     `json:"email,omitempty"`
	Phone          string     `json:"phone,omitempty"`
	UserExternalID string     `json:"userExternalId"`
	DisplayName    string     `json:"displayName,omitempty"`
	CPF            string     `json:"cpf,omitempty"`
	CNPJ           string     `json:"cnpj,omitempty"`
	BirthDate      string     `json:"birthDate,omitempty"`
	OtpChannel     OtpChannel `json:"otpChannel,omitempty"`
}

// ActionMetadata provides metadata for action authentication transactions.
type ActionMetadata struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Reference   string `json:"reference,omitempty"`
}

// DigitalSignatureMetadata provides configuration for digital certificate signing.
type DigitalSignatureMetadata struct {
	SignatureFieldName string `json:"signatureFieldName,omitempty"`
	SignatureReason    string `json:"signatureReason,omitempty"`
	SignatureLocation  string `json:"signatureLocation,omitempty"`
}

// CreateTransactionRequest is the request body for creating a transaction.
type CreateTransactionRequest struct {
	Purpose          TransactionPurpose        `json:"purpose"`
	Policy           Policy                    `json:"policy"`
	Signer           Signer                    `json:"signer"`
	Document         *DocumentInline           `json:"document,omitempty"`
	Action           *ActionMetadata           `json:"action,omitempty"`
	DigitalSignature *DigitalSignatureMetadata `json:"digitalSignature,omitempty"`
	DocumentGroupID  string                    `json:"documentGroupId,omitempty"`
	SignerIndex      *int                      `json:"signerIndex,omitempty"`
	TotalSigners     *int                      `json:"totalSigners,omitempty"`
	Metadata         map[string]string         `json:"metadata,omitempty"`
	ExpiresInMinutes *int                      `json:"expiresInMinutes,omitempty"`
}

// DocumentInline represents an inline document payload.
type DocumentInline struct {
	Content  string `json:"content"`
	Filename string `json:"filename,omitempty"`
}

// LivenessResult contains biometric liveness check results.
type LivenessResult struct {
	Confidence           float64     `json:"confidence"`
	Provider             string      `json:"provider"`
	CaptureMode          CaptureMode `json:"captureMode"`
	ComplianceStandards  []string    `json:"complianceStandards,omitempty"`
}

// MatchResult contains biometric match results.
type MatchResult struct {
	Similarity float64 `json:"similarity"`
	Threshold  float64 `json:"threshold"`
}

// OTPResult contains OTP verification results.
type OTPResult struct {
	Verified bool   `json:"verified"`
	Channel  string `json:"channel"`
}

// ClickResult contains click acceptance results.
type ClickResult struct {
	Accepted    bool   `json:"accepted"`
	TextVersion string `json:"textVersion"`
}

// DigitalSignatureResult contains digital certificate signing results.
type DigitalSignatureResult struct {
	CertificateSubject string `json:"certificateSubject"`
	CertificateSerial  string `json:"certificateSerial"`
	CertificateIssuer  string `json:"certificateIssuer"`
	Algorithm          string `json:"algorithm"`
	SignedAt           string `json:"signedAt"`
	SignedPDFHash      string `json:"signedPdfHash"`
	SignedPDFS3Key     string `json:"signedPdfS3Key,omitempty"`
	SignatureFieldName string `json:"signatureFieldName"`
}

// PurposeDisclosureResult contains the result of a purpose disclosure step.
type PurposeDisclosureResult struct {
	Acknowledged        bool   `json:"acknowledged"`
	DisclosureTextHash  string `json:"disclosureTextHash"`
	DisclosureVersion   string `json:"disclosureVersion"`
	NotificationChannel string `json:"notificationChannel"`
	NotificationSentAt  string `json:"notificationSentAt,omitempty"`
}

// GovernmentDatabase represents a government database source for identity validation.
type GovernmentDatabase string

const (
	GovernmentDatabaseSerproDatavalid GovernmentDatabase = "SERPRO_DATAVALID"
	GovernmentDatabaseTSE             GovernmentDatabase = "TSE"
	GovernmentDatabaseIDRC            GovernmentDatabase = "IDRC"
)

// GovernmentDbValidation contains the result of a government database validation.
type GovernmentDbValidation struct {
	Database              GovernmentDatabase `json:"database"`
	ValidatedAt           string             `json:"validatedAt"`
	CPFHash               string             `json:"cpfHash"`
	BiometricScore        float64            `json:"biometricScore"`
	Cached                bool               `json:"cached"`
	CacheVerifySimilarity *float64           `json:"cacheVerifySimilarity,omitempty"`
	CacheExpiresAt        string             `json:"cacheExpiresAt,omitempty"`
}

// SerproIdentityResult contains the result of a SERPRO identity check.
type SerproIdentityResult struct {
	Valid                bool                `json:"valid"`
	Provider             string              `json:"provider"`
	NameMatch            bool                `json:"nameMatch"`
	BirthDateMatch       bool                `json:"birthDateMatch"`
	BiometricMatch       bool                `json:"biometricMatch"`
	BiometricConfidence  float64             `json:"biometricConfidence"`
	GovernmentDatabase   GovernmentDatabase  `json:"governmentDatabase,omitempty"`
}

// BiographicValidation contains biographic data validation from document extraction.
type BiographicValidation struct {
	NameMatch       *bool    `json:"nameMatch"`
	CPFMatch        *bool    `json:"cpfMatch"`
	BirthDateMatch  *bool    `json:"birthDateMatch"`
	OverallValid    bool     `json:"overallValid"`
	MatchedFields   []string `json:"matchedFields"`
	UnmatchedFields []string `json:"unmatchedFields"`
}

// DocumentPhotoMatchResult contains the result of a document photo match step.
type DocumentPhotoMatchResult struct {
	DocumentType            string                `json:"documentType"`
	ExtractedFaceHash       string                `json:"extractedFaceHash"`
	Similarity              float64               `json:"similarity"`
	Threshold               float64               `json:"threshold"`
	FaceExtractionConfidence float64              `json:"faceExtractionConfidence"`
	BiographicValidation    *BiographicValidation `json:"biographicValidation,omitempty"`
}

// QualityResult contains image quality assessment results.
type QualityResult struct {
	Brightness    float64 `json:"brightness"`
	Sharpness     float64 `json:"sharpness"`
	FaceAreaRatio float64 `json:"faceAreaRatio"`
}

// StepResult contains the outcome of a completed step.
type StepResult struct {
	Liveness           *LivenessResult           `json:"liveness,omitempty"`
	Match              *MatchResult              `json:"match,omitempty"`
	OTP                *OTPResult                `json:"otp,omitempty"`
	Click              *ClickResult              `json:"click,omitempty"`
	PurposeDisclosure  *PurposeDisclosureResult  `json:"purposeDisclosure,omitempty"`
	DigitalSignature   *DigitalSignatureResult   `json:"digitalSignature,omitempty"`
	SerproIdentity         *SerproIdentityResult  `json:"serproIdentity,omitempty"`
	GovernmentDbValidation *GovernmentDbValidation `json:"governmentDbValidation,omitempty"`
	Geolocation            *Geolocation            `json:"geolocation,omitempty"`
	DocumentPhotoMatch     *DocumentPhotoMatchResult `json:"documentPhotoMatch,omitempty"`
	Quality            *QualityResult            `json:"quality,omitempty"`
	ProviderTimestamp  string                    `json:"providerTimestamp,omitempty"`
}

// Step represents a single verification step in a transaction.
type Step struct {
	TenantID      string      `json:"tenantId"`
	TransactionID string      `json:"transactionId"`
	StepID        string      `json:"stepId"`
	Type          StepType    `json:"type"`
	Status        StepStatus  `json:"status"`
	Order         int         `json:"order"`
	Attempts      int         `json:"attempts"`
	MaxAttempts   int         `json:"maxAttempts"`
	CaptureMode   CaptureMode `json:"captureMode,omitempty"`
	StartedAt     string      `json:"startedAt,omitempty"`
	CompletedAt   string      `json:"completedAt,omitempty"`
	Result        *StepResult `json:"result,omitempty"`
	Error         string      `json:"error,omitempty"`
}

// Transaction represents a signing or authentication transaction.
type Transaction struct {
	TenantID           string             `json:"tenantId"`
	TransactionID      string             `json:"transactionId"`
	Status             TransactionStatus  `json:"status"`
	Purpose            TransactionPurpose `json:"purpose"`
	Policy             Policy             `json:"policy"`
	Signer             Signer             `json:"signer"`
	Steps              []Step             `json:"steps"`
	DocumentGroupID    string             `json:"documentGroupId,omitempty"`
	SignerIndex        *int               `json:"signerIndex,omitempty"`
	TotalSigners       *int               `json:"totalSigners,omitempty"`
	Metadata           map[string]string  `json:"metadata,omitempty"`
	ExpiresAt          string             `json:"expiresAt"`
	CreatedAt          string             `json:"createdAt"`
	UpdatedAt          string             `json:"updatedAt"`
	SubmissionDeadline string             `json:"submissionDeadline,omitempty"`
	DeadlineStatus     string             `json:"deadlineStatus,omitempty"`
}

// TransactionListParams are query parameters for listing transactions.
type TransactionListParams struct {
	Status          TransactionStatus `url:"status,omitempty"`
	UserExternalID  string            `url:"userExternalId,omitempty"`
	DocumentGroupID string            `url:"documentGroupId,omitempty"`
	StartDate       string            `url:"startDate,omitempty"`
	EndDate         string            `url:"endDate,omitempty"`
	Limit           int               `url:"limit,omitempty"`
	NextToken       string            `url:"nextToken,omitempty"`
}

// TransactionListResponse is the paginated response for listing transactions.
type TransactionListResponse struct {
	Transactions []Transaction `json:"transactions"`
	NextToken    string        `json:"nextToken,omitempty"`
	Count        int           `json:"count"`
}

// UploadDocumentRequest is the request body for uploading a document.
type UploadDocumentRequest struct {
	Content  string `json:"content"`
	Filename string `json:"filename,omitempty"`
}

// PresignResponse contains the presigned upload URL and parameters.
type PresignResponse struct {
	UploadURL    string `json:"uploadUrl"`
	UploadToken  string `json:"uploadToken"`
	S3Key        string `json:"s3Key"`
	ExpiresIn    int    `json:"expiresIn"`
	ContentType  string `json:"contentType"`
	Instructions string `json:"instructions"`
}

// ConfirmDocumentRequest confirms a presigned upload.
type ConfirmDocumentRequest struct {
	UploadToken string `json:"uploadToken"`
}

// DownloadResponse contains the presigned download URL.
type DownloadResponse struct {
	TransactionID string `json:"transactionId"`
	DocumentHash  string `json:"documentHash,omitempty"`
	OriginalURL   string `json:"originalUrl,omitempty"`
	SignedURL     string `json:"signedUrl,omitempty"`
	ExpiresIn     int    `json:"expiresIn"`
}

// StartStepRequest is the optional request body for starting a step.
type StartStepRequest struct {
	CaptureMode CaptureMode `json:"captureMode,omitempty"`
	OtpChannel  string      `json:"otpChannel,omitempty"`
}

// StartStepResponse is returned when a step is started.
type StartStepResponse struct {
	StepID            string `json:"stepId"`
	Type              string `json:"type"`
	Status            string `json:"status"`
	LivenessSessionID string `json:"livenessSessionId,omitempty"`
	HostedURL         string `json:"hostedUrl,omitempty"`
	Message           string `json:"message,omitempty"`
	OTPCode           string `json:"otpCode,omitempty"`
}

// CompleteStepRequest is the request body for completing a step.
// Use the specific request types (CompleteClickRequest, etc.) and pass them here.
type CompleteStepRequest map[string]any

// CompleteClickRequest is the request body for completing a click-accept step.
type CompleteClickRequest struct {
	Accepted    bool   `json:"accepted"`
	TextVersion string `json:"textVersion,omitempty"`
}

// CompleteOTPRequest is the request body for completing an OTP step.
type CompleteOTPRequest struct {
	Code string `json:"code"`
}

// CompleteLivenessRequest is the request body for completing a liveness step.
type CompleteLivenessRequest struct {
	LivenessSessionID string       `json:"livenessSessionId"`
	Geolocation       *Geolocation `json:"geolocation,omitempty"`
}

// ReferenceImage defines a base64-encoded reference image for biometric match.
type ReferenceImage struct {
	Source string `json:"source"`
	Data   string `json:"data"`
}

// CompleteBiometricMatchRequest is the request body for completing a biometric match step.
type CompleteBiometricMatchRequest struct {
	ReferenceImage    *ReferenceImage `json:"referenceImage,omitempty"`
	SandboxSimilarity *float64        `json:"sandboxSimilarity,omitempty"`
	Geolocation       *Geolocation    `json:"geolocation,omitempty"`
}

// CompletePurposeDisclosureRequest is the request body for completing a purpose disclosure step.
type CompletePurposeDisclosureRequest struct {
	Acknowledged bool `json:"acknowledged"`
}

// CompleteDocumentPhotoMatchRequest is the request body for completing a document photo match step.
type CompleteDocumentPhotoMatchRequest struct {
	DocumentImage string       `json:"documentImage"`
	DocumentType  string       `json:"documentType"`
	Geolocation   *Geolocation `json:"geolocation,omitempty"`
}

// CompleteStepResponse is the response when a step is completed.
type CompleteStepResponse = Step

// PrepareSigningRequest is the request body for preparing a digital signing operation.
type PrepareSigningRequest struct {
	CertificateChainPEMs []string `json:"certificateChainPems"`
}

// PrepareSigningResponse contains the hash to sign.
type PrepareSigningResponse struct {
	SignatureRequestID string `json:"signatureRequestId"`
	HashToSign         string `json:"hashToSign"`
	HashAlgorithm      string `json:"hashAlgorithm"`
	SignatureAlgorithm string `json:"signatureAlgorithm"`
}

// CompleteSigningRequest provides the raw signature to finalize signing.
type CompleteSigningRequest struct {
	SignatureRequestID string `json:"signatureRequestId"`
	RawSignatureBase64 string `json:"rawSignatureBase64"`
}

// CompleteSigningDigitalSignatureResult is the digital signature info in the response.
type CompleteSigningDigitalSignatureResult struct {
	CertificateSubject string `json:"certificateSubject"`
	CertificateSerial  string `json:"certificateSerial"`
	CertificateIssuer  string `json:"certificateIssuer"`
	Algorithm          string `json:"algorithm"`
	SignedAt           string `json:"signedAt"`
	SignedPDFHash      string `json:"signedPdfHash"`
	SignatureFieldName string `json:"signatureFieldName"`
}

// CompleteSigningResult wraps the digital signature result.
type CompleteSigningResult struct {
	DigitalSignature CompleteSigningDigitalSignatureResult `json:"digitalSignature"`
}

// CompleteSigningResponse is returned when signing is completed.
type CompleteSigningResponse struct {
	StepID string                `json:"stepId"`
	Status string                `json:"status"`
	Result CompleteSigningResult `json:"result"`
}

// RegisterWebhookRequest is the request body for registering a webhook.
type RegisterWebhookRequest struct {
	URL    string             `json:"url"`
	Events []WebhookEventType `json:"events"`
}

// RegisterWebhookResponse is returned when a webhook is registered.
type RegisterWebhookResponse struct {
	WebhookID string             `json:"webhookId"`
	URL       string             `json:"url"`
	Secret    string             `json:"secret"`
	Events    []WebhookEventType `json:"events"`
	Status    string             `json:"status"`
	CreatedAt string             `json:"createdAt"`
}

// Webhook represents a registered webhook endpoint.
type Webhook struct {
	WebhookID string             `json:"webhookId"`
	URL       string             `json:"url"`
	Events    []WebhookEventType `json:"events"`
	Status    string             `json:"status"`
	CreatedAt string             `json:"createdAt"`
}

// WebhookPayload is the body of a webhook delivery.
type WebhookPayload struct {
	ID            string           `json:"id"`
	EventType     WebhookEventType `json:"eventType"`
	TenantID      string           `json:"tenantId"`
	TransactionID string           `json:"transactionId,omitempty"`
	Timestamp     string           `json:"timestamp"`
	Data          map[string]any   `json:"data"`
	Test          bool             `json:"test,omitempty"`
}

// WebhookTestDelivery describes the outcome of a single test webhook delivery.
type WebhookTestDelivery struct {
	HTTPStatus int    `json:"httpStatus"`
	Success    bool   `json:"success"`
	Timestamp  string `json:"timestamp"`
	Error      string `json:"error,omitempty"`
}

// WebhookTestResponse is returned when testing a webhook.
type WebhookTestResponse struct {
	WebhookID    string              `json:"webhookId"`
	TestDelivery WebhookTestDelivery `json:"testDelivery"`
}

// EvidenceSigner contains the signer info within evidence.
type EvidenceSigner struct {
	Name           string `json:"name"`
	CPF            string `json:"cpf,omitempty"`
	CNPJ           string `json:"cnpj,omitempty"`
	UserExternalID string `json:"userExternalId"`
}

// EvidenceDocument contains document info within evidence.
type EvidenceDocument struct {
	Hash     string `json:"hash"`
	Filename string `json:"filename"`
}

// EvidenceStep is a step record within evidence.
type EvidenceStep struct {
	Type        string         `json:"type"`
	Status      string         `json:"status"`
	CompletedAt string         `json:"completedAt,omitempty"`
	Result      map[string]any `json:"result,omitempty"`
}

// Evidence is the audit evidence for a completed transaction.
type Evidence struct {
	TenantID      string         `json:"tenantId"`
	TransactionID string         `json:"transactionId"`
	EvidenceID    string         `json:"evidenceId"`
	Status        string         `json:"status"`
	Signer        EvidenceSigner `json:"signer"`
	Steps         []EvidenceStep `json:"steps"`
	Document      *EvidenceDocument `json:"document,omitempty"`
	CreatedAt     string         `json:"createdAt"`
	CompletedAt   string         `json:"completedAt,omitempty"`
}

// VerificationSigner contains the signer info within a verification response.
type VerificationSigner struct {
	DisplayName string `json:"displayName,omitempty"`
	CPFCNPJ     string `json:"cpfCnpj,omitempty"`
}

// VerificationStep represents a single step within a verification response.
type VerificationStep struct {
	Type        string `json:"type"`
	Status      string `json:"status"`
	Order       int    `json:"order"`
	CompletedAt string `json:"completedAt,omitempty"`
}

// VerificationResponse is returned when verifying evidence.
type VerificationResponse struct {
	EvidenceID    string              `json:"evidenceId"`
	Status        string              `json:"status"`
	TransactionID string              `json:"transactionId"`
	Purpose       string              `json:"purpose"`
	DocumentHash  string              `json:"documentHash,omitempty"`
	EvidenceHash  string              `json:"evidenceHash"`
	Policy        *Policy             `json:"policy"`
	Steps         []VerificationStep  `json:"steps"`
	Signer        *VerificationSigner `json:"signer,omitempty"`
	TenantName    string              `json:"tenantName,omitempty"`
	TenantCNPJ    string              `json:"tenantCnpj,omitempty"`
	CreatedAt     string              `json:"createdAt"`
	CompletedAt   string              `json:"completedAt"`
}

// DownloadArtifact represents a single downloadable artifact.
type DownloadArtifact struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

// VerificationDownloads contains the available download artifacts.
//
// SignedSignature is the detached PKCS#7 / CMS (.p7s) for digital-cert
// signing of non-PDF documents. It is only populated by the API for
// standalone signing sessions (single-signer); it is omitted entirely
// from the response when the evidence belongs to a multi-signer
// envelope — use VerificationService.VerifyEnvelope to retrieve the
// consolidated envelope-level .p7s instead.
type VerificationDownloads struct {
	OriginalDocument *DownloadArtifact `json:"originalDocument"`
	EvidencePack     *DownloadArtifact `json:"evidencePack"`
	FinalPDF         *DownloadArtifact `json:"finalPdf"`
	SignedSignature  *DownloadArtifact `json:"signedSignature,omitempty"`
}

// VerificationDownloadsResponse contains download URLs for evidence artifacts.
type VerificationDownloadsResponse struct {
	EvidenceID string                `json:"evidenceId"`
	Downloads  VerificationDownloads `json:"downloads"`
}

// EnvelopeVerificationSigner is a per-signer entry within an envelope
// verification response.
type EnvelopeVerificationSigner struct {
	SignerIndex   int    `json:"signerIndex"`
	DisplayName   string `json:"displayName"`
	CPFCNPJ       string `json:"cpfCnpj,omitempty"`
	Status        string `json:"status"`
	PolicyProfile string `json:"policyProfile,omitempty"`
	EvidenceID    string `json:"evidenceId,omitempty"`
	CompletedAt   string `json:"completedAt,omitempty"`
}

// EnvelopeVerificationDownloads contains envelope-level consolidated downloads.
//
// CombinedSignedPDF is present only for PDF envelopes; ConsolidatedSignature
// is the merged .p7s containing every signer's SignerInfo for non-PDF
// envelopes signed with digital certificates.
type EnvelopeVerificationDownloads struct {
	CombinedSignedPDF     *DownloadArtifact `json:"combinedSignedPdf,omitempty"`
	ConsolidatedSignature *DownloadArtifact `json:"consolidatedSignature,omitempty"`
}

// EnvelopeVerificationResponse is returned when verifying a multi-signer
// envelope via GET /v1/verify/envelope/{envelopeId}.
type EnvelopeVerificationResponse struct {
	EnvelopeID        string                         `json:"envelopeId"`
	Status            string                         `json:"status"`
	SigningMode       string                         `json:"signingMode"`
	TotalSigners      int                            `json:"totalSigners"`
	CompletedSessions int                            `json:"completedSessions"`
	DocumentHash      string                         `json:"documentHash"`
	TenantName        string                         `json:"tenantName,omitempty"`
	TenantCNPJ        string                         `json:"tenantCnpj,omitempty"`
	Signers           []EnvelopeVerificationSigner   `json:"signers"`
	Downloads         *EnvelopeVerificationDownloads `json:"downloads,omitempty"`
	CreatedAt         string                         `json:"createdAt"`
	CompletedAt       string                         `json:"completedAt,omitempty"`
}

// EnrollUserRequest is the request body for enrolling a user's biometric data.
type EnrollUserRequest struct {
	Image  string `json:"image"`           // base64 JPEG (required)
	CPF    string `json:"cpf"`             // 11 digits (required)
	Source string `json:"source,omitempty"` // BANK_PROVIDED, FIRST_LIVENESS, DOCUMENT_PHOTO
}

// EnrollUserResponse is returned when enrollment succeeds.
type EnrollUserResponse struct {
	UserExternalID       string  `json:"userExternalId"`
	EnrollmentHash       string  `json:"enrollmentHash"`
	EnrollmentVersion    int     `json:"enrollmentVersion"`
	EnrollmentSource     string  `json:"enrollmentSource"`
	EnrolledAt           string  `json:"enrolledAt"`
	CPF                  string  `json:"cpf"`
	FaceConfidence       float64 `json:"faceConfidence"`
	DocumentImageHash    string  `json:"documentImageHash,omitempty"`
	ExtractionConfidence float64 `json:"extractionConfidence,omitempty"`
}

// ServiceHealth contains health info for a single backend service.
type ServiceHealth struct {
	Status  string `json:"status"`
	Latency *int   `json:"latency,omitempty"`
}

// HealthCheckResponse is the response from the health endpoint.
type HealthCheckResponse struct {
	Status    string                   `json:"status"`
	Version   string                   `json:"version"`
	Timestamp string                   `json:"timestamp"`
	Services  map[string]ServiceHealth `json:"services,omitempty"`
}

// HealthHistoryResponse is the response from the health history endpoint.
type HealthHistoryResponse struct {
	Entries []HealthCheckResponse `json:"entries"`
}

// CombinedStampResponse is the response when requesting a combined stamp for a document group.
type CombinedStampResponse struct {
	GroupID     string `json:"groupId"`
	SignerCount int    `json:"signerCount"`
	DownloadURL string `json:"downloadUrl"`
	ExpiresIn   int    `json:"expiresIn"`
}

// PresignRequest is the request body for generating a presigned upload URL.
type PresignRequest struct {
	ContentType string `json:"contentType"`
	Filename    string `json:"filename"`
}

// CancelTransactionResponse is returned when a transaction is cancelled.
type CancelTransactionResponse struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	CancelledAt   string `json:"cancelledAt"`
}

// FinalizeResponse is returned when a transaction is finalized.
type FinalizeResponse struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	EvidenceID    string `json:"evidenceId"`
	EvidenceHash  string `json:"evidenceHash"`
	CompletedAt   string `json:"completedAt"`
}

// DocumentUploadResponse is returned when a document is uploaded inline.
type DocumentUploadResponse struct {
	TransactionID string `json:"transactionId"`
	DocumentHash  string `json:"documentHash"`
	Status        string `json:"status"`
	UploadedAt    string `json:"uploadedAt"`
}

// StepListResponse is the response for listing steps.
type StepListResponse struct {
	Steps []StepDetail `json:"steps"`
}

// StepDetail represents a step in the step list response.
type StepDetail struct {
	StepID      string `json:"stepId"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Order       int    `json:"order"`
	Attempts    int    `json:"attempts"`
	MaxAttempts int    `json:"maxAttempts"`
	CaptureMode string `json:"captureMode,omitempty"`
	StartedAt   string `json:"startedAt,omitempty"`
	CompletedAt string `json:"completedAt,omitempty"`
	Error       string `json:"error,omitempty"`
}

// StepCompleteResponse is returned when a step is completed.
type StepCompleteResponse struct {
	StepID   string      `json:"stepId"`
	Type     string      `json:"type"`
	Status   string      `json:"status"`
	Attempts int         `json:"attempts"`
	Result   *StepResult `json:"result,omitempty"`
}

// ConfirmDocumentResponse is returned when a presigned upload is confirmed.
type ConfirmDocumentResponse struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	DocumentHash  string `json:"documentHash"`
}
